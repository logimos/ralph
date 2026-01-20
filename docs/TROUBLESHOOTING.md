# Ralph Troubleshooting Guide

This guide helps diagnose and resolve common issues with Ralph.

## Table of Contents

1. [Installation Issues](#installation-issues)
2. [Configuration Issues](#configuration-issues)
3. [Agent Issues](#agent-issues)
4. [Build System Issues](#build-system-issues)
5. [Iteration Issues](#iteration-issues)
6. [Recovery Issues](#recovery-issues)
7. [Memory and State Issues](#memory-and-state-issues)
8. [Performance Issues](#performance-issues)
9. [CI/CD Issues](#cicd-issues)

---

## Installation Issues

### `ralph: command not found`

**Cause:** Ralph is not in your PATH.

**Solution:**
```bash
# Check if Go bin is in PATH
echo $PATH | grep -q "$(go env GOPATH)/bin" && echo "OK" || echo "Missing"

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:$(go env GOPATH)/bin

# Or use GOBIN
export GOBIN=$HOME/bin
export PATH=$PATH:$GOBIN

# Reinstall
go install github.com/start-it/ralph@latest
```

### `go install` fails with version error

**Cause:** Go version too old.

**Solution:**
```bash
# Check Go version
go version  # Needs 1.21+

# Update Go
# macOS: brew upgrade go
# Linux: download from go.dev
```

### Build fails with missing dependencies

**Cause:** Dependencies not downloaded.

**Solution:**
```bash
cd ralph
go mod download
go mod tidy
go build .
```

---

## Configuration Issues

### Config file not being loaded

**Cause:** File not in expected location or wrong format.

**Diagnosis:**
```bash
# Check if file exists
ls -la .ralph.yaml .ralph.json ralph.config.yaml 2>/dev/null

# Run with verbose to see config loading
ralph -verbose -status
```

**Solution:**
- Ensure file is in current directory or home directory
- Check file name matches exactly (case-sensitive)
- Validate YAML/JSON syntax

### CLI flags not overriding config file

**Cause:** Flag not specified correctly.

**Solution:**
```bash
# Correct - explicit value
ralph -iterations 10 -verbose=true

# Boolean flags need =true or just the flag
ralph -verbose  # Sets to true
```

### Build system not detected

**Cause:** Project files not present or in subdirectory.

**Diagnosis:**
```bash
# Check for build files
ls go.mod package.json pom.xml build.gradle Cargo.toml 2>/dev/null
```

**Solution:**
```bash
# Explicitly specify
ralph -build-system go -iterations 5

# Or create appropriate file
go mod init myproject
```

---

## Agent Issues

### Agent command not found

**Cause:** Agent CLI not installed or not in PATH.

**Diagnosis:**
```bash
# Check if agent exists
which cursor-agent
which claude
```

**Solution:**
```bash
# Install the agent CLI
# For Cursor: check Cursor IDE installation
# For Claude: follow Anthropic CLI installation

# Or specify different agent
ralph -agent "/full/path/to/agent" -iterations 5
```

### Agent produces no output

**Cause:** Agent command incorrect or failing silently.

**Diagnosis:**
```bash
# Test agent directly
cursor-agent --help
cursor-agent --print "test prompt"
```

**Solution:**
- Verify agent installation
- Check agent credentials/authentication
- Use verbose mode to see agent output

### Agent timeout

**Cause:** Agent taking too long to respond.

**Solution:**
```bash
# Use environment with longer timeout
ralph -environment ci -iterations 5

# Or reduce complexity of requests
# Break features into smaller steps
```

---

## Build System Issues

### Type check failing unexpectedly

**Cause:** Custom command or wrong preset.

**Diagnosis:**
```bash
# Test command manually
go build ./...
npm run typecheck
```

**Solution:**
```bash
# Override with working command
ralph -typecheck "make lint" -iterations 5

# Check package.json for correct script name
cat package.json | grep typecheck
```

### Tests failing but code works

**Cause:** Test environment issues or missing setup.

**Solution:**
```bash
# Run tests manually first
go test ./... -v
npm test -- --verbose

# Check for required services
docker compose up -d  # If tests need database
```

### Wrong build system detected

**Cause:** Multiple build system files present.

**Solution:**
```bash
# Explicitly specify
ralph -build-system pnpm -iterations 5

# Or remove conflicting files
rm yarn.lock  # If using pnpm
```

---

## Iteration Issues

### Iteration stuck on same feature

**Cause:** Feature not being marked as tested.

**Diagnosis:**
```bash
# Check plan file
cat plan.json | jq '.[] | select(.tested == false)'

# Check agent output
ralph -iterations 1 -verbose
```

**Solution:**
- Ensure agent updates plan.json correctly
- Check feature has clear completion criteria
- Enable auto-replan to break deadlock

### Completion signal not detected

**Cause:** Agent not outputting `<promise>COMPLETE</promise>`.

**Solution:**
- Check prompt includes completion signal instructions
- Verify agent follows instructions
- Manually check if all features are tested

### Progress file not updating

**Cause:** Permission issues or wrong path.

**Solution:**
```bash
# Check permissions
ls -la progress.txt

# Specify different path
ralph -progress /tmp/progress.txt -iterations 5
```

---

## Recovery Issues

### Understanding Recovery vs Replanning

Ralph uses a two-tier failure handling system. Understanding when to use each is key to effective error recovery:

**Tier 1: Recovery (Per-Feature)**
- Handles failures within a single feature
- Flags: `-max-retries`, `-recovery-strategy`
- Actions: retry, skip, rollback
- Use when: Individual features are failing

**Tier 2: Replanning (Plan-Level)**  
- Restructures the entire plan when recovery isn't enough
- Flags: `-auto-replan`, `-replan-threshold`, `-replan-strategy`
- Actions: adjust plan, AI restructure
- Use when: Multiple features failing repeatedly

See [Two-Tier Failure Handling](FEATURES.md#two-tier-failure-handling) for the complete escalation flow.

### When to Use Recovery

**Scenario 1: Test is flaky (passes sometimes)**
```bash
# Use retry with more attempts
ralph -iterations 10 -max-retries 5 -recovery-strategy retry
```

**Scenario 2: One feature is broken but others are fine**
```bash
# Skip the problematic feature and continue
ralph -iterations 10 -recovery-strategy skip
```

**Scenario 3: Changes corrupted the codebase**
```bash
# Rollback via git and retry fresh
ralph -iterations 10 -recovery-strategy rollback
```

### When to Use Replanning

**Scenario 1: Multiple features failing in sequence**
```bash
# Enable auto-replanning after 3 consecutive failures
ralph -iterations 10 -auto-replan -replan-threshold 3
```

**Scenario 2: Plan structure is fundamentally wrong**
```bash
# Manually trigger replanning
ralph -replan

# Or use AI to restructure the entire plan
ralph -replan -replan-strategy agent
```

**Scenario 3: Requirements changed mid-project**
```bash
# Replanning will detect plan.json was modified externally
# and offer to restructure based on new requirements
ralph -iterations 10 -auto-replan
```

### Recovery not triggering

**Cause:** Failure not detected or recovery disabled.

**Diagnosis:**
```bash
# Check for failure detection
ralph -iterations 1 -verbose 2>&1 | grep -i failure
```

**Solution:**
```bash
# Enable verbose to see detection
ralph -verbose -iterations 5

# Ensure strategy is set
ralph -recovery-strategy retry -iterations 5
```

### Rollback strategy fails

**Cause:** Not in a git repository or no changes to rollback.

**Diagnosis:**
```bash
# Check git status
git status
git rev-parse --git-dir  # Should show .git
```

**Solution:**
```bash
# Initialize git if needed
git init
git add -A
git commit -m "Initial commit"

# Use different strategy if git not available
ralph -recovery-strategy skip -iterations 5
```

### Max retries exceeded immediately

**Cause:** `max_retries` set too low or consistent failures.

**Solution:**
```bash
# Increase retries for flaky tests
ralph -max-retries 5 -iterations 10

# Or enable replanning for systematic issues
ralph -auto-replan -iterations 10
```

### Replanning not triggering

**Cause:** Auto-replan disabled or threshold not reached.

**Solution:**
```bash
# Enable auto-replanning
ralph -auto-replan -iterations 10

# Lower threshold if needed (default is 3)
ralph -auto-replan -replan-threshold 2 -iterations 10

# Or manually trigger
ralph -replan
```

### Replanning creates unwanted changes

**Cause:** Agent-based replanning is too aggressive.

**Solution:**
```bash
# Use incremental strategy instead of agent
ralph -auto-replan -replan-strategy incremental -iterations 10

# Or restore a previous plan version
ralph -list-versions
ralph -restore-version 2

# Or disable replanning entirely
ralph -replan-strategy none -iterations 10
```

---

## Memory and State Issues

### Memory file corrupted

**Cause:** Invalid JSON or concurrent access.

**Solution:**
```bash
# Check file syntax
cat .ralph-memory.json | jq .

# Reset if corrupted
ralph -clear-memory

# Or delete manually
rm .ralph-memory.json
```

### Memories not being injected

**Cause:** Memory file not loaded or relevance too low.

**Diagnosis:**
```bash
# Check memory contents
ralph -show-memory

# Run with verbose
ralph -verbose -iterations 1
```

### Nudges not being detected

**Cause:** File not watched or nudge already acknowledged.

**Solution:**
```bash
# Check nudge status
ralph -show-nudges

# Clear and re-add
ralph -clear-nudges
ralph -nudge "focus:Important task"
```

---

## Performance Issues

### Iterations very slow

**Cause:** Agent response time or large project.

**Diagnosis:**
```bash
# Time a single iteration
time ralph -iterations 1 -verbose
```

**Solutions:**
- Use simpler prompts
- Break features into smaller steps
- Use multi-agent mode for parallelism
- Check agent service status

### High memory usage

**Cause:** Large memory file or many iterations.

**Solution:**
```bash
# Reduce memory retention
ralph -memory-retention 30 -iterations 10

# Prune old memories
ralph -clear-memory
```

### File system operations slow

**Cause:** Large number of files or network filesystem.

**Solution:**
- Work on local filesystem
- Add files to .gitignore to reduce scanning
- Use SSD storage

---

## CI/CD Issues

### Colors/spinners breaking CI output

**Cause:** Non-TTY environment not detected.

**Solution:**
```bash
# Disable colors explicitly
ralph -no-color -iterations 10

# Or use JSON output
ralph -json-output -iterations 10
```

### Different behavior in CI vs local

**Cause:** Environment detection affecting settings.

**Solution:**
```bash
# Force specific environment
ralph -environment local -iterations 10

# Or use same config
ralph -config .ralph.ci.yaml -iterations 10
```

### Timeouts in CI

**Cause:** CI environment slower than local.

**Solution:**
```bash
# Increase retries
ralph -max-retries 5 -iterations 10

# Use CI environment (auto-increases timeout)
ralph -environment ci -iterations 10
```

### Git operations failing in CI

**Cause:** Git not configured or read-only checkout.

**Solution:**
```bash
# Configure git in CI
git config --global user.email "ci@example.com"
git config --global user.name "CI"

# Use skip strategy instead of rollback
ralph -recovery-strategy skip -iterations 10
```

---

## Getting More Help

### Diagnostic Commands

```bash
# Version information
ralph -version

# Full help
ralph -help

# Verbose output for debugging
ralph -verbose -log-level debug -iterations 1

# Check configuration
ralph -status

# Test agent connectivity
cursor-agent --print "Hello"
```

### Reporting Bugs

Include in bug reports:
1. Ralph version: `ralph -version`
2. Go version: `go version`
3. OS and architecture
4. Configuration file (sanitized)
5. Error messages
6. Steps to reproduce
7. Verbose output: `ralph -verbose -log-level debug`

### Common Environment Variables

```bash
# Go paths
echo $GOPATH
echo $GOBIN

# CI detection
echo $CI
echo $GITHUB_ACTIONS
echo $GITLAB_CI
```
