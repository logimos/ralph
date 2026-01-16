// Package config provides configuration management for Ralph.
package config

const (
	// DefaultPlanFile is the default path for the plan file
	DefaultPlanFile = "plan.json"
	// DefaultProgressFile is the default path for the progress file
	DefaultProgressFile = "progress.txt"
	// DefaultAgentCmd is the default AI agent command
	DefaultAgentCmd = "cursor-agent"
)

// Config holds the application configuration
type Config struct {
	PlanFile       string
	ProgressFile   string
	Iterations     int
	AgentCmd       string
	TypeCheckCmd   string
	TestCmd        string
	BuildSystem    string
	Verbose        bool
	ShowVersion    bool
	ListStatus     bool
	ListTested     bool
	ListUntested   bool
	GeneratePlan   bool
	NotesFile      string
	OutputPlanFile string
	ConfigFile     string // Path to config file (if specified via -config flag)
}

// New creates a new Config with default values
func New() *Config {
	return &Config{
		PlanFile:       DefaultPlanFile,
		ProgressFile:   DefaultProgressFile,
		AgentCmd:       DefaultAgentCmd,
		OutputPlanFile: DefaultPlanFile,
	}
}
