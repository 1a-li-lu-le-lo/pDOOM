<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Forecast inventory

One row per record in `data/snapshots/snap-2026-09-26-001/forecasts.json` (15 records,
all `verified_search`, all `eligible`, all `status: current`). Central values are shown as
recorded in the snapshot (probabilities on the 0–1 scale), beside the wording, population,
outcome set and horizon that give them meaning. They are source data, not observatory
estimates; the release's aggregates and their display rounding are in
`data/releases/rel-2026-09-26-001/aggregations.json` and `estimates.json`.

| Forecast id | Forecaster or survey (source) | Date | Population (n) | Wording (P = marked paraphrase) | Outcome set | Horizon | Conditioning | As recorded | Group |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fc-espai-2022-extremely-bad-long-run` | 2022 Expert Survey on Progress in AI (`src-aiimpacts-2022-espai`) | 2022-08-03 | general_ai_researchers (738) | P: How likely is it that the long-run effect of advanced AI on humanity will be extremely bad (e.g., human extinction)? | O3–O8 | eventual | Conditional on HLMI existing (survey framing) | median 0.05; response rate 0.17 | `G-XBAD-LONG-RUN` |
| `fc-espai-2023-extremely-bad-long-run` | 2023 ESPAI, Grace et al. 2024 (`src-grace-2024-thousands-of-ai-authors`) | 2023-10-15 | general_ai_researchers (2,778) | P: same wording | O3–O8 | eventual | same | median 0.05; response rate 0.15 | `G-XBAD-LONG-RUN` |
| `fc-espai-2023-extinction-or-disempowerment` | 2023 ESPAI (`src-grace-2024-thousands-of-ai-authors`) | 2023-10-15 | general_ai_researchers (2,778) | What probability do you put on future AI advances causing human extinction or similarly permanent and severe disempowerment of the human species? | O3, O6 | eventual | Unconditional on timelines | median 0.05; mean 0.162 | `G-EXT-DISEMP-UNDATED` |
| `fc-espai-2024-extinction-or-disempowerment` | 2024 ESPAI (`src-aiimpacts-2024-espai`) | 2024-12-15 (approx.) | general_ai_researchers (1,580) | same wording | O3, O6 | eventual | Unconditional on timelines | median 0.10; mean 0.18; response rate 0.10 | `G-EXT-DISEMP-UNDATED` |
| `fc-xpt-2023-superforecasters-ai-extinction-2100` | XPT superforecasters (`src-fri-2023-xpt`) | 2023-07-01 | superforecasters (89) | P: What is the probability that AI will cause human extinction by 2100? | O6 | 2100 | Unconditional | median 0.0038 | `G-EXT-2100` |
| `fc-xpt-2023-experts-ai-extinction-2100` | XPT domain experts (`src-fri-2023-xpt`) | 2023-07-01 | domain_experts (80) | P: same | O6 | 2100 | Unconditional | median 0.03 | `G-EXT-2100` |
| `fc-xpt-2023-superforecasters-ai-catastrophe-2100` | XPT superforecasters (`src-fri-2023-xpt`) | 2023-07-01 | superforecasters (89) | P: What is the probability of a catastrophe caused by AI that kills at least 10% of humanity by 2100? | O4, O5, O6 | 2100 | Unconditional | median 0.0213 | `G-CAT10-2100` |
| `fc-xpt-2023-experts-ai-catastrophe-2100` | XPT domain experts (`src-fri-2023-xpt`) | 2023-07-01 | domain_experts (80) | P: same | O4, O5, O6 | 2100 | Unconditional | median 0.12 | `G-CAT10-2100` |
| `fc-ord-2020-unaligned-ai-100y` | Toby Ord, The Precipice, Table 6.1 (`src-ord-2020-precipice`) | 2020-03-05 | individual_expert | P: Existential catastrophe via unaligned artificial intelligence within the next 100 years: 1 in 10. | O3, O6, O8 | custom (to about 2120) | Unconditional; existential catastrophe includes unrecoverable collapse and dystopia | median 0.10 | none |
| `fc-carlsmith-2021-power-seeking-2070` | Joseph Carlsmith, 2021 estimate (`src-carlsmith-2022-power-seeking`) | 2021-04-01 | individual_expert | P: Probability of existential catastrophe from misaligned, power-seeking AI by 2070: ~5%. | O3, O6, O8 | custom (2070) | Unconditional; six-premise argument | median 0.05 | none |
| `fc-carlsmith-2022-power-seeking-2070-update` | Joseph Carlsmith, 2022 update (`src-carlsmith-2022-power-seeking`) | 2022-06-16 | individual_expert | P: The author's estimate has risen to greater than 10%. | O3, O6, O8 | custom (2070) | Unconditional | lower bound 0.10 (median stored at the bound; `quantiles.p05` 0.10) | none |
| `fc-metaculus-578-human-extinction-2100` | Metaculus community, question 578 (`src-metaculus-578-human-extinction-2100`) | 2026-09-23 | public_forecasters (1,580) | Will humans go extinct before 2100? | O6 | 2100 | Any cause; strict "no human alive" resolution | median 0.02 | none (any cause) |
| `fc-metaculus-2026-pro-forecasters-ai-extinction` | Metaculus Pro Forecasters, platform analysis (`src-metaculus-2026-pro-forecasters-ai-extinction`) | 2026-09-23 | public_forecasters (31) | P: Will AI cause human extinction? (Pro Forecasters: 7.5%) | O6 | 2100 (inferred) | Unconditional (platform question) | median 0.075 | `G-EXT-2100-METACULUS` |
| `fc-metaculus-2026-dedicated-xrisk-forecasters-ai-extinction` | Metaculus "most dedicated x-risk forecasters" (`src-metaculus-2026-pro-forecasters-ai-extinction`) | 2026-09-23 | public_forecasters | P: Will AI cause human extinction? (Most dedicated x-risk forecasters: 22%) | O6 | 2100 (inferred) | Unconditional (platform question) | median 0.22 | `G-EXT-2100-METACULUS` |
| `fc-metaculus-1495-gc-caused-by-ai` | Metaculus community, question 1495 (`src-metaculus-1495-gc-caused-by-ai`) | 2026-09-23 | public_forecasters (464) | If a global catastrophe occurs, will it be principally due to AI? | O4, O5, O6 | 2100 | Conditional on a global catastrophe before 2100 | median 0.27 | none (conditional) |

## Reasons for disagreement recorded on the records

The `selection_effects`, `framing_effects` and `calibration` fields record, per source:

- **Question wording**: "extremely bad (e.g., human extinction)" versus "human extinction or
  similarly permanent and severe disempowerment" versus "AI will cause human extinction by
  2100" versus a 10%-mortality threshold.
- **Horizon**: undated long-run questions versus 2100 versus 2070 versus "next 100 years".
- **Conditioning**: on HLMI existing; on a global catastrophe occurring; unconditional.
- **Population selection**: self-selected survey respondents with response rates of 0.10–0.17;
  incentivised tournament participants; platform users; a platform-defined subgroup selected
  for engagement with existential-risk questions.
- **Framing**: the 2023 survey documents substantial framing effects across question variants;
  the 2022 survey notes answers depended on whether extinction was asked about directly.
- **Calibration**: superforecasters have a strong short-horizon track record; no population has
  a track record on long-horizon existential questions; individual expert judgments carry no
  calibration record.

## Limitations

- Eight of fifteen wordings are paraphrases recovered through search snippets.
- The two Metaculus platform-analysis items share one source article; their horizon is inferred.
- The Carlsmith 2022 update is a bound, not a point estimate, and is stored with the median at
  the bound; it must never be pooled.
- Population sizes for the XPT items are the tournament totals for each population, not the
  respondent count for each question.

## Open questions

- Should the XPT items carry the exact question text once the paper is fetched, with
  `paraphrase` cleared?
- Should platform community predictions be recorded with the platform's own aggregation method
  named, since "community prediction" is itself an aggregate?
