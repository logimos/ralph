# Failure Recovery

Ralph uses a sophisticated two-tier system to handle failures gracefully.

## Two-Tier System Overview

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
    └───┬───┘             └─────┬─────┘           └───┬───┘
        │                       │                     │
        └───────────────────────┼─────────────────────┘
                                │
                    retries < -max-retries?
                        │           │
                       YES          NO
                        │           │
                    Try again   Skip feature
                                    │
    ════════════════════════════════╪════════════════════════════
    TIER 2: REPLANNING              │ (Plan-Level)
    Flags: -auto-replan, -replan-threshold, -replan-strategy
    ════════════════════════════════╪════════════════════════════
                                    │
                    consecutive failures >= threshold?
                        │           │
                       NO          YES
                        │           │
                    Continue    Trigger Replan
```

## Tier 1: Recovery (Per-Feature)

### Failure Types

| Type | Detection | Example |
|------|-----------|---------|
| `test_failure` | FAIL patterns | Test assertions fail |
| `typecheck_failure` | Syntax, undefined errors | Compilation errors |
| `timeout` | Execution exceeds limit | Long-running operations |
| `agent_error` | Non-zero exit | Agent crashes |

### Recovery Strategies

| Strategy | Action | Best For |
|----------|--------|----------|
| `retry` | Retry with enhanced prompt | Transient issues |
| `skip` | Skip feature, move on | Blocking problems |
| `rollback` | Git reset, then retry | Corrupted state |

### Configuration

```bash
# Set max retries (default: 3)
ralph -iterations 10 -max-retries 5

# Set recovery strategy (default: retry)
ralph -iterations 10 -recovery-strategy skip

# Combine options
ralph -iterations 10 -max-retries 2 -recovery-strategy rollback
```

In config file:

```yaml
# .ralph.yaml
max_retries: 5
recovery_strategy: retry
```

### Strategy Details

#### Retry Strategy

When using `retry`, Ralph generates enhanced prompts based on failure type:

- **Test failures**: Emphasizes fixing tests first
- **Type check failures**: Focuses on compilation issues
- **Timeouts**: Suggests simplification
- **Agent errors**: General guidance to address root cause

#### Rollback Strategy

The `rollback` strategy uses git:

1. Checks if in a git repository
2. Verifies uncommitted changes exist
3. Runs `git reset HEAD --` and `git checkout -- .`
4. Returns to clean state for retry

!!! warning
    Rollback only reverts tracked file changes. Untracked files are preserved.

## Tier 2: Replanning (Plan-Level)

When recovery alone isn't enough, replanning restructures the entire plan.

### Triggers

| Trigger | Condition |
|---------|-----------|
| `test_failure` | Consecutive failures >= threshold |
| `requirement_change` | plan.json modified externally |
| `blocked_feature` | Features become blocked |
| `manual` | User runs `-replan` |

### Strategies

| Strategy | Description |
|----------|-------------|
| `incremental` | Adjust based on current state |
| `agent` | AI agent restructures plan |
| `none` | Disable replanning |

### Configuration

```bash
# Enable auto-replanning
ralph -iterations 10 -auto-replan

# Set failure threshold
ralph -iterations 10 -auto-replan -replan-threshold 5

# Use agent-based strategy
ralph -iterations 10 -auto-replan -replan-strategy agent

# Manual replan
ralph -replan
```

In config file:

```yaml
# .ralph.yaml
auto_replan: true
replan_strategy: incremental
replan_threshold: 3
```

### Plan Versioning

Before replanning, Ralph creates backups:

```bash
# List backup versions
ralph -list-versions

# Restore a version
ralph -restore-version 2
```

## When to Use What

| Situation | Use Recovery | Use Replanning |
|-----------|--------------|----------------|
| Single feature failing | ✓ `-recovery-strategy retry` | |
| Transient errors | ✓ `-max-retries 5` | |
| Multiple features failing | | ✓ `-auto-replan` |
| Plan structure is wrong | | ✓ `-replan` |
| Need to backtrack | | ✓ `-replan-strategy agent` |
| Feature is blocked | ✓ `-recovery-strategy skip` | ✓ if many blocked |

## Example Workflow

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
  → Failure counter: 3 (equals threshold)
  → REPLANNING TRIGGERED

Iteration 7: Continue with restructured plan...
```

## Best Practices

1. **Start with retry**: The default strategy handles most cases
2. **Use appropriate thresholds**: 3 is usually good; adjust based on feature complexity
3. **Enable auto-replan for long runs**: Prevents getting stuck
4. **Review backups**: Use `-list-versions` to understand what changed
5. **Combine with scope control**: Use `-scope-limit` alongside recovery
