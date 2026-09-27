// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { MODES, MODE_LABEL, type Mode } from "./modes";
import { useMode } from "./ModeProvider";

/**
 * Visible in the first viewport on every page. All five modes are offered;
 * in the compact header the two secondary immersive modes are hidden on
 * narrow screens by CSS (they remain on the home page's full switcher).
 */
export function ModeSwitcher({ compact = false }: { compact?: boolean }) {
  const { mode, setMode } = useMode();
  const pathname = usePathname();
  const router = useRouter();
  const onText = pathname === "/text";
  const secondary: Mode[] = ["orrery", "branching"];
  return (
    <div className={`mode-switcher${compact ? " mode-switcher-compact" : ""}`} role="group" aria-label="Presentation mode">
      {MODES.map((m) =>
        m === "text" ? (
          <Link key={m} href="/text" aria-current={onText ? "page" : undefined}>
            {MODE_LABEL[m]}
          </Link>
        ) : (
          <button
            key={m}
            type="button"
            className={compact && secondary.includes(m) ? "mode-optional" : undefined}
            aria-pressed={!onText && mode === m}
            onClick={() => {
              setMode(m);
              if (pathname !== "/") router.push("/");
            }}
          >
            {MODE_LABEL[m]}
          </button>
        ),
      )}
    </div>
  );
}
