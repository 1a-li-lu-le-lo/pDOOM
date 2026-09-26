// Copyright NU Cybernetics. p(DOOM) — research prototype.
import Link from "next/link";
import { AUTHOR, BRAND, BRAND_EXPANDED } from "@pdoom/schemas";

export function Footer({ releaseId, dataCutoff, published }: { releaseId?: string; dataCutoff?: string; published?: string | null }) {
  return (
    <footer className="site-footer">
      <div className="container">
        <div>
          <strong>{BRAND}</strong> — {BRAND_EXPANDED}.<br />
          Authored by {AUTHOR}. Copyright {AUTHOR}. Research prototype.
        </div>
        <div>
          {releaseId ? (
            <>
              Release <Link href={`/releases/${releaseId}`}>{releaseId}</Link>
              <br />
              Data through {dataCutoff ?? "—"} · Published {published ? published.slice(0, 10) : "—"}
            </>
          ) : (
            "No release loaded."
          )}
        </div>
        <div>
          <Link href="/method">Methodology</Link> · <Link href="/changelog">Changelog</Link> · <Link href="/text">Plain text</Link> ·{" "}
          <Link href="/method/content-safety">Content safety</Link>
          <br />
          No official probability is published in this release line. Indexes are not probabilities. Nothing here predicts a date.
        </div>
      </div>
    </footer>
  );
}
