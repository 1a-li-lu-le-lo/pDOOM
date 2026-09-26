// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
/**
 * Orrery: concentric orbits drawn in SVG and turned by CSS. The centre is the
 * withheld official object (a hollow dashed ring, no text). The inner orbit
 * carries the indexes as small labelled planets whose distance and size follow
 * the index value (hollow grey when insufficient); the middle orbit carries the
 * scenarios coloured and shaped by recoverability; the outer orbit carries one
 * blue safeguard marker per intervention. Words may appear as labels; numbers
 * never do. Conceptual only; not a simulation of AI risk.
 */
import { useMemo } from "react";
import type { StageData } from "../HomeStage";
import { RecoverabilityMarker, indexLabel, recoverabilityWord } from "../grammar";
import { clamp, mulberry32 } from "../seeded";

const CX = 500;
const CY = 300;
const TAU = Math.PI * 2;
const CENTRE_R = 58;
const INDEX_NEAR = 118;
const INDEX_FAR = 180;
const INDEX_GUIDE = (INDEX_NEAR + INDEX_FAR) / 2;
const SCENARIO_R = 215;
const SAFEGUARD_R = 262;

const LABEL =
  "Conceptual orrery of the current release; not a simulation of AI risk. The hollow dashed centre stands for the official estimate, which is withheld. The inner orbit carries the indexes as small labelled planets, placed further out and drawn larger as an index rises, and hollow grey where evidence is insufficient. The middle orbit carries the scenarios, coloured and shaped by recoverability. The outer orbit carries one blue safeguard marker per intervention. The orbits turn slowly and stand still when reduced motion is requested.";

interface Planet {
  id: string;
  label: string;
  x: number;
  y: number;
  size: number;
  hollow: boolean;
  left: boolean;
}

interface Node {
  id: string;
  title: string;
  recoverability: string;
  x: number;
  y: number;
  left: boolean;
}

interface Tick {
  id: string;
  title: string;
  x1: number;
  y1: number;
  x2: number;
  y2: number;
}

function spread(i: number, n: number): number {
  return -Math.PI / 2 + (i / Math.max(1, n)) * TAU;
}

function buildLayout(data: StageData, seed: number) {
  const r = mulberry32(seed);
  const jitter = (amt: number) => (r() - 0.5) * amt;
  const planets: Planet[] = data.indexes.map((ix, i, arr) => {
    const v = ix.value === null ? null : clamp(ix.value / 100, 0, 1);
    const a = spread(i, arr.length) + jitter(0.16);
    const dist = INDEX_NEAR + (v ?? 0.5) * (INDEX_FAR - INDEX_NEAR);
    return {
      id: ix.id,
      label: indexLabel(ix.id),
      x: CX + Math.cos(a) * dist,
      y: CY + Math.sin(a) * dist,
      size: v === null ? 6 : 4 + v * 8,
      hollow: v === null,
      left: Math.cos(a) < -0.2,
    };
  });
  const nodes: Node[] = data.scenarios.map((s, i, arr) => {
    const a = spread(i, arr.length) + jitter(0.08);
    const dist = SCENARIO_R + jitter(14);
    return {
      id: s.id,
      title: `${s.id} — ${s.name}; ${recoverabilityWord(s.recoverability)}`,
      recoverability: s.recoverability,
      x: CX + Math.cos(a) * dist,
      y: CY + Math.sin(a) * dist,
      left: Math.cos(a) < -0.2,
    };
  });
  const ticks: Tick[] = data.interventions.map((iv, i, arr) => {
    const a = spread(i, arr.length) + jitter(0.05);
    const c = Math.cos(a);
    const s = Math.sin(a);
    return {
      id: iv.id,
      title: `${iv.name} (safeguard)`,
      x1: CX + c * (SAFEGUARD_R - 5),
      y1: CY + s * (SAFEGUARD_R - 5),
      x2: CX + c * (SAFEGUARD_R + 5),
      y2: CY + s * (SAFEGUARD_R + 5),
    };
  });
  return { planets, nodes, ticks };
}

