import { describe, expect, it } from "vitest";
import { runCli } from "./run.js";

const emptyEnv = {};
const emptyStdin = "";

describe("runCli", () => {
  it("v1 health returns ok JSON and exit 0", async () => {
    const r = await runCli(["v1", "health"], emptyEnv, emptyStdin);
    expect(r.code).toBe(0);
    expect(r.stderr).toBe("");
    const parsed = JSON.parse(r.stdout.trim()) as {
      ok: boolean;
      version: string;
      service: string;
    };
    expect(parsed.ok).toBe(true);
    expect(parsed.service).toBe("layers");
    expect(parsed.version).toMatch(/^\d+\.\d+\.\d+/);
  });

  it("unknown args exit 1 with usage on stderr", async () => {
    const r = await runCli([], emptyEnv, emptyStdin);
    expect(r.code).toBe(1);
    expect(r.stdout).toBe("");
    expect(r.stderr).toContain("Usage:");
  });

  it("v1 alone exits 1", async () => {
    const r = await runCli(["v1"], emptyEnv, emptyStdin);
    expect(r.code).toBe(1);
  });
});
