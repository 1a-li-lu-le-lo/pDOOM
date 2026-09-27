// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { roundForDisplay } from "@pdoom/model-core";

export interface Dot {
  /** Stable unique id (e.g. the forecast id); used as the React key when provided. */
  id?: string;
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
  // Axis ticks are whole points; dot values are member forecasts quoted as their sources state them.
  const fmt = (v: number) => roundForDisplay(v, 1);
  const asRecorded = (v: number) => `${Math.round(v * 1000) / 10}%`;
  const key = (d: Dot) => d.id ?? `${d.group} · ${d.label}`;
  // Group labels are "<outcome set> · <horizon>". The horizon is always shown in full on its own
  // line so rows across horizons stay distinguishable; only the outcome part is shortened.
  const MAX_CHARS = 30;
  const clip = (s: string) => {
    if (s.length <= MAX_CHARS) return s;
    const cut = s.slice(0, MAX_CHARS - 1);
    const atWord = cut.replace(/\s+\S*$/, "");
    return (atWord.length > 0 ? atWord : cut) + "…";
  };
  const splitLabel = (g: string): [string, string | null] => {
    const i = g.lastIndexOf(" · ");
    return i === -1 ? [g, null] : [g.slice(0, i), g.slice(i + 3)];
  };
  return (
    <figure>
      <div className="chart-scroll" tabIndex={0} role="group" aria-label={`${title} (scrolls sideways on narrow screens)`}>
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
        {groups.map((g, gi) => {
          const [outcome, horizon] = splitLabel(g);
          return (
            <g key={g}>
              <text x={0} y={horizon ? gi * rowH + 12 : gi * rowH + 19}>
                <title>{g}</title>
                <tspan x={0}>{clip(outcome)}</tspan>
                {horizon ? (
                  <tspan x={0} dy={12}>
                    {horizon}
                  </tspan>
                ) : null}
              </text>
              <line x1={padL} x2={W - 16} y1={gi * rowH + 15} y2={gi * rowH + 15} stroke="var(--c-border)" />
              {dots
                .filter((d) => d.group === g)
                .map((d) => (
                  <circle key={key(d)} cx={x(d.value)} cy={gi * rowH + 15} r={6} fill={d.colorVar ?? "var(--c-evidence)"} opacity={0.85}>
                    <title>{`${d.label}: ${asRecorded(d.value)} (as the source states it)`}</title>
                  </circle>
                ))}
            </g>
          );
        })}
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
                <th>Group</th>
                <th>Forecast</th>
                <th className="num">Value</th>
              </tr>
            </thead>
            <tbody>
              {dots.map((d) => (
                <tr key={key(d)}>
                  <td>{d.group}</td>
                  <td>{d.label}</td>
                  <td className="num">{asRecorded(d.value)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
