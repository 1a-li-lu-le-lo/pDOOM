// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { getDataSource, getRelease } from "@/lib/data";
import { renderMarkdown } from "@/lib/markdown";
import { fmtDate, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Changelog",
  description: "Every release of the observatory, what changed in it, who approved it, and the diff of every estimate and index against the release before.",
};

export default async function ChangelogPage() {
  const [rel, releases] = await Promise.all([getRelease(), getDataSource().listReleases()]);
  const log = renderMarkdown(rel.documents.changelog);
  const changes = rel.delta.estimate_changes.filter((c) => c.change_points !== null && c.change_points !== 0);
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader title="Changelog" lede="A number that moves without an explanation is a rumour. Every release lists what changed, why, and who signed it." textAnchor="history" />
        <div className="table-wrap">
          <table>
            <caption>Releases, newest first</caption>
            <thead>
              <tr>
                <th>Release</th>
                <th>Published</th>
                <th>Snapshot</th>
                <th>Models</th>
                <th>Editorial level</th>
                <th className="num">Uncertainty</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {releases.map((r) => (
                <tr key={r.release_id}>
                  <td>
                    <Link href={`/releases/${r.release_id}`}>{r.release_id}</Link>
                  </td>
                  <td>{fmtDate(r.published)}</td>
                  <td>{r.data_snapshot}</td>
                  <td>{r.model_versions.join(", ")}</td>
                  <td>{titleCase(r.editorial_risk_level)}</td>
                  <td className="num">{r.uncertainty_score === null ? "—" : Math.round(r.uncertainty_score)}</td>
                  <td>{r.is_current ? <span className="badge badge-evidence">current</span> : r.superseded ? `superseded by ${r.superseded.by}` : ""}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <section className="card stack">
          <h2 style={{ marginTop: 0 }}>Current release: {rel.manifest.release_id}</h2>
          <ul>
            {rel.manifest.changes.map((c) => (
              <li key={c}>{c}</li>
            ))}
          </ul>
          <h3>Estimate changes against {rel.manifest.previous_release_id ?? "no previous release"}</h3>
          {changes.length ? (
            <div className="table-wrap">
              <table>
                <caption>Estimates whose median moved</caption>
                <thead>
                  <tr>
                    <th>Estimate</th>
                    <th>Previous</th>
                    <th>New</th>
                    <th className="num">Change (points)</th>
                  </tr>
                </thead>
                <tbody>
                  {changes.map((c) => (
                    <tr key={c.estimate_id}>
                      <td>{c.estimate_id}</td>
                      <td>{c.previous_display || "—"}</td>
                      <td>{c.new_display}</td>
                      <td className="num">{c.change_points?.toFixed(1)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="muted">No estimate median moved relative to the previous release{rel.delta.previous_release_id ? "" : " (this is the first release, so every value is new)"}.</p>
          )}
          {rel.delta.heightened_review_triggers.length ? (
            <p className="cite">Heightened review triggers: {rel.delta.heightened_review_triggers.join("; ")}</p>
          ) : null}
          {rel.delta.index_changes.length ? (
            <>
              <h3>Index changes</h3>
              <ul>
                {rel.delta.index_changes.map((c) => (
                  <li key={c.index_id}>
                    {titleCase(c.index_id)}: {c.previous === null ? "—" : Math.round(c.previous)} → {c.new === null ? "—" : Math.round(c.new)}
                    {c.delta !== null ? ` (${c.delta > 0 ? "+" : ""}${c.delta.toFixed(1)})` : ""}
                  </li>
                ))}
              </ul>
            </>
          ) : null}
        </section>
        <article className="prose card" aria-label="Release changelog" dangerouslySetInnerHTML={{ __html: log.html.replace(/^<h1[^>]*>[\s\S]*?<\/h1>\n?/, "") }} />
      </div>
    </div>
  );
}
