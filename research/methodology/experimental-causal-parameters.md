<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Experimental causal model: parameter rationale

`docs/method/causal-model.md` points here for the reasoning behind each research-mode parameter.
The parameters live in `model_spec.experimental_causal` of `snap-2026-09-26-001`
(`tools/snapshot/content/drivers.mjs`); the `rationale` object quoted below is stored in the
snapshot and copied into the release manifest's `priors`. Every value is a documented judgment,
not a measurement. The model is research mode; the official value is withheld (ADR-002).

## 1. Structure recap

Per horizon T: P(O_i ≤ T) = P(A) · P(C | A) · P(E | A, C) · P(F | A, C, E) · P(O_i | A, C, E, F),
with each factor a logit-normal distribution fitted to `{p05, p50, p95}` and one common latent
factor with loading λ = 0.5. Seed 20260926, 20,000 samples; identical in Go and TypeScript.
Version `pdoom-model/experimental-causal@0.1.0`.

## 2. Factor parameters by horizon

| Horizon | A (p05 / p50 / p95) | C | E | F |
| --- | --- | --- | --- | --- |
| 1y | 0.02 / 0.08 / 0.25 | 0.35 / 0.65 / 0.90 | 0.08 / 0.25 / 0.55 | 0.04 / 0.18 / 0.45 |
| 3y | 0.08 / 0.25 / 0.55 | 0.35 / 0.65 / 0.90 | 0.10 / 0.30 / 0.60 | 0.05 / 0.20 / 0.50 |
| 5y | 0.15 / 0.40 / 0.75 | 0.35 / 0.65 / 0.90 | 0.10 / 0.30 / 0.60 | 0.05 / 0.20 / 0.50 |
| 10y | 0.30 / 0.60 / 0.90 | 0.40 / 0.70 / 0.92 | 0.12 / 0.35 / 0.65 | 0.05 / 0.22 / 0.55 |
| 25y | 0.50 / 0.80 / 0.97 | 0.40 / 0.72 / 0.93 | 0.15 / 0.40 / 0.70 | 0.05 / 0.25 / 0.60 |
| 2100 | 0.60 / 0.88 / 0.99 | 0.40 / 0.75 / 0.94 | 0.15 / 0.42 / 0.72 | 0.05 / 0.25 / 0.60 |
| eventual | 0.70 / 0.93 / 0.995 | 0.40 / 0.75 / 0.94 | 0.15 / 0.45 / 0.75 | 0.05 / 0.25 / 0.60 |

Outcome shares conditional on safeguard failure (the same for every horizon):

| Outcome | p05 / p50 / p95 |
| --- | --- |
| O3 permanent severe disempowerment | 0.08 / 0.25 / 0.50 |
| O4 civilizational collapse | 0.03 / 0.10 / 0.25 |
| O5 near-extinction | 0.01 / 0.04 / 0.12 |
| O6 human extinction | 0.01 / 0.05 / 0.15 |
| O7 biospheric catastrophe | 0.005 / 0.02 / 0.08 |
| O8 other irreversible loss | 0.03 / 0.10 / 0.25 |

The p50 shares sum to 0.56; the remainder is a serious but recoverable harm. The snapshot
validator requires p05 < p50 < p95 for every triple and Σ p50 shares ≤ 1.

## 3. The recorded rationale

Quoted from `model_spec.experimental_causal.rationale`:

- **A.** "Judgment: 'sufficiently advanced capability' ≈ high-level machine intelligence able to
  run long-horizon autonomous operations. Anchors: ESPAI 2023 aggregate 10% by 2027 and 50% by
  2047 (Grace et al. 2024); ESPAI 2024 medians closer than before; METR horizon doubling of 4–7
  months. The p50 path rises from 8% at one year to 60% at ten years and 88% by 2100; the eventual
  value leaves room for permanent stagnation."
- **C.** "Judgment: given such capability, deployment with meaningful autonomy and access is
  likely because agentic deployment is already the commercial default (METR horizons, AISI and
  Replit incidents). p50 ≈ 0.65–0.75 with wide bands."
