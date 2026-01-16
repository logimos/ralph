// Package prompt provides prompt building logic for Ralph.
package prompt

import (
	"fmt"
	"path/filepath"

	"github.com/logimos/ralph/internal/config"
)

const (
	// CompleteSignal is the marker indicating the plan is complete
	CompleteSignal = "<promise>COMPLETE</promise>"
)

// BuildIterationPrompt builds the prompt for an iteration
func BuildIterationPrompt(cfg *config.Config) string {
	// Resolve absolute paths for the plan and progress files
	planPath, err := filepath.Abs(cfg.PlanFile)
	if err != nil {
		planPath = cfg.PlanFile
	}

	progressPath, err := filepath.Abs(cfg.ProgressFile)
	if err != nil {
		progressPath = cfg.ProgressFile
	}

	// Build the prompt string as a single line (matching bash script behavior)
	// The bash script uses backslash continuation, which results in a single-line string
	prompt := fmt.Sprintf("@%s @%s ", planPath, progressPath)
	prompt += "1. Find the highest-priority feature to work on and work only on that feature. "
	prompt += "This should be the one YOU decide has the highest priority - not necessarily the first in the list. "
	prompt += fmt.Sprintf("2. Check that the types check via %s and that the tests pass via %s. ", cfg.TypeCheckCmd, cfg.TestCmd)
	prompt += "3. Update the PRD with the work that was done. "
	prompt += "4. Append your progress to the progress.txt file. "
	prompt += "Use this to leave a note for the next person working in the codebase. "
	prompt += "5. Make a git commit of that feature. "
	prompt += "ONLY WORK ON A SINGLE FEATURE. "
	prompt += fmt.Sprintf("If, while implementing the feature, you notice the PRD is complete, output %s. ", CompleteSignal)

	return prompt
}

// BuildPlanGenerationPrompt creates the prompt for converting notes to plan.json
func BuildPlanGenerationPrompt(notesPath, outputPath string) string {
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
