// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package fetch implements SafeClient, the only HTTP client the ingestion
// pipeline uses. It is deliberately narrow:
//
//   - https only, allowlisted hostnames only (from config/sources.json), never
//     IP literals or credentials in URLs;
//   - every connection resolves the hostname itself and refuses private,
//     loopback, link-local, multicast, unspecified, documentation, benchmarking
//     and tunnelling addresses (SSRF guard), re-checked on every redirect hop;
//   - at most 3 redirects, each to an allowlisted https URL that the target
//     host's robots.txt permits, each hop waiting the per-host delay;
//   - a robots.txt policy consulted before every request and a per-host delay
//     that honours Crawl-delay up to MaxCrawlDelay (fail closed when the policy
//     cannot answer); waits observe the context, so a run timeout ends them;
//   - robots.txt itself is fetched without following redirects: a 3xx is
//     handed back to the robots package, which treats it as closed;
//   - a hard body cap (8 MiB) enforced with io.LimitReader on the decoded
//     stream, which also bounds gzip decompression bombs: the stdlib transport
//     decompresses transparently and never more than MaxBody+1 bytes are read;
//   - a content-type allowlist, a fixed User-Agent, no cookie jar, no proxy
//     from the environment (a proxy would hide the resolved address from the
//     guard), no authentication of any kind.
package fetch

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Defaults.
const (
	DefaultMaxBody      int64         = 8 << 20
	DefaultTimeout      time.Duration = 20 * time.Second
	DefaultMaxRedirects               = 3
	DefaultMinDelay     time.Duration = 3 * time.Second
	// MaxCrawlDelay caps the Crawl-delay taken from a remote robots.txt. A
	// host asking for more is waited on for MaxCrawlDelay: a file on an
	// allowlisted host must never be able to stall a run for hours.
	MaxCrawlDelay    time.Duration = 60 * time.Second
	DefaultUserAgent               = "pdoom-ingest/0.1 (+https://github.com/1a-li-lu-le-lo/pdoom)"
	maxHeaderBytes                 = 64 << 10
)

// DefaultContentTypes is the default response content-type allowlist. text/xml
// is included because many RSS feeds are served with it.
var DefaultContentTypes = []string{
	"text/html", "application/xml", "text/xml", "application/rss+xml", "application/atom+xml", "application/json", "text/plain",
}

// Sentinel errors; callers use errors.Is.
var (
	ErrNotHTTPS         = errors.New("fetch: only https urls are allowed")
	ErrHostNotAllowed   = errors.New("fetch: host is not on the allowlist")
	ErrIPLiteral        = errors.New("fetch: ip-literal hosts are refused")
	ErrCredentials      = errors.New("fetch: urls with credentials are refused")
	ErrPrivateAddress   = errors.New("fetch: host resolves to a non-public address")
	ErrNoAddresses      = errors.New("fetch: host resolved to no addresses")
	ErrTooManyRedirects = errors.New("fetch: too many redirects")
	ErrBodyTooLarge     = errors.New("fetch: response body exceeds the size cap")
	ErrContentType      = errors.New("fetch: response content-type is not allowed")
	ErrRobotsDisallowed = errors.New("fetch: robots policy disallows the url")
	ErrHTTPStatus       = errors.New("fetch: unexpected http status")
)

// Policy is the robots.txt decision interface (implemented by *robots.Cache).
type Policy interface {
	Allowed(ctx context.Context, rawURL string) (bool, error)
	CrawlDelay(ctx context.Context, rawURL string) time.Duration
}

// Resolver resolves hostnames (net.DefaultResolver in production).
type Resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// Doer performs one HTTP request.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Options configure a SafeClient. Zero values take the defaults. Resolver,
// Dial, TLSConfig, Now and Sleep exist so tests can run against httptest
// servers with the guard fully active.
type Options struct {
	AllowedHosts []string
	UserAgent    string
	Timeout      time.Duration
	MaxBody      int64
	MaxRedirects int
	ContentTypes []string
	MinDelay     time.Duration
	Policy       Policy

	Resolver  Resolver
	Dial      func(ctx context.Context, network, addr string) (net.Conn, error)
	TLSConfig *tls.Config
	Now       func() time.Time
	// Sleep waits for d or until ctx is done, returning ctx.Err() in that case.
	Sleep func(ctx context.Context, d time.Duration) error
}

