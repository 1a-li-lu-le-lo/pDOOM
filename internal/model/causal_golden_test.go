// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// causalGolden is the cross-language contract consumed by @pdoom/model-core.
type causalGolden struct {
	RNG struct {
		Seed      int64     `json:"seed"`
		FirstFive []float64 `json:"first_five"`
	} `json:"rng"`
	Spec     schema.ExperimentalCausalSpec `json:"spec"`
	Params   schema.UserScenarioParams     `json:"params"`
	Samples  int                           `json:"samples"`
	Expected schema.UserScenarioResult     `json:"expected"`
}

func TestCausalGolden(t *testing.T) {
	path := filepath.Join("testdata", "causal-golden.json")
	spec := fixtureCausalSpec()
	params := schema.UserScenarioParams{Horizon: "10y", CapabilityTimeline: 1, SafetyProgress: -1, Resilience: 1}
	res, err := EvaluateUserScenario(spec, params, 2000, 0)
	if err != nil {
		t.Fatal(err)
	}
	g := causalGolden{Spec: spec, Params: params, Samples: 2000, Expected: res}
	g.RNG.Seed = 12345
	r := NewMulberry32(12345)
	for i := 0; i < 5; i++ {
		g.RNG.FirstFive = append(g.RNG.FirstFive, r.Next())
	}
	got, _ := schema.CanonicalJSONIndent(g)
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden missing; run with -update: %v", err)
	}
	var w causalGolden
	if err := json.Unmarshal(want, &w); err != nil {
		t.Fatal(err)
	}
	for k, v := range w.Expected.OutcomeEstimates {
		if got := res.OutcomeEstimates[k]; got != v {
			t.Fatalf("outcome %s changed: %+v vs %+v", k, got, v)
		}
	}
}
