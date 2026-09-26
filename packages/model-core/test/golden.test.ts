// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import type { ExperimentalCausalSpec, UserScenarioParams, UserScenarioResult } from "@pdoom/schemas";
import { evaluateUserScenario, formatInterval, mulberry32, roundForDisplay, simulate } from "../src/index";

interface Golden {
  rng: { seed: number; first_five: number[] };
  spec: ExperimentalCausalSpec;
  params: UserScenarioParams;
  samples: number;
  expected: UserScenarioResult;
}

const golden: Golden = JSON.parse(
  readFileSync(fileURLToPath(new URL("../../../internal/model/testdata/causal-golden.json", import.meta.url)), "utf8"),
);

describe("cross-language contract with internal/model", () => {
  it("mulberry32 reproduces the Go stream exactly", () => {
    const r = mulberry32(golden.rng.seed);
    for (const want of golden.rng.first_five) expect(r()).toBe(want);
  });

  it("evaluateUserScenario matches the Go golden within 0.01 on every quantile", () => {
    const res = evaluateUserScenario(golden.spec, golden.params, golden.samples, 0);
    expect(res.label).toBe("user_scenario");
    expect(res.spec_version).toBe(golden.expected.spec_version);
    expect(res.flags).toEqual(golden.expected.flags);
    for (const [key, want] of Object.entries(golden.expected.outcome_estimates)) {
      const got = res.outcome_estimates[key as keyof typeof res.outcome_estimates];
      expect(got, key).toBeDefined();
      for (const q of ["p05", "p25", "p50", "p75", "p95", "mean"] as const) {
        expect(Math.abs((got as Record<string, number>)[q]! - (want as Record<string, number>)[q]!), `${key}.${q}`).toBeLessThan(0.01);
      }
    }
    for (const f of ["A", "C", "E", "F"] as const) {
      expect(Math.abs(res.factor_summary[f].p50 - golden.expected.factor_summary[f].p50)).toBeLessThan(1e-6);
    }
  });

  it("is deterministic and directionally sane", () => {
    const h = golden.spec.horizons[golden.params.horizon]!;
    const a = simulate(h, golden.spec.common_factor_loading, 42, 3000);
    const b = simulate(h, golden.spec.common_factor_loading, 42, 3000);
    expect(a).toEqual(b);
    const safer = simulate(h, golden.spec.common_factor_loading, 42, 3000, { F: -1 });
    expect(safer.outcomes.P_DOOM!.p50).toBeLessThan(a.outcomes.P_DOOM!.p50);
  });

  it("rejects out-of-range sliders and unknown horizons", () => {
    expect(() => evaluateUserScenario(golden.spec, { ...golden.params, resilience: 3 })).toThrow(/range/);
    expect(() => evaluateUserScenario(golden.spec, { ...golden.params, horizon: "1y" })).toThrow(/horizon/);
  });

  it("rounds like Go", () => {
    expect(roundForDisplay(0.1273, 5)).toBe("15%");
    expect(roundForDisplay(0.1273, 2)).toBe("12%");
    expect(roundForDisplay(0.004, 1)).toBe("<1%");
    expect(roundForDisplay(0.996, 1)).toBe(">99%");
    expect(formatInterval(0.03, 0.3, 5)).toBe("5%–30%");
  });
});
