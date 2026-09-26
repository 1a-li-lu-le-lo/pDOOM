-- Copyright NU Cybernetics. p(DOOM) — research prototype.
-- Reference PostgreSQL schema (ADR-010). The prototype's system of record is the
-- sealed snapshot and release directories under data/; this schema mirrors
-- them for teams that need relational queries. Column names follow the JSON
-- field names in packages/schemas (Appendix A of the build specification).
-- Nothing in the running system depends on this file.

BEGIN;

CREATE SCHEMA IF NOT EXISTS pdoom;
SET search_path TO pdoom, public;

-- Enumerations mirror packages/schemas/src/enums.ts. They are CHECK
-- constraints rather than ENUM types so that adding a value is a migration of
-- one constraint, not a type rebuild.

CREATE TABLE snapshot (
  snapshot_id          text PRIMARY KEY CHECK (snapshot_id ~ '^snap-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{3}$'),
  created_at           timestamptz NOT NULL,
  source_cutoff        date NOT NULL,
  baseline_snapshot_id text REFERENCES snapshot (snapshot_id),
  notes                text NOT NULL DEFAULT '',
  files                jsonb NOT NULL  -- [{path, sha256, count}]
);

-- Every entity carries the same review triple.
-- verification: {status, checked_at, method, note}
-- human_review_status: pending | approved | rejected | needs_changes
-- model_use_status: eligible | used | informational | excluded

CREATE TABLE source (
  snapshot_id        text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id                 text NOT NULL CHECK (id ~ '^src-[a-z0-9-]+$'),
  canonical_url      text NOT NULL,
  title              text NOT NULL,
  publisher          text NOT NULL,
  authors            text[] NOT NULL DEFAULT '{}',
  date_published     date,
  date_updated       date,
  date_retrieved     date NOT NULL,
  source_tier        smallint NOT NULL CHECK (source_tier BETWEEN 1 AND 5),
  source_type        text NOT NULL,
  jurisdiction       text,
  topic              text[] NOT NULL DEFAULT '{}',
  evidence_summary   text NOT NULL,
  counterevidence    text,
  methodology        text,
  sample             text,
  limitations        text,
  conflicts          text[] NOT NULL DEFAULT '{}',
  license            text,
  robots_status      text NOT NULL CHECK (robots_status IN ('allowed', 'disallowed', 'not_applicable', 'unknown')),
  content_hash       text,
  archive_reference  text,
  language           text NOT NULL,
  translation        text,
  duplicate_group    text,
  retraction_status  text NOT NULL CHECK (retraction_status IN ('none', 'retracted', 'corrected', 'disputed')),
  correction_status  text,
  citation           text NOT NULL,
  verification       jsonb NOT NULL,
  human_review_status text NOT NULL,
  model_use_status   text NOT NULL CHECK (model_use_status IN ('eligible', 'used', 'informational', 'excluded')),
  PRIMARY KEY (snapshot_id, id),
  -- Tier 4 is informational at most; tier 5 is always excluded (build-spec §0.5).
  CHECK (NOT (source_tier = 5 AND model_use_status <> 'excluded')),
  CHECK (NOT (source_tier = 4 AND model_use_status IN ('eligible', 'used')))
);
CREATE INDEX source_tier_idx ON source (snapshot_id, source_tier);
CREATE INDEX source_url_idx ON source (snapshot_id, canonical_url);

