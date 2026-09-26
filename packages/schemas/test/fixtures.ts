// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * One valid fixture per kind. Fixtures are synthetic test data: URLs use the
 * reserved example.org domain and no real source, organisation or figure is
 * represented. They exist only to exercise the schemas.
 */
import type { Kind, Verification } from "../src/index";

const SHA = "a".repeat(64);
const SIG = "QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVo=";

const verified: Verification = {
  status: "verified_fetch",
  checked_at: "2026-09-20",
  method: "WebFetch canonical_url",
  note: "fixture",
};

const reviewed = {
  verification: verified,
  human_review_status: "reviewed",
  model_use_status: "eligible",
} as const;

export const definition = {
  id: "def-fixture-term",
  term: "Fixture term",
  short_definition: "A term used only to exercise the schema.",
  definitions: [
    { text: "A longer fixture definition.", source_id: "src-fixture-survey-2025", attribution: "Fixture author", note: "" },
    { text: "An alternative fixture definition.", source_id: null, attribution: "Editorial", note: "contested wording" },
  ],
  consensus: "contested",
  related_ids: ["def-fixture-other"],
  see_also_urls: ["https://example.org/glossary/fixture-term"],
  ...reviewed,
};

export const source = {
  id: "src-fixture-survey-2025",
  canonical_url: "https://example.org/reports/fixture-survey-2025",
  title: "Fixture survey of researchers",
  publisher: "Example Institute",
  authors: ["A. Fixture", "B. Fixture"],
  date_published: "2025-01-15",
  date_updated: null,
  date_retrieved: "2026-09-20",
  source_tier: 2,
  source_type: "survey",
  jurisdiction: null,
  topic: ["forecasts", "surveys"],
  claim_ids: ["clm-fixture-survey-01"],
  evidence_summary: "Fixture summary of what the source reports.",
  counterevidence: null,
  methodology: "Online questionnaire.",
  sample: "Fixture sample description.",
  limitations: "Fixture limitations.",
  conflicts: ["none_known"],
  license: "CC BY 4.0",
  robots_status: "allowed",
  content_hash: null,
  archive_reference: null,
  language: "en",
  translation: null,
  duplicate_group: null,
  retraction_status: "none",
  correction_status: null,
  citation: "Fixture, A. & Fixture, B. (2025). Fixture survey of researchers. Example Institute.",
  ...reviewed,
};

export const claim = {
  id: "clm-fixture-survey-01",
  text: "Fixture claim text.",
  subject: "fixture respondents",
  predicate: "reported",
  object: "a fixture median",
  date: "2025-01-15",
  horizon: null,
  geography: null,
  model_name: null,
  model_version: null,
  source_id: "src-fixture-survey-2025",
  evidence_type: "survey_result",
  quantitative_value: 0.05,
  unit: "probability",
  uncertainty: "interquartile range reported",
  direct_quote_pointer: "Section 3, Figure 2",
  context: "Fixture context.",
  corroboration_ids: [],
  contradiction_ids: [],
  relevance: "moderate",
  status: "candidate",
  ...reviewed,
};

export const forecast = {
  id: "fc-fixture-survey-2025-extinction-2100",
  forecaster_or_survey: "Fixture survey 2025",
  source_id: "src-fixture-survey-2025",
  date: "2025-01-15",
  population: "general_ai_researchers",
  sample_size: 1000,
  expertise: "Fixture expertise description.",
  question_wording_original: "Fixture question wording.",
  paraphrase: true,
  outcome_set: ["O6"],
  horizon: "2100",
  horizon_note: null,
  horizon_end_year: 2100,
  conditions: "Unconditional.",
  mean: 0.08,
  median: 0.05,
  quantiles: { p25: 0.01, p50: 0.05, p75: 0.1 },
  response_rate: 0.15,
  selection_effects: "Fixture selection note.",
  framing_effects: null,
  calibration: null,
  group_id: "grp-fixture-O6-2100-unconditional",
  transformation_note: null,
  status: "current",
  ...reviewed,
};

export const benchmark = {
  id: "bm-fixture-task-horizon",
  name: "Fixture task horizon benchmark",
  maintainer: "Example Evaluations",
  version: "1.0",
  url: "https://example.org/benchmarks/fixture",
  tasks: "Fixture task family.",
  contamination_risk: "moderate",
  saturation: "none",
  scaffold: null,
  model_access: "API",
  unit: "minutes",
  direction: "higher_is_more_capable",
  limitations: "Fixture limitations.",
  pdoom_relevance: "moderate",
  weight_note: null,
  source_ids: ["src-fixture-survey-2025"],
  ...reviewed,
};

