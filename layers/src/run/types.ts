export const RUN_KINDS = ["progress", "commit", "failure", "note", "structured"] as const;
export type RunKind = (typeof RUN_KINDS)[number];

export type RunEvent = {
  sessionKey: string;
  iteration?: number;
  featureId?: number;
  kind: RunKind;
  payload?: unknown;
  ts: string;
};
