// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { Timeline } from "@/components/charts/Timeline";
import { Waterfall } from "@/components/charts/Waterfall";
import { IndexGauge } from "@/components/meter/IndexGauge";
import { SignalTable } from "@/components/meter/SignalTable";
import { getRelease, getSnapshot, indexById } from "@/lib/data";
import { titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Capabilities",
  description: "Frontier capability indicators: benchmark trajectories with their limitations, the driver signals behind the Capability Pressure Index, and what each contributes.",
};

export default async function CapabilitiesPage() {
  const [rel, snap] = await Promise.all([getRelease(), getSnapshot()]);
  const cpi = indexById(rel, "capability_pressure");
  const drivers = snap.drivers.filter((d) => ["D1", "D4"].includes(d.id));
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="Capabilities"
          lede="What frontier systems can do, measured by named benchmarks with their limitations attached. Capability pressure is an index of movement, not a probability of anything."
          textAnchor="capabilities"
        />
        <div className="grid grid-2">
          <div>{cpi ? <IndexGauge index={cpi} /> : null}</div>
          <div className="card">
            <p className="muted" style={{ margin: 0 }}>
              {cpi?.note}
            </p>
          </div>
        </div>
        {cpi ? (
          <Waterfall
            title="Capability Pressure Index contributions"
            description="Each bar is one signal's weighted, tier-adjusted contribution in index points. The order is the model specification's order; the sum is the index."
            items={cpi.components.map((c) => ({ label: c.signal_id, value: c.contribution, direction: c.contribution >= 0 ? "raises" : "lowers", note: `weight ${c.weight}, normalised ${c.value_normalized.toFixed(2)}, tier ${c.tier}` }))}
          />
        ) : null}
        {snap.benchmarks.map((b) => {
          const results = snap.benchmark_results.filter((r) => r.benchmark_id === b.id);
          return (
            <section key={b.id} id={b.id} className="card stack" aria-labelledby={`${b.id}-h`}>
              <div className="row">
                <span className="badge badge-evidence">benchmark</span>
                <span className="badge">{titleCase(b.contamination_risk)} contamination risk</span>
                <span className="badge">saturation: {titleCase(b.saturation)}</span>
                <span className="badge">{titleCase(b.pdoom_relevance)} relevance</span>
              </div>
              <h2 id={`${b.id}-h`} style={{ margin: 0 }}>
                {b.name}
              </h2>
              <p className="muted">
                {b.maintainer}
                {b.version ? ` · ${b.version}` : ""} ·{" "}
                <a href={b.url} rel="noopener noreferrer">
                  {b.url}
                </a>
              </p>
              <p>{b.tasks}</p>
              <Timeline
                points={results.map((r) => ({ date: r.date, value: r.value, label: r.model_name, low: r.ci_low, high: r.ci_high, note: r.note ?? undefined }))}
                title={`${b.name} results`}
                description={`${results.length} published results, ${b.direction.replace(/_/g, " ")}. Whiskers show the reported confidence interval where one exists.`}
                unit={b.unit}
                log={b.unit === "minutes" || b.unit === "hours"}
              />
              <details>
                <summary>Limitations and scaffolding</summary>
                <p>{b.limitations}</p>
                <dl className="kv">
                  <dt>Scaffold</dt>
                  <dd>{b.scaffold ?? "not stated"}</dd>
                  <dt>Model access</dt>
                  <dd>{b.model_access ?? "not stated"}</dd>
                  <dt>Weight note</dt>
                  <dd>{b.weight_note ?? "—"}</dd>
                  <dt>Sources</dt>
                  <dd>
                    {b.source_ids.map((s, i) => (
                      <span key={s}>
                        {i ? ", " : ""}
                        <Link href={`/evidence/sources/${s}`}>{s}</Link>
                      </span>
                    ))}
                  </dd>
                </dl>
              </details>
            </section>
          );
        })}
        <section className="stack">
          <h2>Driver signals (D1 capability, D4 compute)</h2>
          <SignalTable drivers={drivers} observations={snap.driver_observations} weights={snap.model_spec.index_weights.capability_pressure} />
        </section>
        <p className="cite">
          Index method: <Link href="/method/indexes">docs/method/indexes.md</Link>. Download: <Link href="/api/export/snapshot.json">snapshot.json</Link>.
        </p>
      </div>
    </div>
  );
}
