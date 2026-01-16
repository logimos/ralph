package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/logimos/ralph/internal/agent"
	"github.com/logimos/ralph/internal/config"
	"github.com/logimos/ralph/internal/detection"
	"github.com/logimos/ralph/internal/environment"
	"github.com/logimos/ralph/internal/goals"
	"github.com/logimos/ralph/internal/memory"
	"github.com/logimos/ralph/internal/milestone"
	"github.com/logimos/ralph/internal/nudge"
	"github.com/logimos/ralph/internal/plan"
	"github.com/logimos/ralph/internal/prompt"
	"github.com/logimos/ralph/internal/recovery"
	"github.com/logimos/ralph/internal/replan"
	"github.com/logimos/ralph/internal/scope"
	"github.com/logimos/ralph/internal/ui"
	"github.com/logimos/ralph/internal/validation"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
)

func main() {
	cfg := parseFlags()

	// Handle version command (exit early)
	if cfg.ShowVersion {
		fmt.Printf("ralph version %s\n", Version)
		os.Exit(0)
	}

	// Handle generate-plan command
	if cfg.GeneratePlan {
		if err := validateConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := generatePlanFromNotes(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle memory commands (don't require iterations or plan file)
	if cfg.ShowMemory || cfg.ClearMemory || cfg.AddMemory != "" {
		if err := handleMemoryCommands(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle nudge commands (don't require iterations or plan file)
	if cfg.ShowNudges || cfg.ClearNudges || cfg.Nudge != "" {
		if err := handleNudgeCommands(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle milestone commands (require plan file but not iterations)
	if cfg.ListMilestones || cfg.ShowMilestone != "" {
		if err := validateConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := handleMilestoneCommands(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle list commands (don't require iterations)
	if cfg.ListStatus || cfg.ListTested || cfg.ListUntested || cfg.ListDeferred {
		if err := validateConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if cfg.ListDeferred {
			if err := listDeferredFeatures(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		if err := listPlanStatus(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle replan-related commands
	if cfg.ListVersions || cfg.RestoreVersion > 0 || cfg.Replan {
		if err := validateConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := handleReplanCommands(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle validation commands
	if cfg.Validate || cfg.ValidateFeature > 0 {
		if err := validateConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := handleValidationCommands(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle goal commands
	if cfg.Goal != "" || cfg.GoalStatus || cfg.ListGoals || cfg.DecomposeGoal != "" || cfg.DecomposeAll {
		if err := handleGoalCommands(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := runIterations(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *config.Config {
	cfg := config.New()

	// Config file flag (parsed early to load file config before other flags)
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to configuration file (default: auto-discover .ralph.yaml, .ralph.json)")

	flag.StringVar(&cfg.PlanFile, "plan", config.DefaultPlanFile, "Path to the plan file (e.g., plan.json)")
	flag.StringVar(&cfg.ProgressFile, "progress", config.DefaultProgressFile, "Path to the progress file (e.g., progress.txt)")
	flag.IntVar(&cfg.Iterations, "iterations", 0, "Number of iterations to run (required)")
	flag.StringVar(&cfg.AgentCmd, "agent", config.DefaultAgentCmd, "Command name for the AI agent CLI tool")
	flag.StringVar(&cfg.BuildSystem, "build-system", "", "Build system preset (pnpm, npm, yarn, gradle, maven, cargo, go, python) or 'auto' for detection")
	flag.StringVar(&cfg.TypeCheckCmd, "typecheck", "", "Command to run for type checking (overrides build-system preset)")
	flag.StringVar(&cfg.TestCmd, "test", "", "Command to run for testing (overrides build-system preset)")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&cfg.Verbose, "v", false, "Enable verbose output (shorthand)")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&cfg.ListStatus, "status", false, "List plan status (tested and untested features)")
	flag.BoolVar(&cfg.ListTested, "list-tested", false, "List only tested features")
	flag.BoolVar(&cfg.ListUntested, "list-untested", false, "List only untested features")
	flag.BoolVar(&cfg.GeneratePlan, "generate-plan", false, "Generate plan.json from notes file")
	flag.StringVar(&cfg.NotesFile, "notes", "", "Path to notes file (required with -generate-plan)")
	flag.StringVar(&cfg.OutputPlanFile, "output", config.DefaultPlanFile, "Output plan file path (default: plan.json)")
	flag.IntVar(&cfg.MaxRetries, "max-retries", config.DefaultMaxRetries, "Maximum retries per feature before escalation (default: 3)")
	flag.StringVar(&cfg.RecoveryStrategy, "recovery-strategy", config.DefaultRecoveryStrategy, "Recovery strategy: retry, skip, rollback (default: retry)")
	flag.StringVar(&cfg.Environment, "environment", "", "Override detected environment (local, github-actions, gitlab-ci, jenkins, circleci, ci)")
	// UI-related flags
	flag.BoolVar(&cfg.NoColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "Minimal output (errors only)")
	flag.BoolVar(&cfg.Quiet, "q", false, "Minimal output (shorthand for -quiet)")
	flag.BoolVar(&cfg.JSONOutput, "json-output", false, "Machine-readable JSON output")
	flag.StringVar(&cfg.LogLevel, "log-level", config.DefaultLogLevel, "Log level: debug, info, warn, error")
	// Memory-related flags
	flag.StringVar(&cfg.MemoryFile, "memory-file", config.DefaultMemoryFile, "Path to memory file")
	flag.BoolVar(&cfg.ShowMemory, "show-memory", false, "Display stored memories")
	flag.BoolVar(&cfg.ClearMemory, "clear-memory", false, "Clear all stored memories")
	flag.StringVar(&cfg.AddMemory, "add-memory", "", "Add a memory entry (format: type:content where type is decision, convention, tradeoff, or context)")
	flag.IntVar(&cfg.MemoryRetention, "memory-retention", config.DefaultMemoryRetention, "Days to retain memories (default: 90)")
	// Milestone-related flags
	flag.BoolVar(&cfg.ListMilestones, "milestones", false, "List all milestones with progress")
	flag.StringVar(&cfg.ShowMilestone, "milestone", "", "Show features for a specific milestone")
	// Nudge-related flags
	flag.StringVar(&cfg.NudgeFile, "nudge-file", config.DefaultNudgeFile, "Path to nudge file")
	flag.StringVar(&cfg.Nudge, "nudge", "", "Add one-time nudge (format: type:content where type is focus, skip, constraint, or style)")
	flag.BoolVar(&cfg.ClearNudges, "clear-nudges", false, "Clear all nudges")
	flag.BoolVar(&cfg.ShowNudges, "show-nudges", false, "Display current nudges")
	// Scope control flags
	flag.IntVar(&cfg.ScopeLimit, "scope-limit", config.DefaultScopeLimit, "Max iterations per feature (0 = unlimited)")
	flag.StringVar(&cfg.Deadline, "deadline", "", "Deadline duration (e.g., '1h', '30m', '2h30m')")
	flag.BoolVar(&cfg.ListDeferred, "list-deferred", false, "List deferred features")
	// Replanning flags
	flag.BoolVar(&cfg.AutoReplan, "auto-replan", config.DefaultAutoReplan, "Enable automatic replanning when triggers fire")
	flag.BoolVar(&cfg.Replan, "replan", false, "Manually trigger replanning")
	flag.StringVar(&cfg.ReplanStrategy, "replan-strategy", config.DefaultReplanStrategy, "Replanning strategy: incremental, agent, none")
	flag.IntVar(&cfg.ReplanThreshold, "replan-threshold", config.DefaultReplanThreshold, "Consecutive failures before replanning (default: 3)")
	flag.BoolVar(&cfg.ListVersions, "list-versions", false, "List plan backup versions")
	flag.IntVar(&cfg.RestoreVersion, "restore-version", 0, "Restore a specific plan version")
	// Validation flags
	flag.BoolVar(&cfg.Validate, "validate", false, "Run validations for all completed features")
	flag.IntVar(&cfg.ValidateFeature, "validate-feature", 0, "Validate a specific feature by ID")
	// Goal flags
	flag.StringVar(&cfg.GoalsFile, "goals-file", config.DefaultGoalsFile, "Path to goals file")
	flag.StringVar(&cfg.Goal, "goal", "", "Add a high-level goal to decompose into plan items")
	flag.IntVar(&cfg.GoalPriority, "goal-priority", 5, "Priority for the goal (higher = more important)")
	flag.BoolVar(&cfg.GoalStatus, "goal-status", false, "Show progress toward all goals")
	flag.BoolVar(&cfg.ListGoals, "list-goals", false, "List all goals")
	flag.StringVar(&cfg.DecomposeGoal, "decompose-goal", "", "Decompose a specific goal by ID into plan items")
	flag.BoolVar(&cfg.DecomposeAll, "decompose-all", false, "Decompose all pending goals into plan items")

	flag.Usage = func() {
		// Version already includes 'v' prefix from git tags, so don't add another
		versionDisplay := Version
		if !strings.HasPrefix(Version, "v") && Version != "dev" {
			versionDisplay = "v" + Version
		}
		fmt.Fprintf(os.Stderr, "Ralph %s - AI-Assisted Development Workflow CLI\n\n", versionDisplay)
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nBuild System Presets:\n")
		fmt.Fprintf(os.Stderr, "  pnpm    - pnpm typecheck / pnpm test\n")
		fmt.Fprintf(os.Stderr, "  npm     - npm run typecheck / npm test\n")
		fmt.Fprintf(os.Stderr, "  yarn    - yarn typecheck / yarn test\n")
		fmt.Fprintf(os.Stderr, "  gradle  - ./gradlew check / ./gradlew test\n")
		fmt.Fprintf(os.Stderr, "  maven   - mvn compile / mvn test\n")
		fmt.Fprintf(os.Stderr, "  cargo   - cargo check / cargo test\n")
		fmt.Fprintf(os.Stderr, "  go      - go build ./... / go test ./...\n")
		fmt.Fprintf(os.Stderr, "  python  - mypy . / pytest\n")
		fmt.Fprintf(os.Stderr, "  auto    - Auto-detect from project files\n")
		fmt.Fprintf(os.Stderr, "\nConfiguration File:\n")
		fmt.Fprintf(os.Stderr, "  Ralph automatically discovers config files in this order:\n")
		fmt.Fprintf(os.Stderr, "    1. Current directory: .ralph.yaml, .ralph.yml, .ralph.json,\n")
		fmt.Fprintf(os.Stderr, "       ralph.config.yaml, ralph.config.yml, ralph.config.json\n")
		fmt.Fprintf(os.Stderr, "    2. Home directory: same file names\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Config file format (YAML example):\n")
		fmt.Fprintf(os.Stderr, "    agent: cursor-agent\n")
		fmt.Fprintf(os.Stderr, "    build_system: go\n")
		fmt.Fprintf(os.Stderr, "    typecheck: go build ./...\n")
		fmt.Fprintf(os.Stderr, "    test: go test ./...\n")
		fmt.Fprintf(os.Stderr, "    plan: plan.json\n")
		fmt.Fprintf(os.Stderr, "    progress: progress.txt\n")
		fmt.Fprintf(os.Stderr, "    iterations: 5\n")
		fmt.Fprintf(os.Stderr, "    verbose: true\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Priority: CLI flags > config file > defaults\n")
		fmt.Fprintf(os.Stderr, "\nRecovery Strategies:\n")
		fmt.Fprintf(os.Stderr, "  retry    - Retry the feature with enhanced guidance (default)\n")
		fmt.Fprintf(os.Stderr, "  skip     - Skip the feature and move to the next one\n")
		fmt.Fprintf(os.Stderr, "  rollback - Revert changes via git and retry fresh\n")
		fmt.Fprintf(os.Stderr, "\nEnvironment Detection:\n")
		fmt.Fprintf(os.Stderr, "  Ralph automatically detects the execution environment and adapts:\n")
		fmt.Fprintf(os.Stderr, "  - CI environments: longer timeouts, verbose output by default\n")
		fmt.Fprintf(os.Stderr, "  - Local: shorter timeouts, optimized for interactive use\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Supported CI providers (auto-detected):\n")
		fmt.Fprintf(os.Stderr, "    github-actions, gitlab-ci, jenkins, circleci, travis-ci, azure-devops\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Override with -environment flag or config file.\n")
		fmt.Fprintf(os.Stderr, "\nOutput Options:\n")
		fmt.Fprintf(os.Stderr, "  -no-color      Disable colored output (auto-disabled in non-TTY)\n")
		fmt.Fprintf(os.Stderr, "  -quiet, -q     Minimal output (errors only)\n")
		fmt.Fprintf(os.Stderr, "  -json-output   Machine-readable JSON output\n")
		fmt.Fprintf(os.Stderr, "  -log-level     Log verbosity: debug, info, warn, error (default: info)\n")
		fmt.Fprintf(os.Stderr, "\nMemory System:\n")
		fmt.Fprintf(os.Stderr, "  Ralph remembers architectural decisions and conventions across sessions.\n")
		fmt.Fprintf(os.Stderr, "  Memories are stored in %s (configurable with -memory-file).\n", config.DefaultMemoryFile)
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Memory types:\n")
		fmt.Fprintf(os.Stderr, "    decision   - Architectural choices (e.g., 'Use PostgreSQL for persistence')\n")
		fmt.Fprintf(os.Stderr, "    convention - Coding standards (e.g., 'Use snake_case for database columns')\n")
		fmt.Fprintf(os.Stderr, "    tradeoff   - Accepted compromises (e.g., 'Sacrifice type safety for performance')\n")
		fmt.Fprintf(os.Stderr, "    context    - Project knowledge (e.g., 'Main service is in cmd/server')\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  AI agents can create memories using markers in their output:\n")
		fmt.Fprintf(os.Stderr, "    [REMEMBER:DECISION]Use PostgreSQL for all persistence[/REMEMBER]\n")
		fmt.Fprintf(os.Stderr, "\nMilestone Tracking:\n")
		fmt.Fprintf(os.Stderr, "  Ralph supports milestone-based progress tracking.\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Define milestones by adding 'milestone' field to plan.json features:\n")
		fmt.Fprintf(os.Stderr, "    {\"id\": 1, \"description\": \"Feature\", \"milestone\": \"Alpha\"}\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Optional: Use 'milestone_order' to control feature order within a milestone.\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Commands:\n")
		fmt.Fprintf(os.Stderr, "    -milestones          List all milestones with progress\n")
		fmt.Fprintf(os.Stderr, "    -milestone <name>    Show features for a specific milestone\n")
		fmt.Fprintf(os.Stderr, "\nNudge System:\n")
		fmt.Fprintf(os.Stderr, "  Nudges provide lightweight mid-run guidance without stopping execution.\n")
		fmt.Fprintf(os.Stderr, "  Create/edit nudges.json during a run to steer Ralph.\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Nudge types:\n")
		fmt.Fprintf(os.Stderr, "    focus      - Prioritize a specific feature or approach\n")
		fmt.Fprintf(os.Stderr, "    skip       - Defer a feature or skip certain work\n")
		fmt.Fprintf(os.Stderr, "    constraint - Add a requirement or limitation\n")
		fmt.Fprintf(os.Stderr, "    style      - Specify coding style preferences\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Commands:\n")
		fmt.Fprintf(os.Stderr, "    -nudge <type:content>  Add a one-time nudge\n")
		fmt.Fprintf(os.Stderr, "    -show-nudges           Display current nudges\n")
		fmt.Fprintf(os.Stderr, "    -clear-nudges          Clear all nudges\n")
		fmt.Fprintf(os.Stderr, "    -nudge-file <path>     Use custom nudge file\n")
		fmt.Fprintf(os.Stderr, "\nScope Control:\n")
		fmt.Fprintf(os.Stderr, "  Ralph supports smart scope control to prevent over-building.\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Options:\n")
		fmt.Fprintf(os.Stderr, "    -scope-limit <n>       Max iterations per feature (0 = unlimited)\n")
		fmt.Fprintf(os.Stderr, "    -deadline <duration>   Time limit for the run (e.g., '1h', '30m', '2h30m')\n")
		fmt.Fprintf(os.Stderr, "    -list-deferred         List features that have been deferred\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  When a feature exceeds its iteration limit or the deadline is reached,\n")
		fmt.Fprintf(os.Stderr, "  Ralph automatically defers the feature and moves to the next one.\n")
		fmt.Fprintf(os.Stderr, "  Deferred features are marked in plan.json with 'deferred: true'.\n")
		fmt.Fprintf(os.Stderr, "\nAdaptive Replanning:\n")
		fmt.Fprintf(os.Stderr, "  Ralph can dynamically adjust plans when issues occur.\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Options:\n")
		fmt.Fprintf(os.Stderr, "    -auto-replan           Enable automatic replanning when triggers fire\n")
		fmt.Fprintf(os.Stderr, "    -replan                Manually trigger replanning\n")
		fmt.Fprintf(os.Stderr, "    -replan-strategy       Strategy: incremental, agent, none (default: incremental)\n")
		fmt.Fprintf(os.Stderr, "    -replan-threshold <n>  Consecutive failures before replanning (default: 3)\n")
		fmt.Fprintf(os.Stderr, "    -list-versions         List plan backup versions\n")
		fmt.Fprintf(os.Stderr, "    -restore-version <n>   Restore a specific plan version\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Replan triggers:\n")
		fmt.Fprintf(os.Stderr, "    - test_failure:       Repeated test failures (threshold reached)\n")
		fmt.Fprintf(os.Stderr, "    - requirement_change: plan.json externally modified\n")
		fmt.Fprintf(os.Stderr, "    - blocked_feature:    Feature becomes blocked/deferred\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Plan versioning creates backups as plan.json.bak.N before changes.\n")
		fmt.Fprintf(os.Stderr, "\nOutcome Validation:\n")
		fmt.Fprintf(os.Stderr, "  Ralph supports outcome-focused validation beyond tests and type checks.\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Validation types:\n")
		fmt.Fprintf(os.Stderr, "    http_get       - Verify HTTP GET endpoint returns expected response\n")
		fmt.Fprintf(os.Stderr, "    http_post      - Verify HTTP POST endpoint works correctly\n")
		fmt.Fprintf(os.Stderr, "    cli_command    - Verify CLI command executes successfully\n")
		fmt.Fprintf(os.Stderr, "    file_exists    - Verify file exists with expected content\n")
		fmt.Fprintf(os.Stderr, "    output_contains - Verify output contains expected pattern\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Add validations to plan.json features:\n")
		fmt.Fprintf(os.Stderr, "    {\n")
		fmt.Fprintf(os.Stderr, "      \"id\": 1, \"description\": \"API endpoint\",\n")
		fmt.Fprintf(os.Stderr, "      \"validations\": [\n")
		fmt.Fprintf(os.Stderr, "        {\"type\": \"http_get\", \"url\": \"http://localhost:8080/health\"}\n")
		fmt.Fprintf(os.Stderr, "      ]\n")
		fmt.Fprintf(os.Stderr, "    }\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Commands:\n")
		fmt.Fprintf(os.Stderr, "    -validate              Run validations for all completed features\n")
		fmt.Fprintf(os.Stderr, "    -validate-feature <id> Validate a specific feature\n")
		fmt.Fprintf(os.Stderr, "\nGoal-Oriented Planning:\n")
		fmt.Fprintf(os.Stderr, "  Ralph can decompose high-level goals into actionable plans using AI.\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Define goals in goals.json or via command line:\n")
		fmt.Fprintf(os.Stderr, "    {\n")
		fmt.Fprintf(os.Stderr, "      \"goals\": [\n")
		fmt.Fprintf(os.Stderr, "        {\n")
		fmt.Fprintf(os.Stderr, "          \"id\": \"auth\",\n")
		fmt.Fprintf(os.Stderr, "          \"description\": \"Add user authentication with OAuth\",\n")
		fmt.Fprintf(os.Stderr, "          \"priority\": 10,\n")
		fmt.Fprintf(os.Stderr, "          \"success_criteria\": [\"Users can log in via Google\", \"Sessions persist\"]\n")
		fmt.Fprintf(os.Stderr, "        }\n")
		fmt.Fprintf(os.Stderr, "      ]\n")
		fmt.Fprintf(os.Stderr, "    }\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  Commands:\n")
		fmt.Fprintf(os.Stderr, "    -goal <description>       Add a goal and decompose it into plan items\n")
		fmt.Fprintf(os.Stderr, "    -goal-priority <n>        Set priority for the goal (default: 5)\n")
		fmt.Fprintf(os.Stderr, "    -goal-status              Show progress toward all goals\n")
		fmt.Fprintf(os.Stderr, "    -list-goals               List all goals\n")
		fmt.Fprintf(os.Stderr, "    -decompose-goal <id>      Decompose a specific goal into plan items\n")
		fmt.Fprintf(os.Stderr, "    -decompose-all            Decompose all pending goals\n")
		fmt.Fprintf(os.Stderr, "    -goals-file <path>        Use custom goals file\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -version                         # Show version information\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 5                    # Run 5 iterations (auto-detect build system)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 5 -build-system gradle  # Use Gradle preset\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -config my-config.yaml           # Use specific config file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -status                          # Show plan status\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-tested                     # List tested features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-untested                   # List untested features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -milestones                      # Show milestone progress\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -milestone Alpha                 # Show features for 'Alpha' milestone\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -generate-plan -notes notes.md   # Generate plan.json from notes\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -generate-plan -notes notes.md -output my-plan.json  # Custom output file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -show-memory                     # Display stored memories\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -add-memory \"decision:Use PostgreSQL for persistence\"  # Add a memory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -clear-memory                    # Clear all memories\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -nudge \"focus:Work on feature 5 first\"  # Add a one-time nudge\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -show-nudges                     # Display current nudges\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -clear-nudges                    # Clear all nudges\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 5 -scope-limit 3     # Max 3 iterations per feature\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 10 -deadline 2h      # 2 hour time limit\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-deferred                   # Show deferred features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 5 -auto-replan       # Enable automatic replanning\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -replan -replan-strategy agent   # Manually trigger agent-based replanning\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-versions                   # Show plan backup versions\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -restore-version 2               # Restore plan version 2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -validate                        # Run validations for completed features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -validate-feature 5              # Validate specific feature\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -goal \"Add user authentication with OAuth\"  # Add and decompose goal\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -goal-status                     # Show progress toward goals\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-goals                      # List all goals\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -decompose-goal auth             # Decompose specific goal\n", os.Args[0])
	}

	flag.Parse()

	// Load configuration file (if specified or auto-discovered)
	cfg.ConfigFile = configFile
	loadConfigFile(cfg)

	// Apply build system configuration
	detection.ApplyBuildSystemConfig(cfg)

	return cfg
}

// loadConfigFile loads and applies configuration from a file.
// Priority: CLI flags > config file > defaults
func loadConfigFile(cfg *config.Config) {
	var configPath string

	if cfg.ConfigFile != "" {
		// Explicit config file specified
		configPath = cfg.ConfigFile
	} else {
		// Auto-discover config file
		configPath = config.DiscoverConfigFile()
	}

	if configPath == "" {
		return
	}

	// Store which flags were explicitly set on command line
	explicitFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		explicitFlags[f.Name] = true
	})

	fileCfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		if cfg.ConfigFile != "" {
			// Only warn if config file was explicitly specified
			fmt.Fprintf(os.Stderr, "Warning: failed to load config file %s: %v\n", configPath, err)
		}
		return
	}

	// Validate config file
	if err := config.ValidateFileConfig(fileCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: invalid config file %s: %v\n", configPath, err)
		return
	}

	// Apply file config, but only for values not explicitly set via CLI
	applyFileConfigWithPrecedence(cfg, fileCfg, explicitFlags)

	if cfg.Verbose {
		fmt.Printf("Loaded configuration from: %s\n", configPath)
	}
}

// applyFileConfigWithPrecedence applies file config values only when
// the corresponding CLI flag was not explicitly set.
func applyFileConfigWithPrecedence(cfg *config.Config, fileCfg *config.FileConfig, explicitFlags map[string]bool) {
	if fileCfg.Agent != "" && !explicitFlags["agent"] {
		cfg.AgentCmd = fileCfg.Agent
	}
	if fileCfg.BuildSystem != "" && !explicitFlags["build-system"] {
		cfg.BuildSystem = fileCfg.BuildSystem
	}
	if fileCfg.TypeCheck != "" && !explicitFlags["typecheck"] {
		cfg.TypeCheckCmd = fileCfg.TypeCheck
	}
	if fileCfg.Test != "" && !explicitFlags["test"] {
		cfg.TestCmd = fileCfg.Test
	}
	if fileCfg.Plan != "" && !explicitFlags["plan"] {
		cfg.PlanFile = fileCfg.Plan
	}
	if fileCfg.Progress != "" && !explicitFlags["progress"] {
		cfg.ProgressFile = fileCfg.Progress
	}
	if fileCfg.Iterations > 0 && !explicitFlags["iterations"] {
		cfg.Iterations = fileCfg.Iterations
	}
	if fileCfg.Verbose && !explicitFlags["verbose"] && !explicitFlags["v"] {
		cfg.Verbose = fileCfg.Verbose
	}
	if fileCfg.MaxRetries > 0 && !explicitFlags["max-retries"] {
		cfg.MaxRetries = fileCfg.MaxRetries
	}
	if fileCfg.RecoveryStrategy != "" && !explicitFlags["recovery-strategy"] {
		cfg.RecoveryStrategy = fileCfg.RecoveryStrategy
	}
	if fileCfg.Environment != "" && !explicitFlags["environment"] {
		cfg.Environment = fileCfg.Environment
	}
	// UI settings
	if fileCfg.NoColor && !explicitFlags["no-color"] {
		cfg.NoColor = fileCfg.NoColor
	}
	if fileCfg.Quiet && !explicitFlags["quiet"] && !explicitFlags["q"] {
		cfg.Quiet = fileCfg.Quiet
	}
	if fileCfg.JSONOutput && !explicitFlags["json-output"] {
		cfg.JSONOutput = fileCfg.JSONOutput
	}
	if fileCfg.LogLevel != "" && !explicitFlags["log-level"] {
		cfg.LogLevel = fileCfg.LogLevel
	}
	// Memory settings
	if fileCfg.MemoryFile != "" && !explicitFlags["memory-file"] {
		cfg.MemoryFile = fileCfg.MemoryFile
	}
	if fileCfg.MemoryRetention > 0 && !explicitFlags["memory-retention"] {
		cfg.MemoryRetention = fileCfg.MemoryRetention
	}
	// Nudge settings
	if fileCfg.NudgeFile != "" && !explicitFlags["nudge-file"] {
		cfg.NudgeFile = fileCfg.NudgeFile
	}
	// Scope control settings
	if fileCfg.ScopeLimit > 0 && !explicitFlags["scope-limit"] {
		cfg.ScopeLimit = fileCfg.ScopeLimit
	}
	if fileCfg.Deadline != "" && !explicitFlags["deadline"] {
		cfg.Deadline = fileCfg.Deadline
	}
	// Replan settings
	if fileCfg.AutoReplan && !explicitFlags["auto-replan"] {
		cfg.AutoReplan = fileCfg.AutoReplan
	}
	if fileCfg.ReplanStrategy != "" && !explicitFlags["replan-strategy"] {
		cfg.ReplanStrategy = fileCfg.ReplanStrategy
	}
	if fileCfg.ReplanThreshold > 0 && !explicitFlags["replan-threshold"] {
		cfg.ReplanThreshold = fileCfg.ReplanThreshold
	}
	// Goals settings
	if fileCfg.GoalsFile != "" && !explicitFlags["goals-file"] {
		cfg.GoalsFile = fileCfg.GoalsFile
	}
}

func validateConfig(cfg *config.Config) error {
	// Skip validation for generate-plan (handled separately)
	if cfg.GeneratePlan {
		if cfg.NotesFile == "" {
			return fmt.Errorf("notes file is required with -generate-plan (use -notes flag)")
		}
		notesPath := strings.TrimSpace(cfg.NotesFile)
		if notesPath == "" {
			return fmt.Errorf("notes file path cannot be empty")
		}
		if _, err := os.Stat(notesPath); os.IsNotExist(err) {
			return fmt.Errorf("notes file not found: %s", notesPath)
		}
		// Check if agent command exists
		if _, err := exec.LookPath(cfg.AgentCmd); err != nil {
			return fmt.Errorf("agent command not found in PATH: %s", cfg.AgentCmd)
		}
		return nil
	}

	// Skip iteration validation if we're just listing status or milestones
	if cfg.ListStatus || cfg.ListTested || cfg.ListUntested || cfg.ListMilestones || cfg.ShowMilestone != "" || cfg.ListDeferred {
		if _, err := os.Stat(cfg.PlanFile); os.IsNotExist(err) {
			return fmt.Errorf("plan file not found: %s", cfg.PlanFile)
		}
		return nil
	}

	if cfg.Iterations <= 0 {
		return fmt.Errorf("iterations must be a positive integer (use -iterations flag)")
	}

	if _, err := os.Stat(cfg.PlanFile); os.IsNotExist(err) {
		return fmt.Errorf("plan file not found: %s", cfg.PlanFile)
	}

	// Check if agent command exists
	if _, err := exec.LookPath(cfg.AgentCmd); err != nil {
		return fmt.Errorf("agent command not found in PATH: %s", cfg.AgentCmd)
	}

	// Validate recovery strategy
	if _, err := recovery.ParseStrategyType(cfg.RecoveryStrategy); err != nil {
		return err
	}

	// Validate max retries
	if cfg.MaxRetries < 0 {
		return fmt.Errorf("max-retries cannot be negative")
	}

	// Validate scope limit
	if cfg.ScopeLimit < 0 {
		return fmt.Errorf("scope-limit cannot be negative")
	}

	// Validate deadline format
	if cfg.Deadline != "" {
		if _, err := config.ParseDeadline(cfg.Deadline); err != nil {
			return fmt.Errorf("invalid deadline format: %w", err)
		}
	}

	return nil
}

func runIterations(cfg *config.Config) error {
	// Create UI instance
	uiCfg := ui.OutputConfig{
		NoColor:    cfg.NoColor,
		Quiet:      cfg.Quiet,
		JSONOutput: cfg.JSONOutput,
		LogLevel:   ui.ParseLogLevel(cfg.LogLevel),
	}
	output := ui.New(uiCfg)

	// Start timing for summary
	startTime := time.Now()

	// Detect environment
	var envProfile *environment.EnvironmentProfile
	if cfg.Environment != "" {
		envType := environment.ParseEnvironmentType(cfg.Environment)
		envProfile = environment.ForceEnvironment(envType)
	} else {
		envProfile = environment.Detect()
	}

	// Apply environment-based recommendations if not explicitly set
	if !cfg.Verbose && envProfile.RecommendedVerbose {
		cfg.Verbose = true
	}

	// Auto-disable colors in CI if not explicitly set
	if envProfile.CIEnvironment && !cfg.NoColor {
		// Keep colors in CI for modern terminals that support it
	}

	// Load memory store
	memStore := memory.NewStore(cfg.MemoryFile)
	memStore.SetRetentionDays(cfg.MemoryRetention)
	if err := memStore.Load(); err != nil {
		output.Warn("Failed to load memory: %v", err)
	}

	// Prune expired memories
	pruned, _ := memStore.Prune()
	if pruned > 0 && cfg.Verbose {
		output.Debug("Pruned %d expired memories", pruned)
	}

	// Load nudge store
	nudgeStore := nudge.NewStore(cfg.NudgeFile)
	if err := nudgeStore.Load(); err != nil {
		output.Debug("No nudge file loaded: %v", err)
	}

	output.Header("Ralph - Iterative Development Workflow")
	output.Info("Plan file: %s", cfg.PlanFile)
	output.Info("Progress file: %s", cfg.ProgressFile)
	output.Info("Iterations: %d", cfg.Iterations)
	output.Info("Agent command: %s", cfg.AgentCmd)
	output.Info("Recovery strategy: %s (max %d retries)", cfg.RecoveryStrategy, cfg.MaxRetries)
	if memStore.Count() > 0 {
		output.Info("Memory: %d entries loaded from %s", memStore.Count(), cfg.MemoryFile)
	}
	if nudgeStore.ActiveCount() > 0 {
		output.Info("Nudges: %d active nudge(s) from %s", nudgeStore.ActiveCount(), cfg.NudgeFile)
	}
	
	// Load plans and create milestone manager
	plans, planErr := plan.ReadFile(cfg.PlanFile)
	var milestoneMgr *milestone.Manager
	var completedMilestonesBefore map[string]bool
	if planErr == nil {
		milestoneMgr = milestone.NewManager(plans)
		
		// Record which milestones are complete before we start
		completedMilestonesBefore = make(map[string]bool)
		for _, p := range milestoneMgr.GetCompletedMilestones() {
			completedMilestonesBefore[p.Milestone.Name] = true
		}
		
		// Show milestone progress in verbose mode
		if cfg.Verbose && milestoneMgr.HasMilestones() {
			output.SubHeader("Milestone Progress")
			for _, p := range milestoneMgr.CalculateAllProgress() {
				output.Print("  %s", milestone.FormatProgress(p))
			}
		}
	}
	
	if cfg.Verbose {
		output.Debug("Type check command: %s", cfg.TypeCheckCmd)
		output.Debug("Test command: %s", cfg.TestCmd)
		output.Print("")
		output.Print("%s", envProfile.Summary())
	}
	output.Print("")

	// Initialize recovery manager
	strategyType, _ := recovery.ParseStrategyType(cfg.RecoveryStrategy)
	recoveryMgr := recovery.NewRecoveryManager(cfg.MaxRetries, strategyType)

	// Initialize replan manager
	replanMgr := replan.NewReplanManager(cfg.PlanFile, cfg.AgentCmd, cfg.AutoReplan)
	replanStrategyType, _ := replan.ParseStrategyType(cfg.ReplanStrategy)
	consecutiveFailures := 0

	// Initialize scope manager
	scopeConstraints := &scope.Constraints{
		MaxIterationsPerFeature: cfg.ScopeLimit,
		AutoDefer:               true,
	}
	scopeMgr := scope.NewManager(scopeConstraints)

	// Set deadline if specified
	if cfg.Deadline != "" {
		deadline, _ := config.ParseDeadline(cfg.Deadline)
		scopeMgr.SetDeadline(deadline)
	}

	// Show scope info if scope control is enabled
	if cfg.ScopeLimit > 0 || cfg.Deadline != "" {
		output.Info("Scope control: %s", formatScopeInfo(cfg))
	}
	
	// Show replan info if enabled
	if cfg.AutoReplan {
		output.Info("Auto-replan: enabled (strategy: %s, threshold: %d failures)", cfg.ReplanStrategy, cfg.ReplanThreshold)
	}

	// Track metrics for summary
	var summary ui.Summary
	summary.TotalIterations = cfg.Iterations
	summary.StartTime = startTime

	// Track the current feature being worked on (extracted from output if possible)
	currentFeatureID := 0
	currentFeatureSteps := 0
	currentFeatureDesc := ""
	var additionalPromptGuidance string

	for i := 1; i <= cfg.Iterations; i++ {
		// Check deadline before starting iteration
		if scopeMgr.IsDeadlineExceeded() {
			output.Warn("Deadline exceeded - stopping execution")
			break
		}

		// Get current feature from plans (first untested, non-deferred)
		detectedFeatureID, detectedSteps, detectedDesc := extractCurrentFeatureFromPlans(cfg.PlanFile)
		if detectedFeatureID > 0 && detectedFeatureID != currentFeatureID {
			// New feature detected - start tracking it
			currentFeatureID = detectedFeatureID
			currentFeatureSteps = detectedSteps
			currentFeatureDesc = detectedDesc
			scopeMgr.StartFeature(currentFeatureID, currentFeatureSteps, currentFeatureDesc)
			if cfg.Verbose {
				complexity := scope.EstimateComplexity(currentFeatureSteps, currentFeatureDesc)
				output.Debug("Working on feature #%d (%s complexity): %s", 
					currentFeatureID, complexity, currentFeatureDesc)
			}
		}

		output.Header("Iteration %d/%d", i, cfg.Iterations)
		summary.IterationsRun = i

		// Record iteration for scope tracking
		scopeMgr.RecordIteration(currentFeatureID)

		// Check if current feature should be deferred
		if shouldDefer, reason := scopeMgr.ShouldDefer(currentFeatureID); shouldDefer && currentFeatureID > 0 {
			scopeMgr.DeferFeature(currentFeatureID, reason)
			output.Warn("Feature #%d deferred: %s", currentFeatureID, scope.FormatDeferralReason(reason))
			
			// Mark feature as deferred in plan file
			if err := markFeatureDeferred(cfg.PlanFile, currentFeatureID, string(reason)); err != nil {
				output.Debug("Failed to update plan file: %v", err)
			}
			
			// Log deferral to progress file
			deferMsg := fmt.Sprintf("DEFERRED: Feature #%d - %s (iterations used: %d)", 
				currentFeatureID, scope.FormatDeferralReason(reason), 
				scopeMgr.GetFeatureScope(currentFeatureID).IterationsUsed)
			appendProgress(cfg.ProgressFile, deferMsg)
			
			summary.FeaturesSkipped++
			
			// Reset current feature - agent will move to next
			currentFeatureID = 0
		}

		// Check for simplification suggestion
		if currentFeatureID > 0 && scopeMgr.ShouldSuggestSimplification(currentFeatureID) && 
			!scopeMgr.WasSimplificationSuggested(currentFeatureID) {
			suggestions := scope.SuggestSimplification(currentFeatureSteps, currentFeatureDesc)
			output.Warn("Feature #%d may be complex. Suggestions:", currentFeatureID)
			for _, s := range suggestions {
				output.Print("  - %s", s)
			}
			scopeMgr.MarkSimplificationSuggested(currentFeatureID)
		}

		if cfg.Verbose {
			output.Debug("Executing agent command...")
			if cfg.ScopeLimit > 0 && currentFeatureID > 0 {
				remaining := scopeMgr.RemainingIterations(currentFeatureID)
				output.Debug("Scope: %d iterations remaining for current feature", remaining)
			}
		}

		// Show spinner for agent execution if TTY
		var spinner *ui.Spinner
		if output.IsTTY() && !cfg.Quiet && !cfg.JSONOutput {
			spinner = output.NewSpinner("Executing agent...")
			spinner.Start()
		}

		// Check for nudge file changes (allows user to add nudges mid-run)
		if reloaded, _ := nudgeStore.Reload(); reloaded && cfg.Verbose {
			output.Debug("Nudge file updated, reloaded %d nudge(s)", nudgeStore.ActiveCount())
		}

		// Capture active nudges before this iteration
		activeNudges := nudgeStore.GetActive()

		// Build the prompt for the AI agent, including any recovery guidance
		iterPrompt := prompt.BuildIterationPrompt(cfg)
		
		// Inject memory context (relevant memories based on current feature category)
		// Note: category could be extracted from the plan in a future enhancement
		memoryContext := memStore.BuildPromptContext("", 10) // Get top 10 relevant memories
		if memoryContext != "" {
			iterPrompt = memoryContext + iterPrompt
		}

		// Inject nudge context
		nudgeContext := nudgeStore.BuildPromptContext()
		if nudgeContext != "" {
			iterPrompt = nudgeContext + iterPrompt
		}
		
		if additionalPromptGuidance != "" {
			iterPrompt = additionalPromptGuidance + "\n\n" + iterPrompt
			additionalPromptGuidance = "" // Clear after use
		}

		if cfg.Verbose {
			output.Debug("Prompt: %s", iterPrompt)
		}

		// Execute the AI agent CLI tool
		result, err := agent.Execute(cfg, iterPrompt)
		
		// Stop spinner
		if spinner != nil {
			spinner.Stop()
		}

		// Determine exit code for failure detection
		exitCode := 0
		if err != nil {
			exitCode = 1
			// Don't return immediately - handle with recovery
		}

		// Print the agent output
		if result != "" {
			output.Print("%s", result)
		}

		// Extract and store any memories from the agent output
		memoriesStored := extractAndStoreMemories(memStore, result, "")
		if memoriesStored > 0 && cfg.Verbose {
			output.Debug("Extracted and stored %d new memories from agent output", memoriesStored)
		}

		// Acknowledge nudges that were injected into this iteration
		if len(activeNudges) > 0 {
			if err := nudgeStore.AcknowledgeAll(); err != nil {
				output.Debug("Failed to acknowledge nudges: %v", err)
			} else {
				// Log nudge acknowledgment to progress file
				ackMsg := nudge.FormatAcknowledgment(activeNudges)
				if ackMsg != "" {
					appendProgress(cfg.ProgressFile, ackMsg)
				}
				if cfg.Verbose {
					output.Debug("Acknowledged %d nudge(s)", len(activeNudges))
				}
			}
		}

		// Check for completion signal (even if there was an error, the output might contain it)
		if strings.Contains(result, prompt.CompleteSignal) {
			output.Success("Plan complete! Detected completion signal after %d iteration(s).", i)
			summary.FeaturesCompleted++
			summary.EndTime = time.Now()
			summary.FailuresRecovered = recoveryMgr.GetRecoveredCount()
			output.PrintSummary(summary)
			printRecoverySummaryUI(output, recoveryMgr, cfg.Verbose)
			
			// Show scope summary if scope control was active
			if cfg.ScopeLimit > 0 || cfg.Deadline != "" {
				printScopeSummary(output, scopeMgr, cfg.Verbose)
			}
			
			// Show final milestone status
			if milestoneMgr != nil && milestoneMgr.HasMilestones() {
				output.SubHeader("Final Milestone Status")
				output.Print("%s", milestoneMgr.Summary())
			}
			return nil
		}
		
		// Check for newly completed milestones
		if milestoneMgr != nil && milestoneMgr.HasMilestones() {
			// Reload plans to get updated tested status
			updatedPlans, err := plan.ReadFile(cfg.PlanFile)
			if err == nil {
				milestoneMgr = milestone.NewManager(updatedPlans)
				
				// Check for newly completed milestones
				for _, p := range milestoneMgr.GetCompletedMilestones() {
					if !completedMilestonesBefore[p.Milestone.Name] {
						output.Success("%s", milestone.CelebrationMessage(p.Milestone.Name))
						completedMilestonesBefore[p.Milestone.Name] = true
					}
				}
			}
		}

		// Handle failure detection and recovery
		if err != nil || containsFailureIndicators(result) {
			if exitCode == 0 && containsFailureIndicators(result) {
				exitCode = 1 // Treat as failure even if command succeeded
			}

			failure, recoveryResult := recoveryMgr.HandleFailure(result, exitCode, currentFeatureID, i)
			
			if failure != nil {
				output.Warn("Failure detected: %s", failure)
				summary.Errors = append(summary.Errors, failure.String())
				
				// Track consecutive failures for replanning
				consecutiveFailures++
				
				// Log failure to progress file
				logFailureToProgress(cfg.ProgressFile, failure)

				if recoveryResult.ShouldSkip {
					output.Info("Recovery: %s", recoveryResult.Message)
					summary.FeaturesSkipped++
					// Add to blocked features for replan tracking
					replanMgr.AddBlockedFeature(currentFeatureID)
					// Reset consecutive failures when skipping
					consecutiveFailures = 0
				} else if recoveryResult.ShouldRetry {
					output.Info("Recovery: %s", recoveryResult.Message)
					// Set additional guidance for the retry
					if recoveryResult.ModifiedPrompt != "" {
						additionalPromptGuidance = recoveryResult.ModifiedPrompt
					}
				}

				if !recoveryResult.Success {
					output.Error("Recovery action failed: %s", recoveryResult.Message)
					summary.FeaturesFailed++
				}
				
				// Check for replanning triggers
				replanMgr.UpdateState(currentFeatureID, consecutiveFailures, []string{string(failure.Type)}, plans)
				replanMgr.IncrementIterations()
				
				if shouldReplan, trigger := replanMgr.ShouldReplan(); shouldReplan {
					output.SubHeader("Automatic Replanning Triggered")
					output.Info("Trigger: %s", trigger)
					
					replanResult, replanErr := replanMgr.ExecuteReplan(replanStrategyType, trigger)
					if replanErr != nil {
						output.Error("Replanning failed: %v", replanErr)
					} else if replanResult.Success {
						output.Success("Replanning completed: %s", replanResult.Message)
						if replanResult.OldPlanPath != "" {
							output.Debug("Backup created: %s", replanResult.OldPlanPath)
						}
						if replanResult.Diff != nil && !replanResult.Diff.IsEmpty() {
							output.Print("%s", replanResult.Diff.Summary())
						}
						// Update local plans reference
						plans = replanResult.NewPlans
						// Log replan to progress file
						appendProgress(cfg.ProgressFile, fmt.Sprintf("REPLAN: %s triggered, strategy: %s", trigger, replanStrategyType))
						// Reset consecutive failures after replanning
						consecutiveFailures = 0
					}
				}
			} else if err != nil {
				// Agent execution error but no specific failure detected
				output.Error("Agent execution error: %v", err)
				summary.Errors = append(summary.Errors, err.Error())
				consecutiveFailures++
			}
		} else {
			// Iteration completed without obvious failures
			// Reset consecutive failures on success
			consecutiveFailures = 0
			replanMgr.ResetState()
		}

		output.Print("") // Empty line between iterations
	}

	output.Info("Completed %d iteration(s) without completion signal.", cfg.Iterations)
	summary.EndTime = time.Now()
	summary.FailuresRecovered = recoveryMgr.GetRecoveredCount()
	output.PrintSummary(summary)
	printRecoverySummaryUI(output, recoveryMgr, cfg.Verbose)
	
	// Print scope summary if scope control was active
	if cfg.ScopeLimit > 0 || cfg.Deadline != "" {
		printScopeSummary(output, scopeMgr, cfg.Verbose)
	}
	
	// Print memory summary if we have memories
	if memStore.Count() > 0 && cfg.Verbose {
		output.SubHeader("Memory Status")
		output.Print("Total memories: %d (stored in %s)", memStore.Count(), cfg.MemoryFile)
	}
	
	// Print milestone summary if milestones are defined
	if milestoneMgr != nil && milestoneMgr.HasMilestones() {
		// Reload plans to get updated tested status
		updatedPlans, err := plan.ReadFile(cfg.PlanFile)
		if err == nil {
			milestoneMgr = milestone.NewManager(updatedPlans)
		}
		output.SubHeader("Milestone Progress")
		output.Print("%s", milestoneMgr.Summary())
		
		// Show next milestone to complete
		next := milestoneMgr.GetNextMilestoneToComplete()
		if next != nil {
			output.Info("Next milestone: %s (%s)",
				next.Milestone.Name,
				milestone.FormatProgressBar(next, 20))
		}
	}
	
	return nil
}

// containsFailureIndicators checks if the output contains signs of failure
func containsFailureIndicators(output string) bool {
	outputLower := strings.ToLower(output)
	indicators := []string{
		"fail",
		"error:",
		"panic:",
		"cannot compile",
		"build failed",
		"test failed",
		"assertion failed",
	}
	
	for _, indicator := range indicators {
		if strings.Contains(outputLower, indicator) {
			return true
		}
	}
	return false
}

// logFailureToProgress appends failure information to the progress file
func logFailureToProgress(progressFile string, failure *recovery.Failure) {
	message := fmt.Sprintf("FAILURE [%s]: %s (feature #%d, retry %d)",
		failure.Type, failure.Message, failure.FeatureID, failure.RetryCount)
	
	if err := appendProgress(progressFile, message); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to log failure to progress file: %v\n", err)
	}
}

// printRecoverySummary prints a summary of failures and recovery actions (legacy function)
func printRecoverySummary(rm *recovery.RecoveryManager, verbose bool) {
	summary := rm.GetFailureSummary()
	if summary != "No failures recorded" {
		fmt.Println()
		fmt.Printf("=== Recovery Summary ===\n")
		fmt.Println(summary)
	} else if verbose {
		fmt.Println()
		fmt.Println("No failures encountered during execution.")
	}
}

// printRecoverySummaryUI prints a summary using the UI package
func printRecoverySummaryUI(output *ui.UI, rm *recovery.RecoveryManager, verbose bool) {
	summary := rm.GetFailureSummary()
	if summary != "No failures recorded" {
		output.SubHeader("Recovery Summary")
		output.Print("%s", summary)
	} else if verbose {
		output.Print("")
		output.Success("No failures encountered during execution.")
	}
}

// listPlanStatus displays plan status (tested/untested features)
func listPlanStatus(cfg *config.Config) error {
	plans, err := plan.ReadFile(cfg.PlanFile)
	if err != nil {
		return err
	}

	// Determine what to show
	showTested := cfg.ListStatus || cfg.ListTested
	showUntested := cfg.ListStatus || cfg.ListUntested

	if showTested {
		fmt.Printf("=== Tested Features (from %s) ===\n", cfg.PlanFile)
		tested := plan.Filter(plans, true)
		if len(tested) == 0 {
			fmt.Println("No tested features found")
		} else {
			plan.Print(tested)
		}
		fmt.Println()
	}

	if showUntested {
		fmt.Printf("=== Untested Features (from %s) ===\n", cfg.PlanFile)
		untested := plan.Filter(plans, false)
		if len(untested) == 0 {
			fmt.Println("No untested features found")
		} else {
			plan.Print(untested)
		}
	}

	return nil
}

// listDeferredFeatures displays features that have been deferred due to scope constraints
func listDeferredFeatures(cfg *config.Config) error {
	plans, err := plan.ReadFile(cfg.PlanFile)
	if err != nil {
		return err
	}

	deferred := plan.FilterDeferred(plans, true)

	fmt.Printf("=== Deferred Features (from %s) ===\n", cfg.PlanFile)
	if len(deferred) == 0 {
		fmt.Println("No deferred features found")
		fmt.Println()
		fmt.Println("Features are deferred when they exceed scope constraints:")
		fmt.Println("  - Iteration limit reached (-scope-limit flag)")
		fmt.Println("  - Deadline reached (-deadline flag)")
		fmt.Println()
		fmt.Println("To use scope control, run with:")
		fmt.Printf("  %s -iterations 10 -scope-limit 3  # Max 3 iterations per feature\n", os.Args[0])
		fmt.Printf("  %s -iterations 10 -deadline 1h   # 1 hour time limit\n", os.Args[0])
	} else {
		for _, p := range deferred {
			reason := p.DeferReason
			if reason == "" {
				reason = "unspecified"
			}
			fmt.Printf("  %d. %s [%s] - %s\n", p.ID, p.Category, reason, p.Description)
		}
		fmt.Printf("\nTotal deferred: %d features\n", len(deferred))
	}

	return nil
}

// generatePlanFromNotes generates a plan.json file from notes using the AI agent
func generatePlanFromNotes(cfg *config.Config) error {
	fmt.Printf("Generating plan from notes file: %s\n", cfg.NotesFile)
	fmt.Printf("Output plan file: %s\n", cfg.OutputPlanFile)
	fmt.Printf("Agent command: %s\n\n", cfg.AgentCmd)

	// Resolve absolute paths
	notesPath, err := filepath.Abs(cfg.NotesFile)
	if err != nil {
		notesPath = cfg.NotesFile
	}

	outputPath, err := filepath.Abs(cfg.OutputPlanFile)
	if err != nil {
		outputPath = cfg.OutputPlanFile
	}

	// Build the prompt for plan generation
	genPrompt := prompt.BuildPlanGenerationPrompt(notesPath, outputPath)

	if cfg.Verbose {
		fmt.Printf("Prompt: %s\n\n", genPrompt)
	}

	// Execute the agent
	result, err := agent.Execute(cfg, genPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// The agent should have written the file, but let's verify
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// If file doesn't exist, try to extract JSON from the result and write it
		fmt.Println("Plan file not found, attempting to extract from agent output...")
		if err := plan.ExtractAndWrite(result, outputPath); err != nil {
			return fmt.Errorf("failed to extract plan from output: %w\n\nAgent output:\n%s", err, result)
		}
	}

	fmt.Printf("\n✓ Plan generated successfully: %s\n", outputPath)
	return nil
}

// appendProgress appends a message to the progress file
func appendProgress(path string, message string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open progress file: %w", err)
	}
	defer f.Close()

	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("\n[%s] %s\n", timestamp, message)

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write to progress file: %w", err)
	}

	return nil
}

// handleNudgeCommands processes nudge-related CLI commands
func handleNudgeCommands(cfg *config.Config) error {
	store := nudge.NewStore(cfg.NudgeFile)

	if err := store.Load(); err != nil {
		// If file doesn't exist and we're clearing, that's fine
		if !cfg.ClearNudges {
			return fmt.Errorf("failed to load nudges: %w", err)
		}
	}

	// Handle clear nudges command
	if cfg.ClearNudges {
		if err := store.Clear(); err != nil {
			return fmt.Errorf("failed to clear nudges: %w", err)
		}
		fmt.Printf("Nudges cleared: %s\n", cfg.NudgeFile)
		return nil
	}

	// Handle add nudge command
	if cfg.Nudge != "" {
		parts := strings.SplitN(cfg.Nudge, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid nudge format: expected 'type:content' (e.g., 'focus:Work on feature 5')")
		}

		nudgeType, err := nudge.ParseNudgeType(parts[0])
		if err != nil {
			return err
		}

		n, err := store.Add(nudgeType, parts[1], 0)
		if err != nil {
			return fmt.Errorf("failed to add nudge: %w", err)
		}

		fmt.Printf("Nudge added: [%s] %s\n", strings.ToUpper(string(n.Type)), n.Content)
		return nil
	}

	// Handle show nudges command (default if no other nudge command)
	if cfg.ShowNudges {
		fmt.Println(store.Summary())
		return nil
	}

	return nil
}

// handleMemoryCommands processes memory-related CLI commands
func handleMemoryCommands(cfg *config.Config) error {
	store := memory.NewStore(cfg.MemoryFile)
	store.SetRetentionDays(cfg.MemoryRetention)

	if err := store.Load(); err != nil {
		return fmt.Errorf("failed to load memory: %w", err)
	}

	// Handle clear memory command
	if cfg.ClearMemory {
		if err := store.Clear(); err != nil {
			return fmt.Errorf("failed to clear memory: %w", err)
		}
		fmt.Printf("Memory cleared: %s\n", cfg.MemoryFile)
		return nil
	}

	// Handle add memory command
	if cfg.AddMemory != "" {
		parts := strings.SplitN(cfg.AddMemory, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid add-memory format: expected 'type:content' (e.g., 'decision:Use PostgreSQL')")
		}

		entryType, err := memory.ParseEntryType(parts[0])
		if err != nil {
			return err
		}

		entry, err := store.Add(entryType, parts[1], "", "user")
		if err != nil {
			return fmt.Errorf("failed to add memory: %w", err)
		}

		fmt.Printf("Memory added: [%s] %s\n", strings.ToUpper(string(entry.Type)), entry.Content)
		return nil
	}

	// Handle show memory command (default if no other memory command)
	if cfg.ShowMemory {
		// Prune old memories first
		pruned, _ := store.Prune()
		if pruned > 0 {
			fmt.Printf("Pruned %d expired memories\n\n", pruned)
		}

		fmt.Println(store.Summary())
		return nil
	}

	return nil
}

// extractAndStoreMemories extracts memories from agent output and stores them
func extractAndStoreMemories(store *memory.Store, output, category string) int {
	entries := memory.ExtractFromOutput(output)
	if len(entries) == 0 {
		return 0
	}

	stored := 0
	for _, e := range entries {
		e.Category = category
		_, err := store.Add(e.Type, e.Content, category, "agent")
		if err == nil {
			stored++
		}
	}

	return stored
}

// formatScopeInfo returns a formatted string of scope control settings
func formatScopeInfo(cfg *config.Config) string {
	var parts []string
	if cfg.ScopeLimit > 0 {
		parts = append(parts, fmt.Sprintf("max %d iterations/feature", cfg.ScopeLimit))
	}
	if cfg.Deadline != "" {
		parts = append(parts, fmt.Sprintf("deadline %s", cfg.Deadline))
	}
	if len(parts) == 0 {
		return "unlimited"
	}
	return strings.Join(parts, ", ")
}

// markFeatureDeferred updates the plan file to mark a feature as deferred
func markFeatureDeferred(planFile string, featureID int, reason string) error {
	plans, err := plan.ReadFile(planFile)
	if err != nil {
		return err
	}

	if plan.MarkDeferred(plans, featureID, reason) {
		return plan.WriteFile(planFile, plans)
	}
	return nil
}

// printScopeSummary prints a summary of scope control results
func printScopeSummary(output *ui.UI, scopeMgr *scope.Manager, verbose bool) {
	status := scopeMgr.GetStatus()
	
	if status.DeferredCount > 0 || verbose {
		output.SubHeader("Scope Summary")
		output.Print("Elapsed time: %s", status.ElapsedTime.Round(time.Second))
		
		if status.DeadlineSet {
			if status.DeadlineExceeded {
				output.Warn("Deadline: EXCEEDED")
			} else {
				output.Print("Time remaining: %s", status.RemainingTime.Round(time.Second))
			}
		}
		
		if status.DeferredCount > 0 {
			output.Warn("Deferred features: %d (IDs: %v)", status.DeferredCount, status.DeferredFeatureIDs)
			output.Print("")
			output.Print("Deferred features will remain marked in plan.json.")
			output.Print("Review and un-defer them manually when ready to continue.")
		}
	}
}

// extractCurrentFeatureFromPlans tries to get the current feature being worked on
func extractCurrentFeatureFromPlans(planFile string) (int, int, string) {
	plans, err := plan.ReadFile(planFile)
	if err != nil {
		return 0, 0, ""
	}

	// Find first untested, non-deferred feature
	for _, p := range plans {
		if !p.Tested && !p.Deferred {
			return p.ID, len(p.Steps), p.Description
		}
	}
	return 0, 0, ""
}

// handleReplanCommands processes replan-related CLI commands
func handleReplanCommands(cfg *config.Config) error {
	// Create replan manager
	replanMgr := replan.NewReplanManager(cfg.PlanFile, cfg.AgentCmd, cfg.AutoReplan)

	// Handle list versions command
	if cfg.ListVersions {
		versions := replanMgr.GetVersions()
		fmt.Printf("=== Plan Versions (from %s) ===\n", cfg.PlanFile)
		if len(versions) == 0 {
			fmt.Println("No backup versions found.")
			fmt.Println()
			fmt.Println("Backups are created automatically when:")
			fmt.Println("  - Replanning is triggered")
			fmt.Println("  - Plan.json is modified during execution")
			fmt.Println()
			fmt.Println("To enable automatic replanning:")
			fmt.Printf("  %s -iterations 5 -auto-replan\n", os.Args[0])
		} else {
			for _, v := range versions {
				fmt.Printf("  Version %d: %s (trigger: %s)\n", v.Version, v.Timestamp.Format(time.RFC3339), v.Trigger)
				fmt.Printf("    Path: %s\n", v.Path)
			}
			fmt.Printf("\nTotal: %d version(s)\n", len(versions))
			fmt.Println("\nTo restore a version:")
			fmt.Printf("  %s -restore-version <number>\n", os.Args[0])
		}
		return nil
	}

	// Handle restore version command
	if cfg.RestoreVersion > 0 {
		if err := replanMgr.RestoreVersion(cfg.RestoreVersion); err != nil {
			return fmt.Errorf("failed to restore version %d: %w", cfg.RestoreVersion, err)
		}
		fmt.Printf("Restored plan version %d\n", cfg.RestoreVersion)
		return nil
	}

	// Handle manual replan command
	if cfg.Replan {
		// Load current plans
		plans, err := plan.ReadFile(cfg.PlanFile)
		if err != nil {
			return fmt.Errorf("failed to load plan file: %w", err)
		}

		// Find current feature (first untested, non-deferred)
		currentFeatureID := 0
		for _, p := range plans {
			if !p.Tested && !p.Deferred {
				currentFeatureID = p.ID
				break
			}
		}

		// Update state
		replanMgr.UpdateState(currentFeatureID, 0, nil, plans)

		// Parse strategy
		strategyType, err := replan.ParseStrategyType(cfg.ReplanStrategy)
		if err != nil {
			return err
		}

		fmt.Printf("Manual replanning with strategy: %s\n", strategyType)

		// Execute replanning
		result, err := replanMgr.ManualReplan(strategyType)
		if err != nil {
			return fmt.Errorf("replanning failed: %w", err)
		}

		// Display results
		if result.Success {
			fmt.Println("Replanning successful!")
			if result.OldPlanPath != "" {
				fmt.Printf("Backup created: %s\n", result.OldPlanPath)
			}
			if result.Diff != nil && !result.Diff.IsEmpty() {
				fmt.Println()
				fmt.Println(result.Diff.Summary())
			}
		} else {
			fmt.Printf("Replanning completed: %s\n", result.Message)
		}
		return nil
	}

	return nil
}

// handleValidationCommands processes validation-related CLI commands
func handleValidationCommands(cfg *config.Config) error {
	// Create UI instance
	uiCfg := ui.OutputConfig{
		NoColor:    cfg.NoColor,
		Quiet:      cfg.Quiet,
		JSONOutput: cfg.JSONOutput,
		LogLevel:   ui.ParseLogLevel(cfg.LogLevel),
	}
	output := ui.New(uiCfg)

	// Load plans
	plans, err := plan.ReadFile(cfg.PlanFile)
	if err != nil {
		return fmt.Errorf("failed to load plan file: %w", err)
	}

	// Filter plans to validate
	var plansToValidate []plan.Plan
	if cfg.ValidateFeature > 0 {
		// Validate specific feature
		p := plan.GetByID(plans, cfg.ValidateFeature)
		if p == nil {
			return fmt.Errorf("feature #%d not found", cfg.ValidateFeature)
		}
		plansToValidate = append(plansToValidate, *p)
	} else {
		// Validate all completed features that have validations
		for _, p := range plans {
			if p.Tested && len(p.Validations) > 0 {
				plansToValidate = append(plansToValidate, p)
			}
		}
	}

	if len(plansToValidate) == 0 {
		if cfg.ValidateFeature > 0 {
			output.Info("Feature #%d has no validations defined", cfg.ValidateFeature)
		} else {
			output.Info("No completed features with validations found")
		}
		fmt.Println()
		fmt.Println("To add validations, include a 'validations' array in your plan.json features:")
		fmt.Println(`  {
    "id": 1,
    "description": "API endpoint",
    "tested": true,
    "validations": [
      {"type": "http_get", "url": "http://localhost:8080/health", "expected_status": 200},
      {"type": "cli_command", "command": "curl", "args": ["-s", "http://localhost:8080/version"]}
    ]
  }`)
		return nil
	}

	output.Header("Running Validations")
	output.Info("Features to validate: %d", len(plansToValidate))
	output.Print("")

	// Track overall results
	totalValidations := 0
	totalPassed := 0
	totalFailed := 0
	var allResults []validation.ValidationRunResult

	ctx := context.Background()

	for _, p := range plansToValidate {
		if len(p.Validations) == 0 {
			if cfg.ValidateFeature > 0 {
				output.Warn("Feature #%d has no validations defined", p.ID)
			}
			continue
		}

		output.SubHeader("Feature #%d: %s", p.ID, p.Description)

		// Create validation runner
		runner := validation.NewValidationRunner()

		// Convert plan.ValidationDefinition to validation.ValidationDefinition
		for _, vdef := range p.Validations {
			valDef := validation.ValidationDefinition{
				Type:           validation.ValidationType(vdef.Type),
				URL:            vdef.URL,
				Method:         vdef.Method,
				Body:           vdef.Body,
				Headers:        vdef.Headers,
				ExpectedStatus: vdef.ExpectedStatus,
				ExpectedBody:   vdef.ExpectedBody,
				Command:        vdef.Command,
				Args:           vdef.Args,
				Path:           vdef.Path,
				Pattern:        vdef.Pattern,
				Input:          vdef.Input,
				Timeout:        vdef.Timeout,
				Retries:        vdef.Retries,
				Description:    vdef.Description,
				Options:        vdef.Options,
			}
			if err := runner.AddFromDefinitions([]validation.ValidationDefinition{valDef}); err != nil {
				output.Error("Invalid validation: %v", err)
				continue
			}
		}

		// Run validations
		result := runner.Run(ctx)
		result.FeatureID = p.ID
		result.FeatureName = p.Description

		allResults = append(allResults, result)
		totalValidations += result.TotalCount
		totalPassed += result.PassedCount
		totalFailed += result.FailedCount

		// Display results
		for _, vr := range result.Results {
			if vr.Success {
				output.Success("  %s", vr.Message)
			} else {
				output.Error("  %s", vr.Message)
				if vr.Error != "" && cfg.Verbose {
					output.Debug("    Error: %s", vr.Error)
				}
			}
		}

		output.Print("")
	}

	// Print summary
	output.Header("Validation Summary")
	
	status := "PASSED"
	if totalFailed > 0 {
		status = "FAILED"
	}

	output.Print("Overall: %s", status)
	output.Print("  Total validations: %d", totalValidations)
	output.Print("  Passed: %d", totalPassed)
	output.Print("  Failed: %d", totalFailed)
	output.Print("")

	// Show failed features
	if totalFailed > 0 {
		output.Warn("Failed features:")
		for _, r := range allResults {
			if r.FailedCount > 0 {
				output.Print("  - Feature #%d: %s (%d/%d failed)",
					r.FeatureID, r.FeatureName, r.FailedCount, r.TotalCount)
			}
		}
	}

	// Log validation results to progress file
	summaryMsg := fmt.Sprintf("VALIDATION: %s - %d/%d passed across %d features",
		status, totalPassed, totalValidations, len(plansToValidate))
	appendProgress(cfg.ProgressFile, summaryMsg)

	// Return error if any validations failed
	if totalFailed > 0 {
		return fmt.Errorf("%d validation(s) failed", totalFailed)
	}

	return nil
}

// handleMilestoneCommands processes milestone-related CLI commands
func handleMilestoneCommands(cfg *config.Config) error {
	// Load plans
	plans, err := plan.ReadFile(cfg.PlanFile)
	if err != nil {
		return err
	}

	// Create milestone manager
	mgr := milestone.NewManager(plans)

	// Try to load milestones from a separate file if it exists
	// Check for milestones.json in the same directory as plan file
	milestonesFile := strings.TrimSuffix(cfg.PlanFile, ".json") + "-milestones.json"
	if _, err := os.Stat(milestonesFile); err == nil {
		if err := mgr.LoadMilestones(milestonesFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load milestones file: %v\n", err)
		}
	}

	// Handle -milestones flag (list all milestones with progress)
	if cfg.ListMilestones {
		if !mgr.HasMilestones() {
			fmt.Println("No milestones defined.")
			fmt.Println()
			fmt.Println("To define milestones, either:")
			fmt.Println("  1. Add a 'milestone' field to your plan.json features")
			fmt.Println("  2. Create a milestones.json file with milestone definitions")
			fmt.Println()
			fmt.Println("Example plan.json with milestones:")
			fmt.Println("  [")
			fmt.Println("    {\"id\": 1, \"description\": \"Feature A\", \"milestone\": \"Alpha\"},")
			fmt.Println("    {\"id\": 2, \"description\": \"Feature B\", \"milestone\": \"Alpha\"},")
			fmt.Println("    {\"id\": 3, \"description\": \"Feature C\", \"milestone\": \"Beta\"}")
			fmt.Println("  ]")
			return nil
		}

		fmt.Println(mgr.Summary())

		// Check for completed milestones and show celebration
		completed := mgr.GetCompletedMilestones()
		for _, p := range completed {
			fmt.Printf("\n%s\n", milestone.CelebrationMessage(p.Milestone.Name))
		}

		// Show next milestone to complete
		next := mgr.GetNextMilestoneToComplete()
		if next != nil {
			fmt.Printf("\nNext milestone to complete: %s (%s)\n",
				next.Milestone.Name,
				milestone.FormatProgressBar(next, 20))
		}

		return nil
	}

	// Handle -milestone <name> flag (show features for specific milestone)
	if cfg.ShowMilestone != "" {
		progress := mgr.CalculateProgress(cfg.ShowMilestone)

		if progress.TotalFeatures == 0 {
			return fmt.Errorf("milestone '%s' not found or has no features", cfg.ShowMilestone)
		}

		fmt.Printf("=== Milestone: %s ===\n", progress.Milestone.Name)
		if progress.Milestone.Description != "" {
			fmt.Printf("Description: %s\n", progress.Milestone.Description)
		}
		if progress.Milestone.Criteria != "" {
			fmt.Printf("Success Criteria: %s\n", progress.Milestone.Criteria)
		}
		fmt.Printf("Progress: %s\n", milestone.FormatProgressBar(progress, 30))
		fmt.Printf("Status: %s (%d/%d features complete)\n\n",
			progress.Status, progress.CompletedFeatures, progress.TotalFeatures)

		fmt.Println("Features:")
		for _, f := range progress.Features {
			status := "[ ]"
			if f.Tested {
				status = "[x]"
			}
			fmt.Printf("  %s %d. %s\n", status, f.ID, f.Description)
		}

		// Show celebration if milestone is complete
		if progress.Status == milestone.StatusComplete {
			fmt.Printf("\n%s\n", milestone.CelebrationMessage(progress.Milestone.Name))
		}

		return nil
	}

	return nil
}

// handleGoalCommands processes goal-related CLI commands
func handleGoalCommands(cfg *config.Config) error {
	// Create UI instance
	uiCfg := ui.OutputConfig{
		NoColor:    cfg.NoColor,
		Quiet:      cfg.Quiet,
		JSONOutput: cfg.JSONOutput,
		LogLevel:   ui.ParseLogLevel(cfg.LogLevel),
	}
	output := ui.New(uiCfg)

	// Load existing plans (needed for progress tracking and decomposition)
	var plans []plan.Plan
	if _, err := os.Stat(cfg.PlanFile); err == nil {
		plans, _ = plan.ReadFile(cfg.PlanFile)
	}

	// Create goal manager
	goalMgr := goals.NewManager(plans)
	goalMgr.SetGoalsFile(cfg.GoalsFile)

	// Load existing goals
	if err := goalMgr.LoadGoals(cfg.GoalsFile); err != nil && !os.IsNotExist(err) {
		output.Warn("Failed to load goals: %v", err)
	}

	// Handle -list-goals flag
	if cfg.ListGoals {
		if !goalMgr.HasGoals() {
			fmt.Println("No goals defined.")
			fmt.Println()
			fmt.Println("To add a goal, use:")
			fmt.Printf("  %s -goal \"Add user authentication with OAuth\"\n", os.Args[0])
			fmt.Println()
			fmt.Println("Or create a goals.json file:")
			fmt.Println("  {")
			fmt.Println("    \"goals\": [")
			fmt.Println("      {")
			fmt.Println("        \"id\": \"auth\",")
			fmt.Println("        \"description\": \"Add user authentication\",")
			fmt.Println("        \"priority\": 10")
			fmt.Println("      }")
			fmt.Println("    ]")
			fmt.Println("  }")
			return nil
		}

		fmt.Println(goalMgr.Summary())
		return nil
	}

	// Handle -goal-status flag
	if cfg.GoalStatus {
		if !goalMgr.HasGoals() {
			fmt.Println("No goals defined.")
			return nil
		}

		output.Header("Goal Progress")
		allProgress := goalMgr.CalculateAllProgress()
		
		for _, p := range allProgress {
			if p.TotalPlanItems > 0 {
				output.Print("  %s: %s", p.Goal.Description, goals.FormatProgressBar(p, 20))
			} else {
				statusStr := "pending"
				if p.Status == goals.StatusInProgress {
					statusStr = "in progress"
				} else if p.Status == goals.StatusComplete {
					statusStr = "complete"
				} else if p.Status == goals.StatusBlocked {
					statusStr = "blocked"
				}
				output.Print("  %s: [%s] (no plan items)", p.Goal.Description, statusStr)
			}
		}

		// Show next goal to work on
		next := goalMgr.GetNextGoalToWork()
		if next != nil {
			output.Print("")
			output.Info("Next goal to work on: %s (priority: %d)", next.Description, next.Priority)
		}

		return nil
	}

	// Handle -goal flag (add and decompose a new goal)
	if cfg.Goal != "" {
		output.Header("Adding Goal")
		output.Info("Goal: %s", cfg.Goal)
		output.Info("Priority: %d", cfg.GoalPriority)

		// Create the goal
		goal, err := goalMgr.AddGoalFromDescription(cfg.Goal, cfg.GoalPriority)
		if err != nil {
			return fmt.Errorf("failed to add goal: %w", err)
		}

		// Save goals file
		if err := goalMgr.SaveGoals(); err != nil {
			output.Warn("Failed to save goals file: %v", err)
		}

		output.Success("Goal added with ID: %s", goal.ID)

		// Decompose the goal if we have an agent
		if _, err := exec.LookPath(cfg.AgentCmd); err == nil {
			output.Print("")
			output.SubHeader("Decomposing Goal into Plan Items")

			if err := decomposeGoal(cfg, output, goalMgr, goal); err != nil {
				output.Warn("Failed to decompose goal: %v", err)
				output.Print("You can try again later with: %s -decompose-goal %s", os.Args[0], goal.ID)
			}
		} else {
			output.Print("")
			output.Info("Agent not found. To decompose this goal into plan items, run:")
			output.Print("  %s -decompose-goal %s", os.Args[0], goal.ID)
		}

		return nil
	}

	// Handle -decompose-goal flag
	if cfg.DecomposeGoal != "" {
		goal := goalMgr.GetGoalByID(cfg.DecomposeGoal)
		if goal == nil {
			return fmt.Errorf("goal with ID %q not found", cfg.DecomposeGoal)
		}

		output.Header("Decomposing Goal")
		output.Info("Goal: %s", goal.Description)

		if err := decomposeGoal(cfg, output, goalMgr, goal); err != nil {
			return err
		}

		return nil
	}

	// Handle -decompose-all flag
	if cfg.DecomposeAll {
		pendingGoals := goalMgr.GetPendingGoals()
		if len(pendingGoals) == 0 {
			output.Info("No pending goals to decompose")
			return nil
		}

		output.Header("Decomposing All Pending Goals")
		output.Info("Goals to decompose: %d", len(pendingGoals))

		for _, goal := range pendingGoals {
			output.SubHeader("Goal: %s", goal.Description)
			goalRef := goalMgr.GetGoalByID(goal.ID) // Get pointer
			if err := decomposeGoal(cfg, output, goalMgr, goalRef); err != nil {
				output.Error("Failed to decompose goal %q: %v", goal.ID, err)
				continue
			}
			output.Print("")
		}

		return nil
	}

	return nil
}

// decomposeGoal decomposes a single goal into plan items using the AI agent
func decomposeGoal(cfg *config.Config, output *ui.UI, goalMgr *goals.Manager, goal *goals.Goal) error {
	// Load current plans
	var existingPlans []plan.Plan
	if _, err := os.Stat(cfg.PlanFile); err == nil {
		existingPlans, _ = plan.ReadFile(cfg.PlanFile)
	}

	// Get absolute path for output
	outputPath, err := filepath.Abs(cfg.PlanFile)
	if err != nil {
		outputPath = cfg.PlanFile
	}

	// Build the decomposition prompt
	decomposePrompt := goals.BuildGoalDecompositionPrompt(goal, existingPlans, outputPath)

	if cfg.Verbose {
		output.Debug("Prompt: %s", decomposePrompt)
	}

	// Show spinner during agent execution
	var spinner *ui.Spinner
	if output.IsTTY() && !cfg.Quiet && !cfg.JSONOutput {
		spinner = output.NewSpinner("Decomposing goal with AI agent...")
		spinner.Start()
	}

	// Execute the agent
	result, err := agent.Execute(cfg, decomposePrompt)

	if spinner != nil {
		spinner.Stop()
	}

	if err != nil {
		return fmt.Errorf("agent execution failed: %w", err)
	}

	// Parse the result
	decompResult, err := goals.ParseDecompositionResult(result, goal)
	if err != nil {
		output.Debug("Raw agent output: %s", result)
		return fmt.Errorf("failed to parse decomposition result: %w", err)
	}

	if !decompResult.Success || len(decompResult.GeneratedPlans) == 0 {
		// Try to extract from file if agent wrote directly
		if _, err := os.Stat(outputPath); err == nil {
			updatedPlans, readErr := plan.ReadFile(outputPath)
			if readErr == nil && len(updatedPlans) > len(existingPlans) {
				// Plans were written directly
				newCount := len(updatedPlans) - len(existingPlans)
				output.Success("Generated %d plan items (written directly by agent)", newCount)
				
				// Link new plan IDs to the goal
				for i := len(existingPlans); i < len(updatedPlans); i++ {
					goalMgr.LinkPlanToGoal(goal.ID, updatedPlans[i].ID)
				}
				
				// Update goal status
				goal.Status = goals.StatusInProgress
				goalMgr.UpdateGoal(*goal)
				goalMgr.SaveGoals()
				
				return nil
			}
		}
		
		output.Debug("Raw agent output: %s", result)
		return fmt.Errorf("decomposition produced no plan items: %s", decompResult.Message)
	}

	// Merge with existing plans
	mergedPlans := goals.MergePlans(existingPlans, decompResult.GeneratedPlans)

	// Write the merged plan file
	if err := plan.WriteFile(cfg.PlanFile, mergedPlans); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	// Link generated plans to the goal
	for _, p := range decompResult.GeneratedPlans {
		goalMgr.LinkPlanToGoal(goal.ID, p.ID)
	}

	// Update goal status
	goal.Status = goals.StatusInProgress
	goalMgr.UpdateGoal(*goal)
	goalMgr.SaveGoals()

	output.Success("Generated %d plan items", len(decompResult.GeneratedPlans))
	
	// Print generated plan items
	output.Print("")
	output.Print("Generated plan items:")
	for _, p := range decompResult.GeneratedPlans {
		output.Print("  %d. [%s] %s", p.ID, p.Category, p.Description)
	}

	// Log to progress file
	progressMsg := fmt.Sprintf("GOAL DECOMPOSED: %q -> %d plan items (IDs: %v)",
		goal.Description, len(decompResult.GeneratedPlans), getIDs(decompResult.GeneratedPlans))
	appendProgress(cfg.ProgressFile, progressMsg)

	return nil
}

// getIDs extracts IDs from a slice of plans
func getIDs(plans []plan.Plan) []int {
	ids := make([]int, len(plans))
	for i, p := range plans {
		ids[i] = p.ID
	}
	return ids
}
