<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Accessibility

p(DOOM) commits to WCAG 2.2 level AA for the web app, and to the stronger rule of the build
specification that nothing substantive may exist only in a 3D scene (`/text` contains
everything). This page maps each commitment to how it is met and lists the known gaps.

## 1. Commitments and how they are met

| Commitment (WCAG 2.2 reference) | How it is met | Where |
| --- | --- | --- |
| Bypass blocks (2.4.1) | A "Skip to content" link is the first focusable element and targets `<main id="main">` | `apps/web/app/layout.tsx`, `.skip-link` in `base.css` (visible on focus) |
| Landmarks and headings (1.3.1, 2.4.6) | `<header>`, `<nav aria-label="Primary">`, `<nav aria-label="Breadcrumb">`, `<main>`, `<footer>`; exactly one `<h1>` per route through `PageHeader`; sections with `aria-labelledby` | `components/shell/*`, `MeterPanel` |
| Language of page (3.1.1) | `<html lang="en">` | `layout.tsx` |
| Non-text content (1.1.1) | Every SVG chart has `role="img"` and an `aria-label` that states the data in words; decorative bars and the static disk behind the meter are `aria-hidden`; the static disk, when it is the visible stage, has a full-sentence accessible name including the disclaimer | `components/charts/*`, `StaticDisk`, `HomeStage` |
| Info and relationships (1.3.1) | Every chart is a `<figure>` with `<figcaption>`; SVG charts include a `<details>` data table with a `<caption>` and header cells; all tables on pages carry `<caption>`; key–value data uses `<dl>` | `components/charts/*`, `/text` |
| Colour not the only means (1.4.1) | Badges carry shape markers and text; statuses are words; recoverability is written in headings; bars print their numbers | [`../design/color.md`](../design/color.md) §5 |
| Contrast (1.4.3) | Text tokens: `--c-text #f2f3f7`, `--c-text-2 #b6bac8`, `--c-text-3 #7d8394` on `--c-bg #050508`; light theme uses darker semantic colours; `prefers-contrast: more` raises secondary text and borders | `tokens.css` |
| Reflow and text spacing (1.4.10, 1.4.12) | Fluid grid (`minmax(min(100%, …), 1fr)`), 16px gutter, `.table-wrap` horizontal scroll for wide tables, rem-based type scale, `clamp()` display sizes | `globals.css` |
| Keyboard (2.1.1) and focus visible (2.4.7, 2.4.11) | All controls are native buttons, links, inputs and selects; `:focus-visible` gives a 3px outline with offset; node-graph links get a visible stroke on focus; the sticky header does not cover focused content (`scroll-margin-top` on headings) | `base.css`, `globals.css` |
| Name, role, value (4.1.2) | Mode buttons carry `aria-pressed`; horizon buttons in the lab carry `aria-pressed`; horizon and audience tabs use `aria-current="page"`; sliders have `<label for>`, `aria-valuetext` and a `<datalist>`; share group and mode group have `role="group"` with labels | `ModeSwitcher`, `ScenarioLab`, `ShareLink` |
| Status messages (4.1.3) | The Scenario Lab result region is `aria-live="polite"` with `aria-busy` while computing; the copy-link button is `aria-live="polite"` | `ScenarioLab`, `ShareLink` |
| Motion (2.3.3) | No keyframe animations; transitions only on state change; `prefers-reduced-motion` zeroes durations, selects the still Observatory mode and replaces scenes with the static disk | [`../design/motion-semantics.md`](../design/motion-semantics.md) |
| Content on hover or focus (1.4.13) | Nothing is hover-only: details are in `<details>`/`<summary>`; SVG `<title>` tooltips duplicate table rows | `components/charts/*`, `SignalTable` |
| Consistent navigation and identification (3.2.3, 3.2.4) | The same header, footer, breadcrumb and "Plain text" button on every route; status words and badge shapes are fixed site-wide | `PageHeader`, `lib/format.ts` |
| Forms (3.3.1, 3.3.2) | The evidence filter form has labelled controls and submits with GET; submissions to the API return every validation problem at once | `/evidence`, `internal/api` |
| Works without JavaScript | All routes are server-rendered; `/text` is complete with ordinary anchors and no client components; the mode switcher and Scenario Lab are progressive enhancements | `app/text/page.tsx` |
| Print | Black on white, navigation hidden, details opened, URLs printed after links | `base.css` |
| Target size (2.5.8) | Buttons and tabs are pill-shaped with `--s-2`/`--s-3` padding on `--fs-sm` text | `globals.css` |

## 2. Immersive modes

Scenes are progressive enhancement over a complete, server-rendered page. A scene never
carries information that is not in the meter or in `/text`; the mode switcher sits in the
first viewport; reduced motion and missing WebGL both fall back to the static disk. The
scenes themselves are placeholders at present.

## 3. Testing

- The Playwright configuration (`tests/e2e/playwright.config.ts`) defines `desktop`, `mobile`
  (Pixel 7) and `reduced-motion` projects, and `@axe-core/playwright` is installed for automated
  checks of key pages.
- The build specification lists these end-to-end checks: routes, mode switching, reduced
  motion, WebGL failure, `/text` without JavaScript, axe on key pages.

## 4. Known gaps

- **No end-to-end or axe specs exist yet.** `tests/e2e` contains only the configuration, and
  `tests/accessibility` is empty; the commitments above are met by construction and by review,
  not yet by automated tests.
- **Contrast has not been machine-verified.** A hand calculation puts `--c-text-3` on `--c-bg`
  above 4.5:1 in the dark theme and just above it in the light theme; automated verification is
  required before the light theme is exposed through a switcher.
- **Node-graph labels are truncated** at 26 characters in the SVG (the full label is in the
  link's `aria-label` and in the relations table).
- **Wide tables scroll horizontally** inside `.table-wrap`; there is no responsive
  card layout for them on narrow screens.
- **The `/text` table of contents uses two CSS columns**, which read in visual rather than
  DOM order for sighted keyboard users on wide screens; the DOM order is correct.
- **Form validation on `/evidence`** relies on the server ignoring invalid values; there is no
  inline error text.
- **Theme and palette switches** (`data-theme`, `data-palette`) have no UI yet.
- **Scenes are placeholders**; when built they must honour pause-on-hidden, a skippable intro,
  a visible label and the reduced-motion fallback.

## Not yet implemented

- `tests/e2e/*.spec.ts` and `tests/accessibility/*` (routes, mode switching, reduced motion,
  WebGL failure, `/text` without JavaScript, axe on `/`, `/meter`, `/futures`, `/lab`, `/text`).
- An accessibility statement page inside the web app; this document is the statement.
