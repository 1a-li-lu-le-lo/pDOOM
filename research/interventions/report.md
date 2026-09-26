<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Interventions, organisations and actions: research report

This report justifies the 27 interventions (`I01`–`I27`), 18 organisations and 57 actions in
`snap-2026-09-26-001` (`interventions.json`, `organizations.json`, `actions.json`, authored in
`tools/snapshot/content/interventions.mjs`). They are the content of `/safeguards`, `/act` and
`/text#safeguards`–`#act`, and the safeguard arcs of the stage. Effect sizes are qualitative only
(`effect_size: unknown` on every record) and none of these records feeds a computation
(`model_use_status: informational`).

## 1. Interventions

| Id | Name | Category | Targets | Evidence | Cost | Time | Key sources |
| --- | --- | --- | --- | --- | --- | --- | --- |
| I01 | Better dangerous-capability evaluations | technical | S1 S2 S4 S6 S14 | moderate | moderate | months | RSP, Preparedness v2, FSF v3, Inspect, IASR 2026 |
| I02 | Third-party audits | organizational | S15 S14 S1 | weak | moderate | 1-2y | SB 53, EU GPAI guidelines |
| I03 | Mandatory incident reporting | national | S15 S3 S4 | moderate | low | months | EU GPAI guidelines, SB 53, OECD definitions |
| I04 | Strong model-weight security | technical | S4 S6 S15 | moderate | high | 1-2y | RAND 2024 |
| I05 | Compute monitoring | international | S14 S15 S4 | weak | moderate | 1-2y | Sastry et al., Epoch |
| I06 | Sandboxing | technical | S1 S3 S5 | moderate | moderate | months | AISI incident report, NSA guidance |
| I07 | Least-privilege tool access | technical | S3 S5 S1 | moderate | low | months | github-mcp-server #844, AIID 1152, NSA, OWASP |
| I08 | Deterministic agent policies | technical | S2 S3 S11 S13 | moderate | low | months | CSA tool poisoning, Microsoft guidance via THN |
| I09 | Human approval for high-impact actions | technical | S1 S2 S3 S8 | moderate | low | months | AIID 1152, OWASP |
| I10 | Agent cancellation | technical | S3 S13 | weak | moderate | months | AISI incident report |
| I11 | Secure MCP registries | technical | S3 S12 S13 | weak | moderate | 1-2y | MCP registry, NSA, CSA crisis note |
| I12 | Signed skills and dependencies | technical | S12 S3 | weak | moderate | 1-2y | CVE-2025-6514, CSA crisis note |
| I13 | Model organisms of misalignment | technical | S1 | moderate | moderate | months | alignment faking, agentic misalignment |
| I14 | Interpretability | technical | S1 S17 | weak | high | 3-5y | IASR 2026 |
| I15 | Scalable oversight | technical | S1 S2 S17 | weak | high | 3-5y | IASR 2026 |
| I16 | Control evaluations | technical | S1 S3 S13 | moderate | moderate | 1-2y | AI Control paper |
| I17 | International coordination | international | S4 S6 S7 S8 S9 S10 S11 S14 S15 S16 S17 | weak | high | 3-5y | IASR 2026, EU GPAI guidelines, Science consensus paper |
| I18 | Whistleblower protections | national | S15 S9 | moderate | low | months | SB 53 |
| I19 | Safety cases | organizational | S1 S3 | weak | moderate | 1-2y | Safety Cases paper |
| I20 | Staged deployment | organizational | S2 S4 S6 S14 | moderate | low | months | RSP, Preparedness v2 |
| I21 | Emergency response | resilience | S3 S4 S5 S6 S7 | weak | moderate | 1-2y | IASR 2026, AISI incident report |
| I22 | Liability and insurance | national | S8 S15 | weak | low | 1-2y | CBC (Moffatt v. Air Canada) |
| I23 | Public-interest laboratories | national | S9 S15 S16 | weak | very_high | 3-5y | Inspect, SSTI (CAISI) |
| I24 | Nonprofit alignment research | organizational | S1 S2 | moderate | moderate | 1-2y | METR 2025, AI Control paper |
| I25 | Open safety tooling | technical | S9 S10 S12 | moderate | low | months | Inspect |
| I26 | Forecasting | organizational | S7 S16 S18 | weak | low | months | XPT, Metaculus 578 |
| I27 | Resilience engineering | resilience | S2 S5 S8 S10 S11 S12 S16 S18 | weak | high | 3-5y | IASR 2026 |

