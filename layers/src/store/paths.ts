import { mkdirSync } from "node:fs";
import { join, resolve } from "node:path";

/**
 * Request `dataDir` wins; else `LAYERS_DATA_DIR` env (LAYERS_SPEC §7.1).
 */
export function effectiveDataDirOverride(
  fromRequest: string | undefined,
  env: NodeJS.ProcessEnv
): string | undefined {
  const r = fromRequest?.trim();
  if (r) {
    return r;
  }
  const e = env.LAYERS_DATA_DIR?.trim();
  return e || undefined;
}

/**
 * Default data directory under project root (LAYERS_SPEC §7.1).
 */
export function resolveDataDir(projectRoot: string, override?: string): string {
  const root = resolve(projectRoot);
  const dir = override?.trim() || join(root, ".layers");
  mkdirSync(dir, { recursive: true });
  return dir;
}

export function dbPath(dataDir: string): string {
  return join(dataDir, "memory.db");
}
