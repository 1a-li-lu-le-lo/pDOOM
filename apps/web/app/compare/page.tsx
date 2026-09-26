// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { HORIZONS } from "@pdoom/schemas";
import { oneIn } from "@pdoom/model-core";
import { PageHeader } from "@/components/shell/PageHeader";
import { IconArray } from "@/components/charts/IconArray";
import { EstimateCard } from "@/components/meter/EstimateCard";
import { ShareLink } from "@/components/share/ShareLink";
import { DEFAULT_HORIZON, getRelease, headline, isValidHorizon, researchEstimate } from "@/lib/data";
import { horizonLabel, outcomeSetLabel } from "@/lib/format";

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
        {chosen && q ? (
          <>
            <div className="grid grid-2">
              <EstimateCard e={chosen} />
              <div className="card stack">
                <div className="eyebrow">The same number three ways</div>
                <p>
                  <strong>As a grid.</strong> Of one thousand futures consistent with this estimate's assumptions, the median run has about {Math.round(q.p50 * 1000).toLocaleString("en-US")} ending in {outcomeSetLabel(chosen.outcome_set).toLowerCase()} within {horizonLabel(chosen.horizon).toLowerCase()}. The plausible range is {Math.round(q.p05 * 1000).toLocaleString("en-US")} to {Math.round(q.p95 * 1000).toLocaleString("en-US")}.
                </p>
                <p>
                  <strong>As odds.</strong> Roughly {oneIn(q.p50)} at the median; between {oneIn(q.p95)} and {oneIn(q.p05)} across the interval.
                </p>
                <p>
                  <strong>As a complement.</strong> The same estimate says that in about {oneIn(1 - q.p50) === "1 in 1" ? "nearly all" : `${Math.round((1 - q.p50) * 1000).toLocaleString("en-US")} of one thousand`} futures, none of these outcomes occurs within the horizon. Both readings are the same statement.
                </p>
                <p className="cite">
                  No comparison with everyday risks is offered: those have base rates measured from events that happened, and this does not. See <Link href="/method/calibration">calibration</Link>.
                </p>
              </div>
            </div>
            <IconArray p={q.p50} low={q.p05} high={q.p95} label={`${outcomeSetLabel(chosen.outcome_set)}, ${horizonLabel(chosen.horizon)}, ${chosen.status === "research_mode" ? "research-mode model" : "external aggregate"}`} colorVar={chosen.status === "research_mode" ? "var(--c-disagreement)" : "var(--c-evidence)"} />
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
                <div className="cite">median {e.display.central} · interval {e.display.interval}</div>
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
