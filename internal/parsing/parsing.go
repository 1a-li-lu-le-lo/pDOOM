// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package parsing turns fetched bodies into normalized Documents. Three parsers
// exist: feeds (RSS/Atom/JSON Feed via gofeed), html_text (a small
// single-pass stdlib tokenizer that drops script/style and collapses
// whitespace, linear in the input size) and a JSON parser for GitHub
// advisories.
//
// Doctrine: retrieved content is untrusted data. A Document carries text and
// metadata only; it has no field that could express a tier, an allowlist
// decision, an instruction or any configuration. Whatever a page says is
// stored as inert text for a human reviewer.
package parsing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mmcdole/gofeed"

	"github.com/1a-li-lu-le-lo/pdoom/internal/dedup"
	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/sources"
)

// Limits.
const (
	MaxTextChars  = 20000
	MaxTitleChars = 500
	MaxItems      = 500
	MaxAuthors    = 50
	maxAuthorLen  = 200
)

// Document is the normalized unit handed to dedup, claims and the review queue.
type Document struct {
	SourceID     string   `json:"source_id"`
	URL          string   `json:"url"`
	CanonicalURL string   `json:"canonical_url"`
	Title        string   `json:"title"`
	PublishedAt  *string  `json:"published_at"`
	RetrievedAt  string   `json:"retrieved_at"`
	ContentHash  string   `json:"content_hash"`
	Text         string   `json:"text"`
	Language     string   `json:"language"`
	Authors      []string `json:"authors"`
}

// NewDocument assembles a Document, normalizing the URL, bounding title and
// text, hashing the content and guessing the language.
func NewDocument(sourceID, rawURL, title, text string, published *time.Time, retrievedAt string, authors []string) Document {
	title = Truncate(collapse(title), MaxTitleChars)
	text = Truncate(collapse(text), MaxTextChars)
	var pub *string
	if published != nil && !published.IsZero() {
		s := published.UTC().Format(time.RFC3339)
		pub = &s
	}
	clean := make([]string, 0, len(authors))
	seen := map[string]bool{}
	for _, a := range authors {
		a = Truncate(collapse(a), maxAuthorLen)
		if a == "" || seen[a] {
			continue
		}
		seen[a] = true
		clean = append(clean, a)
		if len(clean) >= MaxAuthors {
			break
		}
	}
	d := Document{
		SourceID:     sourceID,
		URL:          strings.TrimSpace(rawURL),
		CanonicalURL: dedup.CanonicalOrRaw(rawURL),
		Title:        title,
		PublishedAt:  pub,
		RetrievedAt:  retrievedAt,
		Text:         text,
		Language:     GuessLanguage(title + " " + text),
		Authors:      clean,
	}
	d.ContentHash = schema.SHA256Hex([]byte(d.Title + "\n" + d.Text))
	return d
}

// Parse dispatches on the configured parser.
func Parse(parser sources.Parser, sourceID, pageURL string, body []byte, retrievedAt string) ([]Document, error) {
	switch parser {
	case sources.ParserFeed:
		return ParseFeed(sourceID, body, retrievedAt)
	case sources.ParserHTMLText:
		d, err := ParseHTMLText(sourceID, pageURL, body, retrievedAt)
		if err != nil {
			return nil, err
		}
		return []Document{d}, nil
	case sources.ParserJSON:
		return ParseGitHubAdvisories(sourceID, body, retrievedAt)
	}
	return nil, fmt.Errorf("parsing: unknown parser %q", parser)
}

