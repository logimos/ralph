# Ralph Loop: Architecture Analysis and Enhancement Opportunities

This document analyzes how the **Ralph iterative loop** works in this codebase, what it does well, where cost and fragility concentrate, how it interacts with **Cursor** (and other agent CLIs), and concrete directions to **optimize the loop** itself—rather than general product features.

**Layers era:** The repo now ships **`layers/`** (TypeScript memory service) and optional **Ralph ↔ Layers** integration (`-layers-enabled`). Much of §6.5–7.12 described **aspirational** mitigations before that stack existed; §6.6 and §10 summarize **what is implemented**, **what is weak**, and **what to do next**. See also **`docs/RALPH_LAYERS_SPEC.md`** (Ralph orchestration using Layers) and **`docs/layers/`** (user docs).

---

## 1. What the Ralph loop is (in this repo)

At its core, Ralph is a **Go CLI orchestrator** that runs a **fixed number of iterations** (`-iterations N`). Each iteration:

1. Loads configuration, optional memory (`-memory-file`), and nudges (`-nudge-file`).
2. Builds a **single-line prompt** that references the plan file and progress file by path (the `@path` convention used by Cursor Agent).
3. **Invokes an external AI agent as a subprocess** and waits for it to finish.
4. Interprets **stdout/stderr** for a completion marker, failure heuristics, and optional memory markers.
5. May append structured-ish lines to `progress.txt`, trigger **recovery**, **replan**, **deferral** (scope), and milestone celebrations—then repeats.

The **authoritative workflow state** for “what is left to do” lives primarily in **`plan.json`** (JSON array of features: `id`, `steps`, `tested`, `deferred`, etc.). The **append-only `progress.txt`** is a human-oriented log and handoff surface between iterations. The **agent** is expected to edit the repo: update the plan/PRD, append progress, run tests, and commit.

So the loop is: **orchestrator (deterministic) → agent (non-deterministic) → filesystem + git → repeat**.

---

## 2. Architecture map (loop-centric)

| Layer | Responsibility |
|--------|----------------|
| `ralph.go` → `runIterations` | The main loop: deadline/scope, prompt assembly, agent execution, completion signal, failure/replan/side effects. |
| `internal/prompt` | Builds the iteration prompt: `@<plan> @<progress>` plus step-by-step instructions (typecheck, tests, update PRD, append progress, commit, `COMPLETE` signal). |
| `internal/agent` | Spawns `cursor-agent --print --force <prompt>` or `claude --permission-mode acceptEdits -p <prompt>`. |
| `internal/plan` | Reads/writes `plan.json`; optional JSON extraction from agent output for plan generation. |
| `internal/recovery` | Classifies failures from output + exit code; retry/skip/rollback strategies. |
| `internal/replan` | Triggers replanning (e.g. consecutive failures, plan file hash change); can invoke agent for replan strategies. |
| `internal/scope` | Per-feature iteration budgets and deadlines; can mark plans deferred. |
| `internal/memory`, `internal/nudge` | Injected text prepended to the prompt; nudges acknowledged to avoid repetition. |
| `internal/layers`, `layers/` (optional) | When **`-layers-enabled`**: **retrieve** / **record** / **append-run** / **compact** via the Layers TS service; see **`docs/LAYERS_SPEC.md`**. |
| `internal/milestone` | Derived progress from plan items tagged with milestones. |

**Legacy parallel:** `ralph.sh` implements the same *idea* in bash (Claude CLI, `@plans/prd.json @progress.txt`, same completion token).

---

## 3. What works well

### 3.1 Thin orchestrator, heavy agent

Ralph does not try to implement coding inside Go. It **delegates** implementation to the agent. That keeps the orchestrator maintainable and matches how Cursor and similar tools are meant to be used.

### 3.2 File-based state is inspectable

`plan.json` and `progress.txt` are easy to **diff in git**, read when debugging, and edit manually (with replan triggers noticing plan changes). This is valuable for long-running “ralph loops” where transparency beats opaque internal state.

