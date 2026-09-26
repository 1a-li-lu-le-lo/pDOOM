// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Icon array: N cells, of which round(p·N) are filled. The most honest way to
 * show a small probability is to show how many of N futures it touches; the
 * grid is deterministic (filled cells first, reading order) so two viewers see
 * the same picture. Server-rendered SVG with an accessible summary.
 */
export function IconArray({ p, cells = 1000, label, colorVar = "var(--c-risk)", low, high }: { p: number; cells?: number; label: string; colorVar?: string; low?: number; high?: number }) {
  const n = Math.max(0, Math.min(cells, Math.round(p * cells)));
  const nLow = low === undefined ? null : Math.round(low * cells);
  const nHigh = high === undefined ? null : Math.round(high * cells);
  const cols = cells >= 1000 ? 50 : cells >= 100 ? 20 : 10;
  const rows = Math.ceil(cells / cols);
  const size = 10;
  const gap = 2;
  const W = cols * (size + gap);
  const H = rows * (size + gap);
  const cellsArr = Array.from({ length: cells }, (_, i) => i);
  const summary = `${label}: ${n.toLocaleString("en-US")} of ${cells.toLocaleString("en-US")} cells filled${nLow !== null && nHigh !== null ? `; plausible range ${nLow.toLocaleString("en-US")} to ${nHigh.toLocaleString("en-US")}` : ""}.`;
  return (
    <figure className="icon-array">
      <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label={summary} style={{ maxWidth: 560 }}>
        {cellsArr.map((i) => {
          const x = (i % cols) * (size + gap);
          const y = Math.floor(i / cols) * (size + gap);
          const filled = i < n;
          const inRange = nHigh !== null && i < nHigh && !filled;
          return <rect key={i} x={x} y={y} width={size} height={size} rx={2} fill={filled ? colorVar : inRange ? colorVar : "var(--c-surface-2)"} opacity={filled ? 1 : inRange ? 0.28 : 1} />;
        })}
      </svg>
      <figcaption>
        {summary} Filled cells are the median; the faint cells extend to the upper end of the plausible interval.
      </figcaption>
    </figure>
  );
}
