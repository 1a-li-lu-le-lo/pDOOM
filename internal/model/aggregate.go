// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// AggregateModelVersion is the version string of the external aggregation model.
const AggregateModelVersion = "pdoom-model/external-aggregate@0.1.0"

// maxSingleWeight caps any one forecast's share in a weighted method
// (docs/method/aggregation.md, "single-figure cap").
const maxSingleWeight = 0.35

// recencyHalfLifeYears is the half-life of the recency-weighted method.
const recencyHalfLifeYears = 3.0

// Group is a compatibility group of forecasts that share outcome set, horizon
// and conditioning.
type Group struct {
	ID           string
	OutcomeSet   []schema.Outcome
	Horizon      string
	Conditioning string
	Forecasts    []schema.Forecast
	Values       []float64
	Preferred    float64
	IQRLogOdds   float64
	Results      []schema.AggregationResult
}

// SourceIDs lists the distinct source ids of the group's forecasts.
func (g Group) SourceIDs() []string {
	set := map[string]bool{}
	for _, f := range g.Forecasts {
		set[f.SourceID] = true
	}
	return sortedKeys(set)
}

// ForecastIDs lists the group's forecast ids in order.
func (g Group) ForecastIDs() []string {
	out := make([]string, len(g.Forecasts))
	for i, f := range g.Forecasts {
		out[i] = f.ID
	}
	return out
}

// PopulationCount counts distinct populations in the group.
func (g Group) PopulationCount() int {
	set := map[schema.ForecastPopulation]bool{}
	for _, f := range g.Forecasts {
		set[f.Population] = true
	}
	return len(set)
}

// Aggregate builds compatibility groups from the snapshot's eligible forecasts
// and computes every method listed in the model spec. Groups need at least two
// members; singleton and ungrouped forecasts are listed individually on the
// forecasts page but never aggregated.
func Aggregate(s *snapshot.Snapshot) []Group {
	byGroup := map[string][]schema.Forecast{}
	for _, f := range s.Forecasts {
		if !f.ModelUseStatus.FeedsModel() || !f.Verification.Status.AllowsModelUse() || f.Status != "current" {
			continue
		}
		if f.GroupID == nil || *f.GroupID == "" {
			continue
		}
		if _, ok := f.CentralValue(); !ok {
			continue
		}
		byGroup[*f.GroupID] = append(byGroup[*f.GroupID], f)
	}
	methods := s.ModelSpec.AggregationMethods
	if len(methods) == 0 {
		methods = []schema.AggregationMethod{schema.MethodUnweightedMedian}
	}
	var groups []Group
	for _, gid := range sortedKeys(byGroup) {
		fs := byGroup[gid]
		if len(fs) < 2 {
			continue
		}
		sort.Slice(fs, func(i, j int) bool { return fs[i].ID < fs[j].ID })
		g := Group{ID: gid, OutcomeSet: fs[0].OutcomeSet, Horizon: string(fs[0].Horizon), Conditioning: fs[0].Conditions, Forecasts: fs}
		for _, f := range fs {
			v, _ := f.CentralValue()
			g.Values = append(g.Values, v)
		}
		g.IQRLogOdds = IQRLogOdds(g.Values)
		for _, m := range methods {
			r := aggregateMethod(s, g, m)
			if m == schema.MethodUnweightedMedian {
				r.Preferred = true
				g.Preferred = r.Value
			}
			g.Results = append(g.Results, r)
		}
		if g.Preferred == 0 && len(g.Results) > 0 {
			g.Results[0].Preferred = true
			g.Preferred = g.Results[0].Value
		}
		groups = append(groups, g)
	}
	return groups
}

func aggregateMethod(s *snapshot.Snapshot, g Group, m schema.AggregationMethod) schema.AggregationResult {
	r := schema.AggregationResult{
		AggregationID:      fmt.Sprintf("agg-%s-%s", strings.ToLower(g.ID), m),
		GroupID:            g.ID,
		Method:             m,
		OutcomeSet:         g.OutcomeSet,
		Horizon:            g.Horizon,
		Conditioning:       g.Conditioning,
		N:                  len(g.Forecasts),
		ForecastIDs:        g.ForecastIDs(),
		SourceIDs:          g.SourceIDs(),
		PopulationCount:    g.PopulationCount(),
		Weights:            map[string]float64{},
		WordingNote:        wordingNote(g),
		TransformationNote: transformationNote(g),
	}
	n := len(g.Values)
	equal := make([]float64, n)
	for i := range equal {
		equal[i] = 1 / float64(n)
	}
	var weights []float64
	switch m {
	case schema.MethodUnweightedMedian:
		r.Value = Median(g.Values)
		weights = equal
		r.Note = "Median of member central values; the preferred method because it is robust to a single extreme forecast."
	case schema.MethodLinearPool:
		r.Value = Mean(g.Values)
		weights = equal
		r.Note = "Arithmetic mean of member central values."
	case schema.MethodLogOddsPool:
		lo := make([]float64, n)
		for i, v := range g.Values {
			lo[i] = Logit(v)
		}
		r.Value = Sigmoid(Mean(lo))
		weights = equal
		r.Note = "Mean on the log-odds scale, mapped back to a probability."
	case schema.MethodTrimmedMean:
		r.Value = TrimmedMean(g.Values, 0.2)
		weights = equal
		r.Note = "Mean after trimming 20% from each end; falls back to the median when fewer than five members."
	case schema.MethodTierWeighted:
		weights = capWeights(tierWeights(s, g))
		r.Value = WeightedMean(g.Values, weights)
		r.Note = "Weighted by source tier multiplier (tier 1 = 1.0, tier 2 = 0.9, tier 3 = 0.5); no member above the single-figure cap."
	case schema.MethodRecencyWeighted:
		weights = capWeights(recencyWeights(g))
		r.Value = WeightedMean(g.Values, weights)
		r.Note = "Weighted by 0.5^(age in years / 3) relative to the newest member; no member above the single-figure cap."
	case schema.MethodEqualWeightByPopulation:
		r.Value = populationMedianMean(g)
		weights = populationWeights(g)
		r.Note = "Mean of per-population medians so that a large survey and a small superforecaster panel count equally."
	default:
		r.Value = Median(g.Values)
		weights = equal
		r.Note = "Unknown method; median used."
	}
	r.Value = round6(clamp01(r.Value))
	for i, f := range g.Forecasts {
		r.Weights[f.ID] = round4(weights[i])
	}
	return r
}

