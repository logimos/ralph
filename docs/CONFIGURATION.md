# Ralph Configuration Reference

This document provides a complete reference for all Ralph configuration options.

## Configuration Methods

Ralph supports three methods of configuration (in order of precedence):

1. **CLI Flags** (highest priority)
2. **Environment Variables**
3. **Configuration File** (lowest priority)

## Configuration File

### Supported File Names

Ralph automatically discovers configuration files in this order:

**Current Directory:**
1. `.ralph.yaml`
2. `.ralph.yml`
3. `.ralph.json`
4. `ralph.config.yaml`
5. `ralph.config.yml`
6. `ralph.config.json`

**Home Directory:** (same names as above, used as fallback)

### File Formats

**YAML Format:**
```yaml
# AI Agent
agent: cursor-agent

# Build System
build_system: go
typecheck: go build ./...
test: go test ./...

# File Paths
plan: plan.json
progress: progress.txt
memory_file: .ralph-memory.json
nudge_file: nudges.json
goals_file: goals.json

# Execution
iterations: 5
verbose: true
environment: local

# Recovery
max_retries: 3
recovery_strategy: retry

# Scope Control
scope_limit: 0
deadline: ""

# Replanning
auto_replan: false
replan_strategy: incremental
replan_threshold: 3

# Output
no_color: false
quiet: false
json_output: false
log_level: info

# Memory
memory_retention: 90

# Multi-Agent
agents_file: agents.json
parallel_agents: 2
enable_multi_agent: false
```

**JSON Format:**
```json
{
  "agent": "cursor-agent",
  "build_system": "go",
  "typecheck": "go build ./...",
  "test": "go test ./...",
  "plan": "plan.json",
  "progress": "progress.txt",
  "iterations": 5,
  "verbose": true
}
```

## All Configuration Options

### Core Settings

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Agent | `-agent` | `agent` | `cursor-agent` | AI agent CLI command |
| Build System | `-build-system` | `build_system` | `auto` | Build system preset |
| Type Check | `-typecheck` | `typecheck` | (from preset) | Type checking command |
| Test | `-test` | `test` | (from preset) | Test command |
| Plan File | `-plan` | `plan` | `plan.json` | Path to plan file |
| Progress File | `-progress` | `progress` | `progress.txt` | Path to progress file |
| Iterations | `-iterations` | `iterations` | 0 | Number of iterations |
| Verbose | `-verbose`, `-v` | `verbose` | `false` | Enable verbose output |
| Config File | `-config` | N/A | (auto-discover) | Custom config file path |

### Plan Display Options

| Option | CLI Flag | Description |
|--------|----------|-------------|
| List All | `-list-all` | List all features (tested and untested) |
| List Tested | `-list-tested` | List only tested features |
| List Untested | `-list-untested` | List only untested features |
| List Deferred | `-list-deferred` | List deferred features |
| Status | `-status` | **DEPRECATED**: Use `-list-all` instead |

> **Note:** The `-status` flag is deprecated and will show a warning. Use `-list-all` for the same functionality.

### Build Systems

| Build System | Type Check Command | Test Command | Detection |
|-------------|-------------------|--------------|-----------|
| `go` | `go build ./...` | `go test ./...` | `go.mod` |
| `pnpm` | `pnpm typecheck` | `pnpm test` | `pnpm-lock.yaml` |
| `npm` | `npm run typecheck` | `npm test` | `package.json` |
| `yarn` | `yarn typecheck` | `yarn test` | `yarn.lock` |
| `gradle` | `./gradlew check` | `./gradlew test` | `build.gradle`, `gradlew` |
| `maven` | `mvn compile` | `mvn test` | `pom.xml` |
| `cargo` | `cargo check` | `cargo test` | `Cargo.toml` |
| `python` | `mypy .` | `pytest` | `setup.py`, `pyproject.toml`, `requirements.txt` |
| `auto` | (detected) | (detected) | Automatic detection |

### Environment Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Environment | `-environment` | `environment` | (detected) | Override environment |

**Environment Values:**
- `local` - Local development
- `github-actions`, `github`, `gh` - GitHub Actions
- `gitlab-ci`, `gitlab`, `gl` - GitLab CI
- `jenkins` - Jenkins
- `circleci`, `circle` - CircleCI
- `travis-ci`, `travis` - Travis CI
- `azure-devops`, `azure` - Azure DevOps
- `ci` - Generic CI

### Recovery Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Max Retries | `-max-retries` | `max_retries` | `3` | Retries before escalation |
| Strategy | `-recovery-strategy` | `recovery_strategy` | `retry` | Recovery strategy |

**Recovery Strategies:**
- `retry` - Retry with enhanced prompt
- `skip` - Skip feature, move to next
- `rollback` - Git rollback, then retry

### Scope Control Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Scope Limit | `-scope-limit` | `scope_limit` | `0` | Max iterations per feature (0=unlimited) |
| Deadline | `-deadline` | `deadline` | `""` | Time limit (e.g., "1h", "30m") |

### Replanning Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Auto Replan | `-auto-replan` | `auto_replan` | `false` | Enable automatic replanning |
| Strategy | `-replan-strategy` | `replan_strategy` | `incremental` | Replanning strategy |
| Threshold | `-replan-threshold` | `replan_threshold` | `3` | Failures before replan |

**Replan Strategies:**
- `incremental` - Adjust based on current state
- `agent` - Use AI agent to restructure
- `none` - Disable replanning

