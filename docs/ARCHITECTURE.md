# Ralph Architecture

This document describes the architecture and design of Ralph.

## Overview

Ralph is a CLI tool that orchestrates AI-assisted development workflows. It processes plan files, executes development tasks through AI agents, validates work, and tracks progress.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          ralph.go                                │
│                     (CLI Entry Point)                           │
├─────────────────────────────────────────────────────────────────┤
│  • Flag parsing and validation                                   │
│  • Configuration loading                                         │
│  • Command routing                                               │
│  • Main iteration loop orchestration                            │
└──────────────────────────┬──────────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
          ▼                ▼                ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────────┐
│   config/   │  │   plan/     │  │     agent/      │
│             │  │             │  │                 │
│ • Config    │  │ • Plan      │  │ • IsCursorAgent │
│ • FileConfig│  │ • ReadFile  │  │ • Execute       │
│ • Load      │  │ • Filter    │  │                 │
│ • Validate  │  │ • Write     │  │                 │
└─────────────┘  └─────────────┘  └─────────────────┘
          │                │                │
          └────────────────┼────────────────┘
                           │
┌──────────────────────────┼──────────────────────────────────────┐
│                    Internal Packages                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │detection │ │environment│ │ prompt   │ │   ui     │           │
│  │          │ │           │ │          │ │          │           │
│  │Build sys │ │CI detect  │ │ Build    │ │ Colors   │           │
│  │detection │ │Resources  │ │ prompts  │ │ Progress │           │
│  │Presets   │ │Adaptation │ │          │ │ Spinner  │           │
│  └──────────┘ └───────────┘ └──────────┘ └──────────┘           │
│                                                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ memory   │ │  nudge   │ │milestone │ │  goals   │           │
│  │          │ │          │ │          │ │          │           │
│  │Cross-sess│ │Mid-run   │ │Progress  │ │Goal mgmt │           │
│  │memory    │ │guidance  │ │tracking  │ │Decompose │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
│                                                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ recovery │ │  replan  │ │  scope   │ │validation│           │
│  │          │ │          │ │          │ │          │           │
│  │Failure   │ │Adaptive  │ │Iteration │ │Outcome   │           │
│  │handling  │ │replanning│ │budgets   │ │validation│           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
│                                                                  │
│  ┌──────────┐                                                    │
│  │multiagent│                                                    │
│  │          │                                                    │
│  │Agent     │                                                    │
│  │orchestra.│                                                    │
│  └──────────┘                                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Package Responsibilities

### Core Packages

#### `ralph.go` (Main Package)
- Entry point for CLI
- Flag parsing and validation
- Configuration precedence handling
- Command routing (iterations, status, generate-plan, etc.)
- Main iteration loop orchestration
- Integration of all subsystems

#### `internal/config`
- `Config` struct with all runtime settings
- `FileConfig` for configuration file parsing
- Configuration file discovery and loading
- YAML and JSON format support
- Validation and defaults

#### `internal/plan`
- `Plan` struct representing feature items
- File operations (read, write, filter)
- Plan validation
- JSON serialization

#### `internal/agent`
- Agent type detection (Cursor, Claude, etc.)
- Command construction for different agents
- Agent execution wrapper

#### `internal/prompt`
- Iteration prompt construction
- Plan generation prompt construction
- Completion signal handling

### Detection and Environment

#### `internal/detection`
- Build system detection (Go, Node, Python, etc.)
- Build system presets (typecheck, test commands)
- Project file scanning

#### `internal/environment`
- CI environment detection (GitHub Actions, GitLab CI, etc.)
- System resource detection (CPU, memory)
- Project complexity estimation
- Execution recommendations (timeout, parallelism)

### User Interface

#### `internal/ui`
- Colored output formatting
- Progress bar and spinner
- Log levels (debug, info, warn, error)
- Summary dashboard
- JSON output mode
- TTY detection

### Memory and State

#### `internal/memory`
- Cross-session memory storage
- Memory entry types (decision, convention, tradeoff, context)
- Memory extraction from agent output
- Memory injection into prompts
- Relevance scoring

#### `internal/nudge`
- User nudge system
- File-based nudge storage
- Mid-run guidance detection
- Nudge injection into prompts
- Acknowledgment tracking

#### `internal/milestone`
- Milestone extraction from plans
- Progress calculation
- Status tracking
- Milestone completion detection

### Execution Control

#### `internal/recovery`
- Failure type detection
- Recovery strategies (retry, skip, rollback)
- Failure tracking per feature
- Strategy escalation