### 3.3 Control surfaces around the loop

The loop is surrounded by useful knobs: **recovery**, **auto-replan**, **scope limits**, **deadlines**, **nudges** (mid-run steering), **memory** (cross-run hints), and **milestones** (coarse checkpoints). These are the right *conceptual* layers for optimizing autonomy vs. safety.

### 3.4 Agent abstraction

`internal/agent` centralizes how the subprocess is invoked and special-cases **Cursor** vs **Claude**, so the rest of the code does not scatter agent-specific flags.

---

## 4. Where the design is weak or inconsistent (loop impact)

### 4.1 “Priority” in the prompt vs. priority in code

The iteration prompt tells the model to choose the **highest-priority** feature, not necessarily the first in the list (`internal/prompt/prompt.go`). Meanwhile, helpers such as `extractCurrentFeatureFromPlans` use the **first untested, non-deferred** row in file order (`ralph.go`). That **semantic mismatch** affects:

- Scope tracking (which feature ID is “current” for deferral and failure attribution).
- Mental model: users think the orchestrator enforces priority; it mostly does not.

**Loop optimization implication:** Either make priority **explicit** in data (e.g. `priority` field, stable sort) and use it everywhere, or simplify the prompt to “work on the next unfinished feature according to plan order.” Mixed messages waste iterations and model reasoning.

### 4.2 No closed-loop verification of “done”

Success for “did the feature land?” is largely **trust in the agent**: the prompt asks for tests and typecheck, but Ralph’s iteration success path does not, by itself, **re-run** `typecheck`/`test` after the agent returns and gate `plan.json` updates. Failure detection uses **heuristics** on agent output (`containsFailureIndicators`) plus exit codes—not structured test reports.

**Consequence:** Expensive agent work can be followed by **false confidence** (green-looking narrative) or **false failure** (substring “error:” in unrelated text).

### 4.3 `progress.txt` is a handoff without a schema

Appending freeform notes is excellent for humans and weak for machines. Ralph cannot reliably **compress**, **summarize**, or **diff** “what changed this iteration” without re-invoking an LLM or heuristics.

### 4.4 Multi-agent mode is not in the hot path

Configuration includes `-multi-agent` / `enable_multi_agent`, but **the main iteration loop does not branch on it** (grep shows no use of `EnableMultiAgent` inside `runIterations`). The **multiagent** package exists and is documented, yet the default loop remains **single subprocess per iteration**. Users may believe parallel agents are optimizing the loop when they are not.

### 4.5 Replan and “agent” strategies are second AI passes

Replanning—especially strategies that call the agent—add **another full LLM invocation** on top of an already expensive iteration. That is appropriate when stuck, but it is a **multiplier** on cost if triggers are too sensitive or thresholds too low.

---

## 5. Where it gets expensive

### 5.1 One full agent run per iteration (dominant cost)

Each iteration is typically **one Cursor/Claude invocation** with a prompt that includes **paths to the whole plan and cumulative progress**. As the repo and `progress.txt` grow, **context size** grows roughly linearly unless you summarize or truncate. This is usually the **largest dollar and latency cost** in the system.

### 5.2 `@` file inclusion semantics (Cursor Agent)

The prompt embeds **absolute paths** to `plan.json` and `progress.txt`. The agent implementation uses Cursor’s CLI pattern (`--print --force`). Whatever the Cursor Agent loads from those paths counts against **context**. Long progress logs + large plans + large codebase reads inside the agent compound quickly.

### 5.3 Ancillary AI calls

- **`-generate-plan`** from notes: full agent run.
- **Goal decomposition** (`decomposeGoal`): full agent run.
- **Replan** with agent-backed strategies: additional runs.

These are fewer than per-iteration work in a long loop but can dominate short runs.

### 5.4 Subprocess overhead vs. API streaming

Spawning CLI agents adds process startup and **no native backpressure** from Ralph’s side beyond waiting on stdout/stderr. Not usually the main cost versus tokens, but it matters for very fast micro-iterations.

