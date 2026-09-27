// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Every enumeration from build-spec §3.3 (and the small per-entity enumerations
 * from Appendix A) as zod v4 enums, plus the display / rule tables that the web
 * app, the MCP server and the Go code mirror.
 *
 * Exact string values are part of the cross-language contract. Do not rename.
 */
import { z } from "zod";
import { HORIZON_KEYS, OUTCOME_CODES, OUTCOME_SLUGS } from "./constants";

// ---------------------------------------------------------------------------
// §3.3 enumerations
// ---------------------------------------------------------------------------

export const Outcome = z.enum(OUTCOME_CODES);
export type { OutcomeCode } from "./constants";

/** Outcomes counted in the combined p(DOOM) definition (O3..O8). */
export const PDoomOutcome = z.enum(["O3", "O4", "O5", "O6", "O7", "O8"]);
export type PDoomOutcomeCode = z.infer<typeof PDoomOutcome>;

export const OutcomeSlug = z.enum(OUTCOME_SLUGS);
export type { OutcomeSlugValue } from "./constants";

/** Horizon keys in specification order. Do not rely on `Horizon.options` for order:
 * zod v4 stores enum members as object keys, and JavaScript orders integer-like keys
 * such as "2100" first. */
export { HORIZON_KEYS } from "./constants";
export const Horizon = z.enum(HORIZON_KEYS);
export type { HorizonKey } from "./constants";

/** Forecast records may carry a non-standard horizon; `custom` requires `horizon_note`. */
export const ForecastHorizon = z.enum([...HORIZON_KEYS, "custom"]);
export type ForecastHorizonKey = z.infer<typeof ForecastHorizon>;

export const SourceTier = z.literal([1, 2, 3, 4, 5]);
export type SourceTierValue = z.infer<typeof SourceTier>;

export const SourceType = z.enum([
  "paper",
  "preprint",
  "dataset",
  "code",
  "model_card",
  "safety_framework",
  "government",
  "standard",
  "court",
  "regulatory",
  "company_disclosure",
  "incident_report",
  "survey",
  "review_article",
  "evaluation",
  "journalism",
  "blog",
  "newsletter",
  "talk",
  "podcast",
  "social",
  "forecast_platform",
  "other",
]);
export type SourceTypeValue = z.infer<typeof SourceType>;

export const ConflictLabel = z.enum([
  "developer_self_report",
  "advocacy_context",
  "government_policy_context",
  "commercial_interest",
  "funder_relationship",
  "none_known",
]);
export type ConflictLabelValue = z.infer<typeof ConflictLabel>;

export const HumanReviewStatus = z.enum(["pending", "reviewed", "disputed"]);
export type HumanReviewStatusValue = z.infer<typeof HumanReviewStatus>;

export const ModelUseStatus = z.enum(["excluded", "informational", "eligible", "used"]);
export type ModelUseStatusValue = z.infer<typeof ModelUseStatus>;

export const VerificationStatus = z.enum([
  "verified_fetch",
  "verified_search",
  "verified_prior_knowledge",
  "unverified",
]);
export type VerificationStatusValue = z.infer<typeof VerificationStatus>;

/** Only these verification statuses permit `eligible` / `used` (build-spec §3.3). */
export const MODEL_USE_VERIFICATION_STATUSES: readonly VerificationStatusValue[] = [
  "verified_fetch",
  "verified_search",
];

export const ClaimEvidenceType = z.enum([
  "measurement",
  "survey_result",
  "forecast",
  "incident_report",
  "policy_text",
  "expert_judgment",
  "model_output",
  "anecdote",
]);
export type ClaimEvidenceTypeValue = z.infer<typeof ClaimEvidenceType>;

export const ClaimStatus = z.enum([
  "candidate",
  "corroborated",
  "contradicted",
  "retracted",
  "superseded",
]);
export type ClaimStatusValue = z.infer<typeof ClaimStatus>;

