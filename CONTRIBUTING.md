# Contributing to Ralph

Thank you for your interest in contributing to Ralph! This document provides guidelines and information for contributors.

## Code of Conduct

Please be respectful and constructive in all interactions. We welcome contributors of all experience levels.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, but recommended)

### Setting Up Development Environment

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/ralph.git
   cd ralph
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the project:**
   ```bash
   make build
   # or
   go build -o ralph .
   ```

4. **Run tests:**
   ```bash
   make test
   # or
   go test ./...
   ```

5. **Install locally for testing:**
   ```bash
   make install-local
   ```

## Project Structure

```
ralph/
├── ralph.go           # Main entry point, CLI parsing, orchestration
├── ralph_test.go      # Main package tests
├── internal/          # Internal packages (not importable externally)
│   ├── agent/         # AI agent execution
│   ├── config/        # Configuration handling
│   ├── detection/     # Build system detection
│   ├── environment/   # Environment detection (CI, resources)
│   ├── goals/         # Goal-oriented planning
│   ├── memory/        # Cross-session memory
│   ├── milestone/     # Milestone tracking
│   ├── multiagent/    # Multi-agent orchestration
│   ├── nudge/         # User nudge system
│   ├── plan/          # Plan file operations
│   ├── prompt/        # Prompt construction
│   ├── recovery/      # Failure recovery strategies
│   ├── replan/        # Adaptive replanning
│   ├── scope/         # Scope control
│   ├── ui/            # CLI output formatting
│   └── validation/    # Outcome validation
├── docs/              # Documentation
├── examples/          # Example projects
└── Makefile           # Build automation
```

## Development Guidelines

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format code (enforced by CI)
- Use `golint` and `go vet` for static analysis
- Keep functions focused and small
- Write descriptive variable and function names

### Package Design

- Each package should have a single responsibility
- Minimize dependencies between packages
- Use interfaces for testability
- Keep exported API surface small

### Testing

- Write unit tests for all new functionality
- Aim for >80% code coverage on core logic
- Use table-driven tests where appropriate
- Create test fixtures in `testdata/` directories
- Mock external dependencies

**Running tests:**
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Generate HTML coverage report
make test-coverage-html
```

### Documentation

- Add godoc comments to all exported functions, types, and constants
- Update README.md for user-facing changes
- Update relevant docs/ files for feature changes
- Include examples in documentation

## Making Changes

### Branching Strategy

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/my-feature
   ```

2. Make your changes with clear, atomic commits

3. Keep your branch up to date:
   ```bash
   git fetch origin
   git rebase origin/main
   ```

### Commit Messages

Follow conventional commit format:
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

**Examples:**
```
feat: add support for Maven build system
fix: handle empty plan file gracefully
docs: update configuration documentation
test: add tests for milestone progress calculation
refactor: extract prompt building to separate package
```

### Pull Requests

1. **Before submitting:**
   - Ensure all tests pass: `make test`
   - Run linting: `make lint`
   - Update documentation if needed
   - Add tests for new functionality

2. **PR description should include:**
   - Summary of changes
   - Motivation/context
   - How to test
   - Breaking changes (if any)

3. **Review process:**
   - PRs require at least one approval
   - Address review comments promptly
   - Keep PRs focused and reasonably sized

## Adding New Features

### Planning

1. Open an issue to discuss the feature
2. Get feedback before starting implementation
3. Consider backward compatibility

### Implementation Checklist

- [ ] Create new package in `internal/` if needed
- [ ] Implement core functionality
- [ ] Add unit tests (>80% coverage)
- [ ] Add integration tests if applicable
- [ ] Update CLI flags in `ralph.go`
- [ ] Update config file support if needed
- [ ] Add godoc comments
- [ ] Update README.md
- [ ] Update relevant docs/ files
- [ ] Add example usage

### Feature Flags and Configuration

New features should typically support:
1. CLI flag for enabling/configuring
2. Configuration file option
3. Sensible defaults

## Bug Reports

When reporting bugs, please include:

1. **Ralph version:** `ralph -version`
2. **Go version:** `go version`
3. **Operating system:** (e.g., Ubuntu 22.04, macOS 14)
4. **Steps to reproduce**
5. **Expected behavior**
6. **Actual behavior**
7. **Relevant logs** (with `-verbose` flag)

## Feature Requests

For feature requests, please include:

1. **Use case:** What problem does this solve?
2. **Proposed solution:** How should it work?
3. **Alternatives:** What other approaches did you consider?
4. **Additional context:** Any relevant examples or references

## Release Process

Releases are managed by maintainers:

1. Update version in relevant files
2. Create release notes
3. Tag the release: `git tag vX.Y.Z`
4. Push tag: `git push origin vX.Y.Z`
5. GitHub Actions builds and publishes release

## Getting Help

- **Questions:** Open a discussion or issue
- **Bugs:** Open an issue with reproduction steps
- **Security issues:** Email maintainers directly (do not open public issues)

## Recognition

Contributors are recognized in:
- Release notes
- GitHub contributors page

Thank you for contributing to Ralph!
