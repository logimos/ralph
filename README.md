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

### Check Version

```bash
# Show version information
ralph -version
```

### Basic Workflow

Run iterative development cycles:

```bash
# Run 5 iterations (auto-detects build system)
ralph -iterations 5

# With verbose output
ralph -iterations 3 -verbose

# Use a different plan file
ralph -iterations 5 -plan my-plan.json

# Use Cursor Agent (default) or Claude
ralph -iterations 5 -agent cursor-agent
ralph -iterations 5 -agent claude

# Specify build system explicitly
ralph -iterations 5 -build-system gradle
ralph -iterations 5 -build-system maven
ralph -iterations 5 -build-system cargo
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
  -build-system string
        Build system preset (pnpm, npm, yarn, gradle, maven, cargo, go, python) or 'auto' for detection
  -config string
        Path to configuration file (default: auto-discover .ralph.yaml, .ralph.json)
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
        Command to run for testing (overrides build-system preset)
  -typecheck string
        Command to run for type checking (overrides build-system preset)
  -v, -verbose
        Enable verbose output
  -version
        Show version information and exit
```

### Build System Support

Ralph supports multiple build systems with auto-detection and presets:

**Supported Build Systems:**
- **pnpm** - `pnpm typecheck` / `pnpm test` (default)
- **npm** - `npm run typecheck` / `npm test`
- **yarn** - `yarn typecheck` / `yarn test`
- **gradle** - `./gradlew check` / `./gradlew test`
- **maven** - `mvn compile` / `mvn test`
- **cargo** - `cargo check` / `cargo test`
- **go** - `go build ./...` / `go test ./...`
- **python** - `mypy .` / `pytest`

**Auto-Detection:**
Ralph automatically detects the build system by checking for common project files:
- `build.gradle` or `gradlew` → Gradle
- `pom.xml` → Maven
- `Cargo.toml` → Cargo (Rust)
- `go.mod` → Go
- `setup.py`, `pyproject.toml`, or `requirements.txt` → Python
- `pnpm-lock.yaml` → pnpm
- `yarn.lock` → Yarn
- `package.json` → npm

**Usage Examples:**
```bash
# Auto-detect build system (default behavior)
ralph -iterations 5

# Explicitly specify build system
ralph -iterations 5 -build-system gradle

# Use auto-detection explicitly
ralph -iterations 5 -build-system auto

# Override individual commands
ralph -iterations 5 -build-system gradle -test "./gradlew test --tests MyTest"
```

## Configuration File

Ralph supports persistent configuration through YAML or JSON configuration files. This allows you to set default options for your project without specifying them on every command.

### Supported File Names

Ralph automatically discovers configuration files in the following order:

1. **Current directory** (first found wins):
   - `.ralph.yaml`
   - `.ralph.yml`
   - `.ralph.json`
   - `ralph.config.yaml`
   - `ralph.config.yml`
   - `ralph.config.json`

2. **Home directory** (same file names as above)

### Configuration Options

All configuration options are optional. Only specify the settings you want to customize.

**YAML Format (`.ralph.yaml`):**
```yaml
# AI agent command
agent: cursor-agent

# Build system preset (pnpm, npm, yarn, gradle, maven, cargo, go, python, auto)
build_system: go

# Custom commands (override build system preset)
typecheck: go build ./...
test: go test -v ./...

# File paths
plan: plan.json
progress: progress.txt

# Execution settings
iterations: 5
verbose: true
```

**JSON Format (`.ralph.json`):**
```json
{
  "agent": "cursor-agent",
  "build_system": "go",
  "typecheck": "go build ./...",
  "test": "go test -v ./...",
  "plan": "plan.json",
  "progress": "progress.txt",
  "iterations": 5,
  "verbose": true
}
```

### Configuration Precedence

Configuration values are applied in the following order (later values override earlier ones):

1. **Defaults** - Built-in default values
2. **Config file** - Values from auto-discovered or specified config file
3. **CLI flags** - Command-line arguments always take highest precedence

This means you can set project defaults in a config file and override specific values on the command line when needed.

### Examples

**Basic project configuration:**
```yaml
# .ralph.yaml
build_system: go
iterations: 3
```

