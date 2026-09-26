// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Objects written by the model run and promoted into a release directory
 * (build-spec §3.6): estimates, index values, aggregation results, sensitivity
 * runs, the PROBABILITY CHANGE POLICY delta record, the release manifest,
 * reviewer approvals and the "Why this number?" driver explanations.
 */
import { z } from "zod";
import {
  Base64String,
  DateString,
  DateTimeString,
  EstimateId,
  GitCommit,
  ModelVersion,
  NonEmptyString,
  NonNegativeInt,
  IndexScore,
  Probability,
  QuantilesSchema,
  ReleaseId,
  Sha256Hex,
  SignalId,
  SnapshotId,
  SourceId,
  UnitInterval,
  checkQuantilesMonotonic,
} from "./common";
import {
  AggregationMethod,
  ApprovalStatus,
  ConfidenceLabel,
  ContributionDirection,
  DriverFamily,
  EstimateStatus,
  ForecastHorizon,
  Horizon,
  IndexId,
  ModelRole,
  Outcome,
  RoundingRule,
  SensitivityKind,
  SourceTier,
  UncertaintyLabel,
} from "./enums";
import { TierMultipliersSchema } from "./entities";

// ---------------------------------------------------------------------------
// Estimate
// ---------------------------------------------------------------------------

export const EstimateDisplaySchema = z.strictObject({
  central: NonEmptyString.describe('Rounded central value or "Insufficiently calibrated"'),
  interval: z.string().nullable().describe("Rounded p05–p95 interval, null when no probability is published"),
  note: z.string(),
});
export type EstimateDisplay = z.infer<typeof EstimateDisplaySchema>;

export const SourceCoverageSchema = z.strictObject({
  forecast_count: NonNegativeInt,
  population_count: NonNegativeInt,
  source_ids: z.array(SourceId),
});
export type SourceCoverage = z.infer<typeof SourceCoverageSchema>;

/** Pointer to the same estimate in the previous release. */
export const PreviousEstimateSchema = z.strictObject({
  release_id: ReleaseId,
  estimate_id: EstimateId,
  quantiles: QuantilesSchema.nullable(),
  mean: Probability.nullable(),
});
export type PreviousEstimate = z.infer<typeof PreviousEstimateSchema>;

export const EstimateSchema = z
  .strictObject({
    estimate_id: EstimateId,
    producer: ModelVersion,
    status: EstimateStatus,
    outcome_set: z.array(Outcome).min(1),
    outcome_label: NonEmptyString,
    horizon: Horizon,
    conditioning: z.string(),
    forecast_origin_date: DateString,
    last_evidence_date: DateString,
    quantiles: QuantilesSchema.nullable(),
    mean: Probability.nullable(),
    disagreement: UncertaintyLabel,
    uncertainty: UncertaintyLabel,
    model_confidence: ConfidenceLabel,
    source_coverage: SourceCoverageSchema,
    previous: PreviousEstimateSchema.nullable(),
    reason_for_change: z.string(),
    rounding_rule: RoundingRule,
    display: EstimateDisplaySchema,
    assumptions: z.array(z.string()),
    method_ref: NonEmptyString,
  })
  .superRefine((value, ctx) => {
    checkQuantilesMonotonic(value.quantiles, ctx);
    if (value.status === "insufficiently_calibrated") {
      if (value.quantiles !== null) {
        ctx.addIssue({
          code: "custom",
          message: "an insufficiently_calibrated estimate publishes no quantiles",
          path: ["quantiles"],
        });
      }
      if (value.mean !== null) {
        ctx.addIssue({
          code: "custom",
          message: "an insufficiently_calibrated estimate publishes no mean",
          path: ["mean"],
        });
      }
    }
  })
  .describe("Published estimate with outcome, horizon, conditioning, interval and status");
export type Estimate = z.infer<typeof EstimateSchema>;

// ---------------------------------------------------------------------------
// IndexValue
// ---------------------------------------------------------------------------

export const IndexComponentSchema = z.strictObject({
  signal_id: SignalId,
  weight: z.number(),
  value_normalized: UnitInterval,
  tier: SourceTier,
  contribution: z.number(),
});
export type IndexComponent = z.infer<typeof IndexComponentSchema>;

export const IndexBaselineSchema = z.strictObject({
  snapshot_id: SnapshotId,
  value: IndexScore.nullable(),
});

export const IndexValueSchema = z
  .strictObject({
    index_id: IndexId,
    value: IndexScore.nullable(),
    label: NonEmptyString,
    scale: NonEmptyString,
    is_probability: z.literal(false),
    baseline: IndexBaselineSchema.nullable(),
    components: z.array(IndexComponentSchema),
    coverage: UnitInterval,
    as_of: DateString,
    method_ref: NonEmptyString,
    note: z.string(),
  })
  .describe("Index value on the 0–100 scale; never a probability");
