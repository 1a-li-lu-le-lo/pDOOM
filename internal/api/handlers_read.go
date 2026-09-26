// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/publishing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

var (
	reReleaseID  = regexp.MustCompile(`^rel-\d{4}-\d{2}-\d{2}-\d{3}$`)
	reScenarioID = regexp.MustCompile(`^S\d{1,3}$`)
	reSourceID   = regexp.MustCompile(`^src-[a-z0-9]+(?:-[a-z0-9]+)*$`)
	reSlug       = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
)

// Headline is the compact per-release summary served by /v1/meter and
// /v1/meter/history. Every estimate row keeps its status, horizon and outcome
// set; the official row publishes no probability in this release line.
type Headline struct {
	ReleaseID                string                    `json:"release_id"`
	DataSnapshot             string                    `json:"data_snapshot"`
	DataCutoff               string                    `json:"data_cutoff"`
	GeneratedAt              string                    `json:"generated_at"`
	Published                *string                   `json:"published"`
	Superseded               *schema.Superseded        `json:"superseded"`
	IsCurrent                bool                      `json:"is_current"`
	ModelVersions            []string                  `json:"model_versions"`
	EditorialRiskLevel       schema.EditorialRiskLevel `json:"editorial_risk_level"`
	UncertaintyScore         *float64                  `json:"uncertainty_score"`
	Official                 []schema.EstimateSummary  `json:"official"`
	ExternalAggregates       []schema.EstimateSummary  `json:"external_aggregates"`
	ResearchMode             []schema.EstimateSummary  `json:"research_mode"`
	Indexes                  []IndexHeadline           `json:"indexes"`
	Changes                  []string                  `json:"changes"`
	HeightenedReviewTriggers []string                  `json:"heightened_review_triggers"`
}

// IndexHeadline is one index value in a headline. Indexes are never probabilities.
type IndexHeadline struct {
	IndexID       schema.IndexID `json:"index_id"`
	Value         *float64       `json:"value"`
	Label         string         `json:"label"`
	IsProbability bool           `json:"is_probability"`
}

func headlineOf(r *schema.Release, isCurrent bool) Headline {
	m := r.Manifest
	h := Headline{
		ReleaseID: m.ReleaseID, DataSnapshot: m.DataSnapshot, DataCutoff: m.SourceCutoff, GeneratedAt: m.GeneratedAt,
		Published: m.Published, Superseded: m.Superseded, IsCurrent: isCurrent, ModelVersions: nonNil(m.ModelVersions),
		EditorialRiskLevel: m.EditorialRiskLevel, UncertaintyScore: m.UncertaintyScore,
		Official: []schema.EstimateSummary{}, ExternalAggregates: []schema.EstimateSummary{}, ResearchMode: []schema.EstimateSummary{},
		Indexes: []IndexHeadline{}, Changes: nonNil(m.Changes), HeightenedReviewTriggers: nonNil(r.Delta.HeightenedReviewTriggers),
	}
	for _, e := range m.Estimates {
		switch e.Status {
		case schema.StatusOfficial, schema.StatusInsufficientlyCalibrate:
			h.Official = append(h.Official, e)
		case schema.StatusExternalAggregate:
			h.ExternalAggregates = append(h.ExternalAggregates, e)
		case schema.StatusResearchMode:
			h.ResearchMode = append(h.ResearchMode, e)
		}
	}
	for _, iv := range r.Indexes {
		h.Indexes = append(h.Indexes, IndexHeadline{IndexID: iv.IndexID, Value: iv.Value, Label: iv.Label, IsProbability: iv.IsProbability})
	}
	return h
}

// MeterResponse is the body of GET /v1/meter.
type MeterResponse struct {
	Meta               Meta                      `json:"meta"`
	Manifest           schema.ReleaseManifest    `json:"manifest"`
	Official           []schema.Estimate         `json:"official"`
	ExternalAggregates []schema.Estimate         `json:"external_aggregates"`
	ResearchMode       []schema.Estimate         `json:"research_mode"`
	Indexes            []schema.IndexValue       `json:"indexes"`
	EditorialRiskLevel schema.EditorialRiskLevel `json:"editorial_risk_level"`
	UncertaintyScore   *float64                  `json:"uncertainty_score"`
	UncertaintyLabel   schema.UncertaintyLabel   `json:"uncertainty_label"`
	DataCutoff         string                    `json:"data_cutoff"`
	LastReviewed       *string                   `json:"last_reviewed"`
	LastPublished      *string                   `json:"last_published"`
	ModelVersions      []string                  `json:"model_versions"`
	Limitations        []string                  `json:"limitations"`
	Headline           Headline                  `json:"headline"`
}

