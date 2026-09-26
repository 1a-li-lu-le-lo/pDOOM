<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Source hierarchy and evidence rules

| Tier | What | Examples |
| --- | --- | --- |
| 1 | Primary and authoritative | research papers, benchmark datasets, source code, model cards, government publications, standards, court records, regulatory findings, audited disclosures, official incident reports, original survey datasets |
| 2 | Independent technical analysis | peer-reviewed reviews, independent evaluations, university and nonprofit research, forecasting research |
| 3 | High-quality journalism | investigative and technically competent reporting linked to primary evidence |
| 4 | Commentary | expert blogs, essays, newsletters, talks, podcasts |
| 5 | Unverified signals | social media, anonymous claims, screenshots, rumours |

Rules enforced by the snapshot validator and the model:
- Tier 5 sources are `excluded` from model use; tier 4 sources cannot be `eligible` or `used` (they may corroborate).
- Tier multipliers in the indexes: 1.0, 0.9, 0.5, 0, 0.
- Only `verified_fetch` or `verified_search` items may be `eligible`/`used`; `unverified` items are `excluded`.
- Company documents carry `developer_self_report`; advocacy organisations `advocacy_context`; governments `government_policy_context`.
- A retracted source cannot feed the model.
- Predictions remain predictions whatever the author's prestige.

In this snapshot direct page fetches were blocked by the build environment's egress policy, so verification is `verified_search` (a search-result snippet quoting the figure was checked) or `verified_prior_knowledge` (informational only). Every source record states which.
