// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { BarList } from "@/components/charts/BarList";
import { getSnapshot } from "@/lib/data";
import { fmtDate, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Evidence",
  description: "The source ledger: every source in the snapshot with its tier, verification status, model-use status, licence and robots status.",
};

const TIER_LABEL: Record<number, string> = {
  1: "Tier 1 · primary, peer-reviewed or official",
  2: "Tier 2 · reputable secondary and institutional",
  3: "Tier 3 · pre-prints, trackers and platforms",
  4: "Tier 4 · informational only",
  5: "Tier 5 · excluded from the model",
};

export default async function EvidencePage({ searchParams }: { searchParams: Promise<{ tier?: string; topic?: string; q?: string; status?: string }> }) {
  const sp = await searchParams;
  const snap = await getSnapshot();
  const tier = sp.tier && /^[1-5]$/.test(sp.tier) ? Number(sp.tier) : undefined;
  const topic = sp.topic?.slice(0, 40);
  const q = sp.q?.trim().toLowerCase().slice(0, 80);
  const status = sp.status?.slice(0, 40);
  const topics = [...new Set(snap.sources.flatMap((s) => s.topic))].sort();
  const all = snap.sources;
  const items = all
    .filter((s) => (tier ? s.source_tier === tier : true))
    .filter((s) => (topic ? s.topic.includes(topic) : true))
    .filter((s) => (status ? s.verification.status === status : true))
    .filter((s) => (q ? `${s.title} ${s.publisher} ${s.authors.join(" ")} ${s.evidence_summary}`.toLowerCase().includes(q) : true))
    .sort((a, b) => a.source_tier - b.source_tier || (b.date_published ?? "").localeCompare(a.date_published ?? ""));
  const byTier = [1, 2, 3, 4, 5].map((t) => ({ label: TIER_LABEL[t]!, value: all.filter((s) => s.source_tier === t).length, display: String(all.filter((s) => s.source_tier === t).length), href: `/evidence?tier=${t}`, colorVar: t <= 2 ? "var(--c-evidence)" : t === 3 ? "var(--c-uncertainty)" : "var(--c-insufficient)" }));
  const byStatus = ["verified_fetch", "verified_search", "verified_prior_knowledge", "unverified"].map((st) => ({ label: titleCase(st), value: all.filter((s) => s.verification.status === st).length, display: String(all.filter((s) => s.verification.status === st).length), href: `/evidence?status=${st}`, colorVar: st.startsWith("verified_fetch") || st === "verified_search" ? "var(--c-evidence)" : "var(--c-insufficient)" }));
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="Evidence ledger"
          lede="Every source the observatory knows about, how it was verified, and whether the model is allowed to use it. Tier 4 and 5 sources are shown for transparency and never feed a computation."
          textAnchor="sources"
        />
        <div className="grid grid-2">
          <BarList title="Sources by tier" description="Count of sources per tier in the current snapshot." items={byTier} />
          <BarList title="Sources by verification status" description="Only sources verified by fetch or by search are eligible for the model; prior-knowledge and unverified entries are informational." items={byStatus} />
        </div>
        <form className="filters card" method="get" action="/evidence" aria-label="Filter sources">
          <label>
            Search
            <input type="search" name="q" defaultValue={sp.q ?? ""} placeholder="title, publisher, author" />
          </label>
          <label>
            Tier
            <select name="tier" defaultValue={tier ? String(tier) : ""}>
              <option value="">All tiers</option>
              {[1, 2, 3, 4, 5].map((t) => (
                <option key={t} value={t}>
                  Tier {t}
                </option>
              ))}
            </select>
          </label>
          <label>
            Topic
            <select name="topic" defaultValue={topic ?? ""}>
              <option value="">All topics</option>
              {topics.map((t) => (
                <option key={t} value={t}>
                  {titleCase(t)}
                </option>
              ))}
            </select>
          </label>
          <button className="btn btn-primary" type="submit">
            Filter
          </button>
          <Link className="btn" href="/evidence">
            Clear
          </Link>
        </form>
        <p className="muted" aria-live="polite">
          {items.length} of {all.length} sources
          {items.length === 0 ? " match these filters. Nothing is hidden: clear a filter to widen the ledger." : ""}
        </p>
        <div className="table-wrap table-wide" tabIndex={0} role="region" aria-label="Scrollable table">
          <table>
            <caption>Source ledger (filtered)</caption>
            <thead>
              <tr>
                <th>Source</th>
                <th>Publisher</th>
                <th>Date</th>
                <th>Tier</th>
                <th>Type</th>
                <th>Verification</th>
                <th>Model use</th>
                <th>Claims</th>
              </tr>
            </thead>
            <tbody>
              {items.map((s) => (
                <tr key={s.id}>
                  <td>
                    <Link href={`/evidence/sources/${s.id}`}>{s.title}</Link>
                    {s.retraction_status !== "none" ? <span className="badge badge-risk"> {titleCase(s.retraction_status)}</span> : null}
                  </td>
                  <td>{s.publisher}</td>
                  <td>{fmtDate(s.date_published)}</td>
                  <td className="num">{s.source_tier}</td>
                  <td>{titleCase(s.source_type)}</td>
                  <td>{titleCase(s.verification.status)}</td>
                  <td>{titleCase(s.model_use_status)}</td>
                  <td className="num">{s.claim_ids.length}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <p className="cite">
          Tier definitions and rules: <Link href="/method/source-hierarchy">docs/method/source-hierarchy.md</Link>. Download: <Link href="/api/export/sources.csv">sources.csv</Link>.
        </p>
      </div>
    </div>
  );
}
