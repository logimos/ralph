# Plan Management

Ralph uses JSON plan files to define and track features through the development process.

## Plan Structure

A plan file is a JSON array of feature objects:

```json
[
  {
    "id": 1,
    "category": "infra",
    "description": "Initialize project structure",
    "steps": [
      "Create directory structure",
      "Initialize package.json",
      "Add README.md"
    ],
    "expected_output": "Project structure exists with basic files",
    "tested": false
  }
]
```

### Plan Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | number | Unique identifier |
| `category` | string | Feature category (chore, infra, feature, etc.) |
| `description` | string | Clear, actionable description |
| `steps` | array | Specific implementation steps |
| `expected_output` | string | What success looks like |
| `tested` | boolean | Whether the feature is complete |
| `milestone` | string | Optional milestone name |
| `milestone_order` | number | Order within milestone |
| `deferred` | boolean | Whether feature was deferred |
| `defer_reason` | string | Reason for deferral |
| `validations` | array | Outcome validations |

## Generating Plans

Convert notes to structured plans:

```bash
# Generate from markdown notes
ralph -generate-plan -notes notes.md

# Custom output file
ralph -generate-plan -notes notes.md -output my-plan.json

# Verbose output
ralph -generate-plan -notes notes.md -verbose
```

### Notes Format

```markdown
# My Project

## Features
- User authentication
- REST API for products
- Admin dashboard

## Requirements
- Must use PostgreSQL
- TypeScript preferred
```

## Viewing Plan Status

```bash
# Show all features
ralph -list-all

# Show completed features
ralph -list-tested

# Show remaining features
ralph -list-untested

# Show deferred features
ralph -list-deferred

# Use different plan file
ralph -list-all -plan other-plan.json
```

## Plan Analysis

Analyze plans for potential improvements:

```bash
ralph -analyze-plan
```

### What It Detects

| Issue Type | Detection | Recommendation |
|------------|-----------|----------------|
| Compound Features | "verb X and verb Y" pattern | Split into separate features |
| Complex Features | >9 steps | Break into smaller features |

### Example Output

```
=== Plan Analysis Report ===

Total plans analyzed: 16
Issues found: 3
  - Compound features (with 'and'): 1
  - Complex features (>9 steps): 2

[WARNING] Feature #5: complex
  Feature #5 has 12 steps (>9), may be too complex
  Suggestions:
    Feature has 12 steps - consider splitting into smaller features
    Detected 4 potential logical groupings:
      Group 1 (3 steps): setup/config
      Group 2 (4 steps): implementation
      Group 3 (3 steps): testing
      Group 4 (2 steps): documentation
```

## Plan Refinement

Apply suggested refinements:

```bash
# Preview changes (writes to plan.refined.json)
ralph -analyze-plan

# Review the diff
diff plan.json plan.refined.json

# Apply refinements
ralph -refine-plan

# Dry run (preview without writing)
ralph -refine-plan -dry-run
```

## Categories

Common category values:

| Category | Description |
|----------|-------------|
| `chore` | Setup, configuration, tooling |
| `infra` | Infrastructure, deployment |
| `feature` | User-facing features |
| `db` | Database-related work |
| `ui` | User interface |
| `api` | API development |
| `security` | Security features |
| `testing` | Test-related work |
| `docs` | Documentation |

## Best Practices

1. **Specific Steps**: Each step should be implementable in one action
2. **Clear Outputs**: Expected output should be verifiable
3. **Right Size**: 3-7 steps per feature is ideal
4. **Dependencies**: Order features by dependency (implement before use)
5. **Categories**: Use consistent categories for filtering
