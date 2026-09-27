// Copyright NU Cybernetics. p(DOOM) — research prototype.
export default function Loading() {
  return (
    <div className="page" aria-busy="true" aria-live="polite">
      <div className="container stack">
        <div className="skeleton skeleton-title" />
        <div className="skeleton skeleton-line" />
        <div className="skeleton skeleton-line short" />
        <p className="cite">Reading the current release…</p>
      </div>
    </div>
  );
}
