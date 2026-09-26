// Copyright NU Cybernetics. p(DOOM) — research prototype.
export const MODES = ["event-horizon", "orrery", "branching", "observatory", "text"] as const;
export type Mode = (typeof MODES)[number];
export const MODE_LABEL: Record<Mode, string> = {
  "event-horizon": "Event Horizon",
  orrery: "Orrery",
  branching: "Branching Futures",
  observatory: "Observatory",
  text: "Plain text",
};
export const MODE_STORAGE_KEY = "pdoom.mode";
export function isMode(v: unknown): v is Mode {
  return typeof v === "string" && (MODES as readonly string[]).includes(v);
}
