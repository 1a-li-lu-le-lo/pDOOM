<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# p(DOOM)

**Probability of Doom, Disempowerment, and Unrecoverable Machine-Caused Catastrophe.**
The AI Existential and Civilizational Risk Observatory. A research prototype authored by
NU Cybernetics.

p(DOOM) is a structured forecasting and evidence-synthesis system about catastrophic and
existential risk from advanced AI. It collects named sources, forecasts, benchmark results,
incidents, scenarios and safeguards into a versioned data snapshot, computes a small number
of clearly labelled quantities from that snapshot with deterministic code, and publishes them
only after signed human review. It is not a thermometer: nothing here measures "doom", no
number is presented without its outcome definition, horizon, conditioning, interval and
sources, and the official p(DOOM) value is withheld until a documented calibration process
exists. It does not predict a date, does not give advice about anyone's situation, and does
not publish operational detail about any dangerous pathway.

## Eight outputs, never blended

| Output | Object | Probability? |
| --- | --- | --- |
| A. External forecast aggregate | named surveys, tournaments and platform forecasts aggregated only inside compatibility groups (same outcome set, horizon, conditioning) | yes, from sources |
| B. p(DOOM) model estimate | research-mode experimental causal model; the **official** object is withheld | research mode only |
| C. Evidence Pressure Index | 0–100 movement relative to the baseline release | no |
| D. Capability Pressure Index | 0–100 | no |
| E. Control Strength Index | 0–100 | no |
| F. Incident Pressure Index | 0–100 | no |
| G. Uncertainty Score | 0–100 | no |
| H. Editorial risk level | rule-based label (`very_low` … `insufficient_evidence`) | no |

The definitions are in [`docs/method/definitions-of-outputs.md`](docs/method/definitions-of-outputs.md);
the reasoning is in [`docs/governance/decision-log.md`](docs/governance/decision-log.md) (ADR-001).

## The official value is withheld

In the current release line the official object has status `insufficiently_calibrated` for
every horizon and publishes no number. Long-horizon existential outcomes have never occurred
in a way that permits scoring, the research-mode parameters are documented judgments, and the
external forecasts inherit the selection and framing effects of their sources. What would
change this is written in [`docs/method/calibration.md`](docs/method/calibration.md); the
decision is ADR-002. When a value from the release is quoted anywhere, it is quoted with its
status ("research mode" or "external forecast aggregate"), its outcome set, its horizon and
its interval.

Current release: `rel-2026-09-26-001`, computed from snapshot `snap-2026-09-26-001`
(source cutoff 2026-09-26; 70 sources, 34 claims, 15 forecast records, 3 benchmarks with
7 results, 11 incidents, 18 scenarios with 30 edges, 10 driver families with 17 observations,
27 interventions, 18 organisations, 57 actions, 29 definitions). Model versions:
`pdoom-model/official-index-only@0.1.0`, `pdoom-model/external-aggregate@0.1.0`,
`pdoom-model/experimental-causal@0.1.0`, `pdoom-model/indexes@0.1.0`.

## Five modes and a plain-text route

The web app has five presentation modes: **Event Horizon**, **Orrery**, **Branching Futures**,
**Observatory** (the still, server-rendered view) and **Plain text**. Every mode reads the same
release; the immersive scenes are labelled "Conceptual risk visualization · not a simulation of
AI risk" and are progressive enhancement. `/text` is a complete, server-rendered, printable
page with every substantive fact in twenty sections and works without JavaScript; every other
route links to its matching section. Readers who prefer reduced motion get the Observatory by
default. The Event Horizon scene (WebGL) pauses when hidden, lowers its quality on slow devices
and can be skipped; the Orrery and Branching scenes are SVG. All three are conceptual and
labelled as such; the grammar is in [`docs/design/motion-semantics.md`](docs/design/motion-semantics.md).

## Repository layout

