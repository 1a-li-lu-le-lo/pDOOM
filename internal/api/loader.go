// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
	"github.com/1a-li-lu-le-lo/pdoom/internal/publishing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// state is one immutable view of the promoted data: the current release, the
// snapshot it was computed from, and every other published release that could
// be read (for /v1/meter/history and /v1/releases). A new state is built
// whenever CURRENT changes; handlers never mutate it.
type state struct {
	release   *schema.Release
	snapshot  *snapshot.Snapshot
	releases  []schema.ReleaseSummary
	history   []Headline
	byID      map[string]*schema.Release
	currentID string
	loadedAt  time.Time
}

// meta builds the release metadata embedded in every data response.
func (st *state) meta() Meta {
	return metaOf(st.release.Manifest)
}

func metaOf(m schema.ReleaseManifest) Meta {
	return Meta{
		ReleaseID:     m.ReleaseID,
		DataSnapshot:  m.DataSnapshot,
		DataCutoff:    m.SourceCutoff,
		GeneratedAt:   m.GeneratedAt,
		Published:     m.Published,
		ModelVersions: nonNil(m.ModelVersions),
		Limitations:   nonNil(m.KnownLimitations),
	}
}

// loader reads the current release lazily and re-checks CURRENT at most once
// per interval. A failed reload keeps the previous state (and logs).
type loader struct {
	paths    config.Paths
	log      *slog.Logger
	now      func() time.Time
	interval time.Duration

	mu        sync.Mutex
	st        *state
	lastCheck time.Time
	lastMod   time.Time
	lastSize  int64
	lastErr   error
}

func newLoader(paths config.Paths, log *slog.Logger, now func() time.Time, interval time.Duration) *loader {
	return &loader{paths: paths, log: log, now: now, interval: interval}
}

// current returns the loaded state, refreshing it when CURRENT changed and the
// refresh interval elapsed. It returns an error only when no state has ever
// been loaded.
func (l *loader) current() (*state, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if l.st != nil && !l.lastCheck.IsZero() && now.Sub(l.lastCheck) < l.interval && now.Sub(l.lastCheck) >= 0 {
		return l.st, nil
	}
	l.lastCheck = now
	fi, err := os.Stat(l.paths.Current)
	if err != nil {
		l.lastErr = fmt.Errorf("api: stat CURRENT: %w", err)
		if l.st != nil {
			l.log.Warn("CURRENT unreadable; keeping loaded release", "error", err.Error())
			return l.st, nil
		}
		return nil, l.lastErr
	}
	id, err := publishing.CurrentReleaseID(l.paths)
	if err != nil {
		l.lastErr = fmt.Errorf("api: read CURRENT: %w", err)
		if l.st != nil {
			return l.st, nil
		}
		return nil, l.lastErr
	}
	if l.st != nil && id == l.st.currentID && fi.ModTime().Equal(l.lastMod) && fi.Size() == l.lastSize {
		return l.st, nil
	}
	st, err := l.load(now)
	if err != nil {
		l.lastErr = err
		l.log.Error("release reload failed", "error", err.Error())
		if l.st != nil {
			return l.st, nil
		}
		return nil, err
	}
	l.lastErr = nil
	l.lastMod = fi.ModTime()
	l.lastSize = fi.Size()
	if l.st == nil || l.st.currentID != st.currentID {
		l.log.Info("release loaded", "release_id", st.currentID, "data_snapshot", st.release.Manifest.DataSnapshot, "published_releases", len(st.releases))
	}
	l.st = st
	return st, nil
}

// status reports the loaded release id (empty when none) and the last error.
func (l *loader) status() (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.st == nil {
		if l.lastErr == nil {
			return "", errors.New("no release loaded yet")
		}
		return "", l.lastErr
	}
	return l.st.currentID, nil
}

// load reads the current release and its snapshot, which must both be valid,
// then adds every other published release directory that can be read. A
// damaged or unpublished sibling directory is skipped and logged rather than
// taking the current release offline: publishing.Promote copies files before
// it writes the published manifest and CURRENT, so a crash mid-promotion
// leaves exactly such a directory behind.
func (l *loader) load(now time.Time) (*state, error) {
	paths := l.paths
	rel, err := publishing.LoadCurrentRelease(paths)
	if err != nil {
		return nil, fmt.Errorf("api: load current release: %w", err)
	}
	if rel.Manifest.Published == nil {
		return nil, fmt.Errorf("api: release %s is not marked published; refusing to serve it", rel.Manifest.ReleaseID)
	}
	snapDir := filepath.Join(paths.Snapshots, rel.Manifest.DataSnapshot)
	snap, err := snapshot.Load(snapDir)
	if err != nil {
		return nil, fmt.Errorf("api: load snapshot %s: %w", rel.Manifest.DataSnapshot, err)
	}
	st := &state{release: rel, snapshot: snap, currentID: rel.Manifest.ReleaseID, loadedAt: now, byID: map[string]*schema.Release{}}
	for _, r := range scanPublishedReleases(paths, l.log, rel) {
		sum := summaryOf(r, r.Manifest.ReleaseID == st.currentID)
		st.releases = append(st.releases, sum)
		st.history = append(st.history, headlineOf(r, sum.IsCurrent))
		st.byID[r.Manifest.ReleaseID] = r
	}
	st.releases = nonNil(st.releases)
	st.history = nonNil(st.history)
	return st, nil
}

// summaryOf builds the /v1/releases row of a release.
func summaryOf(r *schema.Release, isCurrent bool) schema.ReleaseSummary {
	m := r.Manifest
	return schema.ReleaseSummary{ReleaseID: m.ReleaseID, DataSnapshot: m.DataSnapshot, ModelVersions: nonNil(m.ModelVersions), Published: m.Published, Superseded: m.Superseded, IsCurrent: isCurrent, EditorialRiskLevel: m.EditorialRiskLevel, UncertaintyScore: m.UncertaintyScore}
}

// scanPublishedReleases returns every release directory under data/releases
// that is well-formed and marked published, in id order. cur, when non-nil,
// is reused instead of being read again. Directories that cannot be read, or
// whose manifest is unpublished or names a different release id, are skipped
// with a warning; the public API never exposes them (build-spec §0.7).
func scanPublishedReleases(paths config.Paths, log *slog.Logger, cur *schema.Release) []*schema.Release {
	entries, err := os.ReadDir(paths.Releases)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Warn("release directory unreadable", "error", err.Error())
		}
		return nil
	}
	var out []*schema.Release
	for _, e := range entries {
		id := e.Name()
		if !e.IsDir() || !reReleaseID.MatchString(id) {
			continue
		}
		var r *schema.Release
		if cur != nil && cur.Manifest.ReleaseID == id {
			r = cur
		} else if r, err = publishing.LoadRelease(paths, id); err != nil {
			log.Warn("skipping unreadable release directory", "release_id", id, "error", err.Error())
			continue
		}
		if r.Manifest.ReleaseID != id {
			log.Warn("skipping release directory whose manifest names another release", "release_id", id, "manifest_release_id", r.Manifest.ReleaseID)
			continue
		}
		if r.Manifest.Published == nil {
			log.Warn("skipping release directory that is not marked published", "release_id", id)
			continue
		}
		out = append(out, r)
	}
	return out
}

func nonNil[T any](xs []T) []T {
	if xs == nil {
		return []T{}
	}
	return xs
}
