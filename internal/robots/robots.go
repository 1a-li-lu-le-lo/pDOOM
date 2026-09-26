// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package robots fetches and evaluates robots.txt for the ingestion pipeline.
// Decisions are cached per scheme://host for the lifetime of the cache, and
// every failure to obtain a usable robots.txt (network error, 5xx, 401/403,
// parse error) is treated as "disallowed": the crawler fails closed.
package robots

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/temoto/robotstxt"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// UserAgent is the full User-Agent header sent by pdoom-ingest.
const UserAgent = "pdoom-ingest/0.1 (+https://github.com/1a-li-lu-le-lo/pdoom)"

// AgentToken is the product token matched against robots.txt User-agent lines.
const AgentToken = "pdoom-ingest"

// MaxRobotsBytes bounds how much of a robots.txt is read.
const MaxRobotsBytes = 512 * 1024

// Doer performs one HTTP request. The ingestion pipeline passes the
// SafeClient's guarded doer so robots.txt fetches get the same host allowlist
// and SSRF checks as everything else.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// ErrFetch is wrapped into every error returned when robots.txt could not be
// obtained; callers treat it as a denial.
var ErrFetch = errors.New("robots.txt unavailable (fail closed)")

// Check is the decision for one URL.
type Check struct {
	// Allowed is the final decision for the URL (false whenever the robots file
	// could not be obtained).
	Allowed bool
	// Status is "allowed" or "disallowed" for a usable robots file, "unknown"
	// when the file could not be obtained (Err set).
	Status schema.RobotsStatus
	// Fetched reports whether a usable robots.txt (or a definitive 404) was obtained.
	Fetched bool
	// StatusCode is the HTTP status of the robots.txt response (0 on network error).
	StatusCode int
	// CrawlDelay is the crawl-delay of the matching group (0 when none).
	CrawlDelay time.Duration
	// Err explains an unobtainable robots file.
	Err error
}

type entry struct {
	data       *robotstxt.RobotsData
	statusCode int
	err        error
}

// Cache fetches robots.txt at most once per scheme://host.
type Cache struct {
	mu      sync.Mutex
	doer    Doer
	agent   string
	entries map[string]*entry
}

// NewCache returns a cache using d for robots.txt fetches.
func NewCache(d Doer) *Cache {
	return &Cache{doer: d, agent: AgentToken, entries: map[string]*entry{}}
}

// Reset drops every cached decision.
func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = map[string]*entry{}
}

// Check evaluates rawURL against the host's robots.txt.
func (c *Cache) Check(ctx context.Context, rawURL string) Check {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" || u.Scheme == "" {
		return Check{Status: "unknown", Err: fmt.Errorf("robots: invalid url %q: %w", rawURL, ErrFetch)}
	}
	e := c.lookup(ctx, u)
	if e.err != nil {
		return Check{Status: "unknown", StatusCode: e.statusCode, Err: e.err}
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	group := e.data.FindGroup(c.agent)
	var delay time.Duration
	if group != nil {
		delay = group.CrawlDelay
	}
	ok := e.data.TestAgent(path, c.agent)
	status := schema.RobotsStatus("allowed")
	if !ok {
		status = "disallowed"
	}
	return Check{Allowed: ok, Status: status, Fetched: true, StatusCode: e.statusCode, CrawlDelay: delay}
}

// Allowed reports whether rawURL may be fetched. A non-nil error means the
// robots file could not be obtained; the boolean is then false.
func (c *Cache) Allowed(ctx context.Context, rawURL string) (bool, error) {
	ch := c.Check(ctx, rawURL)
	return ch.Allowed, ch.Err
}

// CrawlDelay returns the crawl-delay for the host of rawURL (0 when unknown).
func (c *Cache) CrawlDelay(ctx context.Context, rawURL string) time.Duration {
	return c.Check(ctx, rawURL).CrawlDelay
}

func (c *Cache) lookup(ctx context.Context, u *url.URL) *entry {
	key := strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
	c.mu.Lock()
	if e, ok := c.entries[key]; ok {
		c.mu.Unlock()
		return e
	}
	c.mu.Unlock()
	e := c.fetch(ctx, u.Scheme, u.Host)
	c.mu.Lock()
	// Keep the first result if a concurrent lookup raced us.
	if prev, ok := c.entries[key]; ok {
		e = prev
	} else {
		c.entries[key] = e
	}
	c.mu.Unlock()
	return e
}

func (c *Cache) fetch(ctx context.Context, scheme, host string) *entry {
	if c.doer == nil {
		return &entry{err: fmt.Errorf("robots: no http client configured: %w", ErrFetch)}
	}
	target := scheme + "://" + host + "/robots.txt"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return &entry{err: fmt.Errorf("robots: build request %s: %w", target, ErrFetch)}
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/plain")
	resp, err := c.doer.Do(req)
	if err != nil {
		return &entry{err: fmt.Errorf("robots: fetch %s: %v: %w", target, err, ErrFetch)}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxRobotsBytes))
	if err != nil {
		return &entry{statusCode: resp.StatusCode, err: fmt.Errorf("robots: read %s: %v: %w", target, err, ErrFetch)}
	}
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		data, perr := robotstxt.FromBytes(body)
		if perr != nil {
			return &entry{statusCode: resp.StatusCode, err: fmt.Errorf("robots: parse %s: %v: %w", target, perr, ErrFetch)}
		}
		return &entry{data: data, statusCode: resp.StatusCode}
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		// No robots.txt: the standard meaning is "no restrictions".
		data, _ := robotstxt.FromStatusAndBytes(resp.StatusCode, nil)
		return &entry{data: data, statusCode: resp.StatusCode}
	default:
		// 401/403, other 4xx, 5xx and redirects that were not followed: closed.
		return &entry{statusCode: resp.StatusCode, err: fmt.Errorf("robots: %s returned HTTP %d: %w", target, resp.StatusCode, ErrFetch)}
	}
}
