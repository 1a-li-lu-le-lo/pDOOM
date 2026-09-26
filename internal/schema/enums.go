// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package schema holds the Go structs that mirror @pdoom/schemas: every entity of
// a data snapshot (build-spec Appendix A), every release object (build-spec
// §3.6), the Scenario Lab parameter/result objects, typed enumerations
// (build-spec §3.3) and the JSON file envelope (§3.1).
//
// The package is a leaf: it imports only the standard library so that every
// other Go package (snapshot, model, publishing, audit, api) can depend on it.
package schema

import "strings"

// Outcome is one of the nine outcome codes O0..O8 (build-spec §3.3).
type Outcome string

// Outcome codes.
const (
	O0 Outcome = "O0" // beneficial_or_manageable
	O1 Outcome = "O1" // serious_reversible_harm
	O2 Outcome = "O2" // systemic_authoritarian_or_oligopolistic_control
	O3 Outcome = "O3" // permanent_severe_disempowerment
	O4 Outcome = "O4" // civilizational_collapse
	O5 Outcome = "O5" // near_extinction
	O6 Outcome = "O6" // human_extinction
	O7 Outcome = "O7" // biospheric_catastrophe
	O8 Outcome = "O8" // other_irreversible_loss
)

// Outcomes lists every outcome code in order.
var Outcomes = []Outcome{O0, O1, O2, O3, O4, O5, O6, O7, O8}

// OutcomeSlugs maps outcome codes to their snake_case slug.
var OutcomeSlugs = map[Outcome]string{
	O0: "beneficial_or_manageable",
	O1: "serious_reversible_harm",
	O2: "systemic_authoritarian_or_oligopolistic_control",
	O3: "permanent_severe_disempowerment",
	O4: "civilizational_collapse",
	O5: "near_extinction",
	O6: "human_extinction",
	O7: "biospheric_catastrophe",
	O8: "other_irreversible_loss",
}

// OutcomeLabels maps outcome codes to a short human label.
var OutcomeLabels = map[Outcome]string{
	O0: "Beneficial or manageable",
	O1: "Serious reversible harm",
	O2: "Systemic authoritarian or oligopolistic control",
	O3: "Permanent severe disempowerment",
	O4: "Civilizational collapse",
	O5: "Near extinction",
	O6: "Human extinction",
	O7: "Biospheric catastrophe",
	O8: "Other irreversible loss",
}

// Derived outcome sets (build-spec §3.3). p(DOOM) is O3..O8 and may only be shown
// next to its decomposition (§0.4).
var (
	PDoomOutcomes          = []Outcome{O3, O4, O5, O6, O7, O8}
	ExtinctionOutcomes     = []Outcome{O6}
	DisempowermentOutcomes = []Outcome{O3}
	CollapseOutcomes       = []Outcome{O4, O5}
	BiosphereOutcomes      = []Outcome{O7}
)

// Valid reports whether the outcome code is one of O0..O8.
func (o Outcome) Valid() bool { return OutcomeSlugs[o] != "" }

// Horizon is a forecast horizon key measured from forecast_origin_date.
type Horizon string

// Horizon keys.
const (
	Horizon1y       Horizon = "1y"
	Horizon3y       Horizon = "3y"
	Horizon5y       Horizon = "5y"
	Horizon10y      Horizon = "10y"
	Horizon25y      Horizon = "25y"
	Horizon2100     Horizon = "2100"
	HorizonEventual Horizon = "eventual"
	// HorizonCustom is allowed on forecasts only, together with horizon_note.
	HorizonCustom Horizon = "custom"
)

// Horizons lists the canonical horizon keys in increasing order.
var Horizons = []Horizon{Horizon1y, Horizon3y, Horizon5y, Horizon10y, Horizon25y, Horizon2100, HorizonEventual}

// Valid reports whether h is a canonical horizon key (custom is not).
func (h Horizon) Valid() bool { return HorizonIndex(h) >= 0 }

// HorizonIndex returns the position of h in Horizons, or -1.
func HorizonIndex(h Horizon) int {
	for i, x := range Horizons {
		if x == h {
			return i
		}
	}
	return -1
}

