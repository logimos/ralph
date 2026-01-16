package goals

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/logimos/ralph/internal/plan"
)

func TestNewManager(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Description: "Test plan 1"},
		{ID: 2, Description: "Test plan 2"},
	}

	mgr := NewManager(plans)
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if len(mgr.goals) != 0 {
		t.Errorf("Expected 0 goals, got %d", len(mgr.goals))
	}
}

func TestAddGoal(t *testing.T) {
	mgr := NewManager(nil)

	goal := Goal{
		Description: "Test goal",
		Priority:    5,
	}

	err := mgr.AddGoal(goal)
	if err != nil {
		t.Errorf("AddGoal failed: %v", err)
	}

	if len(mgr.goals) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(mgr.goals))
	}

	// Check that ID was generated
	if mgr.goals[0].ID == "" {
		t.Error("Goal ID was not generated")
	}

	// Check timestamps
	if mgr.goals[0].CreatedAt.IsZero() {
		t.Error("CreatedAt was not set")
	}
}

func TestAddGoalValidation(t *testing.T) {
	mgr := NewManager(nil)

	// Empty description should fail
	err := mgr.AddGoal(Goal{})
	if err == nil {
		t.Error("Expected error for empty description")
	}

	// Duplicate ID should fail
	goal1 := Goal{ID: "test1", Description: "Test 1"}
	mgr.AddGoal(goal1)

	goal2 := Goal{ID: "test1", Description: "Test 2"}
	err = mgr.AddGoal(goal2)
	if err == nil {
		t.Error("Expected error for duplicate ID")
	}
}

func TestAddGoalFromDescription(t *testing.T) {
	mgr := NewManager(nil)

	goal, err := mgr.AddGoalFromDescription("Add user authentication", 10)
	if err != nil {
		t.Errorf("AddGoalFromDescription failed: %v", err)
	}

	if goal.Description != "Add user authentication" {
		t.Errorf("Expected description 'Add user authentication', got %q", goal.Description)
	}

	if goal.Priority != 10 {
		t.Errorf("Expected priority 10, got %d", goal.Priority)
	}

	// Check category was inferred
	if goal.Category == "" {
		t.Error("Category was not inferred")
	}
}

func TestGetGoalByID(t *testing.T) {
	mgr := NewManager(nil)

	goal := Goal{ID: "test-goal", Description: "Test"}
	mgr.AddGoal(goal)

	found := mgr.GetGoalByID("test-goal")
	if found == nil {
		t.Error("GetGoalByID returned nil for existing goal")
	}

	notFound := mgr.GetGoalByID("nonexistent")
	if notFound != nil {
		t.Error("GetGoalByID returned non-nil for nonexistent goal")
	}
}

func TestRemoveGoal(t *testing.T) {
	mgr := NewManager(nil)

	goal := Goal{ID: "test-goal", Description: "Test"}
	mgr.AddGoal(goal)

	if !mgr.RemoveGoal("test-goal") {
		t.Error("RemoveGoal returned false for existing goal")
	}

	if len(mgr.goals) != 0 {
		t.Errorf("Expected 0 goals after removal, got %d", len(mgr.goals))
	}

	if mgr.RemoveGoal("nonexistent") {
		t.Error("RemoveGoal returned true for nonexistent goal")
	}
}

func TestUpdateGoal(t *testing.T) {
	mgr := NewManager(nil)

	goal := Goal{ID: "test-goal", Description: "Original"}
	mgr.AddGoal(goal)

	updated := Goal{ID: "test-goal", Description: "Updated", Priority: 10}
	err := mgr.UpdateGoal(updated)
	if err != nil {
		t.Errorf("UpdateGoal failed: %v", err)
	}

	found := mgr.GetGoalByID("test-goal")
	if found.Description != "Updated" {
		t.Errorf("Expected description 'Updated', got %q", found.Description)
	}

	// Update nonexistent goal should fail
	err = mgr.UpdateGoal(Goal{ID: "nonexistent", Description: "Test"})
	if err == nil {
		t.Error("Expected error for updating nonexistent goal")
	}
}