CREATE TABLE claim (
  snapshot_id          text NOT NULL,
  id                   text NOT NULL CHECK (id ~ '^clm-[a-z0-9-]+$'),
  text                 text NOT NULL,
  subject              text NOT NULL,
  predicate            text NOT NULL,
  object               text NOT NULL,
  date                 date,
  horizon              text,
  geography            text,
  model_name           text,
  model_version        text,
  source_id            text NOT NULL,
  evidence_type        text NOT NULL,
  quantitative_value   double precision,
  unit                 text,
  uncertainty          text,
  direct_quote_pointer text,
  context              text NOT NULL DEFAULT '',
  corroboration_ids    text[] NOT NULL DEFAULT '{}',
  contradiction_ids    text[] NOT NULL DEFAULT '{}',
  relevance            text NOT NULL CHECK (relevance IN ('none', 'weak', 'indirect', 'moderate', 'strong', 'direct_precursor')),
  status               text NOT NULL,
  verification         jsonb NOT NULL,
  human_review_status  text NOT NULL,
  model_use_status     text NOT NULL,
  PRIMARY KEY (snapshot_id, id),
  FOREIGN KEY (snapshot_id, source_id) REFERENCES source (snapshot_id, id) ON DELETE CASCADE
);

CREATE TABLE forecast (
  snapshot_id               text NOT NULL,
  id                        text NOT NULL CHECK (id ~ '^fc-[a-z0-9-]+$'),
  forecaster_or_survey      text NOT NULL,
  source_id                 text NOT NULL,
  date                      date NOT NULL,
  population                text NOT NULL,
  sample_size               integer CHECK (sample_size >= 0),
  expertise                 text NOT NULL,
  question_wording_original text NOT NULL,
  paraphrase                boolean NOT NULL,
  outcome_set               text[] NOT NULL CHECK (cardinality(outcome_set) >= 1),
  horizon                   text NOT NULL,
  horizon_note              text,
  horizon_end_year          integer,
  conditions                text NOT NULL,
  mean                      double precision CHECK (mean BETWEEN 0 AND 1),
  median                    double precision CHECK (median BETWEEN 0 AND 1),
  quantiles                 jsonb,        -- {p05?, p25?, p50?, p75?, p95?} probabilities
  response_rate             double precision CHECK (response_rate BETWEEN 0 AND 1),
  selection_effects         text,
  framing_effects           text,
  calibration               text,
  group_id                  text,
  transformation_note       text,
  status                    text NOT NULL CHECK (status IN ('current', 'superseded', 'withdrawn')),
  verification              jsonb NOT NULL,
  human_review_status       text NOT NULL,
  model_use_status          text NOT NULL,
  PRIMARY KEY (snapshot_id, id),
  FOREIGN KEY (snapshot_id, source_id) REFERENCES source (snapshot_id, id) ON DELETE CASCADE,
  CHECK (horizon <> 'custom' OR horizon_note IS NOT NULL)
);
CREATE INDEX forecast_group_idx ON forecast (snapshot_id, group_id);

CREATE TABLE benchmark (
  snapshot_id        text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id                 text NOT NULL CHECK (id ~ '^bm-[a-z0-9-]+$'),
  name               text NOT NULL,
  maintainer         text NOT NULL,
  version            text,
  url                text NOT NULL,
  tasks              text NOT NULL,
  contamination_risk text NOT NULL CHECK (contamination_risk IN ('low', 'moderate', 'high', 'unknown')),
  saturation         text NOT NULL CHECK (saturation IN ('none', 'partial', 'saturated', 'unknown')),
  scaffold           text,
  model_access       text,
  unit               text NOT NULL,
  direction          text NOT NULL CHECK (direction IN ('higher_is_more_capable', 'lower_is_more_capable')),
  limitations        text NOT NULL,
  pdoom_relevance    text NOT NULL,
  weight_note        text,
  source_ids         text[] NOT NULL DEFAULT '{}',
  verification       jsonb NOT NULL,
  human_review_status text NOT NULL,
  model_use_status   text NOT NULL,
  PRIMARY KEY (snapshot_id, id)
);