export type IndexValue = z.infer<typeof IndexValueSchema>;

// ---------------------------------------------------------------------------
// AggregationResult
// ---------------------------------------------------------------------------

export const AggregationResultSchema = z
  .strictObject({
    group_id: NonEmptyString,
    outcome_set: z.array(Outcome).min(1),
    horizon: ForecastHorizon,
    method: AggregationMethod,
    value: Probability,
    quantiles: QuantilesSchema.optional(),
    n: NonNegativeInt,
    weights_note: z.string(),
    wording_note: z.string(),
    source_ids: z.array(SourceId),
  })
  .superRefine((value, ctx) => checkQuantilesMonotonic(value.quantiles, ctx))
  .describe("Aggregate of one compatibility group under one method");
export type AggregationResult = z.infer<typeof AggregationResultSchema>;

// ---------------------------------------------------------------------------
// SensitivityRun
// ---------------------------------------------------------------------------

export const SensitivityRunSchema = z
  .strictObject({
    id: NonEmptyString,
    kind: SensitivityKind,
    target_estimate_id: EstimateId,
    baseline_p50: Probability,
    perturbed_p50: Probability,
    delta: z.number().min(-1).max(1).describe("perturbed_p50 - baseline_p50"),
    description: z.string(),
  })
  .superRefine((value, ctx) => {
    if (Math.abs(value.perturbed_p50 - value.baseline_p50 - value.delta) > 1e-6) {
      ctx.addIssue({
        code: "custom",
        message: "delta must equal perturbed_p50 - baseline_p50",
        path: ["delta"],
      });
    }
  })
  .describe("One perturbation of the model and its effect on a target estimate");
export type SensitivityRun = z.infer<typeof SensitivityRunSchema>;

/** Summary of all sensitivity runs; reused by the delta record and the manifest. */
export const SensitivitySummarySchema = z.strictObject({
  run_count: NonNegativeInt,
  max_abs_delta: z.number().min(0).max(1).nullable(),
  most_sensitive_kind: SensitivityKind.nullable(),
  note: z.string(),
});
export type SensitivitySummary = z.infer<typeof SensitivitySummarySchema>;

// ---------------------------------------------------------------------------
// DeltaRecord (PROBABILITY CHANGE POLICY)
// ---------------------------------------------------------------------------

/** One side of a probability change: the estimate as it stood in a release or candidate. */
export const DeltaEstimatePointSchema = z.strictObject({
  artifact_id: NonEmptyString.describe("Release id (previous) or candidate id (new)"),
  estimate_id: EstimateId,
  status: EstimateStatus,
  p50: Probability.nullable(),
});
export type DeltaEstimatePoint = z.infer<typeof DeltaEstimatePointSchema>;

export const ApprovalSummarySchema = z.strictObject({
  status: ApprovalStatus,
  required_approvals: NonNegativeInt,
  received_approvals: NonNegativeInt,
  approved_at: DateTimeString.nullable(),
});
export type ApprovalSummary = z.infer<typeof ApprovalSummarySchema>;

export const DeltaRecordSchema = z
  .strictObject({
    previous_estimate: DeltaEstimatePointSchema.nullable(),
    new_candidate: DeltaEstimatePointSchema.nullable(),
    absolute_change: z.number().min(-1).max(1).nullable(),
    relative_change: z.number().nullable(),
    affected_horizons: z.array(Horizon),
    affected_outcomes: z.array(Outcome),
    sources_added: z.array(SourceId),
    sources_removed: z.array(SourceId),
    model_changes: z.array(z.string()),
    weight_changes: z.array(z.string()),
    data_corrections: z.array(z.string()),
    sensitivity_summary: SensitivitySummarySchema.nullable(),
    heightened_review_triggers: z.array(z.string()),
    reviewer: z.string().nullable(),
    approval: ApprovalSummarySchema.nullable(),
    release: ReleaseId.nullable(),
  })
  .describe("PROBABILITY CHANGE POLICY record versus the previous release (null fields on first release)");
export type DeltaRecord = z.infer<typeof DeltaRecordSchema>;

// ---------------------------------------------------------------------------
// ReleaseManifest (MODEL RELEASE ARTIFACT)
// ---------------------------------------------------------------------------

export const OutcomeDefinitionSummarySchema = z.strictObject({
  label: NonEmptyString,
  outcome_set: z.array(Outcome).min(1),
  text: z.string(),
  definition_ids: z.array(z.string()),
});

export const PriorsSummarySchema = z.strictObject({
  description: z.string(),
  model_spec_id: NonEmptyString,
  experimental_causal_version: z.string().nullable(),
  method_ref: NonEmptyString,
});

export const WeightsSummarySchema = z.strictObject({
  description: z.string(),
  tier_multipliers: TierMultipliersSchema,
  weight_bounds: z.tuple([z.number(), z.number()]),
  method_ref: NonEmptyString,
});

