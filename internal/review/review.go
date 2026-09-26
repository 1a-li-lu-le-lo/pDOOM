// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package review is the human review queue: append-only JSONL files under
// data/review-queue/, written through internal/storage. It is the only place
// the ingestion pipeline may write. Items in the queue never change a
// snapshot, a release, a weight or a configuration; a human reads them and,
// if they are worth keeping, adds the evidence to a NEW data snapshot through
// the normal research workflow.
//
// The pipeline owns the files named <date>-ingest.jsonl. The same directory
// also holds submissions.jsonl, which the public API appends community
// submissions to in its own record shape ({id: "sub-…", kind: source |
// correction, submitted_at, payload, contact, status: "received"}); the
// pipeline neither reads nor writes that file, so a foreign record can never
// abort an ingestion run.
package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/claims"
	"github.com/1a-li-lu-le-lo/pdoom/internal/parsing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/storage"
)

// Kind is the type of a review item.
type Kind string

// Kinds.
const (
	KindDocument       Kind = "document"
	KindCandidateClaim Kind = "candidate_claim"
	KindSubmission     Kind = "submission"
)

// Valid reports whether the kind is allowed.
func (k Kind) Valid() bool {
	return k == KindDocument || k == KindCandidateClaim || k == KindSubmission
}

// StatusPending is the only status the pipeline writes.
const StatusPending = "pending"

// ContradictionCheckNotImplemented marks the contradiction detector stub.
const ContradictionCheckNotImplemented = "not_implemented"

// Quality is the source-quality summary attached to every item. It is a triage
// aid for reviewers, never a model input.
type Quality struct {
	Tier             schema.SourceTier      `json:"tier"`
	Conflicts        []schema.ConflictLabel `json:"conflicts"`
	RobotsStatus     schema.RobotsStatus    `json:"robots_status"`
	DuplicateCluster string                 `json:"duplicate_cluster"`
	Duplicate        bool                   `json:"duplicate"`
	// Score is a coarse triage rank in [0, 1]: tier rank minus a penalty per
	// declared conflict, zero when robots did not allow the fetch.
	Score float64 `json:"score"`
}

// Archive is the metadata of the fetched resource an item came from.
type Archive struct {
	FetchURL    string `json:"fetch_url"`
	FinalURL    string `json:"final_url"`
	SHA256      string `json:"sha256"`
	ContentType string `json:"content_type"`
	Bytes       int    `json:"bytes"`
	RetrievedAt string `json:"retrieved_at"`
}

// ReviewItem is one line of a queue file.
type ReviewItem struct {
	ID                 string             `json:"id"`
	Kind               Kind               `json:"kind"`
	CreatedAt          string             `json:"created_at"`
	SourceID           string             `json:"source_id"`
	Document           *parsing.Document  `json:"document"`
	CandidateClaims    []claims.Candidate `json:"candidate_claims"`
	Quality            Quality            `json:"quality"`
	Archive            *Archive           `json:"archive"`
	Classification     []string           `json:"classification"`
	PrimarySourceLinks []string           `json:"primary_source_links"`
	Contradictions     []string           `json:"contradictions"`
	ContradictionCheck string             `json:"contradiction_check"`
	Submission         json.RawMessage    `json:"submission,omitempty"`
	Status             string             `json:"status"`
}

