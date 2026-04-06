import { appendFileSync, mkdirSync } from "node:fs";
import { dirname } from "node:path";
import type { RunEvent } from "./types.js";
import { RUN_KINDS } from "./types.js";

export function validateRunEvent(raw: unknown): RunEvent {
  if (!raw || typeof raw !== "object") {
    throw new Error("event must be an object");
  }
  const o = raw as Record<string, unknown>;
  const sessionKey = typeof o.sessionKey === "string" ? o.sessionKey.trim() : "";
  if (!sessionKey) {
    throw new Error("event.sessionKey is required");
  }
  const kind = typeof o.kind === "string" ? o.kind.trim().toLowerCase() : "";
  if (!RUN_KINDS.includes(kind as (typeof RUN_KINDS)[number])) {
    throw new Error(`event.kind must be one of: ${RUN_KINDS.join(", ")}`);
  }
  const iteration =
    typeof o.iteration === "number" && Number.isInteger(o.iteration) ? o.iteration : undefined;
  const featureId =
    typeof o.featureId === "number" && Number.isInteger(o.featureId) ? o.featureId : undefined;
  const ts = typeof o.ts === "string" && o.ts.trim() ? o.ts.trim() : new Date().toISOString();
  const ev: RunEvent = {
    sessionKey,
    kind: kind as RunEvent["kind"],
    ts,
    ...(iteration !== undefined ? { iteration } : {}),
    ...(featureId !== undefined ? { featureId } : {}),
  };
  if ("payload" in o) {
    ev.payload = o.payload;
  }
  return ev;
}

export function appendRunEvent(absLogPath: string, event: RunEvent): void {
  mkdirSync(dirname(absLogPath), { recursive: true });
  const line = `${JSON.stringify(event)}\n`;
  appendFileSync(absLogPath, line, "utf8");
}
