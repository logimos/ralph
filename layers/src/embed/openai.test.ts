import { afterEach, describe, expect, it, vi } from "vitest";
import { createOpenAiEmbedder } from "./openai.js";

describe("createOpenAiEmbedder", () => {
  const originalFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("throws a concise message when the request times out", async () => {
    const embedder = createOpenAiEmbedder({
      OPENAI_API_KEY: "sk-test",
      LAYERS_EMBEDDING_TIMEOUT_MS: "100",
    });
    expect(embedder).not.toBeNull();

    globalThis.fetch = (_input: RequestInfo | URL, init?: RequestInit) =>
      new Promise<Response>((_resolve, reject) => {
        const signal = init?.signal;
        if (!signal) {
          reject(new Error("expected AbortSignal"));
          return;
        }
        if (signal.aborted) {
          const err = new Error("The operation was aborted");
          err.name = "AbortError";
          reject(err);
          return;
        }
        signal.addEventListener("abort", () => {
          const err = new Error("The operation was aborted");
          err.name = "AbortError";
          reject(err);
        });
      });

    await expect(embedder!.embed(["x"])).rejects.toThrow(/timed out after 100ms/i);
  });
});