export const ForecastPopulation = z.enum([
  "general_ai_researchers",
  "frontier_lab_researchers",
  "ai_safety_researchers",
  "superforecasters",
  "domain_experts",
  "economists",
  "governance_researchers",
  "public_forecasters",
  "prediction_market",
  "individual_expert",
  "organization",
]);
export type ForecastPopulationValue = z.infer<typeof ForecastPopulation>;

export const IncidentCause = z.enum([
  "malicious_use",
  "malfunction",
  "human_misuse",
  "organizational_failure",
  "security_compromise",
  "insufficient_oversight",
  "systemic_interaction",
  "unclear",
]);
export type IncidentCauseValue = z.infer<typeof IncidentCause>;

export const IncidentHarm = z.enum([
  "physical",
  "psychological",
  "financial",
  "informational",
  "political",
  "environmental",
  "privacy",
  "security",
  "civil_rights",
  "institutional",
  "infrastructure",
]);
export type IncidentHarmValue = z.infer<typeof IncidentHarm>;

export const IncidentSeverity = z.enum([
  "negligible",
  "minor",
  "material",
  "major",
  "severe",
  "catastrophic",
]);
export type IncidentSeverityValue = z.infer<typeof IncidentSeverity>;

export const PDoomRelevance = z.enum([
  "none",
  "weak",
  "indirect",
  "moderate",
  "strong",
  "direct_precursor",
]);
export type PDoomRelevanceValue = z.infer<typeof PDoomRelevance>;

export const EvidenceLevel = z.enum([
  "allegation",
  "single_source_report",
  "corroborated_report",
  "official_finding",
  "peer_reviewed_analysis",
  "independently_reproduced",
]);
export type EvidenceLevelValue = z.infer<typeof EvidenceLevel>;

export const DriverFamily = z.enum(["D1", "D2", "D3", "D4", "D5", "D6", "D7", "D8", "D9", "D10"]);
export type DriverFamilyCode = z.infer<typeof DriverFamily>;

export const DriverFamilySlug = z.enum([
  "capability",
  "autonomy",
  "access_exposure",
  "scalability",
  "alignment_control",
  "security",
  "governance",
  "incidents",
  "race_dynamics",
  "resilience",
]);
export type DriverFamilySlugValue = z.infer<typeof DriverFamilySlug>;

export const ObservationKind = z.enum(["observation", "judgment"]);
export type ObservationKindValue = z.infer<typeof ObservationKind>;

export const UncertaintyLabel = z.enum(["low", "moderate", "high", "extreme"]);
export type UncertaintyLabelValue = z.infer<typeof UncertaintyLabel>;
/** Disagreement between forecasters uses the same four-step label. */
export const DisagreementLabel = UncertaintyLabel;
export type DisagreementLabelValue = UncertaintyLabelValue;

export const EditorialRiskLevel = z.enum([
  "very_low",
  "low",
  "guarded",
  "elevated",
  "high",
  "severe_uncertainty",
  "insufficient_evidence",
]);
export type EditorialRiskLevelValue = z.infer<typeof EditorialRiskLevel>;

export const EstimateStatus = z.enum([
  "official",
  "insufficiently_calibrated",
  "external_aggregate",
  "research_mode",
  "user_scenario",
]);
export type EstimateStatusValue = z.infer<typeof EstimateStatus>;

export const IndexId = z.enum([
  "evidence_pressure",
  "capability_pressure",
  "control_strength",
  "incident_pressure",
  "uncertainty",
  "agentic_infrastructure_risk",
  "attention",
]);
export type IndexIdValue = z.infer<typeof IndexId>;

export const AggregationMethod = z.enum([
  "unweighted_median",
  "linear_pool",
  "log_odds_pool",
  "trimmed_mean",
  "tier_weighted",
  "recency_weighted",
  "equal_weight_by_population",
]);
export type AggregationMethodValue = z.infer<typeof AggregationMethod>;

// ---------------------------------------------------------------------------
// Appendix A per-entity enumerations
// ---------------------------------------------------------------------------

