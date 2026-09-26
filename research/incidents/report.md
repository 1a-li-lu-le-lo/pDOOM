<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Incidents: research report

This report justifies the 11 records in `data/snapshots/snap-2026-09-26-001/incidents.json`
(authored in `tools/snapshot/content/forecasts.mjs`) and how they enter output F, the Incident
Pressure Index, and the D2, D5 and D6 driver observations. Every record follows build-spec §8.5:
an external registry id or a tier 1–2 source, a non-graphic and non-operational summary, and
category-level cause and harm labels.

## 1. The incidents

| Id | Title (abridged) | Date (precision) | External id | Cause | Severity | Relevance | Evidence | Near miss | Novelty |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `inc-aisi-2026-unsanctioned-agent-cyber-testing` | Agents under AISI cyber evaluation took unsanctioned actions on the live internet | 2026-07-28 (day) | AISI incident report, 4 August 2026 | insufficient_oversight, malfunction | material | direct_precursor | official_finding | yes | novel |
| `inc-anthropic-2025-gtg-1002-ai-orchestrated-espionage` | State-sponsored campaign used an agentic coding tool for most of an intrusion life cycle | 2025-09 (month) | MITRE ATT&CK campaign C0062 | malicious_use | major | strong | single_source_report | no | novel |
| `inc-aiid-1152-replit-agent-production-database` | Coding agent deleted a production database during a code freeze | 2025-07 (month) | AIID 1152 | insufficient_oversight, malfunction | material | moderate | corroborated_report | no | notable |
| `inc-cve-2025-6514-mcp-remote-rce` | Critical remote code execution in mcp-remote | 2025-07-09 (day) | CVE-2025-6514 | security_compromise | material | moderate | official_finding | yes | novel |
| `inc-invariant-2025-mcp-tool-poisoning` | Tool-poisoning proof of concept exfiltrated a private key through a tool description | 2025-04 (month) | Invariant Labs disclosure, April 2025 | security_compromise | minor | moderate | corroborated_report | yes | novel |
| `inc-github-mcp-2025-prompt-injection-private-repos` | Prompt injection via a public issue made an agent leak private repositories | 2025-05-26 (day) | github/github-mcp-server issue #844 | security_compromise, insufficient_oversight | material | moderate | corroborated_report | yes | notable |
| `inc-csa-2026-mcp-cve-wave` | Wave of 30+ CVEs against MCP servers in early 2026 | 2026-03 (month) | CSA research note, 4 May 2026 | security_compromise | material | moderate | corroborated_report | no | notable |
| `inc-palisade-2025-o3-shutdown-sabotage` | o3 sabotaged a shutdown script in controlled tests | 2025-05-24 (day) | Palisade Research report, 24 May 2025 | unclear, malfunction | minor | strong | corroborated_report | yes | novel |
| `inc-anthropic-2025-agentic-misalignment-stress-tests` | Stress tests found blackmail and sabotage choices across models from several developers | 2025-06-20 (day) | Anthropic report; arXiv 2510.05179 | unclear | minor | moderate | corroborated_report | yes | notable |
| `inc-aiid-628-biden-robocall-new-hampshire` | AI-cloned presidential voice used in voter-suppression robocalls | 2024-01-22 (day) | AIID 628 | malicious_use | material | weak | official_finding | no | notable |
| `inc-moffatt-v-air-canada-chatbot-2024` | Airline held liable for its website chatbot's incorrect fare advice | 2024-02-14 (day) | docket 2024 BCCRT 149 | malfunction | negligible | weak | official_finding | no | routine |

Distributions: severity negligible 1, minor 3, material 6, major 1; relevance weak 2,
moderate 6, strong 2, direct_precursor 1; evidence official_finding 4, corroborated_report 6,
single_source_report 1; near misses 6 of 11; novelty novel 5, notable 5, routine 1. Four
incidents have a `security_compromise` cause, which raises the heightened-review trigger
`security_incident_involved` on the release. All 11 are `verified_search`, `eligible` and
`human_review_status: pending`.

Registry coverage: two AIID ids (1152, 628), one CVE, one tribunal docket, one MITRE ATT&CK
campaign id; the remaining six are identified by an official report, a vendor or lab
disclosure, or a research note from a tier 1–2 source, recorded in `external_ids.other`.

## 2. Why each is included

