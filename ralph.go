package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// Version is set at build time via ldflags, or detected at runtime
	// Use GetVersion() to access the version, which handles runtime detection
	Version = "dev"
)

var (
	versionOnce     sync.Once
	detectedVersion string
)

// GetVersion returns the version, detecting it at runtime if not set via ldflags
func GetVersion() string {
	return getVersion()
}

// getVersion returns the version, with lazy detection if needed
func getVersion() string {
	versionOnce.Do(func() {
		// If Version was set via ldflags (not "dev"), use it directly
		if Version != "dev" {
			detectedVersion = Version
			return
		}
		// Otherwise, try to detect it at runtime
		detectedVersion = detectVersion()
	})
	return detectedVersion
}

// detectVersion attempts to detect the version at runtime
func detectVersion() string {
	// Try to get version from go list (works for go install @version)
	if goVersion := getGoListVersion(); goVersion != "" {
		return goVersion
	}

	// Try git describe (works for local builds and source installs)
	if gitVersion := getGitVersion(); gitVersion != "" {
		return gitVersion
	}

	// Return "dev" as fallback
	return "dev"
}

// getGoListVersion attempts to get version using `go list -m`
func getGoListVersion() string {
	// Try to get the version of the installed module
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Version}}", "github.com/logimos/ralph")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	version := strings.TrimSpace(string(output))
	// Only return if it looks like a version tag
	if strings.HasPrefix(version, "v") && len(version) > 1 && version != "(devel)" {
		return version
	}
	return ""
}

// getGitVersion attempts to get version from git tags
func getGitVersion() string {
	// Try to find the git root by checking current directory and parents
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up the directory tree to find .git
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			// Found git repo, try to get version
			cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
			cmd.Dir = dir
			output, err := cmd.Output()
			if err != nil {
				return ""
			}
			version := strings.TrimSpace(string(output))
			// Strip any suffix like -dirty, -1-g1234567, etc.
			if idx := strings.Index(version, "-"); idx != -1 {
				version = version[:idx]
			}
			// Only return if it looks like a version tag
			if strings.HasPrefix(version, "v") && len(version) > 1 {
				return version
			}
			return ""
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}

	return ""
}

const (
	completeSignal      = "<promise>COMPLETE</promise>"
	defaultPlanFile     = "plan.json"
	defaultProgressFile = "progress.txt"
)

// detectBuildSystem attempts to detect the build system from project files
func detectBuildSystem() string {
	// Check for Gradle
	if _, err := os.Stat("build.gradle"); err == nil {
		return "gradle"
	}
	if _, err := os.Stat("build.gradle.kts"); err == nil {
		return "gradle"
	}
	if _, err := os.Stat("gradlew"); err == nil {
		return "gradle"
	}

	// Check for Maven
	if _, err := os.Stat("pom.xml"); err == nil {
		return "maven"
	}

	// Check for Cargo (Rust)
	if _, err := os.Stat("Cargo.toml"); err == nil {
		return "cargo"
	}

	// Check for Go modules
	if _, err := os.Stat("go.mod"); err == nil {
		return "go"
	}

	// Check for Python (common indicators)
	if _, err := os.Stat("setup.py"); err == nil {
		return "python"
	}
	if _, err := os.Stat("pyproject.toml"); err == nil {
		return "python"
	}
	if _, err := os.Stat("requirements.txt"); err == nil {
		return "python"
	}

	// Check for pnpm (has pnpm-lock.yaml)
	if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		return "pnpm"
	}

	// Check for yarn (has yarn.lock)
	if _, err := os.Stat("yarn.lock"); err == nil {
		return "yarn"
	}

	// Check for npm (has package.json, but no lock file means npm)
	if _, err := os.Stat("package.json"); err == nil {
		return "npm"
	}

	// Default to pnpm for backward compatibility
	return "pnpm"
}

