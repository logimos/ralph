# Goal-Oriented Planning

Define high-level goals and let AI decompose them into actionable plans.

## Overview

Goals let you specify what you want to achieve (like "add user authentication") and Ralph uses AI to break that down into specific plan items.

## Defining Goals

### Via CLI

```bash
# Add and decompose a goal
ralph -goal "Add user authentication with OAuth"

# With priority (higher = more important)
ralph -goal "Add payment processing" -goal-priority 10
```

### Via goals.json

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
      "dependencies": ["auth"]
    }
  ]
}
```

## Goal Fields

| Field | Description |
|-------|-------------|
| `id` | Unique identifier |
| `description` | High-level goal description (required) |
| `priority` | Priority ordering (default: 5) |
| `category` | Category for grouping |
| `success_criteria` | Array of success criteria |
| `tags` | Tags for filtering |
| `dependencies` | IDs of goals this depends on |
| `status` | pending, in_progress, complete, blocked |
| `generated_plan_ids` | IDs of generated plan items |

## Commands

```bash
# Add and decompose a goal
ralph -goal "Add user authentication"
ralph -goal "Add payments" -goal-priority 10

# Show all goals with progress
ralph -goals

# Decompose a specific goal
ralph -decompose-goal auth

# Decompose all pending goals
ralph -decompose-all

# Use custom goals file
ralph -goals-file my-goals.json
```

## Goal Decomposition

When you add a goal with `-goal` or use `-decompose-goal`:

1. **Analysis**: AI agent analyzes the goal
2. **Generation**: Creates plan items with categories, steps, outputs
3. **Dependencies**: Identifies dependencies between items
4. **Update**: Updates plan.json with new items
5. **Linking**: Links items to the goal for tracking

### Example

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

## Goal Progress

Track progress toward goals:

```bash
ralph -goals

# Output:
# === Goal Progress ===
#   Add user authentication: [████████░░░░░░░░░░░░] 40%
#   Add payment processing: [pending] (no plan items)
#   
# Next goal to work on: Add user authentication (priority: 10)
```

Progress is calculated from:
- **Completed items**: Plan items with `tested: true`
- **Deferred items**: Items deferred due to scope constraints
- **Remaining items**: Items still to be completed

## Goal Dependencies

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
      "dependencies": ["auth"]
    }
  ]
}
```

When a goal has dependencies:
- It's marked as **blocked** until dependencies complete
- `-goals` shows which goals are blocking
- `GetNextGoalToWork()` skips blocked goals

## Goal Status

| Status | Description |
|--------|-------------|
| `pending` | Goal hasn't been started |
| `in_progress` | Work has started (has linked plan items) |
| `complete` | All generated plan items are complete |
| `blocked` | Waiting on dependent goals |

## Category Inference

If no category is specified, Ralph infers from description:

| Keywords | Category |
|----------|----------|
| add, implement, create | feature |
| setup, configure, deploy | infrastructure |
| database, db, migration | database |
| ui, frontend, component | ui |
| api, endpoint, rest | api |
| security, auth | security |
| test, testing | testing |
| document, docs | documentation |
| refactor, clean | refactor |

## Example Workflow

1. **Define high-level goals:**
   ```bash
   ralph -goal "Add user authentication with OAuth" -goal-priority 10
   ralph -goal "Implement payment processing" -goal-priority 8
   ralph -goal "Create admin dashboard" -goal-priority 5
   ```

2. **Check goals:**
   ```bash
   ralph -goals
   ```

3. **Run iterations:**
   ```bash
   ralph -iterations 10 -verbose
   ```

4. **Track progress:**
   ```bash
   ralph -goals
   ```

5. **When one goal completes, continue:**
   ```bash
   ralph -iterations 10  # Continues with next priority goal
   ```

## Goals vs Plans

| Aspect | Goals | Plans |
|--------|-------|-------|
| Level | High-level outcomes | Specific tasks |
| Creation | Manual or CLI | Manual or decomposed |
| Tracking | Progress percentage | Individual completion |
| Dependencies | Goal-to-goal | Within decomposed items |
| Persistence | goals.json | plan.json |

Use goals for project outcomes, use plans for specific implementation tasks.

## Best Practices

1. **Clear descriptions**: Be specific about what success looks like
2. **Success criteria**: Define measurable criteria
3. **Reasonable scope**: One goal = one logical feature area
4. **Use dependencies**: Order work correctly
5. **Track progress**: Check `-goals` regularly
