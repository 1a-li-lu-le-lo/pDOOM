// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { MetadataRoute } from "next";
import { getDataSource, getRelease, getSnapshot } from "@/lib/data";

const base = process.env.PDOOM_PUBLIC_URL ?? "http://localhost:3000";
const STATIC = ["", "/meter", "/futures", "/evidence", "/capabilities", "/agents", "/incidents", "/forecasts", "/safeguards", "/act", "/method", "/method/definitions", "/lab", "/changelog", "/compare", "/text"];

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  let lastModified: Date | undefined;
  const entries: MetadataRoute.Sitemap = [];
  try {
    const [rel, snap, docs, releases] = await Promise.all([getRelease(), getSnapshot(), getDataSource().getMethodology(), getDataSource().listReleases()]);
    lastModified = rel.manifest.published ? new Date(rel.manifest.published) : undefined;
    for (const s of snap.scenarios) entries.push({ url: `${base}/futures/${s.id}`, lastModified });
    for (const s of snap.sources) entries.push({ url: `${base}/evidence/sources/${s.id}`, lastModified, priority: 0.4 });
    for (const d of docs) entries.push({ url: `${base}/method/${d.slug}`, lastModified });
    for (const r of releases) entries.push({ url: `${base}/releases/${r.release_id}`, lastModified: r.published ? new Date(r.published) : undefined });
  } catch {
    /* no release: the static routes still exist */
  }
  return [...STATIC.map((p) => ({ url: `${base}${p}`, lastModified, priority: p === "" || p === "/meter" ? 1 : 0.7 })), ...entries];
}
