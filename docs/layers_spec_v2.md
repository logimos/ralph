# Layers specification — v2 themes (forward-looking)

**Status:** Draft / roadmap. **`LAYERS_SPEC.md`** remains the **normative v1** contract for shipped behavior. This document captures **gaps**, **inconsistencies**, and **evolution** after Phases 0–6.

**Audience:** implementers extending `layers/` or integrating new transports.

---

## 1. Purpose

v1 delivered: SQLite + FTS, optional OpenAI embeddings, hybrid retrieve, run log + compact, HTTP mirror, Ralph Go client. v2 addresses **what v1 deliberately deferred** or **implemented minimally**.

---

## 2. Gaps in v1 (canonical list)

### 2.1 Data model and indexing

| Gap | v1 behavior | v2 direction |
|-----|-------------|--------------|
| **Chunking** | Full row in FTS; `LAYERS_SPEC` mentions chunks as deferred | Split long `content` into chunk rows; FTS/vector per chunk; parent join on retrieve |
| **Memory types** | Fixed enum in TS validation | Extensible types or namespaced tags; migration policy |
| **Temporal decay** | Implicit via `updated_at` ordering in fallbacks | Explicit half-life or decay factor in scoring (oc-spec style) |
| **Branch / workspace scope** | Project-scoped only | Optional `gitBranch` or `workspaceId` metadata + filter at retrieve |

### 2.2 Embeddings and providers

| Gap | v1 behavior | v2 direction |
|-----|-------------|--------------|
| **Provider plugins** | OpenAI only (`fetch`) | **Ollama**, local file, or interface registry via env (`LAYERS_EMBEDDER=…`) |
| **Dimension mismatch** | Assumes single model per DB | Store `embedding_model` + `dim` per row; reject mixed without re-embed |
| **Backfill** | Lazy on retrieve | CLI: `v1 re-embed` or background job for all rows |

### 2.3 Operations and reliability

| Gap | v1 behavior | v2 direction |
|-----|-------------|--------------|
| **Transactions** | Per-statement | Batch `record` in one transaction; atomic import |
| **Backup / export** | Ad hoc SQLite file | `v1 export` / `v1 import` archive format |
| **Migrations** | Manual `user_version` | Tested migration ladder; optional `PRAGMA` integrity check on open |

### 2.4 HTTP / API

| Gap | v1 behavior | v2 direction |
|-----|-------------|--------------|
| **Auth** | Localhost-only by default | Optional shared secret header or mTLS for remote bind |
| **Idempotency** | None for POST | `Idempotency-Key` for record |
| **Streaming** | N/A | Optional NDJSON stream for large retrieve (unlikely priority) |
| **OpenAPI** | Spec in markdown only | Generated OpenAPI 3 from TS types or vice versa |

### 2.5 Run log (Layer A)

| Gap | v1 behavior | v2 direction |
|-----|-------------|--------------|
| **Schema validation** | Loose `kind` + `payload` | JSON Schema per `kind`; version field per line |
| **Compaction policy** | Single `compact` command | Scheduled compaction, retention by age, separate “archive” file |
| **Correlation** | `sessionKey` freeform | UUID run id from Ralph; link to CI job id |

### 2.6 Contract drift with Ralph

| Issue | Detail |
|-------|--------|
| **Single writer** | Spec §11 still warns “no double writer”; Ralph **falls back** to `.ralph-memory.json` if `record` fails — document as **supported degradation** or tighten policy in v2. |
| **Fact type** | TS allows `fact`; Go `memory` uses four types — align enums across import/export. |

---

## 3. v2 feature themes (prioritized themes, not commitments)

1. **Multi-provider embeddings** + explicit model metadata.  
2. **Chunk index** for long memories + retrieval merge.  
3. **Operational hardening**: export/import, idempotent HTTP, auth for remote serve.  
4. **Run log schema** + richer compaction strategies.  
5. **Optional git-branch / environment** scoping for monorepos and parallel workstreams.

---

## 4. Non-goals for v2 (likely)

- Replacing **`plan.json`** authority (remains Ralph).  
- Running inside the Cursor host process.  
- Distributed multi-node Layers (single workspace process remains the default mental model).

---

## 5. Relationship to other docs

| Document | Role |
|----------|------|
| **`LAYERS_SPEC.md`** | Shipped v1 — keep accurate for CLI/HTTP JSON |
| **`docs/layers/`** | User + integration guides |
| **`RALPH_LAYERS_SPEC.md`** | Ralph orchestration using Layers (smarter loops) |
| **`RALPH_LOOP_ENHANCEMENTS.md`** | Whole-loop analysis including Layers-era updates |

---

## 6. Revision

**v0.1** — Initial v2 themes document after Phase 6 completion.
