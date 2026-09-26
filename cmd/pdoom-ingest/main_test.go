// Copyright NU Cybernetics. p(DOOM) — research prototype.

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/fetch"
	"github.com/1a-li-lu-le-lo/pdoom/internal/observability"
	"github.com/1a-li-lu-le-lo/pdoom/internal/parsing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/sources"
)

// parsingDoc lets the pipeline tests build documents without importing parsing twice.
type parsingDoc = parsing.Document

func writeConfig(t *testing.T) string {
	t.Helper()
	raw, err := json.Marshal(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "sources.json")
	if err := os.WriteFile(p, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunRequiresNow(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"--config", writeConfig(t)}, &out, &errb); code != exitConfig {
		t.Fatalf("exit %d, stderr %s", code, errb.String())
	}
	if !strings.Contains(errb.String(), "--now is required") {
		t.Fatalf("stderr: %s", errb.String())
	}
	errb.Reset()
	if code := run([]string{"--config", writeConfig(t), "--now", "soon"}, &out, &errb); code != exitConfig {
		t.Fatalf("exit %d", code)
	}
}

func TestRunConfigErrorsExitNonZero(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"--config", filepath.Join(t.TempDir(), "missing.json"), "--now", testNow}, &out, &errb); code != exitConfig {
		t.Fatalf("exit %d", code)
	}
	bad := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(bad, []byte(`{"version":2,"generated_at":"x","sources":[]}`), 0o644)
	errb.Reset()
	if code := run([]string{"--config", bad, "--now", testNow}, &out, &errb); code != exitConfig || !strings.Contains(errb.String(), "invalid config") {
		t.Fatalf("exit %d stderr %s", code, errb.String())
	}
	errb.Reset()
	if code := run([]string{"--config", writeConfig(t), "--now", testNow, "--source", "nope"}, &out, &errb); code != exitConfig {
		t.Fatalf("unknown source exit %d", code)
	}
	if code := run([]string{"--no-such-flag"}, &out, &errb); code != exitConfig {
		t.Fatalf("bad flag exit %d", code)
	}
	errb.Reset()
	if code := run([]string{"--config", writeConfig(t), "--now", testNow, "--check-robots"}, &out, &errb); code != exitConfig || !strings.Contains(errb.String(), "--allow-network") {
		t.Fatalf("check-robots without network exit %d stderr %s", code, errb.String())
	}
}

func TestRunPlanOnlyIsDefault(t *testing.T) {
	var out, errb bytes.Buffer
	data := t.TempDir()
	code := run([]string{"--config", writeConfig(t), "--now", testNow, "--data-dir", data}, &out, &errb)
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, errb.String())
	}
	s := out.String()
	if !strings.Contains(s, "plan only") || !strings.Contains(s, "plan   feed-a") || !strings.Contains(s, "skip   manual-c") || !strings.Contains(s, "summary ") {
		t.Fatalf("stdout:\n%s", s)
	}
	if entries, _ := os.ReadDir(data); len(entries) != 0 {
		t.Fatalf("plan must not write into the data dir: %v", entries)
	}
	out.Reset()
	code = run([]string{"--config", writeConfig(t), "--now", testNow, "--data-dir", data, "--json"}, &out, &errb)
	if code != exitOK {
		t.Fatalf("exit %d", code)
	}
	var sum Summary
	if err := json.Unmarshal(out.Bytes(), &sum); err != nil {
		t.Fatalf("json summary: %v\n%s", err, out.String())
	}
	if sum.Network || sum.Queued != 0 || len(sum.Sources) != 5 {
		t.Fatalf("summary = %+v", sum)
	}
}

// mapResolver is a deterministic fake DNS for the robots check harness.
type mapResolver map[string][]string

func (m mapResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	ips, ok := m[host]
	if !ok {
		return nil, fmt.Errorf("no such host %q", host)
	}
	var out []net.IPAddr
	for _, s := range ips {
		out = append(out, net.IPAddr{IP: net.ParseIP(s)})
	}
	return out, nil
}

