<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Corrections policy

A number that moves without an explanation is a rumour; a number that is wrong and stays is
worse. This page describes how corrections reach the observatory, how they are decided, and
how a corrected release is issued. Appeals against editorial decisions follow the same path
and are described at the end.

## 1. Intake: everything goes through the review queue

There are three intake channels. None of them can change a published value.

| Channel | What it writes | Status |
| --- | --- | --- |
| `POST /v1/submissions/corrections` on `cmd/pdoom-api` | One JSON line in `data/review-queue/submissions.jsonl` (`kind: correction`, `status: received`) | Implemented ([`../api/README.md`](../api/README.md)) |
| `POST /v1/submissions/sources` | One JSON line, `kind: source` | Implemented |
| MCP submission tools in `services/mcp` | The same file or the same API route (build-spec §7) | Not yet implemented; `services/mcp/src/server.ts` is a placeholder |

A correction payload (`CorrectionSubmissionPayloadSchema` in `packages/schemas/src/lab.ts`,
mirrored by `internal/api/handlers_submit.go`) has:

```json
{ "payload": { "target_id": "src-example", "field": "date_published",
               "correction": "Published in May, not June.",
               "evidence_url": "https://example.org/report" },
  "contact": "reviewer@example.org" }
```

`target_id` may name any entity or estimate id (`src-…`, `fc-…`, `inc-…`, `S3`, `I04`,
`est-research-P_DOOM-10y`). The API strips control characters, bounds every field
(500/4000/2048/320 characters), rate-limits the route (burst 5, 10 per minute per client) and
answers `202` with a submission id and the note that nothing published changes until
reviewers assemble a new snapshot and a signed release is promoted. No raw IP address is
logged.

The queue directory is gitignored except its README; reviewers read `submissions.jsonl`
directly (`pdoom-ingest` neither reads nor writes it).

## 2. Triage

A reviewer classifies each submission:

| Class | Example | Consequence |
| --- | --- | --- |
| Metadata | wrong `date_published`, author list, licence, citation | fixed in the content module (`tools/snapshot/content/*.mjs`) or research fragment; enters the next snapshot |
| Value | a forecast median transcribed wrongly; a benchmark result misread | as above, plus a note in `verification.note`; if the value feeds the model, a corrected release is required (section 3) |
| Wording | question wording not verbatim, `paraphrase` flag missing | fixed; group membership re-checked against `research/forecasts/compatibility-groups.md` |
| Classification | tier, conflict label, relevance, severity, evidence level disputed | the record is set to `human_review_status: disputed` until two reviewers agree; the decision is written into the relevant research report |
| Retraction | a source is retracted or corrected by its publisher | section 4 |
| Content safety | a record contains operational or graphic detail | removed immediately in the next snapshot; `content_safety_note` records the removal; a corrected release is issued whether or not a number changes |
| Method | a documented error in `internal/model` or `packages/model-core` | code fix, golden fixtures regenerated, model version bumped; heightened-review trigger `model_version_changed` |

Submissions that cannot be verified from a canonical URL are left in the queue with no
action; the queue is never edited or deleted.

## 3. Issuing a corrected release

Releases are immutable: `internal/publishing.LoadDir` verifies the SHA-256 of every data
file against the manifest and refuses a directory that was altered, and `Promote` refuses a
release id that already exists. A correction therefore always produces a new release:

1. Apply the fix to the content modules or fragments and build a new snapshot directory
   (`node tools/snapshot/build.mjs snap-YYYY-MM-DD-NNN`), then seal and validate it
   (`pdoomctl snapshot seal`, `pdoomctl snapshot validate`). The corrected snapshot's
   `manifest.notes` names the correction and the submission id.
2. Run `pdoomctl model run --snapshot … --out data/candidates/cand-…`. The candidate's
   `delta.json` compares every estimate and index with the current release and lists the
   heightened-review triggers.
3. `pdoomctl release diff <candidate>` shows the movement. The changelog entry states the
   correction in plain words ("Corrected: <what>, reported via submission <id>").
4. Collect approvals and promote ([`update-governance.md`](update-governance.md)). A
   correction that moves the research-mode headline by two or more points, or any extinction
   estimate at all, is a heightened-review release.
5. On promotion the previous release's manifest receives `superseded: { by, at }`, `CURRENT`
   moves, and an audit event `release.promote` is appended. The superseded release stays
   readable at `/releases/<id>` and `/v1/releases/<id>`, marked "Superseded by … on …".

If the current release is itself defective (for example a content-safety failure), the
operator may first roll back to the previous release: `pdoomctl release rollback <release-id>
--at <ts> --reason "<why>"`. Rollback moves only the `CURRENT` pointer and appends
`release.rollback` to the audit log; the defective release directory is kept, so the record
of what was published is preserved.

## 4. Retraction handling for sources

When a publisher retracts or corrects a source:

1. Set `retraction_status` to `retracted` (or `corrected`/`disputed`) and describe it in
   `correction_status`.
2. Set `model_use_status: excluded`. The snapshot validator refuses a retracted source that is
   `eligible` or `used`, so the model cannot be re-run with it.
3. Set each dependent claim to `status: retracted` (or `superseded` when a corrected figure
   replaces it) and add a new claim for the corrected value where one exists.
4. Re-run the model. Forecasts, driver observations and incidents that cite only the retracted
   source drop out of compatibility groups and index coverage; the delta record shows the
   effect, and coverage gaps are listed on the index.
5. Issue the corrected release as in section 3. The source remains in the ledger with its
   retraction status visible; it is not deleted, because readers must be able to see what the
   earlier release relied on.

No source in `snap-2026-09-26-001` has a retraction status other than `none`.

## 5. Public record

- `/changelog` and `/releases/[releaseId]` show every release, its `changes`, approvals and the
  per-estimate delta.
- `data/releases/<id>/changelog.md` and `delta.json` are shipped with each release and served
  by `/v1/releases/{id}`.
- The hash-chained audit log (`data/audit/audit.jsonl`) records every promotion and rollback
  with actor, timestamp, approvers and manifest hash; `pdoomctl audit verify` checks the chain.

## 6. Appeals

Anyone may appeal an editorial decision (a tier, a conflict label, a group membership, an
exclusion, a wording judgement) through the corrections route with `target_id` set to the
record and `correction` describing the disputed decision and the evidence. An appeal is handled
as a "classification" triage item: the record is marked `disputed`, two reviewers who did not
make the original decision consider it, and the outcome is written into the research report
for that area with the reasoning. The record is only returned to `reviewed` when they agree;
otherwise it stays `disputed` and is displayed as such.

## Not yet implemented

- A service-level target for acknowledging or resolving submissions; the prototype
  acknowledges receipt synchronously (`202`) and makes no time commitment.
- A public list of open submissions and their triage state; today only the outcomes (changelog,
  delta record, research reports) are public.
- MCP submission tools (`services/mcp`).
- Re-queuing a submission as a `ReviewItem` (`review.KindSubmission` exists in
  `internal/review` but the API does not write that shape).
