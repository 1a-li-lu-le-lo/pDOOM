// Copyright NU Cybernetics. p(DOOM) — research prototype.

package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type rec struct {
	B int    `json:"b"`
	A string `json:"a"`
}

func TestAppendReadRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "queue")
	st, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("New must not create the directory")
	}
	path, n, err := st.Append("a.jsonl", []any{rec{B: 1, A: "x"}, map[string]any{"z": true, "y": []int{1}}})
	if err != nil || n != 2 {
		t.Fatalf("append: n=%d err=%v", n, err)
	}
	if path != filepath.Join(dir, "a.jsonl") {
		t.Fatalf("path = %s", path)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "{\"a\":\"x\",\"b\":1}\n{\"y\":[1],\"z\":true}\n" {
		t.Fatalf("lines are not canonical JSON:\n%s", raw)
	}
	// A second append adds lines; nothing is rewritten.
	if _, n, err := st.Append("a.jsonl", []any{rec{B: 3, A: "w"}}); err != nil || n != 1 {
		t.Fatalf("second append: n=%d err=%v", n, err)
	}
	items, err := ReadAll[map[string]any](st, "a.jsonl")
	if err != nil || len(items) != 3 {
		t.Fatalf("read all: %d %v", len(items), err)
	}
	if items[0]["a"] != "x" || items[2]["b"] != float64(3) {
		t.Fatalf("items = %v", items)
	}
	// Blank lines are skipped and line numbers refer to the file.
	os.WriteFile(filepath.Join(dir, "b.jsonl"), []byte("\n{\"n\":1}\n\n   \n{\"n\":2}\n"), 0o644)
	var lines []int
	err = st.Read("b.jsonl", func(line int, raw []byte) error {
		lines = append(lines, line)
		if raw[0] != '{' {
			t.Fatalf("raw must be trimmed: %q", raw)
		}
		return nil
	})
	if err != nil || len(lines) != 2 || lines[0] != 2 || lines[1] != 5 {
		t.Fatalf("lines = %v err=%v", lines, err)
	}
}

func TestAppendWritesNothingOnBadRecordOrNoRecords(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "queue")
	st, _ := New(dir)
	if _, n, err := st.Append("a.jsonl", nil); err != nil || n != 0 {
		t.Fatalf("empty append: n=%d err=%v", n, err)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("empty append must not create anything")
	}
	_, n, err := st.Append("a.jsonl", []any{rec{B: 1}, make(chan int)})
	if err == nil || n != 0 || !strings.Contains(err.Error(), "record 1") {
		t.Fatalf("unencodable record must abort before writing: n=%d err=%v", n, err)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("nothing may be written when a record fails to encode")
	}
	big := strings.Repeat("x", MaxLineBytes)
	if _, err := Encode(map[string]string{"s": big}); err == nil {
		t.Fatal("a record above the line bound must be refused")
	}
}

func TestPathConfinement(t *testing.T) {
	st, _ := New("/data/review-queue/")
	if st.Dir() != "/data/review-queue" {
		t.Fatalf("dir = %s", st.Dir())
	}
	for _, bad := range []string{"", "../x.jsonl", "a/b.jsonl", "x.json", "/abs.jsonl", ".jsonl", "a..b.jsonl", "..jsonl",
		"a\\b.jsonl", "a b.jsonl", strings.Repeat("a", 200) + ".jsonl", "x.jsonl\n"} {
		if p, err := st.Path(bad); !errors.Is(err, ErrBadName) {
			t.Errorf("%q must be refused, got %q %v", bad, p, err)
		}
	}
	p, err := st.Path("2026-09-26-ingest.jsonl")
	if err != nil || p != "/data/review-queue/2026-09-26-ingest.jsonl" {
		t.Fatalf("path = %q %v", p, err)
	}
	if _, err := New("  "); err == nil {
		t.Fatal("empty dir must be refused")
	}
	// Read and Append refuse bad names before touching the file system.
	if err := st.Read("../etc.jsonl", func(int, []byte) error { return nil }); !errors.Is(err, ErrBadName) {
		t.Fatalf("read bad name: %v", err)
	}
	if _, _, err := st.Append("../etc.jsonl", []any{1}); !errors.Is(err, ErrBadName) {
		t.Fatalf("append bad name: %v", err)
	}
}

func TestReadErrorsNameTheLine(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	os.WriteFile(filepath.Join(dir, "c.jsonl"), []byte("{\"ok\":1}\n\nnot json\n"), 0o644)
	if _, err := ReadAll[map[string]any](st, "c.jsonl"); err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Fatalf("corrupt line must be reported with its position: %v", err)
	}
	if items, err := ReadAll[rec](st, "missing.jsonl"); err != nil || items != nil {
		t.Fatalf("missing file must be empty: %v %v", items, err)
	}
	// A callback error is wrapped with the position.
	sentinel := errors.New("stop")
	err := st.Read("c.jsonl", func(int, []byte) error { return sentinel })
	if !errors.Is(err, sentinel) || !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("callback error: %v", err)
	}
	// A line above the bound is an error, not a silent truncation.
	long := append([]byte("{\"s\":\""), make([]byte, MaxLineBytes)...)
	for i := 6; i < len(long); i++ {
		long[i] = 'x'
	}
	long = append(long, '"', '}', '\n')
	if err := os.WriteFile(filepath.Join(dir, "long.jsonl"), long, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadAll[map[string]any](st, "long.jsonl"); err == nil || !strings.Contains(err.Error(), "long.jsonl") {
		t.Fatalf("oversized line must fail: %v", err)
	}
}

func TestFiles(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	for _, name := range []string{"b.jsonl", "a.jsonl", "notes.txt"} {
		os.WriteFile(filepath.Join(dir, name), []byte("{}\n"), 0o644)
	}
	os.MkdirAll(filepath.Join(dir, "sub.jsonl"), 0o755)
	files, err := st.Files(nil)
	if err != nil || len(files) != 2 || filepath.Base(files[0]) != "a.jsonl" || filepath.Base(files[1]) != "b.jsonl" {
		t.Fatalf("files = %v err=%v", files, err)
	}
	files, _ = st.Files(func(name string) bool { return name == "b.jsonl" })
	if len(files) != 1 || filepath.Base(files[0]) != "b.jsonl" {
		t.Fatalf("filtered files = %v", files)
	}
	empty, _ := New(filepath.Join(dir, "nope"))
	if files, err := empty.Files(nil); err != nil || files != nil {
		t.Fatalf("missing dir: %v %v", files, err)
	}
}
