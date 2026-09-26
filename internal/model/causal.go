// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"
	"sort"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// CausalModelVersion is the version string of the experimental causal model.
// The model_spec carries its own experimental_causal.version; they must agree.
const CausalModelVersion = "pdoom-model/experimental-causal@0.1.0"

// outcomeOrder is the fixed order in which outcome-share deviates are drawn.
var outcomeOrder = []string{"O3", "O4", "O5", "O6", "O7", "O8"}

// Shifts are logit-scale location shifts applied before simulation. Keys are
// "A", "C", "E", "F" or an outcome code "O3".."O8".
type Shifts map[string]float64

// SimResult is the output of one simulation of one horizon.
type SimResult struct {
	// Outcomes holds summaries keyed by O3..O8, P_DOOM and P_COLLAPSE.
	Outcomes map[string]schema.OutcomeEstimate
	// Factors holds the {p05,p50,p95} implied by the (possibly shifted) factor distributions.
	Factors map[string]schema.TriQuantile
}

// Simulate runs the experimental causal decomposition for one horizon.
//
// Draw order per sample (each normal consumes two uniforms):
//
//	Z (common factor), εA, εC, εE, εF, εO3, εO4, εO5, εO6, εO7, εO8
//
// For each factor X ∈ {A, C, E, F}: x = sigmoid(mu_X + sigma_X·(λZ + √(1−λ²)·εX)).
// reach = A·C·E·F is the probability that a sufficiently capable system is
// deployed with autonomy, dangerous exposure occurs and safeguards fail.
// Outcome shares s_i (conditional on reach) are sampled the same way; if their
// sum exceeds 1 they are normalized to sum to 1. P(O_i) = reach·s_i,
// P_DOOM = Σ_{O3..O8} P(O_i), P_COLLAPSE = P(O4) + P(O5).
func Simulate(h schema.HorizonCausalSpec, lambda float64, seed int64, samples int, shifts Shifts) SimResult {
	if samples <= 0 {
		samples = 1
	}
	factors := map[string]LogitNormal{
		"A": LogitNormalFromTri(h.A).Shifted(shifts["A"]),
		"C": LogitNormalFromTri(h.C).Shifted(shifts["C"]),
		"E": LogitNormalFromTri(h.E).Shifted(shifts["E"]),
		"F": LogitNormalFromTri(h.F).Shifted(shifts["F"]),
	}
	shares := map[string]LogitNormal{}
	for _, o := range outcomeOrder {
		if t, ok := h.O[o]; ok {
			shares[o] = LogitNormalFromTri(t).Shifted(shifts[o])
		}
	}
	lam := lambda
	if lam < 0 {
		lam = 0
	}
	if lam > 1 {
		lam = 1
	}
	rest := sqrt1m(lam)

	rng := NewMulberry32(seed)
	out := map[string][]float64{"P_DOOM": make([]float64, samples), "P_COLLAPSE": make([]float64, samples)}
	for _, o := range outcomeOrder {
		out[o] = make([]float64, samples)
	}
	for i := 0; i < samples; i++ {
		z := rng.Normal()
		reach := 1.0
		for _, k := range []string{"A", "C", "E", "F"} {
			eps := rng.Normal()
			reach *= factors[k].At(lam*z + rest*eps)
		}
		s := map[string]float64{}
		sum := 0.0
		for _, o := range outcomeOrder {
			eps := rng.Normal()
			d, ok := shares[o]
			if !ok {
				continue
			}
			v := d.At(lam*z + rest*eps)
			s[o] = v
			sum += v
		}
		if sum > 1 {
			for o := range s {
				s[o] /= sum
			}
		}
		doom := 0.0
		for _, o := range outcomeOrder {
			p := reach * s[o]
			out[o][i] = p
			doom += p
		}
		out["P_DOOM"][i] = doom
		out["P_COLLAPSE"][i] = out["O4"][i] + out["O5"][i]
	}
	res := SimResult{Outcomes: map[string]schema.OutcomeEstimate{}, Factors: map[string]schema.TriQuantile{}}
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sum := Summarize(out[k])
		res.Outcomes[k] = schema.OutcomeEstimate{P05: round6(sum.P05), P25: round6(sum.P25), P50: round6(sum.P50), P75: round6(sum.P75), P95: round6(sum.P95), Mean: round6(sum.Mean)}
	}
	for k, d := range factors {
		t := d.Tri()
		res.Factors[k] = schema.TriQuantile{P05: round6(t.P05), P50: round6(t.P50), P95: round6(t.P95)}
	}
	return res
}

