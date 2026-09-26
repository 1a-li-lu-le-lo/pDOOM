// Copyright NU Cybernetics. p(DOOM) — research prototype.

package fetch

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/robots"
)

// mapResolver is a deterministic fake DNS.
type mapResolver map[string][]string

func (m mapResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	ips, ok := m[host]
	if !ok {
		return nil, fmt.Errorf("no such host %q", host)
	}
	var out []net.IPAddr
	for _, s := range ips {
		out = append(out, net.IPAddr{IP: net.ParseIP(s)})
	}
	return out, nil
}

type fakePolicy struct {
	allowed    bool
	denyPrefix string // paths with this prefix are denied even when allowed is true
	err        error
	delay      time.Duration
	calls      int32
}

func (p *fakePolicy) Allowed(_ context.Context, rawURL string) (bool, error) {
	atomic.AddInt32(&p.calls, 1)
	if p.err != nil {
		return false, p.err
	}
	if p.denyPrefix != "" {
		if u, err := url.Parse(rawURL); err == nil && strings.HasPrefix(u.Path, p.denyPrefix) {
			return false, nil
		}
	}
	return p.allowed, nil
}
func (p *fakePolicy) CrawlDelay(context.Context, string) time.Duration { return p.delay }

// harness wires an httptest TLS server behind the guard: the allowlisted hosts
// resolve to public-looking addresses and the dialer is redirected to the
// local listener; everything else goes through the real checks. The client
// verifies every host against the test certificate's name (ServerName), so
// several hostnames can share one listener.
type harness struct {
	srv      *httptest.Server
	dials    int32
	requests []*http.Request
	mu       sync.Mutex
}

func newHarness(t *testing.T) *harness {
	h := &harness{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		h.requests = append(h.requests, r.Clone(context.Background()))
		h.mu.Unlock()
		switch r.URL.Path {
		case "/feed":
			w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
			w.Header().Set("Set-Cookie", "session=abc")
			fmt.Fprint(w, `<rss><channel><title>t</title></channel></rss>`)
		case "/redirect-http-loopback":
			http.Redirect(w, r, "http://127.0.0.1:8080/x", http.StatusFound)
		case "/redirect-https-loopback":
			http.Redirect(w, r, "https://127.0.0.1/x", http.StatusFound)
		case "/redirect-other-host":
			http.Redirect(w, r, "https://other.example/x", http.StatusFound)
		case "/redirect-private-host":
			http.Redirect(w, r, "https://private.example.com/x", http.StatusFound)
		case "/redirect-private-path":
			http.Redirect(w, r, "https://example.com/private/x", http.StatusFound)
		case "/private/x":
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, "members only")
		case "/redirect-ok":
			http.Redirect(w, r, "https://example.com/feed", http.StatusFound)
		case "/r1":
			http.Redirect(w, r, "https://example.com/r2", http.StatusFound)
		case "/r2":
			http.Redirect(w, r, "https://example.com/r3", http.StatusFound)
		case "/r3":
			http.Redirect(w, r, "https://example.com/feed", http.StatusFound)
		case "/r0":
			http.Redirect(w, r, "https://example.com/r1", http.StatusFound)
		case "/big":
			w.Header().Set("Content-Type", "text/plain")
			w.Write(bytes.Repeat([]byte("a"), 4096))
		case "/csv":
			w.Header().Set("Content-Type", "text/csv")
			fmt.Fprint(w, "a,b")
		case "/no-type":
			w.Header()["Content-Type"] = nil
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("x"))
		case "/error":
			http.Error(w, "boom", http.StatusInternalServerError)
		case "/gzip":
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			gz.Write([]byte(strings.Repeat("hello ", 100)))
			gz.Close()
		case "/gzip-bomb":
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			gz.Write(bytes.Repeat([]byte("z"), 200000))
			gz.Close()
		case "/robots.txt":
			if r.Host == "redirect.example.com" {
				http.Redirect(w, r, "https://example.com/robots.txt", http.StatusFound)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, "User-agent: *\nAllow: /\n")
		default:
			http.NotFound(w, r)
		}
	})
	h.srv = httptest.NewTLSServer(mux)
	t.Cleanup(h.srv.Close)
	return h
}

