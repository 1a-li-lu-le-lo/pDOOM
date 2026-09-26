// Copyright NU Cybernetics. p(DOOM) — research prototype.
// Server-side data access. The web app never computes an estimate: it reads the
// promoted release (data/releases/CURRENT) and its snapshot through @pdoom/sdk.
import { cache } from "react";
import { createFileDataSource, type PDoomDataSource, type Release, type Snapshot } from "@pdoom/sdk";
import type { Estimate, IndexValue } from "@pdoom/schemas";

let source: PDoomDataSource | null = null;

export function getDataSource(): PDoomDataSource {
  if (!source) source = createFileDataSource({});
  return source;
}

export const getRelease = cache(async (): Promise<Release> => getDataSource().getRelease());
export const getSnapshot = cache(async (): Promise<Snapshot> => getDataSource().getSnapshot());

export interface Headline {
  official: Estimate[];
  external: Estimate[];
  research: Estimate[];
}

export function headline(rel: Release): Headline {
  return {
    official: rel.estimates.filter((e) => e.status === "insufficiently_calibrated"),
    external: rel.estimates.filter((e) => e.status === "external_aggregate"),
    research: rel.estimates.filter((e) => e.status === "research_mode"),
  };
}

export function estimateById(rel: Release, id: string): Estimate | undefined {
  return rel.estimates.find((e) => e.estimate_id === id);
}

export function researchEstimate(rel: Release, key: string, horizon: string): Estimate | undefined {
  return estimateById(rel, `est-research-${key}-${horizon}`);
}

export function indexById(rel: Release, id: IndexValue["index_id"]): IndexValue | undefined {
  return rel.indexes.find((i) => i.index_id === id);
}

/** The horizon shown by default on the meter. */
export const DEFAULT_HORIZON = "10y";

export function isValidHorizon(h: string | undefined): h is string {
  return !!h && ["1y", "3y", "5y", "10y", "25y", "2100", "eventual"].includes(h);
}
