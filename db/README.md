<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Reference database schema

`migrations/0001_reference_schema.sql` is a PostgreSQL mirror of the snapshot and release
entities (ADR-010 in `docs/governance/decision-log.md`). The prototype does not run a database:
the API, web app and MCP server read `data/releases/CURRENT` and the sealed snapshot directly.

Apply with `psql -f db/migrations/0001_reference_schema.sql` and load the JSON files with any
`jsonb`-aware loader; column names equal the JSON field names in `packages/schemas`, and the
CHECK constraints repeat the rules the Go validator enforces (tier 4–5 sources never eligible,
the withheld official object publishes no number, quantiles are ordered).
