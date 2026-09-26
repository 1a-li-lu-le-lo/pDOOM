// Copyright NU Cybernetics. p(DOOM) — research prototype.

package schema

// Envelope is the JSON file envelope of every snapshot entity file (build-spec §3.1).
type Envelope[T any] struct {
	Kind          string `json:"kind"`
	SchemaVersion int    `json:"schema_version"`
	Items         []T    `json:"items"`
}

// Verification is the verification record carried by every data-bearing entity.
type Verification struct {
	Status    VerificationStatus `json:"status"`
	CheckedAt string             `json:"checked_at"`
	Method    string             `json:"method"`
	Note      string             `json:"note"`
}

// Review holds the two review fields present on every reviewed entity.
type Review struct {
	Verification      Verification      `json:"verification"`
	HumanReviewStatus HumanReviewStatus `json:"human_review_status"`
	ModelUseStatus    ModelUseStatus    `json:"model_use_status"`
}

// Definition is a glossary definition (Appendix A: definition).
type Definition struct {
	ID              string           `json:"id"`
	Term            string           `json:"term"`
	ShortDefinition string           `json:"short_definition"`
	Definitions     []DefinitionText `json:"definitions"`
	Consensus       Consensus        `json:"consensus"`
	RelatedIDs      []string         `json:"related_ids"`
	SeeAlsoURLs     []string         `json:"see_also_urls"`
	Review
}

// DefinitionText is one attributed definition text.
type DefinitionText struct {
	Text        string  `json:"text"`
	SourceID    *string `json:"source_id"`
	Attribution string  `json:"attribution"`
	Note        string  `json:"note"`
}

// Source is a source record (Appendix A: source).
type Source struct {
	ID               string           `json:"id"`
	CanonicalURL     string           `json:"canonical_url"`
	Title            string           `json:"title"`
	Publisher        string           `json:"publisher"`
	Authors          []string         `json:"authors"`
	DatePublished    *string          `json:"date_published"`
	DateUpdated      *string          `json:"date_updated"`
	DateRetrieved    string           `json:"date_retrieved"`
	SourceTier       SourceTier       `json:"source_tier"`
	SourceType       SourceType       `json:"source_type"`
	Jurisdiction     *string          `json:"jurisdiction"`
	Topic            []string         `json:"topic"`
	ClaimIDs         []string         `json:"claim_ids"`
	EvidenceSummary  string           `json:"evidence_summary"`
	Counterevidence  *string          `json:"counterevidence"`
	Methodology      *string          `json:"methodology"`
	Sample           *string          `json:"sample"`
	Limitations      *string          `json:"limitations"`
	Conflicts        []ConflictLabel  `json:"conflicts"`
	License          *string          `json:"license"`
	RobotsStatus     RobotsStatus     `json:"robots_status"`
	ContentHash      *string          `json:"content_hash"`
	ArchiveReference *string          `json:"archive_reference"`
	Language         string           `json:"language"`
	Translation      *string          `json:"translation"`
	DuplicateGroup   *string          `json:"duplicate_group"`
	RetractionStatus RetractionStatus `json:"retraction_status"`
	CorrectionStatus *string          `json:"correction_status"`
	Citation         string           `json:"citation"`
	Review
}

// Claim is an atomic claim extracted from a source (Appendix A: claim).
type Claim struct {
	ID                 string            `json:"id"`
	Text               string            `json:"text"`
	Subject            string            `json:"subject"`
	Predicate          string            `json:"predicate"`
	Object             string            `json:"object"`
	Date               *string           `json:"date"`
	Horizon            *string           `json:"horizon"`
	Geography          *string           `json:"geography"`
	ModelName          *string           `json:"model_name"`
	ModelVersion       *string           `json:"model_version"`
	SourceID           string            `json:"source_id"`
	EvidenceType       ClaimEvidenceType `json:"evidence_type"`
	QuantitativeValue  *float64          `json:"quantitative_value"`
	Unit               *string           `json:"unit"`
	Uncertainty        *string           `json:"uncertainty"`
	DirectQuotePointer *string           `json:"direct_quote_pointer"`
	Context            string            `json:"context"`
	CorroborationIDs   []string          `json:"corroboration_ids"`
	ContradictionIDs   []string          `json:"contradiction_ids"`
	Relevance          Relevance         `json:"relevance"`
	Status             ClaimStatus       `json:"status"`
	Review
}

