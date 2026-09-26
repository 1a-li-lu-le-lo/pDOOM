// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot/fixture"
)

var update = flag.Bool("update", false, "rewrite golden files")

const generatedAt = "2026-09-26T12:00:00Z"

func runFixture(t *testing.T) *Result {
	t.Helper()
	res, err := Run(fixture.New(), RunOptions{GeneratedAt: generatedAt, CodeCommit: "0000000"})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestRunIsDeterministic(t *testing.T) {
	a, b := runFixture(t), runFixture(t)
	ja, _ := schema.CanonicalJSON(a)
	jb, _ := schema.CanonicalJSON(b)
	if !bytes.Equal(ja, jb) {
		t.Fatal("two runs on the same snapshot produced different canonical JSON")
	}
}

func TestRunInvariantsAndShape(t *testing.T) {
	res := runFixture(t)
	if errs := CheckInvariants(res); len(errs) > 0 {
		t.Fatalf("invariants violated: %v", errs)
	}
	var official, external, research int
	for _, e := range res.Estimates {
		switch e.Status {
		case schema.StatusInsufficientlyCalibrate:
			official++
			if e.Quantiles != nil || e.Display.Central != "Insufficiently calibrated" {
				t.Fatalf("official object leaks a probability: %+v", e)
			}
		case schema.StatusExternalAggregate:
			external++
			if e.GroupID == nil || *e.GroupID != "G-FIX-EXT-2100" || e.SourceCoverage.ForecastCount != 4 {
				t.Fatalf("unexpected external estimate %+v", e)
			}
			if e.Quantiles.P50 != 0.04 { // median of 0.004, 0.03, 0.05, 0.15
				t.Fatalf("median should be 0.04, got %v", e.Quantiles.P50)
			}
		case schema.StatusResearchMode:
			research++
		}
	}
	if official != 7 || external != 1 || research != 15 {
		t.Fatalf("counts official=%d external=%d research=%d", official, external, research)
	}
	if len(res.Aggregations) != 7 {
		t.Fatalf("expected 7 aggregation methods, got %d", len(res.Aggregations))
	}
	byID := map[schema.IndexID]schema.IndexValue{}
	for _, iv := range res.Indexes {
		byID[iv.IndexID] = iv
	}
	cpi := byID[schema.IndexCapabilityPressure]
	if cpi.Value == nil || cpi.Coverage != 1 {
		t.Fatalf("cpi: %+v", cpi)
	}
	// 0.35·1·0.62 + 0.35·1·0.70 + 0.30·1·0.50 over weights 1.0 → 61.2
	if *cpi.Value < 61.1 || *cpi.Value > 61.3 {
		t.Fatalf("cpi value %v", *cpi.Value)
	}
	csi := byID[schema.IndexControlStrength]
	if csi.Value == nil || csi.Coverage != 0.7 {
		t.Fatalf("csi should have coverage 0.7 (D7 missing): %+v", csi)
	}
	epi := byID[schema.IndexEvidencePressure]
	if epi.Value == nil || *epi.Value != 50 {
		t.Fatalf("baseline evidence pressure must be 50: %+v", epi)
	}
	if byID[schema.IndexAttention].Value != nil {
		t.Fatal("attention index must be null without media data")
	}
	if byID[schema.IndexUncertainty].Value == nil {
		t.Fatal("uncertainty score missing")
	}
	if !res.EditorialRiskLevel.Valid() {
		t.Fatalf("editorial level %q", res.EditorialRiskLevel)
	}
	if len(res.Sensitivity) == 0 || res.Sensitivity[0].Rank == 0 {
		t.Fatal("sensitivity runs missing or unranked")
	}
	if len(res.Delta.HeightenedReviewTriggers) == 0 || res.Delta.PreviousReleaseID != nil {
		t.Fatalf("first-release delta wrong: %+v", res.Delta)
	}
	if len(res.DriversExplained.Items) == 0 {
		t.Fatal("no driver explanations")
	}
}

func TestRunHandlesMissingData(t *testing.T) {
	s := fixture.New()
	s.Forecasts = nil
	s.DriverObservations = nil
	s.Incidents = nil
	res, err := Run(s, RunOptions{GeneratedAt: generatedAt})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Aggregations) != 0 {
		t.Fatal("no forecasts should mean no aggregations")
	}
	for _, iv := range res.Indexes {
		if (iv.IndexID == schema.IndexCapabilityPressure || iv.IndexID == schema.IndexIncidentPressure) && iv.Value != nil {
			t.Fatalf("%s should be null without data", iv.IndexID)
		}
	}
	if errs := CheckInvariants(res); len(errs) > 0 {
		t.Fatalf("invariants: %v", errs)
	}
	if res.EditorialRiskLevel != "insufficient_evidence" {
		t.Fatalf("expected insufficient_evidence, got %s", res.EditorialRiskLevel)
	}
}

func TestRunRequiresGeneratedAt(t *testing.T) {
	if _, err := Run(fixture.New(), RunOptions{}); err == nil {
		t.Fatal("expected error without GeneratedAt")
	}
}

func TestRoundForDisplay(t *testing.T) {
	cases := []struct {
		p    float64
		step int
		want string
	}{{0.1273, 5, "15%"}, {0.1273, 2, "12%"}, {0.1273, 1, "13%"}, {0.004, 1, "<1%"}, {0.996, 1, ">99%"}, {0, 5, "0%"}, {1, 1, "100%"}, {0.03, 5, "5%"}, {0.02, 5, "<1%"}}
	for _, c := range cases {
		if got := RoundForDisplay(c.p, c.step); got != c.want {
			t.Errorf("RoundForDisplay(%v,%d)=%s want %s", c.p, c.step, got, c.want)
		}
	}
	if got := FormatInterval(0.03, 0.3, 5); got != "5%–30%" {
		t.Errorf("interval %s", got)
	}
}

func TestGolden(t *testing.T) {
	res := runFixture(t)
	got, err := schema.CanonicalJSONIndent(res)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("testdata", "golden", "snap-1999-01-01-001.json")
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden file missing; run go test ./internal/model -update: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("golden mismatch; if the change is intended run go test ./internal/model -update")
	}
}
