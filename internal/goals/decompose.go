// Package goals provides high-level goal management and automatic plan decomposition for Ralph.
package goals

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/logimos/ralph/internal/plan"
)

// DecompositionResult represents the result of goal decomposition
type DecompositionResult struct {
	Goal           *Goal
	GeneratedPlans []plan.Plan
	Dependencies   map[int][]int // Map of plan ID to IDs it depends on
	Success        bool
	Message        string
	RawOutput      string // Raw agent output for debugging
}

// BuildGoalDecompositionPrompt creates the prompt for decomposing a goal into plan items
func BuildGoalDecompositionPrompt(goal *Goal, existingPlans []plan.Plan, outputPath string) string {
	var sb strings.Builder

	sb.WriteString("Analyze the following high-level goal and decompose it into a detailed, actionable implementation plan.\n\n")

	// Goal description
	sb.WriteString("## Goal\n")
	sb.WriteString(fmt.Sprintf("Description: %s\n", goal.Description))

	// Success criteria if provided
	if len(goal.SuccessCriteria) > 0 {
		sb.WriteString("\n## Success Criteria\n")
		for i, criteria := range goal.SuccessCriteria {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, criteria))
		}
	}

	// Category hint
	if goal.Category != "" {
		sb.WriteString(fmt.Sprintf("\n## Category Hint: %s\n", goal.Category))
	}

	// Tags for context
	if len(goal.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Tags: %s\n", strings.Join(goal.Tags, ", ")))
	}

	// Existing plans for context
	if len(existingPlans) > 0 {
		sb.WriteString("\n## Existing Plan Items (for context and ID assignment)\n")
		maxID := 0
		for _, p := range existingPlans {
			sb.WriteString(fmt.Sprintf("- ID %d: %s [%s] - %s\n", p.ID, p.Category, statusString(p.Tested), p.Description))
			if p.ID > maxID {
				maxID = p.ID
			}
		}
		sb.WriteString(fmt.Sprintf("\nStart new plan IDs from %d\n", maxID+1))
	}

	// Instructions
	sb.WriteString("\n## Instructions\n")
	sb.WriteString("Create a JSON array of plan items that will achieve this goal. ")
	sb.WriteString("Each plan item should follow this structure:\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"id\": <unique integer>,\n")
	sb.WriteString("  \"category\": \"<chore|infra|db|ui|feature|api|security|other>\",\n")
	sb.WriteString("  \"description\": \"<clear, actionable description>\",\n")
	sb.WriteString("  \"steps\": [\"<specific step 1>\", \"<specific step 2>\", ...],\n")
	sb.WriteString("  \"expected_output\": \"<what success looks like>\",\n")
	sb.WriteString("  \"tested\": false,\n")
	sb.WriteString("  \"depends_on\": [<IDs of plan items this depends on>] // optional\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("Requirements:\n")
	sb.WriteString("1. Break down the goal into small, implementable tasks (each doable in 1-3 iterations)\n")
	sb.WriteString("2. Order tasks logically - dependencies should come first\n")
	sb.WriteString("3. Each task should be self-contained and testable\n")
	sb.WriteString("4. Include setup/infrastructure tasks if needed\n")
	sb.WriteString("5. Include testing tasks where appropriate\n")
	sb.WriteString("6. Be specific in steps - avoid vague instructions\n")
	sb.WriteString("7. Use the 'depends_on' field to indicate task dependencies\n\n")

	sb.WriteString(fmt.Sprintf("Write the complete JSON array to: %s\n", outputPath))
	sb.WriteString("The file should contain ONLY the JSON array of new plan items (not existing ones).\n")

	return sb.String()
}

// BuildMultiGoalDecompositionPrompt creates a prompt for decomposing multiple goals
func BuildMultiGoalDecompositionPrompt(goals []Goal, existingPlans []plan.Plan, outputPath string) string {
	var sb strings.Builder

	sb.WriteString("Analyze the following high-level goals and decompose them into a unified, actionable implementation plan.\n\n")

	// List all goals
	sb.WriteString("## Goals (in priority order)\n")
	for i, goal := range goals {
		sb.WriteString(fmt.Sprintf("\n### Goal %d: %s\n", i+1, goal.Description))
		sb.WriteString(fmt.Sprintf("Priority: %d\n", goal.Priority))
		if goal.Category != "" {
			sb.WriteString(fmt.Sprintf("Category: %s\n", goal.Category))
		}
		if len(goal.SuccessCriteria) > 0 {
			sb.WriteString("Success Criteria:\n")
			for _, c := range goal.SuccessCriteria {
				sb.WriteString(fmt.Sprintf("  - %s\n", c))
			}
		}
		if len(goal.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf("Depends on goals: %s\n", strings.Join(goal.Dependencies, ", ")))
		}
	}

	// Existing plans for context
	if len(existingPlans) > 0 {
		sb.WriteString("\n## Existing Plan Items\n")
		maxID := 0
		for _, p := range existingPlans {
			sb.WriteString(fmt.Sprintf("- ID %d: [%s] %s\n", p.ID, p.Category, p.Description))
			if p.ID > maxID {
				maxID = p.ID
			}
		}
		sb.WriteString(fmt.Sprintf("\nStart new plan IDs from %d\n", maxID+1))
	}

	// Instructions
	sb.WriteString("\n## Instructions\n")
	sb.WriteString("Create a unified plan that addresses all goals. The plan should:\n")
	sb.WriteString("1. Respect goal dependencies and priorities\n")
	sb.WriteString("2. Identify shared work between goals to avoid duplication\n")
	sb.WriteString("3. Order tasks to maximize efficiency\n")
	sb.WriteString("4. Include a 'goal_id' field to track which goal each task serves\n\n")

	sb.WriteString("Plan item structure:\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"id\": <unique integer>,\n")
	sb.WriteString("  \"category\": \"<chore|infra|db|ui|feature|api|security|other>\",\n")
	sb.WriteString("  \"description\": \"<clear, actionable description>\",\n")
	sb.WriteString("  \"steps\": [\"<specific step 1>\", ...],\n")
	sb.WriteString("  \"expected_output\": \"<what success looks like>\",\n")
	sb.WriteString("  \"tested\": false,\n")
	sb.WriteString("  \"goal_id\": \"<ID of the goal this serves>\",\n")
	sb.WriteString("  \"depends_on\": [<IDs of plan items this depends on>]\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString(fmt.Sprintf("Write the complete JSON array to: %s\n", outputPath))

	return sb.String()
}

