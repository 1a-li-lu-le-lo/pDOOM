// Copyright NU Cybernetics. p(DOOM) — research prototype.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/1a-li-lu-le-lo/pdoom/internal/claims"
	"github.com/1a-li-lu-le-lo/pdoom/internal/dedup"
	"github.com/1a-li-lu-le-lo/pdoom/internal/fetch"
	"github.com/1a-li-lu-le-lo/pdoom/internal/observability"
	"github.com/1a-li-lu-le-lo/pdoom/internal/parsing"
	"github.com/1a-li-lu-le-lo/pdoom/internal/review"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/sources"
)

// Fetcher is the network dependency of the pipeline (fetch.SafeClient in
// production, a fake in tests).
type Fetcher interface {
	Get(ctx context.Context, rawURL string) (*fetch.Result, error)
}

// Pipeline runs DISCOVER → PERMISSION CHECK → FETCH → HASH → ARCHIVE METADATA →
// PARSE → NORMALIZE → DEDUPLICATE → CLASSIFY → EXTRACT CLAIMS → LINK PRIMARY
// SOURCES → SCORE SOURCE QUALITY → DETECT CONTRADICTIONS (stub) → HUMAN REVIEW
// QUEUE for every allowed source. Its only side effect is appending review
// items; with DryRun it has none.
type Pipeline struct {
	Config  *sources.Config
	Fetcher Fetcher // nil: plan only, no network
	Queue   *review.Queue
	Metrics *observability.Registry
	Logger  *slog.Logger
	Out     io.Writer
	// Now is the caller-supplied RFC 3339 UTC timestamp used for retrieved_at,
	// created_at and the queue file name. The pipeline never reads the clock.
	Now    string
	DryRun bool
	Since  string // optional YYYY-MM-DD; documents published earlier are skipped
	OnlyID string // optional source id filter
}

// SourceOutcome summarizes one source.
type SourceOutcome struct {
	SourceID       string `json:"source_id"`
	URL            string `json:"url"`
	Skipped        bool   `json:"skipped"`
	Reason         string `json:"reason,omitempty"`
	Fetched        bool   `json:"fetched"`
	Bytes          int    `json:"bytes"`
	Documents      int    `json:"documents"`
	Duplicates     int    `json:"duplicates"`
	NearDuplicates int    `json:"near_duplicates"`
	FilteredSince  int    `json:"filtered_before_since"`
	Claims         int    `json:"candidate_claims"`
	Queued         int    `json:"queued"`
	WouldQueue     int    `json:"would_queue"`
	Error          string `json:"error,omitempty"`
}

// Summary is the machine-readable result of a run.
type Summary struct {
	Now       string                 `json:"now"`
	DryRun    bool                   `json:"dry_run"`
	Network   bool                   `json:"network"`
	Sources   []SourceOutcome        `json:"sources"`
	Queued    int                    `json:"queued"`
	QueueFile string                 `json:"queue_file,omitempty"`
	Metrics   observability.Snapshot `json:"metrics"`
}

var reURL = regexp.MustCompile(`https?://[^\s<>"'\)\]]+`)

