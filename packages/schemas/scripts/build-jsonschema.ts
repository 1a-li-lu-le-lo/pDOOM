// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Writes data/schemas/<kind>.schema.json (+ envelope schemas and index.json)
 * from the zod schemas. Deterministic: same input → byte-identical output.
 *
 *   pnpm --filter @pdoom/schemas build:jsonschema          # write
 *   pnpm --filter @pdoom/schemas build:jsonschema -- --check   # exit 1 if stale
 */
import { existsSync, mkdirSync, readdirSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { buildJsonSchemaFiles } from "../src/jsonschema";

const here = dirname(fileURLToPath(import.meta.url));
const outDir = resolve(here, "..", "..", "..", "data", "schemas");
const checkOnly = process.argv.includes("--check");

function isGeneratedName(name: string): boolean {
  return name === "index.json" || name.endsWith(".schema.json");
}

function main(): number {
  const files = buildJsonSchemaFiles();
  mkdirSync(outDir, { recursive: true });

  let written = 0;
  let unchanged = 0;
  let stale = 0;

  for (const [name, content] of files) {
    const target = join(outDir, name);
    const current = existsSync(target) ? readFileSync(target, "utf8") : null;
    if (current === content) {
      unchanged += 1;
      continue;
    }
    stale += 1;
    if (!checkOnly) {
      writeFileSync(target, content, "utf8");
      written += 1;
    } else {
      console.error(`stale: ${name}`);
    }
  }

  for (const name of readdirSync(outDir)) {
    if (isGeneratedName(name) && !files.has(name)) {
      stale += 1;
      if (!checkOnly) {
        unlinkSync(join(outDir, name));
        console.info(`removed: ${name}`);
      } else {
        console.error(`orphan: ${name}`);
      }
    }
  }

  if (checkOnly) {
    if (stale > 0) {
      console.error(`${stale} schema file(s) out of date; run pnpm --filter @pdoom/schemas build:jsonschema`);
      return 1;
    }
    console.info(`data/schemas up to date (${files.size} files)`);
    return 0;
  }

  console.info(`data/schemas: ${written} written, ${unchanged} unchanged (${files.size} files)`);
  return 0;
}

process.exitCode = main();
