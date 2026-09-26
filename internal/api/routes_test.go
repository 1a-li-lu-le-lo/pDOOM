// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
)

var (
	reSpecPath   = regexp.MustCompile(`^  (/[^\s:]*):\s*$`)
	reSpecMethod = regexp.MustCompile(`^    (get|post|put|patch|delete|head|options|trace):\s*$`)
)

// specRoutes parses api/openapi.yaml minimally: path keys at two-space indent
// under "paths:" and method keys at four-space indent below each path.
func specRoutes(t *testing.T) []string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	inPaths := false
	current := ""
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) > 0 && line[0] != ' ' {
			inPaths = strings.HasPrefix(line, "paths:")
			continue
		}
		if !inPaths {
			continue
		}
		if m := reSpecPath.FindStringSubmatch(line); m != nil {
			current = m[1]
			continue
		}
		if m := reSpecMethod.FindStringSubmatch(line); m != nil && current != "" {
			out = append(out, strings.ToUpper(m[1])+" "+current)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return out
}

func TestRouteTableMatchesOpenAPI(t *testing.T) {
	spec := specRoutes(t)
	if len(spec) == 0 {
		t.Fatal("no routes parsed from api/openapi.yaml")
	}
	registered := Routes(Deps{Paths: config.Resolve(t.TempDir(), "")})
	specSet := map[string]bool{}
	for _, r := range spec {
		specSet[r] = true
	}
	regSet := map[string]bool{}
	for _, r := range registered {
		regSet[r] = true
	}
	for _, r := range spec {
		if !regSet[r] {
			t.Errorf("documented in api/openapi.yaml but not registered: %s", r)
		}
	}
	for _, r := range registered {
		if !specSet[r] {
			t.Errorf("registered but missing from api/openapi.yaml: %s", r)
		}
	}
	// No administrative surface: every mutating route is one of the two
	// documented public write endpoints.
	for _, r := range registered {
		if strings.HasPrefix(r, "GET ") {
			continue
		}
		switch r {
		case "POST /v1/scenario-lab/evaluate", "POST /v1/submissions/sources", "POST /v1/submissions/corrections":
		default:
			t.Errorf("unexpected non-GET route %s (admin operations are CLI-only)", r)
		}
		if strings.Contains(r, "admin") || strings.Contains(r, "release") || strings.Contains(r, "promote") {
			t.Errorf("route %s looks administrative", r)
		}
	}
	if len(spec) != len(registered) {
		t.Fatalf("route counts differ: spec %d, registered %d", len(spec), len(registered))
	}
}

func TestOpenAPIDeclaresNoAdminEndpoints(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.HasPrefix(text, "# Copyright NU Cybernetics. p(DOOM) — research prototype.") {
		t.Fatal("openapi.yaml lacks the copyright line")
	}
	if !strings.Contains(text, "openapi: 3.1.0") {
		t.Fatal("openapi.yaml must declare OpenAPI 3.1")
	}
	if !strings.Contains(text, "CLI-only") || !strings.Contains(text, "ADR-004") {
		t.Fatal("openapi.yaml must state that administrative operations are CLI-only (ADR-004)")
	}
	for _, forbidden := range []string{"PDUM", "pDOOM", "PDOOM)"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("openapi.yaml contains the misspelt brand %q", forbidden)
		}
	}
}
