# Configuration

Ralph supports three methods of configuration with the following precedence (highest to lowest):

1. **CLI Flags** - Command-line arguments
2. **Environment Variables** - CI detection
3. **Configuration File** - `.ralph.yaml` or `.ralph.json`

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

**Home Directory:** Same files as above (fallback)

### File Formats

=== "YAML"

    ```yaml
    # .ralph.yaml
    agent: cursor-agent
    build_system: go
    typecheck: go build ./...
    test: go test ./...
    
    plan: plan.json
    progress: progress.txt
    
    iterations: 5
    verbose: true
    
    # Recovery settings
    max_retries: 3
    recovery_strategy: retry
    
    # Scope control
    scope_limit: 0  # 0 = unlimited
    deadline: ""    # e.g., "2h", "30m"
    
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
    memory_file: .ralph-memory.json
    memory_retention: 90  # days
    ```

=== "JSON"

    ```json
    {
      "agent": "cursor-agent",
      "build_system": "go",
      "typecheck": "go build ./...",
      "test": "go test ./...",
      "plan": "plan.json",
      "progress": "progress.txt",
      "iterations": 5,
      "verbose": true,
      "max_retries": 3,
      "recovery_strategy": "retry"
    }
    ```

### Custom Config File

Use `-config` to specify a custom location:

```bash
ralph -config production.yaml -iterations 10
```

## Build System Support

Ralph auto-detects build systems based on project files:

| Build System | Detection | Type Check | Test |
|-------------|-----------|------------|------|
| Go | `go.mod` | `go build ./...` | `go test ./...` |
| pnpm | `pnpm-lock.yaml` | `pnpm typecheck` | `pnpm test` |
| npm | `package.json` | `npm run typecheck` | `npm test` |
| Yarn | `yarn.lock` | `yarn typecheck` | `yarn test` |
| Gradle | `build.gradle` | `./gradlew check` | `./gradlew test` |
| Maven | `pom.xml` | `mvn compile` | `mvn test` |
| Cargo | `Cargo.toml` | `cargo check` | `cargo test` |
| Python | `setup.py`, `pyproject.toml` | `mypy .` | `pytest` |

Override with `-build-system` or custom commands:

```bash
# Explicit build system
ralph -iterations 5 -build-system gradle

# Custom commands
ralph -iterations 5 -typecheck "make lint" -test "make test"
```

## Environment Detection

Ralph automatically detects CI environments and adapts:

| Environment | Detection | Adjustments |
|-------------|-----------|-------------|
| Local | Default | Shorter timeouts, interactive output |
| GitHub Actions | `GITHUB_ACTIONS` | Longer timeouts, verbose output |
| GitLab CI | `GITLAB_CI` | Longer timeouts, verbose output |
| Jenkins | `JENKINS_URL` | Longer timeouts, verbose output |
| CircleCI | `CIRCLECI` | Longer timeouts, verbose output |
| Generic CI | `CI` | Longer timeouts, verbose output |

Override with:

```bash
ralph -iterations 5 -environment github-actions
```

## Common Configurations

### Go Project

```yaml
# .ralph.yaml
build_system: go
iterations: 5
verbose: true
```

### Node.js with pnpm

```yaml
# .ralph.yaml
build_system: pnpm
plan: tasks/plan.json
verbose: true
```

### CI-Optimized

```yaml
# .ralph.yaml
build_system: go
verbose: true
max_retries: 5
auto_replan: true
json_output: true
```

### Full Configuration

```yaml
# .ralph.yaml
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
```

## Next Steps

- [CLI Reference](../reference/cli.md) - Complete CLI options
- [Full Configuration Reference](../reference/configuration.md) - All options
