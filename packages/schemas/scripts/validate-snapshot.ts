import { readFileSync } from "node:fs";
import { KINDS, SnapshotManifestSchema } from "@pdoom/schemas";
const dir = process.argv[2];
for (const k of KINDS) {
  if (k.location !== "snapshot" || !k.file) continue;
  let raw: string;
  try { raw = readFileSync(`${dir}/${k.file}`, "utf8"); } catch { continue; }
  const doc = JSON.parse(raw);
  if (k.kind === "snapshot_manifest") {
    const r = SnapshotManifestSchema.safeParse(doc);
    if (!r.success) console.log(k.file, JSON.stringify(r.error.issues.slice(0, 8)));
    continue;
  }
  const seen = new Set<string>();
  for (const it of doc.items) {
    const r = k.schema.safeParse(it);
    if (!r.success) for (const iss of r.error.issues) { const key = iss.path.join(".") + ":" + iss.message; if (!seen.has(key)) { seen.add(key); console.log(k.file, it.id, key); } }
  }
}
