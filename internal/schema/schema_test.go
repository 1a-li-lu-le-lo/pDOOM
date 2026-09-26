// Copyright NU Cybernetics. p(DOOM) — research prototype.

package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEnumsValid(t *testing.T) {
	cases := []struct {
		name  string
		valid bool
	}{
		{"outcome O6", Outcome("O6").Valid()},
		{"outcome O9 invalid", !Outcome("O9").Valid()},
		{"horizon 2100", Horizon("2100").Valid()},
		{"horizon custom invalid", !Horizon("custom").Valid()},
		{"tier 3", SourceTier(3).Valid()},
		{"tier 6 invalid", !SourceTier(6).Valid()},
		{"source type paper", SourceType("paper").Valid()},
		{"model use eligible feeds", ModelUseStatus("eligible").FeedsModel()},
		{"model use informational does not feed", !ModelUseStatus("informational").FeedsModel()},
		{"verified_fetch allows", VerifiedFetch.AllowsModelUse()},
		{"prior knowledge does not allow", !VerifiedPriorKnowledge.AllowsModelUse()},
		{"population", ForecastPopulation("superforecasters").Valid()},
		{"cause", IncidentCause("security_compromise").Valid()},
		{"harm", Harm("infrastructure").Valid()},
		{"severity", Severity("catastrophic").Valid()},
		{"relevance", Relevance("direct_precursor").Valid()},
		{"evidence level", EvidenceLevel("official_finding").Valid()},
		{"driver family D10", DriverFamily("D10").Valid()},
		{"driver family D11 invalid", !DriverFamily("D11").Valid()},
		{"uncertainty", UncertaintyLabel("extreme").Valid()},
		{"editorial", EditorialRiskLevel("severe_uncertainty").Valid()},
		{"estimate status", EstimateStatus("insufficiently_calibrated").Valid()},
		{"index id", IndexID("agentic_infrastructure_risk").Valid()},
		{"method", AggregationMethod("equal_weight_by_population").Valid()},
		{"kind", Kind("driver_observation").Valid()},
		{"kind invalid", !Kind("drivers").Valid()},
		{"conflict", ConflictLabel("none_known").Valid()},
		{"claim status", ClaimStatus("corroborated").Valid()},
		{"claim evidence", ClaimEvidenceType("survey_result").Valid()},
		{"consensus", Consensus("emerging").Valid()},
		{"robots", RobotsStatus("not_applicable").Valid()},
		{"retraction", RetractionStatus("disputed").Valid()},
		{"contamination", ContaminationRisk("unknown").Valid()},
		{"saturation", Saturation("partial").Valid()},
		{"direction", BenchmarkDirection("lower_is_more_capable").Valid()},
		{"confidence", Confidence("moderate").Valid()},
		{"date precision", DatePrecision("month").Valid()},
		{"novelty", Novelty("novel").Valid()},
		{"probability source", ProbabilitySource("not_assigned").Valid()},
		{"recoverability", Recoverability("none").Valid()},
		{"edge relation", EdgeRelation("prevents_response").Valid()},
		{"signal direction", SignalDirection("higher_strengthens_control").Valid()},
		{"evidence strength", EvidenceStrength("weak").Valid()},
		{"cost", Cost("very_high").Valid()},
		{"time to deploy", TimeToDeploy("5y+").Valid()},
		{"effect size", EffectSize("large").Valid()},
		{"intervention category", InterventionCategory("resilience").Valid()},
		{"audience", Audience("auditors_red_teams").Valid()},
		{"effort", Effort("high").Valid()},
	}
	for _, c := range cases {
		if !c.valid {
			t.Errorf("%s: unexpected validity result", c.name)
		}
	}
	if len(Outcomes) != 9 || len(Horizons) != 7 || len(DriverFamilies) != 10 || len(AggregationMethods) != 7 || len(IndexIDs) != 7 {
		t.Fatalf("enum tables incomplete")
	}
	if HorizonIndex("10y") != 3 || HorizonIndex("nope") != -1 {
		t.Fatalf("HorizonIndex wrong")
	}
}

