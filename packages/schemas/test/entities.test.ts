// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { describe, expect, it } from "vitest";
import {
  ActionSchema,
  BenchmarkResultSchema,
  DriverObservationSchema,
  DriverSchema,
  ForecastSchema,
  KINDS,
  ModelSpecSchema,
  ScenarioEdgeSchema,
  SourceSchema,
  VerificationSchema,
} from "../src/index";
import { FIXTURES, driver, driver_observation, forecast, model_spec, source } from "./fixtures";

const snapshotKinds = KINDS.filter((k) => k.location === "snapshot");

describe("snapshot entity fixtures", () => {
  it.each(snapshotKinds.map((k) => [k.kind, k] as const))("%s fixture parses", (kind, entry) => {
    const result = entry.schema.safeParse(FIXTURES[kind]);
    expect(result.error?.issues ?? []).toEqual([]);
    expect(result.success).toBe(true);
  });

  it.each(snapshotKinds.map((k) => [k.kind, k] as const))(
    "%s rejects an unknown field",
    (kind, entry) => {
      const withExtra = { ...(FIXTURES[kind] as object), unexpected_field: 1 };
      const result = entry.schema.safeParse(withExtra);
      expect(result.success).toBe(false);
      expect(result.error?.issues.some((i) => i.code === "unrecognized_keys")).toBe(true);
    },
  );

  it.each(snapshotKinds.map((k) => [k.kind, k] as const))(
    "%s rejects a missing required field",
    (kind, entry) => {
      const fixture = FIXTURES[kind] as Record<string, unknown>;
      const firstKey = Object.keys(fixture)[0] as string;
      const { [firstKey]: _dropped, ...rest } = fixture;
      expect(entry.schema.safeParse(rest).success).toBe(false);
    },
  );
});

describe("verification and model use", () => {
  it("verification requires a YYYY-MM-DD date", () => {
    expect(VerificationSchema.safeParse({ ...source.verification, checked_at: "2026-9-2" }).success).toBe(false);
    expect(VerificationSchema.safeParse({ ...source.verification, checked_at: "2026-13-01" }).success).toBe(false);
    expect(VerificationSchema.safeParse(source.verification).success).toBe(true);
  });

  it("unverified items cannot be eligible or used", () => {
    const unverified = {
      ...source,
      verification: { ...source.verification, status: "unverified" },
    };
    expect(SourceSchema.safeParse({ ...unverified, model_use_status: "used" }).success).toBe(false);
    expect(SourceSchema.safeParse({ ...unverified, model_use_status: "eligible" }).success).toBe(false);
    expect(SourceSchema.safeParse({ ...unverified, model_use_status: "excluded" }).success).toBe(true);
    const prior = {
      ...source,
      verification: { ...source.verification, status: "verified_prior_knowledge" },
    };
    expect(SourceSchema.safeParse({ ...prior, model_use_status: "eligible" }).success).toBe(false);
    expect(SourceSchema.safeParse({ ...prior, model_use_status: "informational" }).success).toBe(true);
  });

  it("tier 5 must be excluded and tier 4 at most informational", () => {
    expect(SourceSchema.safeParse({ ...source, source_tier: 5, model_use_status: "eligible" }).success).toBe(false);
    expect(SourceSchema.safeParse({ ...source, source_tier: 5, model_use_status: "excluded" }).success).toBe(true);
    expect(SourceSchema.safeParse({ ...source, source_tier: 4, model_use_status: "used" }).success).toBe(false);
    expect(SourceSchema.safeParse({ ...source, source_tier: 4, model_use_status: "informational" }).success).toBe(true);
    expect(SourceSchema.safeParse({ ...source, source_tier: 6 }).success).toBe(false);
  });

  it("canonical_url must be http(s)", () => {
    expect(SourceSchema.safeParse({ ...source, canonical_url: "ftp://example.org/x" }).success).toBe(false);
    expect(SourceSchema.safeParse({ ...source, canonical_url: "not a url" }).success).toBe(false);
  });
});

