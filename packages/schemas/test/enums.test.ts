// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { describe, expect, it } from "vitest";
import {
  AGGREGATION_METHODS,
  AggregationMethod,
  CONFLICT_LABELS,
  ConflictLabel,
  DERIVED_OUTCOME_SETS,
  DRIVER_FAMILIES,
  DriverFamily,
  DriverFamilySlug,
  EDITORIAL_RISK_LEVELS,
  EditorialRiskLevel,
  ForecastHorizon,
  HORIZONS,
  Horizon,
  INDEX_DEFINITIONS,
  IndexId,
  OUTCOMES,
  Outcome,
  OutcomeSlug,
  PDoomOutcome,
  ROUNDING_BY_UNCERTAINTY,
  SOURCE_TIERS,
  SourceTier,
  SourceType,
  UncertaintyLabel, HORIZON_KEYS } from "../src";

describe("outcomes", () => {
  it("OUTCOMES covers O0–O8 exactly, in order", () => {
    expect(Outcome.options).toEqual(["O0", "O1", "O2", "O3", "O4", "O5", "O6", "O7", "O8"]);
    expect(Object.keys(OUTCOMES)).toEqual(Outcome.options);
    for (const code of Outcome.options) {
      expect(OUTCOMES[code].code).toBe(code);
      expect(OutcomeSlug.options).toContain(OUTCOMES[code].slug);
      expect(OUTCOMES[code].label.length).toBeGreaterThan(0);
      expect(OUTCOMES[code].description.length).toBeGreaterThan(0);
    }
    expect(new Set(Object.values(OUTCOMES).map((o) => o.slug)).size).toBe(9);
  });

  it("included_in flags agree with the derived sets", () => {
    for (const code of Outcome.options) {
      const flags = OUTCOMES[code].included_in;
      expect(flags.pdoom).toBe(DERIVED_OUTCOME_SETS.P_DOOM.includes(code as never));
      expect(flags.extinction).toBe(DERIVED_OUTCOME_SETS.P_EXTINCTION.includes(code as never));
      expect(flags.disempowerment).toBe(
        DERIVED_OUTCOME_SETS.P_DISEMPOWERMENT.includes(code as never),
      );
      expect(flags.collapse).toBe(DERIVED_OUTCOME_SETS.P_COLLAPSE.includes(code as never));
      expect(flags.biosphere).toBe(DERIVED_OUTCOME_SETS.P_BIOSPHERE.includes(code as never));
    }
    expect([...DERIVED_OUTCOME_SETS.P_DOOM]).toEqual(PDoomOutcome.options);
    expect(Outcome.options.filter((c) => !OUTCOMES[c].included_in.pdoom)).toEqual(["O0", "O1", "O2"]);
  });
});

describe("horizons", () => {
  it("HORIZONS has the seven keys in spec order", () => {
    expect([...HORIZON_KEYS]).toEqual(["1y", "3y", "5y", "10y", "25y", "2100", "eventual"]);
    expect([...Horizon.options].sort()).toEqual([...HORIZON_KEYS].sort());
    expect(HORIZONS.map((h) => h.key)).toEqual([...HORIZON_KEYS]);
    expect([...ForecastHorizon.options].sort()).toEqual([...HORIZON_KEYS, "custom"].sort());
  });

  it("duration horizons carry years; 2100 and eventual do not", () => {
    const byKey = Object.fromEntries(HORIZONS.map((h) => [h.key, h]));
    expect(byKey["1y"]?.years).toBe(1);
    expect(byKey["25y"]?.years).toBe(25);
    expect(byKey["2100"]?.years).toBeNull();
    expect(byKey["eventual"]?.years).toBeNull();
    expect(byKey["eventual"]?.kind).toBe("open_ended");
    expect(byKey["eventual"]?.note).toMatch(/not comparable/i);
  });
});

describe("driver families and source tiers", () => {
  it("DRIVER_FAMILIES covers D1–D10 with unique slugs", () => {
    expect(DRIVER_FAMILIES.map((d) => d.id)).toEqual(DriverFamily.options);
    expect(DRIVER_FAMILIES.map((d) => d.slug)).toEqual(DriverFamilySlug.options);
    expect(DRIVER_FAMILIES[0]).toMatchObject({ id: "D1", slug: "capability" });
    expect(DRIVER_FAMILIES[9]).toMatchObject({ id: "D10", slug: "resilience" });
  });

  it("SOURCE_TIERS covers 1–5 and encodes the brief's rules", () => {
    expect(SOURCE_TIERS.map((t) => t.tier)).toEqual([1, 2, 3, 4, 5]);
    for (const tier of SOURCE_TIERS) {
      expect(SourceTier.safeParse(tier.tier).success).toBe(true);
      expect(tier.rules.length).toBeGreaterThan(0);
    }
    const byTier = Object.fromEntries(SOURCE_TIERS.map((t) => [t.tier, t]));
    expect(byTier[5]?.may_alter_pdoom).toBe(false);
    expect(byTier[5]?.max_model_use_status).toBe("excluded");
    expect(byTier[4]?.requires_corroboration).toBe(true);
    expect(byTier[4]?.max_model_use_status).toBe("informational");
    expect(byTier[3]?.rules.join(" ")).toMatch(/candidate/);
    expect(byTier[1]?.rules.join(" ")).toMatch(/independent corroboration/);
    expect(SOURCE_TIERS.map((t) => t.default_multiplier)).toEqual([1.0, 0.9, 0.5, 0.0, 0.0]);
  });
});

describe("display tables", () => {
  it("EDITORIAL_RISK_LEVELS matches the enum and has descriptions", () => {
    expect(EDITORIAL_RISK_LEVELS.map((l) => l.level)).toEqual(EditorialRiskLevel.options);
    for (const level of EDITORIAL_RISK_LEVELS) expect(level.description.length).toBeGreaterThan(20);
  });

  it("INDEX_DEFINITIONS matches the enum and never claims to be a probability", () => {
    expect(INDEX_DEFINITIONS.map((i) => i.id)).toEqual(IndexId.options);
    for (const index of INDEX_DEFINITIONS) {
      expect(index.is_probability).toBe(false);
      expect(index.scale).toMatch(/0-100/);
    }
  });

  it("AGGREGATION_METHODS and CONFLICT_LABELS match their enums", () => {
    expect(AGGREGATION_METHODS.map((m) => m.method)).toEqual(AggregationMethod.options);
    expect(CONFLICT_LABELS.map((c) => c.label)).toEqual(ConflictLabel.options);
    for (const c of CONFLICT_LABELS) expect(c.display.length).toBeGreaterThan(0);
  });

  it("rounding follows the uncertainty label (§0.8)", () => {
    expect(Object.keys(ROUNDING_BY_UNCERTAINTY).sort()).toEqual([...UncertaintyLabel.options].sort());
    expect(ROUNDING_BY_UNCERTAINTY.extreme.points).toBe(5);
    expect(ROUNDING_BY_UNCERTAINTY.high.points).toBe(2);
    expect(ROUNDING_BY_UNCERTAINTY.moderate.points).toBe(1);
    expect(ROUNDING_BY_UNCERTAINTY.low.points).toBe(1);
  });

  it("SourceType lists the 23 spec values", () => {
    expect(SourceType.options).toHaveLength(23);
    expect(SourceType.options[0]).toBe("paper");
    expect(SourceType.options.at(-1)).toBe("other");
  });
});
