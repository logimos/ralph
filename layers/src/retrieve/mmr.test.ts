import { describe, expect, it } from "vitest";
import { mmrSelect } from "./mmr.js";

describe("mmrSelect", () => {
  it("prefers diverse items when lambda is moderate", () => {
    const items = [
      { id: "1", relevance: 1, text: "use postgres for db" },
      { id: "2", relevance: 0.95, text: "use postgres for database" },
      { id: "3", relevance: 0.5, text: "write tests first" },
    ];
    const out = mmrSelect(items, 2, 0.5);
    expect(out).toHaveLength(2);
    const ids = out.map((x) => x.id).sort();
    expect(ids).toContain("1");
    expect(ids).toContain("3");
  });
});