// Result is a successful fetch.
type Result struct {
	URL         string `json:"url"`
	FinalURL    string `json:"final_url"`
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type"`
	Bytes       int    `json:"bytes"`
	SHA256      string `json:"sha256"`
	Redirects   int    `json:"redirects"`
	Body        []byte `json:"-"`
}

// SafeClient is the guarded HTTP client. It is safe for concurrent use.
type SafeClient struct {
	opts    Options
	allowed map[string]bool
	types   map[string]bool
	// client follows at most MaxRedirects redirects, each re-validated;
	// robotsClient shares the transport but never follows a redirect.
	client       *http.Client
	robotsClient *http.Client

	mu   sync.Mutex
	next map[string]time.Time // host → earliest time of the next request
}

// New builds a SafeClient. Hostnames are lowercased; an empty allowlist
// refuses every request.
func New(opts Options) (*SafeClient, error) {
	if opts.UserAgent == "" {
		opts.UserAgent = DefaultUserAgent
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}
	if opts.MaxBody <= 0 {
		opts.MaxBody = DefaultMaxBody
	}
	if opts.MaxRedirects < 0 {
		opts.MaxRedirects = 0
	} else if opts.MaxRedirects == 0 {
		opts.MaxRedirects = DefaultMaxRedirects
	}
	if opts.MaxRedirects > DefaultMaxRedirects {
		return nil, fmt.Errorf("fetch: MaxRedirects %d exceeds the hard limit %d", opts.MaxRedirects, DefaultMaxRedirects)
	}
	if len(opts.ContentTypes) == 0 {
		opts.ContentTypes = DefaultContentTypes
	}
	if opts.MinDelay < 0 {
		opts.MinDelay = 0
	}
	if opts.MinDelay > MaxCrawlDelay {
		return nil, fmt.Errorf("fetch: MinDelay %v exceeds MaxCrawlDelay %v", opts.MinDelay, MaxCrawlDelay)
	}
	if opts.Resolver == nil {
		opts.Resolver = net.DefaultResolver
	}
	if opts.Dial == nil {
		d := &net.Dialer{Timeout: opts.Timeout}
		opts.Dial = d.DialContext
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Sleep == nil {
		opts.Sleep = sleepContext
	}
	c := &SafeClient{opts: opts, allowed: map[string]bool{}, types: map[string]bool{}, next: map[string]time.Time{}}
	for _, h := range opts.AllowedHosts {
		h = normalizeHost(h)
		if h == "" {
			return nil, fmt.Errorf("fetch: empty hostname in allowlist")
		}
		if net.ParseIP(h) != nil {
			return nil, fmt.Errorf("fetch: allowlist entry %q is an ip literal", h)
		}
		c.allowed[h] = true
	}
	for _, t := range opts.ContentTypes {
		c.types[strings.ToLower(strings.TrimSpace(t))] = true
	}
	tr := &http.Transport{
		Proxy:                  nil,
		DialContext:            c.guardedDial,
		TLSClientConfig:        opts.TLSConfig,
		ForceAttemptHTTP2:      true,
		MaxIdleConns:           4,
		IdleConnTimeout:        30 * time.Second,
		TLSHandshakeTimeout:    10 * time.Second,
		ResponseHeaderTimeout:  opts.Timeout,
		MaxResponseHeaderBytes: maxHeaderBytes,
		DisableCompression:     false,
	}
	c.client = &http.Client{Transport: tr, Timeout: opts.Timeout, CheckRedirect: c.checkRedirect, Jar: nil}
	c.robotsClient = &http.Client{Transport: tr, Timeout: opts.Timeout, CheckRedirect: noRedirects, Jar: nil}
	return c, nil
}

// sleepContext is the default Sleep: a timer that gives up when ctx is done.
func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// noRedirects makes an http.Client return the 3xx response instead of
// following it.
func noRedirects(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

func normalizeHost(h string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(h)), ".")
}