export const DefinitionConsensus = z.enum(["consensus", "contested", "emerging"]);
export const RobotsStatus = z.enum(["allowed", "disallowed", "not_applicable", "unknown"]);
export const RetractionStatus = z.enum(["none", "retracted", "corrected", "disputed"]);
export const ForecastStatus = z.enum(["current", "superseded", "withdrawn"]);
export const ContaminationRisk = z.enum(["low", "moderate", "high", "unknown"]);
export const Saturation = z.enum(["none", "partial", "saturated", "unknown"]);
export const BenchmarkDirection = z.enum(["higher_is_more_capable", "lower_is_more_capable"]);
/** Three-step confidence used by benchmark results, scenario edges and model confidence. */
export const ConfidenceLabel = z.enum(["low", "moderate", "high"]);
export type ConfidenceLabelValue = z.infer<typeof ConfidenceLabel>;
export const DatePrecision = z.enum(["day", "month", "year", "unknown"]);
export const IncidentNovelty = z.enum(["routine", "notable", "novel"]);
export const ProbabilitySource = z.enum([
  "not_assigned",
  "external_forecast",
  "experimental_model",
  "expert_elicitation",
]);
export const Recoverability = z.enum(["high", "moderate", "low", "none", "unknown"]);
export const ScenarioEdgeRelation = z.enum([
  "enables",
  "amplifies",
  "prevents_response",
  "shares_prerequisite",
  "competes_with",
]);
export const SignalDirection = z.enum(["higher_raises_pressure", "higher_strengthens_control"]);
export const EvidenceStrength = z.enum(["none", "weak", "moderate", "strong"]);
export const InterventionCost = z.enum(["low", "moderate", "high", "very_high", "unknown"]);
export const TimeToDeploy = z.enum(["months", "1-2y", "3-5y", "5y+", "unknown"]);
export const EffectSize = z.enum(["unknown", "small", "moderate", "large"]);
export const InterventionCategory = z.enum([
  "technical",
  "organizational",
  "national",
  "international",
  "resilience",
]);
export const ActionAudience = z.enum([
  "individuals",
  "software_engineers",
  "ai_researchers",
  "laboratories",
  "policymakers",
  "funders",
  "educators",
  "nonprofits",
  "auditors_red_teams",
  "standards_bodies",
]);
export type ActionAudienceValue = z.infer<typeof ActionAudience>;
export const ActionEffort = z.enum(["low", "moderate", "high"]);

// ---------------------------------------------------------------------------
// Release / lab enumerations
// ---------------------------------------------------------------------------

/** Display rounding rule derived from the uncertainty label (build-spec §0.8). */
export const RoundingRule = z.enum(["nearest_5", "nearest_2", "nearest_1", "not_applicable"]);
export type RoundingRuleValue = z.infer<typeof RoundingRule>;

export const SensitivityKind = z.enum([
  "leave_one_source_out",
  "leave_one_survey_out",
  "alternative_weighting",
  "alternative_prior",
  "alternative_dependency",
  "optimistic_safeguards",
  "pessimistic_safeguards",
  "slower_capability",
  "faster_capability",
  "lower_exposure",
  "higher_exposure",
]);
export type SensitivityKindValue = z.infer<typeof SensitivityKind>;

/** Direction of a driver's contribution in "Why this number?" explanations. */

/** How a driver enters the published objects. */


export const SubmissionKind = z.enum(["source", "correction", "incident_reference"]);
export type SubmissionKindValue = z.infer<typeof SubmissionKind>;
export const SubmissionStatus = z.enum(["received"]);

export const CassandraSeverity = z.enum(["critical", "high", "medium", "low"]);
export type CassandraSeverityValue = z.infer<typeof CassandraSeverity>;
export const CassandraStatus = z.enum(["open", "resolved", "accepted_risk", "disputed"]);
export type CassandraStatusValue = z.infer<typeof CassandraStatus>;

// ---------------------------------------------------------------------------
// Constant tables
// ---------------------------------------------------------------------------

