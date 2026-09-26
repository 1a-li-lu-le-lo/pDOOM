<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Definitions: research report

This report justifies the 29 records in `data/snapshots/snap-2026-09-26-001/definitions.json`
(authored in `tools/snapshot/content/definitions.mjs`), rendered at `/method/definitions` and
summarised in `/text#definition`. Every definition record has a `term`, a `short_definition`
(the observatory's working definition), one or more attributed `definitions` with a `source_id`
where a source exists, a `consensus` label, `related_ids` and, in this snapshot, the verification
status `verified_prior_knowledge` with `model_use_status: informational` and
`human_review_status: pending`. Definitions never feed a computation.

## 1. The terms

| Id | Term | Consensus | Attributed definitions (source) |
| --- | --- | --- | --- |
| `def-pdoom` | p(DOOM) | contested | common usage (no source); the observatory's own definition in terms of O3–O8, horizon and model specification |
| `def-agi` | AGI | contested | OpenAI Charter (`src-openai-charter`); Morris et al. Levels of AGI (`src-morris-2023-levels-of-agi`); ESPAI HLMI wording (`src-grace-2024-thousands-of-ai-authors`) |
| `def-asi` | ASI | contested | Bostrom 2014 (`src-bostrom-2014-superintelligence`) |
| `def-frontier-model` | Frontier model | contested | EU AI Act Article 51 (`src-eu-ai-act-oj`); developer frameworks (`src-anthropic-rsp`) |
| `def-agent` | Agent | contested | International AI Safety Report 2026 working usage (`src-iasr-2026`) |
| `def-autonomy` | Autonomy | contested | Morris et al. autonomy levels |
| `def-agentic-harness` | Agentic harness | emerging | METR usage (`src-metr-2025-measuring-long-tasks`) |
| `def-harness-of-harnesses` | Harness of harnesses | emerging | working definition by this observatory (no source) |
| `def-mcp` | MCP | consensus | MCP specification (`src-mcp-specification`) |
| `def-tool-use` | Tool use | consensus | MCP specification |
| `def-compute` | Compute | consensus | Sastry et al. 2024 (`src-sastry-2024-computing-power-governance`) |
| `def-capability` | Capability | contested | OpenAI Preparedness Framework v2 (`src-openai-preparedness-v2`) |
| `def-alignment` | Alignment | contested | alignment faking paper (`src-anthropic-2024-alignment-faking`); Hendrycks et al. 2023 |
| `def-control` | Control | emerging | AI Control paper (`src-greenblatt-2023-ai-control`) |
| `def-containment` | Containment | consensus | NSA MCP guidance (`src-nsa-2026-mcp-security`) |
| `def-evaluation` | Evaluation | consensus | AISI Inspect (`src-aisi-inspect`) |
| `def-incident` | Incident | consensus | OECD 2024 (`src-oecd-2024-defining-ai-incidents`); EU AI Act Article 3(49) |
| `def-near-miss` | Near miss | emerging | OECD "AI hazard" |
| `def-catastrophe` | Catastrophe | contested | XPT 10%-mortality threshold (`src-fri-2023-xpt`) |
| `def-extinction` | Extinction | consensus | Metaculus question 578 resolution criteria |
| `def-disempowerment` | Disempowerment | contested | Kulveit et al. 2025; ESPAI wording |
| `def-lock-in` | Lock-in | contested | Ord 2020 (`src-ord-2020-precipice`) |
| `def-recoverability` | Recoverability | emerging | working definition by this observatory |
| `def-systemic-risk` | Systemic risk | contested | EU AI Act Article 3(65); Critch & Russell TASRA |
| `def-malicious-use` | Malicious use | consensus | Hendrycks et al. 2023 |
| `def-malfunction` | Malfunction | consensus | International AI Safety Report 2026 |
| `def-loss-of-control` | Loss of control | contested | DeepMind FSF v3 (`src-gdm-fsf-v3`) |
| `def-epistemic-uncertainty` | Epistemic uncertainty | consensus | Der Kiureghian & Ditlevsen 2009 |
| `def-aleatory-uncertainty` | Aleatory uncertainty | consensus | Der Kiureghian & Ditlevsen 2009 |

Consensus labels: consensus 11, contested 13, emerging 5. Two records (`def-harness-of-harnesses`,
`def-recoverability`) carry a working definition attributed to "This observatory" because no
published consensus definition exists; the note on the first says so explicitly.

## 2. How definitions are used

- `def-pdoom` fixes the meaning of every p(DOOM) object: outcomes O3–O8, a stated horizon,
  conditional on a stated model specification, a model output never a measurement, always shown
  with its decomposition. It is the textual counterpart of `OUTCOMES` in `@pdoom/schemas` and of
  the release manifest's `outcome_definition`.
- `def-recoverability` is the dividing line between O1–O2 and O3–O8 and therefore between
  "serious harm" and "doom" in the outcome ladder; it is a working definition and is labelled
  `emerging`.
- `def-incident` and `def-near-miss` govern what may enter `incidents.json` (registry or tier 1–2
  source; contained failures and controlled-test findings recorded as near misses).
- `def-catastrophe` documents that the XPT 10%-mortality threshold is a threshold, not a
  trajectory class, which is why the forecast conventions map it to `[O4, O5, O6]` with
  reservations.
- Agent-related terms (`def-agent`, `def-autonomy`, `def-agentic-harness`,
  `def-harness-of-harnesses`, `def-mcp`, `def-tool-use`, `def-containment`) anchor the D2, D3
  and D6 driver signals and the agentic infrastructure risk index.

## 3. Verification

All 29 records are `verified_prior_knowledge`: the attributed wordings were written from prior
knowledge of the cited documents and not re-fetched. Eleven of the cited sources are themselves
prior-knowledge records ([`../sources/report.md`](../sources/report.md) §2). Because definitions
are informational, this does not affect any computed value, but the attributed quotations must be
checked against the documents at review, and any that are paraphrases must be marked as such.

## Limitations

- No definition has been verified by fetch or by search; quotations may be paraphrases.
- Contested terms (AGI, alignment, catastrophe, disempowerment, loss of control) are given the
  observatory's working definition in `short_definition`, which is an editorial choice.
- `see_also_urls` is empty on every record.
- The definitions page shows the records but does not yet show which entities depend on each term.

## Open questions

- Should `def-pdoom` list the ordinary-usage meanings it departs from (undated, unconditioned
  "p(doom)" figures) so that readers see why the observatory's numbers differ from quoted ones?
- Should `def-recoverability` be given a time bound ("within generations" is the current wording)
  or an institutional test?
- Which attributed wordings are verbatim under fair-use limits and which should be marked as
  paraphrases?
