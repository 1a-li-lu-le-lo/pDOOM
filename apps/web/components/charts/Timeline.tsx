// Copyright NU Cybernetics. p(DOOM) — research prototype.
export interface TimelinePoint {
  date: string;
  value: number;
  label: string;
  low?: number | null;
  high?: number | null;
  note?: string;
}

/**
 * Dated measurements on a time axis with optional confidence whiskers
 * (server-rendered SVG). A log scale is used when the range spans more than
 * two orders of magnitude, and the caption says so.
 */
export function Timeline({ points, title, description, unit, log }: { points: TimelinePoint[]; title: string; description: string; unit: string; log?: boolean }) {
  const pts = [...points].filter((p) => p.value > 0 || !log).sort((a, b) => a.date.localeCompare(b.date));
  if (!pts.length) return <p className="muted">No measurements recorded.</p>;
  const W = 640;
  const H = 240;
  const padL = 56;
  const padB = 34;
  const t0 = Date.parse(pts[0]!.date);
  const t1 = Math.max(Date.parse(pts[pts.length - 1]!.date), t0 + 86_400_000 * 30);
  const vals = pts.flatMap((p) => [p.value, p.high ?? p.value, p.low ?? p.value]).filter((v) => v > 0 || !log);
  const vmax = Math.max(...vals);
  const vmin = log ? Math.min(...vals.filter((v) => v > 0)) : 0;
  const useLog = !!log && vmax / Math.max(vmin, 1e-9) > 100;
  const fy = (v: number) => (useLog ? Math.log10(Math.max(v, vmin)) : v);
  const ymin = fy(vmin);
  const ymax = fy(vmax) + (useLog ? 0.1 : (vmax - vmin) * 0.1);
  const xs = (d: string) => padL + ((Date.parse(d) - t0) / (t1 - t0)) * (W - padL - 14);
  const ys = (v: number) => H - padB - ((fy(v) - ymin) / Math.max(1e-9, ymax - ymin)) * (H - padB - 12);
  const fmtV = (v: number) => (v >= 100 ? Math.round(v).toLocaleString("en-US") : v >= 10 ? v.toFixed(0) : v.toFixed(1));
  const ticks = useLog ? Array.from({ length: Math.ceil(ymax) - Math.floor(ymin) + 1 }, (_, i) => 10 ** (Math.floor(ymin) + i)).filter((t) => t >= vmin && t <= vmax * 1.3) : [0, 0.25, 0.5, 0.75, 1].map((f) => vmin + (vmax - vmin) * f);
  const years = [...new Set(pts.map((p) => p.date.slice(0, 4)))];
  return (
    <figure>
      <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${title}: ${pts.length} measurements from ${pts[0]!.date} to ${pts[pts.length - 1]!.date}`}>
        <g className="grid">
          {ticks.map((t) => (
            <line key={t} x1={padL} x2={W - 14} y1={ys(t)} y2={ys(t)} />
          ))}
        </g>
        {ticks.map((t) => (
          <text key={`l${t}`} x={padL - 6} y={ys(t) + 4} textAnchor="end">
            {fmtV(t)}
          </text>
        ))}
        {years.map((y) => (
          <text key={y} x={xs(`${y}-07-01`)} y={H - 8} textAnchor="middle">
            {y}
          </text>
        ))}
        <path d={pts.map((p, i) => `${i === 0 ? "M" : "L"}${xs(p.date).toFixed(1)},${ys(p.value).toFixed(1)}`).join(" ")} fill="none" stroke="var(--c-risk)" strokeOpacity={0.5} strokeWidth={1.5} />
        {pts.map((p) => (
          <g key={`${p.date}-${p.label}`}>
            {p.low != null && p.high != null ? <line x1={xs(p.date)} x2={xs(p.date)} y1={ys(p.high)} y2={ys(p.low)} stroke="var(--c-risk)" strokeOpacity={0.5} /> : null}
            <circle cx={xs(p.date)} cy={ys(p.value)} r={4.5} fill="var(--c-risk)" stroke="var(--c-bg)" strokeWidth={1.5}>
              <title>
                {p.label}: {fmtV(p.value)} {unit} ({p.date})
              </title>
            </circle>
          </g>
        ))}
        <text x={padL} y={12}>
          {unit}
          {useLog ? " (log scale)" : ""}
        </text>
      </svg>
      <figcaption>
        {description}
        {useLog ? " The vertical axis is logarithmic." : ""}
        <details>
          <summary>Data table</summary>
          <table>
            <caption>{title}</caption>
            <thead>
              <tr>
                <th>Date</th>
                <th>Model</th>
                <th className="num">{unit}</th>
                <th>Interval</th>
                <th>Note</th>
              </tr>
            </thead>
            <tbody>
              {pts.map((p) => (
                <tr key={`${p.date}-${p.label}-r`}>
                  <td>{p.date}</td>
                  <td>{p.label}</td>
                  <td className="num">{fmtV(p.value)}</td>
                  <td>{p.low != null && p.high != null ? `${fmtV(p.low)}–${fmtV(p.high)}` : "—"}</td>
                  <td className="cite">{p.note ?? ""}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
