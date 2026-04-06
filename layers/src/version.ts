import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

/**
 * Runtime version from layers/package.json (single source of truth).
 * Resolved from this module's location under dist/ after build.
 */
export function getPackageVersion(): string {
  const here = dirname(fileURLToPath(import.meta.url));
  const pkgPath = join(here, "..", "package.json");
  const raw = readFileSync(pkgPath, "utf8");
  const pkg = JSON.parse(raw) as { version: string };
  if (typeof pkg.version !== "string" || !pkg.version) {
    throw new Error(`Invalid or missing version in ${pkgPath}`);
  }
  return pkg.version;
}
