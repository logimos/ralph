// Package plan provides plan file operations for Ralph.
package plan

import (
	"encoding/json"
	"fmt"
	"os"
)

// Plan represents the structure of a plan file
type Plan struct {
	ID             int      `json:"id"`
	Category       string   `json:"category,omitempty"`
	Command        string   `json:"command,omitempty"`
	Description    string   `json:"description"`
	Steps          []string `json:"steps,omitempty"`
	ExpectedOutput string   `json:"expected_output,omitempty"`
	Tested         bool     `json:"tested,omitempty"`
	Milestone      string   `json:"milestone,omitempty"`       // Optional milestone this feature belongs to
	MilestoneOrder int      `json:"milestone_order,omitempty"` // Order within the milestone (for prioritization)
	Deferred       bool     `json:"deferred,omitempty"`        // Whether this feature has been deferred due to scope constraints
	DeferReason    string   `json:"defer_reason,omitempty"`    // Reason for deferral (if deferred)
}

// ReadFile reads and parses a plan file
func ReadFile(path string) ([]Plan, error) {
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

// WriteFile writes plans to a plan file
func WriteFile(path string, plans []Plan) error {
	data, err := json.MarshalIndent(plans, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal plans: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	return nil
}

// MarkDeferred marks a plan as deferred with the given reason
func MarkDeferred(plans []Plan, featureID int, reason string) bool {
	for i := range plans {
		if plans[i].ID == featureID {
			plans[i].Deferred = true
			plans[i].DeferReason = reason
			return true
		}
	}
	return false
}

// FilterDeferred returns plans filtered by deferred status
func FilterDeferred(plans []Plan, deferred bool) []Plan {
	var result []Plan
	for _, plan := range plans {
		if plan.Deferred == deferred {
			result = append(result, plan)
		}
	}
	return result
}

// GetByID returns a plan by its ID, or nil if not found
func GetByID(plans []Plan, id int) *Plan {
	for i := range plans {
		if plans[i].ID == id {
			return &plans[i]
		}
	}
	return nil
}

// Filter filters plans by tested status
func Filter(plans []Plan, tested bool) []Plan {
	var result []Plan
	for _, plan := range plans {
		if plan.Tested == tested {
			result = append(result, plan)
		}
	}
	return result
}

// Print prints plans in a formatted table
func Print(plans []Plan) {
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

// ExtractAndWrite attempts to extract JSON from agent output and write it to file
func ExtractAndWrite(output, outputPath string) error {
	// Try to find JSON array in the output
	// Look for content between ```json and ``` or just find the JSON array
	jsonStart := -1
	for i := 0; i < len(output); i++ {
		if output[i] == '[' {
			jsonStart = i
			break
		}
	}

	if jsonStart == -1 {
		// Try to find code block
		jsonBlockStart := indexOf(output, "```json")
		if jsonBlockStart != -1 {
			jsonStart = jsonBlockStart + 7
		} else {
			jsonBlockStart = indexOf(output, "```")
			if jsonBlockStart != -1 {
				jsonStart = jsonBlockStart + 3
			}
		}
	}

	if jsonStart == -1 {
		return fmt.Errorf("could not find JSON in output")
	}

	// Find the end of the JSON array
	jsonEnd := lastIndexOf(output, "]")
	if jsonEnd == -1 || jsonEnd <= jsonStart {
		return fmt.Errorf("could not find end of JSON array")
	}

	// Extract JSON
	jsonContent := trimSpace(output[jsonStart : jsonEnd+1])

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

// indexOf returns the index of the first occurrence of substr in s, or -1
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// lastIndexOf returns the index of the last occurrence of substr in s, or -1
func lastIndexOf(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
