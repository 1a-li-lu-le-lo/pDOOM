// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package snapshot loads, validates and seals a data snapshot directory
// (build-spec §3.5). Loading is strict: every file is decoded with
// DisallowUnknownFields and, when the manifest lists a sha256 for a file, the
// bytes on disk must match it.
package snapshot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Manifest is the snapshot manifest (alias of schema.SnapshotManifest).
type Manifest = schema.SnapshotManifest

// FileSpec describes one entity file of a snapshot.
type FileSpec struct {
	Path     string
	Kind     schema.Kind
	Optional bool
}

// Files lists the entity files of a snapshot in canonical order. actions.json is
// optional; everything else is required.
var Files = []FileSpec{
	{"definitions.json", "definition", false},
	{"sources.json", "source", false},
	{"claims.json", "claim", false},
	{"forecasts.json", "forecast", false},
	{"benchmarks.json", "benchmark", false},
	{"benchmark_results.json", "benchmark_result", false},
	{"incidents.json", "incident", false},
	{"scenarios.json", "scenario", false},
	{"scenario_edges.json", "scenario_edge", false},
	{"drivers.json", "driver", false},
	{"driver_observations.json", "driver_observation", false},
	{"interventions.json", "intervention", false},
	{"organizations.json", "organization", false},
	{"model_spec.json", "model_spec", false},
	{"actions.json", "action", true},
}

// Snapshot is a fully decoded snapshot directory.
type Snapshot struct {
	Dir                string
	Manifest           Manifest
	Definitions        []schema.Definition
	Sources            []schema.Source
	Claims             []schema.Claim
	Forecasts          []schema.Forecast
	Benchmarks         []schema.Benchmark
	BenchmarkResults   []schema.BenchmarkResult
	Incidents          []schema.Incident
	Scenarios          []schema.Scenario
	ScenarioEdges      []schema.ScenarioEdge
	Drivers            []schema.Driver
	DriverObservations []schema.DriverObservation
	Interventions      []schema.Intervention
	Organizations      []schema.Organization
	Actions            []schema.Action
	ModelSpec          schema.ModelSpec

	// RawFiles holds the bytes of every file that was read (keyed by relative
	// path) so that Validate can apply JSON Schemas to the original documents.
	RawFiles map[string][]byte
}

// Load reads manifest.json and every entity file of dir.
func Load(dir string) (*Snapshot, error) {
	s := &Snapshot{Dir: dir, RawFiles: map[string][]byte{}}
	mraw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("snapshot: read manifest: %w", err)
	}
	if err := decodeStrict(mraw, &s.Manifest); err != nil {
		return nil, fmt.Errorf("snapshot: manifest.json: %w", err)
	}
	s.RawFiles["manifest.json"] = mraw
	listed := map[string]schema.SnapshotFile{}
	for _, f := range s.Manifest.Files {
		listed[f.Path] = f
	}

	var errs []error
	for _, spec := range Files {
		raw, err := os.ReadFile(filepath.Join(dir, spec.Path))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && spec.Optional {
				continue
			}
			errs = append(errs, fmt.Errorf("%s: %w", spec.Path, err))
			continue
		}
		s.RawFiles[spec.Path] = raw
		if lf, ok := listed[spec.Path]; ok && lf.SHA256 != "" {
			if got := schema.SHA256Hex(raw); got != lf.SHA256 {
				errs = append(errs, fmt.Errorf("%s: sha256 mismatch (manifest %s, file %s)", spec.Path, lf.SHA256, got))
				continue
			}
		}
		if err := s.decodeFile(spec, raw); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", spec.Path, err))
		}
	}
	// Files listed in the manifest that we do not know are an error: the loader
	// must not silently ignore data.
	for _, f := range s.Manifest.Files {
		if !knownFile(f.Path) && f.Path != "manifest.json" {
			errs = append(errs, fmt.Errorf("%s: listed in manifest but not a known snapshot file", f.Path))
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("snapshot: %w", errors.Join(errs...))
	}
	return s, nil
}

func knownFile(path string) bool {
	for _, f := range Files {
		if f.Path == path {
			return true
		}
	}
	return false
}

