# Features Overview

Ralph provides a comprehensive set of features for orchestrating AI-assisted development workflows.

## Core Capabilities

### Plan Management
Track features, generate plans from notes, and organize work into milestones.

- **Plan Files**: JSON-based feature definitions with steps and expected outputs
- **Generation**: Convert notes to structured plans using AI
- **Status Tracking**: View tested, untested, and deferred features
- **Analysis**: Detect complex features and suggest refinements

[Learn more about Plan Management →](plan-management.md)

### Failure Handling (Two-Tier System)

Ralph uses a sophisticated two-tier system to handle failures:

1. **Tier 1: Recovery** (per-feature)
    - Retry with enhanced prompts
    - Skip problematic features
    - Rollback via git

2. **Tier 2: Replanning** (plan-level)
    - Automatic when failures exceed threshold
    - Restructures the plan dynamically

[Learn more about Failure Recovery →](failure-recovery.md)

### Environment Adaptation

Ralph automatically detects and adapts to:

- Local development
- CI/CD environments (GitHub Actions, GitLab CI, Jenkins, etc.)
- System resources (CPU, memory)
- Project complexity

## Advanced Features

### Memory System

Persistent storage for architectural decisions and coding conventions:

- **Types**: decisions, conventions, tradeoffs, context
- **Extraction**: AI agents can create memories using markers
- **Injection**: Relevant memories added to prompts automatically

[Learn more about Memory →](memory.md)

### Nudge System

Lightweight mid-run guidance:

- **Types**: focus, skip, constraint, style
- **Real-time**: Edit nudges while Ralph runs
- **Acknowledgment**: Nudges processed once and marked complete

[Learn more about Nudges →](nudges.md)

### Milestone Tracking

Organize features into project milestones:

- Progress visualization with bars
- Completion celebrations
- Milestone-based filtering

[Learn more about Milestones →](milestones.md)

### Goal-Oriented Planning

Define high-level goals that get decomposed into actionable plans:

- AI-powered decomposition
- Progress tracking toward goals
- Dependency management between goals

[Learn more about Goals →](goals.md)

### Outcome Validation

Validate outcomes beyond unit tests:

- HTTP endpoint validation
- CLI command verification
- File existence checks
- Output pattern matching

[Learn more about Validation →](validation.md)

### Multi-Agent Collaboration

Coordinate multiple AI agents:

- **Roles**: implementer, tester, reviewer, refactorer
- **Parallel execution**: Multiple agents working simultaneously
- **Conflict resolution**: Priority, merge, or vote strategies

[Learn more about Multi-Agent →](multi-agent.md)

## Feature Matrix

| Feature | Local | CI | Config File | CLI Flag |
|---------|-------|----|-----------|-----------| 
| Plan Management | ✓ | ✓ | - | ✓ |
| Failure Recovery | ✓ | ✓ | ✓ | ✓ |
| Scope Control | ✓ | ✓ | ✓ | ✓ |
| Replanning | ✓ | ✓ | ✓ | ✓ |
| Memory System | ✓ | ✓ | ✓ | ✓ |
| Nudge System | ✓ | - | ✓ | ✓ |
| Milestones | ✓ | ✓ | - | ✓ |
| Goals | ✓ | ✓ | ✓ | ✓ |
| Validation | ✓ | ✓ | - | ✓ |
| Multi-Agent | ✓ | ✓ | ✓ | ✓ |
| CLI Output | ✓ | ✓ | ✓ | ✓ |
