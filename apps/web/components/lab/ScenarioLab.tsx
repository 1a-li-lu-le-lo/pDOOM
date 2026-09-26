// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { HORIZON_KEYS, USER_SCENARIO_SLIDER_KEYS, type ExperimentalCausalSpec, type UserScenarioParams, type UserScenarioResult, type UserScenarioSliderKey } from "@pdoom/schemas";
import { evaluateUserScenario, formatInterval, roundForDisplay, roundingStep } from "@pdoom/model-core";
import { QuantileStrip } from "../charts/QuantileStrip";
import { ShareLink } from "../share/ShareLink";
import { horizonLabel, outcomeLabel } from "@/lib/format";

const SLIDER_LABEL: Record<UserScenarioSliderKey, { label: string; low: string; high: string; factor: string }> = {
  capability_timeline: { label: "Capability timeline", low: "much slower", high: "much faster", factor: "A (capability pressure)" },
  autonomy_growth: { label: "Autonomy growth", low: "much slower", high: "much faster", factor: "C (control) ↓" },
  access_level: { label: "System access", low: "much narrower", high: "much broader", factor: "C ↓, E (exposure) ↑" },
  safety_progress: { label: "Safety research progress", low: "much slower", high: "much faster", factor: "F (fragility) ↓" },
  governance_strength: { label: "Governance strength", low: "much weaker", high: "much stronger", factor: "F ↓" },
  model_security: { label: "Model-weight security", low: "much weaker", high: "much stronger", factor: "E ↓" },
  open_weight_diffusion: { label: "Open-weight diffusion", low: "much less", high: "much more", factor: "E ↑" },
  international_coordination: { label: "International coordination", low: "much weaker", high: "much stronger", factor: "F ↓" },
  incident_frequency: { label: "Incident frequency", low: "much lower", high: "much higher", factor: "E ↑" },
  resilience: { label: "Civilizational resilience", low: "much lower", high: "much higher", factor: "O4, O5, O6 base rates ↓" },
};

const SHORT: Record<UserScenarioSliderKey, string> = {
  capability_timeline: "ct",
  autonomy_growth: "ag",
  access_level: "al",
  safety_progress: "sp",
  governance_strength: "gs",
  model_security: "ms",
  open_weight_diffusion: "od",
  international_coordination: "ic",
  incident_frequency: "if",
  resilience: "rs",
};

type Baseline = Record<string, { p05: number; p50: number; p95: number; display: string; interval: string } | null>;

function readParams(sp: URLSearchParams, horizons: string[]): UserScenarioParams {
  const h = sp.get("h");
  const p: Record<string, number | string> = { horizon: h && horizons.includes(h) ? h : horizons.includes("10y") ? "10y" : horizons[0]! };
  for (const k of USER_SCENARIO_SLIDER_KEYS) {
    const v = Number(sp.get(SHORT[k]) ?? 0);
    p[k] = Number.isInteger(v) && v >= -2 && v <= 2 ? v : 0;
  }
  return p as unknown as UserScenarioParams;
}

function toQuery(p: UserScenarioParams): string {
  const q = new URLSearchParams();
  q.set("h", p.horizon);
  for (const k of USER_SCENARIO_SLIDER_KEYS) if (p[k] !== 0) q.set(SHORT[k], String(p[k]));
  return q.toString();
}

/**
 * Scenario Lab. The model runs in the browser (model-core mirrors the Go
 * implementation bit for bit). The URL is the state: every result is
 * reproducible from its link, which is what makes it worth sharing.
 */
