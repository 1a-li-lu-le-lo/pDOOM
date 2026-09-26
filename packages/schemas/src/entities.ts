// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * One strict zod object per snapshot entity (build-spec Appendix A), plus the
 * shared verification record and the snapshot manifest.
 *
 * Field names follow Appendix A exactly. Cross-field invariants that JSON Schema
 * cannot express (verification ⇒ model use, custom horizon ⇒ note, signal ⇒
 * family, quantile ordering) are enforced with refinements; Go re-implements them
 * in internal/snapshot.Validate.
 */
import { z } from "zod";
import {
  ActionId,
  BenchmarkId,
  BenchmarkResultId,
  ClaimId,
  DateString,
  DateTimeString,
  DefinitionId,
  DriverFamilyId,
  ForecastId,
  GenericId,
  HttpUrl,
  IncidentId,
  InterventionId,
  ModelSpecId,
  NonEmptyString,
  NonNegativeInt,
  OrganizationId,
  PartialQuantilesSchema,
  PositiveInt,
  Probability,
  ScenarioEdgeId,
  ScenarioId,
  Sha256Hex,
  SignalId,
  SnapshotId,
  SourceId,
  TriQuantilesSchema,
  UnitInterval,
  checkQuantilesMonotonic,
} from "./common.js";
import {
  ActionAudience,
  ActionEffort,
  AggregationMethod,
  BenchmarkDirection,
  ClaimEvidenceType,
  ClaimStatus,
  ConfidenceLabel,
  ConflictLabel,
  ContaminationRisk,
  DatePrecision,
  DefinitionConsensus,
  DriverFamily,
  EditorialRiskLevel,
  EffectSize,
  EvidenceLevel,
  EvidenceStrength,
  ForecastHorizon,
  ForecastPopulation,
  ForecastStatus,
  HumanReviewStatus,
  IncidentCause,
  IncidentHarm,
  IncidentNovelty,
  IncidentSeverity,
  InterventionCategory,
  InterventionCost,
  MODEL_USE_VERIFICATION_STATUSES,
  ModelUseStatus,
  ObservationKind,
  Outcome,
  PDoomOutcome,
  PDoomRelevance,
  ProbabilitySource,
  Recoverability,
  RetractionStatus,
  RobotsStatus,
  Saturation,
  ScenarioEdgeRelation,
  SignalDirection,
  SourceTier,
  SourceType,
  TimeToDeploy,
  UncertaintyLabel,
  VerificationStatus,
  Horizon,
} from "./enums.js";

// ---------------------------------------------------------------------------
// Shared fragments
// ---------------------------------------------------------------------------

export const VerificationSchema = z
  .strictObject({
    status: VerificationStatus,
    checked_at: DateString,
    method: z.string(),
    note: z.string(),
  })
  .describe("Verification record: how and when the item was checked against its canonical source");
export type Verification = z.infer<typeof VerificationSchema>;

/** Fields present on every data-bearing entity (build-spec §3.4). */
const reviewFields = {
  verification: VerificationSchema,
  human_review_status: HumanReviewStatus,
  model_use_status: ModelUseStatus,
} as const;

type Reviewed = {
  verification: Verification;
  model_use_status: z.infer<typeof ModelUseStatus>;
};

/** Only verified_fetch / verified_search items may be eligible or used (build-spec §3.3). */
export function checkModelUseEligibility(value: Reviewed, ctx: z.RefinementCtx): void {
  const wantsModelUse = value.model_use_status === "eligible" || value.model_use_status === "used";
  if (wantsModelUse && !MODEL_USE_VERIFICATION_STATUSES.includes(value.verification.status)) {
    ctx.addIssue({
      code: "custom",
      message: `model_use_status "${value.model_use_status}" requires verification.status verified_fetch or verified_search`,
      path: ["model_use_status"],
    });
  }
}

const StringArray = z.array(z.string());
const SourceIdArray = z.array(SourceId);

// ---------------------------------------------------------------------------
// definition
// ---------------------------------------------------------------------------

export const DefinitionEntrySchema = z.strictObject({
  text: z.string(),
  source_id: SourceId.nullable(),
  attribution: z.string(),
  note: z.string(),
});

