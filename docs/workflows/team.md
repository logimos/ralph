# Team Collaboration

Workflows for multiple developers using Ralph together.

## Shared Configuration

Commit `.ralph.yaml` to your repository:

```yaml
# .ralph.yaml - shared team configuration
agent: cursor-agent
build_system: go
verbose: true

# Consistent recovery settings
max_retries: 3
recovery_strategy: retry

# Scope control
scope_limit: 5
```

## Shared Plan Files

### Centralized Plan

Keep one `plan.json` that the team collaborates on:

```bash
# Developer A adds features
ralph -generate-plan -notes new-features.md

# Developer B checks status
ralph -list-untested

# Developer C runs iterations
ralph -iterations 5
```

### Feature Branches

Each developer can work on separate plan files:

```bash
# Developer A: auth features
ralph -plan auth-plan.json -iterations 5

# Developer B: api features  
ralph -plan api-plan.json -iterations 5
```

## Using Goals

Define team goals in `goals.json`:

```json
{
  "goals": [
    {
      "id": "auth",
      "description": "Complete user authentication",
      "priority": 10,
      "assignee": "developer-a"
    },
    {
      "id": "api",
      "description": "Build REST API",
      "priority": 8,
      "assignee": "developer-b"
    }
  ]
}
```

Track team progress:

```bash
ralph -goals
```

## Using Milestones

Define milestones for sprints:

```json
[
  {
    "id": 1,
    "description": "User login",
    "milestone": "Sprint-1",
    "tested": true
  },
  {
    "id": 2,
    "description": "User registration",
    "milestone": "Sprint-1",
    "tested": false
  },
  {
    "id": 3,
    "description": "Admin dashboard",
    "milestone": "Sprint-2",
    "tested": false
  }
]
```

Track sprint progress:

```bash
ralph -milestones
```

## Code Review Workflow

### 1. AI Implements, Human Reviews

```bash
# AI implements feature
ralph -iterations 1

# Human reviews the changes
git diff HEAD~1

# Continue or fix
ralph -iterations 1
```

### 2. Multi-Agent Review

Configure reviewer agent:

```json
{
  "agents": [
    {"id": "impl", "role": "implementer", "command": "cursor-agent"},
    {"id": "review", "role": "reviewer", "command": "claude"}
  ]
}
```

Run with multi-agent:

```bash
ralph -iterations 5 -multi-agent
```

## Shared Memory

Share architectural decisions across the team:

```bash
# Add team-wide conventions
ralph -add-memory "convention:Use snake_case for DB columns"
ralph -add-memory "decision:PostgreSQL for all persistence"

# Commit memory file
git add .ralph-memory.json
git commit -m "Add team conventions"
```

## Conflict Resolution

### Plan Conflicts

When merging plan.json changes:

```bash
# After merge with conflicts
ralph -analyze-plan

# Fix and continue
ralph -iterations 5
```

### Progress Conflicts

Each developer's progress.txt is append-only, so conflicts are rare. If they occur, keep both changes.

## Communication via Nudges

Leave guidance for teammates:

```bash
# Developer A leaves a nudge
ralph -nudge "focus:Feature 5 is blocked on API changes"

# Developer B sees it
ralph -show-nudges
```

## Best Practices

1. **Commit config files**: Share `.ralph.yaml` and `.ralph-memory.json`
2. **Use milestones**: Align on sprint boundaries
3. **Review AI output**: Human review is still essential
4. **Document decisions**: Use memory for architectural choices
5. **Communicate via nudges**: Leave notes for teammates
6. **Small iterations**: Run 3-5 iterations at a time
7. **Frequent syncs**: Pull before running iterations
