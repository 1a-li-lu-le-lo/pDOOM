// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package dedup normalizes URLs and detects near-duplicate documents so that
// one underlying event (a press release and the articles that repeat it)
// counts once in the review queue.
//
// Everything here is deterministic and stdlib-only: canonical URL rules,
// a 64-bit simhash over word shingles and a small clusterer that assigns a
// stable cluster id to each document.
package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// DefaultHammingThreshold is the simhash distance at or below which two
// documents are considered near-duplicates. For unrelated texts the distance
// of two 64-bit simhashes is centred on 32 with a standard deviation of 4, so
// 8 keeps false positives negligible while tolerating a rewritten lead
// sentence or a few substituted words.
const DefaultHammingThreshold = 8

// minShingles is the number of shingles below which simhash is too noisy;
// shorter texts fall back to exact normalized-text comparison.
const minShingles = 4

var (
	reArxivPath = regexp.MustCompile(`^/(?:abs|pdf|html|format|e-print|ps)/([a-z\-]+(?:\.[A-Z]{2})?/\d{7}|\d{4}\.\d{4,5})(?:v\d+)?(?:\.pdf)?/?$`)
	reDOIPrefix = regexp.MustCompile(`^(?i)doi:\s*`)
)

// trackingParams are removed from query strings (exact, lowercase names).
var trackingParams = map[string]bool{
	"fbclid": true, "gclid": true, "dclid": true, "msclkid": true, "mc_cid": true, "mc_eid": true,
	"ref_src": true, "igshid": true, "_hsenc": true, "_hsmi": true, "yclid": true,
}

// CanonicalURL normalizes a URL for identity comparison:
//
//   - scheme and host lowercased, default ports and trailing host dots removed
//   - fragment removed; utm_* and known click-tracking parameters removed;
//     remaining query parameters sorted
//   - trailing slash removed from non-root paths
//   - arXiv abs/pdf/html/version variants unified to https://arxiv.org/abs/<id>
//   - DOI links (doi.org, dx.doi.org, "doi:" prefix) unified to
//     https://doi.org/<lowercase doi>
//
// An error is returned for empty or unparsable input and for URLs without a host.
func CanonicalURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("dedup: empty url")
	}
	if reDOIPrefix.MatchString(s) {
		doi := strings.TrimSpace(reDOIPrefix.ReplaceAllString(s, ""))
		if doi == "" {
			return "", fmt.Errorf("dedup: empty doi")
		}
		return "https://doi.org/" + strings.ToLower(doi), nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("dedup: parse %q: %w", raw, err)
	}
	if u.Host == "" || u.Scheme == "" {
		return "", fmt.Errorf("dedup: url %q has no scheme or host", raw)
	}
	u.Scheme = strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	host = strings.TrimSuffix(host, ".")
	port := u.Port()
	if (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
		port = ""
	}
	u.User = nil
	u.Fragment = ""
	u.RawFragment = ""

	// arXiv: any host in the arxiv.org family, abs/pdf/version variants.
	if host == "arxiv.org" || strings.HasSuffix(host, ".arxiv.org") {
		if m := reArxivPath.FindStringSubmatch(u.Path); m != nil {
			return "https://arxiv.org/abs/" + m[1], nil
		}
	}
	// DOI resolvers.
	if host == "doi.org" || host == "dx.doi.org" || host == "www.doi.org" {
		doi := strings.TrimPrefix(u.Path, "/")
		if doi == "" {
			return "", fmt.Errorf("dedup: doi url %q has no doi", raw)
		}
		return "https://doi.org/" + strings.ToLower(doi), nil
	}

	if port != "" {
		u.Host = host + ":" + port
	} else {
		u.Host = host
	}
	// Query: drop tracking parameters, sort the rest.
	if u.RawQuery != "" {
		q := u.Query()
		keys := make([]string, 0, len(q))
		for k := range q {
			lk := strings.ToLower(k)
			if strings.HasPrefix(lk, "utm_") || trackingParams[lk] {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			vals := q[k]
			sort.Strings(vals)
			for _, v := range vals {
				if b.Len() > 0 {
					b.WriteByte('&')
				}
				b.WriteString(url.QueryEscape(k))
				if v != "" {
					b.WriteByte('=')
					b.WriteString(url.QueryEscape(v))
				}
			}
		}
		u.RawQuery = b.String()
	}
	// Path: drop trailing slash except for the root.
	if len(u.Path) > 1 && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimRight(u.Path, "/")
		u.RawPath = ""
	}
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String(), nil
}

// CanonicalOrRaw returns the canonical form or, when the input cannot be
// normalized, the trimmed input unchanged. It never fails.
func CanonicalOrRaw(raw string) string {
	c, err := CanonicalURL(raw)
	if err != nil {
		return strings.TrimSpace(raw)
	}
	return c
}

