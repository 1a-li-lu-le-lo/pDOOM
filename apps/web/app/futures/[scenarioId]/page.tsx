// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { PageHeader } from "@/components/shell/PageHeader";
import { getSnapshot } from "@/lib/data";
import { outcomeLabel, outcomeSetLabel, titleCase } from "@/lib/format";

export async function generateStaticParams() {
  const snap = await getSnapshot();
  return snap.scenarios.map((s) => ({ scenarioId: s.id }));
}

export async function generateMetadata({ params }: { params: Promise<{ scenarioId: string }> }): Promise<Metadata> {
  const { scenarioId } = await params;
  const snap = await getSnapshot();
  const s = snap.scenarios.find((x) => x.id === scenarioId);
  return { title: s ? `${s.id} ${s.name}` : "Scenario", description: s?.description };
}

function List({ title, items }: { title: string; items: readonly string[] }) {
  if (!items.length) return null;
  return (
    <section>
      <h2>{title}</h2>
      <ul>
        {items.map((i) => (
          <li key={i}>{i}</li>
        ))}
      </ul>
    </section>
  );
}

export default async function ScenarioPage({ params }: { params: Promise<{ scenarioId: string }> }) {
  const { scenarioId } = await params;
  const snap = await getSnapshot();
  const s = snap.scenarios.find((x) => x.id === scenarioId);
  if (!s) notFound();
  const out = snap.scenario_edges.filter((e) => e.from_id === s.id);
  const inn = snap.scenario_edges.filter((e) => e.to_id === s.id);
  const name = (id: string) => snap.scenarios.find((x) => x.id === id)?.name ?? id;
  const interventions = snap.interventions.filter((i) => s.intervention_ids.includes(i.id) || i.target_scenario_ids.includes(s.id));
  const sources = snap.sources.filter((src) => s.source_ids.includes(src.id));
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          crumbs={[{ href: "/futures", label: "Futures" }]}
          title={`${s.id} · ${s.name}`}
          lede={s.description}
          textAnchor="scenarios"
        >
          <div className="row">
            <span className="badge badge-uncertainty">{s.uncertainty} uncertainty</span>
            <span className="badge" style={{ color: s.recoverability === "none" ? "var(--c-risk)" : s.recoverability === "high" || s.recoverability === "moderate" ? "var(--c-resilience)" : s.recoverability === "unknown" ? "var(--c-insufficient)" : "var(--c-uncertainty)" }}>
              {s.recoverability === "none" ? "Irreversible" : `${titleCase(s.recoverability)} recoverability`}
            </span>
            <span className="badge">{s.probability_source.replace(/_/g, " ")}</span>
            <span className="badge">{titleCase(s.human_review_status)}</span>
          </div>
        </PageHeader>

        <div className="grid grid-2">
          <div className="card stack">
            <h2 style={{ marginTop: 0 }}>Outcomes this pathway can reach</h2>
            <ul>
              {s.outcome_set.map((o) => (
                <li key={o}>
                  <strong>{o}</strong> {outcomeLabel(o)}
                </li>
              ))}
            </ul>
            <p className="muted">{outcomeSetLabel(s.outcome_set)}</p>
            <h2>Exposure</h2>
            <p>{s.exposure}</p>
            <h2>Time horizon</h2>
            <p>{s.time_horizon_note}</p>
            {s.content_safety_note ? (
              <p className="cite">Content note: {s.content_safety_note}</p>
            ) : null}
          </div>
          <div className="card stack">
            <h2 style={{ marginTop: 0 }}>Evidence</h2>
            <p>{s.evidence_summary}</p>
            {sources.length ? (
              <ul>
                {sources.map((src) => (
                  <li key={src.id}>
                    <Link href={`/evidence/sources/${src.id}`}>{src.title}</Link> <span className="cite">tier {src.source_tier}</span>
                  </li>
                ))}
              </ul>
            ) : null}
          </div>
        </div>

        <div className="grid grid-2">
          <List title="Prerequisites" items={s.prerequisites} />
          <List title="Capability thresholds" items={s.capability_thresholds} />
          <List title="Early indicators (what would be visible first)" items={s.early_indicators} />
          <List title="Counterindicators (what would make this less likely)" items={s.counterindicators} />
          <List title="Control failures involved" items={s.control_failures} />
          <List title="Human contributions" items={s.human_contributions} />
          <List title="AI contributions" items={s.ai_contributions} />
          <List title="Open questions" items={s.open_questions} />
        </div>

        <section className="card stack">
          <h2 style={{ marginTop: 0 }}>Relations</h2>
          {inn.length ? (
            <>
              <h3>Enabled or amplified by</h3>
              <ul>
                {inn.map((e) => (
                  <li key={e.id}>
                    <Link href={`/futures/${e.from_id}`}>
                      {e.from_id} {name(e.from_id)}
                    </Link>{" "}
                    <em>{titleCase(e.relation)}</em> this pathway ({e.confidence} confidence): {e.rationale}
                  </li>
                ))}
              </ul>
            </>
          ) : null}
          {out.length ? (
            <>
              <h3>Leads to</h3>
              <ul>
                {out.map((e) => (
                  <li key={e.id}>
                    <em>{titleCase(e.relation)}</em>{" "}
                    <Link href={`/futures/${e.to_id}`}>
                      {e.to_id} {name(e.to_id)}
                    </Link>{" "}
                    ({e.confidence} confidence): {e.rationale}
                  </li>
                ))}
              </ul>
            </>
          ) : null}
          {s.dependencies.length ? (
            <p className="muted">
              Depends on:{" "}
              {s.dependencies.map((d, i) => (
                <span key={d}>
                  {i ? ", " : ""}
                  <Link href={`/futures/${d}`}>{d}</Link>
                </span>
              ))}
            </p>
          ) : null}
        </section>

        <section className="stack">
          <h2>Safeguards that target this pathway</h2>
          {interventions.length ? (
            <div className="grid grid-3">
              {interventions.map((i) => (
                <article key={i.id} className="card">
                  <div className="row">
                    <span className="badge badge-safeguard">{i.category.replace(/_/g, " ")}</span>
                    <span className="badge">{i.evidence_strength} evidence</span>
                  </div>
                  <h3>
                    <Link href={`/safeguards#${i.id}`}>
                      {i.id} · {i.name}
                    </Link>
                  </h3>
                  <p className="muted">{i.mechanism}</p>
                </article>
              ))}
            </div>
          ) : (
            <p className="muted">No safeguard in the snapshot targets this pathway yet. That is itself a finding.</p>
          )}
        </section>
      </div>
    </div>
  );
}
