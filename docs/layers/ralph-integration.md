# Ralph ↔ Layers integration

This document explains how the **Ralph Go CLI** uses the **Layers** service. The normative contract snippets remain in **`docs/LAYERS_SPEC.md` §7.4**; this page is the **operator’s view**.

---

## 1. When to enable Layers

Enable Layers when you want:

- **Bounded, query-aware memory** instead of only a flat `.ralph-memory.json` relevance list.
- **Optional hybrid search** (FTS + embeddings) and a consistent **`contextBlock`** format.
- **Structured run history** (JSONL) and a **compact snapshot** for review or future prompt wiring.

Layers is **optional**. Without `-layers-enabled`, Ralph behaves as before (flat memory file + `BuildPromptContext`).

---

## 2. Configuration

### CLI flags

| Flag | Meaning |
|------|---------|
| `-layers-enabled` | Turn on Layers for retrieve / record / append-run / compact in the iteration loop |
| `-layers-command` | How to invoke Layers (default: `layers` on `PATH`). Supports quoted paths, e.g. `node "/path/with spaces/main.js"` |
| `-layers-url` | If set (e.g. `http://127.0.0.1:7847`), Ralph uses **HTTP** instead of a subprocess |
| `-layers-data-dir` | Passed to Layers as `dataDir` / `LAYERS_DATA_DIR` |

### Config file (YAML / JSON)

`layers_enabled`, `layers_command`, `layers_url`, `layers_data_dir` — merged with usual **file < CLI** precedence.

### Project root for Layers

Ralph sets **`projectRoot`** to the **directory containing `plan.json`** (from `-plan`). Layers stores data under **`<projectRoot>/.layers/`** by default.

---

## 3. What happens each iteration (high level)

1. **Compact (if enabled):** Ralph calls **`compact`** so **`context-snapshot.md`** is up to date before building the iteration prompt. If that file exists and is non-empty, the prompt includes **`@<absolute-path>/context-snapshot.md`** between **`plan.json`** and **`progress.txt`**, with a short instruction to prefer the snapshot for iteration context.

2. **Retrieve (if enabled):** Ralph builds a **retrieve** query from the **current untested** plan feature (category, description, feature id when known). The **`contextBlock`** from Layers is **prepended** to the iteration prompt.  
   - If retrieve **fails** or Layers is unavailable, Ralph **falls back** to `memory.BuildPromptContext` (flat JSON file).

3. **Agent runs** as usual (`cursor-agent`, etc.).

4. **`[REMEMBER:…]` markers** in stdout: if Layers is enabled, Ralph sends a **`record`** batch.  
   - If **`record` fails**, Ralph **warns** and writes memories to the **flat** `.ralph-memory.json` instead (durability over strict single-writer purity).

5. **append-run:** Ralph appends a **structured** JSONL event (iteration number, feature id, etc.). Failures are **debug-logged** in verbose mode.

---

## 4. End of `runIterations`

Ralph calls **`compact`** again so **`context-snapshot.md`** reflects the final run log state (it was also refreshed **before** each iteration prompt when Layers is enabled). Between iterations, **`append-run`** adds lines to the run log; the pre-prompt **`compact`** folds those into the snapshot for the **next** turn.

---

## 5. Operational checklist

1. **Build Layers:** `npm run build -w layers`
2. **Install on PATH** *or* set **`-layers-command`** *or* run **`layers v1 serve`** and set **`-layers-url`**
3. **Optional:** `LAYERS_OPENAI_API_KEY` for hybrid retrieval
4. Run: `ralph -layers-enabled -iterations N …`

---

## 6. Limitations (today)

- **Progress file:** Ralph still **`@`**-references **`progress.txt`** (append-only archive). With Layers enabled, a **bounded** **`context-snapshot.md`** is also **`@`**-referenced when it exists; the prompt tells the agent to prefer the snapshot for recent context.
- **Empty first run:** Until **`compact`** produces a non-empty **`context-snapshot.md`**, the prompt may omit the snapshot **`@`** (only plan + progress).
- **Smarter orchestration** (re-query Layers after failures, policy-driven retrieve) is **not** implemented; see **`docs/RALPH_LAYERS_SPEC.md`**.

---

## 7. Code references

- Go client: **`internal/layers/`**
- Loop wiring: **`ralph.go`** (`runIterations`)