```
apps/web/                 Next.js 15 web app (App Router, React 19); reads the promoted release only
cmd/pdoomctl/             snapshot seal/validate, model run, approvals, promote, rollback, audit verify, keygen
cmd/pdoom-api/            public read API + Scenario Lab + submissions (api/openapi.yaml)
cmd/pdoom-ingest/         bounded, allowlisted crawler; writes only to the review queue
cmd/pdoom-model/          deterministic model run to stdout (writes nothing)
internal/                 Go: schema, snapshot, model, publishing, audit, review, fetch, robots, parsing, dedup, sources, claims, api, config, observability, storage
packages/schemas/         @pdoom/schemas — zod schemas, enums, brand constants, JSON Schema export
packages/model-core/      @pdoom/model-core — TypeScript port of the causal model and rounding (Scenario Lab)
packages/sdk/             @pdoom/sdk — file and HTTP data sources
packages/design-system/   tokens.css, fonts.css, tokens.json, base.css
services/mcp/             @pdoom/mcp — stdio MCP server: twelve read tools, three submission tools
skills/pdoom/             SKILL.md: how an assistant should read p(DOOM)
tools/snapshot/           build.mjs (content modules → snapshot), merge-fragments.mjs
data/                     snapshots/, releases/ (+ CURRENT), schemas/, keys/reviewers/*.pub, audit/audit.jsonl, review-queue/, candidates/
config/sources.json       allowlisted ingestion sources (all disabled until an operator verifies them)
research/                 reports that justify every record in the snapshot
docs/                     architecture, method, governance, design, accessibility, security, operations, api
tests/e2e/                Playwright configuration (desktop, mobile, reduced-motion projects)
```

The binding specification is [`docs/architecture/build-spec.md`](docs/architecture/build-spec.md);
the architecture summary is [`docs/architecture/overview.md`](docs/architecture/overview.md).

## Quick start

```sh
pnpm install                         # Node 22, pnpm 10; Go 1.24 for the Go tools
make check                           # gofmt, go build/vet/test, typecheck, lint, vitest, builds
pnpm --filter @pdoom/web dev         # web app on http://localhost:3000, reading data/releases/CURRENT
go run ./cmd/pdoom-api               # API on http://localhost:8080 (GET /v1/meter, /readyz, …)
node tools/snapshot/build.mjs        # rebuild data/snapshots/snap-2026-09-26-001 from the content modules
go run ./cmd/pdoomctl snapshot validate data/snapshots/snap-2026-09-26-001
```

The full release path (build → seal → validate → model run → diff → approve → promote → audit)
is in [`docs/operations/runbook.md`](docs/operations/runbook.md). Environment variables and
the process layout are in [`docs/operations/deployment.md`](docs/operations/deployment.md).

## Data provenance

Every source, forecast, benchmark result, incident and organisation is a real, cited record
with a canonical URL and a verification object. In `snap-2026-09-26-001`, 53 of 70 sources are
`verified_search` (a search-result snippet quoting the figure and date was checked, because
direct page fetches were blocked in the build environment) and 17 are
`verified_prior_knowledge`, which are informational only and excluded from every computation.
Human review is pending for all records. Tier 4–5, unverified and retracted sources never feed
the model. The research reports under [`research/`](research/README.md) state the evidence
and its limitations for each area; the source hierarchy is in
[`docs/method/source-hierarchy.md`](docs/method/source-hierarchy.md).

## Contributing a correction

Send a correction or suggest a source through the public API:

```sh
curl -s -X POST http://localhost:8080/v1/submissions/corrections \
  -H 'Content-Type: application/json' \
  -d '{"payload":{"target_id":"src-example","field":"date_published",
       "correction":"Published in May, not June.","evidence_url":"https://example.org/report"}}'
```

Submissions land in `data/review-queue/submissions.jsonl` and never change a published value
directly; a reviewer triages them, fixes enter a new snapshot, and a corrected release is
promoted with signed approvals. Details: [`docs/governance/corrections-policy.md`](docs/governance/corrections-policy.md).

## Governance

- [Editorial policy](docs/governance/editorial-policy.md) · [Update governance](docs/governance/update-governance.md) · [Probability change policy](docs/governance/probability-change-policy.md)
- [Conflict of interest](docs/governance/conflict-of-interest.md) · [Cassandra charter](docs/governance/cassandra-charter.md) and [review of the first release](docs/governance/cassandra-reviews/rel-2026-09-26-001.md) · [Psychological safety](docs/governance/psychological-safety.md)
- [Decision log](docs/governance/decision-log.md) · [Threat model](docs/security/threat-model.md) · [Crawler safety](docs/security/crawler-safety.md)
- [Public methodology](docs/method/model.md) · [Content safety](docs/method/content-safety.md) · [Accessibility](docs/accessibility/accessibility.md) · [API](docs/api/README.md)

Releases are immutable, promoted only by `pdoomctl release promote` after ed25519-signed
approvals, a byte-for-byte reproducibility re-run and invariant checks, and recorded in a
hash-chained audit log (`pdoomctl audit verify`).

## Licence

See [`NOTICE`](NOTICE). The licence decision is pending (ADR-008 in the decision log); third-party
data are cited per record with their own licences.

Authored by NU Cybernetics. Copyright NU Cybernetics. Research prototype.