---

## 6. How Ralph interacts with Cursor AI (specifics)

### 6.1 Integration model

Ralph does **not** embed the Cursor IDE or use a private RPC. It runs **`cursor-agent`** as a **child process** with:

- `--print` (non-interactive output),
- `--force` (non-interactive behavior),
- The **entire prompt as one positional argument** (`internal/agent/agent.go`).

So Ralph’s relationship to Cursor is: **CLI contract + whatever file reads the agent performs when resolving `@path` references**.

### 6.2 Detection

`IsCursorAgent` treats commands containing `cursor-agent` or `cursor` (with a special case to avoid confusing `claude`-like names) as Cursor (`internal/agent/agent.go`). This keeps one code path for “Cursor-style” vs “Claude-style” flags.

### 6.3 Session and continuity

From Ralph’s perspective, each iteration is **stateless**: new process, new prompt. Any **conversation memory** depends on Cursor Agent’s own behavior (caching, project index, etc.), not on Ralph storing chat history. The **durable** continuity Ralph provides is **files**: `plan.json`, `progress.txt`, memory JSON, git history.

### 6.4 Implications for optimization

- **Prompt design** and **file size** directly drive Cursor cost.
- **Forcing structured outputs** (e.g. JSON summary blocks) improves what Ralph can do **without** another huge call.
- **Tighter repository scoping** (what the agent is allowed to read) is outside Ralph today but affects Cursor behavior.

---

## 6.5 Two-layer memory: OpenClaw (oc-spec) vs Ralph

**Implementation direction:** Memory for Ralph is specified as the TypeScript app **`layers/`** with a versioned **Ralph ↔ Layers contract**. See **[`LAYERS_SPEC.md`](LAYERS_SPEC.md)** (phased plan, CLI/HTTP API, cross-references to oc-spec and current Ralph memory). **User-facing docs:** [`docs/layers/README.md`](layers/README.md). **Gaps / v2 themes:** [`layers_spec_v2.md`](layers_spec_v2.md). **Ralph orchestration roadmap (smarter use of Layers):** [`RALPH_LAYERS_SPEC.md`](RALPH_LAYERS_SPEC.md).

The `oc-spec/` folder documents **OpenClaw’s** persistent memory architecture. It is useful as a **reference pattern** for token-efficient continuity—not as something to copy line-for-line in Go, but as a **separation of concerns** Ralph can emulate.

### 6.5.1 What OpenClaw specifies (summary from oc-spec)

| Layer | Role | Storage / behavior (per `spec.md`) |
|--------|------|-------------------------------------|
| **Layer A — Session history** | **Authoritative** append-only record of what happened in conversation | `sessions.json` metadata + `<sessionId>.jsonl` transcripts; appends go through **SessionManager** semantics (ordering/parent chain) for **history + compaction** correctness |
| **Layer B — Semantic memory** | **Retrieved** facts, not full history | `memory-core`: markdown files, SQLite index with **vector + FTS**, **hybrid merge** (`spec.algorithms.md`: vector + keyword, temporal decay, **MMR** for diversity), **FTS-only fallback** when embeddings fail |

**Retrieval** (`memory_search`): normalize query → optional embeddings → FTS → merge with weights → threshold + **top‑k** — so the model sees **a small relevant slice**, not the entire transcript.

**Maintenance** on Layer A: locks, atomic writes, prune/archive, **disk budget** — history is **durable** but **bounded**.

### 6.5.2 What Ralph does today

| Artifact | Acts like OpenClaw… | Token / clarity issue |
|----------|---------------------|------------------------|
| **`progress.txt`** | Layer A–ish (append-only log) | **Unbounded** growth; included via `@progress` → **full file** tends to land in context → **killer for tokens** as runs lengthen. |
| **`.ralph-memory.json`** | Partial Layer B (structured entries) | **Flat list**; `BuildPromptContext` injects top‑**N** by simple **relevance score** (`internal/memory/memory.go`) — better than raw progress, but **no hybrid search**, no chunking, no MMR, no embedding fallback path. |
| **`plan.json`** | Workflow truth | Usually must stay in context in some form; can be **projected** (see §7.4) to shrink. |

