# Ralph - AI-Assisted Development Workflow CLI

Ralph is a Golang CLI application that automates iterative development workflows by orchestrating AI-assisted development cycles. It processes plan files, executes development tasks through an AI agent CLI tool (like Cursor Agent or Claude), validates work, tracks progress, and commits changes iteratively until completion.

## Features

- **Iterative Development**: Process features one at a time based on priority
- **AI Agent Integration**: Works with Cursor Agent, Claude, or any compatible CLI tool
- **Plan Management**: Generate plans from notes, list tested/untested features
- **Progress Tracking**: Automatically updates plan files and progress logs
- **Milestone Tracking**: Organize features into milestones with progress visualization
- **Validation**: Integrates with type checking and testing commands
- **Outcome Validation**: Validate HTTP endpoints, CLI commands, and file existence beyond unit tests
- **Git Integration**: Creates commits for completed features
- **Completion Detection**: Automatically detects when all work is complete
- **Failure Recovery**: Automatically handles failures with configurable recovery strategies
- **Environment Detection**: Automatically adapts to CI and local environments
- **Long-Running Memory**: Remembers architectural decisions and conventions across sessions
- **Nudge System**: Lightweight mid-run guidance without stopping execution
- **Smart Scope Control**: Iteration budgets and time limits to prevent over-building
- **Adaptive Replanning**: Dynamically adjusts plans when tests fail repeatedly or requirements change
- **Goal-Oriented Planning**: Define high-level goals and let AI decompose them into actionable plans
- **Multi-Agent Collaboration**: Coordinate multiple AI agents (implementer, tester, reviewer, refactorer) working in parallel
- **Plan Analysis**: Analyze plans for refinement opportunities (compound and complex features)

## Installation

### Using Go CLI (Recommended)

Install the latest version directly from GitHub:

```bash
go install github.com/start-it/ralph@latest
```

Install a specific version:

```bash
go install github.com/start-it/ralph@v1.0.0
```

Install from the main branch:

```bash
go install github.com/start-it/ralph@main
```

After installation, make sure `$GOPATH/bin` or `$GOBIN` is in your PATH. The binary will be available as `ralph`.

### From Source

```bash
git clone <repository-url>
cd start-it
make build
```

### Local Development Install

```bash
make install-local
```

This builds and installs the current version to `$GOPATH/bin` or `$GOBIN` for testing your local changes. Perfect for iterating on code changes without manually copying binaries.

## Usage

### Check Version

```bash
# Show version information
ralph -version
```

### Basic Workflow

Run iterative development cycles:

```bash
# Run 5 iterations (auto-detects build system)
ralph -iterations 5

# With verbose output
ralph -iterations 3 -verbose

# Use a different plan file
ralph -iterations 5 -plan my-plan.json

# Use Cursor Agent (default) or Claude
ralph -iterations 5 -agent cursor-agent
ralph -iterations 5 -agent claude

# Specify build system explicitly
ralph -iterations 5 -build-system gradle
ralph -iterations 5 -build-system maven
ralph -iterations 5 -build-system cargo
```

### Plan Management

#### Generate Plan from Notes

Convert detailed notes into a structured plan.json:

```bash
# Generate plan.json from notes
ralph -generate-plan -notes .hidden/notes.md

# Custom output file
ralph -generate-plan -notes notes.md -output my-plan.json

# With verbose output
ralph -generate-plan -notes notes.md -verbose
```

#### View Plan Status

```bash
# Show both tested and untested features
ralph -status

# List only tested features
ralph -list-tested

# List only untested features
ralph -list-untested

# Use a different plan file
ralph -status -plan test.json
```

### Command-Line Options

```
Usage: ralph [options]

Options:
  -agent string
        Command name for the AI agent CLI tool (default "cursor-agent")
  -build-system string
        Build system preset (pnpm, npm, yarn, gradle, maven, cargo, go, python) or 'auto' for detection
  -config string
        Path to configuration file (default: auto-discover .ralph.yaml, .ralph.json)
  -environment string
        Override detected environment (local, github-actions, gitlab-ci, jenkins, circleci, ci)
  -generate-plan
        Generate plan.json from notes file
  -iterations int
        Number of iterations to run (required)
  -json-output
        Machine-readable JSON output for automation
  -list-tested
        List only tested features
  -list-untested
        List only untested features
  -log-level string
        Log verbosity: debug, info, warn, error (default "info")
  -max-retries int
        Maximum retries per feature before escalation (default: 3)
  -no-color
        Disable colored output (auto-disabled in non-TTY)
  -notes string
        Path to notes file (required with -generate-plan)
  -output string
        Output plan file path (default: plan.json)
  -plan string
        Path to the plan file (e.g., plan.json) (default "plan.json")
  -progress string
        Path to the progress file (e.g., progress.txt) (default "progress.txt")
  -quiet, -q
        Minimal output (errors only)
  -recovery-strategy string
        Recovery strategy: retry, skip, rollback (default: retry)
  -status
        List plan status (tested and untested features)
  -test string
        Command to run for testing (overrides build-system preset)
  -typecheck string
        Command to run for type checking (overrides build-system preset)
  -v, -verbose
        Enable verbose output
  -version
        Show version information and exit

Memory Options:
  -show-memory
        Display stored memories
  -clear-memory
        Clear all stored memories
  -add-memory string
        Add a memory entry (format: type:content)
  -memory-file string
        Path to memory file (default ".ralph-memory.json")
  -memory-retention int
        Days to retain memories (default: 90)

Milestone Options:
  -milestones
        List all milestones with progress
  -milestone string
        Show features for a specific milestone

Nudge Options:
  -nudge string
        Add one-time nudge (format: type:content)
  -show-nudges
        Display current nudges
  -clear-nudges
        Clear all nudges
  -nudge-file string
        Path to nudge file (default "nudges.json")

Scope Control Options:
  -scope-limit int
        Max iterations per feature (0 = unlimited)
  -deadline string
        Time limit for the run (e.g., "1h", "30m", "2h30m")
  -list-deferred
        List features that have been deferred

Replanning Options:
  -auto-replan
        Enable automatic replanning when triggers fire
  -replan
        Manually trigger replanning
  -replan-strategy string
        Replanning strategy: incremental, agent, none (default "incremental")
  -replan-threshold int
        Consecutive failures before replanning (default: 3)
  -list-versions
        List plan backup versions
  -restore-version int
        Restore a specific plan version

Validation Options:
  -validate
        Run validations for all completed features
  -validate-feature int
        Validate a specific feature by ID

Goal Options:
  -goal string
        Add a high-level goal to decompose into plan items
  -goal-priority int
        Priority for the goal (higher = more important, default: 5)
  -goal-status
        Show progress toward all goals
  -list-goals
        List all goals
  -decompose-goal string
        Decompose a specific goal by ID into plan items
  -decompose-all
        Decompose all pending goals into plan items
  -goals-file string
        Path to goals file (default "goals.json")

Multi-Agent Options:
  -multi-agent
        Enable multi-agent collaboration mode
  -agents string
        Path to multi-agent configuration file (default "agents.json")
  -parallel-agents int
        Maximum number of agents to run in parallel (default: 2)
  -list-agents
        List configured agents
```

### Output Options

Control how Ralph displays information:

```bash
# Minimal output (errors only)
ralph -iterations 5 -quiet

# Machine-readable JSON output
ralph -iterations 5 -json-output

# Disable colored output
ralph -iterations 5 -no-color

# Set log verbosity level
ralph -iterations 5 -log-level debug   # Show all messages
ralph -iterations 5 -log-level warn    # Only warnings and errors
ralph -iterations 5 -log-level error   # Only errors
```

### Build System Support

Ralph supports multiple build systems with auto-detection and presets:

**Supported Build Systems:**
- **pnpm** - `pnpm typecheck` / `pnpm test` (default)
- **npm** - `npm run typecheck` / `npm test`
- **yarn** - `yarn typecheck` / `yarn test`
- **gradle** - `./gradlew check` / `./gradlew test`
- **maven** - `mvn compile` / `mvn test`
- **cargo** - `cargo check` / `cargo test`
- **go** - `go build ./...` / `go test ./...`
- **python** - `mypy .` / `pytest`

**Auto-Detection:**
Ralph automatically detects the build system by checking for common project files:
- `build.gradle` or `gradlew` → Gradle
- `pom.xml` → Maven
- `Cargo.toml` → Cargo (Rust)
- `go.mod` → Go
- `setup.py`, `pyproject.toml`, or `requirements.txt` → Python
- `pnpm-lock.yaml` → pnpm
- `yarn.lock` → Yarn
- `package.json` → npm

**Usage Examples:**
```bash
# Auto-detect build system (default behavior)
ralph -iterations 5

# Explicitly specify build system
ralph -iterations 5 -build-system gradle

# Use auto-detection explicitly
ralph -iterations 5 -build-system auto

# Override individual commands
ralph -iterations 5 -build-system gradle -test "./gradlew test --tests MyTest"
```

## Configuration File

Ralph supports persistent configuration through YAML or JSON configuration files. This allows you to set default options for your project without specifying them on every command.

### Supported File Names

Ralph automatically discovers configuration files in the following order:

