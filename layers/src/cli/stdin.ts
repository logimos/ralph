import { readFileSync } from "node:fs";

/**
 * Read entire stdin when piped. Returns empty when stdin is a TTY (interactive)
 * so commands like `v1 health` do not block waiting for EOF.
 */
export function readStdinSync(): string {
  if (process.stdin.isTTY) {
    return "";
  }
  try {
    return readFileSync(0, "utf8");
  } catch {
    return "";
  }
}
