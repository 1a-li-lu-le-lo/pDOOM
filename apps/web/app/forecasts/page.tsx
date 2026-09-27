// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { DistributionDots } from "@/components/charts/DistributionDots";
import { getRelease, getSnapshot } from "@/lib/data";
import { fmtDate, horizonLabel, outcomeSetLabel, pct, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Forecasts",
  description: "Every external forecast in the snapshot with its original question wording, population, outcome set, horizon and conditioning, aggregated only inside compatibility groups.",
};

// Member forecasts are quoted as their sources state them; computed aggregates follow the release rounding rule.
const asRecorded = (p: number | null) => (p === null ? "—" : `${Math.round(p * 1000) / 10}%`);

export default async function ForecastsPage() {
  const [rel, snap] = await Promise.all([getRelease(), getSnapshot()]);
  const src = new Map(snap.sources.map((s) => [s.id, s]));
  const groups = [...new Set(rel.aggregations.map((a) => a.group_id))];
  const forecastsByGroup = (g: string) => snap.forecasts.filter((f) => f.group_id === g);
  const ungrouped = snap.forecasts.filter((f) => !f.group_id || !groups.includes(f.group_id));
  const dots = snap.forecasts
    .filter((f) => f.median !== null || f.mean !== null)
    .map((f) => ({ label: f.forecaster_or_survey, value: (f.median ?? f.mean)!, group: `${outcomeSetLabel(f.outcome_set)} · ${horizonLabel(f.horizon)}${f.horizon === "custom" && f.horizon_end_year ? ` (${f.horizon_end_year})` : ""}`, colorVar: f.population === "superforecasters" ? "var(--c-safeguard)" : f.population === "prediction_market" ? "var(--c-uncertainty)" : "var(--c-evidence)" }));
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="External forecasts"
          lede="Surveys, tournaments and platform forecasts, each shown with the question that was actually asked. They are only ever combined inside a compatibility group: the same outcome set, the same horizon, comparable conditioning."
          textAnchor="forecasts"
        />

        <DistributionDots
          dots={dots}
          title="All forecasts with a central value"
          description="One dot per forecast, one row per outcome set and horizon. Dots on different rows are answers to different questions and must not be compared directly. Blue dots are superforecaster populations, amber dots are prediction markets."
        />

        {groups.map((g) => {
          const aggs = rel.aggregations.filter((a) => a.group_id === g);
          const pref = aggs.find((a) => a.preferred) ?? aggs[0]!;
          const unc = rel.estimates.find((e) => e.group_id === g)?.uncertainty ?? "high";
          const members = forecastsByGroup(g);
          return (
            <section key={g} id={g} className="card stack" aria-labelledby={`${g}-h`}>
              <div className="row">
                <span className="badge badge-evidence">group {g}</span>
                <span className="badge">{horizonLabel(pref.horizon)}</span>
                <span className="badge">{members.length} forecasts</span>
              </div>
              <h2 id={`${g}-h`} style={{ margin: 0 }}>
                {outcomeSetLabel(pref.outcome_set)}
              </h2>
              <p className="muted">{pref.conditioning}</p>
              <div className="table-wrap" tabIndex={0} role="region" aria-label="Scrollable table">
                <table>
                  <caption>Aggregation methods for {g}; the preferred method is marked</caption>
                  <thead>
                    <tr>
                      <th>Method</th>
                      <th className="num">Value (rounded as the release displays it)</th>
                      <th>Note</th>
                    </tr>
                  </thead>
                  <tbody>
                    {aggs.map((a) => (
                      <tr key={a.aggregation_id}>
                        <td>
                          {titleCase(a.method)} {a.preferred ? <span className="badge badge-evidence">preferred</span> : null}
                        </td>
                        <td className="num">{pct(a.value, unc)}</td>
                        <td className="muted">{a.note}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <details>
                <summary>Member forecasts and original wording</summary>
                <div className="stack">
                  {members.map((f) => (
                    <ForecastRow key={f.id} f={f} sourceTitle={src.get(f.source_id)?.title} />
                  ))}
                </div>
              </details>
              {pref.transformation_note ? (
                <p className="cite">Mapping to outcome taxonomy: {pref.transformation_note}</p>
              ) : null}
            </section>
          );
        })}

        <section className="card stack" aria-labelledby="ungrouped-h">
          <h2 id="ungrouped-h" style={{ margin: 0 }}>
            Forecasts shown alone
          </h2>
          <p className="muted">These ask a question no other forecast in the snapshot asks, so they are displayed but never aggregated.</p>
          {ungrouped.map((f) => (
            <ForecastRow key={f.id} f={f} sourceTitle={src.get(f.source_id)?.title} />
          ))}
        </section>

        <p className="cite">
          Aggregation method: <Link href="/method/aggregation">docs/method/aggregation.md</Link>. Disagreement labels: <Link href="/method/disagreement">docs/method/disagreement.md</Link>.
        </p>
      </div>
    </div>
  );
}

function ForecastRow({ f, sourceTitle }: { f: import("@pdoom/schemas").Forecast; sourceTitle?: string }) {
  return (
    <article className="forecast-row" aria-label={f.forecaster_or_survey}>
      <div className="row">
        <strong>{f.forecaster_or_survey}</strong>
        <span className="badge">{titleCase(f.population)}</span>
        <span className="badge">{fmtDate(f.date)}</span>
        {f.sample_size ? <span className="badge">n = {f.sample_size.toLocaleString("en-US")}</span> : null}
        {f.paraphrase ? <span className="badge badge-uncertainty">paraphrased</span> : null}
        {f.status !== "current" ? <span className="badge badge-insufficient">{titleCase(f.status)}</span> : null}
      </div>
      <p className="quote">“{f.question_wording_original}”</p>
      <dl className="kv">
        <dt>Central value</dt>
        <dd>
          {f.median !== null ? `median ${asRecorded(f.median)}` : f.mean !== null ? `mean ${asRecorded(f.mean)}` : "not stated"}
          {f.quantiles?.p25 !== undefined && f.quantiles?.p75 !== undefined ? ` · IQR ${asRecorded(f.quantiles.p25)}–${asRecorded(f.quantiles.p75)}` : ""}
          {f.median !== null || f.mean !== null ? " (as the source states it)" : ""}
        </dd>
        <dt>Outcome set</dt>
        <dd>{outcomeSetLabel(f.outcome_set)}</dd>
        <dt>Horizon</dt>
        <dd>
          {horizonLabel(f.horizon)}
          {f.horizon_note ? ` — ${f.horizon_note}` : ""}
        </dd>
        <dt>Conditions</dt>
        <dd>{f.conditions}</dd>
        {f.selection_effects ? (
          <>
            <dt>Selection effects</dt>
            <dd>{f.selection_effects}</dd>
          </>
        ) : null}
        {f.framing_effects ? (
          <>
            <dt>Framing effects</dt>
            <dd>{f.framing_effects}</dd>
          </>
        ) : null}
        <dt>Source</dt>
        <dd>
          <Link href={`/evidence/sources/${f.source_id}`}>{sourceTitle ?? f.source_id}</Link>
        </dd>
      </dl>
    </article>
  );
}
