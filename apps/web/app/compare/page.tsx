// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { HORIZONS } from "@pdoom/schemas";
import { oneIn, roundingStep } from "@pdoom/model-core";
import { PageHeader } from "@/components/shell/PageHeader";
import { IconArray } from "@/components/charts/IconArray";
import { EstimateCard } from "@/components/meter/EstimateCard";
import { ShareLink } from "@/components/share/ShareLink";
import { DEFAULT_HORIZON, getRelease, headline, isValidHorizon, researchEstimate } from "@/lib/data";
import { horizonLabel, outcomeSetLabel, tidyInterval } from "@/lib/format";

export const metadata: Metadata = {
  title: "Compare",
  description: "See a probability as a count of futures rather than a percentage: how many of a thousand paths, one in how many, and how the horizon changes the picture.",
};

export default async function ComparePage({ searchParams }: { searchParams: Promise<{ horizon?: string; estimate?: string }> }) {
  const sp = await searchParams;
  const rel = await getRelease();
  const horizon = isValidHorizon(sp.horizon) ? sp.horizon : DEFAULT_HORIZON;
  const h = headline(rel);
  const candidates = [...h.research.filter((e) => e.outcome_set.length === 6 && e.horizon === horizon), ...h.external.filter((e) => e.horizon === horizon)];
  const chosen = candidates.find((e) => e.estimate_id === sp.estimate) ?? candidates[0];
  const q = chosen?.quantiles ?? null;
  // Build-spec §0.8: counts and odds derive from the display-rounded value, never from the raw quantile,
  // so the comparator is exactly as precise as the release card beside it.
  const step = chosen ? roundingStep(chosen.uncertainty) : 1;
  const rounded = (p: number) => (Math.round((p * 100) / step) * step) / 100;
  const r = q ? { p05: rounded(q.p05), p50: rounded(q.p50), p95: rounded(q.p95) } : null;
  const CELLS = 1000;
  const ONE_POINT = CELLS / 100;
  /** True when the release itself shows the below-one-point guard: a positive value that rounds to zero at this step. */
  const belowOnePoint = (raw: number, rp: number) => raw > 0 && rp * CELLS < ONE_POINT;
  const countOf = (raw: number, rp: number) => (belowOnePoint(raw, rp) ? `fewer than ${ONE_POINT.toLocaleString("en-US")}` : `about ${Math.round(rp * CELLS).toLocaleString("en-US")}`);
  const oddsOf = (raw: number, rp: number) => (belowOnePoint(raw, rp) ? `rarer than 1 in ${(CELLS / ONE_POINT).toLocaleString("en-US")}` : `roughly ${oneIn(rp)}`);
  const complementOf = (raw: number, rp: number) => (belowOnePoint(raw, rp) ? `more than ${(CELLS - ONE_POINT).toLocaleString("en-US")}` : `about ${(CELLS - Math.round(rp * CELLS)).toLocaleString("en-US")}`);
  const allHorizons = HORIZONS.map((hz) => researchEstimate(rel, "P_DOOM", hz.key)).filter((e): e is NonNullable<typeof e> => !!e && !!e.quantiles);
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="A probability as a count of futures"
          lede="Percentages hide their own size. The same number reads differently as cells in a grid, as one path in so many, and across horizons. Nothing here is more precise than the release it comes from."
          textAnchor="estimate"
        >
          <nav className="horizon-tabs" aria-label="Horizon">
            {HORIZONS.map((hz) => (
              <Link key={hz.key} href={`/compare?horizon=${hz.key}`} aria-current={hz.key === horizon ? "page" : undefined}>
                {hz.label}
              </Link>
            ))}
          </nav>
        </PageHeader>
        {candidates.length > 1 ? (
          <nav className="row" aria-label="Estimate">
            {candidates.map((e) => (
              <Link key={e.estimate_id} className="btn" href={`/compare?horizon=${horizon}&estimate=${e.estimate_id}`} aria-current={chosen?.estimate_id === e.estimate_id ? "page" : undefined}>
                {e.status === "research_mode" ? "Research-mode model" : `External: ${e.group_id ?? outcomeSetLabel(e.outcome_set)}`}
              </Link>
            ))}
          </nav>
        ) : null}
        {chosen && q && r ? (
          <>
            <div className="grid grid-2">
              <EstimateCard e={chosen} />
              <div className="card stack">
                <div className="eyebrow">The same number three ways</div>
                <p>
                  <strong>As a grid.</strong> Of one thousand futures consistent with this estimate's assumptions, {countOf(q.p50, r.p50)} end in {outcomeSetLabel(chosen.outcome_set)} ({horizonLabel(chosen.horizon).toLowerCase()}) at the median. The plausible range is {countOf(q.p05, r.p05)} to {countOf(q.p95, r.p95)}.
                </p>
                <p>
                  <strong>As odds.</strong> At the median, {oddsOf(q.p50, r.p50)}; across the interval, between {oddsOf(q.p95, r.p95)} and {oddsOf(q.p05, r.p05)}.
                </p>
                <p>
                  <strong>As a complement.</strong> The same estimate says that in {complementOf(q.p50, r.p50)} of one thousand futures none of these outcomes occurs within the horizon. Both readings are the same statement.
                </p>
                <p className="cite">
                  No comparison with everyday risks is offered: those have base rates measured from events that happened, and this does not. See <Link href="/method/calibration">calibration</Link>.
                </p>
              </div>
            </div>
            <IconArray p={r.p50} low={r.p05} high={r.p95} cells={CELLS} label={`${outcomeSetLabel(chosen.outcome_set)}, ${horizonLabel(chosen.horizon)}, ${chosen.status === "research_mode" ? "research-mode model" : "external aggregate"}`} colorVar={chosen.status === "research_mode" ? "var(--c-disagreement)" : "var(--c-evidence)"} />
          </>
        ) : (
          <p className="muted">No estimate with a published interval exists for this horizon.</p>
        )}
        <section className="card stack">
          <h2 style={{ marginTop: 0 }}>How the horizon changes the picture</h2>
          <p className="muted">Research-mode medians per horizon, each as cells of one hundred. Longer horizons include everything shorter ones do, plus more time for anything to happen.</p>
          <div className="grid grid-4">
            {allHorizons.map((e) => (
              <div key={e.estimate_id} className="stack" style={{ gap: "var(--s-1)" }}>
                <div className="eyebrow">{horizonLabel(e.horizon)}</div>
                <IconArray p={e.quantiles!.p50} high={e.quantiles!.p95} cells={100} label={`Research-mode p(DOOM), ${horizonLabel(e.horizon)}`} colorVar="var(--c-disagreement)" />
                <div className="cite">median {e.display.central} · interval {tidyInterval(e.display.interval)}</div>
              </div>
            ))}
          </div>
        </section>
        <div className="row no-print">
          <ShareLink label="Copy link to this comparison" />
          <Link className="btn" href="/lab">
            Change the assumptions
          </Link>
        </div>
      </div>
    </div>
  );
}
