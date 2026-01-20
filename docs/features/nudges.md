# User Nudge Hooks

Lightweight mid-run guidance without stopping execution.

## Overview

Nudges allow you to steer the AI agent during execution. You can create or edit nudges while Ralph runs, and they'll be incorporated into subsequent iterations.

## Nudge Types

| Type | Purpose | Examples |
|------|---------|----------|
| `focus` | Prioritize specific work | "Work on feature 5 first" |
| `skip` | Defer certain work | "Skip feature 3 for now" |
| `constraint` | Add requirements | "Don't use external libraries" |
| `style` | Coding preferences | "Use functional style" |

## Nudge File

Nudges are stored in `nudges.json`:

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
    }
  ],
  "last_updated": "2026-01-16T12:00:00Z"
}
```

## Commands

```bash
# Add a nudge
ralph -nudge "focus:Work on feature 5 first"
ralph -nudge "skip:Skip feature 3 for now"
ralph -nudge "constraint:Don't use external libraries"
ralph -nudge "style:Use functional programming style"

# Display all nudges
ralph -show-nudges

# Clear all nudges
ralph -clear-nudges

# Use custom nudge file
ralph -iterations 5 -nudge-file project-nudges.json
```

## Configuration

```yaml
# .ralph.yaml
nudge_file: nudges.json
```

## Nudge Priority

Nudges can have a priority (higher = more important). Default priority is 0.

```json
{
  "type": "focus",
  "content": "Critical bug fix needed",
  "priority": 10
}
```

## Nudge Injection

During iterations, active nudges are injected into prompts:

```
[USER GUIDANCE - Please follow these instructions carefully:]
- [FOCUS (priority: 10)] Prioritize feature 5 - it's blocking other work
- [CONSTRAINT (priority: 5)] Don't add external dependencies without approval
- [STYLE] Use functional programming style
[END USER GUIDANCE]
```

## Nudge Acknowledgment

After a nudge is used in an iteration, it's automatically marked as acknowledged. This prevents repetition in subsequent iterations.

## Mid-Run Guidance

The key feature of nudges is real-time steering:

1. **Start an iteration run:**
   ```bash
   ralph -iterations 10 -verbose
   ```

2. **While running**, add guidance in another terminal:
   ```bash
   ralph -nudge "focus:Stop working on feature 2, switch to feature 7"
   ```

3. **Ralph detects the change** and incorporates the nudge into the next iteration.

## Example Workflow

1. **Start iterations:**
   ```bash
   ralph -iterations 10 -verbose
   ```

2. **Notice agent is working on wrong feature:**
   ```bash
   ralph -nudge "focus:Feature 7 is more urgent - please switch to that"
   ```

3. **Need to add a constraint:**
   ```bash
   ralph -nudge "constraint:The API must remain backward compatible"
   ```

4. **Check current nudges:**
   ```bash
   ralph -show-nudges
   ```

5. **After run completes**, clear if no longer needed:
   ```bash
   ralph -clear-nudges
   ```

## Nudge vs Memory

| Aspect | Nudges | Memory |
|--------|--------|--------|
| Purpose | Real-time guidance | Long-term knowledge |
| Duration | Single iteration | Persistent across sessions |
| Use case | Steering current work | Maintaining consistency |
| Creation | Manual (user adds) | Manual or automatic |

Use nudges for immediate guidance, use memory for architectural decisions.

## Best Practices

1. **Be specific**: Clear nudges produce better results
2. **Use priority**: Higher priority for urgent guidance
3. **Check status**: Use `-show-nudges` to see what's active
4. **Clean up**: Clear nudges after they're no longer needed
5. **One at a time**: Too many nudges can confuse the agent
