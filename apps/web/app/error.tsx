// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import Link from "next/link";

export default function ErrorPage({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <div className="page">
      <div className="container stack" style={{ maxWidth: "60ch" }}>
        <h1>Something did not render</h1>
        <p className="lede">The page failed before it could show its data. No estimate has changed; the release on disk is unaffected.</p>
        {error.digest ? <p className="cite">Reference {error.digest}</p> : null}
        <div className="row">
          <button type="button" className="btn btn-primary" onClick={reset}>
            Try again
          </button>
          <Link className="btn" href="/text">
            Read the plain-text version
          </Link>
        </div>
      </div>
    </div>
  );
}