export const DefinitionSchema = z
  .strictObject({
    id: DefinitionId,
    term: NonEmptyString,
    short_definition: z.string(),
    definitions: z.array(DefinitionEntrySchema),
    consensus: DefinitionConsensus,
    related_ids: StringArray,
    see_also_urls: z.array(HttpUrl),
    ...reviewFields,
  })
  .superRefine(checkModelUseEligibility)
  .describe("Glossary definition with attributed variants");
export type Definition = z.infer<typeof DefinitionSchema>;

// ---------------------------------------------------------------------------
// source
// ---------------------------------------------------------------------------

export const SourceSchema = z
  .strictObject({
    id: SourceId,
    canonical_url: HttpUrl,
    title: NonEmptyString,
    publisher: z.string(),
    authors: StringArray,
    date_published: DateString.nullable(),
    date_updated: DateString.nullable(),
    date_retrieved: DateString,
    source_tier: SourceTier,
    source_type: SourceType,
    jurisdiction: z.string().nullable(),
    topic: StringArray,
    claim_ids: z.array(ClaimId),
    evidence_summary: z.string(),
    counterevidence: z.string().nullable(),
    methodology: z.string().nullable(),
    sample: z.string().nullable(),
    limitations: z.string().nullable(),
    conflicts: z.array(ConflictLabel),
    license: z.string().nullable(),
    robots_status: RobotsStatus,
    content_hash: z.string().nullable(),
    archive_reference: z.string().nullable(),
    language: NonEmptyString,
    translation: z.string().nullable(),
    duplicate_group: z.string().nullable(),
    retraction_status: RetractionStatus,
    correction_status: z.string().nullable(),
    citation: z.string(),
    ...reviewFields,
  })
  .superRefine((value, ctx) => {
    checkModelUseEligibility(value, ctx);
    if (value.source_tier === 5 && value.model_use_status !== "excluded") {
      ctx.addIssue({
        code: "custom",
        message: "tier 5 sources must be model_use_status excluded",
        path: ["model_use_status"],
      });
    }
    if (
      value.source_tier === 4 &&
      (value.model_use_status === "eligible" || value.model_use_status === "used")
    ) {
      ctx.addIssue({
        code: "custom",
        message: "tier 4 sources are informational at most",
        path: ["model_use_status"],
      });
    }
  })
  .describe("Source record (build-spec Appendix A)");
export type Source = z.infer<typeof SourceSchema>;

// ---------------------------------------------------------------------------
// claim
// ---------------------------------------------------------------------------

export const ClaimSchema = z
  .strictObject({
    id: ClaimId,
    text: NonEmptyString,
    subject: z.string(),
    predicate: z.string(),
    object: z.string(),
    date: DateString.nullable(),
    horizon: z.string().nullable(),
    geography: z.string().nullable(),
    model_name: z.string().nullable(),
    model_version: z.string().nullable(),
    source_id: SourceId,
    evidence_type: ClaimEvidenceType,
    quantitative_value: z.number().nullable(),
    unit: z.string().nullable(),
    uncertainty: z.string().nullable(),
    direct_quote_pointer: z
      .string()
      .nullable()
      .describe("Section / page / figure locator, never the full quoted text"),
    context: z.string(),
    corroboration_ids: z.array(ClaimId),
    contradiction_ids: z.array(ClaimId),
    relevance: PDoomRelevance,
    status: ClaimStatus,
    ...reviewFields,
  })
  .superRefine(checkModelUseEligibility)
  .describe("Atomic claim extracted from a source");
export type Claim = z.infer<typeof ClaimSchema>;

// ---------------------------------------------------------------------------
// forecast
// ---------------------------------------------------------------------------

