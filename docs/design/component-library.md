<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Component library

Inventory of `apps/web/components`, with each component's props and the accessibility
contract it honours. All chart components are pure SVG or HTML React components that render on
the server; client components are marked. Data enters components as props from pages that read
the promoted release through `@pdoom/sdk`; no component fetches, stores or computes an estimate
(the Scenario Lab computes a `user_scenario` result in the browser and stores nothing).

## Shared contract

- Every chart renders `<figure>` with a `<figcaption>` that holds the `description` prop and,
  for SVG charts, a `<details><summary>Data table</summary>` containing the values.
- SVG charts carry `role="img"` and an `aria-label` that summarises the data in words.
- Colours are semantic tokens passed as `colorVar` (default `var(--c-evidence)`); meaning is
  also carried by text ([`color.md`](color.md) §5).
- Numbers are formatted by the caller or by `@pdoom/model-core` rounding; components never
  contain a numeric literal that is a probability (guard test).
- Nothing depends on hover: SVG `<title>` tooltips are supplements to the table.

## charts/

| Component | Props | Renders | Accessibility |
| --- | --- | --- | --- |
| `QuantileStrip` | `p05 p25 p50 p75 p95: number`, `max?`, `label`, `description`, `colorVar?`, `compact?`, `children?` | Interval bar: thin p05–p95 band, thick p25–p75 band, median tick; axis ticks unless compact | `aria-label` "{label}: median x, 5th to 95th percentile a to b"; data table with the five quantiles |
| `LineWithBand` | `points: {x, low, mid, high}[]`, `title`, `description`, `max?`, `colorVar?` | Categorical-x curve of medians with a p05–p95 band (used for the research-mode curve across horizons) | `aria-label` with first and last median; table Horizon/p05/median/p95 |
| `BarList` | `items: {label, value, display?, note?, colorVar?, href?}[]`, `max?`, `title`, `description`, `unit?` | Labelled horizontal bars on a shared scale with the value printed at the right | `<ul aria-label={title}>`; bars `aria-hidden`; value always printed |
| `Waterfall` | `items: {label, value, direction, note?}[]`, `title`, `description` | Contribution bars in index points, safeguard colour when `direction === "strengthens_control"`, risk colour otherwise | as `BarList`; "+x.x pts" printed |
| `DistributionDots` | `dots: {label, value, group, colorVar?}[]`, `title`, `description`, `max?` | One row per outcome-set/horizon group; each forecast a dot on a shared probability axis | `aria-label` with dot and group counts; table Group/Forecast/Value; `<title>` per dot |
| `NodeGraph` | `nodes: {id, label, group, href?, colorVar?}[]`, `edges: {from, to, label, strength?}[]`, `title`, `description` | Deterministic circular layout ordered by group; quadratic arcs with arrowheads; stroke weight by `strength` | `aria-label` with node and relation counts; nodes with `href` are links with `aria-label` "{id}: {label}"; the relations table is the authoritative form |
| `Timeline` | `points: {date, value, label, low?, high?, note?}[]`, `title`, `description`, `unit`, `log?` | Dated measurements with optional whiskers; log axis when the range spans more than two orders of magnitude, stated in the caption | `aria-label` with count and date span; table Date/Model/value/Interval/Note; "No measurements recorded." when empty |
| `IconArray` | `p: number`, `cells?` (default 1000), `label`, `colorVar?`, `low?`, `high?` | N cells, round(p·N) filled first in reading order; faint cells to the upper interval | `aria-label` "{label}: n of N cells filled; plausible range a to b"; caption repeats it |

## meter/

| Component | Props | Renders | Accessibility |
| --- | --- | --- | --- |
| `EstimateCard` | `e: Estimate`, `compact?`, `hideStrip?` | The one card through which every probability is shown: status badge, horizon badge, disagreement and uncertainty badges, outcome-set eyebrow, display value in the status colour, interval with rounding rule, `QuantileStrip`, meta (producer, origin date, evidence-through date, forecast and population counts), `<details>` with conditioning, assumptions and method link. Withheld estimates show the note instead of a strip | `<article aria-label="{status}: {outcome set}, {horizon}">`; all badges are text |
| `IndexGauge` | `index: IndexValue` | 0–100 value in the index's colour, bar, and "index, not a probability · coverage N%" (or "—"/"not computed" when `value` is null) | value has `aria-label` "{label}: N out of 100"; bar `aria-hidden` |
| `MeterPanel` | `rel: Release`, `horizon: string` | The above-the-fold meter: withheld headline, external and research-mode cards, decomposition (O3, O4+O5, O6, O7), six gauges, top upward and downward drivers, "What changed", action buttons | `<section aria-labelledby="meter-heading">`; h2 is the headline |
| `SignalTable` | `drivers: Driver[]`, `observations: DriverObservation[]`, `weights?` | Table of signals with latest observation, normalised value, confidence, kind, weight and the normalisation recipe; rationale and sources in a `<details>` per row | `<caption>`; rows have `id={signal_id}` for deep links |

## shell/