export const benchmark_result = {
  id: "bmr-fixture-task-horizon-fixture-model",
  benchmark_id: "bm-fixture-task-horizon",
  model_name: "Fixture model",
  model_developer: "Example Developer",
  date: "2026-03-01",
  value: 60,
  unit: "minutes",
  ci_low: 40,
  ci_high: 90,
  scaffold: null,
  confidence: "moderate",
  note: null,
  source_ids: ["src-fixture-survey-2025"],
  ...reviewed,
};

export const incident = {
  id: "inc-fixture-0001-service-outage",
  title: "Fixture incident",
  date: "2026-02-10",
  date_precision: "month",
  external_ids: { other: "fixture-registry-0001" },
  summary: "A fixture incident described at category level.",
  cause: ["malfunction"],
  harm: ["financial"],
  severity: "minor",
  pdoom_relevance: "weak",
  evidence_level: "corroborated_report",
  systems_involved: ["fixture system"],
  jurisdiction: null,
  near_miss: false,
  novelty: "routine",
  exposure_note: null,
  source_ids: ["src-fixture-survey-2025"],
  ...reviewed,
};

export const scenario = {
  id: "S3",
  name: "Fixture scenario",
  outcome_set: ["O3", "O4"],
  description: "Category-level fixture description.",
  prerequisites: ["fixture prerequisite"],
  early_indicators: ["fixture indicator"],
  counterindicators: ["fixture counterindicator"],
  capability_thresholds: ["fixture threshold"],
  exposure: "Fixture exposure note.",
  control_failures: ["fixture control failure"],
  human_contributions: ["fixture human contribution"],
  ai_contributions: ["fixture AI contribution"],
  dependencies: ["S1"],
  time_horizon_note: "Fixture horizon note.",
  probability_source: "not_assigned",
  uncertainty: "extreme",
  intervention_ids: ["I04"],
  recoverability: "low",
  evidence_summary: "Fixture evidence summary.",
  source_ids: ["src-fixture-survey-2025"],
  open_questions: ["fixture open question"],
  content_safety_note: null,
  ...reviewed,
};

export const scenario_edge = {
  id: "se-S1-S3",
  from_id: "S1",
  to_id: "S3",
  relation: "enables",
  confidence: "moderate",
  rationale: "Fixture rationale.",
  source_ids: [],
};

export const driver = {
  id: "D1",
  name: "Capability",
  description: "Fixture driver description.",
  signals: [
    {
      signal_id: "D1.task_horizon_50pct",
      name: "Task horizon (50 %)",
      description: "Fixture signal description.",
      direction: "higher_raises_pressure",
      normalization: "log-scaled minutes mapped to 0..1",
      raw_unit: "minutes",
      preferred_source_types: ["evaluation", "paper"],
      observation_vs_judgment: "observation",
    },
  ],
};

export const driver_observation = {
  id: "obs-d1-task-horizon-2026-03",
  signal_id: "D1.task_horizon_50pct",
  family: "D1",
  value_normalized: 0.7,
  raw_value: 60,
  raw_unit: "minutes",
  confidence: 0.6,
  observation_kind: "observation",
  as_of: "2026-03-01",
  rationale: "Fixture rationale.",
  counterevidence: null,
  source_ids: ["src-fixture-survey-2025"],
  ...reviewed,
};

export const intervention = {
  id: "I04",
  name: "Fixture intervention",
  target_scenario_ids: ["S3"],
  mechanism: "Fixture mechanism at category level.",
  evidence_summary: "Fixture evidence summary.",
  evidence_strength: "weak",
  cost: "moderate",
  time_to_deploy: "1-2y",
  effect_size: "unknown",
  uncertainty: "high",
  possible_failure: "Fixture failure mode.",
  possible_backfire: "Fixture backfire note.",
  owner_types: ["laboratories"],
  user_actions: [{ audience: "individuals", action: "Fixture action." }],
  category: "technical",
  source_ids: ["src-fixture-survey-2025"],
  ...reviewed,
};

export const organization = {
  id: "org-fixture-institute",
  name: "Example Institute",
  url: "https://example.org",
  mission: "Fixture mission.",
  legal_status: "non-profit",
  jurisdiction: "Fixture jurisdiction",
  focus: ["evaluation"],
  programs: ["fixture program"],
  open_outputs: ["fixture report"],
  funding_disclosure: "Fixture funding disclosure.",
  conflicts: [],
  evidence_of_impact: "Fixture evidence of impact.",
  ways_to_help: ["fixture way to help"],
  inclusion_criteria_met: ["publishes open outputs"],
  last_verified: "2026-09-20",
  source_ids: ["src-fixture-survey-2025"],
  ...reviewed,
};