export const ForecastSchema = z
  .strictObject({
    id: ForecastId,
    forecaster_or_survey: NonEmptyString,
    source_id: SourceId,
    date: DateString,
    population: ForecastPopulation,
    sample_size: NonNegativeInt.nullable(),
    expertise: z.string(),
    question_wording_original: z.string(),
    paraphrase: z.boolean(),
    outcome_set: z.array(Outcome).min(1),
    horizon: ForecastHorizon,
    horizon_note: z.string().nullable(),
    horizon_end_year: z.int().nullable(),
    conditions: z.string(),
    mean: Probability.nullable(),
    median: Probability.nullable(),
    quantiles: PartialQuantilesSchema,
    response_rate: UnitInterval.nullable(),
    selection_effects: z.string().nullable(),
    framing_effects: z.string().nullable(),
    calibration: z.string().nullable(),
    group_id: z.string().nullable(),
    transformation_note: z.string().nullable(),
    status: ForecastStatus,
    ...reviewFields,
  })
  .superRefine((value, ctx) => {
    checkModelUseEligibility(value, ctx);
    checkQuantilesMonotonic(value.quantiles, ctx);
    if (value.horizon === "custom" && !value.horizon_note) {
      ctx.addIssue({
        code: "custom",
        message: 'horizon "custom" requires a horizon_note',
        path: ["horizon_note"],
      });
    }
    if (
      value.median !== null &&
      value.quantiles.p50 !== undefined &&
      Math.abs(value.median - value.quantiles.p50) > 1e-9
    ) {
      ctx.addIssue({
        code: "custom",
        message: "median and quantiles.p50 must agree when both are present",
        path: ["median"],
      });
    }
  })
  .describe("Forecast record with original wording, outcome set, horizon and conditioning");
export type Forecast = z.infer<typeof ForecastSchema>;

// ---------------------------------------------------------------------------
// benchmark, benchmark_result
// ---------------------------------------------------------------------------

export const BenchmarkSchema = z
  .strictObject({
    id: BenchmarkId,
    name: NonEmptyString,
    maintainer: z.string(),
    version: z.string().nullable(),
    url: HttpUrl,
    tasks: z.string(),
    contamination_risk: ContaminationRisk,
    saturation: Saturation,
    scaffold: z.string().nullable(),
    model_access: z.string().nullable(),
    unit: z.string(),
    direction: BenchmarkDirection,
    limitations: z.string(),
    pdoom_relevance: PDoomRelevance,
    weight_note: z.string().nullable(),
    source_ids: SourceIdArray,
    ...reviewFields,
  })
  .superRefine(checkModelUseEligibility)
  .describe("Capability benchmark");
export type Benchmark = z.infer<typeof BenchmarkSchema>;

export const BenchmarkResultSchema = z
  .strictObject({
    id: BenchmarkResultId,
    benchmark_id: BenchmarkId,
    model_name: NonEmptyString,
    model_developer: z.string(),
    date: DateString,
    value: z.number(),
    unit: z.string(),
    ci_low: z.number().nullable(),
    ci_high: z.number().nullable(),
    scaffold: z.string().nullable(),
    confidence: ConfidenceLabel,
    note: z.string().nullable(),
    source_ids: SourceIdArray,
    ...reviewFields,
  })
  .superRefine((value, ctx) => {
    checkModelUseEligibility(value, ctx);
    if (value.ci_low !== null && value.ci_high !== null && value.ci_low > value.ci_high) {
      ctx.addIssue({ code: "custom", message: "ci_low must not exceed ci_high", path: ["ci_low"] });
    }
  })
  .describe("A single benchmark measurement for one model");
export type BenchmarkResult = z.infer<typeof BenchmarkResultSchema>;

// ---------------------------------------------------------------------------
// incident
// ---------------------------------------------------------------------------

export const IncidentExternalIdsSchema = z.strictObject({
  aiid: z.string().optional(),
  oecd_aim: z.string().optional(),
  mit_tracker: z.string().optional(),
  cve: z.string().optional(),
  docket: z.string().optional(),
  other: z.string().optional(),
});
export type IncidentExternalIds = z.infer<typeof IncidentExternalIdsSchema>;

export const IncidentSchema = z
  .strictObject({
    id: IncidentId,
    title: NonEmptyString,
    date: DateString.nullable(),
    date_precision: DatePrecision,
    external_ids: IncidentExternalIdsSchema,
    summary: z.string().describe("Non-graphic, non-operational summary"),
    cause: z.array(IncidentCause),
    harm: z.array(IncidentHarm),
    severity: IncidentSeverity,
    pdoom_relevance: PDoomRelevance,
    evidence_level: EvidenceLevel,
    systems_involved: StringArray,
    jurisdiction: z.string().nullable(),
    near_miss: z.boolean(),
    novelty: IncidentNovelty,
    exposure_note: z.string().nullable(),
    source_ids: SourceIdArray,
    ...reviewFields,
  })
  .superRefine(checkModelUseEligibility)
  .describe("Verified AI incident or near miss, described at category level");
