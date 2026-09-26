// Copyright NU Cybernetics. p(DOOM) — research prototype.

package schema

// FiveQuantiles is the full quantile set carried by estimates and Scenario Lab results.
type FiveQuantiles struct {
	P05 float64 `json:"p05"`
	P25 float64 `json:"p25"`
	P50 float64 `json:"p50"`
	P75 float64 `json:"p75"`
	P95 float64 `json:"p95"`
}

// Ordered reports whether the quantiles are non-decreasing.
func (q FiveQuantiles) Ordered() bool {
	return q.P05 <= q.P25 && q.P25 <= q.P50 && q.P50 <= q.P75 && q.P75 <= q.P95
}

// SourceCoverage summarizes what fed an estimate.
type SourceCoverage struct {
	ForecastCount   int      `json:"forecast_count"`
	PopulationCount int      `json:"population_count"`
	SourceIDs       []string `json:"source_ids"`
}

// PreviousEstimate points at the same estimate in the previous release.
type PreviousEstimate struct {
	ReleaseID string   `json:"release_id"`
	P50       *float64 `json:"p50"`
	Display   string   `json:"display"`
}

// Display holds the rounded, human-readable rendering of an estimate.
type Display struct {
	Central  string `json:"central"`
	Interval string `json:"interval"`
	Note     string `json:"note"`
}

// Estimate is a release estimate object (build-spec §3.6). Quantiles and mean
// are null for the official insufficiently_calibrated object.
type Estimate struct {
	EstimateID         string            `json:"estimate_id"`
	Producer           string            `json:"producer"`
	Status             EstimateStatus    `json:"status"`
	OutcomeSet         []Outcome         `json:"outcome_set"`
	OutcomeLabel       string            `json:"outcome_label"`
	Horizon            string            `json:"horizon"`
	HorizonNote        *string           `json:"horizon_note"`
	Conditioning       string            `json:"conditioning"`
	ForecastOriginDate string            `json:"forecast_origin_date"`
	LastEvidenceDate   string            `json:"last_evidence_date"`
	Quantiles          *FiveQuantiles    `json:"quantiles"`
	Mean               *float64          `json:"mean"`
	Disagreement       *UncertaintyLabel `json:"disagreement"`
	Uncertainty        UncertaintyLabel  `json:"uncertainty"`
	ModelConfidence    string            `json:"model_confidence"`
	SourceCoverage     SourceCoverage    `json:"source_coverage"`
	Previous           *PreviousEstimate `json:"previous"`
	ReasonForChange    string            `json:"reason_for_change"`
	RoundingRule       string            `json:"rounding_rule"`
	Display            Display           `json:"display"`
	Assumptions        []string          `json:"assumptions"`
	MethodRef          string            `json:"method_ref"`
	GroupID            *string           `json:"group_id"`
}

// IndexBaseline is the baseline an index is compared against.
type IndexBaseline struct {
	SnapshotID string   `json:"snapshot_id"`
	ReleaseID  string   `json:"release_id"`
	Value      *float64 `json:"value"`
}

// IndexComponent is one contribution to an index.
type IndexComponent struct {
	SignalID        string  `json:"signal_id"`
	Weight          float64 `json:"weight"`
	ValueNormalized float64 `json:"value_normalized"`
	Tier            int     `json:"tier"`
	Contribution    float64 `json:"contribution"`
}

// IndexValue is a release index object (build-spec §3.6). Value is null when the
// index cannot be computed from the snapshot (for example attention without
// media-volume data). Indexes are never probabilities.
type IndexValue struct {
	IndexID       IndexID          `json:"index_id"`
	Value         *float64         `json:"value"`
	Label         string           `json:"label"`
	Scale         string           `json:"scale"`
	IsProbability bool             `json:"is_probability"`
	Baseline      *IndexBaseline   `json:"baseline"`
	Components    []IndexComponent `json:"components"`
	Coverage      float64          `json:"coverage"`
	AsOf          string           `json:"as_of"`
	MethodRef     string           `json:"method_ref"`
	Note          string           `json:"note"`
}

