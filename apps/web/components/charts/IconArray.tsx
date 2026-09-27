// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Icon array: N cells, of which round(p·N) are filled. The most honest way to
 * show a small probability is to show how many of N futures it touches; the
 * grid is deterministic (filled cells first, reading order) so two viewers see
 * the same picture. Server-rendered SVG with an accessible summary. Unfilled
 * cells are a pattern fill and filled cells are drawn as whole rows plus one
 * partial row, so a thousand-cell grid is a handful of elements.
 */
export function IconArray({ p, cells = 1000, label, colorVar = "var(--c-risk)", low, high }: { p: number; cells?: number; label: string; colorVar?: string; low?: number; high?: number }) {
  const n = Math.max(0, Math.min(cells, Math.round(p * cells)));
  const nLow = low === undefined ? null : Math.round(low * cells);
  const nHigh = high === undefined ? null : Math.min(cells, Math.round(high * cells));
  const cols = cells >= 1000 ? 50 : cells >= 100 ? 20 : 10;
  const rows = Math.ceil(cells / cols);
  const size = 10;
  const gap = 2;
  const step = size + gap;
  const W = cols * step;
  const H = rows * step;
  const id = `ia-${cells}-${Math.round(p * 1e6)}`;
  const summary = `${label}: ${n.toLocaleString("en-US")} of ${cells.toLocaleString("en-US")} cells filled${nLow !== null && nHigh !== null ? `; plausible range ${nLow.toLocaleString("en-US")} to ${nHigh.toLocaleString("en-US")}` : ""}.`;
  /** Rectangles covering cells [from, to) in reading order. */
  const span = (from: number, to: number, fill: string, opacity: number) => {
    const out: React.ReactNode[] = [];
    let i = from;
    while (i < to) {
      const row = Math.floor(i / cols);
      const col = i % cols;
      const end = Math.min(to, (row + 1) * cols);
      const count = end - i;
      out.push(<rect key={`${fill}-${i}`} x={col * step} y={row * step} width={count * step - gap} height={size} rx={2} fill={fill} opacity={opacity} mask={`url(#${id}-m)`} />);
      i = end;
    }
    return out;
  };
  return (
    <figure className="icon-array">
      <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={summary} style={{ maxWidth: 560 }}>
        <defs>
          <pattern id={`${id}-p`} width={step} height={step} patternUnits="userSpaceOnUse">
            <rect width={size} height={size} rx={2} fill="var(--c-surface-2)" />
          </pattern>
          <mask id={`${id}-m`}>
            <rect width={W} height={H} fill={`url(#${id}-mp)`} />
          </mask>
          <pattern id={`${id}-mp`} width={step} height={step} patternUnits="userSpaceOnUse">
            <rect width={size} height={size} rx={2} fill="#fff" />
          </pattern>
        </defs>
        <rect width={W} height={H} fill={`url(#${id}-p)`} />
        {nHigh !== null && nHigh > n ? span(n, nHigh, colorVar, 0.28) : null}
        {span(0, n, colorVar, 1)}
      </svg>
      <figcaption>
        {summary} Filled cells are the median; the faint cells extend to the upper end of the plausible interval.
        <details>
          <summary>Data table</summary>
          <table>
            <caption>{label}</caption>
            <thead>
              <tr>
                <th>Cells</th>
                <th className="num">Filled (median)</th>
                <th className="num">Lower end</th>
                <th className="num">Upper end</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="num">{cells.toLocaleString("en-US")}</td>
                <td className="num">{n.toLocaleString("en-US")}</td>
                <td className="num">{nLow === null ? "—" : nLow.toLocaleString("en-US")}</td>
                <td className="num">{nHigh === null ? "—" : nHigh.toLocaleString("en-US")}</td>
              </tr>
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
