// Copyright NU Cybernetics. p(DOOM) — research prototype.

package snapshot_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot/fixture"
)

func schemaDir(t *testing.T) string {
	t.Helper()
	dir, _ := filepath.Abs(filepath.Join("..", "..", "data", "schemas"))
	if _, err := os.Stat(dir); err != nil {
		t.Skip("data/schemas not built")
	}
	return dir
}

func TestWriteSealLoadValidateRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "snap-1999-01-01-001")
	if err := snapshot.Write(dir, fixture.New()); err != nil {
		t.Fatal(err)
	}
	s, err := snapshot.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Manifest.Files) == 0 || s.Manifest.Files[0].SHA256 == "" {
		t.Fatal("manifest not sealed")
	}
	problems := snapshot.Validate(s, schemaDir(t))
	for _, p := range problems {
		if p.Severity == snapshot.SeverityError {
			t.Errorf("unexpected error: %s", p)
		}
	}
}

func TestLoadRejectsTamperedFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "snap-1999-01-01-001")
	if err := snapshot.Write(dir, fixture.New()); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "forecasts.json")
	raw, _ := os.ReadFile(p)
	if err := os.WriteFile(p, []byte(strings.Replace(string(raw), "0.05", "0.5", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot.Load(dir); err == nil || !strings.Contains(err.Error(), "sha256 mismatch") {
		t.Fatalf("expected sha256 mismatch, got %v", err)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "snap-1999-01-01-001")
	if err := snapshot.Write(dir, fixture.New()); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "definitions.json")
	raw, _ := os.ReadFile(p)
	out := strings.Replace(string(raw), `"term":`, `"unexpected_field": 1, "term":`, 1)
	if err := os.WriteFile(p, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot.Seal(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot.Load(dir); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestValidateCatchesRuleViolations(t *testing.T) {
	s := fixture.New()
	// Unverified but eligible → error. Tier 4 eligible → error. Mixed group → error.
	s.Forecasts[0].Verification.Status = "unverified"
	s.Sources[5].ModelUseStatus = "eligible"
	g := "G-FIX-EXT-2100"
	s.Forecasts[4].GroupID = &g
	problems := snapshot.Validate(s, "")
	want := []string{"unverified items must be excluded", "tier 4 sources cannot be eligible/used", "mixes outcome sets or horizons"}
	for _, w := range want {
		found := false
		for _, p := range problems {
			if strings.Contains(p.Message, w) {
				found = true
			}
		}
		if !found {
			t.Errorf("expected a problem containing %q; got %v", w, problems)
		}
	}
}
