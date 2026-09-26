// Copyright NU Cybernetics. p(DOOM) — research prototype.

package robots

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeDoer serves robots.txt bodies per host without a network.
type fakeDoer struct {
	mu     sync.Mutex
	status map[string]int
	body   map[string]string
	err    map[string]error
	calls  map[string]int
	agents []string
}

func newFake() *fakeDoer {
	return &fakeDoer{status: map[string]int{}, body: map[string]string{}, err: map[string]error{}, calls: map[string]int{}}
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	host := req.URL.Host
	f.calls[host]++
	f.agents = append(f.agents, req.Header.Get("User-Agent"))
	if req.URL.Path != "/robots.txt" {
		return nil, errors.New("fake: only robots.txt is served")
	}
	if err, ok := f.err[host]; ok {
		return nil, err
	}
	code, ok := f.status[host]
	if !ok {
		code = 200
	}
	return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(f.body[host])), Header: http.Header{}}, nil
}

const sampleRobots = `User-agent: *
Disallow: /private/
Crawl-delay: 2

User-agent: pdoom-ingest
Disallow: /no-pdoom/
Allow: /
Crawl-delay: 7
`

func TestAllowedDisallowedAndCrawlDelay(t *testing.T) {
	f := newFake()
	f.body["site.example"] = sampleRobots
	c := NewCache(f)
	ctx := context.Background()

	ok, err := c.Allowed(ctx, "https://site.example/feed.xml")
	if err != nil || !ok {
		t.Fatalf("feed should be allowed: ok=%v err=%v", ok, err)
	}
	ok, err = c.Allowed(ctx, "https://site.example/no-pdoom/x")
	if err != nil || ok {
		t.Fatalf("agent-specific disallow must apply: ok=%v err=%v", ok, err)
	}
	// The agent-specific group wins over *, so /private/ is allowed for us.
	if ok, _ := c.Allowed(ctx, "https://site.example/private/x"); !ok {
		t.Fatal("agent group should override the wildcard group")
	}
	if d := c.CrawlDelay(ctx, "https://site.example/anything"); d != 7*time.Second {
		t.Fatalf("crawl delay = %v, want 7s", d)
	}
	ch := c.Check(ctx, "https://site.example/no-pdoom/y?q=1")
	if ch.Allowed || ch.Status != "disallowed" || !ch.Fetched || ch.StatusCode != 200 {
		t.Fatalf("check = %+v", ch)
	}
	if f.calls["site.example"] != 1 {
		t.Fatalf("robots.txt must be fetched once per host, got %d", f.calls["site.example"])
	}
	if len(f.agents) == 0 || f.agents[0] != UserAgent {
		t.Fatalf("user agent header = %q", f.agents)
	}
}

func TestWildcardGroupWhenNoAgentGroup(t *testing.T) {
	f := newFake()
	f.body["w.example"] = "User-agent: *\nDisallow: /private/\nCrawl-delay: 3\n"
	c := NewCache(f)
	ctx := context.Background()
	if ok, _ := c.Allowed(ctx, "https://w.example/private/doc"); ok {
		t.Fatal("wildcard disallow must apply")
	}
	if ok, _ := c.Allowed(ctx, "https://w.example/public"); !ok {
		t.Fatal("public path must be allowed")
	}
	if d := c.CrawlDelay(ctx, "https://w.example/"); d != 3*time.Second {
		t.Fatalf("delay = %v", d)
	}
}

func TestFailClosed(t *testing.T) {
	f := newFake()
	f.err["down.example"] = errors.New("connection refused")
	f.status["broken.example"] = 500
	f.status["forbidden.example"] = 403
	f.status["redirect.example"] = 302
	c := NewCache(f)
	ctx := context.Background()
	for _, host := range []string{"down.example", "broken.example", "forbidden.example", "redirect.example"} {
		ok, err := c.Allowed(ctx, "https://"+host+"/feed")
		if ok || err == nil || !errors.Is(err, ErrFetch) {
			t.Errorf("%s: want closed with ErrFetch, got ok=%v err=%v", host, ok, err)
		}
		ch := c.Check(ctx, "https://"+host+"/feed")
		if ch.Allowed || ch.Status != "unknown" || ch.Fetched {
			t.Errorf("%s: check = %+v", host, ch)
		}
		if c.CrawlDelay(ctx, "https://"+host+"/") != 0 {
			t.Errorf("%s: crawl delay must be zero when unknown", host)
		}
	}
	// Failures are cached too (no hammering a broken host).
	if f.calls["down.example"] != 1 {
		t.Fatalf("failed fetch should be cached, got %d calls", f.calls["down.example"])
	}
}