func tierWeights(s *snapshot.Snapshot, g Group) []float64 {
	w := make([]float64, len(g.Forecasts))
	for i, f := range g.Forecasts {
		tier := s.BestTier([]string{f.SourceID})
		w[i] = s.ModelSpec.TierMultiplier(tier)
	}
	return normalize(w)
}

func recencyWeights(g Group) []float64 {
	var newest time.Time
	dates := make([]time.Time, len(g.Forecasts))
	for i, f := range g.Forecasts {
		d, _ := time.Parse("2006-01-02", f.Date)
		dates[i] = d
		if d.After(newest) {
			newest = d
		}
	}
	w := make([]float64, len(g.Forecasts))
	for i, d := range dates {
		years := newest.Sub(d).Hours() / 24 / 365.25
		w[i] = math.Pow(0.5, years/recencyHalfLifeYears)
	}
	return normalize(w)
}

func populationWeights(g Group) []float64 {
	counts := map[schema.ForecastPopulation]int{}
	for _, f := range g.Forecasts {
		counts[f.Population]++
	}
	w := make([]float64, len(g.Forecasts))
	for i, f := range g.Forecasts {
		w[i] = 1 / float64(len(counts)) / float64(counts[f.Population])
	}
	return w
}

func populationMedianMean(g Group) float64 {
	byPop := map[schema.ForecastPopulation][]float64{}
	for i, f := range g.Forecasts {
		byPop[f.Population] = append(byPop[f.Population], g.Values[i])
	}
	keys := make([]string, 0, len(byPop))
	for k := range byPop {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	meds := make([]float64, 0, len(keys))
	for _, k := range keys {
		meds = append(meds, Median(byPop[schema.ForecastPopulation(k)]))
	}
	return Mean(meds)
}

// capWeights enforces the single-figure cap by iteratively clipping weights at
// maxSingleWeight and redistributing the excess proportionally.
func capWeights(w []float64) []float64 {
	w = normalize(w)
	if len(w) <= 2 {
		return w // with two members the cap cannot be honoured; both get ≥ 0.5 by construction
	}
	for iter := 0; iter < 10; iter++ {
		excess := 0.0
		free := 0.0
		for _, x := range w {
			if x > maxSingleWeight {
				excess += x - maxSingleWeight
			} else {
				free += x
			}
		}
		if excess == 0 || free == 0 {
			break
		}
		for i, x := range w {
			if x > maxSingleWeight {
				w[i] = maxSingleWeight
			} else {
				w[i] = x + excess*(x/free)
			}
		}
	}
	return w
}

func normalize(w []float64) []float64 {
	sum := 0.0
	for _, x := range w {
		sum += x
	}
	out := make([]float64, len(w))
	if sum == 0 {
		for i := range out {
			out[i] = 1 / float64(len(w))
		}
		return out
	}
	for i, x := range w {
		out[i] = x / sum
	}
	return out
}

func wordingNote(g Group) string {
	parts := make([]string, 0, len(g.Forecasts))
	for _, f := range g.Forecasts {
		q := f.QuestionWordingOriginal
		if f.Paraphrase {
			q = "(paraphrase) " + q
		}
		parts = append(parts, fmt.Sprintf("%s [%s, %s]: %q", f.ID, f.ForecasterOrSurvey, f.Date, q))
	}
	return strings.Join(parts, " | ")
}

func transformationNote(g Group) string {
	var parts []string
	for _, f := range g.Forecasts {
		if f.TransformationNote != nil && strings.TrimSpace(*f.TransformationNote) != "" {
			parts = append(parts, fmt.Sprintf("%s: %s", f.ID, *f.TransformationNote))
		}
	}
	if len(parts) == 0 {
		return "No transformation applied; members answer the same question in the same units."
	}
	return strings.Join(parts, " | ")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
