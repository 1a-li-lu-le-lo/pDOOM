<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Editorial policy

This policy says what p(DOOM) may publish, who decides, and in what words. It is the
editorial reading of the binding constitution in
[`docs/architecture/build-spec.md`](../architecture/build-spec.md) (§0). Where the two
disagree, the build specification wins.

## 1. The only thing that is published is a promoted release

Every number, label and sentence shown by the web app (`apps/web`), the API
(`cmd/pdoom-api`) and the exports (`/api/export/*`) is read from the release that
`data/releases/CURRENT` points at. In this repository that is `rel-2026-09-26-001`,
computed from snapshot `snap-2026-09-26-001` (70 sources, 15 forecast records,
11 incidents, 18 scenarios, 27 interventions).

Nothing else is publication:

- A snapshot is not published until a candidate computed from it is promoted.
- A candidate (`data/candidates/<id>`) is not published; `pdoomctl model run` refuses to
  write a candidate under `data/releases`.
- Ingestion output (`data/review-queue/`) is never published (build-spec rule 0.7 and
  [`../operations/ingestion.md`](../operations/ingestion.md)).
- Scenario Lab results are labelled `user_scenario` and are not stored.
- Prose in `docs/` and `research/` describes method and evidence; it carries no numbers of
  its own beyond what a release states.

## 2. The eight outputs, kept separate

The release publishes eight objects (build-spec rule 0.2, [`../method/definitions-of-outputs.md`](../method/definitions-of-outputs.md)).
They are stored as separate records and are never blended, averaged, or converted into one
another:

| Output | Object in the release | Editorial rule |
| --- | --- | --- |
| A. External forecast aggregate | `estimates.json` items with `status: external_aggregate`; `aggregations.json` | Only inside a compatibility group; original wording beside every value |
| B. p(DOOM) model estimate | `estimates.json` items with `status: research_mode`; the official object with `status: insufficiently_calibrated` | Research-mode values are always called "research mode"; the official value is withheld |
| C. Evidence Pressure Index | `indexes.json`, `evidence_pressure` | Not a probability |
| D. Capability Pressure Index | `indexes.json`, `capability_pressure` | Not a probability |
| E. Control Strength Index | `indexes.json`, `control_strength` | Not a probability |
| F. Incident Pressure Index | `indexes.json`, `incident_pressure` | Not a probability |
| G. Uncertainty Score | `indexes.json`, `uncertainty` | Not a probability; sets display rounding |
| H. Editorial risk level | `manifest.json`, `editorial_risk_level` | A rule-based label, shown with its rule |

The release also carries `agentic_infrastructure_risk` (an index) and an `attention` index
slot that is `null` in this release line because no media-volume data is ingested.

## 3. The official value is withheld

The official object exists for every horizon (`est-official-P_DOOM-1y` … `-eventual`) and
has `status: insufficiently_calibrated`, `quantiles: null` and the display text
"Insufficiently calibrated". The model invariants (`internal/model/invariants.go`) fail a
candidate whose official object carries quantiles or a mean. The decision and what would
change it are recorded in [`decision-log.md`](decision-log.md) (ADR-002) and
[`../method/calibration.md`](../method/calibration.md).

Editors therefore never write a headline probability for p(DOOM). When a value from the
release is quoted, it is quoted with its status, outcome set, horizon and interval, for
example: "the release states that the research-mode model's median for the combined
outcome set O3–O8 at the 10-year horizon is shown as 2% with a plausible interval of <1%–14%
(status `research_mode`, not the official estimate)".

## 4. Who approves

Publication is a signed human decision. The mechanism is described in
[`update-governance.md`](update-governance.md); the roles are:

| Role | What they sign | Key |
| --- | --- | --- |
| Editorial reviewer | That the candidate's language, labels, content-safety and provenance meet this policy | ed25519 key whose public half lives in `data/keys/reviewers/<id>.pub` |
| Methodology reviewer | That the model run is reproducible, the invariants hold and the delta record is explained | same |
| Acknowledging reviewer | When heightened-review triggers exist, a second distinct approver acknowledges them (`--heightened-review-ack`) | must have signed an approval |
| Cassandra (adversarial) reviewer | Attacks every number before promotion ([`cassandra-charter.md`](cassandra-charter.md)) | findings, not a signature gate, in this release line |

`internal/publishing.RequiredApprovals` is 1 in the prototype; heightened review always
requires two distinct reviewers. Release `rel-2026-09-26-001` was approved by
`nuc-editorial-dev` and `nuc-methodology-dev`, with the latter acknowledging the triggers
`first_release` and `security_incident_involved`.

