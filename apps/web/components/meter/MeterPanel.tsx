// Copyright NU Cybernetics. p(DOOM) — research prototype.
import Link from "next/link";
import type { Release } from "@pdoom/sdk";
import { headline, indexById, researchEstimate } from "@/lib/data";
import { fmtDate, horizonLabel, titleCase } from "@/lib/format";
import { EstimateCard } from "./EstimateCard";
import { IndexGauge } from "./IndexGauge";

/**
 * The CURRENT METER PANEL. Server-rendered so the estimate is readable before
 * any 3D asset loads. The official value is withheld in this release line and
 * the panel says so; the external aggregate and research-mode model appear
 * beside it, separately labelled and never blended.
 */
export function MeterPanel({ rel, horizon }: { rel: Release; horizon: string }) {
  const h = headline(rel);
  const official = h.official.find((e) => e.horizon === horizon) ?? h.official[0];
  const research = researchEstimate(rel, "P_DOOM", horizon);
  const decomposition = ["O3", "P_COLLAPSE", "O6", "O7"].map((k) => researchEstimate(rel, k, horizon)).filter((e): e is NonNullable<typeof e> => !!e);
  const indexes = ["capability_pressure", "control_strength", "incident_pressure", "agentic_infrastructure_risk", "evidence_pressure", "uncertainty"]
    .map((id) => indexById(rel, id as never))
    .filter((i): i is NonNullable<typeof i> => !!i);
  const up = rel.drivers_explained.items.filter((d) => d.direction === "raises_pressure").sort((a, b) => b.magnitude - a.magnitude).slice(0, 3);
  const down = rel.drivers_explained.items.filter((d) => d.direction === "strengthens_control").sort((a, b) => b.magnitude - a.magnitude).slice(0, 3);
  return (
    <section className="meter" aria-labelledby="meter-heading">
      <div className="meter-headline">
        <div className="eyebrow">p(DOOM) · {horizonLabel(horizon)} · O3–O8 combined</div>
        <h2 id="meter-heading" className="meter-value withheld" style={{ marginTop: 0 }}>
          {official?.display.central ?? "No release"}
        </h2>
        <p className="lede" style={{ margin: 0 }}>
          {official?.display.note}
        </p>
        <div className="estimate-meta">
          <span>model {rel.manifest.model_versions[0]}</span>
          <span>data through {fmtDate(rel.manifest.source_cutoff)}</span>
          <span>last reviewed {fmtDate(rel.manifest.published)}</span>
          <span>editorial level: {titleCase(rel.manifest.editorial_risk_level)}</span>
        </div>
      </div>

      <div className="grid grid-2">
        {h.external
          .filter((e) => e.horizon === horizon || horizon === "2100" || horizon === "eventual" ? e.horizon === horizon : false)
          .slice(0, 2)
          .map((e) => (
            <EstimateCard key={e.estimate_id} e={e} compact />
          ))}
        {research ? <EstimateCard e={research} compact /> : null}
      </div>
      {h.external.filter((e) => e.horizon === horizon).length === 0 ? (
        <p className="muted" style={{ fontSize: "var(--fs-sm)" }}>
          No external forecast group asks about the {horizonLabel(horizon)} horizon; external aggregates exist for 2100 and open-ended horizons. See <Link href="/forecasts">Forecasts</Link>.
        </p>
      ) : null}

      <div>
        <div className="eyebrow">Decomposition (research mode, {horizonLabel(horizon)})</div>
        <div className="grid grid-4" style={{ marginTop: "var(--s-2)" }}>
          {decomposition.map((e) => (
            <EstimateCard key={e.estimate_id} e={e} compact hideStrip />
          ))}
        </div>
      </div>

      <div>
        <div className="eyebrow">Indexes (0–100; not probabilities)</div>
        <div className="gauge-row" style={{ marginTop: "var(--s-2)" }}>
          {indexes.map((i) => (
            <IndexGauge key={i.index_id} index={i} />
          ))}
        </div>
      </div>

      <div className="grid grid-2">
        <div>
          <div className="eyebrow">Top upward drivers (index points)</div>
          <ol style={{ paddingLeft: "1.2em" }}>
            {up.map((d) => (
              <li key={d.driver + d.index_id}>
                <code>{d.driver}</code> +{Math.round(d.magnitude * 10) / 10} on {titleCase(d.index_id)}
              </li>
            ))}
          </ol>
        </div>
        <div>
          <div className="eyebrow">Top downward drivers (control strength)</div>
          <ol style={{ paddingLeft: "1.2em" }}>
            {down.map((d) => (
              <li key={d.driver + d.index_id}>
                <code>{d.driver}</code> +{Math.round(d.magnitude * 10) / 10} on {titleCase(d.index_id)}
              </li>
            ))}
          </ol>
        </div>
      </div>

      <div>
        <div className="eyebrow">What changed</div>
        <ul>
          {rel.manifest.changes.map((c) => (
            <li key={c}>{c}</li>
          ))}
        </ul>
      </div>

      <div className="row no-print">
        <Link className="btn btn-primary" href="/meter#why">Inspect the assumptions</Link>
        <Link className="btn" href="/forecasts">Compare forecasts</Link>
        <Link className="btn" href={`/meter?horizon=${horizon === "10y" ? "2100" : "10y"}`}>Change horizon</Link>
        <Link className="btn" href="/futures">Explore scenarios</Link>
        <Link className="btn" href="/text">Read plain text</Link>
        <Link className="btn" href="/api/export/release.json">Download data</Link>
        <Link className="btn" href="/act">Help reduce risk</Link>
      </div>
    </section>
  );
}
