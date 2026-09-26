// Copyright NU Cybernetics. p(DOOM) — research prototype.

package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Seal rewrites manifest.json so that its files list carries the sha256 and item
// count of every entity file present in dir. All other manifest fields are kept.
// Entity files are never modified. It returns the sealed manifest.
func Seal(dir string) (*Manifest, error) {
	mraw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("seal: read manifest: %w", err)
	}
	var m Manifest
	if err := decodeStrict(mraw, &m); err != nil {
		return nil, fmt.Errorf("seal: manifest.json: %w", err)
	}
	var files []schema.SnapshotFile
	for _, spec := range Files {
		raw, err := os.ReadFile(filepath.Join(dir, spec.Path))
		if err != nil {
			if os.IsNotExist(err) && spec.Optional {
				continue
			}
			return nil, fmt.Errorf("seal: %s: %w", spec.Path, err)
		}
		var env struct {
			Kind  string            `json:"kind"`
			Items []json.RawMessage `json:"items"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			return nil, fmt.Errorf("seal: %s: %w", spec.Path, err)
		}
		if env.Kind != string(spec.Kind) {
			return nil, fmt.Errorf("seal: %s: kind %q, expected %q", spec.Path, env.Kind, spec.Kind)
		}
		files = append(files, schema.SnapshotFile{Path: spec.Path, SHA256: schema.SHA256Hex(raw), Count: len(env.Items)})
	}
	m.Files = files
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	out = append(out, '\n')
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), out, 0o644); err != nil {
		return nil, fmt.Errorf("seal: write manifest: %w", err)
	}
	return &m, nil
}
