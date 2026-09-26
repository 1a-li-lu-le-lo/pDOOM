<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Update governance

How a new release comes to exist, who signs it, what triggers heightened review, and how a
release is withdrawn. The mechanics are in `internal/publishing`, `internal/audit` and
`cmd/pdoomctl`; the operator commands are in [`../operations/runbook.md`](../operations/runbook.md).

## 1. Cadence

There is no timer. A release is made when a reviewed snapshot exists, and only then. The
prototype has one release, `rel-2026-09-26-001`, from `snap-2026-09-26-001` (source cutoff
2026-09-26), generated at `2026-09-26T21:45:00Z` and published at `2026-09-26T22:00:00Z`. The
first backtest is scheduled once a second snapshot exists
([`../method/backtesting.md`](../method/backtesting.md)). Ingestion runs
(`pdoom-ingest`) may happen on any schedule; they never produce a release.

## 2. The path from evidence to release

```
research fragments / content modules
        │  node tools/snapshot/build.mjs <snap-id>
        ▼
data/snapshots/<snap-id>/           pdoomctl snapshot seal   → file hashes and counts in manifest.json
                                    pdoomctl snapshot validate → JSON Schema + invariants (exit 1 on errors)
        │  pdoomctl model run --snapshot … --out data/candidates/<cand-id> --generated-at <ts> [--code-commit <sha>]
        ▼
data/candidates/<cand-id>/          estimates, indexes, aggregations, sensitivity, drivers_explained, delta,
                                    manifest (signature = manifest hash), approvals.json ([]), changelog.md, model-card.md
        │  pdoomctl release diff <cand-dir>          → change record versus the current release
        │  pdoomctl release approve <cand-dir> --reviewer <id> --key <file> --signed-at <ts> --conflicts "<text>"   (× N)
        ▼
        │  pdoomctl release promote <cand-dir> --published-at <ts> [--heightened-review-ack <id>] [--actor <name>]
        ▼
data/releases/<rel-id>/ + data/releases/CURRENT + data/audit/audit.jsonl (release.promote)
```

Timestamps are always supplied by the operator; no tool reads the clock for anything that
becomes part of a release, so the same inputs always produce the same bytes.

## 3. What promotion verifies

`internal/publishing.Promote` runs these gates in order and refuses on the first failure
(`GateError{Gate, Reason}`):

| Gate | Check |
| --- | --- |
| `options` | `--published-at` present |
| `integrity` | Every data file's SHA-256 matches the manifest; the candidate is not already marked published |
| `approvals` | The manifest hash equals the recorded `signature`; every approval signs that hash with a key whose public half is in `data/keys/reviewers/<key_id>.pub` (ed25519); distinct approvers ≥ required (default `RequiredApprovals = 1`, overridable with `--required-approvals`) |
| `heightened_review` | If `delta.heightened_review_triggers` is non-empty: an acknowledging reviewer id is given, at least two distinct reviewers have signed, and the acknowledging reviewer is one of them |
| `reproducibility` | The referenced snapshot loads and validates; the model is re-run with the manifest's `generated_at` and `code_commit`; the invariants pass; the re-run's six data files (`estimates.json`, `indexes.json`, `aggregations.json`, `sensitivity.json`, `drivers_explained.json`, `delta.json`) are byte-for-byte identical to the candidate's |
| `invariants` | `internal/model.CheckInvariants` (official object has no quantiles; quantiles ordered and in [0,1]; research-mode medians monotone across horizons within 0.005; P_DOOM mean equals the sum of component means; indexes in [0,100] with `is_probability: false` and weights in [0, 0.35]; every aggregation has at least two members and weights that sum to 1; a valid editorial level) |
| `uniqueness` | The release id does not exist yet, and no existing release publishes the same snapshot, commit and generation time |

Only after all gates pass does the tool copy the directory, write `published`, `reviewers`
and the `approval` policy block into the manifest, mark the previous release
`superseded: { by, at }`, write `CURRENT`, and append the audit event. The web app
(`@pdoom/sdk`, re-reads `CURRENT` at most every 5 s) and the API (same interval) pick the
new release up without a restart.

## 4. Heightened-review triggers

