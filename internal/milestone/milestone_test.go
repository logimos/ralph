package milestone

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/logimos/ralph/internal/plan"
)

func TestNewManager(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Description: "Test feature 1"},
		{ID: 2, Description: "Test feature 2"},
	}
	
	mgr := NewManager(plans)
	
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
	if len(mgr.plans) != 2 {
		t.Errorf("Expected 2 plans, got %d", len(mgr.plans))
	}
}

func TestExtractMilestonesFromPlans(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Description: "Feature 1", Milestone: "Alpha"},
		{ID: 2, Description: "Feature 2", Milestone: "Alpha"},
		{ID: 3, Description: "Feature 3", Milestone: "Beta"},
		{ID: 4, Description: "Feature 4"}, // No milestone
	}
	
	mgr := NewManager(plans)
	mgr.ExtractMilestonesFromPlans()
	
	milestones := mgr.GetMilestones()
	if len(milestones) != 2 {
		t.Errorf("Expected 2 milestones, got %d", len(milestones))
	}
	
	// Check that both Alpha and Beta are present
	found := make(map[string]bool)
	for _, m := range milestones {
		found[m.Name] = true
	}
	
	if !found["Alpha"] {
		t.Error("Expected milestone 'Alpha' to be extracted")
	}
	if !found["Beta"] {
		t.Error("Expected milestone 'Beta' to be extracted")
	}
}

func TestGetFeaturesForMilestone(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Description: "Feature 1", Milestone: "Alpha", MilestoneOrder: 2},
		{ID: 2, Description: "Feature 2", Milestone: "Alpha", MilestoneOrder: 1},
		{ID: 3, Description: "Feature 3", Milestone: "Beta"},
		{ID: 4, Description: "Feature 4"}, // No milestone
	}
	
	mgr := NewManager(plans)
	
	// Test getting Alpha features
	alphaFeatures := mgr.GetFeaturesForMilestone("Alpha")
	if len(alphaFeatures) != 2 {
		t.Errorf("Expected 2 Alpha features, got %d", len(alphaFeatures))
	}
	
	// Check ordering (should be sorted by MilestoneOrder)
	if len(alphaFeatures) >= 2 {
		if alphaFeatures[0].ID != 2 {
			t.Errorf("Expected first Alpha feature to be ID 2 (order 1), got ID %d", alphaFeatures[0].ID)
		}
		if alphaFeatures[1].ID != 1 {
			t.Errorf("Expected second Alpha feature to be ID 1 (order 2), got ID %d", alphaFeatures[1].ID)
		}
	}
	
	// Test case-insensitive matching
	alphaFeaturesLower := mgr.GetFeaturesForMilestone("alpha")
	if len(alphaFeaturesLower) != 2 {
		t.Errorf("Expected case-insensitive match for 'alpha', got %d features", len(alphaFeaturesLower))
	}
	
	// Test non-existent milestone
	noFeatures := mgr.GetFeaturesForMilestone("Gamma")
	if len(noFeatures) != 0 {
		t.Errorf("Expected 0 features for non-existent milestone, got %d", len(noFeatures))
	}
}

func TestGetFeaturesForMilestoneWithExplicitIDs(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Description: "Feature 1"},
		{ID: 2, Description: "Feature 2"},
		{ID: 3, Description: "Feature 3"},
	}
	
	milestones := []Milestone{
		{
			ID:       "alpha",
			Name:     "Alpha",
			Features: []int{1, 3}, // Explicitly list feature IDs
		},
	}
	
	mgr := NewManager(plans)
	mgr.SetMilestones(milestones)
	
	features := mgr.GetFeaturesForMilestone("Alpha")
	if len(features) != 2 {
		t.Errorf("Expected 2 features from explicit IDs, got %d", len(features))
	}
	
	// Check that the correct features are included
	featureIDs := make(map[int]bool)
	for _, f := range features {
		featureIDs[f.ID] = true
	}
	
	if !featureIDs[1] {
		t.Error("Expected feature 1 to be included")
	}
	if !featureIDs[3] {
		t.Error("Expected feature 3 to be included")
	}
	if featureIDs[2] {
		t.Error("Feature 2 should not be included")
	}
}

