import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { insertMemory } from "./insert.js";
import { openDatabase } from "./open.js";

describe("insertMemory + FTS", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("inserts and FTS index matches token", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-test-"));
    const db = openDatabase(dir);
    insertMemory(db, {
      type: "decision",
      content: "Use PostgreSQL for persistence",
      source: "agent",
    });
    const row = db
      .prepare(`SELECT memory_id FROM memory_fts WHERE memory_fts MATCH ? LIMIT 1`)
      .get("PostgreSQL") as { memory_id: string } | undefined;
    expect(row?.memory_id).toBeTruthy();
    db.close();
  });
});
