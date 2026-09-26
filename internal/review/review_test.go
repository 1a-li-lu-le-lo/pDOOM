// Copyright NU Cybernetics. p(DOOM) — research prototype.

package review

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1a-li-lu-le-lo/pdoom/internal/claims"
	"github.com/1a-li-lu-le-lo/pdoom/internal/parsing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

func sampleItem(url string) ReviewItem {
	doc := parsing.NewDocument("src-a", url, "Title", "Researchers estimate a 5% chance of failure.", nil, "2026-09-26T00:00:00Z", []string{"A"})
	return ReviewItem{
		ID:                 NewItemID(KindDocument, "src-a", doc.CanonicalURL, doc.ContentHash),
		Kind:               KindDocument,
		CreatedAt:          "2026-09-26T00:00:00Z",
		SourceID:           "src-a",
		Document:           &doc,
		CandidateClaims:    claims.Extract(doc.Text),
		Quality:            Quality{Tier: 2, Conflicts: []schema.ConflictLabel{"none_known"}, RobotsStatus: "allowed", DuplicateCluster: "dup-0000000000000000", Score: 0.8},
		Archive:            &Archive{FetchURL: "https://a.example/feed", FinalURL: "https://a.example/feed", SHA256: strings.Repeat("0", 64), ContentType: "application/rss+xml", Bytes: 10, RetrievedAt: "2026-09-26T00:00:00Z"},
		Classification:     []string{"forecast"},
		PrimarySourceLinks: []string{},
		Contradictions:     []string{},
		ContradictionCheck: ContradictionCheckNotImplemented,
		Status:             StatusPending,
	}
}

func TestAppendListAndIdempotence(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "review-queue")
	q := NewQueue(dir)
	a := sampleItem("https://a.example/1")
	b := sampleItem("https://a.example/2")
	path, n, err := q.Append("2026-09-26", []ReviewItem{a, b})
	if err != nil || n != 2 {
		t.Fatalf("append: n=%d err=%v", n, err)
	}
	if filepath.Base(path) != "2026-09-26-ingest.jsonl" {
		t.Fatalf("path = %s", path)
	}
	// Appending the same items again writes nothing new; a third item is added.
	c := sampleItem("https://a.example/3")
	_, n, err = q.Append("2026-09-26", []ReviewItem{a, b, c})
	if err != nil || n != 1 {
		t.Fatalf("second append: n=%d err=%v", n, err)
	}
	items, err := q.List()
	if err != nil || len(items) != 3 {
		t.Fatalf("list: %d %v", len(items), err)
	}
	if items[0].ID != a.ID || items[2].ID != c.ID || items[0].Document.URL != "https://a.example/1" {
		t.Fatalf("order/content: %+v", items)
	}
	if len(items[0].CandidateClaims) != 1 || items[0].CandidateClaims[0].ModelUseStatus != "excluded" {
		t.Fatalf("claims round trip: %+v", items[0].CandidateClaims)
	}
	// Lines are canonical JSON: sorted keys, one object per line.
	raw, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %d", len(lines))
	}
	for _, l := range lines {
		want, _ := schema.CanonicalJSON(json.RawMessage(l))
		if string(want) != l {
			t.Fatalf("line is not canonical JSON: %s", l)
		}
		if !strings.HasPrefix(l, `{"archive":`) {
			t.Fatalf("keys must be sorted: %s", l[:40])
		}
	}
	docs, err := q.ListKind(KindDocument)
	if err != nil || len(docs) != 3 {
		t.Fatalf("ListKind: %d %v", len(docs), err)
	}
	subs, _ := q.ListKind(KindSubmission)
	if len(subs) != 0 {
		t.Fatal("no submissions expected")
	}
}

func TestQueueNeverTouchesOtherFiles(t *testing.T) {
	root := t.TempDir()
	other := filepath.Join(root, "releases", "CURRENT")
	if err := os.MkdirAll(filepath.Dir(other), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte("rel-x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	q := NewQueue(filepath.Join(root, "review-queue"))
	if _, _, err := q.Append("2026-09-26", []ReviewItem{sampleItem("https://a.example/1")}); err != nil {
		t.Fatal(err)
	}
	files, _ := q.Files()
	if len(files) != 1 || !strings.HasPrefix(files[0], q.Dir) {
		t.Fatalf("files = %v", files)
	}
	if got, _ := os.ReadFile(other); string(got) != "rel-x\n" {
		t.Fatal("queue must not touch files outside its directory")
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 2 {
		t.Fatalf("unexpected entries in root: %v", entries)
	}
}

func TestAppendRejectsBadDatesAndInvalidItems(t *testing.T) {
	q := NewQueue(t.TempDir())
	if _, _, err := q.Append("2026/09/26", nil); err == nil {
		t.Fatal("bad date must fail")
	}
	if _, _, err := q.Append("../escape", nil); err == nil {
		t.Fatal("path-like date must fail")
	}
	good := sampleItem("https://a.example/1")
	bad := sampleItem("https://a.example/2")
	bad.Status = "approved"
	_, n, err := q.Append("2026-09-26", []ReviewItem{good, bad})
	if err == nil || n != 0 {
		t.Fatalf("invalid item must abort the whole append: n=%d err=%v", n, err)
	}
	if files, _ := q.Files(); len(files) != 0 {
		t.Fatal("nothing may be written when validation fails")
	}
}

func TestValidate(t *testing.T) {
	ok := sampleItem("https://a.example/1")
	if err := Validate(ok); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		f    func(*ReviewItem)
		want string
	}{
		{"id", func(i *ReviewItem) { i.ID = "x" }, "id"},
		{"kind", func(i *ReviewItem) { i.Kind = "note" }, "kind"},
		{"created", func(i *ReviewItem) { i.CreatedAt = "" }, "created_at"},
		{"source", func(i *ReviewItem) { i.SourceID = "" }, "source_id"},
		{"status", func(i *ReviewItem) { i.Status = "reviewed" }, "status"},
		{"document", func(i *ReviewItem) { i.Document = nil }, "require a document"},
		{"document hash", func(i *ReviewItem) { i.Document.ContentHash = "" }, "content_hash"},
		{"tier", func(i *ReviewItem) { i.Quality.Tier = 0 }, "quality.tier"},
		{"robots", func(i *ReviewItem) { i.Quality.RobotsStatus = "maybe" }, "robots_status"},
		{"claim status", func(i *ReviewItem) { i.CandidateClaims[0].ModelUseStatus = "eligible" }, "candidate/excluded"},
		{"submission payload", func(i *ReviewItem) { i.Kind = KindSubmission; i.Submission = nil }, "submission payload"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			it := sampleItem("https://a.example/1")
			it.CandidateClaims = append([]claims.Candidate{}, it.CandidateClaims...)
			c.f(&it)
			err := Validate(it)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want error containing %q, got %v", c.want, err)
			}
		})
	}
	sub := sampleItem("https://a.example/1")
	sub.Kind = KindSubmission
	sub.Document = nil
	sub.Submission = json.RawMessage(`{"kind":"source","url":"https://x.example"}`)
	sub.ID = NewItemID(KindSubmission, "src-a", "https://x.example", "")
	if err := Validate(sub); err != nil {
		t.Fatalf("submission item: %v", err)
	}
}

