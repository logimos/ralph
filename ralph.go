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
	PlanFile     string
	ProgressFile string
	Iterations   int
	AgentCmd     string
	TypeCheckCmd string
	TestCmd      string
	Verbose      bool
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

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	return config
}

func validateConfig(config *Config) error {
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

// Helper function to read plan file (for potential future use)
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
