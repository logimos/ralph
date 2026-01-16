# Simple API Example

This example demonstrates using Ralph to build a simple REST API in Go with SQLite.

## Project Overview

A basic user management API with:
- Health check endpoint
- SQLite database for persistence
- CRUD operations for users
- Configuration via environment variables

## Getting Started

### Prerequisites

- Go 1.21+
- Ralph installed (`go install github.com/start-it/ralph@latest`)

### Using Ralph

1. **Check the plan status:**
   ```bash
   ralph -status -plan plan.json
   ```

2. **Run development iterations:**
   ```bash
   ralph -iterations 10 -plan plan.json -verbose
   ```

3. **Track milestone progress:**
   ```bash
   ralph -milestones -plan plan.json
   ```

## Plan Structure

The plan is organized into milestones:

- **foundation**: Basic infrastructure (HTTP server, database)
- **api**: Core API functionality (CRUD endpoints, tests)
- **production**: Production readiness (config, docs)

## Validations

This example includes outcome-focused validations:
- Health endpoint validation
- API endpoint integration tests

Run validations after completing features:
```bash
ralph -validate -plan plan.json
```

## Ralph Configuration

You can create a `.ralph.yaml` file for this project:

```yaml
build_system: go
plan: plan.json
iterations: 5
verbose: true
```

## Expected Outcome

After Ralph completes all iterations:
- Fully functional REST API at http://localhost:8080
- SQLite database with user management
- Comprehensive test coverage
- Production-ready configuration
