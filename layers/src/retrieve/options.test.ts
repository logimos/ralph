import { describe, expect, it } from "vitest";
import {
  computePoolSize,
  normalizeMaxTokens,
  normalizeMmrLambda,
  normalizeTopK,
} from "./options.js";

describe("retrieve options normalization", () => {
  it("clamps topK and handles invalid", () => {
    expect(normalizeTopK(undefined)).toBe(10);
    expect(normalizeTopK(-1)).toBe(10);
    expect(normalizeTopK(NaN)).toBe(10);
    expect(normalizeTopK(3.7)).toBe(3);
    expect(normalizeTopK(500)).toBe(100);
  });

  it("clamps maxTokens", () => {
    expect(normalizeMaxTokens(undefined)).toBe(2000);
    expect(normalizeMaxTokens(0)).toBe(2000);
    expect(normalizeMaxTokens(-5)).toBe(2000);
    expect(normalizeMaxTokens(42.9)).toBe(42);
  });

  it("computePoolSize is non-negative and capped", () => {
    expect(computePoolSize(0)).toBe(0);
    expect(computePoolSize(5)).toBe(20);
    expect(computePoolSize(100)).toBe(50);
  });

  it("normalizes mmr lambda", () => {
    expect(normalizeMmrLambda(undefined)).toBe(0.5);
    expect(normalizeMmrLambda(2)).toBe(1);
    expect(normalizeMmrLambda(-1)).toBe(0);
  });
});
