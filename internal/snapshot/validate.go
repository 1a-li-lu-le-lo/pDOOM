// Copyright NU Cybernetics. p(DOOM) — research prototype.

package snapshot

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Severity of a validation problem.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// Problem is one validation finding.
type Problem struct {
	Severity string `json:"severity"`
	File     string `json:"file"`
	ID       string `json:"id"`
	Field    string `json:"field"`
	Message  string `json:"message"`
}

func (p Problem) String() string {
	var b strings.Builder
	b.WriteString(p.Severity)
	b.WriteString(": ")
	b.WriteString(p.File)
	if p.ID != "" {
		b.WriteString(" [")
		b.WriteString(p.ID)
		b.WriteString("]")
	}
	if p.Field != "" {
		b.WriteString(" ")
		b.WriteString(p.Field)
	}
	b.WriteString(": ")
	b.WriteString(p.Message)
	return b.String()
}

// HasErrors reports whether any problem has error severity.
func HasErrors(problems []Problem) bool {
	for _, p := range problems {
		if p.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Validate applies JSON Schemas from schemaDir (data/schemas/<kind>.schema.json,
// when present) and the cross-file invariants of build-spec §3 and Appendix A.
// It never mutates the snapshot.
func Validate(s *Snapshot, schemaDir string) []Problem {
	v := &validator{s: s, signals: s.SignalIndex()}
	v.collectIDs()
	v.checkSchemas(schemaDir)
	v.checkManifest()
	v.checkDefinitions()
	v.checkSources()
	v.checkClaims()
	v.checkForecasts()
	v.checkBenchmarks()
	v.checkBenchmarkResults()
	v.checkIncidents()
	v.checkScenarios()
	v.checkScenarioEdges()
	v.checkDrivers()
	v.checkDriverObservations()
	v.checkInterventions()
	v.checkOrganizations()
	v.checkActions()
	v.checkModelSpec()
	sort.SliceStable(v.problems, func(i, j int) bool {
		if v.problems[i].Severity != v.problems[j].Severity {
			return v.problems[i].Severity == SeverityError
		}
		return v.problems[i].File < v.problems[j].File
	})
	return v.problems
}

type validator struct {
	s        *Snapshot
	problems []Problem
	signals  map[string]SignalRef

	sourceIDs, claimIDs, forecastIDs, benchmarkIDs, incidentIDs, scenarioIDs, interventionIDs,
	organizationIDs, definitionIDs, actionIDs, edgeIDs, driverIDs, observationIDs, resultIDs map[string]bool
	allIDs map[string]bool
}

func (v *validator) errf(file, id, field, format string, args ...any) {
	v.problems = append(v.problems, Problem{SeverityError, file, id, field, fmt.Sprintf(format, args...)})
}

func (v *validator) warnf(file, id, field, format string, args ...any) {
	v.problems = append(v.problems, Problem{SeverityWarning, file, id, field, fmt.Sprintf(format, args...)})
}

var (
	reSnapshotID  = regexp.MustCompile(`^snap-\d{4}-\d{2}-\d{2}-\d{3}$`)
	reSnapshotAny = regexp.MustCompile(`^snap-[a-z0-9-]+$`)
	reScenarioID  = regexp.MustCompile(`^S([1-9]|1[0-8])$`)
	reDriverID    = regexp.MustCompile(`^D([1-9]|10)$`)
	reInterID     = regexp.MustCompile(`^I\d{2,}$`)
	reSignalID    = regexp.MustCompile(`^D(?:[1-9]|10)\.[a-z0-9_]+$`)
	reSlug        = regexp.MustCompile(`^[a-z0-9]+(?:[-_.][a-z0-9]+)*$`)
	reEdgeID      = regexp.MustCompile(`^se-S\d+-S\d+$`)
	reClaimID     = regexp.MustCompile(`^clm-[a-z0-9]+(?:-[a-z0-9]+)*-\d{2,}$`)
)

func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func isTimestamp(s string) bool {
	_, err := time.Parse(time.RFC3339, s)
	return err == nil
}

func (v *validator) checkDate(file, id, field string, s *string, required bool) {
	if s == nil {
		if required {
			v.errf(file, id, field, "required date missing")
		}
		return
	}
	if !isDate(*s) {
		v.errf(file, id, field, "not a YYYY-MM-DD date: %q", *s)
	}
}

func (v *validator) checkPrefixedID(file, id, prefix string) {
	if !strings.HasPrefix(id, prefix) || !reSlug.MatchString(strings.TrimPrefix(id, prefix)) {
		v.errf(file, id, "id", "must match %s<slug>", prefix)
	}
}

func (v *validator) checkReview(file, id string, r schema.Review) {
	if !r.Verification.Status.Valid() {
		v.errf(file, id, "verification.status", "invalid value %q", r.Verification.Status)
	}
	if !isDate(r.Verification.CheckedAt) {
		v.errf(file, id, "verification.checked_at", "not a YYYY-MM-DD date: %q", r.Verification.CheckedAt)
	}
	if !r.HumanReviewStatus.Valid() {
		v.errf(file, id, "human_review_status", "invalid value %q", r.HumanReviewStatus)
	}
	if !r.ModelUseStatus.Valid() {
		v.errf(file, id, "model_use_status", "invalid value %q", r.ModelUseStatus)
	}
	if r.Verification.Status == schema.Unverified && r.ModelUseStatus != schema.ModelUseExcluded {
		v.errf(file, id, "model_use_status", "unverified items must be excluded (got %q)", r.ModelUseStatus)
	}
	if r.ModelUseStatus.FeedsModel() && !r.Verification.Status.AllowsModelUse() {
		v.errf(file, id, "model_use_status", "only verified_fetch/verified_search items may be eligible/used (verification %q)", r.Verification.Status)
	}
}

func (v *validator) checkSourceRefs(file, id, field string, ids []string) {
	for _, sid := range ids {
		if !v.sourceIDs[sid] {
			v.errf(file, id, field, "unknown source id %q", sid)
		}
	}
}

func (v *validator) checkUnique(file string, ids []string) map[string]bool {
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" {
			v.errf(file, "", "id", "empty id")
			continue
		}
		if seen[id] {
			v.errf(file, id, "id", "duplicate id")
		}
		seen[id] = true
	}
	return seen
}

func (v *validator) collectIDs() {
	s := v.s
	ids := func(n int, f func(i int) string) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = f(i)
		}
		return out
	}
	v.definitionIDs = v.checkUnique("definitions.json", ids(len(s.Definitions), func(i int) string { return s.Definitions[i].ID }))
	v.sourceIDs = v.checkUnique("sources.json", ids(len(s.Sources), func(i int) string { return s.Sources[i].ID }))
	v.claimIDs = v.checkUnique("claims.json", ids(len(s.Claims), func(i int) string { return s.Claims[i].ID }))
	v.forecastIDs = v.checkUnique("forecasts.json", ids(len(s.Forecasts), func(i int) string { return s.Forecasts[i].ID }))
	v.benchmarkIDs = v.checkUnique("benchmarks.json", ids(len(s.Benchmarks), func(i int) string { return s.Benchmarks[i].ID }))
	v.resultIDs = v.checkUnique("benchmark_results.json", ids(len(s.BenchmarkResults), func(i int) string { return s.BenchmarkResults[i].ID }))
	v.incidentIDs = v.checkUnique("incidents.json", ids(len(s.Incidents), func(i int) string { return s.Incidents[i].ID }))
	v.scenarioIDs = v.checkUnique("scenarios.json", ids(len(s.Scenarios), func(i int) string { return s.Scenarios[i].ID }))
	v.edgeIDs = v.checkUnique("scenario_edges.json", ids(len(s.ScenarioEdges), func(i int) string { return s.ScenarioEdges[i].ID }))
	v.driverIDs = v.checkUnique("drivers.json", ids(len(s.Drivers), func(i int) string { return s.Drivers[i].ID }))
	v.observationIDs = v.checkUnique("driver_observations.json", ids(len(s.DriverObservations), func(i int) string { return s.DriverObservations[i].ID }))
	v.interventionIDs = v.checkUnique("interventions.json", ids(len(s.Interventions), func(i int) string { return s.Interventions[i].ID }))
	v.organizationIDs = v.checkUnique("organizations.json", ids(len(s.Organizations), func(i int) string { return s.Organizations[i].ID }))
	v.actionIDs = v.checkUnique("actions.json", ids(len(s.Actions), func(i int) string { return s.Actions[i].ID }))
	v.allIDs = map[string]bool{}
	for _, m := range []map[string]bool{v.definitionIDs, v.sourceIDs, v.claimIDs, v.forecastIDs, v.benchmarkIDs, v.resultIDs, v.incidentIDs, v.scenarioIDs, v.edgeIDs, v.driverIDs, v.observationIDs, v.interventionIDs, v.organizationIDs, v.actionIDs} {
		for id := range m {
			v.allIDs[id] = true
		}
	}
	for id := range v.signals {
		v.allIDs[id] = true
	}
	for _, o := range schema.Outcomes {
		v.allIDs[string(o)] = true
	}
}

