package baseline

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		path     string
		expected FileType
	}{
		// Source files
		{"main.go", FileTypeSource},
		{"src/app.py", FileTypeSource},
		{"lib/utils.js", FileTypeSource},
		{"components/Button.tsx", FileTypeSource},
		{"handlers/api.java", FileTypeSource},
		{"core/engine.rs", FileTypeSource},
		{"model.swift", FileTypeSource},
		{"style.css", FileTypeSource},
		{"template.html", FileTypeSource},

		// Test files
		{"main_test.go", FileTypeTest},
		{"app.test.js", FileTypeTest},
		{"component.spec.tsx", FileTypeTest},
		{"test/unit_test.py", FileTypeTest},
		{"tests/integration.go", FileTypeTest},
		{"__tests__/Button.test.js", FileTypeTest},

		// Config files
		{"package.json", FileTypeConfig},
		{"go.mod", FileTypeConfig},
		{"Cargo.toml", FileTypeConfig},
		{"requirements.txt", FileTypeConfig},
		{".gitignore", FileTypeConfig},
		{"docker-compose.yml", FileTypeConfig},
		{"tsconfig.json", FileTypeConfig},
		{"config.yaml", FileTypeConfig},
		{".eslintrc.json", FileTypeConfig},

		// Documentation
		{"README.md", FileTypeDocs},
		{"CHANGELOG.md", FileTypeDocs},
		{"CONTRIBUTING.md", FileTypeDocs},
		{"docs/guide.md", FileTypeDocs},
		{"documentation/api.rst", FileTypeDocs},
		{"notes.txt", FileTypeDocs},

		// Assets
		{"logo.png", FileTypeAsset},
		{"icon.svg", FileTypeAsset},
		{"font.woff2", FileTypeAsset},
		{"video.mp4", FileTypeAsset},
		{"archive.zip", FileTypeAsset},

		// Config - Makefile is a build config
		{"Makefile", FileTypeConfig},

		// Other
		{"unknown.xyz", FileTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectFileType(tt.path)
			if result != tt.expected {
				t.Errorf("detectFileType(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "Go"},
		{"app.py", "Python"},
		{"index.js", "JavaScript"},
		{"component.tsx", "TypeScript (React)"},
		{"App.java", "Java"},
		{"main.rs", "Rust"},
		{"script.sh", "Shell"},
		{"style.css", "CSS"},
		{"data.json", "JSON"},
		{"config.yaml", "YAML"},
		{"unknown.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectLanguage(tt.path)
			if result != tt.expected {
				t.Errorf("detectLanguage(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestNewScanner(t *testing.T) {
	scanner := NewScanner("/test/path")

	if scanner.rootPath != "/test/path" {
		t.Errorf("rootPath = %v, want /test/path", scanner.rootPath)
	}

	// Check default ignore dirs
	found := false
	for _, dir := range scanner.ignoreDirs {
		if dir == ".git" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected .git in default ignoreDirs")
	}
}

func TestScannerIgnoreDirs(t *testing.T) {
	scanner := NewScanner(".")

	// Test shouldIgnoreDir
	ignoreTests := []struct {
		name     string
		expected bool
	}{
		{".git", true},
		{"node_modules", true},
		{"vendor", true},
		{"src", false},
		{"lib", false},
	}

	for _, tt := range ignoreTests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.shouldIgnoreDir(tt.name)
			if result != tt.expected {
				t.Errorf("shouldIgnoreDir(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestScannerIgnoreFiles(t *testing.T) {
	scanner := NewScanner(".")

	// Test shouldIgnoreFile
	ignoreTests := []struct {
		name     string
		expected bool
	}{
		{"file.pyc", true},
		{"cache.swp", true},
		{".DS_Store", true},
		{"main.go", false},
		{"app.js", false},
	}

	for _, tt := range ignoreTests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.shouldIgnoreFile(tt.name)
			if result != tt.expected {
				t.Errorf("shouldIgnoreFile(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestAppendUnique(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected int
	}{
		{"add new item", []string{"a", "b"}, "c", 3},
		{"add duplicate", []string{"a", "b", "c"}, "b", 3},
		{"add to empty", []string{}, "a", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendUnique(tt.slice, tt.item)
			if len(result) != tt.expected {
				t.Errorf("len(appendUnique(...)) = %v, want %v", len(result), tt.expected)
			}
		})
	}
}

func TestBaselineSaveAndLoad(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "baseline-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a baseline
	baseline := &Baseline{
		Version:     "1.0",
		GeneratedAt: time.Now(),
		RootPath:    tmpDir,
		TechStack: TechStack{
			Languages:  []string{"Go", "JavaScript"},
			BuildTools: []string{"Go Modules", "npm"},
		},
		Structure: CodebaseStructure{
			RootPath:    tmpDir,
			Directories: []string{"cmd", "internal", "pkg"},
			EntryPoints: []string{"cmd/main.go"},
		},
		Files:      []FileInfo{{Path: "main.go", Type: FileTypeSource}},
		FileCounts: map[FileType]int{FileTypeSource: 1},
		TotalFiles: 1,
		TotalLines: 100,
	}

	// Save the baseline
	baselinePath := filepath.Join(tmpDir, "baseline.json")
	if err := baseline.Save(baselinePath); err != nil {
		t.Fatalf("Failed to save baseline: %v", err)
	}

	// Load the baseline
	loaded, err := Load(baselinePath)
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Verify
	if loaded.Version != baseline.Version {
		t.Errorf("Version = %v, want %v", loaded.Version, baseline.Version)
	}
	if loaded.TotalFiles != baseline.TotalFiles {
		t.Errorf("TotalFiles = %v, want %v", loaded.TotalFiles, baseline.TotalFiles)
	}
	if len(loaded.TechStack.Languages) != len(baseline.TechStack.Languages) {
		t.Errorf("Languages count = %v, want %v", len(loaded.TechStack.Languages), len(baseline.TechStack.Languages))
	}
}

func TestBaselineSummary(t *testing.T) {
	baseline := &Baseline{
		Version:     "1.0",
		GeneratedAt: time.Now(),
		RootPath:    "/test/project",
		TechStack: TechStack{
			Languages:     []string{"Go"},
			BuildTools:    []string{"Go Modules"},
			TestFramework: "Go testing",
		},
		Structure: CodebaseStructure{
			Directories: []string{"cmd", "internal"},
			EntryPoints: []string{"cmd/main.go"},
		},
		FileCounts: map[FileType]int{FileTypeSource: 10, FileTypeTest: 5},
		TotalFiles: 15,
		TotalLines: 1000,
		Conventions: []string{"Standard Go project layout"},
		Patterns:    []string{"Repository pattern"},
	}

	summary := baseline.Summary()

	// Check that summary contains expected information
	checks := []string{
		"Codebase Baseline",
		"/test/project",
		"Total files: 15",
		"Total lines: 1000",
		"Go",
		"Go Modules",
		"Go testing",
		"Standard Go project layout",
		"Repository pattern",
	}

	for _, check := range checks {
		if !contains(summary, check) {
			t.Errorf("Summary missing expected content: %q", check)
		}
	}
}

func TestBaselineBuildPromptContext(t *testing.T) {
	baseline := &Baseline{
		TechStack: TechStack{
			Languages:     []string{"Go", "JavaScript"},
			BuildTools:    []string{"Go Modules"},
			TestFramework: "Go testing",
		},
		Structure: CodebaseStructure{
			EntryPoints: []string{"cmd/main.go"},
			TestDirs:    []string{"internal/pkg_test"},
		},
		FileCounts:  map[FileType]int{FileTypeSource: 50, FileTypeTest: 20, FileTypeConfig: 5},
		TotalFiles:  75,
		Conventions: []string{"Standard Go project layout"},
		Patterns:    []string{"Repository pattern"},
	}

	context := baseline.BuildPromptContext()

	// Check that context contains expected information
	checks := []string{
		"[CODEBASE CONTEXT",
		"Technology Stack",
		"Primary languages: Go, JavaScript",
		"Build tools: Go Modules",
		"Test framework: Go testing",
		"Project Structure",
		"Total files: 75",
		"Entry points: cmd/main.go",
		"Conventions to follow",
		"Standard Go project layout",
		"Patterns in use",
		"Repository pattern",
		"[END CODEBASE CONTEXT]",
	}

	for _, check := range checks {
		if !contains(context, check) {
			t.Errorf("BuildPromptContext missing expected content: %q", check)
		}
	}
}

func TestScannerSetIgnoreDirs(t *testing.T) {
	scanner := NewScanner(".")

	// Set custom ignore dirs
	scanner.SetIgnoreDirs([]string{"custom1", "custom2"})

	if len(scanner.ignoreDirs) != 2 {
		t.Errorf("ignoreDirs length = %v, want 2", len(scanner.ignoreDirs))
	}

	if !scanner.shouldIgnoreDir("custom1") {
		t.Error("Expected custom1 to be ignored")
	}

	// Old defaults should not be ignored anymore
	if scanner.shouldIgnoreDir(".git") {
		t.Error("Expected .git to NOT be ignored after SetIgnoreDirs")
	}
}

func TestScannerAddIgnoreDirs(t *testing.T) {
	scanner := NewScanner(".")
	originalCount := len(scanner.ignoreDirs)

	// Add more ignore dirs
	scanner.AddIgnoreDirs([]string{"custom1", "custom2"})

	if len(scanner.ignoreDirs) != originalCount+2 {
		t.Errorf("ignoreDirs length = %v, want %v", len(scanner.ignoreDirs), originalCount+2)
	}

	if !scanner.shouldIgnoreDir("custom1") {
		t.Error("Expected custom1 to be ignored")
	}

	// Original defaults should still be ignored
	if !scanner.shouldIgnoreDir(".git") {
		t.Error("Expected .git to still be ignored after AddIgnoreDirs")
	}
}

func TestScanRealDirectory(t *testing.T) {
	// Create a temporary directory with some test files
	tmpDir, err := os.MkdirTemp("", "scan-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := map[string]string{
		"main.go":              "package main\n\nfunc main() {}\n",
		"main_test.go":         "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}\n",
		"go.mod":               "module test\n\ngo 1.21\n",
		"README.md":            "# Test Project\n",
		"internal/pkg/pkg.go":  "package pkg\n",
		".git/config":          "[core]\n", // Should be ignored
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// Scan the directory
	scanner := NewScanner(tmpDir)
	baseline, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Verify results
	if baseline.TotalFiles != 5 { // Excluding .git/config
		t.Errorf("TotalFiles = %v, want 5", baseline.TotalFiles)
	}

	// Check file type counts
	if baseline.FileCounts[FileTypeSource] != 2 { // main.go and pkg.go
		t.Errorf("Source files = %v, want 2", baseline.FileCounts[FileTypeSource])
	}
	if baseline.FileCounts[FileTypeTest] != 1 { // main_test.go
		t.Errorf("Test files = %v, want 1", baseline.FileCounts[FileTypeTest])
	}
	if baseline.FileCounts[FileTypeConfig] != 1 { // go.mod
		t.Errorf("Config files = %v, want 1", baseline.FileCounts[FileTypeConfig])
	}
	if baseline.FileCounts[FileTypeDocs] != 1 { // README.md
		t.Errorf("Doc files = %v, want 1", baseline.FileCounts[FileTypeDocs])
	}

	// Check tech stack detection
	if len(baseline.TechStack.Languages) == 0 || baseline.TechStack.Languages[0] != "Go" {
		t.Errorf("Languages = %v, want [Go, ...]", baseline.TechStack.Languages)
	}
	if len(baseline.TechStack.BuildTools) == 0 {
		t.Error("Expected build tools to be detected")
	}

	// Check structure
	if len(baseline.Structure.EntryPoints) == 0 {
		t.Error("Expected entry points to be detected")
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/baseline.json")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestDetectConventions(t *testing.T) {
	scanner := NewScanner(".")

	// Test with Go standard layout files
	files := []FileInfo{
		{Path: "cmd/app/main.go", Type: FileTypeSource},
		{Path: "internal/pkg/service.go", Type: FileTypeSource},
		{Path: "pkg/utils/helper.go", Type: FileTypeSource},
		{Path: "internal/pkg/service_test.go", Type: FileTypeTest},
	}

	conventions := scanner.detectConventions(files)

	found := false
	for _, c := range conventions {
		if contains(c, "Standard Go project layout") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to detect Standard Go project layout convention")
	}
}

func TestDetectPatterns(t *testing.T) {
	scanner := NewScanner(".")

	// Test with MVC-like structure
	files := []FileInfo{
		{Path: "models/user.go", Type: FileTypeSource},
		{Path: "views/home.html", Type: FileTypeSource},
		{Path: "controllers/user.go", Type: FileTypeSource},
	}

	patterns := scanner.detectPatterns(files)

	found := false
	for _, p := range patterns {
		if contains(p, "MVC") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to detect MVC pattern")
	}
}

func TestCountLines(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "lines-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write 5 lines
	content := "line1\nline2\nline3\nline4\nline5"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	lines, err := countLines(tmpFile.Name())
	if err != nil {
		t.Fatalf("countLines failed: %v", err)
	}

	if lines != 5 {
		t.Errorf("countLines = %v, want 5", lines)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s, substr))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
