// Copyright NU Cybernetics. p(DOOM) — research prototype.

package publishing

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// RequiredApprovals is the number of distinct signed reviewer approvals a
// candidate needs before promotion. The prototype requires one; heightened
// review always requires two distinct reviewers (see Promote).
const RequiredApprovals = 1

var (
	reReleaseID   = regexp.MustCompile(`^rel-\d{4}-\d{2}-\d{2}-\d{3}$`)
	reCandidateID = regexp.MustCompile(`^cand-\d{4}-\d{2}-\d{2}-\d{3}$`)
)

// BuildOptions parameterize BuildManifest.
type BuildOptions struct {
	ReleaseID         string
	CandidateID       string
	CodeCommit        string
	PreviousReleaseID *string
	SnapshotDir       string
}

// BuildManifest assembles the MODEL RELEASE ARTIFACT from a model result.
func BuildManifest(res *model.Result, snap *snapshot.Snapshot, opts BuildOptions) (schema.ReleaseManifest, error) {
	if !reReleaseID.MatchString(opts.ReleaseID) {
		return schema.ReleaseManifest{}, fmt.Errorf("release id %q must match rel-YYYY-MM-DD-NNN", opts.ReleaseID)
	}
	if !reCandidateID.MatchString(opts.CandidateID) {
		return schema.ReleaseManifest{}, fmt.Errorf("candidate id %q must match cand-YYYY-MM-DD-NNN", opts.CandidateID)
	}
	if opts.CodeCommit == "" {
		opts.CodeCommit = "uncommitted"
	}
	spec := snap.ModelSpec
	m := schema.ReleaseManifest{
		ReleaseID: opts.ReleaseID, CandidateID: opts.CandidateID, PreviousReleaseID: opts.PreviousReleaseID,
		ModelVersions: res.ModelVersions, CodeCommit: opts.CodeCommit, DataSnapshot: snap.Manifest.SnapshotID,
		SourceCutoff: snap.Manifest.SourceCutoff, GeneratedAt: res.GeneratedAt,
		OutcomeDefinition: "p(DOOM) = probability of O3 (permanent severe disempowerment), O4 (civilizational collapse), O5 (near-extinction), O6 (human extinction), O7 (biospheric catastrophe) or O8 (other irreversible loss) within the stated horizon, conditional on the model specification. The official value in this release line is index-only (insufficiently calibrated); external aggregates and the research-mode model are published separately and never blended.",
		Horizons:          horizonsOf(res),
		Priors: map[string]any{
			"experimental_causal": map[string]any{"version": spec.ExperimentalCausal.Version, "seed": spec.ExperimentalCausal.Seed, "samples": spec.ExperimentalCausal.Samples, "common_factor_loading": spec.ExperimentalCausal.CommonFactorLoading, "horizons": spec.ExperimentalCausal.Horizons, "rationale": spec.ExperimentalCausal.Rationale},
			"note":                "Every prior is a documented judgment stored in model_spec.json; none is a measurement.",
		},
		Weights:      map[string]any{"index_weights": spec.IndexWeights, "tier_multipliers": spec.TierMultipliers, "weight_bounds": spec.WeightBounds, "incident_scoring": spec.IncidentScoring, "single_figure_cap": 0.35},
		Dependencies: map[string]any{"structure": "single common latent factor Z with Gaussian copula loading λ across A, C, E, F and outcome shares", "lambda": spec.ExperimentalCausal.CommonFactorLoading, "not_modelled": []string{"feedback loops", "time-varying hazards", "explicit competing risks", "scenario-graph dependence (edges are descriptive)"}},
		Estimates:    []schema.EstimateSummary{}, Sensitivity: []schema.SensitivitySummary{},
		ExternalForecasts:  schema.ExternalForecastSummary{SourceIDs: []string{}, GroupIDs: []string{}},
		Changes:            changes(res),
		EditorialRiskLevel: res.EditorialRiskLevel, UncertaintyScore: res.UncertaintyScore, DataSummary: res.DataSummary,
		Reviewers: []string{}, Approval: nil,
		KnownLimitations: knownLimitations(res),
		ReproductionCommand: fmt.Sprintf("go run ./cmd/pdoomctl model run --snapshot data/snapshots/%s --out data/candidates/%s --release-id %s --generated-at %s --code-commit %s%s",
			snap.Manifest.SnapshotID, opts.CandidateID, opts.ReleaseID, res.GeneratedAt, opts.CodeCommit, prevFlag(opts.PreviousReleaseID)),
		Files: []schema.SnapshotFile{},
	}
	for _, e := range res.Estimates {
		var p50 *float64
		if e.Quantiles != nil {
			v := e.Quantiles.P50
			p50 = &v
		}
		m.Estimates = append(m.Estimates, schema.EstimateSummary{EstimateID: e.EstimateID, Status: e.Status, OutcomeSet: e.OutcomeSet, Horizon: e.Horizon, P50: p50, Interval: e.Quantiles, Display: e.Display.Central})
	}
	m.Intervals = "Intervals are p05–p95. External aggregates use the min–max range of member forecasts; research-mode intervals are Monte Carlo quantiles under documented parameter uncertainty; the official object publishes none."
	for _, r := range res.Sensitivity {
		if r.Rank <= 10 {
			m.Sensitivity = append(m.Sensitivity, schema.SensitivitySummary{RunID: r.RunID, Kind: r.Kind, TargetID: r.TargetID, Delta: r.Delta})
		}
	}
	sort.Slice(m.Sensitivity, func(i, j int) bool { return m.Sensitivity[i].RunID < m.Sensitivity[j].RunID })
	srcs := map[string]bool{}
	for _, g := range res.Groups {
		m.ExternalForecasts.GroupCount++
		m.ExternalForecasts.ForecastCount += len(g.Forecasts)
		m.ExternalForecasts.GroupIDs = append(m.ExternalForecasts.GroupIDs, g.ID)
		for _, s := range g.SourceIDs() {
			srcs[s] = true
		}
	}
	for s := range srcs {
		m.ExternalForecasts.SourceIDs = append(m.ExternalForecasts.SourceIDs, s)
	}
	sort.Strings(m.ExternalForecasts.SourceIDs)
	return m, nil
}