export { DERIVED_OUTCOME_SETS, HORIZONS, OUTCOMES, type DerivedOutcomeSetKey, type HorizonDefinition, type OutcomeDefinition } from "./constants";

export interface DriverFamilyDefinition {
  id: DriverFamilyCode;
  slug: DriverFamilySlugValue;
  label: string;
  description: string;
}

export const DRIVER_FAMILIES: readonly DriverFamilyDefinition[] = [
  {
    id: "D1",
    slug: "capability",
    label: "Capability",
    description:
      "Measured frontier capability on tracked benchmarks and task horizons, at the indicator level.",
  },
  {
    id: "D2",
    slug: "autonomy",
    label: "Autonomy",
    description:
      "How long and how independently systems act without human checkpoints in deployed settings.",
  },
  {
    id: "D3",
    slug: "access_exposure",
    label: "Access and exposure",
    description:
      "How widely capable systems are available and how many people and systems are exposed to them.",
  },
  {
    id: "D4",
    slug: "scalability",
    label: "Scalability",
    description:
      "Compute, cost and deployment trends that determine how quickly capabilities can be replicated at scale.",
  },
  {
    id: "D5",
    slug: "alignment_control",
    label: "Alignment and control",
    description:
      "Progress on making systems behave as intended and on tools for oversight, interpretability and shutdown.",
  },
  {
    id: "D6",
    slug: "security",
    label: "Security",
    description:
      "Protection of model weights, infrastructure and supply chains against theft, tampering and misuse, described at the safeguard level.",
  },
  {
    id: "D7",
    slug: "governance",
    label: "Governance",
    description:
      "Standards, regulation, institutional capacity and international agreements that constrain development and deployment.",
  },
  {
    id: "D8",
    slug: "incidents",
    label: "Incidents",
    description: "Frequency, severity and novelty of verified AI incidents and near misses.",
  },
  {
    id: "D9",
    slug: "race_dynamics",
    label: "Race dynamics",
    description:
      "Competitive pressure between developers and states that shortens timelines or weakens caution.",
  },
  {
    id: "D10",
    slug: "resilience",
    label: "Resilience",
    description:
      "Societal capacity to absorb, respond to and recover from AI-related shocks.",
  },
];

export interface SourceTierDefinition {
  tier: SourceTierValue;
  name: string;
  description: string;
  /** Default weight multiplier used by the indexes (build-spec §3.5). */
  default_multiplier: number;
  /** Whether evidence at this tier may, on its own, change a p(DOOM)-relevant estimate. */
  may_alter_pdoom: boolean;
  /** Whether corroboration from a higher tier is required before model use. */
  requires_corroboration: boolean;
  /** Highest model_use_status an item at this tier may reach. */
  max_model_use_status: ModelUseStatusValue;
  rules: readonly string[];
}

