// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { MODES, MODE_LABEL, type Mode } from "./modes";
import { useMode } from "./ModeProvider";

/** Visible in the first viewport on every page: immersive modes, the observatory and plain text. */
export function ModeSwitcher({ compact = false }: { compact?: boolean }) {
  const { mode, setMode } = useMode();
  const pathname = usePathname();
  const router = useRouter();
  const onText = pathname === "/text";
  const visible: Mode[] = compact ? ["event-horizon", "observatory", "text"] : [...MODES];
  return (
    <div className="mode-switcher" role="group" aria-label="Presentation mode">
      {visible.map((m) =>
        m === "text" ? (
          <Link key={m} href="/text" aria-current={onText ? "page" : undefined}>
            {MODE_LABEL[m]}
          </Link>
        ) : (
          <button
            key={m}
            type="button"
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
