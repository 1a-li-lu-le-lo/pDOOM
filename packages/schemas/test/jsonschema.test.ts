// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { KINDS } from "../src/index";
import { buildJsonSchemaFiles, canonicalJson, canonicalize } from "../src/jsonschema";

const here = dirname(fileURLToPath(import.meta.url));
const schemaDir = resolve(here, "..", "..", "..", "data", "schemas");

describe("JSON Schema build", () => {
  const first = buildJsonSchemaFiles();
  const second = buildJsonSchemaFiles();

  it("is idempotent", () => {
    expect([...second.keys()]).toEqual([...first.keys()]);
    for (const [name, content] of first) expect(second.get(name)).toBe(content);
  });

  it("emits an item schema for every kind, an envelope schema for envelope kinds, and an index", () => {
    for (const entry of KINDS) {
      expect(first.has(`${entry.kind}.schema.json`)).toBe(true);
      expect(first.has(`${entry.kind}.envelope.schema.json`)).toBe(entry.container === "envelope");
    }
    const index = JSON.parse(first.get("index.json") ?? "{}");
    expect(index.kind).toBe("schema_index");
    expect(index.schema_version).toBe(1);
    expect(index.schemas.map((s: { kind: string }) => s.kind)).toEqual(KINDS.map((k) => k.kind));
  });

  it("every schema is a closed object with a URN id and no timestamps", () => {
    for (const [name, content] of first) {
      if (name === "index.json") continue;
      expect(content.endsWith("\n")).toBe(true);
      const doc = JSON.parse(content);
      expect(doc.$schema).toBe("https://json-schema.org/draft/2020-12/schema");
      expect(doc.$id).toMatch(/^urn:pdoom:schema:[a-z_.]+:v1$/);
      expect(doc.type).toBe("object");
      expect(doc.additionalProperties).toBe(false);
      expect(content).not.toMatch(/"(generated_at|built_at|build_time)"\s*:\s*"\d{4}-/);
    }
  });

  it("uses only RE2-compatible regular expressions", () => {
    for (const content of first.values()) {
      expect(content).not.toMatch(/\(\?[=!<]/);
    }
  });

  it("matches the committed files in data/schemas", () => {
    expect(existsSync(schemaDir)).toBe(true);
    for (const [name, content] of first) {
      const onDisk = readFileSync(join(schemaDir, name), "utf8");
      expect(onDisk, `${name} is stale; run pnpm --filter @pdoom/schemas build:jsonschema`).toBe(content);
    }
    const generatedOnDisk = readdirSync(schemaDir).filter(
      (n) => n === "index.json" || n.endsWith(".schema.json"),
    );
    expect(generatedOnDisk.sort()).toEqual([...first.keys()].sort());
  });
});

describe("canonical JSON", () => {
  it("sorts keys but preserves property declaration order", () => {
    const value = { b: 1, a: { z: 1, y: 2 }, properties: { id: {}, aardvark: {} } };
    expect(canonicalJson(value)).toBe(
      `${JSON.stringify({ a: { y: 2, z: 1 }, b: 1, properties: { id: {}, aardvark: {} } }, null, 2)}\n`,
    );
    expect(canonicalize([{ b: 1, a: 2 }])).toEqual([{ a: 2, b: 1 }]);
  });
});
