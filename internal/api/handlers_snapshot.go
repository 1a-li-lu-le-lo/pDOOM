// Copyright NU Cybernetics. p(DOOM) — research prototype.
package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/1a-li-lu-le-lo/pdoom/internal/publishing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// SnapshotResponse is the body of GET /v1/snapshot: the sealed data snapshot
// the current release was computed from, entity file by entity file. It is the
// same public data the web app exports as snapshot.json and what @pdoom/sdk's
// HTTP source reads to mirror the file source.
type SnapshotResponse struct {
	Meta     Meta         `json:"meta"`
	Snapshot SnapshotBody `json:"snapshot"`
}

// SnapshotBody mirrors the snapshot directory with snake_case keys that match
// the TypeScript Snapshot interface in packages/sdk.
type SnapshotBody struct {
	Manifest           snapshot.Manifest          `json:"manifest"`
	Definitions        []schema.Definition        `json:"definitions"`
	Sources            []schema.Source            `json:"sources"`
	Claims             []schema.Claim             `json:"claims"`
	Forecasts          []schema.Forecast          `json:"forecasts"`
	Benchmarks         []schema.Benchmark         `json:"benchmarks"`
	BenchmarkResults   []schema.BenchmarkResult   `json:"benchmark_results"`
	Incidents          []schema.Incident          `json:"incidents"`
	Scenarios          []schema.Scenario          `json:"scenarios"`
	ScenarioEdges      []schema.ScenarioEdge      `json:"scenario_edges"`
	Drivers            []schema.Driver            `json:"drivers"`
	DriverObservations []schema.DriverObservation `json:"driver_observations"`
	Interventions      []schema.Intervention      `json:"interventions"`
	Organizations      []schema.Organization      `json:"organizations"`
	Actions            []schema.Action            `json:"actions"`
	ModelSpec          schema.ModelSpec           `json:"model_spec"`
}

func (s *server) handleSnapshot(w http.ResponseWriter, r *http.Request, st *state) {
	sn := st.snapshot
	s.ok(w, r, st, SnapshotResponse{Meta: st.meta(), Snapshot: SnapshotBody{
		Manifest:           sn.Manifest,
		Definitions:        nonNil(sn.Definitions),
		Sources:            nonNil(sn.Sources),
		Claims:             nonNil(sn.Claims),
		Forecasts:          nonNil(sn.Forecasts),
		Benchmarks:         nonNil(sn.Benchmarks),
		BenchmarkResults:   nonNil(sn.BenchmarkResults),
		Incidents:          nonNil(sn.Incidents),
		Scenarios:          nonNil(sn.Scenarios),
		ScenarioEdges:      nonNil(sn.ScenarioEdges),
		Drivers:            nonNil(sn.Drivers),
		DriverObservations: nonNil(sn.DriverObservations),
		Interventions:      nonNil(sn.Interventions),
		Organizations:      nonNil(sn.Organizations),
		Actions:            nonNil(sn.Actions),
		ModelSpec:          sn.ModelSpec,
	}})
}

// releaseDocuments are the Markdown files a promoted release ships beside its
// JSON files. Only these two names are ever read from a release directory.
var releaseDocuments = map[string]bool{"changelog.md": true, "model-card.md": true}

// handleReleaseDocument serves GET /v1/releases/{id}/{doc}: the changelog or
// the model card of a published release, as Markdown.
func (s *server) handleReleaseDocument(w http.ResponseWriter, r *http.Request, st *state) {
	id := r.PathValue("id")
	doc := r.PathValue("doc")
	if !reReleaseID.MatchString(id) || !releaseDocuments[doc] {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no document "+strconv.Quote(doc)+" for release "+strconv.Quote(id))
		return
	}
	published := false
	if st != nil {
		_, published = st.byID[id]
	} else if rel, err := publishing.LoadRelease(s.deps.Paths, id); err == nil && rel.Manifest.Published != nil && rel.Manifest.ReleaseID == id {
		published = true
	}
	if !published {
		s.writeError(w, r, http.StatusNotFound, "not_found", "no release with id "+strconv.Quote(id))
		return
	}
	raw, err := os.ReadFile(filepath.Join(s.deps.Paths.Releases, id, doc))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.writeError(w, r, http.StatusNotFound, "not_found", "release "+strconv.Quote(id)+" has no "+doc)
			return
		}
		s.log.Error("read release document", "release_id", id, "doc", doc, "error", err.Error())
		s.writeError(w, r, http.StatusInternalServerError, "internal", "release document could not be read")
		return
	}
	s.serveBody(w, r, raw, contentTypeMarkdown)
}
