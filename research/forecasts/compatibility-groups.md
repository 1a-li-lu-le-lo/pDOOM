<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Forecast compatibility groups

This document declares the `group_id` values used in `data/snapshots/snap-2026-09-26-001/forecasts.json`.
A compatibility group is the only unit inside which p(DOOM) may aggregate external forecasts
(build-spec §0.5). Members of one group must share **all four** of: outcome set, horizon,
conditioning, and comparable question wording. Population may differ inside a group (the
aggregation layer publishes per-population breakdowns and the `equal_weight_by_population`
method), but wording, outcome and horizon may not.

Forecasts that have no compatible partner carry `group_id: null` and are shown as singletons
beside their original wording. Group membership says nothing about quality: an item that is
`model_use_status: "informational"` or `"excluded"` still carries its group id for
documentation, and the aggregation layer must filter on `model_use_status` before pooling.

## Outcome-code conventions used across groups

| Original wording | p(DOOM) outcome set | Reason |
| --- | --- | --- |
| "human extinction", "AI wipes out humanity", "global population below 5,000" | `[O5, O6]` | Operational extinction definitions used by XPT and Metaculus (population < 5,000) include near-extinction; ordinary-language "wipe out" is read the same way so the group shares one set. |
| "extremely bad (e.g., human extinction)" | `[O4, O5, O6, O8]` | The example is extinction, but the bucket is explicitly broader ("extremely bad"); collapse, near-extinction and other irreversible losses are plausible readings. Not O3, because the bucket sits above "on balance bad" in a five-point scale of overall impact and disempowerment without severe harm is unlikely to be placed there. |
| "human extinction or similarly permanent and severe disempowerment of the human species" | `[O3, O5, O6]` | Explicitly extinction plus permanent severe disempowerment. |
| "existential catastrophe" (Ord, Carlsmith, FRI adversarial collaboration) | `[O3, O4, O5, O6, O8]` | Destruction of humanity's long-term potential: extinction, unrecoverable collapse, unrecoverable dystopia/disempowerment, other irreversible loss. |
| "misaligned AI takeover" | `[O3, O5, O6]` | Takeover implies permanent disempowerment and may include extinction. |
| "death of at least 10% of humans alive at the beginning of a five-year period" | `[O1, O4, O5, O6]` | A casualty threshold, not a trajectory class; it includes recoverable (O1) and unrecoverable (O4–O6) continuations. Never combine with extinction-only groups. |
| "AI-driven event causing at least 50 deaths or USD 100 billion in damages" | `[O1]` | Serious but, at this threshold, reversible harm. |

## Declared groups

