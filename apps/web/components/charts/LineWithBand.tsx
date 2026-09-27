// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { roundForDisplay } from "@pdoom/model-core";

export interface BandPoint {
  /** Full horizon label; used in the data table and the accessible name. */
  x: string;
  /** Optional short label for the drawn axis (the full `x` still appears in the table). */
  short?: string;
  low: number;
  mid: number;
  high: number;
  /** Display rounding step in points for this point's estimate (build-spec §0.8, `roundingStep(uncertainty)`). Whole points when omitted. */
  step?: number;
}

/** Split "Eventually (no fixed date)" into two axis lines; other labels stay whole. */
function axisLines(label: string): string[] {
  const i = label.indexOf(" (");
  return i > 0 ? [label.slice(0, i), label.slice(i + 1)] : [label];
}

/** Cumulative-probability curve across horizons with a p05–p95 band. Categorical x axis. */
export function LineWithBand({ points, title, description, max, colorVar = "var(--c-evidence)" }: { points: BandPoint[]; title: string; description: string; max?: number; colorVar?: string }) {
  const W = 640;
  const padL = 44;
  const padR = 12;
  const n = points.length;
  // Axis label layout: labels are anchored inward at the ends so none escapes the drawing, and when the
  // widest label would not fit its slot the labels alternate between two rows instead of colliding.
  const AXIS_FONT = 12;
  const LINE_H = AXIS_FONT + 3;
  const labels = points.map((p) => axisLines(p.short ?? p.x));
  const estWidth = (s: string) => s.length * AXIS_FONT * 0.55;
  const slot = (W - padL - padR) / Math.max(1, n - 1);
  const widest = Math.max(0, ...labels.flat().map(estWidth));
  const extent = (i: number) => {
    const w = Math.max(0, ...(labels[i] ?? []).map(estWidth));
    const x = padL + (i * (W - padL - padR)) / Math.max(1, n - 1);
    const a = n > 1 && i === 0 ? 0 : n > 1 && i === n - 1 ? 1 : 0.5;
    return [x - w * a, x + w * (1 - a)];
  };
  const touching = labels.some((_, i) => i < n - 1 && extent(i)[1]! + 8 > extent(i + 1)[0]!);
  const stagger = n > 2 && (widest > slot * 0.9 || touching);
  const maxLines = Math.max(1, ...labels.map((l) => l.length));
  const rowH = maxLines * LINE_H;
  const padB = 8 + (stagger ? 2 : 1) * rowH;
  const H = 190 + padB;
  const raw = Math.max(0.01, ...points.map((p) => p.high)) * 1.1;
  const top = max ?? ([0.01, 0.02, 0.05, 0.1, 0.2, 0.3, 0.5, 1].find((c) => c >= raw) ?? 1);
  const xs = (i: number) => padL + (i * (W - padL - padR)) / Math.max(1, n - 1);
  const ys = (v: number) => H - padB - (Math.min(top, v) / top) * (H - padB - 10);
  const anchor = (i: number) => (n > 1 && i === 0 ? "start" : n > 1 && i === n - 1 ? "end" : "middle");
  const labelY = (i: number, line: number) => H - padB + AXIS_FONT + 2 + (stagger ? i % 2 : 0) * rowH + line * LINE_H;
  const path = (key: "low" | "mid" | "high") => points.map((p, i) => `${i === 0 ? "M" : "L"}${xs(i).toFixed(1)},${ys(p[key]).toFixed(1)}`).join(" ");
  const band = `${path("high")} ${[...points].reverse().map((p, j) => `L${xs(n - 1 - j).toFixed(1)},${ys(p.low).toFixed(1)}`).join(" ")} Z`;
  // Values follow the release rounding for their estimate (build-spec §0.8); a positive value under one point reads "<1%".
  const fmt = (v: number, step = 1) => roundForDisplay(v, step);
  const first = points[0];
  const last = points[n - 1];
  return (
    <figure>
      <div className="chart-scroll" tabIndex={0} role="group" aria-label={`${title} (scrolls sideways on narrow screens)`}>
      <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${title}: median from ${fmt(first?.mid ?? 0, first?.step)} to ${fmt(last?.mid ?? 0, last?.step)}`}>
        <g className="grid">
          {[0, 0.25, 0.5, 0.75, 1].map((f) => (
            <line key={f} x1={padL} x2={W - padR} y1={ys(f * top)} y2={ys(f * top)} />
          ))}
        </g>
        <g className="axis">
          {[0, 0.5, 1].map((f) => (
            <text key={f} x={padL - 6} y={ys(f * top) + 4} textAnchor="end" style={{ fontSize: AXIS_FONT }}>
              {fmt(f * top)}
            </text>
          ))}
        </g>
        <path d={band} fill={colorVar} opacity={0.18} />
        <path d={path("mid")} fill="none" stroke={colorVar} strokeWidth={2.5} />
        {points.map((p, i) => (
          <g key={p.x}>
            <circle cx={xs(i)} cy={ys(p.mid)} r={3.5} fill={colorVar} />
            {(labels[i] ?? []).map((line, k) => (
              <text key={k} x={xs(i)} y={labelY(i, k)} textAnchor={anchor(i)} style={{ fontSize: AXIS_FONT }}>
                {line}
              </text>
            ))}
          </g>
        ))}
      </svg>
      </div>
      <figcaption>
        {description}
        <details>
          <summary>Data table</summary>
          <table>
            <caption>{title}</caption>
            <thead>
              <tr>
                <th>Horizon</th>
                <th className="num">p05</th>
                <th className="num">median</th>
                <th className="num">p95</th>
              </tr>
            </thead>
            <tbody>
              {points.map((p) => (
                <tr key={p.x}>
                  <td>{p.x}</td>
                  <td className="num">{fmt(p.low, p.step)}</td>
                  <td className="num">{fmt(p.mid, p.step)}</td>
                  <td className="num">{fmt(p.high, p.step)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
