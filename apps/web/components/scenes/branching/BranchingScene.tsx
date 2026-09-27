// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
/**
 * Branching futures: an SVG tree read left to right. A single trunk for the
 * present splits into one branch per recoverability group; branch thickness
 * grows with the number of scenarios in the group and colour marks
 * recoverability. Leaves are the individual scenarios (short ids as labels);
 * faint dashed blue ties across the branches stand for the safeguards in the
 * release. Layout is seeded and deterministic; the grow-in is CSS only and is
 * skipped under reduced motion. Conceptual only; not a simulation of AI risk.
 */
import { useMemo } from "react";
import type { StageData } from "../HomeStage";
import {
  RECOVERABILITY_ORDER,
  RecoverabilityMarker,
  asRecoverability,
  recoverabilityColor,
  recoverabilityWord,
  type Recoverability,
} from "../grammar";
import { mulberry32 } from "../seeded";

// The stage crops the 1000x600 viewBox with "slice"; on narrow phone stages
// only x in [200, 800] stays visible, so the whole tree is laid out inside it.
const X_ROOT = 200;
const X_TRUNK = 300;
const X_GROUP = 480;
const X_LEAF = 690;
const CY = 300;
const Y_TOP = 70;
const Y_BOTTOM = 530;
const GAP = 26;
const MAX_TIES = 14;

const LABEL =
  "Conceptual branching-futures tree; not a simulation of AI risk. A single trunk for the present splits into branches, one per recoverability group. A branch grows thicker with the number of scenarios in its group, and its colour marks recoverability: green for recoverable, amber for hard to recover, red for not recoverable, grey for unknown. The leaves are the individual scenarios; faint dashed blue ties across the branches stand for the safeguards in the release. The tree grows in once and stands still when reduced motion is requested.";

type P = [number, number];
type Cubic = [P, P, P, P];

interface Leaf {
  id: string;
  title: string;
  recoverability: string;
  x: number;
  y: number;
  d: string;
}

interface Branch {
  key: Recoverability;
  word: string;
  color: string;
  width: number;
  y: number;
  d: string;
  ctrl: Cubic;
  leaves: Leaf[];
}

const f = (n: number) => n.toFixed(1);

function pathOf([p0, p1, p2, p3]: Cubic): string {
  return `M ${f(p0[0])} ${f(p0[1])} C ${f(p1[0])} ${f(p1[1])}, ${f(p2[0])} ${f(p2[1])}, ${f(p3[0])} ${f(p3[1])}`;
}

function cubicAt([p0, p1, p2, p3]: Cubic, t: number): P {
  const u = 1 - t;
  const a = u * u * u;
  const b = 3 * u * u * t;
  const c = 3 * u * t * t;
  const d = t * t * t;
  return [
    a * p0[0] + b * p1[0] + c * p2[0] + d * p3[0],
    a * p0[1] + b * p1[1] + c * p2[1] + d * p3[1],
  ];
}

function buildTies(branches: Branch[], count: number, r: () => number): string[] {
  const k = Math.min(MAX_TIES, Math.max(0, count));
  const first = branches[0];
  if (k === 0 || !first) return [];
  const ties: string[] = [];
  for (let j = 0; j < k; j++) {
    const t = 0.28 + (0.55 * j) / Math.max(1, k - 1) + (r() - 0.5) * 0.05;
    if (branches.length === 1) {
      const p = cubicAt(first.ctrl, t);
      ties.push(`M ${f(p[0])} ${f(p[1] - 16)} L ${f(p[0])} ${f(p[1] + 16)}`);
      continue;
    }
    const gi = j % (branches.length - 1);
    const a = branches[gi];
    const b = branches[gi + 1];
    if (!a || !b) continue;
    const pa = cubicAt(a.ctrl, t);
    const pb = cubicAt(b.ctrl, t);
    const mx = (pa[0] + pb[0]) / 2 + 12;
    const my = (pa[1] + pb[1]) / 2;
    ties.push(`M ${f(pa[0])} ${f(pa[1])} Q ${f(mx)} ${f(my)} ${f(pb[0])} ${f(pb[1])}`);
  }
  return ties;
}

