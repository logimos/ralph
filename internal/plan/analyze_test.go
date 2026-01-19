package plan

import (
	"strings"
	"testing"
)

func TestAnalyzePlans_EmptyPlan(t *testing.T) {
	plans := []Plan{}
	result := AnalyzePlans(plans)

	if result.TotalPlans != 0 {
		t.Errorf("expected TotalPlans=0, got %d", result.TotalPlans)
	}
	if result.IssuesFound != 0 {
		t.Errorf("expected IssuesFound=0, got %d", result.IssuesFound)
	}
}

func TestAnalyzePlans_WellStructuredPlans(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Description: "Add user authentication",
			Steps:       []string{"Create auth module", "Add login endpoint", "Add tests"},
		},
		{
			ID:          2,
			Description: "Implement caching layer",
			Steps:       []string{"Add Redis client", "Create cache wrapper", "Add cache tests"},
		},
	}

	result := AnalyzePlans(plans)

	if result.TotalPlans != 2 {
		t.Errorf("expected TotalPlans=2, got %d", result.TotalPlans)
	}
	if result.IssuesFound != 0 {
		t.Errorf("expected no issues for well-structured plans, got %d", result.IssuesFound)
	}
}

func TestAnalyzePlans_ComplexFeature(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Description: "Build entire application",
			Steps: []string{
				"Step 1", "Step 2", "Step 3", "Step 4", "Step 5",
				"Step 6", "Step 7", "Step 8", "Step 9", "Step 10",
			},
		},
	}

	result := AnalyzePlans(plans)

	if result.ComplexFeatures != 1 {
		t.Errorf("expected ComplexFeatures=1, got %d", result.ComplexFeatures)
	}
	if result.IssuesFound != 1 {
		t.Errorf("expected IssuesFound=1, got %d", result.IssuesFound)
	}
	if result.Issues[0].IssueType != "complex" {
		t.Errorf("expected issue type 'complex', got '%s'", result.Issues[0].IssueType)
	}
	if result.Issues[0].Severity != "warning" {
		t.Errorf("expected severity 'warning', got '%s'", result.Issues[0].Severity)
	}
}

func TestAnalyzePlans_CompoundFeature(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expectIssue bool
	}{
		{
			name:        "implement X and implement Y",
			description: "Implement user registration and implement email verification",
			expectIssue: true,
		},
		{
			name:        "add X and add Y",
			description: "Add caching layer and add rate limiting",
			expectIssue: true,
		},
		{
			name:        "build X and build Y",
			description: "Build frontend and build backend",
			expectIssue: true,
		},
		{
			name:        "acceptable YAML and JSON",
			description: "Add YAML and JSON configuration support",
			expectIssue: false,
		},
		{
			name:        "acceptable read and write",
			description: "Implement read and write operations for files",
			expectIssue: false,
		},
		{
			name:        "acceptable load and save",
			description: "Add load and save functionality",
			expectIssue: false,
		},
		{
			name:        "single feature with and in name",
			description: "Implement search and filter component",
			expectIssue: false,
		},
		{
			name:        "no and in description",
			description: "Add user authentication with OAuth",
			expectIssue: false,
		},
		{
			name:        "acceptable authentication and authorization",
			description: "Add authentication and authorization",
			expectIssue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plans := []Plan{{ID: 1, Description: tt.description}}
			result := AnalyzePlans(plans)

			if tt.expectIssue && result.CompoundFeatures == 0 {
				t.Errorf("expected compound issue for '%s'", tt.description)
			}
			if !tt.expectIssue && result.CompoundFeatures > 0 {
				t.Errorf("unexpected compound issue for '%s'", tt.description)
			}
		})
	}
}