func (s *server) handleMeter(w http.ResponseWriter, r *http.Request, st *state) {
	rel := st.release
	m := rel.Manifest
	resp := MeterResponse{
		Meta: st.meta(), Manifest: m,
		Official: []schema.Estimate{}, ExternalAggregates: []schema.Estimate{}, ResearchMode: []schema.Estimate{},
		Indexes: nonNil(rel.Indexes), EditorialRiskLevel: m.EditorialRiskLevel, UncertaintyScore: m.UncertaintyScore,
		DataCutoff: m.SourceCutoff, LastReviewed: lastReviewed(rel.Approvals), LastPublished: m.Published,
		ModelVersions: nonNil(m.ModelVersions), Limitations: nonNil(m.KnownLimitations), Headline: headlineOf(rel, true),
	}
	for _, e := range rel.Estimates {
		switch e.Status {
		case schema.StatusOfficial, schema.StatusInsufficientlyCalibrate:
			resp.Official = append(resp.Official, e)
			if resp.UncertaintyLabel == "" {
				resp.UncertaintyLabel = e.Uncertainty
			}
		case schema.StatusExternalAggregate:
			resp.ExternalAggregates = append(resp.ExternalAggregates, e)
		case schema.StatusResearchMode:
			resp.ResearchMode = append(resp.ResearchMode, e)
		}
	}
	s.ok(w, r, st, resp)
}

func lastReviewed(approvals []schema.Approval) *string {
	var latest string
	for _, a := range approvals {
		if a.SignedAt > latest {
			latest = a.SignedAt
		}
	}
	if latest == "" {
		return nil
	}
	return &latest
}

// HistoryResponse is the body of GET /v1/meter/history.
type HistoryResponse struct {
	Meta     Meta       `json:"meta"`
	Releases []Headline `json:"releases"`
}

func (s *server) handleMeterHistory(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, HistoryResponse{Meta: st.meta(), Releases: nonNil(st.history)})
}

// OutcomeDefinition describes one outcome code (build-spec §3.3).
type OutcomeDefinition struct {
	Code        schema.Outcome `json:"code"`
	Slug        string         `json:"slug"`
	Label       string         `json:"label"`
	Description string         `json:"description"`
	IncludedIn  IncludedIn     `json:"included_in"`
}

// IncludedIn flags the derived sets an outcome belongs to.
type IncludedIn struct {
	PDoom          bool `json:"pdoom"`
	Extinction     bool `json:"extinction"`
	Disempowerment bool `json:"disempowerment"`
	Collapse       bool `json:"collapse"`
	Biosphere      bool `json:"biosphere"`
}

// DerivedSet is a named outcome set such as P_DOOM.
type DerivedSet struct {
	Key      string           `json:"key"`
	Label    string           `json:"label"`
	Outcomes []schema.Outcome `json:"outcomes"`
	Note     string           `json:"note"`
}

// OutcomesResponse is the body of GET /v1/outcomes.
type OutcomesResponse struct {
	Outcomes    []OutcomeDefinition `json:"outcomes"`
	DerivedSets []DerivedSet        `json:"derived_sets"`
	Note        string              `json:"note"`
}

