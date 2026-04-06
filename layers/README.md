# Layers

TypeScript service for **recording and retrieving** project memory for the [Ralph](https://github.com/logimos/ralph) loop. Ralph integrates via a **versioned CLI/HTTP contract** — see the canonical spec:

**[docs/LAYERS_SPEC.md](../docs/LAYERS_SPEC.md)**

## Requirements

- **Node.js 20+** (see `engines` in `package.json`)

## Development

From the **repository root** (npm workspaces):

```bash
npm ci
npm run build -w layers
npm run test -w layers
npm run lint -w layers
npm run format:check -w layers
```

Or use **Make**:

```bash
make layers-test
```

## CLI

After `npm run build -w layers`:

```bash
node layers/dist/cli/main.js v1 health
```

## HTTP (Phase 5)

Same JSON bodies as CLI stdin, via **`POST`**:

| Path                      | Body                       |
| ------------------------- | -------------------------- |
| `/v1/retrieve`            | `RetrieveRequest`          |
| `/v1/record`              | `RecordRequest`            |
| `/v1/append-run`          | `AppendRunRequest`         |
| `/v1/compact`             | `CompactRequest`           |
| `/v1/import-ralph-memory` | `ImportRalphMemoryRequest` |

**`GET /v1/health`** — same JSON as `v1 health`.

Start (default **`127.0.0.1:7847`**):

```bash
node layers/dist/cli/main.js v1 serve
# optional: --host=127.0.0.1 --port=7847
# env: LAYERS_HTTP_HOST, LAYERS_HTTP_PORT, LAYERS_HTTP_ALLOW_REMOTE=1 for non-loopback
```

**Phase 1 — record** (stdin JSON, see `docs/LAYERS_SPEC.md` §7.3):

```bash
echo '{"projectRoot":"/abs/repo","entries":[{"type":"decision","content":"Use SQLite","source":"agent"}]}' \
  | node layers/dist/cli/main.js v1 record
```

Database: `<projectRoot>/.layers/memory.db` unless overridden by `dataDir` in the JSON request or the **`LAYERS_DATA_DIR`** environment variable (see `docs/LAYERS_SPEC.md` §7.1).

**Phase 1 — import** legacy Ralph memory (`.ralph-memory.json` from `internal/memory`):

```bash
echo '{"projectRoot":"/abs/repo"}' | node layers/dist/cli/main.js v1 import-ralph-memory
# or explicit file:
echo '{"projectRoot":"/abs/repo"}' | node layers/dist/cli/main.js v1 import-ralph-memory /path/to/.ralph-memory.json
```

If stdin is omitted, set `LAYERS_PROJECT_ROOT` and optional path as first argument.

**Phase 2 — retrieve** (FTS + MMR-lite, stdin `RetrieveRequest` §7.3):

```bash
echo '{"projectRoot":"/abs/repo","query":{"text":"auth JWT","category":"feature","featureId":2},"options":{"topK":8}}' \
  | node layers/dist/cli/main.js v1 retrieve
```

Response includes `contextBlock`, `memories`, and `meta` (`ftsOnly`, `truncated`, optional `embeddingModel`).

**Hybrid embeddings (Phase 4):** set **`LAYERS_OPENAI_API_KEY`** or **`OPENAI_API_KEY`**. Without a key, retrieval is **FTS-only** (`meta.ftsOnly: true`). Optional **`LAYERS_EMBEDDING_MODEL`** (default `text-embedding-3-small`). **`LAYERS_EMBEDDING_TIMEOUT_MS`** caps embed HTTP calls (default 60000). Tune blend with `options.vectorWeight` / `options.textWeight` (defaults 0.55 / 0.45).

**Phase 3 — run log + snapshot** (Layer A, §3.3 / §7):

Append one JSON line per event (default file `<projectRoot>/.layers/run.jsonl`):

```bash
echo '{"projectRoot":"/abs/repo","event":{"sessionKey":"s1","kind":"structured","payload":{"n":1}}}' \
  | node layers/dist/cli/main.js v1 append-run
```

Rebuild bounded **`context-snapshot.md`** from the last N log lines:

```bash
echo '{"projectRoot":"/abs/repo","maxEvents":50,"maxBytes":32000}' \
  | node layers/dist/cli/main.js v1 compact
```

Default snapshot path: `<projectRoot>/.layers/context-snapshot.md`. Override with `runLog` / `snapshot` (absolute or relative to the data directory).

## Status

- **Phase 0**: scaffold, health, lint, tests, CI.
- **Phase 1**: SQLite + FTS5, `v1 record`, `v1 import-ralph-memory` (idempotent).
- **Phase 2**: `v1 retrieve` — FTS ranking, category/feature boosts, MMR-lite, `contextBlock` + token budget.
- **Phase 3**: `v1 append-run`, `v1 compact` — JSONL run log + markdown snapshot.
- **Phase 4**: optional OpenAI embeddings + hybrid retrieve; lazy embedding cache in SQLite.
- **Phase 5**: optional HTTP (`v1 serve`, loopback by default) — same JSON as CLI.
- **Next**: Phase 6 — Ralph Go integration (`docs/LAYERS_SPEC.md`).

## Quick links

- OpenClaw reference (in-repo): `oc-spec/spec.md` §8, `spec.algorithms.md` §10
- Ralph memory today: `internal/memory/memory.go`
- Loop analysis: `docs/RALPH_LOOP_ENHANCEMENTS.md` §6.5–7.12
