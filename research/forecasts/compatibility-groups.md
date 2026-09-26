<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Forecast compatibility groups

This document declares the `group_id` values used in
`data/snapshots/snap-2026-09-26-001/forecasts.json`. A compatibility group is the only unit
inside which p(DOOM) may aggregate external forecasts (build-spec §0.5, ADR-004). Members of
one group must share **all four** of: outcome set, horizon, conditioning, and comparable
question wording. Population may differ inside a group (the aggregation layer publishes the
`equal_weight_by_population` method and a population count), but wording, outcome and horizon
may not. The snapshot validator (`internal/snapshot/validate.go`) rejects a group that mixes
outcome sets or horizons and warns when the conditioning text differs.

Forecasts that have no compatible partner carry `group_id: null` and are shown as singletons
beside their original wording. Group membership says nothing about quality: the aggregation
layer filters on `model_use_status` before pooling. In this snapshot all 15 forecast records
are `verified_search` and `eligible`.

## Outcome-code conventions used in this snapshot

| Original wording | Outcome set recorded | Reason (from each record's `transformation_note`) |
| --- | --- | --- |
| "long-run effect of advanced AI on humanity will be extremely bad (e.g., human extinction)" | `[O3, O4, O5, O6, O7, O8]` | Respondents were not asked to separate extinction from other irreversible outcomes, so the bucket is mapped to the combined p(DOOM) set |
| "future AI advances causing human extinction or similarly permanent and severe disempowerment of the human species" | `[O3, O6]` | Explicitly extinction plus permanent severe disempowerment |
| "AI will cause human extinction by 2100" (XPT); "Will AI cause human extinction?" (Metaculus) | `[O6]` | Extinction maps to O6 directly |
| "catastrophe caused by AI that kills at least 10% of humanity by 2100" | `[O4, O5, O6]` | Broader than extinction, narrower than p(DOOM); collapse, near-extinction or extinction |
| "existential catastrophe via unaligned artificial intelligence" (Ord); "existential catastrophe from misaligned, power-seeking AI" (Carlsmith) | `[O3, O6, O8]` | Existential catastrophe covers extinction and unrecoverable loss of potential |
| "Will humans go extinct before 2100?" (any cause) | `[O6]` | Extinction, but cause-unrestricted; never grouped with AI-specific questions |
| "If a global catastrophe occurs, will it be principally due to AI?" | `[O4, O5, O6]` | Conditional on a catastrophe; never averaged with unconditional items |

## Declared groups

| group_id | Definition (outcome · horizon · conditioning) | Members | Notes |
| --- | --- | --- | --- |
| `G-XBAD-LONG-RUN` | extremely bad long-run effect, mapped to O3–O8 · `eventual` · "Conditional on high-level machine intelligence existing (survey framing)." | `fc-espai-2022-extremely-bad-long-run`, `fc-espai-2023-extremely-bad-long-run` | Two waves of the same survey series with the same wording (`paraphrase: true` on both); a time series of one population, not two independent samples |
| `G-EXT-DISEMP-UNDATED` | extinction or permanent severe disempowerment, O3 ∪ O6 · `eventual` · "Unconditional on timelines." | `fc-espai-2023-extinction-or-disempowerment`, `fc-espai-2024-extinction-or-disempowerment` | Same wording across the 2023 and 2024 surveys; the 2024 survey date is approximate (December 2024) |
| `G-EXT-2100` | AI-caused human extinction, O6 · `2100` · "Unconditional." | `fc-xpt-2023-superforecasters-ai-extinction-2100`, `fc-xpt-2023-experts-ai-extinction-2100` | Two populations of one tournament (XPT); wording paraphrased |
| `G-CAT10-2100` | AI-caused catastrophe killing ≥10% of humanity, O4–O6 · `2100` · "Unconditional." | `fc-xpt-2023-superforecasters-ai-catastrophe-2100`, `fc-xpt-2023-experts-ai-catastrophe-2100` | As above |
| `G-EXT-2100-METACULUS` | AI-caused human extinction, O6 · `2100` (horizon inferred from the linked platform questions) · "Unconditional (platform question)." | `fc-metaculus-2026-pro-forecasters-ai-extinction`, `fc-metaculus-2026-dedicated-xrisk-forecasters-ai-extinction` | Kept separate from `G-EXT-2100` because wording, population selection and the inferred horizon differ; the "dedicated x-risk forecasters" subgroup is selected for engagement with the topic |

Each group has exactly two members, so the preferred `unweighted_median` is the mean of two
values and every leave-one-source-out run moves the aggregate to one member's value. The
release's `sensitivity.json` shows this for `G-CAT10-2100` and `G-EXT-2100-METACULUS` in
particular; the model card lists those runs among the largest influences.

## Singletons (`group_id: null`)

| Forecast id | Why it has no group |
| --- | --- |
| `fc-ord-2020-unaligned-ai-100y` | Horizon is "next 100 years" from 2020 (`custom`, `horizon_end_year: 2120`); no other record shares wording and horizon |
| `fc-carlsmith-2021-power-seeking-2070` | `custom` horizon (2070); the author's 2022 update supersedes it and is recorded separately |
| `fc-carlsmith-2022-power-seeking-2070-update` | Recorded as a lower bound (">10%", median stored at the bound, `quantiles.p05` at the bound); a bound is not poolable |
| `fc-metaculus-578-human-extinction-2100` | Any cause, not AI-specific |
| `fc-metaculus-1495-gc-caused-by-ai` | Conditional on a global catastrophe occurring before 2100 |

## Reserved labels for forecasts not yet in the snapshot

Earlier drafting for this area used the labels `G-XBAD-HLMI`, `G-XDIS-AI-EVENTUAL`,
`G-XDIS-CONTROL-EVENTUAL`, `G-TAKEOVER-2070`, `G-XCAT-2100`, `G-HARM50-2050`,
`G-EXT-2100-ALLCAUSE`, `G-CAT10-2100-ALLCAUSE`, `G-EXT-2030`, `G-CAT10-2030`, `G-EXT-2050` and
`G-XRISK-100Y` for forecast records (ESPAI 2016 wave, ESPAI control-failure wording, XPT
2030 and 2050 horizons, XPT any-cause questions, the Good Judgment Carlsmith-premise forecast,
the FRI adversarial-collaboration and Conditional Trees figures, LEAP wave-9 items,
Samotsvety, Manifold) that are **not** in `snap-2026-09-26-001`. They are reserved so that
future records reuse the same names, and nothing in the release refers to them. A record using
one of them must first be verified and added to `tools/snapshot/content/forecasts.mjs` or a
fragment.

## Rules for adding a member

1. Copy the outcome set from the conventions table; do not invent a new mapping for a wording
   that is already mapped. A new wording gets a new row in that table.
2. Horizon keys: `2100`, `eventual`, or `custom` with `horizon_note` and `horizon_end_year`.
   "Within 100 years" from a dated source is `custom`.
3. Conditioning strings must match exactly for group membership; the literals used are
   "Unconditional.", "Unconditional on timelines.", "Unconditional (platform question)." and
   "Conditional on high-level machine intelligence existing (survey framing)."
4. A forecast whose wording restricts the cause (for example "inability to control") is a
   different group from one that does not.
5. Any-cause forecasts never join an AI-specific group.
6. Prediction-market and platform items may join a group but should stay `informational` until
   the as-of date of the quoted value is verified from the live page; the two Metaculus platform
   analysis items are `eligible` on the strength of the dated platform article and should be
   re-checked at review.
7. Adding a forecast source id raises the heightened-review trigger
   `expert_survey_or_forecast_source_added`.

## Limitations

- Group membership was assigned from question wording as recovered through search-result
  snippets; the eight records marked `paraphrase: true` must be checked against the primary
  documents before the next release.
- Outcome mappings are editorial judgements; the conventions table is the single place to
  revise them, and a revision changes every aggregate that uses the mapping.
- Two of the five groups are time series from one survey series (AI Impacts); pooling across
  waves conflates population drift, wording stability and genuine belief change.
- The `G-EXT-2100-METACULUS` horizon is inferred, not stated in the source article.

## Open questions

- Should `G-XBAD-LONG-RUN` be treated as conditional on HLMI for aggregation purposes, or
  displayed as an unconditional long-run judgement given that most respondents expect HLMI
  this century?
- Is `[O4, O5, O6]` the right reading of the 10%-mortality threshold, or should threshold
  questions be kept out of outcome-coded aggregation and shown only as raw thresholds?
- Once the XPT wordings are verified from the primary paper, should `G-EXT-2100` and
  `G-EXT-2100-METACULUS` be merged, and how should the release note describe the change?
