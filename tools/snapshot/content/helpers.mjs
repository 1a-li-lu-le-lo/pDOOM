// Copyright NU Cybernetics. p(DOOM) — research prototype.
// Helpers for authoring snapshot content. Every record's verification is honest:
// "search" = a web-search result snippet quoting the figure was checked on CHECKED_AT
// (direct page fetches are blocked from the build environment); "prior" = drawn
// from prior knowledge and NOT fed to the model (informational only).
export const CHECKED_AT = "2026-09-26";
export const RETRIEVED = "2026-09-26";

export function verification(kind, note = "") {
  if (kind === "search") {
    return { status: "verified_search", checked_at: CHECKED_AT, method: "WebSearch result snippet quoting the figure/date; page fetch blocked by egress policy", note };
  }
  return { status: "verified_prior_knowledge", checked_at: CHECKED_AT, method: "prior knowledge, not re-fetched", note: note || "Informational only; excluded from model computations until fetched and reviewed." };
}

export function review(kind, useIfVerified = "eligible") {
  return {
    verification: verification(kind),
    human_review_status: "pending",
    model_use_status: kind === "search" ? useIfVerified : "informational",
  };
}

export function src(id, o) {
  return {
    id,
    canonical_url: o.url,
    title: o.title,
    publisher: o.publisher,
    authors: o.authors ?? [],
    date_published: o.date ?? null,
    date_updated: o.updated ?? null,
    date_retrieved: RETRIEVED,
    source_tier: o.tier,
    source_type: o.type,
    jurisdiction: o.jurisdiction ?? null,
    topic: o.topic ?? [],
    claim_ids: [],
    evidence_summary: o.summary,
    counterevidence: o.counter ?? null,
    methodology: o.methodology ?? null,
    sample: o.sample ?? null,
    limitations: o.limitations ?? null,
    conflicts: o.conflicts ?? ["none_known"],
    license: o.license ?? null,
    robots_status: o.robots ?? "unknown",
    content_hash: null,
    archive_reference: null,
    language: "en",
    translation: null,
    duplicate_group: null,
    retraction_status: "none",
    correction_status: null,
    citation: o.citation ?? `${o.publisher} (${(o.date ?? "n.d.").slice(0, 4)}). ${o.title}. ${o.url}`,
    ...review(o.verified ?? "prior", o.use ?? "eligible"),
  };
}

export function claim(id, o) {
  return {
    id,
    text: o.text,
    subject: o.subject,
    predicate: o.predicate,
    object: o.object,
    date: o.date ?? null,
    horizon: o.horizon ?? null,
    geography: o.geography ?? null,
    model_name: o.model ?? null,
    model_version: null,
    source_id: o.source,
    evidence_type: o.type,
    quantitative_value: o.value ?? null,
    unit: o.unit ?? null,
    uncertainty: o.uncertainty ?? null,
    direct_quote_pointer: o.pointer ?? null,
    context: o.context ?? "",
    corroboration_ids: o.corroboration ?? [],
    contradiction_ids: [],
    relevance: o.relevance ?? "moderate",
    status: o.status ?? "candidate",
    ...review(o.verified ?? "search"),
  };
}
