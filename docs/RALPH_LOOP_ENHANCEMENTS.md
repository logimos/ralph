# Ralph Loop: Architecture Analysis and Enhancement Opportunities

This document analyzes how the **Ralph iterative loop** works in this codebase, what it does well, where cost and fragility concentrate, how it interacts with **Cursor** (and other agent CLIs), and concrete directions to **optimize the loop** itself—rather than general product features.

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

---

## 8. Summary

The Ralph loop is a **simple, robust pattern**: repeated **agent subprocess** calls with **file-backed state**. Its strengths are **transparency** and **composability** (memory, nudges, scope, replan). Its main weaknesses for optimization are **context growth**, **soft verification** of completion, **priority semantics drift**, and **unstructured progress**. The interaction with **Cursor** is entirely through the **CLI and `@` file references**, so **token-efficient prompts and artifacts** are the highest-leverage improvements to the loop itself.

---

## 9. Suggested implementation order (technical only)

1. **Verify gate** (typecheck/test in Ralph) + structured capture of results.  
2. **Priority field** + consistent feature selection + prompt alignment.  
3. **Structured iteration log** + rolling summary / plan projection for prompts.  
4. **Replan trigger** tuning and incremental replan path.  
5. **Multi-agent** integration or config cleanup.

This ordering front-loads **deterministic correctness** and **cost control** before adding **parallelism** or more **LLM-heavy** behaviors.
