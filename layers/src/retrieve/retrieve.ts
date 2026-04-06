import type Database from "better-sqlite3";
import type { Embedder } from "../embed/types.js";
import { buildFtsMatchQuery, normalizeQueryText } from "./ftsQuery.js";
import { mmrSelect, type MmrItem } from "./mmr.js";
import {
  computePoolSize,
  normalizeMaxTokens,
  normalizeMmrLambda,
  normalizeTopK,
} from "./options.js";
import { blobToFloat32Array, cosineSimilarity, float32ArrayToNumbers } from "./vector.js";

export type RetrieveQueryInput = {
  text: string;
  category: string | null;
  featureId: number;
};

export type RetrieveOptionsInput = {
  topK?: number;
  maxTokens?: number;
  mmrLambda?: number;
  /** When false and embeddings fail, throw instead of FTS-only (default true) */
  embeddingFallbackOk?: boolean;
  /** Hybrid weights (default 0.55 / 0.45) */
  vectorWeight?: number;
  textWeight?: number;
};

export type RetrieveEmbeddingContext = {
  embedder: Embedder;
};

export type RetrievedMemory = {
  id: string;
  type: string;
  content: string;
  score: number;
};

export type RetrieveMeta = {
  ftsOnly: boolean;
  truncated: boolean;
  embeddingModel?: string;
};

type Row = {
  id: string;
  type: string;
  content: string;
  category: string | null;
  feature_id: number | null;
  bm25: number;
  embedding: Buffer | null;
};

function normalizeHybridWeights(
  vw?: number,
  tw?: number
): { vectorWeight: number; textWeight: number } {
  if (vw === undefined && tw === undefined) {
    return { vectorWeight: 0.55, textWeight: 0.45 };
  }
  if (vw !== undefined && tw !== undefined && Number.isFinite(vw) && Number.isFinite(tw)) {
    const sum = vw + tw;
    if (sum <= 0) {
      return { vectorWeight: 0.55, textWeight: 0.45 };
    }
    return { vectorWeight: vw / sum, textWeight: tw / sum };
  }
  if (vw !== undefined && Number.isFinite(vw)) {
    const vv = Math.min(1, Math.max(0, vw));
    return { vectorWeight: vv, textWeight: 1 - vv };
  }
  if (tw !== undefined && Number.isFinite(tw)) {
    const tt = Math.min(1, Math.max(0, tw));
    return { vectorWeight: 1 - tt, textWeight: tt };
  }
  return { vectorWeight: 0.55, textWeight: 0.45 };
}

