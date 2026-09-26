// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";

export const NAV = [
  ["/meter", "Meter"],
  ["/futures", "Futures"],
  ["/evidence", "Evidence"],
  ["/capabilities", "Capabilities"],
  ["/agents", "Agents"],
  ["/incidents", "Incidents"],
  ["/forecasts", "Forecasts"],
  ["/safeguards", "Safeguards"],
  ["/act", "Act"],
  ["/method", "Method"],
  ["/text", "Text"],
] as const;

export function Nav() {
  const pathname = usePathname();
  return (
    <nav className="site-nav" aria-label="Primary">
      {NAV.map(([href, label]) => (
        <Link key={href} href={href} aria-current={pathname === href || pathname.startsWith(href + "/") ? "page" : undefined}>
          {label}
        </Link>
      ))}
    </nav>
  );
}
