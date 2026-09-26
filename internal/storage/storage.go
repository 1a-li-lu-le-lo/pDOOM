// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package storage is the append-only JSONL file store under a data
// directory. It is the file primitive of the ingestion side: the review queue
// appends canonical JSON lines through it and reads them back with a bounded
// scanner. It never rewrites, truncates or deletes anything.
//
// Confinement: a Store is rooted at one directory and only ever opens plain
// "<name>.jsonl" files directly inside it. Names with path separators, parent
// references, other extensions or an absolute path are refused with
// ErrBadName before any file system call.
package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// MaxLineBytes bounds one JSON line, on write and on read.
const MaxLineBytes = 16 << 20

// ErrBadName is returned for a file name that is not a plain <name>.jsonl
// directly inside the store.
var ErrBadName = errors.New("storage: file name must be a plain <name>.jsonl inside the store")

var reName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}\.jsonl$`)

// Store is a directory of append-only JSONL files.
type Store struct {
	dir string
}

// New returns a store rooted at dir. The directory is created by the first
// append; nothing is touched before that.
func New(dir string) (*Store, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("storage: empty directory")
	}
	return &Store{dir: filepath.Clean(dir)}, nil
}

// Dir returns the cleaned root directory.
func (s *Store) Dir() string { return s.dir }

// Path returns the path of name inside the store, refusing anything that is
// not a plain <name>.jsonl file name directly inside it.
func (s *Store) Path(name string) (string, error) {
	if !reName.MatchString(name) || strings.Contains(name, "..") {
		return "", fmt.Errorf("%w: %q", ErrBadName, name)
	}
	p := filepath.Join(s.dir, name)
	if filepath.Dir(p) != s.dir || filepath.Base(p) != name {
		return "", fmt.Errorf("%w: %q", ErrBadName, name)
	}
	return p, nil
}

// Encode returns the canonical JSON line for v (sorted keys, no trailing
// newline). Records that do not fit on one bounded line are refused.
func Encode(v any) ([]byte, error) {
	raw, err := schema.CanonicalJSON(v)
	if err != nil {
		return nil, fmt.Errorf("storage: encode: %w", err)
	}
	if bytes.IndexByte(raw, '\n') >= 0 {
		return nil, errors.New("storage: encoded record contains a newline")
	}
	if len(raw) > MaxLineBytes {
		return nil, fmt.Errorf("storage: encoded record is %d bytes, above the %d-byte line bound", len(raw), MaxLineBytes)
	}
	return raw, nil
}

// Append encodes every record first (nothing is written if one fails), then
// appends them as lines to name, opened with O_APPEND so an existing file is
// never rewritten. It returns the file path and the number of lines written;
// with no records it touches nothing.
func (s *Store) Append(name string, records []any) (path string, n int, err error) {
	path, err = s.Path(name)
	if err != nil {
		return "", 0, err
	}
	lines := make([][]byte, 0, len(records))
	for i, r := range records {
		raw, err := Encode(r)
		if err != nil {
			return path, 0, fmt.Errorf("record %d: %w", i, err)
		}
		lines = append(lines, raw)
	}
	if len(lines) == 0 {
		return path, 0, nil
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return path, 0, fmt.Errorf("storage: create %s: %w", s.dir, err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return path, 0, fmt.Errorf("storage: open %s: %w", path, err)
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 64<<10)
	for _, l := range lines {
		if _, err := w.Write(l); err != nil {
			return path, 0, fmt.Errorf("storage: write %s: %w", path, err)
		}
		if err := w.WriteByte('\n'); err != nil {
			return path, 0, fmt.Errorf("storage: write %s: %w", path, err)
		}
	}
	if err := w.Flush(); err != nil {
		return path, 0, fmt.Errorf("storage: flush %s: %w", path, err)
	}
	return path, len(lines), nil
}

// Read calls fn for every non-blank line of name with its 1-based line
// number and trimmed bytes (valid only during the call). A missing file is
// empty. A line above MaxLineBytes, an unreadable file or an error from fn
// ends the read with an error naming the file and line.
func (s *Store) Read(name string, fn func(line int, raw []byte) error) error {
	path, err := s.Path(name)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("storage: open %s: %w", path, err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), MaxLineBytes)
	line := 0
	for sc.Scan() {
		line++
		raw := bytes.TrimSpace(sc.Bytes())
		if len(raw) == 0 {
			continue
		}
		if err := fn(line, raw); err != nil {
			return fmt.Errorf("storage: %s line %d: %w", path, line, err)
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("storage: read %s after line %d: %w", path, line, err)
	}
	return nil
}

// ReadAll decodes every line of name into a T, in file order.
func ReadAll[T any](s *Store, name string) ([]T, error) {
	var out []T
	err := s.Read(name, func(_ int, raw []byte) error {
		var v T
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		out = append(out, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Files returns the sorted paths of the regular *.jsonl files in the store
// whose base name satisfies match (nil matches every one). A missing
// directory is empty.
func (s *Store) Files(match func(name string) bool) ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: read dir %s: %w", s.dir, err)
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !e.Type().IsRegular() || !reName.MatchString(name) {
			continue
		}
		if match != nil && !match(name) {
			continue
		}
		out = append(out, filepath.Join(s.dir, name))
	}
	sort.Strings(out)
	return out, nil
}
