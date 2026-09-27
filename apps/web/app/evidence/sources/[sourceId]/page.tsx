// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { PageHeader } from "@/components/shell/PageHeader";
import { getSnapshot } from "@/lib/data";
import { fmtDate, titleCase } from "@/lib/format";

export async function generateStaticParams() {
  const snap = await getSnapshot();
  return snap.sources.map((s) => ({ sourceId: s.id }));
}

export async function generateMetadata({ params }: { params: Promise<{ sourceId: string }> }): Promise<Metadata> {
  const { sourceId } = await params;
  const snap = await getSnapshot();
  const s = snap.sources.find((x) => x.id === sourceId);
  return { title: s?.title ?? "Source", description: s?.evidence_summary };
}

export default async function SourcePage({ params }: { params: Promise<{ sourceId: string }> }) {
  const { sourceId } = await params;
  const snap = await getSnapshot();
  const s = snap.sources.find((x) => x.id === sourceId);
  if (!s) notFound();
  const claims = snap.claims.filter((c) => c.source_id === s.id);
  const forecasts = snap.forecasts.filter((f) => f.source_id === s.id);
  const incidents = snap.incidents.filter((i) => i.source_ids.includes(s.id));
  const scenarios = snap.scenarios.filter((x) => x.source_ids.includes(s.id));
  const observations = snap.driver_observations.filter((o) => o.source_ids.includes(s.id));
  const usedBy = [
    ...forecasts.map((f) => ({ key: f.id, href: f.group_id ? `/forecasts#${f.group_id}` : "/forecasts", label: `Forecast: ${f.forecaster_or_survey}` })),
    ...incidents.map((i) => ({ key: i.id, href: "/incidents", label: `Incident: ${i.title}` })),
    ...scenarios.map((x) => ({ key: x.id, href: `/futures/${x.id}`, label: `Scenario ${x.id}: ${x.name}` })),
    ...observations.map((o) => ({ key: o.id, href: "/capabilities", label: `Driver observation ${o.signal_id}` })),
  ];
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader crumbs={[{ href: "/evidence", label: "Evidence" }]} title={s.title} lede={s.evidence_summary} textAnchor="sources">
          <div className="row">
            <span className="badge badge-evidence">tier {s.source_tier}</span>
            <span className="badge">{titleCase(s.source_type)}</span>
            <span className={`badge ${s.verification.status.startsWith("verified_fetch") || s.verification.status === "verified_search" ? "badge-evidence" : "badge-insufficient"}`}>{titleCase(s.verification.status)}</span>
            <span className="badge">{titleCase(s.model_use_status)}</span>
            {s.conflicts.length ? <span className="badge badge-uncertainty">{s.conflicts.map(titleCase).join(", ")}</span> : null}
          </div>
        </PageHeader>
        <div className="grid grid-2">
          <div className="card">
            <h2 style={{ marginTop: 0 }}>Record</h2>
            <dl className="kv">
              <dt>Canonical URL</dt>
              <dd>
                <a href={s.canonical_url} rel="noopener noreferrer">
                  {s.canonical_url}
                </a>
              </dd>
              <dt>Publisher</dt>
              <dd>{s.publisher}</dd>
              <dt>Authors</dt>
              <dd>{s.authors.join(", ") || "—"}</dd>
              <dt>Published</dt>
              <dd>
                {fmtDate(s.date_published)}
                {s.date_updated ? ` (updated ${fmtDate(s.date_updated)})` : ""}
              </dd>
              <dt>Retrieved</dt>
              <dd>{fmtDate(s.date_retrieved)}</dd>
              <dt>Jurisdiction</dt>
              <dd>{s.jurisdiction ?? "—"}</dd>
              <dt>Topics</dt>
              <dd>{s.topic.map(titleCase).join(", ")}</dd>
              <dt>Licence</dt>
              <dd>{s.license ?? "not stated"}</dd>
              <dt>Robots</dt>
              <dd>{titleCase(s.robots_status)}</dd>
              <dt>Archive</dt>
              <dd>{s.archive_reference ?? "none recorded"}</dd>
              <dt>Content hash</dt>
              <dd>
                <code>{s.content_hash ?? "—"}</code>
              </dd>
              <dt>Citation</dt>
              <dd>{s.citation}</dd>
            </dl>
          </div>
          <div className="card">
            <h2 style={{ marginTop: 0 }}>Assessment</h2>
            <dl className="kv">
              <dt>Verification</dt>
              <dd>
                {titleCase(s.verification.status)} on {fmtDate(s.verification.checked_at)} · {s.verification.method}
                {s.verification.note ? ` · ${s.verification.note}` : ""}
              </dd>
              <dt>Human review</dt>
              <dd>{titleCase(s.human_review_status)}</dd>
              <dt>Model use</dt>
              <dd>{titleCase(s.model_use_status)}</dd>
              <dt>Methodology</dt>
              <dd>{s.methodology ?? "—"}</dd>
              <dt>Sample</dt>
              <dd>{s.sample ?? "—"}</dd>
              <dt>Limitations</dt>
              <dd>{s.limitations ?? "—"}</dd>
              <dt>Counterevidence</dt>
              <dd>{s.counterevidence ?? "none recorded"}</dd>
              <dt>Retraction status</dt>
              <dd>{titleCase(s.retraction_status)}</dd>
            </dl>
          </div>
        </div>
        <section className="stack">
          <h2>Claims extracted from this source ({claims.length})</h2>
          {claims.length ? (
            <div className="table-wrap" tabIndex={0} role="region" aria-label="Atomic claims">
              <table>
                <caption>Atomic claims</caption>
                <thead>
                  <tr>
                    <th>Claim</th>
                    <th>Type</th>
                    <th>Value</th>
                    <th>Relevance</th>
                    <th>Status</th>
                    <th>Locator</th>
                  </tr>
                </thead>
                <tbody>
                  {claims.map((c) => (
                    <tr key={c.id} id={c.id}>
                      <td>{c.text}</td>
                      <td>{titleCase(c.evidence_type)}</td>
                      <td className="num">{c.quantitative_value !== null ? `${c.quantitative_value}${c.unit === "probability" ? "" : ` ${c.unit ?? ""}`}` : "—"}</td>
                      <td>{titleCase(c.relevance)}</td>
                      <td>{titleCase(c.status)}</td>
                      <td className="cite">{c.direct_quote_pointer ?? "—"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="muted">No atomic claims have been extracted from this source yet.</p>
          )}
        </section>
        {usedBy.length ? (
          <section>
            <h2>Where this source is used</h2>
            <ul>
              {usedBy.map((u) => (
                <li key={u.key}>
                  <Link href={u.href}>{u.label}</Link>
                </li>
              ))}
            </ul>
          </section>
        ) : null}
      </div>
    </div>
  );
}
