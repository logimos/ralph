# Quick Start

Get Ralph running in 5 minutes.

## Step 1: Create a Plan

Ralph uses JSON plan files to define features. You can create one manually or generate it from notes.

### Generate from Notes

Create a `notes.md` file with your project requirements:

```markdown
# My Project

## Features
- User authentication with JWT
- REST API for CRUD operations
- Database integration with PostgreSQL
- Unit tests for all endpoints
```

Generate the plan:

```bash
ralph -generate-plan -notes notes.md
```

### Manual Plan

Or create `plan.json` directly:

```json
[
  {
    "id": 1,
    "category": "infra",
    "description": "Set up project structure",
    "steps": [
      "Create directory structure",
      "Initialize go.mod",
      "Add README.md"
    ],
    "expected_output": "Project structure exists with basic files",
    "tested": false
  },
  {
    "id": 2,
    "category": "feature",
    "description": "Add user authentication",
    "steps": [
      "Create auth middleware",
      "Implement JWT token generation",
      "Add login endpoint"
    ],
    "expected_output": "Users can authenticate via JWT",
    "tested": false
  }
]
```

## Step 2: Run Iterations

Execute development cycles:

```bash
# Run 5 iterations
ralph -iterations 5

# With verbose output
ralph -iterations 5 -verbose
```

Ralph will:

1. Find the first untested feature
2. Call the AI agent with implementation instructions
3. Run type checking and tests
4. Update the plan file
5. Create a git commit
6. Repeat until all features are complete

## Step 3: Monitor Progress

Check the current status:

```bash
# Show all features
ralph -list-all

# Show only completed features
ralph -list-tested

# Show remaining features
ralph -list-untested
```

## Common Options

```bash
# Use a different AI agent
ralph -iterations 5 -agent claude

# Use a specific plan file
ralph -iterations 5 -plan my-plan.json

# Specify build system explicitly
ralph -iterations 5 -build-system go
ralph -iterations 5 -build-system npm
ralph -iterations 5 -build-system cargo
```

## Configuration File

Create a `.ralph.yaml` file for persistent settings:

```yaml
# .ralph.yaml
agent: cursor-agent
build_system: go
iterations: 5
verbose: true
```

Then simply run:

```bash
ralph -iterations 5
```

## Example Workflow

```bash
# 1. Generate plan from notes
ralph -generate-plan -notes notes.md

# 2. Check the generated plan
ralph -list-all

# 3. Run iterations
ralph -iterations 10 -verbose

# 4. Check progress
ralph -list-untested

# 5. Continue if needed
ralph -iterations 5
```

## Next Steps

- [Configuration Guide](configuration.md) - Customize Ralph settings
- [Failure Recovery](../features/failure-recovery.md) - Handle errors automatically
- [Milestones](../features/milestones.md) - Organize features into milestones
