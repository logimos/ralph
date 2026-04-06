# Ralph × Layers — orchestration specification (draft)

**Purpose:** Define how **Ralph (Go)** should **use** the Layers memory service beyond the minimal **Phase 6** wiring, so the loop can make **smarter, cheaper, more consistent** decisions. This complements **`LAYERS_SPEC.md`** (Layers TS contract) and **`RALPH_LOOP_ENHANCEMENTS.md`** (whole-loop analysis).

**Scope:** Ralph behavior, prompts, and file policy when Layers is enabled or available. **Out of scope:** changing Layers’ v1 JSON schema except where noted as future work.

---

## 1. Goals

1. **Use memory as a policy input**, not only static context — retrieve and record at **meaningful decision points**, not only once per iteration preamble.  
2. **Reduce token load** while preserving clarity: prefer **Layers `contextBlock` + bounded snapshot** over unbounded `@progress`.  
3. **Consistent feature semantics:** align “current feature,” retrieve query, and scope tracking (see §4).  
4. **Observable runs:** structured events in Layers **and** optional correlation with Ralph recovery/replan.

---

## 2. Current implementation snapshot (Phase 6)

| Mechanism | Behavior |
|-----------|----------|
| **Retrieve** | Before each iteration; query from first untested plan row; prepend `contextBlock`; 2m timeout; fallback to flat `BuildPromptContext` on failure |
| **Record** | Batch `[REMEMBER:…]` → Layers; on failure, warn + write flat `.ralph-memory.json` |
| **append-run** | One JSONL line per iteration (iteration + feature id) |
| **compact** | End of `runIterations`; writes `context-snapshot.md` |
| **progress.txt** | Unchanged; still referenced in prompt |

---

## 3. Smarter use of Layers (proposed behaviors)

These are **spec-level intentions** for future Ralph PRs, not all implemented today.

### 3.1 Inject compacted history, not raw progress

**Problem:** `@progress.txt` grows without bound (see `RALPH_LOOP_ENHANCEMENTS.md` §5.1, §7.10).

**Direction:**

- Prefer **`@`** reference to **`context-snapshot.md`** (or a Ralph-generated **`progress-context.md`**) produced from Layers **`compact`** or a **tail-only** file.  
- Keep **`progress.txt`** as append-only archive; stop feeding the full file to the agent when a bounded substitute exists.

### 3.2 Re-retrieve after material events

**Problem:** A single retrieve at iteration start ignores new information from failures or recovery text.

**Direction (optional flags):**

- **`layers-retrieve-on-retry`** — extra retrieve when recovery injects `additionalPromptGuidance`.  
- **`layers-retrieve-after-failure`** — after `containsFailureIndicators` or non-zero exit, retrieve with query augmented by **last stderr snippet** or **failure class** (bounded length).

### 3.3 Record structured outcomes, not only REMEMBER markers

**Direction:**

- After verification (future **verify gate**), **`record`** a **`fact`** or **`context`** entry with **test outcome** + **feature id** (even without agent prose).  
- Optionally **`append-run`** with `kind: failure` | `success` matching recovery state machine.

### 3.4 Query construction upgrades

**Direction:**

- Include **step titles** or **milestone name** in retrieve query when available.  
- Optional **`layers-retrieve-query-template`** in config (string with placeholders `{featureId}`, `{category}`, `{description}`).

### 3.5 Policy and safety

**Direction:**

- **Rate-limit** Layers calls per iteration (max N HTTP/subprocess calls).  
- **Redact** secrets from any string sent to `append-run` payload from agent output.  
- **Align** `extractCurrentFeatureFromPlans` with prompt **priority** semantics (`RALPH_LOOP_ENHANCEMENTS.md` §4.1) so retrieve query and scope **refer to the same feature**.

---

## 4. Consistency requirements (hard)

| ID | Requirement |
|----|-------------|
| R-1 | **Current feature** used for retrieve **must** match the feature id used for scope / deferral for that iteration (same plan read). |
| R-2 | **Timeouts** on all Layers calls from Ralph (already: bounded context). |
| R-3 | If Layers disabled or unreachable, behavior **degrades** to flat memory without aborting the loop unless configured strict (future flag). |

---

## 5. Configuration (future unified block)

Illustrative YAML shape (not all keys exist yet):

```yaml
layers:
  enabled: true
  url: "http://127.0.0.1:7847"   # optional; else CLI
  command: "layers"             # optional
  data_dir: ".layers"           # optional
  retrieve:
    max_tokens: 2000
    top_k: 10
  prompt:
    include_snapshot: true      # @ context-snapshot.md
    snapshot_path: ".layers/context-snapshot.md"
  behavior:
    retrieve_on_retry: false
    record_verify_outcomes: false
```

---

## 6. Non-goals (this spec)

- Defining Layers TS internals (see **`LAYERS_SPEC.md`** / **`layers_spec_v2.md`**).  
- Replacing **`plan.json`** as source of truth.  
- Mandating embeddings (FTS-only must remain valid).

---

## 7. Cross-references

| Doc | Use |
|-----|-----|
| `docs/LAYERS_SPEC.md` | Layers v1 API |
| `docs/layers_spec_v2.md` | Layers evolution |
| `docs/layers/ralph-integration.md` | Operator integration guide |
| `docs/RALPH_LOOP_ENHANCEMENTS.md` | Loop-wide issues and status |
| `internal/layers/` | Go client |

---

## 8. Revision

**v0.1** — Initial draft after Phase 6; captures “smarter Layers” roadmap for Ralph.