1. **Current directory** (first found wins):
   - `.ralph.yaml`
   - `.ralph.yml`
   - `.ralph.json`
   - `ralph.config.yaml`
   - `ralph.config.yml`
   - `ralph.config.json`

2. **Home directory** (same file names as above)

### Configuration Options

All configuration options are optional. Only specify the settings you want to customize.

**YAML Format (`.ralph.yaml`):**
```yaml
# AI agent command
agent: cursor-agent

# Build system preset (pnpm, npm, yarn, gradle, maven, cargo, go, python, auto)
build_system: go

# Custom commands (override build system preset)
typecheck: go build ./...
test: go test -v ./...

# File paths
plan: plan.json
progress: progress.txt

# Execution settings
iterations: 5
verbose: true
```

**JSON Format (`.ralph.json`):**
```json
{
  "agent": "cursor-agent",
  "build_system": "go",
  "typecheck": "go build ./...",
  "test": "go test -v ./...",
  "plan": "plan.json",
  "progress": "progress.txt",
  "iterations": 5,
  "verbose": true
}
```

### Configuration Precedence

Configuration values are applied in the following order (later values override earlier ones):

1. **Defaults** - Built-in default values
2. **Config file** - Values from auto-discovered or specified config file
3. **CLI flags** - Command-line arguments always take highest precedence

This means you can set project defaults in a config file and override specific values on the command line when needed.

### Examples

**Basic project configuration:**
```yaml
# .ralph.yaml
build_system: go
iterations: 3
```

**Full configuration for a Node.js project:**
```yaml
# .ralph.yaml
agent: cursor-agent
build_system: pnpm
plan: tasks/plan.json
progress: tasks/progress.txt
verbose: true
```

**Configuration with recovery settings:**
```yaml
# .ralph.yaml
build_system: go
iterations: 10
max_retries: 5
recovery_strategy: retry
```

**Configuration with environment override:**
```yaml
# .ralph.yaml
build_system: go
iterations: 5
environment: ci  # Force CI behavior for consistent builds
verbose: true
```

**Using a custom config file:**
```bash
# Use a specific config file
ralph -config production.yaml -iterations 10

# Override config file settings
ralph -iterations 1 -verbose  # Uses config file but overrides these values
```

## Failure Recovery

Ralph includes an intelligent failure recovery system that automatically detects and handles failures during iterative development. This helps prevent the workflow from getting stuck when issues occur.

### Failure Types

Ralph detects four types of failures:

- **Test Failures** (`test_failure`): Test assertions fail, test suites don't pass
- **Type Check Failures** (`typecheck_failure`): Compilation errors, type mismatches, syntax errors
- **Timeout** (`timeout`): Operations exceed time limits
- **Agent Errors** (`agent_error`): General agent execution failures

### Recovery Strategies

Three recovery strategies are available:

| Strategy | Description | Best For |
|----------|-------------|----------|
| `retry` (default) | Retry with enhanced prompt guidance | Transient failures, test issues |
| `skip` | Skip the feature, move to next | Blocking issues, complex problems |
| `rollback` | Revert changes via git, then retry | Corrupted state, bad code changes |

### Configuration

**Command-line flags:**
```bash
# Set maximum retries before escalation (default: 3)
ralph -iterations 10 -max-retries 5

# Set recovery strategy (default: retry)
ralph -iterations 10 -recovery-strategy skip

# Combine options
ralph -iterations 10 -max-retries 2 -recovery-strategy rollback
```

**Configuration file:**
```yaml
# .ralph.yaml
max_retries: 5
recovery_strategy: retry
```

### How It Works

1. **Failure Detection**: After each agent execution, Ralph analyzes the output and exit code to detect failures
2. **Recording**: Failures are tracked per feature with retry counts
3. **Strategy Application**: The configured strategy is applied to recover
4. **Escalation**: If max retries are exceeded, Ralph automatically escalates to skip the feature
5. **Summary**: At the end, a failure summary shows what happened

### Example Output

```
=== Iteration 3/10 ===
Executing agent command...

⚠ Failure detected: [test_failure] --- FAIL: TestSomething (feature #5, iteration 3, retries: 1)
→ Recovery: Retrying feature #5 (attempt 2/3)

=== Iteration 4/10 ===
...

=== Recovery Summary ===
Failure Summary:
  Feature #5: 2 failure(s)
    - test_failure: 2
Total failures: 2
```

### Retry Strategy Details

When using the `retry` strategy, Ralph generates enhanced prompts based on the failure type:

- **Test failures**: Emphasizes fixing tests first, ensuring assertions pass
- **Type check failures**: Focuses on compilation issues, imports, and type errors
- **Timeouts**: Suggests simplification and breaking down into smaller steps
- **Agent errors**: Provides general guidance to review and address the root cause

### Rollback Strategy Details

The `rollback` strategy uses git to revert changes:

1. Checks if in a git repository
2. Verifies uncommitted changes exist
3. Runs `git reset HEAD --` and `git checkout -- .`
4. Returns to a clean state for retry

**Note**: Rollback only reverts tracked file changes. Untracked files are preserved for safety.

## Environment Detection

Ralph automatically detects the execution environment and adapts its behavior accordingly. This ensures optimal performance in both local development and CI/CD pipelines.

### Supported Environments

| Environment | Detection | Adjustments |
|-------------|-----------|-------------|
| Local | Default (no CI vars) | Shorter timeouts, interactive output |
| GitHub Actions | `GITHUB_ACTIONS` env var | Longer timeouts, verbose output |
| GitLab CI | `GITLAB_CI` env var | Longer timeouts, verbose output |
| Jenkins | `JENKINS_URL` env var | Longer timeouts, verbose output |
| CircleCI | `CIRCLECI` env var | Longer timeouts, verbose output |
| Travis CI | `TRAVIS` env var | Longer timeouts, verbose output |
| Azure DevOps | `TF_BUILD` env var | Longer timeouts, verbose output |
| Generic CI | `CI` env var | Longer timeouts, verbose output |

### Automatic Adaptations

When running in a CI environment, Ralph automatically:

1. **Enables verbose output** - CI logs benefit from detailed information
2. **Increases timeouts** - CI builds may run slower than local machines
3. **Adjusts parallel hints** - Based on detected CPU cores

### System Resources Detection

Ralph detects:

- **CPU cores** - Used to calculate parallel execution hints
- **Available memory** - Detected from `/proc/meminfo` (Linux) or `sysctl` (macOS)
- **Project complexity** - Based on file count (small: <100, medium: 100-1000, large: >1000 files)

### Configuration

**Override detected environment:**
```bash
# Force CI behavior locally
ralph -iterations 5 -environment github-actions

# Force local behavior in CI
ralph -iterations 5 -environment local
```

**Configuration file:**
```yaml
# .ralph.yaml
environment: github-actions  # Force specific environment
```

**Supported environment values:**
- `local` - Local development (default)
- `github-actions` (or `github`, `gh`) - GitHub Actions
- `gitlab-ci` (or `gitlab`, `gl`) - GitLab CI
- `jenkins` - Jenkins
- `circleci` (or `circle`) - CircleCI
- `travis-ci` (or `travis`) - Travis CI
- `azure-devops` (or `azure`) - Azure DevOps
- `ci` - Generic CI environment

### Verbose Output

With `-verbose` flag (or in CI), Ralph displays environment information:

```
Environment: Local development
  CPU cores: 8
  Memory: 16.0 GB
  Project complexity: medium (234 files)
  Recommended timeout: 1m0s
  Parallel hint: 7 workers
```

## Enhanced CLI Output

Ralph provides rich CLI output with colors, progress indicators, and structured logging for better visibility into the development workflow.

### Features

- **Colored Output**: Success (green), errors (red), warnings (yellow), info (blue)
- **Progress Spinner**: Visual feedback during long-running agent executions
- **Summary Dashboard**: End-of-run summary showing completed features, failures, and timing
- **Structured Logging**: Log levels for controlling output verbosity
- **JSON Output**: Machine-readable output for automation and CI integration

### Output Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| Normal | Colored output with symbols | Interactive terminal use |
| Quiet (`-quiet`) | Errors only | Scripting, background jobs |
| JSON (`-json-output`) | Structured JSON lines | CI/CD pipelines, log aggregation |
| No Color (`-no-color`) | Plain text | Non-TTY environments, legacy terminals |

### Log Levels

Control output verbosity with `-log-level`:

| Level | Shows |
|-------|-------|
| `debug` | All messages including detailed debugging info |
| `info` (default) | Standard progress and informational messages |
| `warn` | Only warnings and errors |
| `error` | Only error messages |

### Summary Dashboard

At the end of each run, Ralph displays a summary:

```
=== Execution Summary ===
┌───────────────────────────────────────────┐
│ Progress:              8/10 iterations    │
│ Features completed:                   5   │
│ Features failed:                      1   │
│ Features skipped:                     2   │
│ Failures recovered:                   3   │
│ Duration:                         2m30s   │
└───────────────────────────────────────────┘
```

### Configuration

**Command-line flags:**
```bash
# Quiet mode for scripts
ralph -iterations 10 -quiet

# JSON output for CI pipelines
ralph -iterations 10 -json-output

# Disable colors for legacy terminals
ralph -iterations 10 -no-color

# Debug level logging
ralph -iterations 10 -log-level debug -verbose
```