// AggregationResult is one compatibility group × method result.
type AggregationResult struct {
	AggregationID      string             `json:"aggregation_id"`
	GroupID            string             `json:"group_id"`
	Method             AggregationMethod  `json:"method"`
	Preferred          bool               `json:"preferred"`
	OutcomeSet         []Outcome          `json:"outcome_set"`
	Horizon            string             `json:"horizon"`
	Conditioning       string             `json:"conditioning"`
	Value              float64            `json:"value"`
	N                  int                `json:"n"`
	ForecastIDs        []string           `json:"forecast_ids"`
	SourceIDs          []string           `json:"source_ids"`
	PopulationCount    int                `json:"population_count"`
	Weights            map[string]float64 `json:"weights"`
	WordingNote        string             `json:"wording_note"`
	TransformationNote string             `json:"transformation_note"`
	Note               string             `json:"note"`
}

// SensitivityRun is one perturbation of the model.
type SensitivityRun struct {
	RunID           string    `json:"run_id"`
	Kind            string    `json:"kind"`
	TargetKind      string    `json:"target_kind"`
	TargetID        string    `json:"target_id"`
	RemovedID       *string   `json:"removed_id"`
	ParameterChange string    `json:"parameter_change"`
	Horizon         *string   `json:"horizon"`
	OutcomeSet      []Outcome `json:"outcome_set"`
	BaselineValue   float64   `json:"baseline_value"`
	Value           float64   `json:"value"`
	Delta           float64   `json:"delta"`
	Rank            int       `json:"rank"`
	Note            string    `json:"note"`
}

// EstimateChange is one row of a delta record.
type EstimateChange struct {
	EstimateID      string   `json:"estimate_id"`
	Status          string   `json:"status"`
	PreviousP50     *float64 `json:"previous_p50"`
	NewP50          *float64 `json:"new_p50"`
	ChangePoints    *float64 `json:"change_points"`
	PreviousDisplay string   `json:"previous_display"`
	NewDisplay      string   `json:"new_display"`
}

// IndexChange is one index row of a delta record.
type IndexChange struct {
	IndexID  IndexID  `json:"index_id"`
	Previous *float64 `json:"previous"`
	New      *float64 `json:"new"`
	Delta    *float64 `json:"delta"`
}

// DeltaRecord is the PROBABILITY CHANGE POLICY record versus the previous release
// (null fields for a first release).
type DeltaRecord struct {
	Kind                     string           `json:"kind"`
	SchemaVersion            int              `json:"schema_version"`
	ReleaseID                *string          `json:"release_id"`
	PreviousReleaseID        *string          `json:"previous_release_id"`
	SnapshotID               string           `json:"snapshot_id"`
	PreviousSnapshotID       *string          `json:"previous_snapshot_id"`
	GeneratedAt              string           `json:"generated_at"`
	EstimateChanges          []EstimateChange `json:"estimate_changes"`
	IndexChanges             []IndexChange    `json:"index_changes"`
	HeightenedReviewTriggers []string         `json:"heightened_review_triggers"`
	NewScenarioIDs           []string         `json:"new_scenario_ids"`
	NewForecastSourceIDs     []string         `json:"new_forecast_source_ids"`
	RedefinedBenchmarkIDs    []string         `json:"redefined_benchmark_ids"`
	ModelVersionChanged      bool             `json:"model_version_changed"`
	Notes                    []string         `json:"notes"`
}

// DriverContribution is one "Why this number?" row.
type DriverContribution struct {
	IndexID         IndexID  `json:"index_id"`
	Driver          string   `json:"driver"`
	DriverFamily    string   `json:"driver_family"`
	Direction       string   `json:"direction"`
	Magnitude       float64  `json:"magnitude"`
	SourceIDs       []string `json:"source_ids"`
	Confidence      float64  `json:"confidence"`
	ModelRole       string   `json:"model_role"`
	LastUpdated     string   `json:"last_updated"`
	Sensitivity     float64  `json:"sensitivity"`
	Counterevidence *string  `json:"counterevidence"`
}

// DriversExplained is drivers_explained.json.
type DriversExplained struct {
	Kind          string               `json:"kind"`
	SchemaVersion int                  `json:"schema_version"`
	ReleaseID     *string              `json:"release_id"`
	AsOf          string               `json:"as_of"`
	Items         []DriverContribution `json:"items"`
	Notes         []string             `json:"notes"`
}

