<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Typography

Type tokens live in `packages/design-system/tokens.css`; the base rules in
`packages/design-system/base.css`; page-level rules in `apps/web/app/globals.css`.

## 1. Font tokens

| Token | Stack |
| --- | --- |
| `--font-display` | "Space Grotesk", "Manrope", "Avenir Next", "Segoe UI", system-ui, sans-serif |
| `--font-body` | "IBM Plex Sans", "Source Sans 3", "Inter", system-ui, -apple-system, "Segoe UI", Roboto, sans-serif |
| `--font-mono` | "IBM Plex Mono", "JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace |

No font files ship with the repository and there are no `@font-face` rules, so each stack
resolves to the first face installed on the reader's device and otherwise to the system face.
The web CSP (`font-src 'self' data:`) forbids fetching fonts from a third party, so this is
deliberate: the page never depends on a font CDN.

Roles: `--font-display` for `h1`–`h4`, the wordmark, the eyebrow labels and every large number
(`.meter-value`, `.estimate-card .value`, `.gauge .num`, `.lab-value`); `--font-body` for
everything else; `--font-mono` for ids, commands, hashes and code (`code`, `kbd`, `pre`).

## 2. Size and rhythm

| Token | Value | Typical use |
| --- | --- | --- |
| `--fs-xs` | 0.75rem | badges, eyebrows, `.cite`, table headers, chart labels (11px in SVG) |
| `--fs-sm` | 0.875rem | tables, navigation, buttons, captions, `.kv` lists |
| `--fs-md` | 1rem | body |
| `--fs-lg` | 1.125rem | `.lede`, `.hero-sub`, `h4` |
| `--fs-xl` | 1.375rem | `h3`, wordmark, withheld value in `EstimateCard` |
| `--fs-2xl` | 1.75rem | `h2`, `.gauge .num` |
| `--fs-3xl` | 2.25rem | `h1`, `.estimate-card .value` |
| `--fs-4xl` | 3rem | reserved |
| `--fs-hero` | `clamp(3rem, 9vw, 7rem)` | home `<h1>` |

Line heights: `--lh-tight 1.1` (headings), `--lh-snug 1.3`, `--lh-body 1.6`. Headings use
`text-wrap: balance` and negative letter-spacing (−0.01em base; −0.02em to −0.04em for the
wordmark, page titles and large numbers). The meter value scales with
`clamp(2rem, 6vw, 4rem)` and the Scenario Lab value with `clamp(3rem, 9vw, 6rem)`.

Spacing follows a 4px grid (`--s-1` 0.25rem … `--s-9` 6rem); `.stack > * + *` adds `--s-4`
between siblings.

## 3. Measure

`--measure: 68ch` bounds `p`, `ul`, `ol`, `dl` and `.prose`; `.lede` and `.hero-sub` are
narrower (68ch and 52ch); `/text` uses 78ch for its long tables; the container is 1200px with a
16px gutter (`--gutter`) so nothing touches the viewport edge on a phone.

## 4. Numbers

- `body { font-variant-numeric: tabular-nums }` so digits align in every table and list;
  `td.num, th.num` are right-aligned and repeat `tabular-nums`.
- Large numbers use the display face; the withheld official value is deliberately smaller and
  grey than a number would be (`.meter-value.withheld`).
- Probabilities are formatted only by `roundForDisplay`/`formatInterval` from
  `@pdoom/model-core` with the estimate's rounding step, giving "2%", "<1%", ">99%" and
  "<1%–14%" (en dash, no spaces). Decimals never appear in an estimate.
- Counts use `toLocaleString("en-US")` (for example "20,000 samples", "1,000 cells").
- Indexes are shown as integers out of 100 with the words "index, not a probability".
- Source values (member forecasts, benchmark results) are shown as recorded, with their unit,
  beside the original wording; they are data, not estimates.

## 5. Text elements

- Links: `--c-link`, 1px underline offset 0.15em, thicker on hover; external links print their
  URL after the text in print media.
- Focus: `:focus-visible` gets a 3px `--c-focus` outline with 2px offset.
- Eyebrows (`.eyebrow`): display face, `--fs-xs`, 0.12em tracking, uppercase, `--c-text-3`.
- Badges (`.badge`): `--fs-xs`, weight 600, 0.04em tracking, uppercase, pill, marker before the
  text ([`color.md`](color.md) §5).
- Citations (`.cite`): `--fs-xs`, `--c-text-3`, used for source ids, method references and
  the mandated Scenario Lab disclaimer.
- Quotes of original question wording use `.quote` (italic, left rule) on `/forecasts`.
- Markdown rendered from `docs/method/*.md` uses `.prose` with sticky "On this page" links.

## 6. Print

`@media print` sets black on white, hides navigation and canvases, opens every `<details>` so
data tables print, and appends URLs to external links.

## Not yet implemented

- Shipping and subsetting the named brand fonts (the build specification lists a `fonts.css`
  in the design-system package; it does not exist).
- A typographic scale for the immersive scenes' in-canvas labels (scenes are placeholders).
