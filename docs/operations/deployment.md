<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Deployment

The prototype is two long-running processes and one data volume: a Node.js server for the
Next.js web app, the Go `pdoom-api` binary, and the `data/` directory they both read. Releases
are produced offline by `pdoomctl` and appear on the volume; nothing in production writes to
`data/releases`.

## 1. Processes

| Process | Build | Run | Listens |
| --- | --- | --- | --- |
| Web (`apps/web`) | `pnpm install --frozen-lockfile && pnpm --filter @pdoom/web build` | `pnpm --filter @pdoom/web start -- -p 3000` (Next.js node server; `output` is not `standalone`, `outputFileTracingRoot` is the repository root because the app imports workspace packages) | `:3000` by default |
| API (`cmd/pdoom-api`) | `go build -o bin/pdoom-api ./cmd/pdoom-api` | `bin/pdoom-api --addr :8080 --data-dir /srv/pdoom/data` | `:8080` by default |
| Tooling (`cmd/pdoomctl`, `cmd/pdoom-model`, `cmd/pdoom-ingest`) | `go build ./...` | Operator use only ([`runbook.md`](runbook.md), [`ingestion.md`](ingestion.md)) | none |

The web app reads the data directory directly through `@pdoom/sdk`'s file data source; it does
not call the API at runtime. The API is for external readers and for the two submission routes.
Both processes re-read `data/releases/CURRENT` at most every 5 s, so a promotion is served
without restarts. Both shut down gracefully on `SIGINT`/`SIGTERM` (the API drains for up to
10 s).

## 2. Environment variables

| Variable | Read by | Meaning | Default |
| --- | --- | --- | --- |
| `PDOOM_DATA_DIR` | `@pdoom/sdk` (web, MCP), all Go binaries (`internal/config`) | The data directory containing `releases/`, `snapshots/`, `schemas/`, `keys/`, `audit/`, `review-queue/` | SDK: walk up from the working directory to a `data/` folder holding `releases/CURRENT` or `snapshots/`; Go: `./data` |
| `PDOOM_DOCS_DIR` | `@pdoom/sdk` | Directory whose `method/` sub-folder holds the methodology markdown served at `/method/[slug]` | `<data>/../docs` |
| `PDOOM_PUBLIC_URL` | `apps/web/app/layout.tsx` | `metadataBase` for absolute Open Graph and share-card URLs | `http://localhost:3000` |
| `PDOOM_API_URL` | `@pdoom/sdk` (`createDefaultDataSource`) | HTTP data source for consumers without a data directory (the MCP server, once implemented); the web app uses the file source | unset |
| `PDOOM_ADDR` | `pdoom-api` | Listen address | `:8080` |
| `PDOOM_METHOD_DIR` | `pdoom-api` | Markdown served by `/v1/methodology` | `<data>/../docs/method` |
| `PDOOM_CORS_ORIGINS` | `pdoom-api` | Comma-separated browser origins allowed to POST cross-origin | none (same-origin only) |
| `PDOOM_TRUST_PROXY` | `pdoom-api` | `1` to take the client address from the right-most `X-Forwarded-For` and the scheme from `X-Forwarded-Proto` | off |
| `PDOOM_SCHEMA_DIR` | Go binaries | JSON Schema directory for validation | `<data>/schemas` |
| `PDOOM_E2E_PORT`, `PDOOM_CHROMIUM` | Playwright config | Test server port and Chromium path | `3117`, `/opt/pw-browsers/chromium` |

Flags win over environment variables for the Go binaries. There are no secrets in the
environment: the servers hold no keys and no credentials.

## 3. Data volume layout

```
/srv/pdoom/data/
  releases/
    CURRENT                      text file: the current release id (written only by pdoomctl)
    rel-2026-09-26-001/          immutable release directory (manifest, estimates, indexes, …, approvals, changelog, model card)
  snapshots/
    snap-2026-09-26-001/         sealed snapshot the release was computed from (needed at runtime for entity routes)
  schemas/                       generated JSON Schemas (needed by pdoomctl validation, not by the servers)
  keys/reviewers/*.pub           public keys only
  audit/audit.jsonl              hash-chained log
  review-queue/                  submissions.jsonl written by the API; <date>-ingest.jsonl by pdoom-ingest
  candidates/                    operator workspace; not served
```

In the repository this is `data/`; in production copy the committed `releases/`, `snapshots/`
and `schemas/` directories (they are content-addressed and small), keep `review-queue/`
writable by the API process only, and keep `candidates/` and private keys off the server. A
promotion is deployed by copying the new release directory and snapshot, then replacing
`CURRENT` atomically (write to a temporary file and rename). Both servers pick it up within 5 s.

Back up `data/` as a whole; `audit.jsonl` and `review-queue/` are the only files that grow in
production.

## 4. Security headers and the CSP

The web app sets these headers on every route (`apps/web/next.config.ts`):

```
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';
  img-src 'self' data: blob:; font-src 'self' data:; worker-src 'self' blob:; connect-src 'self';
  frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'
X-Content-Type-Options: nosniff
Referrer-Policy: no-referrer
X-Frame-Options: DENY
Permissions-Policy: camera=(), microphone=(), geolocation=(), interest-cohort=()
```

`poweredByHeader` is off. The API sets `Content-Security-Policy: default-src 'none';
frame-ancestors 'none'; sandbox`, nosniff, no-referrer, `X-Frame-Options: DENY` and a
Permissions-Policy denying device features on every response, plus `ETag` and
`Cache-Control: public, max-age=60` on GET bodies (`no-store` on POST). A reverse proxy in front
of both should terminate TLS and may add `Strict-Transport-Security`; do not let it strip the
CSP.

## 5. No CDN dependencies

Nothing is loaded from a third party at runtime: no font CDN (the font stacks fall back to system
faces; `font-src 'self' data:`), no analytics, no script CDN, no external images. The web app's
`connect-src 'self'` means the browser talks only to the origin that served the page. Static
assets are served by the Next.js server itself; a CDN may cache them in front, but the site does
not depend on one.

## 6. Reverse proxy notes

- Route `/` to the web app and `/v1/`, `/healthz`, `/readyz` to the API, or serve them on
  separate hostnames. When the API is on the same host as the web app, same-origin POSTs need no
  `PDOOM_CORS_ORIGINS` entry; an `https://` origin on the request's own host counts as
  same-origin even when the API sees plain HTTP behind the proxy.
- Set `PDOOM_TRUST_PROXY=1` only when the proxy overwrites `X-Forwarded-For`; otherwise rate
  limits and the hashed client id will key on the proxy address.
- Use `/readyz` for the API and `GET /text` for the web app as health checks.

## 7. Upgrading

1. Build the new code (`make check` green).
2. Deploy the binaries and the Next.js build.
3. If a new release exists, copy its snapshot and release directory, then replace `CURRENT`.
4. Confirm `/readyz` reports the expected `release_id` and `/text` shows it in the lede.
5. Verify the audit chain on the deployed copy: `pdoomctl audit verify /srv/pdoom/data/audit/audit.jsonl`.

## Not yet implemented

- Container images, a compose file or service manifests; the layout above is the contract they
  would implement.
- A `standalone` Next.js output; the server currently runs from the workspace with
  `node_modules` present.
- The MCP server as a deployable (`services/mcp` is a placeholder).
- A reference PostgreSQL deployment: `db/migrations` is empty; the prototype runs entirely
  from files.
