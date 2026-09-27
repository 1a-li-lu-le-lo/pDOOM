// Copyright NU Cybernetics. p(DOOM) — research prototype.
export interface BarListItem {
  label: string;
  value: number;
  display?: string;
  note?: string;
  colorVar?: string;
  href?: string;
}

/** Horizontal bars with labels; values on a shared 0–max scale. Accessible table included. */
export function BarList({ items, max, title, description, unit = "" }: { items: BarListItem[]; max?: number; title: string; description: string; unit?: string }) {
  const top = max ?? Math.max(1e-9, ...items.map((i) => i.value));
  return (
    <figure>
      <ul className="driver-list" aria-label={title}>
        {items.map((it) => (
          <li key={it.label}>
            <div>
              <div>{it.href ? <a href={it.href}>{it.label}</a> : it.label}</div>
              <div className="gauge" style={{ padding: 0, border: 0, background: "transparent" }}>
                <div className="bar" style={{ color: it.colorVar ?? "var(--c-evidence)" }} aria-hidden="true">
                  <span style={{ width: `${Math.min(100, Math.max(0, (it.value / top) * 100))}%` }} />
                </div>
              </div>
              {it.note ? <div className="muted" style={{ fontSize: "var(--fs-xs)" }}>{it.note}</div> : null}
            </div>
            <div className="num">{it.display ?? `${it.value}${unit}`}</div>
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
                <th>Item</th>
                <th className="num">Value</th>
                <th>Note</th>
              </tr>
            </thead>
            <tbody>
              {items.map((it) => (
                <tr key={`${it.label}-row`}>
                  <td>{it.label}</td>
                  <td className="num">{it.display ?? `${it.value}${unit}`}</td>
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
