// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import { useEffect, useState } from "react";

export const THEME_KEY = "pdoom.theme";
export const PALETTE_KEY = "pdoom.palette";
type Theme = "system" | "dark" | "light";
type Palette = "default" | "cvd";

function apply(theme: Theme, palette: Palette) {
  const root = document.documentElement;
  if (theme === "system") delete root.dataset.theme;
  else root.dataset.theme = theme;
  if (palette === "default") delete root.dataset.palette;
  else root.dataset.palette = palette;
}

/**
 * Theme and palette controls. Stored locally only (no cookie, no account);
 * applied before first paint by the inline script in the root layout so the
 * page never flashes. "Colour-vision safe" switches the semantic palette
 * defined in the design tokens; shapes already back every colour.
 */
export function DisplaySettings() {
  const [theme, setTheme] = useState<Theme>("system");
  const [palette, setPalette] = useState<Palette>("default");
  useEffect(() => {
    try {
      const t = window.localStorage.getItem(THEME_KEY);
      const p = window.localStorage.getItem(PALETTE_KEY);
      if (t === "dark" || t === "light") setTheme(t);
      if (p === "cvd") setPalette(p);
    } catch {
      /* storage unavailable */
    }
  }, []);
  const update = (t: Theme, p: Palette) => {
    setTheme(t);
    setPalette(p);
    apply(t, p);
    try {
      window.localStorage.setItem(THEME_KEY, t);
      window.localStorage.setItem(PALETTE_KEY, p);
    } catch {
      /* storage unavailable: applies for this page only */
    }
  };
  return (
    <form className="display-settings" aria-label="Display settings" onSubmit={(e) => e.preventDefault()}>
      <label>
        Theme
        <select value={theme} onChange={(e) => update(e.target.value as Theme, palette)}>
          <option value="system">System</option>
          <option value="dark">Dark</option>
          <option value="light">Light</option>
        </select>
      </label>
      <label>
        Palette
        <select value={palette} onChange={(e) => update(theme, e.target.value as Palette)}>
          <option value="default">Default</option>
          <option value="cvd">Colour-vision safe</option>
        </select>
      </label>
    </form>
  );
}
