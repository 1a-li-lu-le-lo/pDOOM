<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Crawler safety

`pdoom-ingest` fetches a small, allowlisted set of feeds and pages to save reviewers time. It
must be a good citizen of every host it touches and must never become a path to publication.
This page states the allowlist semantics, robots handling, politeness rules, caps, what is
never fetched, and the operator procedure for enabling a source. It is consistent with the
pipeline description in [`../operations/ingestion.md`](../operations/ingestion.md), which is the
detailed reference.

## 1. The allowlist: `config/sources.json`

The file is decoded strictly by `internal/sources` (unknown fields rejected) and validated on
load. Each entry:

| Field | Meaning | Rules enforced by the loader |
| --- | --- | --- |
| `id`, `name` | Stable identifier and display name | required |
| `kind` | Transport: `rss`, `atom`, `api`, `dataset`, `html` | must be one of these |
| `url` | The single URL fetched for this source | https only; the hostname is added to the allowlist only when `allowed` is true |
| `tier` | Default source tier (1–5) for items from this source; per-item tier is set at human review | 1–5 |
| `publisher` | Named publisher | required |
| `allowed` | Whether the source may be fetched at all | only `true` entries are fetched, and an allowed entry must pass every structural check |
| `reason` | Why the source is disabled, or what an operator must verify | free text |
| `robots_checked_at`, `robots_status` | Last robots check date and result: `allowed`, `disallowed`, `not_applicable`, `unknown` | `allowed`/`disallowed`/`not_applicable` require `robots_checked_at`; fetching requires `allowed` or `not_applicable` |
| `license` | Licence covering the retrieved text | required for an allowed entry |
| `max_items_per_run` | Items taken from one fetch | 0–500 (`MaxItemsCap`) |
| `fetch_interval_hours` | Minimum spacing between runs for this source | ≥ 1 for an allowed entry |
| `parser` | `feed`, `json`, `html_text` | `rss`/`atom` require `feed`; `html` requires `html_text`; `api` requires `feed` or `json` |
| `conflicts` | Conflict labels applied to items | `ConflictLabel` values |
| `notes` | Politeness rules and reviewer hints | free text |

The fetcher only knows the hostnames of `allowed: true` entries; a host not in the list cannot
be contacted, not even through a redirect.

### Current state

The file lists ten candidate sources: `arxiv-api-cs-ai-cy`, `metr-blog`, `epoch-ai-blog`,
`nist-news`, `uk-aisi-govuk-atom`, `oecd-ai-observatory`, `github-advisory-database`,
`ai-incident-database`, `metaculus-api`, `federal-register-ai`. Every one is `allowed: false`
with `robots_status: "unknown"`, because none of the hosts could be reached from the environment
in which the file was written. Each `reason` states what an operator must verify. Nothing is
fetched until an operator completes the procedure in section 6.

## 2. Robots handling

- `robots.txt` is fetched once per scheme://host through the same guarded client, without
  following redirects, and cached for the run.
- The agent token matched against `User-agent` lines is `pdoom-ingest`; the full header is
  `pdoom-ingest/0.1 (+https://github.com/1a-li-lu-le-lo/pdoom)`.
- 404 and 410 mean no restrictions. 401, 403, 5xx, a redirect, a parse error or a network error
  all mean **disallowed**. The crawler fails closed.
- `Crawl-delay` is honoured up to a cap of 60 s (`fetch.MaxCrawlDelay`); a larger value is
  waited for 60 s so that one host cannot stall a run for hours.
- Every redirect target is checked against the target host's robots policy as well as the
  allowlist.
- `--check-robots` prints the status per source and exits 1 if any host's file could not be
  obtained; it never writes.

## 3. Politeness

- One request per source per run.
- Per-host delay: the larger of 3 s (the arXiv API rule) and the host's `Crawl-delay`, capped at
  60 s; redirect hops wait too; waits observe the run timeout.
- Conditional requests and caching beyond the robots cache are not used; the per-source
  `fetch_interval_hours` and `max_items_per_run` bound the load instead.
- No cookies, no authentication, no API keys, no retries within a run.

