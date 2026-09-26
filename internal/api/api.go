// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package api implements the public read API of p(DOOM) described in
// api/openapi.yaml.
//
// Handlers only serve files of the current promoted release and of the data
// snapshot that release was computed from; they never compute or alter an
// estimate. The two write endpoints append a submission to the review queue or
// evaluate a user scenario with the experimental model (the result is labelled
// user_scenario and nothing on disk changes). There are no administrative
// endpoints: promotion, rollback and review are CLI-only (pdoomctl, ADR-003).
package api

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
)

// Limits and defaults of the public API.
const (
	// MaxBodyBytes bounds every request body (submissions and scenario lab).
	MaxBodyBytes = 64 * 1024
	// MaxURILength bounds the request line.
	MaxURILength = 2048
	// MaxSamples is the largest Monte Carlo sample count the scenario lab accepts.
	MaxSamples = 50000
	// MaxSeed is the largest seed the scenario lab accepts (uint32 range).
	MaxSeed = 4294967295
	// MaxPageLimit is the largest page size of paginated routes.
	MaxPageLimit = 200
	// DefaultPageLimit is the page size when limit is absent.
	DefaultPageLimit = 50
	// DefaultRefreshInterval bounds how often data/releases/CURRENT is re-checked.
	DefaultRefreshInterval = 5 * time.Second
	// CacheMaxAge is the max-age of public GET responses in seconds.
	CacheMaxAge = 60
)

// Rate-limit classes.
const (
	ClassRead   = "read"
	ClassLab    = "lab"
	ClassSubmit = "submit"
)

// Limit is a token-bucket rate limit: Burst tokens available at once, refilled
// at PerMinute tokens per minute.
type Limit struct {
	Burst     int
	PerMinute float64
}

// RateLimits holds the per-client limits of each route class.
type RateLimits struct {
	Read   Limit
	Lab    Limit
	Submit Limit
}

// DefaultRateLimits are used for zero-valued fields of Deps.RateLimits.
var DefaultRateLimits = RateLimits{
	Read:   Limit{Burst: 120, PerMinute: 600},
	Lab:    Limit{Burst: 10, PerMinute: 30},
	Submit: Limit{Burst: 5, PerMinute: 10},
}

// Deps are the dependencies of the handler.
type Deps struct {
	// Paths locates the data directory (releases, snapshots, review queue).
	Paths config.Paths
	// Logger receives the access log and loader messages. Nil discards.
	Logger *slog.Logger
	// Now is the clock used for rate limiting, refresh checks, access-log
	// durations and submission timestamps. Nil means time.Now. Nothing served
	// from a release depends on it.
	Now func() time.Time
	// MethodDir holds the methodology markdown files. Empty resolves to
	// <DataDir>/../docs/method.
	MethodDir string
	// CORSOrigins allowlists browser origins for POST routes (PDOOM_CORS_ORIGINS).
	CORSOrigins []string
	// TrustProxy derives the client key from X-Forwarded-For (PDOOM_TRUST_PROXY=1).
	TrustProxy bool
	// RateLimits overrides DefaultRateLimits; zero fields keep the defaults.
	RateLimits RateLimits
	// RefreshInterval overrides DefaultRefreshInterval when positive.
	RefreshInterval time.Duration
}

// handlerFunc is the signature of every route handler. st is nil for routes
// that do not need a loaded release (and for releaseOptional routes while no
// release is loaded).
type handlerFunc func(w http.ResponseWriter, r *http.Request, st *state)

// releaseNeed says how a route relates to the loaded release state.
type releaseNeed int

const (
	// releaseNone: the route never looks at the release (static vocabularies,
	// health, methodology documents, submissions).
	releaseNone releaseNeed = iota
	// releaseOptional: the route serves from the loaded state when there is
	// one and falls back to disk otherwise (release listing and detail).
	releaseOptional
	// releaseRequired: 503 until a release is loaded.
	releaseRequired
)

// route is one row of the route table.
type route struct {
	method  string
	pattern string
	class   string
	release releaseNeed
	// query lists the query parameters that influence the body. Only they
	// enter the response-cache key; every other parameter is ignored, so a
	// client cannot grow the cache by appending junk queries.
	query   []string
	handler handlerFunc
}

