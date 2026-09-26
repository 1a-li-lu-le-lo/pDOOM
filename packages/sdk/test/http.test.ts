// Copyright NU Cybernetics. p(DOOM) — research prototype.
// Integration test of the HTTP data source against the real Go API. It spawns
// cmd/pdoom-api on a free port with the repository's data directory and checks
// that the HTTP source returns the same objects as the file source. Skipped
// when Go is not installed (the TypeScript CI job); the Go job covers the API.
import { spawn, spawnSync, type ChildProcess } from "node:child_process";
import { createServer } from "node:net";
import { fileURLToPath } from "node:url";
import { afterAll, beforeAll, describe, expect, it } from "vitest";
import { createFileDataSource, createHttpDataSource } from "../src/index";

const repo = fileURLToPath(new URL("../../../", import.meta.url));
const hasGo = spawnSync("go", ["version"], { encoding: "utf8" }).status === 0;

async function freePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = createServer();
    srv.listen(0, "127.0.0.1", () => {
      const addr = srv.address();
      srv.close(() => (typeof addr === "object" && addr ? resolve(addr.port) : reject(new Error("no port"))));
    });
  });
}

describe.skipIf(!hasGo)("http data source against cmd/pdoom-api", () => {
  let child: ChildProcess | undefined;
  let base = "";

  beforeAll(async () => {
    const port = await freePort();
    base = `http://127.0.0.1:${port}`;
    child = spawn("go", ["run", "./cmd/pdoom-api", "--addr", `127.0.0.1:${port}`, "--data-dir", "data"], { cwd: repo, env: { ...process.env, GOTOOLCHAIN: "local" }, stdio: ["ignore", "ignore", "pipe"] });
    let stderr = "";
    child.stderr?.on("data", (d: Buffer) => (stderr += d.toString()));
    const deadline = Date.now() + 120_000;
    while (Date.now() < deadline) {
      try {
        const r = await fetch(`${base}/readyz`);
        if (r.ok) return;
      } catch {
        /* not up yet */
      }
      if (child.exitCode !== null) throw new Error(`pdoom-api exited: ${stderr}`);
      await new Promise((r) => setTimeout(r, 500));
    }
    throw new Error(`pdoom-api did not become ready: ${stderr}`);
  }, 130_000);

  afterAll(() => {
    child?.kill("SIGTERM");
  });

  it("returns the same release and snapshot objects as the file source", async () => {
    const http = createHttpDataSource({ baseUrl: base });
    const file = createFileDataSource({ dataDir: `${repo}data` });
    const [a, b] = await Promise.all([http.getRelease(), file.getRelease()]);
    expect(a.manifest).toEqual(b.manifest);
    expect(a.estimates).toEqual(b.estimates);
    expect(a.indexes).toEqual(b.indexes);
    expect(a.aggregations).toEqual(b.aggregations);
    expect(a.delta).toEqual(b.delta);
    expect(a.approvals).toEqual(b.approvals);
    const [sa, sb] = await Promise.all([http.getSnapshot(), file.getSnapshot()]);
    expect(sa.manifest.snapshot_id).toBe(sb.manifest.snapshot_id);
    for (const key of ["sources", "claims", "forecasts", "incidents", "scenarios", "scenario_edges", "drivers", "driver_observations", "interventions", "organizations", "actions", "definitions", "benchmarks", "benchmark_results"] as const) {
      expect(sa[key], key).toEqual(sb[key]);
    }
    expect(sa.model_spec).toEqual(sb.model_spec);
  }, 60_000);

  it("unwraps every list route and detail route", async () => {
    const http = createHttpDataSource({ baseUrl: base });
    const file = createFileDataSource({ dataDir: `${repo}data` });
    expect(await http.listReleases()).toEqual(await file.listReleases());
    expect(await http.getForecasts()).toEqual(await file.getForecasts());
    expect(await http.getBenchmarks()).toEqual(await file.getBenchmarks());
    expect(await http.getBenchmarkResults()).toEqual(await file.getBenchmarkResults());
    expect(await http.getIncidents()).toEqual(await file.getIncidents());
    expect(await http.getScenarios()).toEqual(await file.getScenarios());
    expect(await http.getScenarioEdges()).toEqual(await file.getScenarioEdges());
    expect(await http.getDrivers()).toEqual(await file.getDrivers());
    expect(await http.getDriverObservations()).toEqual(await file.getDriverObservations());
    expect(await http.getInterventions()).toEqual(await file.getInterventions());
    expect(await http.getOrganizations()).toEqual(await file.getOrganizations());
    expect(await http.getActions()).toEqual(await file.getActions());
    expect(await http.getDefinitions()).toEqual(await file.getDefinitions());
    expect((await http.getScenario("S1"))?.id).toBe("S1");
    expect(await http.getScenario("S99")).toBeUndefined();
    const sources = await http.getSources({ tier: 1, limit: 5 });
    expect(sources.items.length).toBeLessThanOrEqual(5);
    expect(sources.total).toBeGreaterThan(0);
    for (const s of sources.items) expect(s.source_tier).toBe(1);
    const one = await http.getSource(sources.items[0]!.id);
    expect(one?.id).toBe(sources.items[0]!.id);
    const docs = await http.getMethodology();
    expect(docs.map((d) => d.slug)).toEqual((await file.getMethodology()).map((d) => d.slug));
    expect((await http.getMethodologyDoc("model"))?.markdown).toMatch(/^# /m);
    expect(await http.getMethodologyDoc("nope")).toBeUndefined();
  }, 60_000);
});
