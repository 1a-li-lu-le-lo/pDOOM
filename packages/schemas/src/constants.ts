// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Plain constants with no zod dependency. Import from "@pdoom/schemas/constants"
 * in client code so the schema runtime is not bundled just for a list of keys.
 * The main entry re-exports everything here, so "@pdoom/schemas" still works.
 */

/** Outcome ladder codes O0–O8 (build-spec §0.3). */
export const OUTCOME_CODES = ["O0", "O1", "O2", "O3", "O4", "O5", "O6", "O7", "O8"] as const;
export type OutcomeCode = (typeof OUTCOME_CODES)[number];

/** Outcome slugs, one per code. */
export const OUTCOME_SLUGS = [
  "beneficial_or_manageable",
  "serious_reversible_harm",
  "systemic_authoritarian_or_oligopolistic_control",
  "permanent_severe_disempowerment",
  "civilizational_collapse",
  "near_extinction",
  "human_extinction",
  "biospheric_catastrophe",
  "other_irreversible_loss",
] as const;
export type OutcomeSlugValue = (typeof OUTCOME_SLUGS)[number];

/** Horizons in display order (build-spec §0.4). */
export const HORIZON_KEYS = ["1y", "3y", "5y", "10y", "25y", "2100", "eventual"] as const;
export type HorizonKey = (typeof HORIZON_KEYS)[number];

/** The brand, written exactly like this everywhere it is displayed. */
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

export const USER_SCENARIO_SLIDER_KEYS = [
  "capability_timeline",
  "autonomy_growth",
  "access_level",
  "safety_progress",
  "governance_strength",
  "model_security",
  "open_weight_diffusion",
  "international_coordination",
  "incident_frequency",
  "resilience",
] as const;

export const USER_SCENARIO_MAX_SAMPLES = 50000;

export const USER_SCENARIO_LABEL = "user_scenario" as const;

export const USER_SCENARIO_OUTCOME_KEYS = [
  "O3",
  "O4",
  "O5",
  "O6",
  "O7",
  "O8",
  "P_DOOM",
  "P_COLLAPSE",
] as const;

export interface OutcomeDefinition {
  code: OutcomeCode;
  slug: OutcomeSlugValue;
  label: string;
  description: string;
  included_in: {
    pdoom: boolean;
    extinction: boolean;
    disempowerment: boolean;
    collapse: boolean;
    biosphere: boolean;
  };
}

const NOT_INCLUDED = {
  pdoom: false,
  extinction: false,
  disempowerment: false,
  collapse: false,
  biosphere: false,
} as const;

/**
 * Outcome ladder O0–O8. Descriptions are category-level and contain no
 * operational detail. O0–O2 are tracked but sit outside the combined p(DOOM) set.
 */
export const OUTCOMES: Record<OutcomeCode, OutcomeDefinition> = {
  O0: {
    code: "O0",
    slug: "beneficial_or_manageable",
    label: "Beneficial or manageable",
    description:
      "Advanced AI is developed and deployed with harms that remain bounded, correctable and within the capacity of existing institutions to manage.",
    included_in: { ...NOT_INCLUDED },
  },
  O1: {
    code: "O1",
    slug: "serious_reversible_harm",
    label: "Serious but reversible harm",
    description:
      "Large-scale harm occurs — economic, social or physical — but societies retain the ability to recover and to correct course within a generation.",
    included_in: { ...NOT_INCLUDED },
  },
  O2: {
    code: "O2",
    slug: "systemic_authoritarian_or_oligopolistic_control",
    label: "Systemic authoritarian or oligopolistic control",
    description:
      "AI-enabled concentration of power by a state or a small set of organisations that substantially reduces political and economic freedom but is not judged permanent.",
    included_in: { ...NOT_INCLUDED },
  },
  O3: {
    code: "O3",
    slug: "permanent_severe_disempowerment",
    label: "Permanent severe disempowerment",
    description:
      "Humanity loses, in a way judged practically irreversible, the ability to direct its own future — whether to AI systems or to a narrow group controlling them.",
    included_in: { ...NOT_INCLUDED, pdoom: true, disempowerment: true },
  },
  O4: {
    code: "O4",
    slug: "civilizational_collapse",
    label: "Civilizational collapse",
    description:
      "A breakdown of global-scale institutions, infrastructure and population that is not recovered from within centuries, with humanity surviving in reduced form.",
    included_in: { ...NOT_INCLUDED, pdoom: true, collapse: true },
  },
  O5: {
    code: "O5",
    slug: "near_extinction",
    label: "Near extinction",
    description:
      "Human population falls to a small fraction of its current level with recovery uncertain; the species survives.",
    included_in: { ...NOT_INCLUDED, pdoom: true, collapse: true },
  },
  O6: {
    code: "O6",
    slug: "human_extinction",
    label: "Human extinction",
    description: "No living humans remain.",
    included_in: { ...NOT_INCLUDED, pdoom: true, extinction: true },
  },
  O7: {
    code: "O7",
    slug: "biospheric_catastrophe",
    label: "Biospheric catastrophe",
    description:
      "Irreversible destruction of much of Earth's biosphere attributable to AI-driven activity, whether or not humans survive.",
    included_in: { ...NOT_INCLUDED, pdoom: true, biosphere: true },
  },
  O8: {
    code: "O8",
    slug: "other_irreversible_loss",
    label: "Other irreversible loss",
    description:
      "An unrecoverable loss of value not captured by O3–O7, such as permanent lock-in of a substantially worse trajectory for humanity.",
    included_in: { ...NOT_INCLUDED, pdoom: true },
  },
};

/** Derived outcome sets (build-spec §3.3). Never combine outcomes silently (§0.4). */
export const DERIVED_OUTCOME_SETS = {
  P_DOOM: ["O3", "O4", "O5", "O6", "O7", "O8"],
  P_EXTINCTION: ["O6"],
  P_DISEMPOWERMENT: ["O3"],
  P_COLLAPSE: ["O4", "O5"],
  P_BIOSPHERE: ["O7"],
} as const satisfies Record<string, readonly OutcomeCode[]>;
export type DerivedOutcomeSetKey = keyof typeof DERIVED_OUTCOME_SETS;

export interface HorizonDefinition {
  key: HorizonKey;
  label: string;
  /** Years after `forecast_origin_date`; null for calendar-year and open-ended horizons. */
  years: number | null;
  kind: "duration" | "calendar_year" | "open_ended";
  note: string;
}

export const HORIZONS: readonly HorizonDefinition[] = [
  { key: "1y", label: "Within 1 year", years: 1, kind: "duration", note: "" },
  { key: "3y", label: "Within 3 years", years: 3, kind: "duration", note: "" },
  { key: "5y", label: "Within 5 years", years: 5, kind: "duration", note: "" },
  { key: "10y", label: "Within 10 years", years: 10, kind: "duration", note: "" },
  { key: "25y", label: "Within 25 years", years: 25, kind: "duration", note: "" },
  {
    key: "2100",
    label: "By 2100",
    years: null,
    kind: "calendar_year",
    note: "Calendar-year endpoint; the elapsed span depends on forecast_origin_date, so it is not interchangeable with a fixed duration.",
  },
  {
    key: "eventual",
    label: "Eventually (no fixed date)",
    years: null,
    kind: "open_ended",
    note: "Open-ended horizon with no endpoint. 'Eventual' probabilities are not comparable with dated horizons, cannot be converted to a rate, and must never be displayed as a countdown or as an implied date.",
  },
];