func TestCalculateProgress(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Description: "Task 1", Tested: true},
		{ID: 2, Description: "Task 2", Tested: false},
		{ID: 3, Description: "Task 3", Tested: true},
		{ID: 4, Description: "Task 4", Tested: false, Deferred: true},
	}

	mgr := NewManager(plans)
	goal := Goal{
		ID:               "test-goal",
		Description:      "Test goal",
		GeneratedPlanIDs: []int{1, 2, 3, 4},
	}
	mgr.AddGoal(goal)

	progress := mgr.CalculateProgress("test-goal")
	if progress == nil {
		t.Fatal("CalculateProgress returned nil")
	}

	if progress.TotalPlanItems != 4 {
		t.Errorf("Expected 4 total items, got %d", progress.TotalPlanItems)
	}

	if progress.CompletedItems != 2 {
		t.Errorf("Expected 2 completed items, got %d", progress.CompletedItems)
	}

	if progress.DeferredItems != 1 {
		t.Errorf("Expected 1 deferred item, got %d", progress.DeferredItems)
	}

	if progress.RemainingItems != 1 {
		t.Errorf("Expected 1 remaining item, got %d", progress.RemainingItems)
	}

	expectedPercent := 50.0
	if progress.PercentComplete != expectedPercent {
		t.Errorf("Expected %.0f%% complete, got %.0f%%", expectedPercent, progress.PercentComplete)
	}
}

func TestCalculateProgressWithDependencies(t *testing.T) {
	mgr := NewManager(nil)

	// Create goals with dependencies
	goal1 := Goal{ID: "goal1", Description: "Base goal", Status: StatusComplete}
	goal2 := Goal{ID: "goal2", Description: "Dependent goal", Dependencies: []string{"goal1"}}
	goal3 := Goal{ID: "goal3", Description: "Blocked goal", Dependencies: []string{"goal2"}}

	mgr.AddGoal(goal1)
	mgr.AddGoal(goal2)
	mgr.AddGoal(goal3)

	// goal2 should not be blocked (goal1 is complete)
	progress2 := mgr.CalculateProgress("goal2")
	if len(progress2.BlockedByGoals) > 0 {
		t.Errorf("goal2 should not be blocked, but is blocked by: %v", progress2.BlockedByGoals)
	}

	// goal3 should be blocked (goal2 is not complete)
	progress3 := mgr.CalculateProgress("goal3")
	if len(progress3.BlockedByGoals) != 1 {
		t.Errorf("goal3 should be blocked by 1 goal, got %d", len(progress3.BlockedByGoals))
	}
}

func TestGetGoalsByPriority(t *testing.T) {
	mgr := NewManager(nil)

	mgr.AddGoal(Goal{ID: "low", Description: "Low priority", Priority: 1})
	mgr.AddGoal(Goal{ID: "high", Description: "High priority", Priority: 10})
	mgr.AddGoal(Goal{ID: "medium", Description: "Medium priority", Priority: 5})

	sorted := mgr.GetGoalsByPriority()

	if sorted[0].ID != "high" {
		t.Errorf("Expected first goal to be 'high', got %q", sorted[0].ID)
	}
	if sorted[1].ID != "medium" {
		t.Errorf("Expected second goal to be 'medium', got %q", sorted[1].ID)
	}
	if sorted[2].ID != "low" {
		t.Errorf("Expected third goal to be 'low', got %q", sorted[2].ID)
	}
}

func TestGetNextGoalToWork(t *testing.T) {
	mgr := NewManager(nil)

	// Complete goal should be skipped
	mgr.AddGoal(Goal{ID: "complete", Description: "Complete", Priority: 10, Status: StatusComplete})
	// Blocked goal should be skipped
	mgr.AddGoal(Goal{ID: "blocked", Description: "Blocked", Priority: 8, Dependencies: []string{"pending"}})
	// Pending goal should be returned
	mgr.AddGoal(Goal{ID: "pending", Description: "Pending", Priority: 5, Status: StatusPending})

	next := mgr.GetNextGoalToWork()
	if next == nil {
		t.Fatal("GetNextGoalToWork returned nil")
	}

	if next.ID != "pending" {
		t.Errorf("Expected next goal to be 'pending', got %q", next.ID)
	}
}

