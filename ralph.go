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
	"time"
)

const (
	completeSignal      = "<promise>COMPLETE</promise>"
	defaultPlanFile     = "plan.json"
	defaultProgressFile = "progress.txt"
)

// Config holds the application configuration
type Config struct {
	PlanFile       string
	ProgressFile   string
	Iterations     int
	AgentCmd       string
	TypeCheckCmd   string
	TestCmd        string
	Verbose        bool
	ListStatus     bool
	ListTested     bool
	ListUntested   bool
	GeneratePlan   bool
	NotesFile      string
	OutputPlanFile string
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
	flag.StringVar(&config.TypeCheckCmd, "typecheck", "pnpm typecheck", "Command to run for type checking")
	flag.StringVar(&config.TestCmd, "test", "pnpm test", "Command to run for testing")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.Verbose, "v", false, "Enable verbose output (shorthand)")
	flag.BoolVar(&config.ListStatus, "status", false, "List plan status (tested and untested features)")
	flag.BoolVar(&config.ListTested, "list-tested", false, "List only tested features")
	flag.BoolVar(&config.ListUntested, "list-untested", false, "List only untested features")
	flag.BoolVar(&config.GeneratePlan, "generate-plan", false, "Generate plan.json from notes file")
	flag.StringVar(&config.NotesFile, "notes", "", "Path to notes file (required with -generate-plan)")
	flag.StringVar(&config.OutputPlanFile, "output", defaultPlanFile, "Output plan file path (default: plan.json)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -iterations 5                    # Run 5 iterations\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -status                          # Show plan status\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-tested                     # List tested features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-untested                   # List untested features\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -generate-plan -notes notes.md   # Generate plan.json from notes\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -generate-plan -notes notes.md -output my-plan.json  # Custom output file\n", os.Args[0])
	}

	flag.Parse()

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
	prompt += "The plan should be saved as a JSON file at: " + outputPath + " "
	prompt += "The JSON must be a valid array of plan objects, each with the following structure: "
	prompt += "{ \"id\": number, \"category\": string (e.g., \"chore\", \"infra\", \"db\", \"ui\", \"feature\", \"other\"), "
	prompt += "\"description\": string (clear, actionable description), "
	prompt += "\"steps\": [string] (array of specific, implementable steps), "
	prompt += "\"expected_output\": string (what success looks like), "
	prompt += "\"tested\": boolean (default false) }. "
	prompt += "Break down the notes into logical, sequential features/tasks. "
	prompt += "Each plan item should be self-contained and implementable. "
	prompt += "Categories should reflect the type of work: 'chore' for setup/tooling, 'infra' for infrastructure, "
	prompt += "'db' for database work, 'ui' for frontend, 'feature' for features, 'other' for core logic/services. "
	prompt += "Ensure the JSON is valid and properly formatted. "
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
