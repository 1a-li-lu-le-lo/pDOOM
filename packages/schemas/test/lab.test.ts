// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { describe, expect, it } from "vitest";
import {
  CassandraFindingSchema,
  CorrectionSubmissionPayloadSchema,
  IncidentReferenceSubmissionPayloadSchema,
  KINDS,
  SourceSubmissionPayloadSchema,
  SubmissionSchema,
  USER_SCENARIO_MAX_SAMPLES,
  USER_SCENARIO_OUTCOME_KEYS,
  USER_SCENARIO_SLIDER_KEYS,
  UserScenarioParamsSchema,
  UserScenarioResultSchema,
} from "../src/index";
import {
  FIXTURES,
  cassandra_finding,
  submission,
  user_scenario_params,
  user_scenario_result,
} from "./fixtures";

const runtimeKinds = KINDS.filter((k) => k.location === "runtime" || k.location === "review_queue");

describe("runtime object fixtures", () => {
  it.each(runtimeKinds.map((k) => [k.kind, k] as const))("%s fixture parses", (kind, entry) => {
    const result = entry.schema.safeParse(FIXTURES[kind]);
    expect(result.error?.issues ?? []).toEqual([]);
  });

  it.each(runtimeKinds.map((k) => [k.kind, k] as const))("%s rejects an unknown field", (kind, entry) => {
    expect(entry.schema.safeParse({ ...(FIXTURES[kind] as object), unexpected_field: 1 }).success).toBe(false);
  });
});

describe("UserScenarioParams", () => {
  it("has exactly ten sliders", () => {
    expect(USER_SCENARIO_SLIDER_KEYS).toHaveLength(10);
    const keys = Object.keys(UserScenarioParamsSchema.shape);
    for (const slider of USER_SCENARIO_SLIDER_KEYS) expect(keys).toContain(slider);
    expect(keys).toEqual(["horizon", ...USER_SCENARIO_SLIDER_KEYS, "samples", "seed"]);
  });

  it("sliders are integers in -2..2", () => {
    expect(UserScenarioParamsSchema.safeParse({ ...user_scenario_params, resilience: 3 }).success).toBe(false);
    expect(UserScenarioParamsSchema.safeParse({ ...user_scenario_params, resilience: -3 }).success).toBe(false);
    expect(UserScenarioParamsSchema.safeParse({ ...user_scenario_params, resilience: 0.5 }).success).toBe(false);
    expect(UserScenarioParamsSchema.safeParse({ ...user_scenario_params, resilience: -2 }).success).toBe(true);
  });

  it("samples and seed are optional and bounded", () => {
    const { samples: _s, seed: _d, ...minimal } = user_scenario_params;
    expect(UserScenarioParamsSchema.safeParse(minimal).success).toBe(true);
    expect(UserScenarioParamsSchema.safeParse({ ...minimal, samples: USER_SCENARIO_MAX_SAMPLES }).success).toBe(true);
    expect(UserScenarioParamsSchema.safeParse({ ...minimal, samples: USER_SCENARIO_MAX_SAMPLES + 1 }).success).toBe(
      false,
    );
    expect(UserScenarioParamsSchema.safeParse({ ...minimal, samples: 0 }).success).toBe(false);
    expect(UserScenarioParamsSchema.safeParse({ ...minimal, seed: -1 }).success).toBe(false);
    expect(UserScenarioParamsSchema.safeParse({ ...minimal, horizon: "custom" }).success).toBe(false);
  });
});

describe("UserScenarioResult", () => {
  it("is always labelled user_scenario", () => {
    expect(UserScenarioResultSchema.safeParse({ ...user_scenario_result, label: "official" }).success).toBe(false);
  });

  it("requires every outcome key and rejects unknown ones", () => {
    expect(USER_SCENARIO_OUTCOME_KEYS).toEqual(["O3", "O4", "O5", "O6", "O7", "O8", "P_DOOM", "P_COLLAPSE"]);
    const { P_DOOM: _p, ...partial } = user_scenario_result.outcome_estimates;
    expect(UserScenarioResultSchema.safeParse({ ...user_scenario_result, outcome_estimates: partial }).success).toBe(
      false,
    );
    expect(
      UserScenarioResultSchema.safeParse({
        ...user_scenario_result,
        outcome_estimates: { ...user_scenario_result.outcome_estimates, O0: user_scenario_result.outcome_estimates.O3 },
      }).success,
    ).toBe(false);
  });

  it("echoes params with the same horizon", () => {
    expect(
      UserScenarioResultSchema.safeParse({
        ...user_scenario_result,
        params_echo: { ...user_scenario_params, horizon: "25y" },
      }).success,
    ).toBe(false);
  });
});

describe("Submission and CassandraFinding", () => {
  it("submissions are always received and carry an object payload", () => {
    expect(SubmissionSchema.safeParse({ ...submission, status: "accepted" }).success).toBe(false);
    expect(SubmissionSchema.safeParse({ ...submission, payload: "text" }).success).toBe(false);
    expect(SubmissionSchema.safeParse({ ...submission, kind: "praise" }).success).toBe(false);
    const { contact: _c, ...withoutContact } = submission;
    expect(SubmissionSchema.safeParse(withoutContact).success).toBe(true);
  });

  it("typed payload shapes validate", () => {
    expect(
      SourceSubmissionPayloadSchema.safeParse({
        canonical_url: "https://example.org/report",
        title: "Fixture",
        why_relevant: "fixture",
        claimed_evidence: "fixture",
      }).success,
    ).toBe(true);
    expect(
      CorrectionSubmissionPayloadSchema.safeParse({ target_id: "src-fixture", correction: "fixture" }).success,
    ).toBe(true);
    expect(
      IncidentReferenceSubmissionPayloadSchema.safeParse({
        registry: "aiid",
        external_id: "1",
        url: "https://example.org/incident/1",
        summary: "fixture",
      }).success,
    ).toBe(true);
  });

  it("cassandra findings use the fixed severity and status sets", () => {
    expect(CassandraFindingSchema.safeParse({ ...cassandra_finding, severity: "blocker" }).success).toBe(false);
    expect(CassandraFindingSchema.safeParse({ ...cassandra_finding, status: "wontfix" }).success).toBe(false);
    expect(CassandraFindingSchema.safeParse({ ...cassandra_finding, status: "accepted_risk" }).success).toBe(true);
  });
});
