// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Static SVG accretion disk: the WebGL and reduced-motion fallback, and the
 * server-rendered first frame behind the meter. Purely conceptual — labelled
 * as such wherever it appears. Ring thickness encodes the interval width and
 * the number of orbit dots encodes source coverage; nothing here is a simulation.
 */
export interface StaticDiskProps {
  /** Interval width (p95 − p05) of the featured estimate, 0..1; drives ring thickness. */
  intervalWidth?: number;
  /** Evidence pressure 0..100; drives disk brightness. */
  brightness?: number;
  /** Number of evidence particles to draw (capped). */
  particles?: number;
  /** Safeguard arcs to draw. */
  safeguards?: number;
  seed?: number;
}

function rng(seed: number) {
  let a = seed | 0;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

export function StaticDisk({ intervalWidth = 0.2, brightness = 50, particles = 140, safeguards = 5, seed = 7 }: StaticDiskProps) {
  const r = rng(seed);
  const n = Math.min(400, Math.max(20, particles));
  const dots = Array.from({ length: n }, (_, i) => {
    const a = r() * Math.PI * 2;
    const rad = 120 + r() * 220;
    const tilt = 0.42;
    return { x: 500 + Math.cos(a) * rad, y: 300 + Math.sin(a) * rad * tilt, s: 0.6 + r() * 1.8, o: 0.25 + r() * 0.7, k: i };
  });
  const ring = 18 + Math.min(80, intervalWidth * 240);
  const glow = 0.35 + Math.min(0.55, brightness / 160);
  return (
    <svg className="static-disk" viewBox="0 0 1000 600" preserveAspectRatio="xMidYMid slice" role="img" aria-label="Conceptual risk visualization: an abstract accretion disk. The centre stands for irreversible catastrophe, orbiting points for evidence, the ring thickness for the plausible interval, and blue arcs for safeguards. It is not a simulation of AI risk.">
      <defs>
        <radialGradient id="disk-glow" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#ffd9a0" stopOpacity={glow} />
          <stop offset="35%" stopColor="#d88a5a" stopOpacity={glow * 0.5} />
          <stop offset="70%" stopColor="#3b2a5e" stopOpacity={0.25} />
          <stop offset="100%" stopColor="#050508" stopOpacity={0} />
        </radialGradient>
        <radialGradient id="horizon" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#000" />
          <stop offset="85%" stopColor="#000" />
          <stop offset="100%" stopColor="#1a0d05" stopOpacity={0.6} />
        </radialGradient>
      </defs>
      <ellipse cx={500} cy={300} rx={420} ry={176} fill="url(#disk-glow)" />
      {/* interval ring: thickness = plausible interval */}
      <ellipse cx={500} cy={300} rx={250} ry={105} fill="none" stroke="#f5f5f7" strokeOpacity={0.18} strokeWidth={ring} />
      <ellipse cx={500} cy={300} rx={250} ry={105} fill="none" stroke="#f5f5f7" strokeOpacity={0.7} strokeWidth={1.5} />
      {/* safeguard arcs */}
      {Array.from({ length: Math.min(12, safeguards) }, (_, i) => {
        const start = (i / Math.max(1, safeguards)) * Math.PI * 2;
        const end = start + 0.7;
        const rx = 300 + (i % 3) * 22;
        const ry = rx * 0.42;
        const p = (t: number) => `${(500 + Math.cos(t) * rx).toFixed(1)},${(300 + Math.sin(t) * ry).toFixed(1)}`;
        return <path key={i} d={`M${p(start)} A${rx},${ry} 0 0 1 ${p(end)}`} fill="none" stroke="var(--c-safeguard)" strokeOpacity={0.8} strokeWidth={2} strokeLinecap="round" />;
      })}
      {/* evidence particles */}
      {dots.map((d) => (
        <circle key={d.k} cx={d.x} cy={d.y} r={d.s} fill="#f5f5f7" opacity={d.o} />
      ))}
      {/* event horizon */}
      <circle cx={500} cy={300} r={58} fill="url(#horizon)" />
      <circle cx={500} cy={300} r={58} fill="none" stroke="#ffb27a" strokeOpacity={0.5} strokeWidth={1} />
    </svg>
  );
}