// ParseDecompositionResult parses the agent output to extract generated plans
func ParseDecompositionResult(output string, goal *Goal) (*DecompositionResult, error) {
	result := &DecompositionResult{
		Goal:      goal,
		RawOutput: output,
	}

	// Try to find JSON array in the output
	jsonStart := strings.Index(output, "[")
	if jsonStart == -1 {
		// Try finding it in a code block
		codeStart := strings.Index(output, "```json")
		if codeStart != -1 {
			jsonStart = strings.Index(output[codeStart+7:], "[")
			if jsonStart != -1 {
				jsonStart = codeStart + 7 + jsonStart
			}
		} else {
			codeStart = strings.Index(output, "```")
			if codeStart != -1 {
				jsonStart = strings.Index(output[codeStart+3:], "[")
				if jsonStart != -1 {
					jsonStart = codeStart + 3 + jsonStart
				}
			}
		}
	}

	if jsonStart == -1 {
		result.Success = false
		result.Message = "Could not find JSON array in output"
		return result, fmt.Errorf("no JSON array found in agent output")
	}

	// Find the end of the JSON array
	jsonEnd := strings.LastIndex(output, "]")
	if jsonEnd == -1 || jsonEnd <= jsonStart {
		result.Success = false
		result.Message = "Could not find end of JSON array"
		return result, fmt.Errorf("could not find end of JSON array")
	}

	jsonContent := output[jsonStart : jsonEnd+1]

	// Parse the JSON
	var plans []plan.Plan
	if err := json.Unmarshal([]byte(jsonContent), &plans); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to parse JSON: %v", err)
		return result, err
	}

	result.GeneratedPlans = plans
	result.Success = true
	result.Message = fmt.Sprintf("Generated %d plan items", len(plans))

	// Extract dependencies from the plans (if they have depends_on field)
	result.Dependencies = extractDependencies(plans)

	return result, nil
}

// ExtendedPlan is a plan with additional fields for decomposition
type ExtendedPlan struct {
	plan.Plan
	GoalID    string `json:"goal_id,omitempty"`
	DependsOn []int  `json:"depends_on,omitempty"`
}

// extractDependencies extracts dependency information from plans
// Note: This assumes plans might have a "depends_on" field in their JSON
func extractDependencies(plans []plan.Plan) map[int][]int {
	deps := make(map[int][]int)
	// Dependencies would be extracted from extended plan fields
	// For now, return empty map - dependencies are tracked via plan order
	return deps
}

// statusString returns a status string for a plan
func statusString(tested bool) string {
	if tested {
		return "done"
	}
	return "todo"
}

// MergePlans merges generated plans with existing plans
func MergePlans(existing []plan.Plan, generated []plan.Plan) []plan.Plan {
	// Find the maximum existing ID
	maxID := 0
	for _, p := range existing {
		if p.ID > maxID {
			maxID = p.ID
		}
	}

	// Reassign IDs to generated plans if needed to avoid conflicts
	idMap := make(map[int]int) // old ID -> new ID
	for i := range generated {
		oldID := generated[i].ID
		if oldID <= maxID {
			maxID++
			generated[i].ID = maxID
		} else if generated[i].ID > maxID {
			maxID = generated[i].ID
		}
		idMap[oldID] = generated[i].ID
	}

	// Combine the plans
	return append(existing, generated...)
}

// ValidatePlanDependencies checks that plan dependencies are valid
func ValidatePlanDependencies(plans []plan.Plan, deps map[int][]int) []string {
	var errors []string
	planIDs := make(map[int]bool)

	for _, p := range plans {
		planIDs[p.ID] = true
	}

	for planID, depIDs := range deps {
		if !planIDs[planID] {
			errors = append(errors, fmt.Sprintf("dependency references unknown plan ID %d", planID))
			continue
		}
		for _, depID := range depIDs {
			if !planIDs[depID] {
				errors = append(errors, fmt.Sprintf("plan %d depends on unknown plan ID %d", planID, depID))
			}
			if depID == planID {
				errors = append(errors, fmt.Sprintf("plan %d has circular dependency on itself", planID))
			}
		}
	}

	return errors
}

// GetNextPlanID returns the next available plan ID given existing plans
func GetNextPlanID(plans []plan.Plan) int {
	maxID := 0
	for _, p := range plans {
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	return maxID + 1
}
