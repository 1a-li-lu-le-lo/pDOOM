// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// IndexModelVersion is the version string of the index computations.
const IndexModelVersion = "pdoom-model/indexes@0.1.0"

// indexMeta holds the label, scale and semantic of every published index.
var indexMeta = map[schema.IndexID]struct {
	label, scale, method string
	pressure             bool
}{
	schema.IndexCapabilityPressure:        {"Capability Pressure Index", "0–100, higher = more capability pressure", "docs/method/indexes.md#capability-pressure", true},
	schema.IndexControlStrength:           {"Control Strength Index", "0–100, higher = stronger safeguards and institutions", "docs/method/indexes.md#control-strength", false},
	schema.IndexAgenticInfrastructureRisk: {"Agentic Infrastructure Risk Index", "0–100, higher = more exposure from agent infrastructure", "docs/method/indexes.md#agentic-infrastructure-risk", true},
	schema.IndexIncidentPressure:          {"Incident Pressure Index", "0–100, higher = more verified incident pressure", "docs/method/indexes.md#incident-pressure", true},
	schema.IndexEvidencePressure:          {"Evidence Pressure Index", "0–100, 50 = no movement since the baseline release", "docs/method/indexes.md#evidence-pressure", true},
	schema.IndexUncertainty:               {"Uncertainty Score", "0–100, higher = less is known", "docs/method/indexes.md#uncertainty-score", true},
	schema.IndexAttention:                 {"Attention Index", "0–100, media volume per underlying event", "docs/method/indexes.md#attention-index", true},
}

// observationFor picks the observation used for a signal: the most recent
// as_of among eligible/used, verified observations; ties break on id.
func observationFor(s *snapshot.Snapshot, signalID string) (schema.DriverObservation, bool) {
	var best schema.DriverObservation
	found := false
	for _, o := range s.DriverObservations {
		if o.SignalID != signalID || !o.ModelUseStatus.FeedsModel() || !o.Verification.Status.AllowsModelUse() {
			continue
		}
		if !found || o.AsOf > best.AsOf || (o.AsOf == best.AsOf && o.ID < best.ID) {
			best, found = o, true
		}
	}
	return best, found
}

// signalIndex computes one signal-weighted index. Signals whose direction is
// opposite to the index semantic are inverted (v → 1 − v) so that every index
// reads in one direction. excluded lets sensitivity analysis drop one signal.
func signalIndex(s *snapshot.Snapshot, id schema.IndexID, weights map[string]float64, asOf string, excluded string) schema.IndexValue {
	meta := indexMeta[id]
	signals := s.SignalIndex()
	iv := schema.IndexValue{IndexID: id, Label: meta.label, Scale: meta.scale, IsProbability: false, AsOf: asOf, MethodRef: meta.method, Components: []schema.IndexComponent{}}
	num, den, wObs, wAll := 0.0, 0.0, 0.0, 0.0
	for _, sig := range sortedKeys(weights) {
		w := weights[sig]
		wAll += w
		if sig == excluded {
			continue
		}
		o, ok := observationFor(s, sig)
		if !ok {
			continue
		}
		v := o.ValueNormalized
		if ref, ok := signals[sig]; ok {
			if (ref.Signal.Direction == schema.HigherStrengthensControl) == meta.pressure {
				v = 1 - v
			}
		}
		tier := s.BestTier(o.SourceIDs)
		m := s.ModelSpec.TierMultiplier(tier)
		if m == 0 {
			continue
		}
		wObs += w
		num += w * m * v
		den += w * m
		iv.Components = append(iv.Components, schema.IndexComponent{SignalID: sig, Weight: w, ValueNormalized: round6(v), Tier: int(tier), Contribution: 0})
	}
	if den == 0 || wAll == 0 {
		iv.Note = "No eligible observations; the index cannot be computed from this snapshot."
		iv.Coverage = 0
		return iv
	}
	value := round4(100 * num / den)
	iv.Value = &value
	iv.Coverage = round4(wObs / wAll)
	// Contribution in index points: the signal's share of the weighted sum.
	for i := range iv.Components {
		c := &iv.Components[i]
		m := s.ModelSpec.TierMultiplier(schema.SourceTier(c.Tier))
		c.Contribution = round4(100 * c.Weight * m * c.ValueNormalized / den)
	}
	if iv.Coverage < 1 {
		iv.Note = fmt.Sprintf("Computed over %.0f%% of the specified signal weight; missing signals are listed in the coverage gaps.", iv.Coverage*100)
	}
	return iv
}

