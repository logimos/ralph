package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDiscoverConfigFileCurrentDir tests config file discovery in current directory
func TestDiscoverConfigFileCurrentDir(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	for _, fileName := range ConfigFileNames {
		t.Run(fileName, func(t *testing.T) {
			// Create temp directory
			tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create config file
			configPath := filepath.Join(tempDir, fileName)
			if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			// Change to temp directory
			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}
			defer os.Chdir(originalDir)

			// Test discovery
			found := DiscoverConfigFile()
			if found != configPath {
				t.Errorf("DiscoverConfigFile() = %q, want %q", found, configPath)
			}
		})
	}
}

// TestDiscoverConfigFilePrecedence tests that earlier files in the list take precedence
func TestDiscoverConfigFilePrecedence(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple config files
	yamlPath := filepath.Join(tempDir, ".ralph.yaml")
	jsonPath := filepath.Join(tempDir, ".ralph.json")
	if err := os.WriteFile(yamlPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create yaml config: %v", err)
	}
	if err := os.WriteFile(jsonPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create json config: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// .ralph.yaml should take precedence over .ralph.json
	found := DiscoverConfigFile()
	if found != yamlPath {
		t.Errorf("DiscoverConfigFile() = %q, want %q (yaml should take precedence)", found, yamlPath)
	}
}

// TestDiscoverConfigFileNotFound tests that empty string is returned when no config exists
func TestDiscoverConfigFileNotFound(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create empty temp directory
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalDir)

	found := DiscoverConfigFile()
	if found != "" {
		t.Errorf("DiscoverConfigFile() = %q, want empty string", found)
	}
}

// TestLoadConfigFileYAML tests loading a YAML config file
func TestLoadConfigFileYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	yamlContent := `
agent: custom-agent
build_system: go
typecheck: go vet ./...
test: go test -v ./...
plan: custom-plan.json
progress: custom-progress.txt
iterations: 10
verbose: true
`
	configPath := filepath.Join(tempDir, ".ralph.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if cfg.Agent != "custom-agent" {
		t.Errorf("Agent = %q, want %q", cfg.Agent, "custom-agent")
	}
	if cfg.BuildSystem != "go" {
		t.Errorf("BuildSystem = %q, want %q", cfg.BuildSystem, "go")
	}
	if cfg.TypeCheck != "go vet ./..." {
		t.Errorf("TypeCheck = %q, want %q", cfg.TypeCheck, "go vet ./...")
	}
	if cfg.Test != "go test -v ./..." {
		t.Errorf("Test = %q, want %q", cfg.Test, "go test -v ./...")
	}
	if cfg.Plan != "custom-plan.json" {
		t.Errorf("Plan = %q, want %q", cfg.Plan, "custom-plan.json")
	}
	if cfg.Progress != "custom-progress.txt" {
		t.Errorf("Progress = %q, want %q", cfg.Progress, "custom-progress.txt")
	}
	if cfg.Iterations != 10 {
		t.Errorf("Iterations = %d, want %d", cfg.Iterations, 10)
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
}

// TestLoadConfigFileJSON tests loading a JSON config file
func TestLoadConfigFileJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	jsonContent := `{
	"agent": "json-agent",
	"build_system": "npm",
	"typecheck": "npm run check",
	"test": "npm test",
	"plan": "my-plan.json",
	"progress": "my-progress.txt",
	"iterations": 5,
	"verbose": false
}`
	configPath := filepath.Join(tempDir, ".ralph.json")
	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if cfg.Agent != "json-agent" {
		t.Errorf("Agent = %q, want %q", cfg.Agent, "json-agent")
	}
	if cfg.BuildSystem != "npm" {
		t.Errorf("BuildSystem = %q, want %q", cfg.BuildSystem, "npm")
	}
	if cfg.Iterations != 5 {
		t.Errorf("Iterations = %d, want %d", cfg.Iterations, 5)
	}
}

// TestLoadConfigFileEmpty tests loading an empty config file
func TestLoadConfigFileEmpty(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, ".ralph.yaml")
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	// Should return empty config
	if cfg.Agent != "" {
		t.Errorf("Agent = %q, want empty", cfg.Agent)
	}
}

// TestLoadConfigFileNotFound tests error when file doesn't exist
func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfigFile("/nonexistent/path/.ralph.yaml")
	if err == nil {
		t.Error("LoadConfigFile() should return error for nonexistent file")
	}
}

// TestLoadConfigFileInvalidYAML tests error handling for invalid YAML
func TestLoadConfigFileInvalidYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, ".ralph.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err = LoadConfigFile(configPath)
	if err == nil {
		t.Error("LoadConfigFile() should return error for invalid YAML")
	}
}

// TestLoadConfigFileInvalidJSON tests error handling for invalid JSON
func TestLoadConfigFileInvalidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, ".ralph.json")
	if err := os.WriteFile(configPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err = LoadConfigFile(configPath)
	if err == nil {
		t.Error("LoadConfigFile() should return error for invalid JSON")
	}
}

// TestValidateFileConfigValid tests validation of valid config
func TestValidateFileConfigValid(t *testing.T) {
	validConfigs := []FileConfig{
		{},
		{BuildSystem: "go"},
		{BuildSystem: "auto"},
		{BuildSystem: "pnpm", Iterations: 5},
		{Agent: "custom", Verbose: true},
	}

	for i, cfg := range validConfigs {
		if err := ValidateFileConfig(&cfg); err != nil {
			t.Errorf("ValidateFileConfig() for config %d returned error: %v", i, err)
		}
	}
}

