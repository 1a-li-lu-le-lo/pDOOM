<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Backtesting

Backtests reconstruct what the observatory would have published at earlier dates using only information available then.

Procedure for each historical date:
1. Assemble a snapshot restricted to sources with `date_published` (or `date_retrieved` where publication is undated) before the cutoff; record every exclusion.
2. Reconstruct driver observations from those sources only; mark reconstructed values `observation_kind: judgment` where the original measurement is unavailable.
3. Run the model with the historical snapshot and the model version under test; record indexes and estimates.
4. Compare proxy predictions (capability milestones, policy events, incident counts) with later observations; score with Brier and log scores.
5. Hindsight checks: flag any source whose date is after the cutoff, any observation whose value was revised later, and any narrative that could only be written with later knowledge.
6. Overreaction check: compare index movements around high-attention news events with the movement justified by tier-1 evidence alone.

Rules: no future information; data revisions are recorded as corrections, never silently applied; backtests are published as research reports and never promoted as releases.

Status: no backtests exist for the first release; the first is scheduled once a second snapshot exists.
