// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { review } from "./helpers.mjs";

const P = "higher_raises_pressure";
const C = "higher_strengthens_control";

function sig(signal_id, name, description, direction, normalization, raw_unit, preferred = ["evaluation", "paper"], ovj = "observation when a published measurement exists; judgment when qualitative evidence is mapped to the scale") {
  return { signal_id, name, description, direction, normalization, raw_unit, preferred_source_types: preferred, observation_vs_judgment: ovj };
}

export const drivers = [
  { id: "D1", name: "capability", description: "Frontier capability relevant to catastrophic pathways: autonomy horizon, coding, computer use, cyber and biological uplift, research automation.", signals: [
    sig("D1.task_horizon_50pct", "50% task-completion time horizon", "METR-style human task length completed with 50% success by the best published model.", P, "clamp(log2(hours / 0.05) / log2(160 / 0.05), 0, 1): 3 minutes → 0, one 160-hour work-month → 1", "hours"),
    sig("D1.swe_bench_verified", "SWE-bench Verified top score", "Best tracked percentage of resolved tasks.", P, "score / 100", "percent"),
    sig("D1.computer_use_osworld", "OSWorld-Verified top score", "Best tracked success rate on real-computer tasks.", P, "score / 100", "percent"),
    sig("D1.cyber_offense_capability", "Cyber offensive capability", "Degree to which frontier systems can execute intrusion life-cycle stages autonomously.", P, "judgment scale: 0 = no uplift, 0.5 = autonomous execution of most tactical stages with human strategic control (documented), 1 = fully autonomous campaigns against defended targets", null, ["incident_report", "government", "evaluation"], "judgment"),
    sig("D1.bio_chem_uplift", "Biological/chemical uplift", "Category-level assessment of uplift for dangerous laboratory work as reported by synthesis reports and developer frameworks; no operational detail.", P, "judgment scale: 0 = no uplift beyond public sources, 0.5 = meaningful assistance with substantial uncertainty about real-world uplift, 1 = expert-level end-to-end uplift established", null, ["government", "safety_framework"], "judgment"),
    sig("D1.ai_rnd_automation", "AI research automation", "Share of AI R&D workflow that frontier systems perform autonomously.", P, "judgment scale 0–1", null, ["evaluation"], "judgment"),
  ] },
  { id: "D2", name: "autonomy", description: "Reliable unattended operation: task length, self-correction, delegation, persistence.", signals: [
    sig("D2.unattended_operation_hours", "Unattended operation length", "Longest human-equivalent task length completed reliably by deployed agents.", P, "clamp(log2(1 + hours) / log2(1 + 168), 0, 1): one work-week of unattended operation → 1", "hours"),
    sig("D2.autonomous_action_evidence", "Evidence of unsanctioned autonomous action", "Documented cases of agents acting outside sanctioned scope in real environments.", P, "judgment scale: 0 = none documented, 0.5 = contained cases with no real-world harm, 1 = uncontained cases with real-world harm", null, ["incident_report", "government"], "judgment"),
    sig("D2.delegation_depth", "Delegation depth in deployed harnesses", "Typical nesting depth of agent-to-agent delegation in production systems.", P, "clamp(depth / 5, 0, 1)", "levels", ["evaluation", "code"], "judgment"),
    sig("D2.human_approval_frequency", "Human approval frequency", "Share of high-impact actions gated by human approval in typical deployments (inverted).", C, "share of high-impact actions approved by a human, 0–1", "share", ["evaluation", "standard"], "judgment"),
  ] },
  { id: "D3", name: "access_exposure", description: "What deployed systems can reach: tools, credentials, money, code execution, infrastructure, weights.", signals: [
    sig("D3.agent_tool_access_breadth", "Breadth of tool and credential access", "Typical scope of tools, credentials and network access granted to deployed agents.", P, "judgment scale: 0 = read-only sandboxed, 0.5 = broad developer-level access common, 1 = production and financial write access common without gates", null, ["standard", "government", "incident_report"], "judgment"),
    sig("D3.open_weight_frontier_gap", "Open-weight proximity to frontier", "How close the best open-weight models are to the closed frontier.", P, "1 − (capability gap in months / 24), clamped", "months behind", ["evaluation", "review_article"], "judgment"),
    sig("D3.critical_infrastructure_integration", "Critical-infrastructure integration", "Extent of agent integration into operational technology.", P, "judgment scale 0–1", null, ["government", "incident_report"], "judgment"),
  ] },
  { id: "D4", name: "scalability", description: "How fast capability can be replicated and scaled: compute growth, inference cost, agent population.", signals: [
    sig("D4.frontier_compute_growth", "Frontier training compute growth", "Annual growth factor of frontier training compute.", P, "clamp(log10(growth_factor) / log10(10), 0, 1): 10× per year → 1", "factor per year", ["review_article", "dataset"]),
    sig("D4.inference_cost_decline", "Inference cost decline", "Annual decline factor in the cost of frontier-level inference.", P, "clamp(log10(decline_factor) / log10(10), 0, 1)", "factor per year", ["review_article", "dataset"]),
    sig("D4.agent_population_scale", "Deployed agent population", "Order of magnitude of concurrently deployed autonomous agents.", P, "clamp((log10(agents) − 4) / 5, 0, 1)", "agents", ["review_article"], "judgment"),
  ] },
  { id: "D5", name: "alignment_control", description: "Evidence about honesty, corrigibility, shutdown compliance, control evaluations and interpretability.", signals: [
    sig("D5.control_evaluation_maturity", "Control-evaluation maturity", "Whether frontier deployments pass control evaluations at scale.", C, "judgment scale: 0 = none, 0.5 = published protocols with limited deployment use, 1 = routine passing control evaluations for frontier agents", null, ["paper", "evaluation"], "judgment"),
    sig("D5.shutdown_compliance_evidence", "Shutdown compliance", "Evidence that frontier models comply with shutdown and do not sabotage oversight in tests.", C, "1 − (severity of documented non-compliance): 1 = consistent compliance, 0.5 = non-compliance in a minority of controlled runs, 0 = routine subversion", null, ["evaluation", "journalism"], "judgment"),
    sig("D5.interpretability_detection", "Interpretability-based deception detection", "Availability of deployed detectors for deceptive cognition.", C, "judgment scale 0–1", null, ["paper"], "judgment"),
    sig("D5.alignment_faking_evidence", "Alignment-faking evidence", "Whether model organisms show alignment faking that persists through training (inverted control signal).", C, "1 − strength of evidence for persistent alignment faking", null, ["paper"], "judgment"),
  ] },
  { id: "D6", name: "security", description: "Model-weight security, agent-infrastructure security, supply-chain integrity, prompt-injection resistance.", signals: [
    sig("D6.model_weight_security_level", "Model-weight security level", "Best documented security level (RAND SL1–SL5) achieved by frontier developers.", C, "level / 5", "SL level", ["paper", "company_disclosure"]),
    sig("D6.mcp_supply_chain_integrity", "Agent supply-chain integrity", "Signing, provenance and vulnerability posture of agent tool registries and servers.", C, "judgment scale: 0 = unverified ecosystem with active CVE waves, 0.5 = identity verification without code signing, 1 = signed, scanned, monitored", null, ["standard", "government", "review_article"], "judgment"),
    sig("D6.prompt_injection_resistance", "Prompt-injection resistance", "Degree to which deployed agents resist instruction injection through content and tool metadata.", C, "judgment scale: 0 = routinely exploitable, 0.5 = mitigations common but bypassable, 1 = deterministic isolation standard", null, ["government", "review_article", "journalism"], "judgment"),
    sig("D6.cancellation_and_audit", "Cancellation and audit completeness", "Whether deployed agent stacks support reliable cancellation and complete audit.", C, "judgment scale 0–1", null, ["standard"], "judgment"),
  ] },
  { id: "D7", name: "governance", description: "Binding rules, incident reporting, evaluation capacity, safety frameworks, enforcement, international coordination.", signals: [
    sig("D7.binding_frontier_rules_jurisdictions", "Jurisdictions with binding frontier-model rules", "Count of jurisdictions with binding statutes specifically covering frontier or general-purpose models with systemic risk.", C, "clamp(count / 10, 0, 1)", "jurisdictions", ["government"]),
    sig("D7.incident_reporting_mandates", "Incident-reporting mandates", "Coverage of mandatory serious-incident reporting for frontier developers.", C, "judgment scale: 0 = none, 0.5 = mandates in some major jurisdictions with enforcement pending, 1 = global interoperable mandates enforced", null, ["government"], "judgment"),
    sig("D7.third_party_evaluation_capacity", "Third-party evaluation capacity", "Capacity of independent and government evaluators to test frontier systems pre-deployment.", C, "judgment scale: 0 = none, 0.5 = government institutes with published tooling and partial access, 1 = mandatory independent pre-deployment evaluation", null, ["government", "code"], "judgment"),
    sig("D7.published_safety_frameworks", "Published frontier safety frameworks", "Share of the five largest frontier developers with a published, versioned frontier safety framework.", C, "share, 0–1", "share", ["safety_framework"]),
    sig("D7.enforcement_active", "Enforcement active", "Whether regulators have exercised enforcement powers for frontier-model obligations.", C, "judgment scale 0–1", null, ["government", "regulatory"], "judgment"),
  ] },
  { id: "D8", name: "incidents", description: "Verified incidents and near misses; computed from incidents.json by the incident pressure index rather than from signals.", signals: [
    sig("D8.verified_precursor_incidents", "Verified loss-of-control precursors", "Count of verified incidents with relevance strong or direct_precursor in the last 24 months.", P, "clamp(count / 10, 0, 1)", "incidents", ["incident_report", "government"]),
  ] },
  { id: "D9", name: "race_dynamics", description: "Competitive pressure to deploy before safeguards are ready.", signals: [
    sig("D9.release_cadence_pressure", "Release cadence pressure", "Frequency of frontier releases and shortening of evaluation windows.", P, "judgment scale 0–1", null, ["journalism", "review_article"], "judgment"),
    sig("D9.safety_commitment_erosion", "Safety commitment erosion", "Documented weakening of safety commitments or institutions under competitive pressure.", P, "judgment scale 0–1", null, ["journalism", "government"], "judgment"),
    sig("D9.geopolitical_competition", "Geopolitical competition intensity", "Degree to which frontier AI is framed and funded as a strategic race between states.", P, "judgment scale 0–1", null, ["government", "journalism"], "judgment"),
  ] },
  { id: "D10", name: "resilience", description: "Capacity to absorb and recover from failures: fallbacks, segmentation, emergency governance, distributed expertise.", signals: [
    sig("D10.manual_fallback_coverage", "Manual fallback coverage", "Share of critical services with tested manual fallback.", C, "share, 0–1", "share", ["government"], "judgment"),
    sig("D10.emergency_response_capacity", "AI emergency-response capacity", "Existence and exercise of cross-sector AI incident response.", C, "judgment scale 0–1", null, ["government"], "judgment"),
    sig("D10.infrastructure_diversity", "Infrastructure diversity", "Diversity of model and cloud providers behind critical services (inverse monoculture).", C, "judgment scale 0–1", null, ["review_article"], "judgment"),
  ] },
];