export type Incident = z.infer<typeof IncidentSchema>;

// ---------------------------------------------------------------------------
// scenario, scenario_edge
// ---------------------------------------------------------------------------

export const ScenarioSchema = z
  .strictObject({
    id: ScenarioId,
    name: NonEmptyString,
    outcome_set: z.array(Outcome).min(1),
    description: z.string(),
    prerequisites: StringArray,
    early_indicators: StringArray,
    counterindicators: StringArray,
    capability_thresholds: StringArray,
    exposure: z.string(),
    control_failures: StringArray,
    human_contributions: StringArray,
    ai_contributions: StringArray,
    dependencies: z.array(ScenarioId),
    time_horizon_note: z.string(),
    probability_source: ProbabilitySource,
    uncertainty: UncertaintyLabel,
    intervention_ids: z.array(InterventionId),
    recoverability: Recoverability,
    evidence_summary: z.string(),
    source_ids: SourceIdArray,
    open_questions: StringArray,
    content_safety_note: z.string().nullable(),
    ...reviewFields,
  })
  .superRefine(checkModelUseEligibility)
  .describe("Category-level pathway scenario (S1..S18)");
export type Scenario = z.infer<typeof ScenarioSchema>;

export const ScenarioEdgeSchema = z
  .strictObject({
    id: ScenarioEdgeId,
    from_id: ScenarioId,
    to_id: ScenarioId,
    relation: ScenarioEdgeRelation,
    confidence: ConfidenceLabel,
    rationale: z.string(),
    source_ids: SourceIdArray,
  })
  .superRefine((value, ctx) => {
    if (value.id !== `se-${value.from_id}-${value.to_id}`) {
      ctx.addIssue({
        code: "custom",
        message: "scenario_edge id must be se-<from_id>-<to_id>",
        path: ["id"],
      });
    }
    if (value.from_id === value.to_id) {
      ctx.addIssue({ code: "custom", message: "an edge cannot point at itself", path: ["to_id"] });
    }
  })
  .describe("Directed relation between two scenarios");
export type ScenarioEdge = z.infer<typeof ScenarioEdgeSchema>;

// ---------------------------------------------------------------------------
// driver, driver_observation
// ---------------------------------------------------------------------------

export const DriverSignalSchema = z.strictObject({
  signal_id: SignalId,
  name: NonEmptyString,
  description: z.string(),
  direction: SignalDirection,
  normalization: z.string().describe("How the raw value maps to 0..1"),
  raw_unit: z.string().nullable(),
  preferred_source_types: StringArray,
  observation_vs_judgment: z.string(),
});
export type DriverSignal = z.infer<typeof DriverSignalSchema>;

export const DriverSchema = z
  .strictObject({
    id: DriverFamilyId,
    name: NonEmptyString,
    description: z.string(),
    signals: z.array(DriverSignalSchema),
  })
  .superRefine((value, ctx) => {
    value.signals.forEach((signal, index) => {
      if (!signal.signal_id.startsWith(`${value.id}.`)) {
        ctx.addIssue({
          code: "custom",
          message: `signal_id must start with "${value.id}."`,
          path: ["signals", index, "signal_id"],
        });
      }
    });
  })
  .describe("Driver family (D1..D10) and its tracked signals");
export type Driver = z.infer<typeof DriverSchema>;

export const DriverObservationSchema = z
  .strictObject({
    id: GenericId,
    signal_id: SignalId,
    family: DriverFamily,
    value_normalized: UnitInterval,
    raw_value: z.number().nullable(),
    raw_unit: z.string().nullable(),
    confidence: UnitInterval,
    observation_kind: ObservationKind,
    as_of: DateString,
    rationale: z.string(),
    counterevidence: z.string().nullable(),
    source_ids: SourceIdArray,
    ...reviewFields,
  })
  .superRefine((value, ctx) => {
    checkModelUseEligibility(value, ctx);
    if (!value.signal_id.startsWith(`${value.family}.`)) {
      ctx.addIssue({
        code: "custom",
        message: "signal_id family prefix must equal family",
        path: ["signal_id"],
      });
    }
  })
  .describe("One normalised observation or judgment for a driver signal");
