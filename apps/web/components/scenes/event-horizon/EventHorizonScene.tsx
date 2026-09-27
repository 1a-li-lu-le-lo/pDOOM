// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
/**
 * Event Horizon: the immersive stage behind the meter, rendered with React Three
 * Fiber over the static SVG disk. The disk stays underneath as the first frame
 * and as the fallback, and this canvas fades in over it. The scene is
 * conceptual (see World.tsx for the grammar) and never a simulation of AI risk;
 * it carries no text and no numbers, and everything substantive lives in /text.
 *
 * Runtime behaviour: a quality manager measures the first seconds and lowers
 * pixel ratio, particle count and the halo layer when frames are slow; the loop
 * pauses when the tab is hidden or the stage is scrolled out of view; the short
 * intro is skipped by any pointer, key or scroll; reduced motion renders a still
 * frame; a WebGL failure or a lost context unmounts the canvas quietly.
 */
import { Canvas, useFrame, useThree, type Frameloop } from "@react-three/fiber";
import {
  Component,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import type { StageData } from "../HomeStage";
import { World, deriveModel, type IntroControl, type Palette } from "./World";

const FULL_DPR: [number, number] = [1, 1.5];
const LOW_DPR = 1;
const LOW_PARTICLE_SCALE = 0.5;
/** Frames are ignored for this long after creation (shader compilation, layout). */
const WARMUP_SECONDS = 0.6;
const MEASURE_SECONDS = 2;
const MIN_FPS = 45;
const MEASURE_MAX_DELTA = 0.5;

const FALLBACK_PALETTE: Palette = {
  bg: "#050508",
  evidence: "#f5f5f7",
  safeguard: "#5fd6ea",
  uncertainty: "#e8b45a",
};

interface Quality {
  dpr: number | [number, number];
  particleScale: number;
  halo: boolean;
}

const FULL_QUALITY: Quality = { dpr: FULL_DPR, particleScale: 1, halo: true };
const LOW_QUALITY: Quality = { dpr: LOW_DPR, particleScale: LOW_PARTICLE_SCALE, halo: false };

/**
 * Scene colour tokens (--scene-*) from the document. The stage is dark in every
 * theme by design; the palette switch (colour-vision safe) still reaches the
 * canvas through these tokens.
 */
function readPalette(): Palette {
  if (typeof document === "undefined") return FALLBACK_PALETTE;
  try {
    const style = getComputedStyle(document.documentElement);
    const pick = (name: string, fallback: string) => {
      const v = style.getPropertyValue(name).trim();
      return /^#[0-9a-f]{6}$/i.test(v) ? v : fallback;
    };
    return {
      bg: pick("--scene-bg", FALLBACK_PALETTE.bg),
      evidence: pick("--scene-evidence", FALLBACK_PALETTE.evidence),
      safeguard: pick("--scene-safeguard", FALLBACK_PALETTE.safeguard),
      uncertainty: pick("--scene-uncertainty", FALLBACK_PALETTE.uncertainty),
    };
  } catch {
    return FALLBACK_PALETTE;
  }
}

function prefersReducedMotion(): boolean {
  try {
    return (
      typeof window !== "undefined" && window.matchMedia("(prefers-reduced-motion: reduce)").matches
    );
  } catch {
    return false;
  }
}

/** Renders nothing if anything inside the canvas throws, so the static disk beneath stays. */
class SceneBoundary extends Component<
  { children: ReactNode; onError: () => void },
  { failed: boolean }
> {
  override state = { failed: false };
  static getDerivedStateFromError() {
    return { failed: true };
  }
  override componentDidCatch() {
    this.props.onError();
  }
  override render() {
    return this.state.failed ? null : this.props.children;
  }
}

/** Measures frame time over the first couple of seconds and asks for lower quality once if it is slow. */
function QualityProbe({ onDegrade }: { onDegrade: () => void }) {
  const acc = useRef({ warm: 0, time: 0, frames: 0, done: false });
  useFrame((_, delta) => {
    const a = acc.current;
    if (a.done) return;
    const dt = Math.min(delta, MEASURE_MAX_DELTA);
    if (a.warm < WARMUP_SECONDS) {
      a.warm += dt;
      return;
    }
    a.time += dt;
    a.frames += 1;
    if (a.time >= MEASURE_SECONDS) {
      a.done = true;
      if (a.frames / a.time < MIN_FPS) onDegrade();
    }
  });
  return null;
}

/** Switching the frameloop back from "never" needs one invalidate to restart the loop. */
function LoopKicker() {
  const frameloop = useThree((s) => s.frameloop);
  const invalidate = useThree((s) => s.invalidate);
  useEffect(() => {
    if (frameloop !== "never") invalidate();
  }, [frameloop, invalidate]);
  return null;
}

export function EventHorizonScene({ data }: { data: StageData }) {
  const wrapRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const intro = useRef<IntroControl>({ skipped: false });
  const [failed, setFailed] = useState(false);
  const [ready, setReady] = useState(false);
  const [skipped, setSkipped] = useState(false);
  const [reduced, setReduced] = useState(prefersReducedMotion);
  const [inView, setInView] = useState(true);
  const [hidden, setHidden] = useState(false);
  const [quality, setQuality] = useState<Quality>(FULL_QUALITY);
  const [palette, setPalette] = useState(readPalette);
  const model = useMemo(
    () => deriveModel(data, quality.particleScale, quality.halo),
    [data, quality.particleScale, quality.halo],
  );

  // (d) reduced motion: still frame, no intro, and follow changes while mounted.
  useEffect(() => {
    let mq: MediaQueryList;
    try {
      mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    } catch {
      return;
    }
    const onChange = (e: MediaQueryListEvent) => setReduced(e.matches);
    setReduced(mq.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  // Palette: the display settings flip data-palette / data-theme on <html> at
  // runtime, so re-read the --scene-* tokens whenever those attributes change.
  useEffect(() => {
    if (typeof MutationObserver === "undefined") return;
    const root = document.documentElement;
    const mo = new MutationObserver(() => setPalette(readPalette()));
    mo.observe(root, { attributes: true, attributeFilter: ["data-palette", "data-theme"] });
    return () => mo.disconnect();
  }, []);

  // (b) pause when the document is hidden.
  useEffect(() => {
    const onVisibility = () => setHidden(document.visibilityState === "hidden");
    onVisibility();
    document.addEventListener("visibilitychange", onVisibility);
    return () => document.removeEventListener("visibilitychange", onVisibility);
  }, []);

  // (b) pause when the stage is scrolled out of view.
  useEffect(() => {
    const el = wrapRef.current;
    if (!el || typeof IntersectionObserver === "undefined") return;
    const io = new IntersectionObserver((entries) => {
      for (const entry of entries) setInView(entry.isIntersecting);
    });
    io.observe(el);
    return () => io.disconnect();
  }, []);

  // (c) any pointer, key or scroll skips the intro instantly; listeners leave once it is skipped.
  useEffect(() => {
    if (reduced || skipped) return;
    const skip = () => {
      if (intro.current.skipped) return;
      intro.current.skipped = true;
      setSkipped(true);
    };
    const opts: AddEventListenerOptions = { passive: true };
    const events = ["pointerdown", "keydown", "scroll", "wheel", "touchstart"] as const;
    for (const name of events) window.addEventListener(name, skip, opts);
    return () => {
      for (const name of events) window.removeEventListener(name, skip, opts);
    };
  }, [reduced, skipped]);

  // (e) a lost context unmounts the canvas; the static disk underneath remains.
  const onContextLost = useCallback((event: Event) => {
    event.preventDefault();
    setFailed(true);
  }, []);
  useEffect(
    () => () => {
      canvasRef.current?.removeEventListener("webglcontextlost", onContextLost);
    },
    [onContextLost],
  );

  const degrade = useCallback(() => setQuality(LOW_QUALITY), []);
  const fail = useCallback(() => setFailed(true), []);

  if (failed) return null;
  const frameloop: Frameloop = reduced ? "demand" : inView && !hidden ? "always" : "never";
  return (
    <div
      ref={wrapRef}
      className="scene-canvas"
      aria-hidden="true"
      data-ready={ready ? "true" : "false"}
      data-skipped={skipped ? "true" : "false"}
    >
      <SceneBoundary onError={fail}>
        <Canvas
          frameloop={frameloop}
          dpr={quality.dpr}
          flat
          gl={{
            antialias: false,
            alpha: false,
            stencil: false,
            depth: true,
            powerPreference: "high-performance",
            failIfMajorPerformanceCaveat: false,
          }}
          camera={{ position: [0, 3.33, 7.49], fov: 36, near: 0.5, far: 60 }}
          onCreated={(state) => {
            try {
              state.gl.setClearColor(palette.bg, 1);
              const el = state.gl.domElement;
              canvasRef.current = el;
              el.addEventListener("webglcontextlost", onContextLost, false);
              setReady(true);
            } catch {
              setFailed(true);
            }
          }}
        >
          <World model={model} palette={palette} reduced={reduced} intro={intro} />
          {!reduced && quality.particleScale === 1 ? <QualityProbe onDegrade={degrade} /> : null}
          <LoopKicker />
        </Canvas>
      </SceneBoundary>
    </div>
  );
}