var outcomeDescriptions = map[schema.Outcome]string{
	schema.O0: "Advanced AI is developed and deployed with harms that remain bounded, correctable and within the capacity of existing institutions to manage.",
	schema.O1: "Large-scale harm occurs — economic, social or physical — but societies retain the ability to recover and to correct course within a generation.",
	schema.O2: "AI-enabled concentration of power by a state or a small set of organisations that substantially reduces political and economic freedom but is not judged permanent.",
	schema.O3: "Humanity loses, in a way judged practically irreversible, the ability to direct its own future — whether to AI systems or to a narrow group controlling them.",
	schema.O4: "A breakdown of global-scale institutions, infrastructure and population that is not recovered from within centuries, with humanity surviving in reduced form.",
	schema.O5: "Human population falls to a small fraction of its current level with recovery uncertain; the species survives.",
	schema.O6: "No living humans remain.",
	schema.O7: "Irreversible destruction of much of Earth's biosphere attributable to AI-driven activity, whether or not humans survive.",
	schema.O8: "An unrecoverable loss of value not captured by O3–O7, such as permanent lock-in of a substantially worse trajectory for humanity.",
}

func contains(set []schema.Outcome, o schema.Outcome) bool {
	for _, x := range set {
		if x == o {
			return true
		}
	}
	return false
}

func outcomesResponse() OutcomesResponse {
	resp := OutcomesResponse{Note: "Outcomes are never silently combined. The combined p(DOOM) set (O3–O8) may only be shown next to its decomposition, and every probability carries its outcome set, horizon and conditioning."}
	for _, o := range schema.Outcomes {
		resp.Outcomes = append(resp.Outcomes, OutcomeDefinition{
			Code: o, Slug: schema.OutcomeSlugs[o], Label: schema.OutcomeLabels[o], Description: outcomeDescriptions[o],
			IncludedIn: IncludedIn{
				PDoom: contains(schema.PDoomOutcomes, o), Extinction: contains(schema.ExtinctionOutcomes, o),
				Disempowerment: contains(schema.DisempowermentOutcomes, o), Collapse: contains(schema.CollapseOutcomes, o),
				Biosphere: contains(schema.BiosphereOutcomes, o),
			},
		})
	}
	resp.DerivedSets = []DerivedSet{
		{Key: "P_DOOM", Label: schema.OutcomeSetLabel(schema.PDoomOutcomes), Outcomes: schema.PDoomOutcomes, Note: "Shown only next to its per-outcome decomposition."},
		{Key: "P_EXTINCTION", Label: schema.OutcomeSetLabel(schema.ExtinctionOutcomes), Outcomes: schema.ExtinctionOutcomes, Note: ""},
		{Key: "P_DISEMPOWERMENT", Label: schema.OutcomeSetLabel(schema.DisempowermentOutcomes), Outcomes: schema.DisempowermentOutcomes, Note: ""},
		{Key: "P_COLLAPSE", Label: schema.OutcomeSetLabel(schema.CollapseOutcomes), Outcomes: schema.CollapseOutcomes, Note: ""},
		{Key: "P_BIOSPHERE", Label: schema.OutcomeSetLabel(schema.BiosphereOutcomes), Outcomes: schema.BiosphereOutcomes, Note: ""},
	}
	return resp
}

func (s *server) handleOutcomes(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, outcomesResponse())
}

// HorizonDefinition describes one horizon key.
type HorizonDefinition struct {
	Key   schema.Horizon `json:"key"`
	Label string         `json:"label"`
	Years *int           `json:"years"`
	Kind  string         `json:"kind"`
	Note  string         `json:"note"`
}

// HorizonsResponse is the body of GET /v1/horizons.
type HorizonsResponse struct {
	Horizons []HorizonDefinition `json:"horizons"`
	Note     string              `json:"note"`
}

func horizonsResponse() HorizonsResponse {
	years := map[schema.Horizon]int{schema.Horizon1y: 1, schema.Horizon3y: 3, schema.Horizon5y: 5, schema.Horizon10y: 10, schema.Horizon25y: 25}
	resp := HorizonsResponse{Note: "Horizons are measured from forecast_origin_date. Forecasts with different horizons are never averaged together."}
	for _, h := range schema.Horizons {
		d := HorizonDefinition{Key: h, Label: schema.HorizonLabel(h), Kind: "duration"}
		if y, ok := years[h]; ok {
			yy := y
			d.Years = &yy
		}
		switch h {
		case schema.Horizon2100:
			d.Kind = "calendar_year"
			d.Note = "Calendar-year endpoint; the elapsed span depends on forecast_origin_date, so it is not interchangeable with a fixed duration."
		case schema.HorizonEventual:
			d.Kind = "open_ended"
			d.Note = "Open-ended horizon with no endpoint. Eventual probabilities are not comparable with dated horizons, cannot be converted to a rate, and are never displayed as a countdown or an implied date."
		}
		resp.Horizons = append(resp.Horizons, d)
	}
	return resp
}

