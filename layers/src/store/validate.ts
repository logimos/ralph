import type { MemorySource, MemoryType } from "./types.js";
import { MEMORY_SOURCES, MEMORY_TYPES } from "./types.js";

export function parseMemoryType(s: string): MemoryType {
  const t = s.toLowerCase().trim();
  if ((MEMORY_TYPES as readonly string[]).includes(t)) {
    return t as MemoryType;
  }
  throw new Error(`Invalid memory type: ${JSON.stringify(s)}`);
}

export function parseMemorySource(s: string): MemorySource {
  const t = s.toLowerCase().trim();
  if ((MEMORY_SOURCES as readonly string[]).includes(t)) {
    return t as MemorySource;
  }
  throw new Error(`Invalid memory source: ${JSON.stringify(s)}`);
}