// checkSchemas applies data/schemas/<kind>.schema.json when the file exists. A
// schema whose top level describes the envelope (properties.kind and
// properties.items) is applied to the whole document; otherwise it is applied
// to each item.
func (v *validator) checkSchemas(schemaDir string) {
	if schemaDir == "" {
		return
	}
	apply := func(file, schemaName string) {
		path := filepath.Join(schemaDir, schemaName+".schema.json")
		if _, err := os.Stat(path); err != nil {
			return
		}
		raw, ok := v.s.RawFiles[file]
		if !ok {
			return
		}
		c := jsonschema.NewCompiler()
		sch, err := c.Compile(path)
		if err != nil {
			v.warnf(file, "", "", "json schema %s could not be compiled: %v", path, err)
			return
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
		if err != nil {
			v.errf(file, "", "", "not valid JSON: %v", err)
			return
		}
		schemaDoc, _ := os.ReadFile(path)
		wholeDoc := file == "manifest.json" || looksLikeEnvelopeSchema(schemaDoc)
		if wholeDoc {
			if err := sch.Validate(doc); err != nil {
				v.errf(file, "", "", "json schema: %s", firstLine(err.Error()))
			}
			return
		}
		obj, _ := doc.(map[string]any)
		items, _ := obj["items"].([]any)
		for _, it := range items {
			id := ""
			if m, ok := it.(map[string]any); ok {
				id, _ = m["id"].(string)
			}
			if err := sch.Validate(it); err != nil {
				v.errf(file, id, "", "json schema: %s", firstLine(err.Error()))
			}
		}
	}
	apply("manifest.json", "snapshot_manifest")
	for _, spec := range Files {
		apply(spec.Path, string(spec.Kind))
	}
}

func looksLikeEnvelopeSchema(raw []byte) bool {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return false
	}
	m, ok := doc.(map[string]any)
	if !ok {
		return false
	}
	props, ok := m["properties"].(map[string]any)
	if !ok {
		return false
	}
	_, hasKind := props["kind"]
	_, hasItems := props["items"]
	return hasKind && hasItems
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i] + " …"
	}
	return s
}

