// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { describe, expect, it } from "vitest";
import { fmtDate, outcomeSetLabel, pct, tidyInterval, titleCase, uncertaintyBadge } from "../lib/format";

describe("format helpers", () => {
  it("pct rounds by uncertainty label and never shows decimals", () => {
    expect(pct(0.0163, "high")).toBe("2%");
    expect(pct(0.0163, "low")).toBe("2%");
    expect(pct(0.0163, "extreme")).toBe("<1%");
    expect(pct(0.004, "moderate")).toBe("<1%");
    expect(pct(0.996, "low")).toBe(">99%");
    expect(pct(null)).toBe("—");
    expect(pct(0.1339, "high")).not.toMatch(/\./);
  });

  it("tidyInterval collapses equal endpoints only", () => {
    expect(tidyInterval("<1%–<1%")).toBe("<1%");
    expect(tidyInterval("<1%–14%")).toBe("<1%–14%");
    expect(tidyInterval("")).toBe("");
  });

  it("outcome set labels name the taxonomy without lowercasing the brand", () => {
    expect(outcomeSetLabel(["O3", "O4", "O5", "O6", "O7", "O8"])).toBe("p(DOOM): O3–O8 combined");
    expect(outcomeSetLabel(["O4", "O5"])).toMatch(/O4\+O5/);
    expect(outcomeSetLabel(["O6"])).toBe("Human extinction");
  });

  it("dates, case and badges", () => {
    expect(fmtDate("2026-09-26T22:00:00Z")).toBe("2026-09-26");
    expect(fmtDate(null)).toBe("—");
    expect(titleCase("verified_search")).toBe("Verified search");
    expect(uncertaintyBadge("extreme").text).toBe("extreme uncertainty");
    expect(uncertaintyBadge(null).text).toBe("unknown");
  });
});
