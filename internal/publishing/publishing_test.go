// Copyright NU Cybernetics. p(DOOM) — research prototype.

package publishing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1a-li-lu-le-lo/pdoom/internal/audit"
	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot/fixture"
)

const (
	genAt = "2026-09-26T12:00:00Z"
	pubAt = "2026-09-26T13:00:00Z"
)

type env struct {
	paths     config.Paths
	snap      *snapshot.Snapshot
	candidate string
	keyA      string
	keyB      string
}

func setup(t *testing.T) *env {
	t.Helper()
	root := t.TempDir()
	paths := config.Resolve(filepath.Join(root, "data"), "")
	fx := fixture.New()
	snapDir := filepath.Join(paths.Snapshots, fx.Manifest.SnapshotID)
	if err := snapshot.Write(snapDir, fx); err != nil {
		t.Fatal(err)
	}
	snap, err := snapshot.Load(snapDir)
	if err != nil {
		t.Fatal(err)
	}
	res, err := model.Run(snap, model.RunOptions{GeneratedAt: genAt, CodeCommit: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	cid := "cand-2026-09-26-001"
	m, err := BuildManifest(res, snap, BuildOptions{ReleaseID: "rel-2026-09-26-001", CandidateID: cid, CodeCommit: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	cdir := filepath.Join(paths.Candidates, cid)
	if _, err := WriteCandidate(cdir, res, m); err != nil {
		t.Fatal(err)
	}
	_, keyA, err := GenerateKey(paths.Keys, "reviewer-a")
	if err != nil {
		t.Fatal(err)
	}
	_, keyB, err := GenerateKey(paths.Keys, "reviewer-b")
	if err != nil {
		t.Fatal(err)
	}
	return &env{paths: paths, snap: snap, candidate: cdir, keyA: keyA, keyB: keyB}
}

func TestPromoteHappyPathAndRollback(t *testing.T) {
	e := setup(t)
	// No approvals → gate fails.
	if _, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt}); err == nil || !strings.Contains(err.Error(), "approvals") {
		t.Fatalf("expected approvals gate, got %v", err)
	}
	if _, err := Approve(e.candidate, "reviewer-a", e.keyA, "", pubAt); err != nil {
		t.Fatal(err)
	}
	// First release carries heightened triggers → needs a second reviewer + ack.
	if _, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt}); err == nil || !strings.Contains(err.Error(), "heightened_review") {
		t.Fatalf("expected heightened review gate, got %v", err)
	}
	if _, err := Approve(e.candidate, "reviewer-b", e.keyB, "none", pubAt); err != nil {
		t.Fatal(err)
	}
	if _, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt, HeightenedReviewAck: "reviewer-zz"}); err == nil {
		t.Fatal("ack by a non-approver must fail")
	}
	rel, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt, HeightenedReviewAck: "reviewer-b", Actor: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if rel.Manifest.Published == nil || rel.Manifest.Approval == nil || rel.Manifest.Approval.ReceivedApprovals != 2 {
		t.Fatalf("published manifest incomplete: %+v", rel.Manifest.Approval)
	}
	cur, err := CurrentReleaseID(e.paths)
	if err != nil || cur != "rel-2026-09-26-001" {
		t.Fatalf("CURRENT = %q err %v", cur, err)
	}
	if _, err := LoadCurrentRelease(e.paths); err != nil {
		t.Fatal(err)
	}
	if n, err := audit.Verify(e.paths.Audit); err != nil || n != 1 {
		t.Fatalf("audit: n=%d err=%v", n, err)
	}
	// Promoting again is refused (immutable).
	if _, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt, HeightenedReviewAck: "reviewer-b"}); err == nil {
		t.Fatal("re-promotion must fail")
	}
	// Rollback to a non-existent release fails; to the current one fails.
	if err := Rollback(e.paths, "rel-2026-09-26-009", pubAt, "test", "x"); err == nil {
		t.Fatal("rollback to missing release must fail")
	}
	if err := Rollback(e.paths, "rel-2026-09-26-001", pubAt, "test", "x"); err == nil {
		t.Fatal("rollback to current must fail")
	}
	list, err := ListReleases(e.paths)
	if err != nil || len(list) != 1 || !list[0].IsCurrent {
		t.Fatalf("list %+v err %v", list, err)
	}
}

