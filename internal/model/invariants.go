// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"
	"math"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// monotonicityTolerance is the largest decrease of a research-mode median
// between consecutive horizons that is attributed to Monte Carlo noise.
const monotonicityTolerance = 0.005

// CheckInvariants verifies the structural guarantees every candidate must
// satisfy before it can be promoted. It returns one message per violation.
func CheckInvariants(res *Result) []string {
	var errs []string
	fail := func(format string, args ...any) { errs = append(errs, fmt.Sprintf(format, args...)) }
	seen := map[string]bool{}
	byKeyHorizon := map[string]map[string]float64{}
	for _, e := range res.Estimates {
		if seen[e.EstimateID] {
			fail("duplicate estimate id %s", e.EstimateID)
		}
		seen[e.EstimateID] = true
		if e.Status == schema.StatusInsufficientlyCalibrate {
			if e.Quantiles != nil || e.Mean != nil {
				fail("%s: insufficiently_calibrated estimate must not publish quantiles or a mean", e.EstimateID)
			}
			continue
		}
		if e.Quantiles == nil {
			fail("%s: %s estimate without quantiles", e.EstimateID, e.Status)
			continue
		}
		q := e.Quantiles
		for _, v := range []float64{q.P05, q.P25, q.P50, q.P75, q.P95} {
			if v < 0 || v > 1 || math.IsNaN(v) {
				fail("%s: quantile out of [0,1]", e.EstimateID)
			}
		}
		if !q.Ordered() {
			fail("%s: quantiles not ordered", e.EstimateID)
		}
		if e.Mean != nil && (*e.Mean < 0 || *e.Mean > 1) {
			fail("%s: mean out of [0,1]", e.EstimateID)
		}
		if e.Status == schema.StatusResearchMode {
			key := schema.OutcomeSetKey(e.OutcomeSet)
			if byKeyHorizon[key] == nil {
				byKeyHorizon[key] = map[string]float64{}
			}
			byKeyHorizon[key][e.Horizon] = q.P50
		}
		if e.Display.Central == "" {
			fail("%s: empty display", e.EstimateID)
		}
	}
	// Horizon monotonicity of research-mode medians.
	for key, byH := range byKeyHorizon {
		prev := -1.0
		for _, h := range schema.Horizons {
			v, ok := byH[string(h)]
			if !ok {
				continue
			}
			if prev >= 0 && v < prev-monotonicityTolerance {
				fail("research-mode %s median decreases from %.4f to %.4f at horizon %s", key, prev, v, h)
			}
			if v > prev {
				prev = v
			}
		}
	}
	// No double counting: for each horizon, P_DOOM mean equals the sum of component means.
	for _, h := range schema.Horizons {
		sim, ok := res.Research[string(h)]
		if !ok {
			continue
		}
		sum := 0.0
		for _, o := range outcomeOrder {
			sum += sim.Outcomes[o].Mean
		}
		if d := math.Abs(sum - sim.Outcomes["P_DOOM"].Mean); d > 1e-4 {
			fail("horizon %s: P_DOOM mean %.6f differs from Σ component means %.6f", h, sim.Outcomes["P_DOOM"].Mean, sum)
		}
		if c := sim.Outcomes["O4"].Mean + sim.Outcomes["O5"].Mean; math.Abs(c-sim.Outcomes["P_COLLAPSE"].Mean) > 1e-4 {
			fail("horizon %s: P_COLLAPSE mean inconsistent", h)
		}
	}
	for _, iv := range res.Indexes {
		if iv.IsProbability {
			fail("index %s claims to be a probability", iv.IndexID)
		}
		if iv.Value != nil && (*iv.Value < 0 || *iv.Value > 100 || math.IsNaN(*iv.Value)) {
			fail("index %s out of [0,100]: %v", iv.IndexID, *iv.Value)
		}
		if iv.Coverage < 0 || iv.Coverage > 1 {
			fail("index %s coverage out of [0,1]", iv.IndexID)
		}
		for _, c := range iv.Components {
			if iv.IndexID == schema.IndexCapabilityPressure || iv.IndexID == schema.IndexControlStrength || iv.IndexID == schema.IndexAgenticInfrastructureRisk {
				if c.Weight < 0 || c.Weight > 0.35 {
					fail("index %s: weight %.3f for %s outside [0, 0.35]", iv.IndexID, c.Weight, c.SignalID)
				}
			}
		}
	}
	for _, a := range res.Aggregations {
		if a.Value < 0 || a.Value > 1 {
			fail("aggregation %s value out of [0,1]", a.AggregationID)
		}
		if a.N < 2 {
			fail("aggregation %s has fewer than two members", a.AggregationID)
		}
		wsum := 0.0
		for _, w := range a.Weights {
			wsum += w
		}
		if len(a.Weights) > 0 && math.Abs(wsum-1) > 0.01 {
			fail("aggregation %s weights sum to %.3f", a.AggregationID, wsum)
		}
	}
	if !res.EditorialRiskLevel.Valid() {
		fail("invalid editorial risk level %q", res.EditorialRiskLevel)
	}
	return errs
}
