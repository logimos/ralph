import { afterEach, describe, expect, it } from "vitest";
import { createLayersHttpServer } from "./server.js";

describe("createLayersHttpServer", () => {
  let server: ReturnType<typeof createLayersHttpServer>;

  afterEach(async () => {
    if (server?.listening) {
      await new Promise<void>((resolve, reject) => {
        server.close((err) => (err ? reject(err) : resolve()));
      });
    }
  });

  it("GET /v1/health returns JSON", async () => {
    server = createLayersHttpServer({});
    await new Promise<void>((resolve, reject) => {
      server.listen(0, "127.0.0.1", () => resolve());
      server.on("error", reject);
    });
    const addr = server.address();
    if (!addr || typeof addr === "string") {
      throw new Error("expected socket address");
    }
    const res = await fetch(`http://127.0.0.1:${addr.port}/v1/health`);
    expect(res.status).toBe(200);
    const body = (await res.json()) as { ok: boolean; service: string };
    expect(body.ok).toBe(true);
    expect(body.service).toBe("layers");
  });

  it("POST /v1/record with JSON body", async () => {
    server = createLayersHttpServer({});
    await new Promise<void>((resolve, reject) => {
      server.listen(0, "127.0.0.1", () => resolve());
      server.on("error", reject);
    });
    const addr = server.address();
    if (!addr || typeof addr === "string") {
      throw new Error("expected socket address");
    }
    const tmp = await import("node:fs/promises");
    const os = await import("node:os");
    const path = await import("node:path");
    const dir = await tmp.mkdtemp(path.join(os.tmpdir(), "layers-http-"));
    try {
      const res = await fetch(`http://127.0.0.1:${addr.port}/v1/record`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          projectRoot: dir,
          entries: [{ type: "fact", content: "hello http", source: "agent" }],
        }),
      });
      expect(res.status).toBe(200);
      const body = (await res.json()) as { ok: boolean; count: number };
      expect(body.ok).toBe(true);
      expect(body.count).toBe(1);
    } finally {
      await tmp.rm(dir, { recursive: true, force: true });
    }
  });
});
