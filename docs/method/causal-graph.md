<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Causal graph and scenario dependence

## The conceptual DAG

```
        race dynamics (S15) ──┐
                              ▼
   A capability ──▶ C deployment/autonomy ──▶ E exposure event ──▶ F safeguard failure ──▶ O_i
        ▲                     ▲                    ▲                     ▲
        │                     │                    │                     │
  self-accelerating      access & tools      malicious use (S4)     governance capacity
  research (S14)          (D2, D3)          security gaps (D6)      epistemic collapse (S10)
```

Common-cause nodes (race dynamics, governance capacity, epistemic collapse) act on several factors at once. The experimental model compresses them into one latent factor Z; the scenario graph (`scenario_edges.json`) records the specific edges with a relation (enables, amplifies, prevents_response, shares_prerequisite, competes_with), a confidence and a rationale.

## Why pathway probabilities are not summed

Scenarios overlap: cyber compromise can enable biological misuse; economic concentration can enable authoritarian lock-in; epistemic collapse can prevent a coordinated response; autonomy and access are prerequisites of nearly every pathway. Summing per-pathway probabilities would double count the shared prerequisites and ignore competing risks (one catastrophe pre-empting another). Pathway probabilities are therefore **not assigned** in this release (`probability_source: not_assigned`); only the aggregate decomposition is simulated.

## What would justify richer structure

- A dynamic Bayesian network when time matters (hazards changing with policy or capability events) and when conditional probability tables can be sourced rather than invented.
- Competing-risk models once outcome-specific hazards can be estimated.
- Copulas beyond the single Gaussian factor only if documented dependence evidence exists; otherwise they add parameters without adding knowledge.

The rendering of the graph on the Futures page is a node graph of scenarios and edges; it is labelled "conceptual", not "simulated".
