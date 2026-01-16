package nudge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "default path when empty",
			path:     "",
			wantPath: DefaultNudgeFile,
		},
		{
			name:     "custom path",
			path:     "custom/nudges.json",
			wantPath: "custom/nudges.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore(tt.path)
			if store.Path() != tt.wantPath {
				t.Errorf("NewStore().Path() = %v, want %v", store.Path(), tt.wantPath)
			}
		})
	}
}

func TestStoreLoadSave(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nudgePath := filepath.Join(tmpDir, "nudges.json")
	store := NewStore(nudgePath)

	// Test load when file doesn't exist (should initialize empty)
	if err := store.Load(); err != nil {
		t.Fatalf("Load() with no file failed: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("Expected 0 nudges, got %d", store.Count())
	}

	// Add some nudges
	n1, err := store.Add(NudgeTypeFocus, "Focus on feature 5", 10)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}
	if n1.Type != NudgeTypeFocus {
		t.Errorf("Expected type %v, got %v", NudgeTypeFocus, n1.Type)
	}

	n2, err := store.Add(NudgeTypeConstraint, "Don't use external libs", 5)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if store.Count() != 2 {
		t.Errorf("Expected 2 nudges, got %d", store.Count())
	}

	// Verify file was created and is valid JSON
	data, err := os.ReadFile(nudgePath)
	if err != nil {
		t.Fatalf("Failed to read nudge file: %v", err)
	}

	var nf NudgeFile
	if err := json.Unmarshal(data, &nf); err != nil {
		t.Fatalf("File is not valid JSON: %v", err)
	}

	if len(nf.Nudges) != 2 {
		t.Errorf("Expected 2 nudges in file, got %d", len(nf.Nudges))
	}

	// Test loading existing file
	store2 := NewStore(nudgePath)
	if err := store2.Load(); err != nil {
		t.Fatalf("Load() of existing file failed: %v", err)
	}

	if store2.Count() != 2 {
		t.Errorf("Expected 2 nudges after reload, got %d", store2.Count())
	}

	// Verify the nudges were loaded correctly
	all := store2.GetAll()
	found := false
	for _, n := range all {
		if n.Content == "Focus on feature 5" && n.Type == NudgeTypeFocus && n.Priority == 10 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find the focus nudge after reload")
	}

	// Test with n2
	_ = n2 // Used to verify second nudge was added
}

func TestStoreGetActive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	// Add nudges with different priorities
	store.Add(NudgeTypeFocus, "Low priority", 1)
	store.Add(NudgeTypeFocus, "High priority", 10)
	store.Add(NudgeTypeFocus, "Medium priority", 5)

	active := store.GetActive()
	if len(active) != 3 {
		t.Fatalf("Expected 3 active nudges, got %d", len(active))
	}

	// Should be sorted by priority (highest first)
	if active[0].Priority != 10 {
		t.Errorf("Expected highest priority first, got priority %d", active[0].Priority)
	}
	if active[1].Priority != 5 {
		t.Errorf("Expected medium priority second, got priority %d", active[1].Priority)
	}
	if active[2].Priority != 1 {
		t.Errorf("Expected lowest priority last, got priority %d", active[2].Priority)
	}
}

func TestStoreAcknowledge(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	n, err := store.Add(NudgeTypeFocus, "Test nudge", 0)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if store.ActiveCount() != 1 {
		t.Errorf("Expected 1 active, got %d", store.ActiveCount())
	}

	// Acknowledge the nudge
	if err := store.Acknowledge(n.ID); err != nil {
		t.Fatalf("Acknowledge() failed: %v", err)
	}

	if store.ActiveCount() != 0 {
		t.Errorf("Expected 0 active after acknowledge, got %d", store.ActiveCount())
	}

	// Verify acknowledged nudge is no longer in active list
	active := store.GetActive()
	if len(active) != 0 {
		t.Errorf("Expected no active nudges, got %d", len(active))
	}

	// But it should still be in GetAll
	all := store.GetAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 total nudge, got %d", len(all))
	}
	if !all[0].Acknowledged {
		t.Error("Expected nudge to be acknowledged")
	}
}

func TestStoreAcknowledgeAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	store.Add(NudgeTypeFocus, "Nudge 1", 0)
	store.Add(NudgeTypeSkip, "Nudge 2", 0)
	store.Add(NudgeTypeConstraint, "Nudge 3", 0)

	if store.ActiveCount() != 3 {
		t.Errorf("Expected 3 active, got %d", store.ActiveCount())
	}

	if err := store.AcknowledgeAll(); err != nil {
		t.Fatalf("AcknowledgeAll() failed: %v", err)
	}

	if store.ActiveCount() != 0 {
		t.Errorf("Expected 0 active after AcknowledgeAll, got %d", store.ActiveCount())
	}
}