Triggers are computed by `internal/model/delta.go` and stored in `delta.json` and
`changelog.md`. Any trigger requires two distinct signed reviewers plus an explicit
`--heightened-review-ack <reviewer-id>` from one of them.

| Trigger name | When it fires |
| --- | --- |
| `first_release` | No previous release |
| `model_version_changed` | The set of model versions differs from the previous release |
| `new_scenario_introduced` | A scenario id not present in the previous release's data summary |
| `expert_survey_or_forecast_source_added` | A forecast source id not present in the previous release |
| `benchmark_redefined` | A benchmark's name, version, URL, unit or direction hash differs from the previous release |
| `headline_pdoom_material_change` | Any research-mode P_DOOM median moves by 2.0 points or more (`headlineTriggerPoints`) |
| `extinction_estimate_changed` | Any estimate for the extinction outcome set (O6) moves at all |
| `security_incident_involved` | An eligible incident has cause `security_compromise` |
| `single_laboratory_evidence` | An eligible driver observation rests solely on `developer_self_report` sources from one publisher |

`rel-2026-09-26-001` carried `first_release` and `security_incident_involved` (four eligible
incidents have a `security_compromise` cause) and was acknowledged by `nuc-methodology-dev`.

The build specification's list (model version change, new scenario, new forecast source,
benchmark redefinition, large estimate movement) maps onto the names above; the two
data-driven triggers (`security_incident_involved`, `single_laboratory_evidence`) are
additional and fire regardless of a previous release.

## 5. Reviewers and keys

- Keys are created with `pdoomctl keygen --id <reviewer-id> --out data/keys/reviewers`
  (reviewer ids match `^[a-z0-9]+(?:[-_.][a-z0-9]+)*$`). The `.pub` file is committed; the
  `.key` file is mode 0600 and gitignored (`*.key`, `data/keys/reviewers/*.key`).
- The tool refuses to overwrite an existing key pair.
- Present public keys: `nuc-editorial-dev.pub`, `nuc-methodology-dev.pub`.
- An approval signs the hex manifest hash; the manifest hash is the SHA-256 of the canonical
  manifest with `signature`, `approval`, `reviewers`, `published` and `superseded` cleared, so
  later publication fields do not invalidate signatures, while any other manifest edit does.
- An approval can only be added to a candidate, never to a published release, and one
  reviewer can approve a candidate once.
- Each approval carries a `conflicts_declared` text ([`conflict-of-interest.md`](conflict-of-interest.md)).

## 6. Rollback

`pdoomctl release rollback <release-id> --at <ts> --reason "<text>" [--actor <name>]`
points `CURRENT` at an existing promoted release. It requires a reason, refuses a release that
is already current, and appends `release.rollback` with `from` and `reason` to the audit log.
Nothing is deleted: the rolled-back release keeps its directory and its `published` field, so
the history stays complete. A later promotion computes its delta against whatever release is
current at that time (or `--previous <rel-id>` to choose explicitly; `--no-previous` computes a
baseline release).

## 7. Immutability and the audit chain

- Candidate directories are never overwritten (`WriteCandidate` refuses an existing directory).
- Release directories are never modified except for the two publication-time writes performed
  by `Promote` (own manifest: `published`, `reviewers`, `approval`; previous manifest:
  `superseded`).
- `data/audit/audit.jsonl` is append-only; each event's hash is
  `sha256(prev_hash + canonical_json(event without hash))`, the genesis `prev_hash` is 64 zeros,
  and `Append` refuses to extend a broken chain. CI runs `pdoomctl audit verify` on the
  committed log. The log currently holds one event: the promotion of `rel-2026-09-26-001` by
  actor `nu-cybernetics-build`.

## 8. Release identifiers

`rel-YYYY-MM-DD-NNN`, `cand-YYYY-MM-DD-NNN`, `snap-YYYY-MM-DD-NNN`. `pdoomctl model run`
reserves the next release number for the generation date unless `--release-id` is given.

## Not yet implemented

- A Cassandra findings file as a promotion input ([`cassandra-charter.md`](cassandra-charter.md));
  today findings are handled outside the tool.
- More than one required approval by default (`RequiredApprovals = 1`; heightened review
  already requires two).
- A scheduled cadence or a release calendar.
- `docs/operations/pending-deps.md` named in build-spec §2 does not exist in this repository.
