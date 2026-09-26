<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Capabilities: research report

This report justifies the capability evidence in `snap-2026-09-26-001`: 3 benchmarks
(`benchmarks.json`), 7 benchmark results (`benchmark_results.json`), and the driver
observations under D1 (capability), D2 (autonomy) and D4 (scalability) in
`driver_observations.json`, all authored in `tools/snapshot/content/forecasts.mjs` and
`drivers.mjs`. Capability is a prerequisite or pressure variable in this observatory, never an
outcome; nothing here is converted into a probability.

## 1. Benchmarks

| Id | Name | Maintainer | Unit | Contamination | Saturation | Relevance | Sources | Feeds |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `bm-metr-time-horizon-50` | METR 50% task-completion time horizon (Time Horizon 1.1, January 2026) | METR | minutes | moderate | none | strong | `src-metr-2025-measuring-long-tasks`, `src-metr-time-horizons-page`, `src-metr-2026-time-horizon-limitations` | `D1.task_horizon_50pct`, `D2.unattended_operation_hours` |
| `bm-swe-bench-verified` | SWE-bench Verified (500 tasks) | SWE-bench maintainers (with OpenAI for the subset) | percent resolved | high | partial | moderate | `src-swebench-official`, `src-steel-2026-swe-bench-verified-leaderboard` | `D1.swe_bench_verified` at reduced confidence |
| `bm-osworld-verified` | OSWorld-Verified | OSWorld maintainers | percent tasks succeeded | moderate | partial | moderate | `src-osworld-official`, `src-benchlm-2026-osworld-verified` | `D1.computer_use_osworld` |

Each benchmark record states its `limitations` (wide confidence intervals and a software-centric
suite for the time horizon; near-saturation and flawed tests for SWE-bench Verified; fixed task
suite and heterogeneous scaffolds for OSWorld). All three are `verified_search` and `eligible`.

## 2. Benchmark results

| Id | Model (developer) | Date | Value as recorded | Confidence | Verification / model use | Note |
| --- | --- | --- | --- | --- | --- | --- |
| `bmr-metr-th50-claude-3-7-sonnet` | Claude 3.7 Sonnet (Anthropic) | 2025-03-19 | 50 minutes | moderate | search / eligible | around 50 minutes in the March 2025 paper |
| `bmr-metr-th50-o3` | o3 (OpenAI) | 2025-04-16 (approx.) | 110 minutes | moderate | search / eligible | METR living page |
| `bmr-metr-th50-gpt-5-1-codex-max` | GPT-5.1-Codex-Max (OpenAI) | 2025-11-20 (approx.) | 173 minutes | moderate | search / eligible | 2 h 53 min |
| `bmr-metr-th50-claude-opus-4-5` | Claude Opus 4.5 (Anthropic) | 2025-12-18 (approx.) | 289 minutes, CI 109–1225 | moderate | search / eligible | 4 h 49 min; the highest METR-published horizon in the snapshot |
| `bmr-metr-th50-claude-opus-4-6-third-party` | Claude Opus 4.6 (Anthropic) | 2026-03-01 (month) | 719 minutes | low | prior knowledge / informational | third-party EA Forum estimate, not a METR publication; excluded from the model |
| `bmr-swe-bench-verified-claude-opus-5-2026-09` | Claude Opus 5 (Anthropic) | 2026-09-04 | 97.0 percent resolved (Vals.ai run) | low | search / eligible | tracker-reported; treated as a saturation signal |
| `bmr-osworld-verified-qwen3-8-max-2026-09` | Qwen3.8 Max (Alibaba) | 2026-09-22 | 86.1 percent tasks succeeded | low | search / eligible | tracker-reported; above the 72.36 percent human baseline of the original paper |

The Capabilities page draws the METR results on a `Timeline` with the Opus 4.5 confidence
whiskers and a log axis when the range requires it; the third-party Opus 4.6 estimate is shown
as informational and is not an input.

## 3. Driver observations (D1, D2, D4)

