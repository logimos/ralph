// Package plan provides plan file operations for Ralph.
package plan

import (
	"fmt"
	"regexp"
	"strings"
)

// AnalysisIssue represents an issue found during plan analysis
type AnalysisIssue struct {
	PlanID      int      // ID of the plan with the issue
	IssueType   string   // "compound" or "complex"
	Description string   // Human-readable description of the issue
	Severity    string   // "warning" or "suggestion"
	Suggestions []string // Suggested actions to resolve the issue
}

// AnalysisResult represents the result of analyzing a plan file
type AnalysisResult struct {
	TotalPlans       int             // Total number of plans analyzed
	IssuesFound      int             // Number of issues found
	CompoundFeatures int             // Features with 'and' suggesting multiple features
	ComplexFeatures  int             // Features with >9 steps
	Issues           []AnalysisIssue // List of issues found
}

// AnalyzePlans analyzes a list of plans for potential refinement issues
func AnalyzePlans(plans []Plan) *AnalysisResult {
	result := &AnalysisResult{
		TotalPlans: len(plans),
		Issues:     []AnalysisIssue{},
	}

	for _, plan := range plans {
		// Check for compound features (descriptions with 'and')
		if isCompoundFeature(plan.Description) {
			issue := AnalysisIssue{
				PlanID:      plan.ID,
				IssueType:   "compound",
				Description: fmt.Sprintf("Feature #%d description contains 'and', may represent multiple features", plan.ID),
				Severity:    "suggestion",
				Suggestions: suggestCompoundSplit(plan),
			}
			result.Issues = append(result.Issues, issue)
			result.CompoundFeatures++
		}

		// Check for complex features (>9 steps)
		if len(plan.Steps) > 9 {
			issue := AnalysisIssue{
				PlanID:      plan.ID,
				IssueType:   "complex",
				Description: fmt.Sprintf("Feature #%d has %d steps (>9), may be too complex", plan.ID, len(plan.Steps)),
				Severity:    "warning",
				Suggestions: suggestComplexSplit(plan),
			}
			result.Issues = append(result.Issues, issue)
			result.ComplexFeatures++
		}
	}

	result.IssuesFound = len(result.Issues)
	return result
}

// isCompoundFeature checks if a description suggests multiple features
func isCompoundFeature(description string) bool {
	// Normalize to lowercase for matching
	lower := strings.ToLower(description)

	// Pattern: " and " as a word separator (not part of another word)
	// Check if description contains " and " as a word separator
	if !strings.Contains(lower, " and ") {
		return false
	}

	// Common acceptable "and" pairs that don't indicate multiple features
	// These are closely related concepts that typically belong together
	acceptablePairs := []string{
		"read and write",
		"yaml and json",
		"json and yaml",
		"input and output",
		"request and response",
		"load and save",
		"save and load",
		"encode and decode",
		"serialize and deserialize",
		"success and error",
		"error and success",
		"start and stop",
		"open and close",
		"create and delete",
		"add and remove",
		"push and pull",
		"get and set",
		"search and filter",
		"sort and filter",
		"authentication and authorization", // These are closely related security concepts
		"auth and authz",
		"encrypt and decrypt",
		"compress and decompress",
		"upload and download",
		"import and export",
		"copy and paste",
		"undo and redo",
		"lock and unlock",
		"enable and disable",
		"show and hide",
		"expand and collapse",
		"connect and disconnect",
		"subscribe and unsubscribe",
	}

	// Check if it's an acceptable pair
	for _, pair := range acceptablePairs {
		if strings.Contains(lower, pair) {
			return false
		}
	}

	// Look for patterns that strongly suggest compound features:
	// Must be "verb X and verb Y" pattern where both are action verbs
	// Example: "implement user login and implement admin dashboard"
	// NOT: "implement search and filter component" (single component)
	
	// Split by " and " and check if BOTH parts have distinct action verbs at start
	parts := strings.Split(lower, " and ")
	if len(parts) != 2 {
		return false
	}

	// Check if both parts start with action verbs (suggesting separate features)
	actionVerbs := []string{"add", "create", "implement", "build", "setup", "configure", "enable", "integrate", "develop", "design"}
	
	firstPart := strings.TrimSpace(parts[0])
	secondPart := strings.TrimSpace(parts[1])
	
	firstHasVerb := false
	secondHasVerb := false
	
	for _, verb := range actionVerbs {
		if strings.HasPrefix(firstPart, verb+" ") {
			firstHasVerb = true
		}
		if strings.HasPrefix(secondPart, verb+" ") {
			secondHasVerb = true
		}
	}
	
	// Only flag as compound if BOTH parts start with action verbs
	// This catches: "implement X and implement Y" but not "implement X and Y component"
	return firstHasVerb && secondHasVerb
}

// suggestCompoundSplit provides suggestions for splitting a compound feature
func suggestCompoundSplit(plan Plan) []string {
	suggestions := []string{}

	lower := strings.ToLower(plan.Description)
	parts := strings.Split(lower, " and ")

	if len(parts) >= 2 {
		suggestions = append(suggestions,
			fmt.Sprintf("Consider splitting into %d separate features:", len(parts)))

		for i, part := range parts {
			trimmed := strings.TrimSpace(part)
			// Capitalize first letter
			if len(trimmed) > 0 {
				trimmed = strings.ToUpper(string(trimmed[0])) + trimmed[1:]
			}
			suggestions = append(suggestions,
				fmt.Sprintf("  %d. %s", i+1, trimmed))
		}
	}

	suggestions = append(suggestions,
		"Each feature should have a single, focused objective")

	return suggestions
}

