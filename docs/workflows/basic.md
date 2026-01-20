# Basic Workflow

A step-by-step guide to using Ralph for AI-assisted development.

## The Development Cycle

```
┌─────────────────┐
│  Define Plan    │ ◄── Start here
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Run Iterations  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Monitor Progress│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Review & Ship  │
└─────────────────┘
```

## Step 1: Define Your Plan

### Option A: Generate from Notes

Create a markdown file with your requirements:

```markdown
# My API Project

## Features
- REST API for user management
- JWT authentication
- PostgreSQL database integration
- Rate limiting
- Comprehensive tests

## Technical Requirements
- Go 1.21+
- Chi router
- Clean architecture
```

Generate the plan:

```bash
ralph -generate-plan -notes notes.md
```

### Option B: Create Manually

Create `plan.json`:

```json
[
  {
    "id": 1,
    "category": "infra",
    "description": "Initialize project structure",
    "steps": [
      "Create go.mod with module name",
      "Create directory structure (cmd/, internal/, pkg/)",
      "Add .gitignore"
    ],
    "expected_output": "Project structure with go.mod exists",
    "tested": false
  },
  {
    "id": 2,
    "category": "infra",
    "description": "Set up Chi router with basic middleware",
    "steps": [
      "Install chi package",
      "Create main.go with router setup",
      "Add logging and recovery middleware"
    ],
    "expected_output": "Server starts and responds to requests",
    "tested": false
  }
]
```

## Step 2: Configure Ralph

Create `.ralph.yaml`:

```yaml
# .ralph.yaml
agent: cursor-agent
build_system: go
verbose: true
max_retries: 3
```

## Step 3: Run Iterations

Start the development cycle:

```bash
# Run 5 iterations
ralph -iterations 5

# With verbose output
ralph -iterations 5 -verbose
```

Ralph will:
1. Find the first untested feature
2. Send instructions to the AI agent
3. Run type checking and tests
4. Update plan.json (mark `tested: true`)
5. Create a git commit
6. Repeat for next feature

## Step 4: Monitor Progress

Check status anytime:

```bash
# Show all features
ralph -list-all

# Show completed
ralph -list-tested

# Show remaining
ralph -list-untested
```

## Step 5: Handle Issues

If something goes wrong, Ralph recovers automatically:

```bash
# Default: retry with enhanced prompts
ralph -iterations 10 -recovery-strategy retry

# Or skip problematic features
ralph -iterations 10 -recovery-strategy skip
```

## Step 6: Review and Ship

When all features are complete:

1. Review the code changes
2. Run final tests
3. Create a release

## Example: Complete Session

```bash
# 1. Create notes
cat > notes.md << 'EOF'
# User API
- CRUD operations for users
- Input validation
- Unit tests
EOF

# 2. Generate plan
ralph -generate-plan -notes notes.md

# 3. Review plan
ralph -list-all

# 4. Run iterations
ralph -iterations 10 -verbose

# 5. Check progress
ralph -list-tested

# 6. Continue if needed
ralph -iterations 5

# 7. Verify completion
ralph -list-untested
```

## Tips for Success

1. **Start small**: Begin with simple features
2. **Clear steps**: Each step should be one action
3. **Test often**: Run iterations in small batches
4. **Review output**: Check git commits as you go
5. **Use milestones**: Organize features into milestones

## Next Steps

- [CI/CD Integration](ci-cd.md) - Run Ralph in pipelines
- [Team Collaboration](team.md) - Multi-developer workflows
