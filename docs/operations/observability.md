<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Observability

What the Go binaries report about themselves, what an operator should watch, and what the
project deliberately does not track.

## 1. Counters and gauges (`internal/observability`)

A dependency-free registry of named counters and gauges behind a mutex, with deterministic
string and `slog` rendering so that a summary can appear in logs and golden output. Used by
`cmd/pdoom-ingest`, which prints one summary at the end of every run (`--json` for a
machine-readable object).

| Name | Kind | Meaning |
| --- | --- | --- |
| `sources_considered` | counter | Entries in `config/sources.json` looked at |
| `sources_skipped` | counter | Entries not fetched (disabled, robots status, interval, `--source` filter) |
| `fetches` | counter | HTTP fetches attempted |
| `fetch_errors` | counter | Fetches refused or failed (allowlist, SSRF guard, status, size, content type) |
| `robots_denials` | counter | Requests refused by robots policy, including fail-closed cases |
| `parser_errors` | counter | Bodies that could not be parsed |
| `documents` | counter | Normalised documents produced |
| `duplicates` | counter | Exact duplicates skipped (same canonical URL, hash or text) |
| `near_duplicates` | counter | Simhash near-duplicates queued with a cluster id |
| `candidate_claims` | counter | Rule-based candidate claims extracted |
| `queued_items` | counter | Lines appended to the review queue |
| `filtered_before_since` | counter | Items older than `--since` |
| `bytes_fetched` | gauge | Bytes read across fetches |

The registry is not exported over the network; it exists so that the run's own account of what
it did is truthful and testable (`internal/observability/observability_test.go`).

## 2. Logs

- `pdoom-api` writes a structured JSON access log to stdout via `slog`: method, path, status,
  bytes, duration, request id and a salted 12-hex-character hash of the client address (IPv6
  clients by /64 prefix). No raw IP address, no user agent string is retained beyond what the
  reverse proxy logs. Loader messages (release loaded, reload failed and kept previous) and
  server lifecycle messages (listening, shutting down) use the same logger.
- `pdoom-ingest` logs per source and per stage (`--verbose` for detail) and ends with the
  counter summary.
- `pdoomctl` prints human-readable results or `--json`; promotion and rollback are also recorded
  in the hash-chained audit log `data/audit/audit.jsonl`, which is the durable record.

## 3. Health and readiness

| Endpoint | Meaning |
| --- | --- |
| `GET /healthz` | Process is up; always `200 {"status":"ok"}` |
| `GET /readyz` | `200 {"status":"ready","release_id":"rel-…"}` once a promoted release and its snapshot are loaded; otherwise `503 {"status":"not_ready","detail":"no promoted release is loaded"}` |

The API re-checks `data/releases/CURRENT` at most every 5 s; a reload failure keeps serving the
previous release and logs the error, so readiness stays green through a bad promotion attempt.

The web app has no health endpoint of its own; `/text` returning 200 with the release id in
its lede is the equivalent check.

## 4. What to alert on

| Signal | Why |
| --- | --- |
| `/readyz` not 200 for more than the reload interval | No release is loaded; the site's footer will say "No release loaded." |
| Release id in `/readyz` differs from `data/releases/CURRENT` for longer than 10 s | The loader failed to load the new release (check logs for the reload error) |
| API `429` rate above normal, or `413`/`415`/`400` bursts on `/v1/submissions/*` | Abuse or a misbehaving client |
| `5xx` from the web server | A rendering error (the error page states that no estimate changed) |
| `pdoomctl audit verify` failing in CI | The audit chain was edited or truncated |
| `robots_denials` or `fetch_errors` rising between ingestion runs | A host changed its policy or the allowlist needs review |
| A `release.rollback` audit event | Someone withdrew a release; the changelog should explain why |
| Snapshot validation failing in CI | The committed snapshot or schemas changed inconsistently |

## 5. Deliberately not tracked

- **No analytics.** No page-view counters, no client-side beacons, no third-party scripts; the
  CSP's `connect-src 'self'` prevents them.
- **No cookies.** The only client-side state is the presentation mode in `localStorage`
  (`pdoom.mode`), which never leaves the browser.
- **No raw IP addresses** in API logs; `X-Forwarded-For` is ignored unless `PDOOM_TRUST_PROXY=1`.
- **No user identity** anywhere; submissions carry an optional free-text `contact` field only.
- **No timing of readers**: no scroll depth, dwell time or engagement metrics. Reach is measured,
  if at all, by the operator's own web-server logs, which are outside this repository.

## Not yet implemented

- A metrics endpoint (Prometheus or OpenTelemetry) for the API; the registry is process-local.
- Alerting rules as code; the table above is guidance.
- Structured logging of `pdoomctl` operations beyond the audit chain.
