<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Cassandra charter: adversarial review

Cassandra is the role whose job is to be believed. Before any candidate is promoted, one
reviewer takes the adversarial position: attack every number, every label and every
assumption, and write down what would have to be true for the release to be wrong.

## 1. Mandate

The Cassandra reviewer must try to break:

| Target | Attacks |
| --- | --- |
| External aggregates | Are group members really the same question? Was the transformation to outcome codes justified (`transformation_note`)? Does one source dominate (leave-one-source-out runs in `sensitivity.json`)? Is a `paraphrase: true` wording faithful? Is the as-of date of a platform value verified? |
| Research-mode model | Is each `{p05, p50, p95}` judgment in `model_spec.experimental_causal` anchored to a cited source? Would a defensible alternative anchor move the median materially? Is the common-factor loading reasonable? Are horizons monotone for the right reason? |
| Indexes | Is every eligible observation's normalisation recipe applied correctly (`drivers.json`)? Are `judgment` observations labelled as such? Are tier multipliers and weights the ones in `model_spec`? What is missing from coverage and does the note say so? |
| Incidents | Does every incident have a registry id or a tier 1–2 source? Is the severity, relevance and evidence level defensible? Is any summary graphic or operational? |
| Sources | Is every canonical URL real and the citation right? Is the tier right for the document type? Are conflicts labelled? Are `verified_prior_knowledge` items excluded from computation? |
| Scenarios and interventions | Category level only; indicators observable; no procedures; effect sizes qualitative. |
| Language and display | Any headline probability? Any index read as a probability? Any countdown, inevitability or alarm? Rounding rule applied? Original wording beside transformed values? |
| Process | Does the delta record list every trigger? Can the run be reproduced from the reproduction command? Are the approvals' conflict declarations complete? |

The mandate is to find reasons the release is wrong, not to balance them against reasons it is
right. Balance is the approving reviewers' job.

## 2. The finding schema

Findings are records of `CassandraFindingSchema` (`packages/schemas/src/lab.ts`; JSON Schema in
`data/schemas/cassandra_finding.schema.json`):

| Field | Meaning |
| --- | --- |
| `finding_id` | Stable id, for example `cas-rel-2026-09-26-001-03` |
| `severity` | `critical`, `high`, `medium`, `low` (`CassandraSeverity`) |
| `component` | What is attacked: an estimate id, group id, index id, signal id, record id or page |
| `claim` | The published statement being attacked, quoted |
| `assumption` | The assumption the statement rests on |
| `counterevidence` | What contradicts it, with sources |
| `reproduction` | How anyone can see the problem (command, file, line, page anchor) |
| `impact` | What would change if the finding is right (which outputs, by roughly how much, or which rule is broken) |
| `recommendation` | What to do: exclude, relabel, re-verify, widen an interval, withhold, rewrite |
| `status` | `open`, `resolved`, `accepted_risk`, `disputed` (`CassandraStatus`) |

Severity guide:

- `critical`: a rule of build-spec §0 is broken (a fabricated or unverified input feeding the
  model, a blended aggregate, a headline probability, operational detail, a hardcoded number).
- `high`: a published value would move by more than its rounding step, or a label (tier,
  conflict, status, relevance) is wrong in a way that changes model use.
- `medium`: a limitation is understated or missing from the changelog, model card or known
  limitations; a wording or date is imprecise without changing a value.
- `low`: presentation, typography, links, minor citation form.

## 3. When it runs

Cassandra review happens after `pdoomctl model run` and `pdoomctl release diff` and before any
approval is signed. It runs for every candidate, and again for the corrected candidate when a
finding forced a change. Heightened-review releases ([`update-governance.md`](update-governance.md))
require the Cassandra reviewer to attack the trigger specifically (the new scenario, the new
forecast source, the redefined benchmark, the model version change, the moved estimate).

## 4. How findings block or annotate a release

| Highest open severity | Effect |
| --- | --- |
| `critical` | No approval may be signed. The candidate is abandoned; a corrected snapshot and a new candidate are produced. |
| `high` | No approval until the finding is `resolved` (a new candidate) or `accepted_risk` with a written justification from both approving reviewers, copied into the changelog. |
| `medium` | May be promoted; the finding's `recommendation` text is added to the changelog and, where it concerns a limitation, to the reviewers' notes for the model card. |
| `low` | Tracked for the next release. |
| `disputed` | Treated as one level lower than its severity for gating, and must be listed in the changelog as disputed. |

The gate is a policy today, not a tool check; `pdoomctl release promote` verifies approvals,
reproducibility and invariants and does not read findings.

## 5. Independence

- The Cassandra reviewer for a candidate is not the reviewer who authored the content modules
  or research fragments that changed in it, and is not the acknowledging reviewer for its
  heightened-review triggers.
- A Cassandra reviewer may also sign an approval only when every finding they raised is
  `resolved` or `accepted_risk`; a `disputed` finding of their own disqualifies them from
  approving that candidate.
- Findings are never edited once filed; a change of view is a new finding or a status change
  with a note.
- The role rotates between releases where the reviewer pool allows.

## 6. Record for `rel-2026-09-26-001`

The prototype release was produced with two development keys held by the build operator. No
separate Cassandra findings file exists for it. The known-limitation list in its manifest and
model card (no calibration; judgment parameters; small groups dominated by a few studies;
single-factor dependence; registry-only incidents; rule-based editorial level) is the closest
record, and the largest sensitivity runs published in the model card show that the aggregate
`G-EXT-2100-METACULUS` and the group `G-CAT10-2100` each rest on two members, so removing either
member moves the aggregate by several points. A Cassandra reviewer should treat both as `high`
findings against any use of those aggregates as headline figures.

## Not yet implemented

- `cassandra-findings.json` inside candidate and release directories, hashed into the manifest.
- A promotion gate that refuses `critical` or `high` open findings.
- Serving findings through the API and on `/releases/[releaseId]`.
