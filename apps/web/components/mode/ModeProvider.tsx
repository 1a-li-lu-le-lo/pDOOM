// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { MODE_STORAGE_KEY, isMode, type Mode } from "./modes";

interface ModeContextValue {
  mode: Mode;
  setMode: (m: Mode) => void;
  reducedMotion: boolean;
  hydrated: boolean;
}

const ModeContext = createContext<ModeContextValue>({ mode: "observatory", setMode: () => {}, reducedMotion: false, hydrated: false });

/**
 * Stores the user's presentation mode locally (no account, no cookie). If the
 * system asks for reduced motion the default is the Observatory; otherwise the
 * Event Horizon. Nothing substantive depends on the mode: every mode reads the
 * same release data, and the plain-text route always exists.
 */
export function ModeProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<Mode>("observatory");
  const [reducedMotion, setReducedMotion] = useState(false);
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    setReducedMotion(mq.matches);
    let stored: string | null = null;
    try {
      stored = window.localStorage.getItem(MODE_STORAGE_KEY);
    } catch {
      stored = null;
    }
    if (isMode(stored)) setModeState(stored);
    else setModeState(mq.matches ? "observatory" : "event-horizon");
    setHydrated(true);
    const onChange = (e: MediaQueryListEvent) => setReducedMotion(e.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const setMode = useCallback((m: Mode) => {
    setModeState(m);
    try {
      window.localStorage.setItem(MODE_STORAGE_KEY, m);
    } catch {
      /* storage unavailable: the choice lives for this page only */
    }
  }, []);

  const value = useMemo(() => ({ mode, setMode, reducedMotion, hydrated }), [mode, setMode, reducedMotion, hydrated]);
  return <ModeContext.Provider value={value}>{children}</ModeContext.Provider>;
}

export function useMode() {
  return useContext(ModeContext);
}