Ralph iterations are **stateless subprocesses** (§6.3), so there is **no** transcript compaction loop inside Ralph like OpenClaw’s **preflight compaction** — unless we **build** a deliberate “summarize / rotate progress” step.

### 6.5.3 Mapping OpenClaw ideas onto Ralph (without losing instruction clarity)

**Principle:** Keep **one authoritative, append-only trail** for audit/debug (like OpenClaw’s JSONL), but **stop feeding that entire trail** into the agent every time. Feed **(1) short working memory + (2) retrieved facts**, like OpenClaw’s A/B split.

| OpenClaw idea | Ralph-oriented analogue |
|----------------|-------------------------|
| Layer A authoritative + compaction | Keep **`progress-full.txt` or JSONL** as archive; maintain **`progress-context.txt`** (or generated block) that is **only** the last *k* entries or a **rolling summary** + pointer to git SHAs. Prompt references **`@`** the **small** file. |
| Layer B semantic + hybrid search | Evolve **`.ralph-memory.json`** (or split **`MEMORY.md` + index**): retrieval by **feature id / category / keywords**; optional **SQLite FTS** or **bleve**/embedded search in Go; optional embeddings later. Always **cap** injected lines (`maxEntries` already exists — expose and tune per env). |
| Transcript invariants | Ralph could require **structured append** (timestamp, feature_id, summary line) so **deterministic** compaction can run without an LLM, or **one** periodic summarization call replaces 50 raw paragraphs. |
| MMR / diversity | When selecting memories, avoid **10 near-duplicate** “use TypeScript” lines — prefer diverse entry types (decision vs convention). |
| Disk budget | **Rotate** or **archive** `progress.txt` when size > N KB; keep **tail** in hot path. |

**Clarity of instruction:** The **prompt rules** (single feature, tests, commit, `COMPLETE`) stay in **`BuildIterationPrompt`** unchanged. What shrinks is **evidence / history**, not the **task contract**. Optionally add one line: “Authoritative detail is in `progress-archive.*`; work from the summary below.”

### 6.5.4 Cheap wins vs heavier engineering

**Cheap (mostly policy + files):**

- Stop `@`-including **unbounded** `progress.txt`; generate **`progress-last.md`** (last *k* lines or last 2k tokens equivalent).
- Pin **`memory_retention`** and **lower default** injected memories; tie **`BuildPromptContext(category, …)`** to **current feature’s category** once `extractCurrentFeatureFromPlans` is aligned with prompt priority (§4.1).
- Ask the agent to append **one structured line** per iteration (`FEATURE=3 STATUS=done SUMMARY=…`) so Ralph can **parse** and **summarize** without loading prose.

**Heavier (OpenClaw parity in spirit):**

- **Embedded FTS** over memory entries + progress summaries for **query = current feature description**.
- **Compaction job**: after each iteration or every *n* steps, rewrite “working summary” via **rules** or **single** LLM call **only when** size exceeds threshold (OpenClaw’s preflight compaction analogue).

### 6.6 Post-Layers: what landed vs what is still weak

Phases **0–6** in **`LAYERS_SPEC.md`** are **implemented** for the Layers service and basic Ralph wiring. The table below maps **earlier pain points** to **current state** and **remaining work**.