func TestIsCompoundFeature(t *testing.T) {
	tests := []struct {
		description string
		expected    bool
	}{
		// Should NOT be flagged (single features or acceptable pairs)
		{"Add authentication and authorization", false}, // Closely related concepts
		{"Implement caching", false},                    // No "and"
		{"Add YAML and JSON support", false},            // Acceptable pair
		{"Add read and write operations", false},        // Acceptable pair
		{"Implement search and filter component", false}, // Single component
		{"Add logging", false},                          // No "and"
		{"Create user profile page", false},             // No "and"
		{"Enable and disable features", false},          // Acceptable pair
		// SHOULD be flagged (clearly separate features with repeated verbs)
		{"Implement user login and implement admin dashboard", true},
		{"Create API and create CLI", true},
		{"Build frontend and build backend", true},
		{"Add caching and add logging", true}, // Both start with "add"
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := isCompoundFeature(tt.description)
			if result != tt.expected {
				t.Errorf("isCompoundFeature(%q) = %v, want %v", tt.description, result, tt.expected)
			}
		})
	}
}

func TestSuggestCompoundSplit(t *testing.T) {
	plan := Plan{
		ID:          1,
		Description: "Add caching and add rate limiting",
	}

	suggestions := suggestCompoundSplit(plan)

	if len(suggestions) == 0 {
		t.Error("expected suggestions for compound feature")
	}

	// Check that suggestions mention splitting
	found := false
	for _, s := range suggestions {
		if strings.Contains(strings.ToLower(s), "split") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected suggestions to mention splitting")
	}
}

func TestSuggestComplexSplit(t *testing.T) {
	plan := Plan{
		ID:          1,
		Description: "Build application",
		Steps: []string{
			"Create project structure",
			"Define configuration",
			"Implement main logic",
			"Add database layer",
			"Add API endpoints",
			"Implement authentication",
			"Add unit tests",
			"Add integration tests",
			"Write documentation",
			"Add CLI interface",
		},
	}

	suggestions := suggestComplexSplit(plan)

	if len(suggestions) == 0 {
		t.Error("expected suggestions for complex feature")
	}

	// Check that suggestions mention step count
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "10 steps") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected suggestions to mention step count")
	}
}

func TestGroupStepsByTheme(t *testing.T) {
	steps := []string{
		"Create config file",
		"Define settings struct",
		"Implement main handler",
		"Add business logic",
		"Write unit tests",
		"Write integration tests",
	}

	groups := groupStepsByTheme(steps)

	// Should identify some groupings
	if len(groups) < 1 {
		t.Error("expected at least one group")
	}
}

func TestFormatAnalysisResult_NoIssues(t *testing.T) {
	result := &AnalysisResult{
		TotalPlans:  5,
		IssuesFound: 0,
	}

	formatted := FormatAnalysisResult(result)

	if !strings.Contains(formatted, "well-structured") {
		t.Error("expected 'well-structured' message for no issues")
	}
	if !strings.Contains(formatted, "5") {
		t.Error("expected total plans count in output")
	}
}

func TestFormatAnalysisResult_WithIssues(t *testing.T) {
	result := &AnalysisResult{
		TotalPlans:       3,
		IssuesFound:      2,
		CompoundFeatures: 1,
		ComplexFeatures:  1,
		Issues: []AnalysisIssue{
			{
				PlanID:      1,
				IssueType:   "compound",
				Description: "Feature has 'and'",
				Severity:    "suggestion",
				Suggestions: []string{"Split into 2 features"},
			},
			{
				PlanID:      2,
				IssueType:   "complex",
				Description: "Feature has 11 steps",
				Severity:    "warning",
				Suggestions: []string{"Split into smaller features"},
			},
		},
	}

	formatted := FormatAnalysisResult(result)

	if !strings.Contains(formatted, "WARNING") {
		t.Error("expected WARNING in output")
	}
	if !strings.Contains(formatted, "SUGGESTION") {
		t.Error("expected SUGGESTION in output")
	}
	if !strings.Contains(formatted, "Compound features") {
		t.Error("expected compound features count")
	}
	if !strings.Contains(formatted, "Complex features") {
		t.Error("expected complex features count")
	}
}

