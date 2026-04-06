import { describe, expect, it } from "vitest";
import { blobToFloat32Array, cosineSimilarity } from "./vector.js";

describe("cosineSimilarity", () => {
  it("is 1 for identical vectors", () => {
    expect(cosineSimilarity([1, 0, 0], [1, 0, 0])).toBeCloseTo(1);
  });

  it("is 0 for orthogonal vectors", () => {
    expect(cosineSimilarity([1, 0], [0, 1])).toBeCloseTo(0);
  });
});

describe("blobToFloat32Array", () => {
  it("returns null when byte length is not a multiple of 4", () => {
    expect(blobToFloat32Array(Buffer.from([1, 2, 3]))).toBeNull();
  });
});