func (v *validator) checkManifest() {
	m := v.s.Manifest
	f := "manifest.json"
	if !reSnapshotAny.MatchString(m.SnapshotID) {
		v.errf(f, m.SnapshotID, "snapshot_id", "must start with snap- and use [a-z0-9-]")
	} else if !reSnapshotID.MatchString(m.SnapshotID) {
		v.warnf(f, m.SnapshotID, "snapshot_id", "does not match snap-YYYY-MM-DD-NNN (acceptable only for test fixtures)")
	}
	if base := filepath.Base(v.s.Dir); v.s.Dir != "" && base != m.SnapshotID {
		v.warnf(f, m.SnapshotID, "snapshot_id", "directory name %q differs from snapshot_id", base)
	}
	if !isTimestamp(m.CreatedAt) && !isDate(m.CreatedAt) {
		v.errf(f, m.SnapshotID, "created_at", "not an ISO 8601 timestamp: %q", m.CreatedAt)
	}
	if !isDate(m.SourceCutoff) {
		v.errf(f, m.SnapshotID, "source_cutoff", "not a YYYY-MM-DD date: %q", m.SourceCutoff)
	}
	if m.BaselineSnapshotID != nil && !reSnapshotID.MatchString(*m.BaselineSnapshotID) {
		v.errf(f, m.SnapshotID, "baseline_snapshot_id", "must match snap-YYYY-MM-DD-NNN")
	}
	if strings.TrimSpace(m.Notes) == "" {
		v.warnf(f, m.SnapshotID, "notes", "empty notes")
	}
	listed := map[string]bool{}
	for _, fl := range m.Files {
		listed[fl.Path] = true
		if fl.Count > 0 || fl.SHA256 != "" {
			if got := v.s.ItemCount(fl.Path); got != fl.Count {
				v.errf(f, m.SnapshotID, "files", "%s: manifest count %d, file holds %d items", fl.Path, fl.Count, got)
			}
		}
		if fl.SHA256 == "" {
			v.warnf(f, m.SnapshotID, "files", "%s: no sha256 listed (run `pdoomctl snapshot seal`)", fl.Path)
		}
	}
	for _, p := range v.s.SortedFilePaths() {
		if !listed[p] {
			v.warnf(f, m.SnapshotID, "files", "%s present but not listed in manifest (run `pdoomctl snapshot seal`)", p)
		}
	}
}

func (v *validator) checkDefinitions() {
	f := "definitions.json"
	for _, d := range v.s.Definitions {
		v.checkPrefixedID(f, d.ID, "def-")
		if !d.Consensus.Valid() {
			v.errf(f, d.ID, "consensus", "invalid value %q", d.Consensus)
		}
		if strings.TrimSpace(d.Term) == "" || strings.TrimSpace(d.ShortDefinition) == "" {
			v.errf(f, d.ID, "term", "term and short_definition are required")
		}
		for i, dt := range d.Definitions {
			if dt.SourceID != nil && !v.sourceIDs[*dt.SourceID] {
				v.errf(f, d.ID, fmt.Sprintf("definitions[%d].source_id", i), "unknown source id %q", *dt.SourceID)
			}
		}
		for _, rid := range d.RelatedIDs {
			if !v.allIDs[rid] {
				v.errf(f, d.ID, "related_ids", "unknown id %q", rid)
			}
		}
		v.checkReview(f, d.ID, d.Review)
	}
}

