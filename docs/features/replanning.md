# Adaptive Replanning

Dynamic plan adjustment when issues occur repeatedly.

## Overview

Replanning is **Tier 2** of Ralph's failure handling system. It kicks in when recovery (Tier 1) alone isn't enough to resolve issues.

## Triggers

| Trigger | Condition |
|---------|-----------|
| `test_failure` | Consecutive failures >= threshold |
| `requirement_change` | plan.json modified externally |
| `blocked_feature` | Features become blocked |
| `manual` | User runs `-replan` |

## Strategies

| Strategy | Description | Best For |
|----------|-------------|----------|
| `incremental` | Adjust based on current state | Most situations |
| `agent` | AI agent restructures plan | Major restructuring |
| `none` | Disable replanning | Manual control |

## Usage

```bash
# Enable automatic replanning
ralph -iterations 10 -auto-replan

# Set failure threshold (default: 3)
ralph -iterations 10 -auto-replan -replan-threshold 5

# Use agent-based strategy
ralph -iterations 10 -auto-replan -replan-strategy agent

# Manual replan
ralph -replan

# Manual replan with agent
ralph -replan -replan-strategy agent
```

## Configuration

```yaml
# .ralph.yaml
auto_replan: true              # Enable automatic replanning
replan_strategy: incremental   # Strategy: incremental, agent, none
replan_threshold: 3            # Consecutive failures before replan
```

## Plan Versioning

Before any replanning, Ralph creates a backup:

```bash
# List all backup versions
ralph -list-versions

# Output:
# === Plan Versions ===
#   Version 1: 2026-01-16T12:00:00Z (trigger: test_failure)
#     Path: plan.json.bak.1.json
#   Version 2: 2026-01-16T13:30:00Z (trigger: manual)
#     Path: plan.json.bak.2.json

# Restore a specific version
ralph -restore-version 1
```

## How It Works

1. **Trigger Detection**: Ralph monitors for replan conditions
2. **Backup Creation**: Current plan is saved before changes
3. **Strategy Execution**: Selected strategy analyzes and proposes changes
4. **Diff Display**: Changes shown before being applied
5. **Plan Update**: Plan file updated with new plan
6. **State Reset**: Failure counters reset after replanning

## Strategy Details

### Incremental Strategy

Makes intelligent adjustments based on trigger:

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
- Reports on plan status

### Agent-Based Strategy

Sends full context to the AI agent:

- Current plan with all states
- Failures encountered
- Blocked/deferred features

The agent can suggest extensive restructuring.

## Example Workflow

1. **Start with auto-replan:**
   ```bash
   ralph -iterations 20 -auto-replan -replan-threshold 3
   ```

2. **During execution**, if tests fail 3 times:
   ```
   === Automatic Replanning Triggered ===
   Trigger: test_failure
   
   Replanning completed: Feature #5 marked for review
   Backup created: plan.json.bak.1.json
   
   Plan Changes:
     ~ Modified: 1 change(s)
       - #5.description: Complex feature -> Complex feature [REQUIRES REVIEW]
   ```

3. **If plan is externally modified:**
   ```
   === Automatic Replanning Triggered ===
   Trigger: requirement_change
   
   Replanning completed: Plan reconciled
   ```

4. **Review versions if needed:**
   ```bash
   ralph -list-versions
   ```

5. **Restore earlier version if needed:**
   ```bash
   ralph -restore-version 1
   ```

## Replanning vs Recovery

| Aspect | Recovery (Tier 1) | Replanning (Tier 2) |
|--------|-------------------|---------------------|
| Scope | Single feature | Entire plan |
| Trigger | Any failure | Repeated failures |
| Action | Retry/skip/rollback | Restructure plan |
| Persistence | No plan changes | Updates plan.json |
| Versioning | None | Creates backups |

## Best Practices

1. **Start with incremental**: Safer and faster for most situations
2. **Set appropriate threshold**: Not too low (constant replanning) or too high (stuck states)
3. **Review version history**: Use `-list-versions` to understand changes
4. **Combine with scope control**: Use `-scope-limit` alongside `-auto-replan`
5. **Manual replan for major changes**: Use `-replan -replan-strategy agent`
