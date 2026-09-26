// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
/**
 * The scene graph inside the Event Horizon canvas. Everything is derived from the
 * StageData once (seeded, deterministic) and animated from a single frame
 * callback that only touches uniforms and group rotations. The grammar mirrors
 * StaticDisk: a dark horizon at the centre, an accretion band whose thickness is
 * the interval width, brightness that follows evidence pressure, orbiting
 * evidence particles, blue safeguard arcs outside the band and a faint arc of
 * research-mode points on the far side. Conceptual only; not a simulation.
 */
import { useFrame, useThree } from "@react-three/fiber";
import { useCallback, useEffect, useMemo, useRef, type RefObject } from "react";
import {
  AdditiveBlending,
  BufferAttribute,
  BufferGeometry,
  CanvasTexture,
  CircleGeometry,
  Color,
  DoubleSide,
  MeshBasicMaterial,
  RingGeometry,
  ShaderMaterial,
  SphereGeometry,
  SpriteMaterial,
  type Group,
} from "three";
import type { StageData } from "../HomeStage";
import { clamp, mulberry32 } from "../seeded";
import { BAND_FRAG, GLOW_FRAG, PLANE_VERT, POINT_FRAG, POINT_VERT } from "./shaders";

export const RING_RADIUS = 2.4;
export const HORIZON_RADIUS = 0.72;
export const PARTICLE_CAP = 400;
const PARTICLE_MIN = 20;
const GLOW_RADIUS = RING_RADIUS * 1.65;
/** One full orbit at the ring radius; inner particles are faster, outer slower. */
const ORBIT_SECONDS = 48;
const INTRO_SECONDS = 2.5;
const MAX_DELTA = 0.1;
const TAU = Math.PI * 2;
/** Fixed warm palette of the disk (matches the StaticDisk gradient stops). */
const HOT = "#ffd9a0";
const WARM = "#d88a5a";
const COLD = "#3b2a5e";
const RIM_RGB = "255, 178, 122";
const RIM_EDGE = 0.685;
const RIM_SCALE = (HORIZON_RADIUS / RIM_EDGE) * 2;

export interface Palette {
  bg: string;
  evidence: string;
  safeguard: string;
  uncertainty: string;
}

export interface SceneModel {
  seed: number;
  /** p95 − p05 of the featured estimate, 0..1. */
  interval: number;
  /** Evidence pressure normalised to 0..1. */
  bright: number;
  particles: number;
  safeguards: number;
  halo: boolean;
  curve: StageData["researchCurve"];
}

export interface IntroControl {
  skipped: boolean;
}

export function deriveModel(data: StageData, particleScale: number, halo: boolean): SceneModel {
  return {
    seed: data.seed ?? 7,
    interval: clamp(data.intervalWidth ?? 0.2, 0, 1),
    bright: clamp((data.brightness ?? 50) / 100, 0, 1),
    particles: Math.round(clamp((data.particles ?? 140) * particleScale, PARTICLE_MIN, PARTICLE_CAP)),
    safeguards: Math.round(clamp(data.safeguards ?? 0, 0, 12)),
    halo,
    curve: data.researchCurve ?? [],
  };
}

function easeOutCubic(t: number): number {
  const u = 1 - t;
  return 1 - u * u * u;
}

interface PointSpec {
  radius: number;
  angle: number;
  speed: number;
  size: number;
  alpha: number;
  height: number;
}

function pointGeometry(points: PointSpec[]): BufferGeometry {
  const n = points.length;
  const position = new Float32Array(n * 3);
  const radius = new Float32Array(n);
  const angle = new Float32Array(n);
  const speed = new Float32Array(n);
  const size = new Float32Array(n);
  const alpha = new Float32Array(n);
  const height = new Float32Array(n);
  points.forEach((p, i) => {
    position[i * 3] = Math.cos(p.angle) * p.radius;
    position[i * 3 + 1] = p.height;
    position[i * 3 + 2] = Math.sin(p.angle) * p.radius;
    radius[i] = p.radius;
    angle[i] = p.angle;
    speed[i] = p.speed;
    size[i] = p.size;
    alpha[i] = p.alpha;
    height[i] = p.height;
  });
  const g = new BufferGeometry();
  g.setAttribute("position", new BufferAttribute(position, 3));
  g.setAttribute("aRadius", new BufferAttribute(radius, 1));
  g.setAttribute("aAngle", new BufferAttribute(angle, 1));
  g.setAttribute("aSpeed", new BufferAttribute(speed, 1));
  g.setAttribute("aSize", new BufferAttribute(size, 1));
  g.setAttribute("aAlpha", new BufferAttribute(alpha, 1));
  g.setAttribute("aHeight", new BufferAttribute(height, 1));
  return g;
}

