// Copyright NU Cybernetics. p(DOOM) — research prototype.

package sources

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

type conflictAlias = schema.ConflictLabel

func strp(s string) *string { return &s }

func validConfig() Config {
	return Config{
		Version:     1,
		GeneratedAt: "2026-09-26",
		Sources: []Source{
			{
				ID: "example-feed", Name: "Example feed", Kind: KindRSS, URL: "https://feeds.example.com/rss.xml", Tier: 2,
				Publisher: "Example", Allowed: true, Reason: "", RobotsCheckedAt: strp("2026-09-26"), RobotsStatus: "allowed",
				License: strp("CC BY 4.0"), MaxItemsPerRun: 20, FetchIntervalHours: 24, Parser: ParserFeed, Conflicts: string2Conflict("none_known"), Notes: "",
			},
			{
				ID: "manual-dataset", Name: "Manual dataset", Kind: KindDataset, URL: "https://data.example.org/snapshot", Tier: 1,
				Publisher: "Example Org", Allowed: false, Reason: "manual import only", RobotsCheckedAt: nil, RobotsStatus: "unknown",
				License: strp("CC BY-SA 4.0"), MaxItemsPerRun: 0, FetchIntervalHours: 0, Parser: ParserJSON, Conflicts: string2Conflict(), Notes: "",
			},
		},
	}
}

// string2Conflict builds a conflicts slice from strings (nil-safe: always non-nil).
func string2Conflict(vals ...string) []conflictAlias {
	out := make([]conflictAlias, 0, len(vals))
	for _, v := range vals {
		out = append(out, conflictAlias(v))
	}
	return out
}

func TestParseValid(t *testing.T) {
	raw, _ := json.Marshal(validConfig())
	c, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(c.Allowed()) != 1 || c.Allowed()[0].ID != "example-feed" {
		t.Fatalf("Allowed = %+v", c.Allowed())
	}
	if hosts := c.AllowedHosts(); len(hosts) != 1 || hosts[0] != "feeds.example.com" {
		t.Fatalf("AllowedHosts = %v", hosts)
	}
	if hosts := c.AllHosts(); len(hosts) != 2 || hosts[0] != "data.example.org" {
		t.Fatalf("AllHosts = %v", hosts)
	}
	if _, ok := c.ByID("manual-dataset"); !ok {
		t.Fatal("ByID")
	}
	if _, ok := c.ByID("nope"); ok {
		t.Fatal("ByID unknown")
	}
}

