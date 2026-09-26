// Copyright NU Cybernetics. p(DOOM) — research prototype.
export interface GraphNode {
  id: string;
  label: string;
  group: string;
  href?: string;
  colorVar?: string;
}
export interface GraphEdge {
  from: string;
  to: string;
  label: string;
  strength?: number;
}

/**
 * Deterministic layered node graph (server-rendered SVG). Nodes are placed on
 * a circle ordered by group so the same data always draws the same picture;
 * edges are quadratic arcs with arrowheads. The accessible table lists every
 * edge, which is the authoritative form of the graph.
 */
export function NodeGraph({ nodes, edges, title, description }: { nodes: GraphNode[]; edges: GraphEdge[]; title: string; description: string }) {
  const W = 760;
  const H = 560;
  const cx = W / 2;
  const cy = H / 2;
  const R = Math.min(W, H) / 2 - 70;
  const ordered = [...nodes].sort((a, b) => a.group.localeCompare(b.group) || a.id.localeCompare(b.id));
  const pos = new Map(ordered.map((n, i) => {
    const t = (i / ordered.length) * Math.PI * 2 - Math.PI / 2;
    return [n.id, { x: cx + Math.cos(t) * R, y: cy + Math.sin(t) * R, t }];
  }));
  const byId = new Map(nodes.map((n) => [n.id, n]));
  return (
    <figure>
      <svg className="chart node-graph" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${title}: ${nodes.length} nodes and ${edges.length} relations; see the table below`}>
        <defs>
          <marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M0,0 L10,5 L0,10 z" fill="var(--c-text-3)" />
          </marker>
        </defs>
        {edges.map((e) => {
          const a = pos.get(e.from);
          const b = pos.get(e.to);
          if (!a || !b) return null;
          const mx = (a.x + b.x) / 2;
          const my = (a.y + b.y) / 2;
          const k = 0.35;
          const qx = cx + (mx - cx) * k;
          const qy = cy + (my - cy) * k;
          return <path key={`${e.from}-${e.to}`} d={`M${a.x.toFixed(1)},${a.y.toFixed(1)} Q${qx.toFixed(1)},${qy.toFixed(1)} ${b.x.toFixed(1)},${b.y.toFixed(1)}`} fill="none" stroke="var(--c-text-3)" strokeOpacity={0.35 + 0.4 * (e.strength ?? 0.5)} strokeWidth={1 + (e.strength ?? 0.5)} markerEnd="url(#arrow)" />;
        })}
        {ordered.map((n) => {
          const p = pos.get(n.id)!;
          const right = Math.cos(p.t) >= 0;
          const lx = p.x + Math.cos(p.t) * 16;
          const ly = p.y + Math.sin(p.t) * 16;
          const body = (
            <>
              <circle cx={p.x} cy={p.y} r={9} fill={n.colorVar ?? "var(--c-evidence)"} stroke="var(--c-bg)" strokeWidth={2} />
              <text x={lx} y={ly + 4} textAnchor={right ? "start" : "end"} style={{ fill: "var(--c-text)" }}>
                {n.id} {n.label.length > 26 ? `${n.label.slice(0, 24)}…` : n.label}
              </text>
            </>
          );
          return n.href ? (
            <a key={n.id} href={n.href} aria-label={`${n.id}: ${n.label}`}>
              {body}
            </a>
          ) : (
            <g key={n.id}>{body}</g>
          );
        })}
      </svg>
      <figcaption>
        {description}
        <details>
          <summary>Relations table</summary>
          <table>
            <caption>{title}: every directed relation</caption>
            <thead>
              <tr>
                <th>From</th>
                <th>Relation</th>
                <th>To</th>
              </tr>
            </thead>
            <tbody>
              {edges.map((e) => (
                <tr key={`${e.from}-${e.to}-${e.label}`}>
                  <td>
                    {e.from} {byId.get(e.from)?.label}
                  </td>
                  <td>{e.label}</td>
                  <td>
                    {e.to} {byId.get(e.to)?.label}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
