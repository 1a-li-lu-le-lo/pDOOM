// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { Waterfall } from "@/components/charts/Waterfall";
import { IndexGauge } from "@/components/meter/IndexGauge";
import { getRelease, getSnapshot, indexById } from "@/lib/data";
import { fmtDate, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Incidents",
  description: "Verified AI incidents and near misses described at category level, with severity, relevance, evidence level and the registries that record them.",
};

export default async function IncidentsPage({ searchParams }: { searchParams: Promise<{ severity?: string; relevance?: string }> }) {
  const sp = await searchParams;
  const [rel, snap] = await Promise.all([getRelease(), getSnapshot()]);
  const ipi = indexById(rel, "incident_pressure");
  const items = [...snap.incidents]
    .filter((i) => (sp.severity ? i.severity === sp.severity : true))
    .filter((i) => (sp.relevance ? i.pdoom_relevance === sp.relevance : true))
    .sort((a, b) => (b.date ?? "").localeCompare(a.date ?? ""));
  const severities = [...new Set(snap.incidents.map((i) => i.severity))];
  const relevances = [...new Set(snap.incidents.map((i) => i.pdoom_relevance))];
  const src = new Map(snap.sources.map((s) => [s.id, s]));
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="Incidents"
          lede="What has already happened, described without graphic or operational detail. Each entry is weighted by severity, relevance, evidence and recency into the Incident Pressure Index."
          textAnchor="incidents"
        />
        <div className="grid grid-2">
          <div>{ipi ? <IndexGauge index={ipi} /> : null}</div>
          <div className="card">
            <p className="muted" style={{ margin: 0 }}>
              {ipi?.note}
            </p>
          </div>
        </div>
        {ipi && ipi.components.length ? (
          <Waterfall
            title="Incident Pressure Index contributions"
            description="Each incident's contribution after severity, relevance, evidence and recency weighting, before the saturating squash."
            items={ipi.components.map((c) => ({ label: c.signal_id, value: c.contribution, direction: c.contribution >= 0 ? "raises" : "lowers", note: `weight ${c.weight.toFixed(2)}` }))}
          />
        ) : null}
        <form className="filters card" method="get" action="/incidents" aria-label="Filter incidents">
          <label>
            Severity
            <select name="severity" defaultValue={sp.severity ?? ""}>
              <option value="">All</option>
              {severities.map((s) => (
                <option key={s} value={s}>
                  {titleCase(s)}
                </option>
              ))}
            </select>
          </label>
          <label>
            Relevance
            <select name="relevance" defaultValue={sp.relevance ?? ""}>
              <option value="">All</option>
              {relevances.map((s) => (
                <option key={s} value={s}>
                  {titleCase(s)}
                </option>
              ))}
            </select>
          </label>
          <button className="btn btn-primary" type="submit">
            Filter
          </button>
          <Link className="btn" href="/incidents">
            Clear
          </Link>
        </form>
        <p className="muted" aria-live="polite">
          {items.length} of {snap.incidents.length} incidents
        </p>
        <div className="stack">
          {items.map((i) => (
            <article key={i.id} id={i.id} className="card stack">
              <div className="row">
                <span className={`badge ${i.severity === "catastrophic" || i.severity === "severe" ? "badge-risk" : i.near_miss ? "badge-uncertainty" : ""}`}>{i.near_miss ? "near miss" : titleCase(i.severity)}</span>
                <span className="badge">{fmtDate(i.date)}</span>
                <span className="badge">{titleCase(i.pdoom_relevance)} relevance</span>
                <span className="badge">{titleCase(i.evidence_level)}</span>
                <span className="badge">{titleCase(i.novelty)}</span>
                {i.jurisdiction ? <span className="badge">{i.jurisdiction}</span> : null}
              </div>
              <h2 style={{ margin: 0, fontSize: "var(--fs-xl)" }}>{i.title}</h2>
              <p>{i.summary}</p>
              <dl className="kv">
                <dt>Causes</dt>
                <dd>{i.cause.map(titleCase).join(", ")}</dd>
                <dt>Harm</dt>
                <dd>{i.harm.map(titleCase).join(", ")}</dd>
                <dt>Systems involved</dt>
                <dd>{i.systems_involved.join(", ") || "—"}</dd>
                {i.exposure_note ? (
                  <>
                    <dt>Exposure</dt>
                    <dd>{i.exposure_note}</dd>
                  </>
                ) : null}
                <dt>Registries</dt>
                <dd>
                  {Object.entries(i.external_ids)
                    .filter(([, v]) => v)
                    .map(([k, v]) => `${k.toUpperCase()}: ${v}`)
                    .join(" · ") || "none"}
                </dd>
                <dt>Sources</dt>
                <dd>
                  {i.source_ids.map((s, k) => (
                    <span key={s}>
                      {k ? "; " : ""}
                      <Link href={`/evidence/sources/${s}`}>{src.get(s)?.title ?? s}</Link>
                    </span>
                  ))}
                </dd>
                <dt>Verification</dt>
                <dd>
                  {titleCase(i.verification.status)} · {titleCase(i.human_review_status)} · {titleCase(i.model_use_status)}
                </dd>
              </dl>
            </article>
          ))}
        </div>
        <p className="cite">
          Scoring: <Link href="/method/indexes">docs/method/indexes.md</Link>. Content rules: <Link href="/method/content-safety">content-safety.md</Link>. Download: <Link href="/api/export/incidents.csv">incidents.csv</Link>.
        </p>
      </div>
    </div>
  );
}
