// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// ScenarioLabRequest is the body of POST /v1/scenario-lab/evaluate: the
// UserScenarioParams (horizon and all ten sliders, each required) plus an
// optional sample count and seed. Pointers distinguish absent fields from
// zero so that a missing slider is reported rather than silently treated as
// baseline.
type ScenarioLabRequest struct {
	Horizon                   *string `json:"horizon"`
	CapabilityTimeline        *int    `json:"capability_timeline"`
	AutonomyGrowth            *int    `json:"autonomy_growth"`
	AccessLevel               *int    `json:"access_level"`
	SafetyProgress            *int    `json:"safety_progress"`
	GovernanceStrength        *int    `json:"governance_strength"`
	ModelSecurity             *int    `json:"model_security"`
	OpenWeightDiffusion       *int    `json:"open_weight_diffusion"`
	InternationalCoordination *int    `json:"international_coordination"`
	IncidentFrequency         *int    `json:"incident_frequency"`
	Resilience                *int    `json:"resilience"`
	Samples                   *int    `json:"samples"`
	Seed                      *int64  `json:"seed"`
}

// params converts the request into UserScenarioParams and lists the required
// fields that are absent.
func (r ScenarioLabRequest) params() (schema.UserScenarioParams, []string) {
	var missing []string
	get := func(name string, p *int) int {
		if p == nil {
			missing = append(missing, name)
			return 0
		}
		return *p
	}
	var p schema.UserScenarioParams
	if r.Horizon == nil {
		missing = append(missing, "horizon")
	} else {
		p.Horizon = *r.Horizon
	}
	p.CapabilityTimeline = get("capability_timeline", r.CapabilityTimeline)
	p.AutonomyGrowth = get("autonomy_growth", r.AutonomyGrowth)
	p.AccessLevel = get("access_level", r.AccessLevel)
	p.SafetyProgress = get("safety_progress", r.SafetyProgress)
	p.GovernanceStrength = get("governance_strength", r.GovernanceStrength)
	p.ModelSecurity = get("model_security", r.ModelSecurity)
	p.OpenWeightDiffusion = get("open_weight_diffusion", r.OpenWeightDiffusion)
	p.InternationalCoordination = get("international_coordination", r.InternationalCoordination)
	p.IncidentFrequency = get("incident_frequency", r.IncidentFrequency)
	p.Resilience = get("resilience", r.Resilience)
	return p, missing
}

// ScenarioLabResponse is labelled user_scenario: it is computed on request
// from the user's sliders with the current release's experimental model
// specification and is never an official estimate.
type ScenarioLabResponse struct {
	Label        string                      `json:"label"`
	Status       schema.EstimateStatus       `json:"status"`
	Disclaimer   string                      `json:"disclaimer"`
	Horizon      string                      `json:"horizon"`
	OutcomeSet   []schema.Outcome            `json:"outcome_set"`
	OutcomeSets  map[string][]schema.Outcome `json:"outcome_sets"`
	Producer     string                      `json:"producer"`
	ModelVersion string                      `json:"model_version"`
	ReleaseID    string                      `json:"release_id"`
	DataSnapshot string                      `json:"data_snapshot"`
	DataCutoff   string                      `json:"data_cutoff"`
	Samples      int                         `json:"samples"`
	Seed         int64                       `json:"seed"`
	Result       schema.UserScenarioResult   `json:"result"`
	Limitations  []string                    `json:"limitations"`
}

var labLimitations = []string{
	"This is a user-generated scenario, not a published estimate; it does not change any official value.",
	"Slider positions shift the logit-scale location of judgment-based parameters; the mapping is documented, not calibrated.",
	"Intervals express parameter uncertainty under the experimental model, not forecast error.",
	"Dependence between factors is a single common latent factor; feedback loops and competing risks are not modelled.",
}

// outcomeSetsByKey maps the simulation output keys to their outcome sets.
var outcomeSetsByKey = map[string][]schema.Outcome{
	"O3": {schema.O3}, "O4": {schema.O4}, "O5": {schema.O5}, "O6": {schema.O6}, "O7": {schema.O7}, "O8": {schema.O8},
	"P_DOOM": schema.PDoomOutcomes, "P_COLLAPSE": schema.CollapseOutcomes,
}

