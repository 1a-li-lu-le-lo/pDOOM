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

export interface IndexGaugeProps {
  index: IndexValue;
  /** Full index name; defaults to the release's `label`. */
  title?: string;
  /** Caption text; defaults to the scale reminder and coverage. */
  description?: string;
}

/**
 * A 0–100 index. The label says what the scale means; it is never a probability.
 * Rendered as a figure with a caption and an accessible data table (build-spec §6).
 */
export function IndexGauge({ index, title = index.label, description }: IndexGaugeProps) {
  const v = index.value;
  const shown = v === null ? "not computed" : String(Math.round(v));
  const coverage = `${Math.round(index.coverage * 100)}%`;
  const caption = description ?? (v === null ? "not computed" : `index, not a probability · coverage ${coverage}`);
  return (
    <figure className="gauge" style={{ margin: 0, color: v === null ? "var(--c-insufficient)" : COLOR[index.index_id] }}>
      <div className="label">{INDEX_SHORT[index.index_id]}</div>
      <div className="num">{v === null ? "—" : Math.round(v)}</div>
      <div className="bar" aria-hidden="true">
        <span style={{ width: `${v ?? 0}%` }} />
      </div>
      <figcaption className="muted" style={{ fontSize: "var(--fs-xs)", color: "var(--c-text-3)" }}>
        {caption}
        <details>
          <summary>Data table</summary>
          <table>
            <caption>{title}</caption>
            <thead>
              <tr>
                <th>Index</th>
                <th>Value</th>
                <th>Scale</th>
                <th>Coverage</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>{title}</td>
                <td className="num">{shown}</td>
                <td>{index.scale}</td>
                <td className="num">{coverage}</td>
              </tr>
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