| Component | Props | Renders | Accessibility |
| --- | --- | --- | --- |
| `Header` | none | Wordmark with `PUBLIC_LABEL`, `Nav`, compact `ModeSwitcher` | `<header>`; wordmark link `aria-label="p(DOOM) home"` |
| `Nav` (client) | none | Eleven primary links (Meter, Futures, Evidence, Capabilities, Agents, Incidents, Forecasts, Safeguards, Act, Method, Text) | `<nav aria-label="Primary">`; `aria-current="page"` on the active section |
| `PageHeader` | `crumbs?`, `title`, `lede?`, `textAnchor?`, `children?` | Breadcrumbs, the page's single `<h1>`, a "Plain text" button linking to `/text#{anchor}`, optional lede and children (horizon tabs) | `<nav aria-label="Breadcrumb">`; one h1 per page |
| `Footer` | `releaseId?`, `dataCutoff?`, `published?` | Brand line with `BRAND_EXPANDED` and `AUTHOR`, release link with data-cutoff and publication dates or "No release loaded.", method/changelog/text/content-safety links and the fixed sentence "No official probability is published in this release line. Indexes are not probabilities. Nothing here predicts a date." | `<footer>` |

## mode/

| Component | Props | Renders | Accessibility |
| --- | --- | --- | --- |
| `ModeProvider` (client) | `children` | Context `{mode, setMode, reducedMotion, hydrated}`; reads `localStorage["pdoom.mode"]` and `prefers-reduced-motion` on hydration; default `observatory` under reduced motion, else `event-horizon`; storage failures are caught | none visible |
| `ModeSwitcher` (client) | `compact?` | Buttons for each immersive mode plus a link to `/text`; compact shows Event Horizon, Observatory, Plain text; selecting a mode navigates to `/` | `role="group" aria-label="Presentation mode"`; buttons carry `aria-pressed`; the text link carries `aria-current="page"` on `/text` |
| `modes.ts` | — | `MODES`, `MODE_LABEL`, `MODE_STORAGE_KEY`, `isMode` | — |

## share/

| Component | Props | Renders | Accessibility |
| --- | --- | --- | --- |
| `ShareLink` (client) | `href?`, `label?` | "Copy link" button (states: label, "Link copied", "Copy failed — select the address bar") and a "Share…" button when `navigator.share` exists; copies the current URL without tracking parameters | `role="group" aria-label="Share"`; copy button `aria-live="polite"` |

## lab/

| Component | Props | Renders | Accessibility |
| --- | --- | --- | --- |
| `ScenarioLab` (client) | `spec: ExperimentalCausalSpec`, `baseline: Record<horizon, {p05,p50,p95,display,interval} | null>`, `releaseId` | Horizon buttons; ten range sliders (−2…2) with factor annotations; sample-count select; reset; result card with the mandated sentence "Under your selected assumptions, not the p(DOOM) official model, the median estimate is", the rounded value, the interval, a `QuantileStrip`, the published research-mode baseline for comparison, flags, decomposition and factor tables, `ShareLink`, and the model-core disclaimer. State lives in the URL (`?h=10y&ct=1…`) so every result is reproducible from its link | horizon buttons `aria-pressed`; each slider has a `<label for>`, `aria-valuetext` ("baseline", "much faster", …) and a `<datalist>` of ticks; the result region is `aria-live="polite"` and `aria-busy` while computing; tables have captions |

## scenes/

| Component | Props | Renders | Accessibility |
| --- | --- | --- | --- |
| `HomeStage` (client) | `data: StageData` | Chooses the scene for the mode; static disk before hydration, in Observatory/Text modes, under reduced motion or without WebGL; dynamic-imports the scenes | Static disk wrappers are `aria-hidden`; scene wrappers carry `aria-label` "Conceptual risk visualization — not a simulation of AI risk" |
| `StaticDisk` | `intervalWidth?`, `brightness?`, `particles?`, `safeguards?`, `seed?` | Server-renderable SVG accretion disk following [`motion-semantics.md`](motion-semantics.md) §1 | `role="img"` with a full-sentence `aria-label` describing the grammar and the disclaimer |
| `event-horizon/EventHorizonScene` (client) | `data` | React Three Fiber canvas over the static disk: horizon, accretion band (thickness = interval width), particle field (count = sources), safeguard arcs, faint research-curve points; quality manager, pause when hidden or out of view, skippable intro, reduced-motion still frame, error boundary | Wrapper `aria-hidden`; no text or numbers in the canvas; `pointer-events: none` |
| `orrery/OrreryScene` (client) | `data`, `reducedMotion` | SVG orrery: withheld official object at the centre, indexes, scenarios (recoverability shapes and colours), safeguard markers on slow CSS orbits | `role="img"` with a conceptual label; `<title>` per node; static under reduced motion |
| `branching/BranchingScene` (client) | `data`, `reducedMotion` | SVG branching tree from "now"; branch thickness = scenarios per recoverability group; dashed safeguard ties; grow-in animation | `role="img"` with a conceptual label; static under reduced motion |
| `charts/FlowDiagram` | `left`, `right`, `flows`, `title`, `description` | Sankey-lite: ribbons from scenarios (by recoverability) to the outcomes they can reach; ribbon width = number of scenarios | `role="img"` label plus a flows table |
| `charts/IconArray` | `p`, `cells?`, `label`, `low?`, `high?` | N cells with round(p·N) filled; used on `/compare` | `role="img"` with a sentence summary |

## Not yet implemented

- In-canvas loading and error overlays for the scenes; failures are silent by design (the static disk remains).
- Page-section components under `components/{forecasts,capabilities,evidence,futures,incidents,agents,safeguards,act,method}` named in the ownership matrix; the pages render their sections inline today.
