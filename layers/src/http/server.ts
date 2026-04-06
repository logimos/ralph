import { createServer, type IncomingMessage, type Server, type ServerResponse } from "node:http";
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
import type { HttpServeOptions } from "./parseOptions.js";

const DEFAULT_MAX_BODY_BYTES = 4 * 1024 * 1024;

function maxBodyBytes(env: NodeJS.ProcessEnv): number {
  const raw = env.LAYERS_HTTP_MAX_BODY_BYTES?.trim();
  if (!raw) {
    return DEFAULT_MAX_BODY_BYTES;
  }
  const n = Number(raw);
  if (!Number.isFinite(n) || n <= 0) {
    return DEFAULT_MAX_BODY_BYTES;
  }
  return Math.min(n, 32 * 1024 * 1024);
}

function readBody(req: IncomingMessage, maxBytes: number): Promise<string> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    let total = 0;
    let finished = false;

    const failTooLarge = (): void => {
      if (finished) {
        return;
      }
      finished = true;
      const err = Object.assign(new Error("Request body exceeds maximum size"), {
        code: "PAYLOAD_TOO_LARGE",
        statusCode: 413,
      });
      reject(err);
    };

    req.on("data", (c) => {
      if (finished) {
        return;
      }
      const chunk = c as Buffer;
      total += chunk.length;
      if (total > maxBytes) {
        failTooLarge();
        return;
      }
      chunks.push(chunk);
    });
    req.on("end", () => {
      if (finished) {
        return;
      }
      if (total > maxBytes) {
        failTooLarge();
        return;
      }
      finished = true;
      resolve(Buffer.concat(chunks).toString("utf8"));
    });
    req.on("error", (err) => {
      if (finished) {
        return;
      }
      finished = true;
      reject(err);
    });
  });
}

function sendJson(res: ServerResponse, status: number, body: unknown): void {
  const payload = JSON.stringify(body);
  res.writeHead(status, {
    "Content-Type": "application/json; charset=utf-8",
    "Content-Length": Buffer.byteLength(payload),
  });
  res.end(payload);
}

const ROUTES: Record<
  string,
  (
    body: unknown,
    env: NodeJS.ProcessEnv
  ) => Promise<Record<string, unknown>> | Record<string, unknown>
> = {
  "/v1/retrieve": async (body, env) =>
    handleRetrieve(body as RetrieveRequest, env) as Promise<Record<string, unknown>>,
  "/v1/record": async (body, env) => handleRecord(body as RecordRequest, env),
  "/v1/append-run": async (body, env) => handleAppendRun(body as AppendRunRequest, env),
  "/v1/compact": async (body, env) => handleCompact(body as CompactRequest, env),
  "/v1/import-ralph-memory": async (body, env) =>
    handleImportRalphMemory(body as ImportRalphMemoryRequest, env),
};

export function createLayersHttpServer(env: NodeJS.ProcessEnv): Server {
  return createServer(async (req, res) => {
    const cleanupRequest = (): void => {
      try {
        req.destroy();
      } catch {
        /* ignore */
      }
    };
    try {
      const url = req.url?.split("?")[0] || "";
      const method = req.method || "";

      if (url === "/v1/health" && method === "GET") {
        sendJson(res, 200, handleHealth());
        return;
      }

      if (url === "/v1/health" && method === "POST") {
        sendJson(res, 405, {
          ok: false,
          error: {
            code: "METHOD_NOT_ALLOWED",
            message: "Use GET for /v1/health",
          },
        });
        return;
      }

      if (method !== "POST") {
        sendJson(res, 405, {
          ok: false,
          error: {
            code: "METHOD_NOT_ALLOWED",
            message: "POST required (GET allowed for /v1/health only)",
          },
        });
        return;
      }
      const handler = ROUTES[url];
      if (!handler) {
        sendJson(res, 404, {
          ok: false,
          error: { code: "NOT_FOUND", message: `Unknown path: ${url}` },
        });
        return;
      }

      const raw = await readBody(req, maxBodyBytes(env));
      let parsed: unknown = {};
      if (raw.trim()) {
        try {
          parsed = JSON.parse(raw) as unknown;
        } catch {
          sendJson(res, 400, {
            ok: false,
            error: { code: "INVALID_JSON", message: "Request body must be JSON" },
          });
          return;
        }
      }

      const out = await handler(parsed, env);
      sendJson(res, 200, out);
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      const code = errorCodeFrom(e, "INTERNAL_ERROR");
      const fromErr =
        e && typeof e === "object" && "statusCode" in e
          ? (e as { statusCode: unknown }).statusCode
          : undefined;
      const statusFromError =
        typeof fromErr === "number" && fromErr >= 400 && fromErr < 600 ? fromErr : undefined;
      const status =
        statusFromError ??
        (code === "INVALID_REQUEST"
          ? 400
          : code === "NOT_FOUND"
            ? 404
            : code === "PAYLOAD_TOO_LARGE"
              ? 413
              : 500);
      sendJson(res, status, { ok: false, error: { code, message: msg } });
      if (code === "PAYLOAD_TOO_LARGE") {
        cleanupRequest();
      }
    }
  });
}

/**
 * Starts HTTP server; resolves when the server closes (e.g. SIGINT).
 */
export function startLayersHttpServer(
  env: NodeJS.ProcessEnv,
  opts: HttpServeOptions
): Promise<void> {
  const server = createLayersHttpServer(env);
  return new Promise((resolve, reject) => {
    server.on("error", reject);
    server.listen(opts.port, opts.host, () => {
      process.stderr.write(
        `layers: listening on http://${opts.host}:${opts.port} (POST /v1/retrieve, /v1/record, …)\n`
      );
    });
    const shutdown = (): void => {
      server.close(() => resolve());
    };
    process.once("SIGINT", shutdown);
    process.once("SIGTERM", shutdown);
  });
}
