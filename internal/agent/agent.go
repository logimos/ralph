// Package agent provides AI agent execution and communication.
package agent

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/logimos/ralph/internal/config"
)

// IsCursorAgent checks if the agent command is cursor-agent
// This detects cursor-agent, cursor, or any command containing "cursor-agent"
func IsCursorAgent(agentCmd string) bool {
	cmd := strings.ToLower(agentCmd)
	return strings.Contains(cmd, "cursor-agent") ||
		(strings.Contains(cmd, "cursor") && !strings.Contains(cmd, "claude"))
}

// Execute runs the AI agent with the given prompt and returns the output
func Execute(cfg *config.Config, prompt string) (string, error) {
	// Construct the command based on the agent type
	var cmd *exec.Cmd
	if IsCursorAgent(cfg.AgentCmd) {
		// cursor-agent uses --print --force and prompt as positional argument
		cmd = exec.Command(cfg.AgentCmd, "--print", "--force", prompt)
	} else {
		// claude uses --permission-mode acceptEdits -p format
		cmd = exec.Command(cfg.AgentCmd, "--permission-mode", "acceptEdits", "-p", prompt)
	}

	if cfg.Verbose {
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
