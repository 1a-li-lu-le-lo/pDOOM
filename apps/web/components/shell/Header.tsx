// Copyright NU Cybernetics. p(DOOM) — research prototype.
import Link from "next/link";
import { BRAND, PUBLIC_LABEL } from "@pdoom/schemas";
import { ModeSwitcher } from "../mode/ModeSwitcher";
import { Nav } from "./Nav";

/**
 * Two deliberate rows at every width: brand and mode switcher, then the
 * primary navigation (a single scrollable row on narrow screens). The mode
 * switcher is therefore always in the first viewport (build-spec §6).
 */
export function Header() {
  return (
    <header className="site-header">
      <div className="container">
        <div className="site-header-row">
          <Link href="/" className="brand" aria-label={`${BRAND} home`}>
            {BRAND}
            <small>{PUBLIC_LABEL}</small>
          </Link>
          <ModeSwitcher compact />
        </div>
        <Nav />
      </div>
    </header>
  );
}
