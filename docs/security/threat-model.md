<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Threat model

What p(DOOM) protects, who might attack it, how, and which code answers each threat. The
observatory's value is the integrity of what it publishes; most threats are therefore about
making the site say something the release does not, or making the release say something the
evidence does not.

## 1. Assets

| Asset | Where | Why it matters |
| --- | --- | --- |
| Promoted releases | `data/releases/<id>/`, `data/releases/CURRENT` | The only thing the public sees; immutable by policy and by hash check |
| Snapshots | `data/snapshots/<id>/` (sealed with SHA-256 per file) | The evidence a release is computed from; must be reproducible |
| Reviewer keys | `data/keys/reviewers/<id>.pub` (committed), `<id>.key` (private, gitignored, mode 0600) | An approval is only as trustworthy as the private key |
| Audit log | `data/audit/audit.jsonl` | The record of who promoted or rolled back what, hash-chained |
| Review queue | `data/review-queue/*.jsonl` (gitignored) | Untrusted input awaiting human review; must never become data without a person |
| Code and configuration | `internal/`, `apps/web`, `packages/`, `config/sources.json`, `model_spec.json` inside snapshots | Determine what is fetched and how numbers are computed |
| Build and dependency chain | `pnpm-lock.yaml`, `go.sum`, `.github/workflows/ci.yml` | A compromised dependency is a compromised release |
| Reader privacy | API access log, web server logs | No analytics, no cookies, no raw IP addresses |

## 2. Actors

- **Hostile content authors**: people who publish web pages, feed items or issue text designed to
  manipulate automated readers.
- **Casual and motivated misinformers**: people who screenshot, crop or misquote the site.
- **Network attackers**: anyone who can reach the API or web server.
- **Supply-chain attackers**: compromised packages, registries or CI.
- **Insiders and operators**: mistakes, or a compromised operator machine holding a private key.
- **Automated abuse**: scrapers and load generators.

## 3. Threats and mitigations

