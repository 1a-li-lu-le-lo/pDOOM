// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"
	"sort"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// researchTargets are the outcome sets published in research mode, keyed by
// the simulation output key.
var researchTargets = []struct {
	Key        string
	OutcomeSet []schema.Outcome
}{
	{"P_DOOM", schema.PDoomOutcomes},
	{"O3", schema.DisempowermentOutcomes},
	{"P_COLLAPSE", schema.CollapseOutcomes},
	{"O6", schema.ExtinctionOutcomes},
	{"O7", schema.BiosphereOutcomes},
}

// causalPerturbations are the parameter perturbations of docs/method/sensitivity.md.
var causalPerturbations = []struct {
	Kind   string
	Change string
	Shifts Shifts
	Lambda *float64
}{
	{"optimistic_safeguards", "F logit −1.0", Shifts{"F": -1}, nil},
	{"pessimistic_safeguards", "F logit +1.0", Shifts{"F": 1}, nil},
	{"slower_capability", "A logit −0.7", Shifts{"A": -0.7}, nil},
	{"faster_capability", "A logit +0.7", Shifts{"A": 0.7}, nil},
	{"lower_exposure", "C and E logit −0.5", Shifts{"C": -0.5, "E": -0.5}, nil},
	{"higher_exposure", "C and E logit +0.5", Shifts{"C": 0.5, "E": 0.5}, nil},
	{"alternative_dependency", "common factor loading λ = 0", nil, f64(0)},
	{"alternative_dependency", "common factor loading λ + 0.3 (capped at 1)", nil, f64(-1)}, // -1 = "increase by 0.3"
}

func f64(v float64) *float64 { return &v }

// Sensitivity computes every sensitivity run for a snapshot given the baseline
// aggregation groups and research-mode simulation results.
func Sensitivity(s *snapshot.Snapshot, groups []Group, base map[string]SimResult) []schema.SensitivityRun {
	var runs []schema.SensitivityRun
	spec := s.ModelSpec.ExperimentalCausal

	// Leave-one-source-out and leave-one-survey-out per aggregation group.
	for _, g := range groups {
		for i, f := range g.Forecasts {
			rest := make([]float64, 0, len(g.Values)-1)
			for j, v := range g.Values {
				if j != i {
					rest = append(rest, v)
				}
			}
			if len(rest) == 0 {
				continue
			}
			val := Median(rest)
			id := f.ID
			runs = append(runs, schema.SensitivityRun{
				RunID: fmt.Sprintf("sens-loo-%s-%s", g.ID, f.ID), Kind: "leave_one_source_out", TargetKind: "aggregation_group", TargetID: g.ID,
				RemovedID: &id, ParameterChange: "remove forecast " + f.ID, Horizon: strPtr(g.Horizon), OutcomeSet: g.OutcomeSet,
				BaselineValue: round6(g.Preferred), Value: round6(val), Delta: round6(val - g.Preferred),
				Note: fmt.Sprintf("Preferred aggregate (median) of %s without %s (%s).", g.ID, f.ID, f.ForecasterOrSurvey),
			})
		}
		// leave-one-survey-out: remove every forecast from the same source at once.
		bySource := map[string][]int{}
		for i, f := range g.Forecasts {
			bySource[f.SourceID] = append(bySource[f.SourceID], i)
		}
		for _, src := range sortedKeys(bySource) {
			idx := bySource[src]
			if len(idx) < 2 {
				continue // identical to leave-one-source-out
			}
			skip := map[int]bool{}
			for _, i := range idx {
				skip[i] = true
			}
			var rest []float64
			for j, v := range g.Values {
				if !skip[j] {
					rest = append(rest, v)
				}
			}
			if len(rest) == 0 {
				continue
			}
			val := Median(rest)
			sid := src
			runs = append(runs, schema.SensitivityRun{
				RunID: fmt.Sprintf("sens-loso-%s-%s", g.ID, src), Kind: "leave_one_survey_out", TargetKind: "aggregation_group", TargetID: g.ID,
				RemovedID: &sid, ParameterChange: fmt.Sprintf("remove all %d forecasts from %s", len(idx), src), Horizon: strPtr(g.Horizon), OutcomeSet: g.OutcomeSet,
				BaselineValue: round6(g.Preferred), Value: round6(val), Delta: round6(val - g.Preferred),
				Note: "Removes every forecast drawn from one source document.",
			})
		}
		// alternative_weighting: spread across methods.
		for _, r := range g.Results {
			if r.Preferred {
				continue
			}
			runs = append(runs, schema.SensitivityRun{
				RunID: fmt.Sprintf("sens-method-%s-%s", g.ID, r.Method), Kind: "alternative_weighting", TargetKind: "aggregation_group", TargetID: g.ID,
				ParameterChange: "method " + string(r.Method), Horizon: strPtr(g.Horizon), OutcomeSet: g.OutcomeSet,
				BaselineValue: round6(g.Preferred), Value: round6(r.Value), Delta: round6(r.Value - g.Preferred),
				Note: "Difference between the preferred aggregate and an alternative aggregation method.",
			})
		}
	}

	// Causal perturbations per horizon and research target.
	horizons := make([]string, 0, len(spec.Horizons))
	for h := range spec.Horizons {
		horizons = append(horizons, h)
	}
	sort.Slice(horizons, func(i, j int) bool {
		return schema.HorizonIndex(schema.Horizon(horizons[i])) < schema.HorizonIndex(schema.Horizon(horizons[j]))
	})
	for _, h := range horizons {
		hs := spec.Horizons[h]
		b, ok := base[h]
		if !ok {
			continue
		}
		for _, p := range causalPerturbations {
			lambda := spec.CommonFactorLoading
			if p.Lambda != nil {
				if *p.Lambda < 0 {
					lambda = lambda + 0.3
					if lambda > 1 {
						lambda = 1
					}
				} else {
					lambda = *p.Lambda
				}
			}
			sim := Simulate(hs, lambda, spec.Seed, spec.Samples, p.Shifts)
			for _, t := range researchTargets {
				bv := b.Outcomes[t.Key].P50
				v := sim.Outcomes[t.Key].P50
				runs = append(runs, schema.SensitivityRun{
					RunID: fmt.Sprintf("sens-%s-%s-%s-%s", p.Kind, slug(p.Change), t.Key, h), Kind: p.Kind, TargetKind: "estimate",
					TargetID: researchEstimateID(t.Key, h), ParameterChange: p.Change, Horizon: strPtr(h), OutcomeSet: t.OutcomeSet,
					BaselineValue: bv, Value: v, Delta: round6(v - bv),
					Note: "Research-mode experimental model re-run with one parameter perturbed; all other parameters unchanged.",
				})
			}
		}
	}

	// Rank by absolute delta (1 = largest influence).
	order := make([]int, len(runs))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		da, db := abs(runs[order[a]].Delta), abs(runs[order[b]].Delta)
		if da != db {
			return da > db
		}
		return runs[order[a]].RunID < runs[order[b]].RunID
	})
	for rank, i := range order {
		runs[i].Rank = rank + 1
	}
	return runs
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func strPtr(s string) *string { return &s }

func slug(s string) string {
	out := make([]rune, 0, len(s))
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out = append(out, r)
			prevDash = false
		case r >= 'A' && r <= 'Z':
			out = append(out, r+('a'-'A'))
			prevDash = false
		default:
			if !prevDash && len(out) > 0 {
				out = append(out, '-')
				prevDash = true
			}
		}
	}
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	return string(out)
}

func researchEstimateID(key, horizon string) string {
	return fmt.Sprintf("est-research-%s-%s", key, horizon)
}