func TestCalculateProgress(t *testing.T) {
	tests := []struct {
		name              string
		plans             []plan.Plan
		milestoneName     string
		wantTotal         int
		wantCompleted     int
		wantPercentage    float64
		wantStatus        Status
	}{
		{
			name: "all complete",
			plans: []plan.Plan{
				{ID: 1, Milestone: "Alpha", Tested: true},
				{ID: 2, Milestone: "Alpha", Tested: true},
			},
			milestoneName:  "Alpha",
			wantTotal:      2,
			wantCompleted:  2,
			wantPercentage: 100,
			wantStatus:     StatusComplete,
		},
		{
			name: "partially complete",
			plans: []plan.Plan{
				{ID: 1, Milestone: "Alpha", Tested: true},
				{ID: 2, Milestone: "Alpha", Tested: false},
			},
			milestoneName:  "Alpha",
			wantTotal:      2,
			wantCompleted:  1,
			wantPercentage: 50,
			wantStatus:     StatusInProgress,
		},
		{
			name: "none complete",
			plans: []plan.Plan{
				{ID: 1, Milestone: "Alpha", Tested: false},
				{ID: 2, Milestone: "Alpha", Tested: false},
			},
			milestoneName:  "Alpha",
			wantTotal:      2,
			wantCompleted:  0,
			wantPercentage: 0,
			wantStatus:     StatusNotStarted,
		},
		{
			name: "empty milestone",
			plans: []plan.Plan{
				{ID: 1, Milestone: "Beta", Tested: true},
			},
			milestoneName:  "Alpha",
			wantTotal:      0,
			wantCompleted:  0,
			wantPercentage: 0,
			wantStatus:     StatusNotStarted,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.plans)
			progress := mgr.CalculateProgress(tt.milestoneName)
			
			if progress.TotalFeatures != tt.wantTotal {
				t.Errorf("TotalFeatures = %d, want %d", progress.TotalFeatures, tt.wantTotal)
			}
			if progress.CompletedFeatures != tt.wantCompleted {
				t.Errorf("CompletedFeatures = %d, want %d", progress.CompletedFeatures, tt.wantCompleted)
			}
			if progress.Percentage != tt.wantPercentage {
				t.Errorf("Percentage = %f, want %f", progress.Percentage, tt.wantPercentage)
			}
			if progress.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", progress.Status, tt.wantStatus)
			}
		})
	}
}

func TestCalculateAllProgress(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Milestone: "Alpha", Tested: true},
		{ID: 2, Milestone: "Alpha", Tested: false},
		{ID: 3, Milestone: "Beta", Tested: true},
		{ID: 4, Milestone: "Beta", Tested: true},
	}
	
	mgr := NewManager(plans)
	allProgress := mgr.CalculateAllProgress()
	
	if len(allProgress) != 2 {
		t.Errorf("Expected 2 milestone progress entries, got %d", len(allProgress))
	}
	
	// Check that we have both Alpha and Beta
	progressMap := make(map[string]*Progress)
	for _, p := range allProgress {
		progressMap[p.Milestone.Name] = p
	}
	
	if alpha, ok := progressMap["Alpha"]; ok {
		if alpha.CompletedFeatures != 1 || alpha.TotalFeatures != 2 {
			t.Errorf("Alpha progress wrong: %d/%d", alpha.CompletedFeatures, alpha.TotalFeatures)
		}
	} else {
		t.Error("Alpha milestone not found")
	}
	
	if beta, ok := progressMap["Beta"]; ok {
		if beta.CompletedFeatures != 2 || beta.TotalFeatures != 2 {
			t.Errorf("Beta progress wrong: %d/%d", beta.CompletedFeatures, beta.TotalFeatures)
		}
		if beta.Status != StatusComplete {
			t.Errorf("Beta should be complete, got status %v", beta.Status)
		}
	} else {
		t.Error("Beta milestone not found")
	}
}

