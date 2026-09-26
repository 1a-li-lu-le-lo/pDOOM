// Copyright NU Cybernetics. p(DOOM) — research prototype.
// Builds data/snapshots/<id> from the authored content modules. Every run is
// deterministic; the manifest is sealed afterwards with `pdoomctl snapshot seal`.
//
// Usage: node tools/snapshot/build.mjs [snapshot-id]
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { sources } from "./content/sources.mjs";
import { claims } from "./content/claims.mjs";
import { forecasts, benchmarks, benchmark_results, incidents } from "./content/forecasts.mjs";
import { scenarios, scenario_edges } from "./content/scenarios.mjs";
import { interventions, organizations, actions } from "./content/interventions.mjs";
import { definitions } from "./content/definitions.mjs";
import { drivers, driver_observations, model_spec } from "./content/drivers.mjs";

const id = process.argv[2] ?? "snap-2026-09-26-001";
const dir = join(process.cwd(), "data", "snapshots", id);
mkdirSync(dir, { recursive: true });

// Attach claim ids to sources.
const bySource = new Map();
for (const c of claims) {
  if (!bySource.has(c.source_id)) bySource.set(c.source_id, []);
  bySource.get(c.source_id).push(c.id);
}
for (const s of sources) s.claim_ids = (bySource.get(s.id) ?? []).sort();

// Referential integrity before writing.
const sourceIds = new Set(sources.map((s) => s.id));
const problems = [];
const check = (where, ids) => {
  for (const sid of ids ?? []) if (!sourceIds.has(sid)) problems.push(`${where}: unknown source ${sid}`);
};
for (const c of claims) check(c.id, [c.source_id]);
for (const f of forecasts) check(f.id, [f.source_id]);
for (const b of benchmarks) check(b.id, b.source_ids);
for (const r of benchmark_results) check(r.id, r.source_ids);
for (const i of incidents) check(i.id, i.source_ids);
for (const s of scenarios) check(s.id, s.source_ids);
for (const e of scenario_edges) check(e.id, e.source_ids);
for (const i of interventions) check(i.id, i.source_ids);
for (const o of organizations) check(o.id, o.source_ids);
for (const a of actions) for (const r of a.resources) if (r.source_id) check(a.id, [r.source_id]);
for (const d of definitions) for (const t of d.definitions) if (t.source_id) check(d.id, [t.source_id]);
for (const o of driver_observations) check(o.id, o.source_ids);
check("model_spec", model_spec.experimental_causal.source_ids);
if (problems.length) {
  console.error(problems.join("\n"));
  process.exit(1);
}

const write = (name, kind, items) =>
  writeFileSync(join(dir, name), JSON.stringify({ kind, schema_version: 1, items }, null, 2) + "\n");

write("definitions.json", "definition", definitions);
write("sources.json", "source", [...sources].sort((a, b) => a.id.localeCompare(b.id)));
write("claims.json", "claim", [...claims].sort((a, b) => a.id.localeCompare(b.id)));
write("forecasts.json", "forecast", forecasts);
write("benchmarks.json", "benchmark", benchmarks);
write("benchmark_results.json", "benchmark_result", benchmark_results);
write("incidents.json", "incident", incidents);
write("scenarios.json", "scenario", scenarios);
write("scenario_edges.json", "scenario_edge", scenario_edges);
write("drivers.json", "driver", drivers);
write("driver_observations.json", "driver_observation", driver_observations);
write("interventions.json", "intervention", interventions);
write("organizations.json", "organization", organizations);
write("actions.json", "action", actions);
write("model_spec.json", "model_spec", [model_spec]);

const manifest = {
  snapshot_id: id,
  created_at: "2026-09-26T00:00:00Z",
  source_cutoff: "2026-09-26",
  baseline_snapshot_id: null,
  files: [],
  notes: "First research-prototype snapshot. Records were authored from web-search-verified snippets (direct page fetches were blocked in the build environment) and prior knowledge; every record carries a verification object and only verified_search items with model_use_status eligible feed computations. Human review is pending for all records.",
};
writeFileSync(join(dir, "manifest.json"), JSON.stringify(manifest, null, 2) + "\n");
console.log(`wrote ${dir}: sources ${sources.length}, claims ${claims.length}, forecasts ${forecasts.length}, benchmarks ${benchmarks.length}/${benchmark_results.length}, incidents ${incidents.length}, scenarios ${scenarios.length}/${scenario_edges.length}, drivers ${drivers.length}/${driver_observations.length}, interventions ${interventions.length}, organizations ${organizations.length}, actions ${actions.length}, definitions ${definitions.length}`);
