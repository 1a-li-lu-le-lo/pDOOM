// Copyright NU Cybernetics. p(DOOM) — research prototype.
export interface Contribution {
  label: string;
  value: number;
  direction: string;
  note?: string;
}

/**
 * Contribution bars for the "Why this number?" panel. Index points, never probabilities.
 * Bars use `colorVar` (default: the risk colour); a bar whose `direction` is `"strengthens_control"`
 * always uses the safeguard colour. Pass `colorVar="var(--c-safeguard)"` for a control-strength index.
 */
export function Waterfall({
  items,
  title,
  description,
  colorVar = "var(--c-risk)",
}: {
  items: Contribution[];
  title: string;
  description: string;
  colorVar?: string;
}) {
  const max = Math.max(1e-9, ...items.map((i) => Math.abs(i.value)));
  return (
    <figure>
      <ul className="driver-list" aria-label={title}>
        {items.map((it) => (
          <li key={it.label}>
            <div>
              <div>{it.label}</div>
              <div className="bar" style={{ height: 6, borderRadius: 3, background: "var(--c-surface-2)", overflow: "hidden" }} aria-hidden="true">
                <span
                  style={{
                    display: "block",
                    height: "100%",
                    width: `${(Math.abs(it.value) / max) * 100}%`,
                    background: it.direction === "strengthens_control" ? "var(--c-safeguard)" : colorVar,
                  }}
                />
              </div>
              {it.note ? <div className="muted" style={{ fontSize: "var(--fs-xs)" }}>{it.note}</div> : null}
            </div>
            <div className="num">
              {it.value >= 0 ? "+" : ""}
              {Math.round(it.value * 10) / 10} pts
            </div>
          </li>
        ))}
      </ul>
      <figcaption>
        {description}
        <details>
          <summary>Data table</summary>
          <table>
            <caption>{title}</caption>
            <thead>
              <tr>
                <th>Component</th>
                <th className="num">Index points</th>
                <th>Note</th>
              </tr>
            </thead>
            <tbody>
              {items.map((it) => (
                <tr key={`${it.label}-row`}>
                  <td>{it.label}</td>
                  <td className="num">
                    {it.value >= 0 ? "+" : ""}
                    {Math.round(it.value * 10) / 10}
                  </td>
                  <td className="cite">{it.note ?? ""}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      </figcaption>
    </figure>
  );
}
