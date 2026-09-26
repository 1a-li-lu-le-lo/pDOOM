// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/config"
	"github.com/1a-li-lu-le-lo/pdoom/internal/model"
	"github.com/1a-li-lu-le-lo/pdoom/internal/publishing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot/fixture"
)

const (
	genAt     = "2026-09-26T12:00:00Z"
	pubAt     = "2026-09-26T13:00:00Z"
	firstRel  = "rel-2026-09-26-001"
	secondRel = "rel-2026-10-01-001"
)

// clock is an adjustable fake clock.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

type env struct {
	root  string
	paths config.Paths
	keyA  string
	keyB  string
	snap  *snapshot.Snapshot
	clock *clock
	logs  *bytes.Buffer
}

// buildDataDir builds a complete data directory through the real safety gate:
// fixture snapshot → model run → candidate → two signed approvals → promotion
// with a heightened-review acknowledgement.
func buildDataDir(t *testing.T) *env {
	t.Helper()
	root := t.TempDir()
	paths := config.Resolve(filepath.Join(root, "data"), "")
	fx := fixture.New()
	snapDir := filepath.Join(paths.Snapshots, fx.Manifest.SnapshotID)
	if err := snapshot.Write(snapDir, fx); err != nil {
		t.Fatal(err)
	}
	snap, err := snapshot.Load(snapDir)
	if err != nil {
		t.Fatal(err)
	}
	_, keyA, err := publishing.GenerateKey(paths.Keys, "reviewer-a")
	if err != nil {
		t.Fatal(err)
	}
	_, keyB, err := publishing.GenerateKey(paths.Keys, "reviewer-b")
	if err != nil {
		t.Fatal(err)
	}
	e := &env{root: root, paths: paths, keyA: keyA, keyB: keyB, snap: snap, clock: &clock{t: time.Date(2026, 9, 26, 14, 0, 0, 0, time.UTC)}, logs: &bytes.Buffer{}}
	e.promote(t, snap, nil, firstRel, "cand-2026-09-26-001", genAt, pubAt)
	// Methodology documents next to the data directory.
	mdir := filepath.Join(root, "docs", "method")
	if err := os.MkdirAll(mdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mdir, "aggregation.md"), []byte("<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->\n# Aggregation method\n\nBody & details.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mdir, "indexes.md"), []byte("No heading here.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mdir, "notes.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}
	return e
}

func (e *env) promote(t *testing.T, snap *snapshot.Snapshot, prev *schema.Release, releaseID, candidateID, generatedAt, publishedAt string) *schema.Release {
	t.Helper()
	res, err := model.Run(snap, model.RunOptions{PreviousRelease: prev, GeneratedAt: generatedAt, CodeCommit: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	var prevID *string
	if prev != nil {
		id := prev.Manifest.ReleaseID
		prevID = &id
	}
	m, err := publishing.BuildManifest(res, snap, publishing.BuildOptions{ReleaseID: releaseID, CandidateID: candidateID, CodeCommit: "abc1234", PreviousReleaseID: prevID})
	if err != nil {
		t.Fatal(err)
	}
	cdir := filepath.Join(e.paths.Candidates, candidateID)
	if _, err := publishing.WriteCandidate(cdir, res, m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []struct{ id, key string }{{"reviewer-a", e.keyA}, {"reviewer-b", e.keyB}} {
		if _, err := publishing.Approve(cdir, k.id, k.key, "none", publishedAt); err != nil {
			t.Fatal(err)
		}
	}
	rel, err := publishing.Promote(e.paths, cdir, publishing.PromoteOptions{PublishedAt: publishedAt, HeightenedReviewAck: "reviewer-b", Actor: "test"})
	if err != nil {
		t.Fatal(err)
	}
	return rel
}

func (e *env) deps() Deps {
	logger := slog.New(slog.NewTextHandler(e.logs, nil))
	return Deps{Paths: e.paths, Logger: logger, Now: e.clock.now, CORSOrigins: []string{"https://lab.example"}}
}

func (e *env) handler(t *testing.T, mutate func(*Deps)) http.Handler {
	t.Helper()
	d := e.deps()
	if mutate != nil {
		mutate(&d)
	}
	return NewHandler(d)
}

func do(h http.Handler, method, target string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, target, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func decode(t *testing.T, rr *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), v); err != nil {
		t.Fatalf("decode %d response: %v\n%s", rr.Code, err, rr.Body.String())
	}
}

func expectError(t *testing.T, rr *httptest.ResponseRecorder, status int, code string) ErrorResponse {
	t.Helper()
	if rr.Code != status {
		t.Fatalf("status %d, want %d: %s", rr.Code, status, rr.Body.String())
	}
	var e ErrorResponse
	decode(t, rr, &e)
	if e.Error != code {
		t.Fatalf("error code %q, want %q (%s)", e.Error, code, e.Detail)
	}
	return e
}

// treeHash hashes every file under dir so tests can prove nothing changed.
func treeHash(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		h.Write([]byte(f))
		h.Write(raw)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func labBody(t *testing.T, overrides map[string]any) []byte {
	t.Helper()
	body := map[string]any{"horizon": "10y", "capability_timeline": 1, "autonomy_growth": 0, "access_level": 0, "safety_progress": -1, "governance_strength": 0,
		"model_security": 0, "open_weight_diffusion": 0, "international_coordination": 0, "incident_frequency": 0, "resilience": 1}
	for k, v := range overrides {
		if v == nil {
			delete(body, k)
		} else {
			body[k] = v
		}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestEveryGetRouteServesJSONWithHeaders(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	paths := []string{"/v1/meter", "/v1/meter/history", "/v1/outcomes", "/v1/horizons", "/v1/drivers", "/v1/scenarios", "/v1/scenarios/S1",
		"/v1/forecasts", "/v1/capabilities", "/v1/incidents", "/v1/safeguards", "/v1/sources", "/v1/sources/src-fixture-survey-2025",
		"/v1/methodology", "/v1/releases", "/v1/releases/" + firstRel, "/v1/definitions", "/v1/organizations", "/v1/actions", "/healthz", "/readyz"}
	for _, p := range paths {
		rr := do(h, http.MethodGet, p, nil, nil)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", p, rr.Code, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("%s: content type %q", p, ct)
		}
		var v map[string]any
		decode(t, rr, &v)
		for k, want := range map[string]string{"X-Content-Type-Options": "nosniff", "Referrer-Policy": "no-referrer", "X-Frame-Options": "DENY"} {
			if got := rr.Header().Get(k); got != want {
				t.Fatalf("%s: header %s = %q", p, k, got)
			}
		}
		if !strings.Contains(rr.Header().Get("Content-Security-Policy"), "default-src 'none'") || rr.Header().Get("Permissions-Policy") == "" {
			t.Fatalf("%s: CSP/Permissions-Policy missing", p)
		}
		if rr.Header().Get("X-Request-ID") == "" || rr.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("%s: request id or CORS missing", p)
		}
		if !strings.HasPrefix(p, "/healthz") && !strings.HasPrefix(p, "/readyz") {
			if rr.Header().Get("ETag") == "" || rr.Header().Get("Cache-Control") != "public, max-age=60" {
				t.Fatalf("%s: ETag/Cache-Control missing: %v", p, rr.Header())
			}
		}
	}
	// Markdown document.
	rr := do(h, http.MethodGet, "/v1/methodology/aggregation", nil, nil)
	if rr.Code != 200 || !strings.HasPrefix(rr.Header().Get("Content-Type"), "text/markdown") || !strings.Contains(rr.Body.String(), "# Aggregation method") {
		t.Fatalf("methodology doc: %d %q %s", rr.Code, rr.Header().Get("Content-Type"), rr.Body.String())
	}
	if rr.Header().Get("ETag") == "" {
		t.Fatal("markdown lacks ETag")
	}
}

func TestMeterShapeAndSeparationOfObjects(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	rr := do(h, http.MethodGet, "/v1/meter", nil, nil)
	if rr.Code != 200 {
		t.Fatalf("%d %s", rr.Code, rr.Body.String())
	}
	var m MeterResponse
	decode(t, rr, &m)
	if m.Meta.ReleaseID != firstRel || m.Meta.DataSnapshot != "snap-1999-01-01-001" || m.Meta.DataCutoff != "2026-09-01" || len(m.Meta.Limitations) == 0 || len(m.Meta.ModelVersions) != 4 {
		t.Fatalf("meta: %+v", m.Meta)
	}
	if len(m.Official) != len(schema.Horizons) {
		t.Fatalf("official rows %d, want one per horizon", len(m.Official))
	}
	for _, o := range m.Official {
		if o.Status != schema.StatusInsufficientlyCalibrate || o.Quantiles != nil || o.Mean != nil || o.Display.Central != "Insufficiently calibrated" {
			t.Fatalf("official object leaks a probability: %+v", o)
		}
		if o.Horizon == "" || len(o.OutcomeSet) != 6 || o.Producer == "" || o.Conditioning == "" {
			t.Fatalf("official object lacks labels: %+v", o)
		}
	}
	if len(m.ExternalAggregates) == 0 || m.ExternalAggregates[0].Status != schema.StatusExternalAggregate || m.ExternalAggregates[0].Quantiles == nil {
		t.Fatalf("external aggregates: %+v", m.ExternalAggregates)
	}
	if len(m.ResearchMode) == 0 || m.ResearchMode[0].Status != schema.StatusResearchMode {
		t.Fatalf("research mode: %+v", m.ResearchMode)
	}
	for _, est := range append(append([]schema.Estimate{}, m.ExternalAggregates...), m.ResearchMode...) {
		if est.Horizon == "" || len(est.OutcomeSet) == 0 || est.Producer == "" || est.LastEvidenceDate != "2026-09-01" {
			t.Fatalf("estimate lacks horizon/outcome_set/producer/cutoff: %+v", est)
		}
	}
	if len(m.Indexes) != 7 {
		t.Fatalf("indexes %d", len(m.Indexes))
	}
	for _, iv := range m.Indexes {
		if iv.IsProbability {
			t.Fatalf("index %s claims to be a probability", iv.IndexID)
		}
	}
	if m.EditorialRiskLevel == "" || m.UncertaintyScore == nil || m.UncertaintyLabel == "" {
		t.Fatalf("editorial/uncertainty missing: %+v", m)
	}
	if m.LastPublished == nil || *m.LastPublished != pubAt || m.LastReviewed == nil || *m.LastReviewed != pubAt || m.DataCutoff != "2026-09-01" {
		t.Fatalf("dates: published %v reviewed %v cutoff %s", m.LastPublished, m.LastReviewed, m.DataCutoff)
	}
	if m.Manifest.ReleaseID != firstRel || m.Manifest.Published == nil || len(m.Manifest.Reviewers) != 2 {
		t.Fatalf("manifest: %+v", m.Manifest)
	}
	if len(m.Headline.Official) != len(schema.Horizons) || len(m.Headline.ExternalAggregates) == 0 || len(m.Headline.Indexes) != 7 || !m.Headline.IsCurrent {
		t.Fatalf("headline: %+v", m.Headline)
	}
	// The served estimates are the release file, byte for byte in content.
	rel, err := publishing.LoadRelease(e.paths, firstRel)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := schema.CanonicalJSON(rel.Estimates)
	got, _ := schema.CanonicalJSON(append(append(append([]schema.Estimate{}, m.Official...), m.ExternalAggregates...), m.ResearchMode...))
	if !bytes.Equal(want, got) {
		t.Fatal("meter estimates differ from the release file")
	}
	// History has the single release.
	rr = do(h, http.MethodGet, "/v1/meter/history", nil, nil)
	var hist HistoryResponse
	decode(t, rr, &hist)
	if len(hist.Releases) != 1 || hist.Releases[0].ReleaseID != firstRel || !hist.Releases[0].IsCurrent || len(hist.Releases[0].HeightenedReviewTriggers) == 0 {
		t.Fatalf("history: %+v", hist)
	}
}

func TestOutcomesAndHorizons(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	var o OutcomesResponse
	decode(t, do(h, http.MethodGet, "/v1/outcomes", nil, nil), &o)
	if len(o.Outcomes) != 9 || o.Outcomes[6].Code != schema.O6 || !o.Outcomes[6].IncludedIn.Extinction || !o.Outcomes[6].IncludedIn.PDoom || o.Outcomes[0].IncludedIn.PDoom {
		t.Fatalf("outcomes: %+v", o.Outcomes)
	}
	if len(o.DerivedSets) != 5 || o.DerivedSets[0].Key != "P_DOOM" || len(o.DerivedSets[0].Outcomes) != 6 {
		t.Fatalf("derived sets: %+v", o.DerivedSets)
	}
	for _, d := range o.Outcomes {
		if d.Description == "" || d.Slug == "" || d.Label == "" {
			t.Fatalf("outcome %s incomplete", d.Code)
		}
	}
	var hz HorizonsResponse
	decode(t, do(h, http.MethodGet, "/v1/horizons", nil, nil), &hz)
	if len(hz.Horizons) != 7 || hz.Horizons[0].Years == nil || *hz.Horizons[0].Years != 1 || hz.Horizons[5].Kind != "calendar_year" || hz.Horizons[6].Years != nil || hz.Horizons[6].Note == "" {
		t.Fatalf("horizons: %+v", hz.Horizons)
	}
}

func TestNotFoundAndMethodNotAllowed(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	for _, p := range []string{"/v1/nope", "/v1/scenarios/S99", "/v1/scenarios/bogus", "/v1/sources/src-nope", "/v1/sources/src-Bad_ID", "/v1/releases/rel-2026-09-26-009", "/v1/releases/xyz", "/v1/methodology/nope", "/v1/methodology/notes", "/admin", "/v1/meter/extra"} {
		rr := do(h, http.MethodGet, p, nil, nil)
		expectError(t, rr, http.StatusNotFound, "not_found")
		if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatalf("%s: security headers missing on 404", p)
		}
	}
	rr := do(h, http.MethodPost, "/v1/meter", []byte(`{}`), nil)
	expectError(t, rr, http.StatusMethodNotAllowed, "method_not_allowed")
	if !strings.Contains(rr.Header().Get("Allow"), "GET") {
		t.Fatalf("Allow header: %q", rr.Header().Get("Allow"))
	}
	rr = do(h, http.MethodGet, "/v1/scenario-lab/evaluate", nil, nil)
	expectError(t, rr, http.StatusMethodNotAllowed, "method_not_allowed")
	rr = do(h, http.MethodDelete, "/v1/submissions/sources", nil, nil)
	expectError(t, rr, http.StatusMethodNotAllowed, "method_not_allowed")
	// Overlong request line.
	rr = do(h, http.MethodGet, "/v1/sources?q="+strings.Repeat("a", MaxURILength), nil, nil)
	expectError(t, rr, http.StatusRequestURITooLong, "uri_too_long")
}

func TestScenarioDetailAndSourceDetail(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	var sc ScenarioResponse
	decode(t, do(h, http.MethodGet, "/v1/scenarios/S3", nil, nil), &sc)
	if sc.Scenario.ID != "S3" || len(sc.Edges) != 1 || sc.Edges[0].ID != "se-S1-S3" || len(sc.Interventions) != 2 || sc.Meta.ReleaseID != firstRel {
		t.Fatalf("scenario S3: %+v", sc)
	}
	var list ScenariosResponse
	decode(t, do(h, http.MethodGet, "/v1/scenarios", nil, nil), &list)
	if len(list.Scenarios) != 2 || len(list.Edges) != 1 {
		t.Fatalf("scenarios: %+v", list)
	}
	var src SourceResponse
	decode(t, do(h, http.MethodGet, "/v1/sources/src-fixture-survey-2025", nil, nil), &src)
	if src.Source.ID != "src-fixture-survey-2025" || len(src.Claims) != 1 || src.Claims[0].ID != "clm-fixture-survey-01" {
		t.Fatalf("source: %+v", src)
	}
	var fc ForecastsResponse
	decode(t, do(h, http.MethodGet, "/v1/forecasts", nil, nil), &fc)
	if fc.Status != schema.StatusExternalAggregate || fc.Producer != model.AggregateModelVersion || len(fc.Forecasts) != 5 || len(fc.Aggregations) == 0 || len(fc.Estimates) == 0 {
		t.Fatalf("forecasts: status %s producer %s n=%d agg=%d est=%d", fc.Status, fc.Producer, len(fc.Forecasts), len(fc.Aggregations), len(fc.Estimates))
	}
	var dr DriversResponse
	decode(t, do(h, http.MethodGet, "/v1/drivers", nil, nil), &dr)
	if len(dr.Drivers) != 6 || len(dr.Observations) != 6 || len(dr.Explained.Items) == 0 {
		t.Fatalf("drivers: %d %d %d", len(dr.Drivers), len(dr.Observations), len(dr.Explained.Items))
	}
	var ca CapabilitiesResponse
	decode(t, do(h, http.MethodGet, "/v1/capabilities", nil, nil), &ca)
	if len(ca.Benchmarks) != 1 || len(ca.Results) != 2 {
		t.Fatalf("capabilities: %+v", ca)
	}
	var in IncidentsResponse
	decode(t, do(h, http.MethodGet, "/v1/incidents", nil, nil), &in)
	var sg SafeguardsResponse
	decode(t, do(h, http.MethodGet, "/v1/safeguards", nil, nil), &sg)
	var df DefinitionsResponse
	decode(t, do(h, http.MethodGet, "/v1/definitions", nil, nil), &df)
	var og OrganizationsResponse
	decode(t, do(h, http.MethodGet, "/v1/organizations", nil, nil), &og)
	var ac ActionsResponse
	decode(t, do(h, http.MethodGet, "/v1/actions", nil, nil), &ac)
	if len(in.Incidents) != 2 || len(sg.Interventions) != 2 || len(df.Definitions) != 1 || len(og.Organizations) != 1 || len(ac.Actions) != 1 {
		t.Fatalf("entity counts: %d %d %d %d %d", len(in.Incidents), len(sg.Interventions), len(df.Definitions), len(og.Organizations), len(ac.Actions))
	}
	var md MethodologyResponse
	decode(t, do(h, http.MethodGet, "/v1/methodology", nil, nil), &md)
	if len(md.Documents) != 2 || md.Documents[0].Slug != "aggregation" || md.Documents[0].Title != "Aggregation method" || md.Documents[1].Title != "indexes" || md.Documents[0].Path != "docs/method/aggregation.md" {
		t.Fatalf("methodology: %+v", md.Documents)
	}
	var rl ReleasesResponse
	decode(t, do(h, http.MethodGet, "/v1/releases", nil, nil), &rl)
	if rl.Current != firstRel || len(rl.Releases) != 1 || !rl.Releases[0].IsCurrent {
		t.Fatalf("releases: %+v", rl)
	}
	var rel ReleaseResponse
	decode(t, do(h, http.MethodGet, "/v1/releases/"+firstRel, nil, nil), &rel)
	if !rel.IsCurrent || rel.Release == nil || rel.Release.Manifest.ReleaseID != firstRel || len(rel.Release.Approvals) != 2 || len(rel.Release.Estimates) == 0 || rel.Meta.ReleaseID != firstRel {
		t.Fatalf("release: %+v", rel.Meta)
	}
}

func TestSourcesFiltersAndPagination(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	cases := []struct {
		query string
		total int
		n     int
	}{
		{"", 6, 6}, {"tier=1", 3, 3}, {"tier=2", 2, 2}, {"tier=4", 1, 1}, {"tier=3", 0, 0},
		{"topic=fixture", 6, 6}, {"topic=FIXTURE", 6, 6}, {"topic=none", 0, 0},
		{"q=tournament", 1, 1}, {"q=FIXTURE", 6, 6}, {"q=zzz", 0, 0}, {"tier=1&q=card", 1, 1},
		{"limit=2&offset=2", 6, 2}, {"limit=4&offset=4", 6, 2}, {"offset=100", 6, 0}, {"limit=200", 6, 6},
	}
	for _, c := range cases {
		var resp SourcesResponse
		rr := do(h, http.MethodGet, "/v1/sources?"+c.query, nil, nil)
		if rr.Code != 200 {
			t.Fatalf("%s: %d %s", c.query, rr.Code, rr.Body.String())
		}
		decode(t, rr, &resp)
		if resp.Total != c.total || len(resp.Sources) != c.n {
			t.Fatalf("%s: total %d n %d, want %d %d", c.query, resp.Total, len(resp.Sources), c.total, c.n)
		}
	}
	var page SourcesResponse
	decode(t, do(h, http.MethodGet, "/v1/sources?limit=2&offset=2", nil, nil), &page)
	if page.Limit != 2 || page.Offset != 2 || page.Sources[0].ID != "src-fixture-eval-report" {
		t.Fatalf("page: %+v", page)
	}
	for _, bad := range []string{"tier=9", "tier=abc", "tier=0", "limit=0", "limit=201", "limit=x", "offset=-1", "offset=y", "q=" + strings.Repeat("a", 201), "topic=" + strings.Repeat("b", 101)} {
		expectError(t, do(h, http.MethodGet, "/v1/sources?"+bad, nil, nil), http.StatusBadRequest, "bad_request")
	}
}

func TestETagAndConditionalRequests(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	first := do(h, http.MethodGet, "/v1/meter", nil, nil)
	etag := first.Header().Get("ETag")
	if !strings.HasPrefix(etag, `"`) || len(etag) != 66 {
		t.Fatalf("etag %q", etag)
	}
	sum := sha256.Sum256(first.Body.Bytes())
	if etag != `"`+hex.EncodeToString(sum[:])+`"` {
		t.Fatal("ETag is not the sha256 of the body")
	}
	second := do(h, http.MethodGet, "/v1/meter", nil, nil)
	if second.Header().Get("ETag") != etag || !bytes.Equal(first.Body.Bytes(), second.Body.Bytes()) {
		t.Fatal("repeated GET is not byte-identical")
	}
	for _, inm := range []string{etag, "W/" + etag, `"other", ` + etag, "*"} {
		rr := do(h, http.MethodGet, "/v1/meter", nil, map[string]string{"If-None-Match": inm})
		if rr.Code != http.StatusNotModified || rr.Body.Len() != 0 || rr.Header().Get("ETag") != etag {
			t.Fatalf("If-None-Match %q: %d body %d", inm, rr.Code, rr.Body.Len())
		}
	}
	rr := do(h, http.MethodGet, "/v1/meter", nil, map[string]string{"If-None-Match": `"stale"`})
	if rr.Code != 200 {
		t.Fatalf("stale validator should yield 200, got %d", rr.Code)
	}
	// HEAD carries headers and no body.
	rr = do(h, http.MethodHead, "/v1/meter", nil, nil)
	if rr.Code != 200 || rr.Body.Len() != 0 || rr.Header().Get("ETag") != etag || rr.Header().Get("Content-Length") == "" {
		t.Fatalf("HEAD: %d body %d etag %q", rr.Code, rr.Body.Len(), rr.Header().Get("ETag"))
	}
	// Unknown-field-free error bodies are not cached.
	if do(h, http.MethodGet, "/v1/nope", nil, nil).Header().Get("Cache-Control") != "no-store" {
		t.Fatal("error responses must be no-store")
	}
}

func TestRateLimiting(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, func(d *Deps) { d.RateLimits.Read = Limit{Burst: 3, PerMinute: 60} })
	for i := 0; i < 3; i++ {
		rr := do(h, http.MethodGet, "/v1/outcomes", nil, nil)
		if rr.Code != 200 {
			t.Fatalf("request %d: %d", i, rr.Code)
		}
		if rr.Header().Get("RateLimit-Limit") != "3" || rr.Header().Get("RateLimit-Remaining") == "" {
			t.Fatalf("rate limit headers: %v", rr.Header())
		}
	}
	rr := do(h, http.MethodGet, "/v1/outcomes", nil, nil)
	expectError(t, rr, http.StatusTooManyRequests, "rate_limited")
	if rr.Header().Get("Retry-After") != "1" {
		t.Fatalf("Retry-After %q", rr.Header().Get("Retry-After"))
	}
	// Other classes are independent of the read bucket.
	if do(h, http.MethodPost, "/v1/scenario-lab/evaluate", labBody(t, nil), nil).Code != 200 {
		t.Fatal("lab bucket should be independent")
	}
	e.clock.advance(time.Second)
	if rr := do(h, http.MethodGet, "/v1/outcomes", nil, nil); rr.Code != 200 {
		t.Fatalf("after refill: %d", rr.Code)
	}
	// Another client has its own bucket.
	req := httptest.NewRequest(http.MethodGet, "/v1/outcomes", nil)
	req.RemoteAddr = "198.51.100.7:4444"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("other client: %d", rec.Code)
	}
}

func TestTrustProxyControlsClientKey(t *testing.T) {
	e := buildDataDir(t)
	send := func(h http.Handler, xff string) int {
		req := httptest.NewRequest(http.MethodGet, "/v1/outcomes", nil)
		req.Header.Set("X-Forwarded-For", xff)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	untrusted := e.handler(t, func(d *Deps) { d.RateLimits.Read = Limit{Burst: 1, PerMinute: 60} })
	if send(untrusted, "203.0.113.1") != 200 || send(untrusted, "203.0.113.2") != 429 {
		t.Fatal("X-Forwarded-For must be ignored unless the proxy is trusted")
	}
	trusted := e.handler(t, func(d *Deps) { d.RateLimits.Read = Limit{Burst: 1, PerMinute: 60}; d.TrustProxy = true })
	if send(trusted, "203.0.113.1") != 200 || send(trusted, "203.0.113.2") != 200 || send(trusted, "10.0.0.9, 203.0.113.2") != 429 {
		t.Fatal("trusted proxy should key on the right-most X-Forwarded-For entry")
	}
}

func TestAccessLogHidesAddresses(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/outcomes", nil)
	req.RemoteAddr = "203.0.113.77:5555"
	req.Header.Set("X-Request-ID", "trace-123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	logs := e.logs.String()
	if strings.Contains(logs, "203.0.113.77") {
		t.Fatal("access log leaks the client address")
	}
	if !strings.Contains(logs, "client=") || !strings.Contains(logs, "request_id=trace-123") || !strings.Contains(logs, "path=/v1/outcomes") || !strings.Contains(logs, "status=200") {
		t.Fatalf("access log incomplete:\n%s", logs)
	}
	if rec.Header().Get("X-Request-ID") != "trace-123" {
		t.Fatal("well-formed request id should be echoed")
	}
	req = httptest.NewRequest(http.MethodGet, "/v1/outcomes", nil)
	req.Header.Set("X-Request-ID", "bad id\n")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if id := rec.Header().Get("X-Request-ID"); id == "bad id\n" || len(id) != 24 {
		t.Fatalf("malformed request id must be replaced, got %q", id)
	}
}

func TestCORS(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	rr := do(h, http.MethodGet, "/v1/meter", nil, map[string]string{"Origin": "https://anywhere.example"})
	if rr.Code != 200 || rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("GET CORS: %d %q", rr.Code, rr.Header().Get("Access-Control-Allow-Origin"))
	}
	before := treeHash(t, e.paths.DataDir)
	body := []byte(`{"payload":{"canonical_url":"https://example.org/x","title":"T","why_relevant":"r","claimed_evidence":"e"}}`)
	rr = do(h, http.MethodPost, "/v1/submissions/sources", body, map[string]string{"Origin": "https://evil.example"})
	expectError(t, rr, http.StatusForbidden, "origin_not_allowed")
	if treeHash(t, e.paths.DataDir) != before {
		t.Fatal("a refused cross-origin POST wrote to the data directory")
	}
	rr = do(h, http.MethodPost, "/v1/submissions/sources", body, map[string]string{"Origin": "https://lab.example"})
	if rr.Code != http.StatusAccepted || rr.Header().Get("Access-Control-Allow-Origin") != "https://lab.example" {
		t.Fatalf("allowlisted origin: %d %q %s", rr.Code, rr.Header().Get("Access-Control-Allow-Origin"), rr.Body.String())
	}
	rr = do(h, http.MethodPost, "/v1/submissions/sources", body, map[string]string{"Origin": "http://example.com"})
	if rr.Code != http.StatusAccepted {
		t.Fatalf("same-origin POST: %d %s", rr.Code, rr.Body.String())
	}
	// Preflight.
	rr = do(h, http.MethodOptions, "/v1/meter", nil, map[string]string{"Origin": "https://anywhere.example", "Access-Control-Request-Method": "GET"})
	if rr.Code != http.StatusNoContent || rr.Header().Get("Access-Control-Allow-Origin") != "*" || !strings.Contains(rr.Header().Get("Access-Control-Allow-Methods"), "GET") {
		t.Fatalf("GET preflight: %d %v", rr.Code, rr.Header())
	}
	rr = do(h, http.MethodOptions, "/v1/submissions/sources", nil, map[string]string{"Origin": "https://evil.example", "Access-Control-Request-Method": "POST"})
	if rr.Code != http.StatusForbidden || rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("POST preflight from unknown origin: %d %v", rr.Code, rr.Header())
	}
	rr = do(h, http.MethodOptions, "/v1/submissions/sources", nil, map[string]string{"Origin": "https://lab.example", "Access-Control-Request-Method": "POST"})
	if rr.Code != http.StatusNoContent || rr.Header().Get("Access-Control-Allow-Origin") != "https://lab.example" || rr.Header().Get("Access-Control-Allow-Methods") != "POST, OPTIONS" {
		t.Fatalf("POST preflight from allowlisted origin: %d %v", rr.Code, rr.Header())
	}
	rr = do(h, http.MethodOptions, "/v1/meter", nil, nil)
	if rr.Code != http.StatusNoContent || !strings.Contains(rr.Header().Get("Allow"), "GET") {
		t.Fatalf("plain OPTIONS: %d %v", rr.Code, rr.Header())
	}
}

func TestSubmissionsLandOnlyInTheQueue(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	before := treeHash(t, e.paths.Releases) + treeHash(t, e.paths.Snapshots)
	queue := SubmissionsPath(e.paths.Review)
	if _, err := os.Stat(queue); !os.IsNotExist(err) {
		t.Fatal("queue should not exist before the first submission")
	}
	body := []byte("{\"payload\":{\"canonical_url\":\"https://example.org/report\",\"title\":\"Ti\\u0000tle\\u0007 with\\u001b control chars \",\"publisher\":\"Pub\",\"date_published\":\"2026-05\",\"why_relevant\":\"Line one\\nline two\",\"claimed_evidence\":\"\\tTable 2\"},\"contact\":\"person@example.org\\u0008\"}")
	rr := do(h, http.MethodPost, "/v1/submissions/sources", body, nil)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("%d %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("POST responses must be no-store")
	}
	var ack SubmissionResponse
	decode(t, rr, &ack)
	if !strings.HasPrefix(ack.ID, "sub-") || len(ack.ID) != 20 || ack.Kind != "source" || ack.Status != "received" || ack.SubmittedAt != e.clock.now().Format(time.RFC3339) {
		t.Fatalf("ack: %+v", ack)
	}
	corr := []byte(`{"payload":{"target_id":"src-fixture-survey-2025","field":"date_published","correction":"The survey closed in July.","evidence_url":"https://example.org/survey"}}`)
	rr = do(h, http.MethodPost, "/v1/submissions/corrections", corr, nil)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("correction: %d %s", rr.Code, rr.Body.String())
	}
	var ack2 SubmissionResponse
	decode(t, rr, &ack2)
	if ack2.Kind != "correction" || ack2.ID == ack.ID {
		t.Fatalf("correction ack: %+v", ack2)
	}
	// Exactly two JSONL lines, and nothing else changed.
	f, err := os.Open(queue)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var lines []Submission
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var sub Submission
		if err := json.Unmarshal(sc.Bytes(), &sub); err != nil {
			t.Fatalf("queue line: %v", err)
		}
		lines = append(lines, sub)
	}
	if len(lines) != 2 {
		t.Fatalf("queue has %d lines", len(lines))
	}
	first := lines[0]
	payload := first.Payload.(map[string]any)
	if first.ID != ack.ID || first.Kind != "source" || first.Status != "received" || first.SubmittedAt != ack.SubmittedAt {
		t.Fatalf("stored submission: %+v", first)
	}
	if payload["title"] != "Title with control chars" || payload["why_relevant"] != "Line one\nline two" || payload["claimed_evidence"] != "Table 2" || payload["date_published"] != "2026-05" {
		t.Fatalf("control characters not stripped: %+v", payload)
	}
	if first.Contact == nil || *first.Contact != "person@example.org" {
		t.Fatalf("contact: %v", first.Contact)
	}
	if lines[1].Kind != "correction" || lines[1].Payload.(map[string]any)["target_id"] != "src-fixture-survey-2025" {
		t.Fatalf("second line: %+v", lines[1])
	}
	if treeHash(t, e.paths.Releases)+treeHash(t, e.paths.Snapshots) != before {
		t.Fatal("submissions modified releases or snapshots")
	}
	if _, err := os.Stat(e.paths.Audit); !os.IsNotExist(err) {
		// The audit log exists only from the promotion; make sure it did not grow.
		raw, _ := os.ReadFile(e.paths.Audit)
		if strings.Count(string(raw), "\n") != 1 {
			t.Fatal("submissions must not append audit events")
		}
	}
}

