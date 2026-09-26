// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package publishing implements candidates, signed approvals, promotion and
// rollback of releases (build-spec §3.6). Only Promote and Rollback write into
// data/releases; everything else works on candidate directories.
package publishing

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var reReviewerID = regexp.MustCompile(`^[a-z0-9]+(?:[-_.][a-z0-9]+)*$`)

// GenerateKey creates an ed25519 key pair for a reviewer and writes
// <dir>/<id>.key (private, base64, mode 0600) and <dir>/<id>.pub (public,
// base64). The private key must never be committed.
func GenerateKey(dir, id string) (pubPath, keyPath string, err error) {
	if !reReviewerID.MatchString(id) {
		return "", "", fmt.Errorf("reviewer id %q must match %s", id, reReviewerID)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	pubPath = filepath.Join(dir, id+".pub")
	keyPath = filepath.Join(dir, id+".key")
	if _, err := os.Stat(pubPath); err == nil {
		return "", "", fmt.Errorf("%s already exists; refusing to overwrite a reviewer key", pubPath)
	}
	if err := os.WriteFile(pubPath, []byte(base64.StdEncoding.EncodeToString(pub)+"\n"), 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(keyPath, []byte(base64.StdEncoding.EncodeToString(priv)+"\n"), 0o600); err != nil {
		return "", "", err
	}
	return pubPath, keyPath, nil
}

// LoadPrivateKey reads a base64 ed25519 private key file.
func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("private key %s: %w", path, err)
	}
	if len(b) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("private key %s: wrong length %d", path, len(b))
	}
	return ed25519.PrivateKey(b), nil
}

// LoadPublicKey reads <keysDir>/<keyID>.pub.
func LoadPublicKey(keysDir, keyID string) (ed25519.PublicKey, error) {
	if !reReviewerID.MatchString(keyID) {
		return nil, fmt.Errorf("invalid key id %q", keyID)
	}
	raw, err := os.ReadFile(filepath.Join(keysDir, keyID+".pub"))
	if err != nil {
		return nil, err
	}
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("public key %s: %w", keyID, err)
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key %s: wrong length %d", keyID, len(b))
	}
	return ed25519.PublicKey(b), nil
}

// signHash signs the manifest hash (as ASCII hex bytes) and returns base64.
func signHash(priv ed25519.PrivateKey, hashHex string) string {
	return base64.StdEncoding.EncodeToString(ed25519.Sign(priv, []byte(hashHex)))
}

// verifyHash checks a base64 signature over the manifest hash.
func verifyHash(pub ed25519.PublicKey, hashHex, sigB64 string) bool {
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return false
	}
	return ed25519.Verify(pub, []byte(hashHex), sig)
}
