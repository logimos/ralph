import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import {
  buildSnapshotContent,
  normalizeCompactOptions,
  readLastJsonlLines,
  utf8ByteTruncate,
  writeSnapshot,
} from "./compact.js";

describe("readLastJsonlLines", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("returns last N lines", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-jsonl-"));
    const p = join(dir, "x.jsonl");
    writeFileSync(p, '{"a":1}\n{"b":2}\n{"c":3}\n', "utf8");
    expect(readLastJsonlLines(p, 2)).toEqual(['{"b":2}', '{"c":3}']);
  });

  it("missing file returns empty", () => {
    const d = mkdtempSync(join(tmpdir(), "layers-miss-"));
    expect(readLastJsonlLines(join(d, "nope.jsonl"), 5)).toEqual([]);
    rmSync(d, { recursive: true, force: true });
  });

  it("tail-reads large file without loading full string at once", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-big-jsonl-"));
    const p = join(dir, "big.jsonl");
    const line = '{"n":' + "x".repeat(400_000) + "}\n";
    writeFileSync(p, line.repeat(6), "utf8");
    const lines = readLastJsonlLines(p, 2);
    expect(lines).toHaveLength(2);
    expect(lines[0]).toContain('"n":');
  });
});

describe("buildSnapshotContent", () => {
  it("truncates when over maxBytes", () => {
    const long = "x".repeat(5000);
    const lines = [`{"p":"${long}"}`];
    const out = buildSnapshotContent(lines, 500);
    expect(Buffer.byteLength(out, "utf8")).toBeLessThanOrEqual(500);
    expect(out).toContain("truncated");
  });
});

describe("utf8ByteTruncate", () => {
  it("handles multi-byte chars", () => {
    const out = utf8ByteTruncate("café", 4);
    expect(Buffer.byteLength(out, "utf8")).toBeLessThanOrEqual(4);
    expect(out).toBe("caf");
  });
});

describe("writeSnapshot", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("writes file", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-snap-"));
    const p = join(dir, "out.md");
    const n = writeSnapshot(p, ['{"k":1}'], normalizeCompactOptions(10, 50_000));
    expect(n).toBeGreaterThan(0);
    expect(readFileSync(p, "utf8")).toContain("Event");
  });
});
