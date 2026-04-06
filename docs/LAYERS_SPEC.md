# Layers — TypeScript memory service for Ralph

**Revision:** Phase 6 (**Ralph Go** integration) + Phase 5 HTTP. Ralph: **`-layers-enabled`** (see §7.4). **`layers` v0.6.x**: **`layers v1 serve`** (default **`127.0.0.1:7847`**), hybrid embeddings optional (**`LAYERS_OPENAI_API_KEY`**).

This document specifies **Layers**: a **TypeScript** application in this repository that owns **durable memory** (record + retrieve + compaction) for the **Ralph loop**. It defines a **versioned contract** between **Ralph (Go)** and **Layers (TS)** so orchestration stays thin and memory stays evolvable.

**Related material:**

| Source | Role |
|--------|------|
| `oc-spec/spec.md` §8, `spec.algorithms.md` §10 | **OpenClaw** two-layer memory (session history + semantic index, hybrid search, FTS fallback) — **behavioral reference**, not a line-for-line port |
| `oc-spec/spec.contracts.md` §6–7 | Transcript ordering, memory search result shapes — **inspiration** for JSON contracts |
| `docs/RALPH_LOOP_ENHANCEMENTS.md` §6.5–7.12 | Ralph-specific analysis: `progress.txt` vs `.ralph-memory.json`, token pressure, split history vs context |
| `internal/memory/memory.go` | **Current** Ralph memory: flat JSON, `BuildPromptContext`, `[REMEMBER:…]` extraction — **to be superseded or bridged** by Layers |

---

## 1. Goals and non-goals

### 1.1 Goals

- **Optimal retrieval**: given a **query context** (feature description, category, optional plan excerpt), return a **small, diverse, high-signal** set of memories — not the full store (OpenClaw-style **top‑k + hybrid + MMR** intent; see §7).
- **Durable recording**: append or upsert memories with **metadata** (type, source, feature id, timestamps) suitable for indexing.
- **Layer A (history) support**: optional **append-only run log** (Ralph iterations / progress events) with **compaction** into a **bounded “context” view** for prompts.
- **Clear contract**: Ralph speaks only through **documented CLI and/or HTTP APIs** (§8). No “import TS from Go.”
- **Degraded modes**: **FTS/keyword-only** when embeddings are unavailable (per `oc-spec` hybrid fallback theme).

### 1.2 Non-goals (initial phases)

- Replacing Ralph’s **plan** authority (`plan.json` remains Ralph + agent).
- Running inside the **Cursor** process (Layers is a **sidecar** or **CLI** invoked by Ralph).
- Full parity with OpenClaw’s **ingress / routing / delivery** stack — only the **memory planes** concepts transfer.

---

## 2. Conceptual architecture

```
┌─────────────────┐     contract (CLI/HTTP)     ┌──────────────────────────────┐
│  Ralph (Go)     │ ──────────────────────────► │  Layers (TypeScript)        │
│  runIterations  │   retrieve / record / compact│  index + storage + optional │
└─────────────────┘                             │  embedding provider          │
        │                                       └──────────────────────────────┘
        │  plan.json, progress paths                     │
        ▼                                                ▼
   workspace files                              `.layers/` (default) under project root
```

**OpenClaw mapping (`oc-spec/spec.md` §8):**

| OpenClaw | Layers component |
|----------|------------------|
| Layer A — session history (JSONL transcript, compaction) | **Run log** store + **compaction** job producing **context snapshot** |
| Layer B — semantic memory (markdown + SQLite + vector + FTS) | **Memory store** + **chunk index** + **hybrid retrieval** |

---

## 3. Data model

### 3.1 Memory record (logical)

Stable id, type, content, optional structured fields:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (uuid) | Stable identifier |
| `type` | enum | `decision` \| `convention` \| `tradeoff` \| `context` \| `fact` (extensible) |
| `content` | string | Primary text (searchable) |
| `category` | string? | e.g. plan category (`infra`, `ui`) — boosts retrieval |
| `featureId` | number? | Ralph plan feature id when known |
| `source` | enum | `agent` \| `user` \| `ralph` \| `import` |
| `createdAt` / `updatedAt` | ISO8601 | For temporal decay |
| `embedding` | float[]? | Optional; null if not embedded |