func (v *validator) checkSources() {
	f := "sources.json"
	urls := map[string]string{}
	for _, s := range v.s.Sources {
		v.checkPrefixedID(f, s.ID, "src-")
		if !strings.HasPrefix(s.CanonicalURL, "http://") && !strings.HasPrefix(s.CanonicalURL, "https://") {
			v.errf(f, s.ID, "canonical_url", "must be an absolute http(s) URL")
		}
		if prev, dup := urls[s.CanonicalURL]; dup {
			v.warnf(f, s.ID, "canonical_url", "same canonical_url as %s", prev)
		}
		urls[s.CanonicalURL] = s.ID
		if strings.TrimSpace(s.Title) == "" || strings.TrimSpace(s.Publisher) == "" {
			v.errf(f, s.ID, "title", "title and publisher are required")
		}
		if !s.SourceTier.Valid() {
			v.errf(f, s.ID, "source_tier", "must be 1..5, got %d", s.SourceTier)
		}
		if !s.SourceType.Valid() {
			v.errf(f, s.ID, "source_type", "invalid value %q", s.SourceType)
		}
		if !s.RobotsStatus.Valid() {
			v.errf(f, s.ID, "robots_status", "invalid value %q", s.RobotsStatus)
		}
		if !s.RetractionStatus.Valid() {
			v.errf(f, s.ID, "retraction_status", "invalid value %q", s.RetractionStatus)
		}
		for _, c := range s.Conflicts {
			if !c.Valid() {
				v.errf(f, s.ID, "conflicts", "invalid value %q", c)
			}
		}
		v.checkDate(f, s.ID, "date_published", s.DatePublished, false)
		v.checkDate(f, s.ID, "date_updated", s.DateUpdated, false)
		v.checkDate(f, s.ID, "date_retrieved", &s.DateRetrieved, true)
		for _, cid := range s.ClaimIDs {
			if !v.claimIDs[cid] {
				v.errf(f, s.ID, "claim_ids", "unknown claim id %q", cid)
			}
		}
		if s.SourceTier >= 4 && s.ModelUseStatus.FeedsModel() {
			v.errf(f, s.ID, "model_use_status", "tier %d sources cannot be eligible/used", s.SourceTier)
		}
		if s.RetractionStatus == "retracted" && s.ModelUseStatus.FeedsModel() {
			v.errf(f, s.ID, "model_use_status", "retracted sources cannot be eligible/used")
		}
		if strings.TrimSpace(s.Citation) == "" {
			v.warnf(f, s.ID, "citation", "empty citation")
		}
		v.checkReview(f, s.ID, s.Review)
	}
}

func (v *validator) checkClaims() {
	f := "claims.json"
	for _, c := range v.s.Claims {
		if !reClaimID.MatchString(c.ID) {
			v.errf(f, c.ID, "id", "must match clm-<slug>-<nn>")
		}
		if !v.sourceIDs[c.SourceID] {
			v.errf(f, c.ID, "source_id", "unknown source id %q", c.SourceID)
		}
		if !c.EvidenceType.Valid() {
			v.errf(f, c.ID, "evidence_type", "invalid value %q", c.EvidenceType)
		}
		if !c.Status.Valid() {
			v.errf(f, c.ID, "status", "invalid value %q", c.Status)
		}
		if !c.Relevance.Valid() {
			v.errf(f, c.ID, "relevance", "invalid value %q", c.Relevance)
		}
		v.checkDate(f, c.ID, "date", c.Date, false)
		for _, id := range c.CorroborationIDs {
			if !v.claimIDs[id] {
				v.errf(f, c.ID, "corroboration_ids", "unknown claim id %q", id)
			}
		}
		for _, id := range c.ContradictionIDs {
			if !v.claimIDs[id] {
				v.errf(f, c.ID, "contradiction_ids", "unknown claim id %q", id)
			}
		}
		if strings.TrimSpace(c.Text) == "" {
			v.errf(f, c.ID, "text", "empty claim text")
		}
		v.checkReview(f, c.ID, c.Review)
	}
}

func prob(x *float64) bool { return x == nil || (*x >= 0 && *x <= 1) }

func (v *validator) checkForecasts() {
	f := "forecasts.json"
	type groupSig struct {
		outcomes string
		horizon  string
		cond     string
	}
	groups := map[string]groupSig{}
	groupFirst := map[string]string{}
	for _, fc := range v.s.Forecasts {
		v.checkPrefixedID(f, fc.ID, "fc-")
		if !v.sourceIDs[fc.SourceID] {
			v.errf(f, fc.ID, "source_id", "unknown source id %q", fc.SourceID)
		}
		v.checkDate(f, fc.ID, "date", &fc.Date, true)
		if !fc.Population.Valid() {
			v.errf(f, fc.ID, "population", "invalid value %q", fc.Population)
		}
		if !fc.Status.Valid() {
			v.errf(f, fc.ID, "status", "invalid value %q", fc.Status)
		}
		if len(fc.OutcomeSet) == 0 {
			v.errf(f, fc.ID, "outcome_set", "empty outcome set")
		}
		for _, o := range fc.OutcomeSet {
			if !o.Valid() {
				v.errf(f, fc.ID, "outcome_set", "invalid outcome %q", o)
			}
		}
		if fc.Horizon == schema.HorizonCustom {
			if fc.HorizonNote == nil || strings.TrimSpace(*fc.HorizonNote) == "" {
				v.errf(f, fc.ID, "horizon_note", "custom horizon requires horizon_note")
			}
		} else if !fc.Horizon.Valid() {
			v.errf(f, fc.ID, "horizon", "invalid horizon %q", fc.Horizon)
		}
		if !prob(fc.Mean) || !prob(fc.Median) || !prob(fc.ResponseRate) {
			v.errf(f, fc.ID, "mean", "mean/median/response_rate must be in [0,1]")
		}
		if q := fc.Quantiles; q != nil {
			vals := []*float64{q.P05, q.P25, q.P50, q.P75, q.P95}
			last := -1.0
			for i, p := range vals {
				if p == nil {
					continue
				}
				if *p < 0 || *p > 1 {
					v.errf(f, fc.ID, "quantiles", "quantile out of [0,1]")
				}
				if *p < last {
					v.errf(f, fc.ID, "quantiles", "quantiles must be non-decreasing (index %d)", i)
				}
				last = *p
			}
		}
		if fc.SampleSize != nil && *fc.SampleSize < 0 {
			v.errf(f, fc.ID, "sample_size", "negative sample size")
		}
		if strings.TrimSpace(fc.QuestionWordingOriginal) == "" {
			v.errf(f, fc.ID, "question_wording_original", "question wording is required (§0.5)")
		}
		if _, ok := fc.CentralValue(); !ok && fc.ModelUseStatus.FeedsModel() {
			v.errf(f, fc.ID, "median", "eligible/used forecasts need median, mean or quantiles.p50")
		}
		if fc.GroupID != nil && *fc.GroupID != "" {
			sig := groupSig{schema.OutcomeSetKey(fc.OutcomeSet), string(fc.Horizon), fc.Conditions}
			if prev, ok := groups[*fc.GroupID]; ok {
				if prev.outcomes != sig.outcomes || prev.horizon != sig.horizon {
					v.errf(f, fc.ID, "group_id", "group %q mixes outcome sets or horizons (see %s); never average incompatible forecasts (§0.5)", *fc.GroupID, groupFirst[*fc.GroupID])
				} else if prev.cond != sig.cond {
					v.warnf(f, fc.ID, "group_id", "group %q members differ in conditions text (see %s)", *fc.GroupID, groupFirst[*fc.GroupID])
				}
			} else {
				groups[*fc.GroupID] = sig
				groupFirst[*fc.GroupID] = fc.ID
			}
		}
		v.checkReview(f, fc.ID, fc.Review)
	}
}

