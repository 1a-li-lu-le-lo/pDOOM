// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// middleware wraps the mux with request ids, security headers, the request
// line limit, panic recovery and the access log.
func (s *server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := s.now()
		rid := requestID(r)
		h := w.Header()
		h.Set("X-Request-ID", rid)
		securityHeaders(h)
		rw := &recorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Error("handler panic", "request_id", rid, "path", r.URL.Path, "panic", recoverString(rec))
				if !rw.wrote {
					s.writeError(rw, r, http.StatusInternalServerError, "internal", "internal error")
				}
			}
			s.log.Info("request",
				"request_id", rid,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"bytes", rw.bytes,
				"duration_ms", s.now().Sub(start).Milliseconds(),
				"client", s.hashedClient(r),
			)
		}()
		if len(r.URL.RequestURI()) > MaxURILength {
			s.writeError(rw, r, http.StatusRequestURITooLong, "uri_too_long", "request line exceeds "+strconv.Itoa(MaxURILength)+" bytes")
			return
		}
		next.ServeHTTP(rw, r)
	})
}

func recoverString(v any) string {
	switch t := v.(type) {
	case error:
		return t.Error()
	case string:
		return t
	}
	return "panic"
}

// securityHeaders are set on every response. The API serves JSON and
// markdown only, so scripts, frames and embedding are all denied.
func securityHeaders(h http.Header) {
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; sandbox")
	h.Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Cross-Origin-Resource-Policy", "cross-origin")
	h.Set("Vary", "Origin")
}

// requestID echoes a well-formed client id or mints a random one.
func requestID(r *http.Request) string {
	if v := r.Header.Get("X-Request-ID"); reRequestID.MatchString(v) {
		return v
	}
	return hex.EncodeToString(randomBytes(12))
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is exceptional; fall back to a time-derived value
		// so that the server keeps answering.
		t := time.Now().UnixNano()
		for i := range b {
			b[i] = byte(t >> (8 * (i % 8)))
		}
	}
	return b
}

// recorder captures the status and byte count for the access log.
type recorder struct {
	http.ResponseWriter
	status int
	bytes  int
	wrote  bool
}

func (r *recorder) WriteHeader(code int) {
	if !r.wrote {
		r.status = code
		r.wrote = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(p []byte) (int, error) {
	if !r.wrote {
		r.wrote = true
	}
	n, err := r.ResponseWriter.Write(p)
	r.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// ipv6ClientPrefix is the prefix length IPv6 clients are keyed on. A single
// subscriber usually holds a whole /64, so keying on the full address would
// hand out unlimited fresh buckets.
const ipv6ClientPrefix = 64

// clientIP returns the address used to key rate limits: the IPv4 address, or
// the /64 prefix of an IPv6 address. X-Forwarded-For is honoured only when the
// operator declared a trusted proxy, and then only its right-most entry (the
// one appended by that proxy).
func (s *server) clientIP(r *http.Request) string {
	if s.deps.TrustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				if ip := net.ParseIP(strings.TrimSpace(parts[i])); ip != nil {
					return clientKey(ip)
				}
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil {
		return clientKey(ip)
	}
	return host
}

// clientKey normalizes an address: IPv4 (including IPv4-mapped IPv6) as is,
// IPv6 masked to its /64 prefix.
func clientKey(ip net.IP) string {
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.Mask(net.CIDRMask(ipv6ClientPrefix, 8*net.IPv6len)).String() + "/" + strconv.Itoa(ipv6ClientPrefix)
}

// hashedClient is the only client identifier that reaches the log: a salted,
// truncated hash of the address.
func (s *server) hashedClient(r *http.Request) string {
	sum := sha256.Sum256(append(append([]byte{}, s.salt...), []byte(s.clientIP(r))...))
	return hex.EncodeToString(sum[:6])
}

// normalizeOrigin lower-cases scheme and host of an origin and drops paths.
func normalizeOrigin(o string) string {
	u, err := url.Parse(strings.TrimSpace(o))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
}

// requestOrigin reconstructs the origin the request was addressed to. The
// scheme comes from the connection, or from X-Forwarded-Proto when the
// operator declared a trusted proxy.
func (s *server) requestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if s.deps.TrustProxy {
		if p := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))); p == "https" || p == "http" {
			scheme = p
		}
	}
	return scheme + "://" + strings.ToLower(r.Host)
}