export const action = {
  id: "act-individuals-learn-the-basics",
  audience: "individuals",
  title: "Learn the basics",
  description: "Fixture description.",
  related_intervention_ids: ["I04"],
  resources: [{ title: "Fixture resource", url: "https://example.org/resource", source_id: null }],
  effort: "low",
  ...reviewed,
};

const tri = { p05: 0.1, p50: 0.3, p95: 0.6 };

export const experimental_causal = {
  version: "0.1.0",
  seed: 20260926,
  samples: 20000,
  common_factor_loading: 0.5,
  horizons: {
    "10y": {
      A: tri,
      C: tri,
      E: tri,
      F: tri,
      O: {
        O3: { p05: 0.001, p50: 0.01, p95: 0.05 },
        O4: { p05: 0.001, p50: 0.01, p95: 0.05 },
        O5: { p05: 0.0005, p50: 0.005, p95: 0.03 },
        O6: { p05: 0.0005, p50: 0.005, p95: 0.03 },
        O7: { p05: 0.0001, p50: 0.001, p95: 0.01 },
        O8: { p05: 0.0001, p50: 0.001, p95: 0.01 },
      },
    },
  },
  rationale: { A: "fixture", C: "fixture", E: "fixture", F: "fixture", O: "fixture", dependence: "fixture" },
  source_ids: ["src-fixture-survey-2025"],
};

export const model_spec = {
  id: "pdoom-model-spec@0.1.0",
  index_weights: {
    capability_pressure: { "D1.task_horizon_50pct": 0.2, "D2.autonomy_duration": 0.15 },
    control_strength: { "D5.alignment_progress": 0.25 },
    incident_pressure: { scale: 10 },
    evidence_pressure: { recency_half_life_days: 365 },
    uncertainty: { disagreement_weight: 0.5, sparsity_weight: 0.5 },
    agentic_infrastructure_risk: { "D3.agent_deployments": 0.3 },
  },
  weight_bounds: [0, 0.35],
  tier_multipliers: { "1": 1, "2": 0.9, "3": 0.5, "4": 0, "5": 0 },
  incident_scoring: {
    severity_weights: { negligible: 0, minor: 0.1, material: 0.3, major: 0.6, severe: 0.8, catastrophic: 1 },
    relevance_weights: { none: 0, weak: 0.2, indirect: 0.4, moderate: 0.6, strong: 0.8, direct_precursor: 1 },
    evidence_weights: {
      allegation: 0.1,
      single_source_report: 0.3,
      corroborated_report: 0.6,
      official_finding: 0.8,
      peer_reviewed_analysis: 0.9,
      independently_reproduced: 1,
    },
    recency_half_life_days: 365,
    squash_k: 2,
  },
  editorial_rules: [{ level: "insufficient_evidence", when: "fewer than 3 eligible forecasts" }],
  rounding_rules: { extreme: 5, high: 2, moderate: 1, low: 1 },
  aggregation_methods: ["unweighted_median", "linear_pool"],
  experimental_causal,
};

export const snapshot_manifest = {
  snapshot_id: "snap-2026-09-26-001",
  created_at: "2026-09-26T00:00:00Z",
  source_cutoff: "2026-09-01",
  baseline_snapshot_id: null,
  files: [{ path: "sources.json", sha256: SHA, count: 1 }],
  notes: "fixture",
};

export const estimate = {
  estimate_id: "est-external-O6-2100",
  producer: "pdoom-model/external-aggregate@0.1.0",
  status: "external_aggregate",
  outcome_set: ["O6"],
  outcome_label: "Human extinction",
  horizon: "2100",
  conditioning: "Unconditional.",
  forecast_origin_date: "2026-09-26",
  last_evidence_date: "2026-09-01",
  quantiles: { p05: 0.003, p25: 0.01, p50: 0.05, p75: 0.1, p95: 0.3 },
  mean: 0.08,
  disagreement: "high",
  uncertainty: "extreme",
  model_confidence: "low",
  source_coverage: { forecast_count: 6, population_count: 3, source_ids: ["src-fixture-survey-2025"] },
  previous: null,
  reason_for_change: "first release",
  rounding_rule: "nearest_5",
  display: { central: "5%", interval: "0%–30%", note: "fixture" },
  assumptions: ["fixture assumption"],
  method_ref: "docs/method/aggregation.md",
};