CREATE TABLE benchmark_result (
  snapshot_id     text NOT NULL,
  id              text NOT NULL CHECK (id ~ '^bmr-[a-z0-9-]+$'),
  benchmark_id    text NOT NULL,
  model_name      text NOT NULL,
  model_developer text NOT NULL,
  date            date NOT NULL,
  value           double precision NOT NULL,
  unit            text NOT NULL,
  ci_low          double precision,
  ci_high         double precision,
  scaffold        text,
  confidence      text NOT NULL CHECK (confidence IN ('low', 'moderate', 'high')),
  note            text,
  source_ids      text[] NOT NULL DEFAULT '{}',
  verification    jsonb NOT NULL,
  human_review_status text NOT NULL,
  model_use_status text NOT NULL,
  PRIMARY KEY (snapshot_id, id),
  FOREIGN KEY (snapshot_id, benchmark_id) REFERENCES benchmark (snapshot_id, id) ON DELETE CASCADE,
  CHECK (ci_low IS NULL OR ci_high IS NULL OR ci_low <= ci_high)
);

CREATE TABLE incident (
  snapshot_id      text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id               text NOT NULL CHECK (id ~ '^inc-[a-z0-9-]+$'),
  title            text NOT NULL,
  date             date,
  date_precision   text NOT NULL CHECK (date_precision IN ('day', 'month', 'year', 'unknown')),
  external_ids     jsonb NOT NULL DEFAULT '{}',   -- {aiid?, oecd_aim?, mit_tracker?, cve?, docket?, other?}
  summary          text NOT NULL,                 -- non-graphic, non-operational
  cause            text[] NOT NULL DEFAULT '{}',
  harm             text[] NOT NULL DEFAULT '{}',
  severity         text NOT NULL,
  pdoom_relevance  text NOT NULL,
  evidence_level   text NOT NULL,
  systems_involved text[] NOT NULL DEFAULT '{}',
  jurisdiction     text,
  near_miss        boolean NOT NULL,
  novelty          text NOT NULL CHECK (novelty IN ('routine', 'notable', 'novel')),
  exposure_note    text,
  source_ids       text[] NOT NULL DEFAULT '{}',
  verification     jsonb NOT NULL,
  human_review_status text NOT NULL,
  model_use_status text NOT NULL,
  PRIMARY KEY (snapshot_id, id)
);

CREATE TABLE scenario (
  snapshot_id           text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id                    text NOT NULL CHECK (id ~ '^S[0-9]{1,2}$'),
  name                  text NOT NULL,
  outcome_set           text[] NOT NULL CHECK (cardinality(outcome_set) >= 1),
  description           text NOT NULL,
  prerequisites         text[] NOT NULL DEFAULT '{}',
  early_indicators      text[] NOT NULL DEFAULT '{}',
  counterindicators     text[] NOT NULL DEFAULT '{}',
  capability_thresholds text[] NOT NULL DEFAULT '{}',
  exposure              text NOT NULL,
  control_failures      text[] NOT NULL DEFAULT '{}',
  human_contributions   text[] NOT NULL DEFAULT '{}',
  ai_contributions      text[] NOT NULL DEFAULT '{}',
  dependencies          text[] NOT NULL DEFAULT '{}',
  time_horizon_note     text NOT NULL,
  probability_source    text NOT NULL CHECK (probability_source IN ('not_assigned', 'external_forecast', 'experimental_model', 'expert_elicitation')),
  uncertainty           text NOT NULL,
  intervention_ids      text[] NOT NULL DEFAULT '{}',
  recoverability        text NOT NULL CHECK (recoverability IN ('high', 'moderate', 'low', 'none', 'unknown')),
  evidence_summary      text NOT NULL,
  source_ids            text[] NOT NULL DEFAULT '{}',
  open_questions        text[] NOT NULL DEFAULT '{}',
  content_safety_note   text,
  verification          jsonb NOT NULL,
  human_review_status   text NOT NULL,
  model_use_status      text NOT NULL,
  PRIMARY KEY (snapshot_id, id)
);