// incidentIndex computes the incident pressure index:
// x = Σ severity_w · relevance_w · evidence_w · 0.5^(age_days / half_life);
// index = 100 · (1 − exp(−x / k)).
func incidentIndex(s *snapshot.Snapshot, asOf string, excluded string) schema.IndexValue {
	meta := indexMeta[schema.IndexIncidentPressure]
	iv := schema.IndexValue{IndexID: schema.IndexIncidentPressure, Label: meta.label, Scale: meta.scale, AsOf: asOf, MethodRef: meta.method, Components: []schema.IndexComponent{}}
	sc := s.ModelSpec.IncidentScoring
	cutoff, _ := time.Parse("2006-01-02", s.Manifest.SourceCutoff)
	x := 0.0
	n := 0
	incidents := append([]schema.Incident(nil), s.Incidents...)
	sort.Slice(incidents, func(i, j int) bool { return incidents[i].ID < incidents[j].ID })
	for _, in := range incidents {
		if !in.ModelUseStatus.FeedsModel() || !in.Verification.Status.AllowsModelUse() || in.ID == excluded {
			continue
		}
		n++
		recency := 0.5
		if in.Date != nil {
			if d, err := time.Parse("2006-01-02", *in.Date); err == nil && sc.RecencyHalfLifeDays > 0 {
				age := cutoff.Sub(d).Hours() / 24
				if age < 0 {
					age = 0
				}
				recency = math.Pow(0.5, age/float64(sc.RecencyHalfLifeDays))
			}
		}
		term := sc.SeverityWeights[string(in.Severity)] * sc.RelevanceWeights[string(in.PDoomRelevance)] * sc.EvidenceWeights[string(in.EvidenceLevel)] * recency
		x += term
		iv.Components = append(iv.Components, schema.IndexComponent{SignalID: in.ID, Weight: round4(term), ValueNormalized: round4(sc.SeverityWeights[string(in.Severity)] * sc.RelevanceWeights[string(in.PDoomRelevance)]), Tier: int(s.BestTier(in.SourceIDs)), Contribution: 0})
	}
	if n == 0 {
		iv.Note = "No eligible incidents in this snapshot."
		return iv
	}
	k := sc.SquashK
	if k <= 0 {
		k = 3
	}
	value := round4(100 * (1 - math.Exp(-x/k)))
	iv.Value = &value
	iv.Coverage = 1
	for i := range iv.Components {
		if x > 0 {
			iv.Components[i].Contribution = round4(value * iv.Components[i].Weight / x)
		}
	}
	iv.Note = fmt.Sprintf("%d eligible incidents; raw pressure %.3f squashed with k = %.1f. Counts one underlying event once; media volume is the separate attention index.", n, x, k)
	return iv
}

// evidenceIndex computes the evidence pressure index relative to the baseline
// release. A first release is its own baseline and reads 50.
func evidenceIndex(current map[schema.IndexID]*float64, prev *schema.Release, asOf string) schema.IndexValue {
	meta := indexMeta[schema.IndexEvidencePressure]
	iv := schema.IndexValue{IndexID: schema.IndexEvidencePressure, Label: meta.label, Scale: meta.scale, AsOf: asOf, MethodRef: meta.method, Components: []schema.IndexComponent{}, Coverage: 1}
	if prev == nil {
		v := 50.0
		iv.Value = &v
		iv.Note = "Baseline release: evidence pressure is defined as 50 (no movement) for the first release; later releases show movement relative to this baseline."
		return iv
	}
	// Baseline values: the previous release's stored baselines if present, else its own values.
	base := func(id schema.IndexID) (*float64, string, string) {
		pv := prev.IndexByID(id)
		if pv == nil {
			return nil, "", ""
		}
		if pv.Baseline != nil && pv.Baseline.Value != nil {
			return pv.Baseline.Value, pv.Baseline.SnapshotID, pv.Baseline.ReleaseID
		}
		return pv.Value, prev.Manifest.DataSnapshot, prev.Manifest.ReleaseID
	}
	delta := 0.0
	complete := true
	var baseSnap, baseRel string
	for _, pair := range []struct {
		id   schema.IndexID
		sign float64
	}{{schema.IndexCapabilityPressure, 1}, {schema.IndexControlStrength, -1}, {schema.IndexIncidentPressure, 1}} {
		b, bs, br := base(pair.id)
		c := current[pair.id]
		if b == nil || c == nil {
			complete = false
			continue
		}
		baseSnap, baseRel = bs, br
		d := *c - *b
		delta += pair.sign * d
		iv.Components = append(iv.Components, schema.IndexComponent{SignalID: string(pair.id), Weight: pair.sign, ValueNormalized: round4(*c / 100), Tier: 1, Contribution: round4(pair.sign * d)})
	}
	if !complete {
		iv.Note = "One or more component indexes are unavailable in the current or baseline release; evidence pressure is not computed."
		return iv
	}
	v := round4(50 + 50*math.Tanh(delta/50))
	iv.Value = &v
	iv.Baseline = &schema.IndexBaseline{SnapshotID: baseSnap, ReleaseID: baseRel}
	iv.Note = fmt.Sprintf("50 + 50·tanh((ΔCPI − ΔCSI + ΔIPI)/50) with Δ measured against release %s.", baseRel)
	return iv
}

