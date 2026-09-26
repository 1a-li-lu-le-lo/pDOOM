// Copyright NU Cybernetics. p(DOOM) — research prototype.

package dedup

import (
	"strings"
	"testing"
)

func TestCanonicalURL(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"HTTPS://Example.COM:443/Path/?utm_source=x&b=2&a=1#frag", "https://example.com/Path?a=1&b=2"},
		{"https://example.com/", "https://example.com/"},
		{"https://example.com", "https://example.com/"},
		{"https://example.com/news/?fbclid=abc", "https://example.com/news"},
		{"https://user:pw@example.com./x", "https://example.com/x"},
		{"http://arxiv.org/abs/2403.01234v2", "https://arxiv.org/abs/2403.01234"},
		{"https://arxiv.org/pdf/2403.01234v1.pdf", "https://arxiv.org/abs/2403.01234"},
		{"https://export.arxiv.org/abs/2403.01234", "https://arxiv.org/abs/2403.01234"},
		{"https://arxiv.org/html/2403.01234v3", "https://arxiv.org/abs/2403.01234"},
		{"https://arxiv.org/abs/hep-th/9901001v1", "https://arxiv.org/abs/hep-th/9901001"},
		{"https://arxiv.org/list/cs.AI/recent", "https://arxiv.org/list/cs.AI/recent"},
		{"https://dx.doi.org/10.1000/ABC.DEF", "https://doi.org/10.1000/abc.def"},
		{"doi:10.1000/XYZ", "https://doi.org/10.1000/xyz"},
		{"https://doi.org/10.1000/xyz?utm_medium=email", "https://doi.org/10.1000/xyz"},
		{"https://example.com/a?z=1&z=0&y", "https://example.com/a?y&z=0&z=1"},
	}
	for _, c := range cases {
		got, err := CanonicalURL(c.in)
		if err != nil {
			t.Errorf("CanonicalURL(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("CanonicalURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCanonicalURLErrors(t *testing.T) {
	for _, in := range []string{"", "   ", "not a url", "/relative/path", "doi:", "https://doi.org/", "http://[::1"} {
		if got, err := CanonicalURL(in); err == nil {
			t.Errorf("CanonicalURL(%q) = %q, want error", in, got)
		}
	}
	if CanonicalOrRaw(" not a url ") != "not a url" {
		t.Fatal("CanonicalOrRaw must fall back to the trimmed input")
	}
}

func TestNormalizeTextAndShingles(t *testing.T) {
	if got := NormalizeText("  Hello,   WORLD! It's 42%. "); got != "hello world it s 42" {
		t.Fatalf("NormalizeText = %q", got)
	}
	sh := Shingles("one two three four", 3)
	if len(sh) != 2 || sh[0] != "one two three" || sh[1] != "two three four" {
		t.Fatalf("Shingles = %v", sh)
	}
	if sh := Shingles("one two", 3); len(sh) != 1 || sh[0] != "one two" {
		t.Fatalf("short Shingles = %v", sh)
	}
	if Shingles("", 3) != nil {
		t.Fatal("empty text must have no shingles")
	}
}

const pressRelease = "Today the Institute published its annual evaluation report. The report finds that measured task horizons roughly doubled over the past year across the tracked benchmark suite, while contamination risk remains moderate. The authors caution that the sample is small and that scaffolding differences complicate comparison across laboratories."

func TestSimhashNearDuplicate(t *testing.T) {
	a := Simhash(pressRelease)
	if a != Simhash(pressRelease) {
		t.Fatal("simhash must be deterministic")
	}
	// Same text with a different lead sentence and punctuation.
	b := Simhash("BREAKING: " + strings.Replace(pressRelease, "annual", "yearly", 1))
	if d := Hamming(a, b); d > DefaultHammingThreshold {
		t.Fatalf("near-duplicate distance %d too large", d)
	}
	c := Simhash("An entirely unrelated note about weather patterns in the northern hemisphere and their effect on migratory birds during autumn.")
	if d := Hamming(a, c); d <= DefaultHammingThreshold {
		t.Fatalf("unrelated texts should be far apart, got %d", d)
	}
	if Hamming(0, ^uint64(0)) != 64 || Hamming(5, 5) != 0 {
		t.Fatal("Hamming basic cases")
	}
}

func TestClustererOneEventCountsOnce(t *testing.T) {
	c := NewClusterer(0)
	first := c.Assign("doc-1", "https://institute.example/press/report", "hash-1", "Institute publishes evaluation report", pressRelease)
	if first.Duplicate || !strings.HasPrefix(first.ClusterID, "dup-") {
		t.Fatalf("first document must start a cluster: %+v", first)
	}
	// Same URL with tracking parameters already canonicalized upstream → same canonical.
	again := c.Assign("doc-2", "https://institute.example/press/report", "", "Institute publishes evaluation report", pressRelease)
	if !again.Duplicate || again.ClusterID != first.ClusterID || again.Match != MatchURL {
		t.Fatalf("same canonical url must join the cluster: %+v", again)
	}
	// An article that repeats the press release nearly verbatim.
	article := c.Assign("doc-3", "https://news.example/story/1", "hash-3", "Report: task horizons doubled", "In a release today, "+pressRelease)
	if !article.Duplicate || article.ClusterID != first.ClusterID {
		t.Fatalf("near-duplicate article must join the cluster: %+v", article)
	}
	// Same content hash at another URL.
	mirror := c.Assign("doc-4", "https://mirror.example/x", "hash-1", "anything", "different words entirely here now")
	if !mirror.Duplicate || mirror.Match != MatchContent || mirror.ClusterID != first.ClusterID {
		t.Fatalf("content hash match must join the cluster: %+v", mirror)
	}
	// Unrelated document starts its own cluster.
	other := c.Assign("doc-5", "https://other.example/y", "hash-5", "Weather", "A note about weather patterns in the northern hemisphere and their effect on migratory birds during autumn and winter seasons.")
	if other.Duplicate || other.ClusterID == first.ClusterID {
		t.Fatalf("unrelated document must not join: %+v", other)
	}
	if c.Clusters() != 2 || c.Size(first.ClusterID) != 4 || c.Size(other.ClusterID) != 1 {
		t.Fatalf("cluster bookkeeping: clusters=%d sizes=%d/%d", c.Clusters(), c.Size(first.ClusterID), c.Size(other.ClusterID))
	}
}

func TestClustererShortTextsUseExactMatchOnly(t *testing.T) {
	c := NewClusterer(0)
	a := c.Assign("a", "https://x.example/1", "", "Short", "hello")
	b := c.Assign("b", "https://x.example/2", "", "Short", "hullo")
	if b.Duplicate || b.ClusterID == a.ClusterID {
		t.Fatalf("short texts must not be simhash-matched: %+v", b)
	}
	d := c.Assign("d", "https://x.example/3", "", "short", "HELLO!")
	if !d.Duplicate || d.Match != MatchExact || d.ClusterID != a.ClusterID {
		t.Fatalf("identical normalized short text must match exactly: %+v", d)
	}
}

func TestClusterIDForIsStable(t *testing.T) {
	a := ClusterIDFor("https://a.example/x", "", "")
	b := ClusterIDFor("https://a.example/x", "other", "text")
	if a != b || len(a) != len("dup-")+16 {
		t.Fatalf("cluster id must depend on the canonical url only: %s vs %s", a, b)
	}
	if ClusterIDFor("", "Title", "Body") == ClusterIDFor("", "Title", "Other body") {
		t.Fatal("without a url the text must drive the id")
	}
}
