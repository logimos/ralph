# Long-Running Goal Memory

Persistent memory for architectural decisions and conventions across sessions.

## Overview

Ralph's memory system stores important decisions and patterns, injecting them into prompts so AI agents maintain consistency across runs.

## Memory Types

| Type | Description | Examples |
|------|-------------|----------|
| `decision` | Architectural choices | "Use PostgreSQL for persistence" |
| `convention` | Coding standards | "Use snake_case for database columns" |
| `tradeoff` | Accepted compromises | "Sacrificed type safety for performance" |
| `context` | Project knowledge | "Main service is in cmd/server" |

## Memory File

Memories are stored in `.ralph-memory.json`:

```json
{
  "entries": [
    {
      "id": "mem_1705420800123456789",
      "type": "decision",
      "content": "Use PostgreSQL for all persistence needs",
      "category": "infra",
      "created_at": "2026-01-16T12:00:00Z",
      "updated_at": "2026-01-16T12:00:00Z",
      "source": "agent"
    }
  ],
  "last_updated": "2026-01-16T12:00:00Z",
  "retention_days": 90
}
```

## AI Agent Memory Extraction

AI agents can create memories using markers in their output:

```
[REMEMBER:DECISION]Use PostgreSQL for all persistence needs[/REMEMBER]
[REMEMBER:CONVENTION]Use snake_case for all database column names[/REMEMBER]
[REMEMBER:TRADEOFF]Opted for eventual consistency to improve performance[/REMEMBER]
[REMEMBER:CONTEXT]The main API is served from cmd/api/main.go[/REMEMBER]
```

Ralph automatically extracts these markers and stores them.

## Memory Injection

During iterations, Ralph injects relevant memories into prompts:

```
[MEMORY CONTEXT - Previous decisions and conventions to follow:]
- [DECISION] Use PostgreSQL for all persistence needs
- [CONVENTION] Use snake_case for all database column names
[END MEMORY CONTEXT]
```

### Relevance Scoring

Memories are ranked by:
- **Type weight**: Decisions and conventions ranked higher
- **Category match**: Entries matching current feature's category get priority
- **Recency**: Recently updated memories ranked higher

## Commands

```bash
# Display all memories
ralph -show-memory

# Add a memory manually
ralph -add-memory "decision:Use PostgreSQL for persistence"
ralph -add-memory "convention:All exported functions must have comments"

# Clear all memories
ralph -clear-memory

# Use custom memory file
ralph -iterations 5 -memory-file project-memory.json

# Set retention period (days)
ralph -iterations 5 -memory-retention 30
```

## Configuration

```yaml
# .ralph.yaml
memory_file: .ralph-memory.json
memory_retention: 90  # Days to keep memories
```

## Memory Retention

Memories older than the retention period are automatically pruned at the start of each run. Default is 90 days.

## Example Workflow

1. **First run** - Agent makes decisions:
   ```
   Agent output: "Setting up the database layer.
   [REMEMBER:DECISION]Use PostgreSQL with pgx driver[/REMEMBER]
   [REMEMBER:CONVENTION]All queries go through repository pattern[/REMEMBER]"
   ```

2. **Subsequent runs** - Agent receives context:
   ```
   [MEMORY CONTEXT]
   - [DECISION] Use PostgreSQL with pgx driver
   - [CONVENTION] All queries go through repository pattern
   [END MEMORY CONTEXT]
   ```

3. **Result**: Agent maintains consistency without repeated instructions

## Memory vs Nudges

| Aspect | Memory | Nudges |
|--------|--------|--------|
| Purpose | Long-term knowledge | Real-time guidance |
| Duration | Persistent across sessions | Single iteration |
| Use case | Maintaining consistency | Steering current work |
| Creation | Manual or agent extraction | Manual only |

Use memory for "always do X", use nudges for "right now, focus on Y".

## Best Practices

1. **Let agents create memories**: Use `[REMEMBER:]` markers for important decisions
2. **Review periodically**: Check `-show-memory` to see what's stored
3. **Prune if needed**: Lower retention for fast-moving projects
4. **Be specific**: Clear, specific memories are more useful than vague ones
5. **Categories matter**: Memory injection prioritizes category matches
