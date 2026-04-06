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
});
