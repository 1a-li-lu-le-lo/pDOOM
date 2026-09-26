// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Scenario Lab, public submissions and Cassandra review findings.
 *
 * Scenario Lab results are labelled `user_scenario` and are never the p(DOOM)
 * official model (build-spec §0.2). Submissions never modify snapshots or
 * releases; they only enter the review queue.
 */
import { z } from "zod";
import { DateTimeString, NonEmptyString, SummaryStatsSchema, TriQuantilesSchema } from "./common";
import {
  CassandraSeverity,
  CassandraStatus,
  Horizon,
  SubmissionKind,
  SubmissionStatus,
} from "./enums";

// ---------------------------------------------------------------------------
// UserScenarioParams
// ---------------------------------------------------------------------------

export const USER_SCENARIO_SLIDER_KEYS = [
  "capability_timeline",
  "autonomy_growth",
  "access_level",
  "safety_progress",
  "governance_strength",
  "model_security",
  "open_weight_diffusion",
  "international_coordination",
  "incident_frequency",
  "resilience",
] as const;
export type UserScenarioSliderKey = (typeof USER_SCENARIO_SLIDER_KEYS)[number];

export const USER_SCENARIO_MAX_SAMPLES = 50000;

/** Integer slider position from -2 (much lower than baseline) to +2 (much higher). */
export const SliderValueSchema = z.int().min(-2).max(2).describe("Integer slider, -2..2");
export type SliderValue = z.infer<typeof SliderValueSchema>;

export const UserScenarioParamsSchema = z
  .strictObject({
    horizon: Horizon,
    capability_timeline: SliderValueSchema,
    autonomy_growth: SliderValueSchema,
    access_level: SliderValueSchema,
    safety_progress: SliderValueSchema,
    governance_strength: SliderValueSchema,
    model_security: SliderValueSchema,
    open_weight_diffusion: SliderValueSchema,
    international_coordination: SliderValueSchema,
    incident_frequency: SliderValueSchema,
    resilience: SliderValueSchema,
    samples: z.int().min(1).max(USER_SCENARIO_MAX_SAMPLES).optional(),
    seed: z.int().min(0).max(4294967295).optional(),
  })
  .describe("Scenario Lab input: horizon plus ten integer sliders in -2..2");
export type UserScenarioParams = z.infer<typeof UserScenarioParamsSchema>;

// ---------------------------------------------------------------------------
// UserScenarioResult
// ---------------------------------------------------------------------------

export const USER_SCENARIO_LABEL = "user_scenario" as const;

export const USER_SCENARIO_OUTCOME_KEYS = [
  "O3",
  "O4",
  "O5",
  "O6",
  "O7",
  "O8",
  "P_DOOM",
  "P_COLLAPSE",
] as const;
export type UserScenarioOutcomeKey = (typeof USER_SCENARIO_OUTCOME_KEYS)[number];

export const UserScenarioOutcomeKey = z.enum(USER_SCENARIO_OUTCOME_KEYS);

export const FactorSummarySchema = z.strictObject({
  A: TriQuantilesSchema,
  C: TriQuantilesSchema,
  E: TriQuantilesSchema,
  F: TriQuantilesSchema,
});
export type FactorSummary = z.infer<typeof FactorSummarySchema>;

export const UserScenarioResultSchema = z
  .strictObject({
    label: z.literal(USER_SCENARIO_LABEL),
    disclaimer: NonEmptyString,
    horizon: Horizon,
    outcome_estimates: z.record(UserScenarioOutcomeKey, SummaryStatsSchema),
    factor_summary: FactorSummarySchema,
    flags: z.array(z.string()),
    params_echo: UserScenarioParamsSchema,
    spec_version: NonEmptyString,
  })
  .superRefine((value, ctx) => {
    if (value.params_echo.horizon !== value.horizon) {
      ctx.addIssue({
        code: "custom",
        message: "params_echo.horizon must match horizon",
        path: ["params_echo", "horizon"],
      });
    }
  })
  .describe("Scenario Lab output; labelled user_scenario, never the official model");
export type UserScenarioResult = z.infer<typeof UserScenarioResultSchema>;

// ---------------------------------------------------------------------------
// Submission
// ---------------------------------------------------------------------------

/** Free-form JSON object payload; the typed payload schemas below are recommended shapes. */
export const SubmissionPayloadSchema = z.record(z.string(), z.unknown());

export const SubmissionSchema = z
  .strictObject({
    id: NonEmptyString,
    kind: SubmissionKind,
    submitted_at: DateTimeString,
    payload: SubmissionPayloadSchema,
    contact: z.string().optional(),
    status: SubmissionStatus,
  })
  .describe("Public submission appended to the review queue; never changes published data");
export type Submission = z.infer<typeof SubmissionSchema>;

export const SourceSubmissionPayloadSchema = z.strictObject({
  canonical_url: z.url({ protocol: /^https?$/ }),
  title: NonEmptyString,
  publisher: z.string().optional(),
  date_published: z.string().optional(),
  why_relevant: z.string(),
  claimed_evidence: z.string(),
});
export type SourceSubmissionPayload = z.infer<typeof SourceSubmissionPayloadSchema>;

export const CorrectionSubmissionPayloadSchema = z.strictObject({
  target_id: NonEmptyString.describe("Id of the entity or estimate being corrected"),
  field: z.string().optional(),
  correction: NonEmptyString,
  evidence_url: z.url({ protocol: /^https?$/ }).optional(),
});
export type CorrectionSubmissionPayload = z.infer<typeof CorrectionSubmissionPayloadSchema>;

export const IncidentReferenceSubmissionPayloadSchema = z.strictObject({
  registry: NonEmptyString.describe("e.g. aiid, oecd_aim, docket"),
  external_id: NonEmptyString,
  url: z.url({ protocol: /^https?$/ }),
  summary: z.string().describe("Non-graphic, non-operational"),
});
export type IncidentReferenceSubmissionPayload = z.infer<
  typeof IncidentReferenceSubmissionPayloadSchema
>;

// ---------------------------------------------------------------------------
// CassandraFinding
// ---------------------------------------------------------------------------

export const CassandraFindingSchema = z
  .strictObject({
    finding_id: NonEmptyString,
    severity: CassandraSeverity,
    component: NonEmptyString,
    claim: z.string(),
    assumption: z.string(),
    counterevidence: z.string(),
    reproduction: z.string(),
    impact: z.string(),
    recommendation: z.string(),
    status: CassandraStatus,
  })
  .describe("Adversarial review finding");
export type CassandraFinding = z.infer<typeof CassandraFindingSchema>;