// Run executes the pipeline.
func (p *Pipeline) Run(ctx context.Context) (*Summary, error) {
	if p.Config == nil {
		return nil, fmt.Errorf("ingest: nil config")
	}
	if p.Metrics == nil {
		p.Metrics = observability.New()
	}
	if p.Logger == nil {
		p.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if p.Out == nil {
		p.Out = io.Discard
	}
	now, err := time.Parse(time.RFC3339, p.Now)
	if err != nil {
		return nil, fmt.Errorf("ingest: --now must be RFC 3339 (e.g. 2026-09-26T00:00:00Z): %w", err)
	}
	date := now.UTC().Format("2006-01-02")
	if p.Since != "" {
		if _, err := time.Parse("2006-01-02", p.Since); err != nil {
			return nil, fmt.Errorf("ingest: --since must be YYYY-MM-DD: %w", err)
		}
	}
	if p.OnlyID != "" {
		if _, ok := p.Config.ByID(p.OnlyID); !ok {
			return nil, fmt.Errorf("ingest: --source %q is not in the config", p.OnlyID)
		}
	}
	sum := &Summary{Now: p.Now, DryRun: p.DryRun, Network: p.Fetcher != nil}

	// Seed the clusterer with what is already queued so a document that was
	// queued on an earlier run is recognised as a duplicate.
	clusterer := dedup.NewClusterer(0)
	if p.Queue != nil {
		existing, err := p.Queue.List()
		if err != nil {
			return nil, fmt.Errorf("ingest: read existing queue: %w", err)
		}
		for _, it := range existing {
			if it.Document != nil {
				clusterer.Assign(it.ID, it.Document.CanonicalURL, it.Document.ContentHash, it.Document.Title, it.Document.Text)
			}
		}
	}

	var pending []review.ReviewItem
	for _, src := range p.Config.Sources {
		if p.OnlyID != "" && src.ID != p.OnlyID {
			continue
		}
		// The run timeout (--timeout) cancels ctx: stop between sources and
		// write nothing, so a cancelled run has no side effect at all.
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("ingest: run cancelled before source %s; nothing was written: %w", src.ID, err)
		}
		p.Metrics.Inc(observability.CounterSourcesConsidered, 1)
		oc := SourceOutcome{SourceID: src.ID, URL: src.URL}
		// PERMISSION CHECK (configuration): the allowlist is the first gate.
		if !src.Allowed {
			oc.Skipped = true
			oc.Reason = "not allowed: " + src.Reason
			p.Metrics.Inc(observability.CounterSourcesSkipped, 1)
			sum.Sources = append(sum.Sources, oc)
			fmt.Fprintf(p.Out, "skip   %-28s %s\n", src.ID, firstLine(oc.Reason))
			continue
		}
		if p.Fetcher == nil {
			oc.Skipped = true
			oc.Reason = fmt.Sprintf("plan only (no --allow-network): would fetch up to %d items", src.MaxItemsPerRun)
			sum.Sources = append(sum.Sources, oc)
			fmt.Fprintf(p.Out, "plan   %-28s %s (max %d items, parser %s, tier %d)\n", src.ID, src.URL, src.MaxItemsPerRun, src.Parser, src.Tier)
			continue
		}
		items := p.runSource(ctx, src, clusterer, &oc)
		if p.DryRun {
			oc.WouldQueue = len(items)
			for _, it := range items {
				fmt.Fprintf(p.Out, "would-queue %s %-28s %s\n", it.ID, src.ID, firstLine(it.Document.Title))
			}
			p.Metrics.Inc("dry_run_items", int64(len(items)))
		} else {
			pending = append(pending, items...)
		}
		sum.Sources = append(sum.Sources, oc)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("ingest: run cancelled; nothing was written: %w", err)
	}
	// HUMAN REVIEW QUEUE: the single write of the pipeline.
	if !p.DryRun && len(pending) > 0 {
		if p.Queue == nil {
			return nil, fmt.Errorf("ingest: no review queue configured")
		}
		path, written, err := p.Queue.Append(date, pending)
		if err != nil {
			return nil, fmt.Errorf("ingest: append review queue: %w", err)
		}
		sum.Queued = written
		sum.QueueFile = path
		p.Metrics.Inc(observability.CounterQueued, int64(written))
		// Attribute queued counts back to sources for the summary.
		perSource := map[string]int{}
		for _, it := range pending {
			perSource[it.SourceID]++
		}
		for i := range sum.Sources {
			sum.Sources[i].Queued = perSource[sum.Sources[i].SourceID]
		}
		fmt.Fprintf(p.Out, "queued %d item(s) to %s\n", written, path)
	}
	sum.Metrics = p.Metrics.Snapshot()
	observability.LogSnapshot(p.Logger, "ingest summary", sum.Metrics)
	fmt.Fprintf(p.Out, "summary %s\n", sum.Metrics.String())
	return sum, nil
}

