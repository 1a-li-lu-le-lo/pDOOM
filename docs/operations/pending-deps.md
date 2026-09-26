<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Pending dependencies

Build-spec §2 asks that any dependency a component needs but cannot install be recorded here so the orchestrator adds it through the lockfiles.

| Component | Dependency | Status |
| --- | --- | --- |
| all | none | Every dependency in use is pinned in `pnpm-lock.yaml` or `go.sum`; see `docs/architecture/integration-inventory.md`. |

Brand fonts (`Space Grotesk`, `IBM Plex Sans`, `IBM Plex Mono`) are named in `packages/design-system/fonts.css` but not bundled; the stacks fall back to system faces until font files with a compatible licence are added under `packages/design-system/fonts/`.