Distributions: category technical 14, organizational 5, national 4, international 2,
resilience 2; evidence strength moderate 13, weak 14 (none `strong`, none `none`); cost low 9,
moderate 12, high 5, very_high 1; time to deploy months 12, 1-2y 10, 3-5y 5. Every record is
`uncertainty: high`, `verified_search`, `informational`, `human_review_status: pending`, and
states a `possible_failure`, a `possible_backfire` (or "None significant."), `owner_types` and at
least one `user_actions` item with an audience.

The evidence summaries are category-level and cite documented instances (the AISI containment
for I10; the Replit deletion for I09; the RAND security levels for I04; the 171 Inspect
evaluations for I25) without asserting measured effect sizes, which is why `effect_size` is
`unknown` throughout and `evidence_strength` never exceeds `moderate`.

## 2. Organisations

18 records, listed because they meet stated inclusion criteria (`inclusion_criteria_met`:
independent or public-interest status, public documented outputs, verifiable website; the UK
AISI is listed as a public body), not as endorsements. Each carries mission, legal status as
stated by the organisation, jurisdiction, focus, programmes, open outputs, funding disclosure,
conflicts, evidence of impact, ways to help and `last_verified: 2026-09-26`.

| Verification | Organisations |
| --- | --- |
| `verified_search` (8) | METR, Epoch AI, Forecasting Research Institute, AI Impacts, UK AI Security Institute, Cloud Security Alliance, Institute for AI Policy and Strategy, Ada Lovelace Institute |
| `verified_prior_knowledge` (10) | Apollo Research, Redwood Research, Center for AI Safety, Centre for the Governance of AI, Centre for Long-Term Resilience, CSET, FAR.AI, Future of Life Institute, SecureBio, Transluce |

Conflicts as recorded: `government_policy_context` (UK AISI), `commercial_interest` (Cloud
Security Alliance: membership and sponsorship), `advocacy_context` (Center for AI Safety, Future
of Life Institute). Funding disclosures are recorded as the organisations state them
("Philanthropic; disclosed on website" in most cases). Ten organisations cite a source record
(`source_ids`); eight cite none and rest on the organisation's own website.

## 3. Actions

57 records across ten audiences: individuals 7, software_engineers 10, ai_researchers 10,
laboratories 9, policymakers 7, funders 6, educators 5, nonprofits 1, auditors_red_teams 1,
standards_bodies 1. Each names related interventions, an `effort` level (low, moderate, high) and,
where a resource exists, a linked resource with its source id (the International AI Safety
Report, Inspect, OWASP GenAI, NSA MCP guidance, the MCP Registry, the AI Incident Database, the
NIST AI RMF, RAND, the AI Control and Safety Cases papers, Metaculus, the EU GPAI guidelines,
California SB 53, METR time horizons, the OECD incident definitions). All actions are
`verified_prior_knowledge` and `informational`: they are editorial recommendations, and the
Act page says that none of them requires believing any particular number.

The action wording avoids operational detail: security actions describe practice (permission
boundaries, least privilege, cancellation, logging, signed dependencies, sandboxing, deterministic
policies, coordinated disclosure) at the level of principle.

## 4. How the records are displayed

- `/safeguards` lists interventions with category, evidence strength, cost, time and target
  scenarios; scenario pages list the safeguards that target them; the stage draws up to twelve
  safeguard arcs from the intervention count.
- `/act` groups actions by audience with a filter, then lists organisations with their criteria,
  funding and conflicts; `/text#act` prints all of it.
- Beneficial and protective content is deliberately prominent
  ([`../../docs/governance/psychological-safety.md`](../../docs/governance/psychological-safety.md)).

## Limitations

- Effect sizes are unknown by construction; `evidence_strength` is a judgement about the
  existence of documented practice and instances, not a measurement of impact.
- Ten organisation records and all 57 actions are prior knowledge and were not re-checked
  against the organisations' current pages; legal status and funding statements are as
  remembered, not as fetched.
- Organisation coverage is skewed towards US and UK bodies.
- Intervention–scenario targeting is editorial; there is no evidence-weighted mapping.
- No record has been human-reviewed.

## Open questions

- Should interventions carry a documented "what would count as evidence of effect" field so that
  `evidence_strength` can move on evidence rather than opinion?
- Which organisations outside the US and UK meet the inclusion criteria and should be added?
- Should actions be verified against a named resource each, dropping any action without one?
- Should `possible_backfire: "None significant."` be allowed, or should every intervention name at
  least one failure mode of its own?
