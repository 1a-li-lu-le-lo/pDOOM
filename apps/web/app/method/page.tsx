// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { getDataSource, getRelease, getSnapshot } from "@/lib/data";
import { titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Method",
  description: "The public methodology: what p(DOOM) computes, what it withholds, and every rule that governs a number on this site.",
};

const ORDER = ["model", "definitions-of-outputs", "source-hierarchy", "indexes", "aggregation", "disagreement", "uncertainty", "calibration", "sensitivity", "causal-model", "causal-graph", "backtesting", "content-safety"];

export default async function MethodPage() {
  const [docs, rel, snap] = await Promise.all([getDataSource().getMethodology(), getRelease(), getSnapshot()]);
  const sorted = [...docs].sort((a, b) => (ORDER.indexOf(a.slug) + 1 || 99) - (ORDER.indexOf(b.slug) + 1 || 99));
  const spec = snap.model_spec;
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader title="Method" lede="Nothing here is a black box. Every number traces to a snapshot, a model specification and a signed release. These documents are the contract." textAnchor="method" />
        <div className="grid grid-3">
          {sorted.map((d) => (
            <article key={d.slug} className="card">
              <h3 style={{ margin: 0 }}>
                <Link href={`/method/${d.slug}`}>{d.title}</Link>
              </h3>
              <p className="cite">{d.path}</p>
            </article>
          ))}
          <article className="card">
            <h3 style={{ margin: 0 }}>
              <Link href="/method/definitions">Definitions</Link>
            </h3>
            <p className="cite">{snap.definitions.length} terms with sourced definitions</p>
          </article>
        </div>
        <section className="card stack">
          <h2 style={{ marginTop: 0 }}>Model specification in force</h2>
          <dl className="kv">
            <dt>Specification</dt>
            <dd>{spec.id}</dd>
            <dt>Model versions</dt>
            <dd>{rel.manifest.model_versions.join(", ")}</dd>
            <dt>Weight bounds</dt>
            <dd>
              {spec.weight_bounds[0]} to {spec.weight_bounds[1]} per signal (no single signal dominates)
            </dd>
            <dt>Aggregation methods</dt>
            <dd>{spec.aggregation_methods.map(titleCase).join(", ")}</dd>
            <dt>Rounding</dt>
            <dd>
              low ±{spec.rounding_rules.low}, moderate ±{spec.rounding_rules.moderate}, high ±{spec.rounding_rules.high}, extreme ±{spec.rounding_rules.extreme} points
            </dd>
            <dt>Experimental causal model</dt>
            <dd>
              {spec.experimental_causal.version} · {spec.experimental_causal.samples.toLocaleString("en-US")} samples · seed {spec.experimental_causal.seed} · common-factor loading {spec.experimental_causal.common_factor_loading}
            </dd>
            <dt>Editorial rules</dt>
            <dd>
              <ul style={{ margin: 0 }}>
                {spec.editorial_rules.map((r) => (
                  <li key={r.level}>
                    <strong>{titleCase(r.level)}:</strong> {r.when}
                  </li>
                ))}
              </ul>
            </dd>
          </dl>
        </section>
      </div>
    </div>
  );
}