func (v *validator) checkBenchmarks() {
	f := "benchmarks.json"
	for _, b := range v.s.Benchmarks {
		v.checkPrefixedID(f, b.ID, "bm-")
		if !b.ContaminationRisk.Valid() {
			v.errf(f, b.ID, "contamination_risk", "invalid value %q", b.ContaminationRisk)
		}
		if !b.Saturation.Valid() {
			v.errf(f, b.ID, "saturation", "invalid value %q", b.Saturation)
		}
		if !b.Direction.Valid() {
			v.errf(f, b.ID, "direction", "invalid value %q", b.Direction)
		}
		if !b.PDoomRelevance.Valid() {
			v.errf(f, b.ID, "pdoom_relevance", "invalid value %q", b.PDoomRelevance)
		}
		if !strings.HasPrefix(b.URL, "http") {
			v.errf(f, b.ID, "url", "must be an absolute URL")
		}
		v.checkSourceRefs(f, b.ID, "source_ids", b.SourceIDs)
		v.checkReview(f, b.ID, b.Review)
	}
}

func (v *validator) checkBenchmarkResults() {
	f := "benchmark_results.json"
	for _, r := range v.s.BenchmarkResults {
		v.checkPrefixedID(f, r.ID, "bmr-")
		if !v.benchmarkIDs[r.BenchmarkID] {
			v.errf(f, r.ID, "benchmark_id", "unknown benchmark id %q", r.BenchmarkID)
		}
		v.checkDate(f, r.ID, "date", &r.Date, true)
		if !r.Confidence.Valid() {
			v.errf(f, r.ID, "confidence", "invalid value %q", r.Confidence)
		}
		if r.CILow != nil && r.CIHigh != nil && *r.CILow > *r.CIHigh {
			v.errf(f, r.ID, "ci_low", "ci_low > ci_high")
		}
		if math.IsNaN(r.Value) || math.IsInf(r.Value, 0) {
			v.errf(f, r.ID, "value", "value must be finite")
		}
		v.checkSourceRefs(f, r.ID, "source_ids", r.SourceIDs)
		v.checkReview(f, r.ID, r.Review)
	}
}

func (v *validator) checkIncidents() {
	f := "incidents.json"
	for _, in := range v.s.Incidents {
		v.checkPrefixedID(f, in.ID, "inc-")
		v.checkDate(f, in.ID, "date", in.Date, false)
		if !in.DatePrecision.Valid() {
			v.errf(f, in.ID, "date_precision", "invalid value %q", in.DatePrecision)
		}
		for _, c := range in.Cause {
			if !c.Valid() {
				v.errf(f, in.ID, "cause", "invalid value %q", c)
			}
		}
		for _, h := range in.Harm {
			if !h.Valid() {
				v.errf(f, in.ID, "harm", "invalid value %q", h)
			}
		}
		if !in.Severity.Valid() {
			v.errf(f, in.ID, "severity", "invalid value %q", in.Severity)
		}
		if !in.PDoomRelevance.Valid() {
			v.errf(f, in.ID, "pdoom_relevance", "invalid value %q", in.PDoomRelevance)
		}
		if !in.EvidenceLevel.Valid() {
			v.errf(f, in.ID, "evidence_level", "invalid value %q", in.EvidenceLevel)
		}
		if !in.Novelty.Valid() {
			v.errf(f, in.ID, "novelty", "invalid value %q", in.Novelty)
		}
		hasExt := in.ExternalIDs.AIID != nil || in.ExternalIDs.OECDAIM != nil || in.ExternalIDs.MITTracker != nil || in.ExternalIDs.CVE != nil || in.ExternalIDs.Docket != nil || in.ExternalIDs.Other != nil
		if !hasExt && v.s.BestTier(in.SourceIDs) > 2 {
			v.errf(f, in.ID, "external_ids", "incidents need an external registry id or a tier 1–2 source (§8.5)")
		}
		if strings.TrimSpace(in.Summary) == "" {
			v.errf(f, in.ID, "summary", "empty summary")
		}
		v.checkSourceRefs(f, in.ID, "source_ids", in.SourceIDs)
		v.checkReview(f, in.ID, in.Review)
	}
}

