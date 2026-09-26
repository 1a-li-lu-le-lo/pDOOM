// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { NodeGraph } from "@/components/charts/NodeGraph";
import { getSnapshot } from "@/lib/data";
import { outcomeSetLabel, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Branching futures",
  description: "Category-level pathway scenarios and the relations between them, each with prerequisites, early indicators, counterindicators and the safeguards that target it.",
};

const RECOVER_COLOR: Record<string, string> = {
  high: "var(--c-resilience)",
  moderate: "var(--c-resilience)",
  low: "var(--c-uncertainty)",
  none: "var(--c-risk)",
  unknown: "var(--c-insufficient)",
};
const RECOVER_LABEL: Record<string, string> = {
  high: "High recoverability",
  moderate: "Moderate recoverability",
  low: "Low recoverability",
  none: "Irreversible",
  unknown: "Recoverability unknown",
};

export default async function FuturesPage() {
  const snap = await getSnapshot();
  const scenarios = [...snap.scenarios].sort((a, b) => Number(a.id.slice(1)) - Number(b.id.slice(1)));
  const groups = ["none", "low", "moderate", "high", "unknown"];
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="Branching futures"
          lede="Eighteen category-level pathways, described by their prerequisites and what would be visible early, never by operational detail. Arrows show which pathways enable, amplify or substitute for others."
          textAnchor="scenarios"
        />
        <NodeGraph
          title="Scenario relations"
          description="Nodes are pathways coloured by recoverability (red irreversible, amber low, green moderate or high, grey unknown). Arrow thickness follows the stated confidence in the relation. This is a map of dependencies, not a prediction of any route."
          nodes={scenarios.map((s) => ({ id: s.id, label: s.name, group: s.recoverability, href: `/futures/${s.id}`, colorVar: RECOVER_COLOR[s.recoverability] }))}
          edges={snap.scenario_edges.map((e) => ({ from: e.from_id, to: e.to_id, label: titleCase(e.relation), strength: e.confidence === "high" ? 1 : e.confidence === "moderate" ? 0.6 : 0.3 }))}
        />
        {groups.map((g) => {
          const items = scenarios.filter((s) => s.recoverability === g);
          if (!items.length) return null;
          return (
            <section key={g} aria-labelledby={`grp-${g}`}>
              <h2 id={`grp-${g}`} className="row">
                <span className="badge" style={{ color: RECOVER_COLOR[g] }}>
                  {RECOVER_LABEL[g]}
                </span>
                <span className="muted" style={{ fontSize: "var(--fs-md)", fontWeight: 400 }}>
                  {items.length} pathways
                </span>
              </h2>
              <div className="grid grid-3">
                {items.map((s) => (
                  <article key={s.id} className="card scenario-card">
                    <div className="row">
                      <span className="badge badge-uncertainty">{s.uncertainty} uncertainty</span>
                      <span className="badge">{s.probability_source.replace(/_/g, " ")}</span>
                    </div>
                    <h3>
                      <Link href={`/futures/${s.id}`}>
                        {s.id} · {s.name}
                      </Link>
                    </h3>
                    <p className="muted">{s.description}</p>
                    <p className="cite">Outcomes: {outcomeSetLabel(s.outcome_set)}</p>
                    <p className="cite">
                      {s.intervention_ids.length} safeguards target this pathway · {s.early_indicators.length} early indicators
                    </p>
                  </article>
                ))}
              </div>
            </section>
          );
        })}
        <p className="cite">
          Method: <Link href="/method/causal-graph">docs/method/causal-graph.md</Link>. Content rules: <Link href="/method/content-safety">docs/method/content-safety.md</Link>.
        </p>
      </div>
    </div>
  );
}
