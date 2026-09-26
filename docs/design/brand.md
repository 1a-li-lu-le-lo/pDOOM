<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Brand

## 1. The name

The brand is written exactly **p(DOOM)** in every displayed or prose context: lower-case p,
parentheses, upper-case DOOM. It is never PDUM, pDOOM, PDOOM, P(doom) or p(doom) in prose or
UI. The web app renders the name only through the `BRAND` constant exported by
`@pdoom/schemas` (`packages/schemas/src/enums.ts`), never as a literal in a component.

Code identifiers follow the build specification: `pdoom` (packages, module paths), `PDoom`
(types such as `PDoomDataSource`), `PDOOM_` (environment variables such as `PDOOM_DATA_DIR`).
Packages are `@pdoom/*`; binaries are `pdoomctl`, `pdoom-api`, `pdoom-ingest`, `pdoom-model`;
the combined outcome-set key is `P_DOOM`; model versions are `pdoom-model/<family>@<semver>`.
Every source file begins with the line `Copyright NU Cybernetics. p(DOOM) — research prototype.`
in the comment syntax of its language.

## 2. The constants, quoted

From `packages/schemas/src/enums.ts`:

```ts
export const BRAND = "p(DOOM)" as const;
export const BRAND_EXPANDED =
  "Probability of Doom, Disempowerment, and Unrecoverable Machine-Caused Catastrophe" as const;
export const PUBLIC_LABEL = "The AI Existential and Civilizational Risk Observatory" as const;
export const AUTHOR = "NU Cybernetics" as const;
export const TAGLINES = [
  "The Future Is Not a Single Number.",
  "Measure the Risk. Expose the Assumptions. Change the Trajectory.",
  "A Living Map of Advanced-AI Risk.",
  "Understand the Odds. Improve the Outcome.",
  "Watch the Frontier Without Losing Sight of Humanity.",
] as const;
```

Where each is used:

| Constant | Used in |
| --- | --- |
| `BRAND` | Header wordmark, `<title>` template (`%s · p(DOOM)`), footer, breadcrumbs root, share card, `/text` heading |
| `BRAND_EXPANDED` | Footer line, `/text` lede |
| `PUBLIC_LABEL` | Header sub-line under the wordmark, default `<title>`, share card, `/text` eyebrow |
| `AUTHOR` | Footer ("Authored by NU Cybernetics. Copyright NU Cybernetics. Research prototype.") |
| `TAGLINES[0]` | Home hero `<h1>`, page description, Open Graph description, share card footer |

`TAGLINES[1..4]` are approved alternates for campaigns and documentation; the hero uses the
first.

## 3. Voice

- **Calm and exact.** Short declarative sentences. The number is never the subject of a
  sentence without its outcome set, horizon and status.
- **Shows its work.** Prefers "the release states", "the model computes", "the source reports"
  to "we believe". Assumptions are named as assumptions; judgments as judgments.
- **Plain words for hard ideas.** "Plausible interval", "not a probability", "conceptual
  visualisation", "research mode", "official value withheld".
- **Respectful of the reader's time and nerves.** No urgency rhetoric, no cliffhangers, no
  countdowns ([`../governance/psychological-safety.md`](../governance/psychological-safety.md)).
- **British-leaning spelling in prose** (visualisation, organisation, licence) as used across
  the docs; enum values and identifiers keep their exact spelling
  (`organizations.json`, `license` field).

## 4. What the brand never does

- Publish or imply an official probability while the release line is
  `insufficiently_calibrated` (ADR-002).
- Present an index as a probability, or a probability without its context.
- Use fear, alarm colours, sirens, clocks, skulls, explosions or graphic imagery.
- Claim a scene is a simulation. Every scene is labelled "Conceptual risk visualization ·
  not a simulation of AI risk".
- Endorse organisations; the directory lists organisations that meet stated criteria.
- Track readers: no analytics, no cookies, no external requests at runtime; sharing means
  sharing a reproducible link to the same signed release, without tracking parameters
  (`ShareLink`).
- Change a number silently ([`../governance/probability-change-policy.md`](../governance/probability-change-policy.md)).
- Substitute a model identifier or an AI attribution for NU Cybernetics' authorship in
  repository files.

## 5. Wordmark and typographic treatment

- Header: `.brand` in the display face (`--font-display`), weight 700, `--fs-xl`,
  letter-spacing −0.02em; `PUBLIC_LABEL` beneath it in the body face at `--fs-xs`,
  `--c-text-3`.
- Hero: `TAGLINES[0]` as the `<h1>` at `--fs-hero` (`clamp(3rem, 9vw, 7rem)`), weight 700,
  letter-spacing −0.035em, max 14ch, with a soft text shadow over the scene.
- The parentheses are part of the name; they are not styled differently and the name is
  never split across lines by hand.

## 6. Share card

`apps/web/app/opengraph-image.tsx` renders the 1200×630 card: `BRAND`, `PUBLIC_LABEL`, the
official status for the default horizon with "O3–O8 combined", the research-mode median and
plausible interval sentence, the tagline, the uncertainty score and the release id. A screenshot
of the card is still an honest statement.

## Not yet implemented

- A logo or icon beyond the typographic wordmark; there is no favicon asset in the repository.
- Self-hosted brand fonts; the tokens name Space Grotesk, IBM Plex Sans and IBM Plex Mono with
  system fallbacks, but no font files or `@font-face` rules ship
  ([`typography.md`](typography.md)).
