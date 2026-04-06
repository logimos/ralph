import { describe, expect, it } from "vitest";
import { buildFtsMatchQuery, normalizeQueryText } from "./ftsQuery.js";

describe("buildFtsMatchQuery", () => {
  it("AND-joins tokens and escapes quotes", () => {
    expect(buildFtsMatchQuery("Use PostgreSQL for data")).toBe(
      '"use" AND "postgresql" AND "for" AND "data"'
    );
  });

  it("returns null for empty or tokenless input", () => {
    expect(buildFtsMatchQuery("")).toBeNull();
    expect(buildFtsMatchQuery("a b")).toBeNull();
  });
});

describe("normalizeQueryText", () => {
  it("collapses whitespace", () => {
    expect(normalizeQueryText("  hello   world  ")).toBe("hello world");
  });
});
