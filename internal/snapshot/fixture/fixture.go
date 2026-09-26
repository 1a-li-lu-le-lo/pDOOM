// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package fixture builds a small synthetic snapshot for automated tests. Every
// item is invented: ids carry the word fixture and URLs point at example.org.
// Nothing here describes a real source, forecast, benchmark, incident or
// organization, and nothing here may be published.
package fixture

import (
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

func s(v string) *string   { return &v }
func f(v float64) *float64 { return &v }
func i(v int) *int         { return &v }

func verified(date string) schema.Review {
	return schema.Review{Verification: schema.Verification{Status: schema.VerifiedFetch, CheckedAt: date, Method: "fixture", Note: "synthetic"}, HumanReviewStatus: "reviewed", ModelUseStatus: schema.ModelUseEligible}
}

func informational(date string) schema.Review {
	return schema.Review{Verification: schema.Verification{Status: schema.VerifiedPriorKnowledge, CheckedAt: date, Method: "fixture", Note: "synthetic"}, HumanReviewStatus: "pending", ModelUseStatus: schema.ModelUseInformational}
}

func source(id, title string, tier int, typ string, date string, conflicts ...string) schema.Source {
	cs := []schema.ConflictLabel{"none_known"}
	if len(conflicts) > 0 {
		cs = nil
		for _, c := range conflicts {
			cs = append(cs, schema.ConflictLabel(c))
		}
	}
	return schema.Source{ID: id, CanonicalURL: "https://example.org/fixture/" + id, Title: title, Publisher: "Fixture Publisher", Authors: []string{"Fixture Author"},
		DatePublished: s(date), DateRetrieved: "2026-08-15", SourceTier: schema.SourceTier(tier), SourceType: schema.SourceType(typ), Topic: []string{"fixture"}, ClaimIDs: []string{},
		EvidenceSummary: "Synthetic test source.", Conflicts: cs, RobotsStatus: "not_applicable", Language: "en", RetractionStatus: "none", Citation: title + " (fixture)", Review: verified("2026-08-15")}
}

// New returns the synthetic fixture snapshot.
func New() *snapshot.Snapshot {
	const d = "2026-08-15"
	snap := &snapshot.Snapshot{Manifest: schema.SnapshotManifest{SnapshotID: "snap-1999-01-01-001", CreatedAt: "2026-09-26T00:00:00Z", SourceCutoff: "2026-09-01",
		Notes: "Synthetic test fixture. Every item in this snapshot is invented for automated tests; ids carry the word fixture and URLs point at example.org. Nothing here describes a real source, forecast, benchmark, incident or organization, and nothing here may be published."}}

	snap.Sources = []schema.Source{
		source("src-fixture-survey-2025", "Fixture survey of researcher expectations", 2, "survey", "2025-06-01"),
		source("src-fixture-tournament-2024", "Fixture forecasting tournament report", 2, "paper", "2024-03-01"),
		source("src-fixture-eval-report", "Fixture capability evaluation report", 1, "evaluation", "2026-05-01"),
		source("src-fixture-lab-card", "Fixture laboratory system card", 1, "model_card", "2026-06-01", "developer_self_report"),
		source("src-fixture-regulator", "Fixture regulator incident notice", 1, "regulatory", "2026-02-01", "government_policy_context"),
		source("src-fixture-blog", "Fixture commentary post", 4, "blog", "2026-07-01"),
	}
	snap.Sources[5].Review = informational(d)
	snap.Sources[5].ModelUseStatus = schema.ModelUseInformational
	snap.Sources[0].ClaimIDs = []string{"clm-fixture-survey-01"}
	snap.Sources[2].ClaimIDs = []string{"clm-fixture-eval-01"}

	snap.Claims = []schema.Claim{
		{ID: "clm-fixture-survey-01", Text: "Median respondent gave 5% to the fixture outcome.", Subject: "fixture respondents", Predicate: "assign median probability", Object: "fixture outcome", Date: s("2025-06-01"), SourceID: "src-fixture-survey-2025", EvidenceType: "survey_result", QuantitativeValue: f(0.05), Unit: s("probability"), DirectQuotePointer: s("Table 1"), Context: "synthetic", CorroborationIDs: []string{}, ContradictionIDs: []string{}, Relevance: "moderate", Status: "corroborated", Review: verified(d)},
		{ID: "clm-fixture-eval-01", Text: "Fixture model completes tasks of 120 minutes at 50% reliability.", Subject: "fixture model", Predicate: "time horizon", Object: "120 minutes", Date: s("2026-05-01"), SourceID: "src-fixture-eval-report", EvidenceType: "measurement", QuantitativeValue: f(120), Unit: s("minutes"), Context: "synthetic", CorroborationIDs: []string{}, ContradictionIDs: []string{}, Relevance: "moderate", Status: "corroborated", Review: verified(d)},
	}

	g := "G-FIX-EXT-2100"
	snap.Forecasts = []schema.Forecast{
		{ID: "fc-fixture-survey-2025-researchers", ForecasterOrSurvey: "Fixture survey 2025", SourceID: "src-fixture-survey-2025", Date: "2025-06-01", Population: "general_ai_researchers", SampleSize: i(120), Expertise: "fixture", QuestionWordingOriginal: "What probability do you give to the fixture outcome by 2100?", OutcomeSet: []schema.Outcome{schema.O6}, Horizon: "2100", Conditions: "unconditional", Median: f(0.05), Mean: f(0.09), GroupID: &g, Status: "current", Review: verified(d)},
		{ID: "fc-fixture-survey-2025-safety", ForecasterOrSurvey: "Fixture survey 2025 (safety subgroup)", SourceID: "src-fixture-survey-2025", Date: "2025-06-01", Population: "ai_safety_researchers", SampleSize: i(40), Expertise: "fixture", QuestionWordingOriginal: "What probability do you give to the fixture outcome by 2100?", OutcomeSet: []schema.Outcome{schema.O6}, Horizon: "2100", Conditions: "unconditional", Median: f(0.15), GroupID: &g, Status: "current", Review: verified(d)},
		{ID: "fc-fixture-tournament-superforecasters", ForecasterOrSurvey: "Fixture tournament", SourceID: "src-fixture-tournament-2024", Date: "2024-03-01", Population: "superforecasters", SampleSize: i(80), Expertise: "fixture", QuestionWordingOriginal: "Probability of the fixture outcome by 2100", Paraphrase: true, OutcomeSet: []schema.Outcome{schema.O6}, Horizon: "2100", Conditions: "unconditional", Median: f(0.004), GroupID: &g, Status: "current", Review: verified(d)},
		{ID: "fc-fixture-tournament-experts", ForecasterOrSurvey: "Fixture tournament (experts)", SourceID: "src-fixture-tournament-2024", Date: "2024-03-01", Population: "domain_experts", SampleSize: i(30), Expertise: "fixture", QuestionWordingOriginal: "Probability of the fixture outcome by 2100", Paraphrase: true, OutcomeSet: []schema.Outcome{schema.O6}, Horizon: "2100", Conditions: "unconditional", Median: f(0.03), GroupID: &g, Status: "current", Review: verified(d)},
		{ID: "fc-fixture-blog-eventual", ForecasterOrSurvey: "Fixture commentator", SourceID: "src-fixture-blog", Date: "2026-07-01", Population: "individual_expert", Expertise: "fixture", QuestionWordingOriginal: "Eventually, the fixture outcome is likely.", OutcomeSet: []schema.Outcome{schema.O3, schema.O6}, Horizon: "eventual", Conditions: "unconditional", Median: f(0.3), Status: "current", Review: informational(d)},
	}

	snap.Benchmarks = []schema.Benchmark{{ID: "bm-fixture-horizon", Name: "Fixture time horizon", Maintainer: "Fixture Evaluators", Version: s("1.0"), URL: "https://example.org/fixture/bench", Tasks: "synthetic", ContaminationRisk: "low", Saturation: "none", Unit: "minutes", Direction: "higher_is_more_capable", Limitations: "synthetic", PDoomRelevance: "moderate", SourceIDs: []string{"src-fixture-eval-report"}, Review: verified(d)}}
	snap.BenchmarkResults = []schema.BenchmarkResult{
		{ID: "bmr-fixture-horizon-model-a", BenchmarkID: "bm-fixture-horizon", ModelName: "Fixture Model A", ModelDeveloper: "Fixture Lab", Date: "2025-11-01", Value: 60, Unit: "minutes", Confidence: "high", SourceIDs: []string{"src-fixture-eval-report"}, Review: verified(d)},
		{ID: "bmr-fixture-horizon-model-b", BenchmarkID: "bm-fixture-horizon", ModelName: "Fixture Model B", ModelDeveloper: "Fixture Lab", Date: "2026-05-01", Value: 120, Unit: "minutes", CILow: f(90), CIHigh: f(170), Confidence: "high", SourceIDs: []string{"src-fixture-eval-report"}, Review: verified(d)},
	}

	snap.Incidents = []schema.Incident{
		{ID: "inc-fixture-agent-deletion", Title: "Fixture coding agent removed production data", Date: s("2026-01-15"), DatePrecision: "day", ExternalIDs: schema.ExternalIDs{AIID: s("fixture-1001")}, Summary: "A synthetic agent acted outside its instructions in a test narrative.", Cause: []schema.IncidentCause{"insufficient_oversight"}, Harm: []schema.Harm{"financial"}, Severity: "material", PDoomRelevance: "moderate", EvidenceLevel: "corroborated_report", SystemsInvolved: []string{"fixture agent"}, NearMiss: false, Novelty: "notable", SourceIDs: []string{"src-fixture-regulator"}, Review: verified(d)},
		{ID: "inc-fixture-chatbot-error", Title: "Fixture chatbot gave a wrong answer", Date: s("2025-09-01"), DatePrecision: "month", ExternalIDs: schema.ExternalIDs{AIID: s("fixture-1002")}, Summary: "A synthetic chatbot error with no wider consequence.", Cause: []schema.IncidentCause{"malfunction"}, Harm: []schema.Harm{"informational"}, Severity: "minor", PDoomRelevance: "weak", EvidenceLevel: "single_source_report", SystemsInvolved: []string{"fixture chatbot"}, Novelty: "routine", SourceIDs: []string{"src-fixture-regulator"}, Review: verified(d)},
	}

	snap.Interventions = []schema.Intervention{
		{ID: "I01", Name: "Fixture dangerous-capability evaluations", TargetScenarioIDs: []string{"S1", "S3"}, Mechanism: "synthetic", EvidenceSummary: "synthetic", EvidenceStrength: "moderate", Cost: "moderate", TimeToDeploy: "1-2y", EffectSize: "unknown", Uncertainty: "high", PossibleFailure: "synthetic", PossibleBackfire: "synthetic", OwnerTypes: []string{"laboratories"}, UserActions: []schema.UserAction{{Audience: "ai_researchers", Action: "fixture"}}, Category: "technical", SourceIDs: []string{"src-fixture-eval-report"}, Review: verified(d)},
		{ID: "I02", Name: "Fixture third-party audits", TargetScenarioIDs: []string{"S3"}, Mechanism: "synthetic", EvidenceSummary: "synthetic", EvidenceStrength: "weak", Cost: "moderate", TimeToDeploy: "1-2y", EffectSize: "unknown", Uncertainty: "high", PossibleFailure: "synthetic", PossibleBackfire: "synthetic", OwnerTypes: []string{"policymakers"}, UserActions: []schema.UserAction{}, Category: "organizational", SourceIDs: []string{"src-fixture-regulator"}, Review: verified(d)},
	}
	snap.Scenarios = []schema.Scenario{
		{ID: "S1", Name: "Fixture deliberate misaligned action", OutcomeSet: []schema.Outcome{schema.O3, schema.O6}, Description: "synthetic", Prerequisites: []string{"fixture"}, EarlyIndicators: []string{"fixture"}, Counterindicators: []string{"fixture"}, CapabilityThresholds: []string{"fixture"}, Exposure: "fixture", ControlFailures: []string{"fixture"}, HumanContributions: []string{"fixture"}, AIContributions: []string{"fixture"}, Dependencies: []string{}, TimeHorizonNote: "fixture", ProbabilitySource: "not_assigned", Uncertainty: "extreme", InterventionIDs: []string{"I01"}, Recoverability: "low", EvidenceSummary: "synthetic", SourceIDs: []string{"src-fixture-eval-report"}, OpenQuestions: []string{"fixture"}, Review: verified(d)},
		{ID: "S3", Name: "Fixture control loss", OutcomeSet: []schema.Outcome{schema.O3, schema.O4}, Description: "synthetic", Prerequisites: []string{"fixture"}, EarlyIndicators: []string{"fixture"}, Counterindicators: []string{"fixture"}, CapabilityThresholds: []string{"fixture"}, Exposure: "fixture", ControlFailures: []string{"fixture"}, HumanContributions: []string{"fixture"}, AIContributions: []string{"fixture"}, Dependencies: []string{"S1"}, TimeHorizonNote: "fixture", ProbabilitySource: "not_assigned", Uncertainty: "extreme", InterventionIDs: []string{"I01", "I02"}, Recoverability: "low", EvidenceSummary: "synthetic", SourceIDs: []string{"src-fixture-eval-report"}, OpenQuestions: []string{}, Review: verified(d)},
	}
	snap.ScenarioEdges = []schema.ScenarioEdge{{ID: "se-S1-S3", FromID: "S1", ToID: "S3", Relation: "enables", Confidence: "moderate", Rationale: "synthetic", SourceIDs: []string{"src-fixture-eval-report"}}}

	sig := func(id, name string, dir schema.SignalDirection) schema.Signal {
		return schema.Signal{SignalID: id, Name: name, Description: "synthetic", Direction: dir, Normalization: "synthetic linear map to [0,1]", RawUnit: s("unit"), PreferredSourceTypes: []string{"evaluation"}, ObservationVsJudgment: "observation"}
	}
	snap.Drivers = []schema.Driver{
		{ID: "D1", Name: "capability", Description: "synthetic", Signals: []schema.Signal{sig("D1.task_horizon_50pct", "Task horizon", schema.HigherRaisesPressure), sig("D1.swe_bench", "Coding", schema.HigherRaisesPressure)}},
		{ID: "D2", Name: "autonomy", Description: "synthetic", Signals: []schema.Signal{sig("D2.unattended_duration", "Unattended duration", schema.HigherRaisesPressure)}},
		{ID: "D3", Name: "access_exposure", Description: "synthetic", Signals: []schema.Signal{sig("D3.tool_permission_scope", "Permission scope", schema.HigherRaisesPressure)}},
		{ID: "D5", Name: "alignment_control", Description: "synthetic", Signals: []schema.Signal{sig("D5.control_evaluations", "Control evaluations", schema.HigherStrengthensControl)}},
		{ID: "D6", Name: "security", Description: "synthetic", Signals: []schema.Signal{sig("D6.mcp_registry_signing", "Registry signing", schema.HigherStrengthensControl)}},
		{ID: "D7", Name: "governance", Description: "synthetic", Signals: []schema.Signal{sig("D7.binding_frontier_rules", "Binding rules", schema.HigherStrengthensControl)}},
	}
	obs := func(id, signal, family string, v float64, src string, kind string) schema.DriverObservation {
		return schema.DriverObservation{ID: id, SignalID: signal, Family: schema.DriverFamily(family), ValueNormalized: v, RawValue: f(v * 100), RawUnit: s("unit"), Confidence: 0.7, ObservationKind: schema.ObservationKind(kind), AsOf: "2026-08-01", Rationale: "synthetic", SourceIDs: []string{src}, Review: verified(d)}
	}
	snap.DriverObservations = []schema.DriverObservation{
		obs("do-d1-task-horizon-50pct", "D1.task_horizon_50pct", "D1", 0.62, "src-fixture-eval-report", "observation"),
		obs("do-d1-swe-bench", "D1.swe_bench", "D1", 0.7, "src-fixture-lab-card", "observation"),
		obs("do-d2-unattended-duration", "D2.unattended_duration", "D2", 0.5, "src-fixture-eval-report", "judgment"),
		obs("do-d3-tool-permission-scope", "D3.tool_permission_scope", "D3", 0.6, "src-fixture-eval-report", "judgment"),
		obs("do-d5-control-evaluations", "D5.control_evaluations", "D5", 0.3, "src-fixture-eval-report", "judgment"),
		obs("do-d6-mcp-registry-signing", "D6.mcp_registry_signing", "D6", 0.25, "src-fixture-eval-report", "judgment"),
		// D7 intentionally has no observation → coverage < 1.
	}
	snap.Organizations = []schema.Organization{{ID: "org-fixture-institute", Name: "Fixture Institute", URL: "https://example.org/fixture/org", Mission: "synthetic", LegalStatus: "fixture nonprofit", Jurisdiction: "fixture", Focus: []string{"fixture"}, Programs: []string{}, OpenOutputs: []string{}, FundingDisclosure: "not published", Conflicts: []string{}, EvidenceOfImpact: "synthetic", WaysToHelp: []string{"fixture"}, InclusionCriteriaMet: []string{"fixture"}, LastVerified: d, SourceIDs: []string{}, Review: verified(d)}}
	snap.Actions = []schema.Action{{ID: "act-individuals-fixture-learn", Audience: "individuals", Title: "Fixture action", Description: "synthetic", RelatedInterventionIDs: []string{"I01"}, Resources: []schema.ActionResource{{Title: "Fixture resource", URL: "https://example.org/fixture/resource"}}, Effort: "low", Review: verified(d)}}
	snap.Definitions = []schema.Definition{{ID: "def-fixture-term", Term: "Fixture term", ShortDefinition: "synthetic", Definitions: []schema.DefinitionText{{Text: "synthetic", SourceID: s("src-fixture-eval-report"), Attribution: "Fixture", Note: ""}}, Consensus: "contested", RelatedIDs: []string{"S1"}, SeeAlsoURLs: []string{}, Review: verified(d)}}

	tri := func(a, b, c float64) schema.TriQuantile { return schema.TriQuantile{P05: a, P50: b, P95: c} }
	horizon := func(a schema.TriQuantile) schema.HorizonCausalSpec {
		return schema.HorizonCausalSpec{A: a, C: tri(0.3, 0.6, 0.9), E: tri(0.1, 0.3, 0.6), F: tri(0.05, 0.2, 0.5),
			O: map[string]schema.TriQuantile{"O3": tri(0.05, 0.2, 0.5), "O4": tri(0.02, 0.1, 0.3), "O5": tri(0.01, 0.05, 0.2), "O6": tri(0.01, 0.05, 0.2), "O7": tri(0.005, 0.02, 0.1), "O8": tri(0.01, 0.05, 0.2)}}
	}
	snap.ModelSpec = schema.ModelSpec{
		ID: "pdoom-model-spec@0.1.0",
		IndexWeights: schema.IndexWeights{
			CapabilityPressure:        map[string]float64{"D1.task_horizon_50pct": 0.35, "D1.swe_bench": 0.35, "D2.unattended_duration": 0.30},
			ControlStrength:           map[string]float64{"D5.control_evaluations": 0.35, "D6.mcp_registry_signing": 0.35, "D7.binding_frontier_rules": 0.30},
			AgenticInfrastructureRisk: map[string]float64{"D2.unattended_duration": 0.34, "D3.tool_permission_scope": 0.33, "D6.mcp_registry_signing": 0.33},
			IncidentPressure:          map[string]float64{},
			EvidencePressure:          map[string]float64{},
			Uncertainty:               map[string]float64{"coverage_gap": 0.10, "forecast_disagreement": 0.15, "low_tier_share": 0.10, "judgment_share": 0.15, "sensitivity_spread": 0.10, "calibration_gap": 0.40},
		},
		WeightBounds:    []float64{0, 0.35},
		TierMultipliers: map[string]float64{"1": 1, "2": 0.9, "3": 0.5, "4": 0, "5": 0},
		IncidentScoring: schema.IncidentScoring{
			SeverityWeights:     map[string]float64{"negligible": 0, "minor": 0.1, "material": 0.3, "major": 0.6, "severe": 0.85, "catastrophic": 1},
			RelevanceWeights:    map[string]float64{"none": 0, "weak": 0.1, "indirect": 0.25, "moderate": 0.5, "strong": 0.8, "direct_precursor": 1},
			EvidenceWeights:     map[string]float64{"allegation": 0.1, "single_source_report": 0.3, "corroborated_report": 0.6, "official_finding": 0.85, "peer_reviewed_analysis": 0.9, "independently_reproduced": 1},
			RecencyHalfLifeDays: 730, SquashK: 3,
		},
		EditorialRules:     []schema.EditorialRule{{Level: "insufficient_evidence", When: "coverage < 0.4"}, {Level: "severe_uncertainty", When: "uncertainty >= 70"}, {Level: "high", When: "cpi >= 70 and csi < 40"}, {Level: "elevated", When: "cpi >= 55 and csi < 55"}, {Level: "guarded", When: "cpi >= 40"}, {Level: "low", When: "cpi >= 25"}, {Level: "very_low", When: "true"}},
		RoundingRules:      schema.RoundingRules{Extreme: 5, High: 2, Moderate: 1, Low: 1},
		AggregationMethods: []schema.AggregationMethod{schema.MethodUnweightedMedian, schema.MethodLinearPool, schema.MethodLogOddsPool, schema.MethodTrimmedMean, schema.MethodTierWeighted, schema.MethodRecencyWeighted, schema.MethodEqualWeightByPopulation},
		ExperimentalCausal: schema.ExperimentalCausalSpec{Version: "pdoom-model/experimental-causal@0.1.0", Seed: 20260926, Samples: 4000, CommonFactorLoading: 0.5,
			Horizons:  map[string]schema.HorizonCausalSpec{"5y": horizon(tri(0.1, 0.3, 0.6)), "10y": horizon(tri(0.2, 0.5, 0.8)), "25y": horizon(tri(0.4, 0.7, 0.95))},
			Rationale: schema.CausalRationale{A: "fixture", C: "fixture", E: "fixture", F: "fixture", O: "fixture", Dependence: "fixture"},
			SourceIDs: []string{"src-fixture-survey-2025", "src-fixture-tournament-2024"}},
	}
	return snap
}
