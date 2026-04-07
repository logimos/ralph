// Package prompt provides prompt building logic for Ralph.
package prompt

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/logimos/ralph/internal/config"
)

const (
	// CompleteSignal is the marker indicating the plan is complete
	CompleteSignal = "<promise>COMPLETE</promise>"
)

// BuildIterationPrompt builds the prompt for an iteration.
// layersSnapshotPath: optional absolute path to context-snapshot.md (Layers compact); "" to omit.
// progressReadPath: path to use in @ for reading recent progress; normally the full progress file.
// When bounded context is enabled, pass the path to progress-context.txt (or "") to fall back to full progress.
// planUsesPriority: true if plan.json sets any non-zero "priority" field — prompt matches plan.NextWorkFeature ordering.
func BuildIterationPrompt(cfg *config.Config, layersSnapshotPath string, progressReadPath string, planUsesPriority bool) string {
	// Resolve absolute paths for the plan and progress files
	planPath, err := filepath.Abs(cfg.PlanFile)
	if err != nil {
		planPath = cfg.PlanFile
	}

	fullProgressPath, err := filepath.Abs(cfg.ProgressFile)
	if err != nil {
		fullProgressPath = cfg.ProgressFile
	}

	readPath := strings.TrimSpace(progressReadPath)
	if readPath == "" {
		readPath = fullProgressPath
	} else if p, err := filepath.Abs(readPath); err == nil {
		readPath = filepath.Clean(p)
	} else {
		readPath = filepath.Clean(readPath)
	}

	// Build the prompt string as a single line (matching bash script behavior)
	// The bash script uses backslash continuation, which results in a single-line string
	trimmedSnap := strings.TrimSpace(layersSnapshotPath)
	var prompt string
	if trimmedSnap != "" {
		snap := filepath.Clean(trimmedSnap)
		prompt = fmt.Sprintf("@%s @%s @%s ", planPath, snap, readPath)
		prompt += "The second @ file is a bounded recent run snapshot (Layers compact output). "
		if readPath != fullProgressPath {
			prompt += "The third @ file is a bounded UTF-8 tail of the progress log; prefer the snapshot and bounded file for context over loading the full log. "
		} else {
			prompt += "Prefer the snapshot for recent structured context over reading the full progress file. "
		}
	} else {
		prompt = fmt.Sprintf("@%s @%s ", planPath, readPath)
		if readPath != fullProgressPath {
			prompt += "The second @ file is a bounded UTF-8 tail of the progress log (full log is appended separately). "
		}
	}
	if planUsesPriority {
		prompt += "1. Work on the next unfinished feature by priority: higher \"priority\" in plan.json comes first; when priorities tie or are unset, use file order. Work on only that feature. "
	} else {
		prompt += "1. Work on the next unfinished feature in plan.json file order (first untested, non-deferred). Work on only that feature. "
	}
	prompt += fmt.Sprintf("2. Check that the types check via %s and that the tests pass via %s. ", cfg.TypeCheckCmd, cfg.TestCmd)
	prompt += "3. Update the PRD with the work that was done. "
	prompt += fmt.Sprintf("4. Append your progress to %s (full chronological log). ", fullProgressPath)
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
	prompt += "{ \"id\": number, \"priority\": number (optional, higher = sooner; 0 = default), \"category\": string (e.g., \"chore\", \"infra\", \"db\", \"ui\", \"feature\", \"other\"), "
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