func (s *server) handleHorizons(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, horizonsResponse())
}

// DriversResponse is the body of GET /v1/drivers.
type DriversResponse struct {
	Meta         Meta                       `json:"meta"`
	Drivers      []schema.Driver            `json:"drivers"`
	Observations []schema.DriverObservation `json:"observations"`
	Explained    schema.DriversExplained    `json:"drivers_explained"`
}

func (s *server) handleDrivers(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, DriversResponse{Meta: st.meta(), Drivers: nonNil(st.snapshot.Drivers), Observations: nonNil(st.snapshot.DriverObservations), Explained: st.release.DriversExplained})
}

// ScenariosResponse is the body of GET /v1/scenarios.
type ScenariosResponse struct {
	Meta      Meta                  `json:"meta"`
	Scenarios []schema.Scenario     `json:"scenarios"`
	Edges     []schema.ScenarioEdge `json:"edges"`
	Note      string                `json:"note"`
}

const scenarioNote = "Scenarios are category-level pathway descriptions. No probability is attached to an individual scenario unless probability_source says otherwise; the estimates on /v1/meter carry their own outcome set and horizon."

func (s *server) handleScenarios(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, ScenariosResponse{Meta: st.meta(), Scenarios: nonNil(st.snapshot.Scenarios), Edges: nonNil(st.snapshot.ScenarioEdges), Note: scenarioNote})
}

// ScenarioResponse is the body of GET /v1/scenarios/{id}.
type ScenarioResponse struct {
	Meta          Meta                  `json:"meta"`
	Scenario      schema.Scenario       `json:"scenario"`
	Edges         []schema.ScenarioEdge `json:"edges"`
	Interventions []schema.Intervention `json:"interventions"`
	Note          string                `json:"note"`
}

func (s *server) handleScenario(w http.ResponseWriter, r *http.Request, st *state) {
	id := r.PathValue("id")
	if !reScenarioID.MatchString(id) {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no scenario with id "+strconv.Quote(id))
		return
	}
	for _, sc := range st.snapshot.Scenarios {
		if sc.ID != id {
			continue
		}
		resp := ScenarioResponse{Meta: st.meta(), Scenario: sc, Edges: []schema.ScenarioEdge{}, Interventions: []schema.Intervention{}, Note: scenarioNote}
		for _, e := range st.snapshot.ScenarioEdges {
			if e.FromID == id || e.ToID == id {
				resp.Edges = append(resp.Edges, e)
			}
		}
		for _, iv := range st.snapshot.Interventions {
			if inStrings(sc.InterventionIDs, iv.ID) || inStrings(iv.TargetScenarioIDs, id) {
				resp.Interventions = append(resp.Interventions, iv)
			}
		}
		s.ok(w, r, st, resp)
		return
	}
	s.writeError(w, r, http.StatusNotFound, "not_found", "no scenario with id "+strconv.Quote(id))
}

