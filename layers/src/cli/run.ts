import type {
  AppendRunRequest,
  CompactRequest,
  ImportRalphMemoryRequest,
  RecordRequest,
  RetrieveRequest,
} from "../contracts/v1.js";
import { errorCodeFrom } from "../v1/errors.js";
import {
  handleAppendRun,
  handleCompact,
  handleHealth,
  handleImportRalphMemory,
  handleRecord,
  handleRetrieve,
} from "../v1/handlers.js";
import { parseServeCli } from "../http/parseOptions.js";
import { startLayersHttpServer } from "../http/server.js";
import { parseImportRalphMemoryRequestForCli } from "../v1/importRequest.js";

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

export async function runCli(
  argv: string[],
  env: NodeJS.ProcessEnv,
  stdin: string
): Promise<{ code: number; stdout: string; stderr: string }> {
  if (argv[0] === "v1" && argv[1] === "health") {
    const body = handleHealth();
    return { code: 0, stdout: `${JSON.stringify(body)}\n`, stderr: "" };
  }

  if (argv[0] === "v1" && argv[1] === "serve") {
    try {
      const opts = parseServeCli(argv, env);
      await startLayersHttpServer(env, opts);
      return { code: 0, stdout: "", stderr: "" };
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      return jsonErr("SERVE_FAILED", msg);
    }
  }

  if (argv[0] === "v1" && argv[1] === "append-run") {
    try {
      const req = parseStdinJson<AppendRunRequest>(stdin);
      const out = handleAppendRun(req, env);
      return { code: 0, stdout: `${JSON.stringify(out)}\n`, stderr: "" };
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      return jsonErr("APPEND_RUN_FAILED", msg);
    }
  }

  if (argv[0] === "v1" && argv[1] === "compact") {
    try {
      const req = parseStdinJson<CompactRequest>(stdin);
      const out = handleCompact(req, env);
      return { code: 0, stdout: `${JSON.stringify(out)}\n`, stderr: "" };
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      return jsonErr("COMPACT_FAILED", msg);
    }
  }

  if (argv[0] === "v1" && argv[1] === "retrieve") {
    try {
      const req = parseStdinJson<RetrieveRequest>(stdin);
      const out = await handleRetrieve(req, env);
      return { code: 0, stdout: `${JSON.stringify(out)}\n`, stderr: "" };
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      return jsonErr("RETRIEVE_FAILED", msg);
    }
  }

  if (argv[0] === "v1" && argv[1] === "record") {
    try {
      const req = parseStdinJson<RecordRequest>(stdin);
      const out = handleRecord(req, env);
      return { code: 0, stdout: `${JSON.stringify(out)}\n`, stderr: "" };
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      return jsonErr("RECORD_FAILED", msg);
    }
  }

  if (argv[0] === "v1" && argv[1] === "import-ralph-memory") {
    try {
      const req = parseImportRalphMemoryRequestForCli(stdin, env) as ImportRalphMemoryRequest;
      const cliPath = argv[2];
      const out = handleImportRalphMemory(req, env, cliPath);
      return { code: 0, stdout: `${JSON.stringify(out)}\n`, stderr: "" };
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      const code = errorCodeFrom(e, "IMPORT_FAILED");
      return jsonErr(code, msg);
    }
  }

  const stderr = `Usage:
  layers v1 health
  layers v1 serve [--host=127.0.0.1] [--port=7847]
  layers v1 retrieve < stdin.json
  layers v1 record < stdin.json
  layers v1 append-run < stdin.json
  layers v1 compact < stdin.json
  layers v1 import-ralph-memory [path/to/.ralph-memory.json] < stdin.json
     (stdin optional if LAYERS_PROJECT_ROOT is set; memory path defaults to <root>/.ralph-memory.json)

See docs/LAYERS_SPEC.md
`;
  return { code: 1, stdout: "", stderr };
}