function origin(x: number, y: number) {
  return { transformOrigin: `${x.toFixed(1)}px ${y.toFixed(1)}px` };
}

export function OrreryScene({ data, reducedMotion }: { data: StageData; reducedMotion: boolean }) {
  const seed = data.seed ?? 7;
  const layout = useMemo(() => buildLayout(data, seed), [data, seed]);
  return (
    <svg
      className={`scene-svg orrery${reducedMotion ? " scene-static" : ""}`}
      viewBox="0 0 1000 600"
      preserveAspectRatio="xMidYMid slice"
      role="img"
      aria-label={LABEL}
    >
      <defs>
        <radialGradient id="orrery-glow" cx={0.5} cy={0.5} r={0.5}>
          <stop offset={0} stopColor="#3b2a5e" stopOpacity={0.35} />
          <stop offset={0.55} stopColor="#1a0d05" stopOpacity={0.12} />
          <stop offset={1} stopColor="#050508" stopOpacity={0} />
        </radialGradient>
      </defs>
      <rect x={0} y={0} width={1000} height={600} fill="url(#orrery-glow)" />

      {/* orbit guides */}
      <circle cx={CX} cy={CY} r={INDEX_GUIDE} className="orrery-guide" strokeDasharray="2 6" />
      <circle cx={CX} cy={CY} r={SCENARIO_R} className="orrery-guide" />
      <circle cx={CX} cy={CY} r={SAFEGUARD_R} className="orrery-guide" />

      {/* the withheld official object: hollow, dashed, text-free */}
      <circle cx={CX} cy={CY} r={CENTRE_R} fill="var(--c-bg)" stroke="var(--c-insufficient)" strokeWidth={1.5} strokeDasharray="7 7" />
      <circle cx={CX} cy={CY} r={CENTRE_R - 14} fill="none" stroke="var(--c-insufficient)" strokeOpacity={0.35} strokeDasharray="2 6" />

      {/* inner orbit: indexes */}
      <g className="orrery-orbit orrery-period-inner">
        {layout.planets.map((p) => (
          <g key={p.id} className="orrery-upright orrery-period-inner" style={origin(p.x, p.y)}>
            <title>{p.label}</title>
            {p.hollow ? (
              <circle cx={p.x} cy={p.y} r={p.size} fill="none" stroke="var(--c-insufficient)" strokeWidth={1.5} strokeDasharray="3 3" />
            ) : (
              <circle cx={p.x} cy={p.y} r={p.size} fill="var(--c-evidence)" />
            )}
            <text x={p.left ? p.x - p.size - 6 : p.x + p.size + 6} y={p.y + 4} textAnchor={p.left ? "end" : "start"} className="orrery-label">
              {p.label}
            </text>
          </g>
        ))}
      </g>

      {/* middle orbit: scenarios by recoverability */}
      <g className="orrery-orbit orrery-period-middle">
        {layout.nodes.map((n) => (
          <g key={n.id} className="orrery-upright orrery-period-middle" style={origin(n.x, n.y)}>
            <title>{n.title}</title>
            <RecoverabilityMarker x={n.x} y={n.y} r={5} recoverability={n.recoverability} />
            <text x={n.left ? n.x - 9 : n.x + 9} y={n.y + 3.5} textAnchor={n.left ? "end" : "start"} className="orrery-id">
              {n.id}
            </text>
          </g>
        ))}
      </g>

      {/* outer orbit: safeguards */}
      <g className="orrery-orbit orrery-period-outer">
        {layout.ticks.map((t) => (
          <line key={t.id} x1={t.x1} y1={t.y1} x2={t.x2} y2={t.y2} className="orrery-safeguard">
            <title>{t.title}</title>
          </line>
        ))}
      </g>
    </svg>
  );
}
