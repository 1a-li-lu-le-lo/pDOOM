// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { HORIZONS, OUTCOMES, type EstimateStatusValue, type IndexIdValue, type UncertaintyLabelValue } from "@pdoom/schemas";
import { roundForDisplay, roundingStep } from "@pdoom/model-core";

export function horizonLabel(key: string): string {
  return HORIZONS.find((h) => h.key === key)?.label ?? key;
}

export function outcomeLabel(code: string): string {
  return OUTCOMES[code as keyof typeof OUTCOMES]?.label ?? code;
}

export function outcomeSetLabel(set: readonly string[]): string {
  const key = set.join("+");
  if (key === "O3+O4+O5+O6+O7+O8") return "p(DOOM): O3–O8 combined";
  if (key === "O4+O5") return "Collapse or near-extinction (O4+O5)";
  if (key === "O4+O5+O6") return "Catastrophe: collapse, near-extinction or extinction (O4–O6)";
  if (key === "O3+O6") return "Extinction or permanent disempowerment (O3+O6)";
  return set.map(outcomeLabel).join(" or ");
}

export const STATUS_LABEL: Record<EstimateStatusValue, string> = {
  official: "Official",
  insufficiently_calibrated: "Official value withheld",
  external_aggregate: "External forecast aggregate",
  research_mode: "Research mode",
  user_scenario: "Your scenario",
};

export const INDEX_SHORT: Record<IndexIdValue, string> = {
  evidence_pressure: "Evidence pressure",
  capability_pressure: "Capability pressure",
  control_strength: "Control strength",
  incident_pressure: "Incident pressure",
  uncertainty: "Uncertainty",
  agentic_infrastructure_risk: "Agentic infrastructure risk",
  attention: "Attention",
};

export function pct(p: number | null | undefined, uncertainty: UncertaintyLabelValue = "high"): string {
  if (p === null || p === undefined) return "—";
  return roundForDisplay(p, roundingStep(uncertainty));
}

export function fmtDate(d: string | null | undefined): string {
  if (!d) return "—";
  return d.length > 10 ? d.slice(0, 10) : d;
}

export function titleCase(s: string): string {
  return s.replace(/_/g, " ").replace(/^\w/, (c) => c.toUpperCase());
}

export function badgeClass(kind: "safeguard" | "uncertainty" | "risk" | "disagreement" | "resilience" | "insufficient" | "evidence"): string {
  return `badge badge-${kind}`;
}

export function uncertaintyBadge(label: UncertaintyLabelValue | null | undefined) {
  if (!label) return { cls: badgeClass("insufficient"), text: "unknown" };
  const cls = label === "low" || label === "moderate" ? badgeClass("evidence") : badgeClass("uncertainty");
  return { cls, text: `${label} uncertainty` };
}