func (h *harness) options(policy Policy, now func() time.Time, sleep func(context.Context, time.Duration) error) Options {
	pool := x509.NewCertPool()
	pool.AddCert(h.srv.Certificate())
	return Options{
		AllowedHosts: []string{"Example.COM", "private.example.com", "redirect.example.com"},
		MaxBody:      2048,
		MinDelay:     0,
		Policy:       policy,
		Resolver: mapResolver{
			"example.com":          {"93.184.216.34"},
			"private.example.com":  {"93.184.216.35", "10.0.0.5"},
			"redirect.example.com": {"93.184.216.36"},
		},
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			atomic.AddInt32(&h.dials, 1)
			return (&net.Dialer{}).DialContext(ctx, network, h.srv.Listener.Addr().String())
		},
		TLSConfig: &tls.Config{RootCAs: pool, ServerName: "example.com", MinVersion: tls.VersionTLS12},
		Now:       now,
		Sleep:     sleep,
	}
}

// fakeClock returns Now/Sleep functions over a virtual clock that only
// advances when a sleep is recorded.
func fakeClock() (now func() time.Time, sleep func(context.Context, time.Duration) error, slept *[]time.Duration) {
	var clock time.Time
	var log []time.Duration
	now = func() time.Time { return clock }
	sleep = func(_ context.Context, d time.Duration) error {
		log = append(log, d)
		clock = clock.Add(d)
		return nil
	}
	return now, sleep, &log
}

func TestGetTable(t *testing.T) {
	h := newHarness(t)
	client, err := New(h.options(&fakePolicy{allowed: true}, nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	cases := []struct {
		name    string
		url     string
		wantErr error
	}{
		{"ok feed", "https://example.com/feed", nil},
		{"redirect within allowlist ok", "https://example.com/redirect-ok", nil},
		{"three redirects ok", "https://example.com/r1", nil},
		{"four redirects refused", "https://example.com/r0", ErrTooManyRedirects},
		{"redirect to http loopback refused", "https://example.com/redirect-http-loopback", ErrNotHTTPS},
		{"redirect to https loopback ip refused", "https://example.com/redirect-https-loopback", ErrIPLiteral},
		{"redirect to non-allowlisted host refused", "https://example.com/redirect-other-host", ErrHostNotAllowed},
		{"redirect to host resolving privately refused", "https://example.com/redirect-private-host", ErrPrivateAddress},
		{"oversized body refused", "https://example.com/big", ErrBodyTooLarge},
		{"disallowed content type refused", "https://example.com/csv", ErrContentType},
		{"missing content type refused", "https://example.com/no-type", ErrContentType},
		{"server error", "https://example.com/error", ErrHTTPStatus},
		{"not found", "https://example.com/missing", ErrHTTPStatus},
		{"unknown host refused", "https://unknown.example/feed", ErrHostNotAllowed},
		{"http refused", "http://example.com/feed", ErrNotHTTPS},
		{"ip literal refused", "https://127.0.0.1/feed", ErrIPLiteral},
		{"ipv6 literal refused", "https://[::1]/feed", ErrIPLiteral},
		{"credentials refused", "https://user:pw@example.com/feed", ErrCredentials},
		{"non-443 port refused", "https://example.com:8443/feed", ErrHostNotAllowed},
		{"gzip decoded transparently", "https://example.com/gzip", nil},
		{"gzip bomb capped", "https://example.com/gzip-bomb", ErrBodyTooLarge},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := client.Get(ctx, c.url)
			if c.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if res.SHA256 == "" || res.Bytes != len(res.Body) || res.StatusCode != 200 {
					t.Fatalf("bad result %+v", res)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected %v, got result %+v", c.wantErr, res)
			}
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected %v, got %v", c.wantErr, err)
			}
		})
	}
}

