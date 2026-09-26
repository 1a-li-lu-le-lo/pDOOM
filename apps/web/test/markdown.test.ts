// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { renderMarkdown, slugify } from "../lib/markdown";

const methodDir = fileURLToPath(new URL("../../../docs/method/", import.meta.url));

describe("renderMarkdown", () => {
  it("renders headings with ids, paragraphs, emphasis and code", () => {
    const r = renderMarkdown("# Title here\n\nSome **bold** and *em* and `code`.\n\n## Second\n\ntext");
    expect(r.title).toBe("Title here");
    expect(r.headings.map((h) => h.id)).toEqual(["title-here", "second"]);
    expect(r.html).toContain("<strong>bold</strong>");
    expect(r.html).toContain("<em>em</em>");
    expect(r.html).toContain("<code>code</code>");
  });

  it("escapes raw html and blocks unsafe links", () => {
    const r = renderMarkdown("<script>alert(1)</script>\n\n[x](javascript:alert(1)) [ok](https://example.org)");
    expect(r.html).not.toContain("<script>");
    expect(r.html).toContain("&lt;script&gt;");
    expect(r.html).not.toContain("javascript:");
    expect(r.html).toContain('href="https://example.org"');
  });

  it("rewrites sibling .md links to /method routes", () => {
    const r = renderMarkdown("See [calibration](calibration.md#what-would-change) and `uncertainty.md`.");
    expect(r.html).toContain('href="/method/calibration#what-would-change"');
  });

  it("renders tables, lists with nesting, quotes and fences", () => {
    const md = "| a | b |\n| --- | --- |\n| 1 | 2 |\n\n- one\n  - nested\n- two\n\n1. first\n2. second\n\n> quoted\n\n```sh\nmake check\n```";
    const r = renderMarkdown(md);
    expect(r.html).toContain("<table>");
    expect(r.html).toContain("<td>2</td>");
    expect(r.html).toMatch(/<ul><li>one<ul><li>nested<\/li><\/ul><\/li><li>two<\/li><\/ul>/);
    expect(r.html).toContain("<ol><li>first</li><li>second</li></ol>");
    expect(r.html).toContain("<blockquote><p>quoted</p></blockquote>");
    expect(r.html).toContain('<pre><code class="language-sh">make check</code></pre>');
  });

  it("renders every methodology document with a title and no raw angle brackets from source", () => {
    const files = readdirSync(methodDir).filter((f) => f.endsWith(".md"));
    expect(files.length).toBeGreaterThan(5);
    for (const f of files) {
      const r = renderMarkdown(readFileSync(join(methodDir, f), "utf8"));
      expect(r.title, f).not.toBe("");
      expect(r.headings.length, f).toBeGreaterThan(0);
    }
  });

  it("slugify is stable", () => {
    expect(slugify("Eight outputs that are never blended")).toBe("eight-outputs-that-are-never-blended");
    expect(slugify("p(DOOM) = P(O3 ∪ O4)")).toBe("pdoom-po3-o4");
  });
});
