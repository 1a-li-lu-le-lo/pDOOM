<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Uncertainty, disagreement and display rounding

## Epistemic versus aleatory

Almost all uncertainty here is epistemic: it comes from limited evidence, judgment-based parameters, small samples and unmodelled dependencies, and it could shrink with better evidence. Aleatory uncertainty (irreducible randomness in how the future unfolds) is present too but is not what the intervals mainly express. The Monte Carlo intervals of the research-mode model express parameter uncertainty under the model, not calibrated forecast error.

## Labels

| Label | Uncertainty score | Disagreement (IQR of member log-odds) | Display rounding |
| --- | --- | --- | --- |
| low | < 25 | < 0.5 | nearest 1 point |
| moderate | < 50 | < 1.0 | nearest 1 point |
| high | < 75 | < 2.0 | nearest 2 points |
| extreme | ≥ 75 | ≥ 2.0 | nearest 5 points |

Decimals are never displayed. Values that round below one point show "<1%"; above ninety-nine show ">99%". Intervals use an en dash ("3%–30%") and collapse to a single value when both ends round the same. Rounding rules live in `model_spec.rounding_rules` and are enforced by `internal/model/rounding.go` and `@pdoom/model-core`.

## Every estimate carries

Forecast origin date, last evidence date, horizon, conditioning, median, mean where meaningful, p05/p25/p75/p95, disagreement, model confidence, source coverage, the previous value and the reason for change.
