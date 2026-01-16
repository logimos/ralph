package replan

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/logimos/ralph/internal/plan"
)

func TestTestFailureTrigger(t *testing.T) {
	tests := []struct {
		name               string
		threshold          int
		consecutiveFailures int
		expected           bool
	}{
		{
			name:               "below threshold",
			threshold:          3,
			consecutiveFailures: 2,
			expected:           false,
		},
		{
			name:               "at threshold",
			threshold:          3,
			consecutiveFailures: 3,
			expected:           true,
		},
		{
			name:               "above threshold",
			threshold:          3,
			consecutiveFailures: 5,
			expected:           true,
		},
		{
			name:               "zero failures",
			threshold:          3,
			consecutiveFailures: 0,
			expected:           false,
		},
		{
			name:               "default threshold",
			threshold:          0,
			consecutiveFailures: 3,
			expected:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := NewTestFailureTrigger(tt.threshold)
			state := &ReplanState{
				ConsecutiveFailures: tt.consecutiveFailures,
			}
			result := trigger.Check(state)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTestFailureTriggerName(t *testing.T) {
	trigger := NewTestFailureTrigger(3)
	if trigger.Name() != TriggerTestFailure {
		t.Errorf("expected %v, got %v", TriggerTestFailure, trigger.Name())
	}
}

func TestRequirementChangeTrigger(t *testing.T) {
	tests := []struct {
		name         string
		planHash     string
		lastPlanHash string
		expected     bool
	}{
		{
			name:         "no change",
			planHash:     "abc123",
			lastPlanHash: "abc123",
			expected:     false,
		},
		{
			name:         "hash changed",
			planHash:     "abc123",
			lastPlanHash: "def456",
			expected:     true,
		},
		{
			name:         "no previous hash",
			planHash:     "abc123",
			lastPlanHash: "",
			expected:     false,
		},
		{
			name:         "no current hash",
			planHash:     "",
			lastPlanHash: "abc123",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := NewRequirementChangeTrigger()
			state := &ReplanState{
				PlanHash:     tt.planHash,
				LastPlanHash: tt.lastPlanHash,
			}
			result := trigger.Check(state)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBlockedFeatureTrigger(t *testing.T) {
	tests := []struct {
		name            string
		minBlocked      int
		blockedFeatures []int
		expected        bool
	}{
		{
			name:            "no blocked features",
			minBlocked:      1,
			blockedFeatures: []int{},
			expected:        false,
		},
		{
			name:            "one blocked feature",
			minBlocked:      1,
			blockedFeatures: []int{1},
			expected:        true,
		},
		{
			name:            "multiple blocked features",
			minBlocked:      2,
			blockedFeatures: []int{1, 2, 3},
			expected:        true,
		},
		{
			name:            "below minimum",
			minBlocked:      3,
			blockedFeatures: []int{1, 2},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := NewBlockedFeatureTrigger(tt.minBlocked)
			state := &ReplanState{
				BlockedFeatures: tt.blockedFeatures,
			}
			result := trigger.Check(state)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestManualTrigger(t *testing.T) {
	trigger := NewManualTrigger()
	state := &ReplanState{}

	// Initially not triggered
	if trigger.Check(state) {
		t.Error("manual trigger should not fire initially")
	}

	// After activation
	trigger.Activate()
	if !trigger.Check(state) {
		t.Error("manual trigger should fire after activation")
	}

	// After reset
	trigger.Reset()
	if trigger.Check(state) {
		t.Error("manual trigger should not fire after reset")
	}
}

func TestPlanVersioner(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{
		{ID: 1, Description: "Test feature"},
	}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	versioner := NewPlanVersioner(planPath)

	// Test creating backup
	backupPath, err := versioner.CreateBackup(TriggerManual)
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	if backupPath == "" {
		t.Error("backup path should not be empty")
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file should exist")
	}

	// Test get versions
	versions := versioner.GetVersions()
	if len(versions) != 1 {
		t.Errorf("expected 1 version, got %d", len(versions))
	}

	// Test get latest version
	latest := versioner.GetLatestVersion()
	if latest == nil {
		t.Error("latest version should not be nil")
	}
	if latest.Version != 1 {
		t.Errorf("expected version 1, got %d", latest.Version)
	}

	// Test duplicate backup (same content)
	backupPath2, err := versioner.CreateBackup(TriggerManual)
	if err != nil {
		t.Fatalf("failed to create duplicate backup: %v", err)
	}
	if backupPath2 != backupPath {
		t.Error("duplicate backup should return same path")
	}

	// Modify plan and create new backup
	testPlan[0].Description = "Modified feature"
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	backupPath3, err := versioner.CreateBackup(TriggerTestFailure)
	if err != nil {
		t.Fatalf("failed to create new backup: %v", err)
	}
	if backupPath3 == backupPath {
		t.Error("new backup should have different path")
	}

	// Test restore version
	if err := versioner.RestoreVersion(1); err != nil {
		t.Fatalf("failed to restore version: %v", err)
	}

	// Verify restore worked
	restored, err := plan.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if restored[0].Description != "Test feature" {
		t.Error("restore did not work correctly")
	}

	// Test invalid restore
	if err := versioner.RestoreVersion(999); err == nil {
		t.Error("expected error for invalid version")
	}
}

func TestComputeDiff(t *testing.T) {
	oldPlans := []plan.Plan{
		{ID: 1, Description: "Feature A", Category: "infra"},
		{ID: 2, Description: "Feature B", Category: "chore"},
		{ID: 3, Description: "Feature C", Tested: false},
	}

	newPlans := []plan.Plan{
		{ID: 1, Description: "Feature A Modified", Category: "infra"},
		{ID: 2, Description: "Feature B", Category: "data"},
		{ID: 4, Description: "Feature D", Category: "new"},
	}

	diff := ComputeDiff(oldPlans, newPlans)

	// Check added
	if len(diff.Added) != 1 {
		t.Errorf("expected 1 added, got %d", len(diff.Added))
	}
	if diff.Added[0].ID != 4 {
		t.Error("expected feature 4 to be added")
	}

	// Check removed
	if len(diff.Removed) != 1 {
		t.Errorf("expected 1 removed, got %d", len(diff.Removed))
	}
	if diff.Removed[0].ID != 3 {
		t.Error("expected feature 3 to be removed")
	}

	// Check modified (should have changes for features 1 and 2)
	if len(diff.Modified) < 2 {
		t.Errorf("expected at least 2 modifications, got %d", len(diff.Modified))
	}

	// Check IsEmpty
	if diff.IsEmpty() {
		t.Error("diff should not be empty")
	}

	// Check empty diff
	emptyDiff := ComputeDiff(oldPlans, oldPlans)
	if !emptyDiff.IsEmpty() {
		t.Error("diff of same plans should be empty")
	}
}

func TestPlanDiffSummary(t *testing.T) {
	diff := &PlanDiff{
		Added: []plan.Plan{
			{ID: 4, Description: "New feature"},
		},
		Removed: []plan.Plan{
			{ID: 3, Description: "Old feature"},
		},
		Modified: []PlanChange{
			{ID: 1, Field: "description", OldValue: "old", NewValue: "new"},
		},
	}

	summary := diff.Summary()
	if summary == "" {
		t.Error("summary should not be empty")
	}

	// Check that summary contains expected elements
	if !containsString(summary, "Added: 1") {
		t.Error("summary should mention added features")
	}
	if !containsString(summary, "Removed: 1") {
		t.Error("summary should mention removed features")
	}
	if !containsString(summary, "Modified: 1") {
		t.Error("summary should mention modified features")
	}
}

func TestCalculatePlanHash(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{
		{ID: 1, Description: "Test feature"},
	}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	hash1, err := CalculatePlanHash(planPath)
	if err != nil {
		t.Fatalf("failed to calculate hash: %v", err)
	}
	if hash1 == "" {
		t.Error("hash should not be empty")
	}

	// Same content should have same hash
	hash2, err := CalculatePlanHash(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if hash1 != hash2 {
		t.Error("same file should have same hash")
	}

	// Different content should have different hash
	testPlan[0].Description = "Modified"
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}
	hash3, err := CalculatePlanHash(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if hash1 == hash3 {
		t.Error("different content should have different hash")
	}
}

func TestCalculatePlansHash(t *testing.T) {
	plans1 := []plan.Plan{
		{ID: 1, Description: "Test"},
	}
	plans2 := []plan.Plan{
		{ID: 1, Description: "Test"},
	}
	plans3 := []plan.Plan{
		{ID: 1, Description: "Modified"},
	}

	hash1 := CalculatePlansHash(plans1)
	hash2 := CalculatePlansHash(plans2)
	hash3 := CalculatePlansHash(plans3)

	if hash1 == "" {
		t.Error("hash should not be empty")
	}
	if hash1 != hash2 {
		t.Error("same plans should have same hash")
	}
	if hash1 == hash3 {
		t.Error("different plans should have different hash")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestTriggerDescriptions(t *testing.T) {
	triggers := []ReplanTrigger{
		NewTestFailureTrigger(3),
		NewRequirementChangeTrigger(),
		NewBlockedFeatureTrigger(1),
		NewManualTrigger(),
	}

	for _, trigger := range triggers {
		if trigger.Description() == "" {
			t.Errorf("trigger %s should have a description", trigger.Name())
		}
	}
}

func TestPlanVersionTimestamp(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "replan_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test plan file
	planPath := filepath.Join(tmpDir, "plan.json")
	testPlan := []plan.Plan{{ID: 1, Description: "Test"}}
	if err := plan.WriteFile(planPath, testPlan); err != nil {
		t.Fatal(err)
	}

	versioner := NewPlanVersioner(planPath)
	before := time.Now()
	_, err = versioner.CreateBackup(TriggerManual)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now()

	latest := versioner.GetLatestVersion()
	if latest.Timestamp.Before(before) || latest.Timestamp.After(after) {
		t.Error("version timestamp should be within test execution time")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
