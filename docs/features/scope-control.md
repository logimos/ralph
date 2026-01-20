# Smart Scope Control

Prevent over-building with iteration budgets and time limits.

## Constraints

| Constraint | Flag | Description |
|------------|------|-------------|
| Iteration Limit | `-scope-limit` | Max iterations per feature |
| Deadline | `-deadline` | Total time limit for the run |

## Usage

```bash
# Limit each feature to 3 iterations max
ralph -iterations 10 -scope-limit 3

# Set a 2 hour deadline
ralph -iterations 10 -deadline 2h

# Combine both
ralph -iterations 20 -scope-limit 5 -deadline 1h30m

# View deferred features
ralph -list-deferred
```

## How It Works

### Iteration Budget

Each feature gets a budget of `-scope-limit` iterations:

1. Ralph tracks iterations per feature
2. If budget exceeded, feature is **deferred**
3. Ralph moves to the next feature
4. Deferred features stay in plan for later

### Deadline

When a deadline is set:

1. Ralph checks time before each iteration
2. If deadline reached, execution stops cleanly
3. Current feature may be marked deferred

### Feature Deferral

Deferred features are marked in `plan.json`:

```json
{
  "id": 5,
  "description": "Complex feature",
  "tested": false,
  "deferred": true,
  "defer_reason": "iteration_limit"
}
```

## Deferral Reasons

| Reason | Description |
|--------|-------------|
| `iteration_limit` | Feature exceeded its iteration budget |
| `deadline` | Deadline was reached |
| `complexity` | Feature deemed too complex |
| `manual` | Feature was manually deferred |

## Complexity Estimation

Ralph estimates feature complexity:

| Complexity | Steps | Suggested Iterations |
|------------|-------|---------------------|
| Low | 1-2 | 3 |
| Medium | 3-5 | 5 |
| High | 6+ | 10 |

Keywords that increase complexity:
- refactor, integration, security, migration, performance

## Simplification Suggestions

At 50% of iteration budget, Ralph suggests:

- Breaking large features into smaller pieces
- Implementing minimal version first
- Focusing on core functionality

## Configuration

### CLI Flags

```bash
# Set iteration limit (0 = unlimited)
ralph -iterations 10 -scope-limit 3

# Set deadline (duration format)
ralph -iterations 10 -deadline 2h

# List deferred features
ralph -list-deferred
```

### Config File

```yaml
# .ralph.yaml
scope_limit: 5       # Max iterations per feature
deadline: "2h"       # Time limit for the run
```

## Example Workflow

1. **Start with scope limits:**
   ```bash
   ralph -iterations 20 -scope-limit 3 -deadline 1h
   ```

2. **Ralph works through features:**
   - Feature 1: Complete in 2 iterations ✓
   - Feature 2: Complete in 1 iteration ✓
   - Feature 3: Hit iteration limit → **Deferred**
   - Feature 4: Complete in 2 iterations ✓
   - Feature 5: Deadline reached → **Deferred**

3. **Check what was deferred:**
   ```bash
   ralph -list-deferred
   
   # Output:
   # === Deferred Features ===
   #   3. [iteration_limit] Complex refactoring task
   #   5. [deadline] Large feature implementation
   #
   # Total deferred: 2 features
   ```

4. **Later, work on deferred features:**
   - Remove `deferred` and `defer_reason` from plan.json
   - Run again with higher limits or more time

## Scope Status Output

During execution:

```
=== Scope Summary ===
Elapsed time: 45m30s
Time remaining: 14m30s
Deferred features: 2 (IDs: [3 5])

Deferred features will remain marked in plan.json.
Review and un-defer them manually when ready to continue.
```

## Best Practices

1. **Start conservative**: Lower scope limits identify problematic features quickly
2. **Review deferrals**: Regularly check `-list-deferred`
3. **Simplify complex features**: If repeatedly deferred, break into smaller pieces
4. **Adjust over time**: Tune limits based on your project's complexity
5. **Combine with recovery**: Use alongside `-max-retries` for best results
