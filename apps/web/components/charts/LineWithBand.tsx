// Copyright NU Cybernetics. p(DOOM) — research prototype.
export interface BandPoint {
  x: string;
  low: number;
  mid: number;
  high: number;
}

/** Cumulative-probability curve across horizons with a p05–p95 band. Categorical x axis. */
export function LineWithBand({ points, title, description, max, colorVar = "var(--c-evidence)" }: { points: BandPoint[]; title: string; description: string; max?: number; colorVar?: string }) {
  const W = 640;
  const H = 220;
  const padL = 44;
  const padB = 28;
  const raw = Math.max(0.01, ...points.map((p) => p.high)) * 1.1;
  const top = max ?? ([0.01, 0.02, 0.05, 0.1, 0.2, 0.3, 0.5, 1].find((c) => c >= raw) ?? 1);
  const xs = (i: number) => padL + (i * (W - padL - 12)) / Math.max(1, points.length - 1);
  const ys = (v: number) => H - padB - (Math.min(top, v) / top) * (H - padB - 10);
  const path = (key: "low" | "mid" | "high") => points.map((p, i) => `${i === 0 ? "M" : "L"}${xs(i).toFixed(1)},${ys(p[key]).toFixed(1)}`).join(" ");
  const band = `${path("high")} ${[...points].reverse().map((p, j) => `L${xs(points.length - 1 - j).toFixed(1)},${ys(p.low).toFixed(1)}`).join(" ")} Z`;
  const fmt = (v: number) => (v * 100 < 1 && v > 0 ? `${Math.round(v * 1000) / 10}%` : `${Math.round(v * 100)}%`);
  return (
    <figure>
      <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${title}: median from ${fmt(points[0]?.mid ?? 0)} to ${fmt(points[points.length - 1]?.mid ?? 0)}`}>
        <g className="grid">
          {[0, 0.25, 0.5, 0.75, 1].map((f) => (
            <line key={f} x1={padL} x2={W - 12} y1={ys(f * top)} y2={ys(f * top)} />
          ))}
        </g>
        <g className="axis">
          {[0, 0.5, 1].map((f) => (
            <text key={f} x={padL - 6} y={ys(f * top) + 4} textAnchor="end">
              {fmt(f * top)}
            </text>
          ))}
        </g>
        <path d={band} fill={colorVar} opacity={0.18} />
        <path d={path("mid")} fill="none" stroke={colorVar} strokeWidth={2.5} />
        {points.map((p, i) => (
          <g key={p.x}>
            <circle cx={xs(i)} cy={ys(p.mid)} r={3.5} fill={colorVar} />
            <text x={xs(i)} y={H - 8} textAnchor="middle">
              {p.x}
            </text>
          </g>
        ))}
      </svg>
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
                  <td className="num">{fmt(p.low)}</td>
                  <td className="num">{fmt(p.mid)}</td>
                  <td className="num">{fmt(p.high)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
