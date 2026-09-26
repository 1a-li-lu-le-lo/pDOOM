<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Scenarios: research report

This report justifies the 18 scenario records (`S1`–`S18`) and 30 scenario edges in
`data/snapshots/snap-2026-09-26-001/scenarios.json` and `scenario_edges.json`, authored in
`tools/snapshot/content/scenarios.mjs`. Scenarios are category-level pathways; no pathway
probability is assigned in this release (`probability_source: not_assigned` on every record),
and the graph is descriptive, not simulated
([`../../docs/method/causal-graph.md`](../../docs/method/causal-graph.md)).

## 1. The pathways

| Id | Name | Outcome set | Recoverability | Depends on | Safeguards targeting it | Extra sources beyond the base set |
| --- | --- | --- | --- | --- | --- | --- |
| S1 | Deliberate misaligned action | O3, O4, O5, O6, O8 | low | — | I01 I06 I07 I09 I10 I13 I14 I15 I16 I19 | alignment faking (2024), agentic misalignment (2025), o3 shutdown tests, AISI 2026 incident, Carlsmith 2022 |
| S2 | Specification failure | O1, O2, O3, O4, O7, O8 | moderate | — | I01 I08 I09 I15 I20 I27 | — |
| S3 | Control loss | O3, O4, O5, O6, O8 | low | S1, S13 | I06 I07 I08 I09 I10 I11 I12 I16 I21 | AISI 2026 incident, AIID 1152, NSA MCP guidance |
| S4 | Malicious human use | O1, O2, O4, O5, O6 | moderate | — | I01 I04 I05 I17 I20 I21 | Anthropic 2025 espionage disclosure |
| S5 | Cyber-physical cascade | O1, O4 | moderate | S4, S12 | I06 I07 I21 I27 | Anthropic 2025 espionage disclosure |
| S6 | Biological or chemical enablement | O4, O5, O6, O7 | low | S4 | I01 I04 I17 I20 I21 I27 | OpenAI Preparedness v2, Anthropic RSP |
| S7 | Military escalation | O1, O4, O5, O6 | low | S4, S5 | I17 I21 I26 | — |
| S8 | Autonomous economic capture | O2, O3, O8 | moderate | S16 | I09 I17 I22 I27 | gradual disempowerment (2025) |
| S9 | Authoritarian lock-in | O2, O3, O8 | low | S8, S10 | I17 I18 I23 I25 | — |
| S10 | Epistemic collapse | O1, O2, O3, O8 | moderate | — | I17 I25 I27 | AIID 628, DeepMind FSF v3 |
| S11 | Ecological optimization failure | O7, O1 | low | S2 | I08 I17 I27 | — |
| S12 | Infrastructure monoculture | O1, O4 | high | — | I11 I12 I27 | CSA MCP security crisis (2026) |
| S13 | Agent swarm emergence | O1, O3, O4 | moderate | — | I08 I10 I11 I16 | CSA analysis of the AISI incident |
| S14 | Self-accelerating AI research | O3, O4, O5, O6, O8 | low | S15 | I01 I02 I05 I17 I20 | METR time-horizons page |
| S15 | Governance race failure | O1, O2, O3, O4, O5, O6 | moderate | — | I02 I03 I17 I18 I22 I23 | EU GPAI guidelines, California SB 53, New York RAISE Act, Korea AI Basic Act |
| S16 | Slow disempowerment | O3, O8 | low | — | I17 I23 I26 I27 | gradual disempowerment (2025) |
| S17 | Benign guardianship becomes permanent | O3, O8 | none | — | I14 I15 I17 | Bostrom 2014 |
| S18 | Unknown unknown | O8 | unknown | — | I26 I27 | — |

Every scenario cites the base set of four taxonomy sources: `src-hendrycks-2023-overview-catastrophic-risks`,
`src-critch-russell-2023-tasra`, `src-iasr-2026`, `src-slattery-2024-ai-risk-repository`.
Recoverability distribution: low 8, moderate 7, high 1, none 1, unknown 1. Every record is
`uncertainty: extreme`, `verified_search`, `informational` (scenarios are never model inputs)
and `human_review_status: pending`.

Each record states prerequisites, early indicators, counterindicators, capability thresholds,
exposure, control failures, human and AI contributions, a time-horizon note that implies no
date, an evidence summary and open questions. The `time_horizon_note` fields are written to
avoid dates ("no date is implied", "ongoing", "gradual", "long-run").

## 2. Content safety