**Compatibility:** Import path from legacy **`.ralph-memory.json`** (`internal/memory` entry shape) on first run or via `layers import`.

### 3.2 Chunk index

Long memories may be **split into chunks** (id + parentMemoryId + chunkIndex + text) for FTS/vector rows — aligns with OpenClaw “chunks” in `spec.md` §8.2.2.

### 3.3 Run log (Layer A)

Append-only **JSON lines** (`.jsonl`), one record per event:

| Field | Description |
|-------|-------------|
| `runId` / `sessionKey` | Stable id for a Ralph **invocation** or long-running session |
| `iteration` | Optional iteration number |
| `featureId` | Optional |
| `kind` | `progress` \| `commit` \| `failure` \| `note` \| `structured` |
| `payload` | Object or string |
| `ts` | ISO8601 |

**Invariant (OpenClaw-style):** appends are **ordered**; compaction reads log and writes **derived** files — does not rewrite history silently without policy.

---

## 4. How we record memories (optimally)

1. **Ingestion paths**
   - **From Ralph**: after each iteration, Ralph calls `record` with **structured** payloads (feature id, summary line, optional raw note).
   - **From agent**: continue supporting **`[REMEMBER:TYPE]…[/REMEMBER]`** markers in agent output — Ralph forwards extracted lines to Layers, or Layers accepts **raw stdout snippet** in `record` (configurable).

2. **Normalization**
   - Trim, dedupe **near-identical** text (hash or simhash) before insert.
   - **Upsert policy**: same `featureId` + same `type` + high similarity → update `updatedAt` and merge content **or** skip — configurable.

3. **Indexing**
   - On write: update **SQLite** FTS row; if embeddings enabled, **async** embed + store vector (queue or inline for MVP).

4. **Budget**
   - Optional **max records per project** / **disk budget** (OpenClaw maintenance theme) — LRU or age+lowest-score pruning.

---

## 5. How we retrieve the “right” memories

Retrieval is **query-driven**, not “dump top N by score” only (contrast current `GetRelevant` in `internal/memory/memory.go`).

### 5.1 Query object (from Ralph)

Built from **current work context**:

- `text` — concatenation of **feature title + description + steps** (from `plan.json` slice) or a short **query string**.
- `category` — from plan item when available (fixes empty category in `BuildPromptContext` today).
- `featureId` — optional filter or boost.

### 5.2 Algorithm (order of operations)

Aligned with **`oc-spec/spec.algorithms.md`** hybrid search intent:

1. **Normalize** query (trim, collapse whitespace).
2. **Keyword / FTS** candidate set (broad recall).
3. If embeddings available: **vector** search on chunks → candidate set.
4. **Merge** with weights `vectorWeight`, `textWeight`, **temporal decay** (prefer recent unless strongly matching).
5. **MMR** (Maximal Marginal Relevance) or simplified diversity pass — reduce redundant lines (same idea as oc-spec §10 `mmr`).
6. **Threshold + top‑k**; optional **max output tokens** budget for the **rendered context block**.
7. If embeddings **unavailable**: **FTS-only** path with same top‑k + diversity (OpenClaw “FTS-only fallback”).

### 5.3 Output for Ralph

- **`memories`**: array of `{ id, type, content, score? }` for debugging and UI.
- **`contextBlock`**: single string, fixed template, e.g.  
  `[MEMORY CONTEXT]\n- [DECISION] …\n[END MEMORY CONTEXT]`  
  so Ralph can **prepend** without reformatting (replaces `BuildPromptContext` formatting when Layers is enabled).

---

## 6. Layer A: progress / context compaction

**Problem** (`RALPH_LOOP_ENHANCEMENTS.md`): unbounded `progress.txt` in `@` paths.

**Layers responsibility (configurable):**

- Ingest **structured** lines from Ralph (`record` with `kind: structured`) or tail **file path** registered in config.
- Maintain **`context-snapshot.md`** (or `.txt`) — **bounded** character/lines for prompt use.
- **Compaction** strategies (phased):
  - **MVP**: last *N* JSONL events or last *K* lines.
  - **Next**: rolling summary field updated when size &gt; threshold (optional **LLM** call **outside** Layers or plugin hook — not required for MVP).

Ralph’s `@` path should point at **snapshot** produced by Layers, not the raw archive.

---

## 7. Ralph ↔ Layers contract (versioned)

