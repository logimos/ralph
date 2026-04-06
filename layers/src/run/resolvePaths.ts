import { isAbsolute, resolve } from "node:path";

export const DEFAULT_RUN_LOG_REL = "run.jsonl";
export const DEFAULT_SNAPSHOT_REL = "context-snapshot.md";

/**
 * Resolve path: absolute paths unchanged; relative paths joined to `dataDir`.
 */
export function resolveUnderDataDir(
  dataDir: string,
  p: string | undefined,
  defaultRelative: string
): string {
  const rel = p?.trim() || defaultRelative;
  if (isAbsolute(rel)) {
    return rel;
  }
  return resolve(dataDir, rel);
}
