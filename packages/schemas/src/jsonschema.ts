// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * Deterministic JSON Schema generation for data/schemas.
 *
 * Pure: no filesystem access, no clock. `buildJsonSchemaFiles()` returns the
 * complete set of file contents keyed by file name so that the build script and
 * the idempotence test share one implementation.
 */
import { z } from "zod";
import { EnvelopeSchema, KINDS, SCHEMA_VERSION } from "./index";

export const JSON_SCHEMA_TARGET = "draft-2020-12" as const;
export const JSON_SCHEMA_DIALECT = "https://json-schema.org/draft/2020-12/schema";

/** URN identifying a generated schema; not a fetchable URL by design. */
export function schemaId(name: string): string {
  return `urn:pdoom:schema:${name}:v${SCHEMA_VERSION}`;
}

type JsonValue = null | boolean | number | string | JsonValue[] | { [key: string]: JsonValue };

/** Object keys whose insertion (declaration) order is meaningful for readers and kept as-is. */
const ORDER_PRESERVING_KEYS = new Set(["properties", "$defs", "definitions"]);

/** Recursively sorts object keys (except under order-preserving keys) so output is stable. */
export function canonicalize(value: unknown, preserveOrder = false): JsonValue {
  if (value === null || typeof value !== "object") {
    if (value === undefined) return null;
    return value as JsonValue;
  }
  if (Array.isArray(value)) {
    return value.map((item) => canonicalize(item));
  }
  const record = value as Record<string, unknown>;
  const keys = Object.keys(record).filter((key) => record[key] !== undefined);
  if (!preserveOrder) keys.sort();
  const out: { [key: string]: JsonValue } = {};
  for (const key of keys) {
    out[key] = canonicalize(record[key], ORDER_PRESERVING_KEYS.has(key));
  }
  return out;
}

/** Stable JSON text: sorted keys, two-space indent, trailing newline. */
export function canonicalJson(value: unknown): string {
  return `${JSON.stringify(canonicalize(value), null, 2)}\n`;
}

export interface GeneratedSchema {
  $schema: string;
  $id: string;
  title: string;
  [key: string]: unknown;
}

/** Converts a zod schema to a titled, identified JSON Schema document. */
export function toJsonSchemaDocument(schema: z.ZodType, name: string, title: string): GeneratedSchema {
  const generated = z.toJSONSchema(schema, {
    target: JSON_SCHEMA_TARGET,
    io: "output",
    unrepresentable: "throw",
    cycles: "throw",
    reused: "inline",
  }) as Record<string, unknown>;
  const { $schema: _dialect, ...rest } = generated;
  return { $schema: JSON_SCHEMA_DIALECT, $id: schemaId(name), title, ...rest };
}

export interface SchemaIndexEntry {
  kind: string;
  location: string;
  container: string;
  file: string | null;
  item_schema: string;
  envelope_schema: string | null;
}

export interface SchemaIndex {
  kind: "schema_index";
  schema_version: typeof SCHEMA_VERSION;
  target: typeof JSON_SCHEMA_TARGET;
  schemas: SchemaIndexEntry[];
}

function titleCase(kind: string): string {
  return kind
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

/**
 * Builds every JSON Schema file. Keys are file names relative to data/schemas:
 * `<kind>.schema.json` (item), `<kind>.envelope.schema.json` (envelope kinds)
 * and `index.json`.
 */
export function buildJsonSchemaFiles(): Map<string, string> {
  const files = new Map<string, string>();
  const index: SchemaIndex = {
    kind: "schema_index",
    schema_version: SCHEMA_VERSION,
    target: JSON_SCHEMA_TARGET,
    schemas: [],
  };

  for (const entry of KINDS) {
    const itemFile = `${entry.kind}.schema.json`;
    files.set(
      itemFile,
      canonicalJson(toJsonSchemaDocument(entry.schema, entry.kind, `p(DOOM) ${titleCase(entry.kind)}`)),
    );

    let envelopeFile: string | null = null;
    if (entry.container === "envelope") {
      envelopeFile = `${entry.kind}.envelope.schema.json`;
      files.set(
        envelopeFile,
        canonicalJson(
          toJsonSchemaDocument(
            EnvelopeSchema(entry.schema, entry.kind),
            `${entry.kind}.envelope`,
            `p(DOOM) ${titleCase(entry.kind)} file`,
          ),
        ),
      );
    }

    index.schemas.push({
      kind: entry.kind,
      location: entry.location,
      container: entry.container,
      file: entry.file,
      item_schema: itemFile,
      envelope_schema: envelopeFile,
    });
  }

  files.set("index.json", canonicalJson(index));
  return files;
}