// ParseFeed parses an RSS, Atom or JSON feed. Items without a usable absolute
// link are skipped; at most MaxItems are returned, in feed order.
func ParseFeed(sourceID string, body []byte, retrievedAt string) ([]Document, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, fmt.Errorf("parsing: empty feed body")
	}
	feed, err := gofeed.NewParser().Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parsing: feed: %w", err)
	}
	if feed == nil {
		return nil, fmt.Errorf("parsing: feed: no feed decoded")
	}
	if strings.TrimSpace(feed.Title) == "" && len(feed.Items) == 0 {
		// gofeed accepts any JSON object as a JSON Feed; an untitled feed with
		// no items is not a feed we can use.
		return nil, fmt.Errorf("parsing: feed: no title and no items (not a feed?)")
	}
	var docs []Document
	for _, it := range feed.Items {
		if it == nil {
			continue
		}
		link := itemLink(it)
		if link == "" {
			continue
		}
		text := it.Content
		if strings.TrimSpace(StripHTML(text)) == "" {
			text = it.Description
		}
		var published *time.Time
		switch {
		case it.PublishedParsed != nil:
			published = it.PublishedParsed
		case it.UpdatedParsed != nil:
			published = it.UpdatedParsed
		}
		var authors []string
		for _, p := range it.Authors {
			if p != nil {
				authors = append(authors, p.Name)
			}
		}
		if len(authors) == 0 && it.Author != nil {
			authors = append(authors, it.Author.Name)
		}
		docs = append(docs, NewDocument(sourceID, link, StripHTML(it.Title), StripHTML(text), published, retrievedAt, authors))
		if len(docs) >= MaxItems {
			break
		}
	}
	return docs, nil
}

func itemLink(it *gofeed.Item) string {
	candidates := append([]string{it.Link}, it.Links...)
	candidates = append(candidates, it.GUID)
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		u, err := url.Parse(c)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			continue
		}
		return c
	}
	return ""
}

// ParseHTMLText extracts the visible text of an HTML page as one Document.
func ParseHTMLText(sourceID, pageURL string, body []byte, retrievedAt string) (Document, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return Document{}, fmt.Errorf("parsing: empty html body")
	}
	if !utf8.Valid(body) {
		body = bytes.ToValidUTF8(body, []byte("�"))
	}
	page := string(body)
	title := firstTag(page, "title")
	if title == "" {
		title = firstTag(page, "h1")
	}
	if title == "" {
		title = pageURL
	}
	text := StripHTML(page)
	if strings.TrimSpace(text) == "" {
		return Document{}, fmt.Errorf("parsing: html page has no visible text")
	}
	return NewDocument(sourceID, pageURL, title, text, nil, retrievedAt, nil), nil
}

// firstTag returns the stripped text of the first <tag>...</tag> element.
// Tag names are matched ASCII-case-insensitively in place: no case-folded copy
// is made, so every offset refers to page itself.
func firstTag(page, tag string) string {
	start := openTagIndex(page, tag)
	if start < 0 {
		return ""
	}
	open := strings.IndexByte(page[start:], '>')
	if open < 0 {
		return ""
	}
	rest := start + open + 1
	end := closeTagIndex(page[rest:], tag)
	if end < 0 {
		return ""
	}
	return StripHTML(page[rest : rest+end])
}

// ghAdvisory is the subset of a GitHub advisory that is retained. Description,
// affected packages, CVSS vectors and references are deliberately not decoded:
// the observatory records that an advisory exists and how severe it is rated,
// never operational detail (build-spec rule 0.9).
type ghAdvisory struct {
	GHSAID      string  `json:"ghsa_id"`
	CVEID       *string `json:"cve_id"`
	Summary     string  `json:"summary"`
	Severity    string  `json:"severity"`
	Type        string  `json:"type"`
	HTMLURL     string  `json:"html_url"`
	PublishedAt *string `json:"published_at"`
	UpdatedAt   *string `json:"updated_at"`
}

