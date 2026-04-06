import { readFileSync } from "node:fs";

/**
 * Read entire stdin (for piped JSON). Empty if nothing piped.
 */
export function readStdinSync(): string {
  try {
    return readFileSync(0, "utf8");
  } catch {
    return "";
  }
}