CREATE TABLE scenario_edge (
  snapshot_id text NOT NULL,
  id          text NOT NULL,
  from_id     text NOT NULL,
  to_id       text NOT NULL,
  relation    text NOT NULL CHECK (relation IN ('enables', 'amplifies', 'prevents_response', 'shares_prerequisite', 'competes_with')),
  confidence  text NOT NULL CHECK (confidence IN ('low', 'moderate', 'high')),
  rationale   text NOT NULL,
  source_ids  text[] NOT NULL DEFAULT '{}',
  PRIMARY KEY (snapshot_id, id),
  FOREIGN KEY (snapshot_id, from_id) REFERENCES scenario (snapshot_id, id) ON DELETE CASCADE,
  FOREIGN KEY (snapshot_id, to_id) REFERENCES scenario (snapshot_id, id) ON DELETE CASCADE,
  CHECK (from_id <> to_id),
  CHECK (id = 'se-' || from_id || '-' || to_id)
);

CREATE TABLE driver (
  snapshot_id text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id          text NOT NULL CHECK (id ~ '^D([1-9]|10)$'),
  name        text NOT NULL,
  description text NOT NULL,
  PRIMARY KEY (snapshot_id, id)
);

CREATE TABLE driver_signal (
  snapshot_id             text NOT NULL,
  driver_id               text NOT NULL,
  signal_id               text NOT NULL,
  name                    text NOT NULL,
  description             text NOT NULL,
  direction               text NOT NULL CHECK (direction IN ('higher_raises_pressure', 'higher_strengthens_control')),
  normalization           text NOT NULL,   -- how the raw value maps to 0..1
  raw_unit                text,
  preferred_source_types  text[] NOT NULL DEFAULT '{}',
  observation_vs_judgment text NOT NULL,
  PRIMARY KEY (snapshot_id, signal_id),
  FOREIGN KEY (snapshot_id, driver_id) REFERENCES driver (snapshot_id, id) ON DELETE CASCADE,
  CHECK (signal_id LIKE driver_id || '.%')
);

CREATE TABLE driver_observation (
  snapshot_id       text NOT NULL,
  id                text NOT NULL,
  signal_id         text NOT NULL,
  family            text NOT NULL,
  value_normalized  double precision NOT NULL CHECK (value_normalized BETWEEN 0 AND 1),
  raw_value         double precision,
  raw_unit          text,
  confidence        double precision NOT NULL CHECK (confidence BETWEEN 0 AND 1),
  observation_kind  text NOT NULL CHECK (observation_kind IN ('observation', 'judgment')),
  as_of             date NOT NULL,
  rationale         text NOT NULL,
  counterevidence   text,
  source_ids        text[] NOT NULL DEFAULT '{}',
  verification      jsonb NOT NULL,
  human_review_status text NOT NULL,
  model_use_status  text NOT NULL,
  PRIMARY KEY (snapshot_id, id),
  FOREIGN KEY (snapshot_id, signal_id) REFERENCES driver_signal (snapshot_id, signal_id) ON DELETE CASCADE,
  CHECK (signal_id LIKE family || '.%')
);

CREATE TABLE intervention (
  snapshot_id         text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id                  text NOT NULL CHECK (id ~ '^I[0-9]{2}$'),
  name                text NOT NULL,
  target_scenario_ids text[] NOT NULL DEFAULT '{}',
  mechanism           text NOT NULL,
  evidence_summary    text NOT NULL,
  evidence_strength   text NOT NULL CHECK (evidence_strength IN ('none', 'weak', 'moderate', 'strong')),
  cost                text NOT NULL CHECK (cost IN ('low', 'moderate', 'high', 'very_high', 'unknown')),
  time_to_deploy      text NOT NULL CHECK (time_to_deploy IN ('months', '1-2y', '3-5y', '5y+', 'unknown')),
  effect_size         text NOT NULL CHECK (effect_size IN ('unknown', 'small', 'moderate', 'large')),  -- qualitative only
  uncertainty         text NOT NULL,
  possible_failure    text NOT NULL,
  possible_backfire   text NOT NULL,
  owner_types         text[] NOT NULL DEFAULT '{}',
  user_actions        jsonb NOT NULL DEFAULT '[]',   -- [{audience, action}]
  category            text NOT NULL CHECK (category IN ('technical', 'organizational', 'national', 'international', 'resilience')),
  source_ids          text[] NOT NULL DEFAULT '{}',
  verification        jsonb NOT NULL,
  human_review_status text NOT NULL,
  model_use_status    text NOT NULL,
  PRIMARY KEY (snapshot_id, id)
);

