// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { Waterfall } from "@/components/charts/Waterfall";
import { IndexGauge } from "@/components/meter/IndexGauge";
import { SignalTable } from "@/components/meter/SignalTable";
import { getRelease, getSnapshot, indexById } from "@/lib/data";
import { fmtDate, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Agents",
  description: "Agentic infrastructure: autonomy, access and supply-chain signals, the Agentic Infrastructure Risk Index, and the incidents that involve autonomous systems.",
};

export default async function AgentsPage() {
  const [rel, snap] = await Promise.all([getRelease(), getSnapshot()]);
  const air = indexById(rel, "agentic_infrastructure_risk");
  const drivers = snap.drivers.filter((d) => ["D2", "D6"].includes(d.id));
  const agentIncidents = snap.incidents.filter((i) => i.cause.some((c) => /autonom|agent|injection|tool/i.test(c)) || i.systems_involved.some((s) => /agent/i.test(s)));
  const agentScenarios = snap.scenarios.filter((s) => /agent|autonom|infrastructure|multi/i.test(`${s.name} ${s.description}`));
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="Agentic infrastructure"
          lede="Systems that act: how long they run unattended, what they can reach, how their tools and connectors are secured, and what has already gone wrong at category level."
          textAnchor="agents"
        />
        <div className="grid grid-2">
          <div>{air ? <IndexGauge index={air} /> : null}</div>
          <div className="card">
            <p className="muted" style={{ margin: 0 }}>
              {air?.note}
            </p>
          </div>
        </div>
        {air ? (
          <Waterfall
            title="Agentic Infrastructure Risk Index contributions"
            description="Weighted, tier-adjusted contributions of each autonomy and supply-chain signal in index points."
            items={air.components.map((c) => ({ label: c.signal_id, value: c.contribution, direction: c.contribution >= 0 ? "raises" : "lowers", note: `weight ${c.weight}, normalised ${c.value_normalized.toFixed(2)}, tier ${c.tier}` }))}
          />
        ) : null}
        <section className="stack">
          <h2>Signals (D2 autonomy and access, D6 security and supply chain)</h2>
          <SignalTable drivers={drivers} observations={snap.driver_observations} weights={snap.model_spec.index_weights.agentic_infrastructure_risk} />
        </section>
        <section className="stack">
          <h2>Incidents involving autonomous or tool-using systems</h2>
          {agentIncidents.length ? (
            <div className="grid grid-2">
              {agentIncidents.map((i) => (
                <article key={i.id} className="card">
                  <div className="row">
                    <span className={`badge ${i.near_miss ? "badge-uncertainty" : "badge-risk"}`}>{i.near_miss ? "near miss" : titleCase(i.severity)}</span>
                    <span className="badge">{fmtDate(i.date)}</span>
                    <span className="badge">{titleCase(i.evidence_level)}</span>
                  </div>
                  <h3>
                    <Link href={`/incidents#${i.id}`}>{i.title}</Link>
                  </h3>
                  <p className="muted">{i.summary}</p>
                </article>
              ))}
            </div>
          ) : (
            <p className="muted">No incident in the snapshot involves an autonomous system.</p>
          )}
        </section>
        <section className="stack">
          <h2>Pathways that run through agents</h2>
          <ul>
            {agentScenarios.map((s) => (
              <li key={s.id}>
                <Link href={`/futures/${s.id}`}>
                  {s.id} · {s.name}
                </Link>{" "}
                <span className="cite">{s.recoverability === "none" ? "irreversible" : `${s.recoverability} recoverability`}</span>
              </li>
            ))}
          </ul>
        </section>
        <p className="cite">
          Index method: <Link href="/method/indexes">docs/method/indexes.md</Link>.
        </p>
      </div>
    </div>
  );
}
