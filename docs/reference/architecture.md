# Architecture

Ralph's internal architecture and package structure.

## System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         ralph.go                            │
│                    (CLI Entry Point)                        │
└─────────────────────────────┬───────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│    config/    │    │     plan/     │    │    agent/     │
│ Configuration │    │ Plan Mgmt     │    │ AI Execution  │
└───────────────┘    └───────────────┘    └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│   recovery/   │    │    replan/    │    │    scope/     │
│ Tier 1 Errors │    │ Tier 2 Errors │    │ Scope Control │
└───────────────┘    └───────────────┘    └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│    memory/    │    │    nudge/     │    │  milestone/   │
│ Persistent    │    │ Mid-Run       │    │ Progress      │
│ Memory        │    │ Guidance      │    │ Tracking      │
└───────────────┘    └───────────────┘    └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│    goals/     │    │  validation/  │    │  multiagent/  │
│ Goal-Oriented │    │ Outcome       │    │ Multi-Agent   │
│ Planning      │    │ Validation    │    │ Coordination  │
└───────────────┘    └───────────────┘    └───────────────┘
```

## Package Structure

```
ralph/
├── ralph.go              # Main entry point, CLI parsing
├── ralph_test.go         # Main tests
├── Makefile              # Build and release commands
└── internal/
    ├── config/           # Configuration management
    │   ├── config.go     # Config struct, defaults
    │   └── file.go       # File loading, validation
    │
    ├── plan/             # Plan file operations
    │   ├── plan.go       # Plan struct, CRUD
    │   └── analyze.go    # Plan analysis, refinement
    │
    ├── agent/            # AI agent execution
    │   └── agent.go      # Execute, IsCursorAgent
    │
    ├── prompt/           # Prompt construction
    │   └── prompt.go     # Build prompts
    │
    ├── detection/        # Build system detection
    │   └── detection.go  # Detect, presets
    │
    ├── environment/      # Environment detection
    │   └── environment.go# CI detection, resources
    │
    ├── recovery/         # Tier 1: Per-feature recovery
    │   ├── recovery.go   # Failure detection
    │   └── strategy.go   # Recovery strategies
    │
    ├── replan/           # Tier 2: Plan-level replanning
    │   ├── replan.go     # Triggers, versioning
    │   └── strategy.go   # Replan strategies
    │
    ├── scope/            # Scope control
    │   └── scope.go      # Constraints, deferral
    │
    ├── memory/           # Persistent memory
    │   └── memory.go     # Store, extraction
    │
    ├── nudge/            # Nudge system
    │   └── nudge.go      # Store, file watching
    │
    ├── milestone/        # Milestone tracking
    │   └── milestone.go  # Manager, progress
    │
    ├── goals/            # Goal-oriented planning
    │   ├── goals.go      # Goal management
    │   └── decompose.go  # AI decomposition
    │
    ├── validation/       # Outcome validation
    │   └── validation.go # Validators, runner
    │
    ├── multiagent/       # Multi-agent coordination
    │   └── multiagent.go # Orchestrator
    │
    └── ui/               # CLI output
        └── ui.go         # Colors, progress, summary
```

## Key Interfaces

### RecoveryStrategy

```go
type RecoveryStrategy interface {
    Apply(failure Failure, prompt string) RecoveryResult
    Name() string
}
```

### ReplanStrategy

```go
type ReplanStrategy interface {
    GenerateNewPlan(state ReplanState) (*ReplanResult, error)
    Name() string
}
```

### Validator

```go
type Validator interface {
    Validate() ValidationResult
    Description() string
    Type() string
}
```

## Data Flow

### Iteration Flow

```
1. Load configuration (config.go)
2. Load plan file (plan.go)
3. Detect environment (environment.go)
4. Initialize recovery manager (recovery.go)
5. Initialize scope manager (scope.go)
6. Load memory/nudges (memory.go, nudge.go)

For each iteration:
    7. Find next untested feature (plan.go)
    8. Build prompt with context (prompt.go)
    9. Inject memory and nudges
    10. Execute agent (agent.go)
    11. Detect failures (recovery.go)
    12. Apply recovery strategy if needed
    13. Check replan triggers (replan.go)
    14. Update plan file
    15. Check scope constraints (scope.go)
    16. Extract memories from output (memory.go)
    17. Acknowledge nudges (nudge.go)

18. Print summary (ui.go)
```

## Configuration Flow

```
Defaults → Config File → CLI Flags
    ↓           ↓            ↓
   merge      merge       override
              ↓
         Final Config
```

## Recovery Flow

```
Failure → Detect Type → Select Strategy → Apply
              ↓                               ↓
    test_failure    ←→    RetryStrategy    → Retry with guidance
    typecheck_failure ←→  SkipStrategy     → Skip to next
    timeout         ←→    RollbackStrategy → Git reset, retry
    agent_error
```

## File Formats

| File | Format | Purpose |
|------|--------|---------|
| `plan.json` | JSON array | Feature definitions |
| `progress.txt` | Plain text | Execution log |
| `.ralph.yaml` | YAML | Configuration |
| `.ralph-memory.json` | JSON | Persistent memory |
| `nudges.json` | JSON | User guidance |
| `goals.json` | JSON | High-level goals |
| `agents.json` | JSON | Multi-agent config |
