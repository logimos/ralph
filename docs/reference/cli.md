# CLI Reference

Complete reference for all Ralph command-line options.

## Usage

```bash
ralph [options]
```

## Core Options

| Flag | Default | Description |
|------|---------|-------------|
| `-iterations` | 0 | Number of iterations to run |
| `-agent` | cursor-agent | AI agent command |
| `-plan` | plan.json | Path to plan file |
| `-progress` | progress.txt | Path to progress file |
| `-config` | (auto) | Path to config file |
| `-build-system` | auto | Build system preset |
| `-typecheck` | (preset) | Type check command |
| `-test` | (preset) | Test command |
| `-verbose`, `-v` | false | Enable verbose output |
| `-version` | - | Show version and exit |

## Plan Display

| Flag | Description |
|------|-------------|
| `-list-all` | List all features |
| `-list-tested` | List completed features |
| `-list-untested` | List remaining features |
| `-list-deferred` | List deferred features |
| `-status` | _(deprecated)_ Use `-list-all` |

## Plan Analysis

| Flag | Description |
|------|-------------|
| `-analyze-plan` | Analyze plan, write preview to plan.refined.json |
| `-refine-plan` | Apply refinements to plan.json |
| `-dry-run` | Preview changes without writing |
| `-generate-plan` | Generate plan from notes |
| `-notes` | Path to notes file (with -generate-plan) |
| `-output` | Output plan file path |

## Recovery (Per-Feature)

| Flag | Default | Description |
|------|---------|-------------|
| `-max-retries` | 3 | Max retries before escalation |
| `-recovery-strategy` | retry | Strategy: retry, skip, rollback |

## Replanning (Plan-Level)

| Flag | Default | Description |
|------|---------|-------------|
| `-auto-replan` | false | Enable automatic replanning |
| `-replan` | - | Manually trigger replanning |
| `-replan-strategy` | incremental | Strategy: incremental, agent, none |
| `-replan-threshold` | 3 | Failures before replanning |
| `-list-versions` | - | List plan backup versions |
| `-restore-version` | - | Restore a specific version |

## Scope Control

| Flag | Default | Description |
|------|---------|-------------|
| `-scope-limit` | 0 | Max iterations per feature (0=unlimited) |
| `-deadline` | - | Time limit (e.g., "2h", "30m") |

## Memory System

| Flag | Default | Description |
|------|---------|-------------|
| `-memory-file` | .ralph-memory.json | Memory file path |
| `-show-memory` | - | Display stored memories |
| `-clear-memory` | - | Clear all memories |
| `-add-memory` | - | Add memory (format: type:content) |
| `-memory-retention` | 90 | Days to retain memories |

## Nudge System

| Flag | Default | Description |
|------|---------|-------------|
| `-nudge-file` | nudges.json | Nudge file path |
| `-nudge` | - | Add one-time nudge (format: type:content) |
| `-show-nudges` | - | Display current nudges |
| `-clear-nudges` | - | Clear all nudges |

## Milestones

| Flag | Description |
|------|-------------|
| `-milestones` | List all milestones with progress |
| `-milestone` | Show features for specific milestone |

## Goals

| Flag | Default | Description |
|------|---------|-------------|
| `-goals-file` | goals.json | Goals file path |
| `-goal` | - | Add a goal to decompose |
| `-goal-priority` | 5 | Priority for new goal |
| `-goals` | - | Show all goals with progress |
| `-decompose-goal` | - | Decompose specific goal |
| `-decompose-all` | - | Decompose all pending goals |
| `-goal-status` | _(deprecated)_ | Use `-goals` |
| `-list-goals` | _(deprecated)_ | Use `-goals` |

## Validation

| Flag | Description |
|------|-------------|
| `-validate` | Run validations for completed features |
| `-validate-feature` | Validate specific feature by ID |

## Multi-Agent

| Flag | Default | Description |
|------|---------|-------------|
| `-multi-agent` | false | Enable multi-agent mode |
| `-agents` | agents.json | Agents config file path |
| `-parallel-agents` | 2 | Max parallel agents |
| `-list-agents` | - | List configured agents |

## Output & UI

| Flag | Default | Description |
|------|---------|-------------|
| `-no-color` | false | Disable colored output |
| `-quiet`, `-q` | false | Minimal output (errors only) |
| `-json-output` | false | Machine-readable JSON output |
| `-log-level` | info | Level: debug, info, warn, error |

## Environment

| Flag | Default | Description |
|------|---------|-------------|
| `-environment` | (auto) | Override detected environment |

## Examples

```bash
# Basic usage
ralph -iterations 5

# Verbose with custom agent
ralph -iterations 5 -verbose -agent claude

# With recovery settings
ralph -iterations 10 -max-retries 5 -recovery-strategy retry

# With scope control
ralph -iterations 20 -scope-limit 3 -deadline 1h

# Enable auto-replan
ralph -iterations 10 -auto-replan -replan-threshold 3

# Generate plan from notes
ralph -generate-plan -notes notes.md -output my-plan.json

# Memory operations
ralph -show-memory
ralph -add-memory "decision:Use PostgreSQL"
ralph -clear-memory

# Nudge operations
ralph -nudge "focus:Work on feature 5"
ralph -show-nudges
ralph -clear-nudges

# Milestone tracking
ralph -milestones
ralph -milestone "Alpha"

# Goal management
ralph -goal "Add authentication" -goal-priority 10
ralph -goals
ralph -decompose-all

# Validation
ralph -validate
ralph -validate-feature 5

# Multi-agent
ralph -iterations 10 -multi-agent -parallel-agents 4

# CI-friendly output
ralph -iterations 5 -json-output -quiet
```
