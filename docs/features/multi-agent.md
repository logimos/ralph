# Multi-Agent Collaboration

Coordinate multiple AI agents working in parallel.

## Overview

Multi-agent mode enables specialized AI agents for different roles (implementation, testing, review, refactoring) to work together on features.

## Agent Roles

| Role | Purpose | Description |
|------|---------|-------------|
| `implementer` | Create code | Primary development work |
| `tester` | Write tests | Test writing and validation |
| `reviewer` | Check quality | Code review and suggestions |
| `refactorer` | Improve code | Code cleanup and optimization |

## Configuration

Create `agents.json`:

```json
{
  "agents": [
    {
      "id": "impl-1",
      "role": "implementer",
      "command": "cursor-agent",
      "specialization": "backend",
      "priority": 10,
      "enabled": true,
      "prompt_prefix": "You are a backend developer. Focus on server-side code."
    },
    {
      "id": "impl-2",
      "role": "implementer",
      "command": "cursor-agent",
      "specialization": "frontend",
      "priority": 10,
      "enabled": true,
      "prompt_prefix": "You are a frontend developer. Focus on UI components."
    },
    {
      "id": "test-1",
      "role": "tester",
      "command": "cursor-agent",
      "specialization": "testing",
      "priority": 8,
      "enabled": true,
      "prompt_prefix": "You are a QA engineer. Write comprehensive tests."
    },
    {
      "id": "review-1",
      "role": "reviewer",
      "command": "claude",
      "specialization": "code quality",
      "priority": 6,
      "enabled": true,
      "prompt_prefix": "You are a senior code reviewer. Check for best practices."
    }
  ],
  "max_parallel": 2,
  "conflict_resolution": "priority",
  "context_file": ".ralph-multiagent-context.json"
}
```

## Agent Fields

| Field | Description |
|-------|-------------|
| `id` | Unique identifier |
| `role` | implementer, tester, reviewer, refactorer |
| `command` | CLI command (e.g., "cursor-agent", "claude") |
| `specialization` | What this agent specializes in |
| `priority` | Execution priority (higher = earlier) |
| `enabled` | Whether to use this agent |
| `timeout` | Max execution time (e.g., "5m") |
| `prompt_prefix` | Text prepended to prompts |
| `prompt_suffix` | Text appended to prompts |

## Config Fields

| Field | Description | Default |
|-------|-------------|---------|
| `agents` | Array of agent configs | Required |
| `max_parallel` | Max concurrent agents | 2 |
| `conflict_resolution` | priority, merge, vote | priority |
| `context_file` | Shared context file | .ralph-multiagent-context.json |

## Commands

```bash
# List configured agents
ralph -list-agents

# Enable multi-agent mode
ralph -iterations 10 -multi-agent

# Custom agents config
ralph -iterations 10 -multi-agent -agents my-agents.json

# Set parallel limit
ralph -iterations 10 -multi-agent -parallel-agents 4
```

## Configuration File

```yaml
# .ralph.yaml
agents_file: agents.json
parallel_agents: 2
enable_multi_agent: true
```

## Workflow

When multi-agent mode is enabled:

1. **Implementation Stage**: Implementer agents create code (parallel)
2. **Testing Stage**: Tester agents validate (parallel, depends on impl)
3. **Review Stage**: Reviewer agents check quality (parallel, depends on impl)
4. **Refactoring Stage**: If review finds issues, refactorer agents improve

Each stage runs agents up to `max_parallel`, aggregating results before the next stage.

## Conflict Resolution

| Strategy | Description | Best For |
|----------|-------------|----------|
| `priority` | Use highest priority agent's result | Clear hierarchy |
| `merge` | Combine non-conflicting suggestions | Collaborative work |
| `vote` | Majority wins for conflicts | Democratic decisions |

## Shared Context

Agents communicate via a shared context file:

```json
{
  "feature_id": 1,
  "feature_description": "Add user authentication",
  "iteration": 3,
  "results": [
    {
      "agent_id": "impl-1",
      "role": "implementer",
      "status": "complete",
      "output": "Implementation complete...",
      "suggestions": ["Add error handling"],
      "issues": []
    }
  ],
  "messages": [],
  "decisions": [],
  "last_updated": "2026-01-16T12:00:00Z"
}
```

## Health Monitoring

Ralph monitors agent health:
- Tracks status (idle, running, complete, failed, timeout)
- Detects stuck agents
- Provides health status

## Example Workflow

1. **Create agents config:**
   ```bash
   # Create agents.json with your configurations
   ralph -list-agents
   ```

2. **Run with multi-agent:**
   ```bash
   ralph -iterations 10 -multi-agent -verbose
   ```

3. **Monitor progress:**
   - Implementation stage completes first
   - Testers validate the implementation
   - Reviewers check code quality
   - Refactorers improve based on feedback

4. **Review results:**
   - Check `.ralph-multiagent-context.json`
   - Review suggestions from all agents

## Multi-Agent vs Single Agent

| Aspect | Single Agent | Multi-Agent |
|--------|--------------|-------------|
| Execution | Sequential | Parallel stages |
| Perspectives | One viewpoint | Multiple specialized |
| Quality | Agent-dependent | Cross-validated |
| Speed | Limited by one | Parallelized |
| Complexity | Simple setup | Requires config |

## Best Practices

1. **Start with 2 agents**: Implementer + tester is a good start
2. **Use specializations**: Backend/frontend, unit/integration tests
3. **Set appropriate timeouts**: Longer for complex tasks
4. **Review context file**: Understand agent interactions
5. **Adjust parallelism**: Based on available resources