CREATE TABLE organization (
  snapshot_id            text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id                     text NOT NULL CHECK (id ~ '^org-[a-z0-9-]+$'),
  name                   text NOT NULL,
  url                    text NOT NULL,
  mission                text NOT NULL,
  legal_status           text NOT NULL,
  jurisdiction           text NOT NULL,
  focus                  text[] NOT NULL DEFAULT '{}',
  programs               text[] NOT NULL DEFAULT '{}',
  open_outputs           text[] NOT NULL DEFAULT '{}',
  funding_disclosure     text NOT NULL,
  conflicts              text[] NOT NULL DEFAULT '{}',
  evidence_of_impact     text NOT NULL,
  ways_to_help           text[] NOT NULL DEFAULT '{}',
  inclusion_criteria_met text[] NOT NULL DEFAULT '{}',
  last_verified          date NOT NULL,
  source_ids             text[] NOT NULL DEFAULT '{}',
  verification           jsonb NOT NULL,
  human_review_status    text NOT NULL,
  model_use_status       text NOT NULL,
  PRIMARY KEY (snapshot_id, id)
);

CREATE TABLE action (
  snapshot_id              text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id                       text NOT NULL,
  audience                 text NOT NULL CHECK (audience IN ('individuals', 'software_engineers', 'ai_researchers', 'laboratories', 'policymakers', 'funders', 'educators', 'nonprofits', 'auditors_red_teams', 'standards_bodies')),
  title                    text NOT NULL,
  description              text NOT NULL,
  related_intervention_ids text[] NOT NULL DEFAULT '{}',
  resources                jsonb NOT NULL DEFAULT '[]',  -- [{title, url, source_id}]
  effort                   text NOT NULL CHECK (effort IN ('low', 'moderate', 'high')),
  verification             jsonb NOT NULL,
  human_review_status      text NOT NULL,
  model_use_status         text NOT NULL,
  PRIMARY KEY (snapshot_id, id),
  CHECK (id LIKE 'act-' || audience || '-%')
);

CREATE TABLE definition (
  snapshot_id      text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id               text NOT NULL CHECK (id ~ '^def-[a-z0-9-]+$'),
  term             text NOT NULL,
  short_definition text NOT NULL,
  definitions      jsonb NOT NULL DEFAULT '[]',  -- [{text, source_id, attribution, note}]
  consensus        text NOT NULL CHECK (consensus IN ('consensus', 'contested', 'emerging')),
  related_ids      text[] NOT NULL DEFAULT '{}',
  see_also_urls    text[] NOT NULL DEFAULT '{}',
  verification     jsonb NOT NULL,
  human_review_status text NOT NULL,
  model_use_status text NOT NULL,
  PRIMARY KEY (snapshot_id, id)
);

-- The model specification is versioned as one document per snapshot.
CREATE TABLE model_spec (
  snapshot_id text NOT NULL REFERENCES snapshot (snapshot_id) ON DELETE CASCADE,
  id          text NOT NULL CHECK (id ~ '^pdoom-model-spec@'),
  spec        jsonb NOT NULL,   -- index_weights, weight_bounds, tier_multipliers, incident_scoring, editorial_rules, rounding_rules, aggregation_methods, experimental_causal
  PRIMARY KEY (snapshot_id, id)
);

