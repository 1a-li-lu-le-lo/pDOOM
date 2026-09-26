// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Command pdoom-model runs the deterministic model on a snapshot and prints the
// result as JSON. It writes nothing: candidates and releases are created only
// through pdoomctl.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/publishing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

func main() {
	snapDir := flag.String("snapshot", "", "snapshot directory (required)")
	generatedAt := flag.String("generated-at", "", "RFC 3339 timestamp (required)")
	dataDir := flag.String("data-dir", "", "data directory for the previous release lookup")
	previous := flag.String("previous", "", "previous release id (default: current release, if any)")
	flag.Parse()
	if *snapDir == "" || *generatedAt == "" {
		fmt.Fprintln(os.Stderr, "usage: pdoom-model --snapshot DIR --generated-at RFC3339 [--previous REL]")
		os.Exit(2)
	}
	paths := config.Resolve(*dataDir, "")
	s, err := snapshot.Load(*snapDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if problems := snapshot.Validate(s, paths.SchemaDir); snapshot.HasErrors(problems) {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, p)
		}
		os.Exit(1)
	}
	var prev *schema.Release
	if *previous == "" {
		if cur, err := publishing.CurrentReleaseID(paths); err == nil {
			*previous = cur
		}
	}
	if *previous != "" {
		if prev, err = publishing.LoadRelease(paths, *previous); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	res, err := model.Run(s, model.RunOptions{PreviousRelease: prev, GeneratedAt: *generatedAt})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	raw, err := schema.CanonicalJSONIndent(res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout.Write(raw)
	if errs := model.CheckInvariants(res); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "invariant:", e)
		}
		os.Exit(1)
	}
}
