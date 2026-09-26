// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"math"
	"testing"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Reference values computed with the JavaScript mulberry32 for seed 12345.
// They pin the cross-language contract with @pdoom/model-core.
var mulberryRef12345 = []float64{0.97972826776094735, 0.30675226449966431, 0.484205421525985, 0.81793441250920296, 0.50942836934700608}

func TestMulberry32MatchesJavaScript(t *testing.T) {
	r := NewMulberry32(12345)
	for i, want := range mulberryRef12345 {
		got := r.Next()
		if math.Abs(got-want) > 1e-15 {
			t.Fatalf("draw %d: got %.17g want %.17g", i, got, want)
		}
	}
}

func TestMulberry32Deterministic(t *testing.T) {
	a, b := NewMulberry32(20260926), NewMulberry32(20260926)
	for i := 0; i < 1000; i++ {
		if a.Next() != b.Next() {
			t.Fatal("same seed produced different streams")
		}
	}
	// Uniform-ish: mean of 20k draws near 0.5.
	r := NewMulberry32(7)
	sum := 0.0
	for i := 0; i < 20000; i++ {
		sum += r.Next()
	}
	if m := sum / 20000; math.Abs(m-0.5) > 0.01 {
		t.Fatalf("mean %f far from 0.5", m)
	}
}

func TestLogitNormalFit(t *testing.T) {
	d := LogitNormalFromTri(schema.TriQuantile{P05: 0.02, P50: 0.1, P95: 0.4})
	tri := d.Tri()
	if math.Abs(tri.P50-0.1) > 1e-9 {
		t.Fatalf("median not reproduced: %v", tri)
	}
	// Symmetric-on-logit specs are reproduced at both tails; this one is
	// nearly symmetric so the tails land close.
	if tri.P05 < 0.015 || tri.P05 > 0.03 || tri.P95 < 0.3 || tri.P95 > 0.5 {
		t.Fatalf("tails off: %v", tri)
	}
}

func TestSimulateDeterministicAndBounded(t *testing.T) {
	h := fixtureHorizon()
	a := Simulate(h, 0.5, 20260926, 5000, nil)
	b := Simulate(h, 0.5, 20260926, 5000, nil)
	for k, v := range a.Outcomes {
		if b.Outcomes[k] != v {
			t.Fatalf("nondeterministic outcome %s", k)
		}
		if v.P05 < 0 || v.P95 > 1 || v.P05 > v.P25 || v.P25 > v.P50 || v.P50 > v.P75 || v.P75 > v.P95 {
			t.Fatalf("bad quantiles for %s: %+v", k, v)
		}
	}
	// P_DOOM mean equals the sum of component means (linearity).
	sum := 0.0
	for _, o := range outcomeOrder {
		sum += a.Outcomes[o].Mean
	}
	if math.Abs(sum-a.Outcomes["P_DOOM"].Mean) > 1e-5 {
		t.Fatalf("P_DOOM mean %f != Σ components %f", a.Outcomes["P_DOOM"].Mean, sum)
	}
	// Shifting F down lowers P_DOOM; shifting A up raises it.
	lower := Simulate(h, 0.5, 20260926, 5000, Shifts{"F": -1})
	higher := Simulate(h, 0.5, 20260926, 5000, Shifts{"A": 1})
	if !(lower.Outcomes["P_DOOM"].P50 < a.Outcomes["P_DOOM"].P50 && higher.Outcomes["P_DOOM"].P50 > a.Outcomes["P_DOOM"].P50) {
		t.Fatalf("shift direction wrong: base %f lower %f higher %f", a.Outcomes["P_DOOM"].P50, lower.Outcomes["P_DOOM"].P50, higher.Outcomes["P_DOOM"].P50)
	}
}

func TestEvaluateUserScenario(t *testing.T) {
	spec := fixtureCausalSpec()
	p := schema.UserScenarioParams{Horizon: "10y", SafetyProgress: 2, GovernanceStrength: -2, InternationalCoordination: -2}
	res, err := EvaluateUserScenario(spec, p, 2000, 0)
	if err != nil {
		t.Fatal(err)
	}
	if res.Label != "user_scenario" || len(res.Flags) == 0 || res.SpecVersion != spec.Version {
		t.Fatalf("unexpected result: %+v", res)
	}
	if _, err := EvaluateUserScenario(spec, schema.UserScenarioParams{Horizon: "10y", Resilience: 3}, 100, 0); err == nil {
		t.Fatal("expected range error")
	}
	if _, err := EvaluateUserScenario(spec, schema.UserScenarioParams{Horizon: "99y"}, 100, 0); err == nil {
		t.Fatal("expected horizon error")
	}
}

func fixtureHorizon() schema.HorizonCausalSpec {
	return schema.HorizonCausalSpec{
		A: schema.TriQuantile{P05: 0.2, P50: 0.5, P95: 0.8},
		C: schema.TriQuantile{P05: 0.3, P50: 0.6, P95: 0.9},
		E: schema.TriQuantile{P05: 0.1, P50: 0.3, P95: 0.6},
		F: schema.TriQuantile{P05: 0.05, P50: 0.2, P95: 0.5},
		O: map[string]schema.TriQuantile{
			"O3": {P05: 0.05, P50: 0.2, P95: 0.5}, "O4": {P05: 0.02, P50: 0.1, P95: 0.3}, "O5": {P05: 0.01, P50: 0.05, P95: 0.2},
			"O6": {P05: 0.01, P50: 0.05, P95: 0.2}, "O7": {P05: 0.005, P50: 0.02, P95: 0.1}, "O8": {P05: 0.01, P50: 0.05, P95: 0.2},
		},
	}
}

func fixtureCausalSpec() schema.ExperimentalCausalSpec {
	h := fixtureHorizon()
	return schema.ExperimentalCausalSpec{
		Version: CausalModelVersion, Seed: 20260926, Samples: 4000, CommonFactorLoading: 0.5,
		Horizons:  map[string]schema.HorizonCausalSpec{"5y": h, "10y": h, "25y": h},
		Rationale: schema.CausalRationale{A: "fixture", C: "fixture", E: "fixture", F: "fixture", O: "fixture", Dependence: "fixture"},
		SourceIDs: []string{},
	}
}
