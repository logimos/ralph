import { describe, expect, it } from "vitest";
import { getPackageVersion } from "./version.js";

describe("getPackageVersion", () => {
  it("reads semver from package.json", () => {
    const v = getPackageVersion();
    expect(v).toMatch(/^\d+\.\d+\.\d+/);
  });
});
