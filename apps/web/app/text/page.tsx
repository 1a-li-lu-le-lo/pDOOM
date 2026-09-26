// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import type { ReactNode } from "react";
import { BRAND, BRAND_EXPANDED, OUTCOMES, PUBLIC_LABEL } from "@pdoom/schemas";
import { getDataSource, getRelease, getSnapshot, headline, indexById, researchEstimate } from "@/lib/data";
import { STATUS_LABEL, fmtDate, horizonLabel, outcomeSetLabel, tidyInterval, titleCase } from "@/lib/format";

export const metadata: Metadata = { title: "Plain text", description: `${BRAND} in plain text: every substantive fact of the observatory without graphics or scripts.` };

const SECTIONS = [
  ["estimate", "Current p(DOOM) estimate"],
  ["definition", "Definition"],
  ["horizon", "Selected horizon"],
  ["uncertainty", "Uncertainty"],
  ["decomposition", "Outcome decomposition"],
  ["upward", "Top upward drivers"],
  ["downward", "Top downward drivers"],
  ["changed", "What changed"],
  ["forecasts", "External forecasts"],
  ["capabilities", "Capability indicators"],
  ["agents", "Agentic infrastructure"],
  ["incidents", "Incidents"],
  ["scenarios", "Scenario map"],
  ["safeguards", "Safeguards"],
  ["act", "What users can do"],
  ["method", "Methodology"],
  ["sources", "Source ledger"],
  ["history", "Model history"],
  ["limitations", "Limitations"],
  ["governance", "Governance"],
] as const;

function H2({ id, children }: { id: string; children: ReactNode }) {
  return (
    <h2 id={id}>
      {children}
      <a className="anchor" href={`#${id}`} aria-label={`Link to this section`}>
        #
      </a>
    </h2>
  );
}

const pct = (p: number) => `${Math.round(p * 1000) / 10}%`;

