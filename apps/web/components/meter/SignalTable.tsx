// Copyright NU Cybernetics. p(DOOM) — research prototype.
import Link from "next/link";
import type { Driver, DriverObservation } from "@pdoom/schemas";
import { fmtDate, titleCase } from "@/lib/format";

/** The observations behind one or more driver families, with their normalisation stated. */
export function SignalTable({ drivers, observations, weights }: { drivers: Driver[]; observations: DriverObservation[]; weights?: Record<string, number> }) {
  return (
    <div className="table-wrap table-wide" tabIndex={0} role="region" aria-label="Signals and their latest observations">
      <table>
        <caption>Signals, their latest observation and how each raw value is normalised to 0–1</caption>
        <thead>
          <tr>
            <th>Signal</th>
            <th>Latest observation</th>
            <th className="num">Normalised</th>
            <th className="num">Confidence</th>
            <th>Kind</th>
            <th className="num">Weight</th>
            <th>Normalisation</th>
          </tr>
        </thead>
        <tbody>
          {drivers.flatMap((d) =>
            d.signals.map((sig) => {
              const obs = observations.filter((o) => o.signal_id === sig.signal_id).sort((a, b) => b.as_of.localeCompare(a.as_of))[0];
              return (
                <tr key={sig.signal_id} id={sig.signal_id}>
                  <td>
                    <strong>{sig.name}</strong>
                    <div className="cite">
                      {sig.signal_id} · {sig.direction.replace(/_/g, " ")}
                    </div>
                  </td>
                  <td>
                    {obs ? (
                      <>
                        {obs.raw_value !== null ? `${obs.raw_value} ${obs.raw_unit ?? ""}` : "judgment"} <span className="cite">as of {fmtDate(obs.as_of)}</span>
                        <details>
                          <summary className="cite">rationale</summary>
                          <p className="cite">{obs.rationale}</p>
                          {obs.counterevidence ? <p className="cite">Counterevidence: {obs.counterevidence}</p> : null}
                          <p className="cite">
                            Sources:{" "}
                            {obs.source_ids.map((s, i) => (
                              <span key={s}>
                                {i ? ", " : ""}
                                <Link href={`/evidence/sources/${s}`}>{s}</Link>
                              </span>
                            ))}
                          </p>
                        </details>
                      </>
                    ) : (
                      <span className="muted">no observation</span>
                    )}
                  </td>
                  <td className="num">{obs ? obs.value_normalized.toFixed(2) : "—"}</td>
                  <td className="num">{obs ? obs.confidence.toFixed(1) : "—"}</td>
                  <td>{obs ? titleCase(obs.observation_kind) : "—"}</td>
                  <td className="num">{weights?.[sig.signal_id] !== undefined ? weights[sig.signal_id]!.toFixed(2) : "—"}</td>
                  <td className="cite">{sig.normalization}</td>
                </tr>
              );
            }),
          )}
        </tbody>
      </table>
    </div>
  );
}
