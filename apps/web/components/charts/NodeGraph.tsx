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
 *
 * Two layouts are rendered and a media query shows one of them: the circle
 * for wide viewports, and for narrow viewports (phones) a vertical arc
 * diagram whose viewBox is sized to the phone column, so labels render at
 * their nominal size and each node has a full-row hit area, instead of the
 * circle being scaled down to unreadable text and tiny link targets.
 */
const NARROW_BREAKPOINT = 720;
const NARROW_MAX_WIDTH = 480;
const LAYOUT_CSS = `.chart.node-graph--narrow{display:none;max-width:${NARROW_MAX_WIDTH}px}@media (max-width:${NARROW_BREAKPOINT}px){.chart.node-graph--wide{display:none}.chart.node-graph--narrow{display:block}}`;

const nodeText = (n: GraphNode) => `${n.id} ${n.label.length > 26 ? `${n.label.slice(0, 24)}…` : n.label}`;
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
  // Narrow layout: one node per row down a single column, arcs on the left.
  const NW = 360;
  const rowH = 36;
  const padY = 18;
  const NH = padY * 2 + ordered.length * rowH;
  const nx = 96;
  const nr = 7;
  const rowY = new Map(ordered.map((n, i) => [n.id, padY + rowH * i + rowH / 2]));
  const ariaLabel = `${title}: ${nodes.length} nodes and ${edges.length} relations; each node links to its page and the table below lists every relation`;
  return (
    <figure>
      <svg className="chart node-graph node-graph--wide" viewBox={`0 0 ${W} ${H}`} role="group" aria-label={ariaLabel}>
        <style>{LAYOUT_CSS}</style>
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
                {nodeText(n)}
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
      <svg className="chart node-graph node-graph--narrow" viewBox={`0 0 ${NW} ${NH}`} role="group" aria-label={ariaLabel}>
        <defs>
          <marker id="arrow-narrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M0,0 L10,5 L0,10 z" fill="var(--c-text-3)" />
          </marker>
        </defs>
        {edges.map((e) => {
          const y1 = rowY.get(e.from);
          const y2 = rowY.get(e.to);
          if (y1 === undefined || y2 === undefined || y1 === y2) return null;
          const x = nx - nr - 1;
          const bulge = Math.min(150, 24 + Math.abs(y2 - y1) * 0.45);
          return <path key={`${e.from}-${e.to}`} d={`M${x},${y1.toFixed(1)} Q${(x - bulge).toFixed(1)},${((y1 + y2) / 2).toFixed(1)} ${x},${y2.toFixed(1)}`} fill="none" stroke="var(--c-text-3)" strokeOpacity={0.35 + 0.4 * (e.strength ?? 0.5)} strokeWidth={1 + (e.strength ?? 0.5)} markerEnd="url(#arrow-narrow)" />;
        })}
        {ordered.map((n) => {
          const y = rowY.get(n.id)!;
          const body = (
            <>
              <rect x={nx - rowH / 2} y={y - rowH / 2} width={NW - nx + rowH / 2} height={rowH} fill="transparent" />
              <circle cx={nx} cy={y} r={nr} fill={n.colorVar ?? "var(--c-evidence)"} stroke="var(--c-bg)" strokeWidth={2} />
              <text x={nx + 14} y={y + 4} textAnchor="start" style={{ fill: "var(--c-text)", fontSize: 12 }}>
                {nodeText(n)}
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
