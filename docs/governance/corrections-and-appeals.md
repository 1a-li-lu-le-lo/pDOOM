<!-- Copyright NU Cybernetics. p(DOOM) — research prototype. -->
# Corrections and appeals

Release model cards (`data/releases/<id>/model-card.md`, rendered from
`internal/publishing/templates/model-card.md.tmpl`) point readers here. The full policy lives
in [`corrections-policy.md`](corrections-policy.md); this page is the short version.

- **Report an error or suggest a source**: `POST /v1/submissions/corrections` or
  `POST /v1/submissions/sources` on the public API ([`../api/README.md`](../api/README.md)).
  Submissions land in the human review queue and never change a published value directly.
- **What happens next**: a reviewer triages the submission; fixes enter a new snapshot; if a
  published number changes, a new signed release is promoted and the previous release is
  marked superseded. Releases are never edited in place.
- **Appeal a decision** (a tier, a label, a group membership, an exclusion): use the same route
  with the record id as `target_id`. The record is marked `disputed` until two reviewers who did
  not make the original decision agree, and the reasoning is published in the research report
  for that area.
- **See the record**: `/changelog`, `/releases/<id>`, `delta.json`, `changelog.md`, and the
  audit chain verified by `pdoomctl audit verify`.
