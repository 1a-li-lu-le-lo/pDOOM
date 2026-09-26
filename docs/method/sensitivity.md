<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Sensitivity analysis and "What would change the number?"

Every candidate release publishes `sensitivity.json`, ranked by absolute effect (rank 1 = largest influence).

| Kind | Perturbation | Target |
| --- | --- | --- |
| leave_one_source_out | remove one forecast from its group | preferred aggregate |
| leave_one_survey_out | remove every forecast from one source document | preferred aggregate |
| alternative_weighting | each alternative aggregation method | preferred aggregate |
| optimistic_safeguards / pessimistic_safeguards | F logit ∓ 1.0 | research-mode medians |
| slower_capability / faster_capability | A logit ∓ 0.7 | research-mode medians |
| lower_exposure / higher_exposure | C and E logit ∓ 0.5 | research-mode medians |
| alternative_dependency | λ = 0 and λ + 0.3 (capped at 1) | research-mode medians |

Alternative priors are the horizon-specific parameter sets themselves; a future release will add a second parameter set for comparison.

## The panel

The "What would change the number?" panel lists the ten largest runs, the signals with the largest index contributions and the change in each index if a signal were removed (`drivers_explained.json`). If the model cannot attribute a change causally, the panel says so: causal attribution of probability deltas to individual drivers is not available in this model version.

## Slider contract (Scenario Lab)

| Slider | Effect per step |
| --- | --- |
| capability_timeline | A +0.5 |
| autonomy_growth | C +0.4 |
| access_level | C +0.3, E +0.3 |
| safety_progress | F −0.5 |
| governance_strength | F −0.4 |
| model_security | E −0.3 |
| open_weight_diffusion | E +0.3 |
| international_coordination | F −0.3 |
| incident_frequency | E +0.2 |
| resilience | O4, O5, O6 shares −0.3 |
