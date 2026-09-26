// Copyright NU Cybernetics. p(DOOM) — research prototype.
// Guard: no probability or percentage literal may be written into web source.
// Every displayed number must come from the promoted release or the model.
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const root = fileURLToPath(new URL("..", import.meta.url));
const dirs = ["app", "components", "lib"];

function walk(dir: string, out: string[] = []): string[] {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) walk(p, out);
    else if (/\.(tsx?|mdx?)$/.test(name)) out.push(p);
  }
  return out;
}

// A digit sequence immediately followed by a percent sign inside JSX text or a
// string literal, e.g. "12%" or >5%<. Template placeholders like `${x}%` are fine,
// and CSS/SVG geometry (width: "100%", offset="35%") is stripped before matching.
const GEOMETRY = /(offset|width|height|left|top|right|bottom|padding|margin|inset|flex|translate|rx|ry|cx|cy|r|x|y)\s*[:=]\s*["'{`]?\s*\d+(\.\d+)?%/g;
const PCT = /(?<![\w$}{])\d+(\.\d+)?\s?%/;
// Common disallowed phrasings.
const PHRASES = [/humanity has \w+ years left/i, /doom is (now )?certain/i, /the machines are coming/i];

describe("no hardcoded percentages or panic phrasing in web source", () => {
  const files = dirs.flatMap((d) => walk(join(root, d)));
  it("scans a meaningful number of files", () => {
    expect(files.length).toBeGreaterThan(5);
  });
  for (const f of files) {
    it(`${f.replace(root, "")} has no percentage literal`, () => {
      const src = readFileSync(f, "utf8")
        .split("\n")
        .filter((line) => !line.trim().startsWith("//") && !line.trim().startsWith("*"))
        .join("\n")
        .replace(GEOMETRY, "");
      const m = src.match(PCT);
      expect(m, m ? `found "${m[0]}"` : "").toBeNull();
      for (const p of PHRASES) expect(src.match(p)).toBeNull();
    });
  }
});