func TestGetCompletedMilestones(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Milestone: "Alpha", Tested: true},
		{ID: 2, Milestone: "Alpha", Tested: false},
		{ID: 3, Milestone: "Beta", Tested: true},
		{ID: 4, Milestone: "Beta", Tested: true},
	}
	
	mgr := NewManager(plans)
	completed := mgr.GetCompletedMilestones()
	
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed milestone, got %d", len(completed))
	}
	
	if len(completed) > 0 && completed[0].Milestone.Name != "Beta" {
		t.Errorf("Expected Beta to be completed, got %s", completed[0].Milestone.Name)
	}
}

func TestGetNextMilestoneToComplete(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Milestone: "Alpha", Tested: true},  // 50%
		{ID: 2, Milestone: "Alpha", Tested: false},
		{ID: 3, Milestone: "Beta", Tested: true},   // 100%
		{ID: 4, Milestone: "Beta", Tested: true},
		{ID: 5, Milestone: "Gamma", Tested: true},  // 33%
		{ID: 6, Milestone: "Gamma", Tested: false},
		{ID: 7, Milestone: "Gamma", Tested: false},
	}
	
	mgr := NewManager(plans)
	next := mgr.GetNextMilestoneToComplete()
	
	if next == nil {
		t.Fatal("Expected a milestone to be returned")
	}
	
	// Alpha is 50%, Gamma is 33%, so Alpha should be next
	if next.Milestone.Name != "Alpha" {
		t.Errorf("Expected Alpha (50%%) to be next, got %s (%.0f%%)", 
			next.Milestone.Name, next.Percentage)
	}
}