func TestOutcomeSetKeyAndLabel(t *testing.T) {
	if OutcomeSetKey(PDoomOutcomes) != "O3+O4+O5+O6+O7+O8" {
		t.Fatalf("key: %s", OutcomeSetKey(PDoomOutcomes))
	}
	if !strings.Contains(OutcomeSetLabel(PDoomOutcomes), "p(DOOM)") {
		t.Fatalf("label: %s", OutcomeSetLabel(PDoomOutcomes))
	}
	if OutcomeSetLabel([]Outcome{O6}) != "Human extinction" {
		t.Fatalf("label: %s", OutcomeSetLabel([]Outcome{O6}))
	}
}

func TestEnvelopeRoundTrip(t *testing.T) {
	raw := `{"kind":"driver","schema_version":1,"items":[{"id":"D1","name":"capability","description":"d","signals":[]}]}`
	var env Envelope[Driver]
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatal(err)
	}
	if env.Kind != "driver" || len(env.Items) != 1 || env.Items[0].ID != "D1" {
		t.Fatalf("bad decode: %+v", env)
	}
}

func TestCanonicalJSON(t *testing.T) {
	type inner struct {
		Z int     `json:"z"`
		A string  `json:"a"`
		F float64 `json:"f"`
	}
	type outer struct {
		B inner          `json:"b"`
		A []inner        `json:"a"`
		M map[string]any `json:"m"`
		N *int           `json:"n"`
	}
	v := outer{B: inner{Z: 1, A: "x<y&z", F: 0.5}, A: []inner{{Z: 2, A: "q", F: 1}}, M: map[string]any{"k2": 2, "k1": "v"}}
	got, err := CanonicalJSON(v)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":[{"a":"q","f":1,"z":2}],"b":{"a":"x<y&z","f":0.5,"z":1},"m":{"k1":"v","k2":2},"n":null}`
	if string(got) != want {
		t.Fatalf("canonical mismatch:\n got %s\nwant %s", got, want)
	}
	// Two structurally equal values with differently ordered maps hash identically.
	v2 := v
	v2.M = map[string]any{"k1": "v", "k2": 2}
	got2, _ := CanonicalJSON(v2)
	if SHA256Hex(got) != SHA256Hex(got2) {
		t.Fatalf("hash differs for equal values")
	}
	ind, err := CanonicalJSONIndent(v)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(ind), "\n") {
		if strings.HasSuffix(line, " ") {
			t.Fatalf("trailing space in %q", line)
		}
	}
	if !strings.HasSuffix(string(ind), "}\n") {
		t.Fatalf("indented output must end with newline")
	}
	var back outer
	if err := json.Unmarshal(ind, &back); err != nil {
		t.Fatalf("indented output not valid JSON: %v", err)
	}
}

func TestForecastCentralValue(t *testing.T) {
	m := 0.2
	f := Forecast{Median: &m}
	if v, ok := f.CentralValue(); !ok || v != 0.2 {
		t.Fatal("median expected")
	}
	f = Forecast{Mean: &m}
	if v, ok := f.CentralValue(); !ok || v != 0.2 {
		t.Fatal("mean expected")
	}
	f = Forecast{Quantiles: &Quantiles{P50: &m}}
	if v, ok := f.CentralValue(); !ok || v != 0.2 {
		t.Fatal("p50 expected")
	}
	if _, ok := (Forecast{}).CentralValue(); ok {
		t.Fatal("no central value expected")
	}
}

func TestTierMultiplier(t *testing.T) {
	m := ModelSpec{TierMultipliers: map[string]float64{"1": 1, "2": 0.9, "3": 0.5, "4": 0, "5": 0}}
	if m.TierMultiplier(2) != 0.9 || m.TierMultiplier(5) != 0 || m.TierMultiplier(9) != 0 {
		t.Fatal("tier multiplier wrong")
	}
}