## 4. Caps

| Bound | Value |
| --- | --- |
| Decoded body | 8 MiB (+1 byte to detect overflow), via `io.LimitReader`; gzip decoded only by the standard library transport |
| Redirects | 3 |
| Items per source per run | `max_items_per_run`, hard cap 500 |
| Text per document | 20 000 characters; title 500 |
| Sentences scanned for claims | 2 000; candidate claims 50; primary-source links 10 |
| Content types | `text/html`, `application/xml`, `text/xml`, `application/rss+xml`, `application/atom+xml`, `application/json`, `text/plain`; anything else or a missing `Content-Type` is refused |
| Run timeout | `--timeout`, default 10 min; on expiry the run writes nothing and exits 1 |
| Review-queue lines | 16 MiB per line on read and write; file names confined to the queue directory |

## 5. What is never fetched

- Anything over plain `http`, on a port other than 443, at an IP-literal host, or with
  credentials in the URL.
- Any host not in the allowlist, including redirect targets.
- Any host whose resolved address is loopback, private, link-local, multicast, unspecified,
  carrier-grade NAT, documentation, benchmarking, discard, ORCHID, Teredo, 6to4/NAT64 with an
  embedded private address, or malformed. The check runs inside the dialer on every hop.
- Any URL a host's `robots.txt` disallows for `pdoom-ingest`, or any host whose `robots.txt`
  cannot be obtained.
- Binary files, PDFs, images, archives (content-type allowlist).
- Anything when `--allow-network` is absent (the default is a plan-only run with no network).
- Advisory descriptions or other operational detail: the `json` parser for GitHub advisories
  keeps id, summary, severity, dates and URL only (build-spec rule 0.9).

Nothing fetched is ever sent to a language model, stored as a full body archive (only the hash,
final URL, content type, size and retrieval time are kept), or written anywhere other than
`data/review-queue/<date>-ingest.jsonl`.

## 6. Operator verification procedure (before setting `allowed: true`)

1. **Robots**: `go run ./cmd/pdoom-ingest --now <RFC3339> --allow-network --check-robots
   --source <id>`; copy the printed status into `robots_status` and today's date into
   `robots_checked_at`. Only `allowed`, or `not_applicable` for an API whose published terms
   govern programmatic access, may be enabled.
2. **Terms**: read the site or API terms. If automated access is not clearly permitted, keep
   `allowed: false` and describe the manual route in `reason`. Record rate rules in `notes`.
3. **Licence**: record the licence covering the retrieved text in `license`. Share-alike
   licences (CC BY-SA) need attribution and share-alike on derived records; note it.
4. **Tier and conflicts**: set `tier` per [`../method/source-hierarchy.md`](../method/source-hierarchy.md)
   and `conflicts` per [`../governance/conflict-of-interest.md`](../governance/conflict-of-interest.md).
5. **Kind and parser**: match the table in section 1; `dataset` entries are usually manual.
6. **Bounds**: `max_items_per_run` 1–500, `fetch_interval_hours` ≥ 1.
7. **Content safety**: if the source can carry operational detail, confirm the parser keeps
   category-level metadata only.
8. **Test**: `go test ./internal/sources` validates the file; then run a dry run
   (`--allow-network` without `--dry-run=false`) with `--source <id>` and read what would be
   queued before enabling writes.

The pipeline never edits `config/sources.json`; every change is a reviewed commit.

## 7. Observability of a run

Each run ends with counters (`sources_considered`, `sources_skipped`, `fetches`,
`fetch_errors`, `robots_denials`, `parser_errors`, `documents`, `duplicates`,
`near_duplicates`, `candidate_claims`, `queued_items`, `filtered_before_since`,
`bytes_fetched`) from `internal/observability`; `--json` prints them machine-readably. See
[`../operations/observability.md`](../operations/observability.md).

## Not yet implemented

- Conditional requests (`If-Modified-Since`/`ETag`) to reduce load further.
- A scheduler; runs are started by an operator with an explicit `--now`.
- A published crawler information page at the `+https://…` URL in the User-Agent beyond the
  repository itself.