func TestSubmissionValidation(t *testing.T) {
	e := buildDataDir(t)
	// Many rejections under a frozen clock: give this handler a large bucket
	// (the limiter itself is exercised below and in TestRateLimiting).
	h := e.handler(t, func(d *Deps) { d.RateLimits.Submit = Limit{Burst: 1000, PerMinute: 60} })
	bad := []struct {
		route, body string
		status      int
		code        string
	}{
		{"/v1/submissions/sources", `{"payload":{"canonical_url":"https://example.org/x","why_relevant":"r","claimed_evidence":"e"}}`, 400, "bad_request"},
		{"/v1/submissions/sources", `{"payload":{"canonical_url":"ftp://example.org/x","title":"T","why_relevant":"r","claimed_evidence":"e"}}`, 400, "bad_request"},
		{"/v1/submissions/sources", `{"payload":{"canonical_url":"https://user:pw@example.org/x","title":"T","why_relevant":"r","claimed_evidence":"e"}}`, 400, "bad_request"},
		{"/v1/submissions/sources", `{"payload":{"canonical_url":"https://example.org/x","title":"T","why_relevant":"r","claimed_evidence":"e","extra":1}}`, 400, "bad_request"},
		{"/v1/submissions/sources", `{"payload":{"canonical_url":"https://example.org/x","title":"T","why_relevant":"r","claimed_evidence":"e","date_published":"May 2026"}}`, 400, "bad_request"},
		{"/v1/submissions/sources", `{"payload":{"canonical_url":"https://example.org/x","title":"T","why_relevant":"r"}}`, 400, "bad_request"},
		{"/v1/submissions/sources", `{"contact":"x"}`, 400, "bad_request"},
		{"/v1/submissions/sources", `{"payload":{"canonical_url":"https://example.org/x","title":"T","why_relevant":"r","claimed_evidence":"e"},"admin":true}`, 400, "bad_request"},
		{"/v1/submissions/sources", `not json`, 400, "bad_request"},
		{"/v1/submissions/sources", ``, 400, "bad_request"},
		{"/v1/submissions/sources", `{"payload":{"canonical_url":"https://example.org/x","title":"T","why_relevant":"r","claimed_evidence":"e"}} trailing`, 400, "bad_request"},
		{"/v1/submissions/corrections", `{"payload":{"target_id":"","correction":"c"}}`, 400, "bad_request"},
		{"/v1/submissions/corrections", `{"payload":{"target_id":"src-x","correction":""}}`, 400, "bad_request"},
		{"/v1/submissions/corrections", `{"payload":{"target_id":"src-x","correction":"c","evidence_url":"javascript:alert(1)"}}`, 400, "bad_request"},
		{"/v1/submissions/corrections", `{"payload":{"target_id":"src-x","correction":"c","field":"Not Snake"}}`, 400, "bad_request"},
		{"/v1/submissions/corrections", `{"payload":{"target_id":"src-x","correction":"c"},"contact":"` + strings.Repeat("c", 400) + `"}`, 400, "bad_request"},
	}
	for _, c := range bad {
		rr := do(h, http.MethodPost, c.route, []byte(c.body), nil)
		e := expectError(t, rr, c.status, c.code)
		if e.Detail == "" {
			t.Fatalf("%s %s: empty detail", c.route, c.body)
		}
	}
	// Missing-title error names the field so a client can fix it.
	rr := do(h, http.MethodPost, "/v1/submissions/sources", []byte(`{"payload":{"canonical_url":"https://example.org/x","why_relevant":"r","claimed_evidence":"e"}}`), nil)
	if !strings.Contains(rr.Body.String(), "title") {
		t.Fatalf("detail should name the missing field: %s", rr.Body.String())
	}
	rr = do(h, http.MethodPost, "/v1/submissions/sources", []byte(`{"payload":{"canonical_url":"https://example.org/x","title":"  ","why_relevant":"r","claimed_evidence":"e"}}`), nil)
	if e := expectError(t, rr, http.StatusBadRequest, "bad_request"); !strings.Contains(e.Detail, "title is required") {
		t.Fatalf("blank title should be rejected: %s", e.Detail)
	}
	// Too large and wrong media type.
	big := []byte(`{"payload":{"canonical_url":"https://example.org/x","title":"` + strings.Repeat("a", MaxBodyBytes) + `","why_relevant":"r","claimed_evidence":"e"}}`)
	expectError(t, do(h, http.MethodPost, "/v1/submissions/sources", big, nil), http.StatusRequestEntityTooLarge, "payload_too_large")
	expectError(t, do(h, http.MethodPost, "/v1/submissions/sources", []byte(`{}`), map[string]string{"Content-Type": "text/plain"}), http.StatusUnsupportedMediaType, "unsupported_media_type")
	if _, err := os.Stat(SubmissionsPath(e.paths.Review)); !os.IsNotExist(err) {
		t.Fatal("rejected submissions must not be queued")
	}
	// Submission rate limit.
	limited := e.handler(t, func(d *Deps) { d.RateLimits.Submit = Limit{Burst: 1, PerMinute: 60} })
	good := []byte(`{"payload":{"target_id":"src-x","correction":"c"}}`)
	if do(limited, http.MethodPost, "/v1/submissions/corrections", good, nil).Code != http.StatusAccepted {
		t.Fatal("first submission should pass")
	}
	expectError(t, do(limited, http.MethodPost, "/v1/submissions/corrections", good, nil), http.StatusTooManyRequests, "rate_limited")
}

