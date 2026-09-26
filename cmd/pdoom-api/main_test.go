// Copyright NU Cybernetics. p(DOOM) — research prototype.

package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestParseOptionsFlagsWinOverEnv(t *testing.T) {
	env := map[string]string{"PDOOM_ADDR": ":9000", "PDOOM_CORS_ORIGINS": "https://a.example, https://b.example", "PDOOM_TRUST_PROXY": "1", "PDOOM_METHOD_DIR": "/m"}
	getenv := func(k string) string { return env[k] }
	o, err := parseOptions(nil, getenv, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if o.addr != ":9000" || !o.trustProxy || o.methodDir != "/m" || len(o.corsOrigins) != 2 || o.corsOrigins[1] != "https://b.example" {
		t.Fatalf("env not applied: %+v", o)
	}
	o, err = parseOptions([]string{"--addr", "127.0.0.1:1", "--cors-origins", "https://c.example", "--data-dir", "/d"}, getenv, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if o.addr != "127.0.0.1:1" || len(o.corsOrigins) != 1 || o.corsOrigins[0] != "https://c.example" || o.dataDir != "/d" {
		t.Fatalf("flags did not win: %+v", o)
	}
	if _, err := parseOptions([]string{"extra"}, getenv, io.Discard); err == nil {
		t.Fatal("positional argument must be rejected")
	}
	o, err = parseOptions(nil, func(string) string { return "" }, io.Discard)
	if err != nil || o.addr != ":8080" || o.trustProxy {
		t.Fatalf("defaults: %+v err %v", o, err)
	}
}

func TestRoutesFlagPrintsTable(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--routes"}, func(string) string { return "" }, &out, &errOut); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut.String())
	}
	for _, want := range []string{"GET /v1/meter", "POST /v1/scenario-lab/evaluate", "POST /v1/submissions/sources", "GET /healthz"} {
		if !strings.Contains(out.String(), want+"\n") {
			t.Fatalf("route table lacks %s:\n%s", want, out.String())
		}
	}
	if code := run([]string{"--bogus"}, func(string) string { return "" }, &out, &errOut); code != 2 {
		t.Fatalf("bad flag should exit 2, got %d", code)
	}
}

func TestServeAndGracefulShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	opts := options{dataDir: t.TempDir()}
	go func() { done <- serve(ctx, opts, ln, slog.New(slog.DiscardHandler)) }()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://" + ln.Addr().String() + "/healthz")
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		cancel()
		t.Fatalf("healthz %d", resp.StatusCode)
	}
	resp, err = client.Get("http://" + ln.Addr().String() + "/readyz")
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		cancel()
		t.Fatalf("readyz without a release should be 503, got %d", resp.StatusCode)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("server did not shut down")
	}
}