**Configuration file:**
```yaml
# .ralph.yaml
no_color: false
quiet: false
json_output: false
log_level: info
```

### CI Compatibility

Ralph automatically detects non-TTY environments and adjusts output:
- Disables colors when output is not a terminal
- Disables spinners and progress bars in non-interactive mode
- Enables verbose output by default in CI environments

## Long-Running Goal Memory

Ralph includes a persistent memory system that remembers architectural decisions, coding conventions, tradeoffs, and project context across sessions. This reduces repetitive guidance and maintains consistency throughout your development workflow.

### Memory Types

| Type | Description | Examples |
|------|-------------|----------|
| `decision` | Architectural choices | "Use PostgreSQL for persistence", "Prefer composition over inheritance" |
| `convention` | Coding standards | "Use snake_case for database columns", "All errors must be wrapped" |
| `tradeoff` | Accepted compromises | "Sacrificed type safety for performance in hot path" |
| `context` | Project knowledge | "Main service is in cmd/server", "Config loaded from environment" |

### Memory File

Memories are stored in `.ralph-memory.json` (configurable with `-memory-file`). The file is automatically created when memories are first added.

**Example memory file:**
```json
{
  "entries": [
    {
      "id": "mem_1705420800123456789",
      "type": "decision",
      "content": "Use PostgreSQL for all persistence needs",
      "category": "infra",
      "created_at": "2026-01-16T12:00:00Z",
      "updated_at": "2026-01-16T12:00:00Z",
      "source": "agent"
    }
  ],
  "last_updated": "2026-01-16T12:00:00Z",
  "retention_days": 90
}
```

### AI Agent Memory Extraction

AI agents can create memories by including markers in their output:

```
[REMEMBER:DECISION]Use PostgreSQL for all persistence needs[/REMEMBER]
[REMEMBER:CONVENTION]Use snake_case for all database column names[/REMEMBER]
[REMEMBER:TRADEOFF]Opted for eventual consistency to improve performance[/REMEMBER]
[REMEMBER:CONTEXT]The main API is served from cmd/api/main.go[/REMEMBER]
```

Ralph automatically extracts these markers from agent output and stores them in the memory file.

### Memory Injection

During iterations, Ralph injects relevant memories into the agent prompt as context:

```
[MEMORY CONTEXT - Previous decisions and conventions to follow:]
- [DECISION] Use PostgreSQL for all persistence needs
- [CONVENTION] Use snake_case for all database column names
[END MEMORY CONTEXT]
```

Memories are ranked by relevance based on:
- **Type weight**: Decisions and conventions ranked higher than context
- **Category match**: Entries matching the current feature's category get priority
- **Recency**: Recently updated memories ranked higher

### Memory Commands

```bash
# Display all stored memories
ralph -show-memory

# Add a memory manually
ralph -add-memory "decision:Use PostgreSQL for persistence"
ralph -add-memory "convention:All exported functions must have comments"

# Clear all memories
ralph -clear-memory

# Use a custom memory file
ralph -iterations 5 -memory-file project-memory.json

# Set memory retention period (days)
ralph -iterations 5 -memory-retention 30
```

### Configuration

**Configuration file:**
```yaml
# .ralph.yaml
memory_file: .ralph-memory.json
memory_retention: 90  # Days to keep memories (default: 90)
```

### Memory Retention

Memories older than the retention period (default: 90 days) are automatically pruned at the start of each run. This ensures the memory file doesn't grow indefinitely while keeping recent, relevant information.

### Example Workflow

1. **First run** - Agent makes architectural decisions:
   ```
   Agent output: "Setting up the database layer.
   [REMEMBER:DECISION]Use PostgreSQL with pgx driver for all database operations[/REMEMBER]
   [REMEMBER:CONVENTION]All database queries go through repository pattern[/REMEMBER]"
   ```

2. **Subsequent runs** - Agent receives context:
   ```
   [MEMORY CONTEXT]
   - [DECISION] Use PostgreSQL with pgx driver for all database operations
   - [CONVENTION] All database queries go through repository pattern
   [END MEMORY CONTEXT]
   ```

3. **Result**: Agent maintains consistency without needing repeated instructions

## User Nudge Hooks

Ralph includes a nudge system for lightweight mid-run guidance. Nudges allow you to steer the AI agent during execution without stopping the workflow. You can create or edit nudges.json during a run, and Ralph will incorporate the nudges into subsequent iterations.

### Nudge Types

| Type | Description | Examples |
|------|-------------|----------|
| `focus` | Prioritize a specific feature or approach | "Work on feature 5 first", "Focus on error handling" |
| `skip` | Defer a feature or skip certain work | "Skip feature 3 for now", "Don't implement caching yet" |
| `constraint` | Add a requirement or limitation | "Don't use external libraries", "Must be backward compatible" |
| `style` | Specify coding style preferences | "Use functional style", "Prefer composition over inheritance" |

### Nudge File

Nudges are stored in `nudges.json` (configurable with `-nudge-file`). The file is automatically created when nudges are first added.

**Example nudges.json:**
```json
{
  "nudges": [
    {
      "id": "nudge_1705420800123456789",
      "type": "focus",
      "content": "Prioritize feature 5 - it's blocking other work",
      "priority": 10,
      "created_at": "2026-01-16T12:00:00Z",
      "acknowledged": false
    },
    {
      "id": "nudge_1705420800123456790",
      "type": "constraint",
      "content": "Don't add external dependencies without approval",
      "priority": 5,
      "created_at": "2026-01-16T12:00:00Z",
      "acknowledged": false
    }
  ],
  "last_updated": "2026-01-16T12:00:00Z"
}
```

### Nudge Priority

Nudges can have a priority (higher = more important). Nudges are sorted by priority when injected into the agent prompt. Default priority is 0.

### Nudge Commands

```bash
# Add a one-time nudge
ralph -nudge "focus:Work on feature 5 first"
ralph -nudge "skip:Skip feature 3 for now"
ralph -nudge "constraint:Don't use external libraries"
ralph -nudge "style:Use functional programming style"

# Display all nudges
ralph -show-nudges

# Clear all nudges
ralph -clear-nudges

# Use a custom nudge file
ralph -iterations 5 -nudge-file project-nudges.json
```

### Nudge Injection

During iterations, Ralph injects active nudges into the agent prompt as context:

```
[USER GUIDANCE - Please follow these instructions carefully:]
- [FOCUS (priority: 10)] Prioritize feature 5 - it's blocking other work
- [CONSTRAINT (priority: 5)] Don't add external dependencies without approval
- [STYLE] Use functional programming style
[END USER GUIDANCE]
```

### Nudge Acknowledgment

After a nudge is injected into an iteration, it is automatically marked as acknowledged. This prevents the same nudge from being repeated in subsequent iterations. Nudge acknowledgments are also logged to progress.txt for tracking.

### Mid-Run Guidance

The key feature of nudges is mid-run guidance. You can:

1. **Start an iteration run:**
   ```bash
   ralph -iterations 10 -verbose
   ```

2. **While running**, create/edit `nudges.json` to add guidance:
   ```bash
   # In another terminal
   ralph -nudge "focus:Stop working on feature 2, switch to feature 7"
   ```

3. **Ralph detects the change** and incorporates the nudge into the next iteration.

### Configuration

**Configuration file:**
```yaml
# .ralph.yaml
nudge_file: nudges.json
```

### Nudge vs Memory

| Aspect | Nudges | Memory |
|--------|--------|--------|
| Purpose | Real-time guidance | Long-term knowledge |
| Duration | Single iteration (acknowledged after use) | Persistent across sessions |
| Use case | Steering current work | Maintaining consistency |
| Creation | Manual (user adds) | Manual or automatic (agent extracts) |

Use nudges for immediate guidance, use memory for architectural decisions and conventions that should persist.

### Example Workflow

1. **Start iterations:**
   ```bash
   ralph -iterations 10 -verbose
   ```

2. **Notice the agent is working on the wrong feature.** Add a nudge:
   ```bash
   ralph -nudge "focus:Feature 7 is more urgent - please switch to that"
   ```

3. **Need to add a constraint.** Add another nudge:
   ```bash
   ralph -nudge "constraint:The API must remain backward compatible"
   ```

4. **Check current nudges:**
   ```bash
   ralph -show-nudges
   ```

5. **After run completes**, clear nudges if no longer needed:
   ```bash
   ralph -clear-nudges
   ```

## Smart Scope Control

Ralph includes smart scope control to prevent over-building and ensure timely completion. You can set iteration budgets per feature and time limits for entire runs. When limits are reached, features are automatically deferred rather than blocking progress.

### Scope Constraints

| Constraint | Description | Flag |
|------------|-------------|------|
| Iteration Limit | Max iterations allowed per feature | `-scope-limit` |
| Deadline | Time limit for the entire run | `-deadline` |

### Usage Examples

```bash
# Limit each feature to 3 iterations max
ralph -iterations 10 -scope-limit 3

# Set a 2 hour deadline for the entire run
ralph -iterations 10 -deadline 2h

# Combine both constraints
ralph -iterations 20 -scope-limit 5 -deadline 1h30m

# View deferred features
ralph -list-deferred
```

### How It Works

1. **Iteration Budget**: Each feature gets a budget of `-scope-limit` iterations. If the feature isn't complete after that many iterations, it's marked as deferred and Ralph moves to the next feature.