// Quantiles is a (possibly partial) set of forecast quantiles on the probability scale.
type Quantiles struct {
	P05 *float64 `json:"p05,omitempty"`
	P25 *float64 `json:"p25,omitempty"`
	P50 *float64 `json:"p50,omitempty"`
	P75 *float64 `json:"p75,omitempty"`
	P95 *float64 `json:"p95,omitempty"`
}

// Forecast is an external forecast record (Appendix A: forecast).
type Forecast struct {
	ID                      string             `json:"id"`
	ForecasterOrSurvey      string             `json:"forecaster_or_survey"`
	SourceID                string             `json:"source_id"`
	Date                    string             `json:"date"`
	Population              ForecastPopulation `json:"population"`
	SampleSize              *int               `json:"sample_size"`
	Expertise               string             `json:"expertise"`
	QuestionWordingOriginal string             `json:"question_wording_original"`
	Paraphrase              bool               `json:"paraphrase"`
	OutcomeSet              []Outcome          `json:"outcome_set"`
	Horizon                 Horizon            `json:"horizon"`
	HorizonNote             *string            `json:"horizon_note"`
	HorizonEndYear          *int               `json:"horizon_end_year"`
	Conditions              string             `json:"conditions"`
	Mean                    *float64           `json:"mean"`
	Median                  *float64           `json:"median"`
	Quantiles               *Quantiles         `json:"quantiles"`
	ResponseRate            *float64           `json:"response_rate"`
	SelectionEffects        *string            `json:"selection_effects"`
	FramingEffects          *string            `json:"framing_effects"`
	Calibration             *string            `json:"calibration"`
	GroupID                 *string            `json:"group_id"`
	TransformationNote      *string            `json:"transformation_note"`
	Status                  ForecastStatus     `json:"status"`
	Review
}

// CentralValue returns the value used by aggregation: median, else mean, else
// quantiles.p50. ok is false when none is present.
func (f Forecast) CentralValue() (v float64, ok bool) {
	switch {
	case f.Median != nil:
		return *f.Median, true
	case f.Mean != nil:
		return *f.Mean, true
	case f.Quantiles != nil && f.Quantiles.P50 != nil:
		return *f.Quantiles.P50, true
	}
	return 0, false
}

// Benchmark is a capability benchmark (Appendix A: benchmark).
type Benchmark struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Maintainer        string             `json:"maintainer"`
	Version           *string            `json:"version"`
	URL               string             `json:"url"`
	Tasks             string             `json:"tasks"`
	ContaminationRisk ContaminationRisk  `json:"contamination_risk"`
	Saturation        Saturation         `json:"saturation"`
	Scaffold          *string            `json:"scaffold"`
	ModelAccess       *string            `json:"model_access"`
	Unit              string             `json:"unit"`
	Direction         BenchmarkDirection `json:"direction"`
	Limitations       string             `json:"limitations"`
	PDoomRelevance    Relevance          `json:"pdoom_relevance"`
	WeightNote        *string            `json:"weight_note"`
	SourceIDs         []string           `json:"source_ids"`
	Review
}

// BenchmarkResult is one measured result (Appendix A: benchmark_result).
type BenchmarkResult struct {
	ID             string     `json:"id"`
	BenchmarkID    string     `json:"benchmark_id"`
	ModelName      string     `json:"model_name"`
	ModelDeveloper string     `json:"model_developer"`
	Date           string     `json:"date"`
	Value          float64    `json:"value"`
	Unit           string     `json:"unit"`
	CILow          *float64   `json:"ci_low"`
	CIHigh         *float64   `json:"ci_high"`
	Scaffold       *string    `json:"scaffold"`
	Confidence     Confidence `json:"confidence"`
	Note           *string    `json:"note"`
	SourceIDs      []string   `json:"source_ids"`
	Review
}

