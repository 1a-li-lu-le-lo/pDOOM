// Copyright NU Cybernetics. p(DOOM) — research prototype.
export interface FlowNode {
  id: string;
  label: string;
  colorVar?: string;
}
export interface Flow {
  from: string;
  to: string;
  value: number;
}

/**
 * Sankey-lite: two columns of nodes joined by ribbons whose thickness is the
 * flow value (a count, never a probability). Deterministic layout; server-
 * rendered SVG with a flows table as the accessible form.
 */
export function FlowDiagram({ left, right, flows, title, description, unit = "scenarios" }: { left: FlowNode[]; right: FlowNode[]; flows: Flow[]; title: string; description: string; unit?: string }) {
  const W = 900;
  const H = Math.max(280, Math.max(left.length, right.length) * 44 + 40);
  const colW = 240;
  const gap = 14;
  const padY = 20;
  const total = (id: string, side: "from" | "to") => flows.filter((f) => f[side] === id).reduce((s, f) => s + f.value, 0);
  const layout = (nodes: FlowNode[], side: "from" | "to") => {
    const sum = nodes.reduce((s, n) => s + total(n.id, side), 0) || 1;
    const avail = H - padY * 2 - gap * (nodes.length - 1);
    let y = padY;
    return new Map(
      nodes.map((n) => {
        const h = Math.max(14, (total(n.id, side) / sum) * avail);
        const box = { y, h, used: 0 };
        y += h + gap;
        return [n.id, box];
      }),
    );
  };
  const L = layout(left, "from");
  const R = layout(right, "to");
  const scaleL = (id: string) => (L.get(id)!.h || 1) / (total(id, "from") || 1);
  const scaleR = (id: string) => (R.get(id)!.h || 1) / (total(id, "to") || 1);
  const x0 = colW;
  const x1 = W - colW;
  const ribbons = flows
    .filter((f) => f.value > 0 && L.has(f.from) && R.has(f.to))
    .map((f) => {
      const a = L.get(f.from)!;
      const b = R.get(f.to)!;
      const ha = f.value * scaleL(f.from);
      const hb = f.value * scaleR(f.to);
      const ya = a.y + a.used;
      const yb = b.y + b.used;
      a.used += ha;
      b.used += hb;
      const cx = (x0 + x1) / 2;
      const d = `M${x0},${ya.toFixed(1)} C${cx},${ya.toFixed(1)} ${cx},${yb.toFixed(1)} ${x1},${yb.toFixed(1)} L${x1},${(yb + hb).toFixed(1)} C${cx},${(yb + hb).toFixed(1)} ${cx},${(ya + ha).toFixed(1)} ${x0},${(ya + ha).toFixed(1)} Z`;
      return { ...f, d, color: left.find((n) => n.id === f.from)?.colorVar ?? "var(--c-text-3)" };
    });
  const name = (nodes: FlowNode[], id: string) => nodes.find((n) => n.id === id)?.label ?? id;
  return (
    <figure>
      <div className="chart-scroll" tabIndex={0} role="group" aria-label={`${title} (scrolls sideways on narrow screens)`}>
      <svg className="chart flow-diagram" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${title}: ${flows.length} flows between ${left.length} groups and ${right.length} outcomes; the table lists every flow`}>
        {ribbons.map((r) => (
          <path key={`${r.from}-${r.to}`} d={r.d} fill={r.color} fillOpacity={0.28} stroke={r.color} strokeOpacity={0.5} strokeWidth={0.5}>
            <title>{`${name(left, r.from)} → ${name(right, r.to)}: ${r.value} ${r.value === 1 ? unit.replace(/s$/, "") : unit}`}</title>
          </path>
        ))}
        {left.map((n) => {
          const b = L.get(n.id)!;
          return (
            <g key={n.id}>
              <rect x={x0 - 10} y={b.y} width={10} height={b.h} fill={n.colorVar ?? "var(--c-text-3)"} rx={2} />
              <text x={x0 - 16} y={b.y + b.h / 2 + 4} textAnchor="end" style={{ fill: "var(--c-text)" }}>
                {n.label}
              </text>
            </g>
          );
        })}
        {right.map((n) => {
          const b = R.get(n.id)!;
          return (
            <g key={n.id}>
              <rect x={x1} y={b.y} width={10} height={b.h} fill={n.colorVar ?? "var(--c-risk)"} rx={2} />
              <text x={x1 + 16} y={b.y + b.h / 2 + 4} style={{ fill: "var(--c-text)" }}>
                {n.label}
              </text>
            </g>
          );
        })}
      </svg>
      </div>
      <figcaption>
        {description}
        <details>
          <summary>Flows table</summary>
          <table>
            <caption>{title}</caption>
            <thead>
              <tr>
                <th>From</th>
                <th>To</th>
                <th className="num">{unit}</th>
              </tr>
            </thead>
            <tbody>
              {flows.map((f) => (
                <tr key={`${f.from}-${f.to}-row`}>
                  <td>{name(left, f.from)}</td>
                  <td>{name(right, f.to)}</td>
                  <td className="num">{f.value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