// EstimateSummary is a compact estimate row for the release manifest.
type EstimateSummary struct {
	EstimateID string         `json:"estimate_id"`
	Status     EstimateStatus `json:"status"`
	OutcomeSet []Outcome      `json:"outcome_set"`
	Horizon    string         `json:"horizon"`
	P50        *float64       `json:"p50"`
	Interval   *FiveQuantiles `json:"interval"`
	Display    string         `json:"display"`
}

// SensitivitySummary is a compact sensitivity row for the release manifest.
type SensitivitySummary struct {
	RunID    string  `json:"run_id"`
	Kind     string  `json:"kind"`
	TargetID string  `json:"target_id"`
	Delta    float64 `json:"delta"`
}

// ExternalForecastSummary summarizes the external forecast inputs.
type ExternalForecastSummary struct {
	GroupCount    int      `json:"group_count"`
	ForecastCount int      `json:"forecast_count"`
	SourceIDs     []string `json:"source_ids"`
	GroupIDs      []string `json:"group_ids"`
}

// ApprovalPolicy records the approval requirements and what was received.
type ApprovalPolicy struct {
	RequiredApprovals   int      `json:"required_approvals"`
	ReceivedApprovals   int      `json:"received_approvals"`
	HeightenedReview    bool     `json:"heightened_review"`
	HeightenedReviewAck *string  `json:"heightened_review_ack"`
	ApproverIDs         []string `json:"approver_ids"`
}

// Superseded records which release superseded this one.
type Superseded struct {
	By string `json:"by"`
	At string `json:"at"`
}

// DataSummary records ids present in the snapshot so that a later release can
// compute a delta record without loading the older snapshot.
type DataSummary struct {
	ScenarioIDs          []string          `json:"scenario_ids"`
	ForecastSourceIDs    []string          `json:"forecast_source_ids"`
	BenchmarkDefinitions map[string]string `json:"benchmark_definitions"`
	ForecastCount        int               `json:"forecast_count"`
	SourceCount          int               `json:"source_count"`
	IncidentCount        int               `json:"incident_count"`
}

// ReleaseManifest is data/releases/<id>/manifest.json (build-spec §3.6).
//
// The field `signature` holds the SHA-256 of the canonical JSON of the manifest
// with the mutable publication fields (`signature`, `approval`, `reviewers`,
// `published`, `superseded`) cleared. Reviewer approvals sign that hash.
type ReleaseManifest struct {
	ReleaseID           string                  `json:"release_id"`
	CandidateID         string                  `json:"candidate_id"`
	PreviousReleaseID   *string                 `json:"previous_release_id"`
	ModelVersions       []string                `json:"model_versions"`
	CodeCommit          string                  `json:"code_commit"`
	DataSnapshot        string                  `json:"data_snapshot"`
	SourceCutoff        string                  `json:"source_cutoff"`
	GeneratedAt         string                  `json:"generated_at"`
	OutcomeDefinition   string                  `json:"outcome_definition"`
	Horizons            []string                `json:"horizons"`
	Priors              map[string]any          `json:"priors"`
	Weights             map[string]any          `json:"weights"`
	Dependencies        map[string]any          `json:"dependencies"`
	Estimates           []EstimateSummary       `json:"estimates"`
	Intervals           string                  `json:"intervals"`
	Sensitivity         []SensitivitySummary    `json:"sensitivity"`
	ExternalForecasts   ExternalForecastSummary `json:"external_forecasts"`
	Changes             []string                `json:"changes"`
	EditorialRiskLevel  EditorialRiskLevel      `json:"editorial_risk_level"`
	UncertaintyScore    *float64                `json:"uncertainty_score"`
	DataSummary         DataSummary             `json:"data_summary"`
	Reviewers           []string                `json:"reviewers"`
	Approval            *ApprovalPolicy         `json:"approval"`
	KnownLimitations    []string                `json:"known_limitations"`
	ReproductionCommand string                  `json:"reproduction_command"`
	Signature           string                  `json:"signature"`
	Published           *string                 `json:"published"`
	Superseded          *Superseded             `json:"superseded"`
}

// Approval is one row of approvals.json.
type Approval struct {
	ReviewerID        string `json:"reviewer_id"`
	KeyID             string `json:"key_id"`
	SignedAt          string `json:"signed_at"`
	ManifestSHA256    string `json:"manifest_sha256"`
	SignatureBase64   string `json:"signature_base64"`
	ConflictsDeclared string `json:"conflicts_declared"`
}

