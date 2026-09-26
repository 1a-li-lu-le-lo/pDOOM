// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package sources loads and validates config/sources.json, the allowlist of
// ingestion sources. The file is the only place a host can be granted network
// access by pdoom-ingest; the pipeline itself never writes it.
package sources

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Version is the only supported config version.
const Version = 1

// Kind is the transport kind of a source.
type Kind string

// Kinds.
const (
	KindRSS     Kind = "rss"
	KindAtom    Kind = "atom"
	KindAPI     Kind = "api"
	KindDataset Kind = "dataset"
	KindHTML    Kind = "html"
)

// Kinds lists the allowed values.
var Kinds = []Kind{KindRSS, KindAtom, KindAPI, KindDataset, KindHTML}

// Valid reports whether the kind is allowed.
func (k Kind) Valid() bool {
	for _, x := range Kinds {
		if x == k {
			return true
		}
	}
	return false
}

// Parser names the parser applied to a fetched body.
type Parser string

// Parsers.
const (
	ParserFeed     Parser = "feed"
	ParserJSON     Parser = "json"
	ParserHTMLText Parser = "html_text"
)

// Parsers lists the allowed values.
var Parsers = []Parser{ParserFeed, ParserJSON, ParserHTMLText}

// Valid reports whether the parser is allowed.
func (p Parser) Valid() bool {
	for _, x := range Parsers {
		if x == p {
			return true
		}
	}
	return false
}

// MaxItemsCap bounds max_items_per_run.
const MaxItemsCap = 500

// Source is one allowlist entry.
type Source struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Kind               Kind                   `json:"kind"`
	URL                string                 `json:"url"`
	Tier               schema.SourceTier      `json:"tier"`
	Publisher          string                 `json:"publisher"`
	Allowed            bool                   `json:"allowed"`
	Reason             string                 `json:"reason"`
	RobotsCheckedAt    *string                `json:"robots_checked_at"`
	RobotsStatus       schema.RobotsStatus    `json:"robots_status"`
	License            *string                `json:"license"`
	MaxItemsPerRun     int                    `json:"max_items_per_run"`
	FetchIntervalHours int                    `json:"fetch_interval_hours"`
	Parser             Parser                 `json:"parser"`
	Conflicts          []schema.ConflictLabel `json:"conflicts"`
	Notes              string                 `json:"notes"`
}

// Config is the decoded config/sources.json.
type Config struct {
	Version     int      `json:"version"`
	GeneratedAt string   `json:"generated_at"`
	Sources     []Source `json:"sources"`
}