func prevFlag(p *string) string {
	if p == nil {
		return ""
	}
	return " --previous " + *p
}

func horizonsOf(res *model.Result) []string {
	set := map[string]bool{}
	for _, e := range res.Estimates {
		set[e.Horizon] = true
	}
	var out []string
	for _, h := range schema.Horizons {
		if set[string(h)] {
			out = append(out, string(h))
		}
	}
	return out
}

func changes(res *model.Result) []string {
	if res.Delta.PreviousReleaseID == nil {
		return []string{"First release of this release line.", fmt.Sprintf("Snapshot %s; %d sources, %d forecasts, %d incidents.", res.SnapshotID, res.DataSummary.SourceCount, res.DataSummary.ForecastCount, res.DataSummary.IncidentCount)}
	}
	var out []string
	for _, ch := range res.Delta.EstimateChanges {
		if ch.ChangePoints != nil && *ch.ChangePoints != 0 {
			out = append(out, fmt.Sprintf("%s: %s → %s (%+.1f points)", ch.EstimateID, ch.PreviousDisplay, ch.NewDisplay, *ch.ChangePoints))
		}
	}
	for _, ch := range res.Delta.IndexChanges {
		if ch.Delta != nil && *ch.Delta != 0 {
			out = append(out, fmt.Sprintf("index %s: %+.1f", ch.IndexID, *ch.Delta))
		}
	}
	if len(out) == 0 {
		out = append(out, "No estimate or index moved relative to the previous release.")
	}
	return out
}

func knownLimitations(res *model.Result) []string {
	return []string{
		"No official probability is published: the official object has status insufficiently_calibrated because no documented calibration process exists for long-horizon existential outcomes.",
		"The research-mode model's parameters are judgments informed by cited sources, not measurements; its intervals express parameter uncertainty under the model, not calibrated forecast error.",
		"External forecast aggregates inherit the selection, framing and population effects of their member surveys; groups are small and dominated by a few studies.",
		"Indexes (0–100) are constructed scores and are never probabilities; coverage below 1 means specified signals lacked eligible observations.",
		"Dependence between factors is represented by a single common latent factor; feedback loops, competing risks and time-varying hazards are not modelled.",
		"Incident pressure counts only verified, registry-referenced incidents and therefore under-represents unreported or unverified events.",
		fmt.Sprintf("Editorial risk level (%s) is a rule-based, human-readable classification, not a measurement.", res.EditorialRiskLevel),
	}
}

// ManifestHash returns the SHA-256 (hex) of the canonical manifest with the
// mutable publication fields cleared. Reviewers sign this value.
func ManifestHash(m schema.ReleaseManifest) (string, error) {
	m.Signature = ""
	m.Approval = nil
	m.Reviewers = nil
	m.Published = nil
	m.Superseded = nil
	raw, err := schema.CanonicalJSON(m)
	if err != nil {
		return "", err
	}
	return schema.SHA256Hex(raw), nil
}

// NextReleaseID returns rel-<date>-NNN where NNN is one more than the highest
// existing release or candidate number for that date.
func NextReleaseID(date string, existing []string) string {
	max := 0
	prefix := "rel-" + date + "-"
	for _, id := range existing {
		if strings.HasPrefix(id, prefix) {
			var n int
			fmt.Sscanf(strings.TrimPrefix(id, prefix), "%d", &n)
			if n > max {
				max = n
			}
		}
	}
	return fmt.Sprintf("%s%03d", prefix, max+1)
}
