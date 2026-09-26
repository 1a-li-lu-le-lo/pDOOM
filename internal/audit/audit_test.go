// Copyright NU Cybernetics. p(DOOM) — research prototype.

package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendAndVerify(t *testing.T) {
	p := filepath.Join(t.TempDir(), "audit.jsonl")
	for i := 0; i < 3; i++ {
		if _, err := Append(p, Event{TS: "2026-09-26T00:00:00Z", Actor: "test", Action: "promote", Subject: "rel-x", Details: map[string]any{"i": i}}); err != nil {
			t.Fatal(err)
		}
	}
	n, err := Verify(p)
	if err != nil || n != 3 {
		t.Fatalf("verify: n=%d err=%v", n, err)
	}
	raw, _ := os.ReadFile(p)
	tampered := strings.Replace(string(raw), `"i":1`, `"i":9`, 1)
	if err := os.WriteFile(p, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(p); err == nil {
		t.Fatal("tampering not detected")
	}
	if _, err := Append(p, Event{TS: "t", Actor: "a", Action: "x"}); err == nil {
		t.Fatal("append to broken chain must fail")
	}
	// Deleting a line breaks the chain too.
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if err := os.WriteFile(p, []byte(lines[0]+"\n"+lines[2]+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(p); err == nil {
		t.Fatal("deletion not detected")
	}
}