func TestScenarioLab(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, func(d *Deps) { d.RateLimits.Lab = Limit{Burst: 1000, PerMinute: 60} })
	before := treeHash(t, e.paths.DataDir)
	rr := do(h, http.MethodPost, "/v1/scenario-lab/evaluate", labBody(t, map[string]any{"samples": 500, "seed": 7}), nil)
	if rr.Code != 200 {
		t.Fatalf("%d %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("lab responses must be no-store")
	}
	var resp ScenarioLabResponse
	decode(t, rr, &resp)
	if resp.Label != "user_scenario" || resp.Status != schema.StatusUserScenario || resp.Result.Label != "user_scenario" {
		t.Fatalf("label: %+v", resp)
	}
	if !strings.Contains(resp.Disclaimer, "not the p(DOOM) official model") || resp.Disclaimer != resp.Result.Disclaimer {
		t.Fatalf("disclaimer: %q", resp.Disclaimer)
	}
	if resp.Horizon != "10y" || len(resp.OutcomeSet) != 6 || resp.Producer != e.snap.ModelSpec.ExperimentalCausal.Version || resp.ModelVersion != resp.Producer {
		t.Fatalf("labels: %+v", resp)
	}
	if resp.ReleaseID != firstRel || resp.DataCutoff != "2026-09-01" || resp.DataSnapshot != "snap-1999-01-01-001" || len(resp.Limitations) == 0 {
		t.Fatalf("provenance: %+v", resp)
	}
	if resp.Samples != 500 || resp.Seed != 7 {
		t.Fatalf("samples/seed echo: %d %d", resp.Samples, resp.Seed)
	}
	if _, ok := resp.Result.OutcomeEstimates["P_DOOM"]; !ok || len(resp.OutcomeSets["P_DOOM"]) != 6 || len(resp.OutcomeSets["O6"]) != 1 {
		t.Fatalf("outcome sets: %+v", resp.OutcomeSets)
	}
	if resp.Result.ParamsEcho.CapabilityTimeline != 1 || resp.Result.ParamsEcho.SafetyProgress != -1 || resp.Result.ParamsEcho.Resilience != 1 {
		t.Fatalf("params echo: %+v", resp.Result.ParamsEcho)
	}
	// Deterministic and identical to a direct model call.
	want, err := model.EvaluateUserScenario(e.snap.ModelSpec.ExperimentalCausal, schema.UserScenarioParams{Horizon: "10y", CapabilityTimeline: 1, SafetyProgress: -1, Resilience: 1}, 500, 7)
	if err != nil {
		t.Fatal(err)
	}
	if wj, _ := schema.CanonicalJSON(want); !bytes.Equal(wj, must(schema.CanonicalJSON(resp.Result))) {
		t.Fatal("API result differs from model.EvaluateUserScenario")
	}
	again := do(h, http.MethodPost, "/v1/scenario-lab/evaluate", labBody(t, map[string]any{"samples": 500, "seed": 7}), nil)
	if !bytes.Equal(again.Body.Bytes(), rr.Body.Bytes()) {
		t.Fatal("scenario lab is not deterministic")
	}
	// Defaults: samples and seed from the spec.
	var def ScenarioLabResponse
	decode(t, do(h, http.MethodPost, "/v1/scenario-lab/evaluate", labBody(t, nil), nil), &def)
	if def.Samples != e.snap.ModelSpec.ExperimentalCausal.Samples || def.Seed != e.snap.ModelSpec.ExperimentalCausal.Seed {
		t.Fatalf("defaults: %d %d", def.Samples, def.Seed)
	}
	// Samples above the spec count are capped to it.
	var capped ScenarioLabResponse
	decode(t, do(h, http.MethodPost, "/v1/scenario-lab/evaluate", labBody(t, map[string]any{"samples": 40000}), nil), &capped)
	if capped.Samples != e.snap.ModelSpec.ExperimentalCausal.Samples {
		t.Fatalf("samples should be capped to the spec: %d", capped.Samples)
	}
	// Rejections.
	rejections := []struct {
		name string
		body []byte
		ct   string
		want int
		code string
		text string
	}{
		{"slider high", labBody(t, map[string]any{"capability_timeline": 3}), "", 400, "bad_request", "capability_timeline"},
		{"slider low", labBody(t, map[string]any{"resilience": -3}), "", 400, "bad_request", "resilience"},
		{"slider float", labBody(t, map[string]any{"resilience": 1.5}), "", 400, "bad_request", "resilience"},
		{"missing slider", labBody(t, map[string]any{"access_level": nil}), "", 400, "bad_request", "access_level"},
		{"missing horizon", labBody(t, map[string]any{"horizon": nil}), "", 400, "bad_request", "horizon"},
		{"unknown horizon", labBody(t, map[string]any{"horizon": "soon"}), "", 400, "bad_request", "horizon"},
		{"uncovered horizon", labBody(t, map[string]any{"horizon": "1y"}), "", 400, "bad_request", "not covered"},
		{"samples too many", labBody(t, map[string]any{"samples": MaxSamples + 1}), "", 400, "bad_request", "samples"},
		{"samples zero", labBody(t, map[string]any{"samples": 0}), "", 400, "bad_request", "samples"},
		{"seed negative", labBody(t, map[string]any{"seed": -1}), "", 400, "bad_request", "seed"},
		{"unknown field", labBody(t, map[string]any{"official": true}), "", 400, "bad_request", "unknown field"},
		{"not json", []byte("hello"), "", 400, "bad_request", "syntax"},
		{"empty", []byte(""), "", 400, "bad_request", "empty"},
		{"wrong media type", labBody(t, nil), "text/plain", 415, "unsupported_media_type", "application/json"},
	}
	for _, c := range rejections {
		headers := map[string]string{}
		if c.ct != "" {
			headers["Content-Type"] = c.ct
		}
		rr := do(h, http.MethodPost, "/v1/scenario-lab/evaluate", c.body, headers)
		e := expectError(t, rr, c.want, c.code)
		if !strings.Contains(e.Detail, c.text) {
			t.Fatalf("%s: detail %q lacks %q", c.name, e.Detail, c.text)
		}
	}
	// Every slider out of range is reported at once.
	all := labBody(t, map[string]any{"capability_timeline": 5, "autonomy_growth": -5})
	e2 := expectError(t, do(h, http.MethodPost, "/v1/scenario-lab/evaluate", all, nil), 400, "bad_request")
	if !strings.Contains(e2.Detail, "capability_timeline") || !strings.Contains(e2.Detail, "autonomy_growth") {
		t.Fatalf("all problems should be reported: %s", e2.Detail)
	}
	big := append([]byte(`{"horizon":"10y","capability_timeline":0,"autonomy_growth":0,"access_level":0,"safety_progress":0,"governance_strength":0,"model_security":0,"open_weight_diffusion":0,"international_coordination":0,"incident_frequency":0,"resilience":0,"samples":`), []byte(strings.Repeat("1", MaxBodyBytes)+"}")...)
	expectError(t, do(h, http.MethodPost, "/v1/scenario-lab/evaluate", big, nil), http.StatusRequestEntityTooLarge, "payload_too_large")
	if treeHash(t, e.paths.DataDir) != before {
		t.Fatal("scenario lab wrote to the data directory")
	}
	// Lab rate limit.
	limited := e.handler(t, func(d *Deps) { d.RateLimits.Lab = Limit{Burst: 2, PerMinute: 60} })
	for i := 0; i < 2; i++ {
		if do(limited, http.MethodPost, "/v1/scenario-lab/evaluate", labBody(t, nil), nil).Code != 200 {
			t.Fatal("within burst")
		}
	}
	expectError(t, do(limited, http.MethodPost, "/v1/scenario-lab/evaluate", labBody(t, nil), nil), http.StatusTooManyRequests, "rate_limited")
}