// ExternalIDs holds registry identifiers of an incident.
type ExternalIDs struct {
	AIID       *string `json:"aiid,omitempty"`
	OECDAIM    *string `json:"oecd_aim,omitempty"`
	MITTracker *string `json:"mit_tracker,omitempty"`
	CVE        *string `json:"cve,omitempty"`
	Docket     *string `json:"docket,omitempty"`
	Other      *string `json:"other,omitempty"`
}

// Incident is an incident record (Appendix A: incident). Summaries are
// non-graphic and non-operational by rule.
type Incident struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Date            *string         `json:"date"`
	DatePrecision   DatePrecision   `json:"date_precision"`
	ExternalIDs     ExternalIDs     `json:"external_ids"`
	Summary         string          `json:"summary"`
	Cause           []IncidentCause `json:"cause"`
	Harm            []Harm          `json:"harm"`
	Severity        Severity        `json:"severity"`
	PDoomRelevance  Relevance       `json:"pdoom_relevance"`
	EvidenceLevel   EvidenceLevel   `json:"evidence_level"`
	SystemsInvolved []string        `json:"systems_involved"`
	Jurisdiction    *string         `json:"jurisdiction"`
	NearMiss        bool            `json:"near_miss"`
	Novelty         Novelty         `json:"novelty"`
	ExposureNote    *string         `json:"exposure_note"`
	SourceIDs       []string        `json:"source_ids"`
	Review
}

// Scenario is a category-level pathway description (Appendix A: scenario).
type Scenario struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	OutcomeSet           []Outcome         `json:"outcome_set"`
	Description          string            `json:"description"`
	Prerequisites        []string          `json:"prerequisites"`
	EarlyIndicators      []string          `json:"early_indicators"`
	Counterindicators    []string          `json:"counterindicators"`
	CapabilityThresholds []string          `json:"capability_thresholds"`
	Exposure             string            `json:"exposure"`
	ControlFailures      []string          `json:"control_failures"`
	HumanContributions   []string          `json:"human_contributions"`
	AIContributions      []string          `json:"ai_contributions"`
	Dependencies         []string          `json:"dependencies"`
	TimeHorizonNote      string            `json:"time_horizon_note"`
	ProbabilitySource    ProbabilitySource `json:"probability_source"`
	Uncertainty          UncertaintyLabel  `json:"uncertainty"`
	InterventionIDs      []string          `json:"intervention_ids"`
	Recoverability       Recoverability    `json:"recoverability"`
	EvidenceSummary      string            `json:"evidence_summary"`
	SourceIDs            []string          `json:"source_ids"`
	OpenQuestions        []string          `json:"open_questions"`
	ContentSafetyNote    *string           `json:"content_safety_note"`
	Review
}

// ScenarioEdge is a relation between two scenarios (Appendix A: scenario_edge).
type ScenarioEdge struct {
	ID         string       `json:"id"`
	FromID     string       `json:"from_id"`
	ToID       string       `json:"to_id"`
	Relation   EdgeRelation `json:"relation"`
	Confidence Confidence   `json:"confidence"`
	Rationale  string       `json:"rationale"`
	SourceIDs  []string     `json:"source_ids"`
}

// Signal is one observable signal of a driver family.
type Signal struct {
	SignalID              string          `json:"signal_id"`
	Name                  string          `json:"name"`
	Description           string          `json:"description"`
	Direction             SignalDirection `json:"direction"`
	Normalization         string          `json:"normalization"`
	RawUnit               *string         `json:"raw_unit"`
	PreferredSourceTypes  []string        `json:"preferred_source_types"`
	ObservationVsJudgment string          `json:"observation_vs_judgment"`
}

// Driver is a driver family D1..D10 with its signals (Appendix A: driver).
type Driver struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Signals     []Signal `json:"signals"`
}