func TestLinkPlanToGoal(t *testing.T) {
	plans := []plan.Plan{{ID: 1, Description: "Test"}}
	mgr := NewManager(plans)

	goal := Goal{ID: "test-goal", Description: "Test"}
	mgr.AddGoal(goal)

	err := mgr.LinkPlanToGoal("test-goal", 1)
	if err != nil {
		t.Errorf("LinkPlanToGoal failed: %v", err)
	}

	found := mgr.GetGoalByID("test-goal")
	if len(found.GeneratedPlanIDs) != 1 {
		t.Errorf("Expected 1 linked plan, got %d", len(found.GeneratedPlanIDs))
	}

	// Linking same plan again should not duplicate
	mgr.LinkPlanToGoal("test-goal", 1)
	found = mgr.GetGoalByID("test-goal")
	if len(found.GeneratedPlanIDs) != 1 {
		t.Errorf("Expected 1 linked plan after duplicate link, got %d", len(found.GeneratedPlanIDs))
	}
}

func TestMarkGoalComplete(t *testing.T) {
	mgr := NewManager(nil)

	goal := Goal{ID: "test-goal", Description: "Test"}
	mgr.AddGoal(goal)

	err := mgr.MarkGoalComplete("test-goal")
	if err != nil {
		t.Errorf("MarkGoalComplete failed: %v", err)
	}

	found := mgr.GetGoalByID("test-goal")
	if found.Status != StatusComplete {
		t.Errorf("Expected status 'complete', got %q", found.Status)
	}

	if found.CompletedAt == nil {
		t.Error("CompletedAt was not set")
	}
}

func TestSaveAndLoadGoals(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "goals_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goalsFile := filepath.Join(tmpDir, "goals.json")

	// Create manager and add goals
	mgr := NewManager(nil)
	mgr.SetGoalsFile(goalsFile)

	mgr.AddGoal(Goal{ID: "goal1", Description: "Goal 1", Priority: 10})
	mgr.AddGoal(Goal{ID: "goal2", Description: "Goal 2", Priority: 5})

	// Save goals
	err = mgr.SaveGoals()
	if err != nil {
		t.Errorf("SaveGoals failed: %v", err)
	}

	// Load into new manager
	mgr2 := NewManager(nil)
	err = mgr2.LoadGoals(goalsFile)
	if err != nil {
		t.Errorf("LoadGoals failed: %v", err)
	}

	if len(mgr2.goals) != 2 {
		t.Errorf("Expected 2 goals, got %d", len(mgr2.goals))
	}

	// Verify goals
	g1 := mgr2.GetGoalByID("goal1")
	if g1 == nil || g1.Description != "Goal 1" {
		t.Error("Goal 1 not loaded correctly")
	}
}

func TestLoadGoalsSimpleFormat(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "goals_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goalsFile := filepath.Join(tmpDir, "goals.json")

	// Write goals in simple array format
	simpleGoals := []Goal{
		{ID: "goal1", Description: "Goal 1"},
		{ID: "goal2", Description: "Goal 2"},
	}
	data, _ := json.MarshalIndent(simpleGoals, "", "  ")
	os.WriteFile(goalsFile, data, 0644)

	// Load goals
	mgr := NewManager(nil)
	err = mgr.LoadGoals(goalsFile)
	if err != nil {
		t.Errorf("LoadGoals failed for simple format: %v", err)
	}

	if len(mgr.goals) != 2 {
		t.Errorf("Expected 2 goals, got %d", len(mgr.goals))
	}
}