2. **Deadline**: When the deadline is reached, Ralph stops execution cleanly rather than abandoning mid-iteration.

3. **Feature Deferral**: Deferred features are marked in plan.json with `"deferred": true` and a `"defer_reason"` field. This allows you to:
   - See which features need more attention
   - Manually un-defer features when ready
   - Track which features consistently exceed budgets

4. **Complexity Estimation**: Ralph estimates feature complexity based on step count and description keywords, using this to suggest when simplification might help.

### Deferral Reasons

| Reason | Description |
|--------|-------------|
| `iteration_limit` | Feature exceeded its iteration budget |
| `deadline` | Deadline was reached during feature work |
| `complexity` | Feature deemed too complex for current scope |
| `manual` | Feature was manually deferred |

### Simplification Suggestions

When a feature reaches half its iteration budget or is detected as high complexity, Ralph suggests simplification strategies:

- Breaking large features into smaller pieces
- Implementing a minimal version first
- Focusing on core functionality and deferring edge cases

### Configuration

**Command-line flags:**
```bash
# Set iteration limit per feature (default: 0 = unlimited)
ralph -iterations 10 -scope-limit 3

# Set deadline (duration format: 1h, 30m, 2h30m)
ralph -iterations 10 -deadline 2h

# List features that were deferred
ralph -list-deferred
```

**Configuration file:**
```yaml
# .ralph.yaml
scope_limit: 5       # Max iterations per feature
deadline: "2h"       # Time limit for the run
```

### Deferred Features in plan.json

When a feature is deferred, it's updated in plan.json:

```json
{
  "id": 5,
  "description": "Complex feature",
  "tested": false,
  "deferred": true,
  "defer_reason": "iteration_limit"
}
```

### Listing Deferred Features

```bash
ralph -list-deferred

# Output:
# === Deferred Features (from plan.json) ===
#   5. chore [iteration_limit] - Complex refactoring task
#   8. feature [deadline] - Large feature implementation
#
# Total deferred: 2 features
```

### Example Workflow

1. **Start with scope limits:**
   ```bash
   ralph -iterations 20 -scope-limit 3 -deadline 1h
   ```

2. **Ralph works through features:**
   - Feature 1: Complete in 2 iterations ✓
   - Feature 2: Complete in 1 iteration ✓
   - Feature 3: Hit iteration limit after 3 iterations → Deferred
   - Feature 4: Complete in 2 iterations ✓
   - Feature 5: Deadline reached → Deferred

3. **Check what was deferred:**
   ```bash
   ralph -list-deferred
   ```

4. **Later, work on deferred features:**
   - Manually remove `deferred` and `defer_reason` from plan.json
   - Run again with a higher scope limit or more time

### Scope Status Output

During execution with scope control enabled, Ralph displays:

```
=== Scope Summary ===
Elapsed time: 45m30s
Time remaining: 14m30s
Deferred features: 2 (IDs: [3 5])

Deferred features will remain marked in plan.json.
Review and un-defer them manually when ready to continue.
```

### Best Practices

1. **Start conservative**: Begin with lower scope limits to identify problematic features quickly
2. **Review deferrals**: Regularly check `-list-deferred` to understand what needs attention
3. **Simplify complex features**: If a feature is repeatedly deferred, break it into smaller pieces
4. **Adjust over time**: Tune your scope limits based on your project's complexity

## Adaptive Plan Replanning

Ralph includes an adaptive replanning system that dynamically adjusts the plan when issues occur. This enables self-correction without full restart, reducing stuck states and improving autonomy.

### Replan Triggers

The replanning system monitors for conditions that indicate the current plan may need adjustment:

| Trigger | Condition | Description |
|---------|-----------|-------------|
| `test_failure` | Consecutive failures >= threshold | Repeated test failures suggest the feature may need to be broken down |
| `requirement_change` | plan.json externally modified | External changes to the plan file are detected and reconciled |
| `blocked_feature` | Feature becomes blocked | When a feature is deferred or blocked, the plan may need adjustment |
| `manual` | User triggers explicitly | Manual replanning via `-replan` flag |

### Replan Strategies

| Strategy | Description | Best For |
|----------|-------------|----------|
| `incremental` (default) | Adjust remaining features based on current state | Most situations, quick adjustments |
| `agent` | Use AI agent to generate a new plan | Complex situations, major restructuring |
| `none` | Disable replanning | When you want full manual control |

### Plan Versioning

Before any replanning occurs, Ralph creates a backup of the current plan:

- Backups are stored as `plan.json.bak.N` (where N is the version number)
- Each backup includes timestamp and trigger information
- You can restore any version if needed

### Usage Examples

```bash
# Enable automatic replanning during iterations
ralph -iterations 10 -auto-replan

# Set replan threshold (consecutive failures before replanning)
ralph -iterations 10 -auto-replan -replan-threshold 5

# Use agent-based replanning strategy
ralph -iterations 10 -auto-replan -replan-strategy agent

# Manually trigger replanning
ralph -replan

# Manual replan with agent strategy
ralph -replan -replan-strategy agent

# List plan backup versions
ralph -list-versions

# Restore a specific plan version
ralph -restore-version 2
```

### Configuration

**Command-line flags:**
```bash
# Enable automatic replanning
ralph -iterations 10 -auto-replan

# Set consecutive failure threshold (default: 3)
ralph -iterations 10 -auto-replan -replan-threshold 5

# Set replanning strategy
ralph -iterations 10 -replan-strategy incremental
ralph -iterations 10 -replan-strategy agent
ralph -iterations 10 -replan-strategy none

# Manual replanning
ralph -replan

# Version management
ralph -list-versions
ralph -restore-version 2
```

**Configuration file:**
```yaml
# .ralph.yaml
auto_replan: true              # Enable automatic replanning
replan_strategy: incremental   # Strategy: incremental, agent, none
replan_threshold: 3            # Consecutive failures before replanning
```

### How It Works

1. **Trigger Detection**: During iterations, Ralph monitors for replan triggers
2. **Backup Creation**: Before any changes, the current plan is backed up
3. **Strategy Execution**: The configured strategy analyzes the situation and proposes changes
4. **Diff Display**: Changes are shown before being applied
5. **Plan Update**: If successful, the plan file is updated with the new plan
6. **State Reset**: Failure counters are reset after replanning

### Incremental Strategy Details

The incremental strategy makes intelligent adjustments based on the trigger:

**For test failures:**
- Marks complex features for review
- Suggests breaking large features into smaller steps
- Identifies potential prerequisite dependencies

**For blocked features:**
- Marks blocked features as deferred
- Identifies the next viable feature to work on
- Reorders remaining work if needed

**For requirement changes:**
- Validates the updated plan for consistency
- Reconciles changes with execution state
- Reports on plan status (tested/untested/deferred counts)

### Agent-Based Strategy Details

The agent strategy sends the current state to the AI agent for analysis:

- Provides full context: current plan, failures, blocked features
- Agent can suggest more extensive restructuring
- Useful when incremental adjustments aren't sufficient

### Version Management

```bash
# List all backup versions
ralph -list-versions

# Output:
# === Plan Versions (from plan.json) ===
#   Version 1: 2026-01-16T12:00:00Z (trigger: test_failure)
#     Path: plan.json.bak.1.json
#   Version 2: 2026-01-16T13:30:00Z (trigger: manual)
#     Path: plan.json.bak.2.json
#
# Total: 2 version(s)
#
# To restore a version:
#   ralph -restore-version <number>

# Restore a specific version
ralph -restore-version 1
# Output: Restored plan version 1
```

### Example Workflow

1. **Start with auto-replan enabled:**
   ```bash
   ralph -iterations 20 -auto-replan -replan-threshold 3
   ```

2. **During execution**, if tests fail 3 times consecutively:
   ```
   === Automatic Replanning Triggered ===
   Trigger: test_failure
   
   Replanning completed: Feature #5 marked for review; Consider breaking into smaller steps
   Backup created: plan.json.bak.1.json
   
   Plan Changes:
     ~ Modified: 1 change(s)
       - #5.description: Complex feature -> Complex feature [REQUIRES REVIEW]
   ```

3. **If the plan is externally modified** while running:
   ```
   === Automatic Replanning Triggered ===
   Trigger: requirement_change
   
   Replanning completed: Plan reconciled: 3 tested, 5 untested, 2 deferred
   ```

4. **After execution**, review versions if needed:
   ```bash
   ralph -list-versions
   ```

5. **Restore an earlier version** if the replan didn't help:
   ```bash
   ralph -restore-version 1
   ```

### Replanning vs Recovery

| Aspect | Recovery | Replanning |
|--------|----------|------------|
| Scope | Single feature | Entire plan |
| Trigger | Any failure | Repeated/systemic failures |
| Action | Retry/skip/rollback feature | Restructure plan |
| Persistence | No plan changes | Updates plan.json |
| Versioning | None | Creates backups |

Use recovery for individual failures, use replanning when the plan itself needs adjustment.

### Best Practices

1. **Start with incremental**: The incremental strategy is safer and faster for most situations
2. **Set appropriate threshold**: Don't set the threshold too low (constant replanning) or too high (stuck states)
3. **Review version history**: Use `-list-versions` to understand what changes were made
4. **Combine with scope control**: Use `-scope-limit` alongside `-auto-replan` for best results
5. **Manual replan for major changes**: Use `-replan -replan-strategy agent` when you need significant restructuring

