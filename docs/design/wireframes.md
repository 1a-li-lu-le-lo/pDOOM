<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Wireframes

Text wireframes for the five routes that define the experience. "Above the fold" means the
first viewport on a desktop; on a phone the hero collapses to a 320px stage and the same order
continues below. The wireframes describe the order of information; every box is rendered from
the current release unless marked static.

Shared chrome on every route: a skip link, the sticky header (wordmark + public label,
primary navigation, compact mode switcher with Event Horizon / Observatory / Plain text), and
the footer (release id, data cutoff, publication date, method/changelog/text/content-safety
links, the fixed sentence that no official probability is published).

## `/` Home (Meter home)

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ p(DOOM)  The AI Existential and…   Meter Futures Evidence … Text  [EH][Obs][Text] │
├──────────────────────────────────────────────────────────────────────────────┤
│  (stage: static disk or scene; ring = interval, glow = evidence pressure,    │
│   dots = sources, cyan arcs = safeguards)                                    │
│                                                                              │
│  CONCEPTUAL RISK VISUALIZATION · NOT A SIMULATION OF AI RISK   (scene label) │
│  The Future Is Not a Single Number.                       (h1 = TAGLINES[0]) │
│  An open-methodology observatory … The official value is withheld until it   │
│  can be calibrated; the assumptions are open for inspection.                 │
│  [Inspect the assumptions]  [Read the plain-text version]                    │
├────────────────────────────────── fold ──────────────────────────────────────┤
│ MeterPanel                                                                   │
│  p(DOOM) · 10 years · O3–O8 combined                               (eyebrow) │
│  Insufficiently calibrated                       (h2, grey, withheld style)  │
│  The observatory publishes indexes, an external forecast aggregate and a     │
│  research-mode model; it does not publish an official probability yet.       │
│  model … · data through … · last reviewed … · editorial level: Elevated      │
│  ┌ External forecast aggregate ┐ ┌ Research mode ┐   (EstimateCards, compact)│
│  │ badge row · outcome · value │ │ … interval … │                            │
│  │ quantile strip · meta       │ │              │                            │
│  └─────────────────────────────┘ └──────────────┘                            │
│  "No external forecast group asks about the 10 years horizon…" (if none)     │
│  Decomposition (research mode): [O3] [O4+O5] [O6] [O7]      (compact cards)  │
│  Indexes (0–100; not probabilities): [CPI][CSI][IPI][AIR][EPI][Uncertainty]  │
│  Top upward drivers (index points)   │ Top downward drivers (control strength)│
│  What changed: • First release … • Snapshot …; 70 sources, 15 forecasts, …   │
│  [Inspect the assumptions][Compare forecasts][Change horizon][Explore        │
│   scenarios][Read plain text][Download data][Help reduce risk]               │
└──────────────────────────────────────────────────────────────────────────────┘
```

Order of information: disclaimer, tagline, what the site is, then the withheld headline,
then the two labelled quantities, then the decomposition, then the indexes, drivers, changes
and actions. The tagline is the only thing on the page that is not read from the release.

## `/meter`

```
┌ Breadcrumb: p(DOOM)                                        [Plain text] ┐
│ The meter                                                          (h1) │
│ Everything on this page comes from one signed release. Change the        │
│ horizon to see how each object moves; open the assumptions to see why.   │
│ Horizon: [1 year][3 years][5 years][10 years*][25 years][2100][Eventual]  │
├──────────────────────────────── fold ───────────────────────────────────┤
│ MeterPanel (as on the home page, for the selected horizon)               │
│ ── Why these numbers ─────────────────────────────────────── (#why) ──  │
│ 1. What an official value would need   (assumptions + conditioning of   │
│    the official object; links to calibration.md and model.md)           │
│ 2. External forecasts, aggregated inside compatibility groups           │
│    BarList: preferred aggregate per group → /forecasts#<group>           │
│ 3. The research-mode model across horizons                              │
│    LineWithBand (median + p05–p95 band) · full EstimateCard              │
│ Which single change matters most                                        │
│    Table of the top 8 sensitivity runs (kind, target, change, baseline, │
│    result, delta)                                                       │
│ Uncertainty score: N out of 100 · BarList of weighted components        │
│ Editorial risk level: [Elevated]  Rule: cpi >= 55 and csi < 55           │
│ Provenance: release · snapshot · model versions · cutoff · reviewers ·  │
│    reproduce command                                                    │
│ [Copy link to this view][Download release JSON][Try your own assumptions]│
└─────────────────────────────────────────────────────────────────────────┘
```

## `/futures`

```
┌ Breadcrumb                                                  [Plain text] ┐
│ Branching futures                                                   (h1) │
│ Eighteen category-level pathways, described by their prerequisites and    │
│ what would be visible early, never by operational detail. Arrows show     │
│ which pathways enable, amplify or substitute for others.                  │
│ NodeGraph "Scenario relations": 18 nodes on a circle coloured by          │
│ recoverability, 30 arcs weighted by confidence; caption states it is a   │
│ map of dependencies, not a prediction; relations table in <details>       │
├──────────────────────────────── fold ────────────────────────────────────┤
│ [Irreversible] 1 pathway         (h2 with shape-coded badge and count)     │
│   ┌ S17 · Benign guardianship becomes permanent ┐  (card: uncertainty      │
│   │ badges · description · Outcomes: … · N safeguards target this pathway │
│   │ · N early indicators)                                                 │
│ [Low recoverability] 8 pathways   grid of cards                             │
│ [Moderate recoverability] 7 pathways                                       │
│ [High recoverability] 1 pathway                                            │
│ [Recoverability unknown] 1 pathway                                         │
│ Method: causal-graph.md · Content rules: content-safety.md                 │
└───────────────────────────────────────────────────────────────────────────┘
```

Each card links to `/futures/[scenarioId]`, which shows the scenario's prerequisites, early
indicators, counterindicators, capability thresholds, exposure, control failures, human and AI
contributions, incoming and outgoing relations, targeting safeguards and sources.

## `/lab`

```
┌ Breadcrumb                                                  [Plain text] ┐
│ Scenario Lab                                                        (h1) │
│ Ten dials, one honest model. … Nothing you do here changes a published    │
│ number, and the result is always labelled as yours.                       │
├──────────────────────────────── fold ────────────────────────────────────┤
│ ┌ Controls (sticky on desktop) ┐ ┌ Result (aria-live) ────────────────┐   │
│ │ Horizon [1y][3y]…[10y*]…     │ │ [Your scenario][10 years][baseline │   │
│ │ Capability timeline  A       │ │  assumptions][pdoom-model/…]       │   │
│ │  much slower ──●── much faster│ │ Under your selected assumptions,   │   │
│ │ Autonomy growth      C ↓     │ │ not the p(DOOM) official model,    │   │
│ │ System access        C ↓ E ↑ │ │ the median estimate is              │   │
│ │ Safety research      F ↓     │ │        (large amber value)          │   │
│ │ Governance strength  F ↓     │ │ plausible interval … · p(DOOM),     │   │
│ │ Model-weight security E ↓    │ │ O3–O8 combined, 10 years · about    │   │
│ │ Open-weight diffusion E ↑    │ │ the same median as the published    │   │
│ │ International coord. F ↓     │ │ research-mode run                   │   │
│ │ Incident frequency   E ↑     │ │ QuantileStrip (compact)             │   │
│ │ Civilizational resil. O ↓    │ │ Published research-mode run for     │   │
│ │ Samples [5,000 ▾] [Reset]    │ │ this horizon (release …): … Your    │   │
│ └──────────────────────────────┘ │ scenario does not change it.        │   │
│                                  │ flags (if any)                      │   │
│                                  │ Decomposition table · Factor table  │   │
│                                  │ [Copy link to this scenario]        │   │
│                                  │ [Back to the published meter]       │   │
│                                  │ disclaimer (.cite)                  │   │
│                                  └─────────────────────────────────────┘   │
│ Model: causal-model.md · Dial mapping: model.md · same code as Go, golden   │
└───────────────────────────────────────────────────────────────────────────┘
```

The URL is the state (`/lab?h=10y&ct=1&sp=-1…`); sharing the link reproduces the result.

## `/text`

```
┌ THE AI EXISTENTIAL AND CIVILIZATIONAL RISK OBSERVATORY · PLAIN-TEXT MODE ┐
│ p(DOOM) in plain text                                               (h1) │
│ Probability of Doom, Disempowerment, and Unrecoverable Machine-Caused     │
│ Catastrophe. This page contains every substantive fact on the site        │
│ without graphics, scripts or hidden interactions. It is server-rendered,  │
│ printable and linkable. Release …, data through …, published ….           │
│ Switch to the immersive version.                                          │
│ Sections (two-column list, 20 anchors):                                   │
│  1 Current p(DOOM) estimate   2 Definition   3 Selected horizon            │
│  4 Uncertainty   5 Outcome decomposition   6 Top upward drivers            │
│  7 Top downward drivers   8 What changed   9 External forecasts            │
│ 10 Capability indicators  11 Agentic infrastructure  12 Incidents         │
│ 13 Scenario map  14 Safeguards  15 What users can do  16 Methodology       │
│ 17 Source ledger  18 Model history  19 Limitations  20 Governance          │
├──────────────────────────────── fold ────────────────────────────────────┤
│ 1. Current p(DOOM) estimate  (definition list: official status, horizon,  │
│    definition, model versions, data through, last reviewed by, editorial  │
│    level; then the two labelled quantities as list items with citations)  │
│ 2. Definition (outcome taxonomy table O0–O8)                              │
│ 3. Selected horizon (research-mode table by horizon)                      │
│ … each section a heading with a `#` anchor link, tables with captions,    │
│    citations as [src-…] links to the source ledger rows …                 │
│ 17. Source ledger (70 rows: id, source, tier, verification, model use,    │
│     conflicts)                                                            │
│ 18. Model history (releases table; reproduction command; manifest hash;   │
│     approvals)                                                            │
│ 19. Limitations (known limitations + snapshot notes)                      │
│ 20. Governance (paragraph) · Downloads: release.json, estimates.csv, …    │
└───────────────────────────────────────────────────────────────────────────┘
```

`/text` needs no JavaScript: it is server-rendered HTML with ordinary anchors, and every
route's header links to the matching section.

## Not yet implemented

- Wireframes for in-canvas overlays; the scenes intentionally render no text today.
- A mobile-specific arrangement of the Scenario Lab controls beyond stacking.
