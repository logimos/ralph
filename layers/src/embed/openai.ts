import type { Embedder } from "./types.js";

const DEFAULT_MODEL = "text-embedding-3-small";

/**
 * OpenAI embeddings via `fetch` (no SDK). Uses `LAYERS_OPENAI_API_KEY` or `OPENAI_API_KEY`.
 */
export function createOpenAiEmbedder(
  env: NodeJS.ProcessEnv,
  model = env.LAYERS_EMBEDDING_MODEL?.trim() || DEFAULT_MODEL
): Embedder | null {
  const key = env.LAYERS_OPENAI_API_KEY?.trim() || env.OPENAI_API_KEY?.trim();
  if (!key) {
    return null;
  }

  return {
    model,
    async embed(texts: string[]): Promise<number[][]> {
      if (texts.length === 0) {
        return [];
      }
      const res = await fetch("https://api.openai.com/v1/embeddings", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${key}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ model, input: texts }),
      });
      if (!res.ok) {
        const errText = await res.text();
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
    },
  };
}
