# Ralph Features Reference

This document provides detailed documentation for all Ralph features.

## Table of Contents

1. [Plan Management](#plan-management)
2. [Build System Support](#build-system-support)
3. [Configuration Files](#configuration-files)
4. [Environment Detection](#environment-detection)
5. [Two-Tier Failure Handling](#two-tier-failure-handling) ← **Understanding Recovery vs Replanning**
6. [Failure Recovery (Tier 1)](#failure-recovery-tier-1)
7. [Smart Scope Control](#smart-scope-control)
8. [Adaptive Replanning (Tier 2)](#adaptive-replanning-tier-2)
9. [Long-Running Memory](#long-running-memory)
10. [User Nudge Hooks](#user-nudge-hooks)
11. [Milestone Tracking](#milestone-tracking)
12. [Goal-Oriented Planning](#goal-oriented-planning)
13. [Outcome Validation](#outcome-validation)
14. [Multi-Agent Collaboration](#multi-agent-collaboration)
15. [Enhanced CLI Output](#enhanced-cli-output)

---

## Plan Management

Ralph uses JSON plan files to define features to implement.

### Plan Structure

```json
[
  {
    "id": 1,
    "category": "feature",
    "description": "Add user authentication",
    "steps": [
      "Create auth middleware",
      "Add login endpoint",
      "Add logout endpoint"
    ],
    "expected_output": "Users can log in and out",
    "tested": false,
    "milestone": "Alpha",
    "validations": []
  }
]
```

### Commands

```bash
# Generate plan from notes
ralph -generate-plan -notes notes.md

# View plan status
ralph -status

# List tested features
ralph -list-tested

# List untested features
ralph -list-untested
```

---

## Build System Support

Ralph auto-detects and supports multiple build systems.

### Supported Systems

| System | Detection | Type Check | Test |
|--------|-----------|------------|------|
| Go | `go.mod` | `go build ./...` | `go test ./...` |
| pnpm | `pnpm-lock.yaml` | `pnpm typecheck` | `pnpm test` |
| npm | `package.json` | `npm run typecheck` | `npm test` |
| Yarn | `yarn.lock` | `yarn typecheck` | `yarn test` |
| Gradle | `build.gradle` | `./gradlew check` | `./gradlew test` |
| Maven | `pom.xml` | `mvn compile` | `mvn test` |
| Cargo | `Cargo.toml` | `cargo check` | `cargo test` |
| Python | `setup.py`, `pyproject.toml` | `mypy .` | `pytest` |

### Usage

```bash
# Auto-detect (default)
ralph -iterations 5

# Explicit selection
ralph -iterations 5 -build-system gradle

# Override commands
ralph -iterations 5 -typecheck "make lint" -test "make test"
```

---

## Configuration Files

Persistent configuration via YAML or JSON files.

### Discovery Order

1. `.ralph.yaml` (current directory)
2. `.ralph.yml`
3. `.ralph.json`
4. `ralph.config.yaml`
5. `ralph.config.yml`
6. `ralph.config.json`
7. Same files in home directory (fallback)

### Example

```yaml
# .ralph.yaml
agent: cursor-agent
build_system: go
plan: plan.json
iterations: 5
verbose: true
max_retries: 3
```

### Precedence

```
Defaults < Config File < CLI Flags
```

---

## Environment Detection

Automatic adaptation to different execution environments.

### Detected Environments

- Local development
- GitHub Actions
- GitLab CI
- Jenkins
- CircleCI
- Travis CI
- Azure DevOps
- Generic CI

### Automatic Adaptations

| Environment | Timeout | Verbose | Colors |
|-------------|---------|---------|--------|
| Local | 30s | As configured | Enabled |
| CI | 120s | Enabled | Disabled |

### Resource Detection

- CPU cores (for parallelism hints)
- Available memory
- Project complexity (file count)

### Override

```bash
ralph -iterations 5 -environment github-actions
```

---

## Two-Tier Failure Handling

Ralph uses a sophisticated two-tier system to handle failures gracefully, minimizing stuck states and maximizing progress.

### Escalation Diagram

```
                    ┌───────────────────────┐
                    │  Feature Fails        │
                    └───────────┬───────────┘
                                │
    ════════════════════════════╪════════════════════════════
    TIER 1: RECOVERY            │ (Per-Feature)
    Flags: -max-retries, -recovery-strategy
    ════════════════════════════╪════════════════════════════
                                ▼
                    ┌───────────────────────┐
                    │  Detect Failure Type  │
                    │  (test/typecheck/etc) │
                    └───────────┬───────────┘
                                │
                    ┌───────────▼───────────┐
                    │  Apply Recovery       │
                    │  Strategy             │
                    └───────────┬───────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
    ┌───────┐             ┌───────────┐           ┌───────┐
    │ RETRY │             │ ROLLBACK  │           │ SKIP  │
    │(prompt│             │(git reset)│           │(next) │
    └───┬───┘             └─────┬─────┘           └───┬───┘
        │                       │                     │
        └───────────────────────┼─────────────────────┘
                                │
                    retries < -max-retries?
                        │           │
                       YES          NO
                        │           │
                    Try again   Skip feature, increment failure counter
                                    │
    ════════════════════════════════╪════════════════════════════
    TIER 2: REPLANNING              │ (Plan-Level)
    Flags: -auto-replan, -replan-threshold, -replan-strategy
    ════════════════════════════════╪════════════════════════════
                                    │
                    consecutive failures >= -replan-threshold?
                        │           │
                       NO          YES
                        │           │
                    Continue    ┌───▼───────────────┐
                    execution   │  Trigger Replan   │
                                │  Strategy         │
                                └─────────┬─────────┘
                                          │
            ┌─────────────────────────────┼─────────────────────────────┐
            ▼                             ▼                             ▼
    ┌───────────────┐           ┌─────────────────┐           ┌─────────────┐
    │ INCREMENTAL   │           │     AGENT       │           │    NONE     │
    │ (adjust plan) │           │ (AI restructure)│           │ (disabled)  │
    └───────────────┘           └─────────────────┘           └─────────────┘
            │                             │
            └─────────────────────────────┘
                          │
                    Create backup (plan.json.bak.N)
                    Apply restructured plan
                    Reset failure counters
                    Resume with new plan
```

### When to Use Recovery vs Replanning

| Situation | Use Recovery | Use Replanning |
|-----------|--------------|----------------|
| Single feature failing | ✓ `-recovery-strategy retry` | |
| Transient errors | ✓ `-max-retries 5` | |
| Multiple features failing in sequence | | ✓ `-auto-replan` |
| Plan structure is wrong | | ✓ `-replan` |
| Need to backtrack significantly | | ✓ `-replan-strategy agent` |
| Feature is fundamentally blocked | ✓ `-recovery-strategy skip` | ✓ if many blocked |

### Configuration Summary

| Tier | Flags | Default Values |
|------|-------|----------------|
| Recovery | `-max-retries`, `-recovery-strategy` | 3 retries, retry strategy |
| Replanning | `-auto-replan`, `-replan-threshold`, `-replan-strategy` | disabled, 3 failures, incremental |

---

## Failure Recovery (Tier 1)

Intelligent handling of failures during a single feature's implementation. This is the **first line of defense**.

### Failure Types

| Type | Detection | Example |
|------|-----------|---------|
| `test_failure` | FAIL patterns, assertion errors | Test assertions fail |
| `typecheck_failure` | Syntax, undefined errors | Compilation errors |
| `timeout` | Execution exceeds limit | Long-running operations |
| `agent_error` | Non-zero exit, error output | Agent crashes |

### Recovery Strategies

| Strategy | Action | Best For |
|----------|--------|----------|
| `retry` | Retry with enhanced prompt | Transient issues |
| `skip` | Skip feature, move on | Blocking problems |
| `rollback` | Git reset, then retry | Corrupted state |

### Configuration

```bash
# Set max retries before escalating to skip (default: 3)
ralph -iterations 10 -max-retries 5 -recovery-strategy retry

# Skip problematic features immediately
ralph -iterations 10 -recovery-strategy skip

# Use git to reset and retry fresh
ralph -iterations 10 -recovery-strategy rollback
```

---

## Smart Scope Control

Prevent over-building with iteration and time budgets.

### Constraints

| Constraint | Flag | Description |
|------------|------|-------------|
| Iteration Limit | `-scope-limit` | Max iterations per feature |
| Deadline | `-deadline` | Total time limit |

### Behavior

- Features exceeding limits are **deferred**
- Deferred features marked in plan.json
- Simplification suggested at 50% of limit

### Usage

```bash
# 3 iterations max per feature
ralph -iterations 20 -scope-limit 3

# 2 hour deadline
ralph -iterations 20 -deadline 2h

# View deferred features
ralph -list-deferred
```

---

## Adaptive Replanning (Tier 2)

Dynamic plan adjustment when issues occur repeatedly. This is the **second line of defense** that kicks in when Recovery (Tier 1) alone isn't enough. See [Two-Tier Failure Handling](#two-tier-failure-handling) for the full escalation flow.

### Triggers

| Trigger | Condition | Description |
|---------|-----------|-------------|
| `test_failure` | Consecutive failures >= `-replan-threshold` | Recovery couldn't fix repeated failures |
| `requirement_change` | plan.json modified externally | User edited the plan manually |
| `blocked_feature` | Features become blocked | Multiple features deferred by scope control |
| `manual` | User runs `-replan` | Explicit user request |

### Strategies

| Strategy | Description | Best For |
|----------|-------------|----------|
| `incremental` | Adjust based on current state | Minor plan adjustments |
| `agent` | AI agent restructures plan | Major restructuring needed |
| `none` | Disable replanning | When manual control preferred |

### Plan Versioning

- Backups created before replanning (`plan.json.bak.N`)
- Listed with `-list-versions`
- Restored with `-restore-version N`
- Hash-based deduplication (identical content = one backup)

### Usage

```bash
# Enable auto-replanning when consecutive failures reach threshold
ralph -iterations 10 -auto-replan -replan-threshold 3

# Use AI to restructure the plan
ralph -iterations 10 -auto-replan -replan-strategy agent

# Manual replan (useful when you know the plan structure is wrong)
ralph -replan

# Manage versions
ralph -list-versions
ralph -restore-version 2
```

### Example: Recovery vs Replanning in Action

```
Iteration 1: Feature A fails (test error)
  → Recovery: Retry with enhanced guidance

Iteration 2: Feature A fails again
  → Recovery: Retry (2/3 retries used)

Iteration 3: Feature A fails again
  → Recovery: Max retries exceeded, skip to Feature B
  → Failure counter: 1

Iteration 4: Feature B fails
  → Recovery: Retry with enhanced guidance

Iteration 5: Feature B fails
  → Recovery: Skip to Feature C
  → Failure counter: 2

Iteration 6: Feature C fails
  → Recovery: Skip
  → Failure counter: 3 (equals -replan-threshold)
  → REPLANNING TRIGGERED: Plan restructured
  → Failure counter reset to 0

Iteration 7: Continue with restructured plan...
```

---

## Long-Running Memory

Persistent memory across sessions.

### Memory Types

| Type | Description |
|------|-------------|
| `decision` | Architectural choices |
| `convention` | Coding standards |
| `tradeoff` | Accepted compromises |
| `context` | Project knowledge |

### Agent Extraction

Agents can create memories using markers:
```
[REMEMBER:DECISION]Use PostgreSQL for persistence[/REMEMBER]
```

### Usage

```bash
# View memories
ralph -show-memory

# Add memory
ralph -add-memory "decision:Use PostgreSQL"

# Clear memories
ralph -clear-memory
```

---

## User Nudge Hooks

Lightweight mid-run guidance.

### Nudge Types

| Type | Purpose |
|------|---------|
| `focus` | Prioritize specific work |
| `skip` | Defer certain work |
| `constraint` | Add requirements |
| `style` | Coding preferences |

### Mid-Run Guidance

Edit `nudges.json` while Ralph runs to steer behavior.

### Usage

```bash
# Add nudge
ralph -nudge "focus:Work on feature 5 first"

# View nudges
ralph -show-nudges

# Clear nudges
ralph -clear-nudges
```

---

## Milestone Tracking

Organize features into project milestones.

### Defining Milestones

Add `milestone` field to features:
```json
{
  "id": 1,
  "description": "User auth",
  "milestone": "Alpha",
  "tested": false
}
```

### Progress Tracking

- Automatic progress calculation
- Visual progress bars
- Completion celebrations

### Usage

```bash
# View all milestones
ralph -milestones

# View specific milestone
ralph -milestone Alpha
```

---

## Goal-Oriented Planning

High-level goals decomposed into actionable plans.

### Defining Goals

```bash
# Via CLI
ralph -goal "Add user authentication with OAuth"

# Via goals.json
{
  "goals": [{
    "id": "auth",
    "description": "Add user authentication",
    "priority": 10,
    "success_criteria": ["Users can log in"]
  }]
}
```

### Decomposition

AI agent analyzes goal and generates plan items.

### Usage

```bash
# Add and decompose goal
ralph -goal "Add authentication" -goal-priority 10

# View goals
ralph -list-goals
ralph -goal-status

# Decompose existing goal
ralph -decompose-goal auth
ralph -decompose-all
```

---

## Outcome Validation

Validation beyond tests and type checks.

### Validation Types

| Type | Purpose |
|------|---------|
| `http_get` | Verify HTTP GET endpoint |
| `http_post` | Verify HTTP POST endpoint |
| `cli_command` | Verify CLI execution |
| `file_exists` | Verify file presence |
| `output_contains` | Verify output patterns |

### Defining Validations

```json
{
  "id": 1,
  "description": "Health check",
  "validations": [{
    "type": "http_get",
    "url": "http://localhost:8080/health",
    "expected_status": 200
  }]
}
```

### Usage

```bash
# Validate all completed features
ralph -validate

# Validate specific feature
ralph -validate-feature 5
```

---

## Multi-Agent Collaboration

Coordinate multiple AI agents working in parallel.

### Agent Roles

| Role | Purpose |
|------|---------|
| `implementer` | Create code |
| `tester` | Write tests |
| `reviewer` | Check quality |
| `refactorer` | Improve code |

### Workflow

1. Implementers create code (parallel)
2. Testers validate (parallel)
3. Reviewers check quality (parallel)
4. Refactorers improve (if needed)

### Configuration

```json
{
  "agents": [
    {"id": "impl", "role": "implementer", "command": "cursor-agent"},
    {"id": "test", "role": "tester", "command": "cursor-agent"}
  ],
  "max_parallel": 2,
  "conflict_resolution": "priority"
}
```

### Usage

```bash
# Enable multi-agent
ralph -iterations 10 -multi-agent

# Custom config
ralph -iterations 10 -multi-agent -agents my-agents.json

# List agents
ralph -list-agents
```

---

## Enhanced CLI Output

Rich terminal output with colors and progress.

### Features

- Colored output (success, error, warning)
- Progress spinner during agent execution
- Summary dashboard
- Log levels

### Output Modes

| Mode | Flag | Description |
|------|------|-------------|
| Normal | (default) | Colored terminal output |
| Quiet | `-quiet` | Errors only |
| JSON | `-json-output` | Machine-readable |
| No Color | `-no-color` | Plain text |

### Log Levels

| Level | Shows |
|-------|-------|
| `debug` | Everything |
| `info` | Normal messages |
| `warn` | Warnings and errors |
| `error` | Errors only |

### Usage

```bash
# Quiet mode
ralph -iterations 5 -quiet

# JSON output
ralph -iterations 5 -json-output

# Debug logging
ralph -iterations 5 -log-level debug -verbose
```