**Principle:** Ralph only uses **stdio JSON** and/or **HTTP**. Version prefix **`v1`**.

### 7.1 Transport

| Mode | Use |
|------|-----|
| **CLI** | `layers <command> [options]` — JSON on stdin / stdout for batch-friendly Ralph `exec` |
| **HTTP** | `127.0.0.1` server (optional) — **same JSON** as CLI stdin per path (§7.2) |

Environment:

- `LAYERS_PROJECT_ROOT` — required for CLI default cwd resolution.
- `LAYERS_DATA_DIR` — default `.layers` under project root.
- `LAYERS_OPENAI_API_KEY` or `OPENAI_API_KEY` — enables hybrid retrieval embeddings.
- `LAYERS_EMBEDDING_MODEL` — OpenAI embedding model id (optional).
- `LAYERS_EMBEDDING_TIMEOUT_MS` — embed request timeout in ms (default **60000**, max **600000**).
- `LAYERS_HTTP_HOST` / `LAYERS_HTTP_PORT` — HTTP bind (defaults **`127.0.0.1`** / **`7847`**); override CLI: `layers v1 serve --host=… --port=…`.
- `LAYERS_HTTP_ALLOW_REMOTE` — set to **`1`** or **`true`** to allow binding a **non-loopback** host (e.g. `0.0.0.0`); otherwise only **`127.0.0.1`**, **`::1`**, **`localhost`** are accepted.
- `LAYERS_HTTP_MAX_BODY_BYTES` — max JSON body size per request (default **4 MiB**, max **32 MiB**); larger bodies get **413** `PAYLOAD_TOO_LARGE`.

### 7.2 CLI commands (normative v1 sketch)

| Command | Stdin | Stdout | Purpose |
|---------|-------|--------|---------|
| `layers v1 serve` | — | logs to stderr; blocks until SIGINT | **HTTP** — same contract as table below |
| `layers v1 retrieve` | `RetrieveRequest` JSON | `RetrieveResponse` JSON | Hybrid search → `contextBlock` |
| `layers v1 record` | `RecordRequest` JSON | `RecordResponse` JSON | Add/update memories |
| `layers v1 append-run` | `AppendRunRequest` JSON | `AppendRunResponse` JSON | Layer A append (JSONL) |
| `layers v1 compact` | `CompactRequest` JSON | `CompactResponse` JSON | Last‑N JSONL → `context-snapshot.md` |
| `layers v1 import-ralph-memory` | (optional path) | `ImportResponse` JSON | Migrate `.ralph-memory.json` |
| `layers v1 health` | — | `{ "ok": true, "version": "..." }` | Probe |

**HTTP (mirror CLI):** `POST` with **JSON body** = stdin payload for the matching CLI command. **`GET /v1/health`** returns the same shape as **`layers v1 health`**; **`POST /v1/health`** is **405**. Default **loopback-only** bind; see env above. Import **`memoryFile`** must be a **relative path under `projectRoot`** (no absolute paths, no `..`).

| HTTP | Body (same as CLI stdin) |
|------|--------------------------|
| `POST /v1/retrieve` | `RetrieveRequest` |
| `POST /v1/record` | `RecordRequest` |
| `POST /v1/append-run` | `AppendRunRequest` |
| `POST /v1/compact` | `CompactRequest` |
| `POST /v1/import-ralph-memory` | `ImportRalphMemoryRequest` |
| `GET /v1/health` | — |

### 7.3 JSON types (v1)

**RetrieveRequest**

```json
{
  "projectRoot": "/abs/path/to/repo",
  "query": {
    "text": "string",
    "category": "string | null",
    "featureId": 0
  },
  "options": {
    "topK": 10,
    "maxTokens": 2000,
    "mmrLambda": 0.5,
    "embeddingFallbackOk": true,
    "vectorWeight": 0.55,
    "textWeight": 0.45
  }
}
```

When embeddings are available, scores blend **cosine(query, row)** (normalized to [0,1]) with **FTS relevance** using `vectorWeight` / `textWeight` (default 0.55 / 0.45). If the API fails and `embeddingFallbackOk` is true (default), **FTS-only** is used.

**RetrieveResponse**

```json
{
  "contextBlock": "string",
  "memories": [{ "id": "uuid", "type": "decision", "content": "string", "score": 0.0 }],
  "meta": {
    "ftsOnly": false,
    "truncated": false,
    "embeddingModel": "text-embedding-3-small"
  }
}
```