var (
	reDate      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	reItemID    = regexp.MustCompile(`^rq-[0-9a-f]{16}$`)
	reQueueFile = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-ingest\.jsonl$`)
)

// ScoreQuality computes Quality.Score from tier, conflicts and robots status.
func ScoreQuality(tier schema.SourceTier, conflicts []schema.ConflictLabel, robots schema.RobotsStatus) float64 {
	if robots != "allowed" && robots != "not_applicable" {
		return 0
	}
	var base float64
	switch tier {
	case 1:
		base = 1.0
	case 2:
		base = 0.8
	case 3:
		base = 0.5
	case 4:
		base = 0.2
	default:
		base = 0
	}
	for _, c := range conflicts {
		if c != "none_known" && c != "" {
			base -= 0.1
		}
	}
	if base < 0 {
		base = 0
	}
	return base
}

// NewItemID derives a stable id from the item's identity fields so that the
// same document queued twice gets the same id.
func NewItemID(kind Kind, sourceID, canonicalURL, contentHash string) string {
	sum := schema.SHA256Hex([]byte(string(kind) + "\x00" + sourceID + "\x00" + canonicalURL + "\x00" + contentHash))
	return "rq-" + sum[:16]
}

// Validate checks an item before it is written.
func Validate(it ReviewItem) error {
	var errs []error
	if !reItemID.MatchString(it.ID) {
		errs = append(errs, fmt.Errorf("id %q must match %s", it.ID, reItemID))
	}
	if !it.Kind.Valid() {
		errs = append(errs, fmt.Errorf("kind %q invalid", it.Kind))
	}
	if strings.TrimSpace(it.CreatedAt) == "" {
		errs = append(errs, fmt.Errorf("created_at is required"))
	}
	if strings.TrimSpace(it.SourceID) == "" {
		errs = append(errs, fmt.Errorf("source_id is required"))
	}
	if it.Status != StatusPending {
		errs = append(errs, fmt.Errorf("status must be %q, got %q", StatusPending, it.Status))
	}
	switch it.Kind {
	case KindDocument, KindCandidateClaim:
		if it.Document == nil {
			errs = append(errs, fmt.Errorf("%s items require a document", it.Kind))
		} else if it.Document.URL == "" || it.Document.ContentHash == "" {
			errs = append(errs, fmt.Errorf("document must have url and content_hash"))
		}
	case KindSubmission:
		if len(it.Submission) == 0 {
			errs = append(errs, fmt.Errorf("submission items require a submission payload"))
		}
	}
	if !it.Quality.Tier.Valid() {
		errs = append(errs, fmt.Errorf("quality.tier %d must be 1..5", it.Quality.Tier))
	}
	if !it.Quality.RobotsStatus.Valid() {
		errs = append(errs, fmt.Errorf("quality.robots_status %q invalid", it.Quality.RobotsStatus))
	}
	for _, c := range it.CandidateClaims {
		if c.Status != "candidate" || c.ModelUseStatus != schema.ModelUseExcluded {
			errs = append(errs, fmt.Errorf("candidate claim %q must be candidate/excluded", c.DirectQuotePointer))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("review: invalid item: %w", errors.Join(errs...))
	}
	return nil
}

// Queue is the pipeline's view of the review-queue directory: the
// <date>-ingest.jsonl files, read and appended through internal/storage.
// Other files in the directory are ignored.
type Queue struct {
	Dir string
}

// NewQueue returns a queue rooted at dir (created on first append).
func NewQueue(dir string) *Queue { return &Queue{Dir: dir} }

func (q *Queue) store() (*storage.Store, error) {
	st, err := storage.New(q.Dir)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	return st, nil
}

// fileName returns the queue file name for a run dated date (YYYY-MM-DD).
func fileName(date string) (string, error) {
	if !reDate.MatchString(date) {
		return "", fmt.Errorf("review: date %q must be YYYY-MM-DD", date)
	}
	return date + "-ingest.jsonl", nil
}

// FileFor returns the file items of a run dated date (YYYY-MM-DD) go to.
func (q *Queue) FileFor(date string) (string, error) {
	name, err := fileName(date)
	if err != nil {
		return "", err
	}
	return filepath.Join(q.Dir, name), nil
}

// Append validates every item, then appends them as canonical JSON lines to
// <date>-ingest.jsonl. Nothing is written if any item is invalid. Items whose
// id is already present in that file are skipped (append-only, idempotent).
func (q *Queue) Append(date string, items []ReviewItem) (path string, written int, err error) {
	name, err := fileName(date)
	if err != nil {
		return "", 0, err
	}
	st, err := q.store()
	if err != nil {
		return "", 0, err
	}
	path = filepath.Join(q.Dir, name)
	for i, it := range items {
		if err := Validate(it); err != nil {
			return path, 0, fmt.Errorf("item %d: %w", i, err)
		}
	}
	existing, err := storage.ReadAll[ReviewItem](st, name)
	if err != nil {
		return path, 0, fmt.Errorf("review: read queue: %w", err)
	}
	seen := map[string]bool{}
	for _, e := range existing {
		seen[e.ID] = true
	}
	var records []any
	for _, it := range items {
		if seen[it.ID] {
			continue
		}
		seen[it.ID] = true
		records = append(records, it)
	}
	if len(records) == 0 {
		return path, 0, nil
	}
	path, written, err = st.Append(name, records)
	if err != nil {
		return path, 0, fmt.Errorf("review: append queue: %w", err)
	}
	return path, written, nil
}

// Files lists the pipeline's queue files (<date>-ingest.jsonl, sorted).
func (q *Queue) Files() ([]string, error) {
	st, err := q.store()
	if err != nil {
		return nil, err
	}
	files, err := st.Files(reQueueFile.MatchString)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	return files, nil
}

// List reads every item of every queue file in file order. Lines that do not
// decode are reported as an error naming the file and line; nothing is mutated.
func (q *Queue) List() ([]ReviewItem, error) {
	st, err := q.store()
	if err != nil {
		return nil, err
	}
	files, err := q.Files()
	if err != nil {
		return nil, err
	}
	var out []ReviewItem
	for _, f := range files {
		items, err := storage.ReadAll[ReviewItem](st, filepath.Base(f))
		if err != nil {
			return nil, fmt.Errorf("review: read queue: %w", err)
		}
		out = append(out, items...)
	}
	return out, nil
}

// ListKind filters List by kind.
func (q *Queue) ListKind(kind Kind) ([]ReviewItem, error) {
	all, err := q.List()
	if err != nil {
		return nil, err
	}
	var out []ReviewItem
	for _, it := range all {
		if it.Kind == kind {
			out = append(out, it)
		}
	}
	return out, nil
}
