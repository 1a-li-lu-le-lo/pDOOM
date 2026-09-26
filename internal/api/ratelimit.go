// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"math"
	"sync"
	"time"
)

// limiter is a per-key token bucket. Buckets idle for longer than idleTTL are
// swept periodically so that the map stays bounded.
type limiter struct {
	limit Limit
	rate  float64 // tokens per second
	now   func() time.Time

	mu        sync.Mutex
	buckets   map[string]*bucket
	lastSweep time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

const (
	limiterIdleTTL       = 10 * time.Minute
	limiterSweepInterval = time.Minute
	limiterHardCap       = 50000
)

func newLimiter(l Limit, now func() time.Time) *limiter {
	if l.Burst <= 0 {
		l.Burst = 1
	}
	if l.PerMinute <= 0 {
		l.PerMinute = 60
	}
	return &limiter{limit: l, rate: l.PerMinute / 60, now: now, buckets: map[string]*bucket{}}
}

// allow takes one token for key. It returns whether the request may proceed,
// the tokens left and, when refused, how long until one token is available.
func (l *limiter) allow(key string) (ok bool, remaining int, retryAfter time.Duration) {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweep(now)
	b, exists := l.buckets[key]
	if !exists {
		b = &bucket{tokens: float64(l.limit.Burst), last: now}
		l.buckets[key] = b
	} else {
		elapsed := now.Sub(b.last).Seconds()
		if elapsed > 0 {
			b.tokens = math.Min(float64(l.limit.Burst), b.tokens+elapsed*l.rate)
		}
		b.last = now
	}
	if b.tokens >= 1 {
		b.tokens--
		return true, int(math.Floor(b.tokens)), 0
	}
	need := 1 - b.tokens
	wait := time.Duration(need / l.rate * float64(time.Second))
	return false, 0, wait
}

// sweep drops idle buckets once per interval, or aggressively when the map is
// unexpectedly large.
func (l *limiter) sweep(now time.Time) {
	if len(l.buckets) < limiterHardCap && !l.lastSweep.IsZero() && now.Sub(l.lastSweep) < limiterSweepInterval {
		return
	}
	l.lastSweep = now
	ttl := limiterIdleTTL
	if len(l.buckets) >= limiterHardCap {
		ttl = time.Minute
	}
	for k, b := range l.buckets {
		if now.Sub(b.last) > ttl {
			delete(l.buckets, k)
		}
	}
}

// size reports the number of tracked clients (for tests).
func (l *limiter) size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