func TestLoadFromDisk(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sources.json")
	raw, _ := json.Marshal(validConfig())
	if err := os.WriteFile(p, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := Load(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("missing file must fail")
	}
}

func TestStrictDecoding(t *testing.T) {
	raw, _ := json.Marshal(validConfig())
	withUnknown := strings.Replace(string(raw), `"version":1`, `"version":1,"extra":true`, 1)
	if _, err := Parse([]byte(withUnknown)); err == nil || !strings.Contains(err.Error(), "extra") {
		t.Fatalf("unknown field must be rejected: %v", err)
	}
	if _, err := Parse(append(raw, []byte(" {}")...)); err == nil || !strings.Contains(err.Error(), "trailing") {
		t.Fatalf("trailing data must be rejected: %v", err)
	}
	if _, err := Parse([]byte("not json")); err == nil {
		t.Fatal("garbage must be rejected")
	}
}

func TestValidationRules(t *testing.T) {
	mutate := func(name string, f func(c *Config), wantSubstr string) {
		t.Run(name, func(t *testing.T) {
			c := validConfig()
			f(&c)
			err := Validate(&c)
			if err == nil {
				t.Fatalf("expected error containing %q", wantSubstr)
			}
			if !strings.Contains(err.Error(), wantSubstr) {
				t.Fatalf("error %q does not contain %q", err.Error(), wantSubstr)
			}
		})
	}
	mutate("version", func(c *Config) { c.Version = 2 }, "version must be 1")
	mutate("generated_at", func(c *Config) { c.GeneratedAt = "yesterday" }, "generated_at")
	mutate("empty sources", func(c *Config) { c.Sources = nil }, "must not be empty")
	mutate("duplicate ids", func(c *Config) { c.Sources[1].ID = "example-feed" }, "duplicate id")
	mutate("bad id", func(c *Config) { c.Sources[0].ID = "Bad_ID" }, "must match")
	mutate("http url", func(c *Config) { c.Sources[0].URL = "http://feeds.example.com/rss.xml" }, "must use https")
	mutate("ip url", func(c *Config) { c.Sources[0].URL = "https://10.0.0.1/rss.xml" }, "IP literal")
	mutate("creds url", func(c *Config) { c.Sources[0].URL = "https://u:p@feeds.example.com/rss.xml" }, "credentials")
	mutate("unqualified host", func(c *Config) { c.Sources[0].URL = "https://localhost/rss.xml" }, "fully qualified")
	mutate("port", func(c *Config) { c.Sources[0].URL = "https://feeds.example.com:8443/rss.xml" }, "port 443")
	mutate("tier", func(c *Config) { c.Sources[0].Tier = 6 }, "tier 6")
	mutate("kind", func(c *Config) { c.Sources[0].Kind = "telnet" }, "kind")
	mutate("parser", func(c *Config) { c.Sources[0].Parser = "regex" }, "parser")
	mutate("rss needs feed parser", func(c *Config) { c.Sources[0].Parser = ParserJSON }, "requires parser feed")
	mutate("html needs html_text", func(c *Config) { c.Sources[0].Kind = KindHTML }, "requires parser html_text")
	mutate("api not html_text", func(c *Config) { c.Sources[0].Kind = KindAPI; c.Sources[0].Parser = ParserHTMLText }, "feed or json")
	mutate("robots status", func(c *Config) { c.Sources[0].RobotsStatus = "maybe" }, "robots_status")
	mutate("robots checked date", func(c *Config) { c.Sources[0].RobotsCheckedAt = strp("today") }, "robots_checked_at")
	mutate("robots status needs date", func(c *Config) { c.Sources[0].RobotsCheckedAt = nil }, "requires robots_checked_at")
	mutate("allowed needs robots", func(c *Config) { c.Sources[0].RobotsStatus = "disallowed" }, "allowed source requires robots_status")
	mutate("allowed unknown robots", func(c *Config) { c.Sources[0].RobotsStatus = "unknown"; c.Sources[0].RobotsCheckedAt = nil }, "allowed source requires robots_status")
	mutate("allowed needs license", func(c *Config) { c.Sources[0].License = nil }, "requires a license")
	mutate("allowed empty license", func(c *Config) { c.Sources[0].License = strp("  ") }, "requires a license")
	mutate("allowed needs max items", func(c *Config) { c.Sources[0].MaxItemsPerRun = 0 }, "max_items_per_run >= 1")
	mutate("max items cap", func(c *Config) { c.Sources[0].MaxItemsPerRun = MaxItemsCap + 1 }, "max_items_per_run")
	mutate("allowed needs interval", func(c *Config) { c.Sources[0].FetchIntervalHours = 0 }, "fetch_interval_hours >= 1")
	mutate("negative interval", func(c *Config) { c.Sources[1].FetchIntervalHours = -1 }, "fetch_interval_hours")
	mutate("disallowed needs reason", func(c *Config) { c.Sources[1].Reason = "" }, "requires a reason")
	mutate("conflicts nil", func(c *Config) { c.Sources[0].Conflicts = nil }, "conflicts must be an array")
	mutate("conflict label", func(c *Config) { c.Sources[0].Conflicts = string2Conflict("bribed") }, "conflict label")
	mutate("name", func(c *Config) { c.Sources[0].Name = " " }, "name is required")
	mutate("publisher", func(c *Config) { c.Sources[0].Publisher = "" }, "publisher is required")
	if err := Validate(nil); err == nil {
		t.Fatal("nil config must fail")
	}
}

func TestNotApplicableRobotsIsAllowedForAPIs(t *testing.T) {
	c := validConfig()
	c.Sources[0].Kind = KindAPI
	c.Sources[0].Parser = ParserJSON
	c.Sources[0].RobotsStatus = "not_applicable"
	if err := Validate(&c); err != nil {
		t.Fatalf("not_applicable must be accepted for allowed sources: %v", err)
	}
}

func TestHostHandlesBadURL(t *testing.T) {
	if (Source{URL: "http://[::1"}).Host() != "" {
		t.Fatal("bad url host must be empty")
	}
	if (Source{URL: "https://Feeds.Example.COM./x"}).Host() != "feeds.example.com" {
		t.Fatal("host must be lowercased without trailing dot")
	}
}

func TestRepositoryConfigIsValid(t *testing.T) {
	p := filepath.Join("..", "..", "config", "sources.json")
	c, err := Load(p)
	if err != nil {
		t.Fatalf("config/sources.json: %v", err)
	}
	if len(c.Sources) < 8 || len(c.Sources) > 12 {
		t.Fatalf("config/sources.json should list 8-12 sources, has %d", len(c.Sources))
	}
	for _, s := range c.Sources {
		if s.Allowed && (s.RobotsStatus != "allowed" && s.RobotsStatus != "not_applicable") {
			t.Fatalf("%s: allowed without a verified robots status", s.ID)
		}
		if !s.Allowed && strings.TrimSpace(s.Reason) == "" {
			t.Fatalf("%s: disallowed without a reason", s.ID)
		}
	}
}
