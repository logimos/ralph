# Ralph — core enhancement phases (trackable)

**Purpose:** Map the **loop-wide** work from **`RALPH_LOOP_ENHANCEMENTS.md`** and **`RALPH_ROADMAP.md`** onto **numbered phases** with **acceptance** and **source docs**. This file is the **checklist**; the roadmap remains the **burndown table**.

**Principles:** Phases are **ordered by leverage** (correctness and clarity before parallelism). Each phase should ship with **tests** and **doc updates** where applicable.

---

## Phase map (Ralph orchestrator)

| Phase | Name | Roadmap IDs | Primary docs | Status |
|-------|------|-------------|--------------|--------|
| **1** | **Plan priority + prompt alignment** | P1, P2 (prompt) | §4.1, §7.1 in **`RALPH_LOOP_ENHANCEMENTS.md`** | **Done** — optional **`priority`**; **`plan.NextWorkFeature`**; **`BuildIterationPrompt(..., planUsesPriority)`** |
| **2** | **Verify gate** | V1 | §7.2 **`RALPH_LOOP_ENHANCEMENTS.md`** | Backlog |
| **3** | **Smarter Layers orchestration** | (extends P2) | **`RALPH_LAYERS_SPEC.md`** §3.2–3.4 | Backlog — re-retrieve on retry/failure, query tuning |
| **4** | **Structured iteration / run log** | §7.3, §7.12 | **`RALPH_LOOP_ENHANCEMENTS.md`**, Layers **`append-run`** | Backlog — tighter schema, optional verify outcomes |
| **5** | **Replan / recovery tuning** | §7.5, §7.8 | **`RALPH_LOOP_ENHANCEMENTS.md`** | Backlog |
| **6** | **Multi-agent** | M1 | **`RALPH_ROADMAP.md` §5** | Backlog — integrate or reduce surface |

**Already largely covered elsewhere (not sequential phases here):**

- **Bounded context (C1, C2)** — Layers snapshot + **`progress-context`**; see **`RALPH_ROADMAP.md` §3**.
- **Layers TS service** — **`docs/LAYERS_SPEC.md`**, **`layers_spec_v2.md`**.

---

## Phase 1 — acceptance

1. **`plan.json`** may include optional **`priority`** (integer, higher = sooner; **0** = unset).
2. **`extractCurrentFeatureFromPlans`** / **`plan.NextWorkFeature`** select the same untested, non-deferred row: **highest priority first**, then **file order**.
3. **`BuildIterationPrompt`**: if any row sets **`priority` ≠ 0**, step 1 describes **priority ordering**; otherwise step 1 describes **file order** (no “you choose highest priority” mismatch).

---

## Revision

**v0.1** — Initial phase map; Phase 1 implemented on branch **`cursor/ralph-enhancement-phases-p1`**.