// ValidateURL applies the static checks: https, no credentials, hostname (not
// IP literal) present and on the allowlist.
func (c *SafeClient) ValidateURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("fetch: parse %q: %w", rawURL, err)
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return nil, fmt.Errorf("%w: %q", ErrNotHTTPS, rawURL)
	}
	if u.User != nil {
		return nil, fmt.Errorf("%w: %q", ErrCredentials, rawURL)
	}
	host := normalizeHost(u.Hostname())
	if host == "" {
		return nil, fmt.Errorf("%w: %q has no host", ErrHostNotAllowed, rawURL)
	}
	if net.ParseIP(host) != nil {
		return nil, fmt.Errorf("%w: %q", ErrIPLiteral, rawURL)
	}
	if !c.allowed[host] {
		return nil, fmt.Errorf("%w: %q", ErrHostNotAllowed, host)
	}
	if p := u.Port(); p != "" && p != "443" {
		return nil, fmt.Errorf("%w: port %s is not 443", ErrHostNotAllowed, p)
	}
	return u, nil
}

// blockedPrefixes lists the non-public ranges that netip.Addr's own
// predicates (unspecified, loopback, private, link-local, multicast) do not
// cover.
var blockedPrefixes = []netip.Prefix{
	// IPv4.
	netip.MustParsePrefix("0.0.0.0/8"),       // "this" network
	netip.MustParsePrefix("100.64.0.0/10"),   // carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),    // TEST-NET-1
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking
	netip.MustParsePrefix("198.51.100.0/24"), // TEST-NET-2
	netip.MustParsePrefix("203.0.113.0/24"),  // TEST-NET-3
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved, including broadcast
	// IPv6.
	netip.MustParsePrefix("::/96"),          // IPv4-compatible (deprecated)
	netip.MustParsePrefix("100::/64"),       // discard-only
	netip.MustParsePrefix("2001::/32"),      // Teredo tunnelling
	netip.MustParsePrefix("2001:10::/28"),   // ORCHID (deprecated)
	netip.MustParsePrefix("2001:20::/28"),   // ORCHIDv2
	netip.MustParsePrefix("2001:db8::/32"),  // documentation
	netip.MustParsePrefix("64:ff9b:1::/48"), // local-use NAT64
	netip.MustParsePrefix("fec0::/10"),      // site-local (deprecated)
}

// embeddedIPv4 lists the IPv6 transition prefixes that carry an IPv4 address
// at a fixed byte offset. Such an address is public only if the embedded IPv4
// address is.
var embeddedIPv4 = []struct {
	prefix netip.Prefix
	offset int
}{
	{netip.MustParsePrefix("64:ff9b::/96"), 12}, // NAT64 well-known prefix
	{netip.MustParsePrefix("2002::/16"), 2},     // 6to4
}

// IsPublicIP reports whether ip is a globally routable unicast address. It
// refuses malformed addresses (anything but 4 or 16 bytes), unspecified,
// loopback, private (RFC 1918 / fc00::/7), link-local, site-local, multicast,
// carrier-grade NAT, 0.0.0.0/8, 192.0.0.0/24, the TEST-NET and benchmarking
// ranges, 240.0.0.0/4, the IPv6 discard, documentation, ORCHID, Teredo and
// IPv4-compatible prefixes, and NAT64 / 6to4 addresses whose embedded IPv4
// address is not itself public. Unknown shapes fail closed.
func IsPublicIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	return isPublicAddr(addr.Unmap(), 0)
}

func isPublicAddr(a netip.Addr, depth int) bool {
	if depth > 1 || !a.IsValid() {
		return false
	}
	if a.IsUnspecified() || a.IsLoopback() || a.IsPrivate() || a.IsLinkLocalUnicast() ||
		a.IsLinkLocalMulticast() || a.IsInterfaceLocalMulticast() || a.IsMulticast() {
		return false
	}
	for _, p := range blockedPrefixes {
		if p.Contains(a) {
			return false
		}
	}
	if a.Is6() {
		b := a.As16()
		for _, e := range embeddedIPv4 {
			if e.prefix.Contains(a) {
				v4 := netip.AddrFrom4([4]byte{b[e.offset], b[e.offset+1], b[e.offset+2], b[e.offset+3]})
				return isPublicAddr(v4, depth+1)
			}
		}
	}
	return true
}

