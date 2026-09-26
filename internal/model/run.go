// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// OfficialModelVersion labels the official object, which publishes no
// probability in this release line (Model G, index-only).
const OfficialModelVersion = "pdoom-model/official-index-only@0.1.0"

// RunOptions parameterize a model run. GeneratedAt must be supplied by the
// caller (the model never reads the clock).
type RunOptions struct {
	PreviousRelease *schema.Release
	GeneratedAt     string // RFC 3339 UTC
	CodeCommit      string
}

// Result is everything a candidate release is made of.
type Result struct {
	SnapshotID         string
	GeneratedAt        string
	ModelVersions      []string
	Estimates          []schema.Estimate
	Indexes            []schema.IndexValue
	Aggregations       []schema.AggregationResult
	Sensitivity        []schema.SensitivityRun
	DriversExplained   schema.DriversExplained
	Delta              schema.DeltaRecord
	EditorialRiskLevel schema.EditorialRiskLevel
	UncertaintyScore   *float64
	UncertaintyLabel   schema.UncertaintyLabel
	DataSummary        schema.DataSummary
	Warnings           []string
	Groups             []Group
	Research           map[string]SimResult
}

// Run executes the full deterministic pipeline on a loaded snapshot.
func Run(s *snapshot.Snapshot, opts RunOptions) (*Result, error) {
	if s == nil {
		return nil, fmt.Errorf("model: nil snapshot")
	}
	if opts.GeneratedAt == "" {
		return nil, fmt.Errorf("model: RunOptions.GeneratedAt is required (the model never reads the clock)")
	}
	asOf := opts.GeneratedAt
	if len(asOf) >= 10 {
		asOf = asOf[:10]
	}
	spec := s.ModelSpec
	res := &Result{SnapshotID: s.Manifest.SnapshotID, GeneratedAt: opts.GeneratedAt,
		ModelVersions: []string{OfficialModelVersion, AggregateModelVersion, spec.ExperimentalCausal.Version, IndexModelVersion}}
	if spec.ExperimentalCausal.Version == "" {
		res.ModelVersions[2] = CausalModelVersion
	}

	// 1. External forecast aggregation.
	res.Groups = Aggregate(s)
	for _, g := range res.Groups {
		res.Aggregations = append(res.Aggregations, g.Results...)
	}
	if res.Aggregations == nil {
		res.Aggregations = []schema.AggregationResult{}
	}

	// 2. Research-mode experimental model per horizon.
	res.Research = map[string]SimResult{}
	horizons := causalHorizons(spec.ExperimentalCausal)
	for _, h := range horizons {
		res.Research[h] = Simulate(spec.ExperimentalCausal.Horizons[h], spec.ExperimentalCausal.CommonFactorLoading, spec.ExperimentalCausal.Seed, spec.ExperimentalCausal.Samples, nil)
	}

	// 3. Sensitivity.
	res.Sensitivity = Sensitivity(s, res.Groups, res.Research)
	if res.Sensitivity == nil {
		res.Sensitivity = []schema.SensitivityRun{}
	}

	// 4. Indexes.
	cpi := signalIndex(s, schema.IndexCapabilityPressure, spec.IndexWeights.CapabilityPressure, asOf, "")
	csi := signalIndex(s, schema.IndexControlStrength, spec.IndexWeights.ControlStrength, asOf, "")
	air := signalIndex(s, schema.IndexAgenticInfrastructureRisk, spec.IndexWeights.AgenticInfrastructureRisk, asOf, "")
	ipi := incidentIndex(s, asOf, "")
	current := map[schema.IndexID]*float64{schema.IndexCapabilityPressure: cpi.Value, schema.IndexControlStrength: csi.Value, schema.IndexIncidentPressure: ipi.Value}
	epi := evidenceIndex(current, opts.PreviousRelease, asOf)
	// Propagate baseline pointers for CPI/CSI/IPI so later releases compare to the first release.
	for _, iv := range []*schema.IndexValue{&cpi, &csi, &ipi} {
		if opts.PreviousRelease != nil {
			if piv := opts.PreviousRelease.IndexByID(iv.IndexID); piv != nil {
				if piv.Baseline != nil {
					iv.Baseline = piv.Baseline
				} else {
					iv.Baseline = &schema.IndexBaseline{SnapshotID: opts.PreviousRelease.Manifest.DataSnapshot, ReleaseID: opts.PreviousRelease.Manifest.ReleaseID, Value: piv.Value}
				}
			}
		}
	}
	uin := UncertaintyInputs{UnknownDependencies: 0.5}
	cov := []float64{cpi.Coverage, csi.Coverage, air.Coverage}
	uin.CoverageGap = 1 - Mean(cov)
	if len(res.Groups) == 0 {
		uin.ForecastDisagreement = 1
	} else {
		var ds []float64
		for _, g := range res.Groups {
			ds = append(ds, clamp01(g.IQRLogOdds/4))
		}
		uin.ForecastDisagreement = Mean(ds)
	}
	uin.LowTierShare = lowTierShare([]schema.IndexValue{cpi, csi, air})
	maxDelta := 0.0
	for _, r := range res.Sensitivity {
		if r.TargetKind == "estimate" && abs(r.Delta) > maxDelta {
			maxDelta = abs(r.Delta)
		}
	}
	uin.SensitivitySpread = clamp01(maxDelta / 0.2)
	unc := uncertaintyIndex(spec, uin, asOf)
	res.Indexes = []schema.IndexValue{epi, cpi, csi, ipi, unc, air, attentionIndex(asOf)}
	res.UncertaintyScore = unc.Value
	res.UncertaintyLabel = schema.UncertaintyExtreme
	if unc.Value != nil {
		res.UncertaintyLabel = UncertaintyLabelFromScore(*unc.Value)
	}

	// 5. Editorial level.
	vars := map[string]float64{"coverage": Mean(cov)}
	for k, v := range map[string]*float64{"uncertainty": unc.Value, "cpi": cpi.Value, "csi": csi.Value, "ipi": ipi.Value, "epi": epi.Value, "air": air.Value} {
		if v != nil {
			vars[k] = *v
		}
	}
	var warns []string
	res.EditorialRiskLevel, warns = editorialLevel(spec, vars)
	res.Warnings = append(res.Warnings, warns...)

	// 6. Estimates.
	res.Estimates = buildEstimates(s, res, asOf)

	// 7. Explanations, data summary, delta.
	res.DriversExplained = Explain(s, res.Indexes, asOf)
	res.DataSummary = dataSummary(s)
	res.Delta = Delta(s, res, opts.PreviousRelease, opts.GeneratedAt)
	if res.Warnings == nil {
		res.Warnings = []string{}
	}
	return res, nil
}