func (s *server) handleScenarioLab(w http.ResponseWriter, r *http.Request, st *state) {
	var req ScenarioLabRequest
	if !s.decodeBody(w, r, &req) {
		return
	}
	spec := st.snapshot.ModelSpec.ExperimentalCausal
	params, missing := req.params()
	if len(missing) > 0 {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "missing required fields: "+strings.Join(missing, ", "))
		return
	}
	if problems := validateLab(params, req.Samples, req.Seed, spec); len(problems) > 0 {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", strings.Join(problems, "; "))
		return
	}
	samples := spec.Samples
	if req.Samples != nil && *req.Samples > 0 && *req.Samples <= spec.Samples {
		samples = *req.Samples
	}
	seed := spec.Seed
	if req.Seed != nil && *req.Seed != 0 {
		seed = *req.Seed
	}
	res, err := model.EvaluateUserScenario(spec, params, samples, seed)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	producer := spec.Version
	if producer == "" {
		producer = model.CausalModelVersion
	}
	m := st.release.Manifest
	sets := map[string][]schema.Outcome{}
	for k := range res.OutcomeEstimates {
		if set, ok := outcomeSetsByKey[k]; ok {
			sets[k] = set
		}
	}
	resp := ScenarioLabResponse{
		Label: res.Label, Status: schema.StatusUserScenario, Disclaimer: res.Disclaimer, Horizon: res.Horizon,
		OutcomeSet: schema.PDoomOutcomes, OutcomeSets: sets, Producer: producer, ModelVersion: producer,
		ReleaseID: m.ReleaseID, DataSnapshot: m.DataSnapshot, DataCutoff: m.SourceCutoff,
		Samples: samples, Seed: seed, Result: res, Limitations: labLimitations,
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// validateLab checks horizon, sliders, samples and seed, reporting every
// problem so that a client can fix them in one round trip.
func validateLab(p schema.UserScenarioParams, samplesOpt *int, seedOpt *int64, spec schema.ExperimentalCausalSpec) []string {
	var problems []string
	if p.Horizon == "" {
		problems = append(problems, "horizon is required")
	} else if !schema.Horizon(p.Horizon).Valid() {
		problems = append(problems, "horizon "+strconv.Quote(p.Horizon)+" is not one of "+joinHorizons())
	} else if _, ok := spec.Horizons[p.Horizon]; !ok {
		problems = append(problems, "horizon "+strconv.Quote(p.Horizon)+" is not covered by the experimental model; covered: "+joinKeys(spec.Horizons))
	}
	sliders := p.Sliders()
	names := make([]string, 0, len(sliders))
	for name := range sliders {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if v := sliders[name]; v < -2 || v > 2 {
			problems = append(problems, fmt.Sprintf("slider %s must be an integer in -2..2, got %d", name, v))
		}
	}
	if samplesOpt != nil && (*samplesOpt < 1 || *samplesOpt > MaxSamples) {
		problems = append(problems, fmt.Sprintf("samples must be in 1..%d", MaxSamples))
	}
	if seedOpt != nil && (*seedOpt < 0 || *seedOpt > MaxSeed) {
		problems = append(problems, fmt.Sprintf("seed must be in 0..%d", MaxSeed))
	}
	return problems
}

func joinHorizons() string {
	var out []string
	for _, h := range schema.Horizons {
		out = append(out, string(h))
	}
	return strings.Join(out, ", ")
}

func joinKeys(m map[string]schema.HorizonCausalSpec) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return schema.HorizonIndex(schema.Horizon(keys[i])) < schema.HorizonIndex(schema.Horizon(keys[j]))
	})
	return strings.Join(keys, ", ")
}

// decodeBody reads a JSON body of at most MaxBodyBytes into v with unknown
// fields rejected. It writes the error response and returns false on failure.
func (s *server) decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if ct := r.Header.Get("Content-Type"); ct != "" {
		mt, _, err := mime.ParseMediaType(ct)
		if err != nil || mt != "application/json" {
			s.writeError(w, r, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
			return false
		}
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			s.writeError(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "body exceeds "+strconv.Itoa(MaxBodyBytes)+" bytes")
			return false
		}
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "body could not be read")
		return false
	}
	if len(bytes.TrimSpace(body)) == 0 {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "empty body")
		return false
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "invalid JSON body: "+jsonErrorDetail(err))
		return false
	}
	if _, err := dec.Token(); err != io.EOF {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "invalid JSON body: trailing data")
		return false
	}
	return true
}

// jsonErrorDetail turns encoding/json errors into short, non-leaking messages.
func jsonErrorDetail(err error) string {
	var syn *json.SyntaxError
	var typ *json.UnmarshalTypeError
	switch {
	case errors.As(err, &syn):
		return "syntax error at offset " + strconv.FormatInt(syn.Offset, 10)
	case errors.As(err, &typ):
		field := typ.Field
		if field == "" {
			field = "body"
		}
		return "field " + field + " has the wrong type (expected " + typ.Type.String() + ")"
	}
	msg := err.Error()
	if strings.HasPrefix(msg, "json: unknown field") {
		return strings.TrimPrefix(msg, "json: ")
	}
	return "malformed JSON"
}
