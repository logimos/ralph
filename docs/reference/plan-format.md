# Plan Format Reference

Complete reference for the plan.json file format.

## Structure

A plan file is a JSON array of feature objects:

```json
[
  {
    "id": 1,
    "category": "infra",
    "description": "Initialize project structure",
    "steps": [
      "Create directory structure",
      "Initialize go.mod",
      "Add README.md"
    ],
    "expected_output": "Project structure exists with basic files",
    "tested": false
  }
]
```

## Fields

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | number | Unique identifier |
| `description` | string | Clear, actionable description |
| `tested` | boolean | Completion status |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `category` | string | Feature category |
| `steps` | array | Implementation steps |
| `expected_output` | string | What success looks like |
| `milestone` | string | Milestone name |
| `milestone_order` | number | Order within milestone |
| `deferred` | boolean | Whether feature is deferred |
| `defer_reason` | string | Reason for deferral |
| `validations` | array | Outcome validations |

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
| `other` | Miscellaneous |

## Steps

Steps should be:

- **Specific**: One clear action per step
- **Implementable**: AI can complete in one iteration
- **Ordered**: Listed in execution order

```json
"steps": [
  "Create users table schema",
  "Implement User model struct",
  "Add CRUD repository methods",
  "Write unit tests for repository"
]
```

## Milestones

Group features into milestones:

```json
{
  "id": 1,
  "description": "User login",
  "milestone": "Alpha",
  "milestone_order": 1,
  "tested": false
}
```

## Deferral

Features can be deferred with reason:

```json
{
  "id": 5,
  "description": "Complex feature",
  "tested": false,
  "deferred": true,
  "defer_reason": "iteration_limit"
}
```

Deferral reasons:
- `iteration_limit` - Exceeded iteration budget
- `deadline` - Deadline reached
- `complexity` - Too complex
- `manual` - Manually deferred

## Validations

Add outcome validations:

```json
{
  "id": 1,
  "description": "Health endpoint",
  "tested": true,
  "validations": [
    {
      "type": "http_get",
      "url": "http://localhost:8080/health",
      "expected_status": 200,
      "expected_body": "healthy",
      "description": "Health check returns healthy"
    }
  ]
}
```

### Validation Types

| Type | Required Fields |
|------|-----------------|
| `http_get` | `url` |
| `http_post` | `url` |
| `cli_command` | `command` |
| `file_exists` | `path` |
| `output_contains` | `pattern` |

## Complete Example

```json
[
  {
    "id": 1,
    "category": "infra",
    "description": "Initialize project structure",
    "steps": [
      "Create go.mod with module name",
      "Create directory structure",
      "Add .gitignore"
    ],
    "expected_output": "Project structure exists",
    "tested": true,
    "milestone": "Setup"
  },
  {
    "id": 2,
    "category": "feature",
    "description": "Add health endpoint",
    "steps": [
      "Create health handler",
      "Add route to router",
      "Write health check test"
    ],
    "expected_output": "GET /health returns 200",
    "tested": true,
    "milestone": "Alpha",
    "milestone_order": 1,
    "validations": [
      {
        "type": "http_get",
        "url": "http://localhost:8080/health",
        "expected_status": 200,
        "description": "Health endpoint works"
      }
    ]
  },
  {
    "id": 3,
    "category": "feature",
    "description": "Complex user management",
    "steps": ["...many steps..."],
    "expected_output": "User CRUD works",
    "tested": false,
    "deferred": true,
    "defer_reason": "iteration_limit"
  }
]
```

## Best Practices

1. **Unique IDs**: Each feature needs a unique ID
2. **Clear descriptions**: Be specific about what to implement
3. **Right-sized steps**: 3-7 steps per feature
4. **Ordered dependencies**: Earlier features shouldn't depend on later ones
5. **Testable outputs**: Expected output should be verifiable
6. **Consistent categories**: Use the same categories throughout