export type DriverObservation = z.infer<typeof DriverObservationSchema>;

// ---------------------------------------------------------------------------
// intervention
// ---------------------------------------------------------------------------

export const UserActionSchema = z.strictObject({
  audience: z.string(),
  action: z.string(),
});

export const InterventionSchema = z
  .strictObject({
    id: InterventionId,
    name: NonEmptyString,
    target_scenario_ids: z.array(ScenarioId),
    mechanism: z.string(),
    evidence_summary: z.string(),
    evidence_strength: EvidenceStrength,
    cost: InterventionCost,
    time_to_deploy: TimeToDeploy,
    effect_size: EffectSize.describe("Qualitative only"),
    uncertainty: UncertaintyLabel,
    possible_failure: z.string(),
    possible_backfire: z.string(),
    owner_types: StringArray,
    user_actions: z.array(UserActionSchema),
    category: InterventionCategory,
    source_ids: SourceIdArray,
    ...reviewFields,
  })
  .superRefine(checkModelUseEligibility)
  .describe("Category-level safeguard or intervention");
export type Intervention = z.infer<typeof InterventionSchema>;

// ---------------------------------------------------------------------------
// organization
// ---------------------------------------------------------------------------

export const OrganizationSchema = z
  .strictObject({
    id: OrganizationId,
    name: NonEmptyString,
    url: HttpUrl,
    mission: z.string(),
    legal_status: z.string(),
    jurisdiction: z.string(),
    focus: StringArray,
    programs: StringArray,
    open_outputs: StringArray,
    funding_disclosure: z.string(),
    conflicts: StringArray,
    evidence_of_impact: z.string(),
    ways_to_help: StringArray,
    inclusion_criteria_met: StringArray,
    last_verified: DateString,
    source_ids: SourceIdArray,
    ...reviewFields,
  })
  .superRefine(checkModelUseEligibility)
  .describe("Organisation listed on the Act page");
export type Organization = z.infer<typeof OrganizationSchema>;

// ---------------------------------------------------------------------------
// action
// ---------------------------------------------------------------------------

export const ActionResourceSchema = z.strictObject({
  title: z.string(),
  url: HttpUrl,
  source_id: SourceId.nullable(),
});

export const ActionSchema = z
  .strictObject({
    id: ActionId,
    audience: ActionAudience,
    title: NonEmptyString,
    description: z.string(),
    related_intervention_ids: z.array(InterventionId),
    resources: z.array(ActionResourceSchema),
    effort: ActionEffort,
    ...reviewFields,
  })
  .superRefine((value, ctx) => {
    checkModelUseEligibility(value, ctx);
    if (!value.id.startsWith(`act-${value.audience}-`)) {
      ctx.addIssue({
        code: "custom",
        message: "action id must be act-<audience>-<slug>",
        path: ["id"],
      });
    }
  })
  .describe("Concrete action for a named audience");
export type Action = z.infer<typeof ActionSchema>;

// ---------------------------------------------------------------------------
// model_spec
// ---------------------------------------------------------------------------

/** Per-signal weight, bounded by `weight_bounds` (default [0, 0.35]). */
const SignalWeights = z.record(SignalId, z.number().min(0).max(1));
const NumericConstants = z.record(z.string(), z.number());

export const IndexWeightsSchema = z.strictObject({
  capability_pressure: SignalWeights,
  control_strength: SignalWeights,
  incident_pressure: NumericConstants,
  evidence_pressure: NumericConstants,
  uncertainty: NumericConstants,
  agentic_infrastructure_risk: SignalWeights,
});
export type IndexWeights = z.infer<typeof IndexWeightsSchema>;

export const TierMultipliersSchema = z.strictObject({
  "1": UnitInterval,
  "2": UnitInterval,
  "3": UnitInterval,
  "4": UnitInterval,
  "5": UnitInterval,
});
export type TierMultipliers = z.infer<typeof TierMultipliersSchema>;