// originAllowedForPOST reports whether a browser origin may POST: same-origin
// or allowlisted through PDOOM_CORS_ORIGINS. An https origin on the request's
// own host counts as same-origin even when this process sees plain HTTP,
// which is the usual deployment with TLS terminated by a reverse proxy; the
// reverse (an http origin against an https request) does not.
func (s *server) originAllowedForPOST(r *http.Request, origin string) bool {
	n := normalizeOrigin(origin)
	if n == "" {
		return false
	}
	if s.origins[n] {
		return true
	}
	return n == s.requestOrigin(r) || n == "https://"+strings.ToLower(r.Host)
}

const exposeHeaders = "ETag, X-Request-ID, RateLimit-Limit, RateLimit-Remaining, Retry-After"

// applyCORS sets the CORS headers of an actual request. GET is open to any
// origin; POST is same-origin or allowlisted, otherwise 403.
func (s *server) applyCORS(w http.ResponseWriter, r *http.Request, rt route) bool {
	h := w.Header()
	if rt.method == http.MethodGet {
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Expose-Headers", exposeHeaders)
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if s.originAllowedForPOST(r, origin) {
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Expose-Headers", exposeHeaders)
		return true
	}
	s.writeError(w, r, http.StatusForbidden, "origin_not_allowed", "cross-origin POST is limited to allowlisted origins (PDOOM_CORS_ORIGINS)")
	return false
}

// preflight answers OPTIONS, including CORS preflight requests.
func (s *server) preflight(w http.ResponseWriter, r *http.Request, methods map[string]route, allow string) {
	h := w.Header()
	h.Set("Allow", allow)
	origin := r.Header.Get("Origin")
	reqMethod := strings.ToUpper(strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")))
	if origin == "" || reqMethod == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if reqMethod == http.MethodHead {
		reqMethod = http.MethodGet
	}
	rt, ok := methods[reqMethod]
	if !ok {
		s.writeError(w, r, http.StatusForbidden, "method_not_allowed", "no such method on this route; allowed: "+allow)
		return
	}
	h.Set("Access-Control-Max-Age", "600")
	if rt.method == http.MethodGet {
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, If-None-Match, X-Request-ID")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if !s.originAllowedForPOST(r, origin) {
		s.writeError(w, r, http.StatusForbidden, "origin_not_allowed", "cross-origin POST is limited to allowlisted origins (PDOOM_CORS_ORIGINS)")
		return
	}
	h.Set("Access-Control-Allow-Origin", origin)
	h.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
	w.WriteHeader(http.StatusNoContent)
}

// rateLimit applies the class limiter to the client; 429 when exhausted.
func (s *server) rateLimit(w http.ResponseWriter, r *http.Request, class string) bool {
	lim := s.limits[class]
	if lim == nil {
		return true
	}
	ok, remaining, retry := lim.allow(class + ":" + s.clientIP(r))
	h := w.Header()
	h.Set("RateLimit-Limit", strconv.Itoa(lim.limit.Burst))
	h.Set("RateLimit-Remaining", strconv.Itoa(remaining))
	if ok {
		return true
	}
	secs := int(retry / time.Second)
	if retry%time.Second != 0 {
		secs++
	}
	if secs < 1 {
		secs = 1
	}
	h.Set("Retry-After", strconv.Itoa(secs))
	s.writeError(w, r, http.StatusTooManyRequests, "rate_limited", "per-client limit for "+class+" routes reached; retry after "+strconv.Itoa(secs)+"s")
	return false
}