func TestInferCategory(t *testing.T) {
	// Test that inferCategory returns a valid category (not necessarily a specific one
	// since map iteration order is non-deterministic when multiple keywords match)
	validCategories := map[string]bool{
		"feature": true, "infrastructure": true, "database": true,
		"ui": true, "api": true, "security": true, "testing": true,
		"performance": true, "documentation": true, "refactor": true, "other": true,
	}

	tests := []struct {
		description     string
		validCategories []string // Any of these is acceptable
	}{
		{"Add user authentication", []string{"security", "feature"}},      // Contains "auth" (security) and "add" (feature)
		{"Setup CI/CD pipeline", []string{"infrastructure"}},               // Contains "setup", "ci"
		{"Create user dashboard UI", []string{"ui", "feature"}},           // Contains "ui" and "create"
		{"Add payment API endpoint", []string{"api", "feature"}},          // Contains "api" and "add"
		{"Write unit tests for auth", []string{"testing", "security"}},    // Contains "test" and "auth"
		{"Optimize database queries", []string{"database", "performance"}}, // Contains "database" and "optimize"
		{"Refactor the codebase", []string{"refactor"}},                   // Contains "refactor"
		{"Add documentation", []string{"documentation", "feature"}},       // Contains "document" and "add"
		{"Random task", []string{"other"}},                                 // No keywords match
	}

	for _, tt := range tests {
		result := inferCategory(tt.description)
		
		// Check that the result is a valid category
		if !validCategories[result] {
			t.Errorf("inferCategory(%q) = %q, which is not a valid category", tt.description, result)
			continue
		}

		// Check that the result is one of the expected categories
		found := false
		for _, expected := range tt.validCategories {
			if result == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("inferCategory(%q) = %q, expected one of %v", tt.description, result, tt.validCategories)
		}
	}
}

func TestFormatProgressBar(t *testing.T) {
	tests := []struct {
		progress *GoalProgress
		width    int
		contains string
	}{
		{
			progress: &GoalProgress{TotalPlanItems: 0},
			width:    10,
			contains: "0%",
		},
		{
			progress: &GoalProgress{TotalPlanItems: 4, CompletedItems: 2, PercentComplete: 50},
			width:    10,
			contains: "50%",
		},
		{
			progress: &GoalProgress{TotalPlanItems: 4, CompletedItems: 4, PercentComplete: 100},
			width:    10,
			contains: "100%",
		},
	}

	for _, tt := range tests {
		result := FormatProgressBar(tt.progress, tt.width)
		if len(result) == 0 {
			t.Error("FormatProgressBar returned empty string")
		}
		if !containsString(result, tt.contains) {
			t.Errorf("FormatProgressBar result %q doesn't contain %q", result, tt.contains)
		}
	}
}

func TestSummary(t *testing.T) {
	mgr := NewManager(nil)

	// Empty manager
	summary := mgr.Summary()
	if summary != "No goals defined." {
		t.Errorf("Expected 'No goals defined.', got %q", summary)
	}

	// With goals
	mgr.AddGoal(Goal{ID: "pending", Description: "Pending goal", Status: StatusPending})
	mgr.AddGoal(Goal{ID: "active", Description: "Active goal", Status: StatusInProgress})
	mgr.AddGoal(Goal{ID: "complete", Description: "Complete goal", Status: StatusComplete})

	summary = mgr.Summary()
	if !containsString(summary, "Pending Goals:") {
		t.Error("Summary should contain 'Pending Goals:'")
	}
	if !containsString(summary, "Active Goals:") {
		t.Error("Summary should contain 'Active Goals:'")
	}
	if !containsString(summary, "Completed Goals:") {
		t.Error("Summary should contain 'Completed Goals:'")
	}
}

func TestHasGoals(t *testing.T) {
	mgr := NewManager(nil)

	if mgr.HasGoals() {
		t.Error("HasGoals should return false for empty manager")
	}

	mgr.AddGoal(Goal{Description: "Test"})

	if !mgr.HasGoals() {
		t.Error("HasGoals should return true after adding a goal")
	}
}

func TestCount(t *testing.T) {
	mgr := NewManager(nil)

	if mgr.Count() != 0 {
		t.Errorf("Expected count 0, got %d", mgr.Count())
	}

	mgr.AddGoal(Goal{Description: "Test 1"})
	mgr.AddGoal(Goal{Description: "Test 2"})

	if mgr.Count() != 2 {
		t.Errorf("Expected count 2, got %d", mgr.Count())
	}
}

