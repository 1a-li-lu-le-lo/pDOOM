<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Motion semantics

Every animated or drawn element on p(DOOM) encodes something from the release, and the
encoding is fixed so that two readers looking at the same release see the same picture. The
motion tokens in `packages/design-system/tokens.css` carry the comment "every animation has a
documented meaning"; this is the document.

## 1. Visual grammar of the stage

The home hero (`apps/web/app/page.tsx`) computes one `StageData` object from the current
release and snapshot and hands it to `HomeStage`, which renders the static disk
(`components/scenes/StaticDisk.tsx`) and, when a mode and WebGL allow, a scene over it.

| Element | Encodes | Source of the value | Mapping in `StaticDisk` |
| --- | --- | --- | --- |
| Ring thickness | Width of the plausible interval of the featured estimate (p95 − p05) | research-mode P_DOOM at the default horizon, else the first external aggregate | `ring = 18 + min(80, intervalWidth × 240)` |
| Disk brightness | Evidence Pressure Index (0–100) | `indexes.json`, `evidence_pressure` | `glow = 0.35 + min(0.55, brightness / 160)` |
| Orbit dots | Sources in the snapshot | `min(400, sources × 2)` | deterministic positions from a seeded generator (seed 20260926) |
| Cyan arcs | Safeguards | `min(12, interventions)` | arcs at equal angles on outer ellipses in `--c-safeguard` |
| The dark centre | Irreversible catastrophe (O3–O8) as a concept, not a quantity | fixed | radial gradient; no size mapping |
| Scene label | The disclaimer | fixed text | "Conceptual risk visualization · not a simulation of AI risk" above the hero; the same sentence is the SVG's accessible name |

The same grammar is used off-stage: `QuantileStrip` draws the p05–p95 band thin and the
p25–p75 band thick with the median as a tick, so "band thickness = interval" is the site-wide
rule for confidence; `LineWithBand` draws the research-mode curve across horizons with its
band; `IconArray` fills round(p × N) of N cells for the median and faint cells to the upper end
of the interval.

## 2. What motion may mean

- **A transition on state change.** Gauge bars grow to their value (`--dur-slow`, `--ease-out`)
  once when the page renders; buttons change surface on hover (`--dur-fast`); the Scenario Lab
  value changes colour with the result (`--dur-fast`).
- **Continuous motion in an immersive scene** may only represent the orbit of evidence around
  the concept, at a speed that does not depend on any estimate, so that speed is never read as
  urgency.
- **Nothing pulses, flashes, throbs or counts.** There are no CSS keyframe animations in
  `globals.css` or `base.css`; alarm motion is forbidden by
  [`../governance/psychological-safety.md`](../governance/psychological-safety.md).

Duration tokens: `--dur-fast 160ms`, `--dur-base 320ms`, `--dur-slow 700ms`,
`--dur-scene 2400ms`; easings `--ease-out cubic-bezier(0.2, 0.8, 0.2, 1)` and
`--ease-in-out cubic-bezier(0.65, 0, 0.35, 1)`.

## 3. Reduced motion

- `@media (prefers-reduced-motion: reduce)` sets all four duration tokens to `0ms`
  (`tokens.css`) and `base.css` forces `animation-duration: 0.001ms`,
  `animation-iteration-count: 1`, `transition-duration: 0.001ms` and `scroll-behavior: auto`
  on every element.
- `ModeProvider` reads the media query on hydration: with reduced motion and no stored choice,
  the default mode is `observatory` (the still disk) instead of `event-horizon`; the value
  updates live if the preference changes.
- `HomeStage` renders the static disk whenever `reducedMotion` is true, even if the reader
  chose an immersive mode, so a WebGL scene never runs against the preference.
- The Playwright configuration has a `reduced-motion` project (`contextOptions.reducedMotion:
  "reduce"`) for testing this path.

## 4. Pause when hidden and skippable intro

The build specification requires the Event Horizon scene to pause when the tab is hidden and
to have a skippable intro. `HomeStage` always renders the static disk first so the page never
waits on 3D assets, and loads scenes with `next/dynamic` and `ssr: false`.

## 5. WebGL and fallbacks

`HomeStage` probes for `webgl2`/`webgl` on mount; if unavailable it renders the static disk
with `data-scene="static"`. The Orrery and Branching modes are SVG/CSS by design and do not
need WebGL. See [`states.md`](states.md).

## Not yet implemented

- A motion budget test (frame-time or CPU) for the scenes; the quality manager measures the
  first two seconds at runtime and halves particle count and pixel ratio under 45 fps, but CI
  asserts nothing about frame time.
- User-facing controls to pause or replay a scene; today pausing is automatic (hidden tab,
  out of view, reduced motion) and the intro is skipped by any input.