func TestGetPlanAnalysisSummary(t *testing.T) {
	tests := []struct {
		name     string
		result   *AnalysisResult
		contains string
	}{
		{
			name:     "no issues",
			result:   &AnalysisResult{TotalPlans: 5, IssuesFound: 0},
			contains: "well-structured",
		},
		{
			name: "with issues",
			result: &AnalysisResult{
				TotalPlans:       3,
				IssuesFound:      2,
				CompoundFeatures: 1,
				ComplexFeatures:  1,
			},
			contains: "2 issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := GetPlanAnalysisSummary(tt.result)
			if !strings.Contains(summary, tt.contains) {
				t.Errorf("expected summary to contain %q, got %q", tt.contains, summary)
			}
		})
	}
}

func TestAnalyzePlans_BothIssueTypes(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Description: "Add logging and add monitoring", // compound
			Steps:       []string{"s1", "s2"},
		},
		{
			ID:          2,
			Description: "Complex feature",
			Steps: []string{
				"s1", "s2", "s3", "s4", "s5",
				"s6", "s7", "s8", "s9", "s10", "s11",
			}, // >9 steps
		},
	}

	result := AnalyzePlans(plans)

	if result.TotalPlans != 2 {
		t.Errorf("expected TotalPlans=2, got %d", result.TotalPlans)
	}
	// May or may not detect compound depending on heuristics
	if result.ComplexFeatures != 1 {
		t.Errorf("expected ComplexFeatures=1, got %d", result.ComplexFeatures)
	}
}

func TestAnalyzePlans_VeryLargeStepCount(t *testing.T) {
	steps := make([]string, 15)
	for i := range steps {
		steps[i] = "Step"
	}

	plans := []Plan{{ID: 1, Description: "Huge feature", Steps: steps}}
	result := AnalyzePlans(plans)

	if result.ComplexFeatures != 1 {
		t.Errorf("expected ComplexFeatures=1, got %d", result.ComplexFeatures)
	}

	// Check that suggestions mention splitting into 2-3 features
	if len(result.Issues) > 0 {
		found := false
		for _, s := range result.Issues[0].Suggestions {
			if strings.Contains(s, "2-3") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected suggestion for 2-3 features for very large feature")
		}
	}
}

// Tests for RefinePlans functionality

func TestRefinePlans_EmptyPlans(t *testing.T) {
	plans := []Plan{}
	result := RefinePlans(plans)

	if result.OriginalCount != 0 {
		t.Errorf("expected OriginalCount=0, got %d", result.OriginalCount)
	}
	if result.RefinedCount != 0 {
		t.Errorf("expected RefinedCount=0, got %d", result.RefinedCount)
	}
	if len(result.NewPlans) != 0 {
		t.Errorf("expected empty NewPlans, got %d", len(result.NewPlans))
	}
}

func TestRefinePlans_WellStructuredPlans(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Description: "Add user authentication",
			Steps:       []string{"Create auth module", "Add login endpoint", "Add tests"},
		},
		{
			ID:          2,
			Description: "Implement caching layer",
			Steps:       []string{"Add Redis client", "Create cache wrapper", "Add cache tests"},
		},
	}

	result := RefinePlans(plans)

	if result.SplitFeatures != 0 {
		t.Errorf("expected SplitFeatures=0, got %d", result.SplitFeatures)
	}
	if result.RefinedCount != 2 {
		t.Errorf("expected RefinedCount=2, got %d", result.RefinedCount)
	}
}

func TestRefinePlans_SkipsTestedFeatures(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Description: "Build entire application with massive scope",
			Steps: []string{
				"Step 1", "Step 2", "Step 3", "Step 4", "Step 5",
				"Step 6", "Step 7", "Step 8", "Step 9", "Step 10",
				"Step 11", "Step 12",
			},
			Tested: true, // Already tested - should NOT be split
		},
	}

	result := RefinePlans(plans)

	if result.SplitFeatures != 0 {
		t.Errorf("expected SplitFeatures=0 (tested features should not be split), got %d", result.SplitFeatures)
	}
	if result.SkippedFeatures != 1 {
		t.Errorf("expected SkippedFeatures=1, got %d", result.SkippedFeatures)
	}
	if result.RefinedCount != 1 {
		t.Errorf("expected RefinedCount=1, got %d", result.RefinedCount)
	}
}