#### `internal/scope`
- Iteration budget tracking
- Time-based deadlines
- Feature complexity estimation
- Automatic deferral
- Simplification suggestions

#### `internal/replan`
- Replan triggers (test failure, requirement change, blocked)
- Replan strategies (incremental, agent-based)
- Plan versioning and backup
- Plan diff computation

### Advanced Features

#### `internal/goals`
- High-level goal management
- Goal-to-plan decomposition
- Progress tracking
- Dependency handling

#### `internal/validation`
- Outcome-focused validation
- HTTP endpoint validation
- CLI command validation
- File existence validation
- Output pattern validation

#### `internal/multiagent`
- Agent role system
- Agent orchestration
- Parallel execution
- Shared context
- Conflict resolution

## Data Flow

### Iteration Execution Flow

```
1. Configuration Loading
   ralph.go → config.Load() → FileConfig → Config

2. Plan Loading
   ralph.go → plan.ReadFile() → []Plan

3. Environment Detection
   ralph.go → environment.Detect() → EnvironmentProfile

4. Memory/Nudge Loading
   ralph.go → memory.Load() + nudge.Load()

5. For each iteration:
   a. Select feature (first untested, non-deferred)
   b. Build prompt (prompt.BuildIterationPrompt)
   c. Inject memory/nudge context
   d. Execute agent (agent.Execute)
   e. Detect failures (recovery.DetectFailure)
   f. Apply recovery if needed
   g. Extract memories from output
   h. Check scope constraints
   i. Update progress file
   j. Check for completion signal

6. Summary and cleanup
```

### Configuration Precedence

```
Defaults → Config File → Environment Variables → CLI Flags
   ↓           ↓                   ↓                ↓
 Lowest                                          Highest
Priority                                        Priority
```

## Design Principles

### 1. Package Independence
Each internal package is designed to be as independent as possible:
- Minimal cross-package dependencies
- Clear interfaces for integration
- Testable in isolation

### 2. Configuration Flexibility
Multiple layers of configuration support different use cases:
- Defaults for quick start
- Config files for project-specific settings
- CLI flags for one-off overrides

### 3. Graceful Degradation
Features are optional and degrade gracefully:
- Memory system works without memory file
- Milestones are optional
- Recovery is configurable

### 4. Extensibility
Design supports future extensions:
- New build systems via presets
- New agent types via detection
- New validation types via interface
- New recovery strategies via interface

## Key Interfaces

### RecoveryStrategy
```go
type RecoveryStrategy interface {
    Name() string
    Apply(failure Failure, tracker *FailureTracker) RecoveryResult
}
```

### ReplanTrigger
```go
type ReplanTrigger interface {
    Name() string
    ShouldTrigger(state ReplanState) bool
}
```

### ReplanStrategy
```go
type ReplanStrategy interface {
    Name() string
    GenerateNewPlan(state ReplanState) (*ReplanResult, error)
}
```

### Validator
```go
type Validator interface {
    Validate() ValidationResult
}
```

## File Formats

### plan.json
```json
[
  {
    "id": 1,
    "category": "feature",
    "description": "...",
    "steps": ["..."],
    "expected_output": "...",
    "tested": false,
    "milestone": "Alpha",
    "deferred": false,
    "validations": [...]
  }
]
```

### .ralph.yaml
```yaml
agent: cursor-agent
build_system: go
plan: plan.json
iterations: 5
verbose: true
```

### .ralph-memory.json
```json
{
  "entries": [...],
  "last_updated": "...",
  "retention_days": 90
}
```

## Testing Strategy

### Unit Tests
- Each package has corresponding `*_test.go` files
- Table-driven tests for comprehensive coverage
- Mock interfaces for external dependencies

### Integration Tests
- Test full iteration flow with mock agent
- Test configuration loading precedence
- Test file operations

### Coverage Targets
- Core logic: >80%
- Utility functions: >70%
- Integration points: manual testing

## Performance Considerations

1. **File I/O**: Minimize file reads during iteration loop
2. **Memory**: Prune old entries to prevent unbounded growth
3. **Agent Execution**: Main bottleneck is AI agent response time
4. **Parallel Execution**: Multi-agent mode supports concurrent execution

## Security Considerations

1. **File Access**: Only accesses files in working directory
2. **Secrets**: Never store secrets in memory/nudge files
3. **Agent Commands**: Validate agent command before execution
4. **Git Operations**: Rollback strategy only affects tracked files