// NormalizeText lowercases, strips punctuation and collapses whitespace so
// that trivially different renderings of the same text compare equal.
func NormalizeText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	space := true
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			space = false
		default:
			if !space {
				b.WriteByte(' ')
				space = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// Shingles returns the word n-grams of the normalized text, in order.
func Shingles(text string, n int) []string {
	if n <= 0 {
		n = 3
	}
	words := strings.Fields(NormalizeText(text))
	if len(words) == 0 {
		return nil
	}
	if len(words) < n {
		return []string{strings.Join(words, " ")}
	}
	out := make([]string, 0, len(words)-n+1)
	for i := 0; i+n <= len(words); i++ {
		out = append(out, strings.Join(words[i:i+n], " "))
	}
	return out
}

// Simhash computes a 64-bit simhash over the 3-word shingles of text. Equal
// texts give equal hashes; texts differing in a few words differ in a few bits.
func Simhash(text string) uint64 {
	return simhashShingles(Shingles(text, 3))
}

func simhashShingles(shingles []string) uint64 {
	var acc [64]int
	for _, sh := range shingles {
		h := fnv.New64a()
		_, _ = h.Write([]byte(sh))
		v := h.Sum64()
		for i := 0; i < 64; i++ {
			if v&(1<<uint(i)) != 0 {
				acc[i]++
			} else {
				acc[i]--
			}
		}
	}
	var out uint64
	for i := 0; i < 64; i++ {
		if acc[i] > 0 {
			out |= 1 << uint(i)
		}
	}
	return out
}

// Hamming returns the number of differing bits between two hashes.
func Hamming(a, b uint64) int {
	x := a ^ b
	n := 0
	for x != 0 {
		x &= x - 1
		n++
	}
	return n
}

// Match is the reason a document was attached to an existing cluster.
type Match string

// Match reasons.
const (
	MatchNone     Match = ""
	MatchURL      Match = "canonical_url"
	MatchExact    Match = "exact_text"
	MatchSimhash  Match = "simhash"
	MatchContent  Match = "content_hash"
	clusterPrefix       = "dup-"
)

type entry struct {
	docID     string
	clusterID string
	hash      uint64
	exact     string
	shingles  int
}

// Clusterer assigns cluster ids. It is not safe for concurrent use.
type Clusterer struct {
	threshold int
	byURL     map[string]string // canonical url → cluster id
	byHash    map[string]string // content hash → cluster id
	entries   []entry
	sizes     map[string]int
}

// NewClusterer returns a clusterer with the given Hamming threshold (≤ 0 uses
// DefaultHammingThreshold).
func NewClusterer(threshold int) *Clusterer {
	if threshold <= 0 {
		threshold = DefaultHammingThreshold
	}
	return &Clusterer{threshold: threshold, byURL: map[string]string{}, byHash: map[string]string{}, sizes: map[string]int{}}
}

// Assignment is the result of adding one document to the clusterer.
type Assignment struct {
	ClusterID string `json:"cluster_id"`
	Duplicate bool   `json:"duplicate"`
	Match     Match  `json:"match,omitempty"`
	// Distance is the simhash Hamming distance for simhash matches.
	Distance int `json:"distance,omitempty"`
}

// ClusterIDFor derives the stable cluster id of a document from its canonical
// URL (or, failing that, its normalized text).
func ClusterIDFor(canonicalURL, title, text string) string {
	key := strings.TrimSpace(canonicalURL)
	if key == "" {
		key = NormalizeText(title + " " + text)
	}
	sum := sha256.Sum256([]byte(key))
	return clusterPrefix + hex.EncodeToString(sum[:])[:16]
}

// Assign adds a document and returns its cluster. contentHash may be empty.
// The first document of an event starts a cluster named after itself; later
// documents whose canonical URL, content hash, normalized text or simhash match
// join that cluster and are reported as duplicates.
func (c *Clusterer) Assign(docID, canonicalURL, contentHash, title, text string) Assignment {
	canon := strings.TrimSpace(canonicalURL)
	if canon != "" {
		if id, ok := c.byURL[canon]; ok {
			c.sizes[id]++
			return Assignment{ClusterID: id, Duplicate: true, Match: MatchURL}
		}
	}
	if contentHash != "" {
		if id, ok := c.byHash[contentHash]; ok {
			c.sizes[id]++
			if canon != "" {
				c.byURL[canon] = id
			}
			return Assignment{ClusterID: id, Duplicate: true, Match: MatchContent}
		}
	}
	// Headlines vary most between outlets, so the simhash is taken over the
	// body text when there is enough of it; short bodies fall back to title+body.
	shingles := Shingles(text, 3)
	if len(shingles) < minShingles {
		shingles = Shingles(title+" "+text, 3)
	}
	exact := NormalizeText(title + " " + text)
	h := simhashShingles(shingles)
	best := -1
	var bestID string
	var bestMatch Match
	for _, e := range c.entries {
		if exact != "" && e.exact == exact {
			best, bestID, bestMatch = 0, e.clusterID, MatchExact
			break
		}
		if len(shingles) < minShingles || e.shingles < minShingles {
			continue
		}
		d := Hamming(h, e.hash)
		if d <= c.threshold && (best < 0 || d < best) {
			best, bestID, bestMatch = d, e.clusterID, MatchSimhash
		}
	}
	if best >= 0 {
		c.sizes[bestID]++
		c.entries = append(c.entries, entry{docID: docID, clusterID: bestID, hash: h, exact: exact, shingles: len(shingles)})
		if canon != "" {
			c.byURL[canon] = bestID
		}
		if contentHash != "" {
			c.byHash[contentHash] = bestID
		}
		return Assignment{ClusterID: bestID, Duplicate: true, Match: bestMatch, Distance: best}
	}
	id := ClusterIDFor(canon, title, text)
	c.entries = append(c.entries, entry{docID: docID, clusterID: id, hash: h, exact: exact, shingles: len(shingles)})
	if canon != "" {
		c.byURL[canon] = id
	}
	if contentHash != "" {
		c.byHash[contentHash] = id
	}
	c.sizes[id]++
	return Assignment{ClusterID: id}
}

// Size returns the number of documents assigned to a cluster.
func (c *Clusterer) Size(clusterID string) int { return c.sizes[clusterID] }

// Clusters returns the number of distinct clusters.
func (c *Clusterer) Clusters() int { return len(c.sizes) }
