<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Testing strategy

Tests exist to keep the epistemic rules true, not only to keep the code running. The layers
below are ordered from the contract outward.

## 1. Contract tests (TypeScript, vitest)

Run with `pnpm test` (root `vitest.config.ts` includes `packages/*/test`, `packages/*/src`,
`services/mcp/test`, `apps/web/test`).

| Suite | What it pins |
| --- | --- |
| `packages/schemas/test/enums.test.ts` | Every enumeration of build-spec §3.3 is complete with exact string values |
| `packages/schemas/test/entities.test.ts`, `release.test.ts`, `lab.test.ts` | Fixtures parse under the zod schemas; release, delta, approval, user-scenario, submission and Cassandra finding shapes |
| `packages/schemas/test/jsonschema.test.ts` | The JSON Schema export matches the zod schemas; CI additionally regenerates `data/schemas` and fails on a diff |
| `packages/sdk/test/sdk.test.ts` | The file data source reads a release and snapshot, resolves the data directory, lists methodology documents |
| `packages/model-core/test/golden.test.ts` | Cross-language contract with Go (section 3) |
| `apps/web/test/markdown.test.ts` | The markdown renderer used for `docs/method/*.md` |
| `apps/web/test/no-hardcoded-numbers.test.ts` | The guard (section 2) |

## 2. The no-hardcoded-numbers guard

`apps/web/test/no-hardcoded-numbers.test.ts` walks `apps/web/{app,components,lib}`, drops
comment lines, strips CSS and SVG geometry (`width: "100%"`, `offset="35%"`, `r`, `cx`, …), and
fails if any digit sequence followed by a percent sign remains (`/(?<![\w$}{])\d+(\.\d+)?\s?%/`).
Template placeholders such as `${x}%` are allowed. The same test fails on the phrases
"humanity has N years left", "doom is certain" and "the machines are coming". It asserts that it
scanned more than five files so that an empty directory cannot pass silently. This is the code
form of build-spec rule 0.13 and ADR-005.

## 3. Golden tests shared between Go and TypeScript

- `internal/model/rng_test.go` pins `mulberry32` to reference values computed in JavaScript for
  seed 12345 (tolerance 1e-15) and checks determinism and uniformity.
- `internal/model/causal_golden_test.go` writes or checks `internal/model/testdata/causal-golden.json`:
  the fixture spec, the parameters (`horizon 10y`, `capability_timeline 1`, `safety_progress −1`,
  `resilience 1`, 2000 samples), the expected `UserScenarioResult` and the first five RNG draws.
  Run with `-update` to regenerate after an intentional model change (which is also a
  `model_version_changed` heightened-review trigger).
- `packages/model-core/test/golden.test.ts` reads the same file and requires: identical RNG
  stream; every outcome quantile and mean within 0.01 of Go; factor p50s within 1e-6; the label
  `user_scenario`; rejection of out-of-range sliders and unknown horizons; rounding identical to
  Go (`roundForDisplay(0.1273, 5) === "15%"`, `(0.1273, 2) === "12%"`, `(0.004, 1) === "<1%"`).
- `internal/model/testdata/golden/snap-1999-01-01-001.json` is the full `Run` golden on the
  synthetic fixture snapshot (`internal/snapshot/fixture`).

## 4. Go tests

`go test ./...` (also `make go-check`, which adds `gofmt -l`, `go build` and `go vet`). Every
Go package has tests:

| Package | Focus |
| --- | --- |
| `internal/schema` | Canonical JSON, hashing, enum validity |
| `internal/snapshot` | Loading with hash verification, validation errors and warnings, sealing |
| `internal/model` | Run on the fixture snapshot, invariants, aggregation, indexes, delta, rounding, RNG, causal golden |
| `internal/publishing` | Candidate writing, manifest hash, approvals, promotion gates (including reproducibility and heightened review), rollback, listing |
| `internal/audit` | Chain append, verify, tamper detection |
| `internal/api` (+ `cmd/pdoom-api`) | End-to-end against a temporary data directory: snapshot written, model run, candidate promoted with two keys and an acknowledgement; every route, 404s, filters and pagination, ETag/304, rate limiting including IPv6 /64 keying, CORS and same-origin behind a TLS-terminating proxy, submission validation and queueing, scenario-lab labelling and slider ranges, live reload after a second promotion, damaged sibling release directories, the bounded cache, and the route table versus `api/openapi.yaml` (`routes_test.go`); `hardening_test.go` and `sanitize_test.go` cover limits and control-character stripping |
| `internal/fetch`, `internal/robots` | SSRF guard, redirect policy, size and content-type caps, robots fail-closed, crawl-delay cap |
| `internal/parsing`, `internal/dedup`, `internal/claims`, `internal/review`, `internal/storage`, `internal/sources` | Parsers (fuzz-tested, hostile markup degrades to text; injection payload stays inert), canonical URLs and simhash, rule-based claims, queue append-only semantics, config validation |
| `internal/observability` | Registry rendering |
| `cmd/pdoom-ingest` | Flags, planning without network, pipeline behaviour |

## 5. End-to-end (Playwright)

`tests/e2e/playwright.config.ts` starts the built web app (`pnpm --filter @pdoom/web start -p
3117`, or `PDOOM_E2E_PORT`), uses the Chromium at `PDOOM_CHROMIUM` or
`/opt/pw-browsers/chromium`, and defines three projects:

| Project | Device | Purpose |
| --- | --- | --- |
| `desktop` | Desktop Chrome | Routes, mode switching, share links |
| `mobile` | Pixel 7 | Reflow, header and mode switcher at phone width |
| `reduced-motion` | Desktop Chrome with `reducedMotion: "reduce"` | Observatory default, static disk, zero durations |

`@axe-core/playwright` is installed for automated accessibility checks. The build specification
names the required specs: routes render with an `<h1>`, mode switching, reduced motion, WebGL
failure fallback, `/text` without JavaScript, axe on key pages. Run with `pnpm e2e` or
`make e2e`.

## 6. `make check` and CI

`make check` = `go-check` (gofmt, build, vet, test) + `ts-check` (typecheck, lint, vitest) +
`build` (all workspace builds, including `next build`).

`.github/workflows/ci.yml` runs three jobs on pushes to `main` and on pull requests:

| Job | Steps |
| --- | --- |
| `typescript` | `pnpm install --frozen-lockfile`; regenerate JSON Schemas and `git diff --exit-code -- data/schemas`; `pnpm typecheck`; `pnpm lint`; `pnpm test`; `pnpm --filter @pdoom/web build` |
| `go` | `gofmt -l` must be empty; `go build`, `go vet`, `go test`; `pdoomctl snapshot validate` on the newest committed snapshot; `pdoomctl audit verify` on the committed audit log |
| `e2e` (needs `typescript`) | install Chromium with dependencies, build the web app, `PDOOM_CHROMIUM= pnpm e2e` |

Node 22 and Go 1.24 with caching; `permissions: contents: read`.

## 7. What a change must prove

| Change | Required evidence |
| --- | --- |
| Model code or `model_spec` parameters | Go tests pass; golden fixtures regenerated deliberately (`-update`) and the TypeScript golden still within tolerance; the candidate's `delta.json` shows `model_version_changed` |
| Schema change | `pnpm build:jsonschema` committed; Go structs updated; fixtures parse; snapshot validates |
| New snapshot | `pdoomctl snapshot validate` clean of errors; research report updated |
| Web copy or component | Guard test passes; page has one `<h1>`, a "Plain text" link and no hover-only content; `/text` carries the same substance |
| API route | Added to `api/openapi.yaml` (the route table test fails otherwise) and to `docs/api/README.md`; read-only unless it is one of the two submission routes or the lab |
| Ingestion source | `go test ./internal/sources`; dry run reviewed ([`../security/crawler-safety.md`](../security/crawler-safety.md)) |

## Not yet implemented

- End-to-end specs: `tests/e2e` contains only `playwright.config.ts`; the `e2e` CI job will
  report that no tests were found until specs are added.
- `tests/accessibility`, `tests/security`, `tests/property`, `tests/visual` and
  `tests/performance` are empty directories.
- Visual regression baselines and a performance budget for the immersive scenes (which are
  placeholders).
- Tests for `services/mcp` (the server is a placeholder; `vitest` is configured with
  `passWithNoTests`).
