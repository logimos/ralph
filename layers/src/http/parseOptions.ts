const DEFAULT_HOST = "127.0.0.1";
const DEFAULT_PORT = 7847;

function parseIntOpt(s: string | undefined, fallback: number): number {
  if (!s?.trim()) {
    return fallback;
  }
  const n = Number.parseInt(s, 10);
  if (!Number.isFinite(n) || n <= 0 || n > 65535) {
    return fallback;
  }
  return n;
}

export type HttpServeOptions = {
  host: string;
  port: number;
};

/**
 * CLI: `layers v1 serve [--host=127.0.0.1] [--port=7847]`
 * Env: LAYERS_HTTP_HOST, LAYERS_HTTP_PORT (defaults above).
 */
const LOOPBACK_HOSTS = new Set(["127.0.0.1", "::1", "localhost"]);

/**
 * By default only loopback interfaces are allowed. Set `LAYERS_HTTP_ALLOW_REMOTE=1` to bind other hosts (e.g. `0.0.0.0`).
 */
export function assertBindHostAllowed(host: string, env: NodeJS.ProcessEnv): void {
  const allow = env.LAYERS_HTTP_ALLOW_REMOTE?.trim();
  if (allow === "1" || allow?.toLowerCase() === "true") {
    return;
  }
  if (LOOPBACK_HOSTS.has(host)) {
    return;
  }
  throw new Error(
    `refusing to bind HTTP on "${host}" (loopback only). Set LAYERS_HTTP_ALLOW_REMOTE=1 to override.`
  );
}

export function parseServeCli(argv: string[], env: NodeJS.ProcessEnv): HttpServeOptions {
  let host = env.LAYERS_HTTP_HOST?.trim() || DEFAULT_HOST;
  let port = parseIntOpt(env.LAYERS_HTTP_PORT, DEFAULT_PORT);

  for (const arg of argv.slice(2)) {
    if (arg.startsWith("--host=")) {
      host = arg.slice("--host=".length).trim() || host;
    } else if (arg.startsWith("--port=")) {
      port = parseIntOpt(arg.slice("--port=".length), port);
    }
  }

  assertBindHostAllowed(host, env);
  return { host, port };
}