| group_id | Definition (outcome · horizon · conditioning) | Members in this snapshot | Notes |
| --- | --- | --- | --- |
| `G-EXT-2100` | AI-caused human extinction (population < 5,000 or equivalent) · by 2100 · unconditional | `fc-xpt-2022-superforecasters-ai-extinction-2100`, `fc-xpt-2022-ai-domain-experts-ai-extinction-2100`, `fc-fri-2024-conditional-trees-ai-experts-extinction-2100` (informational), `fc-metaculus-q27035-community-ai-extinction-2100` (informational), `fc-manifold-dillard-ai-wipe-out-humanity-2100` (informational) | Only the two XPT items are `eligible`. The Metaculus and Manifold items lack a verified as-of date; the Conditional Trees figure lacks confirmation against the paper. |
| `G-CAT10-2100` | AI-caused catastrophe killing ≥10% of humanity within five years · by 2100 · unconditional | `fc-xpt-2022-superforecasters-ai-catastrophe-2100`, `fc-xpt-2022-domain-experts-ai-catastrophe-2100` (excluded, value unverified) | Effectively a singleton until the domain-expert value is verified by a human reviewer. |
| `G-XBAD-HLMI` | Long-run impact of HLMI on humanity is "extremely bad (e.g., human extinction)" · eventual · conditional on HLMI being built | `fc-espai-2016-hlmi-extremely-bad`, `fc-espai-2022-hlmi-extremely-bad`, `fc-espai-2023-hlmi-extremely-bad`, `fc-espai-2024-hlmi-extremely-bad` | Same question wording across four survey waves by the same team; the group is a time series, not four independent samples of one population, so pooling across waves is only meaningful as a trend. |
| `G-XDIS-AI-EVENTUAL` | "future AI advances causing human extinction or similarly permanent and severe disempowerment of the human species" · eventual (no horizon stated) · unconditional | `fc-espai-2022-ai-extinction-or-disempowerment`, `fc-espai-2023-ai-extinction-or-disempowerment`, `fc-espai-2024-ai-extinction-or-disempowerment` | Same caveat as above: a time series of one survey series. |
| `G-XDIS-CONTROL-EVENTUAL` | "human inability to control future advanced AI systems causing human extinction or similarly permanent and severe disempowerment" · eventual · unconditional (cause-specific: control failure) | `fc-espai-2022-control-failure-extinction-or-disempowerment`, `fc-espai-2023-control-failure-extinction-or-disempowerment` | Cause-restricted wording; must not be pooled with `G-XDIS-AI-EVENTUAL` even though the outcome set is identical. |
| `G-TAKEOVER-2070` | Existential catastrophe from misaligned, power-seeking AI (Carlsmith's six-premise argument) · by 2070 · unconditional | `fc-carlsmith-2021-power-seeking-ai-existential-catastrophe-2070` (superseded), `fc-carlsmith-2022-update-power-seeking-ai-existential-catastrophe-2070` (informational, lower bound only), `fc-good-judgment-2023-superforecasters-carlsmith-premises-2070` | The superforecaster figure is the only current numeric member. |
| `G-XCAT-2100` | AI-caused existential catastrophe · by 2100 · unconditional (FRI adversarial collaboration wording) | `fc-fri-2023-roots-concerned-ai-existential-catastrophe-2100`, `fc-fri-2023-roots-skeptics-ai-existential-catastrophe-2100` | Two self-selected sides of one study; the group exists to keep them side by side, not to average them. |
| `G-HARM50-2050` | AI-driven event causing ≥50 deaths or ≥USD 100 billion damages · by end of 2050 · unconditional | `fc-leap-2026-wave9-experts-ai-harm-event-2050`, `fc-leap-2026-wave9-superforecasters-ai-harm-event-2050` | Outcome `[O1]`; outside the p(DOOM) (O3–O8) definition; kept as an incident-pressure context group. |
| `G-EXT-2100-ALLCAUSE` | Human extinction from **any** cause · by 2100 · unconditional | `fc-xpt-2022-superforecasters-any-cause-extinction-2100`, `fc-xpt-2022-domain-experts-any-cause-extinction-2100`, `fc-metaculus-2026-community-any-cause-extinction-2100`, `fc-metaculus-2026-pro-forecasters-any-cause-extinction-2100` | Context group: an upper bound on AI-caused extinction. Never blend with `G-EXT-2100`. |
| `G-CAT10-2100-ALLCAUSE` | ≥10% of humanity dies within five years, **any** cause · by 2100 · unconditional | `fc-xpt-2022-superforecasters-any-cause-catastrophe-2100`, `fc-xpt-2022-domain-experts-any-cause-catastrophe-2100` | Context group. Never blend with `G-CAT10-2100`. |

## Singletons (group_id null)

| Forecast id | Why it has no group |
| --- | --- |
| `fc-ord-2020-unaligned-ai-existential-catastrophe-100y` | Horizon is "next 100 years" from 2020 (custom, end year 2120) and the outcome is Ord's existential-catastrophe definition; no other verified item shares both. The reserved label `G-XRISK-100Y` will be used if a compatible item is added. |
| `fc-samotsvety-2022-misaligned-ai-takeover-2100` | Conditioning "barring pre-APS-AI catastrophe" and takeover wording differ from `G-XCAT-2100` and `G-EXT-2100`. |
| `fc-samotsvety-taisc-2023-ai-catastrophe-95pct` | Horizon could not be verified; excluded. |
| `fc-xpt-2022-superforecasters-ai-extinction-2030`, `fc-xpt-2022-superforecasters-ai-catastrophe-2030`, `fc-xpt-2022-superforecasters-ai-extinction-2050` | Only one verified population per horizon. Reserved labels `G-EXT-2030`, `G-CAT10-2030`, `G-EXT-2050`. |
| `fc-xpt-2022-superforecasters-ai-extinction-2100-conditional-agi-2070` | Conditional on AGI by 2070. |
| `fc-leap-2026-wave9-experts-ai-catastrophe-10pct-*` (six items) | Each is conditional on a slow- or rapid-progress scenario; conditioning differs from `G-CAT10-2100`, and slow/rapid items differ from each other. |
| `fc-metaculus-2026-dedicated-xrisk-forecasters-extinction-2100` | The search snippet leaves the exact question (any cause vs. broader AI doom) ambiguous. |
| `fc-ses-2026-ai-safety-leaders-extinction-2100` | Cause (AI-specific vs. any) not verified. |

## Rules for adding a member

1. Copy the outcome set from the table above; do not invent a new mapping for a wording that is already mapped.
2. Horizon keys: `2100`, `eventual`, or `custom` with `horizon_end_year`. "Within 100 years" from a dated source is `custom`.
3. Conditioning strings must match exactly for group membership; "unconditional" is the literal used.
4. A forecast whose wording restricts the cause (e.g. "inability to control") is a different group from one that does not.
5. Any-cause forecasts live only in `*-ALLCAUSE` groups.
6. Prediction-market and platform items may join a group but should stay `informational` until the as-of date of the quoted value is verified from the live page.

## Limitations

- Group membership was assigned from question wording as recovered through search-result snippets, because canonical pages could not be fetched from this environment. Wordings marked `paraphrase: true` should be checked against the originals before the first release.
- Outcome mappings are editorial judgements; the mapping table is the single place to revise them.
- Several groups are time series from one survey series (AI Impacts). Pooling across waves conflates population drift, wording stability and genuine belief change.

## Open questions

- Should `G-XBAD-HLMI` be treated as conditional on HLMI for aggregation purposes, or displayed as an unconditional long-run judgement given that most respondents expect HLMI this century?
- Is `[O1, O4, O5, O6]` the right reading of the 10%-mortality threshold, or should the threshold questions be kept out of outcome-coded aggregation altogether and shown only as raw thresholds?
- Once the XPT domain-expert catastrophe figure and the Conditional Trees expert figure are verified from the primary documents, should `G-CAT10-2100` and `G-EXT-2100` be re-aggregated, and how should the release note describe the change?
