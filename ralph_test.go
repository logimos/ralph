package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectBuildSystem tests build system detection based on project files
func TestDetectBuildSystem(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tests := []struct {
		name           string
		files          []string
		expectedSystem string
	}{
		{
			name:           "Gradle with build.gradle",
			files:          []string{"build.gradle"},
			expectedSystem: "gradle",
		},
		{
			name:           "Gradle with build.gradle.kts",
			files:          []string{"build.gradle.kts"},
			expectedSystem: "gradle",
		},
		{
			name:           "Gradle with gradlew",
			files:          []string{"gradlew"},
			expectedSystem: "gradle",
		},
		{
			name:           "Maven with pom.xml",
			files:          []string{"pom.xml"},
			expectedSystem: "maven",
		},
		{
			name:           "Cargo with Cargo.toml",
			files:          []string{"Cargo.toml"},
			expectedSystem: "cargo",
		},
		{
			name:           "Go with go.mod",
			files:          []string{"go.mod"},
			expectedSystem: "go",
		},
		{
			name:           "Python with setup.py",
			files:          []string{"setup.py"},
			expectedSystem: "python",
		},
		{
			name:           "Python with pyproject.toml",
			files:          []string{"pyproject.toml"},
			expectedSystem: "python",
		},
		{
			name:           "Python with requirements.txt",
			files:          []string{"requirements.txt"},
			expectedSystem: "python",
		},
		{
			name:           "pnpm with pnpm-lock.yaml",
			files:          []string{"pnpm-lock.yaml"},
			expectedSystem: "pnpm",
		},
		{
			name:           "Yarn with yarn.lock",
			files:          []string{"yarn.lock"},
			expectedSystem: "yarn",
		},
		{
			name:           "npm with package.json",
			files:          []string{"package.json"},
			expectedSystem: "npm",
		},
		{
			name:           "No project files defaults to pnpm",
			files:          []string{},
			expectedSystem: "pnpm",
		},
		{
			name:           "Gradle takes precedence over npm",
			files:          []string{"build.gradle", "package.json"},
			expectedSystem: "gradle",
		},
		{
			name:           "Maven takes precedence over npm",
			files:          []string{"pom.xml", "package.json"},
			expectedSystem: "maven",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test
			tempDir, err := os.MkdirTemp("", "ralph-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create test files
			for _, file := range tt.files {
				filePath := filepath.Join(tempDir, file)
				if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
					t.Fatalf("Failed to create test file %s: %v", file, err)
				}
			}

			// Change to temp directory
			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}
			defer os.Chdir(originalDir)

			// Test detection
			result := detectBuildSystem()
			if result != tt.expectedSystem {
				t.Errorf("detectBuildSystem() = %q, want %q", result, tt.expectedSystem)
			}
		})
	}
}