// suggestComplexSplit provides suggestions for splitting a complex feature
func suggestComplexSplit(plan Plan) []string {
	suggestions := []string{}
	stepCount := len(plan.Steps)

	suggestions = append(suggestions,
		fmt.Sprintf("Feature has %d steps - consider splitting into smaller features", stepCount))

	// Suggest logical groupings based on step content
	groups := groupStepsByTheme(plan.Steps)
	if len(groups) > 1 {
		suggestions = append(suggestions,
			fmt.Sprintf("Detected %d potential logical groupings:", len(groups)))

		for i, group := range groups {
			suggestions = append(suggestions,
				fmt.Sprintf("  Group %d (%d steps): %s", i+1, len(group.Steps), group.Theme))
		}
	}

	// General guidance
	if stepCount > 12 {
		suggestions = append(suggestions,
			"Recommended: Split into 2-3 smaller features with 4-6 steps each")
	} else {
		suggestions = append(suggestions,
			"Recommended: Split into 2 smaller features with 4-5 steps each")
	}

	return suggestions
}

// StepGroup represents a logical grouping of steps
type StepGroup struct {
	Theme string
	Steps []string
}

// groupStepsByTheme attempts to identify logical groupings in steps
func groupStepsByTheme(steps []string) []StepGroup {
	if len(steps) < 4 {
		return nil
	}

	// Theme keywords to look for
	themes := map[string][]string{
		"setup/config": {"create", "define", "configure", "setup", "initialize", "add.*config"},
		"implementation": {"implement", "build", "add.*logic", "add.*function", "write"},
		"integration":    {"integrate", "connect", "wire", "hook", "inject"},
		"testing":        {"test", "verify", "validate", "check", "ensure"},
		"documentation":  {"document", "readme", "doc", "comment"},
		"cli/flags":      {"flag", "cli", "command", "argument", "option"},
	}

	groups := []StepGroup{}
	currentGroup := StepGroup{}

	for _, step := range steps {
		lower := strings.ToLower(step)
		matchedTheme := ""

		for theme, keywords := range themes {
			for _, keyword := range keywords {
				matched, _ := regexp.MatchString(keyword, lower)
				if matched {
					matchedTheme = theme
					break
				}
			}
			if matchedTheme != "" {
				break
			}
		}

		if matchedTheme == "" {
			matchedTheme = "general"
		}

		// If theme changes and we have steps, save the group
		if currentGroup.Theme != "" && currentGroup.Theme != matchedTheme && len(currentGroup.Steps) >= 2 {
			groups = append(groups, currentGroup)
			currentGroup = StepGroup{Theme: matchedTheme, Steps: []string{step}}
		} else {
			if currentGroup.Theme == "" {
				currentGroup.Theme = matchedTheme
			}
			currentGroup.Steps = append(currentGroup.Steps, step)
		}
	}

	// Add final group
	if len(currentGroup.Steps) > 0 {
		groups = append(groups, currentGroup)
	}

	// Filter out tiny groups by merging them
	finalGroups := []StepGroup{}
	for _, group := range groups {
		if len(group.Steps) >= 2 {
			finalGroups = append(finalGroups, group)
		} else if len(finalGroups) > 0 {
			// Merge with previous group
			finalGroups[len(finalGroups)-1].Steps = append(finalGroups[len(finalGroups)-1].Steps, group.Steps...)
		} else {
			finalGroups = append(finalGroups, group)
		}
	}

	return finalGroups
}

// FormatAnalysisResult formats an analysis result for display
func FormatAnalysisResult(result *AnalysisResult) string {
	var sb strings.Builder

	sb.WriteString("=== Plan Analysis Report ===\n\n")
	sb.WriteString(fmt.Sprintf("Total plans analyzed: %d\n", result.TotalPlans))
	sb.WriteString(fmt.Sprintf("Issues found: %d\n", result.IssuesFound))

	if result.IssuesFound == 0 {
		sb.WriteString("\n✓ All plans appear well-structured and self-contained.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("  - Compound features (with 'and'): %d\n", result.CompoundFeatures))
	sb.WriteString(fmt.Sprintf("  - Complex features (>9 steps): %d\n", result.ComplexFeatures))

	sb.WriteString("\n--- Issues ---\n")

	for _, issue := range result.Issues {
		severity := "SUGGESTION"
		if issue.Severity == "warning" {
			severity = "WARNING"
		}

		sb.WriteString(fmt.Sprintf("\n[%s] Feature #%d: %s\n", severity, issue.PlanID, issue.IssueType))
		sb.WriteString(fmt.Sprintf("  %s\n", issue.Description))

		if len(issue.Suggestions) > 0 {
			sb.WriteString("  Suggestions:\n")
			for _, suggestion := range issue.Suggestions {
				sb.WriteString(fmt.Sprintf("    %s\n", suggestion))
			}
		}
	}

	sb.WriteString("\n--- Summary ---\n")
	sb.WriteString("Plan refinement can improve:\n")
	sb.WriteString("  • Code review efficiency (smaller, focused changes)\n")
	sb.WriteString("  • Testing reliability (isolated test cases)\n")
	sb.WriteString("  • Progress tracking (more granular milestones)\n")
	sb.WriteString("  • Recovery from failures (less work to redo)\n")

	return sb.String()
}

// GetPlanAnalysisSummary returns a short summary of the analysis
func GetPlanAnalysisSummary(result *AnalysisResult) string {
	if result.IssuesFound == 0 {
		return fmt.Sprintf("Plan analysis: %d plans, all well-structured", result.TotalPlans)
	}
	return fmt.Sprintf("Plan analysis: %d plans, %d issues (%d compound, %d complex)",
		result.TotalPlans, result.IssuesFound, result.CompoundFeatures, result.ComplexFeatures)
}