func TestMissingRobotsMeansAllowed(t *testing.T) {
	f := newFake()
	f.status["none.example"] = 404
	f.status["gone.example"] = 410
	c := NewCache(f)
	for _, host := range []string{"none.example", "gone.example"} {
		ch := c.Check(context.Background(), "https://"+host+"/anything/at/all")
		if !ch.Allowed || ch.Status != "allowed" || !ch.Fetched || ch.Err != nil {
			t.Errorf("%s: %+v", host, ch)
		}
	}
}

func TestEmptyAndOversizedBodies(t *testing.T) {
	f := newFake()
	f.body["empty.example"] = ""
	f.body["huge.example"] = "User-agent: *\nDisallow: /x\n" + strings.Repeat("# padding\n", MaxRobotsBytes/10+10)
	c := NewCache(f)
	if ok, err := c.Allowed(context.Background(), "https://empty.example/a"); !ok || err != nil {
		t.Fatalf("empty robots.txt means allow-all: ok=%v err=%v", ok, err)
	}
	if ok, err := c.Allowed(context.Background(), "https://huge.example/x"); ok || err != nil {
		t.Fatalf("oversized body must still be parsed up to the cap: ok=%v err=%v", ok, err)
	}
}

func TestInvalidURLAndNilDoer(t *testing.T) {
	c := NewCache(newFake())
	for _, u := range []string{"", "not a url", "/relative", "http://[::1"} {
		ok, err := c.Allowed(context.Background(), u)
		if ok || err == nil {
			t.Errorf("%q must be refused", u)
		}
	}
	nc := NewCache(nil)
	if ok, err := nc.Allowed(context.Background(), "https://x.example/"); ok || !errors.Is(err, ErrFetch) {
		t.Fatalf("nil doer must fail closed: ok=%v err=%v", ok, err)
	}
}

func TestResetAndPerHostCaching(t *testing.T) {
	f := newFake()
	f.body["a.example"] = "User-agent: *\nDisallow: /\n"
	f.body["b.example"] = "User-agent: *\nAllow: /\n"
	c := NewCache(f)
	ctx := context.Background()
	if ok, _ := c.Allowed(ctx, "https://a.example/"); ok {
		t.Fatal("a must be disallowed")
	}
	if ok, _ := c.Allowed(ctx, "https://b.example/"); !ok {
		t.Fatal("b must be allowed")
	}
	if ok, _ := c.Allowed(ctx, "https://A.EXAMPLE/other"); ok {
		t.Fatal("host matching is case-insensitive")
	}
	if f.calls["a.example"]+f.calls["A.EXAMPLE"] != 1 || f.calls["b.example"] != 1 {
		t.Fatalf("calls = %v", f.calls)
	}
	c.Reset()
	_, _ = c.Allowed(ctx, "https://a.example/")
	if f.calls["a.example"] != 2 {
		t.Fatalf("reset should force a refetch, calls=%v", f.calls)
	}
}

func TestConcurrentLookupsFetchOnce(t *testing.T) {
	f := newFake()
	f.body["c.example"] = "User-agent: *\nAllow: /\n"
	c := NewCache(f)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.Allowed(context.Background(), "https://c.example/p")
		}()
	}
	wg.Wait()
	if f.calls["c.example"] < 1 || f.calls["c.example"] > 20 {
		t.Fatalf("calls=%d", f.calls["c.example"])
	}
	// After the race settles, further lookups never refetch.
	before := f.calls["c.example"]
	_, _ = c.Allowed(context.Background(), "https://c.example/q")
	if f.calls["c.example"] != before {
		t.Fatal("cached entry must be reused")
	}
}
