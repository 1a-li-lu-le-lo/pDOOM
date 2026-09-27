// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { PageHeader } from "@/components/shell/PageHeader";
import { NodeGraph } from "@/components/charts/NodeGraph";
import { FlowDiagram } from "@/components/charts/FlowDiagram";
import { getSnapshot } from "@/lib/data";
import { outcomeLabel, outcomeSetLabel, titleCase } from "@/lib/format";

export const metadata: Metadata = {
  title: "Branching futures",
  description: "Category-level pathway scenarios and the relations between them, each with prerequisites, early indicators, counterindicators and the safeguards that target it.",
};

/** Count-aware label so single items never read as "1 pathways". */
const plural = (n: number, singular: string, pluralForm = `${singular}s`) => `${n} ${n === 1 ? singular : pluralForm}`;

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
  const outcomes = ["O3", "O4", "O5", "O6", "O7", "O8"];
  const flows = groups.flatMap((g) => outcomes.map((o) => ({ from: g, to: o, value: scenarios.filter((s) => s.recoverability === g && s.outcome_set.includes(o as never)).length }))).filter((f) => f.value > 0);
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader
          title="Branching futures"
          lede={`${scenarios.length} category-level pathways, described by their prerequisites and what would be visible early, never by operational detail. Arrows show which pathways enable, amplify or substitute for others.`}
          textAnchor="scenarios"
        />
        <NodeGraph
          title="Scenario relations"
          description="Nodes are pathways coloured by recoverability (red irreversible, amber low, green moderate or high, grey unknown). Arrow thickness follows the stated confidence in the relation. This is a map of dependencies, not a prediction of any route."
          nodes={scenarios.map((s) => ({ id: s.id, label: s.name, group: s.recoverability, href: `/futures/${s.id}`, colorVar: RECOVER_COLOR[s.recoverability] }))}
          edges={snap.scenario_edges.map((e) => ({ from: e.from_id, to: e.to_id, label: titleCase(e.relation), strength: e.confidence === "high" ? 1 : e.confidence === "moderate" ? 0.6 : 0.3 }))}
        />
        <FlowDiagram
          title="Which pathways can reach which outcomes"
          description="Ribbons run from recoverability groups of pathways to the outcomes those pathways can reach; ribbon thickness is the number of pathways, never a probability. Most pathways can reach several outcomes, which is why the outcome decomposition on the meter is shown together."
          left={groups.filter((g) => scenarios.some((s) => s.recoverability === g)).map((g) => ({ id: g, label: RECOVER_LABEL[g] ?? g, colorVar: RECOVER_COLOR[g] }))}
          right={outcomes.map((o) => ({ id: o, label: `${o} ${outcomeLabel(o)}`, colorVar: "var(--c-risk)" }))}
          flows={flows}
          unit="pathways"
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
                  {plural(items.length, "pathway")}
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
                      {plural(s.intervention_ids.length, "safeguard")} {s.intervention_ids.length === 1 ? "targets" : "target"} this pathway · {plural(s.early_indicators.length, "early indicator")}
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
