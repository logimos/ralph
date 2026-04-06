/**
 * Text embedding provider (OpenAI, Ollama, etc.).
 */
export type Embedder = {
  /** Model id for logging / meta */
  readonly model: string;
  /** Embed many texts in one provider call when possible */
  embed(texts: string[]): Promise<number[][]>;
};
