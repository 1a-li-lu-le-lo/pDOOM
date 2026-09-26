// Copyright NU Cybernetics. p(DOOM) — research prototype.

package api

import (
	"testing"
)

func TestStripControl(t *testing.T) {
	cases := map[string]string{
		"plain":                     "plain",
		"  padded  ":                "padded",
		"a\x00b\x07c\x1bd":          "abcd",
		"line\nbreak\tkept":         "line\nbreak\tkept",
		"cr\r\nlf":                  "cr\nlf",
		"del\x7fchar":               "delchar",
		"c1\u0085\u009fcontrols":    "c1controls",
		"sep ara tors":              "separators",
		"unicode é ✓ stays":         "unicode é ✓ stays",
		"\x00\x01\x02":              "",
		"zero​width stays (format)": "zero​width stays (format)",
	}
	for in, want := range cases {
		if got := stripControl(in); got != want {
			t.Errorf("stripControl(%q) = %q, want %q", in, got, want)
		}
	}
	if cleanPtr(nil) != nil {
		t.Fatal("nil stays nil")
	}
	empty := "\x00 "
	if cleanPtr(&empty) != nil {
		t.Fatal("a string that empties after stripping becomes nil")
	}
	v := " x\x01y "
	if got := cleanPtr(&v); got == nil || *got != "xy" {
		t.Fatalf("cleanPtr: %v", got)
	}
}

func TestValidateSourcePayloadReportsEveryProblem(t *testing.T) {
	bad := "May"
	p := SourceSubmissionPayload{CanonicalURL: "ftp://x", DatePublished: &bad}
	problems := validateSourcePayload(&p)
	if len(problems) != 3 {
		t.Fatalf("expected 3 problems (url, title, date), got %v", problems)
	}
	good := SourceSubmissionPayload{CanonicalURL: "https://example.org/a?b=c", Title: "T", WhyRelevant: "w", ClaimedEvidence: "e"}
	if problems := validateSourcePayload(&good); len(problems) != 0 {
		t.Fatalf("unexpected problems %v", problems)
	}
	field := "Bad"
	url := "http://"
	c := CorrectionSubmissionPayload{TargetID: "a b", Field: &field, EvidenceURL: &url}
	if problems := validateCorrectionPayload(&c); len(problems) != 4 {
		t.Fatalf("expected 4 problems (target, field, correction, url), got %v", problems)
	}
}

func TestEtagMatches(t *testing.T) {
	etag := `"abc"`
	for _, ok := range []string{`"abc"`, `W/"abc"`, `"x", "abc"`, "*"} {
		if !etagMatches(ok, etag) {
			t.Errorf("%q should match", ok)
		}
	}
	for _, no := range []string{"", `"abd"`, `abc`} {
		if etagMatches(no, etag) {
			t.Errorf("%q should not match", no)
		}
	}
}

func TestNormalizeOrigin(t *testing.T) {
	if normalizeOrigin("HTTPS://Lab.Example/path") != "https://lab.example" || normalizeOrigin("nope") != "" || normalizeOrigin("") != "" {
		t.Fatal("normalizeOrigin")
	}
}

func TestMarkdownTitle(t *testing.T) {
	if markdownTitle([]byte("<!-- c -->\n\n#  Title here \nbody"), "slug") != "Title here" {
		t.Fatal("heading not found")
	}
	if markdownTitle([]byte("## not level one\n"), "slug") != "slug" {
		t.Fatal("fallback expected")
	}
}
