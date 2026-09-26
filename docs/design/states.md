<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# States

Every page has a small number of states beyond "release loaded and rendered". This page lists
them with the exact copy and the code that produces it, so that a state is never improvised.

## 1. Loading

Pages are server-rendered from the release, so a page arrives complete; there is no client
fetch for data and no skeleton. The exceptions:

| Where | Behaviour |
| --- | --- |
| Home hero | `HomeStage` renders the static SVG disk (`StaticDisk`) before any scene; immersive scenes load with `next/dynamic` (`ssr: false`, `loading: () => null`) over it |
| `/lab` | The client component is wrapped in `<Suspense fallback="Loading the Scenario Lab…">`; while a run computes, the result region has `aria-busy` and the value shows "…" with the caption "computing" |
| Method documents | Rendered on the server from `docs/method/*.md`; no loading state |

## 2. No release

When `data/releases/CURRENT` is missing or unreadable:

| Where | Behaviour |
| --- | --- |
| Layout footer | "No release loaded." (`Footer.tsx`; the layout catches the SDK error) |
| Pages that call `getRelease()` | The SDK throws; Next.js renders `app/error.tsx` (section 6) |
| `MeterPanel` | Headline "No release" when the official object is absent |
| Share card | Renders the brand alone with "No release" as the official text |
| API | `/readyz` returns `503 {"status":"not_ready","detail":"no promoted release is loaded"}`; data routes are unavailable; `/healthz` still returns `200` |

The SDK re-reads `CURRENT` at most every 5 s, so a promotion appears without a restart; a
failed reload keeps the previous release.

## 3. WebGL unavailable

`HomeStage` probes `canvas.getContext("webgl2") || getContext("webgl")` on mount. If the probe
fails, the stage renders `StaticDisk` with `data-scene="static"` and `aria-hidden="true"`; the
meter beneath is unaffected. The Orrery and Branching modes are SVG/CSS and do not depend on
WebGL.

## 4. Reduced motion

Default mode `observatory`; scenes replaced by the static disk; durations zeroed
([`motion-semantics.md`](motion-semantics.md) §3).

## 5. Empty and partial data

| Where | Copy |
| --- | --- |
| `MeterPanel`, horizon without an external group | "No external forecast group asks about the {horizon} horizon; external aggregates exist for 2100 and open-ended horizons. See Forecasts." |
| `Timeline` with no points | "No measurements recorded." |
| `SignalTable`, signal without an observation | "no observation" in the observation cell; "—" in the numeric cells |
| `IndexGauge`, `value: null` (the attention index) | "—" as the value and "not computed" beneath; the accessible name reads "not computed" |
| `IndexGauge`, coverage below 1 | "index, not a probability · coverage 85%" (control strength in `rel-2026-09-26-001` covers 85% of its weight; the missing signal is `D5.control_evaluation_maturity`) |
| `/evidence` with filters that match nothing | The table renders with its caption "Source ledger (filtered)" and no rows; the tier and status bar lists above it still show the totals |
| `/act?audience=<unknown>` | The audience is ignored and all audiences are shown |
| `/compare` with no estimate for the horizon | The horizon tabs render; the card area is empty |
| Scenario Lab, `result.flags` | Implausible combinations are flagged in amber badges, never refused |
| Share button, clipboard unavailable | "Copy failed — select the address bar" for 2.2 s, then the label returns |
| `localStorage` unavailable | The mode choice lives for the page only (`ModeProvider` catches) |

## 6. Error page

`apps/web/app/error.tsx` (client boundary):

> **Something did not render**
> The page failed before it could show its data. No estimate has changed; the release on
> disk is unaffected.
> Reference {digest} · [Try again] [Read the plain-text version]

The copy states that nothing published changed, because a rendering error must not be read as
a change in the risk picture.

## 7. Not found

`apps/web/app/not-found.tsx`:

> **Not found**
> There is no page at this address. Nothing is hidden: everything published is reachable from
> the pages below. The meter · Plain-text observatory · Methodology

Dynamic routes call `notFound()` for ids that fail their pattern (`rel-YYYY-MM-DD-NNN`,
`[a-z0-9-]+` method slugs, unknown scenario ids). The export route returns
`404 {"error":"unknown export","available":[…]}` for an unknown name and
`{"error":"release not found"}` for an unknown `?release=`.

## 8. Insufficient calibration (the withheld official value)

This is the normal state of the headline in the v0 release line, not an error:

- `MeterPanel` prints the official object's `display.central` ("Insufficiently calibrated") in
  `.meter-value.withheld`: grey (`--c-insufficient`) and smaller than a number would be, with
  the eyebrow "p(DOOM) · 10 years · O3–O8 combined" and the lede taken from `display.note`
  ("The observatory publishes indexes, an external forecast aggregate and a research-mode
  model; it does not publish an official probability yet.").
- `EstimateCard` for a withheld estimate shows the "Official value withheld" badge with the
  hollow marker, no interval line and no quantile strip, and the note in muted text.
- The share card carries the same text.
- `/text#estimate` prints it first, followed by the two separately labelled quantities.
- The API returns the object with `quantiles: null` and the invariants guarantee it stays so.

## 9. Superseded release

`/releases/[releaseId]` states "Superseded by {id} on {date}." in the lede when the manifest has
`superseded`; the changelog table marks the current release with a "current" badge and others
with "superseded by …". Superseded releases remain fully readable.

## 10. Print

Navigation, `.no-print` controls and canvases are hidden; `<details>` tables open; colours
switch to black on white.

## Not yet implemented

- An explicit empty-state sentence on `/evidence` when filters match no source (today the
  filtered table is simply empty).
- Visible loading and error states for the immersive scenes: today a scene fades in when ready and
  disappears silently on failure, leaving the static disk.
- A dedicated "release superseded since you opened this page" notice; the SDK reload is
  silent.
