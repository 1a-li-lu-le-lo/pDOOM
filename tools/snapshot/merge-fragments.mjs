// Copyright NU Cybernetics. PDUM — research prototype.
// Merge research fragment files (research/*/fragments/{sources,claims}.json) into a
// snapshot's sources.json / claims.json, deduplicating sources by canonical URL and
// rewriting every remapped id across all snapshot entity files.
//
// Usage: node tools/snapshot/merge-fragments.mjs <snapshot-dir> [--dry-run]
import { readFileSync, writeFileSync, existsSync, readdirSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const [, , snapDirArg, ...flags] = process.argv;
if (!snapDirArg) {
  console.error("usage: merge-fragments.mjs <snapshot-dir> [--dry-run]");
  process.exit(2);
}
const dryRun = flags.includes("--dry-run");
const root = process.cwd();
const snapDir = join(root, snapDirArg);
const researchDir = join(root, "research");

const ENTITY_FILES = [
  "definitions.json", "sources.json", "claims.json", "forecasts.json", "benchmarks.json",
  "benchmark_results.json", "incidents.json", "scenarios.json", "scenario_edges.json",
  "drivers.json", "driver_observations.json", "interventions.json", "organizations.json",
  "actions.json", "model_spec.json",
];

function readJSON(p) {
  return JSON.parse(readFileSync(p, "utf8"));
}
function writeJSON(p, v) {
  if (dryRun) return;
  writeFileSync(p, JSON.stringify(v, null, 2) + "\n");
}
function normalizeUrl(u) {
  if (typeof u !== "string") return "";
  try {
    const url = new URL(u.trim());
    url.hash = "";
    url.hostname = url.hostname.toLowerCase().replace(/^www\./, "");
    // arXiv: unify abs/pdf and version suffixes
    if (url.hostname === "arxiv.org") {
      url.pathname = url.pathname.replace(/^\/pdf\//, "/abs/").replace(/v\d+$/, "").replace(/\.pdf$/, "");
    }
    // doi.org: lowercase path
    if (url.hostname === "doi.org") url.pathname = url.pathname.toLowerCase();
    let s = url.toString();
    if (s.endsWith("/")) s = s.slice(0, -1);
    return s.replace(/^http:\/\//, "https://");
  } catch {
    return u.trim().toLowerCase();
  }
}

// 1. Collect fragments.
const fragmentFiles = [];
function walk(dir) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    const st = statSync(p);
    if (st.isDirectory()) walk(p);
    else if (/fragments[\\/](sources|claims)\.json$/.test(p)) fragmentFiles.push(p);
  }
}
if (existsSync(researchDir)) walk(researchDir);

const sources = new Map(); // finalId -> item
const urlToId = new Map(); // normalized url -> finalId
const idRemap = new Map(); // originalId -> finalId
const claims = new Map(); // id -> item
const problems = [];

// Seed with existing snapshot sources/claims so re-runs are idempotent.
for (const [file, store] of [["sources.json", sources], ["claims.json", claims]]) {
  const p = join(snapDir, file);
  if (existsSync(p)) {
    const env = readJSON(p);
    for (const it of env.items ?? []) {
      store.set(it.id, it);
      if (file === "sources.json") urlToId.set(normalizeUrl(it.canonical_url), it.id);
    }
  }
}

const perFragmentStats = [];
for (const f of fragmentFiles.sort()) {
  let env;
  try {
    env = readJSON(f);
  } catch (e) {
    problems.push(`${relative(root, f)}: invalid JSON (${e.message})`);
    continue;
  }
  const items = Array.isArray(env) ? env : env.items ?? [];
  const isSource = /sources\.json$/.test(f);
  let added = 0, merged = 0;
  for (const it of items) {
    if (!it || typeof it.id !== "string") {
      problems.push(`${relative(root, f)}: item without string id`);
      continue;
    }
    if (isSource) {
      const key = normalizeUrl(it.canonical_url);
      const existingId = urlToId.get(key);
      if (existingId && existingId !== it.id) {
        idRemap.set(it.id, existingId);
        // Merge: keep existing record, but union topics/claim_ids and prefer better verification.
        const ex = sources.get(existingId);
        ex.topic = [...new Set([...(ex.topic ?? []), ...(it.topic ?? [])])];
        ex.claim_ids = [...new Set([...(ex.claim_ids ?? []), ...(it.claim_ids ?? [])])];
        if ((ex.verification?.status ?? "unverified") === "unverified" && it.verification?.status?.startsWith("verified")) {
          ex.verification = it.verification;
          ex.model_use_status = it.model_use_status;
        }
        merged++;
      } else if (sources.has(it.id) && existingId === it.id) {
        merged++; // identical record already present
      } else if (sources.has(it.id) && !existingId) {
        // id collision with a different URL: suffix
        let n = 2;
        let newId = `${it.id}-${n}`;
        while (sources.has(newId)) newId = `${it.id}-${++n}`;
        idRemap.set(`${relative(root, f)}::${it.id}`, newId);
        problems.push(`${relative(root, f)}: id ${it.id} collides with a different URL; renamed to ${newId}`);
        sources.set(newId, { ...it, id: newId });
        urlToId.set(key, newId);
        added++;
      } else {
        sources.set(it.id, it);
        urlToId.set(key, it.id);
        added++;
      }
    } else {
      if (claims.has(it.id)) {
        const ex = claims.get(it.id);
        if (JSON.stringify(ex) !== JSON.stringify(it)) {
          let n = 2;
          let newId = `${it.id}-${n}`;
          while (claims.has(newId)) newId = `${it.id}-${++n}`;
          problems.push(`${relative(root, f)}: claim id ${it.id} collides; renamed to ${newId}`);
          claims.set(newId, { ...it, id: newId });
        }
        merged++;
      } else {
        claims.set(it.id, it);
        added++;
      }
    }
  }
  perFragmentStats.push({ file: relative(root, f), items: items.length, added, merged });
}

// 2. Rewrite remapped source ids in claims and every entity file.
function remapValue(v) {
  if (typeof v === "string") return idRemap.has(v) ? idRemap.get(v) : v;
  if (Array.isArray(v)) return [...new Set(v.map(remapValue))];
  if (v && typeof v === "object") {
    const out = {};
    for (const [k, val] of Object.entries(v)) out[k] = remapValue(val);
    return out;
  }
  return v;
}

const sourceItems = [...sources.values()].map(remapValue).sort((a, b) => a.id.localeCompare(b.id));
const claimItems = [...claims.values()].map(remapValue).sort((a, b) => a.id.localeCompare(b.id));

// Attach claim ids to sources.
const claimIdsBySource = new Map();
for (const c of claimItems) {
  if (typeof c.source_id === "string") {
    if (!claimIdsBySource.has(c.source_id)) claimIdsBySource.set(c.source_id, new Set());
    claimIdsBySource.get(c.source_id).add(c.id);
  }
}
for (const s of sourceItems) {
  const set = new Set([...(s.claim_ids ?? []), ...(claimIdsBySource.get(s.id) ?? [])]);
  s.claim_ids = [...set].sort();
}

writeJSON(join(snapDir, "sources.json"), { kind: "source", schema_version: 1, items: sourceItems });
writeJSON(join(snapDir, "claims.json"), { kind: "claim", schema_version: 1, items: claimItems });

let rewritten = 0;
if (idRemap.size) {
  for (const file of ENTITY_FILES) {
    if (file === "sources.json" || file === "claims.json") continue;
    const p = join(snapDir, file);
    if (!existsSync(p)) continue;
    const before = readFileSync(p, "utf8");
    const after = JSON.stringify(remapValue(JSON.parse(before)), null, 2) + "\n";
    if (before !== after) {
      writeJSON(p, JSON.parse(after));
      rewritten++;
    }
  }
}

// 3. Referential integrity report.
const sourceIds = new Set(sourceItems.map((s) => s.id));
const dangling = [];
for (const file of ENTITY_FILES) {
  const p = join(snapDir, file);
  if (!existsSync(p)) continue;
  const env = readJSON(p);
  const scan = (v, where) => {
    if (Array.isArray(v)) v.forEach((x, i) => scan(x, `${where}[${i}]`));
    else if (v && typeof v === "object") {
      for (const [k, val] of Object.entries(v)) {
        if ((k === "source_ids" || k === "source_id") && val) {
          for (const id of Array.isArray(val) ? val : [val]) {
            if (typeof id === "string" && !sourceIds.has(id)) dangling.push(`${file}:${where}.${k} → ${id}`);
          }
        } else scan(val, `${where}.${k}`);
      }
    }
  };
  scan(env.items ?? [], "items");
}

const report = {
  snapshot: snapDirArg,
  dry_run: dryRun,
  fragments: perFragmentStats,
  sources_total: sourceItems.length,
  claims_total: claimItems.length,
  ids_remapped: idRemap.size,
  entity_files_rewritten: rewritten,
  dangling_source_refs: dangling.length,
  dangling_examples: dangling.slice(0, 25),
  problems,
};
console.log(JSON.stringify(report, null, 2));
process.exit(dangling.length || problems.length ? 1 : 0);