- **E.** "Judgment: given deployed autonomous capability, the probability that a dangerous
  exposure, misuse, malfunction or loss-of-control event occurs within the horizon. Anchors:
  verified precursor incidents in 2025–2026 (AISI, GTG-1002) at sub-catastrophic scale; p50
  0.25–0.45 rising with horizon."
- **F.** "Judgment: given such an event, the probability that technical and institutional
  safeguards fail to prevent catastrophic escalation. Anchors: control strength index inputs
  (weight security below SL3 in 2024, early enforcement of binding rules, uneven agent-security
  practice); p50 0.18–0.25."
- **O.** "Judgment: conditional on safeguard failure, shares of outcomes O3–O8 (remainder =
  serious but recoverable outcomes). Disempowerment (O3) is weighted highest following the
  gradual-disempowerment literature and survey wording pairing extinction with permanent
  disempowerment; extinction (O6) p50 0.05 reflects XPT and ESPAI extinction-specific medians
  being far below broader 'extremely bad' medians."
- **Dependence.** "A single common latent factor with loading 0.5 induces positive correlation
  between capability, deployment, exposure and safeguard failure (race dynamics raise all four
  together). λ = 0 and λ + 0.3 are published as sensitivity runs. This is a coarse stand-in for
  the scenario graph; feedback loops and competing risks are not modelled."

Cited anchors (`source_ids`): `src-grace-2024-thousands-of-ai-authors`, `src-aiimpacts-2024-espai`,
`src-fri-2023-xpt`, `src-metr-time-horizons-page`, `src-aisi-2026-incident-report`,
`src-rand-2024-securing-model-weights`, `src-kulveit-2025-gradual-disempowerment`, `src-iasr-2026`.
All eight are `verified_search`, `eligible` sources; the release copies them into the research-mode
estimates' `source_coverage.source_ids`.

## 4. How the parameters are used and tested

- The A anchor is a survey aggregate about "high-level machine intelligence", which is broader
  than "able to run long-horizon autonomous operations"; the judgment narrows it and the p95
  values leave room for faster timelines.
- The C, E and F anchors are the same records that feed the control-strength and
  agentic-infrastructure indexes, so the research-mode model and the indexes are correlated by
  construction; the release's known limitations say the intervals express parameter uncertainty
  under the model, not calibration.
- Sensitivity runs published in the release perturb F by ∓1.0 logit (optimistic/pessimistic
  safeguards), A by ∓0.7, C and E by ∓0.5, and λ to 0 and 0.8, for every horizon (35 runs each
  kind, 70 for dependence). The largest research-mode influences in `rel-2026-09-26-001` are the
  pessimistic-safeguards runs at the 2100 and eventual horizons and the optimistic-safeguards run
  at the eventual horizon (model card, "Sensitivity").
- The invariants require research-mode medians to be non-decreasing across horizons within a
  Monte Carlo tolerance of 0.005 and P_DOOM to equal the sum of its components.
- The Scenario Lab shifts these same parameters on the logit scale by documented amounts per dial
  ([`../../docs/method/sensitivity.md`](../../docs/method/sensitivity.md)); the golden fixture
  `internal/model/testdata/causal-golden.json` pins the computation across languages.

## Limitations

- Every parameter is a judgment by the authors of the snapshot; no parameter is fitted to data
  and none has a calibration record.
- Outcome shares do not vary with horizon, although the relative likelihood of, say,
  disempowerment versus extinction plausibly does.
- One common factor stands in for the scenario graph; feedback, time-varying hazards and
  competing risks are absent.
- The anchors are themselves `verified_search` records that have not been human-reviewed.
- A second parameter set for an "alternative prior" sensitivity run does not yet exist.

## Open questions

- Should O shares be horizon-specific in the next model version, and on what evidence?
- Which resolvable proxies (benchmark thresholds, framework activations, enforcement actions)
  could be registered now so that the A, C and F judgments become scorable
  ([`../../docs/method/calibration.md`](../../docs/method/calibration.md))?
- Is λ = 0.5 too strong or too weak a dependence given that race dynamics are argued to raise all
  four factors together?
- Should the eventual-horizon A value (p50 0.93) be lowered to reflect permanent stagnation more
  strongly, or is the p05 of 0.70 sufficient?
