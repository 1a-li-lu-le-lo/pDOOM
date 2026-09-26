// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { describe, expect, it } from "vitest";
import { existsSync, mkdirSync, writeFileSync, cpSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { fileURLToPath } from "node:url";
import { createFileDataSource, resolveDataDir } from "../src/index";

const sample = fileURLToPath(new URL("../../schemas/test/release-sample/", import.meta.url));

function tempDataDir(): string {
  const dir = join(tmpdir(), `pdoom-sdk-${process.pid}-${Math.random().toString(36).slice(2)}`);
  mkdirSync(join(dir, "releases", "rel-2026-09-26-001"), { recursive: true });
  cpSync(sample, join(dir, "releases", "rel-2026-09-26-001"), { recursive: true });
  writeFileSync(join(dir, "releases", "CURRENT"), "rel-2026-09-26-001\n");
  return dir;
}

describe("file data source", () => {
  it("loads the current release from CURRENT and validates it", async () => {
    const ds = createFileDataSource({ dataDir: tempDataDir() });
    const rel = await ds.getRelease();
    expect(rel.manifest.release_id).toBe("rel-2026-09-26-001");
    expect(rel.estimates.some((e) => e.status === "insufficiently_calibrated")).toBe(true);
    expect(rel.indexes.every((i) => i.is_probability === false)).toBe(true);
    const list = await ds.listReleases();
    expect(list[0]?.is_current).toBe(true);
  });

  it("rejects invalid release ids and missing releases", async () => {
    const ds = createFileDataSource({ dataDir: tempDataDir() });
    await expect(ds.getReleaseById("../etc")).rejects.toThrow(/invalid/);
    await expect(ds.getReleaseById("rel-2000-01-01-001")).rejects.toThrow(/not found/);
  });

  it("resolveDataDir honours PDOOM_DATA_DIR", () => {
    const dir = tempDataDir();
    process.env.PDOOM_DATA_DIR = dir;
    expect(resolveDataDir()).toBe(dir);
    delete process.env.PDOOM_DATA_DIR;
    expect(existsSync(dir)).toBe(true);
  });
});
