// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Submission is one line of data/review-queue/submissions.jsonl. It is the
// only thing a public submission ever writes; snapshots and releases are
// never touched.
type Submission struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	SubmittedAt string  `json:"submitted_at"`
	Payload     any     `json:"payload"`
	Contact     *string `json:"contact,omitempty"`
	Status      string  `json:"status"`
}

// SubmissionRequest is the body of the two submission routes.
type SubmissionRequest struct {
	Payload json.RawMessage `json:"payload"`
	Contact *string         `json:"contact"`
}

// SourceSubmissionPayload mirrors SourceSubmissionPayloadSchema in
// packages/schemas/src/lab.ts.
type SourceSubmissionPayload struct {
	CanonicalURL    string  `json:"canonical_url"`
	Title           string  `json:"title"`
	Publisher       *string `json:"publisher,omitempty"`
	DatePublished   *string `json:"date_published,omitempty"`
	WhyRelevant     string  `json:"why_relevant"`
	ClaimedEvidence string  `json:"claimed_evidence"`
}

// CorrectionSubmissionPayload mirrors CorrectionSubmissionPayloadSchema.
type CorrectionSubmissionPayload struct {
	TargetID    string  `json:"target_id"`
	Field       *string `json:"field,omitempty"`
	Correction  string  `json:"correction"`
	EvidenceURL *string `json:"evidence_url,omitempty"`
}

// SubmissionResponse acknowledges a queued submission.
type SubmissionResponse struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	SubmittedAt string `json:"submitted_at"`
	Status      string `json:"status"`
	Note        string `json:"note"`
}

const (
	submissionsFile = "submissions.jsonl"
	maxShortText    = 500
	maxLongText     = 4000
	maxURLText      = 2048
	maxContactText  = 320
	submissionNote  = "Received into the human review queue. Nothing published changes until reviewers assemble a new snapshot and a signed release is promoted."
)

var (
	reDateLoose   = regexp.MustCompile(`^\d{4}(?:-\d{2}(?:-\d{2})?)?$`)
	reTargetID    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._@+:/-]{0,199}$`)
	reFieldName   = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,99}$`)
	reSubmitKinds = map[string]bool{"source": true, "correction": true}
)

// stripControl removes control characters (C0, C1 and DEL) and the Unicode
// line/paragraph separators, keeping newline and tab, then trims surrounding
// whitespace. U+FFFD (replacement character) is kept: it marks where a client
// sent invalid UTF-8 and is useful to reviewers.
func stripControl(s string) string {
	if !strings.ContainsFunc(s, isDisallowedControl) {
		return strings.TrimSpace(s)
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isDisallowedControl(r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func isDisallowedControl(r rune) bool {
	if r == '\n' || r == '\t' {
		return false
	}
	return unicode.IsControl(r) || r == '\u2028' || r == '\u2029'
}

func cleanPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := stripControl(*p)
	if v == "" {
		return nil
	}
	return &v
}

func checkLen(problems *[]string, name, v string, max int) {
	if len(v) > max {
		*problems = append(*problems, fmt.Sprintf("%s exceeds %d bytes", name, max))
	}
}

func checkHTTPURL(problems *[]string, name, v string) {
	if len(v) > maxURLText {
		*problems = append(*problems, name+" exceeds "+strconv.Itoa(maxURLText)+" bytes")
		return
	}
	u, err := url.Parse(v)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		*problems = append(*problems, name+" must be an absolute http(s) URL")
	}
}

func validateSourcePayload(p *SourceSubmissionPayload) []string {
	var problems []string
	p.CanonicalURL = stripControl(p.CanonicalURL)
	p.Title = stripControl(p.Title)
	p.Publisher = cleanPtr(p.Publisher)
	p.DatePublished = cleanPtr(p.DatePublished)
	p.WhyRelevant = stripControl(p.WhyRelevant)
	p.ClaimedEvidence = stripControl(p.ClaimedEvidence)
	if p.CanonicalURL == "" {
		problems = append(problems, "canonical_url is required")
	} else {
		checkHTTPURL(&problems, "canonical_url", p.CanonicalURL)
	}
	if p.Title == "" {
		problems = append(problems, "title is required")
	}
	checkLen(&problems, "title", p.Title, maxShortText)
	if p.Publisher != nil {
		checkLen(&problems, "publisher", *p.Publisher, maxShortText)
	}
	if p.DatePublished != nil && !reDateLoose.MatchString(*p.DatePublished) {
		problems = append(problems, "date_published must be YYYY, YYYY-MM or YYYY-MM-DD")
	}
	checkLen(&problems, "why_relevant", p.WhyRelevant, maxLongText)
	checkLen(&problems, "claimed_evidence", p.ClaimedEvidence, maxLongText)
	return problems
}

func validateCorrectionPayload(p *CorrectionSubmissionPayload) []string {
	var problems []string
	p.TargetID = stripControl(p.TargetID)
	p.Field = cleanPtr(p.Field)
	p.Correction = stripControl(p.Correction)
	p.EvidenceURL = cleanPtr(p.EvidenceURL)
	if p.TargetID == "" {
		problems = append(problems, "target_id is required")
	} else if !reTargetID.MatchString(p.TargetID) {
		problems = append(problems, "target_id must be an entity or estimate id (letters, digits, . _ @ + : / -; ≤ 200 bytes)")
	}
	if p.Field != nil && !reFieldName.MatchString(*p.Field) {
		problems = append(problems, "field must be a snake_case field name")
	}
	if p.Correction == "" {
		problems = append(problems, "correction is required")
	}
	checkLen(&problems, "correction", p.Correction, maxLongText)
	if p.EvidenceURL != nil {
		checkHTTPURL(&problems, "evidence_url", *p.EvidenceURL)
	}
	return problems
}

func (s *server) handleSubmitSource(w http.ResponseWriter, r *http.Request, _ *state) {
	var req SubmissionRequest
	if !s.decodeBody(w, r, &req) {
		return
	}
	var payload SourceSubmissionPayload
	if !s.decodePayload(w, r, req.Payload, &payload, "canonical_url", "title", "why_relevant", "claimed_evidence") {
		return
	}
	if problems := validateSourcePayload(&payload); len(problems) > 0 {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", strings.Join(problems, "; "))
		return
	}
	s.queueSubmission(w, r, "source", payload, req.Contact)
}

func (s *server) handleSubmitCorrection(w http.ResponseWriter, r *http.Request, _ *state) {
	var req SubmissionRequest
	if !s.decodeBody(w, r, &req) {
		return
	}
	var payload CorrectionSubmissionPayload
	if !s.decodePayload(w, r, req.Payload, &payload, "target_id", "correction") {
		return
	}
	if problems := validateCorrectionPayload(&payload); len(problems) > 0 {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", strings.Join(problems, "; "))
		return
	}
	s.queueSubmission(w, r, "correction", payload, req.Contact)
}

// decodePayload strictly decodes the payload object of a submission and
// checks that every required key is present (an absent string would otherwise
// decode as empty).
func (s *server) decodePayload(w http.ResponseWriter, r *http.Request, raw json.RawMessage, v any, required ...string) bool {
	if len(bytes.TrimSpace(raw)) == 0 {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "payload object is required")
		return false
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "payload must be a JSON object")
		return false
	}
	var missing []string
	for _, k := range required {
		if _, ok := keys[k]; !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "payload lacks required fields: "+strings.Join(missing, ", "))
		return false
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "invalid payload: "+jsonErrorDetail(err))
		return false
	}
	if _, err := dec.Token(); err != io.EOF {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "invalid payload: trailing data")
		return false
	}
	return true
}

