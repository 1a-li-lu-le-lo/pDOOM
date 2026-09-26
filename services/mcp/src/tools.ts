// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Tool definitions of the p(DOOM) MCP server.
 *
 * Every tool reads the promoted release and its sealed snapshot through
 * @pdoom/sdk; none of them computes or alters an estimate. Wherever a
 * probability appears the response carries its horizon, outcome set, status,
 * producing model version, data cutoff and limitations (build-spec §7).
 * The three submission tools append to the human review queue and nothing else.
 */
import { createHash, randomBytes } from "node:crypto";
import { appendFileSync, existsSync, mkdirSync } from "node:fs";
import { join } from "node:path";
import { z } from "zod";
import type { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { PDoomDataSource, Release } from "@pdoom/sdk";
import {
  BRAND,
  CorrectionSubmissionPayloadSchema,
  HORIZON_KEYS,
  Horizon,
  IncidentReferenceSubmissionPayloadSchema,
  OUTCOMES,
  SourceSubmissionPayloadSchema,
  USER_SCENARIO_MAX_SAMPLES,
  UserScenarioParamsSchema,
  type Estimate,
  type Submission,
} from "@pdoom/schemas";
import { evaluateUserScenario, describeUserScenario, roundingStep } from "@pdoom/model-core";

export interface ToolContext {
  source: PDoomDataSource;
  /** Where submissions are appended (file mode). */
  reviewQueueDir?: string;
  /** Where submissions are POSTed (http mode) when no queue dir is configured. */
  apiUrl?: string;
  fetchImpl?: typeof fetch;
  now?: () => Date;
}

const LIMITATIONS = [
  "The official p(DOOM) value is withheld (status insufficiently_calibrated); no headline probability is published.",
  "Indexes are 0–100 scores of movement and are never probabilities.",
  "External forecasts are aggregated only inside compatibility groups (same outcome set and horizon) and inherit their sources' selection and framing effects.",
  "Research-mode estimates come from an experimental causal model whose parameters are documented judgments.",
  "Nothing here predicts a date. Nothing here should be quoted without its horizon, outcome set and status.",
];

function meta(rel: Release) {
  const m = rel.manifest;
  return {
    brand: BRAND,
    release_id: m.release_id,
    data_snapshot: m.data_snapshot,
    data_cutoff: m.source_cutoff,
    published: m.published,
    model_versions: m.model_versions,
    editorial_risk_level: m.editorial_risk_level,
    uncertainty_score: m.uncertainty_score,
    limitations: LIMITATIONS,
  };
}

function estimateView(e: Estimate) {
  return {
    estimate_id: e.estimate_id,
    status: e.status,
    outcome_set: e.outcome_set,
    outcome_label: e.outcome_label,
    horizon: e.horizon,
    horizon_note: e.horizon_note,
    display: e.display,
    quantiles: e.quantiles,
    uncertainty: e.uncertainty,
    disagreement: e.disagreement,
    rounding_rule: e.rounding_rule,
    model_version: e.producer,
    forecast_origin_date: e.forecast_origin_date,
    last_evidence_date: e.last_evidence_date,
    conditioning: e.conditioning,
    assumptions: e.assumptions,
    source_coverage: e.source_coverage,
    method_ref: e.method_ref,
  };
}

function text(payload: unknown) {
  return { content: [{ type: "text" as const, text: JSON.stringify(payload, null, 2) }] };
}

function canonicalJSON(v: unknown): string {
  if (Array.isArray(v)) return `[${v.map(canonicalJSON).join(",")}]`;
  if (v && typeof v === "object") {
    const o = v as Record<string, unknown>;
    return `{${Object.keys(o)
      .sort()
      .filter((k) => o[k] !== undefined)
      .map((k) => `${JSON.stringify(k)}:${canonicalJSON(o[k])}`)
      .join(",")}}`;
  }
  return JSON.stringify(v);
}

let seq = 0;
function submissionId(kind: string, submittedAt: string, payload: unknown): string {
  // Same construction as internal/api/handlers_submit.go: a hash over kind,
  // timestamp, a per-process nonce and the canonical payload.
  const nonce = `${randomBytes(8).toString("hex")}-${++seq}`;
  return `sub-${createHash("sha256").update(`${kind}\n${submittedAt}\n${nonce}\n${canonicalJSON(payload)}`).digest("hex").slice(0, 16)}`;
}

async function submit(ctx: ToolContext, kind: Submission["kind"], payload: Record<string, unknown>, contact?: string) {
  const submittedAt = (ctx.now ?? (() => new Date()))().toISOString();
  const sub: Submission = { id: submissionId(kind, submittedAt, payload), kind, submitted_at: submittedAt, payload, status: "received", ...(contact ? { contact } : {}) };
  if (ctx.reviewQueueDir) {
    mkdirSync(ctx.reviewQueueDir, { recursive: true });
    appendFileSync(join(ctx.reviewQueueDir, "submissions.jsonl"), `${canonicalJSON(sub)}\n`, { flag: "a" });
    return { ...sub, stored_in: "review_queue_file", note: "Appended to the human review queue. Nothing published changes until a reviewer acts and a new release is promoted." };
  }
  if (ctx.apiUrl) {
    const route = kind === "source" ? "/v1/submissions/sources" : kind === "correction" ? "/v1/submissions/corrections" : null;
    if (!route) throw new Error("The public API accepts source and correction submissions only; incident references need a configured review queue directory (PDOOM_DATA_DIR).");
    const f = ctx.fetchImpl ?? fetch;
    const res = await f(`${ctx.apiUrl.replace(/\/$/, "")}${route}`, { method: "POST", headers: { "content-type": "application/json", accept: "application/json" }, body: JSON.stringify({ payload, ...(contact ? { contact } : {}) }) });
    if (!res.ok) throw new Error(`submission rejected: HTTP ${res.status}`);
    return { ...((await res.json()) as object), stored_in: "review_queue_api" };
  }
  throw new Error("No review queue configured: set PDOOM_DATA_DIR (file mode) or PDOOM_API_URL (http mode).");
}

const HorizonArg = z.enum(HORIZON_KEYS as unknown as [string, ...string[]]).describe("One of the seven horizons: 1y, 3y, 5y, 10y, 25y, 2100, eventual");

export function registerTools(server: McpServer, ctx: ToolContext) {
  const src = ctx.source;

  server.registerTool(
    "pdoom_get_meter",
    {
      title: "Current p(DOOM) meter",
      description:
        "The current state of the observatory for one horizon: the official object (withheld in this release line), external forecast aggregates, the research-mode model estimate with its decomposition, the six indexes and the editorial level. Every number carries its horizon, outcome set, status and model version.",
      inputSchema: { horizon: HorizonArg.optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ horizon }) => {
      const rel = await src.getRelease();
      const h = horizon ?? "10y";
      const official = rel.estimates.filter((e) => e.status === "insufficiently_calibrated" || e.status === "official").filter((e) => e.horizon === h);
      const external = rel.estimates.filter((e) => e.status === "external_aggregate" && e.horizon === h);
      const research = rel.estimates.filter((e) => e.status === "research_mode" && e.horizon === h);
      return text({
        meta: meta(rel),
        horizon: h,
        outcome_definition: rel.manifest.outcome_definition,
        official: official.map(estimateView),
        external_aggregates: external.map(estimateView),
        research_mode: research.map(estimateView),
        indexes: rel.indexes.map((i) => ({ index_id: i.index_id, label: i.label, value: i.value, scale: i.scale, is_probability: i.is_probability, coverage: i.coverage, note: i.note })),
        top_drivers: rel.drivers_explained.items.slice(0, 10),
        changes: rel.manifest.changes,
      });
    },
  );

  server.registerTool(
    "pdoom_get_estimate_history",
    {
      title: "Release history",
      description: "Every promoted release with its snapshot, model versions, editorial level and uncertainty score, plus the delta record of the current release (what moved and why).",
      inputSchema: {},
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async () => {
      const [rel, releases] = await Promise.all([src.getRelease(), src.listReleases()]);
      return text({ meta: meta(rel), releases, current_delta: rel.delta, changelog_markdown: rel.documents.changelog });
    },
  );

  server.registerTool(
    "pdoom_list_scenarios",
    {
      title: "Scenario map",
      description: "Category-level pathway scenarios (S1..S18) with outcome sets, recoverability, uncertainty, targeting safeguards and the directed relations between them.",
      inputSchema: { recoverability: z.enum(["high", "moderate", "low", "none", "unknown"]).optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ recoverability }) => {
      const [rel, scenarios, edges] = await Promise.all([src.getRelease(), src.getScenarios(), src.getScenarioEdges()]);
      const items = scenarios.filter((s) => (recoverability ? s.recoverability === recoverability : true));
      return text({ meta: meta(rel), scenarios: items.map((s) => ({ id: s.id, name: s.name, outcome_set: s.outcome_set, description: s.description, recoverability: s.recoverability, uncertainty: s.uncertainty, probability_source: s.probability_source, intervention_ids: s.intervention_ids, early_indicators: s.early_indicators, counterindicators: s.counterindicators })), edges });
    },
  );

  server.registerTool(
    "pdoom_get_scenario",
    {
      title: "Scenario detail",
      description: "Full record of one pathway scenario including prerequisites, indicators, control failures, evidence and open questions. Category level only; no operational detail exists in the data.",
      inputSchema: { id: z.string().regex(/^S\d{1,2}$/) },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ id }) => {
      const [rel, s, interventions] = await Promise.all([src.getRelease(), src.getScenario(id), src.getInterventions()]);
      if (!s) return { ...text({ error: `scenario ${id} not found` }), isError: true };
      return text({ meta: meta(rel), scenario: s, safeguards: interventions.filter((i) => i.target_scenario_ids.includes(s.id) || s.intervention_ids.includes(i.id)).map((i) => ({ id: i.id, name: i.name, evidence_strength: i.evidence_strength, category: i.category })) });
    },
  );

  server.registerTool(
    "pdoom_list_forecasts",
    {
      title: "External forecasts",
      description: "Every external forecast with its original question wording, population, outcome set, horizon and conditioning, and the aggregation results per compatibility group. Forecasts in different groups answer different questions and must not be compared directly.",
      inputSchema: { group_id: z.string().optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ group_id }) => {
      const [rel, forecasts] = await Promise.all([src.getRelease(), src.getForecasts()]);
      const fc = forecasts.filter((f) => (group_id ? f.group_id === group_id : true));
      const aggs = rel.aggregations.filter((a) => (group_id ? a.group_id === group_id : true));
      return text({ meta: meta(rel), forecasts: fc, aggregations: aggs });
    },
  );

  server.registerTool(
    "pdoom_list_incidents",
    {
      title: "Incident registry",
      description: "Verified incidents and near misses at category level (non-graphic, non-operational), with severity, relevance, evidence level, registry ids and sources.",
      inputSchema: { severity: z.string().optional(), near_miss: z.boolean().optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ severity, near_miss }) => {
      const [rel, incidents] = await Promise.all([src.getRelease(), src.getIncidents()]);
      const items = incidents.filter((i) => (severity ? i.severity === severity : true)).filter((i) => (near_miss === undefined ? true : i.near_miss === near_miss));
      const ipi = rel.indexes.find((i) => i.index_id === "incident_pressure");
      return text({ meta: meta(rel), incident_pressure_index: ipi ? { value: ipi.value, scale: ipi.scale, is_probability: false, note: ipi.note } : null, incidents: items });
    },
  );

  server.registerTool(
    "pdoom_list_safeguards",
    {
      title: "Safeguards",
      description: "Interventions with mechanism, evidence strength, cost, time to deploy, qualitative effect size, failure and backfire modes and the scenarios they target.",
      inputSchema: { category: z.enum(["technical", "organizational", "national", "international", "resilience"]).optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ category }) => {
      const [rel, interventions] = await Promise.all([src.getRelease(), src.getInterventions()]);
      const csi = rel.indexes.find((i) => i.index_id === "control_strength");
      return text({ meta: meta(rel), control_strength_index: csi ? { value: csi.value, scale: csi.scale, is_probability: false, note: csi.note } : null, safeguards: interventions.filter((i) => (category ? i.category === category : true)) });
    },
  );

  server.registerTool(
    "pdoom_list_sources",
    {
      title: "Source ledger",
      description: "Sources in the snapshot with tier (1–5), type, verification status, model-use status, licence and robots status. Tier 4–5 sources never feed the model.",
      inputSchema: { tier: z.number().int().min(1).max(5).optional(), topic: z.string().optional(), q: z.string().max(80).optional(), limit: z.number().int().min(1).max(200).optional(), offset: z.number().int().min(0).optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async (filter) => {
      const [rel, res] = await Promise.all([src.getRelease(), src.getSources(filter)]);
      return text({ meta: meta(rel), total: res.total, sources: res.items.map((s) => ({ id: s.id, title: s.title, publisher: s.publisher, authors: s.authors, date_published: s.date_published, canonical_url: s.canonical_url, source_tier: s.source_tier, source_type: s.source_type, topic: s.topic, verification: s.verification, model_use_status: s.model_use_status, conflicts: s.conflicts, license: s.license, citation: s.citation })) });
    },
  );

  server.registerTool(
    "pdoom_get_source",
    {
      title: "Source detail",
      description: "One source record with its evidence summary, limitations, counterevidence and every atomic claim extracted from it.",
      inputSchema: { id: z.string().regex(/^src-[a-z0-9-]+$/) },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ id }) => {
      const [rel, s, claims] = await Promise.all([src.getRelease(), src.getSource(id), src.getClaims()]);
      if (!s) return { ...text({ error: `source ${id} not found` }), isError: true };
      return text({ meta: meta(rel), source: s, claims: claims.filter((c) => c.source_id === id) });
    },
  );

  server.registerTool(
    "pdoom_get_methodology",
    {
      title: "Methodology",
      description: "The public methodology documents. Without a slug, lists them; with a slug, returns the Markdown of that document.",
      inputSchema: { slug: z.string().regex(/^[a-z0-9-]+$/).optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ slug }) => {
      const rel = await src.getRelease();
      if (!slug) return text({ meta: meta(rel), documents: await src.getMethodology(), model_card_markdown: rel.documents.model_card });
      const doc = await src.getMethodologyDoc(slug);
      if (!doc) return { ...text({ error: `methodology document ${slug} not found` }), isError: true };
      return { content: [{ type: "text" as const, text: doc.markdown }] };
    },
  );

  server.registerTool(
    "pdoom_get_definitions",
    {
      title: "Definitions",
      description: "Working definitions of every term (p(DOOM), the outcome taxonomy O0–O8, horizons, indexes) with the attributed definitions found in the literature and whether they are contested.",
      inputSchema: { term: z.string().max(80).optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ term }) => {
      const [rel, defs] = await Promise.all([src.getRelease(), src.getDefinitions()]);
      const q = term?.toLowerCase();
      return text({ meta: meta(rel), outcomes: OUTCOMES, horizons: HORIZON_KEYS, definitions: defs.filter((d) => (q ? d.term.toLowerCase().includes(q) || d.id.includes(q) : true)) });
    },
  );

  server.registerTool(
    "pdoom_list_actions",
    {
      title: "What people can do",
      description: "Concrete actions per audience linked to the safeguards they support, and the organisations register with funding and conflict disclosures.",
      inputSchema: { audience: z.string().optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ audience }) => {
      const [rel, actions, orgs] = await Promise.all([src.getRelease(), src.getActions(), src.getOrganizations()]);
      return text({ meta: meta(rel), actions: actions.filter((a) => (audience ? a.audience === audience : true)), organizations: orgs });
    },
  );

  server.registerTool(
    "pdoom_evaluate_user_scenario",
    {
      title: "Scenario Lab",
      description:
        "Runs the experimental research-mode causal model under caller-chosen assumptions (ten integer dials in -2..2 and a horizon). The result is labelled user_scenario, is NOT the p(DOOM) official model, and never changes any published value. Deterministic for a given seed.",
      inputSchema: { params: UserScenarioParamsSchema, samples: z.number().int().min(1).max(USER_SCENARIO_MAX_SAMPLES).optional() },
      annotations: { readOnlyHint: true, openWorldHint: false },
    },
    async ({ params, samples }) => {
      const [rel, spec] = await Promise.all([src.getRelease(), src.getModelSpec()]);
      const parsed = UserScenarioParamsSchema.parse(params);
      const result = evaluateUserScenario(spec.experimental_causal, parsed, samples);
      return text({ meta: meta(rel), label: result.label, disclaimer: result.disclaimer, sentence: describeUserScenario(result, roundingStep("high")), horizon: result.horizon, outcome_set: ["O3", "O4", "O5", "O6", "O7", "O8"], model_version: `pdoom-model/experimental-causal@${spec.experimental_causal.version}`, result });
    },
  );

  server.registerTool(
    "pdoom_submit_source",
    {
      title: "Submit a source",
      description: "Proposes a source for human review. Appends to the review queue only; it cannot change a snapshot, a release or any estimate.",
      inputSchema: { payload: SourceSubmissionPayloadSchema, contact: z.string().max(200).optional() },
      annotations: { readOnlyHint: false, destructiveHint: false, idempotentHint: false, openWorldHint: false },
    },
    async ({ payload, contact }) => text(await submit(ctx, "source", SourceSubmissionPayloadSchema.parse(payload), contact)),
  );

  server.registerTool(
    "pdoom_submit_correction",
    {
      title: "Submit a correction",
      description: "Reports an error in a published entity or estimate for human review. Appends to the review queue only.",
      inputSchema: { payload: CorrectionSubmissionPayloadSchema, contact: z.string().max(200).optional() },
      annotations: { readOnlyHint: false, destructiveHint: false, idempotentHint: false, openWorldHint: false },
    },
    async ({ payload, contact }) => text(await submit(ctx, "correction", CorrectionSubmissionPayloadSchema.parse(payload), contact)),
  );

  server.registerTool(
    "pdoom_submit_incident_reference",
    {
      title: "Submit an incident reference",
      description: "Points reviewers at an incident recorded in an external registry (AIID, OECD AIM, a docket). Summary must be non-graphic and non-operational. Appends to the review queue only.",
      inputSchema: { payload: IncidentReferenceSubmissionPayloadSchema, contact: z.string().max(200).optional() },
      annotations: { readOnlyHint: false, destructiveHint: false, idempotentHint: false, openWorldHint: false },
    },
    async ({ payload, contact }) => text(await submit(ctx, "incident_reference", IncidentReferenceSubmissionPayloadSchema.parse(payload), contact)),
  );

  return server;
}

export function resolveContext(env: NodeJS.ProcessEnv, source: PDoomDataSource): ToolContext {
  const dataDir = env.PDOOM_DATA_DIR;
  const queue = dataDir && existsSync(dataDir) ? join(dataDir, "review-queue") : undefined;
  return { source, reviewQueueDir: queue, apiUrl: env.PDOOM_API_URL };
}

export { Horizon };