func TestFormatProgress(t *testing.T) {
	tests := []struct {
		progress *Progress
		want     string
	}{
		{
			progress: &Progress{
				Milestone:         &Milestone{Name: "Alpha"},
				CompletedFeatures: 3,
				TotalFeatures:     5,
				Percentage:        60,
				Status:            StatusInProgress,
			},
			want: "◐ Alpha: 3/5 (60%)",
		},
		{
			progress: &Progress{
				Milestone:         &Milestone{Name: "Beta"},
				CompletedFeatures: 5,
				TotalFeatures:     5,
				Percentage:        100,
				Status:            StatusComplete,
			},
			want: "● Beta: 5/5 (100%)",
		},
		{
			progress: &Progress{
				Milestone:         &Milestone{Name: "Gamma"},
				CompletedFeatures: 0,
				TotalFeatures:     3,
				Percentage:        0,
				Status:            StatusNotStarted,
			},
			want: "○ Gamma: 0/3 (0%)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.progress.Milestone.Name, func(t *testing.T) {
			got := FormatProgress(tt.progress)
			if got != tt.want {
				t.Errorf("FormatProgress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatProgressBar(t *testing.T) {
	tests := []struct {
		name       string
		progress   *Progress
		width      int
		wantPrefix string
	}{
		{
			name: "50%",
			progress: &Progress{
				Percentage: 50,
			},
			width:      10,
			wantPrefix: "[█████░░░░░]",
		},
		{
			name: "100%",
			progress: &Progress{
				Percentage: 100,
			},
			width:      10,
			wantPrefix: "[██████████]",
		},
		{
			name: "0%",
			progress: &Progress{
				Percentage: 0,
			},
			width:      10,
			wantPrefix: "[░░░░░░░░░░]",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatProgressBar(tt.progress, tt.width)
			if !containsSubstring(got, tt.wantPrefix) {
				t.Errorf("FormatProgressBar() = %q, want prefix %q", got, tt.wantPrefix)
			}
		})
	}
}

func TestCelebrationMessage(t *testing.T) {
	msg := CelebrationMessage("Alpha")
	if msg == "" {
		t.Error("Expected a celebration message, got empty string")
	}
	if !containsSubstring(msg, "Alpha") {
		t.Errorf("Celebration message should contain milestone name, got %q", msg)
	}
}

func TestLoadMilestones(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "milestone_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Test loading milestones from JSON file
	milestonesJSON := `[
		{"id": "alpha", "name": "Alpha", "description": "First milestone"},
		{"id": "beta", "name": "Beta", "description": "Second milestone"}
	]`
	
	milestonesPath := filepath.Join(tmpDir, "milestones.json")
	if err := os.WriteFile(milestonesPath, []byte(milestonesJSON), 0644); err != nil {
		t.Fatal(err)
	}
	
	mgr := NewManager(nil)
	if err := mgr.LoadMilestones(milestonesPath); err != nil {
		t.Fatalf("LoadMilestones failed: %v", err)
	}
	
	milestones := mgr.GetMilestones()
	if len(milestones) != 2 {
		t.Errorf("Expected 2 milestones, got %d", len(milestones))
	}
}

func TestLoadMilestonesFromEmbeddedFormat(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "milestone_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Test loading milestones from embedded format in plan.json
	embeddedJSON := `{
		"milestones": [
			{"id": "alpha", "name": "Alpha", "features": [1, 2]},
			{"id": "beta", "name": "Beta", "features": [3]}
		]
	}`
	
	planPath := filepath.Join(tmpDir, "plan.json")
	if err := os.WriteFile(planPath, []byte(embeddedJSON), 0644); err != nil {
		t.Fatal(err)
	}
	
	mgr := NewManager(nil)
	if err := mgr.LoadMilestones(planPath); err != nil {
		t.Fatalf("LoadMilestones failed: %v", err)
	}
	
	milestones := mgr.GetMilestones()
	if len(milestones) != 2 {
		t.Errorf("Expected 2 milestones, got %d", len(milestones))
	}
}

func TestHasMilestones(t *testing.T) {
	tests := []struct {
		name       string
		plans      []plan.Plan
		milestones []Milestone
		want       bool
	}{
		{
			name: "has milestone in plans",
			plans: []plan.Plan{
				{ID: 1, Milestone: "Alpha"},
			},
			want: true,
		},
		{
			name: "has explicit milestones",
			plans: []plan.Plan{
				{ID: 1},
			},
			milestones: []Milestone{
				{ID: "alpha", Name: "Alpha"},
			},
			want: true,
		},
		{
			name: "no milestones",
			plans: []plan.Plan{
				{ID: 1},
				{ID: 2},
			},
			want: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.plans)
			if len(tt.milestones) > 0 {
				mgr.SetMilestones(tt.milestones)
			}
			
			got := mgr.HasMilestones()
			if got != tt.want {
				t.Errorf("HasMilestones() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Milestone: "Alpha", Tested: true},
		{ID: 2, Milestone: "Alpha", Tested: false},
		{ID: 3, Milestone: "Beta", Tested: true},
	}
	
	mgr := NewManager(plans)
	summary := mgr.Summary()
	
	if summary == "" {
		t.Error("Expected non-empty summary")
	}
	
	// Check that summary contains expected content
	if !containsSubstring(summary, "Alpha") {
		t.Error("Summary should contain 'Alpha'")
	}
	if !containsSubstring(summary, "Beta") {
		t.Error("Summary should contain 'Beta'")
	}
	if !containsSubstring(summary, "Overall") {
		t.Error("Summary should contain 'Overall'")
	}
}

func TestSummaryNoMilestones(t *testing.T) {
	plans := []plan.Plan{
		{ID: 1, Tested: true},
		{ID: 2, Tested: false},
	}
	
	mgr := NewManager(plans)
	summary := mgr.Summary()
	
	if !containsSubstring(summary, "No milestones defined") {
		t.Errorf("Expected 'No milestones defined', got: %s", summary)
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
