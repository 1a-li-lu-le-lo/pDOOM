// Copyright NU Cybernetics. p(DOOM) — research prototype.

package publishing

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/audit"
	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// PromoteOptions parameterize Promote.
type PromoteOptions struct {
	// RequiredApprovals defaults to RequiredApprovals when zero.
	RequiredApprovals int
	// HeightenedReviewAck names the second reviewer acknowledging the
	// heightened-review triggers. Required when the delta lists any trigger.
	HeightenedReviewAck string
	// PublishedAt is the RFC 3339 publication timestamp (caller-supplied).
	PublishedAt string
	// Actor is recorded in the audit log.
	Actor string
}

// GateError is returned when a promotion gate fails.
type GateError struct {
	Gate   string
	Reason string
}

func (e *GateError) Error() string { return fmt.Sprintf("gate %s failed: %s", e.Gate, e.Reason) }

func gate(name, format string, args ...any) error {
	return &GateError{Gate: name, Reason: fmt.Sprintf(format, args...)}
}

// Promote verifies every gate and copies the candidate into data/releases.
//
// Gates, in order: manifest integrity (file hashes and manifest hash), signed
// approvals (≥ required, distinct reviewers, valid signatures), heightened
// review (two distinct reviewers and an explicit acknowledgement when triggers
// exist), reproducibility (the model is re-run on the referenced snapshot and
// must reproduce every data file byte-for-byte), invariants, and uniqueness of
// the release id. Only after all gates pass are files copied, the previous
// release marked superseded, CURRENT moved and the audit event appended.
func Promote(paths config.Paths, candidateDir string, opts PromoteOptions) (*schema.Release, error) {
	if opts.PublishedAt == "" {
		return nil, gate("options", "published_at is required")
	}
	if opts.Actor == "" {
		opts.Actor = "pdoomctl"
	}
	required := opts.RequiredApprovals
	if required <= 0 {
		required = RequiredApprovals
	}
	cand, err := LoadDir(candidateDir)
	if err != nil {
		return nil, gate("integrity", "%v", err)
	}
	m := cand.Manifest
	if m.Published != nil {
		return nil, gate("integrity", "candidate is already marked published")
	}
	approvers, err := VerifyApprovals(cand, paths.Keys)
	if err != nil {
		return nil, gate("approvals", "%v", err)
	}
	if len(approvers) < required {
		return nil, gate("approvals", "%d distinct signed approval(s), %d required", len(approvers), required)
	}
	heightened := len(cand.Delta.HeightenedReviewTriggers) > 0
	if heightened {
		if opts.HeightenedReviewAck == "" {
			return nil, gate("heightened_review", "triggers %v require --heightened-review-ack <reviewer-id> from a second reviewer", cand.Delta.HeightenedReviewTriggers)
		}
		if len(approvers) < 2 {
			return nil, gate("heightened_review", "heightened review requires approvals from at least two distinct reviewers (have %v)", approvers)
		}
		if !contains(approvers, opts.HeightenedReviewAck) {
			return nil, gate("heightened_review", "acknowledging reviewer %s has not signed an approval", opts.HeightenedReviewAck)
		}
	}
	// Reproducibility.
	snapDir := filepath.Join(paths.Snapshots, m.DataSnapshot)
	snap, err := snapshot.Load(snapDir)
	if err != nil {
		return nil, gate("reproducibility", "load snapshot: %v", err)
	}
	if problems := snapshot.Validate(snap, paths.SchemaDir); snapshot.HasErrors(problems) {
		return nil, gate("reproducibility", "snapshot has validation errors: %s", problems[0])
	}
	var prev *schema.Release
	if m.PreviousReleaseID != nil {
		prev, err = LoadRelease(paths, *m.PreviousReleaseID)
		if err != nil {
			return nil, gate("reproducibility", "load previous release: %v", err)
		}
	}
	res, err := model.Run(snap, model.RunOptions{PreviousRelease: prev, GeneratedAt: m.GeneratedAt, CodeCommit: m.CodeCommit})
	if err != nil {
		return nil, gate("reproducibility", "model run: %v", err)
	}
	if errs := model.CheckInvariants(res); len(errs) > 0 {
		return nil, gate("invariants", "%s", strings.Join(errs, "; "))
	}
	rerun := filepath.Join(os.TempDir(), "pdoom-reproduce-"+m.CandidateID+"-"+schema.SHA256Hex([]byte(candidateDir))[:8])
	_ = os.RemoveAll(rerun)
	defer os.RemoveAll(rerun)
	m2, err := BuildManifest(res, snap, BuildOptions{ReleaseID: m.ReleaseID, CandidateID: m.CandidateID, CodeCommit: m.CodeCommit, PreviousReleaseID: m.PreviousReleaseID})
	if err != nil {
		return nil, gate("reproducibility", "%v", err)
	}
	if _, err := WriteCandidate(rerun, res, m2); err != nil {
		return nil, gate("reproducibility", "%v", err)
	}
	for _, name := range dataFiles {
		a, _ := os.ReadFile(filepath.Join(candidateDir, name))
		b, _ := os.ReadFile(filepath.Join(rerun, name))
		if string(a) != string(b) {
			return nil, gate("reproducibility", "%s differs when the model is re-run on snapshot %s", name, m.DataSnapshot)
		}
	}
	// Uniqueness.
	dest := filepath.Join(paths.Releases, m.ReleaseID)
	if _, err := os.Stat(dest); err == nil {
		return nil, gate("uniqueness", "release %s already exists; releases are immutable", m.ReleaseID)
	}
	existing, _ := listDirs(paths.Releases)
	for _, id := range existing {
		if r, err := LoadDir(filepath.Join(paths.Releases, id)); err == nil && r.Manifest.DataSnapshot == m.DataSnapshot && r.Manifest.CodeCommit == m.CodeCommit && r.Manifest.GeneratedAt == m.GeneratedAt {
			return nil, gate("uniqueness", "release %s already publishes this exact snapshot, commit and generation time", id)
		}
	}

	// All gates passed: publish.
	if err := copyDir(candidateDir, dest); err != nil {
		return nil, err
	}
	published := opts.PublishedAt
	m.Published = &published
	m.Reviewers = approvers
	var ack *string
	if opts.HeightenedReviewAck != "" {
		a := opts.HeightenedReviewAck
		ack = &a
	}
	m.Approval = &schema.ApprovalPolicy{RequiredApprovals: required, ReceivedApprovals: len(approvers), HeightenedReview: heightened, HeightenedReviewAck: ack, ApproverIDs: approvers}
	if err := writeManifest(dest, m); err != nil {
		return nil, err
	}
	if prev != nil {
		pm := prev.Manifest
		pm.Superseded = &schema.Superseded{By: m.ReleaseID, At: published}
		if err := writeManifest(prev.Dir, pm); err != nil {
			return nil, err
		}
	}
	if err := os.WriteFile(paths.Current, []byte(m.ReleaseID+"\n"), 0o644); err != nil {
		return nil, err
	}
	_, err = audit.Append(paths.Audit, audit.Event{TS: published, Actor: opts.Actor, Action: "release.promote", Subject: m.ReleaseID,
		Details: map[string]any{"candidate": m.CandidateID, "snapshot": m.DataSnapshot, "code_commit": m.CodeCommit, "approvers": approvers, "heightened_review": heightened, "heightened_review_ack": opts.HeightenedReviewAck, "manifest_hash": m.Signature, "previous_release": m.PreviousReleaseID}})
	if err != nil {
		return nil, err
	}
	return LoadDir(dest)
}

