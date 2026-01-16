# Example: Build a fictional app with Ralph

This walkthrough shows a complete flow:
1. Start from a fictional requirement spec.
2. Generate a plan from notes.
3. Run iterations and explore iteration flags.

## 1) Fictional requirement spec (notes)

We will build a small app called **TrailBuddy**, a web app that helps hikers log routes,
share notes, and save offline maps.

Save the following as `notes.md`:

```md
Project: TrailBuddy

Requirements:
- Users can sign up, log in, and reset passwords.
- Users can create trail entries with name, location, distance, difficulty, and notes.
- Users can upload a GPX file and see a simplified route preview.
- Provide a public share link for each trail entry.
- Add offline map caching for saved trails.
- Include a simple admin view to delete abusive content.

Non-functional:
- API must respond in under 300ms for trail list queries.
- UI should work on mobile and desktop.
- Use SQLite for local development.
```

## 2) Generate a plan from notes

Use the notes to generate `plan.json`:

```bash
ralph -generate-plan -notes notes.md -output plan.json -verbose
```

At this point you can open `plan.json` to review the features, steps,
and expected outputs that Ralph will iterate on.

## 3) Start iterations to build the app

Run a baseline iteration loop:

```bash
ralph -iterations 10 -verbose
```

Ralph will read `plan.json`, pick the highest-priority feature, and execute
the agent workflow for each iteration until features are complete or the run ends.

## 4) Iteration CLI examples and what they do

These flags change how Ralph behaves while it is building your app.

### Basic control

```bash
# Run 5 iterations (auto-detects build system)
ralph -iterations 5
```

Effect: runs a short iteration loop using auto-detected build/test commands.

```bash
# Use a specific plan file
ralph -iterations 5 -plan plans/trailbuddy.json
```

Effect: swaps in a custom plan file without changing other settings.

### Output and debugging

```bash
# Verbose output
ralph -iterations 5 -verbose
```

Effect: shows detailed step-by-step reasoning and execution flow.

```bash
# Machine-readable output
ralph -iterations 5 -json-output
```

Effect: emits structured JSON suitable for CI pipelines and automation.

### Scope control

```bash
# Limit iterations per feature
ralph -iterations 20 -scope-limit 3
```

Effect: prevents any single feature from consuming the entire run; deferred
features are marked in `plan.json` for later.

```bash
# Stop after a time budget
ralph -iterations 20 -deadline 1h30m
```

Effect: enforces a hard stop when the deadline is reached.

### Recovery and replanning

```bash
# Automatic replanning after repeated failures
ralph -iterations 10 -auto-replan -replan-threshold 3 -replan-strategy agent
```

Effect: if a feature repeatedly fails, Ralph backs up the plan and asks
the agent to restructure remaining work.

```bash
# Adjust recovery behavior
ralph -iterations 10 -max-retries 2 -recovery-strategy rollback
```

Effect: limits retries per feature and rolls back on failure before moving on.

### Build system and environment

```bash
# Force a build system
ralph -iterations 5 -build-system go
```

Effect: tells Ralph which build/test commands to use instead of auto-detecting.

```bash
# Force CI-style behavior locally
ralph -iterations 5 -environment github-actions
```

Effect: makes Ralph behave like it is running in CI (useful for reproducibility).

### Agent options

```bash
# Switch the agent provider
ralph -iterations 5 -agent claude
```

Effect: uses a different AI agent for code changes and reasoning.

```bash
# Parallel execution with multi-agent mode
ralph -iterations 10 -multi-agent -parallel-agents 3
```

Effect: splits feature work across multiple agents in parallel to speed up delivery.

### Guidance during iterations

```bash
# Inject long-term project memory
ralph -iterations 5 -memory-file project-memory.json
```

Effect: adds persistent context to each iteration (naming conventions, design rules).

```bash
# Provide mid-run nudges
ralph -iterations 5 -nudge-file nudges.json
```

Effect: allows you to steer the next iteration without restarting the run.

## 5) Suggested minimal workflow

```bash
# 1) Create notes
vim notes.md

# 2) Generate plan
ralph -generate-plan -notes notes.md -output plan.json

# 3) Build the app
ralph -iterations 10 -verbose
```
