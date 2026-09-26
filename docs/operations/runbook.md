<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Runbook

Operator procedures for building a snapshot, computing a candidate, collecting approvals,
promoting, rolling back and verifying. Every command is run from the repository root. Nothing
here is automatic: the tools refuse to read the clock for anything that becomes part of a
release, so every timestamp is typed by the operator in RFC 3339 UTC.

Prerequisites: Node 22 (`.nvmrc`), pnpm 10 (`packageManager` in `package.json`), Go 1.24
(`go.mod`), and `pnpm install --frozen-lockfile` once (or `make install`).

## 1. Build a snapshot

The snapshot content is authored in `tools/snapshot/content/*.mjs` (sources, claims,
forecasts with benchmarks and incidents, scenarios and edges, interventions with organisations
and actions, definitions, drivers with observations and the model spec). Research fragments
under `research/<area>/fragments/` can be merged with `tools/snapshot/merge-fragments.mjs`.

```sh
# 1. Write the entity files and an unsealed manifest (default id shown).
node tools/snapshot/build.mjs snap-2026-09-26-001
#    → data/snapshots/snap-2026-09-26-001/{definitions,sources,claims,forecasts,benchmarks,
#      benchmark_results,incidents,scenarios,scenario_edges,drivers,driver_observations,
#      interventions,organizations,actions,model_spec}.json + manifest.json
#    The script checks referential integrity of every source id first and exits 1 on a problem.

# 2. (Optional) merge research fragments, deduplicated by canonical URL.
node tools/snapshot/merge-fragments.mjs data/snapshots/snap-2026-09-26-001 --dry-run
node tools/snapshot/merge-fragments.mjs data/snapshots/snap-2026-09-26-001

# 3. Seal: write sha256 and item counts of every entity file into manifest.json.
go run ./cmd/pdoomctl snapshot seal data/snapshots/snap-2026-09-26-001

# 4. Validate: JSON Schemas (data/schemas) + invariants; exit 1 on errors.
go run ./cmd/pdoomctl snapshot validate data/snapshots/snap-2026-09-26-001
#    or: make validate     (uses the newest snap-* directory)
```

A new snapshot always gets a new id (`snap-YYYY-MM-DD-NNN`); the current snapshot is immutable
once a release depends on it. If the JSON Schemas changed, regenerate them first
(`pnpm build:jsonschema`; CI fails if the committed schemas differ).

Validation covers, among other things: required dates and id patterns; unknown source, claim,
benchmark and scenario references; `unverified` items must be `excluded` and only
`verified_fetch`/`verified_search` items may be `eligible`/`used`; tier 4–5 and retracted
sources cannot feed the model; forecasts need original wording and a central value, and a
`group_id` may not mix outcome sets or horizons; quantiles must be ordered in [0,1]; manifest
counts must match the files. Warnings (for example an unsealed file) do not fail validation;
errors do.

## 2. Run the model into a candidate

```sh
go run ./cmd/pdoomctl model run \
  --snapshot data/snapshots/snap-2026-09-26-001 \
  --out data/candidates/cand-2026-09-26-001 \
  --generated-at 2026-09-26T21:45:00Z \
  --code-commit "$(git rev-parse --short HEAD)"
#  optional: --release-id rel-2026-09-26-001   (default: next rel-<date>-NNN)
#            --previous rel-…                   (default: the current release, if any)
#            --no-previous                      (compute a baseline release)
```

The command validates the snapshot, runs `internal/model.Run`, checks the invariants and writes
`estimates.json`, `indexes.json`, `aggregations.json`, `sensitivity.json`,
`drivers_explained.json`, `delta.json`, `manifest.json` (with `signature` = manifest hash),
`approvals.json` (`[]`), `changelog.md` and `model-card.md`. It refuses an output directory that
already exists or that lies under `data/releases`. The human-readable output ends with the next
command to run. `make candidate` runs the same with a timestamped candidate id.

`go run ./cmd/pdoom-model --snapshot <dir> --generated-at <ts>` prints the result as JSON and
writes nothing; use it to inspect a run.

Review the change record before anyone signs:

```sh
go run ./cmd/pdoomctl release diff data/candidates/cand-2026-09-26-001
#  prints every estimate and index versus the current release and the heightened-review triggers
```

Read `changelog.md` in the candidate and add the human explanation of every movement
([`../governance/probability-change-policy.md`](../governance/probability-change-policy.md)).
Run the adversarial review ([`../governance/cassandra-charter.md`](../governance/cassandra-charter.md)).

## 3. Keys

```sh
go run ./cmd/pdoomctl keygen --id nuc-editorial-dev --out data/keys/reviewers
#  public key  data/keys/reviewers/nuc-editorial-dev.pub   (commit this)
#  private key data/keys/reviewers/nuc-editorial-dev.key   (NEVER commit; *.key is gitignored)
```

Reviewer ids match `^[a-z0-9]+(?:[-_.][a-z0-9]+)*$`. The tool refuses to overwrite an existing
key. Private keys stay on the reviewer's machine (mode 0600); only `.pub` files live in the
repository. The present keys are `nuc-editorial-dev.pub` and `nuc-methodology-dev.pub`.

