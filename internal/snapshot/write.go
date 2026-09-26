// Copyright NU Cybernetics. p(DOOM) — research prototype.

package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Write serializes a snapshot to dir (creating it) and seals the manifest. It
// is used by tests and by tooling that assembles snapshots; it never touches
// data/releases.
func Write(dir string, s *Snapshot) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	write := func(path string, kind string, items any) error {
		env := map[string]any{"kind": kind, "schema_version": 1, "items": items}
		raw, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, path), append(raw, '\n'), 0o644)
	}
	steps := []struct {
		path, kind string
		items      any
	}{
		{"definitions.json", "definition", nonNil(s.Definitions)},
		{"sources.json", "source", nonNil(s.Sources)},
		{"claims.json", "claim", nonNil(s.Claims)},
		{"forecasts.json", "forecast", nonNil(s.Forecasts)},
		{"benchmarks.json", "benchmark", nonNil(s.Benchmarks)},
		{"benchmark_results.json", "benchmark_result", nonNil(s.BenchmarkResults)},
		{"incidents.json", "incident", nonNil(s.Incidents)},
		{"scenarios.json", "scenario", nonNil(s.Scenarios)},
		{"scenario_edges.json", "scenario_edge", nonNil(s.ScenarioEdges)},
		{"drivers.json", "driver", nonNil(s.Drivers)},
		{"driver_observations.json", "driver_observation", nonNil(s.DriverObservations)},
		{"interventions.json", "intervention", nonNil(s.Interventions)},
		{"organizations.json", "organization", nonNil(s.Organizations)},
		{"model_spec.json", "model_spec", []schema.ModelSpec{s.ModelSpec}},
	}
	for _, st := range steps {
		if err := write(st.path, st.kind, st.items); err != nil {
			return fmt.Errorf("write %s: %w", st.path, err)
		}
	}
	if len(s.Actions) > 0 {
		if err := write("actions.json", "action", s.Actions); err != nil {
			return err
		}
	}
	m := s.Manifest
	m.Files = nil
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	_, err = Seal(dir)
	return err
}

func nonNil[T any](xs []T) []T {
	if xs == nil {
		return []T{}
	}
	return xs
}