| Topic | Original issue (this doc) | Implemented? | Still weak / poorly implemented |
|--------|---------------------------|--------------|----------------------------------|
| Layer B retrieval | Flat memory, simple score | **Yes** — Layers FTS + MMR + optional embeddings; Ralph prepends **`contextBlock`** when **`-layers-enabled`** | Retrieve runs **once per iteration** at prompt build; no **re-query after failure**; query tied to first untested row — **priority mismatch** with prompt (§4.1) still matters |
| Layer A history | Unbounded **`progress.txt`** | **Partial** — **`run.jsonl`** + **`compact` → `context-snapshot.md`** exist; Ralph **append-run** each iteration | **`progress.txt` still `@`-referenced in full**; snapshot **not auto-injected** into prompt — token killer **not fully removed** |
| Hybrid / embeddings | N/A in old Ralph | **Yes** in Layers (optional OpenAI) | Extra **cost** if enabled; **Ollama** / other providers **not** in v1 Layers |
| Chunking long memories | Mentioned as future | **Not** in v1 Layers | Long entries still single FTS row — see **`layers_spec_v2.md`** |
| Verification gate | No deterministic test/typecheck gate | **Not** in Ralph loop | §7.2 still **open** — high leverage for correctness |
| Multi-agent | Flag not in hot path | **Partial** — **`-multi-agent` is rejected** when running iterations (clear error); **`-list-agents`** still works; see **`RALPH_ROADMAP.md`** | Orchestrated multi-agent loop **not** implemented |
| Priority alignment | Prompt vs `extractCurrentFeatureFromPlans` | **Partially improved** for Layers retrieve (category + description from same plan row) | **Full** priority field + sort still **not** done (§7.1) |
| Smarter orchestration | N/A | **Minimal** — retrieve/record/append/compact | **Policy-driven** use (re-retrieve on retry, record verify outcomes) — see **`RALPH_LAYERS_SPEC.md`** |

**How to fix (directional):** Prefer **`RALPH_LAYERS_SPEC.md`** for Ralph-side behavior; **`layers_spec_v2.md`** for the TS service evolution; keep **§7** below for non-Layers loop work (verify gate, multi-agent, etc.).

---

## 7. Enhancement directions (focused on the loop)

Below are **actionable** directions ordered roughly by impact on loop quality and cost. They are architectural/product choices, not calendar estimates.

### 7.1 Align priority semantics end-to-end

- Add an optional **`priority`** (integer) or **`order`** field to plan items and **sort** consistently in the orchestrator.
- Update `extractCurrentFeatureFromPlans`, replan “current feature” selection, and the **prompt** so they agree.
- Optionally support **explicit “blocked” / dependencies** so the model does not “creatively” skip prerequisites.

**Why:** Reduces wasted iterations from conflicting instructions.

### 7.2 Add an optional “verify gate” after the agent returns

After `agent.Execute`, optionally run **`cfg.TypeCheckCmd` and `cfg.TestCmd`** in Ralph (with timeouts) **before** treating the iteration as clean. If verification fails, **do not** reset consecutive-failure counters; feed stderr into recovery/replan.

**Why:** Closes the loop with **deterministic** signals; reduces reliance on substring heuristics in model prose.

**Cost note:** Adds local CPU time but can **reduce** LLM spend by failing fast without another agent turn.

### 7.3 Structured iteration record instead of only append-only prose

Introduce a machine-readable artifact, e.g. `ralph-state.json` or JSONL `iterations.jsonl`, with fields like: `iteration`, `feature_id`, `git_head_before/after`, `commands_run`, `test_exit_code`, `summary`, `tokens_estimate` (if available). Keep `progress.txt` as optional human narrative or generate it from the structured log.

**Why:** Enables **compaction**: pass a **short structured last-N** into the prompt instead of an ever-growing `progress.txt`.

### 7.4 Progressive context budgeting

- **Rolling summary:** Periodically replace the tail of `progress.txt` with a summarized block (agent- or rule-based) once size exceeds a threshold.
- **Plan projection:** Instead of `@` the full `plan.json`, generate a **minimal JSON** containing only the next K features or the active milestone slice.
- **Diff-based prompts:** Pass `git diff` summaries or file lists instead of narrative repetition.

**Why:** Directly attacks the **dominant token cost** in long Ralph loops.

### 7.5 Harden failure detection

Replace or augment `containsFailureIndicators` with:

