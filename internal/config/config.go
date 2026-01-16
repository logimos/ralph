// Package config provides configuration management for Ralph.
package config

const (
	// DefaultPlanFile is the default path for the plan file
	DefaultPlanFile = "plan.json"
	// DefaultProgressFile is the default path for the progress file
	DefaultProgressFile = "progress.txt"
	// DefaultAgentCmd is the default AI agent command
	DefaultAgentCmd = "cursor-agent"
	// DefaultMaxRetries is the default maximum retries per feature before escalation
	DefaultMaxRetries = 3
	// DefaultRecoveryStrategy is the default recovery strategy
	DefaultRecoveryStrategy = "retry"
	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"
	// DefaultMemoryFile is the default path for the memory file
	DefaultMemoryFile = ".ralph-memory.json"
	// DefaultMemoryRetention is the default number of days to retain memories
	DefaultMemoryRetention = 90
	// DefaultNudgeFile is the default path for the nudge file
	DefaultNudgeFile = "nudges.json"
	// DefaultScopeLimit is the default max iterations per feature (0 = unlimited)
	DefaultScopeLimit = 0
	// DefaultAutoReplan controls whether replanning happens automatically
	DefaultAutoReplan = false
	// DefaultReplanStrategy is the default replanning strategy
	DefaultReplanStrategy = "incremental"
	// DefaultReplanThreshold is the default number of consecutive failures before replanning
	DefaultReplanThreshold = 3
	// DefaultGoalsFile is the default path for the goals file
	DefaultGoalsFile = "goals.json"
)

// Config holds the application configuration
type Config struct {
	PlanFile         string
	ProgressFile     string
	Iterations       int
	AgentCmd         string
	TypeCheckCmd     string
	TestCmd          string
	BuildSystem      string
	Verbose          bool
	ShowVersion      bool
	ListStatus       bool
	ListTested       bool
	ListUntested     bool
	GeneratePlan     bool
	NotesFile        string
	OutputPlanFile   string
	ConfigFile       string // Path to config file (if specified via -config flag)
	MaxRetries       int    // Maximum retries per feature before recovery escalation
	RecoveryStrategy string // Recovery strategy: retry, skip, rollback
	Environment      string // Environment override (local, github-actions, gitlab-ci, etc.)
	// UI-related configuration
	NoColor    bool   // Disable colored output
	Quiet      bool   // Minimal output (errors only)
	JSONOutput bool   // Machine-readable JSON output
	LogLevel   string // Log level: debug, info, warn, error
	// Memory-related configuration
	MemoryFile      string // Path to memory file (default: .ralph-memory.json)
	ShowMemory      bool   // Display stored memories
	ClearMemory     bool   // Clear all memories
	AddMemory       string // Add a manual memory entry (format: "type:content")
	MemoryRetention int    // Number of days to retain memories (default: 90)
	// Milestone-related configuration
	ListMilestones  bool   // List all milestones with progress
	ShowMilestone   string // Show features for a specific milestone
	// Nudge-related configuration
	NudgeFile    string // Path to nudge file (default: nudges.json)
	Nudge        string // One-time inline nudge (format: "type:content")
	ClearNudges  bool   // Clear all nudges
	ShowNudges   bool   // Display current nudges
	// Scope control configuration
	ScopeLimit   int    // Max iterations per feature (0 = unlimited)
	Deadline     string // Deadline duration (e.g., "1h", "30m", "2h30m")
	ListDeferred bool   // List deferred features
	// Replanning configuration
	AutoReplan       bool   // Enable automatic replanning when triggers fire
	Replan           bool   // Manually trigger replanning
	ReplanStrategy   string // Replanning strategy: incremental, agent
	ReplanThreshold  int    // Number of consecutive failures before replanning
	ListVersions     bool   // List plan versions
	RestoreVersion   int    // Restore a specific plan version
	// Validation configuration
	Validate        bool // Run validations for all completed features
	ValidateFeature int  // Validate a specific feature by ID
	// Goal-oriented configuration
	GoalsFile     string // Path to goals file (default: goals.json)
	Goal          string // Single goal to add and decompose
	GoalPriority  int    // Priority for the goal (when using -goal)
	GoalStatus    bool   // Show status of all goals
	ListGoals     bool   // List all goals
	DecomposeGoal string // Decompose a specific goal by ID
	DecomposeAll  bool   // Decompose all pending goals
}

// New creates a new Config with default values
func New() *Config {
	return &Config{
		PlanFile:         DefaultPlanFile,
		ProgressFile:     DefaultProgressFile,
		AgentCmd:         DefaultAgentCmd,
		OutputPlanFile:   DefaultPlanFile,
		MaxRetries:       DefaultMaxRetries,
		RecoveryStrategy: DefaultRecoveryStrategy,
		LogLevel:         DefaultLogLevel,
		MemoryFile:       DefaultMemoryFile,
		MemoryRetention:  DefaultMemoryRetention,
		NudgeFile:        DefaultNudgeFile,
		ScopeLimit:       DefaultScopeLimit,
		AutoReplan:       DefaultAutoReplan,
		ReplanStrategy:   DefaultReplanStrategy,
		ReplanThreshold:  DefaultReplanThreshold,
		GoalsFile:        DefaultGoalsFile,
	}
}
