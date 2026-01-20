# Frequently Asked Questions

Common questions about using Ralph.

## General Questions

### What AI agents does Ralph support?

Ralph works with any CLI-based AI agent. Built-in support includes:

- **Cursor Agent** (default): Uses `--print --force` flags
- **Claude CLI**: Uses `--permission-mode acceptEdits -p` format

Any other agent can be used by specifying the `-agent` flag.

### Do I need to write plan.json manually?

No! You can generate a plan from notes:

```bash
ralph -generate-plan -notes my-notes.md
```

Ralph will use the AI agent to convert your notes into a structured plan.

### How does Ralph know when a feature is complete?

Features are marked complete when:

1. The AI agent updates `tested: true` in plan.json
2. Type checking passes
3. Tests pass

### Can I use Ralph in CI/CD pipelines?

Yes! Ralph automatically detects CI environments and adapts:

- Enables verbose output for better logging
- Increases timeouts for slower CI runners
- Supports JSON output for machine parsing
- Disables colors in non-TTY environments

## Configuration Questions

### Where should I put my configuration file?

Ralph looks for config files in this order:

1. Current directory (`.ralph.yaml`, `.ralph.json`, etc.)
2. Home directory (same file names)

Use `-config path/to/config.yaml` for a custom location.

### How do CLI flags interact with config files?

Precedence from lowest to highest:

1. Built-in defaults
2. Configuration file
3. CLI flags (always win)

### Can I use environment variables?

Ralph detects environment variables for CI detection (`GITHUB_ACTIONS`, `CI`, etc.), but doesn't read configuration from environment variables. Use config files or CLI flags instead.

## Troubleshooting Questions

### Ralph is stuck on the same feature. What do I do?

Try these approaches:

1. Enable verbose mode:
   ```bash
   ralph -verbose -iterations 5
   ```

2. Enable auto-replan:
   ```bash
   ralph -auto-replan -iterations 10
   ```

3. Use scope limits:
   ```bash
   ralph -scope-limit 3 -iterations 10
   ```

4. Check feature complexity:
   ```bash
   ralph -analyze-plan
   ```

### Tests pass locally but Ralph says they fail. Why?

Common causes:

- Different working directory (Ralph runs from project root)
- Missing environment variables
- Race conditions in tests
- Test database not initialized

Run `ralph -verbose` to see the exact commands being executed.

### How do I recover from a bad state?

Options include:

1. Use rollback recovery:
   ```bash
   ralph -recovery-strategy rollback
   ```

2. Restore a plan version:
   ```bash
   ralph -restore-version N
   ```

3. Git reset manually and restart.

## Feature Questions

### What's the difference between memory and nudges?

| Aspect | Memory | Nudges |
|--------|--------|--------|
| Purpose | Long-term knowledge | Real-time guidance |
| Duration | Persistent across sessions | Single iteration |
| Use case | Maintaining consistency | Steering current work |

Use memory for "always do X", use nudges for "right now, focus on Y".

### How do milestones differ from goals?

| Aspect | Milestones | Goals |
|--------|------------|-------|
| Purpose | Group existing features | Define outcomes |
| Creation | Add to plan.json | Decomposed by AI |
| Tracking | Feature completion | Progress percentage |

Milestones organize what you have; goals define what you want.

### Can I run multiple AI agents in parallel?

Yes! Enable multi-agent mode:

```bash
ralph -multi-agent -agents agents.json -iterations 10
```

Configure different agents for implementation, testing, and review roles.

## Integration Questions

### Does Ralph work with monorepos?

Yes. Use a `.ralph.yaml` file with custom commands:

```yaml
typecheck: make typecheck-all
test: make test-all
```

### Can I integrate Ralph with GitHub Actions?

Yes! Example workflow:

```yaml
- name: Run Ralph
  run: |
    go install github.com/start-it/ralph@latest
    ralph -iterations 5 -json-output
```

Ralph auto-detects GitHub Actions and adjusts behavior.

### How do I validate that my API works, not just tests pass?

Use outcome validations in plan.json:

```json
{
  "validations": [
    {
      "type": "http_get",
      "url": "http://localhost:8080/health",
      "expected_status": 200
    }
  ]
}
```

Run validations: `ralph -validate`

## Best Practices

### How many iterations should I run?

Start with small batches:

- **Development**: 3-5 iterations at a time
- **CI/CD**: 5-10 iterations per run
- **Long runs**: Use scope limits and deadlines

### How should I structure my features?

- **3-7 steps** per feature is ideal
- **Specific steps**: One action per step
- **Clear outputs**: Verifiable expected output
- **Dependencies**: Order features correctly

### Should I use recovery or replanning?

Use recovery for:
- Single feature failing
- Transient errors
- Quick fixes

Use replanning for:
- Multiple failures
- Plan structure issues
- Major restructuring

### How do I manage large projects?

1. **Use milestones**: Group features into sprints
2. **Use goals**: Define high-level outcomes
3. **Use scope limits**: Prevent over-iteration
4. **Use deadlines**: Set time boundaries
5. **Review regularly**: Check `-list-deferred`
