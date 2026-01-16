# Ralph - AI-Assisted Development Workflow CLI

Ralph is a Golang CLI application that automates iterative development workflows by orchestrating AI-assisted development cycles. It processes plan files, executes development tasks through an AI agent CLI tool (like Cursor Agent or Claude), validates work, tracks progress, and commits changes iteratively until completion.

## Features

- **Iterative Development**: Process features one at a time based on priority
- **AI Agent Integration**: Works with Cursor Agent, Claude, or any compatible CLI tool
- **Plan Management**: Generate plans from notes, list tested/untested features
- **Progress Tracking**: Automatically updates plan files and progress logs
- **Milestone Tracking**: Organize features into milestones with progress visualization
- **Validation**: Integrates with type checking and testing commands
- **Git Integration**: Creates commits for completed features
- **Completion Detection**: Automatically detects when all work is complete
- **Failure Recovery**: Automatically handles failures with configurable recovery strategies
- **Environment Detection**: Automatically adapts to CI and local environments
- **Long-Running Memory**: Remembers architectural decisions and conventions across sessions
- **Nudge System**: Lightweight mid-run guidance without stopping execution
- **Smart Scope Control**: Iteration budgets and time limits to prevent over-building

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

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]

