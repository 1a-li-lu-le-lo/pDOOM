// Copyright NU Cybernetics. p(DOOM) — research prototype.
import Link from "next/link";
import { BRAND, PUBLIC_LABEL } from "@pdoom/schemas";
import { ModeSwitcher } from "../mode/ModeSwitcher";
import { Nav } from "./Nav";

export function Header() {
  return (
    <header className="site-header">
      <div className="container">
        <Link href="/" className="brand" aria-label={`${BRAND} home`}>
          {BRAND}
          <small>{PUBLIC_LABEL}</small>
        </Link>
        <Nav />
        <ModeSwitcher compact />
      </div>
    </header>
  );
}