func sqrt1m(l float64) float64 {
	v := 1 - l*l
	if v <= 0 {
		return 0
	}
	// math.Sqrt is correctly rounded in both Go and JavaScript.
	return sqrt(v)
}

// SliderShifts maps Scenario Lab sliders to logit-scale shifts. The table is the
// documented contract of docs/method/model.md ("Scenario Lab"):
//
//	capability_timeline      → A  +0.5 per step
//	autonomy_growth          → C  +0.4
//	access_level             → C, E +0.3 each
//	safety_progress          → F  −0.5
//	governance_strength      → F  −0.4
//	model_security           → E  −0.3
//	open_weight_diffusion    → E  +0.3
//	international_coordination → F −0.3
//	incident_frequency       → E  +0.2
//	resilience               → O4, O5, O6 shares −0.3
func SliderShifts(p schema.UserScenarioParams) Shifts {
	s := Shifts{}
	add := func(k string, v float64) { s[k] += v }
	add("A", 0.5*float64(p.CapabilityTimeline))
	add("C", 0.4*float64(p.AutonomyGrowth))
	add("C", 0.3*float64(p.AccessLevel))
	add("E", 0.3*float64(p.AccessLevel))
	add("F", -0.5*float64(p.SafetyProgress))
	add("F", -0.4*float64(p.GovernanceStrength))
	add("E", -0.3*float64(p.ModelSecurity))
	add("E", 0.3*float64(p.OpenWeightDiffusion))
	add("F", -0.3*float64(p.InternationalCoordination))
	add("E", 0.2*float64(p.IncidentFrequency))
	add("O4", -0.3*float64(p.Resilience))
	add("O5", -0.3*float64(p.Resilience))
	add("O6", -0.3*float64(p.Resilience))
	return s
}

// ScenarioFlags reports implausible or tense slider combinations. The lab never
// refuses a combination; it labels it.
func ScenarioFlags(p schema.UserScenarioParams) []string {
	var flags []string
	if p.SafetyProgress >= 2 && p.GovernanceStrength <= -2 && p.InternationalCoordination <= -2 {
		flags = append(flags, "implausible: strong safety progress alongside collapsing governance and coordination")
	}
	if p.CapabilityTimeline <= -2 && p.AutonomyGrowth >= 2 {
		flags = append(flags, "implausible: much slower capability growth alongside much faster autonomy growth")
	}
	if p.OpenWeightDiffusion >= 2 && p.ModelSecurity >= 2 {
		flags = append(flags, "tension: maximal open-weight diffusion alongside maximal model-weight security")
	}
	if p.AccessLevel <= -2 && p.IncidentFrequency >= 2 {
		flags = append(flags, "tension: minimal system access alongside a much higher incident rate")
	}
	return flags
}

// UserScenarioDisclaimer is the sentence every Scenario Lab result carries.
const UserScenarioDisclaimer = "Under your selected assumptions—not the p(DOOM) official model—the median estimate is shown below. This is a user-generated scenario computed by the experimental research-mode model; it is not a published estimate and does not change any official value."

// EvaluateUserScenario runs the experimental model with the user's sliders. It
// validates the sliders and horizon, applies the documented shifts and returns
// a result labelled user_scenario.
func EvaluateUserScenario(spec schema.ExperimentalCausalSpec, p schema.UserScenarioParams, samples int, seed int64) (schema.UserScenarioResult, error) {
	for name, v := range p.Sliders() {
		if v < -2 || v > 2 {
			return schema.UserScenarioResult{}, fmt.Errorf("slider %s out of range -2..2: %d", name, v)
		}
	}
	h, ok := spec.Horizons[p.Horizon]
	if !ok {
		return schema.UserScenarioResult{}, fmt.Errorf("horizon %q is not covered by the experimental model", p.Horizon)
	}
	if samples <= 0 || samples > spec.Samples {
		samples = spec.Samples
	}
	if seed == 0 {
		seed = spec.Seed
	}
	sim := Simulate(h, spec.CommonFactorLoading, seed, samples, SliderShifts(p))
	flags := ScenarioFlags(p)
	if flags == nil {
		flags = []string{}
	}
	return schema.UserScenarioResult{
		Label:            "user_scenario",
		Disclaimer:       UserScenarioDisclaimer,
		Horizon:          p.Horizon,
		OutcomeEstimates: sim.Outcomes,
		FactorSummary:    sim.Factors,
		Flags:            flags,
		ParamsEcho:       p,
		SpecVersion:      spec.Version,
	}, nil
}
