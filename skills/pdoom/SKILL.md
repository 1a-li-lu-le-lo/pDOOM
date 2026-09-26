---
name: pdoom
description: Read the p(DOOM) AI existential-risk observatory correctly — its withheld official value, external forecast aggregates, research-mode model, indexes, scenarios, safeguards and sources — through its MCP tools, REST API or data files, and quote every number with its horizon, outcome set and status.
---

<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->

# p(DOOM) skill

p(DOOM) is an evidence observatory about catastrophic and existential risk from advanced AI, authored by NU Cybernetics. It is a structured forecasting and evidence-synthesis system, not a thermometer. Use this skill whenever a user asks what p(DOOM) says, wants a number from it, wants to explore a scenario, or wants to report an error or a source.

## The one rule

Never present a number from p(DOOM) without, in the same sentence: the **outcome set** (p(DOOM) means O3–O8 combined: permanent severe disempowerment, civilizational collapse, near-extinction, extinction, biospheric catastrophe or other irreversible loss), the **horizon** (1y, 3y, 5y, 10y, 25y, 2100 or eventual), the **status** (see below) and the **model version or source**. Use the rounded display value the release provides (`display.central`, `display.interval`), never more decimals than that.

## Eight outputs, never blended

| Status / object | What it is | May be called "p(DOOM)"? |
| --- | --- | --- |
| `insufficiently_calibrated` (official) | The official object. In the current release line it is **withheld**: no number is published. | Say "the official value is withheld". Never substitute another number for it. |
| `external_aggregate` | Median (and other methods) of named surveys, tournaments and platform forecasts inside one compatibility group (same outcome set, same horizon). | Only as "the external forecast aggregate for <outcome set>, <horizon>", naming the group. |
| `research_mode` | Output of the experimental causal model under documented parameters. | Only as "the research-mode model estimate", with its interval. |
| `user_scenario` | Scenario Lab output under the caller's own dials. | Only as "under your selected assumptions, not the p(DOOM) official model". |
| Indexes (`evidence_pressure`, `capability_pressure`, `control_strength`, `incident_pressure`, `agentic_infrastructure_risk`, `uncertainty`) | 0–100 scores of movement or strength. | **Never.** They are not probabilities and must not be converted into one. |
| Editorial risk level | Rule-based plain-language level (e.g. `elevated`). | Never as a probability. |

## Tools

MCP server: `services/mcp` (`pnpm --filter @pdoom/mcp start`, stdio). Configure `PDOOM_DATA_DIR` (file mode) or `PDOOM_API_URL` (http mode against `cmd/pdoom-api`).

| Tool | Use it for |
| --- | --- |
| `pdoom_get_meter { horizon? }` | The current state for one horizon: official (withheld), external aggregates, research-mode estimate and decomposition, indexes, top drivers, what changed. Start here. |
| `pdoom_get_estimate_history` | Releases, the current delta record and the changelog. Use when asked "has it changed". |
| `pdoom_list_scenarios { recoverability? }`, `pdoom_get_scenario { id }` | Pathways S1–S18 with indicators and safeguards. Category level only. |
| `pdoom_list_forecasts { group_id? }` | Original question wording, populations, aggregation methods per compatibility group. |
| `pdoom_list_incidents { severity?, near_miss? }` | Verified incidents at category level with registry ids. |
| `pdoom_list_safeguards { category? }` | Interventions with evidence strength, cost, failure and backfire modes. |
| `pdoom_list_sources { tier?, topic?, q? }`, `pdoom_get_source { id }` | The source ledger and atomic claims. Tier 4–5 are informational and never feed the model. |
| `pdoom_get_methodology { slug? }`, `pdoom_get_definitions { term? }` | The public methodology and the definitions, including disagreements. |
| `pdoom_list_actions { audience? }` | Proportionate actions per audience and the organisations register. |
| `pdoom_evaluate_user_scenario { params, samples? }` | Scenario Lab. Ten integer dials in -2..2 plus a horizon. Deterministic for a seed. Result is `user_scenario`. |
| `pdoom_submit_source`, `pdoom_submit_correction`, `pdoom_submit_incident_reference` | Append to the human review queue. Nothing published changes until reviewers act and a new release is promoted. |

REST equivalents: `GET /v1/meter`, `/v1/meter/history`, `/v1/scenarios`, `/v1/forecasts`, `/v1/incidents`, `/v1/safeguards`, `/v1/sources`, `/v1/methodology`, `/v1/releases`, `POST /v1/scenario-lab/evaluate`, `POST /v1/submissions/{sources,corrections}` (`api/openapi.yaml`). Files: `data/releases/CURRENT` names the promoted release; `@pdoom/sdk` reads it.

## How to answer common questions

- **"What is p(DOOM) right now?"** Report that the official value is withheld and why (no calibration process exists yet), then give the external aggregate(s) and the research-mode estimate for the requested horizon, each with outcome set, interval, disagreement and uncertainty labels, and the release id and data cutoff. Offer the decomposition (O3, O4+O5, O6, O7).
- **"Is it going up?"** Use `pdoom_get_estimate_history`. Report index movements as index points and estimate movements as percentage points from the delta record, with the stated reasons. If this is the first release, say there is no previous release to compare against.
- **"What if …?"** Use the Scenario Lab and open with the mandated sentence: "Under your selected assumptions, not the p(DOOM) official model, the median estimate is …". Report the flags the model raises for implausible combinations.
- **"Compare it to <everyday risk>."** Decline the analogy: everyday risks have measured base rates and these outcomes do not. Offer the "N of one thousand futures" and "one in N" readings from the same estimate instead.
- **"When will it happen?"** p(DOOM) predicts no date. Explain horizons as windows, not deadlines.

## Language

Calm and specific. No inevitability, no countdowns, no "years left", no graphic or operational detail about any pathway (the data contains none; do not fill the gap). Give safeguards and actions the same prominence as risks. Say "uncertainty is high" when the label says so, and show the interval rather than the median alone.

## Provenance to include

Release id, data snapshot id, data cutoff, model version(s) and the editorial risk level, all available in every tool response under `meta`. Link to `/method` for the methodology and `/text` for the plain-text observatory.