export function ScenarioLab({ spec, baseline, releaseId }: { spec: ExperimentalCausalSpec; baseline: Baseline; releaseId: string }) {
  const router = useRouter();
  const sp = useSearchParams();
  const horizons = HORIZON_KEYS.filter((h) => h in spec.horizons);
  const [params, setParams] = useState<UserScenarioParams>(() => readParams(new URLSearchParams(sp.toString()), horizons));
  const [samples, setSamples] = useState(Math.min(spec.samples, 5000));
  const [result, setResult] = useState<UserScenarioResult | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    setBusy(true);
    const id = window.setTimeout(() => {
      try {
        setResult(evaluateUserScenario(spec, params, samples));
      } finally {
        setBusy(false);
      }
    }, 30);
    return () => window.clearTimeout(id);
  }, [spec, params, samples]);

  useEffect(() => {
    const q = toQuery(params);
    if (q !== sp.toString()) router.replace(`/lab?${q}`, { scroll: false });
  }, [params, router, sp]);

  const set = useCallback((k: UserScenarioSliderKey, v: number) => setParams((p) => ({ ...p, [k]: v })), []);
  const reset = () => setParams((p) => ({ ...readParams(new URLSearchParams(), horizons), horizon: p.horizon }));
  const touched = USER_SCENARIO_SLIDER_KEYS.filter((k) => params[k] !== 0).length;
  const base = baseline[params.horizon];
  const doom = result?.outcome_estimates.P_DOOM;
  const step = roundingStep("high");
  const deltaText = useMemo(() => {
    if (!doom || !base) return null;
    const d = (doom.p50 - base.p50) * 100;
    if (Math.abs(d) < 0.5) return "about the same median as the published research-mode run";
    return `${Math.abs(d) < 1 ? "under one point" : `${Math.round(Math.abs(d))} points`} ${d > 0 ? "above" : "below"} the published research-mode median`;
  }, [doom, base]);

  return (
    <div className="lab">
      <div className="lab-controls card stack">
        <div className="row" role="group" aria-label="Horizon">
          {horizons.map((h) => (
            <button key={h} type="button" className="btn" aria-pressed={params.horizon === h} onClick={() => setParams((p) => ({ ...p, horizon: h }))}>
              {horizonLabel(h)}
            </button>
          ))}
        </div>
        {USER_SCENARIO_SLIDER_KEYS.map((k) => {
          const m = SLIDER_LABEL[k];
          const v = params[k];
          return (
            <div key={k} className="slider">
              <label htmlFor={`s-${k}`}>
                <span>{m.label}</span>
                <span className="cite">{m.factor}</span>
              </label>
              <div className="slider-row">
                <span className="cite">{m.low}</span>
                <input id={`s-${k}`} type="range" min={-2} max={2} step={1} value={v} onChange={(e) => set(k, Number(e.target.value))} aria-valuetext={v === 0 ? "baseline" : v > 0 ? `${v > 1 ? "much " : ""}${m.high}` : `${v < -1 ? "much " : ""}${m.low}`} list={`ticks-${k}`} />
                <span className="cite">{m.high}</span>
              </div>
              <datalist id={`ticks-${k}`}>
                {[-2, -1, 0, 1, 2].map((t) => (
                  <option key={t} value={t} label={t === 0 ? "baseline" : String(t)} />
                ))}
              </datalist>
            </div>
          );
        })}
        <div className="row">
          <label className="cite">
            Samples{" "}
            <select value={samples} onChange={(e) => setSamples(Number(e.target.value))}>
              {[2000, 5000, 10000, spec.samples].filter((n, i, a) => n <= spec.samples && a.indexOf(n) === i).map((n) => (
                <option key={n} value={n}>
                  {n.toLocaleString("en-US")}
                </option>
              ))}
            </select>
          </label>
          <button type="button" className="btn" onClick={reset}>
            Reset dials
          </button>
        </div>
      </div>

      <div className="lab-result stack" aria-live="polite" aria-busy={busy}>
        <div className="card stack">
          <div className="row">
            <span className="badge badge-uncertainty">Your scenario</span>
            <span className="badge">{horizonLabel(params.horizon)}</span>
            <span className="badge">{touched === 0 ? "baseline assumptions" : `${touched} of ${USER_SCENARIO_SLIDER_KEYS.length} dials moved`}</span>
            <span className="badge">{spec.version}</span>
          </div>
          <p className="lede" style={{ margin: 0 }}>
            Under your selected assumptions, not the p(DOOM) official model, the median estimate is
          </p>
          <div className="lab-value" style={{ color: "var(--c-uncertainty)" }}>
            {doom ? roundForDisplay(doom.p50, step) : "…"}
          </div>
          <div className="interval">
            {doom ? `plausible interval ${formatInterval(doom.p05, doom.p95, step)} · p(DOOM), O3–O8 combined, ${horizonLabel(params.horizon)}` : "computing"}
            {deltaText ? ` · ${deltaText}` : ""}
          </div>
          {doom ? <QuantileStrip p05={doom.p05} p25={doom.p25} p50={doom.p50} p75={doom.p75} p95={doom.p95} label={`Your scenario, ${horizonLabel(params.horizon)}`} description="Distribution of the combined outcome under your assumptions. The band is the p05–p95 interval, the darker band p25–p75, the tick the median." colorVar="var(--c-uncertainty)" compact /> : null}
          {base ? (
            <p className="cite">
              Published research-mode run for this horizon (release {releaseId}): median {base.display}, interval {base.interval}. Your scenario does not change it.
            </p>
          ) : null}
          {result?.flags.length ? (
            <ul className="flags">
              {result.flags.map((f) => (
                <li key={f} className="badge badge-uncertainty">
                  {f}
                </li>
              ))}
            </ul>
          ) : null}
        </div>

        {result ? (
          <div className="card stack">
            <div className="eyebrow">Decomposition under your assumptions</div>
            <div className="table-wrap">
              <table>
                <caption>Outcome medians and intervals for your scenario</caption>
                <thead>
                  <tr>
                    <th>Outcome</th>
                    <th className="num">Median</th>
                    <th>p05–p95</th>
                  </tr>
                </thead>
                <tbody>
                  {(["O3", "O4", "O5", "O6", "O7", "O8", "P_COLLAPSE", "P_DOOM"] as const).map((k) => {
                    const s = result.outcome_estimates[k];
                    if (!s) return null;
                    return (
                      <tr key={k}>
                        <td>{k === "P_DOOM" ? "p(DOOM) O3–O8" : k === "P_COLLAPSE" ? "Collapse or near-extinction (O4+O5)" : `${k} ${outcomeLabel(k)}`}</td>
                        <td className="num">{roundForDisplay(s.p50, step)}</td>
                        <td>{formatInterval(s.p05, s.p95, step)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            <div className="eyebrow">Factor summary (p05 / p50 / p95)</div>
            <div className="table-wrap">
              <table>
                <caption>Latent factors after your shifts</caption>
                <thead>
                  <tr>
                    <th>Factor</th>
                    <th className="num">p05</th>
                    <th className="num">p50</th>
                    <th className="num">p95</th>
                  </tr>
                </thead>
                <tbody>
                  {(["A", "C", "E", "F"] as const).map((f) => (
                    <tr key={f}>
                      <td>
                        {f} · {f === "A" ? "capability pressure" : f === "C" ? "control strength" : f === "E" ? "exposure" : "fragility"}
                      </td>
                      <td className="num">{result.factor_summary[f].p05.toFixed(2)}</td>
                      <td className="num">{result.factor_summary[f].p50.toFixed(2)}</td>
                      <td className="num">{result.factor_summary[f].p95.toFixed(2)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ) : null}

        <div className="row no-print">
          <ShareLink label="Copy link to this scenario" />
          <Link className="btn" href="/meter">
            Back to the published meter
          </Link>
        </div>
        <p className="cite">{result?.disclaimer}</p>
      </div>
    </div>
  );
}