func TestGetPendingActiveCompletedGoals(t *testing.T) {
	mgr := NewManager(nil)

	mgr.AddGoal(Goal{ID: "p1", Description: "Pending 1", Status: StatusPending})
	mgr.AddGoal(Goal{ID: "p2", Description: "Pending 2", Status: StatusPending})
	mgr.AddGoal(Goal{ID: "a1", Description: "Active 1", Status: StatusInProgress})
	mgr.AddGoal(Goal{ID: "c1", Description: "Complete 1", Status: StatusComplete})

	pending := mgr.GetPendingGoals()
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending goals, got %d", len(pending))
	}

	active := mgr.GetActiveGoals()
	if len(active) != 1 {
		t.Errorf("Expected 1 active goal, got %d", len(active))
	}

	completed := mgr.GetCompletedGoals()
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed goal, got %d", len(completed))
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Additional tests for decompose.go

func TestBuildGoalDecompositionPrompt(t *testing.T) {
	goal := &Goal{
		ID:          "test-goal",
		Description: "Add user authentication",
		SuccessCriteria: []string{
			"Users can log in",
			"Sessions persist",
		},
		Category: "security",
		Tags:     []string{"auth", "security"},
	}

	existingPlans := []plan.Plan{
		{ID: 1, Category: "infra", Description: "Setup project"},
	}

	prompt := BuildGoalDecompositionPrompt(goal, existingPlans, "plan.json")

	// Check prompt contains key elements
	if !containsSubstring(prompt, "Add user authentication") {
		t.Error("Prompt should contain goal description")
	}
	if !containsSubstring(prompt, "Users can log in") {
		t.Error("Prompt should contain success criteria")
	}
	if !containsSubstring(prompt, "security") {
		t.Error("Prompt should contain category")
	}
	if !containsSubstring(prompt, "plan.json") {
		t.Error("Prompt should contain output path")
	}
}

func TestBuildMultiGoalDecompositionPrompt(t *testing.T) {
	goals := []Goal{
		{ID: "goal1", Description: "Goal 1", Priority: 10},
		{ID: "goal2", Description: "Goal 2", Priority: 5, Dependencies: []string{"goal1"}},
	}

	prompt := BuildMultiGoalDecompositionPrompt(goals, nil, "plan.json")

	if !containsSubstring(prompt, "Goal 1") {
		t.Error("Prompt should contain first goal")
	}
	if !containsSubstring(prompt, "Goal 2") {
		t.Error("Prompt should contain second goal")
	}
	if !containsSubstring(prompt, "priority order") {
		t.Error("Prompt should mention priority order")
	}
}

func TestParseDecompositionResult(t *testing.T) {
	goal := &Goal{ID: "test", Description: "Test"}

	// Valid JSON in output
	output := `Some text before
[
    {"id": 1, "category": "infra", "description": "Task 1", "tested": false},
    {"id": 2, "category": "feature", "description": "Task 2", "tested": false}
]
Some text after`

	result, err := ParseDecompositionResult(output, goal)
	if err != nil {
		t.Errorf("ParseDecompositionResult failed: %v", err)
	}

	if !result.Success {
		t.Error("Result should be successful")
	}

	if len(result.GeneratedPlans) != 2 {
		t.Errorf("Expected 2 plans, got %d", len(result.GeneratedPlans))
	}
}

func TestParseDecompositionResultCodeBlock(t *testing.T) {
	goal := &Goal{ID: "test", Description: "Test"}

	output := "Here are the tasks:\n```json\n[{\"id\": 1, \"description\": \"Task 1\"}]\n```"

	result, err := ParseDecompositionResult(output, goal)
	if err != nil {
		t.Errorf("ParseDecompositionResult failed for code block: %v", err)
	}

	if !result.Success {
		t.Error("Result should be successful")
	}
}

