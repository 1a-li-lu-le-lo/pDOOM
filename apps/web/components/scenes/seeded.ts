// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Seeded pseudo-random numbers for the scenes. Every placement in a scene is
 * derived from the release seed through this generator, so the same release
 * always draws the same picture and nothing in a scene depends on Math.random.
 */

/** mulberry32: a small, fast 32-bit generator returning floats in [0, 1). */
export function mulberry32(seed: number): () => number {
  let a = seed | 0;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** Clamp a value into [lo, hi]; NaN and non-finite values fall to lo. */
export function clamp(v: number, lo: number, hi: number): number {
  if (!Number.isFinite(v)) return lo;
  return Math.min(hi, Math.max(lo, v));
}