describe("identifier patterns", () => {
  it("source ids must be src-<slug>", () => {
    expect(SourceSchema.safeParse({ ...source, id: "source-1" }).success).toBe(false);
    expect(SourceSchema.safeParse({ ...source, id: "src-Upper-Case" }).success).toBe(false);
    expect(SourceSchema.safeParse({ ...source, id: "src-ok-slug-2025" }).success).toBe(true);
  });

  it("scenario edge id must be se-<from>-<to> and not self-referential", () => {
    const edge = FIXTURES.scenario_edge as Record<string, unknown>;
    expect(ScenarioEdgeSchema.safeParse({ ...edge, id: "se-S3-S1" }).success).toBe(false);
    expect(ScenarioEdgeSchema.safeParse({ ...edge, id: "se-S1-S1", to_id: "S1" }).success).toBe(false);
    expect(ScenarioEdgeSchema.safeParse({ ...edge, from_id: "S19", id: "se-S19-S3" }).success).toBe(false);
  });

  it("action ids embed the audience", () => {
    const action = FIXTURES.action as Record<string, unknown>;
    expect(ActionSchema.safeParse({ ...action, audience: "funders" }).success).toBe(false);
    expect(
      ActionSchema.safeParse({ ...action, audience: "funders", id: "act-funders-learn-the-basics" }).success,
    ).toBe(true);
  });

  it("signals must belong to their driver family", () => {
    expect(
      DriverSchema.safeParse({
        ...driver,
        signals: [{ ...driver.signals[0], signal_id: "D2.task_horizon_50pct" }],
      }).success,
    ).toBe(false);
    expect(DriverObservationSchema.safeParse({ ...driver_observation, family: "D2" }).success).toBe(false);
    expect(DriverObservationSchema.safeParse({ ...driver_observation, signal_id: "D1.BadName" }).success).toBe(false);
    expect(DriverObservationSchema.safeParse({ ...driver_observation, value_normalized: 1.2 }).success).toBe(false);
  });
});

describe("forecast invariants", () => {
  it("custom horizon requires a note", () => {
    expect(ForecastSchema.safeParse({ ...forecast, horizon: "custom", horizon_note: null }).success).toBe(false);
    expect(
      ForecastSchema.safeParse({ ...forecast, horizon: "custom", horizon_note: "within 100 years" }).success,
    ).toBe(true);
    expect(ForecastSchema.safeParse({ ...forecast, horizon: "50y" }).success).toBe(false);
  });

  it("quantiles must be probabilities and non-decreasing", () => {
    expect(ForecastSchema.safeParse({ ...forecast, quantiles: { p50: 1.5 } }).success).toBe(false);
    expect(
      ForecastSchema.safeParse({ ...forecast, median: null, quantiles: { p25: 0.2, p50: 0.1 } }).success,
    ).toBe(false);
    expect(ForecastSchema.safeParse({ ...forecast, quantiles: {} }).success).toBe(true);
  });

  it("median and p50 must agree when both present", () => {
    expect(ForecastSchema.safeParse({ ...forecast, median: 0.2 }).success).toBe(false);
    expect(ForecastSchema.safeParse({ ...forecast, median: null }).success).toBe(true);
  });

  it("outcome_set must be non-empty and use outcome codes", () => {
    expect(ForecastSchema.safeParse({ ...forecast, outcome_set: [] }).success).toBe(false);
    expect(ForecastSchema.safeParse({ ...forecast, outcome_set: ["extinction"] }).success).toBe(false);
  });
});

describe("benchmark results and model spec", () => {
  it("confidence interval must be ordered", () => {
    const result = FIXTURES.benchmark_result as Record<string, unknown>;
    expect(BenchmarkResultSchema.safeParse({ ...result, ci_low: 90, ci_high: 40 }).success).toBe(false);
  });

  it("model spec weights respect weight_bounds", () => {
    const bad = {
      ...model_spec,
      index_weights: {
        ...model_spec.index_weights,
        capability_pressure: { "D1.task_horizon_50pct": 0.5 },
      },
    };
    expect(ModelSpecSchema.safeParse(bad).success).toBe(false);
  });

  it("model spec requires every incident weight key", () => {
    const { catastrophic: _c, ...partial } = model_spec.incident_scoring.severity_weights;
    const bad = {
      ...model_spec,
      incident_scoring: { ...model_spec.incident_scoring, severity_weights: partial },
    };
    expect(ModelSpecSchema.safeParse(bad).success).toBe(false);
  });

  it("experimental causal horizons must be spec horizons with O3..O8", () => {
    const horizonSpec = model_spec.experimental_causal.horizons["10y"];
    const bad = {
      ...model_spec,
      experimental_causal: {
        ...model_spec.experimental_causal,
        horizons: { "7y": horizonSpec },
      },
    };
    expect(ModelSpecSchema.safeParse(bad).success).toBe(false);
    const { O8: _o8, ...partialO } = horizonSpec.O;
    const missingOutcome = {
      ...model_spec,
      experimental_causal: {
        ...model_spec.experimental_causal,
        horizons: { "10y": { ...horizonSpec, O: partialO } },
      },
    };
    expect(ModelSpecSchema.safeParse(missingOutcome).success).toBe(false);
  });
});