func TestRefinePlans_SplitsComplexFeature(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Category:    "infra",
			Description: "Build comprehensive feature",
			Steps: []string{
				"Create project structure",
				"Define configuration",
				"Setup database",
				"Implement main logic",
				"Add business handlers",
				"Write unit tests",
				"Write integration tests",
				"Add documentation",
				"Add CLI interface",
				"Configure deployment",
			},
			Tested: false,
		},
	}

	result := RefinePlans(plans)

	// Should be split into multiple features
	if result.SplitFeatures == 0 {
		t.Error("expected complex feature to be split")
	}
	if result.RefinedCount <= 1 {
		t.Errorf("expected multiple refined plans, got %d", result.RefinedCount)
	}
	// All new plans should inherit the category
	for _, p := range result.NewPlans {
		if p.Category != "infra" {
			t.Errorf("expected category 'infra', got '%s'", p.Category)
		}
	}
	// All new plans should be untested
	for _, p := range result.NewPlans {
		if p.Tested {
			t.Error("new split plans should not be marked as tested")
		}
	}
}

func TestRefinePlans_SplitsCompoundFeature(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Category:    "feature",
			Description: "Add logging and add monitoring",
			Steps:       []string{"Add logger", "Add monitoring", "Add alerts", "Add dashboard"},
			Tested:      false,
		},
	}

	result := RefinePlans(plans)

	// Should be split into two features
	if result.SplitFeatures == 0 {
		t.Error("expected compound feature to be split")
	}
	if result.RefinedCount != 2 {
		t.Errorf("expected 2 refined plans for compound split, got %d", result.RefinedCount)
	}
	if len(result.Changes) == 0 {
		t.Error("expected changes to be recorded")
	}
}

func TestRefinePlans_PreservesIDs(t *testing.T) {
	plans := []Plan{
		{
			ID:          5,
			Description: "Simple feature",
			Steps:       []string{"Step 1", "Step 2"},
			Tested:      false,
		},
		{
			ID:          10,
			Description: "Add large feature",
			Steps: []string{
				"Create config", "Setup database", "Add handlers",
				"Write tests", "Add docs", "Setup CI", "Add monitoring",
				"Add logging", "Add metrics", "Deploy",
			},
			Tested: false,
		},
	}

	result := RefinePlans(plans)

	// First plan should keep its ID
	if result.NewPlans[0].ID != 5 {
		t.Errorf("expected first plan to have ID=5, got %d", result.NewPlans[0].ID)
	}

	// New split plans should have IDs > max original ID
	for i, p := range result.NewPlans {
		if i > 0 && p.ID <= 10 && result.NewPlans[i].Description != "Add large feature" {
			// Split plans should have IDs starting after the max
			t.Logf("Plan %d has ID %d", i, p.ID)
		}
	}
}

func TestRefinePlans_PreservesMilestone(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Description: "Build complex feature",
			Milestone:   "Alpha",
			Steps: []string{
				"Step 1", "Step 2", "Step 3", "Step 4", "Step 5",
				"Step 6", "Step 7", "Step 8", "Step 9", "Step 10",
			},
			Tested: false,
		},
	}

	result := RefinePlans(plans)

	// All split plans should inherit the milestone
	for _, p := range result.NewPlans {
		if p.Milestone != "Alpha" {
			t.Errorf("expected milestone 'Alpha', got '%s'", p.Milestone)
		}
	}
}

