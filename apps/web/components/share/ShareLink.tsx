// Copyright NU Cybernetics. p(DOOM) — research prototype.
"use client";
import { useEffect, useState } from "react";

/**
 * Copies the current URL (or a given one) to the clipboard. Sharing here means
 * sharing a reproducible view of the same signed release, never a screenshot of
 * a number without its context. No tracking parameters are added.
 */
export function ShareLink({ href, label = "Copy link" }: { href?: string; label?: string }) {
  const [state, setState] = useState<"idle" | "copied" | "failed">("idle");
  const [canShare, setCanShare] = useState(false);
  useEffect(() => setCanShare(typeof navigator !== "undefined" && typeof navigator.share === "function"), []);
  const url = () => (href ? new URL(href, window.location.origin).toString() : window.location.href);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(url());
      setState("copied");
    } catch {
      setState("failed");
    }
    window.setTimeout(() => setState("idle"), 2200);
  };
  const share = async () => {
    try {
      await navigator.share({ title: document.title, url: url() });
    } catch {
      /* the user dismissed the sheet */
    }
  };
  return (
    <span className="row" role="group" aria-label="Share">
      <button type="button" className="btn" onClick={copy} aria-live="polite">
        {state === "copied" ? "Link copied" : state === "failed" ? "Copy failed — select the address bar" : label}
      </button>
      {canShare ? (
        <button type="button" className="btn" onClick={share}>
          Share…
        </button>
      ) : null}
    </span>
  );
}
