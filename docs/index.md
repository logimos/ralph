# Ralph - AI-Assisted Development Workflow CLI

Ralph is a Golang CLI application that automates iterative development workflows by orchestrating AI-assisted development cycles. It processes plan files, executes development tasks through an AI agent CLI tool (like Cursor Agent or Claude), validates work, tracks progress, and commits changes iteratively until completion.

## Key Features

<div class="grid cards" markdown>

-   :material-format-list-checks:{ .lg .middle } **Plan Management**

    ---

    Generate plans from notes, track features, and monitor progress with milestone-based tracking.

    [:octicons-arrow-right-24: Learn more](features/plan-management.md)

-   :material-refresh:{ .lg .middle } **Failure Recovery**

    ---

    Intelligent two-tier failure handling with automatic retry, skip, and rollback strategies.

    [:octicons-arrow-right-24: Learn more](features/failure-recovery.md)

-   :material-brain:{ .lg .middle } **Long-Running Memory**

    ---

    Persistent memory for architectural decisions and coding conventions across sessions.

    [:octicons-arrow-right-24: Learn more](features/memory.md)

-   :material-robot:{ .lg .middle } **Multi-Agent Collaboration**

    ---

    Coordinate multiple AI agents working in parallel for implementation, testing, and review.

    [:octicons-arrow-right-24: Learn more](features/multi-agent.md)

</div>

## Quick Start

```bash
# Install Ralph
go install github.com/start-it/ralph@latest

# Generate a plan from your notes
ralph -generate-plan -notes notes.md

# Run development iterations
ralph -iterations 5 -verbose
```

## Why Ralph?

Ralph solves the challenge of orchestrating AI-assisted development workflows:

- **Autonomous Iteration**: Run multiple development cycles without manual intervention
- **Smart Recovery**: Automatically handles failures and adapts the plan when issues occur
- **Progress Tracking**: Maintains context across sessions with memory and progress files
- **Quality Assurance**: Validates outcomes beyond just running tests - verify endpoints, check file existence, run CLI commands
- **Team Coordination**: Multi-agent support enables specialized AI agents for different roles

## Getting Started

Ready to dive in? Start with our [Installation Guide](getting-started/installation.md) or jump straight to the [Quick Start](getting-started/quickstart.md).

## Features at a Glance

| Feature | Description |
|---------|-------------|
| [Plan Management](features/plan-management.md) | JSON-based feature tracking with generation from notes |
| [Failure Recovery](features/failure-recovery.md) | Per-feature retry, skip, and rollback strategies |
| [Scope Control](features/scope-control.md) | Iteration budgets and time limits |
| [Adaptive Replanning](features/replanning.md) | Dynamic plan adjustment when issues occur |
| [Memory System](features/memory.md) | Persistent architectural decisions and conventions |
| [Nudge System](features/nudges.md) | Mid-run guidance without stopping execution |
| [Milestones](features/milestones.md) | Organize features into project milestones |
| [Goals](features/goals.md) | High-level goals decomposed into actionable plans |
| [Validation](features/validation.md) | Outcome-focused validation beyond unit tests |
| [Multi-Agent](features/multi-agent.md) | Parallel AI agent coordination |

## Support

- [GitHub Issues](https://github.com/start-it/ralph/issues) - Bug reports and feature requests
- [GitHub Discussions](https://github.com/start-it/ralph/discussions) - Questions and community support
