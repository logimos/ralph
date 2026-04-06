import { readFileSync } from "node:fs";
import { resolve } from "node:path";
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
export function importRalphMemoryIntoDb(db: Database.Database, absPath: string): ImportRalphResult {
  const data = readRalphMemoryFile(absPath);
  const checkLegacy = db.prepare("SELECT 1 AS ok FROM memories WHERE legacy_id = ? LIMIT 1");

  let imported = 0;
  let skipped = 0;
  const errors: string[] = [];

  for (const e of data.entries!) {
    try {
      const legacyId = e.id?.trim();
      if (!legacyId) {
        errors.push("entry missing id");
        continue;
      }
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
      errors.push(err instanceof Error ? err.message : String(err));
    }
  }

  return { imported, skipped, errors };
}

export function resolveMemoryFilePath(projectRoot: string, memoryFile?: string): string {
  const rel = memoryFile?.trim() || ".ralph-memory.json";
  return resolve(projectRoot, rel);
}
