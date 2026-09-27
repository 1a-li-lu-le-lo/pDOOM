// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { HORIZONS } from "@pdoom/schemas";
import { PageHeader } from "@/components/shell/PageHeader";
import { MeterPanel } from "@/components/meter/MeterPanel";
import { EstimateCard } from "@/components/meter/EstimateCard";
import { BarList } from "@/components/charts/BarList";
import { LineWithBand } from "@/components/charts/LineWithBand";
import { ShareLink } from "@/components/share/ShareLink";
import { DEFAULT_HORIZON, getRelease, getSnapshot, headline, indexById, isValidHorizon, researchEstimate } from "@/lib/data";
import { INDEX_SHORT, fmtDate, horizonLabel, outcomeSetLabel, pct, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Meter",
  description: "The current p(DOOM) meter: official status, external forecast aggregates, the research-mode model, indexes and the assumptions behind every number.",
};

// Computed values follow the release's rounding rule for the matching estimate (high uncertainty when unknown).
const roundedAs = (rel: Awaited<ReturnType<typeof getRelease>>, groupId: string | null, p: number) => pct(p, rel.estimates.find((e) => e.group_id === groupId)?.uncertainty ?? "high");

export default async function MeterPage({ searchParams }: { searchParams: Promise<{ horizon?: string }> }) {
  const sp = await searchParams;
  const horizon = isValidHorizon(sp.horizon) ? sp.horizon : DEFAULT_HORIZON;
  const [rel, snap] = await Promise.all([getRelease(), getSnapshot()]);
  const h = headline(rel);
  const research = researchEstimate(rel, "P_DOOM", horizon);
  const unc = indexById(rel, "uncertainty");
  const curve = HORIZONS.map((hz) => researchEstimate(rel, "P_DOOM", hz.key))
    .filter((e): e is NonNullable<typeof e> => !!e && !!e.quantiles)
    .map((e) => ({ x: horizonLabel(e.horizon), low: e.quantiles!.p05, mid: e.quantiles!.p50, high: e.quantiles!.p95 }));
  const groups = rel.aggregations.filter((a) => a.preferred);
  const sens = [...rel.sensitivity].sort((a, b) => Math.abs(b.delta) - Math.abs(a.delta)).slice(0, 8);
  const spec = snap.model_spec;
  const rule = spec.editorial_rules.find((r) => r.level === rel.manifest.editorial_risk_level);
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="The meter"
          lede="Everything on this page comes from one signed release. Change the horizon to see how each object moves; open the assumptions to see why."
          textAnchor="estimate"
        >
          <nav className="horizon-tabs" aria-label="Horizon">
            {HORIZONS.map((hz) => (
              <Link key={hz.key} href={`/meter?horizon=${hz.key}`} aria-current={hz.key === horizon ? "page" : undefined}>
                {hz.label}
              </Link>
            ))}
          </nav>
        </PageHeader>

        <MeterPanel rel={rel} horizon={horizon} />

        <section id="why" className="card stack" aria-labelledby="why-heading">
          <h2 id="why-heading" style={{ marginTop: 0 }}>Why these numbers</h2>
          <p className="lede">
            The official object is withheld because no documented calibration process exists yet. Below are the three things that are published instead, and what would move each one.
          </p>

          <h3>1. What an official value would need</h3>
          <ul>
            {(h.official[0]?.assumptions ?? []).map((a) => (
              <li key={a}>{a}</li>
            ))}
            <li>{h.official[0]?.conditioning}</li>
            <li>
              Read <Link href="/method/calibration">what would change this</Link> and the decision record in the <Link href="/method/model">public methodology</Link>.
            </li>
          </ul>

          <h3>2. External forecasts, aggregated inside compatibility groups</h3>
          <p>
            Forecasts are only combined when they ask about the same outcome set and horizon. The preferred method is the unweighted median because it is robust to a single extreme forecast; every other method is published beside it.
          </p>
          <BarList
            title="Preferred aggregate per compatibility group"
            description="Each bar is the preferred aggregate for one group. The label names the outcome set and horizon; hover or read the table for the method and member count."
            items={groups.map((g) => ({
              label: `${g.group_id} · ${outcomeSetLabel(g.outcome_set)} · ${horizonLabel(g.horizon)}`,
              value: g.value,
              display: roundedAs(rel, g.group_id, g.value),
              note: `${titleCase(g.method)}, ${g.n} ${g.n === 1 ? "forecast" : "forecasts"} from ${g.population_count} ${g.population_count === 1 ? "population" : "populations"}`,
              href: `/forecasts#${g.group_id}`,
            }))}
          />

          <h3>3. The research-mode model across horizons</h3>
          {curve.length ? (
            <LineWithBand
              points={curve}
              title="Research-mode p(DOOM) by horizon"
              description="Median with the p05–p95 band. The curve rises with the horizon because more time allows more pathways; it is a model output under stated parameters, not a forecast of when anything happens."
              colorVar="var(--c-disagreement)"
            />
          ) : null}
          {research ? <EstimateCard e={research} /> : null}

          <h3>Which single change matters most</h3>
          <p>
            Leave-one-out and parameter runs, ranked by the size of their effect. A large effect from removing one source means the aggregate depends on that source.
          </p>
          <div className="table-wrap table-wide" tabIndex={0} role="region" aria-label="Sensitivity runs">
            <table>
              <caption>Top sensitivity runs in this release</caption>
              <thead>
                <tr>
                  <th>Run</th>
                  <th>Target</th>
                  <th>Change</th>
                  <th className="num">Baseline</th>
                  <th className="num">Result</th>
                  <th className="num">Change (whole points)</th>
                </tr>
              </thead>
              <tbody>
                {sens.map((s) => (
                  <tr key={s.run_id}>
                    <td>{titleCase(s.kind)}</td>
                    <td>{s.target_id}</td>
                    <td>{s.parameter_change}</td>
                    <td className="num">{roundedAs(rel, s.target_kind === "aggregation_group" ? s.target_id : null, s.baseline_value)}</td>
                    <td className="num">{roundedAs(rel, s.target_kind === "aggregation_group" ? s.target_id : null, s.value)}</td>
                    <td className="num">{Math.round(s.delta * 100) === 0 ? "under 1" : `${s.delta > 0 ? "+" : ""}${Math.round(s.delta * 100)}`}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <h3>Uncertainty score</h3>
          {unc ? (
            <>
              <p>
                {INDEX_SHORT.uncertainty}: <strong>{Math.round(unc.value ?? 0)}</strong> out of 100. {unc.note}
              </p>
              <BarList
                title="Uncertainty components"
                description="Weighted components of the uncertainty score. The calibration gap is fixed at its maximum until a calibration process exists."
                items={unc.components.map((c) => ({ label: titleCase(c.signal_id), value: c.contribution, display: c.contribution.toFixed(1), note: `weight ${c.weight}, value ${c.value_normalized.toFixed(2)}` }))}
                max={Math.max(1, ...unc.components.map((c) => c.contribution))}
              />
            </>
          ) : null}

          <h3>Editorial risk level</h3>
          <p>
            <span className="badge badge-uncertainty">{titleCase(rel.manifest.editorial_risk_level)}</span> {rule ? <>Rule: {rule.when}</> : null}
          </p>

          <h3>Provenance</h3>
          <dl className="kv">
            <dt>Release</dt>
            <dd>
              <Link href={`/releases/${rel.manifest.release_id}`}>{rel.manifest.release_id}</Link>
            </dd>
            <dt>Data snapshot</dt>
            <dd>{rel.manifest.data_snapshot}</dd>
            <dt>Model versions</dt>
            <dd>{rel.manifest.model_versions.join(", ")}</dd>
            <dt>Source cutoff</dt>
            <dd>{fmtDate(rel.manifest.source_cutoff)}</dd>
            <dt>Reviewers</dt>
            <dd>{rel.manifest.reviewers.join(", ")}</dd>
            <dt>Reproduce</dt>
            <dd>
              <code>{rel.manifest.reproduction_command}</code>
            </dd>
          </dl>
          <div className="row no-print">
            <ShareLink label="Copy link to this view" />
            <Link className="btn" href="/api/export/release.json">Download release JSON</Link>
            <Link className="btn" href="/lab">Try your own assumptions</Link>
          </div>
        </section>
      </div>
    </div>
  );
}