## 4. Approve

Each reviewer, on their own machine with their private key:

```sh
go run ./cmd/pdoomctl release approve data/candidates/cand-2026-09-26-001 \
  --reviewer nuc-editorial-dev \
  --key /secure/path/nuc-editorial-dev.key \
  --signed-at 2026-09-26T21:50:00Z \
  --conflicts "Development key held by the build operator; no financial conflicts; NU Cybernetics authorship"
```

The approval signs the manifest hash and is appended to `approvals.json`. The tool refuses if the
manifest hash does not match its recorded signature (the manifest was altered), if the candidate
is already published, or if the reviewer has already approved. If the candidate's `delta.json`
lists heightened-review triggers, a second distinct reviewer must approve as well.

## 5. Promote

```sh
go run ./cmd/pdoomctl release promote data/candidates/cand-2026-09-26-001 \
  --published-at 2026-09-26T22:00:00Z \
  --heightened-review-ack nuc-methodology-dev \
  --actor nu-cybernetics-build
#  optional: --required-approvals N   (override the default of 1; recorded in the manifest)
```

Gates, in order: integrity, approvals, heightened review, reproducibility (the model is re-run
on the snapshot and every data file must be byte-identical), invariants, uniqueness
([`../governance/update-governance.md`](../governance/update-governance.md) §3). On success the
candidate is copied to `data/releases/<release-id>/`, the manifest gains `published`,
`reviewers` and `approval`, the previous release is marked `superseded`, `data/releases/CURRENT`
is written and an audit event is appended. Commit the release directory, `CURRENT`, the audit
log and the snapshot together.

Common gate failures:

| Message | Cause | Fix |
| --- | --- | --- |
| `gate integrity failed: … sha256 mismatch` | A data file in the candidate was edited | Rebuild the candidate; never edit candidate files by hand |
| `gate approvals failed: manifest hash does not match its recorded signature` | `manifest.json` was edited after approvals | Rebuild; approvals sign the hash |
| `gate approvals failed: approval by X: … no such file` | Missing `data/keys/reviewers/X.pub` | Commit the reviewer's public key |
| `gate heightened_review failed: triggers […] require --heightened-review-ack` | Triggers present | Get a second distinct approval and pass the acknowledging reviewer's id |
| `gate reproducibility failed: <file> differs when the model is re-run` | Code or snapshot changed since the candidate was computed | Rebuild the candidate from the current code and snapshot |
| `gate uniqueness failed: release … already exists` | Release ids are immutable | Choose the next id |

## 6. How the web app and API pick up `CURRENT`

- `@pdoom/sdk` (`createFileDataSource`) resolves the data directory from `PDOOM_DATA_DIR`, or by
  walking up from the working directory to a `data/` folder containing `releases/CURRENT` or
  `snapshots/`, and re-reads `CURRENT` at most every 5 s (`refreshMs`). The Next.js server
  therefore serves a promotion within seconds without a restart.
- `cmd/pdoom-api` reads `CURRENT` on start and re-checks it at most once every 5 s; a failed
  reload keeps the previous release and logs the failure; `/readyz` returns 200 only when a
  release and its snapshot are loaded.
- Both read only `published` releases; a directory left behind by an interrupted promotion is
  skipped with a warning.

## 7. Roll back

```sh
go run ./cmd/pdoomctl release rollback rel-2026-09-26-001 \
  --at 2026-09-27T09:00:00Z \
  --reason "content-safety finding cas-…; corrected release in preparation" \
  --actor nuc-editorial-dev
```

Rollback points `CURRENT` at an existing promoted release, requires a reason, refuses the
release that is already current, and appends `release.rollback`. Nothing is deleted. Then follow
[`../governance/corrections-policy.md`](../governance/corrections-policy.md) to issue a corrected
release.

## 8. Verify

```sh
go run ./cmd/pdoomctl release list                       # every release; * marks CURRENT
go run ./cmd/pdoomctl audit verify data/audit/audit.jsonl # or: make audit-verify
make check                                               # gofmt, go build/vet/test, typecheck, lint, vitest, builds
```

To reproduce a published release independently: check out the release's `code_commit`, run the
`reproduction_command` from its manifest into a fresh candidate directory, and compare the six
data files' hashes with the release manifest's `files` list. `rel-2026-09-26-001` records:

```
go run ./cmd/pdoomctl model run --snapshot data/snapshots/snap-2026-09-26-001 --out data/candidates/cand-2026-09-26-001 --release-id rel-2026-09-26-001 --generated-at 2026-09-26T21:45:00Z --code-commit cd74db7
```

## 9. Ingestion runs

See [`ingestion.md`](ingestion.md). Ingestion writes only to `data/review-queue/` and is not
part of the release path.

## Not yet implemented

- A scheduler for ingestion or releases.
- Container images or a service definition for the two servers
  ([`deployment.md`](deployment.md) describes the process layout).
- `pdoomctl review`, mentioned in `data/review-queue/README.md`; reviewers read the queue files
  directly and write research fragments.
