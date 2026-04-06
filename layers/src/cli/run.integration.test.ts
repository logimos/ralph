import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { runCli } from "./run.js";

describe("runCli record + import integration", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("v1 record writes sqlite under .layers", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-cli-"));
    const stdin = JSON.stringify({
      projectRoot: dir,
      entries: [
        {
          type: "fact",
          content: "API base is /v1",
          category: "infra",
          featureId: 7,
          source: "ralph",
        },
      ],
    });
    const r = runCli(["v1", "record"], {}, stdin);
    expect(r.code).toBe(0);
    const out = JSON.parse(r.stdout) as { ok: boolean; count: number };
    expect(out.ok).toBe(true);
    expect(out.count).toBe(1);

    const dbFile = join(dir, ".layers", "memory.db");
    expect(existsSync(dbFile)).toBe(true);
  });

  it("v1 import-ralph-memory with stdin", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-cli-"));
    mkdirSync(join(dir, ".layers"), { recursive: true });
    writeFileSync(
      join(dir, ".ralph-memory.json"),
      JSON.stringify({
        entries: [
          {
            id: "x1",
            type: "context",
            content: "Redis for cache",
            created_at: "2025-03-01T00:00:00.000Z",
            updated_at: "2025-03-01T00:00:00.000Z",
          },
        ],
      }),
      "utf8"
    );

    const stdin = JSON.stringify({ projectRoot: dir });
    const r = runCli(["v1", "import-ralph-memory"], {}, stdin);
    expect(r.code).toBe(0);
    const out = JSON.parse(r.stdout) as { imported: number; skipped: number };
    expect(out.imported).toBe(1);
  });

  it("v1 retrieve returns contextBlock after record", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-cli-retrieve-"));
    const rec = JSON.stringify({
      projectRoot: dir,
      entries: [
        {
          type: "decision",
          content: "Use JWT for auth tokens",
          category: "feature",
          featureId: 2,
          source: "agent",
        },
      ],
    });
    expect(runCli(["v1", "record"], {}, rec).code).toBe(0);

    const retrieveStdin = JSON.stringify({
      projectRoot: dir,
      query: {
        text: "authentication JWT",
        category: "feature",
        featureId: 2,
      },
      options: { topK: 5, maxTokens: 2000 },
    });
    const r = runCli(["v1", "retrieve"], {}, retrieveStdin);
    expect(r.code).toBe(0);
    const out = JSON.parse(r.stdout) as {
      contextBlock: string;
      memories: unknown[];
      meta: { ftsOnly: boolean };
    };
    expect(out.contextBlock).toContain("[MEMORY CONTEXT]");
    expect(out.memories.length).toBeGreaterThan(0);
    expect(out.meta.ftsOnly).toBe(true);
  });
});
