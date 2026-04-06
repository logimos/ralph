import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname } from "node:path";

export type CompactOptions = {
  maxEvents: number;
  maxBytes: number;
};

const DEFAULT_MAX_EVENTS = 50;
const DEFAULT_MAX_BYTES = 32_000;

export function normalizeCompactOptions(maxEvents?: number, maxBytes?: number): CompactOptions {
  let me = DEFAULT_MAX_EVENTS;
  if (maxEvents !== undefined && Number.isFinite(maxEvents)) {
    me = Math.min(10_000, Math.max(1, Math.trunc(maxEvents)));
  }
  let mb = DEFAULT_MAX_BYTES;
  if (maxBytes !== undefined && Number.isFinite(maxBytes)) {
    mb = Math.min(2_000_000, Math.max(256, Math.trunc(maxBytes)));
  }
  return { maxEvents: me, maxBytes: mb };
}

/**
 * Read last N non-empty lines from a JSONL file (MVP: full read).
 */
export function readLastJsonlLines(absPath: string, maxLines: number): string[] {
  let raw: string;
  try {
    raw = readFileSync(absPath, "utf8");
  } catch (e) {
    const err = e as NodeJS.ErrnoException;
    if (err.code === "ENOENT") {
      return [];
    }
    throw e;
  }
  const lines = raw.split(/\r?\n/).filter((l) => l.trim().length > 0);
  if (lines.length <= maxLines) {
    return lines;
  }
  return lines.slice(-maxLines);
}

const TRUNC_MARKER = "\n\n… _(truncated to maxBytes)_\n";

/** Truncate string to fit UTF-8 byte budget without splitting a code point. */
export function utf8ByteTruncate(s: string, maxBytes: number): string {
  const buf = Buffer.from(s, "utf8");
  if (buf.length <= maxBytes) {
    return s;
  }
  const decoder = new TextDecoder("utf-8", { fatal: true });
  let end = maxBytes;
  while (end > 0) {
    try {
      decoder.decode(buf.subarray(0, end));
      return buf.subarray(0, end).toString("utf8");
    } catch {
      end--;
    }
  }
  return "";
}

export function buildSnapshotContent(lines: string[], maxBytes: number): string {
  const header = `# Layers run context (last ${lines.length} event(s))\n\n`;
  const body = lines
    .map((line, i) => {
      let parsed: unknown;
      try {
        parsed = JSON.parse(line) as unknown;
      } catch {
        return `## Line ${i + 1}\n\n\`\`\`\n${line}\n\`\`\`\n`;
      }
      return `## Event ${i + 1}\n\n\`\`\`json\n${JSON.stringify(parsed, null, 2)}\n\`\`\`\n`;
    })
    .join("\n");

  const text = header + body;
  const budget = maxBytes - Buffer.byteLength(TRUNC_MARKER, "utf8");
  if (budget < 64) {
    return utf8ByteTruncate(text, maxBytes);
  }
  if (Buffer.byteLength(text, "utf8") <= maxBytes) {
    return text;
  }
  return utf8ByteTruncate(text, budget) + TRUNC_MARKER;
}

export function writeSnapshot(absPath: string, lines: string[], opts: CompactOptions): number {
  mkdirSync(dirname(absPath), { recursive: true });
  const content = buildSnapshotContent(lines, opts.maxBytes);
  writeFileSync(absPath, content, "utf8");
  return Buffer.byteLength(content, "utf8");
}