// UncertaintyInputs are the components of the uncertainty score.
type UncertaintyInputs struct {
	CoverageGap          float64 // 1 − mean coverage of the signal indexes
	ForecastDisagreement float64 // mean over groups of min(IQR_logodds / 4, 1); 1 when no groups
	LowTierShare         float64 // share of tier-3 effective weight across signal indexes
	SensitivitySpread    float64 // min(max |Δp50| / 0.2, 1) over research-mode sensitivity runs
	UnknownDependencies  float64 // constant 0.5 in v0
}

// defaultUncertaintyWeights is used when the model spec does not provide weights.
var defaultUncertaintyWeights = map[string]float64{"coverage_gap": 0.3, "forecast_disagreement": 0.2, "low_tier_share": 0.15, "sensitivity_spread": 0.2, "unknown_dependencies": 0.15}

// uncertaintyIndex computes the 0–100 uncertainty score as a weighted mean of
// its components.
func uncertaintyIndex(spec schema.ModelSpec, in UncertaintyInputs, asOf string) schema.IndexValue {
	meta := indexMeta[schema.IndexUncertainty]
	w := spec.IndexWeights.Uncertainty
	if len(w) == 0 {
		w = defaultUncertaintyWeights
	}
	vals := map[string]float64{"coverage_gap": in.CoverageGap, "forecast_disagreement": in.ForecastDisagreement, "low_tier_share": in.LowTierShare, "sensitivity_spread": in.SensitivitySpread, "unknown_dependencies": in.UnknownDependencies}
	iv := schema.IndexValue{IndexID: schema.IndexUncertainty, Label: meta.label, Scale: meta.scale, AsOf: asOf, MethodRef: meta.method, Components: []schema.IndexComponent{}, Coverage: 1}
	num, den := 0.0, 0.0
	for _, k := range sortedKeys(w) {
		v, ok := vals[k]
		if !ok {
			continue
		}
		v = clamp01(v)
		num += w[k] * v
		den += w[k]
		iv.Components = append(iv.Components, schema.IndexComponent{SignalID: k, Weight: w[k], ValueNormalized: round4(v), Tier: 1, Contribution: 0})
	}
	if den == 0 {
		iv.Note = "No uncertainty weights."
		return iv
	}
	value := round4(100 * num / den)
	iv.Value = &value
	for i := range iv.Components {
		iv.Components[i].Contribution = round4(100 * iv.Components[i].Weight * iv.Components[i].ValueNormalized / den)
	}
	iv.Note = "Weighted mean of coverage gap, external-forecast disagreement, low-tier evidence share, sensitivity spread and a constant unknown-dependency term. Not a probability."
	return iv
}

func attentionIndex(asOf string) schema.IndexValue {
	meta := indexMeta[schema.IndexAttention]
	return schema.IndexValue{IndexID: schema.IndexAttention, Label: meta.label, Scale: meta.scale, AsOf: asOf, MethodRef: meta.method, Components: []schema.IndexComponent{}, Coverage: 0,
		Note: "Not computed: the prototype ingests no media-volume data. Article counts are never used as incident counts."}
}

// lowTierShare returns the share of effective weight carried by tier-3 sources
// across the given signal indexes.
func lowTierShare(ivs []schema.IndexValue) float64 {
	low, all := 0.0, 0.0
	for _, iv := range ivs {
		for _, c := range iv.Components {
			all += c.Weight
			if c.Tier >= 3 {
				low += c.Weight
			}
		}
	}
	if all == 0 {
		return 0
	}
	return low / all
}

// editorialLevel evaluates the model spec's ordered rules; first match wins.
func editorialLevel(spec schema.ModelSpec, vars map[string]float64) (schema.EditorialRiskLevel, []string) {
	var warnings []string
	for i, r := range spec.EditorialRules {
		rule, err := schema.ParseRule(r.When)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("editorial rule %d unparsable: %v", i, err))
			continue
		}
		ok, err := rule.Eval(vars)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("editorial rule %d (%s) skipped: %v", i, strings.TrimSpace(r.When), err))
			continue
		}
		if ok {
			return r.Level, warnings
		}
	}
	warnings = append(warnings, "no editorial rule matched; defaulting to insufficient_evidence")
	return "insufficient_evidence", warnings
}
