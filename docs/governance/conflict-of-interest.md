<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Conflict-of-interest policy

Conflicts are recorded, displayed and weighed; they are not hidden and they do not by
themselves exclude a source. This page covers conflict labels on data records, conflict
declarations by reviewers, and recusal.

## 1. Conflict labels on sources and organisations

The vocabulary is the `ConflictLabel` enum in `packages/schemas/src/enums.ts` (mirrored in Go
and in the generated JSON Schemas under `data/schemas/`):

| Label | Applied to | Rule (build-spec §8.3, ingestion checklist step 4) |
| --- | --- | --- |
| `developer_self_report` | Documents published by a company or laboratory about its own models, incidents or safety practice | Always applied to company and lab material, including safety frameworks, model cards and incident disclosures |
| `advocacy_context` | Organisations or documents with a stated advocacy mission | Always applied to advocacy organisations; recorded on the organisation record too |
| `government_policy_context` | Government and intergovernmental publications | Always applied, together with `jurisdiction` |
| `commercial_interest` | Vendors and platforms with a product interest in the subject | Applied to security vendors, benchmark trackers with a product, commercial platforms |
| `funder_relationship` | Work funded by an organisation with a position on the subject | Applied when the funder's interest is documented |
| `none_known` | No conflict identified after review | The default; it means "none known", not "none" |

A record may carry several labels. In `snap-2026-09-26-001` the 70 sources carry
`none_known` (45), `government_policy_context` (14), `developer_self_report` (8),
`funder_relationship` (1), `advocacy_context` (1) and `commercial_interest` (1). The 18
organisations carry `government_policy_context` (1), `commercial_interest` (1) and
`advocacy_context` (2); the remainder declare no conflict.

### How labels are used by the model

- Labels do not change a source's tier. Tier is about the kind of document
  ([`../method/source-hierarchy.md`](../method/source-hierarchy.md)); conflicts are about
  interest. A developer's incident disclosure can be tier 1 and `developer_self_report` at
  the same time (for example `src-anthropic-2025-disrupting-ai-espionage`).
- `counterevidence` and `limitations` fields on the source, and `counterevidence` on the
  driver observation, are where a reviewer records what the conflicted party's account may
  understate or overstate. These are shown on `/evidence/sources/[sourceId]` and in the
  "Why this number" rows of `drivers_explained.json`.
- The delta code (`internal/model/delta.go`, `singleLab`) raises the heightened-review trigger
  `single_laboratory_evidence` when any eligible driver observation rests solely on
  `developer_self_report` sources from one publisher. A release with that trigger needs two
  distinct reviewers and an acknowledgement ([`update-governance.md`](update-governance.md)).
- `verified_prior_knowledge` and `unverified` items never feed the model regardless of their
  conflict label.

### How labels are displayed

Every source row in the source ledger (`/evidence`, `/text#sources`, `sources.csv`) shows the
conflict labels beside tier, verification and model-use status. Organisation cards on `/act`
show the organisation's own funding disclosure and conflicts as recorded, with the note that
listing is not endorsement.

## 2. Reviewer conflict declarations

Every approval is signed with `pdoomctl release approve … --conflicts "<text>"`. The text is
stored verbatim in `approvals.json` as `conflicts_declared`; when the flag is omitted the
stored value is the literal `none declared` (`internal/publishing/approve.go`). The declaration
is part of the published release: it appears on `/releases/[releaseId]` under Approvals and in
`release.json`.

Declarations for `rel-2026-09-26-001`:

| Reviewer | Declaration as signed |
| --- | --- |
| `nuc-editorial-dev` | "Development key held by the build operator; no financial conflicts; NU Cybernetics authorship" |
| `nuc-methodology-dev` | "Development key held by the build operator; heightened-review acknowledgement for first release and security-incident trigger" |

A declaration must state, at minimum:

1. employment or contracts with any organisation named in the snapshot (`organizations.json`,
   `sources.json` publishers, benchmark maintainers, model developers in `benchmark_results.json`);
2. financial interests in any model developer;
3. authorship of any source in the snapshot;
4. whether the reviewer authored the research fragments or content modules under review.

"No financial conflicts" is acceptable only when the reviewer has checked the list above.

## 3. Recusal

A reviewer recuses from approving a candidate when:

- they authored, or are employed by the publisher of, a source whose tier or
  `model_use_status` changed in the snapshot under review;
- they hold a financial interest in a developer whose model appears in a benchmark result or
  incident added in the snapshot;
- they wrote the research fragment or content module that introduced a heightened-review
  trigger (`new_scenario_introduced`, `expert_survey_or_forecast_source_added`,
  `benchmark_redefined`, `model_version_changed`);
- they are the only person able to reproduce the run and no second reviewer has done so.

A recused reviewer may still sign a Cassandra finding ([`cassandra-charter.md`](cassandra-charter.md))
but not an approval or an acknowledgement. The remaining reviewers record the recusal in
their own `--conflicts` text ("<id> recused: <reason>") so it is visible in the release.
`pdoomctl` refuses a second approval from the same reviewer id and refuses an acknowledgement
from a reviewer who has not signed, so a recused reviewer cannot be the acknowledging reviewer.

## 4. Organisation directory

Organisations in `organizations.json` are listed because they meet stated inclusion criteria
(`inclusion_criteria_met`), not as endorsements. Each record carries `funding_disclosure`,
`conflicts` and `last_verified`. Advocacy organisations (`org-center-for-ai-safety`,
`org-future-of-life-institute`) are labelled `advocacy_context`; their positions are treated
as advocacy, not measurement, and their publications as sources carry the same label.

## Not yet implemented

- A standing register of reviewer interests outside the per-release declaration text.
- A dedicated audit action for recusals; today a recusal is visible only through the other
  reviewers' declarations and the changelog.
- A validator check that a `developer_self_report` source has a non-empty `counterevidence`
  or `limitations` field.
