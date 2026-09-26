<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Indexes

All indexes are 0–100 constructed scores. **None is a probability and none is ever converted into one.** They are computed by `internal/model/indexes.go` from `driver_observations.json`, `incidents.json` and `model_spec.json`.

## General rule for signal-based indexes

For an index with signal weights $w_i$ (each in $[0, 0.35]$, summing to 1), tier multipliers $m(\text{tier})$ = {1: 1.0, 2: 0.9, 3: 0.5, 4: 0, 5: 0} and normalised observation values $v_i \in [0,1]$:

$$\text{index} = 100 \cdot \frac{\sum_{i \in \text{observed}} w_i\, m(t_i)\, v_i}{\sum_{i \in \text{observed}} w_i\, m(t_i)},\qquad \text{coverage} = \frac{\sum_{i \in \text{observed}} w_i}{\sum_{i} w_i}$$

One observation per signal is used: the most recent `as_of` among observations with `model_use_status` eligible/used and a `verified_fetch`/`verified_search` verification; $t_i$ is the best tier among its sources. Signals whose direction is opposite to the index semantic are inverted ($v \to 1 - v$) so that every index reads in one direction. Missing signals reduce coverage and are listed in the coverage gaps; they do not silently count as zero.

Worked example (capability pressure, three signals, all tier 1, weights 0.35/0.35/0.30, values 0.62/0.70/0.50): $100 \times (0.217 + 0.245 + 0.150) / 1.0 = 61.2$.

### Capability Pressure Index (higher = more pressure)
Signals and weights are in `model_spec.index_weights.capability_pressure`: D1.task_horizon_50pct 0.25, D1.swe_bench_verified 0.10, D1.computer_use_osworld 0.10, D1.cyber_offense_capability 0.15, D1.bio_chem_uplift 0.10, D2.unattended_operation_hours 0.10, D2.autonomous_action_evidence 0.10, D4.frontier_compute_growth 0.10. Normalisation recipes are stated per signal in `drivers.json` (for example the time horizon: $\operatorname{clamp}(\log_2(h/0.05)/\log_2(160/0.05), 0, 1)$).

### Control Strength Index (higher = stronger control)
D5.control_evaluation_maturity 0.15, D5.shutdown_compliance_evidence 0.15, D6.model_weight_security_level 0.15, D6.mcp_supply_chain_integrity 0.10, D6.prompt_injection_resistance 0.10, D7.binding_frontier_rules_jurisdictions 0.10, D7.incident_reporting_mandates 0.10, D7.third_party_evaluation_capacity 0.10, D7.published_safety_frameworks 0.05.

### Agentic Infrastructure Risk Index (higher = more exposure)
D2.unattended_operation_hours 0.20, D2.autonomous_action_evidence 0.20, D3.agent_tool_access_breadth 0.25, D6.mcp_supply_chain_integrity 0.20 (inverted), D6.prompt_injection_resistance 0.15 (inverted). The dimensions of the brief's harness/MCP module (permission scope, credential scope, delegation depth, cancellation, registry signing, prompt-injection resistance…) are signals under D2, D3 and D6; those without verified observations appear as coverage gaps.

## Incident Pressure Index

$$x = \sum_{\text{eligible incidents}} \text{sev}_w \cdot \text{rel}_w \cdot \text{ev}_w \cdot 0.5^{\,\text{age\_days}/730},\qquad \text{index} = 100\,(1 - e^{-x/3})$$

Weights (`model_spec.incident_scoring`): severity negligible 0, minor 0.1, material 0.3, major 0.6, severe 0.85, catastrophic 1; relevance none 0, weak 0.1, indirect 0.25, moderate 0.5, strong 0.8, direct_precursor 1; evidence allegation 0.1, single_source_report 0.3, corroborated_report 0.6, official_finding 0.85, peer_reviewed_analysis 0.9, independently_reproduced 1. Age is measured from the snapshot's source cutoff; an incident without a date gets recency 0.5. Each underlying event counts once (duplicate-event clustering happens in ingestion); media volume is the separate attention index. Worked example: eleven eligible incidents with raw pressure 0.693 give $100(1 - e^{-0.231}) = 20.6$.

## Evidence Pressure Index

The first release is its own baseline and reads 50 ("no movement"). Later releases compute

$$\text{EPI} = 50 + 50 \tanh\!\left(\frac{\Delta\text{CPI} - \Delta\text{CSI} + \Delta\text{IPI}}{50}\right)$$

with each Δ measured against the baseline release's index values (propagated through `baseline` pointers so all releases compare to the same origin).

## Uncertainty Score

$$\text{uncertainty} = 100 \cdot \frac{\sum_k w_k c_k}{\sum_k w_k}$$

Components $c_k \in [0,1]$ and weights (`model_spec.index_weights.uncertainty`):

| Component | Meaning | Weight |
| --- | --- | --- |
| coverage_gap | $1 -$ mean coverage of the three signal indexes | 0.10 |
| forecast_disagreement | mean over compatibility groups of $\min(\text{IQR}_{\text{log-odds}}/4, 1)$; 1 when no groups | 0.15 |
| low_tier_share | share of signal-index weight carried by tier-3 sources | 0.10 |
| judgment_share | share of signal-index weight resting on `observation_kind = judgment` | 0.15 |
| sensitivity_spread | $\min(\max\lvert\Delta p_{50}\rvert / 0.2, 1)$ over research-mode sensitivity runs | 0.10 |
| calibration_gap | $1 -$ share of research-mode estimates backed by proxy calibration; **1 in this release line** | 0.40 |

The calibration gap dominates by design: until proxy calibration exists the score cannot fall below 40. Labels: < 25 low, < 50 moderate, < 75 high, ≥ 75 extreme; the label sets display rounding (`uncertainty.md`).

## Attention Index

Not computed in this release line (null): the prototype ingests no media-volume data. When it does, it will count articles per underlying event so that article counts are never mistaken for incident counts.

## Editorial risk level

Ordered rules in `model_spec.editorial_rules`, first match wins, over the variables coverage, uncertainty, cpi, csi, ipi, epi and air: insufficient_evidence (coverage < 0.4), severe_uncertainty (uncertainty ≥ 70), high (cpi ≥ 70 and csi < 40), elevated (cpi ≥ 55 and csi < 55), guarded (cpi ≥ 40), low (cpi ≥ 25), very_low. A rule-based label, not a measurement.
