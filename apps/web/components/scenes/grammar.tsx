// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Shared visual grammar for the SVG scenes. Recoverability maps to a colour
 * token and to a marker shape (colour is never the only channel, mirroring the
 * badge grammar in globals.css). Labels are words only: no helper here ever
 * renders a number as text.
 */
import type { CSSProperties } from "react";

export const RECOVERABILITY_ORDER = ["high", "moderate", "low", "none", "unknown"] as const;
export type Recoverability = (typeof RECOVERABILITY_ORDER)[number];

export function asRecoverability(v: string): Recoverability {
  return (RECOVERABILITY_ORDER as readonly string[]).includes(v)
    ? (v as Recoverability)
    : "unknown";
}

const COLOR: Record<Recoverability, string> = {
  high: "var(--c-resilience)",
  moderate: "var(--c-resilience)",
  low: "var(--c-uncertainty)",
  none: "var(--c-risk)",
  unknown: "var(--c-insufficient)",
};

const WORD: Record<Recoverability, string> = {
  high: "recoverable",
  moderate: "largely recoverable",
  low: "hard to recover",
  none: "not recoverable",
  unknown: "recoverability unknown",
};

export function recoverabilityColor(v: string): string {
  return COLOR[asRecoverability(v)];
}

export function recoverabilityWord(v: string): string {
  return WORD[asRecoverability(v)];
}

/** "capability_pressure" → "Capability pressure". */
export function indexLabel(id: string): string {
  return id.replace(/_/g, " ").replace(/^\w/, (c) => c.toUpperCase());
}

export interface MarkerProps {
  x: number;
  y: number;
  r: number;
  recoverability: string;
  className?: string;
  style?: CSSProperties;
}

/**
 * Marker whose shape follows the badge grammar: circle = recoverable,
 * diamond = hard to recover, square = not recoverable, hollow ring = unknown.
 */
export function RecoverabilityMarker({ x, y, r, recoverability, className, style }: MarkerProps) {
  const kind = asRecoverability(recoverability);
  const fill = COLOR[kind];
  if (kind === "unknown") {
    return (
      <circle
        cx={x}
        cy={y}
        r={r}
        fill="none"
        stroke={fill}
        strokeWidth={1.5}
        className={className}
        style={style}
      />
    );
  }
  if (kind === "none") {
    return (
      <rect
        x={x - r}
        y={y - r}
        width={r * 2}
        height={r * 2}
        fill={fill}
        className={className}
        style={style}
      />
    );
  }
  if (kind === "low") {
    return (
      <rect
        x={x - r}
        y={y - r}
        width={r * 2}
        height={r * 2}
        fill={fill}
        transform={`rotate(45 ${x.toFixed(1)} ${y.toFixed(1)})`}
        className={className}
        style={style}
      />
    );
  }
  return <circle cx={x} cy={y} r={r} fill={fill} className={className} style={style} />;
}