export default async function TextPage() {
  const [rel, snap, releases, method] = await Promise.all([getRelease(), getSnapshot(), getDataSource().listReleases(), getDataSource().getMethodology()]);
  const h = headline(rel);
  const horizon = "10y";
  const official = h.official.find((e) => e.horizon === horizon);
  const research = researchEstimate(rel, "P_DOOM", horizon);
  const unc = indexById(rel, "uncertainty");
  const cite = (ids: readonly string[]) => (
    <span className="cite">
      {" "}
      [
      {ids.map((id, i) => (
        <span key={id}>
          {i ? ", " : ""}
          <a href={`#src-${id}`}>{id}</a>
        </span>
      ))}
      ]
    </span>
  );
  const up = rel.drivers_explained.items.filter((d) => d.direction === "raises_pressure").sort((a, b) => b.magnitude - a.magnitude).slice(0, 5);
  const down = rel.drivers_explained.items.filter((d) => d.direction === "strengthens_control").sort((a, b) => b.magnitude - a.magnitude).slice(0, 5);
  const groups = [...new Set(rel.aggregations.map((a) => a.group_id))];
  const agentSignals = snap.driver_observations.filter((o) => ["D2", "D3", "D6"].includes(o.family));
  const air = indexById(rel, "agentic_infrastructure_risk");
  const ipi = indexById(rel, "incident_pressure");

  return (
    <div className="page">
      <div className="container text-mode">
        <p className="eyebrow">{PUBLIC_LABEL} · plain-text mode</p>
        <h1>{BRAND} in plain text</h1>
        <p className="lede">
          {BRAND_EXPANDED}. This page contains every substantive fact on the site without graphics, scripts or hidden interactions. It is server-rendered, printable and linkable. Release {rel.manifest.release_id}, data through {fmtDate(rel.manifest.source_cutoff)}, published {fmtDate(rel.manifest.published)}. <Link href="/">Switch to the immersive version</Link>.
        </p>
        <nav aria-label="Sections">
          <ol className="toc">
            {SECTIONS.map(([id, label]) => (
              <li key={id}>
                <a href={`#${id}`}>{label}</a>
              </li>
            ))}
          </ol>
        </nav>

        <H2 id="estimate">1. Current p(DOOM) estimate</H2>
        <dl className="kv">
          <dt>p(DOOM) (official)</dt>
          <dd>
            <strong>{official?.display.central}</strong> — {official?.display.note}
          </dd>
          <dt>Horizon</dt>
          <dd>{horizonLabel(horizon)}</dd>
          <dt>Definition</dt>
          <dd>{outcomeSetLabel(["O3", "O4", "O5", "O6", "O7", "O8"])}</dd>
          <dt>Model</dt>
          <dd>{rel.manifest.model_versions.join(", ")}</dd>
          <dt>Data through</dt>
          <dd>{fmtDate(rel.manifest.source_cutoff)}</dd>
          <dt>Last reviewed</dt>
          <dd>
            {fmtDate(rel.manifest.published)} by {rel.manifest.reviewers.join(", ")}
          </dd>
          <dt>Editorial risk level</dt>
          <dd>{titleCase(rel.manifest.editorial_risk_level)} (rule-based classification, not a measurement)</dd>
        </dl>
        <p>Two separately labelled quantities are published beside the withheld official value. Neither is the official estimate and they are never blended.</p>
        <ul>
          {research ? (
            <li>
              <strong>{STATUS_LABEL[research.status]}</strong>, {outcomeSetLabel(research.outcome_set)}, {horizonLabel(research.horizon)}: median {research.display.central}, plausible interval {tidyInterval(research.display.interval)} (rounded to {research.rounding_rule.replace("nearest_", "nearest ")} points). {research.display.note}
            </li>
          ) : null}
          {h.external.map((e) => (
            <li key={e.estimate_id}>
              <strong>{STATUS_LABEL[e.status]}</strong> ({e.group_id}), {outcomeSetLabel(e.outcome_set)}, {horizonLabel(e.horizon)}: median {e.display.central}, member range {tidyInterval(e.display.interval)}, {e.disagreement} disagreement, {e.source_coverage.forecast_count} forecasts from {e.source_coverage.population_count} populations.
              {cite(e.source_coverage.source_ids)}
            </li>
          ))}
        </ul>

        <H2 id="definition">2. Definition</H2>
        <p>p(DOOM) is the probability, within a stated horizon and conditional on a stated model specification, that advanced AI leads to one of the outcomes O3–O8 below. It is a model output, not a measurement. Outcomes O0–O2 are tracked but excluded.</p>
        <table>
          <caption>Outcome taxonomy</caption>
          <thead>
            <tr>
              <th>Code</th>
              <th>Outcome</th>
              <th>In p(DOOM)</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            {Object.values(OUTCOMES).map((o) => (
              <tr key={o.code}>
                <td>{o.code}</td>
                <td>{o.label}</td>
                <td>{o.included_in.pdoom ? "yes" : "no"}</td>
                <td>{o.description}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <p>
          Formal definitions of {snap.definitions.length} terms (AGI, agent, MCP, alignment, control, incident, disempowerment and others) are on the <Link href="/method/definitions">definitions page</Link>; where no consensus exists, several attributed definitions are shown.
        </p>

        <H2 id="horizon">3. Selected horizon</H2>
        <p>
          The default horizon is {horizonLabel(horizon)} from the forecast origin date ({fmtDate(rel.manifest.generated_at)}). Estimates exist for 1, 3, 5, 10 and 25 years, by 2100, and eventually. Eventual probabilities have no endpoint and are never comparable with dated ones. No date for any catastrophe is implied anywhere on this site.
        </p>
        <table>
          <caption>Research-mode p(DOOM) by horizon (not the official estimate)</caption>
          <thead>
            <tr>
              <th>Horizon</th>
              <th className="num">Median</th>
              <th className="num">Interval (p05–p95)</th>
            </tr>
          </thead>
          <tbody>
            {["1y", "3y", "5y", "10y", "25y", "2100", "eventual"].map((hz) => {
              const e = researchEstimate(rel, "P_DOOM", hz);
              return e ? (
                <tr key={hz}>
                  <td>{horizonLabel(hz)}</td>
                  <td className="num">{e.display.central}</td>
                  <td className="num">{tidyInterval(e.display.interval)}</td>
                </tr>
              ) : null;
            })}
          </tbody>
        </table>

        <H2 id="uncertainty">4. Uncertainty</H2>
        <p>
          Uncertainty score: <strong>{unc?.value !== null && unc?.value !== undefined ? Math.round(unc.value) : "—"} / 100</strong> ({research?.uncertainty}). {unc?.note}
        </p>
        <table>
          <caption>Uncertainty components (each 0–1) and weights</caption>
          <thead>
            <tr>
              <th>Component</th>
              <th className="num">Value</th>
              <th className="num">Weight</th>
            </tr>
          </thead>
          <tbody>
            {(unc?.components ?? []).map((c) => (
              <tr key={c.signal_id}>
                <td>{titleCase(c.signal_id)}</td>
                <td className="num">{c.value_normalized}</td>
                <td className="num">{c.weight}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <H2 id="decomposition">5. Outcome decomposition</H2>
        <table>
          <caption>Research-mode estimates by outcome set, {horizonLabel(horizon)}</caption>
          <thead>
            <tr>
              <th>Outcome set</th>
              <th className="num">Median</th>
              <th className="num">Interval</th>
            </tr>
          </thead>
          <tbody>
            {["P_DOOM", "O3", "P_COLLAPSE", "O6", "O7"].map((k) => {
              const e = researchEstimate(rel, k, horizon);
              return e ? (
                <tr key={k}>
                  <td>{outcomeSetLabel(e.outcome_set)}</td>
                  <td className="num">{e.display.central}</td>
                  <td className="num">{tidyInterval(e.display.interval)}</td>
                </tr>
              ) : null;
            })}
          </tbody>
        </table>

        <H2 id="upward">6. Top upward drivers</H2>
        <ol>
          {up.map((d) => (
            <li key={d.driver + d.index_id}>
              {d.driver}: +{Math.round(d.magnitude * 10) / 10} points on {titleCase(d.index_id)} (confidence {d.confidence}; removing it would move the index by {Math.round(d.sensitivity * 10) / 10}; last updated {d.last_updated}).
              {cite(d.source_ids)}
              {d.counterevidence ? ` Counterevidence: ${d.counterevidence}` : ""}
            </li>
          ))}
        </ol>
        <H2 id="downward">7. Top downward drivers</H2>
        <ol>
          {down.map((d) => (
            <li key={d.driver + d.index_id}>
              {d.driver}: +{Math.round(d.magnitude * 10) / 10} points on control strength (confidence {d.confidence}; last updated {d.last_updated}).
              {cite(d.source_ids)}
              {d.counterevidence ? ` Counterevidence: ${d.counterevidence}` : ""}
            </li>
          ))}
        </ol>
        <p>{rel.drivers_explained.notes.join(" ")}</p>

        <H2 id="changed">8. What changed</H2>
        <ul>
          {rel.manifest.changes.map((c) => (
            <li key={c}>{c}</li>
          ))}
        </ul>
        <p>Heightened-review triggers for this release: {rel.delta.heightened_review_triggers.join(", ") || "none"}.</p>
        <table>
          <caption>Indexes (0–100; none is a probability)</caption>
          <thead>
            <tr>
              <th>Index</th>
              <th className="num">Value</th>
              <th className="num">Coverage</th>
              <th>Scale</th>
            </tr>
          </thead>
          <tbody>
            {rel.indexes.map((i) => (
              <tr key={i.index_id}>
                <td>{i.label}</td>
                <td className="num">{i.value === null ? "not computed" : Math.round(i.value)}</td>
                <td className="num">{Math.round(i.coverage * 100)} percent</td>
                <td>{i.scale}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <H2 id="forecasts">9. External forecasts</H2>
        <p>Forecasts are aggregated only inside compatibility groups (same outcome set, horizon and conditioning). Original wording is shown for every record.</p>
        {groups.map((g) => {
          const aggs = rel.aggregations.filter((a) => a.group_id === g);
          const members = snap.forecasts.filter((f) => f.group_id === g);
          const first = aggs[0];
          return (
            <div key={g}>
              <h3>
                {g}: {first ? outcomeSetLabel(first.outcome_set) : ""}, {first ? horizonLabel(first.horizon) : ""}
              </h3>
              <table>
                <caption>Members of {g}</caption>
                <thead>
                  <tr>
                    <th>Forecast</th>
                    <th>Population</th>
                    <th>Date</th>
                    <th className="num">Median</th>
                    <th>Question wording</th>
                  </tr>
                </thead>
                <tbody>
                  {members.map((f) => (
                    <tr key={f.id}>
                      <td>
                        {f.forecaster_or_survey}
                        {cite([f.source_id])}
                      </td>
                      <td>
                        {titleCase(f.population)}
                        {f.sample_size ? ` (n=${f.sample_size})` : ""}
                      </td>
                      <td>{f.date}</td>
                      <td className="num">{f.median !== null ? pct(f.median) : f.mean !== null ? `${pct(f.mean)} (mean)` : "—"}</td>
                      <td>
                        {f.paraphrase ? "(paraphrase) " : ""}
                        {f.question_wording_original}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <table>
                <caption>Aggregation methods for {g}</caption>
                <thead>
                  <tr>
                    <th>Method</th>
                    <th className="num">Value</th>
                    <th>Preferred</th>
                  </tr>
                </thead>
                <tbody>
                  {aggs.map((a) => (
                    <tr key={a.aggregation_id}>
                      <td>{titleCase(a.method)}</td>
                      <td className="num">{pct(a.value)}</td>
                      <td>{a.preferred ? "yes" : ""}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          );
        })}
        <h3>Ungrouped records</h3>
        <ul>
          {snap.forecasts
            .filter((f) => !f.group_id)
            .map((f) => (
              <li key={f.id}>
                {f.forecaster_or_survey} ({f.date}, {titleCase(f.population)}): {f.median !== null ? pct(f.median) : "—"} — “{f.question_wording_original}” — {outcomeSetLabel(f.outcome_set)}, {f.horizon === "custom" ? f.horizon_note : horizonLabel(f.horizon)}; {f.conditions}.{f.transformation_note ? ` ${f.transformation_note}` : ""}
                {cite([f.source_id])}
              </li>
            ))}
        </ul>

        <H2 id="capabilities">10. Capability indicators</H2>
        {snap.benchmarks.map((b) => (
          <div key={b.id}>
            <h3>{b.name}</h3>
            <p>
              Maintainer {b.maintainer}; unit {b.unit}; contamination risk {b.contamination_risk}; saturation {b.saturation}; p(DOOM) relevance {b.pdoom_relevance}. {b.limitations}
              {cite(b.source_ids)}
            </p>
            <table>
              <caption>Results for {b.name}</caption>
              <thead>
                <tr>
                  <th>Model</th>
                  <th>Developer</th>
                  <th>Date</th>
                  <th className="num">Value</th>
                  <th>Confidence</th>
                  <th>Note</th>
                </tr>
              </thead>
              <tbody>
                {snap.benchmark_results
                  .filter((r) => r.benchmark_id === b.id)
                  .map((r) => (
                    <tr key={r.id}>
                      <td>{r.model_name}</td>
                      <td>{r.model_developer}</td>
                      <td>{r.date}</td>
                      <td className="num">
                        {r.value} {r.unit}
                        {r.ci_low !== null && r.ci_high !== null ? ` (${r.ci_low}–${r.ci_high})` : ""}
                      </td>
                      <td>{r.confidence}</td>
                      <td>
                        {r.note}
                        {cite(r.source_ids)}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        ))}
        <p>Benchmark progress is a prerequisite or pressure variable, never an outcome; it is not converted into a probability.</p>

        <H2 id="agents">11. Agentic infrastructure</H2>
        <p>
          Agentic infrastructure risk index: {air?.value === null || air?.value === undefined ? "not computed" : `${Math.round(air.value)} / 100`}. Signals under autonomy (D2), access and exposure (D3) and security (D6):
        </p>
        <table>
          <caption>Agent-related driver observations</caption>
          <thead>
            <tr>
              <th>Signal</th>
              <th className="num">Value (0–1)</th>
              <th>Kind</th>
              <th>As of</th>
              <th>Rationale</th>
            </tr>
          </thead>
          <tbody>
            {agentSignals.map((o) => (
              <tr key={o.id}>
                <td>{o.signal_id}</td>
                <td className="num">{o.value_normalized}</td>
                <td>{o.observation_kind}</td>
                <td>{o.as_of}</td>
                <td>
                  {o.rationale}
                  {cite(o.source_ids)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <H2 id="incidents">12. Incidents</H2>
        <p>
          Incident pressure index: {ipi?.value === null || ipi?.value === undefined ? "not computed" : `${Math.round(ipi.value)} / 100`}. {ipi?.note}
        </p>
        <table>
          <caption>Verified incidents and near misses (non-graphic, non-operational summaries)</caption>
          <thead>
            <tr>
              <th>Date</th>
              <th>Incident</th>
              <th>Severity</th>
              <th>Relevance</th>
              <th>Evidence</th>
              <th>Near miss</th>
            </tr>
          </thead>
          <tbody>
            {snap.incidents.map((i) => (
              <tr key={i.id}>
                <td>{i.date ?? "unknown"}</td>
                <td>
                  <strong>{i.title}</strong>. {i.summary}
                  {cite(i.source_ids)}
                </td>
                <td>{i.severity}</td>
                <td>{titleCase(i.pdoom_relevance)}</td>
                <td>{titleCase(i.evidence_level)}</td>
                <td>{i.near_miss ? "yes" : "no"}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <H2 id="scenarios">13. Scenario map</H2>
        <p>Eighteen category-level pathways. No pathway probability is assigned in this release; overlaps are recorded as edges and never summed.</p>
        <table>
          <caption>Scenarios S1–S18</caption>
          <thead>
            <tr>
              <th>Id</th>
              <th>Pathway</th>
              <th>Outcomes</th>
              <th>Recoverability</th>
              <th>Early indicators</th>
            </tr>
          </thead>
          <tbody>
            {snap.scenarios.map((s) => (
              <tr key={s.id}>
                <td>{s.id}</td>
                <td>
                  <Link href={`/futures/${s.id}`}>{s.name}</Link>. {s.description}
                  {cite(s.source_ids)}
                </td>
                <td>{s.outcome_set.join(", ")}</td>
                <td>{s.recoverability}</td>
                <td>{s.early_indicators.join("; ")}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <p>Dependencies ({snap.scenario_edges.length} edges): {snap.scenario_edges.map((e) => `${e.from_id} ${e.relation.replace(/_/g, " ")} ${e.to_id} (${e.confidence})`).join("; ")}.</p>

        <H2 id="safeguards">14. Safeguards</H2>
        <table>
          <caption>Interventions I01–I27 (effect sizes are qualitative only)</caption>
          <thead>
            <tr>
              <th>Id</th>
              <th>Intervention</th>
              <th>Category</th>
              <th>Evidence</th>
              <th>Cost</th>
              <th>Time</th>
              <th>Targets</th>
            </tr>
          </thead>
          <tbody>
            {snap.interventions.map((i) => (
              <tr key={i.id}>
                <td>{i.id}</td>
                <td>
                  <strong>{i.name}</strong>. {i.mechanism} {i.evidence_summary}
                  {cite(i.source_ids)}
                </td>
                <td>{i.category}</td>
                <td>{i.evidence_strength}</td>
                <td>{i.cost}</td>
                <td>{i.time_to_deploy}</td>
                <td>{i.target_scenario_ids.join(", ")}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <H2 id="act">15. What users can do</H2>
        {[...new Set(snap.actions.map((a) => a.audience))].map((aud) => (
          <div key={aud}>
            <h3>{titleCase(aud)}</h3>
            <ul>
              {snap.actions
                .filter((a) => a.audience === aud)
                .map((a) => (
                  <li key={a.id}>
                    <strong>{a.title}.</strong> {a.description}
                    {a.resources.length ? (
                      <>
                        {" "}
                        Resources:{" "}
                        {a.resources.map((r, i) => (
                          <span key={r.url}>
                            {i ? ", " : ""}
                            <a href={r.url} rel="noopener">
                              {r.title}
                            </a>
                          </span>
                        ))}
                        .
                      </>
                    ) : null}
                  </li>
                ))}
            </ul>
          </div>
        ))}
        <p>
          Organisations meeting the directory&apos;s inclusion criteria are listed on the <Link href="/act#organizations">Act page</Link>: {snap.organizations.map((o) => o.name).join(", ")}.
        </p>

        <H2 id="method">16. Methodology</H2>
        <p>
          p(DOOM) keeps eight outputs separate: the external forecast aggregate, the model estimate (official object withheld; research-mode decomposition published), four 0–100 indexes, an uncertainty score and an editorial level. The research-mode model is P(O_i ≤ T) = P(A ≤ T)·P(C|A)·P(E|A,C)·P(F|A,C,E)·P(O_i|A,C,E,F) with logit-normal factors fitted to documented judgments and one common latent factor; Monte Carlo with seed {snap.model_spec.experimental_causal.seed} and {snap.model_spec.experimental_causal.samples} samples. Indexes are weighted means of normalised, tier-weighted observations. Aggregation uses seven methods with the unweighted median preferred.
        </p>
        <ul>
          {method.map((m) => (
            <li key={m.slug}>
              <Link href={`/method/${m.slug}`}>{m.title}</Link>
            </li>
          ))}
        </ul>

        <H2 id="sources">17. Source ledger</H2>
        <table>
          <caption>{snap.sources.length} sources with tier, verification and model-use status</caption>
          <thead>
            <tr>
              <th>Id</th>
              <th>Source</th>
              <th>Tier</th>
              <th>Verification</th>
              <th>Model use</th>
              <th>Conflicts</th>
            </tr>
          </thead>
          <tbody>
            {snap.sources.map((s) => (
              <tr key={s.id} id={`src-${s.id}`}>
                <td>{s.id}</td>
                <td>
                  <a href={s.canonical_url} rel="noopener">
                    {s.title}
                  </a>
                  , {s.publisher}, {s.date_published ?? "n.d."}. {s.evidence_summary}
                </td>
                <td>{s.source_tier}</td>
                <td>{s.verification.status}</td>
                <td>{s.model_use_status}</td>
                <td>{s.conflicts.join(", ")}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <H2 id="history">18. Model history</H2>
        <table>
          <caption>Releases</caption>
          <thead>
            <tr>
              <th>Release</th>
              <th>Snapshot</th>
              <th>Published</th>
              <th>Editorial level</th>
              <th>Uncertainty</th>
            </tr>
          </thead>
          <tbody>
            {releases.map((r) => (
              <tr key={r.release_id}>
                <td>
                  <Link href={`/releases/${r.release_id}`}>{r.release_id}</Link>
                  {r.is_current ? " (current)" : ""}
                </td>
                <td>{r.data_snapshot}</td>
                <td>{fmtDate(r.published)}</td>
                <td>{titleCase(r.editorial_risk_level)}</td>
                <td>{r.uncertainty_score === null ? "—" : Math.round(r.uncertainty_score)}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <p>
          Reproduction command: <code>{rel.manifest.reproduction_command}</code>. Manifest hash {rel.manifest.signature}. Approvals: {rel.approvals.map((a) => `${a.reviewer_id} (${fmtDate(a.signed_at)})`).join(", ")}.
        </p>

        <H2 id="limitations">19. Limitations</H2>
        <ul>
          {rel.manifest.known_limitations.map((l) => (
            <li key={l}>{l}</li>
          ))}
          <li>{snap.manifest.notes}</li>
        </ul>

        <H2 id="governance">20. Governance</H2>
        <p>
          No automated update reaches production. Candidates are computed deterministically from a sealed snapshot, checked against invariants, signed by reviewers with ed25519 keys, re-run for byte-for-byte reproducibility at promotion, and recorded in a hash-chained audit log. Heightened-review triggers require a second reviewer. Releases are immutable; rollback moves only the current pointer. See the <Link href="/method">methodology</Link>, the <Link href="/changelog">changelog</Link> and the release <Link href={`/releases/${rel.manifest.release_id}`}>manifest</Link>. Authored by NU Cybernetics; corrections and appeals are described in the governance documents.
        </p>
        <p>
          Downloads: <a href="/api/export/release.json">release.json</a>, <a href="/api/export/estimates.csv">estimates.csv</a>, <a href="/api/export/sources.csv">sources.csv</a>, <a href="/api/export/forecasts.csv">forecasts.csv</a>, <a href="/api/export/incidents.csv">incidents.csv</a>, <a href="/api/export/snapshot.json">snapshot.json</a>.
        </p>
      </div>
    </div>
  );
}