/** All PARTICLE_CAP evidence particles, seeded; the visible count is a draw range. */
function buildParticleGeometry(seed: number): BufferGeometry {
  const r = mulberry32(seed);
  const points: PointSpec[] = [];
  for (let i = 0; i < PARTICLE_CAP; i++) {
    const angle = r() * TAU;
    const radius = RING_RADIUS * (0.48 + r() * 0.88);
    points.push({
      radius,
      angle,
      speed: (TAU / ORBIT_SECONDS) * Math.pow(RING_RADIUS / radius, 1.5),
      size: 1.2 + r() * 2,
      alpha: 0.25 + r() * 0.7,
      height: (r() - 0.5) * RING_RADIUS * 0.06,
    });
  }
  return pointGeometry(points);
}

/** The research-mode curve as a faint, static arc of points on the far side of the disk. */
function buildCurveGeometry(curve: SceneModel["curve"]): BufferGeometry | null {
  const n = curve.length;
  if (n === 0) return null;
  const points: PointSpec[] = [];
  curve.forEach((c, k) => {
    const angle = Math.PI * (1.15 + (0.7 * k) / Math.max(1, n - 1));
    const rows: [number, number, number][] = [
      [c.low, 1.3, 0.3],
      [c.mid, 2.2, 0.6],
      [c.high, 1.3, 0.3],
    ];
    for (const [v, size, alpha] of rows) {
      points.push({ radius: RING_RADIUS * (1.5 + 0.45 * clamp(v, 0, 1)), angle, speed: 0, size, alpha, height: 0 });
    }
  });
  return pointGeometry(points);
}

interface Arc {
  radius: number;
  start: number;
  tilt: number;
  speed: number;
  length: number;
}

function buildArcs(seed: number, n: number): Arc[] {
  const r = mulberry32(seed ^ 0x9e3779b9);
  return Array.from({ length: n }, (_, i) => ({
    radius: RING_RADIUS * (1.2 + (i % 3) * 0.088),
    start: (i / Math.max(1, n)) * TAU + (r() - 0.5) * 0.3,
    tilt: (r() - 0.5) * 0.16,
    speed: TAU / (60 + (i % 4) * 12 + r() * 8),
    length: 0.6 + r() * 0.3,
  }));
}

function pointMaterial(color: string) {
  const uniforms = {
    uTime: { value: 0 },
    uOpacity: { value: 0 },
    uPixelRatio: { value: 1 },
    uScale: { value: 14 },
    uColor: { value: new Color(color) },
  };
  const material = new ShaderMaterial({
    uniforms,
    vertexShader: POINT_VERT,
    fragmentShader: POINT_FRAG,
    transparent: true,
    depthWrite: false,
    blending: AdditiveBlending,
  });
  return { uniforms, material };
}

function bandMaterial(color: string, interval: number) {
  const thickness = RING_RADIUS * (0.072 + Math.min(0.32, interval * 0.96));
  const inner = RING_RADIUS - thickness / 2;
  const outer = RING_RADIUS + thickness / 2;
  const uniforms = {
    uInner: { value: inner },
    uOuter: { value: outer },
    uFeather: { value: thickness * 0.35 },
    uMid: { value: RING_RADIUS },
    uTime: { value: 0 },
    uOpacity: { value: 0 },
    uLine: { value: 0 },
    uColor: { value: new Color(color) },
  };
  const material = new ShaderMaterial({
    uniforms,
    vertexShader: PLANE_VERT,
    fragmentShader: BAND_FRAG,
    transparent: true,
    depthWrite: false,
    side: DoubleSide,
    blending: AdditiveBlending,
  });
  return { uniforms, material, inner, outer };
}

