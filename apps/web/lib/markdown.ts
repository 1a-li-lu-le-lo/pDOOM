// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * A small Markdown renderer for the repository's own methodology and
 * governance documents (headings, paragraphs, lists, tables, fenced code,
 * block quotes, links, emphasis and inline code). All text is HTML-escaped
 * before markup is applied, and only http(s), mailto and relative links are
 * emitted; raw HTML in the source is shown as text. It renders trusted repo
 * content only and is intentionally not a general-purpose parser.
 */

export interface Heading {
  depth: number;
  text: string;
  id: string;
}

export interface Rendered {
  html: string;
  headings: Heading[];
  title: string;
}

const ESC: Record<string, string> = { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" };
export function escapeHtml(s: string): string {
  return s.replace(/[&<>"']/g, (c) => ESC[c] ?? c);
}

export function slugify(s: string): string {
  return s
    .toLowerCase()
    .replace(/`/g, "")
    .replace(/[^a-z0-9\s-]/g, "")
    .trim()
    .replace(/\s+/g, "-")
    .slice(0, 80);
}

function safeHref(href: string): string | null {
  const h = href.trim();
  if (/^(https?:|mailto:)/i.test(h)) return h;
  if (/^[./#a-z0-9_-]/i.test(h) && !/^[a-z]+:/i.test(h)) {
    // Relative links to other method docs: `calibration.md` → `/method/calibration`
    const m = h.match(/^(?:\.\/)?([a-z0-9-]+)\.md(#.*)?$/i);
    if (m) return `/method/${m[1]}${m[2] ?? ""}`;
    return h;
  }
  return null;
}

function inline(src: string): string {
  let s = escapeHtml(src);
  // inline code first so its contents are not styled
  const codes: string[] = [];
  s = s.replace(/`([^`]+)`/g, (_m, c: string) => {
    codes.push(`<code>${c}</code>`);
    return `\uE000${codes.length - 1}\uE000`;
  });
  s = s.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (_m, text: string, href: string) => {
    const h = safeHref(href.replace(/&amp;/g, "&"));
    return h ? `<a href="${escapeHtml(h)}"${/^https?:/i.test(h) ? ' rel="noopener noreferrer"' : ""}>${text}</a>` : text;
  });
  s = s.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
  s = s.replace(/(^|[\s(])\*([^*\n]+)\*(?=[\s).,;:]|$)/g, "$1<em>$2</em>");
  s = s.replace(/(^|[\s(])_([^_\n]+)_(?=[\s).,;:]|$)/g, "$1<em>$2</em>");
  s = s.replace(/\uE000(\d+)\uE000/g, (_m, i: string) => codes[Number(i)] ?? "");
  return s;
}

function tableRow(line: string): string[] {
  return line
    .trim()
    .replace(/^\|/, "")
    .replace(/\|$/, "")
    .split("|")
    .map((c) => c.trim());
}

export function renderMarkdown(md: string): Rendered {
  const lines = md.replace(/\r\n/g, "\n").replace(/<!--[\s\S]*?-->/g, "").split("\n");
  const out: string[] = [];
  const headings: Heading[] = [];
  const ids = new Map<string, number>();
  let title = "";
  let i = 0;
  const para: string[] = [];
  const flushPara = () => {
    if (para.length) {
      out.push(`<p>${inline(para.join(" "))}</p>`);
      para.length = 0;
    }
  };
  while (i < lines.length) {
    const line = lines[i] ?? "";
    if (/^\s*$/.test(line)) {
      flushPara();
      i++;
      continue;
    }
    const fence = line.match(/^```(\w*)/);
    if (fence) {
      flushPara();
      const buf: string[] = [];
      i++;
      while (i < lines.length && !/^```/.test(lines[i] ?? "")) {
        buf.push(lines[i] ?? "");
        i++;
      }
      i++;
      out.push(`<pre><code${fence[1] ? ` class="language-${escapeHtml(fence[1])}"` : ""}>${escapeHtml(buf.join("\n"))}</code></pre>`);
      continue;
    }
    const h = line.match(/^(#{1,6})\s+(.*?)\s*#*\s*$/);
    if (h) {
      flushPara();
      const depth = h[1]!.length;
      const text = h[2]!;
      let id = slugify(text.replace(/\[([^\]]+)\]\([^)]*\)/g, "$1"));
      const n = ids.get(id) ?? 0;
      ids.set(id, n + 1);
      if (n) id = `${id}-${n}`;
      if (depth === 1 && !title) title = text.replace(/`/g, "");
      headings.push({ depth, text: text.replace(/`/g, ""), id });
      out.push(`<h${depth} id="${id}">${inline(text)}</h${depth}>`);
      i++;
      continue;
    }
    if (/^\s*(-{3,}|\*{3,})\s*$/.test(line)) {
      flushPara();
      out.push("<hr />");
      i++;
      continue;
    }
    if (/^\|/.test(line) && /^\s*\|?\s*:?-{3,}/.test(lines[i + 1] ?? "")) {
      flushPara();
      const head = tableRow(line);
      i += 2;
      const rows: string[][] = [];
      while (i < lines.length && /^\|/.test(lines[i] ?? "")) {
        rows.push(tableRow(lines[i] ?? ""));
        i++;
      }
      out.push(
        `<div class="table-wrap" tabindex="0" role="region" aria-label="Scrollable table"><table><thead><tr>${head.map((c) => `<th>${inline(c)}</th>`).join("")}</tr></thead><tbody>${rows
          .map((r) => `<tr>${r.map((c) => `<td>${inline(c)}</td>`).join("")}</tr>`)
          .join("")}</tbody></table></div>`,
      );
      continue;
    }
    if (/^>\s?/.test(line)) {
      flushPara();
      const buf: string[] = [];
      while (i < lines.length && /^>\s?/.test(lines[i] ?? "")) {
        buf.push((lines[i] ?? "").replace(/^>\s?/, ""));
        i++;
      }
      out.push(`<blockquote>${renderMarkdown(buf.join("\n")).html}</blockquote>`);
      continue;
    }
    const li = line.match(/^(\s*)([-*+]|\d+[.)])\s+(.*)$/);
    if (li) {
      flushPara();
      const ordered = /\d/.test(li[2]!);
      const items: { text: string; nested: string }[] = [];
      const baseIndent = li[1]!.length;
      const itemAt = (idx: number) => (lines[idx] ?? "").match(/^(\s*)([-*+]|\d+[.)])\s+(.*)$/);
      while (i < lines.length) {
        const m = itemAt(i);
        if (!m || m[1]!.length !== baseIndent) break;
        i++;
        const nested: string[] = [];
        while (i < lines.length && /^\s+\S/.test(lines[i] ?? "") && itemAt(i)?.[1]?.length !== baseIndent) {
          const raw = lines[i] ?? "";
          const indent = raw.length - raw.trimStart().length;
          nested.push(raw.slice(Math.min(baseIndent + 2, indent)));
          i++;
        }
        items.push({ text: m[3]!, nested: nested.length ? renderMarkdown(nested.join("\n")).html.replace(/^<p>([\s\S]*)<\/p>$/, " $1") : "" });
      }
      const tag = ordered ? "ol" : "ul";
      out.push(`<${tag}>${items.map((it) => `<li>${inline(it.text)}${it.nested}</li>`).join("")}</${tag}>`);
      continue;
    }
    para.push(line.trim());
    i++;
  }
  flushPara();
  return { html: out.join("\n"), headings, title };
}
