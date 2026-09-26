// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Shared primitives used by every p(DOOM) schema.
 *
 * All regular expressions here must stay RE2-compatible (no lookaheads, no
 * backreferences) because Go validates the generated JSON Schemas with a
 * Go regexp engine. Patterns are exported as strings in ID_PATTERNS so that
 * internal/schema can mirror them verbatim.
 */
import { z } from "zod";

// ---------------------------------------------------------------------------
// Scalars
// ---------------------------------------------------------------------------

/** Calendar date, `YYYY-MM-DD` (month 01–12, day 01–31). */
export const DATE_PATTERN = "^\\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\\d|3[01])$";
export const DateString = z
  .string()
  .regex(new RegExp(DATE_PATTERN), "expected a date formatted YYYY-MM-DD")
  .describe("Calendar date, YYYY-MM-DD");

/** ISO 8601 timestamp in UTC, e.g. `2026-09-26T00:00:00Z`. */
export const DateTimeString = z
  .iso.datetime({ offset: false })
  .describe("ISO 8601 timestamp in UTC (…Z)");

/** A probability on the [0, 1] scale. */
export const Probability = z.number().min(0).max(1).describe("Probability in [0, 1]");

/** A dimensionless value in [0, 1] (normalised signal, confidence, coverage). */
export const UnitInterval = z.number().min(0).max(1).describe("Value in [0, 1]");

/** Index values are 0–100 and are never probabilities (build-spec §0.3). */
export const IndexScore = z
  .number()
  .min(0)
  .max(100)
  .describe("Index score on the 0–100 scale; not a probability");

export const NonNegativeInt = z.int().min(0);
export const PositiveInt = z.int().min(1);

/** A non-empty, trimmed string. */
export const NonEmptyString = z.string().trim().min(1);

/** http(s) URL. */
export const HttpUrl = z.url({ protocol: /^https?$/ }).describe("http(s) URL");

/** Lower-case hexadecimal SHA-256 digest. */
export const SHA256_PATTERN = "^[a-f0-9]{64}$";
export const Sha256Hex = z.string().regex(new RegExp(SHA256_PATTERN), "expected a sha256 hex digest");

/** Git commit hash (7–40 hex characters). */
export const GIT_COMMIT_PATTERN = "^[a-f0-9]{7,40}$";
export const GitCommit = z.string().regex(new RegExp(GIT_COMMIT_PATTERN), "expected a git commit hash");

/** Standard base64 (RFC 4648 §4) with optional padding. */
export const BASE64_PATTERN = "^[A-Za-z0-9+/]+={0,2}$";
export const Base64String = z.string().regex(new RegExp(BASE64_PATTERN), "expected base64 text");

// ---------------------------------------------------------------------------
// Quantiles
// ---------------------------------------------------------------------------

export const QUANTILE_KEYS = ["p05", "p25", "p50", "p75", "p95"] as const;
export type QuantileKey = (typeof QUANTILE_KEYS)[number];

/** Full five-point quantile set on the probability scale. */
export const QuantilesSchema = z
  .strictObject({
    p05: Probability,
    p25: Probability,
    p50: Probability,
    p75: Probability,
    p95: Probability,
  })
  .describe("Five-point quantiles (probabilities)");
export type Quantiles = z.infer<typeof QuantilesSchema>;

/** Any subset of the five quantiles (forecast records often report only a median). */
export const PartialQuantilesSchema = z
  .strictObject({
    p05: Probability.optional(),
    p25: Probability.optional(),
    p50: Probability.optional(),
    p75: Probability.optional(),
    p95: Probability.optional(),
  })
  .describe("Any subset of p05..p95 (probabilities)");
export type PartialQuantiles = z.infer<typeof PartialQuantilesSchema>;

/** Three-point summary used by the experimental causal model parameters. */
export const TriQuantilesSchema = z
  .strictObject({ p05: Probability, p50: Probability, p95: Probability })
  .describe("Three-point quantiles (probabilities)");
export type TriQuantiles = z.infer<typeof TriQuantilesSchema>;

/** Six-point summary returned by the Scenario Lab. */
export const SummaryStatsSchema = z
  .strictObject({
    p05: Probability,
    p25: Probability,
    p50: Probability,
    p75: Probability,
    p95: Probability,
    mean: Probability,
  })
  .describe("Five quantiles plus mean (probabilities)");
export type SummaryStats = z.infer<typeof SummaryStatsSchema>;

