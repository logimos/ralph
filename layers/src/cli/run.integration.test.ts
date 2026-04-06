import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
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

  it("v1 record writes sqlite under .layers", async () => {
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
    const r = await runCli(["v1", "record"], {}, stdin);
    expect(r.code).toBe(0);
    const out = JSON.parse(r.stdout) as { ok: boolean; count: number };
    expect(out.ok).toBe(true);
    expect(out.count).toBe(1);

    const dbFile = join(dir, ".layers", "memory.db");
    expect(existsSync(dbFile)).toBe(true);
  });

  it("v1 import-ralph-memory with stdin", async () => {
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
    const r = await runCli(["v1", "import-ralph-memory"], {}, stdin);
    expect(r.code).toBe(0);
    const out = JSON.parse(r.stdout) as { imported: number; skipped: number };
    expect(out.imported).toBe(1);
  });

  it("v1 retrieve returns contextBlock after record", async () => {
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
    expect((await runCli(["v1", "record"], {}, rec)).code).toBe(0);

    const retrieveStdin = JSON.stringify({
      projectRoot: dir,
      query: {
        text: "authentication JWT",
        category: "feature",
        featureId: 2,
      },
      options: { topK: 5, maxTokens: 2000 },
    });
    const r = await runCli(["v1", "retrieve"], {}, retrieveStdin);
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

  it("v1 append-run then v1 compact writes context-snapshot.md", async () => {
    dir = mkdtempSync(join(tmpdir(), "layers-cli-phase3-"));
    const ev = JSON.stringify({
      projectRoot: dir,
      event: {
        sessionKey: "ralph-session-1",
        kind: "structured",
        iteration: 2,
        featureId: 5,
        payload: { status: "ok" },
      },
    });
    const a = await runCli(["v1", "append-run"], {}, ev);
    expect(a.code).toBe(0);
    const appendOut = JSON.parse(a.stdout) as { path: string };
    expect(existsSync(appendOut.path)).toBe(true);

    const compactStdin = JSON.stringify({
      projectRoot: dir,
      maxEvents: 10,
      maxBytes: 50_000,
    });
    const c = await runCli(["v1", "compact"], {}, compactStdin);
    expect(c.code).toBe(0);
    const compactOut = JSON.parse(c.stdout) as {
      snapshotPath: string;
      eventCount: number;
      bytesWritten: number;
    };
    expect(compactOut.eventCount).toBe(1);
    expect(compactOut.bytesWritten).toBeGreaterThan(0);
    const snap = readFileSync(compactOut.snapshotPath, "utf8");
    expect(snap).toContain("Layers run context");
    expect(snap).toContain("ralph-session-1");
  });
});
