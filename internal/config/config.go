// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package config resolves the few settings shared by the Go binaries.
package config

import (
	"os"
	"path/filepath"
)

// Paths groups the on-disk locations used by the tools.
type Paths struct {
	DataDir    string
	SchemaDir  string
	Snapshots  string
	Candidates string
	Releases   string
	Current    string
	Audit      string
	Keys       string
	Review     string
}

// Resolve builds Paths from a data directory (flag) falling back to
// PDOOM_DATA_DIR and then ./data; the schema directory falls back to
// PDOOM_SCHEMA_DIR and then <data>/schemas.
func Resolve(dataDir, schemaDir string) Paths {
	if dataDir == "" {
		dataDir = os.Getenv("PDOOM_DATA_DIR")
	}
	if dataDir == "" {
		dataDir = "data"
	}
	if schemaDir == "" {
		schemaDir = os.Getenv("PDOOM_SCHEMA_DIR")
	}
	if schemaDir == "" {
		schemaDir = filepath.Join(dataDir, "schemas")
	}
	return Paths{
		DataDir:    dataDir,
		SchemaDir:  schemaDir,
		Snapshots:  filepath.Join(dataDir, "snapshots"),
		Candidates: filepath.Join(dataDir, "candidates"),
		Releases:   filepath.Join(dataDir, "releases"),
		Current:    filepath.Join(dataDir, "releases", "CURRENT"),
		Audit:      filepath.Join(dataDir, "audit", "audit.jsonl"),
		Keys:       filepath.Join(dataDir, "keys", "reviewers"),
		Review:     filepath.Join(dataDir, "review-queue"),
	}
}

// Env returns an environment variable or a default.
func Env(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}
