<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# External forecast aggregation

## Compatibility groups

Forecasts are aggregated only inside a compatibility group (`forecast.group_id`): the same outcome set, the same horizon and the same conditioning, with comparable wording. The snapshot validator rejects a group that mixes outcome sets or horizons. Every forecast keeps its original question wording, and any mapping from wording to outcome codes is recorded in `transformation_note` and displayed beside the transformed value. Singletons and conditional questions are shown individually and never averaged. A group needs at least two members.

Groups in the current snapshot are documented in `research/forecasts/compatibility-groups.md`.

## Methods

For member central values $x_j$ (median, else mean, else p50):

| Method | Formula | Note |
| --- | --- | --- |
| unweighted_median (**preferred**) | median of $x_j$ | robust to one extreme forecast |
| linear_pool | mean of $x_j$ | |
| log_odds_pool | $\operatorname{sigmoid}(\text{mean}(\operatorname{logit} x_j))$ | |
| trimmed_mean | mean after dropping 20% from each end; median when $n < 5$ | |
| tier_weighted | weights $m(\text{tier})$, capped | |
| recency_weighted | weights $0.5^{\text{age years}/3}$ relative to the newest member, capped | |
| equal_weight_by_population | mean of per-population medians | a large survey and a small superforecaster panel count equally |

**Single-figure cap:** in weighted methods no member exceeds 35% of the weight (excess is redistributed) so that one public figure cannot dominate. Extreme confidence is never rewarded: no method weights by how far a forecast sits from 50%.

## What is published

For each group: every method's value, the preferred value, the member forecasts with wording, and an estimate object whose p05/p95 is the min–max of member values and whose p25/p75 are member quartiles. Disagreement is labelled from the interquartile range on the log-odds scale (`uncertainty.md`). Leave-one-source-out, leave-one-survey-out and alternative-method runs are published in `sensitivity.json`.

## Why not one number across groups

Surveys ask "extremely bad outcome, e.g. extinction" (broad, undated), tournaments ask "extinction by 2100" (narrow, dated) and platform questions differ again. Averaging them would hide the definitions that explain most of the spread. The forecasts page shows each group with its wording.