// TestValidateFileConfigInvalid tests validation catches invalid config
func TestValidateFileConfigInvalid(t *testing.T) {
	tests := []struct {
		name string
		cfg  FileConfig
	}{
		{
			name: "Invalid build system",
			cfg:  FileConfig{BuildSystem: "invalid-system"},
		},
		{
			name: "Negative iterations",
			cfg:  FileConfig{Iterations: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateFileConfig(&tt.cfg); err == nil {
				t.Error("ValidateFileConfig() should return error for invalid config")
			}
		})
	}
}

// TestApplyFileConfig tests applying file config to Config struct
func TestApplyFileConfig(t *testing.T) {
	cfg := New()
	fileCfg := &FileConfig{
		Agent:       "file-agent",
		BuildSystem: "cargo",
		TypeCheck:   "cargo check",
		Test:        "cargo test",
		Plan:        "file-plan.json",
		Progress:    "file-progress.txt",
		Iterations:  3,
		Verbose:     true,
	}

	ApplyFileConfig(cfg, fileCfg)

	if cfg.AgentCmd != "file-agent" {
		t.Errorf("AgentCmd = %q, want %q", cfg.AgentCmd, "file-agent")
	}
	if cfg.BuildSystem != "cargo" {
		t.Errorf("BuildSystem = %q, want %q", cfg.BuildSystem, "cargo")
	}
	if cfg.TypeCheckCmd != "cargo check" {
		t.Errorf("TypeCheckCmd = %q, want %q", cfg.TypeCheckCmd, "cargo check")
	}
	if cfg.TestCmd != "cargo test" {
		t.Errorf("TestCmd = %q, want %q", cfg.TestCmd, "cargo test")
	}
	if cfg.PlanFile != "file-plan.json" {
		t.Errorf("PlanFile = %q, want %q", cfg.PlanFile, "file-plan.json")
	}
	if cfg.ProgressFile != "file-progress.txt" {
		t.Errorf("ProgressFile = %q, want %q", cfg.ProgressFile, "file-progress.txt")
	}
	if cfg.Iterations != 3 {
		t.Errorf("Iterations = %d, want %d", cfg.Iterations, 3)
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
}

// TestApplyFileConfigDoesNotOverrideExisting tests that existing values are not overridden
func TestApplyFileConfigDoesNotOverrideExisting(t *testing.T) {
	cfg := New()
	cfg.AgentCmd = "cli-agent"
	cfg.TypeCheckCmd = "cli-typecheck"
	cfg.PlanFile = "cli-plan.json"
	cfg.Iterations = 10

	fileCfg := &FileConfig{
		Agent:      "file-agent",
		TypeCheck:  "file-typecheck",
		Plan:       "file-plan.json",
		Iterations: 3,
	}

	ApplyFileConfig(cfg, fileCfg)

	// Values already set should not be overridden
	if cfg.AgentCmd != "cli-agent" {
		t.Errorf("AgentCmd = %q, should remain %q", cfg.AgentCmd, "cli-agent")
	}
	if cfg.TypeCheckCmd != "cli-typecheck" {
		t.Errorf("TypeCheckCmd = %q, should remain %q", cfg.TypeCheckCmd, "cli-typecheck")
	}
	if cfg.PlanFile != "cli-plan.json" {
		t.Errorf("PlanFile = %q, should remain %q", cfg.PlanFile, "cli-plan.json")
	}
	if cfg.Iterations != 10 {
		t.Errorf("Iterations = %d, should remain %d", cfg.Iterations, 10)
	}
}

// TestConfigFileNamesOrder tests that ConfigFileNames are in the expected order
func TestConfigFileNamesOrder(t *testing.T) {
	expected := []string{
		".ralph.yaml",
		".ralph.yml",
		".ralph.json",
		"ralph.config.yaml",
		"ralph.config.yml",
		"ralph.config.json",
	}

	if len(ConfigFileNames) != len(expected) {
		t.Fatalf("ConfigFileNames has %d entries, want %d", len(ConfigFileNames), len(expected))
	}

	for i, name := range expected {
		if ConfigFileNames[i] != name {
			t.Errorf("ConfigFileNames[%d] = %q, want %q", i, ConfigFileNames[i], name)
		}
	}
}

// TestFindConfigInDir tests the findConfigInDir helper function
func TestFindConfigInDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with no config file
	result := findConfigInDir(tempDir)
	if result != "" {
		t.Errorf("findConfigInDir() = %q, want empty string", result)
	}

	// Create a config file
	configPath := filepath.Join(tempDir, ".ralph.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test with config file
	result = findConfigInDir(tempDir)
	if result != configPath {
		t.Errorf("findConfigInDir() = %q, want %q", result, configPath)
	}
}

// TestLoadConfigFilePartialYAML tests loading config with only some fields set
func TestLoadConfigFilePartialYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ralph-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	yamlContent := `
build_system: python
iterations: 7
`
	configPath := filepath.Join(tempDir, ".ralph.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	// Set fields should have values
	if cfg.BuildSystem != "python" {
		t.Errorf("BuildSystem = %q, want %q", cfg.BuildSystem, "python")
	}
	if cfg.Iterations != 7 {
		t.Errorf("Iterations = %d, want %d", cfg.Iterations, 7)
	}

	// Unset fields should be empty/zero
	if cfg.Agent != "" {
		t.Errorf("Agent = %q, want empty", cfg.Agent)
	}
	if cfg.TypeCheck != "" {
		t.Errorf("TypeCheck = %q, want empty", cfg.TypeCheck)
	}
}