// DriverObservation is one normalized observation of a signal (Appendix A: driver_observation).
type DriverObservation struct {
	ID              string          `json:"id"`
	SignalID        string          `json:"signal_id"`
	Family          DriverFamily    `json:"family"`
	ValueNormalized float64         `json:"value_normalized"`
	RawValue        *float64        `json:"raw_value"`
	RawUnit         *string         `json:"raw_unit"`
	Confidence      float64         `json:"confidence"`
	ObservationKind ObservationKind `json:"observation_kind"`
	AsOf            string          `json:"as_of"`
	Rationale       string          `json:"rationale"`
	Counterevidence *string         `json:"counterevidence"`
	SourceIDs       []string        `json:"source_ids"`
	Review
}

// UserAction is an audience/action pair on an intervention.
type UserAction struct {
	Audience string `json:"audience"`
	Action   string `json:"action"`
}

// Intervention is a safeguard (Appendix A: intervention).
type Intervention struct {
	ID                string               `json:"id"`
	Name              string               `json:"name"`
	TargetScenarioIDs []string             `json:"target_scenario_ids"`
	Mechanism         string               `json:"mechanism"`
	EvidenceSummary   string               `json:"evidence_summary"`
	EvidenceStrength  EvidenceStrength     `json:"evidence_strength"`
	Cost              Cost                 `json:"cost"`
	TimeToDeploy      TimeToDeploy         `json:"time_to_deploy"`
	EffectSize        EffectSize           `json:"effect_size"`
	Uncertainty       UncertaintyLabel     `json:"uncertainty"`
	PossibleFailure   string               `json:"possible_failure"`
	PossibleBackfire  string               `json:"possible_backfire"`
	OwnerTypes        []string             `json:"owner_types"`
	UserActions       []UserAction         `json:"user_actions"`
	Category          InterventionCategory `json:"category"`
	SourceIDs         []string             `json:"source_ids"`
	Review
}

// Organization is an organization record (Appendix A: organization).
type Organization struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	URL                  string   `json:"url"`
	Mission              string   `json:"mission"`
	LegalStatus          string   `json:"legal_status"`
	Jurisdiction         string   `json:"jurisdiction"`
	Focus                []string `json:"focus"`
	Programs             []string `json:"programs"`
	OpenOutputs          []string `json:"open_outputs"`
	FundingDisclosure    string   `json:"funding_disclosure"`
	Conflicts            []string `json:"conflicts"`
	EvidenceOfImpact     string   `json:"evidence_of_impact"`
	WaysToHelp           []string `json:"ways_to_help"`
	InclusionCriteriaMet []string `json:"inclusion_criteria_met"`
	LastVerified         string   `json:"last_verified"`
	SourceIDs            []string `json:"source_ids"`
	Review
}

// ActionResource is a linked resource on an action.
type ActionResource struct {
	Title    string  `json:"title"`
	URL      string  `json:"url"`
	SourceID *string `json:"source_id"`
}

// Action is an audience-specific action (Appendix A: action; optional file actions.json).
type Action struct {
	ID                     string           `json:"id"`
	Audience               Audience         `json:"audience"`
	Title                  string           `json:"title"`
	Description            string           `json:"description"`
	RelatedInterventionIDs []string         `json:"related_intervention_ids"`
	Resources              []ActionResource `json:"resources"`
	Effort                 Effort           `json:"effort"`
	Review
}

// TriQuantile is a {p05,p50,p95} specification on the probability scale. It is
// used both for experimental_causal parameters and for factor summaries.
type TriQuantile struct {
	P05 float64 `json:"p05"`
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
}

// Ordered reports whether p05 < p50 < p95.
func (t TriQuantile) Ordered() bool { return t.P05 < t.P50 && t.P50 < t.P95 }

// HorizonCausalSpec holds the per-horizon factor and outcome-share specs.
type HorizonCausalSpec struct {
	A TriQuantile            `json:"A"`
	C TriQuantile            `json:"C"`
	E TriQuantile            `json:"E"`
	F TriQuantile            `json:"F"`
	O map[string]TriQuantile `json:"O"`
}

