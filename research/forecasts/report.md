<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Forecasts: research report

This report justifies the 15 forecast records in `data/snapshots/snap-2026-09-26-001/forecasts.json`,
the five compatibility groups declared in [`compatibility-groups.md`](compatibility-groups.md), and
the way they enter output A (the external forecast aggregate) of release `rel-2026-09-26-001`.
The row-by-row inventory is [`forecast-inventory.md`](forecast-inventory.md).

## 1. What was collected

Nine sources feed the forecast records (`forecast_source_ids` in the release manifest's data
summary): `src-aiimpacts-2022-espai`, `src-aiimpacts-2024-espai`,
`src-carlsmith-2022-power-seeking`, `src-fri-2023-xpt`,
`src-grace-2024-thousands-of-ai-authors`, `src-metaculus-1495-gc-caused-by-ai`,
`src-metaculus-2026-pro-forecasters-ai-extinction`, `src-metaculus-578-human-extinction-2100`,
`src-ord-2020-precipice`. A tenth source, `src-80000hours-2023-karger-xpt` (tier 4, podcast), is
`informational` and records only that the XPT medians moved during the tournament.

By population: general AI researchers (4 records, from three survey waves of the AI Impacts
series), superforecasters (2, XPT), domain experts (2, XPT), individual experts (3: Ord,
Carlsmith 2021, Carlsmith 2022), public forecasters (4: three Metaculus platform items and one
platform analysis with two subgroups). By horizon: `2100` (8), `eventual` (4), `custom` (3).

Every record carries the original wording (eight as marked paraphrases), `outcome_set`,
`horizon`, `conditions`, the recorded central value, and where the source states them
`sample_size`, `response_rate`, `selection_effects`, `framing_effects` and `calibration`.

## 2. How the records were verified

All 15 records are `verified_search` (`checked_at: 2026-09-26`): a search-result snippet
quoting the figure and date was checked, because page fetches were blocked in the build
environment. None is `verified_fetch`. All are `human_review_status: pending`. The
corresponding claims in `claims.json` (`clm-grace-2024-*`, `clm-aiimpacts-*`, `clm-xpt-2023-*`,
`clm-ord-2020-*`, `clm-carlsmith-*`, `clm-metaculus-*`) carry `direct_quote_pointer` values such
as "Abstract; Section 4", "Table 6.1", "Results tables (AI extinction by 2100)" and "Question
page" so that a reviewer with the primary document can confirm each figure; the Metaculus
Pro-Forecaster claim is `status: candidate` rather than `corroborated` because the platform
article, not the underlying questions, is the source.

## 3. How the records enter the model

- Only records inside a compatibility group are aggregated. The release contains five groups
  of two members each: `G-XBAD-LONG-RUN`, `G-EXT-DISEMP-UNDATED`, `G-EXT-2100`,
  `G-CAT10-2100`, `G-EXT-2100-METACULUS`. For each, `aggregations.json` holds seven method
  results (35 in total); `unweighted_median` is preferred. Each group yields one estimate
  (`est-external-<group>`) whose p05/p95 are the member minimum and maximum and whose p25/p75
  are member quartiles; with two members the interval is simply the two values.
- The five singletons are displayed on `/forecasts` and `/text#forecasts` with their wording
  and are never averaged.
- Disagreement per group is labelled from the interquartile range of member log-odds; with two
  members this is the gap between them. All five groups are labelled `extreme` disagreement in
  the release; the uncertainty score's `forecast_disagreement` component reads 0.2961.
- Leave-one-source-out runs exist for every member (10 runs) and alternative-weighting runs for
  every non-preferred method (30 runs). Because every group has two members, removing either
  member moves the aggregate to the other member's value; the release model card lists the
  `G-CAT10-2100` and `G-EXT-2100-METACULUS` leave-one-out runs among the largest influences.
- The research-mode model does not use forecast records numerically. Its rationale cites the
  ESPAI aggregate timeline and the gap between extinction-specific and "extremely bad" medians
  as anchors for its A and O judgments
  ([`../methodology/experimental-causal-parameters.md`](../methodology/experimental-causal-parameters.md)).

## 4. Transformations to the outcome taxonomy

Each record's `transformation_note` states the mapping; the conventions are collected in
[`compatibility-groups.md`](compatibility-groups.md). The two consequential judgements are:

1. Mapping "extremely bad (e.g., human extinction)" to the full O3–O8 set rather than to a
   narrower set. This makes `G-XBAD-LONG-RUN` comparable in definition to the combined p(DOOM)
   object, at the cost of assuming respondents included non-extinction irreversible outcomes.
2. Keeping "10% of humanity" threshold questions as `[O4, O5, O6]`, which treats a casualty
   threshold as equivalent to the collapse/near-extinction/extinction classes.

Both are editorial choices and are open questions below.

## 5. Populations compared

The disagreement display ([`../../docs/method/disagreement.md`](../../docs/method/disagreement.md))
places groups side by side with a warning that sample sizes and quality differ. In this
snapshot: general AI researchers (thousands of respondents, response rates 0.10–0.17, not
forecasting-trained), superforecasters (89, strong short-horizon record), domain experts (80),
public forecasters (hundreds to thousands, self-selected; and a 31-person paid subgroup), and
individual experts shown as single points.

## Limitations

- Every group has exactly two members; the aggregates are fragile by construction and the
  release says so in its known limitations ("groups are small and dominated by a few studies").
- Two of the five groups are successive waves of one survey series; pooling across waves is a
  trend, not an independent replication.
- Eight wordings are paraphrases; the XPT wording in particular should be replaced by the
  paper's exact question text.
- The Metaculus horizon is inferred; the "dedicated x-risk forecasters" subgroup is selected on
  engagement with the topic.
- The Carlsmith 2022 record is a lower bound stored as a median; it is a singleton and is never
  pooled, but readers of `forecasts.csv` should not treat its `median` as a point estimate.
- No `verified_fetch` record exists; none has been human-reviewed.
- Forecast records were not collected for the `1y`, `3y`, `5y`, `10y` or `25y` horizons, so no
  external aggregate exists at the default meter horizon; the meter states this.

## Open questions

- Should `G-XBAD-LONG-RUN` be mapped to O3–O8 or to a narrower set, and should it be treated as
  conditional on HLMI?
- Should the two Metaculus subgroups be one group, or should a platform-defined subgroup be
  `informational` until its selection is understood?
- Which additional verified records would give a group a third member (XPT 2030/2050 items,
  ESPAI 2016 wave, control-failure wording, LEAP wave items are reserved in the groups file)?
- Should the release display the member range or a sampling-style interval for two-member
  groups, given that the range of two values reads like an interval but is not one?