function glowMaterial() {
  const uniforms = {
    uRadius: { value: GLOW_RADIUS },
    uHole: { value: (HORIZON_RADIUS * 1.18) / GLOW_RADIUS },
    uGlow: { value: 0 },
    uOpacity: { value: 0 },
    uHot: { value: new Color(HOT) },
    uWarm: { value: new Color(WARM) },
    uCold: { value: new Color(COLD) },
  };
  const material = new ShaderMaterial({
    uniforms,
    vertexShader: PLANE_VERT,
    fragmentShader: GLOW_FRAG,
    transparent: true,
    depthWrite: false,
    side: DoubleSide,
    blending: AdditiveBlending,
  });
  return { uniforms, material };
}

/** A soft rim around the horizon, drawn once into a small canvas and shown as a sprite. */
function rimMaterial(): { material: SpriteMaterial | null; texture: CanvasTexture | null } {
  if (typeof document === "undefined") return { material: null, texture: null };
  const canvas = document.createElement("canvas");
  canvas.width = 256;
  canvas.height = 256;
  const ctx = canvas.getContext("2d");
  if (!ctx) return { material: null, texture: null };
  const g = ctx.createRadialGradient(128, 128, 0, 128, 128, 128);
  g.addColorStop(0, `rgba(${RIM_RGB}, 0)`);
  g.addColorStop(0.62, `rgba(${RIM_RGB}, 0)`);
  g.addColorStop(RIM_EDGE, `rgba(${RIM_RGB}, 0.55)`);
  g.addColorStop(0.73, `rgba(${RIM_RGB}, 0.3)`);
  g.addColorStop(0.86, `rgba(${RIM_RGB}, 0.06)`);
  g.addColorStop(1, `rgba(${RIM_RGB}, 0)`);
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, 256, 256);
  const texture = new CanvasTexture(canvas);
  const material = new SpriteMaterial({ map: texture, transparent: true, depthWrite: false, blending: AdditiveBlending, opacity: 0 });
  return { material, texture };
}