/** Adds an issue when the present quantiles are not non-decreasing. */
export function checkQuantilesMonotonic(
  q: Partial<Record<QuantileKey, number>> | null | undefined,
  ctx: z.RefinementCtx,
  path: (string | number)[] = ["quantiles"],
): void {
  if (!q) return;
  let previous: number | undefined;
  for (const key of QUANTILE_KEYS) {
    const value = q[key];
    if (value === undefined) continue;
    if (previous !== undefined && value < previous) {
      ctx.addIssue({
        code: "custom",
        message: `quantiles must be non-decreasing (${key} < preceding quantile)`,
        path: [...path, key],
      });
      return;
    }
    previous = value;
  }
}

// ---------------------------------------------------------------------------
// Identifier patterns (build-spec §3.2). Exported as strings for Go mirroring.
// ---------------------------------------------------------------------------

const SLUG = "[a-z0-9]+(?:-[a-z0-9]+)*";

export const ID_PATTERNS = {
  definition: `^def-${SLUG}$`,
  source: `^src-${SLUG}$`,
  claim: `^clm-${SLUG}-\\d{2,}$`,
  forecast: `^fc-${SLUG}$`,
  benchmark: `^bm-${SLUG}$`,
  benchmark_result: `^bmr-${SLUG}$`,
  incident: `^inc-${SLUG}$`,
  scenario: "^S(?:[1-9]|1[0-8])$",
  scenario_edge: "^se-S(?:[1-9]|1[0-8])-S(?:[1-9]|1[0-8])$",
  driver_family: "^D(?:[1-9]|10)$",
  signal: "^D(?:[1-9]|10)\\.[a-z][a-z0-9]*(?:_[a-z0-9]+)*$",
  intervention: "^I\\d{2,}$",
  organization: `^org-${SLUG}$`,
  action: `^act-[a-z_]+-${SLUG}$`,
  outcome: "^O[0-8]$",
  snapshot: "^snap-\\d{4}-\\d{2}-\\d{2}-\\d{3}$",
  candidate: "^cand-\\d{4}-\\d{2}-\\d{2}-\\d{3}$",
  release: "^rel-\\d{4}-\\d{2}-\\d{2}-\\d{3}$",
  model_version: `^pdoom-model/${SLUG}@\\d+\\.\\d+\\.\\d+(?:-[0-9A-Za-z.-]+)?$`,
  model_spec: "^pdoom-model-spec@\\d+\\.\\d+\\.\\d+$",
  estimate: "^est-[A-Za-z0-9_+-]+$",
} as const;

export type IdKind = keyof typeof ID_PATTERNS;

function idSchema(kind: IdKind, example: string) {
  return z
    .string()
    .regex(new RegExp(ID_PATTERNS[kind]), `expected a ${kind} id such as ${example}`)
    .describe(`${kind} id, e.g. ${example}`);
}

export const DefinitionId = idSchema("definition", "def-agi");
export const SourceId = idSchema("source", "src-grace-2024-thousands-of-ai-authors");
export const ClaimId = idSchema("claim", "clm-metr-time-horizon-01");
export const ForecastId = idSchema("forecast", "fc-xpt-2022-superforecasters-ai-extinction-2100");
export const BenchmarkId = idSchema("benchmark", "bm-metr-time-horizon-50");
export const BenchmarkResultId = idSchema("benchmark_result", "bmr-metr-th50-example-model");
export const IncidentId = idSchema("incident", "inc-aiid-0001-example");
export const ScenarioId = idSchema("scenario", "S3");
export const ScenarioEdgeId = idSchema("scenario_edge", "se-S14-S15");
export const DriverFamilyId = idSchema("driver_family", "D5");
export const SignalId = idSchema("signal", "D1.task_horizon_50pct");
export const InterventionId = idSchema("intervention", "I04");
export const OrganizationId = idSchema("organization", "org-metr");
export const ActionId = idSchema("action", "act-individuals-learn-the-basics");
export const SnapshotId = idSchema("snapshot", "snap-2026-09-26-001");
export const CandidateId = idSchema("candidate", "cand-2026-09-26-001");
export const ReleaseId = idSchema("release", "rel-2026-09-26-001");
export const ModelVersion = idSchema("model_version", "pdoom-model/external-aggregate@0.1.0");
export const ModelSpecId = idSchema("model_spec", "pdoom-model-spec@0.1.0");
export const EstimateId = idSchema("estimate", "est-external-O6-2100");

/** Any snapshot entity id that we do not constrain further (e.g. driver observations). */
export const GenericId = NonEmptyString.describe("Entity id");
