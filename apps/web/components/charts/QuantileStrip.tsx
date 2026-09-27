// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { ReactNode } from "react";
import { roundForDisplay, roundingStep } from "@pdoom/model-core";

export interface QuantileStripProps {
  p05: number;
  p25: number;
  p50: number;
  p75: number;
  p95: number;
  /** Upper bound of the axis (probability). Defaults to a round value above p95. */
  max?: number;
  /** Display rounding step in percentage points (build-spec §0.8); pass `roundingStep(uncertainty)`. */
  step?: number;
  /** What the outer band is: a model interval (default) or, for external aggregates, the min–max range of member forecasts. */
  rangeLabel?: string;
  label: string;
  description: string;
  colorVar?: string;
  compact?: boolean;
  children?: ReactNode;
}

function niceMax(p95: number): number {
  const candidates = [0.01, 0.02, 0.05, 0.1, 0.2, 0.3, 0.5, 1];
  return candidates.find((c) => c >= p95 * 1.15) ?? 1;
}

/**
 * Interval bar: p05–p95 as a thin band, p25–p75 as a thick band, p50 as a tick.
 * Ring thickness (band width) is the visual grammar for the confidence interval
 * everywhere on the site. Server-rendered SVG with an accessible table.
 */
export function QuantileStrip({ p05, p25, p50, p75, p95, max, label, description, colorVar = "var(--c-evidence)", compact = false, children, step = roundingStep("high"), rangeLabel = "5th to 95th percentile" }: QuantileStripProps) {
  const top = max ?? niceMax(p95);
  const W = 600;
  const H = compact ? 34 : 40;
  const x = (p: number) => Math.min(W, Math.max(0, (p / top) * W));
  const pctLabel = (p: number) => roundForDisplay(p, step);
  // Axis ticks are geometry, not estimates; only whole-point ticks get a label so no decimal is ever shown.
  const isWholePoint = (t: number) => Math.abs(t * 100 - Math.round(t * 100)) < 1e-9;
  const tickLabel = (t: number) => roundForDisplay(t, 1);
  const ticks = [0, top / 4, top / 2, (3 * top) / 4, top];
  return (
    <figure className="quantile-strip">
      <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${label}: median ${pctLabel(p50)}, ${rangeLabel} ${pctLabel(p05)} to ${pctLabel(p95)}`} preserveAspectRatio="none" style={{ height: H }}>
        <g className="grid">
          {ticks.map((t) => (
            <line key={t} x1={x(t)} x2={x(t)} y1={0} y2={H} />
          ))}
        </g>
        <rect x={x(p05)} y={compact ? 12 : 14} width={Math.max(2, x(p95) - x(p05))} height={compact ? 10 : 12} fill={colorVar} opacity={0.28} rx={3} />
        <rect x={x(p25)} y={compact ? 10 : 10} width={Math.max(2, x(p75) - x(p25))} height={compact ? 14 : 20} fill={colorVar} opacity={0.65} rx={3} />
        <rect x={x(p50) - 1.5} y={compact ? 6 : 4} width={3} height={compact ? 22 : 32} fill={colorVar} />
      </svg>
      {!compact ? (
        <div className="strip-ticks" aria-hidden="true">
          {ticks.map((t) => (
            <span key={`t${t}`}>{isWholePoint(t) ? tickLabel(t) : ""}</span>
          ))}
        </div>
      ) : null}
      <figcaption>
        {description}
        {children}
        <details>
          <summary>Data table</summary>
          <table>
            <caption>{label}: quantiles as probabilities</caption>
            <thead>
              <tr>
                <th>p05</th>
                <th>p25</th>
                <th>p50 (median)</th>
                <th>p75</th>
                <th>p95</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="num">{pctLabel(p05)}</td>
                <td className="num">{pctLabel(p25)}</td>
                <td className="num">{pctLabel(p50)}</td>
                <td className="num">{pctLabel(p75)}</td>
                <td className="num">{pctLabel(p95)}</td>
              </tr>
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