func must(b []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return b
}

func TestNoReleaseIsNotReadyButAlive(t *testing.T) {
	root := t.TempDir()
	paths := config.Resolve(filepath.Join(root, "data"), "")
	h := NewHandler(Deps{Paths: paths})
	if rr := do(h, http.MethodGet, "/healthz", nil, nil); rr.Code != 200 {
		t.Fatalf("healthz %d", rr.Code)
	}
	rr := do(h, http.MethodGet, "/readyz", nil, nil)
	if rr.Code != http.StatusServiceUnavailable || !strings.Contains(rr.Body.String(), "not_ready") {
		t.Fatalf("readyz %d %s", rr.Code, rr.Body.String())
	}
	expectError(t, do(h, http.MethodGet, "/v1/meter", nil, nil), http.StatusServiceUnavailable, "unavailable")
	expectError(t, do(h, http.MethodPost, "/v1/scenario-lab/evaluate", []byte(`{}`), nil), http.StatusServiceUnavailable, "unavailable")
	if rr := do(h, http.MethodGet, "/v1/outcomes", nil, nil); rr.Code != 200 {
		t.Fatalf("outcomes should not need a release: %d", rr.Code)
	}
	var rl ReleasesResponse
	decode(t, do(h, http.MethodGet, "/v1/releases", nil, nil), &rl)
	if rl.Current != "" || len(rl.Releases) != 0 {
		t.Fatalf("releases: %+v", rl)
	}
	var md MethodologyResponse
	decode(t, do(h, http.MethodGet, "/v1/methodology", nil, nil), &md)
	if len(md.Documents) != 0 {
		t.Fatalf("methodology without a directory: %+v", md)
	}
	// Submissions still work without a release.
	if rr := do(h, http.MethodPost, "/v1/submissions/corrections", []byte(`{"payload":{"target_id":"x","correction":"c"}}`), nil); rr.Code != http.StatusAccepted {
		t.Fatalf("submission without release: %d %s", rr.Code, rr.Body.String())
	}
}