## Milestone-Based Progress Tracking

Ralph supports milestone-based progress tracking to help you organize features into meaningful project milestones like "Alpha", "Beta", or "MVP". This provides a higher-level view of progress beyond individual features.

### Defining Milestones

There are two ways to define milestones:

**1. Add milestone field to features in plan.json:**

```json
[
  {
    "id": 1,
    "description": "User authentication",
    "milestone": "Alpha",
    "milestone_order": 1,
    "tested": true
  },
  {
    "id": 2,
    "description": "User registration",
    "milestone": "Alpha",
    "milestone_order": 2,
    "tested": false
  },
  {
    "id": 3,
    "description": "Password reset",
    "milestone": "Beta",
    "tested": false
  }
]
```

**2. Create a separate milestones file (plan-milestones.json):**

```json
[
  {
    "id": "alpha",
    "name": "Alpha",
    "description": "Core authentication features",
    "criteria": "All auth features working with tests",
    "order": 1,
    "features": [1, 2]
  },
  {
    "id": "beta",
    "name": "Beta",
    "description": "User management features",
    "order": 2,
    "features": [3, 4, 5]
  }
]
```

### Milestone Fields

| Field | Description |
|-------|-------------|
| `milestone` | Name of the milestone this feature belongs to (in plan.json) |
| `milestone_order` | Optional order within the milestone for prioritization |
| `id` | Unique identifier for the milestone (in milestones file) |
| `name` | Display name for the milestone |
| `description` | Description of what the milestone represents |
| `criteria` | Success criteria for completing the milestone |
| `order` | Display/priority order for the milestone |
| `features` | List of feature IDs belonging to this milestone |

### Viewing Milestone Progress

```bash
# List all milestones with progress
ralph -milestones

# Output:
# Milestone Progress:
#   ◐ Alpha: 1/2 (50%)
#   ○ Beta: 0/3 (0%)
#
# Overall: 0/2 milestones complete, 1/5 features (20%)
#
# Next milestone to complete: Alpha ([██████████░░░░░░░░░░] 50%)

# Show features for a specific milestone
ralph -milestone Alpha

# Output:
# === Milestone: Alpha ===
# Description: Core authentication features
# Success Criteria: All auth features working with tests
# Progress: [██████████████░░░░░░░░░░░░░░░░] 50%
# Status: in_progress (1/2 features complete)
#
# Features:
#   [x] 1. User authentication
#   [ ] 2. User registration
```

### Progress Indicators

| Symbol | Status |
|--------|--------|
| ○ | Not started (0% complete) |
| ◐ | In progress (1-99% complete) |
| ● | Complete (100%) |

### Milestone Completion Celebration

When all features in a milestone are marked as tested, Ralph displays a celebration message:

```
Congratulations! Milestone 'Alpha' is done!
```

During iterations, Ralph automatically detects newly completed milestones and celebrates them in real-time.

### Configuration

**Command-line flags:**
```bash
# List all milestones with progress
ralph -milestones

# Show features for a specific milestone
ralph -milestone Alpha
ralph -milestone "Beta Release"
```

### Integration with Iterations

During `ralph -iterations`, milestone progress is:

1. **Displayed at start** (verbose mode) - Shows current progress for all milestones
2. **Monitored during execution** - Detects and celebrates newly completed milestones
3. **Summarized at end** - Shows final milestone status and suggests next milestone to focus on

### Example Workflow

1. **Define milestones** in your plan.json:
   ```json
   [
     {"id": 1, "description": "Setup CI/CD", "milestone": "Infrastructure", "tested": true},
     {"id": 2, "description": "Add database", "milestone": "Infrastructure", "tested": false},
     {"id": 3, "description": "User auth", "milestone": "MVP", "tested": false},
     {"id": 4, "description": "User dashboard", "milestone": "MVP", "tested": false}
   ]
   ```

2. **Check progress**:
   ```bash
   ralph -milestones
   # ◐ Infrastructure: 1/2 (50%)
   # ○ MVP: 0/2 (0%)
   ```

3. **Run iterations** and watch milestones complete:
   ```bash
   ralph -iterations 5 -verbose
   # ... during execution ...
   # Congratulations! Milestone 'Infrastructure' is done!
   ```

4. **View final status**:
   ```bash
   ralph -milestone MVP
   ```

## Outcome-Focused Validation

Ralph supports outcome-focused validation that goes beyond tests and type checks. You can define validations to verify that API endpoints respond correctly, CLI commands work end-to-end, files exist with expected content, and outputs match expected patterns.

### Validation Types

| Type | Description | Example Use Case |
|------|-------------|------------------|
| `http_get` | Verify HTTP GET endpoint | Health checks, API responses |
| `http_post` | Verify HTTP POST endpoint | Form submissions, API writes |
| `cli_command` | Verify CLI command execution | Tool integration, scripts |
| `file_exists` | Verify file exists with content | Config files, generated outputs |
| `output_contains` | Verify output matches pattern | Log validation, response checking |

### Defining Validations

Add validations to features in your plan.json:

```json
[
  {
    "id": 1,
    "description": "Health check endpoint",
    "category": "infra",
    "tested": true,
    "validations": [
      {
        "type": "http_get",
        "url": "http://localhost:8080/health",
        "expected_status": 200,
        "expected_body": "\"status\":\\s*\"healthy\"",
        "description": "Health endpoint returns healthy"
      }
    ]
  },
  {
    "id": 2,
    "description": "User API endpoint",
    "tested": true,
    "validations": [
      {
        "type": "http_post",
        "url": "http://localhost:8080/api/users",
        "body": "{\"name\": \"test\"}",
        "headers": {"Content-Type": "application/json"},
        "expected_status": 201,
        "description": "Create user returns 201"
      },
      {
        "type": "http_get",
        "url": "http://localhost:8080/api/users/1",
        "expected_status": 200,
        "description": "Get user returns 200"
      }
    ]
  },
  {
    "id": 3,
    "description": "CLI tool works",
    "tested": true,
    "validations": [
      {
        "type": "cli_command",
        "command": "./mytool",
        "args": ["--version"],
        "expected_body": "v\\d+\\.\\d+\\.\\d+",
        "description": "Tool version command works"
      }
    ]
  },
  {
    "id": 4,
    "description": "Config file generation",
    "tested": true,
    "validations": [
      {
        "type": "file_exists",
        "path": "config/settings.yaml",
        "pattern": "database:",
        "description": "Config file contains database settings"
      }
    ]
  }
]
```

### Validation Definition Fields

**Common Fields:**
| Field | Description |
|-------|-------------|
| `type` | Validation type (required): `http_get`, `http_post`, `cli_command`, `file_exists`, `output_contains` |
| `description` | Human-readable description of what's being validated |
| `timeout` | Timeout duration (e.g., "30s", "1m") |
| `retries` | Number of retries on failure (default: 3) |

**HTTP Validation Fields:**
| Field | Description |
|-------|-------------|
| `url` | The URL to request (required for HTTP validations) |
| `method` | HTTP method (defaults based on type: GET or POST) |
| `body` | Request body for POST requests |
| `headers` | Map of HTTP headers to send |
| `expected_status` | Expected HTTP status code (default: 200) |
| `expected_body` | Regex pattern to match in response body |

**CLI Validation Fields:**
| Field | Description |
|-------|-------------|
| `command` | Command to execute (required for CLI validation) |
| `args` | Array of command arguments |
| `expected_body` | Regex pattern to match in stdout |
| `options.expected_exit_code` | Expected exit code (default: 0) |

**File Validation Fields:**
| Field | Description |
|-------|-------------|
| `path` | File path to check (required for file_exists) |
| `pattern` | Regex pattern to match in file content |
| `options.should_exist` | Whether file should exist (default: true) |
| `options.min_size` | Minimum file size in bytes |

**Output Validation Fields:**
| Field | Description |
|-------|-------------|
| `input` | The text to check against |
| `pattern` | Regex pattern to match (required for output_contains) |
| `options.inverse` | If true, pattern should NOT match |

### Running Validations

```bash
# Run validations for all completed features
ralph -validate

# Validate a specific feature by ID
ralph -validate-feature 5

# With verbose output
ralph -validate -verbose

# Output example:
# === Running Validations ===
# Features to validate: 3
#
# === Feature #1: Health check endpoint ===
# ✓ GET http://localhost:8080/health returned 200
#
# === Feature #2: User API endpoint ===
# ✓ POST http://localhost:8080/api/users returned 201
# ✓ GET http://localhost:8080/api/users/1 returned 200
#
# === Validation Summary ===
# Overall: PASSED
#   Total validations: 3
#   Passed: 3
#   Failed: 0
```

### Validation Behavior

1. **Retries**: Validations automatically retry on failure (default: 3 retries with exponential backoff)
2. **Timeout**: Each validation has a timeout (default: 30 seconds)
3. **Pattern Matching**: Uses Go's regular expressions for body/output matching
4. **Progress Tracking**: Validation results are logged to progress.txt

### Best Practices