func TestPromoteDetectsTampering(t *testing.T) {
	e := setup(t)
	if _, err := Approve(e.candidate, "reviewer-a", e.keyA, "", pubAt); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(e.candidate, "reviewer-b", e.keyB, "", pubAt); err != nil {
		t.Fatal(err)
	}
	// Alter estimates after signing.
	p := filepath.Join(e.candidate, "estimates.json")
	raw, _ := os.ReadFile(p)
	if err := os.WriteFile(p, []byte(strings.Replace(string(raw), `"Insufficiently calibrated"`, `"12%"`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt, HeightenedReviewAck: "reviewer-b"}); err == nil || !strings.Contains(err.Error(), "integrity") {
		t.Fatalf("expected integrity gate, got %v", err)
	}
}

func TestPromoteDetectsManifestEdit(t *testing.T) {
	e := setup(t)
	if _, err := Approve(e.candidate, "reviewer-a", e.keyA, "", pubAt); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(e.candidate, "reviewer-b", e.keyB, "", pubAt); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(e.candidate, "manifest.json")
	raw, _ := os.ReadFile(p)
	if err := os.WriteFile(p, []byte(strings.Replace(string(raw), `"editorial_risk_level": "`, `"editorial_risk_level": "very_low", "x_": "`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt, HeightenedReviewAck: "reviewer-b"}); err == nil {
		t.Fatal("edited manifest must be rejected")
	}
}

func TestPromoteRejectsUnknownKeyAndForeignSignature(t *testing.T) {
	e := setup(t)
	if _, err := Approve(e.candidate, "reviewer-a", e.keyA, "", pubAt); err != nil {
		t.Fatal(err)
	}
	// Remove reviewer-a's public key → signature cannot be verified.
	if err := os.Remove(filepath.Join(e.paths.Keys, "reviewer-a.pub")); err != nil {
		t.Fatal(err)
	}
	if _, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt}); err == nil || !strings.Contains(err.Error(), "approvals") {
		t.Fatalf("expected approvals gate, got %v", err)
	}
}

func TestSecondReleaseDeltaAndEvidencePressure(t *testing.T) {
	e := setup(t)
	for _, k := range []struct{ id, key string }{{"reviewer-a", e.keyA}, {"reviewer-b", e.keyB}} {
		if _, err := Approve(e.candidate, k.id, k.key, "", pubAt); err != nil {
			t.Fatal(err)
		}
	}
	first, err := Promote(e.paths, e.candidate, PromoteOptions{PublishedAt: pubAt, HeightenedReviewAck: "reviewer-b"})
	if err != nil {
		t.Fatal(err)
	}
	// Second snapshot with a stronger capability observation.
	fx := fixture.New()
	fx.Manifest.SnapshotID = "snap-1999-02-01-001"
	fx.DriverObservations[0].ValueNormalized = 0.9
	snapDir := filepath.Join(e.paths.Snapshots, fx.Manifest.SnapshotID)
	if err := snapshot.Write(snapDir, fx); err != nil {
		t.Fatal(err)
	}
	snap, _ := snapshot.Load(snapDir)
	res, err := model.Run(snap, model.RunOptions{PreviousRelease: first, GeneratedAt: "2026-10-01T00:00:00Z", CodeCommit: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	epi := res.Indexes[0]
	if epi.Value == nil || *epi.Value <= 50 {
		t.Fatalf("evidence pressure should rise above 50: %+v", epi)
	}
	if res.Delta.PreviousReleaseID == nil || *res.Delta.PreviousReleaseID != first.Manifest.ReleaseID {
		t.Fatalf("delta lacks previous release: %+v", res.Delta)
	}
	prevID := first.Manifest.ReleaseID
	m, err := BuildManifest(res, snap, BuildOptions{ReleaseID: NextReleaseID("2026-10-01", ExistingIDs(e.paths)), CandidateID: "cand-2026-10-01-001", CodeCommit: "abc1234", PreviousReleaseID: &prevID})
	if err != nil {
		t.Fatal(err)
	}
	cdir := filepath.Join(e.paths.Candidates, "cand-2026-10-01-001")
	if _, err := WriteCandidate(cdir, res, m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []struct{ id, key string }{{"reviewer-a", e.keyA}, {"reviewer-b", e.keyB}} {
		if _, err := Approve(cdir, k.id, k.key, "", "2026-10-01T01:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}
	second, err := Promote(e.paths, cdir, PromoteOptions{PublishedAt: "2026-10-01T02:00:00Z", HeightenedReviewAck: "reviewer-a"})
	if err != nil {
		t.Fatal(err)
	}
	prev, _ := LoadRelease(e.paths, prevID)
	if prev.Manifest.Superseded == nil || prev.Manifest.Superseded.By != second.Manifest.ReleaseID {
		t.Fatal("previous release not marked superseded")
	}
	if err := Rollback(e.paths, prevID, "2026-10-01T03:00:00Z", "test", "regression found"); err != nil {
		t.Fatal(err)
	}
	if cur, _ := CurrentReleaseID(e.paths); cur != prevID {
		t.Fatalf("rollback did not move CURRENT: %s", cur)
	}
	if n, err := audit.Verify(e.paths.Audit); err != nil || n != 3 {
		t.Fatalf("audit chain: n=%d err=%v", n, err)
	}
}
