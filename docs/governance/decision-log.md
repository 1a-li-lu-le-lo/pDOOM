<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Decision log

Architecture decision records for p(DOOM). Each record states its status, the context, the
decision, the consequences and the alternatives considered. Records are appended, not edited;
a reversal is a new record that supersedes the old one.

## ADR-001: Eight separate outputs, never blended

**Status:** Accepted (build-spec rules 0.2–0.5).

**Context.** Public discussion of "p(doom)" collapses many different things into one number:
survey medians for "extremely bad" outcomes, tournament questions about extinction by 2100,
individual judgements about existential catastrophe within a century, index-like impressions of
capability growth, and editorial mood. These have different outcome sets, horizons, conditioning
and populations, and their spread is mostly explained by those definitions.

**Decision.** The release publishes eight objects as separate records: (A) external forecast
aggregate, (B) model estimate with the official object withheld, (C) Evidence Pressure Index,
(D) Capability Pressure Index, (E) Control Strength Index, (F) Incident Pressure Index,
(G) Uncertainty Score, (H) editorial risk level. Each has its own producer, status and method
reference. Indexes carry `is_probability: false` and are never converted into probabilities.
Outcomes O0–O8 are never silently combined; the combined set O3–O8 appears only beside its
decomposition (O3, O4+O5, O6, O7).

**Consequences.** The meter shows several labelled quantities rather than one; the invariants
fail a candidate that marks an index as a probability; the API and the MCP skill require every
number to travel with its status, outcome set and horizon; the site is harder to quote in one
line, which is intended.

**Alternatives considered.** A single blended "p(DOOM) score" (rejected: pseudo-precision and
definitional confusion); publishing only the external aggregate (rejected: hides the
capability, control and incident evidence and inherits survey framing effects); publishing
only indexes (rejected: readers would convert them into probabilities themselves).

## ADR-002: Official p(DOOM) value withheld as `insufficiently_calibrated` in the v0 release line

**Status:** Accepted (build-spec rule 0.2; referenced by [`../method/model.md`](../method/model.md)).

**Context.** Long-horizon existential outcomes have not occurred in a way that permits scoring.
The research-mode parameters are documented judgments; the external forecasts inherit the
selection, framing and population effects of their sources; compatibility groups in the first
snapshot have two members each. No documented mapping exists from proxy calibration to
long-horizon estimates.

**Decision.** The official object (`est-official-P_DOOM-<horizon>`, producer
`pdoom-model/official-index-only@0.1.0`) is present for every horizon with status
`insufficiently_calibrated`, `quantiles: null`, `mean: null`, display "Insufficiently
calibrated" and a conditioning text that names this record. The research-mode decomposition and
the external aggregates are published beside it under their own statuses. The invariants refuse
an official object with quantiles.

**Consequences.** The meter headline is a status, not a number; the share card carries the
status and the research-mode interval; readers who want a number are shown which number they are
getting and under what assumptions. What would allow an official value is written in
[`../method/calibration.md`](../method/calibration.md): a documented mapping from scored proxy
performance to long-horizon estimates that survives adversarial review, and stable groups with
more than a handful of members.

**Alternatives considered.** Publishing the research-mode median as official (rejected: it is a
judgment model); publishing the external median as official (rejected: two-member groups and
incompatible wordings); publishing a range only (rejected: a range without calibration is still
pseudo-precision at its ends).

## ADR-003: Signed human approvals and a byte-for-byte reproducibility gate before promotion

**Status:** Accepted (build-spec rule 0.7; `internal/publishing/promote.go`).

**Context.** An observatory that ingests web content must not let that content, or a code
change, alter what the public sees without a person taking responsibility. It must also be
possible for anyone to confirm that the published numbers are what the committed code and data
produce.

**Decision.** Releases are created only by `pdoomctl release promote`, which verifies ed25519
approvals over the manifest hash against committed public keys, requires two distinct reviewers
plus an acknowledgement when heightened-review triggers exist, re-runs the model on the sealed
snapshot and requires the six data files to be byte-identical, checks the invariants, and
appends a hash-chained audit event. Promotion, rollback and approvals are CLI-only; the API and
the web app have no administrative surface. Deterministic code paths read no clock and use only
seeded randomness.

**Consequences.** Every release carries its reproduction command, code commit, manifest hash and
approvals; the audit chain can be verified in CI; a change that is not reproducible cannot ship;
operators must supply timestamps explicitly; the prototype's single required approval
(`RequiredApprovals = 1`) is a known weakness mitigated by the two-reviewer rule for heightened
review.