1. **Start simple**: Begin with health checks and basic endpoint validation
2. **Use descriptive descriptions**: Makes output easier to understand
3. **Set appropriate timeouts**: Increase for slow endpoints or commands
4. **Test patterns**: Verify regex patterns work before adding to validations
5. **Validate completed features**: Validations run only on features with `tested: true`

### Example: Full-Stack App Validation

```json
{
  "id": 1,
  "description": "Full-stack app deployment",
  "tested": true,
  "validations": [
    {
      "type": "cli_command",
      "command": "docker",
      "args": ["compose", "ps"],
      "expected_body": "Up",
      "description": "Docker containers are running"
    },
    {
      "type": "http_get",
      "url": "http://localhost:3000",
      "expected_status": 200,
      "description": "Frontend is accessible"
    },
    {
      "type": "http_get",
      "url": "http://localhost:8080/api/health",
      "expected_status": 200,
      "expected_body": "healthy",
      "description": "Backend API is healthy"
    },
    {
      "type": "http_post",
      "url": "http://localhost:8080/api/login",
      "body": "{\"email\":\"test@example.com\",\"password\":\"test\"}",
      "headers": {"Content-Type": "application/json"},
      "expected_status": 200,
      "expected_body": "token",
      "description": "Authentication works"
    }
  ]
}
```

## Goal-Oriented Planning

Ralph supports goal-oriented project outcomes where you specify high-level goals like "add user authentication" and Ralph automatically decomposes them into actionable plan items using AI.

### Defining Goals

There are two ways to define goals:

**1. Via command line (single goal):**

```bash
# Add a goal and decompose it into plan items
ralph -goal "Add user authentication with OAuth"

# Set priority (higher = more important)
ralph -goal "Add payment processing" -goal-priority 10
```

**2. Via goals.json file:**

```json
{
  "goals": [
    {
      "id": "auth",
      "description": "Add user authentication with OAuth",
      "priority": 10,
      "category": "security",
      "success_criteria": [
        "Users can log in via Google",
        "Sessions persist across browser restarts",
        "Logout properly clears session"
      ],
      "tags": ["auth", "security", "oauth"]
    },
    {
      "id": "payments",
      "description": "Integrate Stripe payment processing",
      "priority": 8,
      "category": "feature",
      "success_criteria": [
        "Users can add payment method",
        "Subscriptions work correctly",
        "Invoices are generated"
      ],
      "dependencies": ["auth"]
    }
  ]
}
```

### Goal Fields

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the goal |
| `description` | High-level goal description (required) |
| `priority` | Priority for ordering (higher = more important, default: 5) |
| `category` | Category for grouping (auto-inferred from description if not set) |
| `success_criteria` | Array of success criteria (helps AI understand what done looks like) |
| `tags` | Tags for filtering and organization |
| `dependencies` | IDs of goals this depends on (goals must complete first) |
| `status` | Current status: pending, in_progress, complete, blocked |
| `generated_plan_ids` | IDs of plan items generated from this goal |

### Goal Commands

```bash
# Add and decompose a single goal
ralph -goal "Add user authentication with OAuth"
ralph -goal "Add payment processing" -goal-priority 10

# List all goals
ralph -list-goals

# Show progress toward all goals
ralph -goal-status

# Decompose a specific goal into plan items
ralph -decompose-goal auth

# Decompose all pending goals
ralph -decompose-all

# Use a custom goals file
ralph -goal "New feature" -goals-file my-goals.json
```

### Goal Decomposition

When you add a goal with `-goal` or explicitly decompose with `-decompose-goal`, Ralph:

1. **Analyzes the goal** using the AI agent
2. **Generates plan items** with appropriate categories, steps, and expected outputs
3. **Identifies dependencies** between generated items
4. **Updates plan.json** with the new items
5. **Links items to the goal** for progress tracking

**Example decomposition:**

```bash
ralph -goal "Add user authentication with OAuth"

# Output:
# === Adding Goal ===
# Goal: Add user authentication with OAuth
# Priority: 5
# 
# ✓ Goal added with ID: goal_1705420800123456789
# 
# === Decomposing Goal into Plan Items ===
# 
# ✓ Generated 5 plan items
# 
# Generated plan items:
#   16. [infra] Set up OAuth provider configuration
#   17. [security] Implement OAuth callback handler
#   18. [db] Create users table and session storage
#   19. [feature] Add login/logout UI components
#   20. [chore] Write authentication tests
```

### Goal Progress Tracking

Track progress toward your goals:

```bash
ralph -goal-status

# Output:
# === Goal Progress ===
#   Add user authentication: [████████░░░░░░░░░░░░] 40%
#   Add payment processing: [pending] (no plan items)
#   
# Next goal to work on: Add user authentication (priority: 10)
```

Progress is calculated from:
- **Completed items**: Plan items linked to the goal with `tested: true`
- **Deferred items**: Items that were deferred due to scope constraints
- **Remaining items**: Items still to be completed

### Goal Dependencies

Goals can depend on other goals:

```json
{
  "goals": [
    {
      "id": "auth",
      "description": "User authentication",
      "priority": 10
    },
    {
      "id": "payments",
      "description": "Payment processing",
      "priority": 8,
      "dependencies": ["auth"]
    }
  ]
}
```

When a goal has dependencies:
- It's marked as **blocked** until dependencies complete
- `-goal-status` shows which goals are blocking
- `-list-goals` indicates blocked goals

### Goal Status

| Status | Description |
|--------|-------------|
| `pending` | Goal hasn't been started |
| `in_progress` | Work on the goal has started (has linked plan items) |
| `complete` | All generated plan items are complete |
| `blocked` | Waiting on dependent goals |

### Configuration

**Command-line flags:**
```bash
# Add a goal with decomposition
ralph -goal "Feature description"

# Set goal priority
ralph -goal "Feature" -goal-priority 10

# List all goals
ralph -list-goals

# Show goal progress
ralph -goal-status

# Decompose specific goal
ralph -decompose-goal <goal-id>

# Decompose all pending goals
ralph -decompose-all

# Use custom goals file
ralph -goals-file project-goals.json
```

**Configuration file:**
```yaml
# .ralph.yaml
goals_file: goals.json
```

### Category Inference

When adding a goal without an explicit category, Ralph infers it from the description:

| Keywords | Category |
|----------|----------|
| add, implement, create, build | feature |
| setup, configure, deploy, docker | infrastructure |
| database, db, migration, schema | database |
| ui, frontend, component, page | ui |
| api, endpoint, rest, graphql | api |
| security, auth, authentication | security |
| test, testing, coverage | testing |
| performance, optimize, speed | performance |
| document, docs, readme | documentation |
| refactor, clean, reorganize | refactor |

### Example Workflow

1. **Define high-level goals:**
   ```bash
   ralph -goal "Add user authentication with OAuth" -goal-priority 10
   ralph -goal "Implement payment processing with Stripe" -goal-priority 8
   ralph -goal "Create admin dashboard" -goal-priority 5
   ```

2. **Check goals:**
   ```bash
   ralph -list-goals
   ```

3. **Run iterations to work on generated plans:**
   ```bash
   ralph -iterations 10 -verbose
   ```

4. **Track progress:**
   ```bash
   ralph -goal-status
   ```

5. **When one goal is complete, continue with next:**
   ```bash
   ralph -iterations 10  # Continues with next priority goal
   ```

### Goals vs Plans

| Aspect | Goals | Plans |
|--------|-------|-------|
| Level | High-level outcomes | Specific, actionable tasks |
| Creation | Manual or via `-goal` flag | Manual or decomposed from goals |
| Tracking | Progress percentage | Individual completion (tested) |
| Dependencies | Goal-to-goal | Within decomposed items |
| Persistence | goals.json | plan.json |

Use goals for project outcomes you want to achieve; use plans for specific tasks to implement.

## Multi-Agent Collaboration

Ralph supports multi-agent collaboration for parallel AI coordination. This enables multiple AI agents with different roles (implementer, tester, reviewer, refactorer) to work together on features, improving quality and development speed.

### Agent Roles

| Role | Description | Purpose |
|------|-------------|---------|
| `implementer` | Creates code and implements features | Primary development work |
| `tester` | Validates code through tests | Test writing and validation |
| `reviewer` | Checks code quality | Code review and suggestions |
| `refactorer` | Improves existing code structure | Code cleanup and optimization |

### Configuring Agents

Create an `agents.json` file to configure multi-agent collaboration:

```json
{
  "agents": [
    {
      "id": "impl-1",
      "role": "implementer",
      "command": "cursor-agent",
      "specialization": "backend",
      "priority": 10,
      "enabled": true,
      "prompt_prefix": "You are a backend developer. Focus on server-side code."
    },
    {
      "id": "impl-2",
      "role": "implementer",
      "command": "cursor-agent",
      "specialization": "frontend",
      "priority": 10,
      "enabled": true,
      "prompt_prefix": "You are a frontend developer. Focus on UI components."
    },
    {
      "id": "test-1",
      "role": "tester",
      "command": "cursor-agent",
      "specialization": "testing",
      "priority": 8,
      "enabled": true,
      "prompt_prefix": "You are a QA engineer. Write comprehensive tests."
    },
    {
      "id": "review-1",
      "role": "reviewer",
      "command": "claude",
      "specialization": "code quality",
      "priority": 6,
      "enabled": true,
      "prompt_prefix": "You are a senior code reviewer. Check for best practices."
    }
  ],
  "max_parallel": 2,
  "conflict_resolution": "priority",
  "context_file": ".ralph-multiagent-context.json"
}
```

