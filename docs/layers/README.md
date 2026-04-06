# Layers documentation

**Layers** is the TypeScript memory service in this repository (`layers/`). These pages are the **human-oriented** companion to the normative technical spec.

| Document | Purpose |
|----------|---------|
| **[User guide](user-guide.md)** | Install, run standalone, CLI/HTTP, env vars, data layout, troubleshooting |
| **[Ralph integration](ralph-integration.md)** | How the Go Ralph CLI uses Layers (`-layers-enabled`, flows, fallbacks) |
| **[`../LAYERS_SPEC.md`](../LAYERS_SPEC.md)** | Canonical **v1** contract (phases 0–6, JSON types) |
| **[`../layers_spec_v2.md`](../layers_spec_v2.md)** | **Forward-looking** gaps and evolution (v2 themes) |

**Quick start (with Ralph):** build `layers`, ensure `layers` is on `PATH` or pass `-layers-url`, then run Ralph with `-layers-enabled`. Details: [Ralph integration](ralph-integration.md).