func TestStoreClear(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	store.Add(NudgeTypeFocus, "Test 1", 0)
	store.Add(NudgeTypeFocus, "Test 2", 0)

	if store.Count() != 2 {
		t.Errorf("Expected 2 nudges, got %d", store.Count())
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("Expected 0 nudges after clear, got %d", store.Count())
	}

	// Verify file was updated
	store2 := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store2.Load()
	if store2.Count() != 0 {
		t.Errorf("Expected 0 nudges in reloaded store, got %d", store2.Count())
	}
}

func TestStoreGetByType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	store.Add(NudgeTypeFocus, "Focus 1", 0)
	store.Add(NudgeTypeFocus, "Focus 2", 0)
	store.Add(NudgeTypeSkip, "Skip 1", 0)
	store.Add(NudgeTypeConstraint, "Constraint 1", 0)

	focus := store.GetByType(NudgeTypeFocus)
	if len(focus) != 2 {
		t.Errorf("Expected 2 focus nudges, got %d", len(focus))
	}

	skip := store.GetByType(NudgeTypeSkip)
	if len(skip) != 1 {
		t.Errorf("Expected 1 skip nudge, got %d", len(skip))
	}

	style := store.GetByType(NudgeTypeStyle)
	if len(style) != 0 {
		t.Errorf("Expected 0 style nudges, got %d", len(style))
	}
}

func TestStoreHasChanged(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nudgePath := filepath.Join(tmpDir, "nudges.json")
	store := NewStore(nudgePath)
	store.Load()
	store.Add(NudgeTypeFocus, "Initial nudge", 0)

	// Should not have changed immediately after save
	if store.HasChanged() {
		t.Error("Expected no change right after save")
	}

	// Simulate external modification
	time.Sleep(100 * time.Millisecond)
	nf := NudgeFile{
		Nudges: []Nudge{
			{ID: "ext_1", Type: NudgeTypeSkip, Content: "External nudge", CreatedAt: time.Now()},
		},
		LastUpdated: time.Now(),
	}
	data, _ := json.Marshal(nf)
	os.WriteFile(nudgePath, data, 0644)

	// Now should detect change
	if !store.HasChanged() {
		t.Error("Expected change after external modification")
	}
}

func TestStoreReload(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nudgePath := filepath.Join(tmpDir, "nudges.json")
	store := NewStore(nudgePath)
	store.Load()
	store.Add(NudgeTypeFocus, "Initial", 0)

	// Reload when no change
	changed, err := store.Reload()
	if err != nil {
		t.Fatalf("Reload() failed: %v", err)
	}
	if changed {
		t.Error("Expected no change on reload")
	}

	// Simulate external modification
	time.Sleep(100 * time.Millisecond)
	nf := NudgeFile{
		Nudges: []Nudge{
			{ID: "new_1", Type: NudgeTypeConstraint, Content: "New constraint", CreatedAt: time.Now()},
			{ID: "new_2", Type: NudgeTypeStyle, Content: "New style", CreatedAt: time.Now()},
		},
		LastUpdated: time.Now(),
	}
	data, _ := json.Marshal(nf)
	os.WriteFile(nudgePath, data, 0644)

	// Reload should detect and load changes
	changed, err = store.Reload()
	if err != nil {
		t.Fatalf("Reload() after change failed: %v", err)
	}
	if !changed {
		t.Error("Expected change on reload")
	}

	if store.Count() != 2 {
		t.Errorf("Expected 2 nudges after reload, got %d", store.Count())
	}
}

func TestBuildPromptContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	// Empty store should return empty string
	ctx := store.BuildPromptContext()
	if ctx != "" {
		t.Errorf("Expected empty context for empty store, got %q", ctx)
	}

	// Add some nudges
	store.Add(NudgeTypeFocus, "Work on feature 5", 10)
	store.Add(NudgeTypeConstraint, "No external dependencies", 5)
	store.Add(NudgeTypeStyle, "Use camelCase", 0)
	store.Add(NudgeTypeSkip, "Skip feature 3", 0)

	ctx = store.BuildPromptContext()
	if ctx == "" {
		t.Error("Expected non-empty context")
	}

	// Check for expected content
	if !contains(ctx, "[USER GUIDANCE") {
		t.Error("Expected header in context")
	}
	if !contains(ctx, "[FOCUS") {
		t.Error("Expected FOCUS label in context")
	}
	if !contains(ctx, "Work on feature 5") {
		t.Error("Expected focus content in context")
	}
	if !contains(ctx, "[CONSTRAINT") {
		t.Error("Expected CONSTRAINT label in context")
	}
	if !contains(ctx, "No external dependencies") {
		t.Error("Expected constraint content in context")
	}
	if !contains(ctx, "priority: 10") {
		t.Error("Expected priority in context for high-priority nudge")
	}
}

