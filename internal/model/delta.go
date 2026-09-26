// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"
	"math"
	"sort"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// headlineTriggerPoints is the absolute change (in percentage points) of any
// research-mode P_DOOM median that requires heightened review.
const headlineTriggerPoints = 2.0

// Delta compares a new result with the previous release and records the
// PROBABILITY CHANGE POLICY fields and heightened-review triggers.
func Delta(s *snapshot.Snapshot, res *Result, prev *schema.Release, generatedAt string) schema.DeltaRecord {
	d := schema.DeltaRecord{Kind: "delta_record", SchemaVersion: 1, SnapshotID: s.Manifest.SnapshotID, GeneratedAt: generatedAt,
		EstimateChanges: []schema.EstimateChange{}, IndexChanges: []schema.IndexChange{}, HeightenedReviewTriggers: []string{},
		NewScenarioIDs: []string{}, NewForecastSourceIDs: []string{}, RedefinedBenchmarkIDs: []string{}, Notes: []string{}}
	triggers := map[string]bool{}
	add := func(t string) { triggers[t] = true }

	// Data-driven triggers independent of a previous release.
	for _, in := range s.Incidents {
		if !in.ModelUseStatus.FeedsModel() {
			continue
		}
		for _, c := range in.Cause {
			if c == "security_compromise" {
				add("security_incident_involved")
			}
		}
	}
	if singleLab(s) {
		add("single_laboratory_evidence")
	}
	if prev == nil {
		d.Notes = append(d.Notes, "First release: no previous estimate to compare against.")
		add("first_release")
		d.HeightenedReviewTriggers = sortedKeys(triggers)
		for _, e := range res.Estimates {
			d.EstimateChanges = append(d.EstimateChanges, schema.EstimateChange{EstimateID: e.EstimateID, Status: string(e.Status), NewP50: p50(e), NewDisplay: e.Display.Central})
		}
		for _, iv := range res.Indexes {
			d.IndexChanges = append(d.IndexChanges, schema.IndexChange{IndexID: iv.IndexID, New: iv.Value})
		}
		return d
	}

	pid := prev.Manifest.ReleaseID
	d.PreviousReleaseID = &pid
	psnap := prev.Manifest.DataSnapshot
	d.PreviousSnapshotID = &psnap
	for _, e := range res.Estimates {
		pe := prev.EstimateByID(e.EstimateID)
		ch := schema.EstimateChange{EstimateID: e.EstimateID, Status: string(e.Status), NewP50: p50(e), NewDisplay: e.Display.Central}
		if pe != nil {
			ch.PreviousP50 = p50(*pe)
			ch.PreviousDisplay = pe.Display.Central
			if ch.PreviousP50 != nil && ch.NewP50 != nil {
				pts := round4((*ch.NewP50 - *ch.PreviousP50) * 100)
				ch.ChangePoints = &pts
				if e.Status == schema.StatusResearchMode && schema.OutcomeSetKey(e.OutcomeSet) == schema.OutcomeSetKey(schema.PDoomOutcomes) && math.Abs(pts) >= headlineTriggerPoints {
					add("headline_pdoom_material_change")
				}
				if schema.OutcomeSetKey(e.OutcomeSet) == schema.OutcomeSetKey(schema.ExtinctionOutcomes) && pts != 0 {
					add("extinction_estimate_changed")
				}
			}
		} else {
			d.Notes = append(d.Notes, "new estimate "+e.EstimateID)
		}
		d.EstimateChanges = append(d.EstimateChanges, ch)
	}
	for _, iv := range res.Indexes {
		ch := schema.IndexChange{IndexID: iv.IndexID, New: iv.Value}
		if piv := prev.IndexByID(iv.IndexID); piv != nil {
			ch.Previous = piv.Value
			if piv.Value != nil && iv.Value != nil {
				dv := round4(*iv.Value - *piv.Value)
				ch.Delta = &dv
			}
		}
		d.IndexChanges = append(d.IndexChanges, ch)
	}
	prevScen := set(prev.Manifest.DataSummary.ScenarioIDs)
	for _, sc := range s.Scenarios {
		if !prevScen[sc.ID] {
			d.NewScenarioIDs = append(d.NewScenarioIDs, sc.ID)
		}
	}
	if len(d.NewScenarioIDs) > 0 {
		add("new_scenario_introduced")
	}
	prevFS := set(prev.Manifest.DataSummary.ForecastSourceIDs)
	for _, id := range res.DataSummary.ForecastSourceIDs {
		if !prevFS[id] {
			d.NewForecastSourceIDs = append(d.NewForecastSourceIDs, id)
		}
	}
	if len(d.NewForecastSourceIDs) > 0 {
		add("expert_survey_or_forecast_source_added")
	}
	for id, def := range res.DataSummary.BenchmarkDefinitions {
		if pdef, ok := prev.Manifest.DataSummary.BenchmarkDefinitions[id]; ok && pdef != def {
			d.RedefinedBenchmarkIDs = append(d.RedefinedBenchmarkIDs, id)
		}
	}
	sort.Strings(d.RedefinedBenchmarkIDs)
	if len(d.RedefinedBenchmarkIDs) > 0 {
		add("benchmark_redefined")
	}
	if !sameStrings(res.ModelVersions, prev.Manifest.ModelVersions) {
		d.ModelVersionChanged = true
		add("model_version_changed")
	}
	d.HeightenedReviewTriggers = sortedKeys(triggers)
	return d
}

func p50(e schema.Estimate) *float64 {
	if e.Quantiles == nil {
		return nil
	}
	v := e.Quantiles.P50
	return &v
}

func set(xs []string) map[string]bool {
	m := map[string]bool{}
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as, bs := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(as)
	sort.Strings(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

// singleLab reports whether any eligible observation rests solely on
// developer self-reports from a single publisher.
func singleLab(s *snapshot.Snapshot) bool {
	for _, o := range s.DriverObservations {
		if !o.ModelUseStatus.FeedsModel() || len(o.SourceIDs) == 0 {
			continue
		}
		publishers := map[string]bool{}
		allSelf := true
		for _, sid := range o.SourceIDs {
			src := s.SourceByID(sid)
			if src == nil {
				allSelf = false
				break
			}
			self := false
			for _, c := range src.Conflicts {
				if c == "developer_self_report" {
					self = true
				}
			}
			if !self {
				allSelf = false
				break
			}
			publishers[src.Publisher] = true
		}
		if allSelf && len(publishers) == 1 {
			return true
		}
	}
	return false
}

// benchmarkDefinition hashes the identity of a benchmark so that later
// releases can detect redefinitions.
func benchmarkDefinition(b schema.Benchmark) string {
	v := ""
	if b.Version != nil {
		v = *b.Version
	}
	return schema.SHA256Hex([]byte(fmt.Sprintf("%s|%s|%s|%s|%s", b.Name, v, b.URL, b.Unit, b.Direction)))[:16]
}