type server struct {
	deps      Deps
	log       *slog.Logger
	now       func() time.Time
	loader    *loader
	limits    map[string]*limiter
	origins   map[string]bool
	salt      []byte
	methodDir string
	cache     *bodyCache
	submitMu  sync.Mutex
	patterns  []string
}

func newServer(d Deps) *server {
	s := &server{deps: d, log: d.Logger, now: d.Now, origins: map[string]bool{}, cache: newBodyCache()}
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}
	if s.now == nil {
		s.now = time.Now
	}
	s.salt = randomBytes(16)
	interval := d.RefreshInterval
	if interval <= 0 {
		interval = DefaultRefreshInterval
	}
	s.loader = newLoader(d.Paths, s.log, s.now, interval)
	s.methodDir = d.MethodDir
	if s.methodDir == "" {
		s.methodDir = filepath.Join(filepath.Dir(filepath.Clean(d.Paths.DataDir)), "docs", "method")
	}
	for _, o := range d.CORSOrigins {
		if n := normalizeOrigin(o); n != "" {
			s.origins[n] = true
		}
	}
	rl := d.RateLimits
	s.limits = map[string]*limiter{
		ClassRead:   newLimiter(orLimit(rl.Read, DefaultRateLimits.Read), s.now),
		ClassLab:    newLimiter(orLimit(rl.Lab, DefaultRateLimits.Lab), s.now),
		ClassSubmit: newLimiter(orLimit(rl.Submit, DefaultRateLimits.Submit), s.now),
	}
	return s
}

func orLimit(l, def Limit) Limit {
	if l.Burst <= 0 {
		l.Burst = def.Burst
	}
	if l.PerMinute <= 0 {
		l.PerMinute = def.PerMinute
	}
	return l
}

// routeTable lists every route of the API. The table is the single source of
// truth: NewHandler registers it and the route test compares it with
// api/openapi.yaml.
func (s *server) routeTable() []route {
	get := func(pattern string, need releaseNeed, h handlerFunc) route {
		return route{method: http.MethodGet, pattern: pattern, class: ClassRead, release: need, handler: h}
	}
	sources := get("/v1/sources", releaseRequired, s.handleSources)
	sources.query = sourcesQueryKeys
	return []route{
		get("/healthz", releaseNone, s.handleHealthz),
		get("/readyz", releaseNone, s.handleReadyz),
		get("/v1/meter", releaseRequired, s.handleMeter),
		get("/v1/meter/history", releaseRequired, s.handleMeterHistory),
		get("/v1/outcomes", releaseNone, s.handleOutcomes),
		get("/v1/horizons", releaseNone, s.handleHorizons),
		get("/v1/drivers", releaseRequired, s.handleDrivers),
		get("/v1/scenarios", releaseRequired, s.handleScenarios),
		get("/v1/scenarios/{id}", releaseRequired, s.handleScenario),
		get("/v1/forecasts", releaseRequired, s.handleForecasts),
		get("/v1/capabilities", releaseRequired, s.handleCapabilities),
		get("/v1/incidents", releaseRequired, s.handleIncidents),
		get("/v1/safeguards", releaseRequired, s.handleSafeguards),
		sources,
		get("/v1/sources/{id}", releaseRequired, s.handleSource),
		get("/v1/methodology", releaseNone, s.handleMethodology),
		get("/v1/methodology/{slug}", releaseNone, s.handleMethodologyDoc),
		get("/v1/releases", releaseOptional, s.handleReleases),
		get("/v1/releases/{id}", releaseOptional, s.handleRelease),
		get("/v1/releases/{id}/{doc}", releaseOptional, s.handleReleaseDocument),
		get("/v1/snapshot", releaseRequired, s.handleSnapshot),
		get("/v1/definitions", releaseRequired, s.handleDefinitions),
		get("/v1/organizations", releaseRequired, s.handleOrganizations),
		get("/v1/actions", releaseRequired, s.handleActions),
		{method: http.MethodPost, pattern: "/v1/scenario-lab/evaluate", class: ClassLab, release: releaseRequired, handler: s.handleScenarioLab},
		{method: http.MethodPost, pattern: "/v1/submissions/sources", class: ClassSubmit, release: releaseNone, handler: s.handleSubmitSource},
		{method: http.MethodPost, pattern: "/v1/submissions/corrections", class: ClassSubmit, release: releaseNone, handler: s.handleSubmitCorrection},
	}
}