// ParseGitHubAdvisories parses the JSON array returned by the GitHub
// advisories REST endpoint. Unknown fields are ignored (the payload is
// external data, not our schema); an error object is reported as an error.
func ParseGitHubAdvisories(sourceID string, body []byte, retrievedAt string) ([]Document, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("parsing: empty json body")
	}
	if trimmed[0] == '{' {
		var obj struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(trimmed, &obj)
		if obj.Message != "" {
			return nil, fmt.Errorf("parsing: advisories endpoint returned an error object: %s", Truncate(obj.Message, 200))
		}
		return nil, fmt.Errorf("parsing: advisories payload is an object, expected an array")
	}
	var items []ghAdvisory
	if err := json.Unmarshal(trimmed, &items); err != nil {
		return nil, fmt.Errorf("parsing: advisories json: %w", err)
	}
	var docs []Document
	for _, a := range items {
		id := strings.TrimSpace(a.GHSAID)
		if id == "" {
			continue
		}
		link := strings.TrimSpace(a.HTMLURL)
		if link == "" {
			link = "https://github.com/advisories/" + url.PathEscape(id)
		}
		var published *time.Time
		for _, ts := range []*string{a.PublishedAt, a.UpdatedAt} {
			if ts == nil {
				continue
			}
			if t, err := time.Parse(time.RFC3339, *ts); err == nil {
				published = &t
				break
			}
		}
		title := id
		if s := collapse(a.Summary); s != "" {
			title = id + ": " + s
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Security advisory %s", id)
		if a.CVEID != nil && strings.TrimSpace(*a.CVEID) != "" {
			fmt.Fprintf(&b, " (%s)", strings.TrimSpace(*a.CVEID))
		}
		if a.Type != "" {
			fmt.Fprintf(&b, ", type %s", collapse(a.Type))
		}
		if a.Severity != "" {
			fmt.Fprintf(&b, ", severity %s", collapse(a.Severity))
		}
		b.WriteString(". Category-level record only; see the advisory page for details.")
		docs = append(docs, NewDocument(sourceID, link, title, b.String(), published, retrievedAt, nil))
		if len(docs) >= MaxItems {
			break
		}
	}
	return docs, nil
}

// dropContent lists elements whose content is never visible text.
var dropContent = map[string]bool{"script": true, "style": true, "noscript": true, "template": true, "svg": true}

// StripHTML removes tags, drops script/style/noscript/template/svg content,
// comments and CDATA markers, decodes entities and collapses whitespace. It is
// a tokenizer, not a parser: malformed markup degrades to text, never to an
// error or a panic.
//
// It is a single pass over the input: every search either advances past what
// it found or ends the region being tokenized, so time and memory are linear
// in len(s) whatever the markup looks like (a body made of nothing but
// <script></script> pairs or <![CDATA[ markers costs the same as plain text).
// Tag names are compared ASCII-case-insensitively in place; offsets are never
// taken from a case-folded copy, because strings.ToLower changes the byte
// length of some runes.
func StripHTML(s string) string {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "\uFFFD")
	}
	var out strings.Builder
	out.Grow(len(s))
	i := 0
	// limit is the end of the region being tokenized: len(s), or the position
	// of the "]]>" that closes the CDATA section currently being read. CDATA
	// content is often escaped markup (RSS descriptions), so it is tokenized
	// like the rest of the document instead of being emitted raw.
	limit := len(s)
	inCDATA := false
	// noCDATAEnd records that a search for "]]>" already failed; a later search
	// from a later position cannot succeed, so it is not repeated.
	noCDATAEnd := false
	for i < len(s) {
		if i >= limit {
			// End of the CDATA section: skip "]]>" and resume on the full input.
			i = limit + 3
			limit = len(s)
			inCDATA = false
			out.WriteByte(' ')
			continue
		}
		if s[i] != '<' {
			j := strings.IndexByte(s[i:limit], '<')
			if j < 0 {
				out.WriteString(s[i:limit])
				i = limit
				continue
			}
			out.WriteString(s[i : i+j])
			i += j
			continue
		}
		rest := s[i:limit]
		// Comment.
		if strings.HasPrefix(rest, "<!--") {
			end := strings.Index(rest[4:], "-->")
			if end < 0 {
				// Unterminated comment: the rest of the region is dropped.
				i = limit
				continue
			}
			i += 4 + end + 3
			out.WriteByte(' ')
			continue
		}
		// CDATA marker.
		if strings.HasPrefix(rest, "<![CDATA[") {
			i += 9
			if inCDATA || noCDATAEnd {
				// A nested marker, or a marker that can have no terminator: the
				// marker itself is dropped and tokenizing simply continues.
				continue
			}
			end := strings.Index(s[i:], "]]>")
			if end < 0 {
				// Unterminated section: the rest of the input is its content.
				noCDATAEnd = true
				continue
			}
			limit = i + end
			inCDATA = true
			continue
		}
		gt := strings.IndexByte(rest, '>')
		if gt < 0 {
			// Unterminated tag: treat the rest of the region as text.
			out.WriteString(rest)
			i = limit
			continue
		}
		name, closing := tagName(rest[1:gt])
		i += gt + 1
		out.WriteByte(' ')
		if !closing && dropContent[name] {
			// Skip to the matching close tag (ASCII case-insensitive, in place).
			end := closeTagIndex(s[i:limit], name)
			if end < 0 {
				// Never closed: the rest of the region is its content.
				i = limit
				continue
			}
			i += end
			gt2 := strings.IndexByte(s[i:limit], '>')
			if gt2 < 0 {
				i = limit
				continue
			}
			i += gt2 + 1
		}
	}
	return collapse(html.UnescapeString(out.String()))
}