// guardedDial resolves the hostname, refuses non-public addresses and dials
// the first resolved address so the checked address is the one connected to.
func (c *SafeClient) guardedDial(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("fetch: dial %q: %w", addr, err)
	}
	host = normalizeHost(host)
	if net.ParseIP(host) != nil {
		return nil, fmt.Errorf("%w: %q", ErrIPLiteral, host)
	}
	if !c.allowed[host] {
		return nil, fmt.Errorf("%w: %q", ErrHostNotAllowed, host)
	}
	addrs, err := c.opts.Resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("fetch: resolve %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrNoAddresses, host)
	}
	for _, a := range addrs {
		if !IsPublicIP(a.IP) {
			return nil, fmt.Errorf("%w: %s → %s", ErrPrivateAddress, host, a.IP)
		}
	}
	return c.opts.Dial(ctx, network, net.JoinHostPort(addrs[0].IP.String(), port))
}

// checkRedirect validates every redirect hop like a fresh request: https,
// allowlist, robots policy for the new URL, and the per-host delay for the
// new host.
func (c *SafeClient) checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > c.opts.MaxRedirects {
		return fmt.Errorf("%w: %d > %d", ErrTooManyRedirects, len(via), c.opts.MaxRedirects)
	}
	u, err := c.ValidateURL(req.URL.String())
	if err != nil {
		return fmt.Errorf("fetch: redirect refused: %w", err)
	}
	ctx := req.Context()
	if err := c.robotsAllow(ctx, u); err != nil {
		return fmt.Errorf("fetch: redirect refused: %w", err)
	}
	host := normalizeHost(u.Hostname())
	if _, err := c.wait(ctx, host, c.delayFor(ctx, u)); err != nil {
		return fmt.Errorf("fetch: redirect to %s: waiting: %w", host, err)
	}
	// Never let a redirect carry credentials or cookies (none are set, but be explicit).
	req.Header.Del("Authorization")
	req.Header.Del("Cookie")
	return nil
}

// policy returns the robots policy (nil when none is installed).
func (c *SafeClient) policy() Policy {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.opts.Policy
}

// robotsAllow consults the robots policy for u. A denial or a policy error is
// reported as ErrRobotsDisallowed (fail closed); without a policy every
// allowlisted URL is allowed.
func (c *SafeClient) robotsAllow(ctx context.Context, u *url.URL) error {
	p := c.policy()
	if p == nil {
		return nil
	}
	ok, err := p.Allowed(ctx, u.String())
	if err != nil {
		return fmt.Errorf("%w: %s: %v", ErrRobotsDisallowed, normalizeHost(u.Hostname()), err)
	}
	if !ok {
		return fmt.Errorf("%w: %s", ErrRobotsDisallowed, u.String())
	}
	return nil
}

// delayFor returns the delay reserved after a request to u: MinDelay or the
// host's Crawl-delay, whichever is larger, capped at MaxCrawlDelay.
func (c *SafeClient) delayFor(ctx context.Context, u *url.URL) time.Duration {
	delay := c.opts.MinDelay
	if p := c.policy(); p != nil {
		if d := p.CrawlDelay(ctx, u.String()); d > delay {
			delay = d
		}
	}
	if delay > MaxCrawlDelay {
		delay = MaxCrawlDelay
	}
	return delay
}

// wait enforces the per-host delay and returns the sleep applied. The sleep
// observes ctx: when the context ends first, ctx.Err() is returned and the
// request is not made. delay is capped at MaxCrawlDelay whatever the caller
// passes.
func (c *SafeClient) wait(ctx context.Context, host string, delay time.Duration) (time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if delay > MaxCrawlDelay {
		delay = MaxCrawlDelay
	}
	if delay < 0 {
		delay = 0
	}
	now := c.opts.Now()
	c.mu.Lock()
	var sleep time.Duration
	if next, ok := c.next[host]; ok && next.After(now) {
		sleep = next.Sub(now)
	}
	c.next[host] = now.Add(sleep + delay)
	c.mu.Unlock()
	if sleep > 0 {
		if err := c.opts.Sleep(ctx, sleep); err != nil {
			return sleep, err
		}
	}
	return sleep, nil
}

