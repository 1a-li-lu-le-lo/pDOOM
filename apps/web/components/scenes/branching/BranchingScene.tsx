// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import type { StageData } from "../HomeStage";
import { StaticDisk } from "../StaticDisk";

export function BranchingScene({ data }: { data: StageData; reducedMotion: boolean }) {
  return <StaticDisk {...data} />;
}