func TestBuildPromptContextWithAcknowledged(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	n, _ := store.Add(NudgeTypeFocus, "Active nudge", 0)
	store.Add(NudgeTypeConstraint, "Also active", 0)

	// Acknowledge one
	store.Acknowledge(n.ID)

	ctx := store.BuildPromptContext()

	// Should not contain acknowledged nudge
	if contains(ctx, "Active nudge") {
		t.Error("Should not contain acknowledged nudge")
	}
	// Should contain non-acknowledged nudge
	if !contains(ctx, "Also active") {
		t.Error("Should contain non-acknowledged nudge")
	}
}

func TestStoreSummary(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	// Empty summary
	summary := store.Summary()
	if !contains(summary, "No nudges defined") {
		t.Errorf("Expected 'No nudges defined' for empty store, got %q", summary)
	}

	// With nudges
	n, _ := store.Add(NudgeTypeFocus, "Focus test", 5)
	store.Add(NudgeTypeSkip, "Skip test", 0)
	store.Acknowledge(n.ID)

	summary = store.Summary()
	if !contains(summary, "2 total") {
		t.Error("Expected total count in summary")
	}
	if !contains(summary, "1 active") {
		t.Error("Expected active count in summary")
	}
	if !contains(summary, "1 acknowledged") {
		t.Error("Expected acknowledged count in summary")
	}
}

func TestParseNudgeType(t *testing.T) {
	tests := []struct {
		input   string
		want    NudgeType
		wantErr bool
	}{
		{"focus", NudgeTypeFocus, false},
		{"FOCUS", NudgeTypeFocus, false},
		{"Focus", NudgeTypeFocus, false},
		{"skip", NudgeTypeSkip, false},
		{"constraint", NudgeTypeConstraint, false},
		{"style", NudgeTypeStyle, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseNudgeType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNudgeType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseNudgeType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidNudgeTypes(t *testing.T) {
	types := ValidNudgeTypes()
	expected := []string{"focus", "skip", "constraint", "style"}

	if len(types) != len(expected) {
		t.Errorf("Expected %d types, got %d", len(expected), len(types))
	}

	for _, e := range expected {
		found := false
		for _, got := range types {
			if got == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected type %q not found", e)
		}
	}
}

func TestFormatAcknowledgment(t *testing.T) {
	// Empty list
	msg := FormatAcknowledgment(nil)
	if msg != "" {
		t.Errorf("Expected empty string for nil, got %q", msg)
	}

	msg = FormatAcknowledgment([]Nudge{})
	if msg != "" {
		t.Errorf("Expected empty string for empty slice, got %q", msg)
	}

	// With nudges
	nudges := []Nudge{
		{Type: NudgeTypeFocus, Content: "Focus on X"},
		{Type: NudgeTypeSkip, Content: "Skip Y"},
	}
	msg = FormatAcknowledgment(nudges)
	if !contains(msg, "Acknowledged nudges") {
		t.Error("Expected 'Acknowledged nudges' in message")
	}
	if !contains(msg, "[FOCUS]") {
		t.Error("Expected [FOCUS] in message")
	}
	if !contains(msg, "[SKIP]") {
		t.Error("Expected [SKIP] in message")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nudgePath := filepath.Join(tmpDir, "nudges.json")
	// Create empty file
	os.WriteFile(nudgePath, []byte{}, 0644)

	store := NewStore(nudgePath)
	if err := store.Load(); err != nil {
		t.Fatalf("Load() of empty file failed: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("Expected 0 nudges from empty file, got %d", store.Count())
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nudgePath := filepath.Join(tmpDir, "nudges.json")
	// Create invalid JSON
	os.WriteFile(nudgePath, []byte("not valid json{"), 0644)

	store := NewStore(nudgePath)
	if err := store.Load(); err == nil {
		t.Error("Expected error loading invalid JSON")
	}
}

func TestAcknowledgeNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nudge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "nudges.json"))
	store.Load()

	err = store.Acknowledge("nonexistent_id")
	if err == nil {
		t.Error("Expected error acknowledging non-existent nudge")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