// sourcesQueryKeys are the query parameters GET /v1/sources reads.
var sourcesQueryKeys = []string{"tier", "topic", "q", "limit", "offset"}

// Routes returns the registered "METHOD /path" pairs, sorted. It exists for
// tests and for the CLI's --routes flag.
func Routes(d Deps) []string {
	s := newServer(d)
	var out []string
	for _, rt := range s.routeTable() {
		out = append(out, rt.method+" "+rt.pattern)
	}
	sort.Strings(out)
	return out
}

// NewHandler builds the HTTP handler. The current release is loaded on the
// first request that needs it and re-checked at most once per refresh
// interval, so a promotion is picked up without a restart.
func NewHandler(d Deps) http.Handler {
	return newServer(d).handler()
}

// handler assembles the mux and warms the loader.
func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	byPattern := map[string]map[string]route{}
	for _, rt := range s.routeTable() {
		if byPattern[rt.pattern] == nil {
			byPattern[rt.pattern] = map[string]route{}
		}
		byPattern[rt.pattern][rt.method] = rt
	}
	for pattern, methods := range byPattern {
		s.patterns = append(s.patterns, pattern)
		mux.HandleFunc(pattern, s.dispatch(methods))
	}
	sort.Strings(s.patterns)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no such route; see api/openapi.yaml")
	})
	// Warm the loader so /readyz reflects the state at start; failures are
	// logged and retried on the next request.
	if _, err := s.loader.current(); err != nil {
		s.log.Warn("no release loaded at start", "error", err.Error())
	}
	return s.middleware(mux)
}

// dispatch routes one pattern: CORS, method check, rate limit, release state
// and the response cache, then the handler.
func (s *server) dispatch(methods map[string]route) http.HandlerFunc {
	var names []string
	for m := range methods {
		names = append(names, m)
	}
	if _, ok := methods[http.MethodGet]; ok {
		names = append(names, http.MethodHead)
	}
	names = append(names, http.MethodOptions)
	sort.Strings(names)
	allow := strings.Join(names, ", ")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			s.preflight(w, r, methods, allow)
			return
		}
		method := r.Method
		if method == http.MethodHead {
			method = http.MethodGet
		}
		rt, ok := methods[method]
		if !ok {
			w.Header().Set("Allow", allow)
			s.writeError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "allowed: "+allow)
			return
		}
		if !s.applyCORS(w, r, rt) {
			return
		}
		if !s.rateLimit(w, r, rt.class) {
			return
		}
		var st *state
		switch rt.release {
		case releaseRequired:
			var err error
			if st, err = s.loader.current(); err != nil {
				s.writeError(w, r, http.StatusServiceUnavailable, "unavailable", "no promoted release is loaded")
				return
			}
		case releaseOptional:
			st, _ = s.loader.current()
		}
		if r.Method == http.MethodHead {
			w = headWriter{w}
		}
		if st != nil && method == http.MethodGet {
			key := cacheKey(r, rt.query)
			if body, ok := s.cache.get(st, key); ok {
				s.serveBody(w, r, body, contentTypeJSON)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), cacheKeyContext{}, key))
		}
		rt.handler(w, r, st)
	}
}

// cacheKeyContext carries the response-cache key from dispatch to ok.
type cacheKeyContext struct{}

// cacheKey is the response-cache key of a GET: the path plus, in a fixed
// order, only the query parameters the route reads. Parameters a route
// ignores never reach the key.
func cacheKey(r *http.Request, keys []string) string {
	if len(keys) == 0 || r.URL.RawQuery == "" {
		return r.URL.Path
	}
	q := r.URL.Query()
	var b strings.Builder
	b.WriteString(r.URL.Path)
	sep := "?"
	for _, k := range keys {
		if !q.Has(k) {
			continue
		}
		b.WriteString(sep)
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(url.QueryEscape(q.Get(k)))
		sep = "&"
	}
	return b.String()
}

// cacheKeyOf returns the key dispatch computed for this request, or "" when
// the response must not be cached.
func cacheKeyOf(r *http.Request) string {
	k, _ := r.Context().Value(cacheKeyContext{}).(string)
	return k
}