// robotsHarness serves one robots.txt per configured host over TLS. The
// client's resolver answers with public addresses, its dialer is pointed at
// the listener and every host is verified against the test certificate's
// name, so the whole SafeClient guard stays active.
func robotsHarness(t *testing.T, cfg *sources.Config) (fetch.Options, *int32) {
	t.Helper()
	var requests int32
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/robots.txt" {
			t.Errorf("robots check fetched %s on %s", r.URL.Path, r.Host)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		switch r.Host {
		case "feeds.example.com":
			fmt.Fprint(w, "User-agent: *\nAllow: /\nCrawl-delay: 2\n")
		case "api.example.org":
			http.NotFound(w, r)
		case "denied.example.com":
			fmt.Fprint(w, "User-agent: pdoom-ingest\nDisallow: /\n")
		case "broken.example.com":
			http.Error(w, "down", http.StatusInternalServerError)
		default:
			http.Error(w, "forbidden", http.StatusForbidden)
		}
	})
	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	resolver := mapResolver{}
	for i, h := range cfg.AllHosts() {
		resolver[h] = []string{fmt.Sprintf("93.184.216.%d", 10+i)}
	}
	return fetch.Options{
		Resolver: resolver,
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, srv.Listener.Addr().String())
		},
		TLSConfig: &tls.Config{RootCAs: pool, ServerName: "example.com", MinVersion: tls.VersionTLS12},
		Sleep:     func(context.Context, time.Duration) error { return nil },
	}, &requests
}

// checkRows parses the table printed by runCheckRobots: source → fields.
func checkRows(out string) map[string][]string {
	rows := map[string][]string{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 4 || f[0] == "source" || f[0] == "summary" || strings.HasPrefix(f[0], "config/") {
			continue
		}
		rows[f[0]] = f
	}
	return rows
}

func TestRunCheckRobots(t *testing.T) {
	cfgPath := writeConfig(t)
	before, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := sources.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	opts, requests := robotsHarness(t, cfg)
	var out, errb bytes.Buffer
	code := runCheckRobots(context.Background(), cfg, "", &out, &errb, observability.New(), opts)
	if code != exitRun {
		t.Fatalf("two hosts fail closed, want exit %d, got %d\nstdout:\n%s\nstderr:\n%s", exitRun, code, out.String(), errb.String())
	}
	rows := checkRows(out.String())
	if len(rows) != 5 {
		t.Fatalf("want one row per configured source, got %d:\n%s", len(rows), out.String())
	}
	want := map[string][3]string{
		"feed-a":   {"feeds.example.com", "allowed", "2s"},
		"adv-b":    {"api.example.org", "allowed", "0s"},
		"manual-c": {"data.example.net", "unknown", "0s"},
		"broken-d": {"broken.example.com", "unknown", "0s"},
		"denied-e": {"denied.example.com", "disallowed", "0s"},
	}
	for id, w := range want {
		r, ok := rows[id]
		if !ok {
			t.Fatalf("no row for %s:\n%s", id, out.String())
		}
		if r[1] != w[0] || r[2] != w[1] || r[3] != w[2] {
			t.Errorf("%s: row = %v, want host %s status %s delay %s", id, r, w[0], w[1], w[2])
		}
	}
	note := func(id string) string { return strings.Join(rows[id][4:], " ") }
	if !strings.Contains(note("adv-b"), "HTTP 404") || !strings.Contains(note("manual-c"), "HTTP 403") || !strings.Contains(note("broken-d"), "HTTP 500") {
		t.Errorf("notes: adv-b=%q manual-c=%q broken-d=%q", note("adv-b"), note("manual-c"), note("broken-d"))
	}
	if *requests != 5 {
		t.Errorf("robots.txt requests = %d, want one per host", *requests)
	}
	if !strings.Contains(out.String(), "was not modified") || !strings.Contains(out.String(), "summary ") {
		t.Errorf("output lacks the summary or the not-modified note:\n%s", out.String())
	}
	after, _ := os.ReadFile(cfgPath)
	if !bytes.Equal(before, after) {
		t.Fatal("the robots check must never modify the configuration file")
	}

	// A single healthy source: one row, exit 0. The disallowed host of a
	// disabled source is on the allowlist too (AllHosts): manual-c was checked
	// above; an unconfigured host is refused before any network access.
	out.Reset()
	if code := runCheckRobots(context.Background(), cfg, "feed-a", &out, &errb, observability.New(), opts); code != exitOK {
		t.Fatalf("single healthy source: exit %d\n%s", code, out.String())
	}
	if rows := checkRows(out.String()); len(rows) != 1 || rows["feed-a"] == nil {
		t.Fatalf("rows = %v", rows)
	}
	client, err := fetch.New(func() fetch.Options { o := opts; o.AllowedHosts = cfg.AllHosts(); return o }())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get(context.Background(), "https://other.example.com/robots.txt"); err == nil {
		t.Fatal("a host outside the configuration must be refused")
	}
	// Exit 2 when the client cannot be built (an unusable option).
	bad := opts
	bad.MaxRedirects = 99
	if code := runCheckRobots(context.Background(), cfg, "", &out, &errb, observability.New(), bad); code != exitConfig {
		t.Fatalf("unbuildable client: exit %d", code)
	}
}
