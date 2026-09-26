// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * @pdoom/model-core — the TypeScript mirror of the Go experimental causal
 * model (internal/model/{rng,dist,stats,causal,rounding}.go). It exists so the
 * Scenario Lab can run in the browser and the MCP server can evaluate user
 * scenarios without a Go binary. The Go implementation is authoritative; the
 * golden fixture internal/model/testdata/causal-golden.json pins agreement.
 *
 * Nothing here reads the clock or uses unseeded randomness, and nothing here
 * ever produces an "official" estimate: every result is labelled user_scenario.
 */
import type {
  CausalHorizonSpec,
  ExperimentalCausalSpec,
  SummaryStats,
  TriQuantiles,
  UserScenarioParams,
  UserScenarioResult,
} from "@pdoom/schemas";
import { USER_SCENARIO_SLIDER_KEYS } from "@pdoom/schemas";

export * from "./rounding";

// ---------------------------------------------------------------------------
// RNG — bit-identical to internal/model/rng.go
// ---------------------------------------------------------------------------

/** mulberry32: the reference 32-bit generator shared with Go. */
export function mulberry32(seed: number): () => number {
  let a = seed | 0;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** Seeded generator with the same Box–Muller normal as Go (two uniforms per normal). */
export class Rng {
  private readonly next: () => number;
  constructor(seed: number) {
    this.next = mulberry32(seed);
  }
  uniform(): number {
    return this.next();
  }
  normal(): number {
    let u1 = this.next();
    const u2 = this.next();
    if (u1 < 1e-12) u1 = 1e-12;
    return Math.sqrt(-2 * Math.log(u1)) * Math.cos(2 * Math.PI * u2);
  }
}

// ---------------------------------------------------------------------------
// Distributions
// ---------------------------------------------------------------------------

const Z95 = 1.6448536269514722;

export function logit(p: number): number {
  const eps = 1e-9;
  const q = Math.min(Math.max(p, eps), 1 - eps);
  return Math.log(q / (1 - q));
}

export function sigmoid(x: number): number {
  return 1 / (1 + Math.exp(-x));
}

export interface LogitNormal {
  mu: number;
  sigma: number;
}

/** mu = logit(p50), sigma = (logit(p95) − logit(p05)) / (2 · 1.6448536269514722). */
export function logitNormalFromTri(t: TriQuantiles): LogitNormal {
  return { mu: logit(t.p50), sigma: (logit(t.p95) - logit(t.p05)) / (2 * Z95) };
}

export function shifted(d: LogitNormal, delta: number): LogitNormal {
  return { mu: d.mu + delta, sigma: d.sigma };
}

export function at(d: LogitNormal, z: number): number {
  return sigmoid(d.mu + d.sigma * z);
}

export function tri(d: LogitNormal): TriQuantiles {
  return { p05: at(d, -Z95), p50: at(d, 0), p95: at(d, Z95) };
}

// ---------------------------------------------------------------------------
// Statistics
// ---------------------------------------------------------------------------

/** Nearest-rank quantile of a sorted array. */
export function quantileSorted(sorted: number[], p: number): number {
  if (sorted.length === 0) return 0;
  let k = Math.ceil(p * sorted.length) - 1;
  if (k < 0) k = 0;
  if (k >= sorted.length) k = sorted.length - 1;
  return sorted[k] as number;
}

export function round6(x: number): number {
  return Math.round(x * 1e6) / 1e6;
}

export function summarize(samples: number[]): SummaryStats {
  const s = [...samples].sort((a, b) => a - b);
  const mean = s.length ? s.reduce((acc, v) => acc + v, 0) / s.length : 0;
  return {
    p05: round6(quantileSorted(s, 0.05)),
    p25: round6(quantileSorted(s, 0.25)),
    p50: round6(quantileSorted(s, 0.5)),
    p75: round6(quantileSorted(s, 0.75)),
    p95: round6(quantileSorted(s, 0.95)),
    mean: round6(mean),
  };
}

// ---------------------------------------------------------------------------
// Experimental causal model
// ---------------------------------------------------------------------------

export const OUTCOME_ORDER = ["O3", "O4", "O5", "O6", "O7", "O8"] as const;
export type Shifts = Partial<Record<"A" | "C" | "E" | "F" | (typeof OUTCOME_ORDER)[number], number>>;

export interface SimResult {
  outcomes: Record<string, SummaryStats>;
  factors: Record<"A" | "C" | "E" | "F", TriQuantiles>;
}

/**
 * Draw order per sample (each normal consumes two uniforms):
 * Z (common factor), εA, εC, εE, εF, εO3 … εO8 — identical to Go's Simulate.
 */
export function simulate(
  h: CausalHorizonSpec,
  lambda: number,
  seed: number,
  samples: number,
  shifts: Shifts = {},
): SimResult {
  const n = samples > 0 ? samples : 1;
  const factorKeys = ["A", "C", "E", "F"] as const;
  const factors = {
    A: shifted(logitNormalFromTri(h.A), shifts.A ?? 0),
    C: shifted(logitNormalFromTri(h.C), shifts.C ?? 0),
    E: shifted(logitNormalFromTri(h.E), shifts.E ?? 0),
    F: shifted(logitNormalFromTri(h.F), shifts.F ?? 0),
  };
  const shares: Partial<Record<(typeof OUTCOME_ORDER)[number], LogitNormal>> = {};
  for (const o of OUTCOME_ORDER) {
    const t = h.O[o];
    if (t) shares[o] = shifted(logitNormalFromTri(t), shifts[o] ?? 0);
  }
  const lam = Math.min(Math.max(lambda, 0), 1);
  const restSq = 1 - lam * lam;
  const rest = restSq <= 0 ? 0 : Math.sqrt(restSq);
  const rng = new Rng(seed);
  const out: Record<string, number[]> = { P_DOOM: new Array<number>(n), P_COLLAPSE: new Array<number>(n) };
  for (const o of OUTCOME_ORDER) out[o] = new Array<number>(n);
  for (let i = 0; i < n; i++) {
    const z = rng.normal();
    let reach = 1;
    for (const k of factorKeys) {
      const eps = rng.normal();
      reach *= at(factors[k], lam * z + rest * eps);
    }
    const s: Partial<Record<(typeof OUTCOME_ORDER)[number], number>> = {};
    let sum = 0;
    for (const o of OUTCOME_ORDER) {
      const eps = rng.normal();
      const d = shares[o];
      if (!d) continue;
      const v = at(d, lam * z + rest * eps);
      s[o] = v;
      sum += v;
    }
    if (sum > 1) for (const o of OUTCOME_ORDER) if (s[o] !== undefined) s[o] = (s[o] as number) / sum;
    let doom = 0;
    for (const o of OUTCOME_ORDER) {
      const p = reach * (s[o] ?? 0);
      (out[o] as number[])[i] = p;
      doom += p;
    }
    (out.P_DOOM as number[])[i] = doom;
    (out.P_COLLAPSE as number[])[i] = (out.O4 as number[])[i]! + (out.O5 as number[])[i]!;
  }
  const outcomes: Record<string, SummaryStats> = {};
  for (const k of Object.keys(out).sort()) outcomes[k] = summarize(out[k] as number[]);
  const r6 = (t: TriQuantiles): TriQuantiles => ({ p05: round6(t.p05), p50: round6(t.p50), p95: round6(t.p95) });
  return {
    outcomes,
    factors: { A: r6(tri(factors.A)), C: r6(tri(factors.C)), E: r6(tri(factors.E)), F: r6(tri(factors.F)) },
  };
}

/** Slider → logit-scale shifts; the documented contract of docs/method/model.md. */
export function sliderShifts(p: UserScenarioParams): Shifts {
  const s: Record<string, number> = {};
  const add = (k: string, v: number) => {
    s[k] = (s[k] ?? 0) + v;
  };
  add("A", 0.5 * p.capability_timeline);
  add("C", 0.4 * p.autonomy_growth);
  add("C", 0.3 * p.access_level);
  add("E", 0.3 * p.access_level);
  add("F", -0.5 * p.safety_progress);
  add("F", -0.4 * p.governance_strength);
  add("E", -0.3 * p.model_security);
  add("E", 0.3 * p.open_weight_diffusion);
  add("F", -0.3 * p.international_coordination);
  add("E", 0.2 * p.incident_frequency);
  add("O4", -0.3 * p.resilience);
  add("O5", -0.3 * p.resilience);
  add("O6", -0.3 * p.resilience);
  return s as Shifts;
}

/** Implausible or tense combinations are labelled, never refused. */
export function scenarioFlags(p: UserScenarioParams): string[] {
  const flags: string[] = [];
  if (p.safety_progress >= 2 && p.governance_strength <= -2 && p.international_coordination <= -2)
    flags.push("implausible: strong safety progress alongside collapsing governance and coordination");
  if (p.capability_timeline <= -2 && p.autonomy_growth >= 2)
    flags.push("implausible: much slower capability growth alongside much faster autonomy growth");
  if (p.open_weight_diffusion >= 2 && p.model_security >= 2)
    flags.push("tension: maximal open-weight diffusion alongside maximal model-weight security");
  if (p.access_level <= -2 && p.incident_frequency >= 2)
    flags.push("tension: minimal system access alongside a much higher incident rate");
  return flags;
}

export const USER_SCENARIO_DISCLAIMER =
  "Under your selected assumptions—not the p(DOOM) official model—the median estimate is shown below. This is a user-generated scenario computed by the experimental research-mode model; it is not a published estimate and does not change any official value.";

export function evaluateUserScenario(
  spec: ExperimentalCausalSpec,
  params: UserScenarioParams,
  samples?: number,
  seed?: number,
): UserScenarioResult {
  for (const key of USER_SCENARIO_SLIDER_KEYS) {
    const v = params[key];
    if (!Number.isInteger(v) || v < -2 || v > 2) throw new Error(`slider ${key} out of range -2..2: ${v}`);
  }
  const h = spec.horizons[params.horizon];
  if (!h) throw new Error(`horizon ${params.horizon} is not covered by the experimental model`);
  let n = samples ?? params.samples ?? 0;
  if (n <= 0 || n > spec.samples) n = spec.samples;
  let s = seed ?? params.seed ?? 0;
  if (s === 0) s = spec.seed;
  const sim = simulate(h, spec.common_factor_loading, s, n, sliderShifts(params));
  const { samples: _s, seed: _seed, ...echo } = params;
  return {
    label: "user_scenario",
    disclaimer: USER_SCENARIO_DISCLAIMER,
    horizon: params.horizon,
    outcome_estimates: sim.outcomes as UserScenarioResult["outcome_estimates"],
    factor_summary: sim.factors,
    flags: scenarioFlags(params),
    params_echo: echo,
    spec_version: spec.version,
  };
}

/** The mandated sentence for user scenarios, with the rounded median. */
export function describeUserScenario(result: UserScenarioResult, step = 5): string {
  const doom = result.outcome_estimates.P_DOOM;
  const central = doom ? roundForDisplayImpl(doom.p50, step) : "unavailable";
  const interval = doom ? `${roundForDisplayImpl(doom.p05, step)}–${roundForDisplayImpl(doom.p95, step)}` : "";
  return `Under your selected assumptions—not the p(DOOM) official model—the median estimate is ${central} (plausible interval ${interval}) for the ${result.horizon} horizon.`;
}

import { roundForDisplay as roundForDisplayImpl } from "./rounding";
