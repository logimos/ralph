import { isAbsolute, relative, resolve } from "node:path";

export const DEFAULT_RUN_LOG_REL = "run.jsonl";
export const DEFAULT_SNAPSHOT_REL = "context-snapshot.md";

/**
 * Resolve `p` relative to `dataDir`, or validate an absolute path that still lies
 * under `dataDir` (no `..` escape).
 */
export function resolvePathInDataDir(
  dataDir: string,
  p: string | undefined,
  defaultRelative: string
): string {
  const absData = resolve(dataDir);
  const raw = p?.trim() || defaultRelative;
  const absTarget = isAbsolute(raw) ? resolve(raw) : resolve(absData, raw);
  const rel = relative(absData, absTarget);
  if (rel.startsWith("..") || rel === "..") {
    throw new Error(`path must stay within the Layers data directory: ${p ?? defaultRelative}`);
  }
  return absTarget;
}