// Rollback points CURRENT at an existing release and records the reason.
func Rollback(paths config.Paths, releaseID, at, actor, reason string) error {
	if at == "" || reason == "" {
		return fmt.Errorf("rollback requires a timestamp and a reason")
	}
	if actor == "" {
		actor = "pdoomctl"
	}
	if _, err := LoadRelease(paths, releaseID); err != nil {
		return fmt.Errorf("release %s: %w", releaseID, err)
	}
	cur, _ := CurrentReleaseID(paths)
	if cur == releaseID {
		return fmt.Errorf("release %s is already current", releaseID)
	}
	if err := os.WriteFile(paths.Current, []byte(releaseID+"\n"), 0o644); err != nil {
		return err
	}
	_, err := audit.Append(paths.Audit, audit.Event{TS: at, Actor: actor, Action: "release.rollback", Subject: releaseID, Details: map[string]any{"from": cur, "reason": reason}})
	return err
}

// CurrentReleaseID reads data/releases/CURRENT.
func CurrentReleaseID(paths config.Paths) (string, error) {
	raw, err := os.ReadFile(paths.Current)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

// LoadRelease loads a promoted release by id.
func LoadRelease(paths config.Paths, releaseID string) (*schema.Release, error) {
	if !reReleaseID.MatchString(releaseID) {
		return nil, fmt.Errorf("invalid release id %q", releaseID)
	}
	return LoadDir(filepath.Join(paths.Releases, releaseID))
}

// LoadCurrentRelease loads the release CURRENT points at.
func LoadCurrentRelease(paths config.Paths) (*schema.Release, error) {
	id, err := CurrentReleaseID(paths)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no current release (data/releases/CURRENT missing)")
		}
		return nil, err
	}
	return LoadRelease(paths, id)
}

// ListReleases summarizes every promoted release.
func ListReleases(paths config.Paths) ([]schema.ReleaseSummary, error) {
	ids, err := listDirs(paths.Releases)
	if err != nil {
		return nil, err
	}
	cur, _ := CurrentReleaseID(paths)
	var out []schema.ReleaseSummary
	for _, id := range ids {
		if !reReleaseID.MatchString(id) {
			continue
		}
		r, err := LoadDir(filepath.Join(paths.Releases, id))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		out = append(out, schema.ReleaseSummary{ReleaseID: id, DataSnapshot: r.Manifest.DataSnapshot, ModelVersions: r.Manifest.ModelVersions, Published: r.Manifest.Published, Superseded: r.Manifest.Superseded, IsCurrent: id == cur, EditorialRiskLevel: r.Manifest.EditorialRiskLevel, UncertaintyScore: r.Manifest.UncertaintyScore})
	}
	return out, nil
}

// ExistingIDs lists release and candidate ids (for NextReleaseID).
func ExistingIDs(paths config.Paths) []string {
	var ids []string
	rels, _ := listDirs(paths.Releases)
	ids = append(ids, rels...)
	cands, _ := listDirs(paths.Candidates)
	for _, c := range cands {
		if r, err := LoadDir(filepath.Join(paths.Candidates, c)); err == nil {
			ids = append(ids, r.Manifest.ReleaseID)
		}
	}
	return ids
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		in, err := os.Open(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		out, err := os.Create(filepath.Join(dst, e.Name()))
		if err != nil {
			in.Close()
			return err
		}
		_, err = io.Copy(out, in)
		in.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