// runSource performs the per-source stages and returns the items to queue.
func (p *Pipeline) runSource(ctx context.Context, src sources.Source, clusterer *dedup.Clusterer, oc *SourceOutcome) []review.ReviewItem {
	log := p.Logger.With("source", src.ID)
	// FETCH (robots policy, SSRF guard, size and type caps live in the fetcher).
	p.Metrics.Inc(observability.CounterFetches, 1)
	res, err := p.Fetcher.Get(ctx, src.URL)
	if err != nil {
		if errors.Is(err, fetch.ErrRobotsDisallowed) {
			p.Metrics.Inc(observability.CounterRobotsDenials, 1)
		} else {
			p.Metrics.Inc(observability.CounterFetchErrors, 1)
		}
		oc.Error = err.Error()
		log.Warn("fetch failed", "error", err)
		fmt.Fprintf(p.Out, "error  %-28s %s\n", src.ID, firstLine(err.Error()))
		return nil
	}
	oc.Fetched = true
	oc.Bytes = res.Bytes
	p.Metrics.Add(observability.GaugeBytesFetched, float64(res.Bytes))
	// HASH + ARCHIVE METADATA.
	archive := &review.Archive{FetchURL: src.URL, FinalURL: res.FinalURL, SHA256: res.SHA256, ContentType: res.ContentType, Bytes: res.Bytes, RetrievedAt: p.Now}
	// PARSE + NORMALIZE.
	docs, err := parsing.Parse(src.Parser, src.ID, src.URL, res.Body, p.Now)
	if err != nil {
		p.Metrics.Inc(observability.CounterParserErrors, 1)
		oc.Error = err.Error()
		log.Warn("parse failed", "error", err)
		fmt.Fprintf(p.Out, "error  %-28s %s\n", src.ID, firstLine(err.Error()))
		return nil
	}
	if src.MaxItemsPerRun > 0 && len(docs) > src.MaxItemsPerRun {
		docs = docs[:src.MaxItemsPerRun]
	}
	var items []review.ReviewItem
	for _, doc := range docs {
		if ctx.Err() != nil {
			// The run is over; Run reports the cancellation and writes nothing.
			return nil
		}
		if p.Since != "" && doc.PublishedAt != nil && len(*doc.PublishedAt) >= 10 && (*doc.PublishedAt)[:10] < p.Since {
			oc.FilteredSince++
			p.Metrics.Inc(observability.CounterFilteredSince, 1)
			continue
		}
		oc.Documents++
		p.Metrics.Inc(observability.CounterDocuments, 1)
		id := review.NewItemID(review.KindDocument, src.ID, doc.CanonicalURL, doc.ContentHash)
		// DEDUPLICATE.
		a := clusterer.Assign(id, doc.CanonicalURL, doc.ContentHash, doc.Title, doc.Text)
		if a.Duplicate && a.Match != dedup.MatchSimhash {
			oc.Duplicates++
			p.Metrics.Inc(observability.CounterDuplicates, 1)
			log.Debug("duplicate skipped", "url", doc.CanonicalURL, "match", string(a.Match), "cluster", a.ClusterID)
			continue
		}
		if a.Duplicate {
			oc.NearDuplicates++
			p.Metrics.Inc(observability.CounterNearDuplicates, 1)
		}
		// CLASSIFY (rule-based, category-level tags).
		tags := classify(doc)
		// EXTRACT CLAIMS (rule-based skeleton; LLM classification is off).
		cands := claims.Extract(doc.Text)
		oc.Claims += len(cands)
		p.Metrics.Inc(observability.CounterCandidateClaims, int64(len(cands)))
		// LINK PRIMARY SOURCES.
		links := primaryLinks(doc)
		// SCORE SOURCE QUALITY.
		q := review.Quality{
			Tier: src.Tier, Conflicts: conflictsOrEmpty(src), RobotsStatus: src.RobotsStatus,
			DuplicateCluster: a.ClusterID, Duplicate: a.Duplicate,
			Score: review.ScoreQuality(src.Tier, src.Conflicts, src.RobotsStatus),
		}
		kind := review.KindDocument
		if len(cands) > 0 {
			kind = review.KindCandidateClaim
		}
		d := doc
		items = append(items, review.ReviewItem{
			ID:                 review.NewItemID(kind, src.ID, doc.CanonicalURL, doc.ContentHash),
			Kind:               kind,
			CreatedAt:          p.Now,
			SourceID:           src.ID,
			Document:           &d,
			CandidateClaims:    cands,
			Quality:            q,
			Archive:            archive,
			Classification:     tags,
			PrimarySourceLinks: links,
			// DETECT CONTRADICTIONS: stub. Contradiction detection needs the
			// reviewed claim graph of a snapshot and is out of scope here.
			Contradictions:     []string{},
			ContradictionCheck: review.ContradictionCheckNotImplemented,
			Status:             review.StatusPending,
		})
	}
	return items
}