func (v *validator) checkScenarios() {
	f := "scenarios.json"
	for _, sc := range v.s.Scenarios {
		if !reScenarioID.MatchString(sc.ID) {
			v.errf(f, sc.ID, "id", "must be S1..S18")
		}
		if len(sc.OutcomeSet) == 0 {
			v.errf(f, sc.ID, "outcome_set", "empty outcome set")
		}
		for _, o := range sc.OutcomeSet {
			if !o.Valid() {
				v.errf(f, sc.ID, "outcome_set", "invalid outcome %q", o)
			}
		}
		if !sc.ProbabilitySource.Valid() {
			v.errf(f, sc.ID, "probability_source", "invalid value %q", sc.ProbabilitySource)
		}
		if !sc.Uncertainty.Valid() {
			v.errf(f, sc.ID, "uncertainty", "invalid value %q", sc.Uncertainty)
		}
		if !sc.Recoverability.Valid() {
			v.errf(f, sc.ID, "recoverability", "invalid value %q", sc.Recoverability)
		}
		for _, d := range sc.Dependencies {
			if !v.scenarioIDs[d] {
				v.errf(f, sc.ID, "dependencies", "unknown scenario id %q", d)
			}
		}
		for _, iid := range sc.InterventionIDs {
			if !v.interventionIDs[iid] {
				v.errf(f, sc.ID, "intervention_ids", "unknown intervention id %q", iid)
			}
		}
		v.checkSourceRefs(f, sc.ID, "source_ids", sc.SourceIDs)
		v.checkReview(f, sc.ID, sc.Review)
	}
}

func (v *validator) checkScenarioEdges() {
	f := "scenario_edges.json"
	for _, e := range v.s.ScenarioEdges {
		if !reEdgeID.MatchString(e.ID) {
			v.errf(f, e.ID, "id", "must match se-<from>-<to>")
		}
		if !v.scenarioIDs[e.FromID] {
			v.errf(f, e.ID, "from_id", "unknown scenario id %q", e.FromID)
		}
		if !v.scenarioIDs[e.ToID] {
			v.errf(f, e.ID, "to_id", "unknown scenario id %q", e.ToID)
		}
		if e.ID != "se-"+e.FromID+"-"+e.ToID {
			v.warnf(f, e.ID, "id", "id does not match se-%s-%s", e.FromID, e.ToID)
		}
		if !e.Relation.Valid() {
			v.errf(f, e.ID, "relation", "invalid value %q", e.Relation)
		}
		if !e.Confidence.Valid() {
			v.errf(f, e.ID, "confidence", "invalid value %q", e.Confidence)
		}
		v.checkSourceRefs(f, e.ID, "source_ids", e.SourceIDs)
	}
}

func (v *validator) checkDrivers() {
	f := "drivers.json"
	seenSignals := map[string]string{}
	for _, d := range v.s.Drivers {
		if !reDriverID.MatchString(d.ID) {
			v.errf(f, d.ID, "id", "must be D1..D10")
		}
		for _, sig := range d.Signals {
			if !reSignalID.MatchString(sig.SignalID) {
				v.errf(f, d.ID, "signals", "signal id %q must match <family>.<snake_case>", sig.SignalID)
			} else if !strings.HasPrefix(sig.SignalID, d.ID+".") {
				v.errf(f, d.ID, "signals", "signal id %q does not belong to family %s", sig.SignalID, d.ID)
			}
			if prev, dup := seenSignals[sig.SignalID]; dup {
				v.errf(f, d.ID, "signals", "duplicate signal id %q (also in %s)", sig.SignalID, prev)
			}
			seenSignals[sig.SignalID] = d.ID
			if !sig.Direction.Valid() {
				v.errf(f, d.ID, "signals", "signal %q has invalid direction %q", sig.SignalID, sig.Direction)
			}
			if strings.TrimSpace(sig.Normalization) == "" {
				v.errf(f, d.ID, "signals", "signal %q lacks a normalization description", sig.SignalID)
			}
		}
	}
}

func (v *validator) checkDriverObservations() {
	f := "driver_observations.json"
	for _, o := range v.s.DriverObservations {
		ref, ok := v.signals[o.SignalID]
		if !ok {
			v.errf(f, o.ID, "signal_id", "unknown signal id %q", o.SignalID)
		} else if string(o.Family) != ref.Family {
			v.errf(f, o.ID, "family", "family %q does not match signal family %s", o.Family, ref.Family)
		}
		if !o.Family.Valid() {
			v.errf(f, o.ID, "family", "invalid value %q", o.Family)
		}
		if o.ValueNormalized < 0 || o.ValueNormalized > 1 {
			v.errf(f, o.ID, "value_normalized", "must be in [0,1], got %g", o.ValueNormalized)
		}
		if o.Confidence < 0 || o.Confidence > 1 {
			v.errf(f, o.ID, "confidence", "must be in [0,1], got %g", o.Confidence)
		}
		if !o.ObservationKind.Valid() {
			v.errf(f, o.ID, "observation_kind", "invalid value %q", o.ObservationKind)
		}
		v.checkDate(f, o.ID, "as_of", &o.AsOf, true)
		if strings.TrimSpace(o.Rationale) == "" {
			v.errf(f, o.ID, "rationale", "empty rationale")
		}
		if len(o.SourceIDs) == 0 && o.ModelUseStatus.FeedsModel() {
			v.errf(f, o.ID, "source_ids", "eligible/used observations need at least one source")
		}
		v.checkSourceRefs(f, o.ID, "source_ids", o.SourceIDs)
		v.checkReview(f, o.ID, o.Review)
	}
}

