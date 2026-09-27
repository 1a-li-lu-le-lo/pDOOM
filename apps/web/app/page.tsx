// Copyright NU Cybernetics. p(DOOM) — research prototype.
import Link from "next/link";
import { TAGLINES } from "@pdoom/schemas";
import { HomeStage, type StageData } from "@/components/scenes/HomeStage";
import { MeterPanel } from "@/components/meter/MeterPanel";
import { ModeSwitcher } from "@/components/mode/ModeSwitcher";
import { DEFAULT_HORIZON, getRelease, getSnapshot, headline, indexById, researchEstimate } from "@/lib/data";
import { horizonLabel } from "@/lib/format";

export default async function Home() {
  const [rel, snap] = await Promise.all([getRelease(), getSnapshot()]);
  const h = headline(rel);
  const research = researchEstimate(rel, "P_DOOM", DEFAULT_HORIZON);
  const featured = research ?? h.external[0];
  const epi = indexById(rel, "evidence_pressure");
  const stage: StageData = {
    intervalWidth: featured?.quantiles ? featured.quantiles.p95 - featured.quantiles.p05 : 0.2,
    brightness: epi?.value ?? 50,
    particles: Math.min(400, snap.sources.length * 2),
    safeguards: Math.min(12, snap.interventions.length),
    seed: snap.model_spec.experimental_causal.seed,
    headlineText: h.official[0]?.display.central ?? "",
    intervalText: featured?.display.interval ?? "",
    horizonLabel: horizonLabel(DEFAULT_HORIZON),
    scenarios: snap.scenarios.map((s) => ({ id: s.id, name: s.name, recoverability: s.recoverability })),
    interventions: snap.interventions.map((i) => ({ id: i.id, name: i.name })),
    indexes: rel.indexes.map((i) => ({ id: i.index_id, value: i.value })),
    researchCurve: ["1y", "3y", "5y", "10y", "25y", "2100", "eventual"]
      .map((hz) => researchEstimate(rel, "P_DOOM", hz))
      .filter((e): e is NonNullable<typeof e> => !!e && !!e.quantiles)
      .map((e) => ({ x: e.horizon, low: e.quantiles!.p05, mid: e.quantiles!.p50, high: e.quantiles!.p95 })),
  };
  return (
    <>
      <section className="hero">
        <HomeStage data={stage} />
        <div className="container hero-content">
          <div>
            <div className="scene-label">Conceptual risk visualization · not a simulation of AI risk</div>
            <h1 className="hero-title">{TAGLINES[0]}</h1>
            <p className="hero-sub">
              An open-methodology observatory of evidence about catastrophic and existential risk from advanced AI. Every number comes with its outcome, horizon, interval, disagreement and sources. The official value is withheld until it can be calibrated; the assumptions are open for inspection.
            </p>
            <div className="row">
              <Link className="btn btn-primary" href="/meter">Inspect the assumptions</Link>
              <Link className="btn" href="/text">Read the plain-text version</Link>
            </div>
            <div className="hero-modes">
              <span className="eyebrow">See it as</span>
              <ModeSwitcher />
            </div>
          </div>
        </div>
      </section>
      <section className="page">
        <div className="container">
          <MeterPanel rel={rel} horizon={DEFAULT_HORIZON} />
        </div>
      </section>
    </>
  );
}