S6 (biological or chemical enablement), S7 (military escalation) and S11 (ecological
optimisation failure) carry a `content_safety_note` stating that the record omits agents,
procedures, targets, mitigation-evasion, operational, targeting and attack-mechanism detail.
S4 and S5 are written at the same level without a note. Indicators everywhere are observable,
non-operational events (evaluation thresholds crossed under published frameworks, disclosed
campaigns, registry incidents, policy milestones). This follows
[`../../docs/method/content-safety.md`](../../docs/method/content-safety.md).

## 3. The edges

30 directed edges, each with a relation, a confidence and a rationale:

| Relation | Count | Examples |
| --- | --- | --- |
| `enables` | 14 | S4→S6, S3→S1, S8→S9, S16→S3, S4→S5, S2→S11, S1→S8 |
| `amplifies` | 11 | S15→S3, S15→S14, S14→S15, S1→S3, S8→S16, S12→S5, S13→S3 |
| `prevents_response` | 2 | S10→S15, S10→S3 |
| `shares_prerequisite` | 2 | S2→S3, S18→S3 |
| `competes_with` | 1 | S9→S17 |

Confidence: high 8, moderate 16, low 6. Edges default to `src-hendrycks-2023-overview-catastrophic-risks`
as their source; seven edges cite a specific record instead (AIID 628 and DeepMind FSF v3 for
S10→S9; alignment faking for S1→S3; the CSA AISI analysis for S13→S3; the CSA MCP crisis note for
S12→S5; the espionage disclosure for S4→S5; the mcp-remote CVE for S12→S3). The Futures page draws
the edges as a node graph labelled "a map of dependencies, not a prediction of any route", and
the relations table is the authoritative form.

The edges are why pathway probabilities are not summed: overlapping prerequisites (autonomy and
access, S2→S3), amplification loops (S14↔S15, S1↔S3) and competing endpoints (S9 vs S17) would
double-count or ignore competing risks.

## 4. How the scenarios relate to the model

- The experimental causal model compresses the common-cause nodes of the graph (race dynamics
  S15, governance capacity, epistemic collapse S10) into one latent factor; the graph itself is
  not simulated.
- The Scenario Lab's dials are factor shifts, not scenario selections.
- Interventions reference scenarios through `target_scenario_ids`, and scenarios reference
  interventions through `intervention_ids`; the two lists are unioned on the scenario page.

## 5. Research fragments not yet merged

`research/scenarios/fragments/sources.json` (43 sources: tier 1 ×5, tier 2 ×33, tier 4 ×5;
39 `verified_search` and 4 `verified_fetch`; 37 `eligible`, 6 `informational`) and
`claims.json` (22 claims) were authored for this area with their own identifiers. Only 7 of the
43 match a snapshot source by id and 15 by canonical URL; the rest (for example
`src-rand-2018-geist-lohn-ai-nuclear`, `src-rivera-2024-escalation-risks-llm-wargames`,
`src-hammond-2025-multi-agent-risks`, `src-kleinberg-raghavan-2021-algorithmic-monoculture`,
`src-feldstein-2019-global-expansion-ai-surveillance`, `src-seger-2020-epistemic-security`,
`src-meinke-2024-in-context-scheming`, `src-korinek-suh-2024-scenarios-transition-agi`) are
not in `snap-2026-09-26-001`. Several fragment ids name the same document as a content-module
id under a different slug (`src-hendrycks-2023-overview-catastrophic-ai-risks` versus
`src-hendrycks-2023-overview-catastrophic-risks`; `src-ord-2020-the-precipice` versus
`src-ord-2020-precipice`; `src-greenblatt-2024-alignment-faking` versus
`src-anthropic-2024-alignment-faking`), which `merge-fragments.mjs` deduplicates by canonical
URL and remaps. The four `verified_fetch` fragment sources (Anthropic pages) would be the first
fetch-verified records in the snapshot if merged. Merging is a snapshot change and therefore a
new release.

## Limitations

- Every scenario is `informational` and `extreme` uncertainty; the taxonomy is an editorial
  synthesis of four published taxonomies and the cited incident and evaluation records, not a
  measured structure.
- Edge confidences are judgements; only seven edges cite record-level evidence.
- S17 and S18 rest on philosophical or definitional grounds and have no empirical instances.
- The scenario fragments with per-scenario sources are unmerged, so scenario pages cite the base
  set plus a few records rather than the fuller bibliography.
- No scenario has been human-reviewed.

## Open questions

- Should pathway-specific evidence summaries be strengthened by merging the fragments before
  the second release, given the id remapping that entails?
- Is `recoverability: high` for S12 (infrastructure monoculture) defensible when the outcome set
  includes O4?
- Which early indicators can be turned into tracked driver signals so that the scenario map and
  the indexes share observables (for example agent-to-agent recruitment for S13, release cadence
  for S15)?
- How should residual probability for S18 be represented, if at all, in a future model version?
