// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * @pdoom/schemas — zod v4 schemas, enumerations, constant tables and TypeScript
 * types for every p(DOOM) data object. Go (internal/schema) and the web app mirror
 * the names exported here; see build-spec §3, §5 and Appendix A.
 */
import { z } from "zod";
import {
  ActionSchema,
  BenchmarkResultSchema,
  BenchmarkSchema,
  ClaimSchema,
  DefinitionSchema,
  DriverObservationSchema,
  DriverSchema,
  ForecastSchema,
  IncidentSchema,
  InterventionSchema,
  ModelSpecSchema,
  OrganizationSchema,
  ScenarioEdgeSchema,
  ScenarioSchema,
  SnapshotManifestSchema,
  SourceSchema,
} from "./entities";
import {
  AggregationResultSchema,
  ApprovalSchema,
  DeltaRecordSchema,
  DriversExplainedSchema,
  EstimateSchema,
  IndexValueSchema,
  ReleaseManifestSchema,
  SensitivityRunSchema,
} from "./release";
import {
  CassandraFindingSchema,
  SubmissionSchema,
  UserScenarioParamsSchema,
  UserScenarioResultSchema,
} from "./lab";

export * from "./common";
export * from "./enums";
export * from "./entities";
export * from "./release";
export * from "./lab";

// ---------------------------------------------------------------------------
// Envelope (build-spec §3.1)
// ---------------------------------------------------------------------------

export const SCHEMA_VERSION = 1 as const;

/**
 * Wraps an item schema in the file envelope `{ kind, schema_version: 1, items }`.
 * Pass `kind` to pin the `kind` field to that literal.
 */
export function EnvelopeSchema<T extends z.ZodType>(itemSchema: T, kind?: string) {
  return z.strictObject({
    kind: kind === undefined ? z.string().min(1) : z.literal(kind),
    schema_version: z.literal(SCHEMA_VERSION),
    items: z.array(itemSchema),
  });
}

export interface Envelope<T> {
  kind: string;
  schema_version: typeof SCHEMA_VERSION;
  items: T[];
}

// ---------------------------------------------------------------------------
// Kinds → schema → file
// ---------------------------------------------------------------------------

/** Snapshot entity kinds (build-spec §3.1 plus `action` from Appendix A). */
export const EntityKind = z.enum([
  "definition",
  "source",
  "claim",
  "forecast",
  "benchmark",
  "benchmark_result",
  "incident",
  "scenario",
  "scenario_edge",
  "driver",
  "driver_observation",
  "intervention",
  "organization",
  "action",
  "model_spec",
]);
export type EntityKindValue = z.infer<typeof EntityKind>;

/** Kinds of the envelope files inside a release directory (build-spec §3.6). */
export const ReleaseKind = z.enum([
  "estimate",
  "index_value",
  "aggregation",
  "sensitivity_run",
  "drivers_explained",
]);
export type ReleaseKindValue = z.infer<typeof ReleaseKind>;

export type KindLocation = "snapshot" | "release" | "review_queue" | "runtime";
/**
 * How the object is stored on disk:
 * - `envelope`: `{ kind, schema_version, items: [...] }`
 * - `object`: a single JSON object
 * - `array`: a bare JSON array of items
 * - `jsonl`: one JSON object per line
 * - `none`: request / response object, never written by the pipeline
 */
export type KindContainer = "envelope" | "object" | "array" | "jsonl" | "none";

export interface KindEntry<S extends z.ZodType = z.ZodType> {
  kind: string;
  schema: S;
  /** File name inside the snapshot / release directory, or null for runtime objects. */
  file: string | null;
  location: KindLocation;
  container: KindContainer;
}

