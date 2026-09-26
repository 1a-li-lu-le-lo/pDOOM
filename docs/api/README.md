<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# p(DOOM) public API

`cmd/pdoom-api` serves the read API described in [`api/openapi.yaml`](../../api/openapi.yaml).
It is the HTTP face of the observatory: everything it returns comes from the **current
promoted release** in `data/releases` and from the sealed data snapshot that release was
computed from.

## Run

```sh
go run ./cmd/pdoom-api --addr :8080 --data-dir data
```

| Flag | Environment | Default | Meaning |
| --- | --- | --- | --- |
| `--addr` | `PDOOM_ADDR` | `:8080` | Listen address. |
| `--data-dir` | `PDOOM_DATA_DIR` | `./data` | Data directory holding `releases/`, `snapshots/`, `review-queue/`. |
| `--method-dir` | `PDOOM_METHOD_DIR` | `<data>/../docs/method` | Markdown files served by `/v1/methodology`. |
| `--cors-origins` | `PDOOM_CORS_ORIGINS` | none | Comma-separated browser origins allowed to POST cross-origin. |
| `--trust-proxy` | `PDOOM_TRUST_PROXY=1` | off | Trust the reverse proxy: derive the client address from the right-most `X-Forwarded-For` entry and the request scheme from `X-Forwarded-Proto` (used for same-origin detection of POSTs). |
| `--routes` | | | Print the route table and exit. |

The process shuts down gracefully on `SIGINT`/`SIGTERM` (10 s drain). `GET /healthz` reports
liveness; `GET /readyz` reports `200` only once a promoted release and its snapshot are loaded.

The server reads `data/releases/CURRENT` on start and re-checks it at most once every 5 s.
A promotion performed with `pdoomctl release promote` (or a rollback) is therefore served
within a few seconds, without a restart. If a reload fails the previously loaded release keeps
being served and the failure is logged.

Only the current release and its snapshot must be valid. Any other directory under
`data/releases` that cannot be read (for example one left behind by an interrupted promotion,
or an old manifest with unknown fields) or that is not marked `published` is skipped with a
warning and never appears in `/v1/meter/history`, `/v1/releases` or `/v1/releases/{id}`.
Once a release is loaded, the release list and the release detail routes are served from
memory; they fall back to a best-effort read of the directory only while no release is loaded.

## Guarantees

* **Read-only and release-backed.** Handlers never compute or alter an estimate. Every data
  response embeds a `meta` object (`release_id`, `data_snapshot`, `data_cutoff`,
  `generated_at`, `published`, `model_versions`, `limitations`), and every estimate carries
  its `status`, `producer`, `horizon`, `outcome_set` and `conditioning`. Indexes are
  `0–100` scores with `is_probability: false` and are never converted to probabilities.
* **No official probability.** The official object has status `insufficiently_calibrated`
  and publishes no number. External aggregates, the research-mode model, the indexes and the
  editorial level are separate objects and are never blended.
* **No automatic publication.** Nothing reachable over HTTP changes a release or a snapshot.
  `POST /v1/submissions/*` appends one JSON line to `data/review-queue/submissions.jsonl`
  for human review; `POST /v1/scenario-lab/evaluate` computes a `user_scenario` result per
  request and stores nothing.
* **No administrative endpoints.** Promotion, rollback, review and reviewer keys are
  CLI-only (`pdoomctl`, ADR-004). The OpenAPI document contains no such routes and the route
  table test (`internal/api/routes_test.go`) fails if one appears without being specified.
* **Deterministic bodies.** GET responses are encoded once per release and carry a strong
  `ETag` (SHA-256 of the body) with `Cache-Control: public, max-age=60`; `If-None-Match`
  yields `304 Not Modified`. POST responses are `no-store`. The response cache is bounded
  (2048 bodies, 64 MB, least-recently-used eviction) and keyed by path plus only the query
  parameters a route reads (`tier`, `topic`, `q`, `limit`, `offset` on `/v1/sources`), so
  junk query strings cannot grow it.
* **Security headers** on every response: `X-Content-Type-Options: nosniff`,
  `Referrer-Policy: no-referrer`, `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; sandbox`,
  `Permissions-Policy` denying device features, `X-Frame-Options: DENY`.
* **CORS.** GET/HEAD are open to any origin (`Access-Control-Allow-Origin: *`). POST is
  accepted from the same origin or from origins listed in `PDOOM_CORS_ORIGINS`; other browser
  origins receive `403 origin_not_allowed`. An `https://` origin on the request's own host
  counts as same-origin even when this process sees plain HTTP behind a TLS-terminating
  reverse proxy, so a web app served from the API's host needs no allowlist entry; with
  `--trust-proxy` the scheme is taken from `X-Forwarded-Proto` instead.
* **Request limits.** Request line ≤ 2048 bytes, bodies ≤ 64 KiB (`413`), JSON only (`415`),
  unknown fields rejected (`400`), control characters stripped from submitted text.
* **Privacy.** The access log (structured `slog`, JSON on stdout) records method, path,
  status, bytes, duration, request id and a salted 12-hex-character hash of the client
  address. No raw IP address is logged. `X-Forwarded-For` is ignored unless
  `PDOOM_TRUST_PROXY=1`. IPv6 clients are identified by their `/64` prefix, IPv4 clients by
  their address, both for rate limiting and for the logged hash.
* **Errors** are always `{"error": "<code>", "detail": "<text>"}`; `detail` lists every
  validation problem found so that a client can fix them in one round trip.

## Rate limits