export const official_estimate = {
  estimate_id: "est-official-P_DOOM-10y",
  producer: "pdoom-model/indexes@0.1.0",
  status: "insufficiently_calibrated",
  outcome_set: ["O3", "O4", "O5", "O6", "O7", "O8"],
  outcome_label: "p(DOOM) (O3–O8)",
  horizon: "10y",
  conditioning: "Unconditional.",
  forecast_origin_date: "2026-09-26",
  last_evidence_date: "2026-09-01",
  quantiles: null,
  mean: null,
  disagreement: "extreme",
  uncertainty: "extreme",
  model_confidence: "low",
  source_coverage: { forecast_count: 0, population_count: 0, source_ids: [] },
  previous: null,
  reason_for_change: "first release",
  rounding_rule: "not_applicable",
  display: { central: "Insufficiently calibrated", interval: null, note: "No official probability is published." },
  assumptions: [],
  method_ref: "docs/method/official-status.md",
};

export const index_value = {
  index_id: "capability_pressure",
  value: 62,
  label: "Capability Pressure Index",
  scale: "0-100, higher = more pressure",
  is_probability: false,
  baseline: { snapshot_id: "snap-2026-09-26-001", value: 62 },
  components: [
    { signal_id: "D1.task_horizon_50pct", weight: 0.2, value_normalized: 0.7, tier: 1, contribution: 14 },
  ],
  coverage: 0.8,
  as_of: "2026-09-26",
  method_ref: "docs/method/indexes.md#capability-pressure",
  note: "",
};

export const aggregation = {
  group_id: "grp-fixture-O6-2100-unconditional",
  outcome_set: ["O6"],
  horizon: "2100",
  method: "unweighted_median",
  value: 0.05,
  quantiles: { p05: 0.003, p25: 0.01, p50: 0.05, p75: 0.1, p95: 0.3 },
  n: 6,
  weights_note: "Each forecast counted once.",
  wording_note: "Original wordings shown beside the aggregate.",
  source_ids: ["src-fixture-survey-2025"],
};

export const sensitivity_run = {
  id: "sens-loo-src-fixture-survey-2025",
  kind: "leave_one_source_out",
  target_estimate_id: "est-external-O6-2100",
  baseline_p50: 0.05,
  perturbed_p50: 0.04,
  delta: -0.01,
  description: "Fixture leave-one-out run.",
};

export const drivers_explained = {
  driver: "D1",
  signal_id: "D1.task_horizon_50pct",
  index_id: "capability_pressure",
  direction: "increases",
  magnitude: 14,
  source_ids: ["src-fixture-survey-2025"],
  confidence: "moderate",
  model_role: "index_component",
  last_updated: "2026-09-26",
  sensitivity: { leave_one_out_delta: -14, note: "Fixture sensitivity note." },
  counterevidence: null,
};

export const sensitivity_summary = {
  run_count: 1,
  max_abs_delta: 0.01,
  most_sensitive_kind: "leave_one_source_out",
  note: "fixture",
};

export const delta_record = {
  previous_estimate: null,
  new_candidate: {
    artifact_id: "cand-2026-09-26-001",
    estimate_id: "est-external-O6-2100",
    status: "external_aggregate",
    p50: 0.05,
  },
  absolute_change: null,
  relative_change: null,
  affected_horizons: ["2100"],
  affected_outcomes: ["O6"],
  sources_added: ["src-fixture-survey-2025"],
  sources_removed: [],
  model_changes: ["first release"],
  weight_changes: [],
  data_corrections: [],
  sensitivity_summary,
  heightened_review_triggers: [],
  reviewer: null,
  approval: null,
  release: null,
};

export const approval = {
  reviewer_id: "reviewer-fixture",
  key_id: "fixture-key-01",
  signed_at: "2026-09-26T00:00:00Z",
  manifest_sha256: SHA,
  signature_base64: SIG,
  conflicts_declared: [],
};