export const KINDS = [
  // snapshot
  { kind: "definition", schema: DefinitionSchema, file: "definitions.json", location: "snapshot", container: "envelope" },
  { kind: "source", schema: SourceSchema, file: "sources.json", location: "snapshot", container: "envelope" },
  { kind: "claim", schema: ClaimSchema, file: "claims.json", location: "snapshot", container: "envelope" },
  { kind: "forecast", schema: ForecastSchema, file: "forecasts.json", location: "snapshot", container: "envelope" },
  { kind: "benchmark", schema: BenchmarkSchema, file: "benchmarks.json", location: "snapshot", container: "envelope" },
  { kind: "benchmark_result", schema: BenchmarkResultSchema, file: "benchmark_results.json", location: "snapshot", container: "envelope" },
  { kind: "incident", schema: IncidentSchema, file: "incidents.json", location: "snapshot", container: "envelope" },
  { kind: "scenario", schema: ScenarioSchema, file: "scenarios.json", location: "snapshot", container: "envelope" },
  { kind: "scenario_edge", schema: ScenarioEdgeSchema, file: "scenario_edges.json", location: "snapshot", container: "envelope" },
  { kind: "driver", schema: DriverSchema, file: "drivers.json", location: "snapshot", container: "envelope" },
  { kind: "driver_observation", schema: DriverObservationSchema, file: "driver_observations.json", location: "snapshot", container: "envelope" },
  { kind: "intervention", schema: InterventionSchema, file: "interventions.json", location: "snapshot", container: "envelope" },
  { kind: "organization", schema: OrganizationSchema, file: "organizations.json", location: "snapshot", container: "envelope" },
  { kind: "action", schema: ActionSchema, file: "actions.json", location: "snapshot", container: "envelope" },
  { kind: "model_spec", schema: ModelSpecSchema, file: "model_spec.json", location: "snapshot", container: "envelope" },
  { kind: "snapshot_manifest", schema: SnapshotManifestSchema, file: "manifest.json", location: "snapshot", container: "object" },
  // release / candidate
  { kind: "estimate", schema: EstimateSchema, file: "estimates.json", location: "release", container: "envelope" },
  { kind: "index_value", schema: IndexValueSchema, file: "indexes.json", location: "release", container: "envelope" },
  { kind: "aggregation", schema: AggregationResultSchema, file: "aggregations.json", location: "release", container: "envelope" },
  { kind: "sensitivity_run", schema: SensitivityRunSchema, file: "sensitivity.json", location: "release", container: "envelope" },
  { kind: "drivers_explained", schema: DriversExplainedSchema, file: "drivers_explained.json", location: "release", container: "object" },
  { kind: "delta_record", schema: DeltaRecordSchema, file: "delta.json", location: "release", container: "object" },
  { kind: "release_manifest", schema: ReleaseManifestSchema, file: "manifest.json", location: "release", container: "object" },
  { kind: "approval", schema: ApprovalSchema, file: "approvals.json", location: "release", container: "array" },
  // review queue / runtime
  { kind: "submission", schema: SubmissionSchema, file: "submissions.jsonl", location: "review_queue", container: "jsonl" },
  { kind: "cassandra_finding", schema: CassandraFindingSchema, file: null, location: "runtime", container: "none" },
  { kind: "user_scenario_params", schema: UserScenarioParamsSchema, file: null, location: "runtime", container: "none" },
  { kind: "user_scenario_result", schema: UserScenarioResultSchema, file: null, location: "runtime", container: "none" },
] as const satisfies readonly KindEntry[];

export type Kind = (typeof KINDS)[number]["kind"];

/** Look up a KINDS entry by kind name. */
export function kindEntry(kind: Kind): KindEntry {
  const entry = KINDS.find((k) => k.kind === kind);
  if (!entry) throw new Error(`unknown kind: ${kind}`);
  return entry;
}

/** Envelope schema for a snapshot or release kind, with `kind` pinned. */
export function envelopeFor(kind: Kind) {
  const entry = kindEntry(kind);
  if (entry.container !== "envelope") {
    throw new Error(`kind ${kind} is stored as ${entry.container}, not as an envelope`);
  }
  return EnvelopeSchema(entry.schema, kind);
}

/** Snapshot files in the order they appear in build-spec §3.5. */
export const SNAPSHOT_FILES: readonly { kind: Kind; file: string }[] = KINDS.filter(
  (k) => k.location === "snapshot" && k.file !== null,
).map((k) => ({ kind: k.kind, file: k.file as string }));