export async function retrieveMemories(
  db: Database.Database,
  query: RetrieveQueryInput,
  options: RetrieveOptionsInput,
  embeddingCtx?: RetrieveEmbeddingContext | null
): Promise<{ memories: RetrievedMemory[]; contextBlock: string; meta: RetrieveMeta }> {
  const topK = normalizeTopK(options.topK);
  const maxTokens = normalizeMaxTokens(options.maxTokens);
  const mmrLambda = normalizeMmrLambda(options.mmrLambda);
  const poolSize = computePoolSize(topK);
  const fallbackOk = options.embeddingFallbackOk !== false;
  const { vectorWeight, textWeight } = normalizeHybridWeights(
    options.vectorWeight,
    options.textWeight
  );

  const qText = normalizeQueryText(query.text);
  const ftsMatch = buildFtsMatchQuery(qText);
  const catLower = query.category?.trim().toLowerCase() ?? null;
  const featureId = Number.isFinite(query.featureId) ? query.featureId : 0;

  let rows: Row[] = [];

  if (ftsMatch) {
    const stmt = db.prepare(`
      SELECT
        m.id,
        m.type,
        m.content,
        m.category,
        m.feature_id,
        bm25(memory_fts) AS bm25,
        m.embedding
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
      SELECT id, type, content, category, feature_id, 1.0 AS bm25, embedding
      FROM memories
      ORDER BY updated_at DESC
      LIMIT ?
    `);
    rows = fallback.all(poolSize) as Row[];
  }

  const ftsBoosted = rows.map((r) => {
    const bm = Number(r.bm25);
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
    return { row: r, ftsRel: rel };
  });

  const maxFts = Math.max(...ftsBoosted.map((b) => b.ftsRel), 1e-9);

  let ftsOnly = true;
  let embeddingModel: string | undefined;

  const updateStmt = db.prepare(`UPDATE memories SET embedding = ? WHERE id = ?`);

  if (embeddingCtx && rows.length > 0) {
    const { embedder } = embeddingCtx;
    embeddingModel = embedder.model;
    try {
      const qEmb = (await embedder.embed([qText]))[0];
      if (!qEmb || qEmb.length === 0) {
        throw new Error("empty query embedding");
      }

      const toEmbed: { id: string; content: string }[] = [];
      for (const b of ftsBoosted) {
        const blob = b.row.embedding;
        if (!blob || blob.length === 0) {
          toEmbed.push({ id: b.row.id, content: b.row.content });
        }
      }

      if (toEmbed.length > 0) {
        const byId = new Map(ftsBoosted.map((b) => [b.row.id, b]));
        const texts = toEmbed.map((x) => x.content);
        const vectors = await embedder.embed(texts);
        for (let i = 0; i < toEmbed.length; i++) {
          const buf = Buffer.from(Float32Array.from(vectors[i]!).buffer);
          const id = toEmbed[i]!.id;
          updateStmt.run(buf, id);
          const hit = byId.get(id);
          if (hit) {
            hit.row.embedding = buf;
          }
        }
      }

      ftsOnly = false;

      for (const b of ftsBoosted) {
        const arr = blobToFloat32Array(b.row.embedding);
        if (!arr) {
          b.ftsRel = textWeight * (b.ftsRel / maxFts);
          continue;
        }
        const nums = float32ArrayToNumbers(arr);
        const cos = cosineSimilarity(qEmb, nums);
        const vecNorm = (cos + 1) / 2;
        const ftsNorm = b.ftsRel / maxFts;
        b.ftsRel = vectorWeight * vecNorm + textWeight * ftsNorm;
      }
    } catch (e) {
      if (!fallbackOk) {
        throw e;
      }
      ftsOnly = true;
      embeddingModel = undefined;
      for (const b of ftsBoosted) {
        const bm = Number(b.row.bm25);
        let rel = Number.isFinite(bm) ? 1 / (1 + Math.max(0, bm)) : 0.5;
        if (rel === 0 && rows.length > 0) {
          rel = 0.01;
        }
        if (catLower && b.row.category?.toLowerCase() === catLower) {
          rel *= 1.35;
        }
        if (featureId > 0 && b.row.feature_id === featureId) {
          rel *= 1.35;
        }
        b.ftsRel = rel;
      }
    }
  }

  const maxRel = Math.max(...ftsBoosted.map((b) => b.ftsRel), 1e-9);
  const mmrItems: MmrItem[] = ftsBoosted.map((b) => ({
    id: b.row.id,
    relevance: b.ftsRel / maxRel,
    text: b.row.content,
  }));

  const picked = mmrSelect(mmrItems, topK, mmrLambda);
  const relById = new Map(ftsBoosted.map((b) => [b.row.id, b.ftsRel]));
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
    meta: {
      ftsOnly,
      truncated,
      ...(embeddingModel ? { embeddingModel } : {}),
    },
  };
}

function wordCount(s: string): number {
  const t = s.trim();
  return t.length === 0 ? 0 : t.split(/\s+/).length;
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
  const fullBody = memories.length === 0 ? "(no matching memories)" : lines.join("\n");
  const fullBlock = header + fullBody + footer;

  if (wordCount(fullBlock) <= maxTokens) {
    return { block: fullBlock, truncated: false };
  }

  const bodyWords = fullBody.trim().length === 0 ? [] : fullBody.trim().split(/\s+/);
  const marker = "\n…";

  for (let n = bodyWords.length - 1; n >= 0; n--) {
    const bodyPart = bodyWords.slice(0, n).join(" ") + marker;
    const block = header + bodyPart + footer;
    if (wordCount(block) <= maxTokens) {
      return { block, truncated: true };
    }
  }

  const allWords = fullBlock.trim().split(/\s+/);
  const clipped = allWords.slice(0, Math.max(0, maxTokens)).join(" ");
  const needEllipsis = allWords.length > maxTokens;
  return {
    block: needEllipsis ? `${clipped}\n…` : clipped,
    truncated: needEllipsis || wordCount(fullBlock) > maxTokens,
  };
}