func (v *validator) checkInterventions() {
	f := "interventions.json"
	for _, in := range v.s.Interventions {
		if !reInterID.MatchString(in.ID) {
			v.errf(f, in.ID, "id", "must match I01..")
		}
		for _, sid := range in.TargetScenarioIDs {
			if !v.scenarioIDs[sid] {
				v.errf(f, in.ID, "target_scenario_ids", "unknown scenario id %q", sid)
			}
		}
		if !in.EvidenceStrength.Valid() {
			v.errf(f, in.ID, "evidence_strength", "invalid value %q", in.EvidenceStrength)
		}
		if !in.Cost.Valid() {
			v.errf(f, in.ID, "cost", "invalid value %q", in.Cost)
		}
		if !in.TimeToDeploy.Valid() {
			v.errf(f, in.ID, "time_to_deploy", "invalid value %q", in.TimeToDeploy)
		}
		if !in.EffectSize.Valid() {
			v.errf(f, in.ID, "effect_size", "invalid value %q", in.EffectSize)
		}
		if !in.Uncertainty.Valid() {
			v.errf(f, in.ID, "uncertainty", "invalid value %q", in.Uncertainty)
		}
		if !in.Category.Valid() {
			v.errf(f, in.ID, "category", "invalid value %q", in.Category)
		}
		v.checkSourceRefs(f, in.ID, "source_ids", in.SourceIDs)
		v.checkReview(f, in.ID, in.Review)
	}
}

func (v *validator) checkOrganizations() {
	f := "organizations.json"
	for _, o := range v.s.Organizations {
		v.checkPrefixedID(f, o.ID, "org-")
		if !strings.HasPrefix(o.URL, "http") {
			v.errf(f, o.ID, "url", "must be an absolute URL")
		}
		v.checkDate(f, o.ID, "last_verified", &o.LastVerified, true)
		v.checkSourceRefs(f, o.ID, "source_ids", o.SourceIDs)
		v.checkReview(f, o.ID, o.Review)
	}
}

func (v *validator) checkActions() {
	f := "actions.json"
	for _, a := range v.s.Actions {
		if !strings.HasPrefix(a.ID, "act-") {
			v.errf(f, a.ID, "id", "must match act-<audience>-<slug>")
		}
		if !a.Audience.Valid() {
			v.errf(f, a.ID, "audience", "invalid value %q", a.Audience)
		}
		if !a.Effort.Valid() {
			v.errf(f, a.ID, "effort", "invalid value %q", a.Effort)
		}
		for _, iid := range a.RelatedInterventionIDs {
			if !v.interventionIDs[iid] {
				v.errf(f, a.ID, "related_intervention_ids", "unknown intervention id %q", iid)
			}
		}
		for i, r := range a.Resources {
			if r.SourceID != nil && !v.sourceIDs[*r.SourceID] {
				v.errf(f, a.ID, fmt.Sprintf("resources[%d].source_id", i), "unknown source id %q", *r.SourceID)
			}
		}
		v.checkReview(f, a.ID, a.Review)
	}
}

