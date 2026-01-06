# Ralph - AI-Assisted Development Workflow CLI

Ralph is a Golang CLI application that automates iterative development workflows by orchestrating AI-assisted development cycles. It processes plan files, executes development tasks through an AI agent CLI tool (like Cursor Agent or Claude), validates work, tracks progress, and commits changes iteratively until completion.

## Features

- **Iterative Development**: Process features one at a time based on priority
- **AI Agent Integration**: Works with Cursor Agent, Claude, or any compatible CLI tool
- **Plan Management**: Generate plans from notes, list tested/untested features
- **Progress Tracking**: Automatically updates plan files and progress logs
- **Validation**: Integrates with type checking and testing commands
- **Git Integration**: Creates commits for completed features
- **Completion Detection**: Automatically detects when all work is complete

## Installation

### Using Go CLI (Recommended)

Install the latest version directly from GitHub:

```bash
go install github.com/start-it/ralph@latest
```

Install a specific version:

```bash
go install github.com/start-it/ralph@v1.0.0
```

Install from the main branch:

```bash
go install github.com/start-it/ralph@main
```

After installation, make sure `$GOPATH/bin` or `$GOBIN` is in your PATH. The binary will be available as `ralph`.

### From Source

```bash
git clone <repository-url>
cd start-it
make build
```

### Local Development Install

```bash
make install-local
```

This builds and installs the current version to `$GOPATH/bin` or `$GOBIN` for testing your local changes. Perfect for iterating on code changes without manually copying binaries.

## Usage

### Basic Workflow

Run iterative development cycles:

```bash
# Run 5 iterations
ralph -iterations 5

# With verbose output
ralph -iterations 3 -verbose

# Use a different plan file
ralph -iterations 5 -plan my-plan.json

# Use Cursor Agent (default) or Claude
ralph -iterations 5 -agent cursor-agent
ralph -iterations 5 -agent claude
```

### Plan Management

#### Generate Plan from Notes

Convert detailed notes into a structured plan.json:

```bash
# Generate plan.json from notes
ralph -generate-plan -notes .hidden/notes.md

# Custom output file
ralph -generate-plan -notes notes.md -output my-plan.json

# With verbose output
ralph -generate-plan -notes notes.md -verbose
```

#### View Plan Status

```bash
# Show both tested and untested features
ralph -status

# List only tested features
ralph -list-tested

# List only untested features
ralph -list-untested

# Use a different plan file
ralph -status -plan test.json
```

### Command-Line Options

```
Usage: ralph [options]

Options:
  -agent string
        Command name for the AI agent CLI tool (default "cursor-agent")
  -generate-plan
        Generate plan.json from notes file
  -iterations int
        Number of iterations to run (required)
  -list-tested
        List only tested features
  -list-untested
        List only untested features
  -notes string
        Path to notes file (required with -generate-plan)
  -output string
        Output plan file path (default: plan.json)
  -plan string
        Path to the plan file (e.g., plan.json) (default "plan.json")
  -progress string
        Path to the progress file (e.g., progress.txt) (default "progress.txt")
  -status
        List plan status (tested and untested features)
  -test string
        Command to run for testing (default "pnpm test")
  -typecheck string
        Command to run for type checking (default "pnpm typecheck")
  -v, -verbose
        Enable verbose output
```

## Plan File Format

Plans are JSON files containing an array of feature objects:

```json
[
  {
    "id": 1,
    "category": "chore",
    "description": "Initialize project structure",
    "steps": [
      "Create project root directory",
      "Initialize package.json",
      "Add README.md"
    ],
    "expected_output": "Project structure exists with basic files",
    "tested": false
  }
]
```

### Plan Fields

- `id` (number): Unique identifier for the feature
- `category` (string): Feature category (e.g., "chore", "infra", "db", "ui", "feature", "other")
- `description` (string): Clear, actionable description
- `steps` (array): Array of specific, implementable steps
- `expected_output` (string): What success looks like
- `tested` (boolean): Whether the feature has been tested (default: false)

## Workflow

1. **Plan Generation**: Create a plan.json from notes using `-generate-plan`
2. **Iteration**: Run `ralph -iterations N` to process features
3. **Agent Execution**: Ralph calls the AI agent with instructions to:
   - Find the highest-priority feature
   - Implement it
   - Run type checking and tests
   - Update the plan file
   - Append progress notes
   - Create a git commit
4. **Completion**: Ralph detects `<promise>COMPLETE</promise>` signal and exits

## AI Agent Integration

Ralph works with any CLI tool that accepts prompts. Supported agents:

- **Cursor Agent** (default): Uses `--print --force` flags
- **Claude**: Uses `--permission-mode acceptEdits -p` format

The agent receives a prompt that includes:
- References to the plan file and progress file
- Instructions for feature implementation
- Validation requirements
- Git commit instructions

## Examples

### Complete Workflow

```bash
# 1. Generate plan from notes
ralph -generate-plan -notes .hidden/notes.md

# 2. Review plan status
ralph -status

# 3. Run development iterations
ralph -iterations 10 -verbose

# 4. Check progress
ralph -status
```

### Using Makefile Commands

```bash
# Build
make build

# Install locally
make install-local

# Run with iterations
make run ITERATIONS=5

# Run with custom options
make run ITERATIONS=3 ARGS='-verbose -agent cursor-agent'

# View plan status
make jq-status

# List tested/untested
make jq-tested
make jq-untested
```

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
make lint
```

### Local Installation

After making changes, test locally:

```bash
make install-local
```

This builds and installs to `$GOPATH/bin` or `$GOBIN`, allowing you to test your changes immediately.

## Releases

### Creating Releases

Ralph includes Makefile commands to create releases with pre-built binaries for GitHub:

```bash
# Create a patch release (v1.0.0 -> v1.0.1)
make release-patch

# Create a minor release (v1.0.0 -> v1.1.0)
make release-minor

# Create a major release (v1.0.0 -> v2.0.0)
make release-major
```

These commands will:
1. Calculate the new version based on the latest git tag
2. Build binaries for multiple platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
3. Create checksums for all binaries
4. Output everything to the `dist/` directory

### Release Workflow

The release process is automated via GitHub Actions:

1. Run a release command to build binaries locally (optional, for testing):
   ```bash
   make release-patch  # or release-minor, release-major
   ```

2. Create and push the git tag:
   ```bash
   git tag v1.0.1  # Use the version shown in .version file
   git push origin v1.0.1
   ```

3. **GitHub Actions automatically:**
   - Detects the tag push
   - Builds binaries for all platforms (Linux, macOS, Windows)
   - Creates a GitHub release
   - Uploads all binaries and checksums

**Note**: If no git tags exist yet, the first release will start at `v0.0.1` (patch), `v0.1.0` (minor), or `v1.0.0` (major).

### Manual Release (Local Testing)

If you want to build release binaries locally for testing:

```bash
make release-patch
# Binaries will be in dist/ directory
```

The local build is useful for testing before pushing the tag, but the GitHub Action will rebuild everything when you push the tag.

## Requirements

- Go 1.21 or later
- An AI agent CLI tool (Cursor Agent, Claude, etc.)
- Git (for commit functionality)
- jq (optional, for Makefile plan status commands)

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]

