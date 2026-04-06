import { randomUUID } from "node:crypto";
import type Database from "better-sqlite3";
import { parseMemorySource, parseMemoryType } from "./validate.js";
import type { MemorySource, MemoryType } from "./types.js";

export type NewMemoryInput = {
  type: string;
  content: string;
  category?: string | null;
  featureId?: number | null;
  source: string;
  id?: string;
  createdAt?: string;
  updatedAt?: string;
  legacyId?: string | null;
};

export type InsertedMemory = {
  id: string;
  type: MemoryType;
  content: string;
  category: string | null;
  featureId: number | null;
  source: MemorySource;
  createdAt: string;
  updatedAt: string;
};

function isoNow(): string {
  return new Date().toISOString();
}

export function insertMemory(db: Database.Database, input: NewMemoryInput): InsertedMemory {
  const type = parseMemoryType(input.type);
  const source = parseMemorySource(input.source);
  const content = input.content.trim();
  if (!content) {
    throw new Error("content must be non-empty");
  }

  const id = input.id?.trim() || randomUUID();
  const createdAt = input.createdAt || isoNow();
  const updatedAt = input.updatedAt || createdAt;
  const category =
    input.category === undefined || input.category === null
      ? null
      : String(input.category).trim() || null;
  const featureId =
    input.featureId === undefined || input.featureId === null ? null : Number(input.featureId);
  if (featureId !== null && !Number.isInteger(featureId)) {
    throw new Error("featureId must be an integer when set");
  }

  const legacyId = input.legacyId?.trim() || null;

  const stmt = db.prepare(`
    INSERT INTO memories (id, type, content, category, feature_id, source, created_at, updated_at, legacy_id)
    VALUES (@id, @type, @content, @category, @feature_id, @source, @created_at, @updated_at, @legacy_id)
  `);

  stmt.run({
    id,
    type,
    content,
    category,
    feature_id: featureId,
    source,
    created_at: createdAt,
    updated_at: updatedAt,
    legacy_id: legacyId,
  });

  return {
    id,
    type,
    content,
    category,
    featureId,
    source,
    createdAt,
    updatedAt,
  };
}
