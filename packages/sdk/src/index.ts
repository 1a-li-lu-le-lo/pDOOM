// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * @pdoom/sdk — read-only access to the promoted release and its data snapshot.
 *
 * Two data sources implement the same interface: a file source (Node, reads
 * data/releases/CURRENT and data/snapshots/<id>) used by the web app's server
 * components and the MCP server, and an HTTP source that talks to pdoom-api.
 * Neither can write anything: every mutation path lives in pdoomctl.
 */
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { join, resolve } from "node:path";
import {
  ActionSchema,
  AggregationResultSchema,
  ApprovalSchema,
  BenchmarkResultSchema,
  BenchmarkSchema,
  ClaimSchema,
  DefinitionSchema,
  DeltaRecordSchema,
  DriverObservationSchema,
  DriverSchema,
  DriversExplainedSchema,
  EstimateSchema,
  ForecastSchema,
  IncidentSchema,
  IndexValueSchema,
  InterventionSchema,
  ModelSpecSchema,
  OrganizationSchema,
  ReleaseManifestSchema,
  ScenarioEdgeSchema,
  ScenarioSchema,
  SensitivityRunSchema,
  SnapshotManifestSchema,
  SourceSchema,
  type Action,
  type AggregationResult,
  type Approval,
  type Benchmark,
  type BenchmarkResult,
  type Claim,
  type Definition,
  type DeltaRecord,
  type Driver,
  type DriverObservation,
  type DriversExplained,
  type Estimate,
  type Forecast,
  type Incident,
  type IndexValue,
  type Intervention,
  type ModelSpec,
  type Organization,
  type ReleaseManifest,
  type Scenario,
  type ScenarioEdge,
  type SensitivityRun,
  type SnapshotManifest,
  type Source,
} from "@pdoom/schemas";
import type { z } from "zod";

export interface Release {
  manifest: ReleaseManifest;
  estimates: Estimate[];
  indexes: IndexValue[];
  aggregations: AggregationResult[];
  sensitivity: SensitivityRun[];
  delta: DeltaRecord;
  drivers_explained: DriversExplained;
  approvals: Approval[];
  /** Markdown files shipped with the release (changelog, model card). */
  documents: { changelog: string; model_card: string };
}

export interface Snapshot {
  manifest: SnapshotManifest;
  definitions: Definition[];
  sources: Source[];
  claims: Claim[];
  forecasts: Forecast[];
  benchmarks: Benchmark[];
  benchmark_results: BenchmarkResult[];
  incidents: Incident[];
  scenarios: Scenario[];
  scenario_edges: ScenarioEdge[];
  drivers: Driver[];
  driver_observations: DriverObservation[];
  interventions: Intervention[];
  organizations: Organization[];
  actions: Action[];
  model_spec: ModelSpec;
}

export interface ReleaseSummary {
  release_id: string;
  data_snapshot: string;
  model_versions: string[];
  published: string | null;
  superseded: { by: string; at: string } | null;
  is_current: boolean;
  editorial_risk_level: string;
  uncertainty_score: number | null;
}

export interface MethodologyDoc {
  slug: string;
  path: string;
  title: string;
}

export interface SourceFilter {
  tier?: number;
  topic?: string;
  q?: string;
  limit?: number;
  offset?: number;
}