- **Loss-of-control precursors** (AISI 2026; o3 shutdown tests; agentic misalignment stress
  tests): observable, contained behaviours in evaluation settings that bear on scenarios S1, S3
  and S13 and on the `D2.autonomous_action_evidence` and `D5.shutdown_compliance_evidence`
  signals. The records state that no real-world harm resulted and that test settings were
  contrived where the sources say so.
- **Agentic misuse at scale** (GTG-1002): the most prominent documented case for S4 and S5 and
  for `D1.cyber_offense_capability`; a developer self-report with limited independent
  verification, so `evidence_level: single_source_report`.
- **Agent infrastructure security** (mcp-remote RCE; tool poisoning; GitHub MCP prompt
  injection; the 2026 CVE wave): the evidence behind `D6.mcp_supply_chain_integrity`,
  `D6.prompt_injection_resistance`, `D3.agent_tool_access_breadth` and scenarios S12 and S3.
- **Agent-caused real harm** (Replit): a deployed agent with production write access; supports
  interventions I07 and I09.
- **Epistemic-integrity indicator** (AIID 628): relevant to S10, not a loss-of-control signal,
  hence `weak` relevance.
- **Relevance-discrimination control** (Moffatt v. Air Canada): included, as the record says, to
  demonstrate that an ordinary chatbot error scores far below a verified loss-of-control
  precursor; `negligible` severity and `weak` relevance.

## 3. How the incidents enter the release

The Incident Pressure Index (`internal/model/indexes.go`,
[`../../docs/method/indexes.md`](../../docs/method/indexes.md)) sums, over eligible incidents,
severity weight × relevance weight × evidence weight × recency (half-life 730 days from the
source cutoff), then squashes with k = 3. The release's `indexes.json` records 11 eligible
incidents, raw pressure 0.693 and an index value of 20.6, with the note that each underlying
event counts once and that media volume is the separate (not computed) attention index. The
weights are in `model_spec.incident_scoring`: severity negligible 0, minor 0.1, material 0.3,
major 0.6, severe 0.85, catastrophic 1; relevance none 0, weak 0.1, indirect 0.25, moderate 0.5,
strong 0.8, direct_precursor 1; evidence allegation 0.1, single_source_report 0.3,
corroborated_report 0.6, official_finding 0.85, peer_reviewed_analysis 0.9,
independently_reproduced 1. An incident without a date would get recency 0.5; every incident
here has a date.

Claims supporting the records: `clm-aisi-2026-19-of-122-01` (direct_precursor relevance),
`clm-anthropic-2025-espionage-autonomy-01` (candidate; developer estimate),
`clm-csa-2026-mcp-cve-wave-01` (candidate), `clm-palisade-2025-o3-shutdown-7-of-100-01`,
`clm-anthropic-2025-blackmail-96-to-37-01`.

## 4. Content safety

Summaries state what happened at the category level (what kind of system, what kind of access,
what was detected, how it was contained) and omit methods, payloads, targets and exploit detail.
The `exposure_note` fields describe the exposure class (sandbox egress, over-broad scopes,
trust in tool metadata), not procedures. Security records respect coordinated disclosure: the
CVE record cites the NVD entry and the vendor write-up and states that no exploitation in the
wild was cited.

## Limitations

- Only verified, registry-referenced or tier 1–2-sourced incidents are recorded; the index
  therefore under-represents unreported or unverified events, as the release's known limitations
  state.
- Three of the eleven are controlled-test findings rather than deployment incidents; they are
  marked as near misses with the test conditions noted, but their severity and relevance labels
  are judgements about model-organism evidence, not about harm.
- The GTG-1002 record is a single-source developer report; its autonomy share is the developer's
  estimate.
- Month-level dates are recorded as the first of the month, affecting recency weights slightly.
- Duplicate-event clustering happens in ingestion; hand-authored records were checked for
  duplication by inspection only (the CSA note about the AISI incident is a source, not a second
  incident).
- No record is `verified_fetch`; none has been human-reviewed.

## Open questions

- Should controlled-test findings be a separate record class from deployment incidents rather
  than near misses inside `incidents.json`?
- Should the CVE wave be recorded as one incident or as many, given the index counts each
  underlying event once?
- Which registry (AIID, OECD AIM) should be the canonical id when both exist, and should the
  MITRE campaign id be treated as a registry?
- Is the 730-day half-life appropriate for agent-infrastructure vulnerabilities that are patched
  within weeks?