func TestResultDetails(t *testing.T) {
	h := newHarness(t)
	client, _ := New(h.options(&fakePolicy{allowed: true}, nil, nil))
	res, err := client.Get(context.Background(), "https://example.com/redirect-ok")
	if err != nil {
		t.Fatal(err)
	}
	if res.FinalURL != "https://example.com/feed" || res.Redirects != 1 || res.ContentType != "application/rss+xml" {
		t.Fatalf("result = %+v", res)
	}
	if len(res.SHA256) != 64 {
		t.Fatalf("sha256 = %q", res.SHA256)
	}
	gz, err := client.Get(context.Background(), "https://example.com/gzip")
	if err != nil {
		t.Fatal(err)
	}
	if string(gz.Body) != strings.Repeat("hello ", 100) {
		t.Fatalf("gzip body not decoded: %d bytes", len(gz.Body))
	}
}

func TestHeadersNoCookiesNoCredentials(t *testing.T) {
	h := newHarness(t)
	client, _ := New(h.options(&fakePolicy{allowed: true}, nil, nil))
	for i := 0; i < 2; i++ {
		if _, err := client.Get(context.Background(), "https://example.com/feed"); err != nil {
			t.Fatal(err)
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.requests) != 2 {
		t.Fatalf("requests = %d", len(h.requests))
	}
	for _, r := range h.requests {
		if r.Header.Get("User-Agent") != DefaultUserAgent {
			t.Errorf("user agent = %q", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Cookie") != "" || r.Header.Get("Authorization") != "" {
			t.Errorf("request must not carry cookies or credentials: %v", r.Header)
		}
	}
}

func TestRobotsPolicyFailsClosed(t *testing.T) {
	h := newHarness(t)
	denied := &fakePolicy{allowed: false}
	client, _ := New(h.options(denied, nil, nil))
	if _, err := client.Get(context.Background(), "https://example.com/feed"); !errors.Is(err, ErrRobotsDisallowed) {
		t.Fatalf("want ErrRobotsDisallowed, got %v", err)
	}
	broken := &fakePolicy{allowed: true, err: errors.New("robots.txt unreachable")}
	client, _ = New(h.options(broken, nil, nil))
	if _, err := client.Get(context.Background(), "https://example.com/feed"); !errors.Is(err, ErrRobotsDisallowed) {
		t.Fatalf("policy error must fail closed, got %v", err)
	}
	if atomic.LoadInt32(&h.dials) != 0 {
		t.Fatal("no connection may be opened when robots denies")
	}
	// A refused host never reaches the policy or the network.
	p := &fakePolicy{allowed: true}
	client, _ = New(h.options(p, nil, nil))
	_, _ = client.Get(context.Background(), "https://unknown.example/feed")
	if atomic.LoadInt32(&p.calls) != 0 || atomic.LoadInt32(&h.dials) != 0 {
		t.Fatal("unknown host must be refused before policy and network")
	}
}

func TestRateLimitHonoursCrawlDelay(t *testing.T) {
	h := newHarness(t)
	now, sleep, slept := fakeClock()
	opts := h.options(&fakePolicy{allowed: true, delay: 10 * time.Second}, now, sleep)
	opts.MinDelay = 3 * time.Second
	client, _ := New(opts)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := client.Get(ctx, "https://example.com/feed"); err != nil {
			t.Fatal(err)
		}
	}
	if len(*slept) != 2 || (*slept)[0] != 10*time.Second || (*slept)[1] != 10*time.Second {
		t.Fatalf("slept = %v, want two 10s waits (crawl-delay > min delay)", *slept)
	}
	// With no crawl-delay the minimum applies.
	now, sleep, slept = fakeClock()
	opts = h.options(&fakePolicy{allowed: true}, now, sleep)
	opts.MinDelay = 3 * time.Second
	client, _ = New(opts)
	_, _ = client.Get(ctx, "https://example.com/feed")
	_, _ = client.Get(ctx, "https://example.com/feed")
	if len(*slept) != 1 || (*slept)[0] != 3*time.Second {
		t.Fatalf("slept = %v, want one 3s wait", *slept)
	}
	// Elapsed time reduces the wait.
	*slept = nil
	client.opts.Now = func() time.Time { return now().Add(2 * time.Second) }
	_, _ = client.Get(ctx, "https://example.com/feed")
	if len(*slept) != 1 || (*slept)[0] != 1*time.Second {
		t.Fatalf("slept = %v, want one 1s wait", *slept)
	}
}

// TestCrawlDelayIsCappedAndWaitsObserveContext: a robots.txt asking for a day
// between requests gets MaxCrawlDelay, and a context that ends during the
// wait ends the request with ctx.Err() instead of sleeping on.
func TestCrawlDelayIsCappedAndWaitsObserveContext(t *testing.T) {
	h := newHarness(t)
	now, sleep, slept := fakeClock()
	opts := h.options(&fakePolicy{allowed: true, delay: 24 * time.Hour}, now, sleep)
	opts.MinDelay = 3 * time.Second
	client, _ := New(opts)
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if _, err := client.Get(ctx, "https://example.com/feed"); err != nil {
			t.Fatal(err)
		}
	}
	if len(*slept) != 1 || (*slept)[0] != MaxCrawlDelay {
		t.Fatalf("slept = %v, want one wait of exactly MaxCrawlDelay (%v)", *slept, MaxCrawlDelay)
	}

	// Real sleeper, frozen clock: the second request would wait 60 s, but
	// the context ends after 50 ms.
	opts = h.options(&fakePolicy{allowed: true, delay: 24 * time.Hour}, now, nil)
	client, _ = New(opts)
	if _, err := client.Get(ctx, "https://example.com/feed"); err != nil {
		t.Fatal(err)
	}
	dials := atomic.LoadInt32(&h.dials)
	short, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := client.Get(short, "https://example.com/feed")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want context.DeadlineExceeded, got %v", err)
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("wait ignored the context: took %v", time.Since(start))
	}
	if atomic.LoadInt32(&h.dials) != dials {
		t.Fatal("no connection may be opened after the context ended")
	}
	// An already-cancelled context never sleeps, never dials.
	done, cancelDone := context.WithCancel(ctx)
	cancelDone()
	if _, err := client.Get(done, "https://example.com/feed"); !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
	if atomic.LoadInt32(&h.dials) != dials {
		t.Fatal("cancelled context must not dial")
	}
}

