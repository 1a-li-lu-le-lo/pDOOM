// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { UncertaintyLabelValue } from "@pdoom/schemas";

/** Display rounding step in percentage points per uncertainty label (build-spec §0.8). */
export function roundingStep(label: UncertaintyLabelValue): number {
  switch (label) {
    case "extreme":
      return 5;
    case "high":
      return 2;
    default:
      return 1;
  }
}

/** Renders a probability as a rounded percentage; never decimals; "<1%" and ">99%" guards. */
export function roundForDisplay(p: number, step: number): string {
  const s = step > 0 ? step : 1;
  const pct = p * 100;
  const rounded = Math.round(pct / s) * s;
  if (p > 0 && rounded < 1) return "<1%";
  if (p < 1 && rounded > 99) return ">99%";
  if (rounded <= 0) return "0%";
  if (rounded >= 100) return "100%";
  return `${rounded}%`;
}

/** "3%–30%" with an en dash. */
export function formatInterval(lo: number, hi: number, step: number): string {
  return `${roundForDisplay(lo, step)}–${roundForDisplay(hi, step)}`;
}

/** "1 in N" phrasing for the probability comparator (magnitude only). */
export function oneIn(p: number): string {
  if (p <= 0) return "0";
  const n = Math.round(1 / p);
  return `1 in ${n.toLocaleString("en-US")}`;
}