export const SOURCE_TIERS: readonly SourceTierDefinition[] = [
  {
    tier: 1,
    name: "Primary / authoritative",
    description:
      "Peer-reviewed research, official datasets, primary documents, government and regulatory texts, court records, standards and other authoritative primary material.",
    default_multiplier: 1.0,
    may_alter_pdoom: true,
    requires_corroboration: false,
    max_model_use_status: "used",
    rules: [
      "May be used directly as model input.",
      "A material change to a p(DOOM)-relevant estimate requires Tier 1 evidence or independent corroboration from at least two Tier 2 sources.",
      "Developer-authored primary documents keep the developer_self_report conflict label.",
    ],
  },
  {
    tier: 2,
    name: "Independent technical",
    description:
      "Independent technical analyses, evaluations, reproductions and expert reports by parties without a developer or funder conflict on the matter at hand.",
    default_multiplier: 0.9,
    may_alter_pdoom: true,
    requires_corroboration: false,
    max_model_use_status: "used",
    rules: [
      "May be used directly as model input at slightly reduced weight.",
      "Two independent Tier 2 sources together count as corroboration for a material change.",
    ],
  },
  {
    tier: 3,
    name: "High-quality journalism",
    description:
      "Reporting from outlets with editorial standards, corrections policies and named authors.",
    default_multiplier: 0.5,
    may_alter_pdoom: false,
    requires_corroboration: true,
    max_model_use_status: "used",
    rules: [
      "Creates candidate evidence: claims supported only by Tier 3 sources stay in status candidate.",
      "May be used at reduced weight once corroborated by a Tier 1 or Tier 2 source.",
      "Cannot on its own justify a material change to a p(DOOM)-relevant estimate.",
    ],
  },
  {
    tier: 4,
    name: "Commentary",
    description: "Opinion, blogs, newsletters, talks, podcasts and similar secondary commentary.",
    default_multiplier: 0.0,
    may_alter_pdoom: false,
    requires_corroboration: true,
    max_model_use_status: "informational",
    rules: [
      "Needs corroboration from Tier 1 or Tier 2 before any claim it supports can be used.",
      "Informational only: may be shown as context, never as model input.",
    ],
  },
  {
    tier: 5,
    name: "Unverified",
    description:
      "Social posts, anonymous material and anything whose provenance could not be verified.",
    default_multiplier: 0.0,
    may_alter_pdoom: false,
    requires_corroboration: true,
    max_model_use_status: "excluded",
    rules: [
      "Cannot alter p(DOOM) or any published index in any way.",
      "Always model_use_status excluded; may appear only in the review queue as unverified.",
    ],
  },
];

export interface EditorialRiskLevelDefinition {
  level: EditorialRiskLevelValue;
  label: string;
  description: string;
}

/** Plain-language editorial levels. These are judgments, not probabilities. */
export const EDITORIAL_RISK_LEVELS: readonly EditorialRiskLevelDefinition[] = [
  {
    level: "very_low",
    label: "Very low",
    description:
      "Tracked evidence and safeguards suggest the risk is small and well managed for the stated horizon.",
  },
  {
    level: "low",
    label: "Low",
    description:
      "Some pressure is visible, but safeguards and institutional capacity appear adequate for the stated horizon.",
  },
  {
    level: "guarded",
    label: "Guarded",
    description:
      "Capability and exposure are rising faster than control; the situation warrants sustained attention.",
  },
  {
    level: "elevated",
    label: "Elevated",
    description:
      "Several tracked indicators point the same way and safeguards are lagging; the editors judge the risk to be materially above baseline.",
  },
  {
    level: "high",
    label: "High",
    description:
      "Strong, corroborated evidence of rising pressure with weak or absent safeguards; the editors judge the risk to be serious for the stated horizon.",
  },
  {
    level: "severe_uncertainty",
    label: "Severe uncertainty",
    description:
      "Evidence exists but disagrees so widely, or is so model-dependent, that a single level would mislead; the range of credible views is shown instead.",
  },
  {
    level: "insufficient_evidence",
    label: "Insufficient evidence",
    description:
      "Too little verified evidence has been assembled to assign a level; nothing should be read into the absence.",
  },
];

export interface IndexDefinition {
  id: IndexIdValue;
  label: string;
  scale: string;
  is_probability: false;
  description: string;
}