- Parsed **JUnit / go test JSON** output when tests are run by Ralph.
- Exit-code-only paths when Ralph runs verify commands.
- Configurable **allowlist** for benign lines containing “error” (logs, UI copy).

**Why:** Fewer mistaken recovery branches and replans—both save money and reduce thrash.

### 7.6 Wire or remove multi-agent configuration

Either:

- **Integrate** `EnableMultiAgent` into `runIterations` (e.g. staged pipeline: implementer → tester → summarizer with separate prompts), or
- **Remove/disable** the flag from user-facing docs until implemented.

**Why:** Avoids false expectations and allows real parallelization **only** where file locking and git conflicts are handled.

**Status:** **`validateConfig`** now **fails fast** if **`-multi-agent`** is passed together with **`-iterations`** (main loop is still single-agent). **`flag.Usage`** text no longer implies parallel multi-agent runs. Full integration remains **backlog** — see **`docs/RALPH_ROADMAP.md`** §5.

### 7.7 Idempotency and commit strategy

The prompt asks for a **git commit per feature**. The loop does not verify commit boundaries. Enhancements:

- Check `git status` / last commit message pattern before/after iteration.
- Optionally **squash** or **tag** Ralph iterations for traceability.

**Why:** Makes rollbacks and automation safer when replan/recovery fires.

### 7.8 Smarter replan triggers

- Increase default **threshold** or require **two different failure types** before agent replan.
- Cache **plan hashes** to avoid replanning on whitespace-only edits.
- Prefer **incremental** replan when possible (cheaper than full agent replan).

**Why:** Replan is an extra LLM-sized operation; tightening triggers cuts spend.

### 7.9 Cursor-specific optimizations (without forking the agent)

- Document recommended **`.cursorignore`** / repo layout so `@` pulls stay small.
- Prefer **absolute paths** only where needed; ensure redundant huge files are not referenced in prompts.
- Consider a **`--prompt-profile minimal`** that omits nonessential instructions for small iterations.

### 7.10 Split “history” from “working context” (OpenClaw Layer A pattern)

- **Archive** full append-only history to e.g. `progress-archive.jsonl` or rotate `progress.txt` by size with numbered parts.
- Point **`@`** in the iteration prompt at a **small** file: `progress-context.txt` / `progress-last.md` containing only **last N entries**, a **rolling summary** block, and/or **pointers** (commit SHAs, PR links).
- Optionally run a **compaction** step when `progress-context` exceeds a byte or line threshold (rule-based truncation first; **optional** one LLM summarization call as last resort).

**Why:** Preserves **auditability** (full history on disk) while **capping** what Cursor loads every iteration — same separation as OpenClaw’s **transcript + compaction** story.

**Status (Layers era):** Layers **`compact`** + **`context-snapshot.md`** implement **bounded derived context** from **`run.jsonl`**. **Gap:** Ralph still does not switch **`@`** from **`progress.txt`** to that snapshot by default — **poorly integrated** for token savings until **`RALPH_LAYERS_SPEC.md`** items land.

### 7.11 Strengthen Layer B memory retrieval (OpenClaw Layer B pattern)

- Pass **category from the current plan item** into `BuildPromptContext` instead of always using `""` in `runIterations` (so memories match **infra** vs **ui** work).
- Add **deduplication** or **MMR-style** selection: penalize near-duplicate entry text so 10 memories do not repeat one convention.
- Optional **keyword / FTS** index over memory entries (embedded SQLite or similar) with **query = feature description + steps** and **top‑k** with a **min score** threshold — mirrors OpenClaw’s **hybrid retrieval + fallback** without requiring embeddings on day one.

**Why:** Keeps **instruction text** in the prompt full-size while **facts** stay **small and relevant**.

**Status (Layers era):** **Implemented in Layers** (FTS + boosts + MMR + optional embeddings). Ralph passes **category + feature id** into **retrieve** when **`-layers-enabled`**. **Gap:** Flat **`BuildPromptContext("", 10)`** still runs every iteration as **fallback baseline**; **no min-score threshold** exposed in Ralph flags; **chunking** still future (**`layers_spec_v2.md`**).