### Agent Configuration Fields

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the agent |
| `role` | Agent role: implementer, tester, reviewer, refactorer |
| `command` | CLI command to execute (e.g., "cursor-agent", "claude") |
| `specialization` | What this agent specializes in (e.g., "frontend", "backend") |
| `priority` | Execution priority (higher = earlier in workflow) |
| `enabled` | Whether this agent should be used |
| `timeout` | Maximum execution time (e.g., "5m", "10m") |
| `prompt_prefix` | Text prepended to all prompts for this agent |
| `prompt_suffix` | Text appended to all prompts for this agent |

### Multi-Agent Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `agents` | Array of agent configurations | Required |
| `max_parallel` | Max agents running simultaneously | 2 |
| `conflict_resolution` | How to resolve agent conflicts: priority, merge, vote | priority |
| `context_file` | Shared context file path | .ralph-multiagent-context.json |

### Conflict Resolution Strategies

| Strategy | Description | Best For |
|----------|-------------|----------|
| `priority` | Use highest priority agent's result | Clear hierarchy |
| `merge` | Combine non-conflicting suggestions | Collaborative work |
| `vote` | Majority wins for conflicting suggestions | Democratic decisions |

### Multi-Agent Commands

```bash
# List configured agents
ralph -list-agents

# Enable multi-agent mode during iterations
ralph -iterations 10 -multi-agent

# Specify custom agents config file
ralph -iterations 10 -multi-agent -agents custom-agents.json

# Set max parallel agents
ralph -iterations 10 -multi-agent -parallel-agents 4
```

### Multi-Agent Workflow

When multi-agent mode is enabled, Ralph executes a coordinated workflow:

1. **Implementation Stage**: Implementer agents create the code
2. **Testing Stage**: Tester agents validate with tests
3. **Review Stage**: Reviewer agents check code quality
4. **Refactoring Stage** (optional): If review finds issues, refactorer agents improve the code

Each stage runs agents in parallel (up to `max_parallel`), and results are aggregated before moving to the next stage.

### Shared Context

Agents communicate via a shared context file that tracks:
- Current feature being worked on
- Results from all agents
- Inter-agent messages
- Agreed-upon decisions

```json
{
  "feature_id": 1,
  "feature_description": "Add user authentication",
  "iteration": 3,
  "results": [
    {
      "agent_id": "impl-1",
      "role": "implementer",
      "status": "complete",
      "output": "Implementation complete...",
      "suggestions": ["Add error handling"],
      "issues": []
    }
  ],
  "messages": [],
  "decisions": [],
  "last_updated": "2026-01-16T12:00:00Z"
}
```

### Health Monitoring

Ralph monitors agent health during execution:
- Tracks agent status (idle, running, complete, failed, timeout)
- Detects stuck agents
- Provides health status via internal API

### Configuration

**Command-line flags:**
```bash
# List agents
ralph -list-agents

# Enable multi-agent mode
ralph -iterations 10 -multi-agent

# Custom agents file
ralph -iterations 10 -multi-agent -agents my-agents.json

# Set parallel limit
ralph -iterations 10 -multi-agent -parallel-agents 4
```

**Configuration file:**
```yaml
# .ralph.yaml
agents_file: agents.json
parallel_agents: 2
enable_multi_agent: true
```

### Example Workflow

1. **Create agents configuration:**
   ```bash
   # Create agents.json with your agent configurations
   ralph -list-agents  # View help if no file exists
   ```

2. **Run with multi-agent mode:**
   ```bash
   ralph -iterations 10 -multi-agent -verbose
   ```

3. **Monitor progress:**
   - Implementation stage completes first
   - Testers validate the implementation
   - Reviewers check code quality
   - Refactorers improve based on feedback (if needed)

4. **Review results:**
   - Check `.ralph-multiagent-context.json` for detailed results
   - Review suggestions and issues from all agents

### Multi-Agent vs Single Agent

| Aspect | Single Agent | Multi-Agent |
|--------|--------------|-------------|
| Execution | Sequential | Parallel stages |
| Perspectives | One viewpoint | Multiple specialized viewpoints |
| Quality | Agent-dependent | Cross-validated |
| Speed | Limited by one agent | Parallelized within stages |
| Complexity | Simple setup | Requires configuration |

Use multi-agent for complex projects where quality and multiple perspectives are important. Use single agent for simple tasks or when resources are limited.

## Plan File Format

Plans are JSON files containing an array of feature objects:

```json
[
  {
    "id": 1,
    "category": "chore",
    "description": "Initialize project structure",
    "steps": [
      "Create project root directory",
      "Initialize package.json",
      "Add README.md"
    ],
    "expected_output": "Project structure exists with basic files",
    "tested": false
  }
]
```

### Plan Fields

- `id` (number): Unique identifier for the feature
- `category` (string): Feature category (e.g., "chore", "infra", "db", "ui", "feature", "other")
- `description` (string): Clear, actionable description
- `steps` (array): Array of specific, implementable steps
- `expected_output` (string): What success looks like
- `tested` (boolean): Whether the feature has been tested (default: false)
- `milestone` (string): Optional milestone name this feature belongs to
- `milestone_order` (number): Optional order within the milestone for prioritization
- `deferred` (boolean): Whether the feature has been deferred due to scope constraints
- `defer_reason` (string): Reason for deferral (e.g., "iteration_limit", "deadline")
- `validations` (array): Optional outcome-focused validations (see Outcome-Focused Validation section)

## Workflow

1. **Plan Generation**: Create a plan.json from notes using `-generate-plan`
2. **Iteration**: Run `ralph -iterations N` to process features
3. **Agent Execution**: Ralph calls the AI agent with instructions to:
   - Find the highest-priority feature
   - Implement it
   - Run type checking and tests
   - Update the plan file
   - Append progress notes
   - Create a git commit
4. **Completion**: Ralph detects `<promise>COMPLETE</promise>` signal and exits

## AI Agent Integration

Ralph works with any CLI tool that accepts prompts. Supported agents:

- **Cursor Agent** (default): Uses `--print --force` flags
- **Claude**: Uses `--permission-mode acceptEdits -p` format

The agent receives a prompt that includes:
- References to the plan file and progress file
- Instructions for feature implementation
- Validation requirements
- Git commit instructions

## Examples

### Complete Workflow

```bash
# 1. Generate plan from notes
ralph -generate-plan -notes .hidden/notes.md

# 2. Review plan status
ralph -status

# 3. Run development iterations
ralph -iterations 10 -verbose

# 4. Check progress
ralph -status
```

### Using Makefile Commands

```bash
# Build
make build

# Install locally
make install-local

# Run with iterations
make run ITERATIONS=5

# Run with custom options
make run ITERATIONS=3 ARGS='-verbose -agent cursor-agent'

# View plan status
make jq-status

# List tested/untested
make jq-tested
make jq-untested
```

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
make lint
```

### Local Installation

After making changes, test locally:

```bash
make install-local
```

This builds and installs to `$GOPATH/bin` or `$GOBIN`, allowing you to test your changes immediately.

## Versioning

Ralph uses [Semantic Versioning](https://semver.org/) (SemVer) for version numbers in the format `MAJOR.MINOR.PATCH` (e.g., `v1.2.3`).

### Version Numbering

- **MAJOR** version: Incremented for incompatible API changes or breaking changes
  - Breaking changes to CLI flags or behavior
  - Removing deprecated features
  - Major architectural changes
  
- **MINOR** version: Incremented for new functionality in a backward-compatible manner
  - New features (e.g., new build system support)
  - New CLI flags or options
  - Enhancements that don't break existing workflows
  
- **PATCH** version: Incremented for backward-compatible bug fixes
  - Bug fixes
  - Security patches
  - Documentation improvements
  - Performance improvements

### Examples

- `v1.0.0` → `v1.0.1`: Bug fix (patch) - e.g., fixing a build system detection issue
- `v1.0.0` → `v1.1.0`: New feature (minor) - e.g., adding Gradle support
- `v1.0.0` → `v2.0.0`: Breaking change (major) - e.g., changing required flag names

### Checking Your Version

```bash
# Check installed version
ralph -version

# Output: ralph version v1.2.3
```

The version is embedded at build time and displayed in:
- `ralph -version` command output
- `ralph -help` usage message
- GitHub release binaries

**Version Detection:**
- **Local builds**: Version is automatically detected from git tags using `git describe --tags`
- **GitHub Actions**: Version is extracted from the git tag that triggers the release workflow
- **Development builds**: Shows "dev" if no git tags are found

When building locally with `make build`, ralph will automatically use the latest git tag version (e.g., `v0.0.1`). When GitHub Actions builds a release, it uses the semantic version tag (e.g., `v1.2.3`) that triggered the workflow.

## Releases

### Creating Releases

Ralph includes Makefile commands to create releases with pre-built binaries for GitHub:

```bash
# Create a patch release (v1.0.0 -> v1.0.1)
make release-patch

# Create a minor release (v1.0.0 -> v1.1.0)
make release-minor

