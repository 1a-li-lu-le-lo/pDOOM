// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"bytes"
	"container/list"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

const (
	contentTypeJSON     = "application/json; charset=utf-8"
	contentTypeMarkdown = "text/markdown; charset=utf-8"
)

// ErrorResponse is the body of every error (api/openapi.yaml: Error).
type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

// Meta is the release metadata embedded in every data response so that any
// probability shown can be traced to a release, a snapshot and a cutoff.
type Meta struct {
	ReleaseID     string   `json:"release_id"`
	DataSnapshot  string   `json:"data_snapshot"`
	DataCutoff    string   `json:"data_cutoff"`
	GeneratedAt   string   `json:"generated_at"`
	Published     *string  `json:"published"`
	ModelVersions []string `json:"model_versions"`
	Limitations   []string `json:"limitations"`
}

// encodeJSON renders v with HTML escaping off and a trailing newline.
func encodeJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ok serves v as JSON for a GET route, caching the body for the current state
// under the key dispatch computed (none when no state is loaded).
func (s *server) ok(w http.ResponseWriter, r *http.Request, st *state, v any) {
	body, err := encodeJSON(v)
	if err != nil {
		s.log.Error("encode response", "error", err.Error(), "path", r.URL.Path)
		s.writeError(w, r, http.StatusInternalServerError, "internal", "response could not be encoded")
		return
	}
	if key := cacheKeyOf(r); st != nil && key != "" {
		s.cache.put(st, key, body)
	}
	s.serveBody(w, r, body, contentTypeJSON)
}

// serveBody writes a cacheable GET body with a strong ETag, honouring
// If-None-Match.
func (s *server) serveBody(w http.ResponseWriter, r *http.Request, body []byte, contentType string) {
	etag := `"` + schema.SHA256Hex(body) + `"`
	h := w.Header()
	h.Set("ETag", etag)
	h.Set("Cache-Control", "public, max-age="+strconv.Itoa(CacheMaxAge))
	h.Set("Content-Type", contentType)
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// writeJSON writes a non-cacheable JSON response (POST results, errors).
func (s *server) writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := encodeJSON(v)
	if err != nil {
		body = []byte("{\"error\":\"internal\",\"detail\":\"response could not be encoded\"}\n")
		status = http.StatusInternalServerError
	}
	h := w.Header()
	h.Set("Content-Type", contentTypeJSON)
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// writeError writes the {error, detail} body.
func (s *server) writeError(w http.ResponseWriter, _ *http.Request, status int, code, detail string) {
	s.writeJSON(w, status, ErrorResponse{Error: code, Detail: detail})
}

// etagMatches implements If-None-Match for strong and weak validators.
func etagMatches(header, etag string) bool {
	if header == "" {
		return false
	}
	for _, part := range strings.Split(header, ",") {
		p := strings.TrimSpace(part)
		if p == "*" {
			return true
		}
		p = strings.TrimPrefix(p, "W/")
		if p == etag {
			return true
		}
	}
	return false
}

// bodyCache keeps encoded GET bodies for one state. It is dropped whenever
// the state changes and bounded both in entries and in total bytes; when full
// the least recently used body is evicted, so a client sending many distinct
// keys can only churn the cache, never grow it.
type bodyCache struct {
	maxEntries int
	maxBytes   int

	mu    sync.Mutex
	st    *state
	m     map[string]*list.Element
	order *list.List // front = most recently used
	bytes int
}

type cacheEntry struct {
	key  string
	body []byte
}

// Cache bounds. A single body larger than an eighth of the byte budget is
// re-encoded per request rather than allowed to dominate the cache.
const (
	bodyCacheMaxEntries = 2048
	bodyCacheMaxBytes   = 64 << 20
)

func newBodyCache() *bodyCache { return newBodyCacheWithLimits(bodyCacheMaxEntries, bodyCacheMaxBytes) }

func newBodyCacheWithLimits(maxEntries, maxBytes int) *bodyCache {
	if maxEntries < 1 {
		maxEntries = 1
	}
	if maxBytes < 1 {
		maxBytes = 1
	}
	return &bodyCache{maxEntries: maxEntries, maxBytes: maxBytes, m: map[string]*list.Element{}, order: list.New()}
}

func (c *bodyCache) get(st *state, key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.st != st {
		return nil, false
	}
	el, ok := c.m[key]
	if !ok {
		return nil, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*cacheEntry).body, true
}

func (c *bodyCache) put(st *state, key string, body []byte) {
	if len(body) > c.maxBytes/8 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.st != st {
		c.st = st
		c.m = map[string]*list.Element{}
		c.order.Init()
		c.bytes = 0
	}
	if el, ok := c.m[key]; ok {
		c.bytes += len(body) - len(el.Value.(*cacheEntry).body)
		el.Value.(*cacheEntry).body = body
		c.order.MoveToFront(el)
		return
	}
	for c.order.Len() > 0 && (c.order.Len() >= c.maxEntries || c.bytes+len(body) > c.maxBytes) {
		c.evictOldest()
	}
	c.m[key] = c.order.PushFront(&cacheEntry{key: key, body: body})
	c.bytes += len(body)
}

func (c *bodyCache) evictOldest() {
	el := c.order.Back()
	if el == nil {
		return
	}
	e := c.order.Remove(el).(*cacheEntry)
	delete(c.m, e.key)
	c.bytes -= len(e.body)
}

// stats reports the entry count and byte total (for tests).
func (c *bodyCache) stats() (entries, total int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len(), c.bytes
}

// has reports whether key is cached for st (for tests).
func (c *bodyCache) has(st *state, key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.st != st {
		return false
	}
	_, ok := c.m[key]
	return ok
}

// headWriter discards the body of HEAD responses while keeping headers.
type headWriter struct{ http.ResponseWriter }

func (h headWriter) Write(p []byte) (int, error) { return len(p), nil }