func (s *Snapshot) decodeFile(spec FileSpec, raw []byte) error {
	var kindErr error
	check := func(kind string) {
		if kind != string(spec.Kind) {
			kindErr = fmt.Errorf("kind %q, expected %q", kind, spec.Kind)
		}
	}
	switch spec.Kind {
	case "definition":
		env, err := decodeEnvelope[schema.Definition](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Definitions = env.Items
	case "source":
		env, err := decodeEnvelope[schema.Source](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Sources = env.Items
	case "claim":
		env, err := decodeEnvelope[schema.Claim](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Claims = env.Items
	case "forecast":
		env, err := decodeEnvelope[schema.Forecast](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Forecasts = env.Items
	case "benchmark":
		env, err := decodeEnvelope[schema.Benchmark](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Benchmarks = env.Items
	case "benchmark_result":
		env, err := decodeEnvelope[schema.BenchmarkResult](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.BenchmarkResults = env.Items
	case "incident":
		env, err := decodeEnvelope[schema.Incident](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Incidents = env.Items
	case "scenario":
		env, err := decodeEnvelope[schema.Scenario](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Scenarios = env.Items
	case "scenario_edge":
		env, err := decodeEnvelope[schema.ScenarioEdge](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.ScenarioEdges = env.Items
	case "driver":
		env, err := decodeEnvelope[schema.Driver](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Drivers = env.Items
	case "driver_observation":
		env, err := decodeEnvelope[schema.DriverObservation](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.DriverObservations = env.Items
	case "intervention":
		env, err := decodeEnvelope[schema.Intervention](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Interventions = env.Items
	case "organization":
		env, err := decodeEnvelope[schema.Organization](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Organizations = env.Items
	case "action":
		env, err := decodeEnvelope[schema.Action](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		s.Actions = env.Items
	case "model_spec":
		env, err := decodeEnvelope[schema.ModelSpec](raw)
		if err != nil {
			return err
		}
		check(env.Kind)
		if len(env.Items) != 1 {
			return fmt.Errorf("model_spec must hold exactly one item, got %d", len(env.Items))
		}
		s.ModelSpec = env.Items[0]
	default:
		return fmt.Errorf("unknown kind %q", spec.Kind)
	}
	return kindErr
}

func decodeEnvelope[T any](raw []byte) (schema.Envelope[T], error) {
	var env schema.Envelope[T]
	err := decodeStrict(raw, &env)
	return env, err
}

// decodeStrict decodes JSON with DisallowUnknownFields and rejects trailing data.
func decodeStrict(raw []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		return fmt.Errorf("trailing data after JSON document")
	}
	return nil
}

// ItemCount returns the number of items in the named file.
func (s *Snapshot) ItemCount(path string) int {
	switch path {
	case "definitions.json":
		return len(s.Definitions)
	case "sources.json":
		return len(s.Sources)
	case "claims.json":
		return len(s.Claims)
	case "forecasts.json":
		return len(s.Forecasts)
	case "benchmarks.json":
		return len(s.Benchmarks)
	case "benchmark_results.json":
		return len(s.BenchmarkResults)
	case "incidents.json":
		return len(s.Incidents)
	case "scenarios.json":
		return len(s.Scenarios)
	case "scenario_edges.json":
		return len(s.ScenarioEdges)
	case "drivers.json":
		return len(s.Drivers)
	case "driver_observations.json":
		return len(s.DriverObservations)
	case "interventions.json":
		return len(s.Interventions)
	case "organizations.json":
		return len(s.Organizations)
	case "actions.json":
		return len(s.Actions)
	case "model_spec.json":
		return 1
	}
	return 0
}

// SourceByID returns the source with the given id, or nil.
func (s *Snapshot) SourceByID(id string) *schema.Source {
	for i := range s.Sources {
		if s.Sources[i].ID == id {
			return &s.Sources[i]
		}
	}
	return nil
}

// SignalIndex maps signal_id → (driver family id, signal) over drivers.json.
func (s *Snapshot) SignalIndex() map[string]SignalRef {
	out := map[string]SignalRef{}
	for _, d := range s.Drivers {
		for _, sig := range d.Signals {
			out[sig.SignalID] = SignalRef{Family: d.ID, Signal: sig}
		}
	}
	return out
}

// SignalRef is a signal together with its driver family.
type SignalRef struct {
	Family string
	Signal schema.Signal
}

// BestTier returns the best (lowest) tier among the given source ids, or 5 when
// none resolves (unknown sources are treated as unverified).
func (s *Snapshot) BestTier(sourceIDs []string) schema.SourceTier {
	best := schema.SourceTier(5)
	found := false
	for _, id := range sourceIDs {
		src := s.SourceByID(id)
		if src == nil {
			continue
		}
		if !found || src.SourceTier < best {
			best = src.SourceTier
			found = true
		}
	}
	return best
}

// SortedFilePaths returns the entity file paths present in the snapshot.
func (s *Snapshot) SortedFilePaths() []string {
	paths := make([]string, 0, len(s.RawFiles))
	for p := range s.RawFiles {
		if p != "manifest.json" {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	return paths
}
