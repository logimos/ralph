import { mkdirSync, mkdtempSync, rmSync } from "node:fs";
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

  it("respects maxTokens on contextBlock (word budget)", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-retrieve-tok-"));
    mkdirSync(join(dir, ".layers"), { recursive: true });
    const db = openDatabase(join(dir, ".layers"));
    try {
      const longContent = Array.from({ length: 80 }, (_, i) => `w${i}`).join(" ");
      insertMemory(db, {
        type: "context",
        content: longContent,
        source: "agent",
      });
      const maxTok = 35;
      const { contextBlock } = retrieveMemories(
        db,
        { text: "w0 w1", category: null, featureId: 0 },
        { topK: 3, maxTokens: maxTok, mmrLambda: 1 }
      );
      const wc = contextBlock.trim().length === 0 ? 0 : contextBlock.trim().split(/\s+/).length;
      expect(wc).toBeLessThanOrEqual(maxTok);
    } finally {
      db.close();
    }
  });
});
