import type Database from "better-sqlite3";
import { buildFtsMatchQuery, normalizeQueryText } from "./ftsQuery.js";
import { mmrSelect, type MmrItem } from "./mmr.js";

export type RetrieveQueryInput = {
  text: string;
  category: string | null;
  featureId: number;
};

export type RetrieveOptionsInput = {
  topK?: number;
  maxTokens?: number;
  mmrLambda?: number;
};

export type RetrievedMemory = {
  id: string;
  type: string;
  content: string;
  score: number;
};

export type RetrieveMeta = {
  ftsOnly: true;
  truncated: boolean;
};

export function retrieveMemories(
  db: Database.Database,
  query: RetrieveQueryInput,
  options: RetrieveOptionsInput
): { memories: RetrievedMemory[]; contextBlock: string; meta: RetrieveMeta } {
  const topK = options.topK ?? 10;
  const maxTokens = options.maxTokens ?? 2000;
  const mmrLambda = options.mmrLambda ?? 0.5;
  const poolSize = Math.min(50, Math.max(topK * 4, topK));

  const qText = normalizeQueryText(query.text);
  const ftsMatch = buildFtsMatchQuery(qText);
  const catLower = query.category?.trim().toLowerCase() ?? null;
  const featureId = Number.isFinite(query.featureId) ? query.featureId : 0;

  type Row = {
    id: string;
    type: string;
    content: string;
    category: string | null;
    feature_id: number | null;
    bm25: number;
  };

  let rows: Row[] = [];

  if (ftsMatch) {
    const stmt = db.prepare(`
      SELECT
        m.id,
        m.type,
        m.content,
        m.category,
        m.feature_id,
        bm25(memory_fts) AS bm25
      FROM memory_fts
      INNER JOIN memories m ON m.id = memory_fts.memory_id
      WHERE memory_fts MATCH ?
      ORDER BY bm25 ASC
      LIMIT ?
    `);
    rows = stmt.all(ftsMatch, poolSize) as Row[];
  }

  if (rows.length === 0) {
    const fallback = db.prepare(`
      SELECT id, type, content, category, feature_id, 1.0 AS bm25
      FROM memories
      ORDER BY updated_at DESC
      LIMIT ?
    `);
    rows = fallback.all(poolSize) as Row[];
  }

  const boosted = rows.map((r) => {
    const bm = Number(r.bm25);
    // bm25(): smaller is a better match — invert for positive relevance
    let rel = Number.isFinite(bm) ? 1 / (1 + Math.max(0, bm)) : 0.5;
    if (rel === 0 && rows.length > 0) {
      rel = 0.01;
    }
    if (catLower && r.category?.toLowerCase() === catLower) {
      rel *= 1.35;
    }
    if (featureId > 0 && r.feature_id === featureId) {
      rel *= 1.35;
    }
    return { row: r, relevance: rel };
  });

  const maxRel = Math.max(...boosted.map((b) => b.relevance), 1e-9);
  const mmrItems: MmrItem[] = boosted.map((b) => ({
    id: b.row.id,
    relevance: b.relevance / maxRel,
    text: b.row.content,
  }));

  const picked = mmrSelect(mmrItems, topK, mmrLambda);
  const relById = new Map(boosted.map((b) => [b.row.id, b.relevance]));
  const memories: RetrievedMemory[] = picked.map((p) => {
    const r = rows.find((row) => row.id === p.id)!;
    return {
      id: r.id,
      type: r.type,
      content: r.content,
      score: relById.get(r.id) ?? 0,
    };
  });

  const { block, truncated } = buildContextBlock(memories, maxTokens);
  return {
    memories,
    contextBlock: block,
    meta: { ftsOnly: true, truncated },
  };
}

function buildContextBlock(
  memories: RetrievedMemory[],
  maxTokens: number
): { block: string; truncated: boolean } {
  const header = "[MEMORY CONTEXT]\n";
  const footer =
    "\n[END MEMORY CONTEXT]\n" +
    "If you make important decisions or conventions, note them for future runs.\n";
  const lines = memories.map((m) => {
    const label = m.type.toUpperCase();
    return `- [${label}] ${m.content}`;
  });
  let body = lines.join("\n");
  let truncated = false;
  const approxTokens = (header + body + footer).split(/\s+/).length;
  if (approxTokens > maxTokens && memories.length > 0) {
    const budget = Math.max(50, maxTokens - 20);
    const words = body.split(/\s+/);
    body = words.slice(0, budget).join(" ");
    if (words.length > budget) {
      truncated = true;
      body += "\n…";
    }
  }
  if (memories.length === 0) {
    body = "(no matching memories)";
  }
  return { block: header + body + footer, truncated };
}