func inStrings(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// ForecastsResponse is the body of GET /v1/forecasts. Records are quoted as
// published; aggregations exist only inside documented compatibility groups.
type ForecastsResponse struct {
	Meta         Meta                       `json:"meta"`
	Status       schema.EstimateStatus      `json:"status"`
	Producer     string                     `json:"producer"`
	Forecasts    []schema.Forecast          `json:"forecasts"`
	Aggregations []schema.AggregationResult `json:"aggregations"`
	Estimates    []schema.Estimate          `json:"estimates"`
	Note         string                     `json:"note"`
}

func (s *server) handleForecasts(w http.ResponseWriter, r *http.Request, st *state) {
	resp := ForecastsResponse{Meta: st.meta(), Status: schema.StatusExternalAggregate, Forecasts: nonNil(st.snapshot.Forecasts), Aggregations: nonNil(st.release.Aggregations), Estimates: []schema.Estimate{},
		Note: "Each forecast record keeps its original question wording, outcome set, horizon and conditioning. Aggregations combine only forecasts of one compatibility group and are shown with every alternative method."}
	for _, e := range st.release.Estimates {
		if e.Status == schema.StatusExternalAggregate {
			resp.Estimates = append(resp.Estimates, e)
			if resp.Producer == "" {
				resp.Producer = e.Producer
			}
		}
	}
	if resp.Producer == "" {
		for _, v := range st.release.Manifest.ModelVersions {
			if strings.Contains(v, "external-aggregate") {
				resp.Producer = v
			}
		}
	}
	s.ok(w, r, st, resp)
}

// CapabilitiesResponse is the body of GET /v1/capabilities.
type CapabilitiesResponse struct {
	Meta       Meta                     `json:"meta"`
	Benchmarks []schema.Benchmark       `json:"benchmarks"`
	Results    []schema.BenchmarkResult `json:"results"`
}

func (s *server) handleCapabilities(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, CapabilitiesResponse{Meta: st.meta(), Benchmarks: nonNil(st.snapshot.Benchmarks), Results: nonNil(st.snapshot.BenchmarkResults)})
}

// IncidentsResponse is the body of GET /v1/incidents.
type IncidentsResponse struct {
	Meta      Meta              `json:"meta"`
	Incidents []schema.Incident `json:"incidents"`
}

func (s *server) handleIncidents(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, IncidentsResponse{Meta: st.meta(), Incidents: nonNil(st.snapshot.Incidents)})
}

// SafeguardsResponse is the body of GET /v1/safeguards.
type SafeguardsResponse struct {
	Meta          Meta                  `json:"meta"`
	Interventions []schema.Intervention `json:"interventions"`
}

func (s *server) handleSafeguards(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, SafeguardsResponse{Meta: st.meta(), Interventions: nonNil(st.snapshot.Interventions)})
}

