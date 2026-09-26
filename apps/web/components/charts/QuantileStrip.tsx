// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { ReactNode } from "react";

export interface QuantileStripProps {
  p05: number;
  p25: number;
  p50: number;
  p75: number;
  p95: number;
  /** Upper bound of the axis (probability). Defaults to a round value above p95. */
  max?: number;
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
export function QuantileStrip({ p05, p25, p50, p75, p95, max, label, description, colorVar = "var(--c-evidence)", compact = false, children }: QuantileStripProps) {
  const top = max ?? niceMax(p95);
  const W = 600;
  const H = compact ? 34 : 56;
  const x = (p: number) => Math.min(W, Math.max(0, (p / top) * W));
  const pctLabel = (p: number) => `${Math.round(p * 1000) / 10}%`;
  const ticks = [0, top / 4, top / 2, (3 * top) / 4, top];
  return (
    <figure className="quantile-strip">
      <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${label}: median ${pctLabel(p50)}, 5th to 95th percentile ${pctLabel(p05)} to ${pctLabel(p95)}`} preserveAspectRatio="none" style={{ height: H }}>
        <g className="grid">
          {ticks.map((t) => (
            <line key={t} x1={x(t)} x2={x(t)} y1={0} y2={compact ? H : H - 16} />
          ))}
        </g>
        <rect x={x(p05)} y={compact ? 12 : 14} width={Math.max(2, x(p95) - x(p05))} height={compact ? 10 : 12} fill={colorVar} opacity={0.28} rx={3} />
        <rect x={x(p25)} y={compact ? 10 : 10} width={Math.max(2, x(p75) - x(p25))} height={compact ? 14 : 20} fill={colorVar} opacity={0.65} rx={3} />
        <rect x={x(p50) - 1.5} y={compact ? 6 : 4} width={3} height={compact ? 22 : 32} fill={colorVar} />
        {!compact &&
          ticks.map((t) => (
            <text key={`t${t}`} x={x(t)} y={H - 2} textAnchor={t === 0 ? "start" : t === top ? "end" : "middle"}>
              {pctLabel(t)}
            </text>
          ))}
      </svg>
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
