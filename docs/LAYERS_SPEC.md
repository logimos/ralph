# Layers — TypeScript memory service for Ralph

**Revision:** Phase 2 (`v1 retrieve`, FTS-only) implemented in **`layers` v0.3.0** (see §10).

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
| **HTTP** | `127.0.0.1` server (optional) — same payloads as REST bodies |

Environment:

- `LAYERS_PROJECT_ROOT` — required for CLI default cwd resolution.
- `LAYERS_DATA_DIR` — default `.layers` under project root.

### 7.2 CLI commands (normative v1 sketch)

| Command | Stdin | Stdout | Purpose |
|---------|-------|--------|---------|
| `layers v1 retrieve` | `RetrieveRequest` JSON | `RetrieveResponse` JSON | Hybrid search → `contextBlock` |
| `layers v1 record` | `RecordRequest` JSON | `RecordResponse` JSON | Add/update memories |
| `layers v1 append-run` | `RunEvent` JSON | `AppendResponse` JSON | Layer A append |
| `layers v1 compact` | `CompactRequest` JSON | `CompactResponse` JSON | Regenerate context snapshot |
| `layers v1 import-ralph-memory` | (optional path) | `ImportResponse` JSON | Migrate `.ralph-memory.json` |
| `layers v1 health` | — | `{ "ok": true, "version": "..." }` | Probe |

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
    "embeddingFallbackOk": true
  }
}
```

**RetrieveResponse**

```json
{
  "contextBlock": "string",
  "memories": [{ "id": "uuid", "type": "decision", "content": "string", "score": 0.0 }],
  "meta": {
    "ftsOnly": false,
    "truncated": false
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

Errors: **exit code non-zero**, stderr human message; stdout may still carry `{ "error": { "code": "...", "message": "..." } }` for machine parsing.

### 7.4 Ralph integration points (Go)

| Location | Action |
|----------|--------|
| Start of iteration (before `agent.Execute`) | If `LAYERS_ENABLED` / config: run **`retrieve`** with query from **current plan feature**; prepend `contextBlock` instead of/in addition to `memory.BuildPromptContext` |
| After agent returns | Parse `[REMEMBER:…]` (existing) → **`record`** batch |
| Optional | **`append-run`** with iteration id, feature id, structured summary |
| End of run / size threshold | **`compact`** (or cron) |

Configuration keys (future Ralph PR): `layers_enabled`, `layers_command`, `layers_url`, `layers_data_dir`.

---

## 8. Repository layout (this repo)

```
layers/
  package.json
  tsconfig.json
  src/
    cli/
    server/          # optional HTTP
    index/           # SQLite + FTS + optional vectors
    retrieve/        # hybrid + MMR
    record/
    compact/
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

- [x] `layers v1 retrieve`: stdin JSON `RetrieveRequest` (§7.3) — FTS5 `bm25()` with Porter tokenizer; category + featureId **boost**; **MMR-lite** (word Jaccard) for diversity; **maxTokens** trims `contextBlock`.
- [x] **`contextBlock`** uses `[MEMORY CONTEXT]` … `[END MEMORY CONTEXT]` (§5.3).
- [x] Tests: unit (`ftsQuery`, `mmr`, `retrieve`) + CLI integration for `retrieve`.

### Phase 3 — Run log + compact

- [ ] `append-run` JSONL + `compact` last‑N → **`context-snapshot.md`**.
- [ ] Config: paths, max lines, max bytes.

### Phase 4 — Embeddings (optional provider)

- [ ] Pluggable embedder interface; local or API key from env.
- [ ] Hybrid merge + **embeddingFallbackOk** behavior.

### Phase 5 — HTTP server (optional)

- [ ] Mirror CLI payloads at `POST /v1/retrieve`, etc.
- [ ] Bind localhost only by default.

### Phase 6 — Ralph integration (Go)

- [ ] Config flags / env for Layers.
- [ ] Replace or gate `memory.BuildPromptContext` behind **Layers retrieve** when enabled.
- [ ] Forward `[REMEMBER:…]` extracts to **`record`**.
- [ ] E2E: one repo with `.layers` + iteration loop.

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
