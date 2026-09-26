<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Research reports

The research directory holds the reports that justify every record in the data snapshot.
A record in `data/snapshots/<id>/*.json` may exist only if a report here explains where it
came from, how it was verified, what it means for the model and what is still uncertain. The
reports are prose; the records are data; the two must agree.

## Layout

```
research/
  README.md                          this page
  forecasts/report.md                external forecasts: inventory, groups, transformations
  forecasts/forecast-inventory.md    one row per forecast record with wording, population, horizon, group
  forecasts/compatibility-groups.md  the declared group_id values and the outcome-code conventions
  forecasts/fragments/               (empty) fragment files for future forecast sources
  scenarios/report.md                S1–S18, edges, content-safety notes
  scenarios/fragments/               sources.json, claims.json authored for the scenario area (not yet merged)
  sources/report.md                  tier distribution, verification statuses, conflicts, licences
  sources/fragments/                 (empty)
  definitions/report.md              the 29 terms and their attributed definitions
  definitions/fragments/             (empty)
  capabilities/report.md             benchmarks, results, D1/D2/D4 observations
  incidents/report.md                the 11 incidents and the incident pressure inputs
  interventions/report.md            I01–I27, organisations, actions
  methodology/experimental-causal-parameters.md   the rationale behind every research-mode parameter
```

The snapshot itself is built from `tools/snapshot/content/*.mjs` by
`node tools/snapshot/build.mjs`; fragments under `research/<area>/fragments/` are merged by
`node tools/snapshot/merge-fragments.mjs <snapshot-dir>` (deduplicated by canonical URL). The
reports describe what is in the built snapshot `snap-2026-09-26-001`; where a fragment has not
been merged, the report says so.

## The rules (build-spec §8)

1. **Real sources only.** Canonical URL, publisher, authors, dates. Verify by fetching the URL
   or through a search result; record the method in `verification.method`. If a figure cannot
   be verified, omit it or record `verification.status: "unverified"` with
   `model_use_status: "excluded"`.
2. **Quote question wording verbatim** where licensing permits short quotes; otherwise a marked
   paraphrase (`paraphrase: true`).
3. **Conflicts.** Company and lab documents carry `developer_self_report`; advocacy
   organisations `advocacy_context`; government publications `government_policy_context` with
   a `jurisdiction`.
4. **Predictions are predictions**, whatever the author's standing; `source_type` and
   `evidence_type` say so (`forecast`, `expert_judgment`).
5. **Incidents** need an external registry id (AIID, OECD AIM, court docket, regulator
   reference, CVE) or a tier 1–2 source. No graphic detail; no operational detail.
6. **Scenarios and interventions** are category-level; indicators must be observable; no
   procedures.
7. **Every report ends with "Limitations" and "Open questions".**

## How verification was actually done for `snap-2026-09-26-001`

Direct page fetches were blocked by the build environment's egress policy. The helpers in
`tools/snapshot/content/helpers.mjs` therefore record one of two honest statuses:

- `verified_search` — method text: "WebSearch result snippet quoting the figure/date; page
  fetch blocked by egress policy", `checked_at: 2026-09-26`. Such items may be `eligible` and
  feed computations.
- `verified_prior_knowledge` — method text: "prior knowledge, not re-fetched", with the note
  "Informational only; excluded from model computations until fetched and reviewed." Such
  items are `informational` and never feed the model.

Across the snapshot: sources 53 `verified_search` / 17 `verified_prior_knowledge`; claims 34/0;
forecasts 15/0; benchmarks 3/0; benchmark results 6/1; incidents 11/0; scenarios 18/0
(all `informational` by design); interventions 27/0 (`informational`); organisations 8/10;
driver observations 17/0; definitions 0/29; actions 0/57. Every record has
`human_review_status: pending`. Reviewers must fetch each canonical URL, confirm the figure
and wording against the primary document, and change the status to `verified_fetch` and
`reviewed` before the next snapshot; the snapshot manifest's `notes` field says the same.

## What "justify" means

For each record class the report must let a reviewer answer:

| Question | Where the answer lives |
| --- | --- |
| Is it real and where is it? | `canonical_url`, `citation`, the report's inventory table |
| What exactly did it say? | `question_wording_original`, `direct_quote_pointer`, `evidence_summary` |
| How was it verified? | `verification.status`, `method`, `checked_at`, `note` |
| May the model use it? | `model_use_status`, tier, conflicts, retraction status; the validator enforces the rules |
| How does it enter a number? | `group_id` and `transformation_note` (forecasts); `signal_id`, `normalization`, `value_normalized` and `rationale` (observations); severity, relevance and evidence weights (incidents) |
| What is wrong with it? | `limitations`, `counterevidence`, and the report's Limitations section |

## Cross-references

- Public methodology: [`../docs/method/model.md`](../docs/method/model.md) and the other
  documents in `docs/method/`.
- Editorial rules: [`../docs/governance/editorial-policy.md`](../docs/governance/editorial-policy.md).
- Corrections: [`../docs/governance/corrections-policy.md`](../docs/governance/corrections-policy.md).
- Building and validating a snapshot: [`../docs/operations/runbook.md`](../docs/operations/runbook.md).

## Limitations

- The reports describe `snap-2026-09-26-001` only; there is no second snapshot to compare
  against and no backtest.
- The scenario fragments (`research/scenarios/fragments/`) contain 43 sources and 22 claims
  authored under different ids from the content modules and are not merged into the snapshot;
  see [`scenarios/report.md`](scenarios/report.md).

## Open questions

- Which records should be re-verified first once page fetches are possible: the forecast
  wordings marked `paraphrase: true`, or the driver observations that carry the largest index
  weights?
- Should `verified_prior_knowledge` remain a permitted status at all in a second snapshot, or
  should such records be dropped until fetched?
