# Troubleshooting

Common issues and solutions when using Ralph.

## Installation Issues

### "ralph: command not found"

**Problem**: After installation, `ralph` command is not recognized.

**Solution**: Ensure `$GOPATH/bin` is in your PATH:

```bash
# Add to your shell profile (.bashrc, .zshrc)
export PATH=$PATH:$(go env GOPATH)/bin

# Reload shell
source ~/.bashrc  # or ~/.zshrc
```

### "go: command not found"

**Problem**: Go is not installed.

**Solution**: Install Go 1.21+ from [golang.org](https://golang.org/dl/).

## Configuration Issues

### "config file not found"

**Problem**: Ralph can't find configuration file.

**Solution**: Ensure the file exists in the correct location:

```bash
# Check current directory
ls -la .ralph.yaml .ralph.json

# Or specify explicitly
ralph -config path/to/config.yaml
```

### "invalid build system"

**Problem**: Specified build system is not recognized.

**Solution**: Use a valid build system:

```bash
# Valid options
ralph -build-system go
ralph -build-system npm
ralph -build-system pnpm
ralph -build-system yarn
ralph -build-system gradle
ralph -build-system maven
ralph -build-system cargo
ralph -build-system python
ralph -build-system auto
```

## Agent Issues

### "agent command not found"

**Problem**: The AI agent CLI is not installed.

**Solution**: Install the agent and ensure it's in PATH:

```bash
# For Cursor Agent
# Follow installation instructions at cursor.so

# Verify installation
which cursor-agent
cursor-agent --version
```

### "agent execution failed"

**Problem**: Agent returns non-zero exit code.

**Solution**:

1. Check agent is authenticated:
   ```bash
   cursor-agent --print "Hello"
   ```

2. Enable verbose mode to see details:
   ```bash
   ralph -iterations 1 -verbose
   ```

3. Check progress.txt for error details.

## Build Issues

### "type check failed"

**Problem**: Code doesn't compile.

**Solution**:

1. Run type check manually:
   ```bash
   go build ./...
   # or for Node.js
   npm run typecheck
   ```

2. Fix the errors before continuing.

3. Consider using recovery strategy:
   ```bash
   ralph -iterations 5 -recovery-strategy skip
   ```

### "tests failed"

**Problem**: Tests are failing.

**Solution**:

1. Run tests manually:
   ```bash
   go test ./...
   # or
   npm test
   ```

2. Fix failing tests.

3. Enable retry strategy:
   ```bash
   ralph -iterations 5 -max-retries 5
   ```

## Iteration Issues

### "stuck on same feature"

**Problem**: Ralph keeps working on the same feature without progress.

**Solution**:

1. Enable verbose mode:
   ```bash
   ralph -iterations 5 -verbose
   ```

2. Enable auto-replan:
   ```bash
   ralph -iterations 5 -auto-replan -replan-threshold 3
   ```

3. Use scope limits:
   ```bash
   ralph -iterations 10 -scope-limit 3
   ```

4. Check feature complexity - consider breaking into smaller features:
   ```bash
   ralph -analyze-plan
   ```

### "no features to process"

**Problem**: All features are already tested or deferred.

**Solution**:

1. Check plan status:
   ```bash
   ralph -list-all
   ```

2. If features are deferred, un-defer them:
   - Edit plan.json and remove `"deferred": true` and `"defer_reason"`

3. If all are tested, add new features to plan.json.

## Recovery Issues

### "recovery not working"

**Problem**: Features keep failing despite recovery settings.

**Solution**:

1. Increase max retries:
   ```bash
   ralph -iterations 10 -max-retries 5
   ```

2. Try different strategy:
   ```bash
   ralph -iterations 10 -recovery-strategy rollback
   ```

3. Enable replanning:
   ```bash
   ralph -iterations 10 -auto-replan
   ```

### Understanding Recovery vs Replanning

**Recovery (Tier 1)**: Per-feature, handles individual failures
- Use when: Single feature failing, transient errors
- Flags: `-max-retries`, `-recovery-strategy`

**Replanning (Tier 2)**: Plan-level, restructures the plan
- Use when: Multiple features failing, plan structure wrong
- Flags: `-auto-replan`, `-replan-threshold`, `-replan-strategy`

### When to Use Recovery

- Test is failing due to bug in implementation
- Type check failing due to syntax error
- Agent produced invalid code
- Need to retry with more guidance

### When to Use Replanning

- Multiple consecutive features are failing
- Plan structure seems wrong
- Features have incorrect dependencies
- Need to restructure the approach

## Replanning Issues

### "replanning not triggering"

**Problem**: Auto-replan is enabled but not triggering.

**Solution**:

1. Check threshold:
   ```bash
   # Lower the threshold
   ralph -iterations 10 -auto-replan -replan-threshold 2
   ```

2. Verify failures are consecutive (not recovered).

3. Check replan strategy isn't `none`:
   ```bash
   ralph -iterations 10 -replan-strategy incremental
   ```

### "replanning creates unwanted changes"

**Problem**: Incremental replan modifies plan unexpectedly.

**Solution**:

1. Use `-list-versions` to see backups:
   ```bash
   ralph -list-versions
   ```

2. Restore previous version:
   ```bash
   ralph -restore-version 1
   ```

3. Use manual replanning with control:
   ```bash
   ralph -replan -replan-strategy incremental
   ```

## Memory Issues

### "memories not being used"

**Problem**: Stored memories not appearing in prompts.

**Solution**:

1. Check memory file exists:
   ```bash
   cat .ralph-memory.json
   ```

2. Verify memory format is correct:
   ```bash
   ralph -show-memory
   ```

3. Ensure memory isn't expired (check retention):
   ```bash
   ralph -iterations 5 -memory-retention 365
   ```

## Performance Issues

### "iterations are slow"

**Problem**: Each iteration takes too long.

**Solution**:

1. Check agent performance:
   ```bash
   time cursor-agent --print "test"
   ```

2. Reduce feature complexity - split large features.

3. Use scope limits to prevent over-iteration:
   ```bash
   ralph -iterations 10 -scope-limit 3
   ```

4. Set a deadline:
   ```bash
   ralph -iterations 10 -deadline 1h
   ```

## Getting Help

If these solutions don't help:

1. Enable debug logging:
   ```bash
   ralph -iterations 1 -verbose -log-level debug
   ```

2. Check progress.txt for detailed logs.

3. [Open an issue](https://github.com/start-it/ralph/issues) with:
   - Ralph version (`ralph -version`)
   - Command you ran
   - Error message
   - Relevant logs