export function World({ model, palette, reduced, intro }: { model: SceneModel; palette: Palette; reduced: boolean; intro: RefObject<IntroControl> }) {
  const invalidate = useThree((s) => s.invalidate);
  const dpr = useThree((s) => s.viewport.dpr);

  const particleGeo = useMemo(() => buildParticleGeometry(model.seed), [model.seed]);
  const curveGeo = useMemo(() => buildCurveGeometry(model.curve), [model.curve]);
  const arcs = useMemo(() => buildArcs(model.seed, model.safeguards), [model.seed, model.safeguards]);
  const particles = useMemo(() => pointMaterial(palette.evidence), [palette.evidence]);
  const curve = useMemo(() => pointMaterial(palette.uncertainty), [palette.uncertainty]);
  const band = useMemo(() => bandMaterial(palette.evidence, model.interval), [palette.evidence, model.interval]);
  const glow = useMemo(() => glowMaterial(), []);
  const rim = useMemo(() => rimMaterial(), []);
  const arcMaterial = useMemo(
    () => new MeshBasicMaterial({ color: palette.safeguard, transparent: true, opacity: 0, depthWrite: false, side: DoubleSide }),
    [palette.safeguard],
  );
  const horizonMaterial = useMemo(() => new MeshBasicMaterial({ color: "#000000" }), []);
  const geometries = useMemo(
    () => ({
      horizon: new SphereGeometry(HORIZON_RADIUS, 48, 32),
      band: new RingGeometry(band.inner - 0.05, band.outer + 0.05, 192, 1),
      glow: new CircleGeometry(GLOW_RADIUS, 96),
      arcs: arcs.map((a) => new RingGeometry(a.radius, a.radius + 0.018, 64, 1, 0, a.length)),
    }),
    [band.inner, band.outer, arcs],
  );

  // GPU resources are owned here, so they are released here.
  useEffect(() => () => particleGeo.dispose(), [particleGeo]);
  useEffect(() => () => curveGeo?.dispose(), [curveGeo]);
  useEffect(() => () => particles.material.dispose(), [particles]);
  useEffect(() => () => curve.material.dispose(), [curve]);
  useEffect(() => () => band.material.dispose(), [band]);
  useEffect(() => () => glow.material.dispose(), [glow]);
  useEffect(() => () => arcMaterial.dispose(), [arcMaterial]);
  useEffect(() => () => horizonMaterial.dispose(), [horizonMaterial]);
  useEffect(
    () => () => {
      rim.material?.dispose();
      rim.texture?.dispose();
    },
    [rim],
  );
  useEffect(
    () => () => {
      geometries.horizon.dispose();
      geometries.band.dispose();
      geometries.glow.dispose();
      geometries.arcs.forEach((g) => g.dispose());
    },
    [geometries],
  );

  // The quality manager lowers the particle count by shrinking the draw range; no buffer is rebuilt.
  useEffect(() => {
    particleGeo.setDrawRange(0, model.particles);
    invalidate();
  }, [particleGeo, model.particles, invalidate]);

  const worldRef = useRef<Group>(null);
  const arcRefs = useRef<(Group | null)[]>([]);
  const clock = useRef({ t: 0, intro: reduced ? 1 : 0 });

  const apply = useCallback(
    (dt: number, pixelRatio: number) => {
      const c = clock.current;
      c.t += dt;
      if (reduced || intro.current.skipped) c.intro = 1;
      else c.intro = Math.min(1, c.intro + dt / INTRO_SECONDS);
      const e = easeOutCubic(c.intro);
      const t = reduced ? 0 : c.t;
      const bright = model.bright;

      glow.uniforms.uGlow.value = 0.35 + Math.min(0.55, bright * 0.625);
      glow.uniforms.uOpacity.value = e;

      band.uniforms.uTime.value = t;
      band.uniforms.uOpacity.value = (0.12 + 0.28 * bright) * e;
      band.uniforms.uLine.value = 0.7 * e;

      particles.uniforms.uTime.value = t;
      particles.uniforms.uPixelRatio.value = pixelRatio;
      particles.uniforms.uOpacity.value = (0.55 + 0.45 * bright) * e;

      curve.uniforms.uPixelRatio.value = pixelRatio;
      curve.uniforms.uOpacity.value = 0.6 * e;

      if (rim.material) rim.material.opacity = 0.9 * e;
      arcMaterial.opacity = 0.8 * e;
      arcs.forEach((a, i) => {
        const g = arcRefs.current[i];
        if (g) g.rotation.y = a.start + t * a.speed;
      });
      if (worldRef.current) worldRef.current.scale.setScalar(0.94 + 0.06 * e);
    },
    [reduced, intro, model.bright, glow, band, particles, curve, rim, arcMaterial, arcs],
  );

  useFrame((state, delta) => apply(Math.min(delta, MAX_DELTA), state.viewport.dpr));

  // A still frame for demand mode (reduced motion) and after any prop change.
  useEffect(() => {
    apply(0, dpr);
    invalidate();
  }, [apply, dpr, invalidate]);

  return (
    <group ref={worldRef}>
      <mesh geometry={geometries.horizon} material={horizonMaterial} />
      {rim.material ? <sprite material={rim.material} scale={[RIM_SCALE, RIM_SCALE, 1]} /> : null}
      {model.halo ? <mesh geometry={geometries.glow} material={glow.material} rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.02, 0]} /> : null}
      <mesh geometry={geometries.band} material={band.material} rotation={[-Math.PI / 2, 0, 0]} />
      <points geometry={particleGeo} material={particles.material} frustumCulled={false} />
      {curveGeo ? <points geometry={curveGeo} material={curve.material} frustumCulled={false} /> : null}
      {arcs.map((a, i) => (
        <group
          key={i}
          ref={(el) => {
            arcRefs.current[i] = el;
          }}
          rotation={[a.tilt, a.start, 0]}
        >
          <mesh geometry={geometries.arcs[i]} material={arcMaterial} rotation={[-Math.PI / 2, 0, 0]} />
        </group>
      ))}
    </group>
  );
}