// SourcesResponse is the body of GET /v1/sources.
type SourcesResponse struct {
	Meta    Meta            `json:"meta"`
	Total   int             `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	Sources []schema.Source `json:"sources"`
}

func (s *server) handleSources(w http.ResponseWriter, r *http.Request, st *state) {
	q := r.URL.Query()
	tier := 0
	if v := q.Get("tier"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || !schema.SourceTier(n).Valid() {
			s.writeError(w, r, http.StatusBadRequest, "bad_request", "tier must be an integer 1..5")
			return
		}
		tier = n
	}
	topic := strings.ToLower(strings.TrimSpace(q.Get("topic")))
	text := strings.ToLower(strings.TrimSpace(q.Get("q")))
	if len(topic) > 100 || len(text) > 200 {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "topic (≤100) or q (≤200) is too long")
		return
	}
	limit, offset, err := pagination(q.Get("limit"), q.Get("offset"))
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	var matched []schema.Source
	for _, src := range st.snapshot.Sources {
		if tier != 0 && int(src.SourceTier) != tier {
			continue
		}
		if topic != "" && !hasTopic(src.Topic, topic) {
			continue
		}
		if text != "" && !sourceMatches(src, text) {
			continue
		}
		matched = append(matched, src)
	}
	resp := SourcesResponse{Meta: st.meta(), Total: len(matched), Limit: limit, Offset: offset, Sources: []schema.Source{}}
	if offset < len(matched) {
		end := offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		resp.Sources = matched[offset:end]
	}
	s.ok(w, r, st, resp)
}

func pagination(limitStr, offsetStr string) (limit, offset int, err error) {
	limit = DefaultPageLimit
	if limitStr != "" {
		n, e := strconv.Atoi(limitStr)
		if e != nil || n < 1 || n > MaxPageLimit {
			return 0, 0, errors.New("limit must be an integer 1.." + strconv.Itoa(MaxPageLimit))
		}
		limit = n
	}
	if offsetStr != "" {
		n, e := strconv.Atoi(offsetStr)
		if e != nil || n < 0 || n > 1_000_000 {
			return 0, 0, errors.New("offset must be a non-negative integer")
		}
		offset = n
	}
	return limit, offset, nil
}

func hasTopic(topics []string, want string) bool {
	for _, t := range topics {
		if strings.ToLower(t) == want {
			return true
		}
	}
	return false
}

func sourceMatches(src schema.Source, text string) bool {
	fields := []string{src.ID, src.Title, src.Publisher, src.Citation, src.EvidenceSummary, src.CanonicalURL}
	fields = append(fields, src.Authors...)
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), text) {
			return true
		}
	}
	return false
}

// SourceResponse is the body of GET /v1/sources/{id}.
type SourceResponse struct {
	Meta   Meta           `json:"meta"`
	Source schema.Source  `json:"source"`
	Claims []schema.Claim `json:"claims"`
}

func (s *server) handleSource(w http.ResponseWriter, r *http.Request, st *state) {
	id := r.PathValue("id")
	if !reSourceID.MatchString(id) {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no source with id "+strconv.Quote(id))
		return
	}
	src := st.snapshot.SourceByID(id)
	if src == nil {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no source with id "+strconv.Quote(id))
		return
	}
	resp := SourceResponse{Meta: st.meta(), Source: *src, Claims: []schema.Claim{}}
	for _, c := range st.snapshot.Claims {
		if c.SourceID == id {
			resp.Claims = append(resp.Claims, c)
		}
	}
	s.ok(w, r, st, resp)
}

// MethodologyDoc is one entry of GET /v1/methodology.
type MethodologyDoc struct {
	Slug  string `json:"slug"`
	Path  string `json:"path"`
	Title string `json:"title"`
}

// MethodologyResponse is the body of GET /v1/methodology.
type MethodologyResponse struct {
	Documents []MethodologyDoc `json:"documents"`
}

func (s *server) listMethodology() ([]MethodologyDoc, error) {
	entries, err := os.ReadDir(s.methodDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []MethodologyDoc{}, nil
		}
		return nil, err
	}
	docs := []MethodologyDoc{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		slug := strings.TrimSuffix(name, ".md")
		if !reSlug.MatchString(slug) {
			continue
		}
		title := slug
		if raw, err := os.ReadFile(filepath.Join(s.methodDir, name)); err == nil {
			title = markdownTitle(raw, slug)
		}
		docs = append(docs, MethodologyDoc{Slug: slug, Path: "docs/method/" + name, Title: title})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].Slug < docs[j].Slug })
	return docs, nil
}

// markdownTitle returns the first level-one heading, or the slug.
func markdownTitle(raw []byte, fallback string) string {
	for _, line := range strings.Split(string(raw), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "# ") {
			if title := strings.TrimSpace(strings.TrimPrefix(t, "# ")); title != "" {
				return title
			}
		}
	}
	return fallback
}

func (s *server) handleMethodology(w http.ResponseWriter, r *http.Request, st *state) {
	docs, err := s.listMethodology()
	if err != nil {
		s.log.Error("list methodology", "error", err.Error())
		s.writeError(w, r, http.StatusInternalServerError, "internal", "methodology directory could not be read")
		return
	}
	s.ok(w, r, st, MethodologyResponse{Documents: docs})
}

func (s *server) handleMethodologyDoc(w http.ResponseWriter, r *http.Request, st *state) {
	slug := r.PathValue("slug")
	if !reSlug.MatchString(slug) {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no methodology document "+strconv.Quote(slug))
		return
	}
	raw, err := os.ReadFile(filepath.Join(s.methodDir, slug+".md"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.writeError(w, r, http.StatusNotFound, "not_found", "no methodology document "+strconv.Quote(slug))
			return
		}
		s.log.Error("read methodology", "slug", slug, "error", err.Error())
		s.writeError(w, r, http.StatusInternalServerError, "internal", "methodology document could not be read")
		return
	}
	s.serveBody(w, r, raw, contentTypeMarkdown)
}

// ReleasesResponse is the body of GET /v1/releases.
type ReleasesResponse struct {
	Current  string                  `json:"current"`
	Releases []schema.ReleaseSummary `json:"releases"`
}

// handleReleases lists published releases. With a loaded state the list is
// served from memory (and cached); without one it is read best-effort from
// disk so that the route keeps answering while no release is loaded.
func (s *server) handleReleases(w http.ResponseWriter, r *http.Request, st *state) {
	if st != nil {
		s.ok(w, r, st, ReleasesResponse{Current: st.currentID, Releases: nonNil(st.releases)})
		return
	}
	cur, _ := publishing.CurrentReleaseID(s.deps.Paths)
	resp := ReleasesResponse{Current: cur, Releases: []schema.ReleaseSummary{}}
	for _, rel := range scanPublishedReleases(s.deps.Paths, s.log, nil) {
		resp.Releases = append(resp.Releases, summaryOf(rel, rel.Manifest.ReleaseID == cur))
	}
	s.ok(w, r, st, resp)
}

// ReleaseResponse is the body of GET /v1/releases/{id}: the complete release
// directory as promoted.
type ReleaseResponse struct {
	Meta      Meta            `json:"meta"`
	IsCurrent bool            `json:"is_current"`
	Release   *schema.Release `json:"release"`
}

// handleRelease serves one published release: from the loaded state when
// there is one (unpublished or unreadable directories are absent from it),
// otherwise from disk with the same published check.
func (s *server) handleRelease(w http.ResponseWriter, r *http.Request, st *state) {
	id := r.PathValue("id")
	if !reReleaseID.MatchString(id) {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no release with id "+strconv.Quote(id))
		return
	}
	if st != nil {
		rel, ok := st.byID[id]
		if !ok {
			s.writeError(w, r, http.StatusNotFound, "not_found", "no release with id "+strconv.Quote(id))
			return
		}
		s.ok(w, r, st, ReleaseResponse{Meta: metaOf(rel.Manifest), IsCurrent: id == st.currentID, Release: rel})
		return
	}
	rel, err := publishing.LoadRelease(s.deps.Paths, id)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			s.log.Warn("skipping unreadable release directory", "release_id", id, "error", err.Error())
		}
		s.writeError(w, r, http.StatusNotFound, "not_found", "no release with id "+strconv.Quote(id))
		return
	}
	if rel.Manifest.Published == nil || rel.Manifest.ReleaseID != id {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no release with id "+strconv.Quote(id))
		return
	}
	cur, _ := publishing.CurrentReleaseID(s.deps.Paths)
	s.ok(w, r, st, ReleaseResponse{Meta: metaOf(rel.Manifest), IsCurrent: cur == id, Release: rel})
}

// DefinitionsResponse is the body of GET /v1/definitions.
type DefinitionsResponse struct {
	Meta        Meta                `json:"meta"`
	Definitions []schema.Definition `json:"definitions"`
}

func (s *server) handleDefinitions(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, DefinitionsResponse{Meta: st.meta(), Definitions: nonNil(st.snapshot.Definitions)})
}

// OrganizationsResponse is the body of GET /v1/organizations.
type OrganizationsResponse struct {
	Meta          Meta                  `json:"meta"`
	Organizations []schema.Organization `json:"organizations"`
}

func (s *server) handleOrganizations(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, OrganizationsResponse{Meta: st.meta(), Organizations: nonNil(st.snapshot.Organizations)})
}

// ActionsResponse is the body of GET /v1/actions.
type ActionsResponse struct {
	Meta    Meta            `json:"meta"`
	Actions []schema.Action `json:"actions"`
}

func (s *server) handleActions(w http.ResponseWriter, r *http.Request, st *state) {
	s.ok(w, r, st, ActionsResponse{Meta: st.meta(), Actions: nonNil(st.snapshot.Actions)})
}

// HealthResponse is the body of /healthz and /readyz.
type HealthResponse struct {
	Status    string `json:"status"`
	ReleaseID string `json:"release_id,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

func (s *server) handleHealthz(w http.ResponseWriter, _ *http.Request, _ *state) {
	s.writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (s *server) handleReadyz(w http.ResponseWriter, _ *http.Request, _ *state) {
	if _, err := s.loader.current(); err != nil {
		s.writeJSON(w, http.StatusServiceUnavailable, HealthResponse{Status: "not_ready", Detail: "no promoted release is loaded"})
		return
	}
	id, _ := s.loader.status()
	s.writeJSON(w, http.StatusOK, HealthResponse{Status: "ready", ReleaseID: id})
}