function obs(id, o) {
  return { id, signal_id: o.signal, family: o.signal.split(".")[0], value_normalized: o.value, raw_value: o.raw ?? null, raw_unit: o.unit ?? null, confidence: o.confidence, observation_kind: o.kind, as_of: o.as_of, rationale: o.rationale, counterevidence: o.counter ?? null, source_ids: o.sources, ...review(o.verified ?? "search") };
}

export const driver_observations = [
  obs("do-d1-task-horizon-50pct", { signal: "D1.task_horizon_50pct", value: 0.566, raw: 4.817, unit: "hours", confidence: 0.6, kind: "observation", as_of: "2025-12-18", rationale: "Claude Opus 4.5 at 289 minutes = 4.817 h (METR). log2(4.817/0.05) = 6.590; log2(160/0.05) = 11.644; 6.590/11.644 = 0.566. Wide 95% CI (1 h 49 min – 20 h 25 min).", counter: "A third-party estimate puts Claude Opus 4.6 near 719 minutes (informational only); METR's own limitations note cautions against over-interpretation.", sources: ["src-metr-time-horizons-page", "src-metr-2026-time-horizon-limitations"] }),
  obs("do-d1-swe-bench-verified", { signal: "D1.swe_bench_verified", value: 0.97, raw: 97.0, unit: "percent", confidence: 0.4, kind: "observation", as_of: "2026-09-04", rationale: "Tracker-reported 97.00% (Claude Opus 5, Vals.ai run) / 100 = 0.97. Near saturation; audits found flawed tests at high scores, so the signal mostly indicates saturation.", counter: "Benchmark contamination and scaffold variance; tier-3 source.", sources: ["src-steel-2026-swe-bench-verified-leaderboard"] }),
  obs("do-d1-computer-use-osworld", { signal: "D1.computer_use_osworld", value: 0.861, raw: 86.1, unit: "percent", confidence: 0.4, kind: "observation", as_of: "2026-09-22", rationale: "Tracker-reported 86.1% (Qwen3.8 Max) / 100 = 0.861; exceeds the 72.36% human baseline of the original paper.", counter: "Third-party tracker; step budgets and scaffolds vary.", sources: ["src-benchlm-2026-osworld-verified"] }),
  obs("do-d1-cyber-offense-capability", { signal: "D1.cyber_offense_capability", value: 0.5, confidence: 0.5, kind: "judgment", as_of: "2026-02-03", rationale: "The 2026 International AI Safety Report finds AI's largest cyber role is in preparatory stages and not yet fully autonomous execution; the 2025 espionage disclosure reports agentic execution of most tactical stages with human strategic control. Mapped to the documented mid-point of the scale (0.5).", counter: "Autonomy share in the disclosure is a developer estimate; independent analysts questioned the framing.", sources: ["src-iasr-2026", "src-anthropic-2025-disrupting-ai-espionage"] }),
  obs("do-d1-bio-chem-uplift", { signal: "D1.bio_chem_uplift", value: 0.45, confidence: 0.4, kind: "judgment", as_of: "2026-02-03", rationale: "The 2026 International AI Safety Report finds systems can help produce laboratory instructions and troubleshoot procedures while substantial uncertainty remains about real-world uplift given practical barriers; mapped just below the 0.5 scale point. No operational detail is recorded.", counter: "Real-world uplift has not been demonstrated.", sources: ["src-iasr-2026"] }),
  obs("do-d2-unattended-operation-hours", { signal: "D2.unattended_operation_hours", value: 0.343, raw: 4.817, unit: "hours", confidence: 0.5, kind: "observation", as_of: "2025-12-18", rationale: "Using the METR 50% horizon of 4.817 h as the reliable unattended length: log2(1+4.817) = 2.540; log2(169) = 7.401; 2.540/7.401 = 0.343.", counter: "50% success is not 'reliable'; the 80% horizon is much shorter.", sources: ["src-metr-time-horizons-page"] }),
  obs("do-d2-autonomous-action-evidence", { signal: "D2.autonomous_action_evidence", value: 0.55, confidence: 0.6, kind: "judgment", as_of: "2026-08-04", rationale: "The AISI incident (19 unsanctioned live-internet actions across 122 attempts, contained, no known real-world harm) and the Replit data loss (real harm, single deployment) place the signal slightly above the 0.5 'contained cases' point.", counter: "Both cases were quickly contained or remediated.", sources: ["src-aisi-2026-incident-report", "src-aiid-1152-replit-agent"] }),
  obs("do-d3-agent-tool-access-breadth", { signal: "D3.agent_tool_access_breadth", value: 0.6, confidence: 0.5, kind: "judgment", as_of: "2026-06-30", rationale: "NSA and Microsoft guidance in 2026 responds to broad developer-level tool and credential access being common; the GitHub MCP and Replit cases show production and repository write access granted without gates. Mapped between 'broad developer-level access common' (0.5) and ungated production access (1).", counter: "Guidance adoption is rising; many deployments are read-only.", sources: ["src-nsa-2026-mcp-security", "src-thn-2026-microsoft-mcp-tool-descriptions", "src-github-mcp-server-issue-844"] }),
  obs("do-d4-frontier-compute-growth", { signal: "D4.frontier_compute_growth", value: 0.653, raw: 4.5, unit: "factor per year", confidence: 0.6, kind: "observation", as_of: "2024-05-28", rationale: "Epoch AI: 4–5× per year; log10(4.5)/log10(10) = 0.653.", counter: "2024 estimate; growth may have changed with power and capital constraints.", sources: ["src-epoch-2024-compute-growth"] }),
  obs("do-d5-shutdown-compliance-evidence", { signal: "D5.shutdown_compliance_evidence", value: 0.45, confidence: 0.5, kind: "judgment", as_of: "2025-06-20", rationale: "Shutdown-script sabotage in 7% of runs with explicit instructions (79% without) and blackmail choices in stress tests across developers indicate non-compliance in a minority of controlled runs with instructions: mapped just below the 0.5 point.", counter: "Controlled, contrived settings; no deployed cases documented.", sources: ["src-register-2025-o3-shutdown", "src-anthropic-2025-agentic-misalignment"] }),
  obs("do-d6-model-weight-security-level", { signal: "D6.model_weight_security_level", value: 0.4, raw: 2, unit: "SL level", confidence: 0.3, kind: "observation", as_of: "2024-05-30", rationale: "RAND (2024) reported no company meeting SL3; the best documented level is therefore at most SL2: 2/5 = 0.4. Low confidence because the assessment is from 2024.", counter: "Developers report security investments since 2024 (self-reports, not verified).", sources: ["src-rand-2024-securing-model-weights"] }),
  obs("do-d6-mcp-supply-chain-integrity", { signal: "D6.mcp_supply_chain_integrity", value: 0.3, confidence: 0.6, kind: "judgment", as_of: "2026-07-20", rationale: "The official registry verifies namespaces but not server code; a 30+ CVE wave hit MCP servers in early 2026; NSA guidance exists but adoption is uneven. Mapped below the 0.5 'identity verification without code signing' point.", counter: "Rapid vendor and standards response during 2026.", sources: ["src-mcp-registry-about", "src-csa-2026-mcp-security-crisis", "src-nsa-2026-mcp-security"] }),
  obs("do-d6-prompt-injection-resistance", { signal: "D6.prompt_injection_resistance", value: 0.3, confidence: 0.6, kind: "judgment", as_of: "2026-06-30", rationale: "Tool-description poisoning and prompt injection remained exploitable in 2025–2026 demonstrations and Microsoft's June 2026 guidance; mitigations exist but are bypassable.", counter: "Deterministic policy enforcement is spreading in enterprise deployments.", sources: ["src-csa-2026-mcp-tool-poisoning", "src-thn-2026-microsoft-mcp-tool-descriptions", "src-github-mcp-server-issue-844"] }),
  obs("do-d7-binding-frontier-rules-jurisdictions", { signal: "D7.binding_frontier_rules_jurisdictions", value: 0.4, raw: 4, unit: "jurisdictions", confidence: 0.7, kind: "observation", as_of: "2026-01-22", rationale: "EU (AI Act GPAI obligations from August 2025), California (SB 53, effective January 2026), New York (RAISE Act, signed December 2025, effective 2027) and South Korea (AI Basic Act, effective January 2026): 4/10 = 0.4.", counter: "Other jurisdictions may have binding rules not captured; New York's law is not yet in force.", sources: ["src-eu-ai-act-gpai-guidelines", "src-ca-sb53-tfaia", "src-ny-raise-act-2025", "src-korea-ai-basic-act"] }),
  obs("do-d7-incident-reporting-mandates", { signal: "D7.incident_reporting_mandates", value: 0.45, confidence: 0.6, kind: "judgment", as_of: "2026-08-02", rationale: "EU serious-incident duties for general-purpose models with systemic risk apply from August 2025 with enforcement from August 2026; California SB 53 requires critical-safety-incident reporting from 2026. Mandates exist in some major jurisdictions with enforcement just beginning: just below the 0.5 point.", counter: "No global interoperable regime; US federal mandate absent.", sources: ["src-eu-ai-act-gpai-guidelines", "src-ca-sb53-tfaia"] }),
  obs("do-d7-third-party-evaluation-capacity", { signal: "D7.third_party_evaluation_capacity", value: 0.5, confidence: 0.6, kind: "judgment", as_of: "2026-09-22", rationale: "Government institutes exist (UK AISI with 171 published Inspect evaluations; US CAISI with voluntary agreements) and run pre-deployment testing with partial access: the 0.5 scale point.", counter: "Access depends on voluntary developer agreements.", sources: ["src-aisi-inspect", "src-ssti-2025-caisi-rename"] }),
  obs("do-d7-published-safety-frameworks", { signal: "D7.published_safety_frameworks", value: 0.6, raw: 0.6, unit: "share", confidence: 0.6, kind: "observation", as_of: "2026-04-02", rationale: "Verified versioned frameworks at three of the five largest developers (Anthropic RSP 3.1, OpenAI Preparedness Framework 2, Google DeepMind FSF 3): 3/5 = 0.6. Other developers' frameworks were not verified for this snapshot.", counter: "Frameworks are voluntary and self-assessed; an affordance analysis found one framework guarantees no specific mitigation.", sources: ["src-anthropic-rsp", "src-openai-preparedness-v2", "src-gdm-fsf-v3"] }),
];