| Observation | Signal | Value (0–1) | Raw | Kind | Confidence | As of | Rationale (abridged) | Counterevidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `do-d1-task-horizon-50pct` | D1.task_horizon_50pct | 0.566 | 4.817 h | observation | 0.6 | 2025-12-18 | log2(4.817/0.05)/log2(160/0.05) from the Opus 4.5 horizon | third-party Opus 4.6 estimate (informational); METR's limitations note |
| `do-d1-swe-bench-verified` | D1.swe_bench_verified | 0.97 | 97.0 percent | observation | 0.4 | 2026-09-04 | score/100; mostly indicates saturation | contamination, scaffold variance, tier 3 source |
| `do-d1-computer-use-osworld` | D1.computer_use_osworld | 0.861 | 86.1 percent | observation | 0.4 | 2026-09-22 | score/100 | third-party tracker; step budgets vary |
| `do-d1-cyber-offense-capability` | D1.cyber_offense_capability | 0.5 | — | judgment | 0.5 | 2026-02-03 | IASR 2026 (largest role in preparatory stages) and the 2025 espionage disclosure (agentic execution with human strategic control): the documented mid-point | autonomy share is a developer estimate |
| `do-d1-bio-chem-uplift` | D1.bio_chem_uplift | 0.45 | — | judgment | 0.4 | 2026-02-03 | IASR 2026: assistance with instructions and troubleshooting; real-world uplift uncertain; no operational detail recorded | real-world uplift not demonstrated |
| `do-d2-unattended-operation-hours` | D2.unattended_operation_hours | 0.343 | 4.817 h | observation | 0.5 | 2025-12-18 | log2(1+4.817)/log2(169) | 50 percent success is not "reliable"; the 80 percent horizon is much shorter |
| `do-d2-autonomous-action-evidence` | D2.autonomous_action_evidence | 0.55 | — | judgment | 0.6 | 2026-08-04 | AISI incident (contained, no known harm) and the Replit data loss (real harm, one deployment): slightly above the "contained cases" point | both quickly contained or remediated |
| `do-d4-frontier-compute-growth` | D4.frontier_compute_growth | 0.653 | 4.5 per year | observation | 0.6 | 2024-05-28 | log10(4.5)/log10(10) from Epoch AI | 2024 estimate; growth may have changed |

The normalisation recipes are stated per signal in `drivers.json` and shown in the
`SignalTable` on `/capabilities` and `/agents`. `D1.ai_rnd_automation`, `D2.delegation_depth`,
`D2.human_approval_frequency`, `D4.inference_cost_decline` and `D4.agent_population_scale` are
defined signals with no observation in this snapshot; they are not weighted in the Capability
Pressure Index, whose eight weighted signals are all observed (coverage 1.0 in the release).

## 4. How the observations enter the release

The Capability Pressure Index weights (`model_spec.index_weights.capability_pressure`) are
`D1.task_horizon_50pct` 0.25, `D1.swe_bench_verified` 0.10, `D1.computer_use_osworld` 0.10,
`D1.cyber_offense_capability` 0.15, `D1.bio_chem_uplift` 0.10, `D2.unattended_operation_hours`
0.10, `D2.autonomous_action_evidence` 0.10, `D4.frontier_compute_growth` 0.10. Tier multipliers
apply per observation from its best source tier (the two tracker-based observations rest on tier
3 sources, multiplier 0.5). The release lists every component with its contribution in
`indexes.json` and the removal sensitivity in `drivers_explained.json`; the uncertainty score's
`judgment_share` component (0.5965 in the release) reflects that 9 of the 17 observations across
all indexes are judgments.

Claims supporting the observations: `clm-metr-2025-doubling-7-months-01`,
`clm-metr-2025-doubling-4-months-recent-01` (candidate), `clm-metr-claude-3-7-sonnet-horizon-01`,
`clm-metr-o3-horizon-01`, `clm-metr-gpt-5-1-codex-max-horizon-01` (candidate),
`clm-metr-claude-opus-4-5-horizon-01`, `clm-iasr-2026-agents-30-minutes-01`,
`clm-epoch-2024-compute-4-5x-01`, `clm-steel-2026-swe-bench-verified-97pct-01` (candidate),
`clm-benchlm-2026-osworld-verified-86pct-01` (candidate).

## Limitations

- The time-horizon values carry wide confidence intervals (the Opus 4.5 interval spans an order
  of magnitude) and METR's own limitations note cautions against over-interpretation; the
  observation confidence is 0.6.
- Two observations rest on third-party trackers (tier 3) that aggregate heterogeneous runs;
  audits found flawed tests at high SWE-bench scores, so that signal mostly measures saturation.
- Two D1 signals and one D2 signal are judgments mapped onto documented scales; they carry
  confidence 0.4–0.6 and are listed as judgments in the release.
- The compute-growth observation dates from May 2024.
- The 50 percent horizon is used as the "reliable unattended operation" length for D2, which the
  record's own counterevidence says overstates reliability.
- Several result dates are approximate (month-level), noted on the records.
- No record is `verified_fetch`; none has been human-reviewed.

## Open questions

- Should the 80 percent horizon replace the 50 percent horizon for `D2.unattended_operation_hours`
  once a verified value exists?
- Should tracker-reported benchmark scores be admitted at all, or only maintainer leaderboards?
- What verified source could observe `D1.ai_rnd_automation`, the signal most directly tied to
  scenario S14?
- Is a 160-hour work-month the right upper anchor for the time-horizon normalisation?
