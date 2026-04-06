import { existsSync } from "node:fs";
import { resolve } from "node:path";
import type {
  AppendRunRequest,
  CompactRequest,
  ImportRalphMemoryRequest,
  RecordRequest,
  RetrieveRequest,
} from "../contracts/v1.js";
import { importRalphMemoryIntoDb, resolveMemoryFilePath } from "../import/ralphMemory.js";
import { getPackageVersion } from "../version.js";
import { insertMemory, type InsertedMemory } from "../store/insert.js";
import { openDatabase } from "../store/open.js";
import { effectiveDataDirOverride, resolveDataDir } from "../store/paths.js";
import { createOpenAiEmbedder } from "../embed/openai.js";
import { retrieveMemories } from "../retrieve/retrieve.js";
import { appendRunEvent, validateRunEvent } from "../run/append.js";
import { normalizeCompactOptions, readLastJsonlLines, writeSnapshot } from "../run/compact.js";
import {
  DEFAULT_RUN_LOG_REL,
  DEFAULT_SNAPSHOT_REL,
  resolvePathInDataDir,
} from "../run/resolvePaths.js";

export type HealthResult = {
  ok: true;
  version: string;
  service: "layers";
};

export function handleHealth(): HealthResult {
  return {
    ok: true,
    version: getPackageVersion(),
    service: "layers",
  };
}

export async function handleRetrieve(
  req: RetrieveRequest,
  env: NodeJS.ProcessEnv
): Promise<Record<string, unknown>> {
  if (!req.projectRoot || typeof req.projectRoot !== "string") {
    throw Object.assign(new Error("retrieve: projectRoot is required"), {
      code: "INVALID_REQUEST",
    });
  }
  if (!req.query || typeof req.query.text !== "string") {
    throw Object.assign(new Error("retrieve: query.text is required"), { code: "INVALID_REQUEST" });
  }
  const root = resolve(req.projectRoot);
  const dataDir = resolveDataDir(root, effectiveDataDirOverride(req.dataDir, env));
  const db = openDatabase(dataDir);
  try {
    const q = req.query;
    const embedder = createOpenAiEmbedder(env);
    const embeddingCtx = embedder ? { embedder } : null;
    const { memories, contextBlock, meta } = await retrieveMemories(
      db,
      {
        text: q.text,
        category: q.category ?? null,
        featureId: typeof q.featureId === "number" ? q.featureId : 0,
      },
      {
        topK: req.options?.topK,
        maxTokens: req.options?.maxTokens,
        mmrLambda: req.options?.mmrLambda,
        embeddingFallbackOk: req.options?.embeddingFallbackOk,
        vectorWeight: req.options?.vectorWeight,
        textWeight: req.options?.textWeight,
      },
      embeddingCtx
    );
    return {
      ok: true,
      contextBlock,
      memories,
      meta,
    };
  } finally {
    db.close();
  }
}

export function handleRecord(req: RecordRequest, env: NodeJS.ProcessEnv): Record<string, unknown> {
  if (!req.projectRoot || typeof req.projectRoot !== "string") {
    throw Object.assign(new Error("record: projectRoot is required"), { code: "INVALID_REQUEST" });
  }
  if (!Array.isArray(req.entries) || req.entries.length === 0) {
    throw Object.assign(new Error("record: entries must be a non-empty array"), {
      code: "INVALID_REQUEST",
    });
  }
  const root = resolve(req.projectRoot);
  const dataDir = resolveDataDir(root, effectiveDataDirOverride(req.dataDir, env));
  const db = openDatabase(dataDir);
  const inserted: InsertedMemory[] = [];
  try {
    for (const e of req.entries) {
      inserted.push(insertMemory(db, e));
    }
    return {
      ok: true,
      count: inserted.length,
      inserted: inserted.map((m) => ({
        id: m.id,
        type: m.type,
        content: m.content,
        category: m.category,
        featureId: m.featureId,
        source: m.source,
        createdAt: m.createdAt,
        updatedAt: m.updatedAt,
      })),
    };
  } finally {
    db.close();
  }
}

export function handleAppendRun(
  req: AppendRunRequest,
  env: NodeJS.ProcessEnv
): Record<string, unknown> {
  if (!req.projectRoot || typeof req.projectRoot !== "string") {
    throw Object.assign(new Error("append-run: projectRoot is required"), {
      code: "INVALID_REQUEST",
    });
  }
  if (!req.event || typeof req.event !== "object") {
    throw Object.assign(new Error("append-run: event is required"), { code: "INVALID_REQUEST" });
  }
  const root = resolve(req.projectRoot.trim());
  const dataDir = resolveDataDir(root, effectiveDataDirOverride(req.dataDir, env));
  const logPath = resolvePathInDataDir(dataDir, req.runLog, DEFAULT_RUN_LOG_REL);
  const event = validateRunEvent(req.event);
  appendRunEvent(logPath, event);
  return { ok: true, path: logPath };
}

export function handleCompact(
  req: CompactRequest,
  env: NodeJS.ProcessEnv
): Record<string, unknown> {
  if (!req.projectRoot || typeof req.projectRoot !== "string") {
    throw Object.assign(new Error("compact: projectRoot is required"), { code: "INVALID_REQUEST" });
  }
  const root = resolve(req.projectRoot.trim());
  const dataDir = resolveDataDir(root, effectiveDataDirOverride(req.dataDir, env));
  const runLogPath = resolvePathInDataDir(dataDir, req.runLog, DEFAULT_RUN_LOG_REL);
  const snapshotPath = resolvePathInDataDir(dataDir, req.snapshot, DEFAULT_SNAPSHOT_REL);
  const opts = normalizeCompactOptions(req.maxEvents, req.maxBytes);
  const lines = readLastJsonlLines(runLogPath, opts.maxEvents);
  const bytesWritten = writeSnapshot(snapshotPath, lines, opts);
  return {
    ok: true,
    runLogPath,
    snapshotPath,
    eventCount: lines.length,
    bytesWritten,
  };
}

export function handleImportRalphMemory(
  req: ImportRalphMemoryRequest,
  env: NodeJS.ProcessEnv,
  cliMemoryPathArg?: string
): Record<string, unknown> {
  if (typeof req.projectRoot !== "string" || !req.projectRoot.trim()) {
    throw Object.assign(
      new Error("import-ralph-memory: projectRoot is required and must be a non-empty string"),
      { code: "INVALID_REQUEST" }
    );
  }
  req = { ...req, projectRoot: req.projectRoot.trim() };

  let memoryPath: string;
  if (cliMemoryPathArg) {
    const p = cliMemoryPathArg.trim();
    memoryPath = resolve(p);
  } else if (req.memoryFile) {
    memoryPath = resolveMemoryFilePath(resolve(req.projectRoot), req.memoryFile);
  } else {
    memoryPath = resolveMemoryFilePath(resolve(req.projectRoot));
  }

  if (!existsSync(memoryPath)) {
    throw Object.assign(new Error(`memory file not found: ${memoryPath}`), { code: "NOT_FOUND" });
  }

  const root = resolve(req.projectRoot);
  const dataDir = resolveDataDir(root, effectiveDataDirOverride(req.dataDir, env));
  const db = openDatabase(dataDir);
  try {
    const result = importRalphMemoryIntoDb(db, memoryPath);
    return {
      ok: true,
      path: memoryPath,
      imported: result.imported,
      skipped: result.skipped,
      errors: result.errors,
    };
  } finally {
    db.close();
  }
}