// TestApplyBuildSystemConfig tests build system configuration application
func TestApplyBuildSystemConfig(t *testing.T) {
	tests := []struct {
		name             string
		config           Config
		expectedTypeCmd  string
		expectedTestCmd  string
	}{
		{
			name: "Explicit commands not overridden",
			config: Config{
				TypeCheckCmd: "custom typecheck",
				TestCmd:      "custom test",
				BuildSystem:  "go",
			},
			expectedTypeCmd: "custom typecheck",
			expectedTestCmd: "custom test",
		},
		{
			name: "pnpm preset applied",
			config: Config{
				BuildSystem: "pnpm",
			},
			expectedTypeCmd: "pnpm typecheck",
			expectedTestCmd: "pnpm test",
		},
		{
			name: "npm preset applied",
			config: Config{
				BuildSystem: "npm",
			},
			expectedTypeCmd: "npm run typecheck",
			expectedTestCmd: "npm test",
		},
		{
			name: "yarn preset applied",
			config: Config{
				BuildSystem: "yarn",
			},
			expectedTypeCmd: "yarn typecheck",
			expectedTestCmd: "yarn test",
		},
		{
			name: "gradle preset applied",
			config: Config{
				BuildSystem: "gradle",
			},
			expectedTypeCmd: "./gradlew check",
			expectedTestCmd: "./gradlew test",
		},
		{
			name: "maven preset applied",
			config: Config{
				BuildSystem: "maven",
			},
			expectedTypeCmd: "mvn compile",
			expectedTestCmd: "mvn test",
		},
		{
			name: "cargo preset applied",
			config: Config{
				BuildSystem: "cargo",
			},
			expectedTypeCmd: "cargo check",
			expectedTestCmd: "cargo test",
		},
		{
			name: "go preset applied",
			config: Config{
				BuildSystem: "go",
			},
			expectedTypeCmd: "go build ./...",
			expectedTestCmd: "go test ./...",
		},
		{
			name: "python preset applied",
			config: Config{
				BuildSystem: "python",
			},
			expectedTypeCmd: "mypy .",
			expectedTestCmd: "pytest",
		},
		{
			name: "Unknown build system falls back to pnpm",
			config: Config{
				BuildSystem: "unknown-system",
			},
			expectedTypeCmd: "pnpm typecheck",
			expectedTestCmd: "pnpm test",
		},
		{
			name: "Partial override - only TypeCheckCmd set",
			config: Config{
				TypeCheckCmd: "custom typecheck",
				BuildSystem:  "go",
			},
			expectedTypeCmd: "custom typecheck",
			expectedTestCmd: "go test ./...",
		},
		{
			name: "Partial override - only TestCmd set",
			config: Config{
				TestCmd:     "custom test",
				BuildSystem: "go",
			},
			expectedTypeCmd: "go build ./...",
			expectedTestCmd: "custom test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			applyBuildSystemConfig(&config)

			if config.TypeCheckCmd != tt.expectedTypeCmd {
				t.Errorf("TypeCheckCmd = %q, want %q", config.TypeCheckCmd, tt.expectedTypeCmd)
			}
			if config.TestCmd != tt.expectedTestCmd {
				t.Errorf("TestCmd = %q, want %q", config.TestCmd, tt.expectedTestCmd)
			}
		})
	}
}

// TestIsCursorAgent tests cursor agent detection
func TestIsCursorAgent(t *testing.T) {
	tests := []struct {
		name     string
		agentCmd string
		expected bool
	}{
		{
			name:     "cursor-agent is detected",
			agentCmd: "cursor-agent",
			expected: true,
		},
		{
			name:     "Cursor-Agent (uppercase) is detected",
			agentCmd: "Cursor-Agent",
			expected: true,
		},
		{
			name:     "cursor is detected",
			agentCmd: "cursor",
			expected: true,
		},
		{
			name:     "path with cursor-agent is detected",
			agentCmd: "/usr/local/bin/cursor-agent",
			expected: true,
		},
		{
			name:     "claude is not cursor agent",
			agentCmd: "claude",
			expected: false,
		},
		{
			name:     "claude-code is not cursor agent",
			agentCmd: "claude-code",
			expected: false,
		},
		{
			name:     "other-agent is not cursor agent",
			agentCmd: "other-agent",
			expected: false,
		},
		{
			name:     "empty string is not cursor agent",
			agentCmd: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCursorAgent(tt.agentCmd)
			if result != tt.expected {
				t.Errorf("isCursorAgent(%q) = %v, want %v", tt.agentCmd, result, tt.expected)
			}
		})
	}
}