## 5. What may not be published

- An official p(DOOM) probability in the v0 release line.
- Any index presented as, or converted into, a probability (build-spec rule 0.3; the
  `is_probability: false` flag is checked by the invariants).
- A number without its outcome definition, horizon, conditioning, interval, disagreement,
  model version, data cutoff and last review date (rule 0.1; enforced in the UI by
  `apps/web/components/meter/EstimateCard.tsx`).
- A blended figure across compatibility groups (rule 0.5; `internal/snapshot/validate.go`
  rejects a `group_id` that mixes outcome sets or horizons).
- Any item whose `verification.status` is `unverified` or `verified_prior_knowledge` as a
  model input (rule 0.6; the validator only lets `verified_fetch`/`verified_search` items be
  `eligible`/`used`, and tier 4–5 or retracted sources never are).
- Operational detail in any dangerous domain (rule 0.9;
  [`../method/content-safety.md`](../method/content-safety.md)).
- Countdowns, dates for catastrophe, "years left", inevitability or panic framing (rule 0.10;
  [`psychological-safety.md`](psychological-safety.md)).
- Substantive content that exists only in a 3D scene (rule 0.11; `/text` must contain it).
- A hardcoded percentage in web source (rule 0.13; `apps/web/test/no-hardcoded-numbers.test.ts`).

## 6. Language rules

1. The brand is written exactly `p(DOOM)`; the web app renders it only through the `BRAND`
   constant of `@pdoom/schemas` ([`../design/brand.md`](../design/brand.md)).
2. Display rounding follows the uncertainty label: extreme → nearest 5 points, high → nearest 2,
   moderate and low → nearest 1; never decimals; "<1%" and ">99%" at the ends; intervals with an
   en dash ([`probability-change-policy.md`](probability-change-policy.md)).
3. Status words are fixed: "Official value withheld", "External forecast aggregate",
   "Research mode", "Your scenario" (`STATUS_LABEL` in `apps/web/lib/format.ts`).
4. "eventual" is never displayed beside a dated horizon without its label, and "2100" is a
   calendar endpoint, not a countdown.
5. Predictions are called predictions whatever the author's standing (build-spec §8.4).
6. Indexes are introduced as "index, not a probability" wherever a value appears
   (`apps/web/components/meter/IndexGauge.tsx`).
7. The plain-language editorial level is shown next to the rule that produced it (`/meter`).
8. Company and lab documents are labelled `developer_self_report`, advocacy organisations
   `advocacy_context`, governments `government_policy_context`
   ([`conflict-of-interest.md`](conflict-of-interest.md)).
9. Calm register throughout: no exclamation marks, no imperatives about fear, no imagery of
   harm. The guard test also fails the build on the phrases "humanity has N years left", "doom
   is certain" and "the machines are coming".

## 7. Corrections and retractions

Corrections enter through the review queue and, when they change a published value, through a
new signed release; releases are never edited in place. See
[`corrections-policy.md`](corrections-policy.md).

## 8. Provenance duties of editors

Every source, forecast, benchmark result, incident and organisation must be real, cite a
canonical URL and carry a verification record (build-spec §8). In `snap-2026-09-26-001` the
verification method is `verified_search` for 53 of 70 sources and `verified_prior_knowledge`
for 17, because page fetches were blocked in the build environment; all 70 records have
`human_review_status: pending`. Editors treat this as a stated limitation, not as a
verified corpus. The research reports under [`../../research/`](../../research/README.md)
state it the same way.

## 9. How the policy is enforced in code

| Rule | Enforcement |
| --- | --- |
| No automatic publication | `cmd/pdoomctl` is the only writer of `data/releases`; the API and web app are read-only |
| Signed approvals, reproducibility, invariants | `internal/publishing/promote.go` gates |
| Official value withheld | `internal/model/invariants.go` |
| Compatibility groups only | `internal/snapshot/validate.go` |
| Verified inputs only | `internal/snapshot/validate.go` and the index code's eligibility filter |
| No hardcoded percentages, no panic phrases | `apps/web/test/no-hardcoded-numbers.test.ts` |
| Every number with context | `EstimateCard`, `IndexGauge`, `/text` |
| Audit trail | `internal/audit` hash chain, `pdoomctl audit verify` |

## Not yet implemented

- A named editorial board and a public roster of reviewer identities beyond the two
  development keys in `data/keys/reviewers/`.
- A Cassandra findings file inside the release directory (see
  [`cassandra-charter.md`](cassandra-charter.md)).
- A fixed publication calendar; releases follow reviewed snapshots
  ([`update-governance.md`](update-governance.md)).