function buildTree(data: StageData, seed: number) {
  const r = mulberry32(seed);
  const jitter = (amt: number) => (r() - 0.5) * amt;
  const grouped = RECOVERABILITY_ORDER.map((key) => ({
    key,
    items: data.scenarios.filter((s) => asRecoverability(s.recoverability) === key),
  })).filter((g) => g.items.length > 0);
  const total = grouped.reduce((n, g) => n + g.items.length, 0);
  const avail = Y_BOTTOM - Y_TOP - GAP * Math.max(0, grouped.length - 1);
  let cursor = Y_TOP;
  const branches: Branch[] = grouped.map((g) => {
    const h = total > 0 ? (avail * g.items.length) / total : 0;
    const y0 = cursor;
    cursor += h + GAP;
    const y = y0 + h / 2;
    const ctrl: Cubic = [
      [X_TRUNK, CY],
      [X_TRUNK + 90 + jitter(20), CY],
      [X_GROUP - 90 + jitter(20), y],
      [X_GROUP, y],
    ];
    const leaves: Leaf[] = g.items.map((s, i) => {
      const ly = y0 + (h * (i + 0.5)) / g.items.length;
      const lx = X_LEAF + jitter(30);
      const c: Cubic = [
        [X_GROUP, y],
        [X_GROUP + 80 + jitter(20), y],
        [lx - 80 + jitter(20), ly],
        [lx, ly],
      ];
      return {
        id: s.id,
        title: `${s.id} — ${s.name}; ${recoverabilityWord(s.recoverability)}`,
        recoverability: s.recoverability,
        x: lx,
        y: ly,
        d: pathOf(c),
      };
    });
    return {
      key: g.key,
      word: recoverabilityWord(g.key),
      color: recoverabilityColor(g.key),
      width: Math.min(14, 2 + g.items.length * 1.5),
      y,
      d: pathOf(ctrl),
      ctrl,
      leaves,
    };
  });
  const ties = buildTies(branches, data.interventions.length, r);
  return { branches, ties, trunkWidth: Math.min(22, 4 + total * 0.9) };
}

function delay(seconds: number) {
  return { animationDelay: `${seconds.toFixed(2)}s` };
}

export function BranchingScene({
  data,
  reducedMotion,
}: {
  data: StageData;
  reducedMotion: boolean;
}) {
  const seed = data.seed ?? 7;
  const tree = useMemo(() => buildTree(data, seed), [data, seed]);
  return (
    <svg
      className={`scene-svg branching${reducedMotion ? " scene-static" : ""}`}
      viewBox="0 0 1000 600"
      preserveAspectRatio="xMidYMid slice"
      role="img"
      aria-label={LABEL}
    >
      <text x={X_ROOT} y={CY - 18} className="branch-label">
        now
      </text>
      <path
        d={`M ${X_ROOT} ${CY} L ${X_TRUNK} ${CY}`}
        pathLength={1}
        className="branch-path"
        stroke="var(--c-text-2)"
        strokeWidth={tree.trunkWidth}
        strokeOpacity={0.85}
      />
      {tree.branches.map((b, gi) => (
        <g key={b.key}>
          <path
            d={b.d}
            pathLength={1}
            className="branch-path"
            stroke={b.color}
            strokeWidth={b.width}
            strokeOpacity={0.9}
            style={delay(0.45)}
          />
          <text
            x={X_GROUP - 8}
            y={b.y - b.width / 2 - 7}
            textAnchor="end"
            className="branch-label branch-fade"
            style={delay(1.1 + gi * 0.05)}
          >
            {b.word}
          </text>
          {b.leaves.map((l, li) => (
            <g key={l.id}>
              <title>{l.title}</title>
              <path
                d={l.d}
                pathLength={1}
                className="branch-path"
                stroke={b.color}
                strokeWidth={1.6}
                strokeOpacity={0.75}
                style={delay(1 + gi * 0.08 + li * 0.04)}
              />
              <RecoverabilityMarker
                x={l.x}
                y={l.y}
                r={4.5}
                recoverability={l.recoverability}
                className="branch-fade"
                style={delay(1.8 + gi * 0.08 + li * 0.04)}
              />
              <text
                x={l.x + 9}
                y={l.y + 3.5}
                className="branch-id branch-fade"
                style={delay(1.9 + gi * 0.08 + li * 0.04)}
              >
                {l.id}
              </text>
            </g>
          ))}
        </g>
      ))}
      {tree.ties.map((d, i) => (
        <path key={i} d={d} className="branch-tie branch-fade" style={delay(2.2 + i * 0.06)} />
      ))}
    </svg>
  );
}