// TestBuildPrompt tests prompt construction
func TestBuildPrompt(t *testing.T) {
	config := &Config{
		PlanFile:     "test-plan.json",
		ProgressFile: "test-progress.txt",
		TypeCheckCmd: "go build ./...",
		TestCmd:      "go test ./...",
	}

	prompt := buildPrompt(config)

	// Check that prompt contains expected elements
	if !strings.Contains(prompt, "test-plan.json") {
		t.Error("Prompt should contain plan file path")
	}
	if !strings.Contains(prompt, "test-progress.txt") {
		t.Error("Prompt should contain progress file path")
	}
	if !strings.Contains(prompt, "go build ./...") {
		t.Error("Prompt should contain typecheck command")
	}
	if !strings.Contains(prompt, "go test ./...") {
		t.Error("Prompt should contain test command")
	}
	if !strings.Contains(prompt, "highest-priority feature") {
		t.Error("Prompt should mention priority")
	}
	if !strings.Contains(prompt, completeSignal) {
		t.Error("Prompt should contain completion signal")
	}
	if !strings.Contains(prompt, "ONLY WORK ON A SINGLE FEATURE") {
		t.Error("Prompt should contain single feature instruction")
	}

	// Check @ references for files
	if !strings.Contains(prompt, "@") {
		t.Error("Prompt should use @ references for files")
	}
}

// TestBuildPromptAbsolutePaths tests that buildPrompt uses absolute paths
func TestBuildPromptAbsolutePaths(t *testing.T) {
	config := &Config{
		PlanFile:     "plan.json",
		ProgressFile: "progress.txt",
		TypeCheckCmd: "go build ./...",
		TestCmd:      "go test ./...",
	}

	prompt := buildPrompt(config)

	// The paths should be converted to absolute paths
	// Check that the prompt starts with @ and contains a path separator
	if !strings.Contains(prompt, "@/") && !strings.Contains(prompt, "@\\") {
		// On Windows, the path might use backslashes
		// If running on a system where we can resolve paths, check for absolute path
		cwd, _ := os.Getwd()
		if cwd != "" && !strings.Contains(prompt, cwd) {
			t.Error("Prompt should contain absolute paths")
		}
	}
}