### 7.12 Structured progress lines for machine-safe compaction

- Define a **one-line schema** (or JSON line in JSONL) per iteration: feature id, status, commit hash, short summary.
- Teach the prompt: “Append **both** a human paragraph **and** one machine line matching …”

**Why:** Lets Ralph **truncate** or **rebuild** `progress-context` **without** guessing from prose — reducing reliance on extra LLM calls for compaction.

**Status (Layers era):** **`append-run`** emits **structured JSONL** per iteration — **partial** fulfillment. **Gap:** No strict schema validation; **not** wired to auto-replace **`progress.txt`** in prompts; **human prose** in `progress.txt` remains **unstructured** for machine compaction.

---

## 8. Summary

The Ralph loop is a **simple, robust pattern**: repeated **agent subprocess** calls with **file-backed state**. Its strengths are **transparency** and **composability** (memory, nudges, scope, replan). Its main weaknesses for optimization are **context growth** (especially **`progress.txt` via `@`**, which behaves like an **unbounded OpenClaw Layer A** fed whole into every turn), **soft verification** of completion, **priority semantics drift**, and **unstructured progress**.

**With Layers:** The **semantic memory** and **run-log + compact** pieces of OpenClaw-style architecture are **addressed in `layers/`** and **partially wired** in Ralph. The **largest remaining gap** is **prompt assembly**: still **`@`** full **`progress.txt`** by default, and **bounded snapshot** (`context-snapshot.md`) is **not** the primary handoff file. **Smarter use of Layers** (re-retrieve on failure, inject snapshot, align priority) is specified in **`RALPH_LAYERS_SPEC.md`**.

The interaction with **Cursor** is entirely through the **CLI and `@` file references**, so **token-efficient prompts and artifacts** remain the highest-leverage improvements to the loop itself.

---

## 9. Suggested implementation order (technical only)

**Updated for Layers era** — items already largely covered by **Layers + Phase 6** are noted.

1. **Verify gate** (typecheck/test in Ralph) + structured capture of results — **still open** (§7.2).  
2. **Split progress for context** (§7.10): `@` **bounded** context + archive full history — **Layers provides `compact` / snapshot**; **Ralph must switch `@` targets** — **highest remaining token win**.  
3. **Memory retrieval** (§7.11) — **largely in Layers**; Ralph: expose/tune retrieve options; optional **remove redundant** flat context when Layers succeeds.  
4. **Priority field** + consistent feature selection + prompt alignment — **still open** (§4.1, §7.1).  
5. **Structured iteration lines** (§7.12) — **partial** via **`append-run`**; tighten schema + prompt.  
6. **Replan trigger** tuning and incremental replan path.  
7. **Multi-agent** integration or config cleanup.

This ordering front-loads **deterministic correctness**, **context caps**, and **retrieval quality** before **parallelism** or heavier **LLM compaction** passes.

---

## 10. Layers-era roadmap (cross-doc index)

| Track | Document | Focus |
|-------|----------|--------|
| **Layers TS service** | `docs/LAYERS_SPEC.md` (v1), `docs/layers_spec_v2.md` (gaps) | API, storage, embeddings, HTTP |
| **Ralph + Layers behavior** | `docs/RALPH_LAYERS_SPEC.md` | Smarter retrieve/record, snapshot in prompt, policy |
| **Operator docs** | `docs/layers/README.md`, `user-guide.md`, `ralph-integration.md` | How to run and configure |
| **Loop-wide** | This document (§6.6, §7.x status) | Orchestrator, Cursor, cost, verification |

**“Smarter choices” via Layers:** means using **retrieve** not only as a static preamble but as a **decision-time** tool (e.g. after failures, after replan, with different queries) and feeding **compact** output into **`@`** — specified in **`RALPH_LAYERS_SPEC.md` §3**.
