# Configuration Reference

Complete reference for all Ralph configuration options.

## Configuration Precedence

1. **CLI Flags** (highest priority)
2. **Environment Variables** (CI detection)
3. **Configuration File** (lowest priority)

## Configuration File

### File Names

Ralph discovers config files in order:

**Current Directory:**
1. `.ralph.yaml`
2. `.ralph.yml`
3. `.ralph.json`
4. `ralph.config.yaml`
5. `ralph.config.yml`
6. `ralph.config.json`

**Home Directory:** Same names as fallback

### Complete Configuration

```yaml
# .ralph.yaml - All available options

# ═══════════════════════════════════════════════════════════════
# Core Settings
# ═══════════════════════════════════════════════════════════════

# AI agent CLI command
agent: cursor-agent

# Build system preset: go, npm, pnpm, yarn, gradle, maven, cargo, python, auto
build_system: go

# Override type check command
typecheck: go build ./...

# Override test command
test: go test ./...

# Plan file path
plan: plan.json

# Progress file path
progress: progress.txt

# Number of iterations
iterations: 5

# Verbose output
verbose: true

# ═══════════════════════════════════════════════════════════════
# Recovery (Per-Feature)
# ═══════════════════════════════════════════════════════════════

# Max retries before escalation
max_retries: 3

# Recovery strategy: retry, skip, rollback
recovery_strategy: retry

# ═══════════════════════════════════════════════════════════════
# Scope Control
# ═══════════════════════════════════════════════════════════════

# Max iterations per feature (0 = unlimited)
scope_limit: 0

# Time limit (e.g., "2h", "30m", "1h30m")
deadline: ""

# ═══════════════════════════════════════════════════════════════
# Replanning (Plan-Level)
# ═══════════════════════════════════════════════════════════════

# Enable automatic replanning
auto_replan: false

# Replanning strategy: incremental, agent, none
replan_strategy: incremental

# Consecutive failures before replanning
replan_threshold: 3

# ═══════════════════════════════════════════════════════════════
# Memory System
# ═══════════════════════════════════════════════════════════════

# Memory file path
memory_file: .ralph-memory.json

# Days to retain memories
memory_retention: 90

# ═══════════════════════════════════════════════════════════════
# Nudge System
# ═══════════════════════════════════════════════════════════════

# Nudge file path
nudge_file: nudges.json

# ═══════════════════════════════════════════════════════════════
# Goals
# ═══════════════════════════════════════════════════════════════

# Goals file path
goals_file: goals.json

# ═══════════════════════════════════════════════════════════════
# Multi-Agent
# ═══════════════════════════════════════════════════════════════

# Agents configuration file path
agents_file: agents.json

# Max parallel agents
parallel_agents: 2

# Enable multi-agent mode
enable_multi_agent: false

# ═══════════════════════════════════════════════════════════════
# Output & UI
# ═══════════════════════════════════════════════════════════════

# Disable colored output
no_color: false

# Minimal output (errors only)
quiet: false

# Machine-readable JSON output
json_output: false

# Log level: debug, info, warn, error
log_level: info

# ═══════════════════════════════════════════════════════════════
# Environment
# ═══════════════════════════════════════════════════════════════

# Override detected environment
# Values: local, github-actions, gitlab-ci, jenkins, circleci, travis-ci, azure-devops, ci
environment: ""
```

## Build Systems

| System | Detection | Type Check | Test |
|--------|-----------|------------|------|
| `go` | `go.mod` | `go build ./...` | `go test ./...` |
| `pnpm` | `pnpm-lock.yaml` | `pnpm typecheck` | `pnpm test` |
| `npm` | `package.json` | `npm run typecheck` | `npm test` |
| `yarn` | `yarn.lock` | `yarn typecheck` | `yarn test` |
| `gradle` | `build.gradle` | `./gradlew check` | `./gradlew test` |
| `maven` | `pom.xml` | `mvn compile` | `mvn test` |
| `cargo` | `Cargo.toml` | `cargo check` | `cargo test` |
| `python` | `setup.py`, `pyproject.toml` | `mypy .` | `pytest` |
| `auto` | (detected) | (detected) | (detected) |

## Environment Detection

| Environment | Detection Variable |
|-------------|-------------------|
| Local | Default (no CI vars) |
| GitHub Actions | `GITHUB_ACTIONS` |
| GitLab CI | `GITLAB_CI` |
| Jenkins | `JENKINS_URL` |
| CircleCI | `CIRCLECI` |
| Travis CI | `TRAVIS` |
| Azure DevOps | `TF_BUILD` |
| Generic CI | `CI` |

## Example Configurations

### Minimal Go Project

```yaml
build_system: go
```

### Node.js with pnpm

```yaml
build_system: pnpm
plan: tasks/plan.json
verbose: true
```

### CI-Optimized

```yaml
build_system: go
verbose: true
max_retries: 5
auto_replan: true
json_output: true
```

### Full Feature Set

```yaml
agent: cursor-agent
build_system: go
typecheck: go build ./...
test: go test -race ./...

plan: plan.json
progress: progress.txt
memory_file: .ralph-memory.json
nudge_file: nudges.json
goals_file: goals.json

iterations: 10
verbose: true

max_retries: 3
recovery_strategy: retry

scope_limit: 5
deadline: "2h"

auto_replan: true
replan_strategy: incremental
replan_threshold: 3

log_level: info
no_color: false
memory_retention: 90
```