/** Indexes are 0–100 scores and are never probabilities (build-spec §0.3). */
export const INDEX_DEFINITIONS: readonly IndexDefinition[] = [
  {
    id: "evidence_pressure",
    label: "Evidence Pressure Index",
    scale: "0-100, higher = more verified evidence pointing toward rising risk",
    is_probability: false,
    description:
      "Tier-weighted volume and direction of verified, recent evidence across the tracked driver families.",
  },
  {
    id: "capability_pressure",
    label: "Capability Pressure Index",
    scale: "0-100, higher = more pressure",
    is_probability: false,
    description:
      "How far measured frontier capabilities have moved along tracked signals such as task horizon, autonomy and breadth.",
  },
  {
    id: "control_strength",
    label: "Control Strength Index",
    scale: "0-100, higher = stronger control",
    is_probability: false,
    description:
      "Strength of tracked safeguards: alignment and control progress, security posture and governance capacity.",
  },
  {
    id: "incident_pressure",
    label: "Incident Pressure Index",
    scale: "0-100, higher = more pressure",
    is_probability: false,
    description:
      "Recency-weighted frequency, severity and relevance of verified AI incidents and near misses.",
  },
  {
    id: "uncertainty",
    label: "Uncertainty Score",
    scale: "0-100, higher = less certain",
    is_probability: false,
    description:
      "How much disagreement, sparse evidence and model fragility surround the published estimates.",
  },
  {
    id: "agentic_infrastructure_risk",
    label: "Agentic Infrastructure Risk Index",
    scale: "0-100, higher = more exposure",
    is_probability: false,
    description:
      "Category-level exposure from autonomous agents deployed with tools, credentials and limited oversight.",
  },
  {
    id: "attention",
    label: "Attention Index",
    scale: "0-100, higher = more attention",
    is_probability: false,
    description:
      "How much public, policy and research attention the topic is receiving. Attention is context, not a risk measure.",
  },
];

export interface AggregationMethodDefinition {
  method: AggregationMethodValue;
  label: string;
  description: string;
}

export const AGGREGATION_METHODS: readonly AggregationMethodDefinition[] = [
  {
    method: "unweighted_median",
    label: "Unweighted median",
    description: "Median of the compatible forecasts, each counted once.",
  },
  {
    method: "linear_pool",
    label: "Linear pool",
    description: "Arithmetic mean of the compatible forecasts.",
  },
  {
    method: "log_odds_pool",
    label: "Log-odds pool",
    description: "Mean taken on the log-odds scale and mapped back to a probability.",
  },
  {
    method: "trimmed_mean",
    label: "Trimmed mean",
    description: "Arithmetic mean after removing the most extreme forecasts on each side.",
  },
  {
    method: "tier_weighted",
    label: "Tier-weighted",
    description: "Weighted mean using the source-tier multipliers.",
  },
  {
    method: "recency_weighted",
    label: "Recency-weighted",
    description: "Weighted mean that discounts older forecasts.",
  },
  {
    method: "equal_weight_by_population",
    label: "Equal weight by population",
    description:
      "Each forecaster population is aggregated first, then populations are combined with equal weight.",
  },
];

export interface ConflictLabelDefinition {
  label: ConflictLabelValue;
  display: string;
  description: string;
}

export const CONFLICT_LABELS: readonly ConflictLabelDefinition[] = [
  {
    label: "developer_self_report",
    display: "Developer self-report",
    description: "Published by the organisation that builds or sells the system being described.",
  },
  {
    label: "advocacy_context",
    display: "Advocacy context",
    description: "Published by an organisation that campaigns for a particular policy outcome.",
  },
  {
    label: "government_policy_context",
    display: "Government policy context",
    description: "Published by a government or regulator with a policy position on the matter.",
  },
  {
    label: "commercial_interest",
    display: "Commercial interest",
    description: "The author or publisher has a financial stake in how the topic is perceived.",
  },
  {
    label: "funder_relationship",
    display: "Funder relationship",
    description: "The work was funded by a party with an interest in the conclusion.",
  },
  {
    label: "none_known",
    display: "No known conflict",
    description: "No conflict of interest was identified at review time.",
  },
];

/** Display rounding by uncertainty label (build-spec §0.8): points on the 0–100 % scale. */
export const ROUNDING_BY_UNCERTAINTY: Record<UncertaintyLabelValue, { points: number; rule: RoundingRuleValue }> = {
  extreme: { points: 5, rule: "nearest_5" },
  high: { points: 2, rule: "nearest_2" },
  moderate: { points: 1, rule: "nearest_1" },
  low: { points: 1, rule: "nearest_1" },
};

// ---------------------------------------------------------------------------
// Brand (build-spec name rules). Rendered text uses BRAND, never a literal.
// ---------------------------------------------------------------------------

export { AUTHOR, BRAND, BRAND_EXPANDED, PUBLIC_LABEL, TAGLINES } from "./constants";
