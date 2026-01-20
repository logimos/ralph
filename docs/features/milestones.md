# Milestone Tracking

Organize features into project milestones for high-level progress tracking.

## Overview

Milestones group features into meaningful checkpoints like "Alpha", "Beta", or "MVP", providing a high-level view of progress.

## Defining Milestones

### In Plan.json

Add `milestone` field to features:

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

### Separate Milestones File

Create `plan-milestones.json`:

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

## Milestone Fields

| Field | Description |
|-------|-------------|
| `milestone` | Name in plan.json features |
| `milestone_order` | Order within the milestone |
| `id` | Unique identifier (in milestones file) |
| `name` | Display name |
| `description` | What the milestone represents |
| `criteria` | Success criteria |
| `order` | Display/priority order |
| `features` | List of feature IDs |

## Viewing Progress

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

## Progress Indicators

| Symbol | Status |
|--------|--------|
| ○ | Not started (0%) |
| ◐ | In progress (1-99%) |
| ● | Complete (100%) |

## Completion Celebrations

When all features in a milestone are completed:

```
🎉 Congratulations! Milestone 'Alpha' is done!
```

## Integration with Iterations

During `ralph -iterations`:

1. **At start** (verbose mode): Shows current milestone progress
2. **During execution**: Detects and celebrates completed milestones
3. **At end**: Shows final status and suggests next milestone

## Example Workflow

1. **Define milestones** in plan.json:
   ```json
   [
     {"id": 1, "description": "Setup CI/CD", "milestone": "Infrastructure", "tested": true},
     {"id": 2, "description": "Add database", "milestone": "Infrastructure", "tested": false},
     {"id": 3, "description": "User auth", "milestone": "MVP", "tested": false},
     {"id": 4, "description": "User dashboard", "milestone": "MVP", "tested": false}
   ]
   ```

2. **Check progress:**
   ```bash
   ralph -milestones
   # ◐ Infrastructure: 1/2 (50%)
   # ○ MVP: 0/2 (0%)
   ```

3. **Run iterations** and watch milestones complete:
   ```bash
   ralph -iterations 5 -verbose
   # ... during execution ...
   # 🎉 Congratulations! Milestone 'Infrastructure' is done!
   ```

4. **View final status:**
   ```bash
   ralph -milestone MVP
   ```

## Best Practices

1. **Meaningful names**: Use names that communicate project status
2. **Reasonable size**: 3-10 features per milestone
3. **Clear criteria**: Define what "done" means for each milestone
4. **Order features**: Use `milestone_order` to prioritize
5. **Track progress**: Check `-milestones` regularly
