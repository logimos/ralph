import { existsSync } from "node:fs";
import { resolve } from "node:path";
import type { ImportRalphMemoryRequest, RecordRequest } from "../contracts/v1.js";
import { importRalphMemoryIntoDb, resolveMemoryFilePath } from "../import/ralphMemory.js";
import { getPackageVersion } from "../version.js";
import { insertMemory } from "../store/insert.js";
import { openDatabase } from "../store/open.js";
import { resolveDataDir } from "../store/paths.js";

export type HealthResult = {
  ok: true;
  version: string;
  service: "layers";
};

function jsonErr(code: string, message: string): { code: number; stdout: string; stderr: string } {
  const body = JSON.stringify({ ok: false, error: { code, message } });
  return { code: 1, stdout: `${body}\n`, stderr: `${message}\n` };
}

function parseStdinJson<T>(raw: string): T {
  const trimmed = raw.trim();
  if (!trimmed) {
    throw new Error("stdin is empty (expected JSON)");
  }
  return JSON.parse(trimmed) as T;
}

export function runCli(
  argv: string[],
  env: NodeJS.ProcessEnv,
  stdin: string
): { code: number; stdout: string; stderr: string } {
  if (argv[0] === "v1" && argv[1] === "health") {
    const body: HealthResult = {
      ok: true,
      version: getPackageVersion(),
      service: "layers",
    };
    return { code: 0, stdout: `${JSON.stringify(body)}\n`, stderr: "" };
  }

  if (argv[0] === "v1" && argv[1] === "record") {
    try {
      const req = parseStdinJson<RecordRequest>(stdin);
      if (!req.projectRoot || typeof req.projectRoot !== "string") {
        return jsonErr("INVALID_REQUEST", "record: projectRoot is required");
      }
      if (!Array.isArray(req.entries) || req.entries.length === 0) {
        return jsonErr("INVALID_REQUEST", "record: entries must be a non-empty array");
      }
      const root = resolve(req.projectRoot);
      const dataDir = resolveDataDir(root, req.dataDir);
      const db = openDatabase(dataDir);
      const inserted = [];
      for (const e of req.entries) {
        inserted.push(insertMemory(db, e));
      }
      db.close();
      const out = {
        ok: true as const,
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
      return { code: 0, stdout: `${JSON.stringify(out)}\n`, stderr: "" };
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      return jsonErr("RECORD_FAILED", msg);
    }
  }

  if (argv[0] === "v1" && argv[1] === "import-ralph-memory") {
    try {
      let req: ImportRalphMemoryRequest;
      const raw = stdin.trim();
      if (raw) {
        req = parseStdinJson<ImportRalphMemoryRequest>(raw);
      } else {
        const rootEnv = env.LAYERS_PROJECT_ROOT?.trim();
        if (!rootEnv) {
          return jsonErr(
            "INVALID_REQUEST",
            "import-ralph-memory: provide JSON on stdin or set LAYERS_PROJECT_ROOT"
          );
        }
        req = { projectRoot: rootEnv };
      }

      let memoryPath: string;
      if (argv[2]) {
        const p = argv[2].trim();
        memoryPath = resolve(p);
      } else if (req.memoryFile) {
        memoryPath = resolveMemoryFilePath(resolve(req.projectRoot), req.memoryFile);
      } else {
        memoryPath = resolveMemoryFilePath(resolve(req.projectRoot));
      }

      if (!existsSync(memoryPath)) {
        return jsonErr("NOT_FOUND", `memory file not found: ${memoryPath}`);
      }

      const root = resolve(req.projectRoot);
      const dataDir = resolveDataDir(root, req.dataDir);
      const db = openDatabase(dataDir);
      const result = importRalphMemoryIntoDb(db, memoryPath);
      db.close();

      const out = {
        ok: true as const,
        path: memoryPath,
        imported: result.imported,
        skipped: result.skipped,
        errors: result.errors,
      };
      return { code: 0, stdout: `${JSON.stringify(out)}\n`, stderr: "" };
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      return jsonErr("IMPORT_FAILED", msg);
    }
  }

  const stderr = `Usage:
  layers v1 health
  layers v1 record < stdin.json
  layers v1 import-ralph-memory [path/to/.ralph-memory.json] < stdin.json
     (stdin optional if LAYERS_PROJECT_ROOT is set; memory path defaults to <root>/.ralph-memory.json)

See docs/LAYERS_SPEC.md
`;
  return { code: 1, stdout: "", stderr };
}
