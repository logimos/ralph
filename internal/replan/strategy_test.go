package replan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/logimos/ralph/internal/plan"
)

func TestParseStrategyType(t *testing.T) {
	tests := []struct {
		input    string
		expected StrategyType
		hasError bool
	}{
		{"incremental", StrategyIncremental, false},
		{"inc", StrategyIncremental, false},
		{"INCREMENTAL", StrategyIncremental, false},
		{"agent", StrategyAgentBased, false},
		{"ai", StrategyAgentBased, false},
		{"AGENT", StrategyAgentBased, false},
		{"none", StrategyNone, false},
		{"off", StrategyNone, false},
		{"", StrategyNone, false},
		{"invalid", "", true},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseStrategyType(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestIncrementalStrategyName(t *testing.T) {
	strategy := NewIncrementalStrategy()
	if strategy.Name() != StrategyIncremental {
		t.Errorf("expected %v, got %v", StrategyIncremental, strategy.Name())
	}
}

func TestIncrementalStrategyDescription(t *testing.T) {
	strategy := NewIncrementalStrategy()
	if strategy.Description() == "" {
		t.Error("description should not be empty")
	}
}

func TestIncrementalStrategyExecute(t *testing.T) {
	strategy := NewIncrementalStrategy()

	tests := []struct {
		name    string
		state   *ReplanState
		trigger TriggerType
	}{
		{
			name: "test failure trigger",
			state: &ReplanState{
				FeatureID:           1,
				ConsecutiveFailures: 3,
				Plans: []plan.Plan{
					{ID: 1, Description: "Feature with many steps", Steps: []string{"1", "2", "3", "4", "5", "6"}},
					{ID: 2, Description: "Feature B"},
				},
			},
			trigger: TriggerTestFailure,
		},
		{
			name: "blocked feature trigger",
			state: &ReplanState{
				FeatureID:       1,
				BlockedFeatures: []int{1},
				Plans: []plan.Plan{
					{ID: 1, Description: "Feature A"},
					{ID: 2, Description: "Feature B"},
				},
			},
			trigger: TriggerBlockedFeature,
		},
		{
			name: "requirement change trigger",
			state: &ReplanState{
				FeatureID: 1,
				Plans: []plan.Plan{
					{ID: 1, Description: "Feature A", Tested: true},
					{ID: 2, Description: "Feature B", Tested: false},
					{ID: 3, Description: "Feature C", Deferred: true},
				},
			},
			trigger: TriggerRequirementChange,
		},
		{
			name: "manual trigger",
			state: &ReplanState{
				Plans: []plan.Plan{
					{ID: 1, Description: "Feature A"},
				},
			},
			trigger: TriggerManual,
		},
		{
			name: "empty plans",
			state: &ReplanState{
				Plans: []plan.Plan{},
			},
			trigger: TriggerManual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := strategy.Execute(tt.state, tt.trigger)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("result should not be nil")
			}
			if result.Strategy != StrategyIncremental {
				t.Errorf("expected strategy %v, got %v", StrategyIncremental, result.Strategy)
			}
			if result.Trigger != tt.trigger {
				t.Errorf("expected trigger %v, got %v", tt.trigger, result.Trigger)
			}
		})
	}
}

func TestAgentBasedStrategyName(t *testing.T) {
	strategy := NewAgentBasedStrategy("test-agent")
	if strategy.Name() != StrategyAgentBased {
		t.Errorf("expected %v, got %v", StrategyAgentBased, strategy.Name())
	}
}

func TestAgentBasedStrategyDescription(t *testing.T) {
	strategy := NewAgentBasedStrategy("test-agent")
	if strategy.Description() == "" {
		t.Error("description should not be empty")
	}
}

func TestAgentBasedStrategyBuildPrompt(t *testing.T) {
	strategy := NewAgentBasedStrategy("test-agent")
	state := &ReplanState{
		FeatureID:           1,
		ConsecutiveFailures: 3,
		BlockedFeatures:     []int{2},
		TotalIterations:     10,
		Plans: []plan.Plan{
			{ID: 1, Description: "Feature A", Category: "infra", Tested: true},
			{ID: 2, Description: "Feature B", Category: "chore", Deferred: true},
			{ID: 3, Description: "Feature C", Category: "data"},
		},
	}

	prompt := strategy.buildReplanPrompt(state, TriggerTestFailure)
	
	if prompt == "" {
		t.Error("prompt should not be empty")
	}
	
	// Check that prompt contains expected information
	if !containsString(prompt, "REPLAN TRIGGER") {
		t.Error("prompt should mention trigger")
	}
	if !containsString(prompt, "CURRENT STATE") {
		t.Error("prompt should mention current state")
	}
	if !containsString(prompt, "CURRENT PLAN") {
		t.Error("prompt should mention current plan")
	}
	if !containsString(prompt, "INSTRUCTIONS") {
		t.Error("prompt should contain instructions")
	}
}

func TestReplanManager(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{
		{ID: 1, Description: "Feature A"},
		{ID: 2, Description: "Feature B"},
	}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	// Create manager
	mgr := NewReplanManager(planPath, "test-agent", false)

	// Test initial state
	state := mgr.GetState()
	if state == nil {
		t.Error("state should not be nil")
	}

	// Test update state
	mgr.UpdateState(1, 2, []string{"test_failure"}, testPlan)
	state = mgr.GetState()
	if state.FeatureID != 1 {
		t.Errorf("expected feature ID 1, got %d", state.FeatureID)
	}
	if state.ConsecutiveFailures != 2 {
		t.Errorf("expected 2 failures, got %d", state.ConsecutiveFailures)
	}

	// Test add blocked feature
	mgr.AddBlockedFeature(1)
	if len(state.BlockedFeatures) != 1 {
		t.Errorf("expected 1 blocked feature, got %d", len(state.BlockedFeatures))
	}
	// Adding same feature again shouldn't duplicate
	mgr.AddBlockedFeature(1)
	if len(state.BlockedFeatures) != 1 {
		t.Errorf("expected 1 blocked feature (no duplicates), got %d", len(state.BlockedFeatures))
	}

	// Test clear blocked features
	mgr.ClearBlockedFeatures()
	if len(state.BlockedFeatures) != 0 {
		t.Errorf("expected 0 blocked features after clear, got %d", len(state.BlockedFeatures))
	}

	// Test increment iterations
	mgr.IncrementIterations()
	if state.TotalIterations != 1 {
		t.Errorf("expected 1 iteration, got %d", state.TotalIterations)
	}

	// Test auto-replan setting
	if mgr.IsAutoReplanEnabled() {
		t.Error("auto-replan should be disabled by default")
	}
	mgr.SetAutoReplan(true)
	if !mgr.IsAutoReplanEnabled() {
		t.Error("auto-replan should be enabled after setting")
	}

	// Test trigger descriptions
	descriptions := mgr.GetTriggerDescriptions()
	if len(descriptions) == 0 {
		t.Error("should have trigger descriptions")
	}

	// Test reset state
	mgr.UpdateState(1, 5, []string{"failure"}, testPlan)
	mgr.ResetState()
	state = mgr.GetState()
	if state.ConsecutiveFailures != 0 {
		t.Errorf("consecutive failures should be 0 after reset, got %d", state.ConsecutiveFailures)
	}
}

func TestReplanManagerCheckTriggers(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{{ID: 1, Description: "Feature A"}}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	mgr := NewReplanManager(planPath, "test-agent", true)

	// No triggers should fire initially
	trigger := mgr.CheckTriggers()
	if trigger != TriggerNone {
		t.Errorf("expected no trigger, got %v", trigger)
	}

	// Test failure trigger should fire after threshold
	mgr.UpdateState(1, 3, []string{"test_failure"}, testPlan)
	trigger = mgr.CheckTriggers()
	if trigger != TriggerTestFailure {
		t.Errorf("expected test failure trigger, got %v", trigger)
	}

	// Reset and test blocked feature trigger
	mgr.ResetState()
	mgr.AddBlockedFeature(1)
	trigger = mgr.CheckTriggers()
	if trigger != TriggerBlockedFeature {
		t.Errorf("expected blocked feature trigger, got %v", trigger)
	}
}

func TestReplanManagerShouldReplan(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{{ID: 1, Description: "Feature A"}}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	// With auto-replan disabled
	mgr := NewReplanManager(planPath, "test-agent", false)
	mgr.UpdateState(1, 5, []string{"test_failure"}, testPlan)
	shouldReplan, _ := mgr.ShouldReplan()
	if shouldReplan {
		t.Error("should not replan when auto-replan is disabled")
	}

	// With auto-replan enabled
	mgr.SetAutoReplan(true)
	shouldReplan, trigger := mgr.ShouldReplan()
	if !shouldReplan {
		t.Error("should replan when auto-replan is enabled and trigger fires")
	}
	if trigger != TriggerTestFailure {
		t.Errorf("expected test failure trigger, got %v", trigger)
	}
}

func TestReplanManagerExecuteReplan(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{
		{ID: 1, Description: "Feature A"},
		{ID: 2, Description: "Feature B"},
	}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	mgr := NewReplanManager(planPath, "test-agent", true)
	mgr.UpdateState(1, 3, []string{"test_failure"}, testPlan)

	// Execute incremental replan
	result, err := mgr.ExecuteReplan(StrategyIncremental, TriggerTestFailure)
	if err != nil {
		t.Fatalf("replan failed: %v", err)
	}

	if !result.Success {
		t.Error("replan should succeed")
	}
	if result.OldPlanPath == "" {
		t.Error("should have backup path")
	}
	if result.Strategy != StrategyIncremental {
		t.Errorf("expected incremental strategy, got %v", result.Strategy)
	}

	// Verify backup was created
	versions := mgr.GetVersions()
	if len(versions) == 0 {
		t.Error("should have created a backup version")
	}
}

func TestReplanManagerManualReplan(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{{ID: 1, Description: "Feature A"}}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	mgr := NewReplanManager(planPath, "test-agent", false)
	mgr.UpdateState(1, 0, nil, testPlan)

	result, err := mgr.ManualReplan(StrategyIncremental)
	if err != nil {
		t.Fatalf("manual replan failed: %v", err)
	}

	if result.Trigger != TriggerManual {
		t.Errorf("expected manual trigger, got %v", result.Trigger)
	}
}

func TestReplanManagerRestoreVersion(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{{ID: 1, Description: "Original"}}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	mgr := NewReplanManager(planPath, "test-agent", true)
	mgr.UpdateState(1, 0, nil, testPlan)

	// Create backup
	_, err = mgr.ExecuteReplan(StrategyIncremental, TriggerManual)
	if err != nil {
		t.Fatal(err)
	}

	// Modify plan
	testPlan[0].Description = "Modified"
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	// Restore
	if err := mgr.RestoreVersion(1); err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	// Verify
	restored, err := plan.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if restored[0].Description != "Original" {
		t.Error("restore did not work")
	}
}

func TestContainsAnyWord(t *testing.T) {
	tests := []struct {
		s        string
		words    []string
		expected bool
	}{
		{"this is a testing string", []string{"testing", "other"}, true}, // "testing" is >4 chars
		{"this is a test string", []string{"short", "other"}, false},     // "short" is exactly 5, but not in string
		{"this is a test string", []string{"str", "abc"}, false},         // "str" is too short (<=4)
		{"authentication service", []string{"authentication"}, true},
		{"", []string{"testing"}, false},
		{"testing", []string{}, false},
		{"string testing here", []string{"string"}, true}, // "string" is >4 chars and in the string
	}

	for _, tt := range tests {
		result := containsAnyWord(tt.s, tt.words)
		if result != tt.expected {
			t.Errorf("containsAnyWord(%q, %v) = %v, want %v", tt.s, tt.words, result, tt.expected)
		}
	}
}

func TestReplanResultTimestamp(t *testing.T) {
	strategy := NewIncrementalStrategy()
	state := &ReplanState{
		Plans: []plan.Plan{{ID: 1, Description: "Test"}},
	}

	result, _ := strategy.Execute(state, TriggerManual)
	
	if result.Timestamp.IsZero() {
		t.Error("result timestamp should be set")
	}
}