export const DependenciesSummarySchema = z.strictObject({
  description: z.string(),
  common_factor_loading: UnitInterval.nullable(),
  method_ref: NonEmptyString,
});

export const EstimateSummarySchema = z.strictObject({
  estimate_id: EstimateId,
  status: EstimateStatus,
  outcome_set: z.array(Outcome).min(1),
  horizon: Horizon,
  p50: Probability.nullable(),
  display_central: NonEmptyString,
});

export const IntervalSummarySchema = z.strictObject({
  estimate_id: EstimateId,
  p05: Probability.nullable(),
  p95: Probability.nullable(),
  display_interval: z.string().nullable(),
  uncertainty: UncertaintyLabel,
});

export const ExternalForecastsSummarySchema = z.strictObject({
  forecast_count: NonNegativeInt,
  group_count: NonNegativeInt,
  population_count: NonNegativeInt,
  source_count: NonNegativeInt,
  note: z.string(),
});

export const ReviewerSchema = z.strictObject({
  reviewer_id: NonEmptyString,
  role: z.string(),
  conflicts_declared: z.array(z.string()),
});
export type Reviewer = z.infer<typeof ReviewerSchema>;

export const ManifestSignatureSchema = z.strictObject({
  algorithm: z.literal("ed25519"),
  key_id: NonEmptyString,
  signed_at: DateTimeString,
  signature_base64: Base64String,
});
export type ManifestSignature = z.infer<typeof ManifestSignatureSchema>;

export const ReleaseManifestSchema = z
  .strictObject({
    release_id: ReleaseId,
    model_versions: z.array(ModelVersion).min(1),
    code_commit: GitCommit,
    data_snapshot: SnapshotId,
    source_cutoff: DateString,
    outcome_definition: OutcomeDefinitionSummarySchema,
    horizons: z.array(Horizon).min(1),
    priors: PriorsSummarySchema,
    weights: WeightsSummarySchema,
    dependencies: DependenciesSummarySchema,
    estimates_summary: z.array(EstimateSummarySchema),
    intervals_summary: z.array(IntervalSummarySchema),
    sensitivity_summary: SensitivitySummarySchema,
    external_forecasts_summary: ExternalForecastsSummarySchema,
    changes: z.array(z.string()),
    reviewers: z.array(ReviewerSchema),
    approval: ApprovalSummarySchema,
    known_limitations: z.array(z.string()),
    reproduction_command: NonEmptyString,
    signature: ManifestSignatureSchema.nullable(),
    published: DateTimeString.nullable(),
    superseded: ReleaseId.nullable(),
  })
  .superRefine((value, ctx) => {
    if (value.published !== null && value.approval.status !== "approved") {
      ctx.addIssue({
        code: "custom",
        message: "a published release must carry an approved approval summary",
        path: ["approval", "status"],
      });
    }
    if (value.published !== null && value.signature === null) {
      ctx.addIssue({
        code: "custom",
        message: "a published release must be signed",
        path: ["signature"],
      });
    }
  })
  .describe("MODEL RELEASE ARTIFACT manifest");
export type ReleaseManifest = z.infer<typeof ReleaseManifestSchema>;

// ---------------------------------------------------------------------------
// Approval (approvals.json item)
// ---------------------------------------------------------------------------

export const ApprovalSchema = z
  .strictObject({
    reviewer_id: NonEmptyString,
    key_id: NonEmptyString,
    signed_at: DateTimeString,
    manifest_sha256: Sha256Hex,
    signature_base64: Base64String,
    conflicts_declared: z.array(z.string()),
  })
  .describe("Signed reviewer approval of a release manifest");
export type Approval = z.infer<typeof ApprovalSchema>;

// ---------------------------------------------------------------------------
// DriversExplained ("Why this number?")
// ---------------------------------------------------------------------------

export const DriverSensitivitySchema = z.strictObject({
  leave_one_out_delta: z.number().nullable().describe("Index points moved when this driver is removed"),
  note: z.string(),
});

export const DriversExplainedItemSchema = z
  .strictObject({
    driver: DriverFamily,
    signal_id: SignalId,
    index_id: IndexId,
    direction: ContributionDirection,
    magnitude: z.number().min(0).describe("Absolute contribution in index points"),
    source_ids: z.array(SourceId),
    confidence: ConfidenceLabel,
    model_role: ModelRole,
    last_updated: DateString,
    sensitivity: DriverSensitivitySchema,
    counterevidence: z.string().nullable(),
  })
  .superRefine((value, ctx) => {
    if (!value.signal_id.startsWith(`${value.driver}.`)) {
      ctx.addIssue({
        code: "custom",
        message: "signal_id must belong to driver",
        path: ["signal_id"],
      });
    }
  })
  .describe("One driver contribution shown in the 'Why this number?' panel");
export type DriversExplainedItem = z.infer<typeof DriversExplainedItemSchema>;
