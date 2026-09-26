// Copyright NU Cybernetics. p(DOOM) — research prototype.
import Link from "next/link";
import type { ReactNode } from "react";

export interface Crumb {
  href: string;
  label: string;
}

/**
 * Every route opens with the same header: breadcrumbs, one h1, an optional
 * lede and a link to the same content in plain text (build-spec §6).
 */
export function PageHeader({ crumbs, title, lede, textAnchor, children }: { crumbs?: Crumb[]; title: ReactNode; lede?: ReactNode; textAnchor?: string; children?: ReactNode }) {
  return (
    <header className="page-header">
      <nav className="breadcrumbs" aria-label="Breadcrumb">
        <Link href="/">p(DOOM)</Link>
        {(crumbs ?? []).map((c) => (
          <span key={c.href}>
            {" / "}
            <Link href={c.href}>{c.label}</Link>
          </span>
        ))}
      </nav>
      <div className="page-header-row">
        <h1>{title}</h1>
        <Link className="btn text-link no-print" href={textAnchor ? `/text#${textAnchor}` : "/text"}>
          Plain text
        </Link>
      </div>
      {lede ? <p className="lede">{lede}</p> : null}
      {children}
    </header>
  );
}