func (v *validator) checkModelSpec() {
	f := "model_spec.json"
	m := v.s.ModelSpec
	id := m.ID
	if !strings.HasPrefix(id, "pdoom-model-spec@") {
		v.errf(f, id, "id", "must match pdoom-model-spec@<semver>")
	}
	lo, hi := 0.0, 0.35
	if len(m.WeightBounds) == 2 {
		lo, hi = m.WeightBounds[0], m.WeightBounds[1]
		if lo != 0 || hi != 0.35 {
			v.warnf(f, id, "weight_bounds", "expected [0, 0.35], got %v", m.WeightBounds)
		}
	} else {
		v.errf(f, id, "weight_bounds", "must be [lo, hi]")
	}
	checkWeights := func(name string, w map[string]float64) {
		if len(w) == 0 {
			v.errf(f, id, "index_weights."+name, "no signal weights")
			return
		}
		sum := 0.0
		for sig, wt := range w {
			if _, ok := v.signals[sig]; !ok {
				v.errf(f, id, "index_weights."+name, "unknown signal id %q", sig)
			}
			if wt < lo || wt > hi {
				v.errf(f, id, "index_weights."+name, "weight %g for %s outside [%g, %g]", wt, sig, lo, hi)
			}
			sum += wt
		}
		if math.Abs(sum-1) > 0.01 {
			v.errf(f, id, "index_weights."+name, "weights sum to %.4f, expected 1", sum)
		}
	}
	checkWeights("capability_pressure", m.IndexWeights.CapabilityPressure)
	checkWeights("control_strength", m.IndexWeights.ControlStrength)
	checkWeights("agentic_infrastructure_risk", m.IndexWeights.AgenticInfrastructureRisk)
	if u := m.IndexWeights.Uncertainty; len(u) > 0 {
		sum := 0.0
		for k, wt := range u {
			if wt < 0 {
				v.errf(f, id, "index_weights.uncertainty", "negative weight for %s", k)
			}
			sum += wt
		}
		if math.Abs(sum-1) > 0.01 {
			v.warnf(f, id, "index_weights.uncertainty", "weights sum to %.4f, expected 1", sum)
		}
	}
	for _, t := range []string{"1", "2", "3", "4", "5"} {
		if _, ok := m.TierMultipliers[t]; !ok {
			v.errf(f, id, "tier_multipliers", "missing tier %s", t)
		}
	}
	if m.TierMultipliers["4"] != 0 || m.TierMultipliers["5"] != 0 {
		v.errf(f, id, "tier_multipliers", "tiers 4 and 5 must have multiplier 0")
	}
	is := m.IncidentScoring
	for _, sev := range schema.Severities {
		if _, ok := is.SeverityWeights[sev]; !ok {
			v.errf(f, id, "incident_scoring.severity_weights", "missing %s", sev)
		}
	}
	for _, rel := range schema.Relevances {
		if _, ok := is.RelevanceWeights[rel]; !ok {
			v.errf(f, id, "incident_scoring.relevance_weights", "missing %s", rel)
		}
	}
	for _, ev := range schema.EvidenceLevels {
		if _, ok := is.EvidenceWeights[ev]; !ok {
			v.errf(f, id, "incident_scoring.evidence_weights", "missing %s", ev)
		}
	}
	if is.RecencyHalfLifeDays <= 0 {
		v.errf(f, id, "incident_scoring.recency_half_life_days", "must be positive")
	}
	if is.SquashK <= 0 {
		v.errf(f, id, "incident_scoring.squash_k", "must be positive")
	}
	if len(m.EditorialRules) == 0 {
		v.errf(f, id, "editorial_rules", "at least one rule is required")
	}
	for i, r := range m.EditorialRules {
		if !r.Level.Valid() {
			v.errf(f, id, fmt.Sprintf("editorial_rules[%d].level", i), "invalid level %q", r.Level)
		}
		if _, err := schema.ParseRule(r.When); err != nil {
			v.errf(f, id, fmt.Sprintf("editorial_rules[%d].when", i), "%v", err)
		}
	}
	rr := m.RoundingRules
	if rr.Extreme != 5 || rr.High != 2 || rr.Moderate != 1 || rr.Low != 1 {
		v.errf(f, id, "rounding_rules", "must be {extreme:5, high:2, moderate:1, low:1} (§0.8)")
	}
	if len(m.AggregationMethods) == 0 {
		v.errf(f, id, "aggregation_methods", "empty")
	}
	for _, am := range m.AggregationMethods {
		if !am.Valid() {
			v.errf(f, id, "aggregation_methods", "invalid method %q", am)
		}
	}
	ec := m.ExperimentalCausal
	if ec.Seed == 0 || ec.Samples <= 0 {
		v.errf(f, id, "experimental_causal", "seed and samples must be positive")
	}
	if ec.CommonFactorLoading < 0 || ec.CommonFactorLoading > 1 {
		v.errf(f, id, "experimental_causal.common_factor_loading", "must be in [0,1]")
	}
	if len(ec.Horizons) == 0 {
		v.errf(f, id, "experimental_causal.horizons", "no horizons")
	}
	if !strings.HasPrefix(ec.Version, "pdoom-model/experimental-causal@") {
		v.errf(f, id, "experimental_causal.version", "must match pdoom-model/experimental-causal@<semver>")
	}
	for h, hs := range ec.Horizons {
		if !schema.Horizon(h).Valid() {
			v.errf(f, id, "experimental_causal.horizons", "invalid horizon key %q", h)
		}
		field := "experimental_causal.horizons." + h
		for name, tq := range map[string]schema.TriQuantile{"A": hs.A, "C": hs.C, "E": hs.E, "F": hs.F} {
			v.checkTri(f, id, field+"."+name, tq)
		}
		sum := 0.0
		for o, tq := range hs.O {
			if !schema.Outcome(o).Valid() {
				v.errf(f, id, field+".O", "invalid outcome key %q", o)
			}
			isPDoom := false
			for _, po := range schema.PDoomOutcomes {
				if string(po) == o {
					isPDoom = true
				}
			}
			if !isPDoom {
				v.errf(f, id, field+".O", "outcome %q is not one of O3..O8", o)
			}
			v.checkTri(f, id, field+".O."+o, tq)
			sum += tq.P50
		}
		if sum > 1+1e-9 {
			v.errf(f, id, field+".O", "sum of p50 outcome shares is %.3f > 1", sum)
		}
		if len(hs.O) == 0 {
			v.errf(f, id, field+".O", "no outcome shares")
		}
	}
	v.checkSourceRefs(f, id, "experimental_causal.source_ids", ec.SourceIDs)
}

func (v *validator) checkTri(file, id, field string, t schema.TriQuantile) {
	for _, p := range []float64{t.P05, t.P50, t.P95} {
		if p <= 0 || p >= 1 {
			v.errf(file, id, field, "quantiles must be strictly inside (0,1)")
			return
		}
	}
	if !t.Ordered() {
		v.errf(file, id, field, "quantiles must satisfy p05 < p50 < p95")
	}
}
