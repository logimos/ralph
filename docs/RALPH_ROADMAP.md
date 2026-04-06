# Ralph — product roadmap and burndown

**Purpose:** Single place for **what Ralph claims vs what it does**, **what to build next**, and **acceptance-style notes**. Complements **`RALPH_LOOP_ENHANCEMENTS.md`** (analysis) and **`RALPH_LAYERS_SPEC.md`** (Layers orchestration).

**Principles:** Easy to use; **honest** CLI and docs; **remove or fail fast** on flags that do nothing; delegate heavy memory to **Layers** where appropriate.

---

## Legend

| Status | Meaning |
|--------|---------|
| Done | Shipped on `ralph2` |
| In progress | Active branch / PR |
| Next | Recommended order |
| Backlog | Future / lower priority |
| Rejected / defer | Explicitly not doing now |

---

## Burndown (ordered)

### 1. Truth in advertising (CLI + help)

| ID | Item | Status | Acceptance |
|----|------|--------|------------|
| T1 | **`-multi-agent`** does not change `runIterations` | **Done** (this branch) | `validateConfig` returns a clear error if `-multi-agent` is set with `-iterations` > 0; help text says **not wired**; `-list-agents` still works |
| T2 | **Help / examples** do not imply multi-agent runs parallel agents in the loop | **Done** | `flag.Usage` and examples updated |
| T3 | **Layers** vs **`progress.txt`**: document prompt behavior when bounded context is available | **Done** | **`docs/layers/ralph-integration.md`** |

### 2. Align “current feature” and prompt (priority mismatch)

| ID | Item | Status | Notes |
|----|------|--------|--------|
| P1 | Single source of truth for “current feature” (plan order vs priority field) | Backlog | See **`RALPH_LOOP_ENHANCEMENTS.md` §4.1**, §7.1 |
| P2 | Layers retrieve query uses same feature as scope/deferral | Partial | Same read as iteration tracking; prompt text may still say “highest priority” |

### 3. Bounded context (token cost)

| ID | Item | Status | Notes |
|----|------|--------|--------|
| C1 | Prefer **`@`** bounded file (`context-snapshot.md` or generated tail) when Layers enabled | **Done** | **`compact`** before each prompt when **`-layers-enabled`**; **`@`** order is plan → snapshot (if non-empty) → **`progress.txt`**; see **`RALPH_LAYERS_SPEC.md` §3.1** |
| C2 | Optional **`progress-context.txt`** maintained by Ralph | Backlog | Rule-based tail or compact step |

### 4. Verification

| ID | Item | Status | Notes |
|----|------|--------|--------|
| V1 | Optional **verify gate** (typecheck/test after agent) | Backlog | **`RALPH_LOOP_ENHANCEMENTS.md` §7.2** |

### 5. Multi-agent (future)

| ID | Item | Status | Notes |
|----|------|--------|--------|
| M1 | Integrate `internal/multiagent` into **`runIterations`** **or** remove surface area | Backlog | Until then, **`-multi-agent` is rejected** with iterations |

---

## What already works (reference)

- **Plan / progress / agent loop** — `runIterations`, recovery, replan, scope, milestones.  
- **Flat memory** — `.ralph-memory.json`, `[REMEMBER:…]`.  
- **Layers (optional)** — `-layers-enabled`, retrieve / record / append-run / compact; see **`docs/layers/ralph-integration.md`**.

---

## Revision

**v0.1** — Initial roadmap; **T1–T3** implemented on branch `cursor/ralph-roadmap-honesty`.  
**v0.2** — **C1**: **`context-snapshot.md`** included in iteration **`@`** when Layers is enabled and the snapshot file exists (non-empty).
