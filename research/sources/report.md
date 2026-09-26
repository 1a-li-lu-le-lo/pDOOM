<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Sources: research report

This report describes the 70 source records in `data/snapshots/snap-2026-09-26-001/sources.json`
(authored in `tools/snapshot/content/sources.mjs`) as they are actually recorded: their tiers,
verification statuses, model-use statuses, conflicts, types and licences. It is the ledger
behind `/evidence`, `/text#sources` and `sources.csv`.

## 1. Tier distribution

Tiers follow [`../../docs/method/source-hierarchy.md`](../../docs/method/source-hierarchy.md).

| Tier | Count | What is in it |
| --- | --- | --- |
| 1 | 33 | Primary and authoritative: the JAIR/arXiv survey paper, METR's evaluation posts and living page, both International AI Safety Reports, the SWE-bench and OSWorld benchmark sites, NVD's CVE record, the github-mcp-server issue, NSA and NIST publications, the MCP registry and specification, three developer safety frameworks, two developer disclosures (espionage campaign, agentic misalignment), the AISI incident report, the EU, California, New York and Korean legal texts or official summaries, AISI Inspect, the OECD incident definitions, the Science consensus paper, the OpenAI Charter, and the Der Kiureghian & Ditlevsen paper |
| 2 | 26 | Independent technical analysis: AI Impacts surveys, the XPT paper, Ord, Carlsmith, Metaculus question pages, Epoch AI, RAND, JFrog, three Cloud Security Alliance notes, two AI Incident Database records, the Hendrycks, Critch & Russell, Kulveit and Slattery taxonomies, Bostrom, Morris et al., the AI Control and Safety Cases papers, Sastry et al., alignment faking, OWASP GenAI |
| 3 | 7 | Journalism and third-party trackers: the Metaculus Substack analysis, the Steel.dev SWE-bench tracker, the BenchLM OSWorld tracker, The Hacker News, The Register, CBC News, SSTI |
| 4 | 3 | Commentary: the 80,000 Hours podcast episode, the EA Forum time-horizon estimate, Simon Willison's weblog post |
| 5 | 1 | Unverified signal: METR's post on X about the Claude Opus 4.5 time horizon |

Tier 4 sources are `informational`; the tier 5 source is `excluded`. The validator refuses tier
4–5 sources that are `eligible` or `used`.

## 2. Verification statuses as recorded

| Status | Count | Method text | Model use |
| --- | --- | --- | --- |
| `verified_search` | 53 | "WebSearch result snippet quoting the figure/date; page fetch blocked by egress policy", checked 2026-09-26 | 49 `eligible`, 3 `informational` (the tier 4 items), 1 `excluded` (tier 5) |
| `verified_prior_knowledge` | 17 | "prior knowledge, not re-fetched"; note "Informational only; excluded from model computations until fetched and reviewed." | 17 `informational` |
| `verified_fetch` | 0 | — | — |
| `unverified` | 0 | — | — |

The 17 prior-knowledge records are reference documents whose existence is not in doubt but whose
current page was not checked: `src-swebench-official`, `src-osworld-official`,
`src-mcp-specification`, `src-eu-ai-act-oj`, `src-nist-ai-rmf-100-1`, `src-nist-ai-600-1`,
`src-oecd-2024-defining-ai-incidents`, `src-bengio-2024-science-managing-extreme-risks`,
`src-bostrom-2014-superintelligence`, `src-morris-2023-levels-of-agi`, `src-openai-charter`,
`src-der-kiureghian-2009-aleatory-epistemic`, `src-greenblatt-2023-ai-control`,
`src-clymer-2024-safety-cases`, `src-sastry-2024-computing-power-governance`,
`src-anthropic-2024-alignment-faking`, `src-owasp-genai-security`. They are cited by definitions,
interventions, organisations and actions (all `informational` entities) and by one benchmark
result that is itself `informational`. No driver observation, forecast, incident or benchmark
relies on them.