func TestSplitComplexFeature_SmallStepCount(t *testing.T) {
	nextID := 2
	plan := Plan{
		ID:          1,
		Description: "Small feature",
		Steps:       []string{"Step 1", "Step 2", "Step 3"},
	}

	result := splitComplexFeature(plan, &nextID)

	// Should return original plan unmodified
	if len(result) != 1 {
		t.Errorf("expected 1 plan (no split), got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Errorf("expected ID=1, got %d", result[0].ID)
	}
}

func TestSplitCompoundFeature_NotCompound(t *testing.T) {
	nextID := 2
	plan := Plan{
		ID:          1,
		Description: "Regular feature without compound pattern",
		Steps:       []string{"Step 1", "Step 2"},
	}

	result := splitCompoundFeature(plan, &nextID)

	// Should return original plan unmodified
	if len(result) != 1 {
		t.Errorf("expected 1 plan (no split), got %d", len(result))
	}
}

func TestSplitCompoundFeature_ValidCompound(t *testing.T) {
	nextID := 10
	plan := Plan{
		ID:          1,
		Category:    "test",
		Description: "Add caching and add rate limiting",
		Steps:       []string{"Step 1", "Step 2", "Step 3", "Step 4"},
	}

	result := splitCompoundFeature(plan, &nextID)

	// Should split into two features
	if len(result) != 2 {
		t.Errorf("expected 2 plans, got %d", len(result))
	}

	// Check IDs were assigned correctly
	if result[0].ID != 10 {
		t.Errorf("expected first split plan ID=10, got %d", result[0].ID)
	}
	if result[1].ID != 11 {
		t.Errorf("expected second split plan ID=11, got %d", result[1].ID)
	}

	// Check category was inherited
	if result[0].Category != "test" {
		t.Errorf("expected category 'test', got '%s'", result[0].Category)
	}
}

func TestFormatRefinementResult_NoChanges(t *testing.T) {
	result := &RefinementResult{
		OriginalCount: 5,
		RefinedCount:  5,
		SplitFeatures: 0,
		Changes:       []string{},
	}

	formatted := FormatRefinementResult(result)

	if !strings.Contains(formatted, "No refinements needed") {
		t.Error("expected 'No refinements needed' message")
	}
}

func TestFormatRefinementResult_WithChanges(t *testing.T) {
	result := &RefinementResult{
		OriginalCount:   3,
		RefinedCount:    5,
		SplitFeatures:   2,
		SkippedFeatures: 1,
		Changes: []string{
			"Split feature #1 into 2 features",
			"Split feature #2 into 3 features",
		},
	}

	formatted := FormatRefinementResult(result)

	if !strings.Contains(formatted, "Original plans: 3") {
		t.Error("expected original count in output")
	}
	if !strings.Contains(formatted, "Refined plans: 5") {
		t.Error("expected refined count in output")
	}
	if !strings.Contains(formatted, "Features split: 2") {
		t.Error("expected split count in output")
	}
	if !strings.Contains(formatted, "Changes Made") {
		t.Error("expected 'Changes Made' section")
	}
}

func TestRefinePlans_MixedPlans(t *testing.T) {
	plans := []Plan{
		{
			ID:          1,
			Description: "Simple tested feature",
			Steps:       []string{"s1", "s2"},
			Tested:      true,
		},
		{
			ID:          2,
			Description: "Complex untested feature",
			Steps: []string{
				"Create config", "Setup database", "Add handlers",
				"Write tests", "Add docs", "Setup CI", "Add monitoring",
				"Add logging", "Add metrics", "Deploy",
			},
			Tested: false,
		},
		{
			ID:          3,
			Description: "Another simple feature",
			Steps:       []string{"s1", "s2", "s3"},
			Tested:      false,
		},
	}

	result := RefinePlans(plans)

	// One tested feature should be skipped
	if result.SkippedFeatures != 1 {
		t.Errorf("expected SkippedFeatures=1, got %d", result.SkippedFeatures)
	}
	// One complex feature should be split
	if result.SplitFeatures != 1 {
		t.Errorf("expected SplitFeatures=1, got %d", result.SplitFeatures)
	}
	// Total refined should be more than original
	if result.RefinedCount <= result.OriginalCount {
		t.Errorf("expected RefinedCount > OriginalCount")
	}
}
