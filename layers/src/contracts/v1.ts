export type RecordRequest = {
  projectRoot: string;
  dataDir?: string;
  entries: Array<{
    type: string;
    content: string;
    category?: string | null;
    featureId?: number | null;
    source: string;
  }>;
};

export type RecordResponse = {
  ok: true;
  count: number;
  inserted: Array<{
    id: string;
    type: string;
    content: string;
    category: string | null;
    featureId: number | null;
    source: string;
    createdAt: string;
    updatedAt: string;
  }>;
};

export type ImportRalphMemoryRequest = {
  projectRoot: string;
  dataDir?: string;
  /** Path relative to project root or absolute; default .ralph-memory.json */
  memoryFile?: string;
};

export type ImportRalphMemoryResponse = {
  ok: true;
  path: string;
  imported: number;
  skipped: number;
  errors: string[];
};

export type ErrorResponse = {
  ok: false;
  error: { code: string; message: string };
};