**RecordRequest**

```json
{
  "projectRoot": "/abs/path",
  "entries": [
    {
      "type": "decision",
      "content": "string",
      "category": "ui",
      "featureId": 3,
      "source": "agent"
    }
  ]
}
```

**AppendRunRequest**

```json
{
  "projectRoot": "/abs/path",
  "runLog": "run.jsonl",
  "event": {
    "sessionKey": "ralph-2026-04-06",
    "kind": "structured",
    "iteration": 3,
    "featureId": 12,
    "payload": { "summary": "iteration complete" },
    "ts": "2026-04-06T12:00:00.000Z"
  }
}
```

`kind`: `progress` | `commit` | `failure` | `note` | `structured`. Omit `ts` to let Layers set **now** (ISO8601).

**CompactRequest**

```json
{
  "projectRoot": "/abs/path",
  "runLog": "run.jsonl",
  "snapshot": "context-snapshot.md",
  "maxEvents": 50,
  "maxBytes": 32000
}
```

Reads the last `maxEvents` non-empty lines from the run log, renders markdown (JSON pretty-printed per event), UTF‑8 byte‑truncates to `maxBytes` if needed.

Errors: **exit code non-zero**, stderr human message; stdout may still carry `{ "error": { "code": "...", "message": "..." } }` for machine parsing.

### 7.4 Ralph integration points (Go)

| Location | Action |
|----------|--------|
| Start of iteration (before `agent.Execute`) | If `LAYERS_ENABLED` / config: run **`retrieve`** with query from **current plan feature**; prepend `contextBlock` instead of/in addition to `memory.BuildPromptContext` |
| After agent returns | Parse `[REMEMBER:…]` (existing) → **`record`** batch |
| Optional | **`append-run`** with iteration id, feature id, structured summary |
| End of run / size threshold | **`compact`** (or cron) |

Ralph flags: **`-layers-enabled`**, **`-layers-command`**, **`-layers-url`**, **`-layers-data-dir`**. Config file keys: **`layers_enabled`**, **`layers_command`**, **`layers_url`**, **`layers_data_dir`**. Implementation: **`internal/layers`** (CLI subprocess or HTTP).

---

## 8. Repository layout (this repo)

```
layers/
  package.json
  tsconfig.json
  src/
    cli/
    http/            # optional HTTP server
    v1/              # shared v1 handlers (CLI + HTTP)
    embed/           # OpenAI embeddings (optional)
    retrieve/        # hybrid + MMR
    run/             # JSONL append + compact snapshot
    import/          # .ralph-memory.json
  README.md
docs/
  LAYERS_SPEC.md   # this file
```

---

## 9. Cross-reference checklist

| Topic | oc-spec | Ralph today | Layers |
|-------|---------|-------------|--------|
| Two-layer separation | §8.1 / §8.2 | Mixed `progress` + `.ralph-memory.json` | Run log + memory index |
| Hybrid retrieval | `spec.algorithms.md` §10 | Scored list only | FTS + optional vectors + merge |
| Transcript ordering | `spec.contracts.md` §6.1 | N/A (stateless subprocess) | JSONL append order for run log |
| Bounded context | Maintenance in §8.1.2 | Unbounded `progress.txt` | `compact` → snapshot file |
| Contract to orchestrator | N/A | N/A | §7 CLI/HTTP |

---

## 10. Step-by-step implementation plan

### Phase 0 — Scaffold

- [x] Add `layers/` with **Node + TypeScript**, npm workspace, strict TS, ESLint + Prettier.
- [x] `layers v1 health` CLI — JSON on stdout; **version** read from `layers/package.json`.
- [x] Document **Node version** in `layers/README.md`; tests (**Vitest**), `make layers-test`, CI **`.github/workflows/layers.yml`**.

### Phase 1 — Storage + import

- [x] SQLite schema: `memories` + FTS5 virtual table `memory_fts` (full row content indexed; chunk table deferred to Phase 2+).
- [x] FTS5 on memory content with triggers keeping `memory_fts` in sync.
- [x] `layers v1 import-ralph-memory` — reads legacy `.ralph-memory.json` (Go `internal/memory` shape), idempotent via `legacy_id`.
- [x] `layers v1 record` — stdin JSON per §7.3 `RecordRequest`; writes DB under `<projectRoot>/.layers/` (or `dataDir`).

