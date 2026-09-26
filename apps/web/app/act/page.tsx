// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { getSnapshot } from "@/lib/data";
import { fmtDate, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Act",
  description: "Concrete, proportionate actions for individuals, developers, researchers, policymakers, journalists and organisations, and the organisations working on the problem.",
};

const AUDIENCE_ORDER = ["individuals", "software_engineers", "ai_researchers", "laboratories", "policymakers", "funders", "educators", "nonprofits", "auditors_red_teams", "standards_bodies"];

export default async function ActPage({ searchParams }: { searchParams: Promise<{ audience?: string }> }) {
  const sp = await searchParams;
  const snap = await getSnapshot();
  const audiences = [...new Set(snap.actions.map((a) => a.audience))].sort((a, b) => (AUDIENCE_ORDER.indexOf(a) + 1 || 99) - (AUDIENCE_ORDER.indexOf(b) + 1 || 99));
  const selected = sp.audience && (audiences as string[]).includes(sp.audience) ? sp.audience : undefined;
  const orgs = [...snap.organizations].sort((a, b) => a.name.localeCompare(b.name));
  const interventionName = (id: string) => snap.interventions.find((i) => i.id === id)?.name ?? id;
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="What you can do"
          lede="Fear is not a plan. These are actions with a stated effort level, linked to the safeguard they support and to a source. None of them requires believing any particular number."
          textAnchor="act"
        >
          <nav className="horizon-tabs" aria-label="Audience">
            <Link href="/act" aria-current={!selected ? "page" : undefined}>
              Everyone
            </Link>
            {audiences.map((a) => (
              <Link key={a} href={`/act?audience=${a}`} aria-current={selected === a ? "page" : undefined}>
                {titleCase(a)}
              </Link>
            ))}
          </nav>
        </PageHeader>
        {audiences
          .filter((a) => !selected || a === selected)
          .map((a) => (
            <section key={a} id={`audience-${a}`} className="stack" aria-labelledby={`aud-${a}`}>
              <h2 id={`aud-${a}`}>{titleCase(a)}</h2>
              <div className="grid grid-3">
                {snap.actions
                  .filter((x) => x.audience === a)
                  .map((x) => (
                    <article key={x.id} className="card stack">
                      <div className="row">
                        <span className={`badge ${x.effort === "low" ? "badge-resilience" : x.effort === "high" ? "badge-uncertainty" : ""}`}>{x.effort} effort</span>
                      </div>
                      <h3 style={{ margin: 0 }}>{x.title}</h3>
                      <p className="muted">{x.description}</p>
                      {x.related_intervention_ids.length ? (
                        <p className="cite">
                          Supports:{" "}
                          {x.related_intervention_ids.map((i, k) => (
                            <span key={i}>
                              {k ? ", " : ""}
                              <Link href={`/safeguards#${i}`}>{interventionName(i)}</Link>
                            </span>
                          ))}
                        </p>
                      ) : null}
                      {x.resources.length ? (
                        <ul className="cite">
                          {x.resources.map((r) => (
                            <li key={r.url}>
                              <a href={r.url} rel="noopener noreferrer">
                                {r.title}
                              </a>
                            </li>
                          ))}
                        </ul>
                      ) : null}
                    </article>
                  ))}
              </div>
            </section>
          ))}
        <section id="organizations" className="stack" aria-labelledby="orgs-h">
          <h2 id="orgs-h">Organisations</h2>
          <p className="muted">
            Listed because they meet stated inclusion criteria, not as an endorsement. Funding and conflicts are recorded as the organisation states them; last verified dates are shown.
          </p>
          <div className="grid grid-2">
            {orgs.map((o) => (
              <article key={o.id} id={o.id} className="card stack">
                <div className="row">
                  {o.focus.slice(0, 4).map((f) => (
                    <span key={f} className="badge">
                      {titleCase(f)}
                    </span>
                  ))}
                </div>
                <h3 style={{ margin: 0 }}>
                  <a href={o.url} rel="noopener noreferrer">
                    {o.name}
                  </a>
                </h3>
                <p>{o.mission}</p>
                <dl className="kv">
                  <dt>Status</dt>
                  <dd>
                    {o.legal_status} · {o.jurisdiction}
                  </dd>
                  <dt>Programs</dt>
                  <dd>{o.programs.join("; ")}</dd>
                  <dt>Open outputs</dt>
                  <dd>{o.open_outputs.join("; ")}</dd>
                  <dt>Funding</dt>
                  <dd>{o.funding_disclosure}</dd>
                  <dt>Conflicts</dt>
                  <dd>{o.conflicts.length ? o.conflicts.join("; ") : "none recorded"}</dd>
                  <dt>Evidence of impact</dt>
                  <dd>{o.evidence_of_impact}</dd>
                  <dt>Ways to help</dt>
                  <dd>{o.ways_to_help.join("; ")}</dd>
                  <dt>Inclusion criteria met</dt>
                  <dd>{o.inclusion_criteria_met.join("; ")}</dd>
                  <dt>Last verified</dt>
                  <dd>{fmtDate(o.last_verified)}</dd>
                </dl>
              </article>
            ))}
          </div>
        </section>
        <p className="cite">
          Inclusion rules and conflict-of-interest policy: <Link href="/method">methodology</Link>. Submit a correction or a source through the API's review queue; nothing submitted changes a published number without review.
        </p>
      </div>
    </div>
  );
}
