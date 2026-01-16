# CLI Tool Example

This example demonstrates using Ralph to build a professional CLI tool in Go using Cobra.

## Project Overview

A code generation CLI tool featuring:
- Project initialization
- Code generation with templates
- Configuration file support
- Colored output and progress indicators
- Shell completions and man pages

## Getting Started

### Prerequisites

- Go 1.21+
- Ralph installed

### Using Ralph

1. **Check the plan:**
   ```bash
   ralph -status -plan plan.json
   ```

2. **Run iterations:**
   ```bash
   ralph -iterations 10 -plan plan.json -verbose
   ```

3. **Track progress by milestone:**
   ```bash
   ralph -milestones -plan plan.json
   ```

## Milestones

| Milestone | Description | Features |
|-----------|-------------|----------|
| **Foundation** | Basic CLI setup | Cobra structure, init command |
| **Core** | Main functionality | Code generation, config support |
| **Polish** | Production readiness | UI, tests, release automation |

## Expected CLI Structure

After Ralph completes:

```
mytool/
├── cmd/
│   ├── root.go        # Base command, global flags
│   ├── init.go        # Project initialization
│   ├── generate.go    # Code generation
│   └── config.go      # Configuration management
├── internal/
│   ├── config/        # Config file handling
│   ├── templates/     # Code templates
│   └── ui/            # Output formatting
├── main.go
└── go.mod
```

## CLI Usage (Expected)

```bash
# Initialize a new project
mytool init --name myproject --template default

# Generate code
mytool generate model User --fields "name:string,email:string"
mytool generate controller Users

# View configuration
mytool config list
mytool config set template-dir ./templates
```

## Configuration

Create `.ralph.yaml`:

```yaml
build_system: go
plan: plan.json
iterations: 7
verbose: true

# Validate CLI builds correctly
typecheck: go build -o mytool .
test: go test ./...
```

## Validations

The plan includes CLI validations:
- Version flag works
- Help output is correct
- Commands have proper flags

Run validations:
```bash
ralph -validate -plan plan.json
```

## Key Features to Build

### 1. Cobra-based CLI
- Subcommand structure
- Global and local flags
- Help generation
- Shell completions

### 2. Template System
- Built-in templates
- Custom template support
- Variable interpolation
- Dry-run mode

### 3. Configuration
- YAML/JSON config files
- Environment variables
- Config precedence

### 4. Professional UX
- Colored output
- Progress spinners
- Error formatting
- --quiet and --verbose modes

## Distribution

After completion, the tool will support:
- Multi-platform builds via goreleaser
- Homebrew formula
- Shell completions (bash, zsh, fish)
- Man page generation