# Create a major release (v1.0.0 -> v2.0.0)
make release-major
```

These commands will:
1. Calculate the new version based on the latest git tag using semantic versioning
2. Build binaries for multiple platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
3. Create checksums for all binaries
4. Output everything to the `dist/` directory

### Release Workflow

The release process follows semantic versioning and is automated via GitHub Actions:

1. **Determine version bump type:**
   - **Patch** (`release-patch`): Bug fixes, security patches, minor corrections
   - **Minor** (`release-minor`): New features, enhancements, backward-compatible changes
   - **Major** (`release-major`): Breaking changes, major refactoring, incompatible API changes

2. Run a release command to build binaries locally (optional, for testing):
   ```bash
   make release-patch  # or release-minor, release-major
   ```

3. Review the version in `.version` file and create/push the git tag:
   ```bash
   cat .version  # Verify version (e.g., v1.0.1)
   git tag v1.0.1
   git push origin v1.0.1
   ```

4. **GitHub Actions automatically:**
   - Detects the tag push
   - Builds binaries for all platforms (Linux, macOS, Windows)
   - Creates a GitHub release with the semantic version tag
   - Uploads all binaries and checksums

**Note**: If no git tags exist yet, the first release will start at:
- `v0.0.1` for patch releases
- `v0.1.0` for minor releases  
- `v1.0.0` for major releases

### Semantic Versioning with GitHub Actions

When you push a semantic version tag (e.g., `v1.2.3`), GitHub Actions automatically:

1. **Extracts the version** from the git tag (`v1.2.3`)
2. **Validates the format** to ensure it matches semantic versioning (`vMAJOR.MINOR.PATCH`)
3. **Builds binaries** for all platforms with the version embedded via ldflags
4. **Creates a GitHub release** with the correct version number
5. **Uploads binaries** that will show the correct version when users run `ralph -version`

**Important**: Always use semantic version tags (e.g., `v0.0.1`, `v1.2.3`) - the workflow validates this format and will fail if the tag doesn't match.

### Local Development Versioning

When building locally:
- Use `make build` to automatically detect version from git tags
- The version will be extracted from the latest git tag (e.g., `v0.0.1`)
- If no tags exist, it defaults to `dev`
- Uncommitted changes are handled gracefully (the `-dirty` suffix is stripped)

### Manual Release (Local Testing)

If you want to build release binaries locally for testing:

```bash
make release-patch
# Binaries will be in dist/ directory
```

The local build is useful for testing before pushing the tag, but the GitHub Action will rebuild everything when you push the tag.

## Requirements

- Go 1.21 or later
- An AI agent CLI tool (Cursor Agent, Claude, etc.)
- Git (for commit functionality)
- jq (optional, for Makefile plan status commands)

## Plan Analysis and Refinement

Ralph can analyze your plan.json to identify features that may benefit from refinement, helping you maintain well-structured and manageable tasks.

### Running Analysis

```bash
ralph -analyze-plan
```

### What It Detects

| Issue Type | Detection | Severity | Recommendation |
|------------|-----------|----------|----------------|
| **Compound Features** | Descriptions with "verb X and verb Y" pattern | Suggestion | Split into separate features |
| **Complex Features** | Features with >9 steps | Warning | Break into smaller features |

### Example Output

```
=== Plan Analysis Report ===

Total plans analyzed: 16
Issues found: 3
  - Compound features (with 'and'): 1
  - Complex features (>9 steps): 2

[SUGGESTION] Feature #3: compound
  Feature #3 description contains 'and', may represent multiple features
  Suggestions:
    Consider splitting into 2 separate features:
      1. Add caching layer
      2. Add rate limiting
    Each feature should have a single, focused objective

[WARNING] Feature #5: complex
  Feature #5 has 12 steps (>9), may be too complex
  Suggestions:
    Feature has 12 steps - consider splitting into smaller features
    Detected 4 potential logical groupings:
      Group 1 (3 steps): setup/config
      Group 2 (4 steps): implementation
      Group 3 (3 steps): testing
      Group 4 (2 steps): documentation
    Recommended: Split into 2 smaller features with 4-5 steps each
```

### Why Refine Plans?

Well-structured plans improve:
- **Code review efficiency**: Smaller, focused changes are easier to review
- **Testing reliability**: Isolated test cases for each feature
- **Progress tracking**: More granular milestones show clearer progress
- **Failure recovery**: Less work to redo if something fails

### Acceptable "And" Patterns

Not all "and" patterns indicate compound features. The analysis ignores common acceptable pairs:
- "read and write" (complementary operations)
- "authentication and authorization" (closely related security concepts)
- "YAML and JSON" (format variations)
- "search and filter" (single UI component)
- "load and save" (file operations)

The analysis is conservative, only flagging patterns like "Add X and add Y" where both parts clearly start with action verbs.

## FAQ

### General Questions

**Q: What AI agents does Ralph support?**

A: Ralph works with any CLI-based AI agent. Built-in support includes:
- **Cursor Agent** (default): Uses `--print --force` flags
- **Claude CLI**: Uses `--permission-mode acceptEdits -p` format

Any other agent can be used by specifying the `-agent` flag with the appropriate command.

**Q: Do I need to write a plan.json manually?**

A: No! You can generate a plan from notes:
```bash
ralph -generate-plan -notes my-notes.md
```
Ralph will use the AI agent to convert your notes into a structured plan.

**Q: How does Ralph know when a feature is complete?**

A: Features are marked complete when:
1. The AI agent updates the `tested: true` field in plan.json
2. Type checking passes
3. Tests pass

**Q: Can I use Ralph in CI/CD pipelines?**

A: Yes! Ralph automatically detects CI environments and adapts:
- Enables verbose output for better logging
- Increases timeouts for slower CI runners
- Supports JSON output (`-json-output`) for machine parsing
- Disables colors in non-TTY environments

### Configuration Questions

**Q: Where should I put my configuration file?**

A: Ralph looks for config files in this order:
1. Current directory (`.ralph.yaml`, `.ralph.json`, etc.)
2. Home directory (same file names)

Use `-config path/to/config.yaml` to specify a custom location.

**Q: How do CLI flags interact with config files?**

A: Precedence from lowest to highest:
1. Built-in defaults
2. Configuration file
3. CLI flags (always win)

**Q: Can I use environment variables?**

A: Ralph detects environment variables for CI detection (e.g., `GITHUB_ACTIONS`, `CI`), but doesn't currently read configuration from environment variables. Use config files or CLI flags instead.

### Troubleshooting Questions

**Q: Ralph is stuck on the same feature. What do I do?**

A: Try these approaches:
1. Enable verbose mode: `ralph -verbose -iterations 5`
2. Enable auto-replan: `ralph -auto-replan -iterations 10`
3. Use scope limits: `ralph -scope-limit 3 -iterations 10`
4. Check the feature's steps - they may be too complex

**Q: Tests pass locally but Ralph says they fail. Why?**

A: Common causes:
- Different working directory (Ralph runs from project root)
- Missing environment variables
- Race conditions in tests
- Test database not initialized

Run `ralph -verbose` to see the exact commands being executed.

**Q: How do I recover from a bad state?**

A: Options include:
1. Use rollback recovery: `ralph -recovery-strategy rollback`
2. Restore a plan version: `ralph -restore-version N`
3. Git reset manually and restart

### Feature Questions

**Q: What's the difference between memory and nudges?**

A: 
- **Memory**: Long-term storage for architectural decisions and conventions. Persists across sessions.
- **Nudges**: Short-term guidance for the current run. Acknowledged after use.

Use memory for "always do X", use nudges for "right now, focus on Y".

**Q: How do milestones differ from goals?**

A:
- **Milestones**: Groupings of existing features in plan.json. Track progress toward named checkpoints.
- **Goals**: High-level outcomes that get decomposed into plan items. Start from "what you want" and generate "how to get there".

**Q: Can I run multiple AI agents in parallel?**

A: Yes! Enable multi-agent mode:
```bash
ralph -multi-agent -agents agents.json -iterations 10
```

Configure different agents for implementation, testing, and review roles.

### Integration Questions

**Q: Does Ralph work with monorepos?**

A: Yes. Use a `.ralph.yaml` file with custom commands:
```yaml
typecheck: make typecheck-all  # Your monorepo command
test: make test-all
```

**Q: Can I integrate Ralph with GitHub Actions?**

A: Yes! Example workflow:
```yaml
- name: Run Ralph
  run: |
    go install github.com/start-it/ralph@latest
    ralph -iterations 5 -json-output
```

Ralph auto-detects GitHub Actions and adjusts behavior accordingly.

**Q: How do I validate that my API works, not just tests pass?**

A: Use outcome validations in plan.json:
```json
{
  "validations": [
    {
      "type": "http_get",
      "url": "http://localhost:8080/health",
      "expected_status": 200
    }
  ]
}
```

Run validations: `ralph -validate`

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - System design and package structure
- [Configuration](docs/CONFIGURATION.md) - Complete configuration reference
- [Features](docs/FEATURES.md) - Detailed feature documentation
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and solutions

## Examples

See the [examples/](examples/) directory for sample projects:
- [simple-api](examples/simple-api/) - Basic REST API in Go
- [fullstack-app](examples/fullstack-app/) - React + Go full-stack application
- [cli-tool](examples/cli-tool/) - CLI tool with Cobra

## License

[Add your license here]

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines, code style, and how to submit pull requests.

