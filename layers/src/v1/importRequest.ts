import type { ImportRalphMemoryRequest } from "../contracts/v1.js";

export function parseImportRalphMemoryRequestForCli(
  stdin: string,
  env: NodeJS.ProcessEnv
): ImportRalphMemoryRequest {
  const raw = stdin.trim();
  if (raw) {
    return JSON.parse(raw) as ImportRalphMemoryRequest;
  }
  const rootEnv = env.LAYERS_PROJECT_ROOT?.trim();
  if (!rootEnv) {
    throw Object.assign(
      new Error("import-ralph-memory: provide JSON on stdin or set LAYERS_PROJECT_ROOT"),
      { code: "INVALID_REQUEST" }
    );
  }
  return { projectRoot: rootEnv };
}
