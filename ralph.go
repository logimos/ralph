package main

import (
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
	"github.com/logimos/ralph/internal/plan"
	"github.com/logimos/ralph/internal/prompt"
	"github.com/logimos/ralph/internal/recovery"
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

	// Handle list commands (don't require iterations)
	if cfg.ListStatus || cfg.ListTested || cfg.ListUntested {
		if err := validateConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := listPlanStatus(cfg); err != nil {
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
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -version                         # Show version information\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 5                    # Run 5 iterations (auto-detect build system)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 5 -build-system gradle  # Use Gradle preset\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -config my-config.yaml           # Use specific config file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -status                          # Show plan status\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-tested                     # List tested features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-untested                   # List untested features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -generate-plan -notes notes.md   # Generate plan.json from notes\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -generate-plan -notes notes.md -output my-plan.json  # Custom output file\n", os.Args[0])
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

	// Skip iteration validation if we're just listing status
	if cfg.ListStatus || cfg.ListTested || cfg.ListUntested {
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

	return nil
}

func runIterations(cfg *config.Config) error {
	fmt.Printf("Starting iterative development workflow\n")
	fmt.Printf("Plan file: %s\n", cfg.PlanFile)
	fmt.Printf("Progress file: %s\n", cfg.ProgressFile)
	fmt.Printf("Iterations: %d\n", cfg.Iterations)
	fmt.Printf("Agent command: %s\n", cfg.AgentCmd)
	fmt.Printf("Recovery strategy: %s (max %d retries)\n", cfg.RecoveryStrategy, cfg.MaxRetries)
	if cfg.Verbose {
		fmt.Printf("Type check command: %s\n", cfg.TypeCheckCmd)
		fmt.Printf("Test command: %s\n", cfg.TestCmd)
	}
	fmt.Println()

	// Initialize recovery manager
	strategyType, _ := recovery.ParseStrategyType(cfg.RecoveryStrategy)
	recoveryMgr := recovery.NewRecoveryManager(cfg.MaxRetries, strategyType)

	// Track the current feature being worked on (extracted from output if possible)
	currentFeatureID := 0
	var additionalPromptGuidance string

	for i := 1; i <= cfg.Iterations; i++ {
		fmt.Printf("=== Iteration %d/%d ===\n", i, cfg.Iterations)

		if cfg.Verbose {
			fmt.Printf("Executing agent command...\n")
		}

		// Build the prompt for the AI agent, including any recovery guidance
		iterPrompt := prompt.BuildIterationPrompt(cfg)
		if additionalPromptGuidance != "" {
			iterPrompt = additionalPromptGuidance + "\n\n" + iterPrompt
			additionalPromptGuidance = "" // Clear after use
		}

		if cfg.Verbose {
			fmt.Printf("Prompt: %s\n", iterPrompt)
		}

		// Execute the AI agent CLI tool
		result, err := agent.Execute(cfg, iterPrompt)
		
		// Determine exit code for failure detection
		exitCode := 0
		if err != nil {
			exitCode = 1
			// Don't return immediately - handle with recovery
		}

		// Print the agent output
		if result != "" {
			fmt.Println(result)
		}

		// Check for completion signal (even if there was an error, the output might contain it)
		if strings.Contains(result, prompt.CompleteSignal) {
			fmt.Printf("\n✓ Plan complete! Detected completion signal after %d iteration(s).\n", i)
			printRecoverySummary(recoveryMgr, cfg.Verbose)
			return nil
		}

		// Handle failure detection and recovery
		if err != nil || containsFailureIndicators(result) {
			if exitCode == 0 && containsFailureIndicators(result) {
				exitCode = 1 // Treat as failure even if command succeeded
			}

			failure, recoveryResult := recoveryMgr.HandleFailure(result, exitCode, currentFeatureID, i)
			
			if failure != nil {
				fmt.Printf("\n⚠ Failure detected: %s\n", failure)
				
				// Log failure to progress file
				logFailureToProgress(cfg.ProgressFile, failure)

				if recoveryResult.ShouldSkip {
					fmt.Printf("→ Recovery: %s\n", recoveryResult.Message)
					// Continue to next iteration (which will work on next feature)
				} else if recoveryResult.ShouldRetry {
					fmt.Printf("→ Recovery: %s\n", recoveryResult.Message)
					// Set additional guidance for the retry
					if recoveryResult.ModifiedPrompt != "" {
						additionalPromptGuidance = recoveryResult.ModifiedPrompt
					}
				}

				if !recoveryResult.Success {
					fmt.Printf("→ Recovery action failed: %s\n", recoveryResult.Message)
				}
			} else if err != nil {
				// Agent execution error but no specific failure detected
				fmt.Printf("\n⚠ Agent execution error: %v\n", err)
			}
		}

		fmt.Println() // Empty line between iterations
	}

	fmt.Printf("Completed %d iteration(s) without completion signal.\n", cfg.Iterations)
	printRecoverySummary(recoveryMgr, cfg.Verbose)
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

// printRecoverySummary prints a summary of failures and recovery actions
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