var (
	reID   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,63}$`)
	reDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// Load reads and validates a config file.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sources: read %s: %w", path, err)
	}
	c, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("sources: %s: %w", path, err)
	}
	return c, nil
}

// Parse decodes strictly (unknown fields and trailing data are errors) and validates.
func Parse(raw []byte) (*Config, error) {
	var c Config
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("decode: trailing data after JSON document")
	}
	if err := Validate(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Validate checks the whole config and returns every problem joined.
func Validate(c *Config) error {
	if c == nil {
		return fmt.Errorf("sources: nil config")
	}
	var errs []error
	if c.Version != Version {
		errs = append(errs, fmt.Errorf("version must be %d, got %d", Version, c.Version))
	}
	if !reDate.MatchString(c.GeneratedAt) {
		errs = append(errs, fmt.Errorf("generated_at must be YYYY-MM-DD, got %q", c.GeneratedAt))
	}
	if len(c.Sources) == 0 {
		errs = append(errs, fmt.Errorf("sources must not be empty"))
	}
	seen := map[string]bool{}
	for i, s := range c.Sources {
		label := fmt.Sprintf("sources[%d]", i)
		if s.ID != "" {
			label += " (" + s.ID + ")"
		}
		for _, err := range validateSource(s) {
			errs = append(errs, fmt.Errorf("%s: %w", label, err))
		}
		if seen[s.ID] {
			errs = append(errs, fmt.Errorf("%s: duplicate id", label))
		}
		seen[s.ID] = true
	}
	if len(errs) > 0 {
		return fmt.Errorf("sources: invalid config: %w", errors.Join(errs...))
	}
	return nil
}

func validateSource(s Source) []error {
	var errs []error
	if !reID.MatchString(s.ID) {
		errs = append(errs, fmt.Errorf("id %q must match %s", s.ID, reID))
	}
	if strings.TrimSpace(s.Name) == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	if strings.TrimSpace(s.Publisher) == "" {
		errs = append(errs, fmt.Errorf("publisher is required"))
	}
	if !s.Kind.Valid() {
		errs = append(errs, fmt.Errorf("kind %q invalid", s.Kind))
	}
	if !s.Parser.Valid() {
		errs = append(errs, fmt.Errorf("parser %q invalid", s.Parser))
	} else {
		switch s.Kind {
		case KindRSS, KindAtom:
			if s.Parser != ParserFeed {
				errs = append(errs, fmt.Errorf("kind %s requires parser feed", s.Kind))
			}
		case KindHTML:
			if s.Parser != ParserHTMLText {
				errs = append(errs, fmt.Errorf("kind html requires parser html_text"))
			}
		case KindAPI:
			if s.Parser == ParserHTMLText {
				errs = append(errs, fmt.Errorf("kind api requires parser feed or json"))
			}
		}
	}
	if err := checkURL(s.URL); err != nil {
		errs = append(errs, err)
	}
	if !s.Tier.Valid() {
		errs = append(errs, fmt.Errorf("tier %d must be 1..5", s.Tier))
	}
	if !s.RobotsStatus.Valid() {
		errs = append(errs, fmt.Errorf("robots_status %q invalid", s.RobotsStatus))
	}
	if s.RobotsCheckedAt != nil && !reDate.MatchString(*s.RobotsCheckedAt) {
		errs = append(errs, fmt.Errorf("robots_checked_at must be YYYY-MM-DD or null"))
	}
	if s.RobotsStatus != "unknown" && s.RobotsStatus.Valid() && s.RobotsCheckedAt == nil {
		errs = append(errs, fmt.Errorf("robots_status %q requires robots_checked_at", s.RobotsStatus))
	}
	if s.MaxItemsPerRun < 0 || s.MaxItemsPerRun > MaxItemsCap {
		errs = append(errs, fmt.Errorf("max_items_per_run %d must be 0..%d", s.MaxItemsPerRun, MaxItemsCap))
	}
	if s.FetchIntervalHours < 0 {
		errs = append(errs, fmt.Errorf("fetch_interval_hours must be >= 0"))
	}
	if s.Conflicts == nil {
		errs = append(errs, fmt.Errorf("conflicts must be an array (use [] or [\"none_known\"])"))
	}
	for _, cl := range s.Conflicts {
		if !cl.Valid() {
			errs = append(errs, fmt.Errorf("conflict label %q invalid", cl))
		}
	}
	if s.Allowed {
		if s.RobotsStatus != "allowed" && s.RobotsStatus != "not_applicable" {
			errs = append(errs, fmt.Errorf("allowed source requires robots_status allowed or not_applicable, got %q", s.RobotsStatus))
		}
		if s.License == nil || strings.TrimSpace(*s.License) == "" {
			errs = append(errs, fmt.Errorf("allowed source requires a license"))
		}
		if s.MaxItemsPerRun < 1 {
			errs = append(errs, fmt.Errorf("allowed source requires max_items_per_run >= 1"))
		}
		if s.FetchIntervalHours < 1 {
			errs = append(errs, fmt.Errorf("allowed source requires fetch_interval_hours >= 1"))
		}
	} else if strings.TrimSpace(s.Reason) == "" {
		errs = append(errs, fmt.Errorf("disallowed source requires a reason"))
	}
	return errs
}

func checkURL(raw string) error {
	if strings.TrimSpace(raw) == "" || raw != strings.TrimSpace(raw) {
		return fmt.Errorf("url is required and must not have surrounding whitespace")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("url %q: %w", raw, err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("url %q must use https", raw)
	}
	if u.User != nil {
		return fmt.Errorf("url %q must not carry credentials", raw)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("url %q has no host", raw)
	}
	if net.ParseIP(host) != nil {
		return fmt.Errorf("url %q must use a hostname, not an IP literal", raw)
	}
	if !strings.Contains(host, ".") {
		return fmt.Errorf("url %q host must be fully qualified", raw)
	}
	if p := u.Port(); p != "" && p != "443" {
		return fmt.Errorf("url %q must use port 443", raw)
	}
	return nil
}

// Host returns the lowercase hostname of the source URL ("" when unparsable).
func (s Source) Host() string {
	u, err := url.Parse(s.URL)
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
}

// Allowed returns the sources with allowed: true, in file order.
func (c *Config) Allowed() []Source {
	var out []Source
	for _, s := range c.Sources {
		if s.Allowed {
			out = append(out, s)
		}
	}
	return out
}

// ByID looks a source up.
func (c *Config) ByID(id string) (Source, bool) {
	for _, s := range c.Sources {
		if s.ID == id {
			return s, true
		}
	}
	return Source{}, false
}

// AllowedHosts returns the sorted, de-duplicated hostnames of allowed sources.
// This is the SafeClient host allowlist.
func (c *Config) AllowedHosts() []string {
	return hostsOf(c.Allowed())
}

// AllHosts returns the sorted, de-duplicated hostnames of every source
// (used by the robots.txt check mode, which reads robots files only).
func (c *Config) AllHosts() []string {
	return hostsOf(c.Sources)
}

func hostsOf(list []Source) []string {
	set := map[string]bool{}
	for _, s := range list {
		if h := s.Host(); h != "" {
			set[h] = true
		}
	}
	out := make([]string, 0, len(set))
	for h := range set {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}
