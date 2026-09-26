// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { Suspense } from "react";
import { PageHeader } from "@/components/shell/PageHeader";
import { ScenarioLab } from "@/components/lab/ScenarioLab";
import { getRelease, getSnapshot, researchEstimate } from "@/lib/data";
import { tidyInterval } from "@/lib/format";

export const metadata: Metadata = {
  title: "Scenario Lab",
  description: "Move ten assumptions and watch the research-mode model respond. Results are labelled as your scenario, never as the official value, and every link reproduces the same result.",
};

export default async function LabPage() {
  const [rel, snap] = await Promise.all([getRelease(), getSnapshot()]);
  const spec = snap.model_spec.experimental_causal;
  const baseline = Object.fromEntries(
    Object.keys(spec.horizons).map((h) => {
      const e = researchEstimate(rel, "P_DOOM", h);
      return [h, e?.quantiles ? { p05: e.quantiles.p05, p50: e.quantiles.p50, p95: e.quantiles.p95, display: e.display.central, interval: tidyInterval(e.display.interval) } : null];
    }),
  );
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="Scenario Lab"
          lede="Ten dials, one honest model. Each dial shifts a factor of the experimental causal model on the logit scale by a documented amount. Nothing you do here changes a published number, and the result is always labelled as yours."
          textAnchor="method"
        />
        <Suspense fallback={<p className="muted">Loading the Scenario Lab…</p>}>
          <ScenarioLab spec={spec} baseline={baseline} releaseId={rel.manifest.release_id} />
        </Suspense>
        <p className="cite">
          Model: <Link href="/method/causal-model">docs/method/causal-model.md</Link>. Dial mapping: <Link href="/method/model">docs/method/model.md</Link>. The same code runs in the Go model and in this page, verified against shared golden fixtures.
        </p>
      </div>
    </div>
  );
}