export const IncidentScoringSchema = z.strictObject({
  severity_weights: z.record(IncidentSeverity, z.number()),
  relevance_weights: z.record(PDoomRelevance, z.number()),
  evidence_weights: z.record(EvidenceLevel, z.number()),
  recency_half_life_days: PositiveInt,
  squash_k: z.number(),
});
export type IncidentScoring = z.infer<typeof IncidentScoringSchema>;

export const EditorialRuleSchema = z.strictObject({
  level: EditorialRiskLevel,
  when: z.string(),
});

export const RoundingRulesSchema = z.strictObject({
  extreme: PositiveInt,
  high: PositiveInt,
  moderate: PositiveInt,
  low: PositiveInt,
});

/** Factor / outcome parameters for one horizon of the experimental causal model. */
export const CausalHorizonSpecSchema = z.strictObject({
  A: TriQuantilesSchema.describe("Capability / autonomy pressure factor"),
  C: TriQuantilesSchema.describe("Control strength factor"),
  E: TriQuantilesSchema.describe("Exposure factor"),
  F: TriQuantilesSchema.describe("Fragility / resilience factor"),
  O: z.record(PDoomOutcome, TriQuantilesSchema).describe("Per-outcome base rates, O3..O8"),
});
export type CausalHorizonSpec = z.infer<typeof CausalHorizonSpecSchema>;

export const ExperimentalCausalSpecSchema = z
  .strictObject({
    version: NonEmptyString,
    seed: NonNegativeInt.max(4294967295),
    samples: PositiveInt,
    common_factor_loading: UnitInterval,
    horizons: z.partialRecord(Horizon, CausalHorizonSpecSchema),
    rationale: z.strictObject({
      A: z.string(),
      C: z.string(),
      E: z.string(),
      F: z.string(),
      O: z.string(),
      dependence: z.string(),
    }),
    source_ids: SourceIdArray,
  })
  .describe("Parameters of the experimental causal model (research mode only)");
export type ExperimentalCausalSpec = z.infer<typeof ExperimentalCausalSpecSchema>;

export const ModelSpecSchema = z
  .strictObject({
    id: ModelSpecId,
    index_weights: IndexWeightsSchema,
    weight_bounds: z.tuple([z.number().min(0), z.number().max(1)]),
    tier_multipliers: TierMultipliersSchema,
    incident_scoring: IncidentScoringSchema,
    editorial_rules: z.array(EditorialRuleSchema),
    rounding_rules: RoundingRulesSchema,
    aggregation_methods: z.array(AggregationMethod),
    experimental_causal: ExperimentalCausalSpecSchema,
  })
  .superRefine((value, ctx) => {
    const [low, high] = value.weight_bounds;
    if (low > high) {
      ctx.addIssue({ code: "custom", message: "weight_bounds must be ordered", path: ["weight_bounds"] });
    }
    for (const indexId of ["capability_pressure", "control_strength", "agentic_infrastructure_risk"] as const) {
      for (const [signalId, weight] of Object.entries(value.index_weights[indexId])) {
        if (weight < low || weight > high) {
          ctx.addIssue({
            code: "custom",
            message: `weight ${weight} outside weight_bounds [${low}, ${high}]`,
            path: ["index_weights", indexId, signalId],
          });
        }
      }
    }
  })
  .describe("Model specification: weights, constants, rules and experimental parameters");
export type ModelSpec = z.infer<typeof ModelSpecSchema>;

// ---------------------------------------------------------------------------
// snapshot manifest
// ---------------------------------------------------------------------------

export const SnapshotFileEntrySchema = z.strictObject({
  path: NonEmptyString,
  sha256: Sha256Hex,
  count: NonNegativeInt,
});
export type SnapshotFileEntry = z.infer<typeof SnapshotFileEntrySchema>;

export const SnapshotManifestSchema = z
  .strictObject({
    snapshot_id: SnapshotId,
    created_at: DateTimeString,
    source_cutoff: DateString,
    baseline_snapshot_id: SnapshotId.nullable(),
    files: z.array(SnapshotFileEntrySchema),
    notes: z.string(),
  })
  .describe("Snapshot manifest listing every entity file with its sha256 and item count");
export type SnapshotManifest = z.infer<typeof SnapshotManifestSchema>;
