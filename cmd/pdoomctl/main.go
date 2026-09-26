// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Command pdoomctl operates the p(DOOM) data pipeline: it validates and seals
// snapshots, runs the deterministic model into a candidate directory, records
// signed reviewer approvals, promotes candidates through the safety gate,
// rolls releases back and verifies the audit chain. It is the only way a
// release reaches data/releases; nothing is automatic.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/audit"
	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/publishing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

const usage = `pdoomctl — p(DOOM) snapshot, candidate and release operations

Usage:
  pdoomctl [--data-dir DIR] [--schema-dir DIR] [--json] <command> [args]

Commands:
  snapshot validate <dir>                 apply JSON Schemas and invariants (exit 1 on errors)
  snapshot seal <dir>                     write file hashes and counts into manifest.json
  model run --snapshot <dir> --out <candidate-dir> [--release-id ID] [--previous REL]
            --generated-at RFC3339 [--code-commit SHA]
                                          deterministic model run → candidate (never a release)
  release diff <candidate-dir> [release-id]   print the change record versus a release (default: current)
  release approve <candidate-dir> --reviewer ID --key FILE --signed-at RFC3339 [--conflicts TEXT]
  release promote <candidate-dir> --published-at RFC3339 [--heightened-review-ack ID] [--actor NAME]
  release rollback <release-id> --at RFC3339 --reason TEXT [--actor NAME]
  release list
  audit verify [path]
  keygen --id REVIEWER-ID --out DIR       create an ed25519 reviewer key pair (private key stays out of git)

Timestamps are always supplied explicitly; the tool never reads the clock for
anything that becomes part of a release.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	global := flag.NewFlagSet("pdoomctl", flag.ContinueOnError)
	dataDir := global.String("data-dir", "", "data directory (default $PDOOM_DATA_DIR or ./data)")
	schemaDir := global.String("schema-dir", "", "JSON Schema directory (default <data>/schemas)")
	asJSON := global.Bool("json", false, "machine-readable output")
	global.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	if err := global.Parse(args); err != nil {
		return 2
	}
	paths := config.Resolve(*dataDir, *schemaDir)
	rest := global.Args()
	if len(rest) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	out := func(v any, human string) {
		if *asJSON {
			raw, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(raw))
			return
		}
		fmt.Println(human)
	}
	fail := func(err error) int {
		if *asJSON {
			raw, _ := json.Marshal(map[string]any{"error": err.Error()})
			fmt.Println(string(raw))
		} else {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		return 1
	}
	switch strings.Join(rest[:min(2, len(rest))], " ") {
	case "snapshot validate":
		if len(rest) < 3 {
			return fail(fmt.Errorf("usage: snapshot validate <dir>"))
		}
		s, err := snapshot.Load(rest[2])
		if err != nil {
			return fail(err)
		}
		problems := snapshot.Validate(s, paths.SchemaDir)
		var lines []string
		for _, p := range problems {
			lines = append(lines, p.String())
		}
		summary := fmt.Sprintf("%s: %d problems (%d errors)", s.Manifest.SnapshotID, len(problems), countErrors(problems))
		out(map[string]any{"snapshot_id": s.Manifest.SnapshotID, "problems": problems, "errors": countErrors(problems)}, strings.Join(append(lines, summary), "\n"))
		if snapshot.HasErrors(problems) {
			return 1
		}
		return 0
	case "snapshot seal":
		if len(rest) < 3 {
			return fail(fmt.Errorf("usage: snapshot seal <dir>"))
		}
		m, err := snapshot.Seal(rest[2])
		if err != nil {
			return fail(err)
		}
		out(m, fmt.Sprintf("sealed %s (%d files)", m.SnapshotID, len(m.Files)))
		return 0
	case "model run":
		fs := flag.NewFlagSet("model run", flag.ContinueOnError)
		snapDir := fs.String("snapshot", "", "snapshot directory")
		outDir := fs.String("out", "", "candidate directory to create")
		releaseID := fs.String("release-id", "", "release id to reserve (default next rel-<date>-NNN)")
		previous := fs.String("previous", "", "previous release id (default: current release if any)")
		generatedAt := fs.String("generated-at", "", "RFC 3339 generation timestamp (required)")
		codeCommit := fs.String("code-commit", "", "git commit of the model code")
		noPrevious := fs.Bool("no-previous", false, "compute as a baseline release even if a current release exists")
		if err := fs.Parse(rest[2:]); err != nil {
			return 2
		}
		if *snapDir == "" || *outDir == "" || *generatedAt == "" {
			return fail(fmt.Errorf("--snapshot, --out and --generated-at are required"))
		}
		if strings.HasPrefix(filepath.Clean(*outDir), filepath.Clean(paths.Releases)) {
			return fail(fmt.Errorf("refusing to write a candidate inside %s; only `release promote` writes there", paths.Releases))
		}
		s, err := snapshot.Load(*snapDir)
		if err != nil {
			return fail(err)
		}
		if problems := snapshot.Validate(s, paths.SchemaDir); snapshot.HasErrors(problems) {
			return fail(fmt.Errorf("snapshot has %d validation errors; run `pdoomctl snapshot validate %s`", countErrors(problems), *snapDir))
		}
		var prev *schema.Release
		var prevID *string
		if !*noPrevious {
			if *previous == "" {
				if cur, err := publishing.CurrentReleaseID(paths); err == nil && cur != "" {
					*previous = cur
				}
			}
			if *previous != "" {
				prev, err = publishing.LoadRelease(paths, *previous)
				if err != nil {
					return fail(fmt.Errorf("previous release: %w", err))
				}
				prevID = previous
			}
		}
		res, err := model.Run(s, model.RunOptions{PreviousRelease: prev, GeneratedAt: *generatedAt, CodeCommit: *codeCommit})
		if err != nil {
			return fail(err)
		}
		if errs := model.CheckInvariants(res); len(errs) > 0 {
			return fail(fmt.Errorf("invariants violated:\n  %s", strings.Join(errs, "\n  ")))
		}
		if *releaseID == "" {
			*releaseID = publishing.NextReleaseID((*generatedAt)[:10], publishing.ExistingIDs(paths))
		}
		candID := filepath.Base(*outDir)
		m, err := publishing.BuildManifest(res, s, publishing.BuildOptions{ReleaseID: *releaseID, CandidateID: candID, CodeCommit: *codeCommit, PreviousReleaseID: prevID})
		if err != nil {
			return fail(err)
		}
		m, err = publishing.WriteCandidate(*outDir, res, m)
		if err != nil {
			return fail(err)
		}
		human := fmt.Sprintf("candidate %s written to %s\n  release id     %s\n  snapshot       %s\n  editorial      %s\n  manifest hash  %s\n  triggers       %s\n  estimates      %d, indexes %d, aggregations %d, sensitivity runs %d\nNext: pdoomctl release approve %s --reviewer <id> --key <file> --signed-at <ts>",
			candID, *outDir, m.ReleaseID, m.DataSnapshot, m.EditorialRiskLevel, m.Signature[:16], strings.Join(res.Delta.HeightenedReviewTriggers, ", "), len(res.Estimates), len(res.Indexes), len(res.Aggregations), len(res.Sensitivity), *outDir)
		out(map[string]any{"candidate_dir": *outDir, "release_id": m.ReleaseID, "manifest_hash": m.Signature, "editorial_risk_level": m.EditorialRiskLevel, "heightened_review_triggers": res.Delta.HeightenedReviewTriggers, "warnings": res.Warnings}, human)
		return 0
	case "release diff":
		if len(rest) < 3 {
			return fail(fmt.Errorf("usage: release diff <candidate-dir> [release-id]"))
		}
		cand, err := publishing.LoadDir(rest[2])
		if err != nil {
			return fail(err)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "candidate %s → release %s (previous: %v)\n", cand.Manifest.CandidateID, cand.Manifest.ReleaseID, deref(cand.Delta.PreviousReleaseID))
		for _, c := range cand.Delta.EstimateChanges {
			if c.ChangePoints != nil {
				fmt.Fprintf(&b, "  %-40s %s → %s (%+.2f pts)\n", c.EstimateID, c.PreviousDisplay, c.NewDisplay, *c.ChangePoints)
			} else {
				fmt.Fprintf(&b, "  %-40s %s\n", c.EstimateID, c.NewDisplay)
			}
		}
		for _, c := range cand.Delta.IndexChanges {
			fmt.Fprintf(&b, "  index %-28s %v → %v\n", c.IndexID, fmtF(c.Previous), fmtF(c.New))
		}
		fmt.Fprintf(&b, "heightened-review triggers: %s", strings.Join(cand.Delta.HeightenedReviewTriggers, ", "))
		out(cand.Delta, b.String())
		return 0
	case "release approve":
		fs := flag.NewFlagSet("release approve", flag.ContinueOnError)
		reviewer := fs.String("reviewer", "", "reviewer id (must match a public key in data/keys/reviewers)")
		key := fs.String("key", "", "private key file")
		signedAt := fs.String("signed-at", "", "RFC 3339 timestamp")
		conflicts := fs.String("conflicts", "", "declared conflicts of interest")
		if len(rest) < 3 {
			return fail(fmt.Errorf("usage: release approve <candidate-dir> --reviewer ID --key FILE --signed-at TS"))
		}
		if err := fs.Parse(rest[3:]); err != nil {
			return 2
		}
		ap, err := publishing.Approve(rest[2], *reviewer, *key, *conflicts, *signedAt)
		if err != nil {
			return fail(err)
		}
		out(ap, fmt.Sprintf("approval recorded for %s (manifest %s…)", ap.ReviewerID, ap.ManifestSHA256[:16]))
		return 0
	case "release promote":
		fs := flag.NewFlagSet("release promote", flag.ContinueOnError)
		publishedAt := fs.String("published-at", "", "RFC 3339 timestamp")
		ack := fs.String("heightened-review-ack", "", "second reviewer acknowledging heightened-review triggers")
		actor := fs.String("actor", "", "actor recorded in the audit log")
		required := fs.Int("required-approvals", 0, "override required approvals (default policy)")
		if len(rest) < 3 {
			return fail(fmt.Errorf("usage: release promote <candidate-dir> --published-at TS"))
		}
		if err := fs.Parse(rest[3:]); err != nil {
			return 2
		}
		rel, err := publishing.Promote(paths, rest[2], publishing.PromoteOptions{PublishedAt: *publishedAt, HeightenedReviewAck: *ack, Actor: *actor, RequiredApprovals: *required})
		if err != nil {
			return fail(err)
		}
		out(rel.Manifest, fmt.Sprintf("promoted %s (now CURRENT); approvers %v; audit event appended", rel.Manifest.ReleaseID, rel.Manifest.Reviewers))
		return 0
	case "release rollback":
		fs := flag.NewFlagSet("release rollback", flag.ContinueOnError)
		at := fs.String("at", "", "RFC 3339 timestamp")
		reason := fs.String("reason", "", "why")
		actor := fs.String("actor", "", "actor")
		if len(rest) < 3 {
			return fail(fmt.Errorf("usage: release rollback <release-id> --at TS --reason TEXT"))
		}
		if err := fs.Parse(rest[3:]); err != nil {
			return 2
		}
		if err := publishing.Rollback(paths, rest[2], *at, *actor, *reason); err != nil {
			return fail(err)
		}
		out(map[string]any{"current": rest[2]}, "CURRENT now points at "+rest[2])
		return 0
	case "release list":
		list, err := publishing.ListReleases(paths)
		if err != nil {
			return fail(err)
		}
		var b strings.Builder
		for _, r := range list {
			mark := " "
			if r.IsCurrent {
				mark = "*"
			}
			fmt.Fprintf(&b, "%s %s  snapshot %s  published %s", mark, r.ReleaseID, r.DataSnapshot, deref(r.Published))
			if r.Superseded != nil {
				fmt.Fprintf(&b, "  superseded by %s", r.Superseded.By)
			}
			b.WriteString("\n")
		}
		if len(list) == 0 {
			b.WriteString("no releases")
		}
		out(list, strings.TrimRight(b.String(), "\n"))
		return 0
	case "audit verify":
		path := paths.Audit
		if len(rest) >= 3 {
			path = rest[2]
		}
		n, err := audit.Verify(path)
		if err != nil {
			return fail(err)
		}
		out(map[string]any{"events": n, "ok": true}, fmt.Sprintf("audit chain ok (%d events)", n))
		return 0
	}
	if rest[0] == "keygen" {
		fs := flag.NewFlagSet("keygen", flag.ContinueOnError)
		id := fs.String("id", "", "reviewer id")
		outDir := fs.String("out", paths.Keys, "output directory")
		if err := fs.Parse(rest[1:]); err != nil {
			return 2
		}
		if *id == "" {
			return fail(fmt.Errorf("--id is required"))
		}
		pub, key, err := publishing.GenerateKey(*outDir, *id)
		if err != nil {
			return fail(err)
		}
		out(map[string]any{"public_key": pub, "private_key": key}, fmt.Sprintf("public key  %s (commit this)\nprivate key %s (NEVER commit; *.key is gitignored)", pub, key))
		return 0
	}
	fmt.Fprint(os.Stderr, usage)
	return 2
}

func countErrors(ps []snapshot.Problem) int {
	n := 0
	for _, p := range ps {
		if p.Severity == snapshot.SeverityError {
			n++
		}
	}
	return n
}

func deref(s *string) string {
	if s == nil {
		return "none"
	}
	return *s
}

func fmtF(f *float64) string {
	if f == nil {
		return "null"
	}
	return fmt.Sprintf("%.1f", *f)
}
