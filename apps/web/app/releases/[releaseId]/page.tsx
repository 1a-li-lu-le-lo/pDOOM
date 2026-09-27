// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { PageHeader } from "@/components/shell/PageHeader";
import { EstimateCard } from "@/components/meter/EstimateCard";
import { getDataSource } from "@/lib/data";
import { renderMarkdown } from "@/lib/markdown";
import { fmtDate, titleCase } from "@/lib/format";

export async function generateStaticParams() {
  const releases = await getDataSource().listReleases();
  return releases.map((r) => ({ releaseId: r.release_id }));
}

export async function generateMetadata({ params }: { params: Promise<{ releaseId: string }> }): Promise<Metadata> {
  const { releaseId } = await params;
  return { title: `Release ${releaseId}`, description: `Manifest, approvals, files and model card of release ${releaseId}.` };
}

export default async function ReleasePage({ params }: { params: Promise<{ releaseId: string }> }) {
  const { releaseId } = await params;
  if (!/^rel-\d{4}-\d{2}-\d{2}-\d{3}$/.test(releaseId)) notFound();
  let rel;
  try {
    rel = await getDataSource().getReleaseById(releaseId);
  } catch {
    notFound();
  }
  const m = rel.manifest;
  const card = renderMarkdown(rel.documents.model_card);
  const research = rel.estimates.filter((e) => e.status === "research_mode" && e.outcome_set.length === 6);
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader crumbs={[{ href: "/changelog", label: "Changelog" }]} title={`Release ${m.release_id}`} lede={`Published ${fmtDate(m.published)} from snapshot ${m.data_snapshot}. ${m.superseded ? `Superseded by ${m.superseded.by} on ${fmtDate(m.superseded.at)}.` : "This is the current release."}`} textAnchor="history">
          <div className="row">
            <span className="badge badge-uncertainty">{titleCase(m.editorial_risk_level)}</span>
            <span className="badge">uncertainty {m.uncertainty_score === null ? "—" : Math.round(m.uncertainty_score)}</span>
            {m.approval ? (
              <span className="badge">
                {m.approval.received_approvals} {m.approval.received_approvals === 1 ? "approval" : "approvals"} · {m.approval.required_approvals} required
              </span>
            ) : null}
            {m.approval?.heightened_review ? <span className="badge badge-risk">heightened review</span> : null}
          </div>
        </PageHeader>
        <div className="grid grid-2">
          <section className="card">
            <h2 style={{ marginTop: 0 }}>Manifest</h2>
            <dl className="kv">
              <dt>Candidate</dt>
              <dd>{m.candidate_id}</dd>
              <dt>Generated</dt>
              <dd>{fmtDate(m.generated_at)}</dd>
              <dt>Source cutoff</dt>
              <dd>{fmtDate(m.source_cutoff)}</dd>
              <dt>Code commit</dt>
              <dd>
                <code>{m.code_commit}</code>
              </dd>
              <dt>Model versions</dt>
              <dd>{m.model_versions.join(", ")}</dd>
              <dt>Previous release</dt>
              <dd>{m.previous_release_id ? <Link href={`/releases/${m.previous_release_id}`}>{m.previous_release_id}</Link> : "none"}</dd>
              <dt>Reviewers</dt>
              <dd>{m.reviewers.join(", ")}</dd>
              <dt>Reproduce</dt>
              <dd>
                <code>{m.reproduction_command}</code>
              </dd>
              <dt>Signature</dt>
              <dd>
                <code style={{ wordBreak: "break-all" }}>{m.signature}</code>
              </dd>
            </dl>
          </section>
          <section className="card">
            <h2 style={{ marginTop: 0 }}>Approvals</h2>
            {rel.approvals.length ? (
              <ul>
                {rel.approvals.map((a) => (
                  <li key={a.reviewer_id + a.signed_at}>
                    <strong>{a.reviewer_id}</strong> signed {fmtDate(a.signed_at)} · key <code>{a.key_id}</code>
                    <span className="cite"> · conflicts declared: {a.conflicts_declared || "none"}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="muted">No approval records shipped with this release.</p>
            )}
            {m.approval?.heightened_review_ack ? <p className="cite">Heightened review acknowledged: {m.approval.heightened_review_ack}</p> : null}
            <h2>Files</h2>
            <div className="table-wrap" tabIndex={0} role="region" aria-label="Release files">
              <table>
                <caption>Release files with SHA-256 digests</caption>
                <thead>
                  <tr>
                    <th>File</th>
                    <th className="num">Items</th>
                    <th>SHA-256</th>
                  </tr>
                </thead>
                <tbody>
                  {m.files.map((f) => (
                    <tr key={f.path}>
                      <td>{f.path}</td>
                      <td className="num">{f.count}</td>
                      <td>
                        <code style={{ fontSize: "var(--fs-xs)" }}>{f.sha256.slice(0, 16)}…</code>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        </div>
        <section className="stack">
          <h2>What changed</h2>
          <ul>
            {m.changes.map((c) => (
              <li key={c}>{c}</li>
            ))}
          </ul>
          <h2>Known limitations</h2>
          <ul>
            {m.known_limitations.map((c) => (
              <li key={c}>{c}</li>
            ))}
          </ul>
        </section>
        <section className="stack">
          <h2>Research-mode p(DOOM) estimates in this release</h2>
          <div className="grid grid-3">
            {research.map((e) => (
              <EstimateCard key={e.estimate_id} e={e} compact />
            ))}
          </div>
        </section>
        <article className="prose card" aria-label="Model card" dangerouslySetInnerHTML={{ __html: card.html.replace(/^<h1[^>]*>[\s\S]*?<\/h1>\n?/, "") }} />
        <div className="row no-print">
          <Link className="btn" href={`/api/export/release.json?release=${m.release_id}`}>
            Download this release (JSON)
          </Link>
        </div>
      </div>
    </div>
  );
}
