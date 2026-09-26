<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Model card — release rel-2026-09-26-001

**Data snapshot:** snap-2026-09-26-001 (source cutoff 2026-09-26)
**Generated:** 2026-09-26T21:45:00Z · **Code commit:** cd74db7
**Model versions:** `pdoom-model/official-index-only@0.1.0`, `pdoom-model/external-aggregate@0.1.0`, `pdoom-model/experimental-causal@0.1.0`, `pdoom-model/indexes@0.1.0`
**Editorial risk level:** elevated · **Uncertainty score:** 56 / 100

## Intended use
Public-interest inspection of evidence relevant to catastrophic and existential risk from advanced AI: comparing named external forecasts, reading structured indexes of capability, control, incidents and evidence movement, exploring a clearly labelled research-mode decomposition, and finding interventions. Every number is shown with its outcome definition, horizon, conditioning, interval, disagreement and data cutoff.

## Prohibited uses
Financial trading; military decision-making; emergency-alert triggering; individual mental-health assessment; insurance pricing; legal proof; claims of scientific consensus; predicting a specific catastrophe date; presenting any index as a probability; presenting the research-mode estimate as the official value.

## Outcome definition
p(DOOM) = probability of O3 (permanent severe disempowerment), O4 (civilizational collapse), O5 (near-extinction), O6 (human extinction), O7 (biospheric catastrophe) or O8 (other irreversible loss) within the stated horizon, conditional on the model specification. The official value in this release line is index-only (insufficiently calibrated); external aggregates and the research-mode model are published separately and never blended.

## Horizons
`1y` `3y` `5y` `10y` `25y` `2100` `eventual` — measured from the forecast origin date; "2100" is a calendar endpoint and "eventual" has no endpoint and is never displayed as a date.

## Architecture
1. **Official object** — `pdoom-model/official-index-only`: publishes no probability (status `insufficiently_calibrated`).
2. **External forecast aggregate** — `pdoom-model/external-aggregate`: compatibility groups of named forecasts aggregated under 35 method results; preferred method unweighted median; original wording shown beside every value.
3. **Research-mode experimental model** — `pdoom-model/experimental-causal@0.1.0`: P(O_i ≤ T) = P(A)·P(C|A)·P(E|A,C)·P(F|A,C,E)·P(O_i|A,C,E,F) with logit-normal factors fitted to documented {p05,p50,p95} judgments, one common latent factor, Monte Carlo with a fixed seed.
4. **Indexes** — `pdoom-model/indexes`: capability pressure, control strength, incident pressure, evidence pressure, agentic infrastructure risk, uncertainty score, attention (not computed). None is a probability.

## Data and sources
70 sources, 15 forecast records, 11 incidents in snapshot `snap-2026-09-26-001`. Only items with `verification.status` verified_fetch/verified_search and `model_use_status` eligible/used feed computations; tier 4–5 sources never do. External forecast sources: `src-aiimpacts-2022-espai`, `src-aiimpacts-2024-espai`, `src-fri-2023-xpt`, `src-grace-2024-thousands-of-ai-authors`, `src-metaculus-2026-pro-forecasters-ai-extinction`.

## Priors, assumptions and dependencies
All research-mode parameters are documented judgments stored in `model_spec.json` (see `priors` in the manifest). Dependence: single common factor with loading λ = 0.50. Not modelled: feedback loops, time-varying hazards, explicit competing risks.

## Fitting, calibration and evaluation
No parameters are fitted to outcome data: existential outcomes have not occurred in a way that permits calibration. Proxy calibration (benchmark milestones, policy events, incident occurrence) is planned in `docs/method/calibration.md` and is not yet available; nothing here claims calibration.

## Sensitivity (largest influences)
- `sens-loo-G-CAT10-2100-fc-xpt-2023-experts-ai-catastrophe-2100` (leave_one_source_out) on `G-CAT10-2100`: Δ -0.0493
- `sens-loo-G-CAT10-2100-fc-xpt-2023-superforecasters-ai-catastrophe-2100` (leave_one_source_out) on `G-CAT10-2100`: Δ +0.0493
- `sens-loo-G-EXT-2100-METACULUS-fc-metaculus-2026-dedicated-xrisk-forecasters-ai-extinction` (leave_one_source_out) on `G-EXT-2100-METACULUS`: Δ -0.0725
- `sens-loo-G-EXT-2100-METACULUS-fc-metaculus-2026-pro-forecasters-ai-extinction` (leave_one_source_out) on `G-EXT-2100-METACULUS`: Δ +0.0725
- `sens-loo-G-EXT-DISEMP-UNDATED-fc-espai-2023-extinction-or-disempowerment` (leave_one_source_out) on `G-EXT-DISEMP-UNDATED`: Δ +0.0250
- `sens-loo-G-EXT-DISEMP-UNDATED-fc-espai-2024-extinction-or-disempowerment` (leave_one_source_out) on `G-EXT-DISEMP-UNDATED`: Δ -0.0250
- `sens-optimistic_safeguards-f-logit-1-0-P_DOOM-eventual` (optimistic_safeguards) on `est-research-P_DOOM-eventual`: Δ -0.0221
- `sens-pessimistic_safeguards-f-logit-1-0-P_DOOM-2100` (pessimistic_safeguards) on `est-research-P_DOOM-2100`: Δ +0.0305
- `sens-pessimistic_safeguards-f-logit-1-0-P_DOOM-25y` (pessimistic_safeguards) on `est-research-P_DOOM-25y`: Δ +0.0254
- `sens-pessimistic_safeguards-f-logit-1-0-P_DOOM-eventual` (pessimistic_safeguards) on `est-research-P_DOOM-eventual`: Δ +0.0350