// applyBuildSystemConfig applies build system presets or auto-detection
func applyBuildSystemConfig(config *Config) {
	// If both typecheck and test are explicitly set, don't override
	if config.TypeCheckCmd != "" && config.TestCmd != "" {
		return
	}

	var buildSystem string

	// Determine which build system to use
	if config.BuildSystem != "" {
		if config.BuildSystem == "auto" {
			buildSystem = detectBuildSystem()
			if config.Verbose {
				fmt.Printf("Auto-detected build system: %s\n", buildSystem)
			}
		} else {
			buildSystem = config.BuildSystem
		}
	} else {
		// Auto-detect if neither build-system nor individual commands are set
		if config.TypeCheckCmd == "" && config.TestCmd == "" {
			buildSystem = detectBuildSystem()
			if config.Verbose {
				fmt.Printf("Auto-detected build system: %s\n", buildSystem)
			}
		} else {
			// Use defaults if only one command is set
			buildSystem = "pnpm"
		}
	}

	// Apply preset if available
	if preset, ok := BuildSystemPresets[buildSystem]; ok {
		if config.TypeCheckCmd == "" {
			config.TypeCheckCmd = preset.TypeCheck
		}
		if config.TestCmd == "" {
			config.TestCmd = preset.Test
		}
	} else {
		// Unknown build system, use defaults
		if config.TypeCheckCmd == "" {
			config.TypeCheckCmd = BuildSystemPresets["pnpm"].TypeCheck
		}
		if config.TestCmd == "" {
			config.TestCmd = BuildSystemPresets["pnpm"].Test
		}
		if config.Verbose {
			fmt.Printf("Warning: Unknown build system '%s', using pnpm defaults\n", buildSystem)
		}
	}
}

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
}

// BuildSystemPresets defines commands for common build systems
var BuildSystemPresets = map[string]struct {
	TypeCheck string
	Test      string
}{
	"pnpm": {
		TypeCheck: "pnpm typecheck",
		Test:      "pnpm test",
	},
	"npm": {
		TypeCheck: "npm run typecheck",
		Test:      "npm test",
	},
	"yarn": {
		TypeCheck: "yarn typecheck",
		Test:      "yarn test",
	},
	"gradle": {
		TypeCheck: "./gradlew check",
		Test:      "./gradlew test",
	},
	"maven": {
		TypeCheck: "mvn compile",
		Test:      "mvn test",
	},
	"cargo": {
		TypeCheck: "cargo check",
		Test:      "cargo test",
	},
	"go": {
		TypeCheck: "go build ./...",
		Test:      "go test ./...",
	},
	"python": {
		TypeCheck: "mypy .",
		Test:      "pytest",
	},
}

// Plan represents the structure of a plan file
type Plan struct {
	ID             int      `json:"id"`
	Category       string   `json:"category,omitempty"`
	Command        string   `json:"command,omitempty"`
	Description    string   `json:"description"`
	Steps          []string `json:"steps,omitempty"`
	ExpectedOutput string   `json:"expected_output,omitempty"`
	Tested         bool     `json:"tested,omitempty"`
}