### Output Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| No Color | `-no-color` | `no_color` | `false` | Disable colored output |
| Quiet | `-quiet`, `-q` | `quiet` | `false` | Minimal output |
| JSON Output | `-json-output` | `json_output` | `false` | Machine-readable output |
| Log Level | `-log-level` | `log_level` | `info` | Logging verbosity |

**Log Levels:**
- `debug` - All messages
- `info` - Standard messages (default)
- `warn` - Warnings and errors only
- `error` - Errors only

### Memory Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Memory File | `-memory-file` | `memory_file` | `.ralph-memory.json` | Memory file path |
| Retention | `-memory-retention` | `memory_retention` | `90` | Days to retain memories |
| Show Memory | `-show-memory` | N/A | N/A | Display memories |
| Clear Memory | `-clear-memory` | N/A | N/A | Clear all memories |
| Add Memory | `-add-memory` | N/A | N/A | Add memory entry |

### Nudge Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Nudge File | `-nudge-file` | `nudge_file` | `nudges.json` | Nudge file path |
| Nudge | `-nudge` | N/A | N/A | Add one-time nudge |
| Show Nudges | `-show-nudges` | N/A | N/A | Display nudges |
| Clear Nudges | `-clear-nudges` | N/A | N/A | Clear all nudges |

### Milestone Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Milestones | `-milestones` | N/A | N/A | List all milestones |
| Milestone | `-milestone` | N/A | N/A | Show specific milestone |

### Goal Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Goals File | `-goals-file` | `goals_file` | `goals.json` | Goals file path |
| Show Goals | `-goals` | N/A | N/A | Show all goals with progress |
| Goal | `-goal` | N/A | N/A | Add a goal |
| Goal Priority | `-goal-priority` | N/A | `5` | Priority for new goal |
| Decompose Goal | `-decompose-goal` | N/A | N/A | Decompose specific goal |
| Decompose All | `-decompose-all` | N/A | N/A | Decompose all pending |

> **Note**: The `-goal-status` and `-list-goals` flags are deprecated. Use `-goals` instead to show all goals with progress.

### Validation Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Validate | `-validate` | N/A | N/A | Run all validations |
| Validate Feature | `-validate-feature` | N/A | N/A | Validate specific feature |

### Multi-Agent Options

| Option | CLI Flag | Config Key | Default | Description |
|--------|----------|------------|---------|-------------|
| Enable | `-multi-agent` | `enable_multi_agent` | `false` | Enable multi-agent |
| Agents File | `-agents` | `agents_file` | `agents.json` | Agents config file |
| Parallel | `-parallel-agents` | `parallel_agents` | `2` | Max parallel agents |
| List Agents | `-list-agents` | N/A | N/A | List configured agents |

### Plan Analysis & Refinement Options

Ralph can analyze your plan.json for potential improvements and apply refinements.

| Option | CLI Flag | Description |
|--------|----------|-------------|
| Analyze Plan | `-analyze-plan` | Analyze plan (read-only), write preview to `plan.refined.json` |
| Refine Plan | `-refine-plan` | Apply refinements (writes to `plan.json`, creates backup) |
| Dry Run | `-dry-run` | Preview changes without writing (use with `-refine-plan`) |

**Two-step workflow:**
1. **Analyze first**: Run `-analyze-plan` to see suggestions and generate a preview file
2. **Review**: Compare `plan.json` with `plan.refined.json` using `diff`
3. **Apply**: Run `-refine-plan` to apply the changes (or `-refine-plan -dry-run` to preview)

**What gets detected:**
- **Compound features**: Descriptions with "and" suggesting multiple features
- **Complex features**: Features with >9 steps that may need splitting

**Example workflow:**
```bash
# Step 1: Analyze and generate preview
ralph -analyze-plan
# This shows analysis and writes proposed changes to plan.refined.json

# Step 2: Review the diff
diff plan.json plan.refined.json

# Step 3a: Preview what refinements would be applied
ralph -refine-plan -dry-run

# Step 3b: Apply the refinements
ralph -refine-plan
# This writes to plan.json and creates plan.json.bak as backup
```

### Version Options

| Option | CLI Flag | Description |
|--------|----------|-------------|
| Version | `-version` | Show version and exit |
| Help | `-help`, `-h` | Show help and exit |

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
environment: ci
verbose: true
max_retries: 5
auto_replan: true
json_output: true
```

### Full Feature Set
```yaml
# Agent
agent: cursor-agent

# Build
build_system: go
typecheck: go build ./...
test: go test -race ./...

# Files
plan: plan.json
progress: progress.txt
memory_file: .ralph-memory.json
nudge_file: nudges.json
goals_file: goals.json

# Execution
iterations: 10
verbose: true

# Recovery
max_retries: 3
recovery_strategy: retry

# Scope
scope_limit: 5
deadline: "2h"

# Replanning
auto_replan: true
replan_strategy: incremental
replan_threshold: 3

# Output
log_level: info
no_color: false

# Memory
memory_retention: 90
```

## Configuration Precedence Examples

```bash
# Config file has: iterations: 5
# CLI override wins:
ralph -iterations 10  # Uses 10 iterations

# Config file has: verbose: true
# CLI can override:
ralph -verbose=false  # Disables verbose

# No config file, CLI provides:
ralph -build-system gradle -iterations 3
```
