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
	}
}