## Estimates published
- `est-official-P_DOOM-1y` [insufficiently_calibrated] 1y: Insufficiently calibrated
- `est-official-P_DOOM-3y` [insufficiently_calibrated] 3y: Insufficiently calibrated
- `est-official-P_DOOM-5y` [insufficiently_calibrated] 5y: Insufficiently calibrated
- `est-official-P_DOOM-10y` [insufficiently_calibrated] 10y: Insufficiently calibrated
- `est-official-P_DOOM-25y` [insufficiently_calibrated] 25y: Insufficiently calibrated
- `est-official-P_DOOM-2100` [insufficiently_calibrated] 2100: Insufficiently calibrated
- `est-official-P_DOOM-eventual` [insufficiently_calibrated] eventual: Insufficiently calibrated
- `est-external-G-CAT10-2100` [external_aggregate] 2100: 8%
- `est-external-G-EXT-2100` [external_aggregate] 2100: 2%
- `est-external-G-EXT-2100-METACULUS` [external_aggregate] 2100: 14%
- `est-external-G-EXT-DISEMP-UNDATED` [external_aggregate] eventual: 8%
- `est-external-G-XBAD-LONG-RUN` [external_aggregate] eventual: 6%
- `est-research-P_DOOM-1y` [research_mode] 1y: <1%
- `est-research-O3-1y` [research_mode] 1y: <1%
- `est-research-P_COLLAPSE-1y` [research_mode] 1y: <1%
- `est-research-O6-1y` [research_mode] 1y: <1%
- `est-research-O7-1y` [research_mode] 1y: <1%
- `est-research-P_DOOM-3y` [research_mode] 3y: <1%
- `est-research-O3-3y` [research_mode] 3y: <1%
- `est-research-P_COLLAPSE-3y` [research_mode] 3y: <1%
- `est-research-O6-3y` [research_mode] 3y: <1%
- `est-research-O7-3y` [research_mode] 3y: <1%
- `est-research-P_DOOM-5y` [research_mode] 5y: <1%
- `est-research-O3-5y` [research_mode] 5y: <1%
- `est-research-P_COLLAPSE-5y` [research_mode] 5y: <1%
- `est-research-O6-5y` [research_mode] 5y: <1%
- `est-research-O7-5y` [research_mode] 5y: <1%
- `est-research-P_DOOM-10y` [research_mode] 10y: 2%
- `est-research-O3-10y` [research_mode] 10y: <1%
- `est-research-P_COLLAPSE-10y` [research_mode] 10y: <1%
- `est-research-O6-10y` [research_mode] 10y: <1%
- `est-research-O7-10y` [research_mode] 10y: <1%
- `est-research-P_DOOM-25y` [research_mode] 25y: 2%
- `est-research-O3-25y` [research_mode] 25y: 2%
- `est-research-P_COLLAPSE-25y` [research_mode] 25y: <1%
- `est-research-O6-25y` [research_mode] 25y: <1%
- `est-research-O7-25y` [research_mode] 25y: <1%
- `est-research-P_DOOM-2100` [research_mode] 2100: 4%
- `est-research-O3-2100` [research_mode] 2100: 2%
- `est-research-P_COLLAPSE-2100` [research_mode] 2100: <1%
- `est-research-O6-2100` [research_mode] 2100: <1%
- `est-research-O7-2100` [research_mode] 2100: <1%
- `est-research-P_DOOM-eventual` [research_mode] eventual: 4%
- `est-research-O3-eventual` [research_mode] eventual: 2%
- `est-research-P_COLLAPSE-eventual` [research_mode] eventual: <1%
- `est-research-O6-eventual` [research_mode] eventual: <1%
- `est-research-O7-eventual` [research_mode] eventual: <1%

## Known limitations and biases
- No official probability is published: the official object has status insufficiently_calibrated because no documented calibration process exists for long-horizon existential outcomes.
- The research-mode model's parameters are judgments informed by cited sources, not measurements; its intervals express parameter uncertainty under the model, not calibrated forecast error.
- External forecast aggregates inherit the selection, framing and population effects of their member surveys; groups are small and dominated by a few studies.
- Indexes (0–100) are constructed scores and are never probabilities; coverage below 1 means specified signals lacked eligible observations.
- Dependence between factors is represented by a single common latent factor; feedback loops, competing risks and time-varying hazards are not modelled.
- Incident pressure counts only verified, registry-referenced incidents and therefore under-represents unreported or unverified events.
- Editorial risk level (elevated) is a rule-based, human-readable classification, not a measurement.

## Governance and revision policy
Promotion requires signed reviewer approvals (ed25519), byte-for-byte reproducibility of the candidate from its snapshot, invariant checks, and acknowledgement of heightened-review triggers by a second reviewer. Changes are recorded in `delta.json` and `changelog.md`; releases are immutable and rollback moves only the `CURRENT` pointer. See `docs/governance/update-governance.md`.

## Contact and license
Authored by NU Cybernetics. Corrections: `docs/governance/corrections-and-appeals.md`. License: see `NOTICE`.