### Phase 2 — Retrieve (FTS-only)

- [x] `layers v1 retrieve`: stdin JSON `RetrieveRequest` (§7.3) — FTS5 `bm25()` with Porter tokenizer; category + featureId **boost**; **MMR-lite** (word Jaccard) for diversity; **maxTokens** trims `contextBlock` (word-count budget; no minimum that violates the cap).
- [x] **Option guards**: `topK` / `poolSize` / `maxTokens` / `mmrLambda` validated; FTS **MATCH** tokenization uses Unicode letter/number classes to align with `unicode61`.
- [x] **`contextBlock`** uses `[MEMORY CONTEXT]` … `[END MEMORY CONTEXT]` (§5.3).
- [x] Tests: unit (`ftsQuery`, `mmr`, `retrieve`) + CLI integration for `retrieve`.

### Phase 3 — Run log + compact

- [x] `layers v1 append-run`: stdin **`AppendRunRequest`** — append one **`RunEvent`** line to **`run.jsonl`** (default path under data dir; optional `runLog`).
- [x] `layers v1 compact`: stdin **`CompactRequest`** — read last **`maxEvents`** lines from run log, write **`context-snapshot.md`** (default under data dir); UTF‑8 byte cap via **`maxBytes`**.
- [x] Config: **`runLog`**, **`snapshot`**, **`maxEvents`** (default 50, clamped), **`maxBytes`** (default 32k, clamped).

### Phase 4 — Embeddings (optional provider)

- [x] **OpenAI** embedder via `fetch` (`layers/src/embed/openai.ts`); `LAYERS_OPENAI_API_KEY` / `OPENAI_API_KEY`; model via `LAYERS_EMBEDDING_MODEL`.
- [x] **`memories.embedding`** BLOB (Float32); lazy compute + persist on retrieve; hybrid **vector + FTS** with `vectorWeight` / `textWeight`; **`embeddingFallbackOk`** (default true) → FTS on API failure.
- [x] **`meta.embeddingModel`** / **`meta.ftsOnly`** in retrieve response; **`v1 retrieve`** is async (uses `fetch`).

### Phase 5 — HTTP server (optional)

- [x] **`layers v1 serve`** — default **`127.0.0.1:7847`**; **`POST /v1/retrieve`**, **`/record`**, **`/append-run`**, **`/compact`**, **`/import-ralph-memory`** (same JSON as CLI stdin); **`GET /v1/health`**.
- [x] Loopback-only by default; **`LAYERS_HTTP_ALLOW_REMOTE=1`** to bind other hosts.

### Phase 6 — Ralph integration (Go)

- [x] Config: **`-layers-enabled`**, **`-layers-command`**, **`-layers-url`**, **`-layers-data-dir`**; YAML **`layers_enabled`**, **`layers_command`**, **`layers_url`**, **`layers_data_dir`** (see `internal/config`).
- [x] When **`-layers-enabled`**: **`retrieve`** before each iteration (query from current untested plan feature); prepends **`contextBlock`**; on failure falls back to **`memory.BuildPromptContext`**.
- [x] **`[REMEMBER:…]`** → **`record`** batch (skips writing **`.ralph-memory.json`** for those entries when Layers is on).
- [x] **`append-run`** (iteration + feature id) and **`compact`** at end of run (verbose logs failures).

E2E: run **`layers`** on PATH or **`-layers-url`**, enable **`-layers-enabled`**, run Ralph with **`-iterations`** ≥ 1.

---

## 11. Risks and mitigations

| Risk | Mitigation |
|------|------------|
| TS dependency hell | Pin versions; minimal deps; SQLite via `better-sqlite3` or `sql.js` tradeoff documented |
| Embedding API costs | Off by default; FTS-only path always works |
| Double source of truth | Ralph **does not** write `.ralph-memory.json` when Layers enabled; single writer policy |

---

## 12. Open questions

- **Single global vs per-branch** memory (git branch in metadata?) — recommend **project-scoped** first; branch optional field later.
- **Ollama** vs cloud embeddings — plugin interface in Phase 4.

This spec is the **source of truth** for the `layers` package until superseded by a new revision header at the top of this file.
