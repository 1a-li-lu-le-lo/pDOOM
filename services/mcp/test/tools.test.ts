// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { createHash } from "node:crypto";
import { mkdtempSync, readFileSync, readdirSync, statSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { afterAll, beforeAll, describe, expect, it } from "vitest";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { InMemoryTransport } from "@modelcontextprotocol/sdk/inMemory.js";
import { createFileDataSource } from "@pdoom/sdk";
import { SubmissionSchema } from "@pdoom/schemas";
import { registerTools } from "../src/tools";

const repo = fileURLToPath(new URL("../../../", import.meta.url));
const dataDir = join(repo, "data");

function hashTree(dir: string): string {
  const h = createHash("sha256");
  const walk = (d: string) => {
    for (const name of readdirSync(d).sort()) {
      const p = join(d, name);
      if (statSync(p).isDirectory()) walk(p);
      else h.update(name).update(readFileSync(p));
    }
  };
  walk(dir);
  return h.digest("hex");
}

function textOf(res: Awaited<ReturnType<Client["callTool"]>>): unknown {
  const c = (res.content as { type: string; text?: string }[])[0];
  return c?.type === "text" && c.text ? (c.text.startsWith("{") || c.text.startsWith("[") ? JSON.parse(c.text) : c.text) : null;
}

describe("p(DOOM) MCP tools", () => {
  const queueDir = mkdtempSync(join(tmpdir(), "pdoom-mcp-"));
  const before = { releases: hashTree(join(dataDir, "releases")), snapshots: hashTree(join(dataDir, "snapshots")) };
  const client = new Client({ name: "test", version: "0" });
  let server: McpServer;

  beforeAll(async () => {
    server = new McpServer({ name: "pdoom-test", version: "0" });
    registerTools(server, { source: createFileDataSource({ dataDir }), reviewQueueDir: queueDir, now: () => new Date("2026-09-26T12:00:00Z") });
    const [a, b] = InMemoryTransport.createLinkedPair();
    await Promise.all([server.connect(a), client.connect(b)]);
  });
  afterAll(async () => {
    await client.close();
    await server.close();
  });

  it("exposes the read tools and exactly three submission tools", async () => {
    const { tools } = await client.listTools();
    const names = tools.map((t) => t.name).sort();
    expect(names.filter((n) => n.startsWith("pdoom_submit_"))).toEqual(["pdoom_submit_correction", "pdoom_submit_incident_reference", "pdoom_submit_source"]);
    expect(names).toContain("pdoom_get_meter");
    expect(names).toContain("pdoom_evaluate_user_scenario");
    for (const t of tools) expect(t.name.startsWith("pdoom_")).toBe(true);
  });

  it("meter responses carry status, horizon, outcome set, model version, cutoff and limitations", async () => {
    const r = textOf(await client.callTool({ name: "pdoom_get_meter", arguments: { horizon: "10y" } })) as Record<string, unknown>;
    const meta = r.meta as Record<string, unknown>;
    expect(meta.release_id).toMatch(/^rel-/);
    expect(meta.data_cutoff).toMatch(/^\d{4}-\d{2}-\d{2}/);
    expect(Array.isArray(meta.limitations)).toBe(true);
    const official = r.official as Record<string, unknown>[];
    expect(official.length).toBeGreaterThan(0);
    for (const e of official) {
      expect(e.status).toBe("insufficiently_calibrated");
      expect(e.quantiles).toBeNull();
    }
    for (const e of [...(r.external_aggregates as Record<string, unknown>[]), ...(r.research_mode as Record<string, unknown>[])]) {
      expect(e.horizon).toBe("10y");
      expect(Array.isArray(e.outcome_set)).toBe(true);
      expect(String(e.model_version)).toMatch(/^pdoom-model\//);
    }
    for (const i of r.indexes as Record<string, unknown>[]) expect(i.is_probability).toBe(false);
  });

  it("scenario lab results are labelled user_scenario with the disclaimer", async () => {
    const params = { horizon: "10y", capability_timeline: 1, autonomy_growth: 0, access_level: 0, safety_progress: -1, governance_strength: 0, model_security: 0, open_weight_diffusion: 0, international_coordination: 0, incident_frequency: 0, resilience: 0 };
    const r = textOf(await client.callTool({ name: "pdoom_evaluate_user_scenario", arguments: { params, samples: 2000 } })) as Record<string, unknown>;
    expect(r.label).toBe("user_scenario");
    expect(String(r.disclaimer)).toContain("not the p(DOOM) official model");
    expect(String(r.sentence)).toMatch(/Under your selected assumptions/);
    const again = textOf(await client.callTool({ name: "pdoom_evaluate_user_scenario", arguments: { params, samples: 2000 } })) as Record<string, unknown>;
    expect(again.result).toEqual(r.result);
  });

  it("rejects out-of-range dials", async () => {
    const params = { horizon: "10y", capability_timeline: 3, autonomy_growth: 0, access_level: 0, safety_progress: 0, governance_strength: 0, model_security: 0, open_weight_diffusion: 0, international_coordination: 0, incident_frequency: 0, resilience: 0 };
    const res = await client.callTool({ name: "pdoom_evaluate_user_scenario", arguments: { params } });
    expect(res.isError).toBe(true);
  });

  it("submissions append one valid line to the review queue and touch nothing else", async () => {
    const r = textOf(await client.callTool({ name: "pdoom_submit_correction", arguments: { payload: { target_id: "src-iasr-2026", correction: "test correction", field: "title" }, contact: "tests" } })) as Record<string, unknown>;
    expect(r.status).toBe("received");
    expect(String(r.id)).toMatch(/^sub-[0-9a-f]{16}$/);
    await client.callTool({ name: "pdoom_submit_source", arguments: { payload: { canonical_url: "https://example.org/report", title: "Example", why_relevant: "test", claimed_evidence: "test" } } });
    await client.callTool({ name: "pdoom_submit_incident_reference", arguments: { payload: { registry: "aiid", external_id: "1", url: "https://example.org/incident/1", summary: "test" } } });
    const lines = readFileSync(join(queueDir, "submissions.jsonl"), "utf8").trim().split("\n");
    expect(lines).toHaveLength(3);
    for (const line of lines) expect(SubmissionSchema.safeParse(JSON.parse(line)).success).toBe(true);
    expect(hashTree(join(dataDir, "releases"))).toBe(before.releases);
    expect(hashTree(join(dataDir, "snapshots"))).toBe(before.snapshots);
  });

  it("methodology and definitions are served from the repository", async () => {
    const list = textOf(await client.callTool({ name: "pdoom_get_methodology", arguments: {} })) as Record<string, unknown>;
    expect((list.documents as unknown[]).length).toBeGreaterThan(5);
    const md = textOf(await client.callTool({ name: "pdoom_get_methodology", arguments: { slug: "model" } }));
    expect(String(md)).toMatch(/^# /m);
    const defs = textOf(await client.callTool({ name: "pdoom_get_definitions", arguments: { term: "p(DOOM)" } })) as Record<string, unknown>;
    expect((defs.definitions as unknown[]).length).toBeGreaterThan(0);
  });
});
