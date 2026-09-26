// Copyright NU Cybernetics. p(DOOM) — research prototype.

package observability

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"
)

func TestRegistryCountersAndGauges(t *testing.T) {
	r := New()
	r.Inc(CounterFetches, 1)
	r.Inc(CounterFetches, 2)
	r.Set(GaugeBytesFetched, 1024)
	r.Add(GaugeBytesFetched, 1)
	if got := r.Counter(CounterFetches); got != 3 {
		t.Fatalf("fetches = %d, want 3", got)
	}
	s := r.Snapshot()
	if s.Counters[CounterFetches] != 3 || s.Gauges[GaugeBytesFetched] != 1025 {
		t.Fatalf("snapshot mismatch: %+v", s)
	}
	// The snapshot is a copy: later increments do not leak into it.
	r.Inc(CounterFetches, 10)
	if s.Counters[CounterFetches] != 3 {
		t.Fatal("snapshot must be a copy")
	}
}

func TestSnapshotStringIsSortedAndDeterministic(t *testing.T) {
	r := New()
	r.Inc("zeta", 1)
	r.Inc("alpha", 2)
	r.Set("mid", 0.5)
	want := "alpha=2 mid=0.5 zeta=1"
	for i := 0; i < 5; i++ {
		if got := r.Snapshot().String(); got != want {
			t.Fatalf("String() = %q, want %q", got, want)
		}
	}
}

func TestNilRegistryIsSafe(t *testing.T) {
	var r *Registry
	r.Inc("x", 1)
	r.Set("y", 1)
	r.Add("y", 1)
	if r.Counter("x") != 0 {
		t.Fatal("nil registry counter must be zero")
	}
	if s := r.Snapshot(); len(s.Counters) != 0 || len(s.Gauges) != 0 {
		t.Fatal("nil registry snapshot must be empty")
	}
}

func TestConcurrentIncrements(t *testing.T) {
	r := New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				r.Inc(CounterQueued, 1)
			}
		}()
	}
	wg.Wait()
	if got := r.Counter(CounterQueued); got != 5000 {
		t.Fatalf("queued = %d, want 5000", got)
	}
}

func TestLoggerAndLogSnapshot(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, slog.LevelInfo, false)
	r := New()
	r.Inc(CounterRobotsDenials, 2)
	r.Set(GaugeBytesFetched, 7)
	LogSnapshot(l, "summary", r.Snapshot())
	out := buf.String()
	if !strings.Contains(out, "robots_denials=2") || !strings.Contains(out, "bytes_fetched=7") || !strings.Contains(out, "summary") {
		t.Fatalf("unexpected log output: %s", out)
	}
	buf.Reset()
	jl := NewLogger(&buf, slog.LevelInfo, true)
	LogSnapshot(jl, "summary", r.Snapshot())
	if !strings.Contains(buf.String(), `"robots_denials":2`) {
		t.Fatalf("json log missing counter: %s", buf.String())
	}
	// nil writer and nil logger must not panic.
	NewLogger(nil, slog.LevelInfo, false).Info("discarded")
	LogSnapshot(nil, "x", Snapshot{})
}