// HorizonLabel returns a human-readable label for a horizon key.
func HorizonLabel(h Horizon) string {
	switch h {
	case Horizon1y:
		return "1 year"
	case Horizon3y:
		return "3 years"
	case Horizon5y:
		return "5 years"
	case Horizon10y:
		return "10 years"
	case Horizon25y:
		return "25 years"
	case Horizon2100:
		return "by 2100"
	case HorizonEventual:
		return "eventually"
	case HorizonCustom:
		return "custom"
	}
	return string(h)
}

// SourceTier is 1..5 (1 primary/authoritative ... 5 unverified).
type SourceTier int

// Valid reports whether the tier is in 1..5.
func (t SourceTier) Valid() bool { return t >= 1 && t <= 5 }

// stringEnum is a small helper for typed string enumerations.
func inSet(v string, set []string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// SourceType classifies a source document.
type SourceType string

// SourceTypes lists the allowed values.
var SourceTypes = []string{"paper", "preprint", "dataset", "code", "model_card", "safety_framework", "government", "standard", "court", "regulatory", "company_disclosure", "incident_report", "survey", "review_article", "evaluation", "journalism", "blog", "newsletter", "talk", "podcast", "social", "forecast_platform", "other"}

// Valid reports whether the value is allowed.
func (s SourceType) Valid() bool { return inSet(string(s), SourceTypes) }

// ConflictLabel records a declared conflict of interest for a source.
type ConflictLabel string

// ConflictLabels lists the allowed values.
var ConflictLabels = []string{"developer_self_report", "advocacy_context", "government_policy_context", "commercial_interest", "funder_relationship", "none_known"}

// Valid reports whether the value is allowed.
func (c ConflictLabel) Valid() bool { return inSet(string(c), ConflictLabels) }

// HumanReviewStatus is the human review state of an entity.
type HumanReviewStatus string

// HumanReviewStatuses lists the allowed values.
var HumanReviewStatuses = []string{"pending", "reviewed", "disputed"}

// Valid reports whether the value is allowed.
func (s HumanReviewStatus) Valid() bool { return inSet(string(s), HumanReviewStatuses) }

// ModelUseStatus says whether an entity may feed the model.
type ModelUseStatus string

// Model use statuses.
const (
	ModelUseExcluded      ModelUseStatus = "excluded"
	ModelUseInformational ModelUseStatus = "informational"
	ModelUseEligible      ModelUseStatus = "eligible"
	ModelUseUsed          ModelUseStatus = "used"
)

// ModelUseStatuses lists the allowed values.
var ModelUseStatuses = []string{"excluded", "informational", "eligible", "used"}

// Valid reports whether the value is allowed.
func (s ModelUseStatus) Valid() bool { return inSet(string(s), ModelUseStatuses) }

// FeedsModel reports whether the status allows the entity to enter a computation.
func (s ModelUseStatus) FeedsModel() bool { return s == ModelUseEligible || s == ModelUseUsed }

// VerificationStatus records how an item was verified.
type VerificationStatus string

// Verification statuses.
const (
	VerifiedFetch          VerificationStatus = "verified_fetch"
	VerifiedSearch         VerificationStatus = "verified_search"
	VerifiedPriorKnowledge VerificationStatus = "verified_prior_knowledge"
	Unverified             VerificationStatus = "unverified"
)

// VerificationStatuses lists the allowed values.
var VerificationStatuses = []string{"verified_fetch", "verified_search", "verified_prior_knowledge", "unverified"}

// Valid reports whether the value is allowed.
func (s VerificationStatus) Valid() bool { return inSet(string(s), VerificationStatuses) }

// AllowsModelUse reports whether items with this status may be eligible/used (§3.3).
func (s VerificationStatus) AllowsModelUse() bool { return s == VerifiedFetch || s == VerifiedSearch }

// ClaimEvidenceType classifies the evidence behind a claim.
type ClaimEvidenceType string

// ClaimEvidenceTypes lists the allowed values.
var ClaimEvidenceTypes = []string{"measurement", "survey_result", "forecast", "incident_report", "policy_text", "expert_judgment", "model_output", "anecdote"}

// Valid reports whether the value is allowed.
func (c ClaimEvidenceType) Valid() bool { return inSet(string(c), ClaimEvidenceTypes) }

// ClaimStatus is the corroboration state of a claim.
type ClaimStatus string

// ClaimStatuses lists the allowed values.
var ClaimStatuses = []string{"candidate", "corroborated", "contradicted", "retracted", "superseded"}

// Valid reports whether the value is allowed.
func (c ClaimStatus) Valid() bool { return inSet(string(c), ClaimStatuses) }

// ForecastPopulation is the population a forecast was drawn from.
type ForecastPopulation string

// ForecastPopulations lists the allowed values.
var ForecastPopulations = []string{"general_ai_researchers", "frontier_lab_researchers", "ai_safety_researchers", "superforecasters", "domain_experts", "economists", "governance_researchers", "public_forecasters", "prediction_market", "individual_expert", "organization"}

// Valid reports whether the value is allowed.
func (p ForecastPopulation) Valid() bool { return inSet(string(p), ForecastPopulations) }

// ForecastStatus is the lifecycle state of a forecast record.
type ForecastStatus string

// ForecastStatuses lists the allowed values.
var ForecastStatuses = []string{"current", "superseded", "withdrawn"}

// Valid reports whether the value is allowed.
func (s ForecastStatus) Valid() bool { return inSet(string(s), ForecastStatuses) }

// IncidentCause classifies an incident cause.
type IncidentCause string

// IncidentCauses lists the allowed values.
var IncidentCauses = []string{"malicious_use", "malfunction", "human_misuse", "organizational_failure", "security_compromise", "insufficient_oversight", "systemic_interaction", "unclear"}

// Valid reports whether the value is allowed.
func (c IncidentCause) Valid() bool { return inSet(string(c), IncidentCauses) }

// Harm classifies the kind of harm in an incident.
type Harm string

// Harms lists the allowed values.
var Harms = []string{"physical", "psychological", "financial", "informational", "political", "environmental", "privacy", "security", "civil_rights", "institutional", "infrastructure"}

// Valid reports whether the value is allowed.
func (h Harm) Valid() bool { return inSet(string(h), Harms) }

// Severity is the incident severity.
type Severity string

// Severities lists the allowed values.
var Severities = []string{"negligible", "minor", "material", "major", "severe", "catastrophic"}

// Valid reports whether the value is allowed.
func (s Severity) Valid() bool { return inSet(string(s), Severities) }

// Relevance is p(DOOM) relevance of a claim, benchmark or incident.
type Relevance string

// Relevances lists the allowed values.
var Relevances = []string{"none", "weak", "indirect", "moderate", "strong", "direct_precursor"}

// Valid reports whether the value is allowed.
func (r Relevance) Valid() bool { return inSet(string(r), Relevances) }

// EvidenceLevel is the evidentiary standing of an incident record.
type EvidenceLevel string

// EvidenceLevels lists the allowed values.
var EvidenceLevels = []string{"allegation", "single_source_report", "corroborated_report", "official_finding", "peer_reviewed_analysis", "independently_reproduced"}

// Valid reports whether the value is allowed.
func (e EvidenceLevel) Valid() bool { return inSet(string(e), EvidenceLevels) }

// DriverFamily is D1..D10.
type DriverFamily string

// DriverFamilyNames maps family ids to their names.
var DriverFamilyNames = map[DriverFamily]string{
	"D1": "capability", "D2": "autonomy", "D3": "access_exposure", "D4": "scalability", "D5": "alignment_control",
	"D6": "security", "D7": "governance", "D8": "incidents", "D9": "race_dynamics", "D10": "resilience",
}

// DriverFamilies lists the family ids in order.
var DriverFamilies = []DriverFamily{"D1", "D2", "D3", "D4", "D5", "D6", "D7", "D8", "D9", "D10"}

// Valid reports whether the value is allowed.
func (d DriverFamily) Valid() bool { return DriverFamilyNames[d] != "" }

// ObservationKind distinguishes observations from judgments.
type ObservationKind string

// ObservationKinds lists the allowed values.
var ObservationKinds = []string{"observation", "judgment"}

// Valid reports whether the value is allowed.
func (o ObservationKind) Valid() bool { return inSet(string(o), ObservationKinds) }

// UncertaintyLabel is the uncertainty / disagreement label.
type UncertaintyLabel string

// Uncertainty labels.
const (
	UncertaintyLow      UncertaintyLabel = "low"
	UncertaintyModerate UncertaintyLabel = "moderate"
	UncertaintyHigh     UncertaintyLabel = "high"
	UncertaintyExtreme  UncertaintyLabel = "extreme"
)

// UncertaintyLabels lists the allowed values.
var UncertaintyLabels = []string{"low", "moderate", "high", "extreme"}

// Valid reports whether the value is allowed.
func (u UncertaintyLabel) Valid() bool { return inSet(string(u), UncertaintyLabels) }

// EditorialRiskLevel is the human-set editorial level (never a probability).
type EditorialRiskLevel string

// EditorialRiskLevels lists the allowed values.
var EditorialRiskLevels = []string{"very_low", "low", "guarded", "elevated", "high", "severe_uncertainty", "insufficient_evidence"}

// Valid reports whether the value is allowed.
func (e EditorialRiskLevel) Valid() bool { return inSet(string(e), EditorialRiskLevels) }

// EstimateStatus is the status of an estimate object.
type EstimateStatus string

// Estimate statuses.
const (
	StatusOfficial                EstimateStatus = "official"
	StatusInsufficientlyCalibrate EstimateStatus = "insufficiently_calibrated"
	StatusExternalAggregate       EstimateStatus = "external_aggregate"
	StatusResearchMode            EstimateStatus = "research_mode"
	StatusUserScenario            EstimateStatus = "user_scenario"
)

// EstimateStatuses lists the allowed values.
var EstimateStatuses = []string{"official", "insufficiently_calibrated", "external_aggregate", "research_mode", "user_scenario"}

// Valid reports whether the value is allowed.
func (s EstimateStatus) Valid() bool { return inSet(string(s), EstimateStatuses) }

// IndexID identifies one of the published 0–100 indexes.
type IndexID string

// Index ids.
const (
	IndexEvidencePressure          IndexID = "evidence_pressure"
	IndexCapabilityPressure        IndexID = "capability_pressure"
	IndexControlStrength           IndexID = "control_strength"
	IndexIncidentPressure          IndexID = "incident_pressure"
	IndexUncertainty               IndexID = "uncertainty"
	IndexAgenticInfrastructureRisk IndexID = "agentic_infrastructure_risk"
	IndexAttention                 IndexID = "attention"
)

// IndexIDs lists the allowed values.
var IndexIDs = []string{"evidence_pressure", "capability_pressure", "control_strength", "incident_pressure", "uncertainty", "agentic_infrastructure_risk", "attention"}

// Valid reports whether the value is allowed.
func (i IndexID) Valid() bool { return inSet(string(i), IndexIDs) }

// AggregationMethod names a forecast aggregation method.
type AggregationMethod string

// Aggregation methods.
const (
	MethodUnweightedMedian        AggregationMethod = "unweighted_median"
	MethodLinearPool              AggregationMethod = "linear_pool"
	MethodLogOddsPool             AggregationMethod = "log_odds_pool"
	MethodTrimmedMean             AggregationMethod = "trimmed_mean"
	MethodTierWeighted            AggregationMethod = "tier_weighted"
	MethodRecencyWeighted         AggregationMethod = "recency_weighted"
	MethodEqualWeightByPopulation AggregationMethod = "equal_weight_by_population"
)

// AggregationMethods lists the allowed values.
var AggregationMethods = []string{"unweighted_median", "linear_pool", "log_odds_pool", "trimmed_mean", "tier_weighted", "recency_weighted", "equal_weight_by_population"}

// Valid reports whether the value is allowed.
func (m AggregationMethod) Valid() bool { return inSet(string(m), AggregationMethods) }

// Smaller closed vocabularies used by single entities.

// Consensus is the definition consensus state.
type Consensus string

// Consensuses lists the allowed values.
var Consensuses = []string{"consensus", "contested", "emerging"}

// Valid reports whether the value is allowed.
func (c Consensus) Valid() bool { return inSet(string(c), Consensuses) }

// RobotsStatus records the robots.txt outcome for a source.
type RobotsStatus string

// RobotsStatuses lists the allowed values.
var RobotsStatuses = []string{"allowed", "disallowed", "not_applicable", "unknown"}

// Valid reports whether the value is allowed.
func (r RobotsStatus) Valid() bool { return inSet(string(r), RobotsStatuses) }

// RetractionStatus records whether a source was retracted or corrected.
type RetractionStatus string

// RetractionStatuses lists the allowed values.
var RetractionStatuses = []string{"none", "retracted", "corrected", "disputed"}

// Valid reports whether the value is allowed.
func (r RetractionStatus) Valid() bool { return inSet(string(r), RetractionStatuses) }

// ContaminationRisk is a benchmark contamination label.
type ContaminationRisk string

// ContaminationRisks lists the allowed values.
var ContaminationRisks = []string{"low", "moderate", "high", "unknown"}

// Valid reports whether the value is allowed.
func (c ContaminationRisk) Valid() bool { return inSet(string(c), ContaminationRisks) }

// Saturation is a benchmark saturation label.
type Saturation string

// Saturations lists the allowed values.
var Saturations = []string{"none", "partial", "saturated", "unknown"}

// Valid reports whether the value is allowed.
func (s Saturation) Valid() bool { return inSet(string(s), Saturations) }

// BenchmarkDirection says which direction means more capability.
type BenchmarkDirection string

// BenchmarkDirections lists the allowed values.
var BenchmarkDirections = []string{"higher_is_more_capable", "lower_is_more_capable"}

// Valid reports whether the value is allowed.
func (b BenchmarkDirection) Valid() bool { return inSet(string(b), BenchmarkDirections) }

// Confidence is a three-level qualitative confidence.
type Confidence string

// Confidences lists the allowed values.
var Confidences = []string{"low", "moderate", "high"}

// Valid reports whether the value is allowed.
func (c Confidence) Valid() bool { return inSet(string(c), Confidences) }

// DatePrecision is the precision of an incident date.
type DatePrecision string

// DatePrecisions lists the allowed values.
var DatePrecisions = []string{"day", "month", "year", "unknown"}

// Valid reports whether the value is allowed.
func (d DatePrecision) Valid() bool { return inSet(string(d), DatePrecisions) }

// Novelty is the novelty of an incident.
type Novelty string

// Novelties lists the allowed values.
var Novelties = []string{"routine", "notable", "novel"}

// Valid reports whether the value is allowed.
func (n Novelty) Valid() bool { return inSet(string(n), Novelties) }

// ProbabilitySource says where a scenario's probability (if any) comes from.
type ProbabilitySource string

// ProbabilitySources lists the allowed values.
var ProbabilitySources = []string{"not_assigned", "external_forecast", "experimental_model", "expert_elicitation"}

// Valid reports whether the value is allowed.
func (p ProbabilitySource) Valid() bool { return inSet(string(p), ProbabilitySources) }

// Recoverability is the recoverability of a scenario outcome.
type Recoverability string

// Recoverabilities lists the allowed values.
var Recoverabilities = []string{"high", "moderate", "low", "none", "unknown"}

// Valid reports whether the value is allowed.
func (r Recoverability) Valid() bool { return inSet(string(r), Recoverabilities) }

// EdgeRelation is the relation on a scenario edge.
type EdgeRelation string

// EdgeRelations lists the allowed values.
var EdgeRelations = []string{"enables", "amplifies", "prevents_response", "shares_prerequisite", "competes_with"}

// Valid reports whether the value is allowed.
func (e EdgeRelation) Valid() bool { return inSet(string(e), EdgeRelations) }

// SignalDirection says what a higher normalized signal value means.
type SignalDirection string

// Signal directions.
const (
	HigherRaisesPressure     SignalDirection = "higher_raises_pressure"
	HigherStrengthensControl SignalDirection = "higher_strengthens_control"
)

// SignalDirections lists the allowed values.
var SignalDirections = []string{"higher_raises_pressure", "higher_strengthens_control"}

// Valid reports whether the value is allowed.
func (s SignalDirection) Valid() bool { return inSet(string(s), SignalDirections) }

// EvidenceStrength is the evidence strength for an intervention.
type EvidenceStrength string

// EvidenceStrengths lists the allowed values.
var EvidenceStrengths = []string{"none", "weak", "moderate", "strong"}

// Valid reports whether the value is allowed.
func (e EvidenceStrength) Valid() bool { return inSet(string(e), EvidenceStrengths) }

// Cost is a qualitative intervention cost.
type Cost string

// Costs lists the allowed values.
var Costs = []string{"low", "moderate", "high", "very_high", "unknown"}

// Valid reports whether the value is allowed.
func (c Cost) Valid() bool { return inSet(string(c), Costs) }

// TimeToDeploy is a qualitative deployment time.
type TimeToDeploy string

// TimesToDeploy lists the allowed values.
var TimesToDeploy = []string{"months", "1-2y", "3-5y", "5y+", "unknown"}

// Valid reports whether the value is allowed.
func (t TimeToDeploy) Valid() bool { return inSet(string(t), TimesToDeploy) }

// EffectSize is a qualitative-only effect size.
type EffectSize string

// EffectSizes lists the allowed values.
var EffectSizes = []string{"unknown", "small", "moderate", "large"}

// Valid reports whether the value is allowed.
func (e EffectSize) Valid() bool { return inSet(string(e), EffectSizes) }

// InterventionCategory is the category of an intervention.
type InterventionCategory string

// InterventionCategories lists the allowed values.
var InterventionCategories = []string{"technical", "organizational", "national", "international", "resilience"}

// Valid reports whether the value is allowed.
func (i InterventionCategory) Valid() bool { return inSet(string(i), InterventionCategories) }

// Audience is the audience of an action.
type Audience string

// Audiences lists the allowed values.
var Audiences = []string{"individuals", "software_engineers", "ai_researchers", "laboratories", "policymakers", "funders", "educators", "nonprofits", "auditors_red_teams", "standards_bodies"}

// Valid reports whether the value is allowed.
func (a Audience) Valid() bool { return inSet(string(a), Audiences) }

// Effort is the effort of an action.
type Effort string

// Efforts lists the allowed values.
var Efforts = []string{"low", "moderate", "high"}

// Valid reports whether the value is allowed.
func (e Effort) Valid() bool { return inSet(string(e), Efforts) }

// Kind is the value of the "kind" field of a snapshot file envelope.
type Kind string

// Kinds lists the allowed envelope kinds (build-spec §3.1 plus the optional action file).
var Kinds = []string{"definition", "source", "claim", "forecast", "benchmark", "benchmark_result", "incident", "scenario", "scenario_edge", "driver", "driver_observation", "intervention", "organization", "model_spec", "action"}

// Valid reports whether the value is allowed.
func (k Kind) Valid() bool { return inSet(string(k), Kinds) }

// OutcomeSetKey renders an outcome set as a stable key such as "O4+O5".
func OutcomeSetKey(set []Outcome) string {
	parts := make([]string, 0, len(set))
	for _, o := range set {
		parts = append(parts, string(o))
	}
	return strings.Join(parts, "+")
}

// OutcomeSetLabel renders a human label for an outcome set; the derived sets get
// their derived names, anything else the joined outcome labels.
func OutcomeSetLabel(set []Outcome) string {
	switch OutcomeSetKey(set) {
	case OutcomeSetKey(PDoomOutcomes):
		return "p(DOOM) (O3–O8 combined)"
	case OutcomeSetKey(CollapseOutcomes):
		return "Civilizational collapse or near extinction (O4+O5)"
	}
	parts := make([]string, 0, len(set))
	for _, o := range set {
		parts = append(parts, OutcomeLabels[o])
	}
	return strings.Join(parts, " or ")
}
