// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Command pdoom-ingest runs the bounded, allowlisted ingestion pipeline. It can
// only append items to the review queue (data/review-queue/<date>-ingest.jsonl);
// it never touches snapshots, releases, weights or configuration.
//
//	pdoom-ingest --now 2026-09-26T00:00:00Z                       # plan only (default: dry run, no network)
//	pdoom-ingest --now ... --allow-network                         # fetch, parse, print what would be queued
//	pdoom-ingest --now ... --allow-network --dry-run=false         # fetch and append to the review queue
//	pdoom-ingest --now ... --allow-network --check-robots          # print robots.txt status per source (no writes)
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
	"github.com/1a-li-lu-le-lo/pdoom/internal/fetch"
	"github.com/1a-li-lu-le-lo/pdoom/internal/observability"
	"github.com/1a-li-lu-le-lo/pdoom/internal/review"
	"github.com/1a-li-lu-le-lo/pdoom/internal/robots"
	"github.com/1a-li-lu-le-lo/pdoom/internal/sources"
)

// Exit codes.
const (
	exitOK     = 0
	exitRun    = 1 // a run-time failure (queue write, robots check transport)
	exitConfig = 2 // configuration or flag errors
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("pdoom-ingest", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "config/sources.json", "path to the source allowlist")
	dataDir := fs.String("data-dir", "", "data directory (default: PDOOM_DATA_DIR or ./data)")
	dryRun := fs.Bool("dry-run", true, "print what would be fetched and queued without writing the review queue")
	allowNetwork := fs.Bool("allow-network", false, "permit network access (without it the run is a plan only)")
	only := fs.String("source", "", "run a single source id")
	since := fs.String("since", "", "skip documents published before this date (YYYY-MM-DD)")
	now := fs.String("now", "", "RFC 3339 UTC timestamp used for retrieved_at and queue file names (required)")
	checkRobots := fs.Bool("check-robots", false, "fetch robots.txt for every configured source and print the status; never writes")
	jsonOut := fs.Bool("json", false, "print the run summary as JSON on stdout")
	verbose := fs.Bool("verbose", false, "debug logging on stderr")
	timeout := fs.Duration("timeout", 10*time.Minute, "overall run timeout")
	if err := fs.Parse(args); err != nil {
		return exitConfig
	}
	if *now == "" {
		fmt.Fprintln(stderr, "pdoom-ingest: --now is required (the pipeline never reads the clock; pass e.g. --now "+time.Now().UTC().Format(time.RFC3339)+")")
		return exitConfig
	}
	if _, err := time.Parse(time.RFC3339, *now); err != nil {
		fmt.Fprintf(stderr, "pdoom-ingest: --now: %v\n", err)
		return exitConfig
	}
	cfg, err := sources.Load(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "pdoom-ingest: %v\n", err)
		return exitConfig
	}
	if *only != "" {
		if _, ok := cfg.ByID(*only); !ok {
			fmt.Fprintf(stderr, "pdoom-ingest: --source %q is not in %s\n", *only, *configPath)
			return exitConfig
		}
	}
	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	logger := observability.NewLogger(stderr, level, false)
	metrics := observability.New()
	paths := config.Resolve(*dataDir, "")
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if *checkRobots {
		if !*allowNetwork {
			fmt.Fprintln(stderr, "pdoom-ingest: --check-robots needs --allow-network")
			return exitConfig
		}
		return runCheckRobots(ctx, cfg, *only, stdout, stderr, metrics, fetch.Options{UserAgent: robots.UserAgent, MinDelay: fetch.DefaultMinDelay})
	}

	p := &Pipeline{Config: cfg, Queue: review.NewQueue(paths.Review), Metrics: metrics, Logger: logger, Out: stdout, Now: *now, DryRun: *dryRun, Since: *since, OnlyID: *only}
	if *allowNetwork {
		client, err := fetch.New(fetch.Options{AllowedHosts: cfg.AllowedHosts(), UserAgent: robots.UserAgent, MinDelay: fetch.DefaultMinDelay})
		if err != nil {
			fmt.Fprintf(stderr, "pdoom-ingest: %v\n", err)
			return exitConfig
		}
		client.SetPolicy(robots.NewCache(client.RobotsDoer()))
		p.Fetcher = client
		logger.Info("network enabled", "allowed_hosts", client.AllowedHosts(), "dry_run", *dryRun)
	} else {
		banner := stdout
		if *jsonOut {
			banner = stderr
		}
		fmt.Fprintln(banner, "plan only: no network access (pass --allow-network to fetch; --dry-run=false to write the review queue)")
	}
	if *jsonOut {
		p.Out = io.Discard
	}
	sum, err := p.Run(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "pdoom-ingest: %v\n", err)
		return exitRun
	}
	if *jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(sum); err != nil {
			fmt.Fprintf(stderr, "pdoom-ingest: %v\n", err)
			return exitRun
		}
	}
	return exitOK
}

// runCheckRobots prints the robots.txt decision for every configured source.
// It reads robots.txt only and never modifies the configuration: the operator
// copies the printed status into config/sources.json by hand. The allowlist
// is every configured host (allowed or not), since the point is to find out
// whether a host may be enabled; opts carries the transport (tests inject a
// resolver, dialer and certificate pool). Exit 1 when any host fails closed.
func runCheckRobots(ctx context.Context, cfg *sources.Config, only string, stdout, stderr io.Writer, metrics *observability.Registry, opts fetch.Options) int {
	opts.AllowedHosts = cfg.AllHosts()
	if opts.UserAgent == "" {
		opts.UserAgent = robots.UserAgent
	}
	client, err := fetch.New(opts)
	if err != nil {
		fmt.Fprintf(stderr, "pdoom-ingest: %v\n", err)
		return exitConfig
	}
	cache := robots.NewCache(client.RobotsDoer())
	fmt.Fprintf(stdout, "%-28s %-32s %-11s %-6s %s\n", "source", "host", "robots", "delay", "note")
	failures := 0
	for _, src := range cfg.Sources {
		if only != "" && src.ID != only {
			continue
		}
		metrics.Inc(observability.CounterFetches, 1)
		ch := cache.Check(ctx, src.URL)
		note := ""
		if ch.Err != nil {
			note = ch.Err.Error()
			failures++
			metrics.Inc(observability.CounterRobotsDenials, 1)
		} else if ch.StatusCode != 0 {
			note = fmt.Sprintf("robots.txt HTTP %d", ch.StatusCode)
		}
		fmt.Fprintf(stdout, "%-28s %-32s %-11s %-6s %s\n", src.ID, src.Host(), ch.Status, ch.CrawlDelay, note)
	}
	fmt.Fprintf(stdout, "summary %s\n", metrics.Snapshot().String())
	fmt.Fprintln(stdout, "config/sources.json was not modified; copy the results into robots_status / robots_checked_at by hand.")
	if failures > 0 {
		return exitRun
	}
	return exitOK
}
