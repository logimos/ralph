import type { Embedder } from "./types.js";

const DEFAULT_MODEL = "text-embedding-3-small";
const DEFAULT_TIMEOUT_MS = 60_000;
const MAX_ERROR_BODY_CHARS = 512;

function truncateBody(s: string): string {
  if (s.length <= MAX_ERROR_BODY_CHARS) {
    return s;
  }
  return `${s.slice(0, MAX_ERROR_BODY_CHARS)}…`;
}

function parseTimeoutMs(env: NodeJS.ProcessEnv): number {
  const raw = env.LAYERS_EMBEDDING_TIMEOUT_MS?.trim();
  if (!raw) {
    return DEFAULT_TIMEOUT_MS;
  }
  const n = Number(raw);
  if (!Number.isFinite(n) || n <= 0) {
    return DEFAULT_TIMEOUT_MS;
  }
  return Math.min(n, 600_000);
}

/**
 * OpenAI embeddings via `fetch` (no SDK). Uses `LAYERS_OPENAI_API_KEY` or `OPENAI_API_KEY`.
 * Optional `LAYERS_EMBEDDING_TIMEOUT_MS` (default 60000, max 600000).
 */
export function createOpenAiEmbedder(
  env: NodeJS.ProcessEnv,
  model = env.LAYERS_EMBEDDING_MODEL?.trim() || DEFAULT_MODEL
): Embedder | null {
  const key = env.LAYERS_OPENAI_API_KEY?.trim() || env.OPENAI_API_KEY?.trim();
  if (!key) {
    return null;
  }

  const timeoutMs = parseTimeoutMs(env);

  return {
    model,
    async embed(texts: string[]): Promise<number[][]> {
      if (texts.length === 0) {
        return [];
      }
      const controller = new AbortController();
      const t = setTimeout(() => controller.abort(), timeoutMs);
      try {
        const res = await fetch("https://api.openai.com/v1/embeddings", {
          method: "POST",
          signal: controller.signal,
          headers: {
            Authorization: `Bearer ${key}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ model, input: texts }),
        });
        if (!res.ok) {
          const errText = truncateBody(await res.text());
          throw new Error(`OpenAI embeddings failed: ${res.status} ${errText}`);
        }
        const data = (await res.json()) as {
          data?: Array<{ embedding: number[]; index: number }>;
        };
        if (!data.data || !Array.isArray(data.data)) {
          throw new Error("OpenAI embeddings: invalid response shape");
        }
        const sorted = [...data.data].sort((a, b) => a.index - b.index);
        return sorted.map((d) => d.embedding);
      } catch (e) {
        if (e instanceof Error && e.name === "AbortError") {
          throw new Error(`OpenAI embeddings request timed out after ${timeoutMs}ms`);
        }
        throw e;
      } finally {
        clearTimeout(t);
      }
    },
  };
}