// TestRedirectsConsultRobotsAndWait: a redirect into a path the robots policy
// denies is refused (fail closed, no second connection), and a permitted hop
// waits the per-host delay like any other request.
func TestRedirectsConsultRobotsAndWait(t *testing.T) {
	h := newHarness(t)
	policy := &fakePolicy{allowed: true, denyPrefix: "/private/"}
	client, _ := New(h.options(policy, nil, nil))
	ctx := context.Background()
	if _, err := client.Get(ctx, "https://example.com/private/x"); !errors.Is(err, ErrRobotsDisallowed) {
		t.Fatalf("direct denied path: %v", err)
	}
	_, err := client.Get(ctx, "https://example.com/redirect-private-path")
	if !errors.Is(err, ErrRobotsDisallowed) {
		t.Fatalf("redirect into a robots-disallowed path must be refused, got %v", err)
	}
	if atomic.LoadInt32(&policy.calls) != 3 {
		t.Fatalf("policy consulted %d times, want 3 (direct, initial url, redirect target)", policy.calls)
	}
	h.mu.Lock()
	for _, r := range h.requests {
		if r.URL.Path == "/private/x" {
			t.Fatal("the disallowed redirect target was fetched")
		}
	}
	h.mu.Unlock()

	now, sleep, slept := fakeClock()
	opts := h.options(&fakePolicy{allowed: true}, now, sleep)
	opts.MinDelay = 3 * time.Second
	client, _ = New(opts)
	res, err := client.Get(ctx, "https://example.com/redirect-ok")
	if err != nil || res.Redirects != 1 {
		t.Fatalf("redirect-ok: %v %+v", err, res)
	}
	if len(*slept) != 1 || (*slept)[0] != 3*time.Second {
		t.Fatalf("slept = %v, want one 3s wait before following the hop", *slept)
	}
}