// Get fetches a URL with every guard applied. A robots policy error or denial
// is reported as ErrRobotsDisallowed (fail closed); a context that ends while
// waiting for the per-host delay is reported as its ctx.Err().
func (c *SafeClient) Get(ctx context.Context, rawURL string) (*Result, error) {
	u, err := c.ValidateURL(rawURL)
	if err != nil {
		return nil, err
	}
	if err := c.robotsAllow(ctx, u); err != nil {
		return nil, err
	}
	host := normalizeHost(u.Hostname())
	if _, err := c.wait(ctx, host, c.delayFor(ctx, u)); err != nil {
		return nil, fmt.Errorf("fetch: waiting for %s: %w", host, err)
	}
	resp, err := c.do(ctx, u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	res := &Result{URL: rawURL, FinalURL: resp.Request.URL.String(), StatusCode: resp.StatusCode}
	if resp.Request != nil && resp.Request.Response != nil {
		for r := resp.Request.Response; r != nil; r = r.Request.Response {
			res.Redirects++
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d for %s", ErrHTTPStatus, resp.StatusCode, res.FinalURL)
	}
	ct, err := c.checkContentType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	res.ContentType = ct
	body, err := io.ReadAll(io.LimitReader(resp.Body, c.opts.MaxBody+1))
	if err != nil {
		return nil, fmt.Errorf("fetch: read body of %s: %w", res.FinalURL, err)
	}
	if int64(len(body)) > c.opts.MaxBody {
		return nil, fmt.Errorf("%w: more than %d bytes from %s", ErrBodyTooLarge, c.opts.MaxBody, res.FinalURL)
	}
	res.Body = body
	res.Bytes = len(body)
	res.SHA256 = schema.SHA256Hex(body)
	return res, nil
}

func (c *SafeClient) do(ctx context.Context, target string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch: build request: %w", err)
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)
	req.Header.Set("Accept", strings.Join(c.opts.ContentTypes, ", "))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %s: %w", target, err)
	}
	return resp, nil
}

func (c *SafeClient) checkContentType(header string) (string, error) {
	if strings.TrimSpace(header) == "" {
		return "", fmt.Errorf("%w: missing content-type", ErrContentType)
	}
	mt, _, err := mime.ParseMediaType(header)
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrContentType, header)
	}
	mt = strings.ToLower(mt)
	if !c.types[mt] {
		return "", fmt.Errorf("%w: %q", ErrContentType, mt)
	}
	return mt, nil
}

// RobotsDoer returns a Doer for robots.txt fetches: the same https, allowlist,
// SSRF and per-host delay guards, without consulting the robots policy itself
// (which would recurse) and without following redirects (the 3xx response is
// returned as is, and the robots package treats it as closed). The caller
// reads and bounds the body.
func (c *SafeClient) RobotsDoer() Doer {
	return doerFunc(func(req *http.Request) (*http.Response, error) {
		u, err := c.ValidateURL(req.URL.String())
		if err != nil {
			return nil, err
		}
		ctx := req.Context()
		host := normalizeHost(u.Hostname())
		if _, err := c.wait(ctx, host, c.opts.MinDelay); err != nil {
			return nil, fmt.Errorf("fetch: waiting for %s: %w", host, err)
		}
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", c.opts.UserAgent)
		}
		req.Header.Del("Authorization")
		req.Header.Del("Cookie")
		resp, err := c.robotsClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch: %s: %w", u.String(), err)
		}
		return resp, nil
	})
}

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

// SetPolicy installs the robots policy after construction. The robots cache
// needs the client's RobotsDoer and the client needs the cache, so the
// pipeline builds the client first, the cache second and then calls SetPolicy
// before any Get.
func (c *SafeClient) SetPolicy(p Policy) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.opts.Policy = p
}

// AllowedHosts returns the sorted allowlist (for logging).
func (c *SafeClient) AllowedHosts() []string {
	out := make([]string, 0, len(c.allowed))
	for h := range c.allowed {
		out = append(out, h)
	}
	sortStrings(out)
	return out
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
