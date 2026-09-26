// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { IndexValue } from "@pdoom/schemas";
import { INDEX_SHORT } from "@/lib/format";

const COLOR: Record<string, string> = {
  evidence_pressure: "var(--c-evidence)",
  capability_pressure: "var(--c-risk)",
  control_strength: "var(--c-safeguard)",
  incident_pressure: "var(--c-risk)",
  uncertainty: "var(--c-uncertainty)",
  agentic_infrastructure_risk: "var(--c-risk)",
  attention: "var(--c-insufficient)",
};

/** A 0–100 index. The label says what the scale means; it is never a probability. */
export function IndexGauge({ index }: { index: IndexValue }) {
  const v = index.value;
  return (
    <div className="gauge" style={{ color: v === null ? "var(--c-insufficient)" : COLOR[index.index_id] }}>
      <div className="label">{INDEX_SHORT[index.index_id]}</div>
      <div className="num" aria-label={`${index.label}: ${v === null ? "not computed" : Math.round(v)} out of 100`}>
        {v === null ? "—" : Math.round(v)}
      </div>
      <div className="bar" aria-hidden="true">
        <span style={{ width: `${v ?? 0}%` }} />
      </div>
      <div className="muted" style={{ fontSize: "var(--fs-xs)", color: "var(--c-text-3)" }}>
        {v === null ? "not computed" : `index, not a probability · coverage ${Math.round(index.coverage * 100)}%`}
      </div>
    </div>
  );
}