All 70 records have `human_review_status: pending`, `retraction_status: none`,
`content_hash: null` and `archive_reference: null` (no body was archived because nothing was
fetched). `robots_status` is `not_applicable` for the 12 arXiv-hosted or DOI-resolved records and
`unknown` for the other 58.

## 3. Types, conflicts and licences

Source types: paper 14, government 9, evaluation 8, standard 7, journalism 4, review_article 4,
incident_report 3, survey 3, safety_framework 3, forecast_platform 3, code 2,
company_disclosure 2, dataset 2, blog 2, other 2, podcast 1, social 1.

Conflict labels: `none_known` 45, `government_policy_context` 14 (with `jurisdiction` set:
international, EU, US, UK, US-CA, US-NY, KR), `developer_self_report` 8 (Anthropic ×4, OpenAI ×2,
Google DeepMind ×1, Google DeepMind authors' paper ×1), `funder_relationship` 1 (Carlsmith,
Open Philanthropy), `advocacy_context` 1 (Hendrycks et al., Center for AI Safety),
`commercial_interest` 1 (JFrog). Every developer self-report carries a `limitations` or
`counterevidence` note where the record is used as evidence (for example the espionage
disclosure: "independent verification limited; some analysts questioned the autonomy framing").

Licences are recorded where known: CC BY 4.0 for arXiv preprints, CC BY-SA 4.0 for AI Incident
Database records and OWASP, public domain for US Government works (NVD, NSA, NIST), MIT for the
MCP specification and registry and for Inspect, EU public documents for the Official Journal.
Most journalism, platform and organisation pages have `license: null`.

## 4. Coverage by topic

Forecast sources 9 ([`../forecasts/report.md`](../forecasts/report.md)); capability sources
including benchmarks and trackers 13 ([`../capabilities/report.md`](../capabilities/report.md));
security and agent-infrastructure sources 21 (CVE, MCP registry and specification, CSA, NSA,
Microsoft guidance via THN, developer disclosures, evaluation-behaviour reports, incident records)
([`../incidents/report.md`](../incidents/report.md)); governance sources 13 (frameworks, statutes,
institutes, standards); scenario and definition sources 14
([`../scenarios/report.md`](../scenarios/report.md), [`../definitions/report.md`](../definitions/report.md)).
The 34 claims in `claims.json` attach to 20 of the 70 sources through `claim_ids`.

## 5. Duplicate and merge status

`duplicate_group` is `null` on every record. `research/scenarios/fragments/sources.json` holds
43 additional or overlapping records under different ids (15 match a snapshot record by
canonical URL); they are not merged ([`../scenarios/report.md`](../scenarios/report.md) §5).

## Limitations

- No record is `verified_fetch`. A `verified_search` record has had its figure and date seen in
  a search snippet, not in the document; transcription and context errors are possible and the
  corrections route exists for them.
- 17 records are prior knowledge and are excluded from all computation; the entities that cite
  them are informational, but readers should know that those citations were not re-checked.
- `date_published` is null for `src-aiimpacts-2024-espai` (report dated September 2026, day not
  confirmed) and approximate for several records (month-level dates recorded as the first of the
  month, noted in `limitations`).
- `content_hash` and `archive_reference` are empty everywhere; the archive metadata that
  `pdoom-ingest` would record does not exist for hand-authored records.
- Tier is assigned by document type per the hierarchy; a tier 1 developer disclosure is still a
  self-report, which the conflict label, not the tier, conveys.

## Open questions

- Should a second snapshot require `verified_fetch` for every `eligible` record, dropping the
  `verified_search` allowance that this snapshot depends on?
- Should third-party benchmark trackers (tier 3) be allowed to feed capability signals at all, or
  only the maintainers' own leaderboards?
- How should platform-authored analyses (the Metaculus Substack post, tier 3) be tiered when they
  are the only dated statement of a platform value?
