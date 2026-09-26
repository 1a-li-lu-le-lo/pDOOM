// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { Waterfall } from "@/components/charts/Waterfall";
import { IndexGauge } from "@/components/meter/IndexGauge";
import { SignalTable } from "@/components/meter/SignalTable";
import { getRelease, getSnapshot, indexById } from "@/lib/data";
import { titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Safeguards",
  description: "Interventions that reduce catastrophic AI risk, each with its mechanism, evidence strength, cost, time to deploy, failure modes and the pathways it targets.",
};

const STRENGTH_ORDER: Record<string, number> = { strong: 0, moderate: 1, weak: 2, none: 3, unknown: 4 };

export default async function SafeguardsPage() {
  const [rel, snap] = await Promise.all([getRelease(), getSnapshot()]);
  const csi = indexById(rel, "control_strength");
  const categories = [...new Set(snap.interventions.map((i) => i.category))];
  const drivers = snap.drivers.filter((d) => ["D5", "D7"].includes(d.id));
  const scenarioName = (id: string) => snap.scenarios.find((s) => s.id === id)?.name ?? id;
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="Safeguards"
          lede="What is being done, what it would take, and where the evidence is thin. Effect sizes are qualitative on purpose: nobody has measured how much any of these changes the probability of an outcome that has never happened."
          textAnchor="safeguards"
        />
        <div className="grid grid-2">
          <div>{csi ? <IndexGauge index={csi} /> : null}</div>
          <div className="card">
            <p className="muted" style={{ margin: 0 }}>
              {csi?.note}
            </p>
          </div>
        </div>
        {csi ? (
          <Waterfall
            title="Control Strength Index contributions"
            description="Weighted, tier-adjusted contributions of each control, security, governance and evaluation signal in index points."
            items={csi.components.map((c) => ({ label: c.signal_id, value: c.contribution, direction: c.contribution >= 0 ? "raises" : "lowers", note: `weight ${c.weight}, normalised ${c.value_normalized.toFixed(2)}, tier ${c.tier}` }))}
          />
        ) : null}
        {categories.map((cat) => {
          const items = snap.interventions.filter((i) => i.category === cat).sort((a, b) => (STRENGTH_ORDER[a.evidence_strength] ?? 9) - (STRENGTH_ORDER[b.evidence_strength] ?? 9));
          return (
            <section key={cat} aria-labelledby={`cat-${cat}`} className="stack">
              <h2 id={`cat-${cat}`}>{titleCase(cat)}</h2>
              <div className="grid grid-2">
                {items.map((i) => (
                  <article key={i.id} id={i.id} className="card stack safeguard-card">
                    <div className="row">
                      <span className={`badge ${i.evidence_strength === "strong" || i.evidence_strength === "moderate" ? "badge-evidence" : "badge-insufficient"}`}>{i.evidence_strength} evidence</span>
                      <span className="badge">cost: {titleCase(i.cost)}</span>
                      <span className="badge">deploy: {i.time_to_deploy.replace(/_/g, " ")}</span>
                      <span className="badge badge-uncertainty">{i.uncertainty} uncertainty</span>
                    </div>
                    <h3 style={{ margin: 0 }}>
                      {i.id} · {i.name}
                    </h3>
                    <p>{i.mechanism}</p>
                    <dl className="kv">
                      <dt>Evidence</dt>
                      <dd>{i.evidence_summary}</dd>
                      <dt>Effect size</dt>
                      <dd>{titleCase(i.effect_size)} (qualitative)</dd>
                      <dt>Could fail if</dt>
                      <dd>{i.possible_failure}</dd>
                      <dt>Could backfire if</dt>
                      <dd>{i.possible_backfire}</dd>
                      <dt>Owners</dt>
                      <dd>{i.owner_types.map(titleCase).join(", ")}</dd>
                      <dt>Targets</dt>
                      <dd>
                        {i.target_scenario_ids.map((s, k) => (
                          <span key={s}>
                            {k ? ", " : ""}
                            <Link href={`/futures/${s}`} title={scenarioName(s)}>
                              {s}
                            </Link>
                          </span>
                        ))}
                      </dd>
                    </dl>
                    {i.user_actions.length ? (
                      <details>
                        <summary>What different audiences can do</summary>
                        <ul>
                          {i.user_actions.map((a) => (
                            <li key={a.audience + a.action}>
                              <strong>{titleCase(a.audience)}:</strong> {a.action}
                            </li>
                          ))}
                        </ul>
                      </details>
                    ) : null}
                    <p className="cite">
                      Sources:{" "}
                      {i.source_ids.map((s, k) => (
                        <span key={s}>
                          {k ? ", " : ""}
                          <Link href={`/evidence/sources/${s}`}>{s}</Link>
                        </span>
                      ))}
                    </p>
                  </article>
                ))}
              </div>
            </section>
          );
        })}
        <section className="stack">
          <h2>Signals (D5 control and evaluation, D7 governance)</h2>
          <SignalTable drivers={drivers} observations={snap.driver_observations} weights={snap.model_spec.index_weights.control_strength} />
        </section>
        <div className="row no-print">
          <Link className="btn btn-primary" href="/act">
            What you can do
          </Link>
          <Link className="btn" href="/method/indexes">
            How the index is computed
          </Link>
        </div>
      </div>
    </div>
  );
}