func TestParseDecompositionResultNoJSON(t *testing.T) {
	goal := &Goal{ID: "test", Description: "Test"}

	output := "No JSON here, just text."

	_, err := ParseDecompositionResult(output, goal)
	if err == nil {
		t.Error("Expected error for output without JSON")
	}
}

func TestMergePlans(t *testing.T) {
	existing := []plan.Plan{
		{ID: 1, Description: "Existing 1"},
		{ID: 2, Description: "Existing 2"},
	}

	generated := []plan.Plan{
		{ID: 3, Description: "Generated 1"},
		{ID: 4, Description: "Generated 2"},
	}

	merged := MergePlans(existing, generated)

	if len(merged) != 4 {
		t.Errorf("Expected 4 merged plans, got %d", len(merged))
	}
}

func TestMergePlansConflictingIDs(t *testing.T) {
	existing := []plan.Plan{
		{ID: 1, Description: "Existing"},
	}

	generated := []plan.Plan{
		{ID: 1, Description: "Generated with conflicting ID"},
	}

	merged := MergePlans(existing, generated)

	// Should have 2 plans with unique IDs
	if len(merged) != 2 {
		t.Errorf("Expected 2 merged plans, got %d", len(merged))
	}

	// Check IDs are unique
	ids := make(map[int]bool)
	for _, p := range merged {
		if ids[p.ID] {
			t.Error("Duplicate ID found after merge")
		}
		ids[p.ID] = true
	}
}

func TestGetNextPlanID(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1},
		{ID: 5},
		{ID: 3},
	}

	nextID := GetNextPlanID(plans)
	if nextID != 6 {
		t.Errorf("Expected next ID 6, got %d", nextID)
	}

	// Empty plans
	nextID = GetNextPlanID(nil)
	if nextID != 1 {
		t.Errorf("Expected next ID 1 for empty plans, got %d", nextID)
	}
}

func TestValidatePlanDependencies(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Description: "Plan 1"},
		{ID: 2, Description: "Plan 2"},
	}

	// Valid dependencies
	deps := map[int][]int{
		2: {1},
	}

	errors := ValidatePlanDependencies(plans, deps)
	if len(errors) > 0 {
		t.Errorf("Expected no errors for valid deps, got: %v", errors)
	}

	// Invalid dependency (unknown plan)
	deps = map[int][]int{
		2: {99},
	}

	errors = ValidatePlanDependencies(plans, deps)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error for unknown plan, got %d", len(errors))
	}

	// Circular dependency
	deps = map[int][]int{
		1: {1},
	}

	errors = ValidatePlanDependencies(plans, deps)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error for circular dep, got %d", len(errors))
	}
}

// Benchmark tests

func BenchmarkCalculateProgress(b *testing.B) {
	plans := make([]plan.Plan, 100)
	planIDs := make([]int, 100)
	for i := 0; i < 100; i++ {
		plans[i] = plan.Plan{ID: i + 1, Description: "Plan", Tested: i%2 == 0}
		planIDs[i] = i + 1
	}

	mgr := NewManager(plans)
	goal := Goal{ID: "test", Description: "Test", GeneratedPlanIDs: planIDs}
	mgr.AddGoal(goal)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.CalculateProgress("test")
	}
}

func BenchmarkGetGoalsByPriority(b *testing.B) {
	mgr := NewManager(nil)
	for i := 0; i < 100; i++ {
		mgr.AddGoal(Goal{
			Description: "Goal",
			Priority:    i % 10,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.GetGoalsByPriority()
	}
}

// Test goal timestamps
func TestGoalTimestamps(t *testing.T) {
	mgr := NewManager(nil)

	before := time.Now()
	goal := Goal{Description: "Test goal"}
	mgr.AddGoal(goal)
	after := time.Now()

	found := mgr.GetGoalByID(mgr.goals[0].ID)

	if found.CreatedAt.Before(before) || found.CreatedAt.After(after) {
		t.Error("CreatedAt should be between before and after timestamps")
	}

	if found.UpdatedAt.Before(before) || found.UpdatedAt.After(after) {
		t.Error("UpdatedAt should be between before and after timestamps")
	}
}