Token buckets per client (IPv4 address or IPv6 `/64` prefix), per route class:

| Class | Routes | Burst | Refill |
| --- | --- | --- | --- |
| read | every `GET` | 120 | 600 per minute |
| lab | `POST /v1/scenario-lab/evaluate` | 10 | 30 per minute |
| submit | `POST /v1/submissions/*` | 5 | 10 per minute |

Responses carry `RateLimit-Limit` and `RateLimit-Remaining`; an exhausted bucket returns
`429 rate_limited` with `Retry-After` in seconds.

## Routes

| Method | Path | Body |
| --- | --- | --- |
| GET | `/v1/meter` | Current release summary: manifest, official object, external aggregates, research-mode estimates, indexes, editorial level, uncertainty score, cutoff, review/publication dates, model versions, limitations. |
| GET | `/v1/meter/history` | Every promoted release with a per-release headline. |
| GET | `/v1/outcomes`, `/v1/horizons` | Static vocabularies with derived sets. |
| GET | `/v1/drivers` | Driver families, signals, observations, "why this number" rows. |
| GET | `/v1/scenarios`, `/v1/scenarios/{id}` | Category-level scenarios and edges. |
| GET | `/v1/forecasts` | Forecast records, compatibility-group aggregations, external estimates. |
| GET | `/v1/capabilities` | Benchmarks and results. |
| GET | `/v1/incidents`, `/v1/safeguards`, `/v1/definitions`, `/v1/organizations`, `/v1/actions` | Snapshot entities as promoted. |
| GET | `/v1/sources?tier=&topic=&q=&limit=&offset=` | Paginated sources (`limit` ≤ 200). |
| GET | `/v1/sources/{id}` | One source with its claims. |
| GET | `/v1/methodology`, `/v1/methodology/{slug}` | Methodology documents; content as `text/markdown`. |
| GET | `/v1/releases`, `/v1/releases/{id}` | Published release list and full release directories (unreadable or unpublished directories are omitted). |
| GET | `/v1/releases/{id}/changelog.md`, `/v1/releases/{id}/model-card.md` | The Markdown documents shipped with a published release; content as `text/markdown`. |
| GET | `/v1/snapshot` | The sealed data snapshot behind the current release, entity file by entity file (large; use the ETag). What `@pdoom/sdk`'s HTTP source reads to mirror the file source. |
| POST | `/v1/scenario-lab/evaluate` | UserScenarioParams (+ `samples` ≤ 50000, `seed`). |
| POST | `/v1/submissions/sources`, `/v1/submissions/corrections` | `{ "payload": {...}, "contact"?: "..." }`; the `202` body carries an `id` unique to that submission. |
| GET | `/healthz`, `/readyz` | Liveness, readiness. |

## Examples

```sh
# Current release summary
curl -s http://localhost:8080/v1/meter | jq '.headline'

# Conditional request (304 when unchanged)
ETAG=$(curl -sI http://localhost:8080/v1/meter | awk '/^ETag/ {print $2}' | tr -d '\r')
curl -s -o /dev/null -w '%{http_code}\n' -H "If-None-Match: $ETAG" http://localhost:8080/v1/meter

# Tier-1 sources, second page of 20
curl -s 'http://localhost:8080/v1/sources?tier=1&limit=20&offset=20' | jq '.total, (.sources | length)'

# One scenario
curl -s http://localhost:8080/v1/scenarios/S3 | jq '.scenario.name, .interventions[].id'

# A methodology document
curl -s http://localhost:8080/v1/methodology/aggregation

# Scenario Lab (result is labelled user_scenario; not an official estimate)
curl -s -X POST http://localhost:8080/v1/scenario-lab/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"horizon":"10y","capability_timeline":1,"autonomy_growth":0,"access_level":0,
       "safety_progress":-1,"governance_strength":0,"model_security":0,"open_weight_diffusion":0,
       "international_coordination":0,"incident_frequency":0,"resilience":1,"samples":5000}' \
  | jq '.label, .disclaimer, .result.outcome_estimates.P_DOOM'

# Suggest a source (enters the review queue only)
curl -s -X POST http://localhost:8080/v1/submissions/sources \
  -H 'Content-Type: application/json' \
  -d '{"payload":{"canonical_url":"https://example.org/report","title":"Example report",
       "publisher":"Example Org","date_published":"2026-05","why_relevant":"Measures a tracked signal.",
       "claimed_evidence":"Table 2 reports the value."},"contact":"reviewer@example.org"}'

# Suggest a correction
curl -s -X POST http://localhost:8080/v1/submissions/corrections \
  -H 'Content-Type: application/json' \
  -d '{"payload":{"target_id":"src-example","field":"date_published","correction":"Published in May, not June.",
       "evidence_url":"https://example.org/report"}}'
```

## Tests

`go test ./internal/api/ ./cmd/pdoom-api/` runs an end-to-end suite against a temporary data
directory built from the synthetic fixture (`internal/snapshot/fixture`): the snapshot is
written, the model run, a candidate produced, two reviewer keys generated, two approvals
signed and the candidate promoted with a heightened-review acknowledgement. Every route,
404s, filters and pagination, `ETag`/`304`, rate limiting (including IPv6 `/64` keying),
CORS and same-origin detection behind a TLS-terminating proxy, submission validation, queueing
and id uniqueness, scenario-lab labelling and slider-range rejection, live reload after a second
promotion, damaged or unpublished sibling release directories, the bounded response cache, and
the route table versus `api/openapi.yaml` are covered.