-- ---------------------------------------------------------------------------
-- Releases: the promoted outputs. A release is immutable once published.
-- ---------------------------------------------------------------------------

CREATE TABLE release (
  release_id            text PRIMARY KEY CHECK (release_id ~ '^rel-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{3}$'),
  candidate_id          text NOT NULL,
  data_snapshot         text NOT NULL REFERENCES snapshot (snapshot_id),
  previous_release_id   text REFERENCES release (release_id),
  model_versions        text[] NOT NULL,
  generated_at          timestamptz NOT NULL,
  published             timestamptz,
  superseded_by         text REFERENCES release (release_id),
  superseded_at         timestamptz,
  source_cutoff         date NOT NULL,
  editorial_risk_level  text NOT NULL,
  uncertainty_score     double precision CHECK (uncertainty_score BETWEEN 0 AND 100),
  code_commit           text NOT NULL,
  reproduction_command  text NOT NULL,
  reviewers             text[] NOT NULL DEFAULT '{}',
  changes               text[] NOT NULL DEFAULT '{}',
  known_limitations     text[] NOT NULL DEFAULT '{}',
  manifest              jsonb NOT NULL,   -- the full signed manifest, byte-identical to manifest.json
  signature             text NOT NULL,
  CHECK ((superseded_by IS NULL) = (superseded_at IS NULL))
);

CREATE TABLE estimate (
  release_id           text NOT NULL REFERENCES release (release_id) ON DELETE CASCADE,
  estimate_id          text NOT NULL,
  status               text NOT NULL CHECK (status IN ('official', 'insufficiently_calibrated', 'external_aggregate', 'research_mode', 'user_scenario')),
  outcome_set          text[] NOT NULL CHECK (cardinality(outcome_set) >= 1),
  outcome_label        text NOT NULL,
  horizon              text NOT NULL,
  horizon_note         text,
  producer             text NOT NULL,   -- pdoom-model/<family>@semver
  group_id             text,
  p05                  double precision, p25 double precision, p50 double precision, p75 double precision, p95 double precision,
  mean                 double precision,
  display_central      text NOT NULL,
  display_interval     text NOT NULL,
  display_note         text NOT NULL,
  uncertainty          text NOT NULL,
  disagreement         text,
  rounding_rule        text NOT NULL CHECK (rounding_rule ~ '^(none|nearest_[0-9]+)$'),
  model_confidence     text NOT NULL,
  forecast_origin_date date NOT NULL,
  last_evidence_date   date NOT NULL,
  conditioning         text NOT NULL,
  assumptions          text[] NOT NULL DEFAULT '{}',
  source_coverage      jsonb NOT NULL,   -- {forecast_count, population_count, source_ids}
  method_ref           text NOT NULL,
  reason_for_change    text,
  previous             jsonb,            -- {release_id, p50, display}
  PRIMARY KEY (release_id, estimate_id),
  -- The withheld official object publishes no number (ADR-002).
  CHECK (status <> 'insufficiently_calibrated' OR (p50 IS NULL AND display_interval = '')),
  CHECK (p05 IS NULL OR (p05 <= p25 AND p25 <= p50 AND p50 <= p75 AND p75 <= p95)),
  -- user_scenario results are never stored in a release.
  CHECK (status <> 'user_scenario')
);

CREATE TABLE index_value (
  release_id     text NOT NULL REFERENCES release (release_id) ON DELETE CASCADE,
  index_id       text NOT NULL CHECK (index_id IN ('evidence_pressure', 'capability_pressure', 'control_strength', 'incident_pressure', 'uncertainty', 'agentic_infrastructure_risk', 'attention')),
  value          double precision CHECK (value BETWEEN 0 AND 100),
  label          text NOT NULL,
  scale          text NOT NULL,
  is_probability boolean NOT NULL DEFAULT false CHECK (is_probability = false),
  baseline       jsonb,
  components     jsonb NOT NULL DEFAULT '[]',  -- [{signal_id, weight, value_normalized, tier, contribution}]
  coverage       double precision NOT NULL CHECK (coverage BETWEEN 0 AND 1),
  as_of          date NOT NULL,
  method_ref     text NOT NULL,
  note           text NOT NULL DEFAULT '',
  PRIMARY KEY (release_id, index_id)
);

