<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Psychological safety

p(DOOM) is about a serious subject and is read by people who may already be anxious about it.
The observatory's job is to make the evidence legible, not to make anyone afraid. This page
sets the rules (build-spec rules 0.10 and 0.14) and shows where the code enforces them.

## 1. Principles

1. **Calm language.** Plain declarative sentences; no exclamation marks; no imperatives about
   fear; no rhetoric of urgency.
2. **No inevitability.** Nothing on the site says or implies that a catastrophe will happen.
   Every estimate is conditional, carries an interval and a disagreement label, and the official
   value is withheld.
3. **No countdowns.** No dates for catastrophe, no counting of years, no clocks, no "before it is
   too late". Horizons are stated as periods from the forecast origin date; "2100" is a calendar
   endpoint and "eventual" has none.
4. **Beneficial outcomes and safeguards are visible.** The outcome ladder starts at O0
   "Beneficial or manageable"; the Futures page groups pathways by recoverability; the Safeguards
   and Act pages and the "Help reduce risk" button are part of the meter, not an afterthought.
5. **No alarm UI.** No pulsing, flashing, sirens, red banners, alert sounds or graphic imagery.
   Red in the palette means "elevated risk pressure" on a gauge, never danger signalling.
6. **Fear is not a plan.** The Act page opens with that sentence and lists actions with a stated
   effort level; none requires believing any particular number.

## 2. Language rules with examples

| Do not write | Write instead |
| --- | --- |
| A countdown of how long humanity has | "Estimates exist for the 10-year horizon from the forecast origin date; no date is implied" |
| "p(DOOM) is 5%" | "The official value is withheld; the release states a research-mode median of … for O3–O8 at … (not the official estimate)" |
| "Risk is rising fast" | "The Capability Pressure Index reads N of 100 (an index, not a probability); the Evidence Pressure Index is at its baseline of 50 in the first release" |
| "AI will take over" | "Scenario S1, Deliberate misaligned action, is a category-level pathway with these prerequisites and early indicators; no pathway probability is assigned" |
| "Warning", "Alert", "Critical" as headlines | The editorial level in its own words ("Elevated") beside the rule that produced it |

The uncertainty and disagreement labels (`low`, `moderate`, `high`, `extreme`) describe the
evidence, not the reader's situation, and are always shown as badges with the word "uncertainty"
or "disagreement" attached.

## 3. Help-line neutrality: no medical advice

The observatory is not a health service. It does not diagnose, assess or advise on anyone's
mental state, and it does not give medical advice. The release model card lists "individual
mental-health assessment" and "emergency-alert triggering" among prohibited uses. The site
does not curate or rank crisis services, because doing so would present editorial judgement as
medical guidance; a reader who is distressed is best served by their local health services,
and the observatory's copy should never stand between a reader and that choice. Where copy
touches on how people feel about the subject, it stays factual about the evidence and
neutral about the reader.

## 4. How the UI enforces it

| Mechanism | Where |
| --- | --- |
| Guard test fails the build on percentage literals and on three fixed panic phrasings (a countdown of how long humanity has, certainty of doom, "the machines are coming") | `apps/web/test/no-hardcoded-numbers.test.ts` |
| Every probability rendered through one card with status, outcome set, horizon, interval, disagreement, uncertainty and data cutoff | `apps/web/components/meter/EstimateCard.tsx` |
| The official value is displayed as "Insufficiently calibrated" in the muted `--c-insufficient` colour, smaller than a number would be | `MeterPanel`, `.meter-value.withheld` |
| Footer on every page: "No official probability is published in this release line. Indexes are not probabilities. Nothing here predicts a date." | `apps/web/components/shell/Footer.tsx` |
| Scene label "Conceptual risk visualization · not a simulation of AI risk" above the hero; the SVG disk's accessible name says the same | `apps/web/app/page.tsx`, `StaticDisk.tsx` |
| Mode switcher with a plain-text link in the first viewport; `/text` contains everything | `ModeSwitcher.tsx`, `apps/web/app/text/page.tsx` |
| Reduced-motion preference selects the still Observatory mode by default; motion tokens are zeroed | `ModeProvider.tsx`, `packages/design-system/tokens.css` |
| No CSS keyframe animations anywhere; only short transitions on state change | `apps/web/app/globals.css`, `packages/design-system/base.css` |
| Beneficial and protective content prominent: `/futures` groups by recoverability, `/safeguards` and `/act` are primary navigation, `MeterPanel` ends with "Help reduce risk" | `Nav.tsx`, `MeterPanel.tsx` |
| The share card carries the official status and the research-mode interval, never a bare number | `apps/web/app/opengraph-image.tsx` |
| Content-safety policy: category-level pathways, non-graphic incident summaries | [`../method/content-safety.md`](../method/content-safety.md) |

## 5. Review checklist for copy

Before approving a release or a page, the editorial reviewer confirms:

- no sentence asserts that a catastrophe will occur or when;
- no number appears without its status, outcome set, horizon and interval;
- no index is described with probability language;
- the beneficial outcome, safeguards and actions are reachable within one click of the page;
- no colour, motion or sound is used to signal alarm;
- no medical or crisis advice is given;
- the plain-text route contains the same substance.

## Not yet implemented

- Reader testing of copy for perceived alarm; the rules above are editorial judgement.
- An automated readability or tone check beyond the guard test's fixed phrases.
- Reader testing of the immersive scenes for perceived alarm. The scenes pause when hidden,
  skip their intro on any input and hold still under reduced motion ([`../design/motion-semantics.md`](../design/motion-semantics.md)),
  but no one outside the project has yet rated them.