func TestRobotsDoer(t *testing.T) {
	h := newHarness(t)
	client, _ := New(h.options(&fakePolicy{allowed: false}, nil, nil))
	d := client.RobotsDoer()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/robots.txt", nil)
	resp, err := d.Do(req)
	if err != nil {
		t.Fatalf("robots doer must bypass the policy: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	bad, _ := http.NewRequest(http.MethodGet, "https://unknown.example/robots.txt", nil)
	if _, err := d.Do(bad); !errors.Is(err, ErrHostNotAllowed) {
		t.Fatalf("robots doer must keep the allowlist: %v", err)
	}
	loop, _ := http.NewRequest(http.MethodGet, "https://private.example.com/robots.txt", nil)
	if _, err := d.Do(loop); !errors.Is(err, ErrPrivateAddress) {
		t.Fatalf("robots doer must keep the SSRF guard: %v", err)
	}
}

// TestRobotsDoerDoesNotFollowRedirects: a redirected robots.txt is returned as
// the 3xx response (one connection, no hop), and the robots cache built on the
// doer therefore fails closed for that host.
func TestRobotsDoerDoesNotFollowRedirects(t *testing.T) {
	h := newHarness(t)
	client, _ := New(h.options(&fakePolicy{allowed: true}, nil, nil))
	req, _ := http.NewRequest(http.MethodGet, "https://redirect.example.com/robots.txt", nil)
	resp, err := client.RobotsDoer().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status %d, want the 302 handed back unfollowed", resp.StatusCode)
	}
	if atomic.LoadInt32(&h.dials) != 1 {
		t.Fatalf("dials = %d, want 1", h.dials)
	}
	cache := robots.NewCache(client.RobotsDoer())
	ok, err := cache.Allowed(context.Background(), "https://redirect.example.com/feed")
	if ok || !errors.Is(err, robots.ErrFetch) {
		t.Fatalf("redirected robots.txt must be closed: ok=%v err=%v", ok, err)
	}
	// The un-redirected host still works end to end through the cache.
	if ok, err := cache.Allowed(context.Background(), "https://example.com/feed"); !ok || err != nil {
		t.Fatalf("example.com: ok=%v err=%v", ok, err)
	}
}

func TestNewValidation(t *testing.T) {
	if _, err := New(Options{AllowedHosts: []string{"10.0.0.1"}}); err == nil {
		t.Fatal("ip literal in allowlist must be refused")
	}
	if _, err := New(Options{AllowedHosts: []string{" "}}); err == nil {
		t.Fatal("empty host in allowlist must be refused")
	}
	if _, err := New(Options{MaxRedirects: 10}); err == nil {
		t.Fatal("redirect limit above the hard cap must be refused")
	}
	if _, err := New(Options{MinDelay: MaxCrawlDelay + time.Second}); err == nil {
		t.Fatal("minimum delay above MaxCrawlDelay must be refused")
	}
	c, err := New(Options{AllowedHosts: []string{"B.example", "a.example."}})
	if err != nil {
		t.Fatal(err)
	}
	if hosts := c.AllowedHosts(); len(hosts) != 2 || hosts[0] != "a.example" || hosts[1] != "b.example" {
		t.Fatalf("hosts = %v", hosts)
	}
	if _, err := c.Get(context.Background(), "https://x.example/"); !errors.Is(err, ErrHostNotAllowed) {
		t.Fatalf("empty-ish allowlist: %v", err)
	}
	if _, err := c.Get(context.Background(), "::not a url"); err == nil {
		t.Fatal("garbage url must fail")
	}
}

func TestIsPublicIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "127.8.8.8", "10.1.2.3", "172.16.0.1", "172.31.255.255", "192.168.1.1", "169.254.169.254",
		"0.0.0.0", "0.1.2.3", "100.64.0.1", "100.127.255.255", "192.0.0.1", "224.0.0.1", "239.255.255.255", "240.0.0.1", "255.255.255.255",
		"192.0.2.1", "198.18.0.1", "198.19.255.255", "198.51.100.7", "203.0.113.5", // documentation and benchmarking
		"::1", "::", "fe80::1", "fc00::1", "fd12::1", "ff02::1", "100::1", "2001:db8::1", "::ffff:127.0.0.1", "::ffff:10.0.0.1",
		"fec0::1",               // site-local
		"2001::1",               // Teredo
		"2001:10::1",            // ORCHID
		"2001:20::1",            // ORCHIDv2
		"64:ff9b:1::1",          // local-use NAT64
		"64:ff9b::7f00:1",       // NAT64 of 127.0.0.1
		"64:ff9b::a00:1",        // NAT64 of 10.0.0.1
		"64:ff9b::a9fe:a9fe",    // NAT64 of 169.254.169.254
		"2002:7f00:1::",         // 6to4 of 127.0.0.1
		"2002:c0a8:101::",       // 6to4 of 192.168.1.1
		"::7f00:1", "::808:808", // IPv4-compatible (deprecated) is refused outright
	}
	for _, s := range blocked {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("test address %q does not parse", s)
		}
		if IsPublicIP(ip) {
			t.Errorf("%s must be blocked", s)
		}
	}
	public := []string{
		"8.8.8.8", "1.1.1.1", "93.184.216.34", "172.32.0.1", "100.128.0.1", "198.17.255.255", "198.20.0.1", "192.0.3.1",
		"2606:4700:4700::1111", "2a00:1450:4001:80b::200e",
		"64:ff9b::808:808", // NAT64 of 8.8.8.8
		"2002:808:808::",   // 6to4 of 8.8.8.8
	}
	for _, s := range public {
		if !IsPublicIP(net.ParseIP(s)) {
			t.Errorf("%s must be public", s)
		}
	}
	// Malformed addresses fail closed instead of panicking.
	for _, ip := range []net.IP{nil, {}, {0x20, 0x01, 0x0d}, make(net.IP, 5), make(net.IP, 17)} {
		if IsPublicIP(ip) {
			t.Errorf("malformed address %v must be blocked", []byte(ip))
		}
	}
}