// tagName returns the ASCII-lowercased element name of a tag body (the bytes
// between '<' and '>') and whether it is a closing tag.
func tagName(tag string) (name string, closing bool) {
	tag = strings.TrimSpace(tag)
	if strings.HasPrefix(tag, "/") {
		closing = true
		tag = strings.TrimSpace(tag[1:])
	}
	end := 0
	for end < len(tag) && !isTagNameEnd(tag[end]) {
		end++
	}
	return asciiLower(tag[:end]), closing
}

// isTagNameEnd reports whether c terminates an element name.
func isTagNameEnd(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '/' || c == '>'
}

// asciiLower lowercases ASCII letters only. It never changes the byte length
// of its input, which is what makes offsets into the result valid for the
// original; non-ASCII bytes are copied unchanged.
func asciiLower(s string) string {
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= 'A' && c <= 'Z' {
			b := []byte(s)
			for j := i; j < len(b); j++ {
				if b[j] >= 'A' && b[j] <= 'Z' {
					b[j] += 'a' - 'A'
				}
			}
			return string(b)
		}
	}
	return s
}

// hasFoldPrefix reports whether s starts with name, comparing ASCII letters
// case-insensitively. name must be ASCII lowercase.
func hasFoldPrefix(s, name string) bool {
	if len(s) < len(name) {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != name[i] {
			return false
		}
	}
	return true
}

// openTagIndex returns the offset of the first "<name" element start in s
// (ASCII case-insensitive; the name must be followed by whitespace, '/', '>'
// or the end of s), or -1. Each iteration advances past the '<' it examined,
// so the scan is linear.
func openTagIndex(s, name string) int {
	pos := 0
	for pos < len(s) {
		j := strings.IndexByte(s[pos:], '<')
		if j < 0 {
			return -1
		}
		start := pos + j
		after := start + 1
		if hasFoldPrefix(s[after:], name) {
			k := after + len(name)
			if k == len(s) || isTagNameEnd(s[k]) {
				return start
			}
		}
		pos = after
	}
	return -1
}

// closeTagIndex returns the offset of the first "</name" in s (ASCII
// case-insensitive; the name must be followed by whitespace, '/', '>' or the
// end of s), or -1. Linear for the same reason as openTagIndex.
func closeTagIndex(s, name string) int {
	pos := 0
	for pos < len(s) {
		j := strings.Index(s[pos:], "</")
		if j < 0 {
			return -1
		}
		start := pos + j
		after := start + 2
		if hasFoldPrefix(s[after:], name) {
			k := after + len(name)
			if k == len(s) || isTagNameEnd(s[k]) {
				return start
			}
		}
		pos = after
	}
	return -1
}

// collapse trims and collapses runs of whitespace to one space.
func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// Truncate cuts s to at most max runes.
func Truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	n := 0
	for i := range s {
		if n == max {
			return s[:i]
		}
		n++
	}
	return s
}

// englishStopwords drive the language guess.
var englishStopwords = map[string]bool{
	"the": true, "and": true, "of": true, "to": true, "in": true, "is": true, "that": true, "for": true, "with": true,
	"as": true, "on": true, "are": true, "this": true, "by": true, "be": true, "was": true, "it": true, "an": true,
	"at": true, "from": true, "or": true, "which": true, "we": true, "has": true, "have": true, "not": true, "but": true,
}

// GuessLanguage returns "en" when the text carries enough English function
// words, otherwise "unknown". It is a deliberately small heuristic.
func GuessLanguage(text string) string {
	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return "unknown"
	}
	hits := 0
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?\"'()[]{}“”‘’")
		if englishStopwords[w] {
			hits++
		}
	}
	ratio := float64(hits) / float64(len(words))
	switch {
	case len(words) >= 20 && ratio >= 0.06:
		return "en"
	case len(words) < 20 && hits >= 2 && ratio >= 0.1:
		return "en"
	}
	return "unknown"
}