func main() {
	config := parseFlags()

	// Handle version command (exit early)
	if config.ShowVersion {
		fmt.Printf("ralph version %s\n", getVersion())
		os.Exit(0)
	}

	// Handle generate-plan command
	if config.GeneratePlan {
		if err := validateConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := generatePlanFromNotes(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle list commands (don't require iterations)
	if config.ListStatus || config.ListTested || config.ListUntested {
		if err := validateConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := listPlanStatus(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := runIterations(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *Config {
	config := &Config{
		AgentCmd:     "cursor-agent",
		TypeCheckCmd: "pnpm typecheck",
		TestCmd:      "pnpm test",
	}

	flag.StringVar(&config.PlanFile, "plan", defaultPlanFile, "Path to the plan file (e.g., plan.json)")
	flag.StringVar(&config.ProgressFile, "progress", defaultProgressFile, "Path to the progress file (e.g., progress.txt)")
	flag.IntVar(&config.Iterations, "iterations", 0, "Number of iterations to run (required)")
	flag.StringVar(&config.AgentCmd, "agent", "cursor-agent", "Command name for the AI agent CLI tool")
	flag.StringVar(&config.BuildSystem, "build-system", "", "Build system preset (pnpm, npm, yarn, gradle, maven, cargo, go, python) or 'auto' for detection")
	flag.StringVar(&config.TypeCheckCmd, "typecheck", "", "Command to run for type checking (overrides build-system preset)")
	flag.StringVar(&config.TestCmd, "test", "", "Command to run for testing (overrides build-system preset)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.Verbose, "v", false, "Enable verbose output (shorthand)")
	flag.BoolVar(&config.ShowVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&config.ListStatus, "status", false, "List plan status (tested and untested features)")
	flag.BoolVar(&config.ListTested, "list-tested", false, "List only tested features")
	flag.BoolVar(&config.ListUntested, "list-untested", false, "List only untested features")
	flag.BoolVar(&config.GeneratePlan, "generate-plan", false, "Generate plan.json from notes file")
	flag.StringVar(&config.NotesFile, "notes", "", "Path to notes file (required with -generate-plan)")
	flag.StringVar(&config.OutputPlanFile, "output", defaultPlanFile, "Output plan file path (default: plan.json)")

	flag.Usage = func() {
		// Version already includes 'v' prefix from git tags, so don't add another
		ver := getVersion()
		versionDisplay := ver
		if !strings.HasPrefix(ver, "v") && ver != "dev" {
			versionDisplay = "v" + ver
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
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -version                         # Show version information\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 5                    # Run 5 iterations (auto-detect build system)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -iterations 5 -build-system gradle  # Use Gradle preset\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -status                          # Show plan status\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-tested                     # List tested features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-untested                   # List untested features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -generate-plan -notes notes.md   # Generate plan.json from notes\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -generate-plan -notes notes.md -output my-plan.json  # Custom output file\n", os.Args[0])
	}

	flag.Parse()

	// Apply build system configuration
	applyBuildSystemConfig(config)

	return config
}

func validateConfig(config *Config) error {
	// Skip validation for generate-plan (handled separately)
	if config.GeneratePlan {
		if config.NotesFile == "" {
			return fmt.Errorf("notes file is required with -generate-plan (use -notes flag)")
		}
		notesPath := strings.TrimSpace(config.NotesFile)
		if notesPath == "" {
			return fmt.Errorf("notes file path cannot be empty")
		}
		if _, err := os.Stat(notesPath); os.IsNotExist(err) {
			return fmt.Errorf("notes file not found: %s", notesPath)
		}
		// Check if agent command exists
		if _, err := exec.LookPath(config.AgentCmd); err != nil {
			return fmt.Errorf("agent command not found in PATH: %s", config.AgentCmd)
		}
		return nil
	}

	// Skip iteration validation if we're just listing status
	if config.ListStatus || config.ListTested || config.ListUntested {
		if _, err := os.Stat(config.PlanFile); os.IsNotExist(err) {
			return fmt.Errorf("plan file not found: %s", config.PlanFile)
		}
		return nil
	}

	if config.Iterations <= 0 {
		return fmt.Errorf("iterations must be a positive integer (use -iterations flag)")
	}

	if _, err := os.Stat(config.PlanFile); os.IsNotExist(err) {
		return fmt.Errorf("plan file not found: %s", config.PlanFile)
	}

	// Check if agent command exists
	if _, err := exec.LookPath(config.AgentCmd); err != nil {
		return fmt.Errorf("agent command not found in PATH: %s", config.AgentCmd)
	}

	return nil
}

func runIterations(config *Config) error {
	fmt.Printf("Starting iterative development workflow\n")
	fmt.Printf("Plan file: %s\n", config.PlanFile)
	fmt.Printf("Progress file: %s\n", config.ProgressFile)
	fmt.Printf("Iterations: %d\n", config.Iterations)
	fmt.Printf("Agent command: %s\n", config.AgentCmd)
	if config.Verbose {
		fmt.Printf("Type check command: %s\n", config.TypeCheckCmd)
		fmt.Printf("Test command: %s\n", config.TestCmd)
	}
	fmt.Println()

	for i := 1; i <= config.Iterations; i++ {
		fmt.Printf("=== Iteration %d/%d ===\n", i, config.Iterations)

		if config.Verbose {
			fmt.Printf("Executing agent command...\n")
		}

		// Execute the AI agent CLI tool
		result, err := executeAgent(config, i)
		if err != nil {
			return fmt.Errorf("iteration %d failed: %w", i, err)
		}

		// Print the agent output
		fmt.Println(result)

		// Check for completion signal
		if strings.Contains(result, completeSignal) {
			fmt.Printf("\n✓ Plan complete! Detected completion signal after %d iteration(s).\n", i)
			return nil
		}

		fmt.Println() // Empty line between iterations
	}

	fmt.Printf("Completed %d iteration(s) without completion signal.\n", config.Iterations)
	return nil
}

func executeAgent(config *Config, iteration int) (string, error) {
	// Build the prompt for the AI agent
	prompt := buildPrompt(config)

	if config.Verbose {
		fmt.Printf("Prompt: %s\n", prompt)
	}

	// Construct the command based on the agent type
	var cmd *exec.Cmd
	if isCursorAgent(config.AgentCmd) {
		// cursor-agent uses --print --force and prompt as positional argument
		cmd = exec.Command(config.AgentCmd, "--print", "--force", prompt)
	} else {
		// claude uses --permission-mode acceptEdits -p format
		cmd = exec.Command(config.AgentCmd, "--permission-mode", "acceptEdits", "-p", prompt)
	}

	if config.Verbose {
		fmt.Printf("Command: %s %v\n", cmd.Path, cmd.Args)
	}

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start agent command: %w", err)
	}

	// Read stdout and stderr concurrently
	var stdoutBytes, stderrBytes []byte
	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)

	go func() {
		var err error
		stdoutBytes, err = io.ReadAll(stdout)
		stdoutDone <- err
	}()

	go func() {
		var err error
		stderrBytes, err = io.ReadAll(stderr)
		stderrDone <- err
	}()

	// Wait for both to complete
	if err := <-stdoutDone; err != nil {
		cmd.Process.Kill()
		return "", fmt.Errorf("failed to read stdout: %w", err)
	}

	if err := <-stderrDone; err != nil {
		cmd.Process.Kill()
		return "", fmt.Errorf("failed to read stderr: %w", err)
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		// Include stderr in error message if available
		if len(stderrBytes) > 0 {
			return "", fmt.Errorf("agent command failed: %w\nstderr: %s", err, string(stderrBytes))
		}
		return "", fmt.Errorf("agent command failed: %w", err)
	}

	// Combine stdout and stderr for output
	output := strings.TrimSpace(string(stdoutBytes))
	if len(stderrBytes) > 0 {
		output += "\n" + strings.TrimSpace(string(stderrBytes))
	}

	return output, nil
}

// isCursorAgent checks if the agent command is cursor-agent
// This detects cursor-agent, cursor, or any command containing "cursor-agent"
func isCursorAgent(agentCmd string) bool {
	cmd := strings.ToLower(agentCmd)
	return strings.Contains(cmd, "cursor-agent") ||
		(strings.Contains(cmd, "cursor") && !strings.Contains(cmd, "claude"))
}

func buildPrompt(config *Config) string {
	// Resolve absolute paths for the plan and progress files
	planPath, err := filepath.Abs(config.PlanFile)
	if err != nil {
		planPath = config.PlanFile
	}

	progressPath, err := filepath.Abs(config.ProgressFile)
	if err != nil {
		progressPath = config.ProgressFile
	}

	// Build the prompt string as a single line (matching bash script behavior)
	// The bash script uses backslash continuation, which results in a single-line string
	prompt := fmt.Sprintf("@%s @%s ", planPath, progressPath)
	prompt += "1. Find the highest-priority feature to work on and work only on that feature. "
	prompt += "This should be the one YOU decide has the highest priority - not necessarily the first in the list. "
	prompt += fmt.Sprintf("2. Check that the types check via %s and that the tests pass via %s. ", config.TypeCheckCmd, config.TestCmd)
	prompt += "3. Update the PRD with the work that was done. "
	prompt += "4. Append your progress to the progress.txt file. "
	prompt += "Use this to leave a note for the next person working in the codebase. "
	prompt += "5. Make a git commit of that feature. "
	prompt += "ONLY WORK ON A SINGLE FEATURE. "
	prompt += "If, while implementing the feature, you notice the PRD is complete, output <promise>COMPLETE</promise>. "

	return prompt
}

// Helper function to read plan file
func readPlanFile(path string) ([]Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	var plans []Plan
	if err := json.Unmarshal(data, &plans); err != nil {
		return nil, fmt.Errorf("failed to parse plan file: %w", err)
	}

	return plans, nil
}

// listPlanStatus displays plan status (tested/untested features)
func listPlanStatus(config *Config) error {
	plans, err := readPlanFile(config.PlanFile)
	if err != nil {
		return err
	}

	// Determine what to show
	showTested := config.ListStatus || config.ListTested
	showUntested := config.ListStatus || config.ListUntested

	if showTested {
		fmt.Printf("=== Tested Features (from %s) ===\n", config.PlanFile)
		tested := filterPlans(plans, true)
		if len(tested) == 0 {
			fmt.Println("No tested features found")
		} else {
			printPlans(tested)
		}
		fmt.Println()
	}

	if showUntested {
		fmt.Printf("=== Untested Features (from %s) ===\n", config.PlanFile)
		untested := filterPlans(plans, false)
		if len(untested) == 0 {
			fmt.Println("No untested features found")
		} else {
			printPlans(untested)
		}
	}

	return nil
}

// filterPlans filters plans by tested status
func filterPlans(plans []Plan, tested bool) []Plan {
	var result []Plan
	for _, plan := range plans {
		if plan.Tested == tested {
			result = append(result, plan)
		}
	}
	return result
}

// generatePlanFromNotes generates a plan.json file from notes using the AI agent
func generatePlanFromNotes(config *Config) error {
	fmt.Printf("Generating plan from notes file: %s\n", config.NotesFile)
	fmt.Printf("Output plan file: %s\n", config.OutputPlanFile)
	fmt.Printf("Agent command: %s\n\n", config.AgentCmd)

	// Resolve absolute paths
	notesPath, err := filepath.Abs(config.NotesFile)
	if err != nil {
		notesPath = config.NotesFile
	}

	outputPath, err := filepath.Abs(config.OutputPlanFile)
	if err != nil {
		outputPath = config.OutputPlanFile
	}

	// Build the prompt for plan generation
	prompt := buildPlanGenerationPrompt(notesPath, outputPath)

	if config.Verbose {
		fmt.Printf("Prompt: %s\n\n", prompt)
	}

	// Execute the agent
	result, err := executeAgentForPlanGeneration(config, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// The agent should have written the file, but let's verify
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// If file doesn't exist, try to extract JSON from the result and write it
		fmt.Println("Plan file not found, attempting to extract from agent output...")
		if err := extractAndWritePlan(result, outputPath); err != nil {
			return fmt.Errorf("failed to extract plan from output: %w\n\nAgent output:\n%s", err, result)
		}
	}

	fmt.Printf("\n✓ Plan generated successfully: %s\n", outputPath)
	return nil
}

// buildPlanGenerationPrompt creates the prompt for converting notes to plan.json
func buildPlanGenerationPrompt(notesPath, outputPath string) string {
	prompt := fmt.Sprintf("@%s ", notesPath)
	prompt += "Analyze this notes file and create a comprehensive, step-by-step implementation plan in JSON format. "
	prompt += "The plan should follow a sequential story-based methodology where completion and verification of one story is mandatory before starting the next. "
	prompt += "The plan should be saved as a JSON file at: " + outputPath + " "
	prompt += "The JSON must be a valid array of plan objects, each with the following structure: "
	prompt += "{ \"id\": number, \"category\": string (one of: \"infra\", \"ui\", \"db\", \"data\", \"other\", \"chore\"), "
	prompt += "\"description\": string (clear explanation of what the story accomplishes), "
	prompt += "\"steps\": [string] (array of specific, actionable tasks required to complete the story), "
	prompt += "\"expected_output\": string (concrete, measurable outcome that indicates story completion), "
	prompt += "\"tested\": boolean (default false) indicating whether the story has been tested, "
	prompt += "}. "
	prompt += "Stories should be ordered by logical dependency (e.g., database schema before data population, API endpoints before UI components that consume them, authentication before protected features). "
	prompt += "Each story must be: independently completable and testable, have clear acceptance criteria in the expected_output field, break down complex features into manageable sequential tasks, and specify all technical steps needed for implementation. "
	prompt += "The development approach requires building, testing, and verifying each piece of functionality before moving to the next, ensuring quality and stability at every stage before progression. "
	prompt += "When analyzing the notes, identify: core infrastructure requirements (servers, databases, APIs, authentication), data models and relationships, user-facing features and interfaces, integration points between components, and testing and quality assurance needs. "
	prompt += "Categories should reflect the type of work: 'chore' for setup/tooling, 'infra' for infrastructure, 'db' for database work, 'ui' for frontend, 'data' for data-related work, 'other' for core logic/services. "
	prompt += "Ensure the JSON is valid, properly formatted, and contains no syntax errors. Each story should be sufficiently detailed that a developer can implement it without ambiguity. "
	prompt += "Write the complete JSON array to the file: " + outputPath

	return prompt
}

// executeAgentForPlanGeneration executes the agent with plan generation prompt
func executeAgentForPlanGeneration(config *Config, prompt string) (string, error) {
	// Construct the command based on the agent type
	var cmd *exec.Cmd
	if isCursorAgent(config.AgentCmd) {
		// cursor-agent uses --print --force and prompt as positional argument
		cmd = exec.Command(config.AgentCmd, "--print", "--force", prompt)
	} else {
		// claude uses --permission-mode acceptEdits -p format
		cmd = exec.Command(config.AgentCmd, "--permission-mode", "acceptEdits", "-p", prompt)
	}

	if config.Verbose {
		fmt.Printf("Command: %s %v\n", cmd.Path, cmd.Args)
	}

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start agent command: %w", err)
	}

	// Read stdout and stderr concurrently
	var stdoutBytes, stderrBytes []byte
	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)

	go func() {
		var err error
		stdoutBytes, err = io.ReadAll(stdout)
		stdoutDone <- err
	}()

	go func() {
		var err error
		stderrBytes, err = io.ReadAll(stderr)
		stderrDone <- err
	}()

	// Wait for both to complete
	if err := <-stdoutDone; err != nil {
		cmd.Process.Kill()
		return "", fmt.Errorf("failed to read stdout: %w", err)
	}

	if err := <-stderrDone; err != nil {
		cmd.Process.Kill()
		return "", fmt.Errorf("failed to read stderr: %w", err)
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		// Include stderr in error message if available
		if len(stderrBytes) > 0 {
			return "", fmt.Errorf("agent command failed: %w\nstderr: %s", err, string(stderrBytes))
		}
		return "", fmt.Errorf("agent command failed: %w", err)
	}

	// Combine stdout and stderr for output
	output := strings.TrimSpace(string(stdoutBytes))
	if len(stderrBytes) > 0 {
		output += "\n" + strings.TrimSpace(string(stderrBytes))
	}

	return output, nil
}

// extractAndWritePlan attempts to extract JSON from agent output and write it to file
func extractAndWritePlan(output, outputPath string) error {
	// Try to find JSON array in the output
	// Look for content between ```json and ``` or just find the JSON array
	jsonStart := strings.Index(output, "[")
	if jsonStart == -1 {
		// Try to find code block
		jsonBlockStart := strings.Index(output, "```json")
		if jsonBlockStart != -1 {
			jsonStart = jsonBlockStart + 7
		} else {
			jsonBlockStart = strings.Index(output, "```")
			if jsonBlockStart != -1 {
				jsonStart = jsonBlockStart + 3
			}
		}
	}

	if jsonStart == -1 {
		return fmt.Errorf("could not find JSON in output")
	}

	// Find the end of the JSON array
	jsonEnd := strings.LastIndex(output, "]")
	if jsonEnd == -1 || jsonEnd <= jsonStart {
		return fmt.Errorf("could not find end of JSON array")
	}

	// Extract JSON
	jsonContent := strings.TrimSpace(output[jsonStart : jsonEnd+1])

	// Validate it's valid JSON by parsing it
	var plans []Plan
	if err := json.Unmarshal([]byte(jsonContent), &plans); err != nil {
		return fmt.Errorf("extracted content is not valid JSON: %w", err)
	}

	// Write to file with proper formatting
	formattedJSON, err := json.MarshalIndent(plans, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, formattedJSON, 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	return nil
}

// printPlans prints plans in a formatted table
func printPlans(plans []Plan) {
	// Find max widths for formatting
	maxIDLen := 0
	maxCatLen := 0
	for _, plan := range plans {
		idLen := len(fmt.Sprintf("%d", plan.ID))
		if idLen > maxIDLen {
			maxIDLen = idLen
		}
		if len(plan.Category) > maxCatLen {
			maxCatLen = len(plan.Category)
		}
	}

	// Ensure minimum widths
	if maxIDLen < 2 {
		maxIDLen = 2
	}
	if maxCatLen < 8 {
		maxCatLen = 8
	}

	// Print formatted output
	for _, plan := range plans {
		fmt.Printf("%-*d  %-*s  %s\n", maxIDLen, plan.ID, maxCatLen, plan.Category, plan.Description)
	}
}

// Helper function to append to progress file (for potential future use)
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
