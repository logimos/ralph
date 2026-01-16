// Package config provides configuration management for Ralph.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigFileNames lists the supported configuration file names in order of precedence
var ConfigFileNames = []string{
	".ralph.yaml",
	".ralph.yml",
	".ralph.json",
	"ralph.config.yaml",
	"ralph.config.yml",
	"ralph.config.json",
}

// FileConfig represents the configuration file structure.
// Fields use pointers to distinguish between "not set" and "set to zero/empty value".
type FileConfig struct {
	// Agent configuration
	Agent string `json:"agent,omitempty" yaml:"agent,omitempty"`

	// Build system preset (pnpm, npm, yarn, gradle, maven, cargo, go, python, auto)
	BuildSystem string `json:"build_system,omitempty" yaml:"build_system,omitempty"`

	// Custom commands (override build system preset)
	TypeCheck string `json:"typecheck,omitempty" yaml:"typecheck,omitempty"`
	Test      string `json:"test,omitempty" yaml:"test,omitempty"`

	// File paths
	Plan     string `json:"plan,omitempty" yaml:"plan,omitempty"`
	Progress string `json:"progress,omitempty" yaml:"progress,omitempty"`

	// Execution settings
	Iterations int  `json:"iterations,omitempty" yaml:"iterations,omitempty"`
	Verbose    bool `json:"verbose,omitempty" yaml:"verbose,omitempty"`

	// Recovery settings
	MaxRetries       int    `json:"max_retries,omitempty" yaml:"max_retries,omitempty"`
	RecoveryStrategy string `json:"recovery_strategy,omitempty" yaml:"recovery_strategy,omitempty"`

	// Environment settings
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`

	// UI settings
	NoColor    bool   `json:"no_color,omitempty" yaml:"no_color,omitempty"`
	Quiet      bool   `json:"quiet,omitempty" yaml:"quiet,omitempty"`
	JSONOutput bool   `json:"json_output,omitempty" yaml:"json_output,omitempty"`
	LogLevel   string `json:"log_level,omitempty" yaml:"log_level,omitempty"`

	// Memory settings
	MemoryFile      string `json:"memory_file,omitempty" yaml:"memory_file,omitempty"`
	MemoryRetention int    `json:"memory_retention,omitempty" yaml:"memory_retention,omitempty"`

	// Nudge settings
	NudgeFile string `json:"nudge_file,omitempty" yaml:"nudge_file,omitempty"`

	// Scope control settings
	ScopeLimit int    `json:"scope_limit,omitempty" yaml:"scope_limit,omitempty"` // Max iterations per feature
	Deadline   string `json:"deadline,omitempty" yaml:"deadline,omitempty"`       // Deadline duration (e.g., "1h", "30m")
}

// DiscoverConfigFile searches for a configuration file in the current directory
// and then in the user's home directory. Returns the path to the first file found,
// or empty string if no config file exists.
func DiscoverConfigFile() string {
	// First, check current directory
	cwd, err := os.Getwd()
	if err == nil {
		if path := findConfigInDir(cwd); path != "" {
			return path
		}
	}

	// Then, check home directory
	home, err := os.UserHomeDir()
	if err == nil && home != cwd {
		if path := findConfigInDir(home); path != "" {
			return path
		}
	}

	return ""
}

// findConfigInDir looks for a config file in the specified directory.
// Returns the full path if found, empty string otherwise.
func findConfigInDir(dir string) string {
	for _, name := range ConfigFileNames {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// LoadConfigFile loads and parses a configuration file.
// Supports both YAML and JSON formats based on file extension.
func LoadConfigFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(data) == 0 {
		return &FileConfig{}, nil
	}

	cfg := &FileConfig{}
	ext := filepath.Ext(path)

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config file: %w", err)
		}
	default:
		// Try YAML first (superset of JSON), then JSON
		if err := yaml.Unmarshal(data, cfg); err != nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file (tried YAML and JSON): %w", err)
			}
		}
	}

	return cfg, nil
}

// ValidateFileConfig validates the configuration file contents.
func ValidateFileConfig(cfg *FileConfig) error {
	// Validate build system if specified
	validBuildSystems := map[string]bool{
		"":       true, // empty is valid (use default/auto)
		"auto":   true,
		"pnpm":   true,
		"npm":    true,
		"yarn":   true,
		"gradle": true,
		"maven":  true,
		"cargo":  true,
		"go":     true,
		"python": true,
	}

	if !validBuildSystems[cfg.BuildSystem] {
		return fmt.Errorf("invalid build_system %q: must be one of pnpm, npm, yarn, gradle, maven, cargo, go, python, or auto", cfg.BuildSystem)
	}

	// Validate iterations if specified
	if cfg.Iterations < 0 {
		return fmt.Errorf("iterations cannot be negative")
	}

	// Validate max retries if specified
	if cfg.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	// Validate recovery strategy if specified
	validStrategies := map[string]bool{
		"":         true, // empty is valid (use default)
		"retry":    true,
		"skip":     true,
		"rollback": true,
	}

	if !validStrategies[cfg.RecoveryStrategy] {
		return fmt.Errorf("invalid recovery_strategy %q: must be one of retry, skip, or rollback", cfg.RecoveryStrategy)
	}

	// Validate environment if specified
	validEnvironments := map[string]bool{
		"":               true, // empty is valid (auto-detect)
		"local":          true,
		"github-actions": true,
		"github":         true,
		"gh":             true,
		"gitlab-ci":      true,
		"gitlab":         true,
		"gl":             true,
		"jenkins":        true,
		"circleci":       true,
		"circle":         true,
		"travis-ci":      true,
		"travis":         true,
		"azure-devops":   true,
		"azure":          true,
		"ci":             true,
	}

	if !validEnvironments[cfg.Environment] {
		return fmt.Errorf("invalid environment %q: must be one of local, github-actions, gitlab-ci, jenkins, circleci, travis-ci, azure-devops, or ci", cfg.Environment)
	}

	// Validate log level if specified
	validLogLevels := map[string]bool{
		"":      true, // empty is valid (use default)
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"quiet": true,
	}

	if !validLogLevels[cfg.LogLevel] {
		return fmt.Errorf("invalid log_level %q: must be one of debug, info, warn, error, or quiet", cfg.LogLevel)
	}

	// Validate memory retention if specified
	if cfg.MemoryRetention < 0 {
		return fmt.Errorf("memory_retention cannot be negative")
	}

	// Validate scope limit if specified
	if cfg.ScopeLimit < 0 {
		return fmt.Errorf("scope_limit cannot be negative")
	}

	// Validate deadline format if specified
	if cfg.Deadline != "" {
		if _, err := parseDuration(cfg.Deadline); err != nil {
			return fmt.Errorf("invalid deadline format %q: %w", cfg.Deadline, err)
		}
	}

	return nil
}

// ApplyFileConfig applies file configuration to a Config struct.
// File config values are applied only if the corresponding Config field
// hasn't been explicitly set (i.e., is still at its default value).
// This ensures CLI flags take precedence over file config.
func ApplyFileConfig(cfg *Config, fileCfg *FileConfig) {
	// Apply agent command
	if fileCfg.Agent != "" && cfg.AgentCmd == DefaultAgentCmd {
		cfg.AgentCmd = fileCfg.Agent
	}

	// Apply build system
	if fileCfg.BuildSystem != "" && cfg.BuildSystem == "" {
		cfg.BuildSystem = fileCfg.BuildSystem
	}

	// Apply custom commands
	if fileCfg.TypeCheck != "" && cfg.TypeCheckCmd == "" {
		cfg.TypeCheckCmd = fileCfg.TypeCheck
	}
	if fileCfg.Test != "" && cfg.TestCmd == "" {
		cfg.TestCmd = fileCfg.Test
	}

	// Apply file paths
	if fileCfg.Plan != "" && cfg.PlanFile == DefaultPlanFile {
		cfg.PlanFile = fileCfg.Plan
	}
	if fileCfg.Progress != "" && cfg.ProgressFile == DefaultProgressFile {
		cfg.ProgressFile = fileCfg.Progress
	}

	// Apply execution settings
	if fileCfg.Iterations > 0 && cfg.Iterations == 0 {
		cfg.Iterations = fileCfg.Iterations
	}
	if fileCfg.Verbose && !cfg.Verbose {
		cfg.Verbose = fileCfg.Verbose
	}

	// Apply recovery settings
	if fileCfg.MaxRetries > 0 && cfg.MaxRetries == DefaultMaxRetries {
		cfg.MaxRetries = fileCfg.MaxRetries
	}
	if fileCfg.RecoveryStrategy != "" && cfg.RecoveryStrategy == DefaultRecoveryStrategy {
		cfg.RecoveryStrategy = fileCfg.RecoveryStrategy
	}

	// Apply environment setting
	if fileCfg.Environment != "" && cfg.Environment == "" {
		cfg.Environment = fileCfg.Environment
	}

	// Apply UI settings
	if fileCfg.NoColor && !cfg.NoColor {
		cfg.NoColor = fileCfg.NoColor
	}
	if fileCfg.Quiet && !cfg.Quiet {
		cfg.Quiet = fileCfg.Quiet
	}
	if fileCfg.JSONOutput && !cfg.JSONOutput {
		cfg.JSONOutput = fileCfg.JSONOutput
	}
	if fileCfg.LogLevel != "" && cfg.LogLevel == DefaultLogLevel {
		cfg.LogLevel = fileCfg.LogLevel
	}

	// Apply memory settings
	if fileCfg.MemoryFile != "" && cfg.MemoryFile == DefaultMemoryFile {
		cfg.MemoryFile = fileCfg.MemoryFile
	}
	if fileCfg.MemoryRetention > 0 && cfg.MemoryRetention == DefaultMemoryRetention {
		cfg.MemoryRetention = fileCfg.MemoryRetention
	}

	// Apply nudge settings
	if fileCfg.NudgeFile != "" && cfg.NudgeFile == DefaultNudgeFile {
		cfg.NudgeFile = fileCfg.NudgeFile
	}

	// Apply scope control settings
	if fileCfg.ScopeLimit > 0 && cfg.ScopeLimit == DefaultScopeLimit {
		cfg.ScopeLimit = fileCfg.ScopeLimit
	}
	if fileCfg.Deadline != "" && cfg.Deadline == "" {
		cfg.Deadline = fileCfg.Deadline
	}
}

// parseDuration parses a duration string like "1h", "30m", "2h30m"
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// ParseDeadline parses a deadline string and returns the deadline time
func ParseDeadline(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	d, err := parseDuration(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Now().Add(d), nil
}