func TestResolverFailuresFailClosed(t *testing.T) {
	h := newHarness(t)
	opts := h.options(&fakePolicy{allowed: true}, nil, nil)
	opts.Resolver = mapResolver{"example.com": {}}
	client, _ := New(opts)
	if _, err := client.Get(context.Background(), "https://example.com/feed"); !errors.Is(err, ErrNoAddresses) {
		t.Fatalf("no addresses must fail: %v", err)
	}
	opts.Resolver = mapResolver{}
	client, _ = New(opts)
	if _, err := client.Get(context.Background(), "https://example.com/feed"); err == nil {
		t.Fatal("resolver error must fail")
	}
	// Mixed public/private answers are refused as a whole (DNS rebinding defence).
	opts.Resolver = mapResolver{"example.com": {"93.184.216.34", "127.0.0.1"}}
	client, _ = New(opts)
	if _, err := client.Get(context.Background(), "https://example.com/feed"); !errors.Is(err, ErrPrivateAddress) {
		t.Fatalf("mixed answers must be refused: %v", err)
	}
	// A malformed resolver answer is refused too.
	opts.Resolver = mapResolver{"example.com": {"not-an-ip"}}
	client, _ = New(opts)
	if _, err := client.Get(context.Background(), "https://example.com/feed"); !errors.Is(err, ErrPrivateAddress) {
		t.Fatalf("malformed answer must be refused: %v", err)
	}
	if atomic.LoadInt32(&h.dials) != 0 {
		t.Fatal("no dial may happen when resolution is refused")
	}
}
