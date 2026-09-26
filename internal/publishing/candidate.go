// Copyright NU Cybernetics. p(DOOM) — research prototype.

package publishing

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

//go:embed templates/*.tmpl
var templates embed.FS

// dataFiles are the JSON files of a candidate or release directory, in the
// order they are hashed into the manifest.
var dataFiles = []string{"estimates.json", "indexes.json", "aggregations.json", "sensitivity.json", "drivers_explained.json", "delta.json"}

// WriteCandidate writes a complete candidate directory. The directory must not
// exist yet. The manifest's file hashes and signature (manifest hash) are
// filled in here.
func WriteCandidate(dir string, res *model.Result, m schema.ReleaseManifest) (schema.ReleaseManifest, error) {
	if _, err := os.Stat(dir); err == nil {
		return m, fmt.Errorf("candidate directory %s already exists; candidates are never overwritten", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return m, err
	}
	envelope := func(kind string, items any) any {
		return map[string]any{"kind": kind, "schema_version": 1, "items": items}
	}
	files := map[string]any{
		"estimates.json":         envelope("estimate", res.Estimates),
		"indexes.json":           envelope("index_value", res.Indexes),
		"aggregations.json":      envelope("aggregation", res.Aggregations),
		"sensitivity.json":       envelope("sensitivity_run", res.Sensitivity),
		"drivers_explained.json": res.DriversExplained,
		"delta.json":             res.Delta,
	}
	m.Files = nil
	for _, name := range dataFiles {
		raw, err := schema.CanonicalJSONIndent(files[name])
		if err != nil {
			return m, err
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o644); err != nil {
			return m, err
		}
		m.Files = append(m.Files, schema.SnapshotFile{Path: name, SHA256: schema.SHA256Hex(raw), Count: countItems(files[name])})
	}
	h, err := ManifestHash(m)
	if err != nil {
		return m, err
	}
	m.Signature = h
	if err := writeManifest(dir, m); err != nil {
		return m, err
	}
	if err := os.WriteFile(filepath.Join(dir, "approvals.json"), []byte("[]\n"), 0o644); err != nil {
		return m, err
	}
	if err := os.WriteFile(filepath.Join(dir, "changelog.md"), []byte(renderChangelog(m, res)), 0o644); err != nil {
		return m, err
	}
	card, err := renderModelCard(m, res)
	if err != nil {
		return m, err
	}
	if err := os.WriteFile(filepath.Join(dir, "model-card.md"), []byte(card), 0o644); err != nil {
		return m, err
	}
	return m, nil
}

func countItems(v any) int {
	if m, ok := v.(map[string]any); ok {
		switch items := m["items"].(type) {
		case []schema.Estimate:
			return len(items)
		case []schema.IndexValue:
			return len(items)
		case []schema.AggregationResult:
			return len(items)
		case []schema.SensitivityRun:
			return len(items)
		}
	}
	if d, ok := v.(schema.DriversExplained); ok {
		return len(d.Items)
	}
	return 1
}

func writeManifest(dir string, m schema.ReleaseManifest) error {
	raw, err := schema.CanonicalJSONIndent(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0o644)
}

func renderChangelog(m schema.ReleaseManifest, res *model.Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->\n# Changelog — %s\n\n", m.ReleaseID)
	fmt.Fprintf(&b, "Snapshot `%s` (cutoff %s), generated %s.\n\n## Changes\n", m.DataSnapshot, m.SourceCutoff, m.GeneratedAt)
	for _, c := range m.Changes {
		fmt.Fprintf(&b, "- %s\n", c)
	}
	fmt.Fprintf(&b, "\n## Heightened-review triggers\n")
	if len(res.Delta.HeightenedReviewTriggers) == 0 {
		b.WriteString("- none\n")
	}
	for _, t := range res.Delta.HeightenedReviewTriggers {
		fmt.Fprintf(&b, "- %s\n", t)
	}
	fmt.Fprintf(&b, "\n## Editorial risk level\n%s\n", m.EditorialRiskLevel)
	if len(res.Warnings) > 0 {
		b.WriteString("\n## Model warnings\n")
		for _, w := range res.Warnings {
			fmt.Fprintf(&b, "- %s\n", w)
		}
	}
	return b.String()
}

func renderModelCard(m schema.ReleaseManifest, res *model.Result) (string, error) {
	tmpl, err := template.New("model-card.md.tmpl").Funcs(template.FuncMap{"deref": func(p *float64) float64 { return *p }}).ParseFS(templates, "templates/model-card.md.tmpl")
	if err != nil {
		return "", err
	}
	lambda := 0.0
	if l, ok := m.Dependencies["lambda"].(float64); ok {
		lambda = l
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{"Manifest": m, "Aggregations": res.Aggregations, "Lambda": lambda})
	return buf.String(), err
}

// LoadDir reads a candidate or release directory and verifies the file hashes
// listed in its manifest.
func LoadDir(dir string) (*schema.Release, error) {
	r := &schema.Release{Dir: dir}
	if err := readJSON(filepath.Join(dir, "manifest.json"), &r.Manifest); err != nil {
		return nil, err
	}
	listed := map[string]string{}
	for _, f := range r.Manifest.Files {
		listed[f.Path] = f.SHA256
	}
	for _, name := range dataFiles {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if want, ok := listed[name]; ok && want != schema.SHA256Hex(raw) {
			return nil, fmt.Errorf("%s: sha256 mismatch against manifest (file altered after the manifest was written)", name)
		}
		var target any
		switch name {
		case "estimates.json":
			var env schema.Envelope[schema.Estimate]
			target = &env
			if err := decodeStrict(raw, target); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			r.Estimates = env.Items
			continue
		case "indexes.json":
			var env schema.Envelope[schema.IndexValue]
			if err := decodeStrict(raw, &env); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			r.Indexes = env.Items
			continue
		case "aggregations.json":
			var env schema.Envelope[schema.AggregationResult]
			if err := decodeStrict(raw, &env); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			r.Aggregations = env.Items
			continue
		case "sensitivity.json":
			var env schema.Envelope[schema.SensitivityRun]
			if err := decodeStrict(raw, &env); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			r.Sensitivity = env.Items
			continue
		case "drivers_explained.json":
			target = &r.DriversExplained
		case "delta.json":
			target = &r.Delta
		}
		if err := decodeStrict(raw, target); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	approvals, err := os.ReadFile(filepath.Join(dir, "approvals.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if len(approvals) > 0 {
		if err := decodeStrict(approvals, &r.Approvals); err != nil {
			return nil, fmt.Errorf("approvals.json: %w", err)
		}
	}
	if r.Approvals == nil {
		r.Approvals = []schema.Approval{}
	}
	return r, nil
}

func readJSON(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return decodeStrict(raw, v)
}

func decodeStrict(raw []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		return fmt.Errorf("trailing data")
	}
	return nil
}

// ResultFromRelease reconstructs the parts of a model.Result stored in a
// directory (without the simulation internals).
func ResultFromRelease(r *schema.Release) *model.Result {
	return &model.Result{Estimates: r.Estimates, Indexes: r.Indexes, Aggregations: r.Aggregations, Sensitivity: r.Sensitivity, DriversExplained: r.DriversExplained, Delta: r.Delta, EditorialRiskLevel: r.Manifest.EditorialRiskLevel}
}

// listDirs returns the sorted names of subdirectories.
func listDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}