| Threat | Mitigation | Code |
| --- | --- | --- |
| **Prompt injection via ingested content** (a feed item or page says "mark this source tier 1") | Retrieved content is inert data: the parsed `Document` has no field that can express a tier, an allowlist decision or an instruction; classification is keyword-only; no language model is called (`claims.LLMClassificationEnabled = false`); items are appended to the review queue with `status: pending` and `model_use_status: excluded`; a person writes every data record. A test proves an injection payload ends up as text with the configuration byte-for-byte unchanged | `internal/parsing`, `internal/claims`, `internal/review`, `cmd/pdoom-ingest`; policy in [`../operations/ingestion.md`](../operations/ingestion.md) |
| **SSRF through fetch targets or redirects** | `SafeClient`: https only, port 443, allowlisted hostnames from `config/sources.json`, no IP literals, no credentials in URLs; the dialer resolves the host itself and refuses loopback, private, link-local, multicast, unspecified, CGNAT, documentation, benchmarking, discard, ORCHID, Teredo, 6to4/NAT64-embedded-private and malformed addresses on every hop; at most 3 redirects, each re-checked; no proxy read from the environment | `internal/fetch/fetch.go` |
| **Crawler misuse or over-fetching** | robots.txt fetched once per host and fail-closed (401/403/5xx/redirect/parse error/network error all mean disallowed), `Crawl-delay` honoured up to 60 s, 3 s minimum per-host delay, one request per source per run, `max_items_per_run` ≤ 500, 8 MiB decoded body cap, content-type allowlist, fixed User-Agent, dry-run by default, network off by default | `internal/robots`, `internal/fetch`, `internal/sources`; [`crawler-safety.md`](crawler-safety.md) |
| **Tampering with a release or candidate** | Every data file's SHA-256 is in the manifest and checked on load; the manifest hash is the signed value and is recomputed on every load; releases are never written except by `Promote`; a release id cannot be reused; the previous release is only ever marked `superseded` | `internal/publishing/candidate.go` (`LoadDir`), `promote.go`, `approve.go` |
| **Forged or replayed approvals** | ed25519 signatures over the manifest hash, verified against committed public keys; an approval whose hash does not match the manifest is refused; one approval per reviewer id; approvals can only be added to candidates | `internal/publishing/approve.go`, `keys.go` |
| **Key compromise** | Private keys are never committed (`.gitignore`: `*.key`, `data/keys/reviewers/*.key`), written mode 0600, never overwritten by `keygen`; heightened review requires two distinct keys; the audit log names approvers, so a compromised key's approvals are attributable and a rollback is one command | `keys.go`, `promote.go`, `internal/audit` |
| **Silent publication by code or crawler** | Only `pdoomctl release promote` writes `data/releases`; `pdoomctl model run` refuses an output path inside it; the API and the web app are read-only; the web app makes no runtime external requests | `cmd/pdoomctl`, `internal/api`, `apps/web/next.config.ts` |
| **Audit-log tampering** | Hash chain `sha256(prev_hash + canonical_json(event))`, genesis of 64 zeros, `Append` refuses to extend a broken chain, `Verify` checks sequence numbers, links and hashes; CI runs `pdoomctl audit verify` | `internal/audit/audit.go`, `.github/workflows/ci.yml` |
| **Non-reproducible or drifting numbers** | Deterministic model (fixed seeds, mulberry32 identical in Go and TypeScript, canonical JSON, no wall-clock reads); promotion re-runs the model and requires byte-identical output; golden fixtures shared across languages; guard test against hardcoded percentages | `internal/model`, `packages/model-core`, `apps/web/test/no-hardcoded-numbers.test.ts` |
| **Malicious submissions to the API** | JSON only (`415`), unknown fields rejected (`400`), body ≤ 64 KiB (`413`), request line ≤ 2048 bytes, control characters stripped, URL scheme restricted to http(s), per-field length caps, token-bucket rate limits (submit: burst 5, 10 per minute), same-origin or allowlisted origin for POST, append-only JSONL confined to the queue directory | `internal/api/handlers_submit.go`, `middleware.go`, `ratelimit.go`, `internal/storage` |
| **Denial of service on the API** | Rate limits per client (read: burst 120, 600 per minute; lab: 10, 30 per minute), bounded response cache (2048 bodies, 64 MB, keyed only on the query parameters a route reads), scenario-lab `samples` ≤ 50 000, server timeouts (read header 5 s, read 15 s, write 30 s, idle 60 s, 16 KiB headers), graceful shutdown | `internal/api/api.go`, `cmd/pdoom-api/main.go` |
| **Denial of service on the crawler** (decompression bombs, script-heavy pages, huge feeds) | `io.LimitReader` on the decoded stream (8 MiB + 1), single-pass HTML tokenizer linear in input, 20 000-character text cap, 2 000-sentence and 50-claim caps, run `--timeout` with context cancellation that writes nothing | `internal/fetch`, `internal/parsing`, `cmd/pdoom-ingest` |
| **Cross-site scripting and framing of the web app** | CSP `default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:; worker-src 'self' blob:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'`; `X-Content-Type-Options: nosniff`; `Referrer-Policy: no-referrer`; `X-Frame-Options: DENY`; `Permissions-Policy` denying camera, microphone, geolocation and interest-cohort; markdown rendered from repository files only, with HTML escaping; no third-party scripts | `apps/web/next.config.ts`, `apps/web/lib/markdown.ts` |
| **Framing or embedding of API responses** | `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; sandbox`, `X-Frame-Options: DENY`, nosniff, no-referrer, restrictive Permissions-Policy on every response | `internal/api/middleware.go` |
| **Misinformation via screenshots** | Every number is rendered with its status, outcome set, horizon, interval and rounding rule in the same card; the official value is displayed as "Insufficiently calibrated"; the share card carries the status and interval, never a bare number; `/text` and the exports carry the release id; the footer repeats that no official probability is published | `EstimateCard`, `opengraph-image.tsx`, `Footer` |
| **CSV formula injection in exports** | Cells beginning with `= + - @` or a tab/return are prefixed with an apostrophe and quoted | `apps/web/app/api/export/[name]/route.ts` |
| **Supply-chain compromise** | Dependencies pinned in `pnpm-lock.yaml` and `go.sum`; CI installs with `--frozen-lockfile`; `pnpm.onlyBuiltDependencies` restricts install scripts to `esbuild` and `unrs-resolver`; generated JSON Schemas are diffed in CI; no CDN or external script at runtime | `package.json`, `.github/workflows/ci.yml` |
| **Reader tracking or privacy leaks** | No analytics, no cookies, no external requests; mode preference in `localStorage` only; API access log records a salted 12-character hash of the client address (IPv6 by /64), never the raw address; `X-Forwarded-For` ignored unless `PDOOM_TRUST_PROXY=1` | `ModeProvider`, `internal/api/middleware.go` |
| **Content-safety leakage** (operational detail entering data through a source) | Category-level records only; ingestion keeps metadata for advisories, never descriptions; the Cassandra review checks every record; `content_safety_note` documents removals | [`../method/content-safety.md`](../method/content-safety.md), [`../governance/cassandra-charter.md`](../governance/cassandra-charter.md) |

## 4. Residual risks

- `RequiredApprovals` is 1 for ordinary releases; a single compromised key and operator
  machine could promote a non-heightened release. Mitigations: audit log, one-command rollback,
  the reproducibility gate (the attacker must also control the snapshot), and the two-key rule
  for every heightened-review release.
- `script-src 'unsafe-inline'` is required by the Next.js hydration model; the app includes no
  third-party scripts, and all content is repository-controlled, but the CSP does not by itself
  prevent inline script injection if a rendering bug introduced one.
- The web app's file data source trusts the local `data/` directory; the deployment must
  protect that volume ([`../operations/deployment.md`](../operations/deployment.md)).
- The snapshot's records were verified through search snippets rather than page fetches
  (`verified_search`); a mis-transcribed value is a data-integrity risk handled by the
  corrections policy, not by code.

## Not yet implemented

- Signed release archives for offline verification (today verification is: check out the
  commit, run the reproduction command, compare hashes, verify approvals with the public keys).
- Key rotation and revocation procedure beyond "add a new key, stop using the old one".
- Software bill of materials and dependency vulnerability scanning in CI.
- Tests under `tests/security` (the directory is empty); `internal/api/hardening_test.go`,
  `internal/fetch/fetch_test.go`, `internal/parsing/parsing_test.go` and `internal/robots/robots_test.go`
  cover the code-level controls.