**Alternatives considered.** Continuous deployment from the ingestion pipeline (rejected: rule
0.7); a web admin panel with logins (rejected: widens the attack surface and puts publication
behind a password rather than a key); signing tarballs instead of manifests (deferred; see
[`../security/threat-model.md`](../security/threat-model.md)).

## ADR-004: Compatibility groups: forecasts aggregated only within the same outcome set and horizon

**Status:** Accepted (build-spec rule 0.5; `internal/snapshot/validate.go`, `internal/model/aggregate.go`).

**Context.** Surveys ask about "extremely bad" long-run outcomes, tournaments ask about
extinction by 2100 with an operational definition, platform questions differ again, and some
questions are conditional. Averaging across them hides the definitions that explain most of the
spread.

**Decision.** A forecast may be aggregated only inside a declared compatibility group
(`forecast.group_id`) whose members share outcome set, horizon and conditioning, with comparable
wording. The validator rejects a group that mixes outcome sets or horizons and warns when
conditioning text differs. A group needs at least two members; singletons and conditional
questions are shown individually. Every method's value is published, the unweighted median is
preferred, weighted methods cap any member at 35% of the weight, and the original wording is
shown beside every transformed value. Groups are documented in
[`../../research/forecasts/compatibility-groups.md`](../../research/forecasts/compatibility-groups.md).

**Consequences.** The first release has five groups of two members and five singletons; there
is no single "external p(DOOM)"; leave-one-source-out sensitivity runs expose how fragile each
group is; adding a forecast source is a heightened-review trigger.

**Alternatives considered.** Mapping every question onto the combined O3–O8 set and pooling
(rejected: the mapping is the contested part); pooling by population instead of by question
(rejected: population is a dimension inside a group, published as `equal_weight_by_population`);
using only one preferred survey (rejected: hides disagreement).

## ADR-005: No hardcoded percentages in the web source; every number flows from the signed release

**Status:** Accepted (build-spec rule 0.13; `apps/web/test/no-hardcoded-numbers.test.ts`).

**Context.** The easiest way for a site about probabilities to drift from its own data is a
literal in a component, a default in a chart, or a number in a heading. Once present it survives
releases silently.

**Decision.** No probability, percentage or index value may be written as a literal in
`apps/web` (app, components, lib) or in the service and SDK sources. A vitest guard scans every
`.ts`/`.tsx`/`.md` file, strips CSS and SVG geometry, and fails the build on any digit sequence
followed by a percent sign, and on a fixed list of panic phrasings. All displayed numbers are read
from the promoted release through `@pdoom/sdk` or computed by `@pdoom/model-core` from snapshot
data.

**Consequences.** Copy that needs an example number must read it from the release; chart axes
compute their ticks from data; the guard is a blunt instrument and rejects some innocent text,
which is accepted.

**Alternatives considered.** Code review alone (rejected: does not scale); a lint rule for numeric
literals in JSX (rejected: too noisy for geometry); allowing literals in documentation pages
(rejected: documentation pages are where stale numbers live longest).

## ADR-006: Retrieved web content is inert data; the ingestion pipeline writes only to the review queue

**Status:** Accepted (build-spec rules 0.7, 0.12; `cmd/pdoom-ingest`, `internal/{fetch,robots,parsing,review,storage}`).

**Context.** Anything fetched from the web may contain instructions aimed at automated systems,
may be wrong, and may be hostile. The pipeline must save reviewers time without becoming a path
to publication.

**Decision.** `pdoom-ingest` fetches only allowlisted `https` hosts through `SafeClient`
(SSRF guard, robots.txt fail-closed, size and content-type caps), parses to text with no field
that could express a tier, an allowlist decision or an instruction, extracts rule-based candidate
claims marked `excluded`, and appends canonical JSON lines to `data/review-queue/<date>-ingest.jsonl`.
It never writes to snapshots, releases, configuration or keys, never calls a language model, and
runs in dry-run mode by default. Public API submissions follow the same rule.

**Consequences.** Every data record still has to be written by a person into the research
fragments with a verification record; the queue can grow without consequence; a prompt-injection
payload in a feed is stored as text and proven inert by a test in `internal/parsing`.

