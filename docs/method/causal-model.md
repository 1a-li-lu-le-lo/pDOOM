<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# The experimental causal model (research mode)

**Status: research mode. Not the official estimate. Every parameter is a documented judgment.**

## Structure

Five factors per horizon: A (sufficiently advanced capability emerges by T), C (deployed with meaningful autonomy or access, given A), E (a dangerous exposure, misuse, malfunction or loss-of-control event occurs, given A and C), F (technical and institutional safeguards fail to prevent catastrophic escalation, given A, C and E) and the outcome shares O3–O8 given all of the above. See `model.md` for the equations and `research/methodology/experimental-causal-parameters.md` for the rationale behind each value.

## Parameters (`model_spec.experimental_causal`)

| Horizon | A p50 | C p50 | E p50 | F p50 |
| --- | --- | --- | --- | --- |
| 1y | 0.08 | 0.65 | 0.25 | 0.18 |
| 3y | 0.25 | 0.65 | 0.30 | 0.20 |
| 5y | 0.40 | 0.65 | 0.30 | 0.20 |
| 10y | 0.60 | 0.70 | 0.35 | 0.22 |
| 25y | 0.80 | 0.72 | 0.40 | 0.25 |
| 2100 | 0.88 | 0.75 | 0.42 | 0.25 |
| eventual | 0.93 | 0.75 | 0.45 | 0.25 |

Outcome shares (conditional on safeguard failure), p50: O3 0.25, O4 0.10, O5 0.04, O6 0.05, O7 0.02, O8 0.10; the remaining 0.44 is a serious but recoverable harm. Common factor loading λ = 0.5. Seed 20260926, 20,000 samples. Each value has a p05 and p95 in the spec; the snapshot validator requires p05 < p50 < p95 and Σ p50 shares ≤ 1.

## Anchors used for the judgments

- A: ESPAI 2023 aggregate of 10% for high-level machine intelligence by 2027 and 50% by 2047; ESPAI 2024 medians closer than before; METR time-horizon doubling of four to seven months.
- C: agentic deployment is already the commercial default; incidents in 2025–2026 involved agents with broad access.
- E: verified precursor incidents at sub-catastrophic scale (AISI 2026 evaluation incident; GTG-1002 campaign).
- F: control-strength inputs (weight security below SL3 in 2024, enforcement of binding rules only beginning in 2026, uneven agent-security practice).
- O: gradual-disempowerment literature and survey wording that pairs extinction with permanent disempowerment; extinction-specific medians (XPT, ESPAI) far below broad "extremely bad" medians.

## Reading the output

The medians it produces are small percentages with wide bands (see the release). They are conditional on this structure and these judgments; they are not calibrated forecasts and must never be quoted without the words "research mode".

## Known weaknesses

Judgment parameters; single common factor; no feedback; no competing risks; horizons treated independently (monotonicity is checked as an invariant, not imposed); outcome shares independent of horizon.