func TestPromotionIsPickedUpWithoutRestart(t *testing.T) {
	e := buildDataDir(t)
	h := e.handler(t, nil)
	var m MeterResponse
	decode(t, do(h, http.MethodGet, "/v1/meter", nil, nil), &m)
	if m.Meta.ReleaseID != firstRel {
		t.Fatalf("first: %s", m.Meta.ReleaseID)
	}
	// Second snapshot and release through the gate.
	first, err := publishing.LoadRelease(e.paths, firstRel)
	if err != nil {
		t.Fatal(err)
	}
	fx := fixture.New()
	fx.Manifest.SnapshotID = "snap-1999-02-01-001"
	fx.DriverObservations[0].ValueNormalized = 0.9
	snapDir := filepath.Join(e.paths.Snapshots, fx.Manifest.SnapshotID)
	if err := snapshot.Write(snapDir, fx); err != nil {
		t.Fatal(err)
	}
	snap2, err := snapshot.Load(snapDir)
	if err != nil {
		t.Fatal(err)
	}
	e.promote(t, snap2, first, secondRel, "cand-2026-10-01-001", "2026-10-01T00:00:00Z", "2026-10-01T02:00:00Z")
	// Within the refresh interval the old release is still served.
	e.clock.advance(2 * time.Second)
	decode(t, do(h, http.MethodGet, "/v1/meter", nil, nil), &m)
	if m.Meta.ReleaseID != firstRel {
		t.Fatalf("release changed before the refresh interval: %s", m.Meta.ReleaseID)
	}
	e.clock.advance(4 * time.Second)
	rr := do(h, http.MethodGet, "/v1/meter", nil, nil)
	decode(t, rr, &m)
	if m.Meta.ReleaseID != secondRel || m.Meta.DataSnapshot != "snap-1999-02-01-001" || m.Manifest.PreviousReleaseID == nil || *m.Manifest.PreviousReleaseID != firstRel {
		t.Fatalf("second release not served: %+v", m.Meta)
	}
	var hist HistoryResponse
	decode(t, do(h, http.MethodGet, "/v1/meter/history", nil, nil), &hist)
	if len(hist.Releases) != 2 || hist.Releases[0].ReleaseID != firstRel || hist.Releases[0].IsCurrent || hist.Releases[0].Superseded == nil || hist.Releases[0].Superseded.By != secondRel || !hist.Releases[1].IsCurrent {
		t.Fatalf("history: %+v", hist.Releases)
	}
	var rel ReleaseResponse
	decode(t, do(h, http.MethodGet, "/v1/releases/"+firstRel, nil, nil), &rel)
	if rel.IsCurrent {
		t.Fatal("first release should no longer be current")
	}
	// Rollback moves CURRENT back; picked up after the interval too.
	if err := publishing.Rollback(e.paths, firstRel, "2026-10-01T03:00:00Z", "test", "regression"); err != nil {
		t.Fatal(err)
	}
	e.clock.advance(6 * time.Second)
	decode(t, do(h, http.MethodGet, "/v1/meter", nil, nil), &m)
	if m.Meta.ReleaseID != firstRel {
		t.Fatalf("rollback not served: %s", m.Meta.ReleaseID)
	}
	// A broken CURRENT keeps the last good release.
	if err := os.WriteFile(e.paths.Current, []byte("rel-0000-00-00-000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.clock.advance(6 * time.Second)
	decode(t, do(h, http.MethodGet, "/v1/meter", nil, nil), &m)
	if m.Meta.ReleaseID != firstRel {
		t.Fatalf("broken CURRENT should keep the last good release: %s", m.Meta.ReleaseID)
	}
	if !strings.Contains(e.logs.String(), "release reload failed") {
		t.Fatal("reload failure should be logged")
	}
}

func TestUnpublishedReleaseIsRefused(t *testing.T) {
	e := buildDataDir(t)
	// Point CURRENT at a directory that is a candidate (not published).
	fx := fixture.New()
	res, err := model.Run(e.snap, model.RunOptions{GeneratedAt: genAt, CodeCommit: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	m, err := publishing.BuildManifest(res, e.snap, publishing.BuildOptions{ReleaseID: "rel-2026-09-27-001", CandidateID: "cand-2026-09-27-001", CodeCommit: "abc1234"})
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(e.paths.Releases, "rel-2026-09-27-001")
	if _, err := publishing.WriteCandidate(dir, res, m); err != nil {
		t.Fatal(err)
	}
	_ = fx
	h := e.handler(t, nil)
	expectError(t, do(h, http.MethodGet, "/v1/releases/rel-2026-09-27-001", nil, nil), http.StatusNotFound, "not_found")
	if err := os.WriteFile(e.paths.Current, []byte("rel-2026-09-27-001\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.clock.advance(6 * time.Second)
	var mr MeterResponse
	decode(t, do(h, http.MethodGet, "/v1/meter", nil, nil), &mr)
	if mr.Meta.ReleaseID != firstRel || !strings.Contains(e.logs.String(), "not marked published") {
		t.Fatalf("unpublished release must not be served: %s\n%s", mr.Meta.ReleaseID, e.logs.String())
	}
}