**Full configuration for a Node.js project:**
```yaml
# .ralph.yaml
agent: cursor-agent
build_system: pnpm
plan: tasks/plan.json
progress: tasks/progress.txt
verbose: true
```

**Using a custom config file:**
```bash
# Use a specific config file
ralph -config production.yaml -iterations 10

# Override config file settings
ralph -iterations 1 -verbose  # Uses config file but overrides these values
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

## Versioning

Ralph uses [Semantic Versioning](https://semver.org/) (SemVer) for version numbers in the format `MAJOR.MINOR.PATCH` (e.g., `v1.2.3`).

### Version Numbering

- **MAJOR** version: Incremented for incompatible API changes or breaking changes
  - Breaking changes to CLI flags or behavior
  - Removing deprecated features
  - Major architectural changes
  
- **MINOR** version: Incremented for new functionality in a backward-compatible manner
  - New features (e.g., new build system support)
  - New CLI flags or options
  - Enhancements that don't break existing workflows
  
- **PATCH** version: Incremented for backward-compatible bug fixes
  - Bug fixes
  - Security patches
  - Documentation improvements
  - Performance improvements

### Examples

- `v1.0.0` → `v1.0.1`: Bug fix (patch) - e.g., fixing a build system detection issue
- `v1.0.0` → `v1.1.0`: New feature (minor) - e.g., adding Gradle support
- `v1.0.0` → `v2.0.0`: Breaking change (major) - e.g., changing required flag names

### Checking Your Version

```bash
# Check installed version
ralph -version

# Output: ralph version v1.2.3
```

The version is embedded at build time and displayed in:
- `ralph -version` command output
- `ralph -help` usage message
- GitHub release binaries

**Version Detection:**
- **Local builds**: Version is automatically detected from git tags using `git describe --tags`
- **GitHub Actions**: Version is extracted from the git tag that triggers the release workflow
- **Development builds**: Shows "dev" if no git tags are found

When building locally with `make build`, ralph will automatically use the latest git tag version (e.g., `v0.0.1`). When GitHub Actions builds a release, it uses the semantic version tag (e.g., `v1.2.3`) that triggered the workflow.

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
1. Calculate the new version based on the latest git tag using semantic versioning
2. Build binaries for multiple platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
3. Create checksums for all binaries
4. Output everything to the `dist/` directory

### Release Workflow

The release process follows semantic versioning and is automated via GitHub Actions:

1. **Determine version bump type:**
   - **Patch** (`release-patch`): Bug fixes, security patches, minor corrections
   - **Minor** (`release-minor`): New features, enhancements, backward-compatible changes
   - **Major** (`release-major`): Breaking changes, major refactoring, incompatible API changes

2. Run a release command to build binaries locally (optional, for testing):
   ```bash
   make release-patch  # or release-minor, release-major
   ```

3. Review the version in `.version` file and create/push the git tag:
   ```bash
   cat .version  # Verify version (e.g., v1.0.1)
   git tag v1.0.1
   git push origin v1.0.1
   ```

4. **GitHub Actions automatically:**
   - Detects the tag push
   - Builds binaries for all platforms (Linux, macOS, Windows)
   - Creates a GitHub release with the semantic version tag
   - Uploads all binaries and checksums

**Note**: If no git tags exist yet, the first release will start at:
- `v0.0.1` for patch releases
- `v0.1.0` for minor releases  
- `v1.0.0` for major releases

### Semantic Versioning with GitHub Actions

When you push a semantic version tag (e.g., `v1.2.3`), GitHub Actions automatically:

1. **Extracts the version** from the git tag (`v1.2.3`)
2. **Validates the format** to ensure it matches semantic versioning (`vMAJOR.MINOR.PATCH`)
3. **Builds binaries** for all platforms with the version embedded via ldflags
4. **Creates a GitHub release** with the correct version number
5. **Uploads binaries** that will show the correct version when users run `ralph -version`

**Important**: Always use semantic version tags (e.g., `v0.0.1`, `v1.2.3`) - the workflow validates this format and will fail if the tag doesn't match.

### Local Development Versioning

When building locally:
- Use `make build` to automatically detect version from git tags
- The version will be extracted from the latest git tag (e.g., `v0.0.1`)
- If no tags exist, it defaults to `dev`
- Uncommitted changes are handled gracefully (the `-dirty` suffix is stripped)

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