export interface PDoomDataSource {
  getRelease(): Promise<Release>;
  getReleaseById(id: string): Promise<Release>;
  listReleases(): Promise<ReleaseSummary[]>;
  getSnapshot(): Promise<Snapshot>;
  getDefinitions(): Promise<Definition[]>;
  getSources(filter?: SourceFilter): Promise<{ items: Source[]; total: number }>;
  getSource(id: string): Promise<Source | undefined>;
  getClaims(): Promise<Claim[]>;
  getForecasts(): Promise<Forecast[]>;
  getBenchmarks(): Promise<Benchmark[]>;
  getBenchmarkResults(): Promise<BenchmarkResult[]>;
  getIncidents(): Promise<Incident[]>;
  getScenarios(): Promise<Scenario[]>;
  getScenario(id: string): Promise<Scenario | undefined>;
  getScenarioEdges(): Promise<ScenarioEdge[]>;
  getDrivers(): Promise<Driver[]>;
  getDriverObservations(): Promise<DriverObservation[]>;
  getInterventions(): Promise<Intervention[]>;
  getOrganizations(): Promise<Organization[]>;
  getActions(): Promise<Action[]>;
  getModelSpec(): Promise<ModelSpec>;
  getMethodology(): Promise<MethodologyDoc[]>;
  getMethodologyDoc(slug: string): Promise<{ doc: MethodologyDoc; markdown: string } | undefined>;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** PDOOM_DATA_DIR, else walk up from cwd looking for data/releases/CURRENT. */
export function resolveDataDir(start = process.cwd()): string {
  const env = process.env.PDOOM_DATA_DIR;
  if (env) return resolve(env);
  let dir = resolve(start);
  for (let i = 0; i < 8; i++) {
    const candidate = join(dir, "data");
    if (existsSync(join(candidate, "releases", "CURRENT")) || existsSync(join(candidate, "snapshots"))) return candidate;
    const parent = resolve(dir, "..");
    if (parent === dir) break;
    dir = parent;
  }
  return join(resolve(start), "data");
}

/** docs/method next to the data dir, unless PDOOM_DOCS_DIR is set. */
export function resolveDocsDir(dataDir: string): string {
  return process.env.PDOOM_DOCS_DIR ? resolve(process.env.PDOOM_DOCS_DIR) : join(dataDir, "..", "docs");
}

function parseEnvelope<T extends z.ZodType>(raw: string, schema: T, what: string): z.infer<T>[] {
  const doc = JSON.parse(raw) as { kind?: string; schema_version?: number; items?: unknown[] };
  if (!Array.isArray(doc.items)) throw new Error(`${what}: not an envelope`);
  return doc.items.map((item, i) => {
    const r = schema.safeParse(item);
    if (!r.success) throw new Error(`${what} item ${i}: ${r.error.issues[0]?.path.join(".")} ${r.error.issues[0]?.message}`);
    return r.data as z.infer<T>;
  });
}

function parseObject<T extends z.ZodType>(raw: string, schema: T, what: string): z.infer<T> {
  const r = schema.safeParse(JSON.parse(raw));
  if (!r.success) throw new Error(`${what}: ${r.error.issues[0]?.path.join(".")} ${r.error.issues[0]?.message}`);
  return r.data as z.infer<T>;
}

function titleOf(markdown: string, fallback: string): string {
  const m = markdown.match(/^#\s+(.+)$/m);
  return m?.[1]?.trim() ?? fallback;
}

// ---------------------------------------------------------------------------
// File data source
// ---------------------------------------------------------------------------

export interface FileDataSourceOptions {
  dataDir?: string;
  docsDir?: string;
  /** Re-read CURRENT at most every N ms (default 5000). */
  refreshMs?: number;
}

export function createFileDataSource(opts: FileDataSourceOptions = {}): PDoomDataSource {
  const dataDir = opts.dataDir ? resolve(opts.dataDir) : resolveDataDir();
  const docsDir = opts.docsDir ? resolve(opts.docsDir) : resolveDocsDir(dataDir);
  const refreshMs = opts.refreshMs ?? 5000;
  const releaseCache = new Map<string, Release>();
  const snapshotCache = new Map<string, Snapshot>();
  let currentId: string | null = null;
  let checkedAt = 0;

  const read = (p: string) => readFileSync(p, "utf8");

  const loadRelease = (id: string): Release => {
    const cached = releaseCache.get(id);
    if (cached) return cached;
    const dir = join(dataDir, "releases", id);
    if (!existsSync(join(dir, "manifest.json"))) throw new Error(`release ${id} not found`);
    const rel: Release = {
      manifest: parseObject(read(join(dir, "manifest.json")), ReleaseManifestSchema, `${id}/manifest.json`),
      estimates: parseEnvelope(read(join(dir, "estimates.json")), EstimateSchema, `${id}/estimates.json`),
      indexes: parseEnvelope(read(join(dir, "indexes.json")), IndexValueSchema, `${id}/indexes.json`),
      aggregations: parseEnvelope(read(join(dir, "aggregations.json")), AggregationResultSchema, `${id}/aggregations.json`),
      sensitivity: parseEnvelope(read(join(dir, "sensitivity.json")), SensitivityRunSchema, `${id}/sensitivity.json`),
      delta: parseObject(read(join(dir, "delta.json")), DeltaRecordSchema, `${id}/delta.json`),
      drivers_explained: parseObject(read(join(dir, "drivers_explained.json")), DriversExplainedSchema, `${id}/drivers_explained.json`),
      approvals: existsSync(join(dir, "approvals.json"))
        ? (JSON.parse(read(join(dir, "approvals.json"))) as unknown[]).map((a, i) => parseObject(JSON.stringify(a), ApprovalSchema, `${id}/approvals.json[${i}]`))
        : [],
      documents: {
        changelog: existsSync(join(dir, "changelog.md")) ? read(join(dir, "changelog.md")) : "",
        model_card: existsSync(join(dir, "model-card.md")) ? read(join(dir, "model-card.md")) : "",
      },
    };
    // A published release is immutable, so it is safe to cache forever.
    if (rel.manifest.published) releaseCache.set(id, rel);
    return rel;
  };

  const loadSnapshot = (id: string): Snapshot => {
    const cached = snapshotCache.get(id);
    if (cached) return cached;
    const dir = join(dataDir, "snapshots", id);
    const file = <T extends z.ZodType>(name: string, schema: T): z.infer<T>[] => {
      const p = join(dir, name);
      if (!existsSync(p)) return [];
      return parseEnvelope(read(p), schema, `${id}/${name}`);
    };
    const modelSpec = file("model_spec.json", ModelSpecSchema)[0];
    if (!modelSpec) throw new Error(`snapshot ${id}: model_spec.json missing`);
    const snap: Snapshot = {
      manifest: parseObject(read(join(dir, "manifest.json")), SnapshotManifestSchema, `${id}/manifest.json`),
      definitions: file("definitions.json", DefinitionSchema),
      sources: file("sources.json", SourceSchema),
      claims: file("claims.json", ClaimSchema),
      forecasts: file("forecasts.json", ForecastSchema),
      benchmarks: file("benchmarks.json", BenchmarkSchema),
      benchmark_results: file("benchmark_results.json", BenchmarkResultSchema),
      incidents: file("incidents.json", IncidentSchema),
      scenarios: file("scenarios.json", ScenarioSchema),
      scenario_edges: file("scenario_edges.json", ScenarioEdgeSchema),
      drivers: file("drivers.json", DriverSchema),
      driver_observations: file("driver_observations.json", DriverObservationSchema),
      interventions: file("interventions.json", InterventionSchema),
      organizations: file("organizations.json", OrganizationSchema),
      actions: file("actions.json", ActionSchema),
      model_spec: modelSpec,
    };
    snapshotCache.set(id, snap);
    return snap;
  };

  const current = (): string => {
    const now = Date.now();
    if (currentId && now - checkedAt < refreshMs) return currentId;
    const p = join(dataDir, "releases", "CURRENT");
    if (!existsSync(p)) throw new Error(`no current release: ${p} is missing`);
    currentId = read(p).trim();
    checkedAt = now;
    return currentId;
  };

  const snapshot = async (): Promise<Snapshot> => loadSnapshot(loadRelease(current()).manifest.data_snapshot);

  const listMethod = (): MethodologyDoc[] => {
    const dir = join(docsDir, "method");
    if (!existsSync(dir)) return [];
    return readdirSync(dir)
      .filter((f) => f.endsWith(".md"))
      .sort()
      .map((f) => {
        // `path` is repository-relative, as the API reports it; the file is read from docsDir.
        return { slug: f.replace(/\.md$/, ""), path: `docs/method/${f}`, title: titleOf(read(join(dir, f)), f) };
      });
  };

  return {
    async getRelease() {
      return loadRelease(current());
    },
    async getReleaseById(id) {
      if (!/^rel-\d{4}-\d{2}-\d{2}-\d{3}$/.test(id)) throw new Error(`invalid release id ${id}`);
      return loadRelease(id);
    },
    async listReleases() {
      const dir = join(dataDir, "releases");
      if (!existsSync(dir)) return [];
      const cur = existsSync(join(dir, "CURRENT")) ? current() : null;
      return readdirSync(dir)
        .filter((n) => /^rel-\d{4}-\d{2}-\d{2}-\d{3}$/.test(n) && statSync(join(dir, n)).isDirectory())
        .sort()
        .map((n) => {
          const m = loadRelease(n).manifest;
          return {
            release_id: n,
            data_snapshot: m.data_snapshot,
            model_versions: m.model_versions,
            published: m.published,
            superseded: m.superseded,
            is_current: n === cur,
            editorial_risk_level: m.editorial_risk_level,
            uncertainty_score: m.uncertainty_score,
          };
        });
    },
    getSnapshot: snapshot,
    async getDefinitions() {
      return (await snapshot()).definitions;
    },
    async getSources(filter = {}) {
      let items = (await snapshot()).sources;
      if (filter.tier) items = items.filter((s) => s.source_tier === filter.tier);
      if (filter.topic) items = items.filter((s) => s.topic.includes(filter.topic as string));
      if (filter.q) {
        const q = filter.q.toLowerCase();
        items = items.filter((s) => `${s.title} ${s.publisher} ${s.authors.join(" ")} ${s.evidence_summary}`.toLowerCase().includes(q));
      }
      const total = items.length;
      const offset = Math.max(0, filter.offset ?? 0);
      const limit = Math.min(Math.max(1, filter.limit ?? 200), 200);
      return { items: items.slice(offset, offset + limit), total };
    },
    async getSource(id) {
      return (await snapshot()).sources.find((s) => s.id === id);
    },
    async getClaims() {
      return (await snapshot()).claims;
    },
    async getForecasts() {
      return (await snapshot()).forecasts;
    },
    async getBenchmarks() {
      return (await snapshot()).benchmarks;
    },
    async getBenchmarkResults() {
      return (await snapshot()).benchmark_results;
    },
    async getIncidents() {
      return (await snapshot()).incidents;
    },
    async getScenarios() {
      return (await snapshot()).scenarios;
    },
    async getScenario(id) {
      return (await snapshot()).scenarios.find((s) => s.id === id);
    },
    async getScenarioEdges() {
      return (await snapshot()).scenario_edges;
    },
    async getDrivers() {
      return (await snapshot()).drivers;
    },
    async getDriverObservations() {
      return (await snapshot()).driver_observations;
    },
    async getInterventions() {
      return (await snapshot()).interventions;
    },
    async getOrganizations() {
      return (await snapshot()).organizations;
    },
    async getActions() {
      return (await snapshot()).actions;
    },
    async getModelSpec() {
      return (await snapshot()).model_spec;
    },
    async getMethodology() {
      return listMethod();
    },
    async getMethodologyDoc(slug) {
      if (!/^[a-z0-9-]+$/.test(slug)) return undefined;
      const doc = listMethod().find((d) => d.slug === slug);
      if (!doc) return undefined;
      return { doc, markdown: read(join(docsDir, "method", `${slug}.md`)) };
    },
  };
}

// ---------------------------------------------------------------------------
// HTTP data source (pdoom-api)
// ---------------------------------------------------------------------------

export function createHttpDataSource(opts: { baseUrl: string; fetchImpl?: typeof fetch }): PDoomDataSource {
  // Talks to cmd/pdoom-api. Route shapes follow api/openapi.yaml: every data
  // response wraps its payload beside a `meta` object; the SDK unwraps it so
  // callers see the same objects the file source returns.
  const base = opts.baseUrl.replace(/\/$/, "");
  const f = opts.fetchImpl ?? fetch;
  const get = async <T>(path: string): Promise<T> => {
    const res = await f(`${base}${path}`, { headers: { accept: "application/json" } });
    if (!res.ok) throw new Error(`${path}: HTTP ${res.status}`);
    return (await res.json()) as T;
  };
  const getText = async (path: string): Promise<string | undefined> => {
    const res = await f(`${base}${path}`, { headers: { accept: "text/markdown" } });
    if (!res.ok) return undefined;
    return res.text();
  };
  let snapshotCache: { id: string; snap: Snapshot } | null = null;
  const releaseById = async (id: string): Promise<Release> => {
    if (!/^rel-\d{4}-\d{2}-\d{2}-\d{3}$/.test(id)) throw new Error(`invalid release id: ${id}`);
    const body = await get<{ release: Omit<Release, "documents">; is_current: boolean }>(`/v1/releases/${encodeURIComponent(id)}`);
    const changelog = (await getText(`/v1/releases/${encodeURIComponent(id)}/changelog.md`).catch(() => undefined)) ?? "";
    const modelCard = (await getText(`/v1/releases/${encodeURIComponent(id)}/model-card.md`).catch(() => undefined)) ?? "";
    return { ...body.release, documents: { changelog, model_card: modelCard } };
  };
  const currentId = async () => (await get<{ current: string }>("/v1/releases")).current;
  const snapshot = async (): Promise<Snapshot> => {
    const body = await get<{ meta: { data_snapshot: string }; snapshot: Snapshot }>("/v1/snapshot");
    if (snapshotCache && snapshotCache.id === body.meta.data_snapshot) return snapshotCache.snap;
    snapshotCache = { id: body.meta.data_snapshot, snap: body.snapshot };
    return body.snapshot;
  };
  return {
    getRelease: async () => releaseById(await currentId()),
    getReleaseById: releaseById,
    listReleases: async () => (await get<{ releases: ReleaseSummary[] }>("/v1/releases")).releases,
    getSnapshot: snapshot,
    getDefinitions: async () => (await get<{ definitions: Definition[] }>("/v1/definitions")).definitions,
    getSources: async (filter = {}) => {
      const qs = new URLSearchParams();
      for (const [k, v] of Object.entries(filter)) if (v !== undefined) qs.set(k, String(v));
      const body = await get<{ total: number; sources: Source[] }>(`/v1/sources${qs.size ? `?${qs.toString()}` : ""}`);
      return { items: body.sources, total: body.total };
    },
    getSource: (id) =>
      get<{ source: Source }>(`/v1/sources/${encodeURIComponent(id)}`)
        .then((b) => b.source)
        .catch(() => undefined),
    getClaims: async () => (await snapshot()).claims,
    getForecasts: async () => (await get<{ forecasts: Forecast[] }>("/v1/forecasts")).forecasts,
    getBenchmarks: async () => (await get<{ benchmarks: Benchmark[] }>("/v1/capabilities")).benchmarks,
    getBenchmarkResults: async () => (await get<{ results: BenchmarkResult[] }>("/v1/capabilities")).results,
    getIncidents: async () => (await get<{ incidents: Incident[] }>("/v1/incidents")).incidents,
    getScenarios: async () => (await get<{ scenarios: Scenario[] }>("/v1/scenarios")).scenarios,
    getScenario: (id) =>
      get<{ scenario: Scenario }>(`/v1/scenarios/${encodeURIComponent(id)}`)
        .then((b) => b.scenario)
        .catch(() => undefined),
    getScenarioEdges: async () => (await get<{ edges: ScenarioEdge[] }>("/v1/scenarios")).edges,
    getDrivers: async () => (await get<{ drivers: Driver[] }>("/v1/drivers")).drivers,
    getDriverObservations: async () => (await get<{ observations: DriverObservation[] }>("/v1/drivers")).observations,
    getInterventions: async () => (await get<{ interventions: Intervention[] }>("/v1/safeguards")).interventions,
    getOrganizations: async () => (await get<{ organizations: Organization[] }>("/v1/organizations")).organizations,
    getActions: async () => (await get<{ actions: Action[] }>("/v1/actions")).actions,
    getModelSpec: async () => (await snapshot()).model_spec,
    getMethodology: async () => (await get<{ documents: MethodologyDoc[] }>("/v1/methodology")).documents,
    getMethodologyDoc: async (slug) => {
      if (!/^[a-z0-9-]+$/.test(slug)) return undefined;
      const doc = (await get<{ documents: MethodologyDoc[] }>("/v1/methodology")).documents.find((d) => d.slug === slug);
      if (!doc) return undefined;
      const markdown = await getText(`/v1/methodology/${encodeURIComponent(slug)}`);
      return markdown === undefined ? undefined : { doc, markdown };
    },
  };
}

/** Convenience: file source when a data dir is reachable, else HTTP (PDOOM_API_URL). */
export function createDefaultDataSource(): PDoomDataSource {
  const api = process.env.PDOOM_API_URL;
  const dataDir = resolveDataDir();
  if (existsSync(join(dataDir, "releases", "CURRENT"))) return createFileDataSource({ dataDir });
  if (api) return createHttpDataSource({ baseUrl: api });
  return createFileDataSource({ dataDir });
}
