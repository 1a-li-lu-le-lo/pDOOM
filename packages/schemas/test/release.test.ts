// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { describe, expect, it } from "vitest";
import {
  AggregationResultSchema,
  DeltaRecordSchema,
  DriversExplainedItemSchema,
  EnvelopeSchema,
  EstimateSchema,
  IndexValueSchema,
  KINDS,
  ReleaseManifestSchema,
  SensitivityRunSchema,
  envelopeFor,
} from "../src/index";
import {
  FIXTURES,
  aggregation,
  delta_record,
  drivers_explained,
  estimate,
  index_value,
  official_estimate,
  release_manifest,
  sensitivity_run,
  source,
} from "./fixtures";

const releaseKinds = KINDS.filter((k) => k.location === "release");

describe("release object fixtures", () => {
  it.each(releaseKinds.map((k) => [k.kind, k] as const))("%s fixture parses", (kind, entry) => {
    const result = entry.schema.safeParse(FIXTURES[kind]);
    expect(result.error?.issues ?? []).toEqual([]);
  });

  it.each(releaseKinds.map((k) => [k.kind, k] as const))("%s rejects an unknown field", (kind, entry) => {
    const result = entry.schema.safeParse({ ...(FIXTURES[kind] as object), unexpected_field: true });
    expect(result.success).toBe(false);
  });
});

describe("estimates", () => {
  it("the official object is insufficiently calibrated and publishes no probability", () => {
    expect(EstimateSchema.safeParse(official_estimate).success).toBe(true);
    expect(EstimateSchema.safeParse({ ...official_estimate, quantiles: estimate.quantiles }).success).toBe(false);
    expect(EstimateSchema.safeParse({ ...official_estimate, mean: 0.1 }).success).toBe(false);
  });

  it("quantiles are probabilities in order", () => {
    expect(
      EstimateSchema.safeParse({ ...estimate, quantiles: { ...estimate.quantiles, p95: 1.2 } }).success,
    ).toBe(false);
    expect(
      EstimateSchema.safeParse({ ...estimate, quantiles: { ...estimate.quantiles, p75: 0.01 } }).success,
    ).toBe(false);
  });

  it("producer must be a pdoom-model version and horizon a spec key", () => {
    expect(EstimateSchema.safeParse({ ...estimate, producer: "external-aggregate" }).success).toBe(false);
    expect(EstimateSchema.safeParse({ ...estimate, horizon: "custom" }).success).toBe(false);
  });
});

describe("index values", () => {
  it("is_probability must be literally false and value within 0..100 or null", () => {
    expect(IndexValueSchema.safeParse({ ...index_value, is_probability: true }).success).toBe(false);
    expect(IndexValueSchema.safeParse({ ...index_value, value: 101 }).success).toBe(false);
    expect(IndexValueSchema.safeParse({ ...index_value, value: null }).success).toBe(true);
    expect(IndexValueSchema.safeParse({ ...index_value, index_id: "doom" }).success).toBe(false);
  });
});

describe("aggregations and sensitivity", () => {
  it("aggregation quantiles are optional but validated", () => {
    const { quantiles: _q, ...withoutQuantiles } = aggregation;
    expect(AggregationResultSchema.safeParse(withoutQuantiles).success).toBe(true);
    expect(AggregationResultSchema.safeParse({ ...aggregation, value: 1.01 }).success).toBe(false);
    expect(AggregationResultSchema.safeParse({ ...aggregation, method: "average" }).success).toBe(false);
  });

  it("sensitivity delta must equal perturbed minus baseline", () => {
    expect(SensitivityRunSchema.safeParse({ ...sensitivity_run, delta: 0.02 }).success).toBe(false);
    expect(SensitivityRunSchema.safeParse({ ...sensitivity_run, kind: "random" }).success).toBe(false);
  });
});

describe("delta record and manifest", () => {
  it("first-release delta accepts nulls but not missing fields", () => {
    expect(DeltaRecordSchema.safeParse(delta_record).success).toBe(true);
    const { release: _r, ...missing } = delta_record;
    expect(DeltaRecordSchema.safeParse(missing).success).toBe(false);
    expect(DeltaRecordSchema.safeParse({ ...delta_record, absolute_change: 0.02, relative_change: 0.4 }).success).toBe(
      true,
    );
    expect(DeltaRecordSchema.safeParse({ ...delta_record, affected_horizons: ["50y"] }).success).toBe(false);
  });

  it("a published manifest must be approved and signed", () => {
    expect(ReleaseManifestSchema.safeParse({ ...release_manifest, signature: null }).success).toBe(false);
    expect(
      ReleaseManifestSchema.safeParse({
        ...release_manifest,
        approval: { ...release_manifest.approval, status: "pending" },
      }).success,
    ).toBe(false);
    expect(
      ReleaseManifestSchema.safeParse({
        ...release_manifest,
        published: null,
        signature: null,
        approval: { ...release_manifest.approval, status: "pending", received_approvals: 0, approved_at: null },
      }).success,
    ).toBe(true);
    expect(ReleaseManifestSchema.safeParse({ ...release_manifest, code_commit: "not-a-sha" }).success).toBe(false);
  });

  it("drivers explained items must reference their own family", () => {
    expect(DriversExplainedItemSchema.safeParse({ ...drivers_explained, driver: "D2" }).success).toBe(false);
    expect(DriversExplainedItemSchema.safeParse({ ...drivers_explained, direction: "up" }).success).toBe(false);
  });
});

describe("envelopes", () => {
  it("wraps items with kind and schema_version 1", () => {
    const env = envelopeFor("source");
    expect(env.safeParse({ kind: "source", schema_version: 1, items: [source] }).success).toBe(true);
    expect(env.safeParse({ kind: "claim", schema_version: 1, items: [source] }).success).toBe(false);
    expect(env.safeParse({ kind: "source", schema_version: 2, items: [source] }).success).toBe(false);
    expect(env.safeParse({ kind: "source", schema_version: 1, items: [source], extra: 1 }).success).toBe(false);
    expect(env.safeParse({ kind: "source", schema_version: 1, items: [{ ...source, id: "bad" }] }).success).toBe(false);
  });

  it("accepts any kind string when none is pinned", () => {
    const env = EnvelopeSchema(EstimateSchema);
    expect(env.safeParse({ kind: "estimate", schema_version: 1, items: [] }).success).toBe(true);
    expect(env.safeParse({ kind: "", schema_version: 1, items: [] }).success).toBe(false);
  });

  it("refuses to build an envelope for single-object kinds", () => {
    expect(() => envelopeFor("release_manifest")).toThrow(/not as an envelope/);
  });
});