export const release_manifest = {
  release_id: "rel-2026-09-26-001",
  model_versions: [
    "pdoom-model/external-aggregate@0.1.0",
    "pdoom-model/experimental-causal@0.1.0",
    "pdoom-model/indexes@0.1.0",
  ],
  code_commit: "0123456789abcdef0123456789abcdef01234567",
  data_snapshot: "snap-2026-09-26-001",
  source_cutoff: "2026-09-01",
  outcome_definition: {
    label: "p(DOOM)",
    outcome_set: ["O3", "O4", "O5", "O6", "O7", "O8"],
    text: "Fixture outcome definition text.",
    definition_ids: ["def-fixture-term"],
  },
  horizons: ["1y", "3y", "5y", "10y", "25y", "2100", "eventual"],
  priors: {
    description: "fixture",
    model_spec_id: "pdoom-model-spec@0.1.0",
    experimental_causal_version: "0.1.0",
    method_ref: "docs/method/experimental-model.md",
  },
  weights: {
    description: "fixture",
    tier_multipliers: { "1": 1, "2": 0.9, "3": 0.5, "4": 0, "5": 0 },
    weight_bounds: [0, 0.35],
    method_ref: "docs/method/indexes.md",
  },
  dependencies: {
    description: "fixture",
    common_factor_loading: 0.5,
    method_ref: "docs/method/experimental-model.md",
  },
  estimates_summary: [
    {
      estimate_id: "est-external-O6-2100",
      status: "external_aggregate",
      outcome_set: ["O6"],
      horizon: "2100",
      p50: 0.05,
      display_central: "5%",
    },
  ],
  intervals_summary: [
    {
      estimate_id: "est-external-O6-2100",
      p05: 0.003,
      p95: 0.3,
      display_interval: "0%–30%",
      uncertainty: "extreme",
    },
  ],
  sensitivity_summary,
  external_forecasts_summary: {
    forecast_count: 6,
    group_count: 1,
    population_count: 3,
    source_count: 1,
    note: "fixture",
  },
  changes: ["first release"],
  reviewers: [{ reviewer_id: "reviewer-fixture", role: "method", conflicts_declared: [] }],
  approval: {
    status: "approved",
    required_approvals: 1,
    received_approvals: 1,
    approved_at: "2026-09-26T00:00:00Z",
  },
  known_limitations: ["fixture limitation"],
  reproduction_command:
    "go run ./cmd/pdoomctl model run --snapshot data/snapshots/snap-2026-09-26-001 --out data/candidates/cand-2026-09-26-001",
  signature: {
    algorithm: "ed25519",
    key_id: "fixture-key-01",
    signed_at: "2026-09-26T00:00:00Z",
    signature_base64: SIG,
  },
  published: "2026-09-26T00:00:00Z",
  superseded: null,
};

export const user_scenario_params = {
  horizon: "10y",
  capability_timeline: 1,
  autonomy_growth: 0,
  access_level: -1,
  safety_progress: 2,
  governance_strength: 0,
  model_security: -2,
  open_weight_diffusion: 1,
  international_coordination: 0,
  incident_frequency: 0,
  resilience: 1,
  samples: 20000,
  seed: 20260926,
};

const stats = { p05: 0.001, p25: 0.005, p50: 0.01, p75: 0.02, p95: 0.05, mean: 0.015 };

export const user_scenario_result = {
  label: "user_scenario",
  disclaimer: "Under your selected assumptions—not the p(DOOM) official model—the median estimate is shown below.",
  horizon: "10y",
  outcome_estimates: {
    O3: stats,
    O4: stats,
    O5: stats,
    O6: stats,
    O7: stats,
    O8: stats,
    P_DOOM: stats,
    P_COLLAPSE: stats,
  },
  factor_summary: { A: tri, C: tri, E: tri, F: tri },
  flags: ["extreme_setting"],
  params_echo: user_scenario_params,
  spec_version: "0.1.0",
};

export const submission = {
  id: "sub-fixture-0001",
  kind: "source",
  submitted_at: "2026-09-26T00:00:00Z",
  payload: { canonical_url: "https://example.org/report", title: "Fixture" },
  contact: "fixture@example.org",
  status: "received",
};

export const cassandra_finding = {
  finding_id: "cf-fixture-0001",
  severity: "medium",
  component: "internal/model",
  claim: "Fixture claim.",
  assumption: "Fixture assumption.",
  counterevidence: "Fixture counterevidence.",
  reproduction: "Fixture reproduction steps.",
  impact: "Fixture impact.",
  recommendation: "Fixture recommendation.",
  status: "open",
};

/** One valid fixture per KINDS entry. */
export const FIXTURES: Record<Kind, unknown> = {
  definition,
  source,
  claim,
  forecast,
  benchmark,
  benchmark_result,
  incident,
  scenario,
  scenario_edge,
  driver,
  driver_observation,
  intervention,
  organization,
  action,
  model_spec,
  snapshot_manifest,
  estimate,
  index_value,
  aggregation,
  sensitivity_run,
  drivers_explained,
  delta_record,
  release_manifest,
  approval,
  submission,
  cassandra_finding,
  user_scenario_params,
  user_scenario_result,
};
