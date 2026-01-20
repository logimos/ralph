# Contributing to Ralph

Guidelines for contributing to the Ralph project.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- An AI agent CLI (for testing)

### Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/ralph.git
   cd ralph
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Build and test:
   ```bash
   make build
   make test
   ```

## Project Structure

```
ralph/
├── ralph.go              # Main entry point
├── ralph_test.go         # Main tests
├── Makefile              # Build commands
└── internal/             # Internal packages
    ├── config/           # Configuration
    ├── plan/             # Plan management
    ├── agent/            # AI agent execution
    ├── prompt/           # Prompt building
    ├── detection/        # Build system detection
    ├── environment/      # Environment detection
    ├── recovery/         # Failure recovery
    ├── replan/           # Replanning
    ├── scope/            # Scope control
    ├── memory/           # Memory system
    ├── nudge/            # Nudge system
    ├── milestone/        # Milestone tracking
    ├── goals/            # Goal management
    ├── validation/       # Outcome validation
    ├── multiagent/       # Multi-agent coordination
    └── ui/               # CLI output
```

## Code Style

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use meaningful variable names
- Add comments for exported functions

### Naming Conventions

- **Packages**: lowercase, single word
- **Exported functions**: PascalCase
- **Unexported functions**: camelCase
- **Constants**: PascalCase
- **Interfaces**: Verb-er pattern (Validator, Strategy)

### Comments

```go
// Package memory provides persistent storage for architectural decisions.
package memory

// Store manages memory entries and persistence.
type Store struct {
    // ...
}

// Add creates a new memory entry with the given type and content.
func (s *Store) Add(entryType, content string) (*Entry, error) {
    // ...
}
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/memory/...

# Run specific test
go test -run TestMemoryAdd ./internal/memory/...
```

### Writing Tests

- Use table-driven tests
- Test edge cases
- Use descriptive test names
- Use temporary directories for file tests

```go
func TestMemoryAdd(t *testing.T) {
    tests := []struct {
        name      string
        entryType string
        content   string
        wantErr   bool
    }{
        {"valid decision", "decision", "Use PostgreSQL", false},
        {"empty content", "decision", "", true},
        {"invalid type", "invalid", "content", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            store := NewStore(t.TempDir() + "/memory.json")
            _, err := store.Add(tt.entryType, tt.content)
            if (err != nil) != tt.wantErr {
                t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Pull Request Process

### Before Submitting

1. **Write tests**: All new features need tests
2. **Run tests**: `make test` must pass
3. **Run lint**: `make lint` must pass
4. **Update docs**: Document new features

### PR Guidelines

1. **One feature per PR**: Keep PRs focused
2. **Clear description**: Explain what and why
3. **Link issues**: Reference related issues
4. **Small commits**: Atomic, logical commits

### PR Template

```markdown
## Summary
Brief description of changes.

## Changes
- Added X feature
- Fixed Y bug
- Updated Z documentation

## Testing
- [ ] Unit tests added
- [ ] Integration tests added
- [ ] Manual testing performed

## Documentation
- [ ] README updated
- [ ] Docs updated
- [ ] Help text updated
```

## Adding a New Feature

### 1. Create Package

Create a new package in `internal/`:

```go
// internal/myfeature/myfeature.go
package myfeature

// Feature implements the new feature.
type Feature struct {
    // ...
}

// New creates a new Feature instance.
func New(config Config) *Feature {
    return &Feature{}
}
```

### 2. Add Tests

Create test file:

```go
// internal/myfeature/myfeature_test.go
package myfeature

import "testing"

func TestFeature(t *testing.T) {
    // ...
}
```

### 3. Integrate with CLI

Add CLI flags in `ralph.go`:

```go
// In parseFlags()
flag.BoolVar(&config.MyFeature, "my-feature", false, "Enable my feature")
```

### 4. Add Configuration

Update config file support:

```go
// In internal/config/file.go
type FileConfig struct {
    MyFeature bool `yaml:"my_feature" json:"my_feature,omitempty"`
}
```

### 5. Update Documentation

- Add to CLI help text
- Update README
- Create docs page
- Add to CONFIGURATION.md

## Commit Messages

Use clear, descriptive commit messages:

```
feat: add memory pruning functionality

- Add Prune() method to remove expired entries
- Add memory_retention configuration option
- Update tests for pruning behavior

Closes #123
```

Prefixes:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Tests
- `refactor:` Code refactoring
- `chore:` Build, deps, etc.

## Questions?

- Open an issue for bugs or features
- Start a discussion for questions
- Check existing issues before creating new ones
