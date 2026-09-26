// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Estimate } from "@pdoom/schemas";
import { QuantileStrip } from "../charts/QuantileStrip";
import { STATUS_LABEL, badgeClass, fmtDate, horizonLabel, outcomeSetLabel, tidyInterval, uncertaintyBadge } from "@/lib/format";

const STATUS_COLOR: Record<string, string> = {
  external_aggregate: "var(--c-evidence)",
  research_mode: "var(--c-disagreement)",
  insufficiently_calibrated: "var(--c-insufficient)",
  official: "var(--c-evidence)",
  user_scenario: "var(--c-uncertainty)",
};

/**
 * Every probability on the site is rendered through this card so that the
 * outcome, horizon, status, interval, disagreement and data cutoff travel with it.
 */
export function EstimateCard({ e, compact = false, hideStrip = false }: { e: Estimate; compact?: boolean; hideStrip?: boolean }) {
  const q = e.quantiles;
  const ub = uncertaintyBadge(e.uncertainty);
  const withheld = e.status === "insufficiently_calibrated";
  return (
    <article className="card estimate-card" aria-label={`${STATUS_LABEL[e.status]}: ${outcomeSetLabel(e.outcome_set)}, ${horizonLabel(e.horizon)}`}>
      <div className="row">
        <span className={badgeClass(e.status === "research_mode" ? "disagreement" : withheld ? "insufficient" : "evidence")}>{STATUS_LABEL[e.status]}</span>
        <span className="badge">{horizonLabel(e.horizon)}</span>
        {e.disagreement ? <span className={badgeClass("disagreement")}>{e.disagreement} disagreement</span> : null}
        <span className={ub.cls}>{ub.text}</span>
      </div>
      <div className="eyebrow">{outcomeSetLabel(e.outcome_set)}</div>
      <div className="value" style={{ color: STATUS_COLOR[e.status], fontSize: withheld ? "var(--fs-xl)" : undefined }}>
        {e.display.central}
      </div>
      {q && !withheld ? <div className="interval">plausible interval {tidyInterval(e.display.interval)} · rounded to {e.rounding_rule.replace("nearest_", "nearest ")} points</div> : null}
      {q && !withheld && !hideStrip ? (
        <QuantileStrip p05={q.p05} p25={q.p25} p50={q.p50} p75={q.p75} p95={q.p95} label={`${outcomeSetLabel(e.outcome_set)}, ${horizonLabel(e.horizon)}`} description={e.display.note} colorVar={STATUS_COLOR[e.status]} compact={compact} />
      ) : (
        <p className="muted" style={{ fontSize: "var(--fs-sm)" }}>{e.display.note}</p>
      )}
      <div className="estimate-meta">
        <span>model {e.producer}</span>
        <span>origin {fmtDate(e.forecast_origin_date)}</span>
        <span>evidence through {fmtDate(e.last_evidence_date)}</span>
        {e.source_coverage.forecast_count ? <span>{e.source_coverage.forecast_count} forecasts, {e.source_coverage.population_count} populations</span> : null}
      </div>
      {!compact ? (
        <details>
          <summary>Conditioning and assumptions</summary>
          <p style={{ fontSize: "var(--fs-sm)" }}>{e.conditioning}</p>
          <ul style={{ fontSize: "var(--fs-sm)" }}>
            {e.assumptions.map((a) => (
              <li key={a}>{a}</li>
            ))}
          </ul>
          <p className="cite">
            Method: <a href={`/method/${e.method_ref.replace("docs/method/", "").replace(/\.md.*$/, "")}`}>{e.method_ref}</a>
          </p>
        </details>
      ) : null}
    </article>
  );
}
