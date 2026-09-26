<!-- Copyright NU Cybernetics. PDUM — research prototype. -->
# Integration inventory (Phase 0 — repository discovery)

Date: 2026-09-26. Branch: `claude/pdum-risk-observatory-yq0pve`.

## Findings

The repository contained a single commit with a seven-byte `README.md` and nothing else.
The searches required by the product brief (Three.js scenes, shaders, particle systems,
instanced meshes, noise functions, post-processing, bloom, lensing, orbital mechanics,
camera rigs, texture generators, worker-based simulation, WebGPU experiments, asset
pipelines, performance budgets, frameworks, licensing, deployment configuration, data
infrastructure, design assets, safety gates) all returned no results.

| Component | Path | Purpose | API | Dependencies | Performance | Accessibility | License | Reuse decision | Required changes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| (none found) | — | — | — | — | — | — | — | Build new | — |

## Consequences

- No accretion-disk tooling exists to reuse; the Event Horizon scene is built new in
  `apps/web/components/scenes/event-horizon` on Three.js + React Three Fiber.
- No established production framework exists; Next.js 15 (App Router) is selected
  (see `docs/governance/decision-log.md`, ADR-001).
- No licensing constraints from prior code apply. The project license is a pending
  decision recorded in ADR-005; `NOTICE` carries the NU Cybernetics copyright.
- No deployment configuration exists; the prototype ships a Node server for the web app
  and Go binaries for the API and tooling (see `docs/operations/deployment.md`).
- No data infrastructure exists; the prototype uses versioned snapshot files with a
  reference PostgreSQL schema in `db/migrations` (ADR-003).

## Third-party libraries adopted (with license)

| Library | Version | License | Role |
| --- | --- | --- | --- |
| next / react / react-dom | 15.5 / 19.2 | MIT | web framework |
| three | 0.186 | MIT | WebGL renderer |
| @react-three/fiber, @react-three/drei | 9.8 / 10.7 | MIT | React bindings and helpers |
| d3-scale, d3-shape, d3-array | 4 / 3 / 3 | ISC | chart scales and paths (SVG rendered by React) |
| zod | 4 | MIT | schemas, JSON Schema export |
| @modelcontextprotocol/sdk | 1.30 | MIT | MCP server |
| vitest, @playwright/test, @axe-core/playwright | 3 / 1.63 / 4.13 | MIT / Apache-2.0 / MPL-2.0 | tests |
| github.com/santhosh-tekuri/jsonschema/v6 | 6.0.3 | Apache-2.0 | Go JSON Schema validation |
| github.com/temoto/robotstxt | 1.1.2 | MIT | robots.txt parsing |
| github.com/mmcdole/gofeed | 1.3.0 | MIT | RSS/Atom parsing |

All dependencies are pinned through `pnpm-lock.yaml` and `go.sum`.
