# Layers — user guide

This guide describes how to **run and operate Layers on its own**, without Ralph. For Ralph-specific wiring, see [Ralph integration](ralph-integration.md).

---

## 1. What Layers does

Layers provides:

- **Layer B (semantic memory):** SQLite + FTS5 (+ optional OpenAI embeddings), **retrieve** with MMR-style diversity, **record** batches of memories.
- **Layer A (run history):** append **JSONL** events, **compact** into a bounded **`context-snapshot.md`**.
- **Import:** one-time migration from legacy **`.ralph-memory.json`**.

Default data lives under **`<projectRoot>/.layers/`** unless overridden.

---

## 2. Requirements

- **Node.js 20+**
- Build from repo root: `npm ci && npm run build -w layers`

The CLI entrypoint after build: `layers/dist/cli/main.js` (or `npx`-style invocation via `node path/to/main.js`).

---

## 3. Command overview (v1)

All commands use the shape **`layers v1 <subcommand>`** and expect **JSON on stdin** (except `health`).

| Subcommand | Stdin | Purpose |
|------------|-------|---------|
| `health` | — | Version and service name |
| `retrieve` | `RetrieveRequest` | Hybrid search → `contextBlock` |
| `record` | `RecordRequest` | Insert memories |
| `import-ralph-memory` | `ImportRalphMemoryRequest` | Migrate `.ralph-memory.json` |
| `append-run` | `AppendRunRequest` | Append one line to run JSONL |
| `compact` | `CompactRequest` | Last N log lines → markdown snapshot |
| `serve` | — | HTTP server (same JSON as POST bodies) |

Normative JSON schemas: **`docs/LAYERS_SPEC.md` §7.3**.

---

## 4. Standalone examples

### Health

```bash
node layers/dist/cli/main.js v1 health
```

### Record a memory

```bash
echo '{"projectRoot":"/abs/path/to/repo","entries":[{"type":"decision","content":"Use PostgreSQL for persistence","source":"agent"}]}' \
  | node layers/dist/cli/main.js v1 record
```

### Retrieve for a query

```bash
echo '{"projectRoot":"/abs/path/to/repo","query":{"text":"database persistence","category":"infra","featureId":3},"options":{"topK":8,"maxTokens":2000}}' \
  | node layers/dist/cli/main.js v1 retrieve
```

### Import legacy file

```bash
echo '{"projectRoot":"/abs/path/to/repo"}' \
  | node layers/dist/cli/main.js v1 import-ralph-memory
```

### HTTP server (optional)

```bash
node layers/dist/cli/main.js v1 serve
# POST http://127.0.0.1:7847/v1/retrieve with same JSON body as CLI stdin
```

See **`layers/README.md`** for path table and defaults.

---

## 5. Environment variables

| Variable | Role |
|----------|------|
| `LAYERS_DATA_DIR` | Override default `.layers` directory (also per-request `dataDir` in JSON) |
| `LAYERS_PROJECT_ROOT` | Used by `import-ralph-memory` when stdin is empty |
| `LAYERS_OPENAI_API_KEY` / `OPENAI_API_KEY` | Enable hybrid embeddings on **retrieve** |
| `LAYERS_EMBEDDING_MODEL` | Embedding model id (default `text-embedding-3-small`) |
| `LAYERS_EMBEDDING_TIMEOUT_MS` | Cap for embedding HTTP calls |
| `LAYERS_HTTP_HOST` / `LAYERS_HTTP_PORT` | Bind address for `v1 serve` |
| `LAYERS_HTTP_ALLOW_REMOTE` | `1` to bind non-loopback hosts |
| `LAYERS_HTTP_MAX_BODY_BYTES` | Max JSON body size for HTTP |

---

## 6. On-disk layout (default)

Under **`<projectRoot>/.layers/`** (or `LAYERS_DATA_DIR`):

| Path | Content |
|------|---------|
| `memory.db` | SQLite + FTS (+ optional embedding blobs) |
| `run.jsonl` | Append-only run events |
| `context-snapshot.md` | Output of **compact** (bounded markdown) |

---

## 7. Troubleshooting

| Symptom | Check |
|---------|--------|
| `retrieve` returns FTS-only | No OpenAI key set; or API error (see `meta.ftsOnly`) |
| `import-ralph-memory` fails path | `memoryFile` must be relative under `projectRoot` (no `..`) |
| HTTP 413 | Request body over `LAYERS_HTTP_MAX_BODY_BYTES` |
| Slow first retrieve with embeddings | Lazy embedding fill; subsequent retrieves reuse stored vectors |

---

## 8. Tests and quality

From repo root: `make layers-test` (build, lint, format check, Vitest).

---

## 9. Further reading

- **`docs/LAYERS_SPEC.md`** — full v1 specification  
- **`docs/layers_spec_v2.md`** — planned gaps / evolution  
- **`oc-spec/`** — OpenClaw reference patterns (behavioral, not a port)
