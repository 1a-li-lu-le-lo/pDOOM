// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import dynamic from "next/dynamic";
import { useEffect, useState } from "react";
import { useMode } from "../mode/ModeProvider";
import { StaticDisk, type StaticDiskProps } from "./StaticDisk";

const EventHorizon = dynamic(() => import("./event-horizon/EventHorizonScene").then((m) => m.EventHorizonScene), { ssr: false, loading: () => null });
const Orrery = dynamic(() => import("./orrery/OrreryScene").then((m) => m.OrreryScene), { ssr: false, loading: () => null });
const Branching = dynamic(() => import("./branching/BranchingScene").then((m) => m.BranchingScene), { ssr: false, loading: () => null });

export interface StageData extends StaticDiskProps {
  headlineText: string;
  intervalText: string;
  horizonLabel: string;
  scenarios: { id: string; name: string; recoverability: string }[];
  interventions: { id: string; name: string }[];
  indexes: { id: string; value: number | null }[];
  researchCurve: { x: string; low: number; mid: number; high: number }[];
}

function webglAvailable(): boolean {
  try {
    const c = document.createElement("canvas");
    return !!(c.getContext("webgl2") || c.getContext("webgl"));
  } catch {
    return false;
  }
}

/**
 * The stage behind the meter. Picks the scene for the selected mode, falls back
 * to the static disk when WebGL is unavailable or motion is reduced, and always
 * renders the static disk first so the page never waits on 3D assets.
 */
export function HomeStage({ data }: { data: StageData }) {
  const { mode, reducedMotion, hydrated } = useMode();
  const [webgl, setWebgl] = useState<boolean | null>(null);
  useEffect(() => setWebgl(webglAvailable()), []);
  const label = "Conceptual risk visualization — not a simulation of AI risk";
  if (!hydrated || mode === "observatory" || mode === "text") {
    return (
      <div className="hero-stage" aria-hidden="true">
        <StaticDisk {...data} />
      </div>
    );
  }
  if (mode === "orrery") return <div className="hero-stage" data-scene="orrery" aria-label={label}><Orrery data={data} reducedMotion={reducedMotion} /></div>;
  if (mode === "branching") return <div className="hero-stage" data-scene="branching" aria-label={label}><Branching data={data} reducedMotion={reducedMotion} /></div>;
  if (webgl === false || reducedMotion) {
    return (
      <div className="hero-stage" data-scene="static" aria-hidden="true">
        <StaticDisk {...data} />
      </div>
    );
  }
  return (
    <div className="hero-stage" data-scene="event-horizon" aria-label={label}>
      <StaticDisk {...data} />
      {webgl ? <EventHorizon data={data} /> : null}
    </div>
  );
}
