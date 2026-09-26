<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Colour

The palette lives in `packages/design-system/tokens.css` (CSS custom properties on `:root`)
and is mirrored in `packages/design-system/tokens.json`. Dark is the default. The comment at
the top of `tokens.css` states the intent: "deep black is unknown space, not danger. Every
semantic colour has a text or shape equivalent in the UI; colour is never the only channel."

## 1. Semantic colours (dark default)

| Token | Value | Meaning | Where it is used |
| --- | --- | --- | --- |
| `--c-evidence` | `#f5f5f7` | Documented evidence; the neutral data colour | `EstimateCard` for external aggregates, `IndexGauge` for evidence pressure, chart defaults, forecast dots for most populations |
| `--c-safeguard` | `#5fd6ea` | Safeguards and interventions | `IndexGauge` control strength, safeguard arcs in `StaticDisk`, `Waterfall` bars that strengthen control, superforecaster dots on `/forecasts` |
| `--c-uncertainty` | `#e8b45a` | Uncertainty | Uncertainty gauge, "Your scenario" value and strip in the Scenario Lab, `badge-uncertainty`, low-recoverability nodes on `/futures`, prediction-market dots |
| `--c-risk` | `#ec6a66` | Elevated risk pressure | Capability, incident and agentic-infrastructure gauges, `Waterfall` bars that raise pressure, `Timeline` measurement points, irreversible nodes on `/futures`, `IconArray` filled cells |
| `--c-disagreement` | `#d879e0` | Model or forecaster disagreement | Research-mode estimate cards and curve, `badge-disagreement` |
| `--c-resilience` | `#62cf95` | Resilience and beneficial interventions | Recoverable (high or moderate) nodes on `/futures`, `badge-resilience` |
| `--c-insufficient` | `#8b90a0` | Insufficient evidence, not computed, withheld | The withheld official value (`.meter-value.withheld`), the attention gauge, unknown recoverability, tier 4–5 bars on `/evidence` |

Each semantic colour has a `-soft` variant at 16% alpha (for example
`--c-safeguard-soft: rgba(95, 214, 234, 0.16)`) for fills behind text.

Surfaces and text (dark): `--c-bg #050508`, `--c-bg-elevated #0c0d13`, `--c-surface #12141c`,
`--c-surface-2 #1a1d27`, `--c-border #262a38`, `--c-border-strong #3a4053`; `--c-text #f2f3f7`,
`--c-text-2 #b6bac8`, `--c-text-3 #7d8394`, `--c-text-inverse #05060a`; `--c-focus #8fd8ff`,
`--c-link #9cc9ff`.

## 2. Light theme

`:root[data-theme="light"]` redefines every token with darker semantic colours so that text
and marks keep contrast on light surfaces:

| Token | Light value |
| --- | --- |
| `--c-bg` / `--c-bg-elevated` / `--c-surface` / `--c-surface-2` | `#f7f7fa` / `#ffffff` / `#ffffff` / `#eef0f5` |
| `--c-border` / `--c-border-strong` | `#d8dbe5` / `#b7bccb` |
| `--c-text` / `--c-text-2` / `--c-text-3` / `--c-text-inverse` | `#0b0c12` / `#3c4152` / `#6b7184` / `#ffffff` |
| `--c-evidence` | `#0b0c12` |
| `--c-safeguard` | `#0b7f94` |
| `--c-uncertainty` | `#8a5a00` |
| `--c-risk` | `#b3261e` |
| `--c-disagreement` | `#8e2a99` |
| `--c-resilience` | `#1c7a49` |
| `--c-insufficient` | `#5f6577` |
| `--c-focus` / `--c-link` | `#0a58ca` / `#0a58ca` |

Soft variants drop to 12% alpha in the light theme. `color-scheme` is set to `dark` or `light`
accordingly so form controls follow.

## 3. Colour-vision-deficiency palette

`:root[data-palette="cvd"]` replaces the red/green pair with an Okabe–Ito-inspired set:
`--c-risk #d55e00` (orange-red), `--c-resilience #009e73` (blue-green), `--c-safeguard #56b4e9`,
`--c-uncertainty #e69f00`, `--c-disagreement #cc79a7`. Shapes and labels carry the meaning
regardless of palette (section 5).

## 4. Contrast preference

`@media (prefers-contrast: more)` lifts `--c-text-2` to `#d7dae4`, `--c-text-3` to `#aeb3c2`,
`--c-border` to `#4a5066` and `--c-border-strong` to `#6b7189` in the dark theme.

## 5. Colour is never the only channel

- **Badges have shapes.** `.badge::before` draws a marker before the label: a filled circle by
  default; `badge-risk` a square; `badge-uncertainty` a rotated square (diamond);
  `badge-disagreement` a leaf shape (`border-radius: 50% 0 50% 0`); `badge-insufficient` a
  hollow marker (transparent fill with the border). The label text always names the state
  ("high uncertainty", "extreme disagreement", "Official value withheld").
- **Every chart has a data table.** `QuantileStrip`, `LineWithBand`, `DistributionDots`,
  `NodeGraph` and `Timeline` render a `<details><summary>Data table</summary>` with the values;
  `BarList` and `Waterfall` print the number beside every bar; `IconArray` states the filled
  count in its accessible name and caption.
- **Status words accompany status colours.** `EstimateCard` shows the status label as text;
  `IndexGauge` prints "index, not a probability · coverage N%" under every value.
- **Recoverability on `/futures`** is written in the group heading ("Irreversible", "Low
  recoverability", …) and in each card, not only as node colour.
- **Bars carry `aria-hidden`** where the same value is printed as text next to them.

## 6. Rules for using colour

1. Use semantic tokens only; never a hex literal in a component. The single exception is the
   `StaticDisk` SVG and the share card, whose gradients are fixed art (`#ffd9a0`, `#d88a5a`,
   `#3b2a5e`, `#ffb27a`; `#07070b`, `#141225`, `#2a1a12`) and do not encode data.
2. Red means pressure on a gauge or a raising contribution; it is never a warning banner.
3. The withheld official value is grey (`--c-insufficient`), not red or amber: absence of a
   number is not an alarm.
4. Research-mode values are magenta (`--c-disagreement`) so they can never be confused with
   external aggregates (white) or the official object (grey).
5. Print switches to black on white (`base.css` `@media print`).

## Not yet implemented

- An in-page theme or palette switcher. The `data-theme` and `data-palette` attributes are
  honoured when set on `<html>` but nothing in `apps/web` sets them yet.
- Automated contrast verification (the axe dependency is present but no end-to-end specs exist;
  see [`../accessibility/accessibility.md`](../accessibility/accessibility.md)).