// TestReadPlanFile tests reading and parsing plan files
func TestReadPlanFile(t *testing.T) {
	// Create a temporary plan file
	tempDir, err := os.MkdirTemp("", "ralph-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	planFile := filepath.Join(tempDir, "test-plan.json")

	testPlans := []Plan{
		{
			ID:             1,
			Category:       "feature",
			Description:    "Test feature 1",
			Steps:          []string{"Step 1", "Step 2"},
			ExpectedOutput: "Expected output 1",
			Tested:         true,
		},
		{
			ID:             2,
			Category:       "chore",
			Description:    "Test feature 2",
			Steps:          []string{"Step A", "Step B"},
			ExpectedOutput: "Expected output 2",
			Tested:         false,
		},
		{
			ID:             3,
			Category:       "infra",
			Description:    "Test feature 3",
			Steps:          []string{},
			ExpectedOutput: "",
			Tested:         false,
		},
	}

	// Write test plan file
	data, err := json.MarshalIndent(testPlans, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal test plans: %v", err)
	}
	if err := os.WriteFile(planFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test plan file: %v", err)
	}

	// Test reading the plan file
	plans, err := readPlanFile(planFile)
	if err != nil {
		t.Fatalf("readPlanFile() error = %v", err)
	}

	if len(plans) != 3 {
		t.Errorf("readPlanFile() returned %d plans, want 3", len(plans))
	}

	// Verify first plan
	if plans[0].ID != 1 {
		t.Errorf("First plan ID = %d, want 1", plans[0].ID)
	}
	if plans[0].Category != "feature" {
		t.Errorf("First plan Category = %q, want %q", plans[0].Category, "feature")
	}
	if !plans[0].Tested {
		t.Error("First plan Tested should be true")
	}
}

// TestReadPlanFileNotFound tests error handling for missing plan file
func TestReadPlanFileNotFound(t *testing.T) {
	_, err := readPlanFile("/nonexistent/path/plan.json")
	if err == nil {
		t.Error("readPlanFile() should return error for nonexistent file")
	}
}

// TestReadPlanFileInvalidJSON tests error handling for invalid JSON
func TestReadPlanFileInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempDir, err := os.MkdirTemp("", "ralph-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidFile := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	_, err = readPlanFile(invalidFile)
	if err == nil {
		t.Error("readPlanFile() should return error for invalid JSON")
	}
}

// TestFilterPlans tests filtering plans by tested status
func TestFilterPlans(t *testing.T) {
	plans := []Plan{
		{ID: 1, Tested: true},
		{ID: 2, Tested: false},
		{ID: 3, Tested: true},
		{ID: 4, Tested: false},
		{ID: 5, Tested: false},
	}

	// Test filtering for tested plans
	tested := filterPlans(plans, true)
	if len(tested) != 2 {
		t.Errorf("filterPlans(true) returned %d plans, want 2", len(tested))
	}
	for _, p := range tested {
		if !p.Tested {
			t.Errorf("filterPlans(true) returned untested plan ID %d", p.ID)
		}
	}

	// Test filtering for untested plans
	untested := filterPlans(plans, false)
	if len(untested) != 3 {
		t.Errorf("filterPlans(false) returned %d plans, want 3", len(untested))
	}
	for _, p := range untested {
		if p.Tested {
			t.Errorf("filterPlans(false) returned tested plan ID %d", p.ID)
		}
	}
}

// TestFilterPlansEmpty tests filtering empty plan list
func TestFilterPlansEmpty(t *testing.T) {
	var plans []Plan

	tested := filterPlans(plans, true)
	if len(tested) != 0 {
		t.Errorf("filterPlans(true) on empty list returned %d plans, want 0", len(tested))
	}

	untested := filterPlans(plans, false)
	if len(untested) != 0 {
		t.Errorf("filterPlans(false) on empty list returned %d plans, want 0", len(untested))
	}
}

// TestFilterPlansAllTested tests filtering when all plans are tested
func TestFilterPlansAllTested(t *testing.T) {
	plans := []Plan{
		{ID: 1, Tested: true},
		{ID: 2, Tested: true},
	}

	tested := filterPlans(plans, true)
	if len(tested) != 2 {
		t.Errorf("filterPlans(true) returned %d plans, want 2", len(tested))
	}

	untested := filterPlans(plans, false)
	if len(untested) != 0 {
		t.Errorf("filterPlans(false) returned %d plans, want 0", len(untested))
	}
}

// TestFilterPlansAllUntested tests filtering when all plans are untested
func TestFilterPlansAllUntested(t *testing.T) {
	plans := []Plan{
		{ID: 1, Tested: false},
		{ID: 2, Tested: false},
	}

	tested := filterPlans(plans, true)
	if len(tested) != 0 {
		t.Errorf("filterPlans(true) returned %d plans, want 0", len(tested))
	}

	untested := filterPlans(plans, false)
	if len(untested) != 2 {
		t.Errorf("filterPlans(false) returned %d plans, want 2", len(untested))
	}
}

// TestPlanStructJSON tests Plan struct JSON serialization
func TestPlanStructJSON(t *testing.T) {
	plan := Plan{
		ID:             1,
		Category:       "feature",
		Description:    "Test description",
		Steps:          []string{"Step 1", "Step 2"},
		ExpectedOutput: "Expected output",
		Tested:         true,
	}

	// Serialize to JSON
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("Failed to marshal Plan: %v", err)
	}

	// Deserialize back
	var decoded Plan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Plan: %v", err)
	}

	// Verify fields
	if decoded.ID != plan.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, plan.ID)
	}
	if decoded.Category != plan.Category {
		t.Errorf("Category mismatch: got %q, want %q", decoded.Category, plan.Category)
	}
	if decoded.Description != plan.Description {
		t.Errorf("Description mismatch: got %q, want %q", decoded.Description, plan.Description)
	}
	if len(decoded.Steps) != len(plan.Steps) {
		t.Errorf("Steps length mismatch: got %d, want %d", len(decoded.Steps), len(plan.Steps))
	}
	if decoded.ExpectedOutput != plan.ExpectedOutput {
		t.Errorf("ExpectedOutput mismatch: got %q, want %q", decoded.ExpectedOutput, plan.ExpectedOutput)
	}
	if decoded.Tested != plan.Tested {
		t.Errorf("Tested mismatch: got %v, want %v", decoded.Tested, plan.Tested)
	}
}

