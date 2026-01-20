# Outcome-Focused Validation

Validate outcomes beyond unit tests - verify endpoints, CLI commands, and file existence.

## Overview

Validation goes beyond `go test` or `npm test`. It verifies that your system actually works: APIs respond correctly, CLIs execute properly, and files exist where expected.

## Validation Types

| Type | Description | Example Use Case |
|------|-------------|------------------|
| `http_get` | Verify HTTP GET endpoint | Health checks, API responses |
| `http_post` | Verify HTTP POST endpoint | Form submissions, API writes |
| `cli_command` | Verify CLI command execution | Tool integration, scripts |
| `file_exists` | Verify file exists with content | Config files, generated outputs |
| `output_contains` | Verify output matches pattern | Log validation |

## Defining Validations

Add validations to features in `plan.json`:

```json
{
  "id": 1,
  "description": "Health check endpoint",
  "tested": true,
  "validations": [
    {
      "type": "http_get",
      "url": "http://localhost:8080/health",
      "expected_status": 200,
      "expected_body": "\"status\":\\s*\"healthy\"",
      "description": "Health endpoint returns healthy"
    }
  ]
}
```

## Validation Fields

### Common Fields

| Field | Description |
|-------|-------------|
| `type` | Validation type (required) |
| `description` | Human-readable description |
| `timeout` | Timeout duration (e.g., "30s") |
| `retries` | Number of retries (default: 3) |

### HTTP Validation

| Field | Description |
|-------|-------------|
| `url` | URL to request (required) |
| `method` | HTTP method (defaults from type) |
| `body` | Request body for POST |
| `headers` | Map of HTTP headers |
| `expected_status` | Expected status code (default: 200) |
| `expected_body` | Regex pattern for response body |

### CLI Validation

| Field | Description |
|-------|-------------|
| `command` | Command to execute (required) |
| `args` | Array of arguments |
| `expected_body` | Regex pattern for stdout |
| `options.expected_exit_code` | Expected exit code (default: 0) |

### File Validation

| Field | Description |
|-------|-------------|
| `path` | File path to check (required) |
| `pattern` | Regex pattern for content |
| `options.should_exist` | Whether file should exist (default: true) |
| `options.min_size` | Minimum file size in bytes |

## Running Validations

```bash
# Validate all completed features
ralph -validate

# Validate a specific feature
ralph -validate-feature 5

# With verbose output
ralph -validate -verbose
```

### Example Output

```
=== Running Validations ===
Features to validate: 3

=== Feature #1: Health check endpoint ===
✓ GET http://localhost:8080/health returned 200

=== Feature #2: User API endpoint ===
✓ POST http://localhost:8080/api/users returned 201
✓ GET http://localhost:8080/api/users/1 returned 200

=== Validation Summary ===
Overall: PASSED
  Total validations: 3
  Passed: 3
  Failed: 0
```

## Examples

### API Health Check

```json
{
  "type": "http_get",
  "url": "http://localhost:8080/health",
  "expected_status": 200,
  "expected_body": "healthy",
  "description": "Health endpoint returns healthy"
}
```

### POST Endpoint

```json
{
  "type": "http_post",
  "url": "http://localhost:8080/api/users",
  "body": "{\"name\": \"test\"}",
  "headers": {"Content-Type": "application/json"},
  "expected_status": 201,
  "description": "Create user returns 201"
}
```

### CLI Tool

```json
{
  "type": "cli_command",
  "command": "./mytool",
  "args": ["--version"],
  "expected_body": "v\\d+\\.\\d+\\.\\d+",
  "description": "Tool version command works"
}
```

### Config File

```json
{
  "type": "file_exists",
  "path": "config/settings.yaml",
  "pattern": "database:",
  "description": "Config file contains database settings"
}
```

### Full-Stack App

```json
{
  "validations": [
    {
      "type": "cli_command",
      "command": "docker",
      "args": ["compose", "ps"],
      "expected_body": "Up",
      "description": "Docker containers are running"
    },
    {
      "type": "http_get",
      "url": "http://localhost:3000",
      "expected_status": 200,
      "description": "Frontend is accessible"
    },
    {
      "type": "http_get",
      "url": "http://localhost:8080/api/health",
      "expected_status": 200,
      "expected_body": "healthy",
      "description": "Backend API is healthy"
    },
    {
      "type": "http_post",
      "url": "http://localhost:8080/api/login",
      "body": "{\"email\":\"test@example.com\",\"password\":\"test\"}",
      "headers": {"Content-Type": "application/json"},
      "expected_status": 200,
      "expected_body": "token",
      "description": "Authentication works"
    }
  ]
}
```

## Validation Behavior

1. **Retries**: Automatic retries with exponential backoff (default: 3)
2. **Timeout**: Each validation has a timeout (default: 30s)
3. **Pattern Matching**: Uses Go regular expressions
4. **Progress Tracking**: Results logged to progress.txt

## Best Practices

1. **Start simple**: Begin with health checks and basic endpoints
2. **Descriptive descriptions**: Makes output easier to understand
3. **Appropriate timeouts**: Increase for slow endpoints
4. **Test patterns first**: Verify regex patterns work
5. **Validate completed features**: Run on `tested: true` features
