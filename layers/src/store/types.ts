export const MEMORY_TYPES = ["decision", "convention", "tradeoff", "context", "fact"] as const;
export type MemoryType = (typeof MEMORY_TYPES)[number];

export const MEMORY_SOURCES = ["agent", "user", "ralph", "import"] as const;
export type MemorySource = (typeof MEMORY_SOURCES)[number];

export type MemoryRow = {
  id: string;
  type: MemoryType;
  content: string;
  category: string | null;
  featureId: number | null;
  source: MemorySource;
  createdAt: string;
  updatedAt: string;
  legacyId: string | null;
};
