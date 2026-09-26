// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Command pdoom-api serves the public read API (api/openapi.yaml). It reads
// the current promoted release and its snapshot, re-checks
// data/releases/CURRENT every few seconds so that a promotion is picked up
// without a restart, and shuts down gracefully on SIGINT/SIGTERM. It has no
// administrative surface: releases change only through pdoomctl.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/api"
	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
)

const usage = `pdoom-api — p(DOOM) public read API

Usage:
  pdoom-api [--addr :8080] [--data-dir DIR] [--method-dir DIR] [--cors-origins LIST] [--trust-proxy] [--routes]

Environment:
  PDOOM_DATA_DIR      data directory (default ./data)
  PDOOM_METHOD_DIR    methodology markdown directory (default <data>/../docs/method)
  PDOOM_CORS_ORIGINS  comma-separated origins allowed to POST cross-origin
  PDOOM_TRUST_PROXY   "1" to trust the reverse proxy: client address from X-Forwarded-For,
                      request scheme from X-Forwarded-Proto (same-origin detection of POSTs)
  PDOOM_ADDR          listen address (default :8080)
`

type options struct {
	addr        string
	dataDir     string
	methodDir   string
	corsOrigins []string
	trustProxy  bool
	routes      bool
}

// parseOptions reads flags and environment. Flags win over environment.
func parseOptions(args []string, getenv func(string) string, stderr io.Writer) (options, error) {
	fs := flag.NewFlagSet("pdoom-api", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, usage) }
	var o options
	fs.StringVar(&o.addr, "addr", "", "listen address (default $PDOOM_ADDR or :8080)")
	fs.StringVar(&o.dataDir, "data-dir", "", "data directory (default $PDOOM_DATA_DIR or ./data)")
	fs.StringVar(&o.methodDir, "method-dir", "", "methodology markdown directory (default $PDOOM_METHOD_DIR or <data>/../docs/method)")
	cors := fs.String("cors-origins", "", "comma-separated origins allowed to POST cross-origin (default $PDOOM_CORS_ORIGINS)")
	fs.BoolVar(&o.trustProxy, "trust-proxy", false, "trust the reverse proxy: client address from X-Forwarded-For, request scheme from X-Forwarded-Proto (default $PDOOM_TRUST_PROXY=1)")
	fs.BoolVar(&o.routes, "routes", false, "print the route table and exit")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	if fs.NArg() > 0 {
		return o, fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	if o.addr == "" {
		o.addr = getenv("PDOOM_ADDR")
	}
	if o.addr == "" {
		o.addr = ":8080"
	}
	if o.methodDir == "" {
		o.methodDir = getenv("PDOOM_METHOD_DIR")
	}
	if *cors == "" {
		*cors = getenv("PDOOM_CORS_ORIGINS")
	}
	o.corsOrigins = splitList(*cors)
	if !o.trustProxy && strings.TrimSpace(getenv("PDOOM_TRUST_PROXY")) == "1" {
		o.trustProxy = true
	}
	return o, nil
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (o options) deps(logger *slog.Logger) api.Deps {
	return api.Deps{
		Paths:       config.Resolve(o.dataDir, ""),
		Logger:      logger,
		MethodDir:   o.methodDir,
		CORSOrigins: o.corsOrigins,
		TrustProxy:  o.trustProxy,
	}
}

func main() {
	os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}

func run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, getenv, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	logger := slog.New(slog.NewJSONHandler(stdout, nil))
	if opts.routes {
		for _, r := range api.Routes(opts.deps(nil)) {
			fmt.Fprintln(stdout, r)
		}
		return 0
	}
	ln, err := net.Listen("tcp", opts.addr)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := serve(ctx, opts, ln, logger); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

// serve runs the HTTP server on ln until ctx is cancelled, then drains
// in-flight requests for up to shutdownTimeout.
func serve(ctx context.Context, opts options, ln net.Listener, logger *slog.Logger) error {
	const shutdownTimeout = 10 * time.Second
	handler := api.NewHandler(opts.deps(logger))
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 * 1024,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	logger.Info("pdoom-api listening", "addr", ln.Addr().String(), "data_dir", opts.dataDir, "trust_proxy", opts.trustProxy, "cors_origins", len(opts.corsOrigins))
	select {
	case err := <-errc:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}
	logger.Info("pdoom-api shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		return fmt.Errorf("shutdown: %w", err)
	}
	<-errc
	return nil
}