func TestListReportsCorruptLines(t *testing.T) {
	dir := t.TempDir()
	q := NewQueue(dir)
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-ingest.jsonl"), []byte("{\"id\":\"rq-1\"}\n\nnot json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := q.List(); err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Fatalf("corrupt line must be reported with its position: %v", err)
	}
	empty := NewQueue(filepath.Join(dir, "missing"))
	items, err := empty.List()
	if err != nil || len(items) != 0 {
		t.Fatalf("missing dir must list nothing: %v", err)
	}
}

func TestScoreQualityAndIDs(t *testing.T) {
	if s := ScoreQuality(1, nil, "allowed"); s != 1 {
		t.Fatalf("tier 1 = %v", s)
	}
	if s := ScoreQuality(2, []schema.ConflictLabel{"developer_self_report", "commercial_interest"}, "not_applicable"); s < 0.59 || s > 0.61 {
		t.Fatalf("tier 2 with two conflicts = %v", s)
	}
	if s := ScoreQuality(1, nil, "disallowed"); s != 0 {
		t.Fatalf("robots disallowed = %v", s)
	}
	if s := ScoreQuality(5, []schema.ConflictLabel{"advocacy_context"}, "allowed"); s != 0 {
		t.Fatalf("tier 5 = %v", s)
	}
	if s := ScoreQuality(4, []schema.ConflictLabel{"none_known"}, "allowed"); s != 0.2 {
		t.Fatalf("none_known must not penalise: %v", s)
	}
	a := NewItemID(KindDocument, "s", "https://a.example/1", "h")
	if a != NewItemID(KindDocument, "s", "https://a.example/1", "h") || a == NewItemID(KindCandidateClaim, "s", "https://a.example/1", "h") {
		t.Fatal("ids must be stable and kind-specific")
	}
	if !reItemID.MatchString(a) {
		t.Fatalf("id format: %s", a)
	}
}

// TestListIgnoresForeignFiles: the public API appends its own record shape to
// submissions.jsonl in the same directory, and operators may leave notes
// there. Neither may abort an ingestion run or be mistaken for a queue item.
func TestListIgnoresForeignFiles(t *testing.T) {
	dir := t.TempDir()
	q := NewQueue(dir)
	if _, _, err := q.Append("2026-09-26", []ReviewItem{sampleItem("https://a.example/1")}); err != nil {
		t.Fatal(err)
	}
	apiRecord := `{"id":"sub-0123456789abcdef","kind":"source","submitted_at":"2026-09-26T00:00:00Z","payload":{"canonical_url":"https://x.example","title":"t"},"status":"received"}` + "\n"
	os.WriteFile(filepath.Join(dir, "submissions.jsonl"), []byte(apiRecord), 0o644)
	os.WriteFile(filepath.Join(dir, "notes.jsonl"), []byte("not json at all\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "2026-09-27-ingest.jsonl.bak"), []byte("garbage\n"), 0o644)
	files, err := q.Files()
	if err != nil || len(files) != 1 || filepath.Base(files[0]) != "2026-09-26-ingest.jsonl" {
		t.Fatalf("files = %v err=%v", files, err)
	}
	items, err := q.List()
	if err != nil || len(items) != 1 || items[0].Kind != KindDocument {
		t.Fatalf("list = %d items, err=%v", len(items), err)
	}
	subs, err := q.ListKind(KindSubmission)
	if err != nil || len(subs) != 0 {
		t.Fatalf("api submissions are not queue items: %d %v", len(subs), err)
	}
	if p, err := q.FileFor("2026-09-26"); err != nil || p != filepath.Join(dir, "2026-09-26-ingest.jsonl") {
		t.Fatalf("FileFor = %q %v", p, err)
	}
	if _, err := q.FileFor("../x"); err == nil {
		t.Fatal("path-like date must be refused")
	}
}
