<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Public methodology: what p(DOOM) computes and what it does not

p(DOOM) is a structured forecasting and evidence-synthesis system. It is not a thermometer: nothing here measures "doom". Every number on the site is the output of a documented computation over a versioned data snapshot, published only after a signed human review.

## Eight outputs that are never blended

| Output | What it is | Probability? | Where computed |
| --- | --- | --- | --- |
| A. External forecast aggregate | Named surveys, tournaments and platform forecasts aggregated inside compatibility groups | Yes, aggregated from sources | `internal/model/aggregate.go` |
| B. p(DOOM) model estimate | The research-mode experimental causal model (below); the **official** object is withheld (`insufficiently_calibrated`) | Yes, research mode only | `internal/model/causal.go` |
| C. Evidence Pressure Index | Movement of evidence relative to the baseline release | No | `internal/model/indexes.go` |
| D. Capability Pressure Index | Frontier capability, autonomy, exposure and scaling signals | No | same |
| E. Control Strength Index | Safeguards, security, governance and evaluation signals | No | same |
| F. Incident Pressure Index | Verified incidents weighted by severity, relevance, evidence and recency | No | same |
| G. Uncertainty Score | Coverage, disagreement, evidence quality, judgment share, sensitivity and the calibration gap | No | same |
| H. Editorial risk level | Rule-based plain-language classification | No | `model_spec.editorial_rules` |

Indexes are never converted into probabilities. No calibration process exists that would justify such a conversion (see `calibration.md`).

## Outcome definition

p(DOOM) = P(O3 ∪ O4 ∪ O5 ∪ O6 ∪ O7 ∪ O8 within horizon T | model specification), where O3 is permanent severe disempowerment, O4 civilizational collapse, O5 near-extinction, O6 human extinction, O7 biospheric catastrophe and O8 other irreversible loss. It is always displayed with P-EXTINCTION (O6), P-DISEMPOWERMENT (O3), P-COLLAPSE (O4 ∪ O5) and P-BIOSPHERE (O7) beside it. See `definitions-of-outputs.md`.

## The official value in this release line: withheld

The official object has status `insufficiently_calibrated` for every horizon (Model G, index-only). Reason: long-horizon existential outcomes have never occurred in a way that permits scoring, the research-mode parameters are judgments, and the external forecasts inherit the selection and framing effects of their sources. Publishing a headline probability would be pseudo-precision. Decision record: `docs/governance/decision-log.md`, ADR-002. What would change this is set out in `calibration.md`.

## The experimental causal decomposition (research mode)

$$P(O_i \le T) = P(A \le T)\cdot P(C \mid A)\cdot P(E \mid A, C)\cdot P(F \mid A, C, E)\cdot P(O_i \mid A, C, E, F)$$

In words: a sufficiently capable system emerges (A); it is deployed with meaningful autonomy or access (C); a dangerous exposure, misuse, malfunction or loss-of-control event occurs (E); technical and institutional safeguards fail to prevent catastrophic escalation (F); and, given all that, the outcome is O_i rather than a recoverable harm.

Point estimates are never multiplied blindly. Each factor is a distribution fitted to a documented `{p05, p50, p95}` judgment stored in `model_spec.json` (`experimental_causal`), on the logit scale:

$$\mu_X = \operatorname{logit}(p_{50}),\qquad \sigma_X = \frac{\operatorname{logit}(p_{95}) - \operatorname{logit}(p_{05})}{2 \times 1.6448536269514722}$$

Dependence is modelled by one common latent factor $Z \sim N(0,1)$ with loading $\lambda$ (Gaussian copula):

$$x = \operatorname{sigmoid}\big(\mu_X + \sigma_X (\lambda Z + \sqrt{1-\lambda^2}\,\varepsilon_X)\big),\qquad \varepsilon_X \sim N(0,1)$$

so that race dynamics and similar common causes raise capability, deployment, exposure and safeguard failure together. Outcome shares $s_i$ for O3–O8 are sampled the same way and normalised to sum to at most one (the remainder is the probability of a serious but recoverable harm given safeguard failure). Then $P(O_i) = A\,C\,E\,F\,s_i$ and p(DOOM) is their sum.

Monte Carlo procedure: `samples` draws (20,000) with the fixed `seed` (20260926) from a mulberry32 generator that is bit-identical in Go and TypeScript; normals by Box–Muller; quantiles by nearest rank. The draw order is fixed (Z, ε_A, ε_C, ε_E, ε_F, ε_O3…ε_O8) so the run is reproducible byte-for-byte. Limits: no feedback loops, no time-varying hazards, no explicit competing risks; the scenario graph is descriptive, not simulated.

## Time-to-event view

For each outcome the cumulative probability by time t is $F_i(t) = P(T_i \le t)$ and the hazard is

$$h_i(t) = \lim_{\Delta t \to 0}\frac{P(t \le T_i < t + \Delta t \mid T_i \ge t)}{\Delta t}.$$

Plainly: $F_i$ is the chance the outcome has happened by a date; $h_i$ is the instantaneous chance it happens now given it has not happened yet. The site shows the cumulative curve across the seven horizons and an annualised hazard only where the curve is monotone. Stationarity is never assumed: capability, deployment, regulation and safeguards change the hazard, so no date is implied by any curve.

## Scenario Lab

Users move ten integer sliders (−2…+2). Each slider shifts the logit-scale location of one or more factors by a documented amount (`docs/method/sensitivity.md`, "Slider contract"). Results are labelled `user_scenario`, carry the sentence "Under your selected assumptions—not the p(DOOM) official model—the median estimate is …", and never modify any published value. Implausible combinations are flagged, not refused.

## Model versions in this release line

`pdoom-model/official-index-only@0.1.0`, `pdoom-model/external-aggregate@0.1.0`, `pdoom-model/experimental-causal@0.1.0`, `pdoom-model/indexes@0.1.0`. A change to any of them is a heightened-review trigger (`docs/governance/update-governance.md`).

## Limitations

No calibration; judgment-based parameters; small compatibility groups; incidents limited to verified, registry-referenced events; single-factor dependence. See the model card shipped with each release.