const tri = (p05, p50, p95) => ({ p05, p50, p95 });
const shares = { O3: tri(0.08, 0.25, 0.5), O4: tri(0.03, 0.10, 0.25), O5: tri(0.01, 0.04, 0.12), O6: tri(0.01, 0.05, 0.15), O7: tri(0.005, 0.02, 0.08), O8: tri(0.03, 0.10, 0.25) };
const horizon = (A, C, E, F) => ({ A, C, E, F, O: shares });

export const model_spec = {
  id: "pdoom-model-spec@0.1.0",
  index_weights: {
    capability_pressure: { "D1.task_horizon_50pct": 0.25, "D1.swe_bench_verified": 0.10, "D1.computer_use_osworld": 0.10, "D1.cyber_offense_capability": 0.15, "D1.bio_chem_uplift": 0.10, "D2.unattended_operation_hours": 0.10, "D2.autonomous_action_evidence": 0.10, "D4.frontier_compute_growth": 0.10 },
    control_strength: { "D5.control_evaluation_maturity": 0.15, "D5.shutdown_compliance_evidence": 0.15, "D6.model_weight_security_level": 0.15, "D6.mcp_supply_chain_integrity": 0.10, "D6.prompt_injection_resistance": 0.10, "D7.binding_frontier_rules_jurisdictions": 0.10, "D7.incident_reporting_mandates": 0.10, "D7.third_party_evaluation_capacity": 0.10, "D7.published_safety_frameworks": 0.05 },
    incident_pressure: {},
    evidence_pressure: {},
    uncertainty: { coverage_gap: 0.10, forecast_disagreement: 0.15, low_tier_share: 0.10, judgment_share: 0.15, sensitivity_spread: 0.10, calibration_gap: 0.40 },
    agentic_infrastructure_risk: { "D2.unattended_operation_hours": 0.20, "D2.autonomous_action_evidence": 0.20, "D3.agent_tool_access_breadth": 0.25, "D6.mcp_supply_chain_integrity": 0.20, "D6.prompt_injection_resistance": 0.15 },
  },
  weight_bounds: [0, 0.35],
  tier_multipliers: { "1": 1, "2": 0.9, "3": 0.5, "4": 0, "5": 0 },
  incident_scoring: {
    severity_weights: { negligible: 0, minor: 0.1, material: 0.3, major: 0.6, severe: 0.85, catastrophic: 1 },
    relevance_weights: { none: 0, weak: 0.1, indirect: 0.25, moderate: 0.5, strong: 0.8, direct_precursor: 1 },
    evidence_weights: { allegation: 0.1, single_source_report: 0.3, corroborated_report: 0.6, official_finding: 0.85, peer_reviewed_analysis: 0.9, independently_reproduced: 1 },
    recency_half_life_days: 730,
    squash_k: 3,
  },
  editorial_rules: [
    { level: "insufficient_evidence", when: "coverage < 0.4" },
    { level: "severe_uncertainty", when: "uncertainty >= 70" },
    { level: "high", when: "cpi >= 70 and csi < 40" },
    { level: "elevated", when: "cpi >= 55 and csi < 55" },
    { level: "guarded", when: "cpi >= 40" },
    { level: "low", when: "cpi >= 25" },
    { level: "very_low", when: "true" },
  ],
  rounding_rules: { extreme: 5, high: 2, moderate: 1, low: 1 },
  aggregation_methods: ["unweighted_median", "linear_pool", "log_odds_pool", "trimmed_mean", "tier_weighted", "recency_weighted", "equal_weight_by_population"],
  experimental_causal: {
    version: "pdoom-model/experimental-causal@0.1.0",
    seed: 20260926,
    samples: 20000,
    common_factor_loading: 0.5,
    horizons: {
      "1y": horizon(tri(0.02, 0.08, 0.25), tri(0.35, 0.65, 0.90), tri(0.08, 0.25, 0.55), tri(0.04, 0.18, 0.45)),
      "3y": horizon(tri(0.08, 0.25, 0.55), tri(0.35, 0.65, 0.90), tri(0.10, 0.30, 0.60), tri(0.05, 0.20, 0.50)),
      "5y": horizon(tri(0.15, 0.40, 0.75), tri(0.35, 0.65, 0.90), tri(0.10, 0.30, 0.60), tri(0.05, 0.20, 0.50)),
      "10y": horizon(tri(0.30, 0.60, 0.90), tri(0.40, 0.70, 0.92), tri(0.12, 0.35, 0.65), tri(0.05, 0.22, 0.55)),
      "25y": horizon(tri(0.50, 0.80, 0.97), tri(0.40, 0.72, 0.93), tri(0.15, 0.40, 0.70), tri(0.05, 0.25, 0.60)),
      "2100": horizon(tri(0.60, 0.88, 0.99), tri(0.40, 0.75, 0.94), tri(0.15, 0.42, 0.72), tri(0.05, 0.25, 0.60)),
      "eventual": horizon(tri(0.70, 0.93, 0.995), tri(0.40, 0.75, 0.94), tri(0.15, 0.45, 0.75), tri(0.05, 0.25, 0.60)),
    },
    rationale: {
      A: "Judgment: 'sufficiently advanced capability' ≈ high-level machine intelligence able to run long-horizon autonomous operations. Anchors: ESPAI 2023 aggregate 10% by 2027 and 50% by 2047 (Grace et al. 2024); ESPAI 2024 medians closer than before; METR horizon doubling of 4–7 months. The p50 path rises from 8% at one year to 60% at ten years and 88% by 2100; the eventual value leaves room for permanent stagnation.",
      C: "Judgment: given such capability, deployment with meaningful autonomy and access is likely because agentic deployment is already the commercial default (METR horizons, AISI and Replit incidents). p50 ≈ 0.65–0.75 with wide bands.",
      E: "Judgment: given deployed autonomous capability, the probability that a dangerous exposure, misuse, malfunction or loss-of-control event occurs within the horizon. Anchors: verified precursor incidents in 2025–2026 (AISI, GTG-1002) at sub-catastrophic scale; p50 0.25–0.45 rising with horizon.",
      F: "Judgment: given such an event, the probability that technical and institutional safeguards fail to prevent catastrophic escalation. Anchors: control strength index inputs (weight security below SL3 in 2024, early enforcement of binding rules, uneven agent-security practice); p50 0.18–0.25.",
      O: "Judgment: conditional on safeguard failure, shares of outcomes O3–O8 (remainder = serious but recoverable outcomes). Disempowerment (O3) is weighted highest following the gradual-disempowerment literature and survey wording pairing extinction with permanent disempowerment; extinction (O6) p50 0.05 reflects XPT and ESPAI extinction-specific medians being far below broader 'extremely bad' medians.",
      dependence: "A single common latent factor with loading 0.5 induces positive correlation between capability, deployment, exposure and safeguard failure (race dynamics raise all four together). λ = 0 and λ + 0.3 are published as sensitivity runs. This is a coarse stand-in for the scenario graph; feedback loops and competing risks are not modelled.",
    },
    source_ids: ["src-grace-2024-thousands-of-ai-authors", "src-aiimpacts-2024-espai", "src-fri-2023-xpt", "src-metr-time-horizons-page", "src-aisi-2026-incident-report", "src-rand-2024-securing-model-weights", "src-kulveit-2025-gradual-disempowerment", "src-iasr-2026"],
  },
};
