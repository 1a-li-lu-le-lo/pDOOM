// Copyright NU Cybernetics. p(DOOM) — research prototype.
// The share card. It carries the official status, the horizon and the
// research-mode interval, never a bare number: a screenshot of this card is
// still an honest statement.
import { ImageResponse } from "next/og";
import { BRAND, PUBLIC_LABEL, TAGLINES } from "@pdoom/schemas";
import { DEFAULT_HORIZON, getRelease, headline, indexById, researchEstimate } from "@/lib/data";
import { horizonLabel } from "@/lib/format";

export const alt = `${BRAND} — ${PUBLIC_LABEL}`;
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

const GRADIENT_STOPS: [string, number][] = [
  ["#07070b", 0],
  ["#141225", 60],
  ["#2a1a12", 100],
];
const background = `linear-gradient(160deg, ${GRADIENT_STOPS.map(([c, at]) => `${c} ${at}%`).join(", ")})`;

export default async function Image() {
  let official = "No release";
  let interval = "";
  let unc = "";
  let releaseId = "";
  try {
    const rel = await getRelease();
    const h = headline(rel);
    official = h.official.find((e) => e.horizon === DEFAULT_HORIZON)?.display.central ?? official;
    const r = researchEstimate(rel, "P_DOOM", DEFAULT_HORIZON);
    interval = r ? `Research-mode model, ${horizonLabel(DEFAULT_HORIZON)}: median ${r.display.central}, plausible interval ${r.display.interval}` : "";
    const u = indexById(rel, "uncertainty");
    unc = u?.value != null ? `Uncertainty ${Math.round(u.value)} / 100` : "";
    releaseId = rel.manifest.release_id;
  } catch {
    /* render the brand alone */
  }
  return new ImageResponse(
    (
      <div style={{ width: "100%", height: "100%", display: "flex", flexDirection: "column", justifyContent: "space-between", padding: 64, background, color: "#f5f5f7", fontFamily: "sans-serif" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <div style={{ fontSize: 56, fontWeight: 700, letterSpacing: -2 }}>{BRAND}</div>
          <div style={{ fontSize: 22, color: "#a9a9b8" }}>{PUBLIC_LABEL}</div>
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
          <div style={{ fontSize: 30, color: "#a9a9b8" }}>{`Official value · ${horizonLabel(DEFAULT_HORIZON)} · O3–O8 combined`}</div>
          <div style={{ fontSize: 72, fontWeight: 700, color: "#b7b7c9", letterSpacing: -2 }}>{official}</div>
          <div style={{ fontSize: 26, color: "#e6d3c3", maxWidth: 1000 }}>{interval}</div>
        </div>
        <div style={{ display: "flex", justifyContent: "space-between", fontSize: 22, color: "#a9a9b8" }}>
          <div>{TAGLINES[0]}</div>
          <div>{`${unc}${releaseId ? ` · ${releaseId}` : ""}`}</div>
        </div>
      </div>
    ),
    { ...size },
  );
}
