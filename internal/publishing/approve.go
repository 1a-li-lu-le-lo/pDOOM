// Copyright NU Cybernetics. p(DOOM) — research prototype.

package publishing

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Approve signs the candidate's manifest hash with the reviewer's private key
// and appends the approval to approvals.json. The key id is the reviewer id;
// the matching public key must live in the reviewer keys directory for
// Promote to accept it.
func Approve(candidateDir, reviewerID, keyPath, conflicts, signedAt string) (schema.Approval, error) {
	if !reReviewerID.MatchString(reviewerID) {
		return schema.Approval{}, fmt.Errorf("reviewer id %q must match %s", reviewerID, reReviewerID)
	}
	if signedAt == "" {
		return schema.Approval{}, fmt.Errorf("signed_at is required")
	}
	rel, err := LoadDir(candidateDir)
	if err != nil {
		return schema.Approval{}, fmt.Errorf("load candidate: %w", err)
	}
	if rel.Manifest.Published != nil {
		return schema.Approval{}, fmt.Errorf("%s is already published; approvals are only added to candidates", rel.Manifest.ReleaseID)
	}
	h, err := ManifestHash(rel.Manifest)
	if err != nil {
		return schema.Approval{}, err
	}
	if h != rel.Manifest.Signature {
		return schema.Approval{}, fmt.Errorf("manifest hash %s does not match recorded signature %s; the manifest was altered", h[:12], rel.Manifest.Signature[:12])
	}
	for _, a := range rel.Approvals {
		if a.ReviewerID == reviewerID {
			return schema.Approval{}, fmt.Errorf("reviewer %s has already approved this candidate", reviewerID)
		}
	}
	priv, err := LoadPrivateKey(keyPath)
	if err != nil {
		return schema.Approval{}, err
	}
	if conflicts == "" {
		conflicts = "none declared"
	}
	ap := schema.Approval{ReviewerID: reviewerID, KeyID: reviewerID, SignedAt: signedAt, ManifestSHA256: h, SignatureBase64: signHash(priv, h), ConflictsDeclared: conflicts}
	rel.Approvals = append(rel.Approvals, ap)
	sort.Slice(rel.Approvals, func(i, j int) bool { return rel.Approvals[i].ReviewerID < rel.Approvals[j].ReviewerID })
	raw, err := schema.CanonicalJSONIndent(rel.Approvals)
	if err != nil {
		return ap, err
	}
	return ap, os.WriteFile(filepath.Join(candidateDir, "approvals.json"), raw, 0o644)
}

// VerifyApprovals checks every approval signature against the public keys in
// keysDir and returns the distinct approving reviewer ids.
func VerifyApprovals(rel *schema.Release, keysDir string) ([]string, error) {
	h, err := ManifestHash(rel.Manifest)
	if err != nil {
		return nil, err
	}
	if h != rel.Manifest.Signature {
		return nil, fmt.Errorf("manifest hash does not match its recorded signature; the manifest was altered after it was written")
	}
	seen := map[string]bool{}
	for _, a := range rel.Approvals {
		if a.ManifestSHA256 != h {
			return nil, fmt.Errorf("approval by %s signs hash %s, manifest is %s", a.ReviewerID, a.ManifestSHA256[:12], h[:12])
		}
		pub, err := LoadPublicKey(keysDir, a.KeyID)
		if err != nil {
			return nil, fmt.Errorf("approval by %s: %w", a.ReviewerID, err)
		}
		if !verifyHash(pub, h, a.SignatureBase64) {
			return nil, fmt.Errorf("approval by %s: signature does not verify", a.ReviewerID)
		}
		seen[a.ReviewerID] = true
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}
