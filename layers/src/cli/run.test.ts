import { describe, expect, it } from "vitest";
import { runCli } from "./run.js";

describe("runCli", () => {
  it("v1 health returns ok JSON and exit 0", () => {
    const r = runCli(["v1", "health"]);
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

  it("unknown args exit 1 with usage on stderr", () => {
    const r = runCli([]);
    expect(r.code).toBe(1);
    expect(r.stdout).toBe("");
    expect(r.stderr).toContain("Usage:");
  });

  it("v1 alone exits 1", () => {
    const r = runCli(["v1"]);
    expect(r.code).toBe(1);
  });
});
