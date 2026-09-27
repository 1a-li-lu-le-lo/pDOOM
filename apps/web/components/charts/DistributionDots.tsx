// Copyright NU Cybernetics. p(DOOM) — research prototype.
export interface Dot {
  label: string;
  value: number;
  group: string;
  colorVar?: string;
}

/** One row per group; each forecast is a dot on a shared probability axis. */
export function DistributionDots({ dots, title, description, max }: { dots: Dot[]; title: string; description: string; max?: number }) {
  const groups = [...new Set(dots.map((d) => d.group))];
  const raw = Math.min(1, Math.max(0.01, ...dots.map((d) => d.value)) * 1.2);
  const top = max ?? ([0.01, 0.02, 0.05, 0.1, 0.2, 0.3, 0.5, 1].find((c) => c >= raw) ?? 1);
  const W = 640;
  const rowH = 30;
  const padL = 190;
  const H = groups.length * rowH + 30;
  const x = (v: number) => padL + (Math.min(top, v) / top) * (W - padL - 16);
  const fmt = (v: number) => (v * 100 < 1 && v > 0 ? `${Math.round(v * 1000) / 10}%` : `${Math.round(v * 100)}%`);
  return (
    <figure>
      <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${title}, ${dots.length} forecasts across ${groups.length} groups`}>
        <g className="grid">
          {[0, 0.25, 0.5, 0.75, 1].map((f) => (
            <line key={f} x1={x(f * top)} x2={x(f * top)} y1={0} y2={H - 22} />
          ))}
        </g>
        {[0, 0.5, 1].map((f) => (
          <text key={f} x={x(f * top)} y={H - 6} textAnchor={f === 0 ? "start" : f === 1 ? "end" : "middle"}>
            {fmt(f * top)}
          </text>
        ))}
        {groups.map((g, gi) => (
          <g key={g}>
            <text x={0} y={gi * rowH + 19}>
              {g.length > 30 ? g.slice(0, 29) + "…" : g}
            </text>
            <line x1={padL} x2={W - 16} y1={gi * rowH + 15} y2={gi * rowH + 15} stroke="var(--c-border)" />
            {dots
              .filter((d) => d.group === g)
              .map((d) => (
                <circle key={d.label} cx={x(d.value)} cy={gi * rowH + 15} r={6} fill={d.colorVar ?? "var(--c-evidence)"} opacity={0.85}>
                  <title>{`${d.label}: ${fmt(d.value)}`}</title>
                </circle>
              ))}
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
                <th>Group</th>
                <th>Forecast</th>
                <th className="num">Value</th>
              </tr>
            </thead>
            <tbody>
              {dots.map((d) => (
                <tr key={d.label}>
                  <td>{d.group}</td>
                  <td>{d.label}</td>
                  <td className="num">{fmt(d.value)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
