import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { importRalphMemoryIntoDb, readRalphMemoryFile } from "./ralphMemory.js";
import { openDatabase } from "../store/open.js";

describe("importRalphMemory", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("reads legacy JSON and imports with legacy_id", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-ralph-"));
    const memPath = join(dir, ".ralph-memory.json");
    writeFileSync(
      memPath,
      JSON.stringify({
        entries: [
          {
            id: "mem_legacy_1",
            type: "convention",
            content: "Use fmt for Go",
            category: "chore",
            created_at: "2025-01-01T00:00:00.000Z",
            updated_at: "2025-01-02T00:00:00.000Z",
            source: "user",
          },
        ],
        last_updated: "2025-01-02T00:00:00.000Z",
      }),
      "utf8"
    );

    const parsed = readRalphMemoryFile(memPath);
    expect(parsed.entries?.length).toBe(1);

    const dbDir = join(dir, ".layers");
    mkdirSync(dbDir, { recursive: true });
    const db = openDatabase(dbDir);
    const r = importRalphMemoryIntoDb(db, memPath);
    expect(r.imported).toBe(1);
    expect(r.skipped).toBe(0);

    const r2 = importRalphMemoryIntoDb(db, memPath);
    expect(r2.imported).toBe(0);
    expect(r2.skipped).toBe(1);

    const row = db
      .prepare("SELECT id, legacy_id, type FROM memories WHERE legacy_id = ?")
      .get("mem_legacy_1") as { id: string; legacy_id: string; type: string };
    expect(row.legacy_id).toBe("mem_legacy_1");
    expect(row.type).toBe("convention");

    db.close();
  });

  it("reports validation errors with entry index and id", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-ralph-bad-"));
    const memPath = join(dir, ".ralph-memory.json");
    writeFileSync(
      memPath,
      JSON.stringify({
        entries: [
          {
            id: "bad1",
            type: "",
            content: "x",
            created_at: "2025-01-01T00:00:00.000Z",
            updated_at: "2025-01-01T00:00:00.000Z",
          },
        ],
      }),
      "utf8"
    );
    const dbDir = join(dir, ".layers");
    mkdirSync(dbDir, { recursive: true });
    const db = openDatabase(dbDir);
    const r = importRalphMemoryIntoDb(db, memPath);
    expect(r.imported).toBe(0);
    expect(r.errors.some((m) => m.includes("entries[0]") && m.includes("bad1"))).toBe(true);
    db.close();
  });
});
