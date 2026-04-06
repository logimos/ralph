import { mkdirSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { resolvePathInDataDir } from "./resolvePaths.js";

describe("resolvePathInDataDir", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("joins relative path inside data dir", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-rp-"));
    const data = join(dir, ".layers");
    mkdirSync(data, { recursive: true });
    const p = resolvePathInDataDir(data, "logs/x.jsonl", "run.jsonl");
    expect(p.startsWith(resolve(data))).toBe(true);
    expect(p).toContain("logs");
  });

  it("rejects path that escapes data dir", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-rp-"));
    const data = join(dir, ".layers");
    mkdirSync(data, { recursive: true });
    expect(() => resolvePathInDataDir(data, "../../etc/passwd", "run.jsonl")).toThrow(
      /within the Layers data directory/
    );
  });
});