**Alternatives considered.** LLM-assisted classification in the pipeline (rejected, explicitly
off: `claims.LLMClassificationEnabled = false`); auto-creating source records from feeds
(rejected: rule 0.6 requires verification); a database-backed queue (deferred; the JSONL store is
append-only and confined to the queue directory).

## ADR-007: Immersive modes are conceptual visualisations labelled as such, with a plain-text route as a first-class peer

**Status:** Accepted (build-spec rules 0.11, 0.14; `apps/web/components/mode`, `apps/web/app/text`).

**Context.** The product brief calls for beautiful, shareable scenes (Event Horizon, Orrery,
Branching Futures). A scene can be read as a simulation of the risk, and 3D content is invisible
to many readers and assistive technologies.

**Decision.** Five modes exist: `event-horizon`, `orrery`, `branching`, `observatory`, `text`.
Every scene is labelled "Conceptual risk visualization · not a simulation of AI risk"; the visual
grammar is fixed and documented (ring thickness = interval width, brightness = evidence pressure,
orbit dots = sources, arcs = safeguards). The static SVG disk is the server-rendered first frame,
the WebGL fallback and the reduced-motion default. `/text` is server-rendered, works without
JavaScript and contains every substantive fact (20 sections). The mode switcher with the plain-text
link is in the first viewport of every page, and every route header links to the matching `/text`
section.

**Consequences.** Nothing substantive may live only in a scene; scenes read the same release data
as the meter; reduced-motion users get the Observatory by default; the site is complete without
the scenes, which load after the server-rendered meter and only where WebGL and motion are allowed.

**Alternatives considered.** A scene-first site with a text "accessibility page" (rejected: the
text page would decay); no scenes at all (rejected: clarity and craft are how the site is shared);
letting scene parameters be tuned for drama (rejected: the grammar must map to data).

## ADR-008: Licence decision pending

**Status:** Proposed.

**Context.** `NOTICE` records the copyright of NU Cybernetics and points here for the licence
decision. Third-party data are cited per record with their own licences (`license` field in
`sources.json`; for example CC BY 4.0 for arXiv preprints, CC BY-SA 4.0 for AI Incident Database
records, public domain for US Government works, MIT for Inspect and the MCP Registry).

**Decision.** Not yet taken. The recommendation carried in `NOTICE` is Apache-2.0 for code,
CC BY 4.0 for research and methodology text, with data-source licences preserved per record. Until
a licence file exists, all rights are reserved as `NOTICE` states.

**Consequences.** External contributions cannot yet be accepted under a defined licence; share-alike
sources (CC BY-SA) constrain how derived records may be licensed and are flagged in the ingestion
checklist.

**Alternatives considered.** A single permissive licence for everything (open: data-source
licences may not permit it); a source-available licence for code (open).

## ADR-009: Next.js App Router with server rendering for the web surface

- **Status:** accepted
- **Context:** The observatory must be readable without JavaScript (the `/text` route), render every number from the signed release on the server, and still host three optional immersive modes.
- **Decision:** `apps/web` uses Next.js 15 App Router with React server components for every route; pages render per request so a promotion or rollback is served without a rebuild; immersive scenes load only on the client, after the server-rendered meter, behind WebGL and reduced-motion checks.
- **Consequences:** One framework covers static-like pages, per-request data and client scenes; the Content Security Policy in `apps/web/next.config.ts` allows no external origin. The trade-off is a Node server rather than a static export.
- **Alternatives considered:** A static-site generator (rejected: rollback would require a rebuild); a single-page application (rejected: the plain-text route and no-JavaScript reading are first-class requirements).

## ADR-010: PostgreSQL reference schema, files as the system of record

- **Status:** accepted
- **Context:** The build specification asks for a relational reference schema, while the prototype's system of record is the sealed snapshot and release directories under `data/` with hash-chained audit records.
- **Decision:** `db/migrations/` carries a reference PostgreSQL schema that mirrors the snapshot and release entities for teams that need to query the data relationally. The prototype does not run a database: the API, web app and MCP server read the release files, and every mutation path is the CLI.
- **Consequences:** The schema is documentation with teeth (it can be applied and loaded from the JSON files) but nothing in the running system depends on it; keeping it in step with `packages/schemas` is a review-time duty.
- **Alternatives considered:** Making PostgreSQL the system of record (rejected for v0: it would weaken byte-for-byte reproducibility and the sealed-directory model); no schema at all (rejected: the specification requires one).