// CausalRationale documents the reasoning behind the causal parameters.
type CausalRationale struct {
	A          string `json:"A"`
	C          string `json:"C"`
	E          string `json:"E"`
	F          string `json:"F"`
	O          string `json:"O"`
	Dependence string `json:"dependence"`
}

// ExperimentalCausalSpec parameterizes the experimental research-mode model.
type ExperimentalCausalSpec struct {
	Version             string                       `json:"version"`
	Seed                int64                        `json:"seed"`
	Samples             int                          `json:"samples"`
	CommonFactorLoading float64                      `json:"common_factor_loading"`
	Horizons            map[string]HorizonCausalSpec `json:"horizons"`
	Rationale           CausalRationale              `json:"rationale"`
	SourceIDs           []string                     `json:"source_ids"`
}

// IndexWeights holds signal weights (signal_id → weight) for the signal-based
// indexes and named constants for the derived ones.
type IndexWeights struct {
	CapabilityPressure        map[string]float64 `json:"capability_pressure"`
	ControlStrength           map[string]float64 `json:"control_strength"`
	IncidentPressure          map[string]float64 `json:"incident_pressure"`
	EvidencePressure          map[string]float64 `json:"evidence_pressure"`
	Uncertainty               map[string]float64 `json:"uncertainty"`
	AgenticInfrastructureRisk map[string]float64 `json:"agentic_infrastructure_risk"`
}

// IncidentScoring holds the incident pressure constants.
type IncidentScoring struct {
	SeverityWeights     map[string]float64 `json:"severity_weights"`
	RelevanceWeights    map[string]float64 `json:"relevance_weights"`
	EvidenceWeights     map[string]float64 `json:"evidence_weights"`
	RecencyHalfLifeDays int                `json:"recency_half_life_days"`
	SquashK             float64            `json:"squash_k"`
}

// EditorialRule maps a condition to an editorial risk level; first match wins.
type EditorialRule struct {
	Level EditorialRiskLevel `json:"level"`
	When  string             `json:"when"`
}

// RoundingRules gives the display rounding step (in percentage points) per uncertainty label.
type RoundingRules struct {
	Extreme  int `json:"extreme"`
	High     int `json:"high"`
	Moderate int `json:"moderate"`
	Low      int `json:"low"`
}

// ModelSpec is the single-item model specification (Appendix A: model_spec).
type ModelSpec struct {
	ID                 string                 `json:"id"`
	IndexWeights       IndexWeights           `json:"index_weights"`
	WeightBounds       []float64              `json:"weight_bounds"`
	TierMultipliers    map[string]float64     `json:"tier_multipliers"`
	IncidentScoring    IncidentScoring        `json:"incident_scoring"`
	EditorialRules     []EditorialRule        `json:"editorial_rules"`
	RoundingRules      RoundingRules          `json:"rounding_rules"`
	AggregationMethods []AggregationMethod    `json:"aggregation_methods"`
	ExperimentalCausal ExperimentalCausalSpec `json:"experimental_causal"`
}

// TierMultiplier returns the multiplier for a tier (0 when absent).
func (m ModelSpec) TierMultiplier(t SourceTier) float64 {
	switch t {
	case 1:
		return m.TierMultipliers["1"]
	case 2:
		return m.TierMultipliers["2"]
	case 3:
		return m.TierMultipliers["3"]
	case 4:
		return m.TierMultipliers["4"]
	case 5:
		return m.TierMultipliers["5"]
	}
	return 0
}

// SnapshotFile is one entry of manifest.files.
type SnapshotFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Count  int    `json:"count"`
}

// SnapshotManifest is data/snapshots/<id>/manifest.json (build-spec §3.5).
type SnapshotManifest struct {
	SnapshotID         string         `json:"snapshot_id"`
	CreatedAt          string         `json:"created_at"`
	SourceCutoff       string         `json:"source_cutoff"`
	BaselineSnapshotID *string        `json:"baseline_snapshot_id"`
	Files              []SnapshotFile `json:"files"`
	Notes              string         `json:"notes"`
}
