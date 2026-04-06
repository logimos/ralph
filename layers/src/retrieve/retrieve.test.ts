import { mkdirSync, rmSync } from "node:fs";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { insertMemory } from "../store/insert.js";
import { openDatabase } from "../store/open.js";
import { retrieveMemories } from "./retrieve.js";

describe("retrieveMemories", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("returns FTS-ranked results and contextBlock", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-retrieve-"));
    mkdirSync(join(dir, ".layers"), { recursive: true });
    const db = openDatabase(join(dir, ".layers"));
    try {
      insertMemory(db, {
        type: "decision",
        content: "Use SQLite for local development",
        category: "infra",
        source: "agent",
      });
      insertMemory(db, {
        type: "convention",
        content: "Prefer TypeScript for new services",
        category: "chore",
        source: "agent",
      });

      const { memories, contextBlock, meta } = retrieveMemories(
        db,
        { text: "SQLite development", category: null, featureId: 0 },
        { topK: 5, maxTokens: 2000, mmrLambda: 0.5 }
      );

      expect(meta.ftsOnly).toBe(true);
      expect(memories.length).toBeGreaterThan(0);
      expect(contextBlock).toContain("[MEMORY CONTEXT]");
      expect(contextBlock).toContain("[END MEMORY CONTEXT]");
      expect(contextBlock.toLowerCase()).toContain("sqlite");
    } finally {
      db.close();
    }
  });
});