CREATE TABLE aggregation (
  release_id          text NOT NULL REFERENCES release (release_id) ON DELETE CASCADE,
  aggregation_id      text NOT NULL,
  group_id            text NOT NULL,
  outcome_set         text[] NOT NULL,
  horizon             text NOT NULL,
  method              text NOT NULL,
  preferred           boolean NOT NULL,
  value               double precision NOT NULL CHECK (value BETWEEN 0 AND 1),
  n                   integer NOT NULL CHECK (n >= 1),
  population_count    integer NOT NULL CHECK (population_count >= 1),
  forecast_ids        text[] NOT NULL,
  source_ids          text[] NOT NULL,
  weights             jsonb,
  conditioning        text NOT NULL,
  transformation_note text,
  note                text NOT NULL DEFAULT '',
  PRIMARY KEY (release_id, aggregation_id)
);
CREATE UNIQUE INDEX aggregation_preferred_idx ON aggregation (release_id, group_id) WHERE preferred;

CREATE TABLE sensitivity_run (
  release_id       text NOT NULL REFERENCES release (release_id) ON DELETE CASCADE,
  run_id           text NOT NULL,
  kind             text NOT NULL,
  target_kind      text NOT NULL,
  target_id        text NOT NULL,
  outcome_set      text[] NOT NULL,
  horizon          text NOT NULL,
  parameter_change text NOT NULL,
  removed_id       text,
  baseline_value   double precision NOT NULL,
  value            double precision NOT NULL,
  delta            double precision NOT NULL,
  rank             integer NOT NULL,
  note             text NOT NULL DEFAULT '',
  PRIMARY KEY (release_id, run_id)
);

CREATE TABLE approval (
  release_id         text NOT NULL REFERENCES release (release_id) ON DELETE CASCADE,
  reviewer_id        text NOT NULL,
  key_id             text NOT NULL,
  signed_at          timestamptz NOT NULL,
  manifest_sha256    text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
  signature_base64   text NOT NULL,
  conflicts_declared text NOT NULL DEFAULT '',
  PRIMARY KEY (release_id, reviewer_id, key_id)
);

-- Hash-chained audit log (internal/audit). Each row's hash covers the previous
-- hash, so any edit breaks the chain; verify with `pdoomctl audit verify`.
CREATE TABLE audit_event (
  seq        bigint PRIMARY KEY,
  at         timestamptz NOT NULL,
  actor      text NOT NULL,
  action     text NOT NULL,
  subject    text NOT NULL,
  details    jsonb NOT NULL DEFAULT '{}',
  prev_hash  text NOT NULL,
  hash       text NOT NULL UNIQUE CHECK (hash ~ '^[0-9a-f]{64}$')
);

-- Public submissions (review queue). Never joined to releases by the system.
CREATE TABLE submission (
  id           text PRIMARY KEY CHECK (id ~ '^sub-[0-9a-f]{16}$'),
  kind         text NOT NULL CHECK (kind IN ('source', 'correction', 'incident_reference')),
  submitted_at timestamptz NOT NULL,
  payload      jsonb NOT NULL,
  contact      text,
  status       text NOT NULL CHECK (status IN ('received'))
);

-- Convenience view: the current meter for the latest published release.
CREATE VIEW current_meter AS
SELECT e.*
FROM estimate e
JOIN release r ON r.release_id = e.release_id
WHERE r.published IS NOT NULL AND r.superseded_by IS NULL
ORDER BY r.published DESC;

COMMIT;
