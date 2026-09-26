<!-- Copyright NU Cybernetics. PDUM — research prototype. -->
# Architecture overview

PDUM is a structured forecasting and evidence-synthesis system, not a thermometer. The
architecture enforces that distinction with three hard boundaries:

1. **Evidence never publishes itself.** Ingestion (`cmd/pdum-ingest`) can only write to a
   review queue. Humans assemble reviewed evidence into a *versioned data snapshot*
   (`data/snapshots/<id>`), which is immutable once sealed.
2. **Computation is deterministic and inspectable.** `internal/model` turns a snapshot into a
   *candidate* (`data/candidates/<id>`): indexes, external-forecast aggregations under several
   methods, a research-mode causal model, sensitivity runs and a change record. The same
   snapshot and code always produce byte-identical output (fixed seeds, canonical JSON, no
   wall-clock reads).
3. **Publication is a signed human decision.** `pdumctl release promote` verifies ed25519
   approvals, re-runs the model to prove reproducibility, checks invariants and heightened-review
   triggers, then copies the candidate to `data/releases/<id>` and moves the `CURRENT` pointer.
   The web app, API and MCP server only ever read promoted releases.

```
 sources ─▶ pdum-ingest ─▶ review queue ─▶ human review ─▶ snapshot (sealed, hashed)
                                                              │
                                                pdumctl model run (deterministic)
                                                              │
                                                     candidate release
                                                              │
                                   pdumctl release approve (ed25519, N reviewers)
                                                              │
                          pdumctl release promote (reproduce, invariants, triggers, audit)
                                                              │
                                              data/releases/<id> + CURRENT
                                                              │
                        ┌─────────────────────────┬───────────┴────────────┬───────────────────┐
                    apps/web (SSR)           cmd/pdum-api (OpenAPI)   services/mcp (stdio)   exports (JSON/CSV)
```

## Components

| Layer | Location | Notes |
| --- | --- | --- |
| Contracts | `packages/schemas` (zod), `data/schemas` (JSON Schema), `internal/schema` (Go) | One vocabulary; Go validates snapshots against the generated JSON Schemas. |
| Snapshot & model | `internal/snapshot`, `internal/model` | Loader with hash verification; indexes, aggregation, causal model, sensitivity, rounding. |
| Safety gate | `internal/publishing`, `internal/audit`, `cmd/pdumctl` | Candidates, signed approvals, promotion, rollback, hash-chained audit log. |
| Ingestion | `internal/{robots,fetch,parsing,dedup,sources,claims,review}`, `cmd/pdum-ingest` | Allowlisted, robots-respecting, SSRF-guarded, size-bounded; outputs review items only. |
| Public API | `internal/api`, `cmd/pdum-api`, `api/openapi.yaml` | Read endpoints, scenario-lab evaluation, submissions to the review queue. |
| Web | `apps/web` | Next.js App Router; `/text` is server-rendered and complete; immersive modes are progressive enhancement. |
| Shared TS | `packages/sdk`, `packages/model-core`, `packages/design-system` | Data access, Scenario Lab model (mirrors Go), tokens. |
| MCP | `services/mcp` | Read tools plus three submission tools; no publishing or configuration surface. |
| Research | `research/` | Reports that justify every snapshot record. |
| Governance | `docs/governance`, `docs/security`, `docs/method` | Policies the code enforces or the editors follow. |

## Outputs kept separate

| Output | Object | Where computed | Is a probability? |
| --- | --- | --- | --- |
| A. External forecast aggregate | `estimates.json` (`status: external_aggregate`) and `aggregations.json` | `internal/model/aggregate.go` | Yes (aggregated from named sources) |
| B. PDUM model estimate | `estimates.json` (`status: research_mode`); official object `insufficiently_calibrated` | `internal/model/causal.go`, `run.go` | Yes, research mode only |
| C–F. Indexes | `indexes.json` | `internal/model/indexes.go` | **No** |
| G. Uncertainty score | `indexes.json` (`uncertainty`) | `internal/model/indexes.go` | **No** |
| H. Editorial risk level | `manifest.json` / `Result.EditorialRiskLevel` | rule evaluation over indexes | **No** |

See `build-spec.md` for the binding contract and `docs/method/` for the public methodology.
