// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package observability is a tiny, dependency-free counters/gauges registry
// plus slog helpers for the Go binaries. It exists so that pdoom-ingest can
// print a truthful summary of what it did (fetches, robots denials, parser
// errors, duplicates, queued items) without pulling in a metrics stack.
package observability

import (
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
)

// Well-known counter names used by the ingestion pipeline.
const (
	CounterSourcesConsidered = "sources_considered"
	CounterSourcesSkipped    = "sources_skipped"
	CounterFetches           = "fetches"
	CounterFetchErrors       = "fetch_errors"
	CounterRobotsDenials     = "robots_denials"
	CounterParserErrors      = "parser_errors"
	CounterDocuments         = "documents"
	CounterDuplicates        = "duplicates"
	CounterNearDuplicates    = "near_duplicates"
	CounterCandidateClaims   = "candidate_claims"
	CounterQueued            = "queued_items"
	CounterFilteredSince     = "filtered_before_since"
	GaugeBytesFetched        = "bytes_fetched"
)

// Registry holds named counters and gauges behind a mutex.
type Registry struct {
	mu       sync.Mutex
	counters map[string]int64
	gauges   map[string]float64
}

// New returns an empty registry.
func New() *Registry {
	return &Registry{counters: map[string]int64{}, gauges: map[string]float64{}}
}

// Inc adds delta to a counter (creating it at zero).
func (r *Registry) Inc(name string, delta int64) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[name] += delta
}

// Set stores a gauge value.
func (r *Registry) Set(name string, value float64) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gauges[name] = value
}

// Add adds delta to a gauge.
func (r *Registry) Add(name string, delta float64) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gauges[name] += delta
}

// Counter returns the current value of a counter (zero when absent).
func (r *Registry) Counter(name string) int64 {
	if r == nil {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.counters[name]
}

// Snapshot is a point-in-time copy of the registry.
type Snapshot struct {
	Counters map[string]int64   `json:"counters"`
	Gauges   map[string]float64 `json:"gauges"`
}

// Snapshot copies the registry contents.
func (r *Registry) Snapshot() Snapshot {
	s := Snapshot{Counters: map[string]int64{}, Gauges: map[string]float64{}}
	if r == nil {
		return s
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, v := range r.counters {
		s.Counters[k] = v
	}
	for k, v := range r.gauges {
		s.Gauges[k] = v
	}
	return s
}

// String renders the snapshot as sorted "name=value" pairs, deterministic for
// equal contents so it can appear in golden output and logs.
func (s Snapshot) String() string {
	keys := make([]string, 0, len(s.Counters)+len(s.Gauges))
	for k := range s.Counters {
		keys = append(keys, k)
	}
	for k := range s.Gauges {
		if _, dup := s.Counters[k]; !dup {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if v, ok := s.Counters[k]; ok {
			parts = append(parts, fmt.Sprintf("%s=%d", k, v))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%g", k, s.Gauges[k]))
	}
	return strings.Join(parts, " ")
}

// Attrs converts the snapshot to slog attributes (sorted, deterministic).
func (s Snapshot) Attrs() []any {
	keys := make([]string, 0, len(s.Counters))
	for k := range s.Counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]any, 0, 2*(len(s.Counters)+len(s.Gauges)))
	for _, k := range keys {
		out = append(out, slog.Int64(k, s.Counters[k]))
	}
	gkeys := make([]string, 0, len(s.Gauges))
	for k := range s.Gauges {
		gkeys = append(gkeys, k)
	}
	sort.Strings(gkeys)
	for _, k := range gkeys {
		out = append(out, slog.Float64(k, s.Gauges[k]))
	}
	return out
}

// NewLogger builds a slog logger. JSON output is used when json is true,
// otherwise the text handler. Timestamps are left to slog; the pipeline's own
// data never depends on them.
func NewLogger(w io.Writer, level slog.Level, json bool) *slog.Logger {
	if w == nil {
		w = io.Discard
	}
	opts := &slog.HandlerOptions{Level: level}
	if json {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(slog.NewTextHandler(w, opts))
}

// LogSnapshot logs a registry snapshot at info level with one attribute per metric.
func LogSnapshot(l *slog.Logger, msg string, s Snapshot) {
	if l == nil {
		return
	}
	l.Info(msg, s.Attrs()...)
}
