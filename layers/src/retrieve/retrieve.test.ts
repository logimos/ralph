import { mkdirSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { insertMemory } from "../store/insert.js";
import { openDatabase } from "../store/open.js";
import type { Embedder } from "../embed/types.js";
import { retrieveMemories } from "./retrieve.js";

describe("retrieveMemories", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("returns FTS-ranked results and contextBlock", async () => {
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

      const { memories, contextBlock, meta } = await retrieveMemories(
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

  it("respects maxTokens on contextBlock (word budget)", async () => {
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
      const { contextBlock } = await retrieveMemories(
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

  it("hybrid retrieval uses embeddings when embedder provided", async () => {
    dir = mkdtempSync(join(tmpdir(), "layers-retrieve-hybrid-"));
    mkdirSync(join(dir, ".layers"), { recursive: true });
    const db = openDatabase(join(dir, ".layers"));
    const emb = (t: string): number[] => {
      const v = [0, 0, 0, 0];
      if (t.includes("alpha")) {
        v[0] = 1;
      }
      if (t.includes("beta")) {
        v[1] = 1;
      }
      return v;
    };
    const mockEmbedder: Embedder = {
      model: "mock",
      async embed(texts: string[]) {
        return texts.map(emb);
      },
    };
    try {
      insertMemory(db, {
        type: "fact",
        content: "Topic alpha documentation",
        source: "agent",
      });
      insertMemory(db, {
        type: "fact",
        content: "Topic beta unrelated",
        source: "agent",
      });
      const { memories, meta } = await retrieveMemories(
        db,
        { text: "alpha docs", category: null, featureId: 0 },
        { topK: 2, maxTokens: 500, mmrLambda: 1, vectorWeight: 0.9, textWeight: 0.1 },
        { embedder: mockEmbedder }
      );
      expect(meta.ftsOnly).toBe(false);
      expect(memories[0]?.content).toContain("alpha");
    } finally {
      db.close();
    }
  });
});
