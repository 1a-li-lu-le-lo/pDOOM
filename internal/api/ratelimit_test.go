// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiterBurstAndRefill(t *testing.T) {
	c := &clock{t: time.Unix(1_700_000_000, 0)}
	l := newLimiter(Limit{Burst: 3, PerMinute: 60}, c.now)
	for i := 0; i < 3; i++ {
		ok, remaining, _ := l.allow("k")
		if !ok || remaining != 2-i {
			t.Fatalf("call %d: ok=%v remaining=%d", i, ok, remaining)
		}
	}
	ok, remaining, retry := l.allow("k")
	if ok || remaining != 0 || retry <= 0 || retry > time.Second {
		t.Fatalf("exhausted: ok=%v remaining=%d retry=%v", ok, remaining, retry)
	}
	c.advance(500 * time.Millisecond)
	if ok, _, retry := l.allow("k"); ok || retry > 500*time.Millisecond {
		t.Fatalf("half a token: ok=%v retry=%v", ok, retry)
	}
	c.advance(500 * time.Millisecond)
	if ok, _, _ := l.allow("k"); !ok {
		t.Fatal("one token should have refilled")
	}
	c.advance(time.Hour)
	if _, remaining, _ := l.allow("k"); remaining != 2 {
		t.Fatalf("refill must cap at burst: remaining %d", remaining)
	}
	if ok, _, _ := l.allow("other"); !ok {
		t.Fatal("keys are independent")
	}
}

func TestLimiterSweepsIdleBuckets(t *testing.T) {
	c := &clock{t: time.Unix(1_700_000_000, 0)}
	l := newLimiter(Limit{Burst: 1, PerMinute: 60}, c.now)
	for i := 0; i < 100; i++ {
		l.allow(fmt.Sprintf("k%d", i))
	}
	if l.size() != 100 {
		t.Fatalf("size %d", l.size())
	}
	c.advance(limiterIdleTTL + time.Minute)
	l.allow("fresh")
	if l.size() != 1 {
		t.Fatalf("idle buckets should be swept, size %d", l.size())
	}
}

func TestLimiterDefaults(t *testing.T) {
	l := newLimiter(Limit{}, time.Now)
	if l.limit.Burst != 1 || l.limit.PerMinute != 60 {
		t.Fatalf("defaults: %+v", l.limit)
	}
	if got := orLimit(Limit{Burst: 5}, DefaultRateLimits.Read); got.Burst != 5 || got.PerMinute != DefaultRateLimits.Read.PerMinute {
		t.Fatalf("orLimit: %+v", got)
	}
}