// TestBuildSystemPresetsExist verifies all documented build systems have presets
func TestBuildSystemPresetsExist(t *testing.T) {
	expectedSystems := []string{"pnpm", "npm", "yarn", "gradle", "maven", "cargo", "go", "python"}

	for _, system := range expectedSystems {
		preset, exists := BuildSystemPresets[system]
		if !exists {
			t.Errorf("BuildSystemPresets missing preset for %q", system)
			continue
		}
		if preset.TypeCheck == "" {
			t.Errorf("BuildSystemPresets[%q].TypeCheck is empty", system)
		}
		if preset.Test == "" {
			t.Errorf("BuildSystemPresets[%q].Test is empty", system)
		}
	}
}

// TestConfigStructDefaults tests Config struct initialization
func TestConfigStructDefaults(t *testing.T) {
	config := &Config{}

	// Verify zero values
	if config.Iterations != 0 {
		t.Errorf("Default Iterations = %d, want 0", config.Iterations)
	}
	if config.Verbose {
		t.Error("Default Verbose should be false")
	}
	if config.ShowVersion {
		t.Error("Default ShowVersion should be false")
	}
	if config.ListStatus {
		t.Error("Default ListStatus should be false")
	}
}

// TestAppendProgress tests the progress file append function
func TestAppendProgress(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "ralph-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	progressFile := filepath.Join(tempDir, "progress.txt")

	// Test appending to new file
	if err := appendProgress(progressFile, "First message"); err != nil {
		t.Fatalf("appendProgress() error = %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(progressFile)
	if err != nil {
		t.Fatalf("Failed to read progress file: %v", err)
	}
	if !strings.Contains(string(content), "First message") {
		t.Error("Progress file should contain first message")
	}

	// Test appending second message
	if err := appendProgress(progressFile, "Second message"); err != nil {
		t.Fatalf("appendProgress() second call error = %v", err)
	}

	content, err = os.ReadFile(progressFile)
	if err != nil {
		t.Fatalf("Failed to read progress file after second append: %v", err)
	}
	if !strings.Contains(string(content), "First message") {
		t.Error("Progress file should still contain first message")
	}
	if !strings.Contains(string(content), "Second message") {
		t.Error("Progress file should contain second message")
	}

	// Verify timestamp format
	if !strings.Contains(string(content), "[") || !strings.Contains(string(content), "]") {
		t.Error("Progress entries should have timestamps in brackets")
	}
}

// TestExtractAndWritePlan tests JSON extraction from agent output
func TestExtractAndWritePlan(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "ralph-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		input         string
		shouldSucceed bool
	}{
		{
			name:          "Valid JSON array",
			input:         `Here is the plan: [{"id": 1, "description": "Test"}]`,
			shouldSucceed: true,
		},
		{
			name:          "JSON in code block",
			input:         "```json\n[{\"id\": 1, \"description\": \"Test\"}]\n```",
			shouldSucceed: true,
		},
		{
			name:          "No JSON in output",
			input:         "This is just plain text without any JSON",
			shouldSucceed: false,
		},
		{
			name:          "Invalid JSON structure",
			input:         `[{"id": "not a number"}`,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, tt.name+".json")
			err := extractAndWritePlan(tt.input, testFile)

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("extractAndWritePlan() error = %v, want success", err)
				}
				// Verify file was created and is valid JSON
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					t.Error("extractAndWritePlan() should create output file")
				}
			} else {
				if err == nil {
					t.Error("extractAndWritePlan() should return error for invalid input")
				}
			}
		})
	}
}
