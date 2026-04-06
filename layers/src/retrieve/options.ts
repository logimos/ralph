const MAX_TOP_K = 100;
const MAX_POOL = 50;
const DEFAULT_TOP_K = 10;
const DEFAULT_MAX_TOKENS = 2000;
const MAX_MAX_TOKENS = 100_000;

/** Non-negative integer; invalid → fallback. */
export function normalizeTopK(value: number | undefined, fallback = DEFAULT_TOP_K): number {
  if (value === undefined || !Number.isFinite(value)) {
    return fallback;
  }
  const t = Math.trunc(value);
  if (t < 0) {
    return fallback;
  }
  return Math.min(MAX_TOP_K, t);
}

/** Positive integer budget; zero/negative or non-finite → fallback. */
export function normalizeMaxTokens(
  value: number | undefined,
  fallback = DEFAULT_MAX_TOKENS
): number {
  if (value === undefined || !Number.isFinite(value)) {
    return fallback;
  }
  const t = Math.trunc(value);
  if (t < 1) {
    return fallback;
  }
  return Math.min(MAX_MAX_TOKENS, t);
}

export function normalizeMmrLambda(value: number | undefined, fallback = 0.5): number {
  if (value === undefined || !Number.isFinite(value)) {
    return fallback;
  }
  return Math.min(1, Math.max(0, value));
}

export function computePoolSize(topK: number): number {
  if (topK <= 0) {
    return 0;
  }
  return Math.min(MAX_POOL, Math.max(topK * 4, topK));
}

export { DEFAULT_TOP_K, DEFAULT_MAX_TOKENS, MAX_POOL };
