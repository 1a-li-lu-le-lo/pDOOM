# Review queue

Runtime output of `pdoom-ingest` and of public submissions (`POST /v1/submissions/*` on the
API and the `pdoom_submit_*` MCP tools). Contents are gitignored.

Nothing in this directory can change a published estimate. Items are append-only JSON lines
(`<date>-ingest.jsonl`, `submissions.jsonl`), reviewed by humans; anything accepted is entered
into a *new* data snapshot through `tools/snapshot/content/*.mjs`, validated, sealed and promoted
through the signed release gate (`docs/operations/runbook.md`). A `pdoomctl review` command that
lists and marks items is not yet implemented; reviewers read the files directly.