func causalHorizons(ec schema.ExperimentalCausalSpec) []string {
	hs := make([]string, 0, len(ec.Horizons))
	for h := range ec.Horizons {
		hs = append(hs, h)
	}
	sort.Slice(hs, func(i, j int) bool {
		return schema.HorizonIndex(schema.Horizon(hs[i])) < schema.HorizonIndex(schema.Horizon(hs[j]))
	})
	return hs
}

func buildEstimates(s *snapshot.Snapshot, res *Result, asOf string) []schema.Estimate {
	spec := s.ModelSpec
	step := RoundingStep(spec.RoundingRules, res.UncertaintyLabel)
	cutoff := s.Manifest.SourceCutoff
	var out []schema.Estimate
	prev := func(id string) *schema.PreviousEstimate { return nil }
	// The official object: one per horizon, no probability.
	for _, h := range schema.Horizons {
		out = append(out, schema.Estimate{
			EstimateID: "est-official-P_DOOM-" + string(h), Producer: OfficialModelVersion, Status: schema.StatusInsufficientlyCalibrate,
			OutcomeSet: schema.PDoomOutcomes, OutcomeLabel: schema.OutcomeSetLabel(schema.PDoomOutcomes), Horizon: string(h),
			Conditioning:       "Official value withheld: no documented calibration process supports converting indexes or the research-mode model into a published probability (build-spec §0.2, ADR-002).",
			ForecastOriginDate: asOf, LastEvidenceDate: cutoff, Quantiles: nil, Mean: nil, Disagreement: nil, Uncertainty: res.UncertaintyLabel,
			ModelConfidence: "not_applicable", SourceCoverage: schema.SourceCoverage{SourceIDs: []string{}}, Previous: prev(""), ReasonForChange: "index-only release line",
			RoundingRule: "none", Display: schema.Display{Central: "Insufficiently calibrated", Interval: "", Note: "The observatory publishes indexes, an external forecast aggregate and a research-mode model; it does not publish an official probability yet."},
			Assumptions: []string{"Model G (index-only) is the official producer in this release line."}, MethodRef: "docs/method/model.md#official-value",
		})
	}
	// External aggregates.
	for _, g := range res.Groups {
		vals := append([]float64(nil), g.Values...)
		sort.Float64s(vals)
		q := &schema.FiveQuantiles{P05: round6(vals[0]), P25: round6(Quantile(append([]float64(nil), vals...), 0.25)), P50: round6(g.Preferred), P75: round6(Quantile(append([]float64(nil), vals...), 0.75)), P95: round6(vals[len(vals)-1])}
		fixOrder(q)
		mean := round6(Mean(vals))
		dis := DisagreementLabel(g.IQRLogOdds)
		label := schema.OutcomeSetLabel(g.OutcomeSet)
		gid := g.ID
		var alt []string
		for _, r := range g.Results {
			if !r.Preferred {
				alt = append(alt, fmt.Sprintf("%s %s", r.Method, RoundForDisplay(r.Value, step)))
			}
		}
		out = append(out, schema.Estimate{
			EstimateID: "est-external-" + g.ID, Producer: AggregateModelVersion, Status: schema.StatusExternalAggregate,
			OutcomeSet: g.OutcomeSet, OutcomeLabel: label, Horizon: g.Horizon, Conditioning: g.Conditioning,
			ForecastOriginDate: asOf, LastEvidenceDate: cutoff, Quantiles: q, Mean: &mean, Disagreement: &dis, Uncertainty: res.UncertaintyLabel,
			ModelConfidence: "aggregation_of_named_sources", SourceCoverage: schema.SourceCoverage{ForecastCount: len(g.Forecasts), PopulationCount: g.PopulationCount(), SourceIDs: g.SourceIDs()},
			Previous: prev(""), ReasonForChange: "first computation for this snapshot", RoundingRule: RoundingRuleName(step),
			Display:     schema.Display{Central: RoundForDisplay(g.Preferred, step), Interval: FormatInterval(q.P05, q.P95, step), Note: "Interval is the range of member forecasts (min–max), not a sampling interval. Alternatives: " + strings.Join(alt, "; ")},
			Assumptions: []string{"Members share outcome set, horizon and conditioning (compatibility group " + g.ID + ").", "Central value is the unweighted median; alternative methods are published alongside.", "Original question wording is shown beside every transformed value."},
			MethodRef:   "docs/method/aggregation.md", GroupID: &gid,
		})
	}
	// Research mode.
	for _, h := range causalHorizons(spec.ExperimentalCausal) {
		sim := res.Research[h]
		for _, t := range researchTargets {
			oe, ok := sim.Outcomes[t.Key]
			if !ok {
				continue
			}
			q := &schema.FiveQuantiles{P05: oe.P05, P25: oe.P25, P50: oe.P50, P75: oe.P75, P95: oe.P95}
			mean := oe.Mean
			dis := schema.UncertaintyExtreme
			out = append(out, schema.Estimate{
				EstimateID: researchEstimateID(t.Key, h), Producer: res.ModelVersions[2], Status: schema.StatusResearchMode,
				OutcomeSet: t.OutcomeSet, OutcomeLabel: schema.OutcomeSetLabel(t.OutcomeSet), Horizon: h,
				Conditioning:       "Conditional on the experimental causal decomposition P(A)·P(C|A)·P(E|A,C)·P(F|A,C,E)·P(O_i|A,C,E,F) with judgment-based parameters (model_spec.experimental_causal) and a single common latent factor.",
				ForecastOriginDate: asOf, LastEvidenceDate: cutoff, Quantiles: q, Mean: &mean, Disagreement: &dis, Uncertainty: res.UncertaintyLabel,
				ModelConfidence: "research_mode_low", SourceCoverage: schema.SourceCoverage{SourceIDs: spec.ExperimentalCausal.SourceIDs},
				Previous: prev(""), ReasonForChange: "first computation for this snapshot", RoundingRule: RoundingRuleName(step),
				Display:     schema.Display{Central: RoundForDisplay(oe.P50, step), Interval: FormatInterval(oe.P05, oe.P95, step), Note: "Research mode: not the official estimate. Parameters are documented judgments; the interval reflects parameter uncertainty under the model, not calibration."},
				Assumptions: []string{"Factor distributions are logit-normal fits to documented {p05, p50, p95} judgments.", fmt.Sprintf("Common factor loading λ = %.2f induces positive dependence between factors.", spec.ExperimentalCausal.CommonFactorLoading), fmt.Sprintf("Monte Carlo with seed %d and %d samples; quantiles are nearest-rank.", spec.ExperimentalCausal.Seed, spec.ExperimentalCausal.Samples)},
				MethodRef:   "docs/method/causal-model.md",
			})
		}
	}
	for i := range out {
		if out[i].SourceCoverage.SourceIDs == nil {
			out[i].SourceCoverage.SourceIDs = []string{}
		}
	}
	return out
}

func fixOrder(q *schema.FiveQuantiles) {
	vals := []float64{q.P05, q.P25, q.P50, q.P75, q.P95}
	sort.Float64s(vals)
	q.P05, q.P25, q.P50, q.P75, q.P95 = vals[0], vals[1], vals[2], vals[3], vals[4]
}

func dataSummary(s *snapshot.Snapshot) schema.DataSummary {
	ds := schema.DataSummary{ScenarioIDs: []string{}, ForecastSourceIDs: []string{}, BenchmarkDefinitions: map[string]string{},
		ForecastCount: len(s.Forecasts), SourceCount: len(s.Sources), IncidentCount: len(s.Incidents)}
	for _, sc := range s.Scenarios {
		ds.ScenarioIDs = append(ds.ScenarioIDs, sc.ID)
	}
	sort.Strings(ds.ScenarioIDs)
	fs := map[string]bool{}
	for _, f := range s.Forecasts {
		if f.ModelUseStatus.FeedsModel() {
			fs[f.SourceID] = true
		}
	}
	ds.ForecastSourceIDs = sortedKeys(fs)
	for _, b := range s.Benchmarks {
		ds.BenchmarkDefinitions[b.ID] = benchmarkDefinition(b)
	}
	return ds
}
