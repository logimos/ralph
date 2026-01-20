# CLI Output

Rich terminal output with colors, progress indicators, and structured logging.

## Features

- **Colored Output**: Success (green), errors (red), warnings (yellow), info (blue)
- **Progress Spinner**: Visual feedback during agent execution
- **Summary Dashboard**: End-of-run summary
- **Structured Logging**: Log levels for controlling verbosity
- **JSON Output**: Machine-readable output for automation

## Output Modes

| Mode | Flag | Description | Use Case |
|------|------|-------------|----------|
| Normal | (default) | Colored output | Interactive terminal |
| Quiet | `-quiet` | Errors only | Scripting |
| JSON | `-json-output` | Structured JSON | CI/CD pipelines |
| No Color | `-no-color` | Plain text | Non-TTY environments |

## Log Levels

Control verbosity with `-log-level`:

| Level | Shows |
|-------|-------|
| `debug` | All messages including detailed debugging |
| `info` (default) | Standard progress messages |
| `warn` | Only warnings and errors |
| `error` | Only error messages |

## Usage

```bash
# Quiet mode for scripts
ralph -iterations 10 -quiet

# JSON output for CI
ralph -iterations 10 -json-output

# Disable colors
ralph -iterations 10 -no-color

# Debug logging
ralph -iterations 10 -log-level debug -verbose
```

## Configuration

```yaml
# .ralph.yaml
no_color: false
quiet: false
json_output: false
log_level: info
```

## Summary Dashboard

At the end of each run:

```
=== Execution Summary ===
┌───────────────────────────────────────────┐
│ Progress:              8/10 iterations    │
│ Features completed:                   5   │
│ Features failed:                      1   │
│ Features skipped:                     2   │
│ Failures recovered:                   3   │
│ Duration:                         2m30s   │
└───────────────────────────────────────────┘
```

## CI Compatibility

Ralph automatically detects non-TTY environments:

- Disables colors when output is not a terminal
- Disables spinners and progress bars
- Enables verbose output by default in CI

## JSON Output Format

With `-json-output`, Ralph emits newline-delimited JSON:

```json
{"type":"iteration","number":1,"total":10,"feature_id":1}
{"type":"status","message":"Running type check"}
{"type":"status","message":"Running tests"}
{"type":"complete","feature_id":1,"success":true}
```

## Best Practices

1. **Use verbose in CI**: Better logging for debugging
2. **JSON for automation**: Parse output programmatically
3. **Quiet for scripts**: Only errors when embedding
4. **Debug when troubleshooting**: Full details for issues