func conflictsOrEmpty(src sources.Source) []schema.ConflictLabel {
	if src.Conflicts == nil {
		return []schema.ConflictLabel{}
	}
	return src.Conflicts
}

// classifyRules map category tags to lowercase keywords. They are coarse on
// purpose: the tag is a reviewer hint, never a decision.
var classifyRules = []struct {
	tag      string
	keywords []string
}{
	{"capability", []string{"benchmark", "evaluation", "task horizon", "capabilit", "frontier model", "scaling", "compute"}},
	{"autonomy", []string{"agent", "autonomous", "tool use", "long-horizon"}},
	{"governance", []string{"regulat", "policy", "executive order", "standard", "framework", "governance", "treaty", "oversight", "legislat"}},
	{"safety", []string{"alignment", "safety", "interpretab", "red team", "control", "safeguard"}},
	{"security", []string{"security", "advisory", "vulnerab", "breach", "cve-"}},
	{"incident", []string{"incident", "failure", "harm", "outage", "misuse", "accident"}},
	{"forecast", []string{"forecast", "predict", "probability", "survey", "timeline", "estimate"}},
}

// classify returns sorted category tags for a document.
func classify(doc parsing.Document) []string {
	text := strings.ToLower(doc.Title + " " + doc.Text)
	var tags []string
	for _, r := range classifyRules {
		for _, k := range r.keywords {
			if strings.Contains(text, k) {
				tags = append(tags, r.tag)
				break
			}
		}
	}
	if len(tags) == 0 {
		tags = []string{"uncategorized"}
	}
	sort.Strings(tags)
	return tags
}

// primaryHosts decides whether a link points at a primary source.
func primaryHost(host string) bool {
	host = strings.ToLower(host)
	if host == "arxiv.org" || host == "doi.org" {
		return true
	}
	for _, suffix := range []string{".gov", ".gov.uk", ".europa.eu", ".oecd.org", ".int", ".mil"} {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

// primaryLinks collects canonical links to primary sources found in the
// document (its own URL first), at most ten, sorted and de-duplicated.
func primaryLinks(doc parsing.Document) []string {
	set := map[string]bool{}
	consider := func(raw string) {
		c, err := dedup.CanonicalURL(raw)
		if err != nil {
			return
		}
		u, err := url.Parse(c)
		if err != nil || !primaryHost(u.Hostname()) {
			return
		}
		set[c] = true
	}
	consider(doc.CanonicalURL)
	for _, m := range reURL.FindAllString(doc.Text, 50) {
		consider(strings.TrimRight(m, ".,;:"))
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	if len(out) > 10 {
		out = out[:10]
	}
	return out
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 160 {
		s = s[:160] + "…"
	}
	return s
}