// Release is a loaded candidate or release directory. publishing.Release is an
// alias of this type so that the model package can consume a previous release
// without importing the publishing package.
type Release struct {
	Dir              string              `json:"-"`
	Manifest         ReleaseManifest     `json:"manifest"`
	Estimates        []Estimate          `json:"estimates"`
	Indexes          []IndexValue        `json:"indexes"`
	Aggregations     []AggregationResult `json:"aggregations"`
	Sensitivity      []SensitivityRun    `json:"sensitivity"`
	Delta            DeltaRecord         `json:"delta"`
	DriversExplained DriversExplained    `json:"drivers_explained"`
	Approvals        []Approval          `json:"approvals"`
}

// EstimateByID returns the estimate with the given id, or nil.
func (r *Release) EstimateByID(id string) *Estimate {
	if r == nil {
		return nil
	}
	for i := range r.Estimates {
		if r.Estimates[i].EstimateID == id {
			return &r.Estimates[i]
		}
	}
	return nil
}

// IndexByID returns the index value with the given id, or nil.
func (r *Release) IndexByID(id IndexID) *IndexValue {
	if r == nil {
		return nil
	}
	for i := range r.Indexes {
		if r.Indexes[i].IndexID == id {
			return &r.Indexes[i]
		}
	}
	return nil
}

// ReleaseSummary is one row of `pdoomctl release list`.
type ReleaseSummary struct {
	ReleaseID     string      `json:"release_id"`
	DataSnapshot  string      `json:"data_snapshot"`
	ModelVersions []string    `json:"model_versions"`
	Published     *string     `json:"published"`
	Superseded    *Superseded `json:"superseded"`
	IsCurrent     bool        `json:"is_current"`
}

// UserScenarioParams are the Scenario Lab inputs: a horizon and ten integer
// sliders in -2..2. Each slider shifts the logit-scale location of one or more
// factors of the experimental causal model.
type UserScenarioParams struct {
	Horizon                   string `json:"horizon"`
	CapabilityTimeline        int    `json:"capability_timeline"`
	AutonomyGrowth            int    `json:"autonomy_growth"`
	AccessLevel               int    `json:"access_level"`
	SafetyProgress            int    `json:"safety_progress"`
	GovernanceStrength        int    `json:"governance_strength"`
	ModelSecurity             int    `json:"model_security"`
	OpenWeightDiffusion       int    `json:"open_weight_diffusion"`
	InternationalCoordination int    `json:"international_coordination"`
	IncidentFrequency         int    `json:"incident_frequency"`
	Resilience                int    `json:"resilience"`
}

// Sliders returns the slider values keyed by their JSON names.
func (p UserScenarioParams) Sliders() map[string]int {
	return map[string]int{
		"capability_timeline":        p.CapabilityTimeline,
		"autonomy_growth":            p.AutonomyGrowth,
		"access_level":               p.AccessLevel,
		"safety_progress":            p.SafetyProgress,
		"governance_strength":        p.GovernanceStrength,
		"model_security":             p.ModelSecurity,
		"open_weight_diffusion":      p.OpenWeightDiffusion,
		"international_coordination": p.InternationalCoordination,
		"incident_frequency":         p.IncidentFrequency,
		"resilience":                 p.Resilience,
	}
}

// OutcomeEstimate is one Scenario Lab output distribution summary.
type OutcomeEstimate struct {
	P05  float64 `json:"p05"`
	P25  float64 `json:"p25"`
	P50  float64 `json:"p50"`
	P75  float64 `json:"p75"`
	P95  float64 `json:"p95"`
	Mean float64 `json:"mean"`
}

// UserScenarioResult is the Scenario Lab output. It is labelled user_scenario
// and is never an official estimate.
type UserScenarioResult struct {
	Label            string                     `json:"label"`
	Disclaimer       string                     `json:"disclaimer"`
	Horizon          string                     `json:"horizon"`
	OutcomeEstimates map[string]OutcomeEstimate `json:"outcome_estimates"`
	FactorSummary    map[string]TriQuantile     `json:"factor_summary"`
	Flags            []string                   `json:"flags"`
	ParamsEcho       UserScenarioParams         `json:"params_echo"`
	SpecVersion      string                     `json:"spec_version"`
}