// submissionSeq makes ids unique even if two identical payloads arrive in the
// same second and the random nonce were ever to repeat.
var submissionSeq atomic.Uint64

// submissionID derives the id of a queued submission. The hash covers the
// kind, the timestamp, the canonical payload and a per-request nonce (random
// bytes plus a process counter), so identical payloads submitted in the same
// second, whether retries or duplicates, still receive distinct ids.
func submissionID(kind, submittedAt string, canonicalPayload []byte) string {
	nonce := hex.EncodeToString(randomBytes(8)) + "-" + strconv.FormatUint(submissionSeq.Add(1), 10)
	return "sub-" + schema.SHA256Hex([]byte(kind + "\n" + submittedAt + "\n" + nonce + "\n" + string(canonicalPayload)))[:16]
}

// queueSubmission appends the submission to the review queue. The server
// clock is used for submitted_at: a submission is a runtime request, not part
// of a release.
func (s *server) queueSubmission(w http.ResponseWriter, r *http.Request, kind string, payload any, contact *string) {
	if !reSubmitKinds[kind] {
		s.writeError(w, r, http.StatusInternalServerError, "internal", "unknown submission kind")
		return
	}
	contact = cleanPtr(contact)
	if contact != nil && len(*contact) > maxContactText {
		s.writeError(w, r, http.StatusBadRequest, "bad_request", "contact exceeds "+strconv.Itoa(maxContactText)+" bytes")
		return
	}
	submittedAt := s.now().UTC().Format(time.RFC3339)
	canonicalPayload, err := schema.CanonicalJSON(payload)
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "internal", "payload could not be encoded")
		return
	}
	id := submissionID(kind, submittedAt, canonicalPayload)
	sub := Submission{ID: id, Kind: kind, SubmittedAt: submittedAt, Payload: payload, Contact: contact, Status: "received"}
	line, err := schema.CanonicalJSON(sub)
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "internal", "submission could not be encoded")
		return
	}
	if err := s.appendSubmission(line); err != nil {
		s.log.Error("append submission", "error", err.Error())
		s.writeError(w, r, http.StatusInternalServerError, "internal", "submission could not be stored")
		return
	}
	s.log.Info("submission queued", "submission_id", id, "kind", kind)
	s.writeJSON(w, http.StatusAccepted, SubmissionResponse{ID: id, Kind: kind, SubmittedAt: submittedAt, Status: "received", Note: submissionNote})
}

// SubmissionsPath returns the JSONL file public submissions are appended to.
func SubmissionsPath(reviewDir string) string { return filepath.Join(reviewDir, submissionsFile) }

func (s *server) appendSubmission(line []byte) error {
	s.submitMu.Lock()
	defer s.submitMu.Unlock()
	if err := os.MkdirAll(s.deps.Paths.Review, 0o755); err != nil {
		return fmt.Errorf("review queue dir: %w", err)
	}
	f, err := os.OpenFile(SubmissionsPath(s.deps.Paths.Review), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open review queue: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write review queue: %w", err)
	}
	return nil
}
