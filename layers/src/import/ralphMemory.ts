import { readFileSync } from "node:fs";
import { isAbsolute, relative, resolve } from "node:path";
import type Database from "better-sqlite3";
import { insertMemory } from "../store/insert.js";

/** Matches internal/memory Memory struct (Go). */
type RalphMemoryFile = {
  entries?: RalphEntry[];
  last_updated?: string;
  retention_days?: number;
};

type RalphEntry = {
  id: string;
  type: string;
  content: string;
  category?: string;
  created_at: string;
  updated_at: string;
  source?: string;
};

export function readRalphMemoryFile(absPath: string): RalphMemoryFile {
  const raw = readFileSync(absPath, "utf8");
  const data = JSON.parse(raw) as RalphMemoryFile;
  if (!data.entries || !Array.isArray(data.entries)) {
    throw new Error("Invalid .ralph-memory.json: missing entries array");
  }
  return data;
}

export type ImportRalphResult = {
  imported: number;
  skipped: number;
  errors: string[];
};

/**
 * Import legacy entries; uses legacy_id = Ralph entry id for idempotent re-import.
 */
function validateLegacyEntry(
  e: RalphEntry,
  index: number
): { ok: true } | { ok: false; message: string } {
  const id = e.id?.trim();
  if (!id) {
    return { ok: false, message: `entries[${index}]: missing id` };
  }
  if (typeof e.type !== "string" || !e.type.trim()) {
    return { ok: false, message: `entries[${index}] id=${id}: missing or empty type` };
  }
  if (typeof e.content !== "string" || !e.content.trim()) {
    return { ok: false, message: `entries[${index}] id=${id}: missing or empty content` };
  }
  if (typeof e.created_at !== "string" || !e.created_at.trim()) {
    return { ok: false, message: `entries[${index}] id=${id}: missing or empty created_at` };
  }
  if (typeof e.updated_at !== "string" || !e.updated_at.trim()) {
    return { ok: false, message: `entries[${index}] id=${id}: missing or empty updated_at` };
  }
  return { ok: true };
}

export function importRalphMemoryIntoDb(db: Database.Database, absPath: string): ImportRalphResult {
  const data = readRalphMemoryFile(absPath);
  const checkLegacy = db.prepare("SELECT 1 AS ok FROM memories WHERE legacy_id = ? LIMIT 1");

  let imported = 0;
  let skipped = 0;
  const errors: string[] = [];

  for (let i = 0; i < data.entries!.length; i++) {
    const e = data.entries![i]!;
    try {
      const v = validateLegacyEntry(e, i);
      if (!v.ok) {
        errors.push(v.message);
        continue;
      }
      const legacyId = e.id.trim();
      if (checkLegacy.get(legacyId) != null) {
        skipped++;
        continue;
      }
      insertMemory(db, {
        type: e.type,
        content: e.content,
        category: e.category ?? null,
        featureId: null,
        source: "import",
        legacyId,
        createdAt: e.created_at,
        updatedAt: e.updated_at,
      });
      imported++;
    } catch (err) {
      const id = e.id?.trim() ?? "?";
      const prefix = `entries[${i}] id=${id}`;
      errors.push(`${prefix}: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  return { imported, skipped, errors };
}

/**
 * Resolve a memory file path under `projectRoot`. Rejects absolute paths and `..`
 * traversal so HTTP callers cannot pivot to arbitrary filesystem paths via `memoryFile`.
 */
export function resolveMemoryFilePath(projectRoot: string, memoryFile?: string): string {
  const root = resolve(projectRoot);
  const rel = memoryFile?.trim() || ".ralph-memory.json";
  if (isAbsolute(rel)) {
    throw Object.assign(new Error("memoryFile must be a relative path within projectRoot"), {
      code: "INVALID_REQUEST",
    });
  }
  const absTarget = resolve(root, rel);
  const relToRoot = relative(root, absTarget);
  if (relToRoot.startsWith("..") || relToRoot === "..") {
    throw Object.assign(new Error("memoryFile must stay within projectRoot"), {
      code: "INVALID_REQUEST",
    });
  }
  return absTarget;
}
