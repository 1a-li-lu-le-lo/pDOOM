// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/publishing"
)

// writeCorruptRelease creates a release directory whose manifest is not JSON,
// the shape an interrupted promotion or a hand edit leaves behind.
func writeCorruptRelease(t *testing.T, e *env, id string) string {
	t.Helper()
	dir := filepath.Join(e.paths.Releases, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeDriftedRelease copies the first release into a new directory and adds
// an unknown manifest field, the shape of schema drift in an old release.
func writeDriftedRelease(t *testing.T, e *env, id string) {
	t.Helper()
	src := filepath.Join(e.paths.Releases, firstRel)
	dst := filepath.Join(e.paths.Releases, id)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, en := range entries {
		raw, err := os.ReadFile(filepath.Join(src, en.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if en.Name() == "manifest.json" {
			raw = append([]byte(`{"legacy_field": true,`), bytes.TrimSpace(raw)[1:]...)
		}
		if err := os.WriteFile(filepath.Join(dst, en.Name()), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// writeUnpublishedRelease writes a candidate-shaped directory (no published
// field) under data/releases, as a crash between copyDir and writeManifest in
// publishing.Promote would.
func writeUnpublishedRelease(t *testing.T, e *env, releaseID, candidateID string) {
	t.Helper()
	res, err := model.Run(e.snap, model.RunOptions{GeneratedAt: genAt, CodeCommit: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	m, err := publishing.BuildManifest(res, e.snap, publishing.BuildOptions{ReleaseID: releaseID, CandidateID: candidateID, CodeCommit: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publishing.WriteCandidate(filepath.Join(e.paths.Releases, releaseID), res, m); err != nil {
		t.Fatal(err)
	}
}

func releaseIDs(t *testing.T, h http.Handler) (history, list []string) {
	t.Helper()
	var hist HistoryResponse
	rr := do(h, http.MethodGet, "/v1/meter/history", nil, nil)
	if rr.Code != 200 {
		t.Fatalf("history: %d %s", rr.Code, rr.Body.String())
	}
	decode(t, rr, &hist)
	for _, r := range hist.Releases {
		history = append(history, r.ReleaseID)
	}
	var rl ReleasesResponse
	rr = do(h, http.MethodGet, "/v1/releases", nil, nil)
	if rr.Code != 200 {
		t.Fatalf("releases: %d %s", rr.Code, rr.Body.String())
	}
	decode(t, rr, &rl)
	for _, r := range rl.Releases {
		list = append(list, r.ReleaseID)
	}
	return history, list
}

func TestDamagedSiblingReleaseDoesNotBlockCurrent(t *testing.T) {
	e := buildDataDir(t)
	writeCorruptRelease(t, e, "rel-2026-09-25-001")
	writeDriftedRelease(t, e, "rel-2026-09-20-001")
	h := e.handler(t, nil)
	rr := do(h, http.MethodGet, "/readyz", nil, nil)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), firstRel) {
		t.Fatalf("readyz with a damaged sibling: %d %s", rr.Code, rr.Body.String())
	}
	var m MeterResponse
	rr = do(h, http.MethodGet, "/v1/meter", nil, nil)
	if rr.Code != 200 {
		t.Fatalf("meter with a damaged sibling: %d %s", rr.Code, rr.Body.String())
	}
	decode(t, rr, &m)
	if m.Meta.ReleaseID != firstRel {
		t.Fatalf("meter serves %s", m.Meta.ReleaseID)
	}
	history, list := releaseIDs(t, h)
	if len(history) != 1 || history[0] != firstRel || len(list) != 1 || list[0] != firstRel {
		t.Fatalf("damaged directories must be absent: history %v list %v", history, list)
	}
	for _, id := range []string{"rel-2026-09-25-001", "rel-2026-09-20-001"} {
		expectError(t, do(h, http.MethodGet, "/v1/releases/"+id, nil, nil), http.StatusNotFound, "not_found")
	}
	logs := e.logs.String()
	if !strings.Contains(logs, "skipping unreadable release directory") || !strings.Contains(logs, "rel-2026-09-25-001") || !strings.Contains(logs, "rel-2026-09-20-001") {
		t.Fatalf("skipped directories should be logged:\n%s", logs)
	}
	if strings.Contains(logs, "release reload failed") {
		t.Fatalf("a damaged sibling must not fail the load:\n%s", logs)
	}
}

func TestUnpublishedDirectoryIsAbsentFromHistoryAndList(t *testing.T) {
	e := buildDataDir(t)
	writeUnpublishedRelease(t, e, "rel-2026-09-27-001", "cand-2026-09-27-001")
	h := e.handler(t, nil)
	history, list := releaseIDs(t, h)
	if len(history) != 1 || history[0] != firstRel || len(list) != 1 || list[0] != firstRel {
		t.Fatalf("unpublished directory leaked: history %v list %v", history, list)
	}
	expectError(t, do(h, http.MethodGet, "/v1/releases/rel-2026-09-27-001", nil, nil), http.StatusNotFound, "not_found")
	if !strings.Contains(e.logs.String(), "not marked published") {
		t.Fatalf("skip should be logged:\n%s", e.logs.String())
	}
}

func TestReleaseRoutesWithoutStateFallBackToDisk(t *testing.T) {
	e := buildDataDir(t)
	writeUnpublishedRelease(t, e, "rel-2026-09-27-001", "cand-2026-09-27-001")
	writeCorruptRelease(t, e, "rel-2026-09-25-001")
	if err := os.Remove(e.paths.Current); err != nil {
		t.Fatal(err)
	}
	h := e.handler(t, nil)
	if rr := do(h, http.MethodGet, "/readyz", nil, nil); rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz without CURRENT: %d", rr.Code)
	}
	var rl ReleasesResponse
	rr := do(h, http.MethodGet, "/v1/releases", nil, nil)
	if rr.Code != 200 {
		t.Fatalf("releases: %d %s", rr.Code, rr.Body.String())
	}
	decode(t, rr, &rl)
	if rl.Current != "" || len(rl.Releases) != 1 || rl.Releases[0].ReleaseID != firstRel || rl.Releases[0].IsCurrent {
		t.Fatalf("disk fallback must list only published, readable releases: %+v", rl)
	}
	if rr.Header().Get("ETag") == "" {
		t.Fatal("fallback body should still carry an ETag")
	}
	var rel ReleaseResponse
	rr = do(h, http.MethodGet, "/v1/releases/"+firstRel, nil, nil)
	if rr.Code != 200 {
		t.Fatalf("release detail: %d %s", rr.Code, rr.Body.String())
	}
	decode(t, rr, &rel)
	if rel.IsCurrent || rel.Release == nil || rel.Release.Manifest.ReleaseID != firstRel {
		t.Fatalf("release detail: %+v", rel.Meta)
	}
	for _, id := range []string{"rel-2026-09-27-001", "rel-2026-09-25-001", "rel-2026-09-26-009"} {
		expectError(t, do(h, http.MethodGet, "/v1/releases/"+id, nil, nil), http.StatusNotFound, "not_found")
	}
}

func TestReleasesServedFromMemoryOnceLoaded(t *testing.T) {
	e := buildDataDir(t)
	s := newServer(e.deps())
	h := s.handler()
	st, err := s.loader.current()
	if err != nil {
		t.Fatal(err)
	}
	list := do(h, http.MethodGet, "/v1/releases", nil, nil)
	detail := do(h, http.MethodGet, "/v1/releases/"+firstRel, nil, nil)
	if list.Code != 200 || detail.Code != 200 {
		t.Fatalf("%d %d", list.Code, detail.Code)
	}
	if !s.cache.has(st, "/v1/releases") || !s.cache.has(st, "/v1/releases/"+firstRel) {
		t.Fatal("release bodies should be cached with the state")
	}
	// The directory is not consulted again while the state stands: remove a
	// data file and keep serving byte-identical bodies past the refresh
	// interval (CURRENT is unchanged, so no reload happens).
	est := filepath.Join(e.paths.Releases, firstRel, "estimates.json")
	if err := os.Rename(est, est+".moved"); err != nil {
		t.Fatal(err)
	}
	e.clock.advance(3 * DefaultRefreshInterval)
	again := do(h, http.MethodGet, "/v1/releases/"+firstRel, nil, nil)
	if again.Code != 200 || !bytes.Equal(again.Body.Bytes(), detail.Body.Bytes()) {
		t.Fatalf("detail should be served from memory: %d", again.Code)
	}
	if rr := do(h, http.MethodGet, "/v1/releases", nil, nil); !bytes.Equal(rr.Body.Bytes(), list.Body.Bytes()) {
		t.Fatal("list should be served from memory")
	}
	if rr := do(h, http.MethodGet, "/v1/meter", nil, nil); rr.Code != 200 {
		t.Fatalf("meter: %d", rr.Code)
	}
}

func TestCacheKeyOnlyCarriesQueryParametersTheRouteReads(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/sources?x=1&q=a%20b&tier=1&limit=2", nil)
	if got := cacheKey(req, sourcesQueryKeys); got != "/v1/sources?tier=1&q=a+b&limit=2" {
		t.Fatalf("key %q", got)
	}
	if got := cacheKey(req, nil); got != "/v1/sources" {
		t.Fatalf("query-blind key %q", got)
	}
	plain := httptest.NewRequest(http.MethodGet, "/v1/sources", nil)
	if got := cacheKey(plain, sourcesQueryKeys); got != "/v1/sources" {
		t.Fatalf("no-query key %q", got)
	}
	reordered := httptest.NewRequest(http.MethodGet, "/v1/sources?limit=2&tier=1&q=a+b&zzz=9", nil)
	if cacheKey(reordered, sourcesQueryKeys) != cacheKey(req, sourcesQueryKeys) {
		t.Fatal("parameter order and unread parameters must not change the key")
	}
}

func TestJunkQueriesDoNotGrowTheCache(t *testing.T) {
	e := buildDataDir(t)
	s := newServer(e.deps())
	h := s.handler()
	st, err := s.loader.current()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 60; i++ {
		if rr := do(h, http.MethodGet, fmt.Sprintf("/v1/drivers?junk=%d&x=%d", i, i*7), nil, nil); rr.Code != 200 {
			t.Fatalf("request %d: %d", i, rr.Code)
		}
	}
	if n, _ := s.cache.stats(); n != 1 || !s.cache.has(st, "/v1/drivers") {
		t.Fatalf("junk queries created %d entries", n)
	}
	for _, q := range []string{"tier=1&x=1", "x=2&tier=1", "tier=1"} {
		if rr := do(h, http.MethodGet, "/v1/sources?"+q, nil, nil); rr.Code != 200 {
			t.Fatalf("%s: %d", q, rr.Code)
		}
	}
	if n, _ := s.cache.stats(); n != 2 || !s.cache.has(st, "/v1/sources?tier=1") {
		t.Fatalf("sources filters should share one entry per read parameter set, have %d", n)
	}
	if rr := do(h, http.MethodGet, "/v1/sources?tier=2", nil, nil); rr.Code != 200 {
		t.Fatalf("tier=2: %d", rr.Code)
	}
	if n, _ := s.cache.stats(); n != 3 {
		t.Fatalf("a different filter is a different entry, have %d", n)
	}
	// Errors are never cached.
	expectError(t, do(h, http.MethodGet, "/v1/sources?tier=9", nil, nil), http.StatusBadRequest, "bad_request")
	if n, _ := s.cache.stats(); n != 3 {
		t.Fatalf("error bodies must not be cached, have %d", n)
	}
}

func TestBodyCacheIsBoundedByBytesAndEntriesWithLRUEviction(t *testing.T) {
	c := newBodyCacheWithLimits(3, 1000) // single bodies above 1000/8 = 125 bytes are not kept
	st := &state{}
	c.put(st, "a", make([]byte, 100))
	c.put(st, "b", make([]byte, 100))
	if n, b := c.stats(); n != 2 || b != 200 {
		t.Fatalf("after two puts: %d entries %d bytes", n, b)
	}
	c.put(st, "c", make([]byte, 100)) // 300 bytes, 3 entries: fits exactly
	if n, b := c.stats(); n != 3 || b != 300 {
		t.Fatalf("at the entry cap: %d entries %d bytes", n, b)
	}
	if _, ok := c.get(st, "a"); !ok { // touch a so b becomes the least recently used
		t.Fatal("a should be cached")
	}
	c.put(st, "d", make([]byte, 50)) // 4th entry: evicts the least recently used (b)
	if n, b := c.stats(); n != 3 || b != 250 || c.has(st, "b") || !c.has(st, "a") || !c.has(st, "c") || !c.has(st, "d") {
		t.Fatalf("entry cap eviction: %d entries %d bytes has b=%v", n, b, c.has(st, "b"))
	}
	c.put(st, "huge", make([]byte, 126)) // larger than maxBytes/8 is not kept
	if c.has(st, "huge") {
		t.Fatal("oversized bodies must not be cached")
	}
	c.put(st, "a", make([]byte, 120)) // replacing an entry adjusts the byte total
	if n, b := c.stats(); n != 3 || b != 270 {
		t.Fatalf("replace: %d entries %d bytes", n, b)
	}
	if body, ok := c.get(st, "a"); !ok || len(body) != 120 {
		t.Fatal("replaced body not served")
	}
	// Byte cap: a wide cache with a small byte budget evicts by bytes, oldest first.
	w := newBodyCacheWithLimits(100, 200) // bodies up to 25 bytes
	for i := 0; i < 12; i++ {
		w.put(st, fmt.Sprintf("k%d", i), make([]byte, 20)) // 240 bytes offered
	}
	if n, b := w.stats(); n != 10 || b != 200 || w.has(st, "k0") || w.has(st, "k1") || !w.has(st, "k2") || !w.has(st, "k11") {
		t.Fatalf("byte cap eviction: %d entries %d bytes", n, b)
	}
	st2 := &state{}
	c.put(st2, "z", make([]byte, 1))
	if n, b := c.stats(); n != 1 || b != 1 || c.has(st, "a") || !c.has(st2, "z") {
		t.Fatalf("a new state must reset the cache: %d entries %d bytes", n, b)
	}
	if _, ok := c.get(st, "z"); ok {
		t.Fatal("an old state must never hit")
	}
}

func TestSubmissionIDsAreUniqueUnderAFrozenClock(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	body := []byte(`{"payload":{"target_id":"src-fixture-survey-2025","correction":"Same text, sent three times."}}`)
	re := regexp.MustCompile(`^sub-[a-f0-9]{16}$`)
	ids := map[string]bool{}
	var at string
	for i := 0; i < 3; i++ {
		rr := do(h, http.MethodPost, "/v1/submissions/corrections", body, nil)
		if rr.Code != http.StatusAccepted {
			t.Fatalf("submission %d: %d %s", i, rr.Code, rr.Body.String())
		}
		var ack SubmissionResponse
		decode(t, rr, &ack)
		if !re.MatchString(ack.ID) {
			t.Fatalf("id format %q", ack.ID)
		}
		if at == "" {
			at = ack.SubmittedAt
		} else if ack.SubmittedAt != at {
			t.Fatalf("clock moved: %s vs %s", ack.SubmittedAt, at)
		}
		ids[ack.ID] = true
	}
	if len(ids) != 3 {
		t.Fatalf("identical payloads in the same second must get distinct ids: %v", ids)
	}
	f, err := os.Open(SubmissionsPath(e.paths.Review))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	stored := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var sub Submission
		if err := json.Unmarshal(sc.Bytes(), &sub); err != nil {
			t.Fatal(err)
		}
		stored[sub.ID] = true
	}
	if len(stored) != 3 {
		t.Fatalf("queue holds %d distinct ids", len(stored))
	}
	payload := []byte(`{"a":1}`)
	if submissionID("correction", at, payload) == submissionID("correction", at, payload) {
		t.Fatal("submissionID must not repeat for identical inputs")
	}
}

func TestSameOriginPOSTBehindTLSTerminatingProxy(t *testing.T) {
	e := buildDataDir(t)
	body := []byte(`{"payload":{"target_id":"src-fixture-survey-2025","correction":"c"}}`)
	send := func(h http.Handler, method, host, origin, proto string, secure bool) *httptest.ResponseRecorder {
		var rd *bytes.Reader
		if method == http.MethodPost {
			rd = bytes.NewReader(body)
		} else {
			rd = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, "/v1/submissions/corrections", rd)
		req.Host = host
		req.Header.Set("Origin", origin)
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		}
		if proto != "" {
			req.Header.Set("X-Forwarded-Proto", proto)
		}
		if secure {
			req.TLS = &tls.ConnectionState{}
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	generous := func(d *Deps) { d.RateLimits.Submit = Limit{Burst: 100, PerMinute: 60} }
	plain := e.handler(t, generous)
	cases := []struct {
		name   string
		h      http.Handler
		host   string
		origin string
		proto  string
		secure bool
		want   int
	}{
		{"exact same origin", plain, "api.example.org", "http://api.example.org", "", false, 202},
		{"https origin, TLS terminated upstream, no flag", plain, "api.example.org", "https://api.example.org", "", false, 202},
		{"https origin with port", plain, "api.example.org:8443", "https://api.example.org:8443", "", false, 202},
		{"case-insensitive host", plain, "API.Example.org", "https://api.example.org", "", false, 202},
		{"http origin against an https request", plain, "api.example.org", "http://api.example.org", "", true, 403},
		{"other host over https", plain, "api.example.org", "https://other.example", "", false, 403},
		{"other port", plain, "api.example.org", "https://api.example.org:9443", "", false, 403},
		{"X-Forwarded-Proto ignored when untrusted", plain, "api.example.org", "http://api.example.org", "https", false, 202},
		{"allowlisted origin still works", plain, "api.example.org", "https://lab.example", "", false, 202},
	}
	trusted := e.handler(t, func(d *Deps) { generous(d); d.TrustProxy = true })
	cases = append(cases,
		struct {
			name   string
			h      http.Handler
			host   string
			origin string
			proto  string
			secure bool
			want   int
		}{"trusted proxy: http origin against forwarded https", trusted, "api.example.org", "http://api.example.org", "https", false, 403},
		struct {
			name   string
			h      http.Handler
			host   string
			origin string
			proto  string
			secure bool
			want   int
		}{"trusted proxy: https origin against forwarded https", trusted, "api.example.org", "https://api.example.org", "https", false, 202},
	)
	for _, c := range cases {
		rr := send(c.h, http.MethodPost, c.host, c.origin, c.proto, c.secure)
		if rr.Code != c.want {
			t.Fatalf("%s: %d %s", c.name, rr.Code, rr.Body.String())
		}
		if c.want == 202 && rr.Header().Get("Access-Control-Allow-Origin") != c.origin {
			t.Fatalf("%s: Access-Control-Allow-Origin %q", c.name, rr.Header().Get("Access-Control-Allow-Origin"))
		}
		pre := send(c.h, http.MethodOptions, c.host, c.origin, c.proto, c.secure)
		if (c.want == 202) != (pre.Code == http.StatusNoContent) {
			t.Fatalf("%s: preflight %d disagrees with POST %d", c.name, pre.Code, rr.Code)
		}
	}
}

func TestIPv6ClientsAreKeyedByTheir64Prefix(t *testing.T) {
	for in, want := range map[string]string{
		"2001:db8:1:2:aaaa::1":  "2001:db8:1:2::/64",
		"2001:db8:1:2::":        "2001:db8:1:2::/64",
		"2001:db8:1:3::1":       "2001:db8:1:3::/64",
		"::ffff:198.51.100.7":   "198.51.100.7",
		"198.51.100.7":          "198.51.100.7",
		"fe80::1":               "fe80::/64",
		"2001:DB8:1:2:0:0:0:99": "2001:db8:1:2::/64",
	} {
		ip := net.ParseIP(in)
		if ip == nil {
			t.Fatalf("test address %q does not parse", in)
		}
		if got := clientKey(ip); got != want {
			t.Fatalf("clientKey(%s) = %q, want %q", in, got, want)
		}
	}
	e := buildDataDir(t)
	h := e.handler(t, func(d *Deps) { d.RateLimits.Read = Limit{Burst: 1, PerMinute: 60} })
	send := func(h http.Handler, addr, xff string) int {
		req := httptest.NewRequest(http.MethodGet, "/v1/outcomes", nil)
		req.RemoteAddr = addr
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	if send(h, "[2001:db8:1:2:aaaa::1]:1234", "") != 200 {
		t.Fatal("first request of a /64 should pass")
	}
	if send(h, "[2001:db8:1:2:bbbb::2]:1234", "") != http.StatusTooManyRequests {
		t.Fatal("a second address in the same /64 must share the bucket")
	}
	if send(h, "[2001:db8:1:3::1]:1234", "") != 200 {
		t.Fatal("another /64 has its own bucket")
	}
	if send(h, "198.51.100.7:1", "") != 200 || send(h, "198.51.100.8:1", "") != 200 || send(h, "198.51.100.7:2", "") != http.StatusTooManyRequests {
		t.Fatal("IPv4 clients are keyed by full address")
	}
	trusted := e.handler(t, func(d *Deps) { d.RateLimits.Read = Limit{Burst: 1, PerMinute: 60}; d.TrustProxy = true })
	if send(trusted, "10.0.0.1:1", "2001:db8:9:9::5") != 200 || send(trusted, "10.0.0.1:1", "2001:db8:9:9::6") != http.StatusTooManyRequests {
		t.Fatal("forwarded IPv6 clients are keyed by /64 too")
	}
	e.clock.advance(time.Second)
	if send(h, "[2001:db8:1:2:cccc::3]:1", "") != 200 {
		t.Fatal("the shared bucket refills")
	}
}
