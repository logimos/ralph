import { describe, expect, it } from "vitest";
import { assertBindHostAllowed, parseServeCli } from "./parseOptions.js";

describe("parseServeCli", () => {
  it("defaults to loopback and port 7847", () => {
    const o = parseServeCli(["v1", "serve"], {});
    expect(o.host).toBe("127.0.0.1");
    expect(o.port).toBe(7847);
  });

  it("parses --host and --port", () => {
    const o = parseServeCli(["v1", "serve", "--host=127.0.0.1", "--port=9000"], {});
    expect(o.host).toBe("127.0.0.1");
    expect(o.port).toBe(9000);
  });

  it("refuses non-loopback host without override", () => {
    expect(() => parseServeCli(["v1", "serve", "--host=0.0.0.0"], {})).toThrow(/loopback only/);
  });

  it("allows non-loopback when LAYERS_HTTP_ALLOW_REMOTE=1", () => {
    const o = parseServeCli(["v1", "serve", "--host=0.0.0.0", "--port=9001"], {
      LAYERS_HTTP_ALLOW_REMOTE: "1",
    });
    expect(o.host).toBe("0.0.0.0");
    expect(o.port).toBe(9001);
  });
});

describe("assertBindHostAllowed", () => {
  it("allows localhost with no env", () => {
    expect(() => assertBindHostAllowed("localhost", {})).not.toThrow();
  });
});
