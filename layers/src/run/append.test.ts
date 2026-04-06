import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { appendRunEvent, validateRunEvent } from "./append.js";

describe("validateRunEvent", () => {
  it("accepts minimal structured event", () => {
    const e = validateRunEvent({
      sessionKey: "run-1",
      kind: "structured",
      payload: { iteration: 3 },
    });
    expect(e.sessionKey).toBe("run-1");
    expect(e.kind).toBe("structured");
    expect(e.payload).toEqual({ iteration: 3 });
    expect(e.ts).toBeTruthy();
  });

  it("rejects bad kind", () => {
    expect(() =>
      validateRunEvent({ sessionKey: "x", kind: "nope", ts: new Date().toISOString() })
    ).toThrow();
  });

  it("coerces string iteration and featureId", () => {
    const e = validateRunEvent({
      sessionKey: "s",
      kind: "note",
      iteration: "3",
      featureId: "12",
    });
    expect(e.iteration).toBe(3);
    expect(e.featureId).toBe(12);
  });

  it("throws on invalid iteration type", () => {
    expect(() =>
      validateRunEvent({
        sessionKey: "s",
        kind: "note",
        iteration: 3.5,
      })
    ).toThrow(/iteration/);
  });
});

describe("appendRunEvent", () => {
  let dir: string;
  afterEach(() => {
    if (dir) {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it("appends JSONL line", () => {
    dir = mkdtempSync(join(tmpdir(), "layers-append-"));
    const logPath = join(dir, "nested", "run.jsonl");
    const ev = validateRunEvent({
      sessionKey: "s",
      kind: "note",
      payload: "hello",
    });
    appendRunEvent(logPath, ev);
    const raw = readFileSync(logPath, "utf8").trim();
    const parsed = JSON.parse(raw) as { sessionKey: string };
    expect(parsed.sessionKey).toBe("s");
  });
});
