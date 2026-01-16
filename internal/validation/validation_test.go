package validation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseValidationType(t *testing.T) {
	tests := []struct {
		input    string
		expected ValidationType
		wantErr  bool
	}{
		{"http_get", ValidationTypeHTTPGet, false},
		{"get", ValidationTypeHTTPGet, false},
		{"http-get", ValidationTypeHTTPGet, false},
		{"HTTP_GET", ValidationTypeHTTPGet, false},
		{"http_post", ValidationTypeHTTPPost, false},
		{"post", ValidationTypeHTTPPost, false},
		{"cli_command", ValidationTypeCLI, false},
		{"cli", ValidationTypeCLI, false},
		{"command", ValidationTypeCLI, false},
		{"file_exists", ValidationTypeFileExists, false},
		{"file", ValidationTypeFileExists, false},
		{"exists", ValidationTypeFileExists, false},
		{"output_contains", ValidationTypeOutputContains, false},
		{"output", ValidationTypeOutputContains, false},
		{"contains", ValidationTypeOutputContains, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseValidationType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseValidationType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseValidationType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCreateValidator(t *testing.T) {
	tests := []struct {
		name    string
		def     ValidationDefinition
		wantErr bool
	}{
		{
			name: "http_get with URL",
			def: ValidationDefinition{
				Type: ValidationTypeHTTPGet,
				URL:  "http://example.com",
			},
			wantErr: false,
		},
		{
			name: "http_get without URL",
			def: ValidationDefinition{
				Type: ValidationTypeHTTPGet,
			},
			wantErr: true,
		},
		{
			name: "http_post with URL",
			def: ValidationDefinition{
				Type: ValidationTypeHTTPPost,
				URL:  "http://example.com",
			},
			wantErr: false,
		},
		{
			name: "cli_command with command",
			def: ValidationDefinition{
				Type:    ValidationTypeCLI,
				Command: "echo",
				Args:    []string{"hello"},
			},
			wantErr: false,
		},
		{
			name: "cli_command without command",
			def: ValidationDefinition{
				Type: ValidationTypeCLI,
			},
			wantErr: true,
		},
		{
			name: "file_exists with path",
			def: ValidationDefinition{
				Type: ValidationTypeFileExists,
				Path: "/tmp/test",
			},
			wantErr: false,
		},
		{
			name: "file_exists without path",
			def: ValidationDefinition{
				Type: ValidationTypeFileExists,
			},
			wantErr: true,
		},
		{
			name: "output_contains with pattern",
			def: ValidationDefinition{
				Type:    ValidationTypeOutputContains,
				Pattern: "hello",
				Input:   "hello world",
			},
			wantErr: false,
		},
		{
			name: "output_contains without pattern",
			def: ValidationDefinition{
				Type:  ValidationTypeOutputContains,
				Input: "hello world",
			},
			wantErr: true,
		},
		{
			name: "unknown type",
			def: ValidationDefinition{
				Type: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := CreateValidator(tt.def)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateValidator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && v == nil {
				t.Error("CreateValidator() returned nil validator without error")
			}
		})
	}
}

func TestEndpointValidator(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy"}`))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"something went wrong"}`))
		case "/post":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"created":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name        string
		def         ValidationDefinition
		wantSuccess bool
	}{
		{
			name: "successful GET",
			def: ValidationDefinition{
				Type:           ValidationTypeHTTPGet,
				URL:            server.URL + "/health",
				ExpectedStatus: 200,
			},
			wantSuccess: true,
		},
		{
			name: "GET with body pattern match",
			def: ValidationDefinition{
				Type:           ValidationTypeHTTPGet,
				URL:            server.URL + "/health",
				ExpectedStatus: 200,
				ExpectedBody:   `"status":\s*"healthy"`,
			},
			wantSuccess: true,
		},
		{
			name: "GET with body pattern mismatch",
			def: ValidationDefinition{
				Type:           ValidationTypeHTTPGet,
				URL:            server.URL + "/health",
				ExpectedStatus: 200,
				ExpectedBody:   `"status":\s*"unhealthy"`,
			},
			wantSuccess: false,
		},
		{
			name: "GET with wrong status",
			def: ValidationDefinition{
				Type:           ValidationTypeHTTPGet,
				URL:            server.URL + "/error",
				ExpectedStatus: 200,
			},
			wantSuccess: false,
		},
		{
			name: "successful POST",
			def: ValidationDefinition{
				Type:           ValidationTypeHTTPPost,
				URL:            server.URL + "/post",
				ExpectedStatus: 201,
			},
			wantSuccess: true,
		},
		{
			name: "404 endpoint",
			def: ValidationDefinition{
				Type:           ValidationTypeHTTPGet,
				URL:            server.URL + "/notfound",
				ExpectedStatus: 200,
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewEndpointValidator(tt.def)
			v.Config.MaxRetries = 0 // No retries for faster tests

			ctx := context.Background()
			result := v.Validate(ctx)

			if result.Success != tt.wantSuccess {
				t.Errorf("EndpointValidator.Validate() success = %v, want %v, message: %s",
					result.Success, tt.wantSuccess, result.Message)
			}
		})
	}
}

func TestCLIValidator(t *testing.T) {
	tests := []struct {
		name        string
		def         ValidationDefinition
		wantSuccess bool
	}{
		{
			name: "echo command",
			def: ValidationDefinition{
				Type:    ValidationTypeCLI,
				Command: "echo",
				Args:    []string{"hello"},
			},
			wantSuccess: true,
		},
		{
			name: "echo with output pattern",
			def: ValidationDefinition{
				Type:         ValidationTypeCLI,
				Command:      "echo",
				Args:         []string{"hello world"},
				ExpectedBody: "hello",
			},
			wantSuccess: true,
		},
		{
			name: "echo with non-matching pattern",
			def: ValidationDefinition{
				Type:         ValidationTypeCLI,
				Command:      "echo",
				Args:         []string{"hello"},
				ExpectedBody: "goodbye",
			},
			wantSuccess: false,
		},
		{
			name: "non-existent command",
			def: ValidationDefinition{
				Type:    ValidationTypeCLI,
				Command: "nonexistent_command_xyz",
			},
			wantSuccess: false,
		},
		{
			name: "true command (exit 0)",
			def: ValidationDefinition{
				Type:    ValidationTypeCLI,
				Command: "true",
			},
			wantSuccess: true,
		},
		{
			name: "false command (exit 1)",
			def: ValidationDefinition{
				Type:    ValidationTypeCLI,
				Command: "false",
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewCLIValidator(tt.def)
			v.Config.MaxRetries = 0 // No retries for faster tests

			ctx := context.Background()
			result := v.Validate(ctx)

			if result.Success != tt.wantSuccess {
				t.Errorf("CLIValidator.Validate() success = %v, want %v, message: %s, error: %s",
					result.Success, tt.wantSuccess, result.Message, result.Error)
			}
		})
	}
}

func TestFileExistsValidator(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		def         ValidationDefinition
		wantSuccess bool
	}{
		{
			name: "existing file",
			def: ValidationDefinition{
				Type: ValidationTypeFileExists,
				Path: testFile,
			},
			wantSuccess: true,
		},
		{
			name: "non-existing file",
			def: ValidationDefinition{
				Type: ValidationTypeFileExists,
				Path: filepath.Join(tmpDir, "nonexistent.txt"),
			},
			wantSuccess: false,
		},
		{
			name: "file with content pattern match",
			def: ValidationDefinition{
				Type:    ValidationTypeFileExists,
				Path:    testFile,
				Pattern: "hello",
			},
			wantSuccess: true,
		},
		{
			name: "file with content pattern mismatch",
			def: ValidationDefinition{
				Type:    ValidationTypeFileExists,
				Path:    testFile,
				Pattern: "goodbye",
			},
			wantSuccess: false,
		},
		{
			name: "file should not exist (but does)",
			def: ValidationDefinition{
				Type: ValidationTypeFileExists,
				Path: testFile,
				Options: map[string]interface{}{
					"should_exist": false,
				},
			},
			wantSuccess: false,
		},
		{
			name: "file should not exist (and doesn't)",
			def: ValidationDefinition{
				Type: ValidationTypeFileExists,
				Path: filepath.Join(tmpDir, "nonexistent.txt"),
				Options: map[string]interface{}{
					"should_exist": false,
				},
			},
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewFileExistsValidator(tt.def)

			ctx := context.Background()
			result := v.Validate(ctx)

			if result.Success != tt.wantSuccess {
				t.Errorf("FileExistsValidator.Validate() success = %v, want %v, message: %s",
					result.Success, tt.wantSuccess, result.Message)
			}
		})
	}
}

func TestOutputValidator(t *testing.T) {
	tests := []struct {
		name        string
		def         ValidationDefinition
		wantSuccess bool
	}{
		{
			name: "pattern matches",
			def: ValidationDefinition{
				Type:    ValidationTypeOutputContains,
				Input:   "hello world",
				Pattern: "hello",
			},
			wantSuccess: true,
		},
		{
			name: "pattern does not match",
			def: ValidationDefinition{
				Type:    ValidationTypeOutputContains,
				Input:   "hello world",
				Pattern: "goodbye",
			},
			wantSuccess: false,
		},
		{
			name: "regex pattern",
			def: ValidationDefinition{
				Type:    ValidationTypeOutputContains,
				Input:   "error: code 123",
				Pattern: `error:\s*code\s*\d+`,
			},
			wantSuccess: true,
		},
		{
			name: "inverse - pattern should not match (and doesn't)",
			def: ValidationDefinition{
				Type:    ValidationTypeOutputContains,
				Input:   "hello world",
				Pattern: "goodbye",
				Options: map[string]interface{}{
					"inverse": true,
				},
			},
			wantSuccess: true,
		},
		{
			name: "inverse - pattern should not match (but does)",
			def: ValidationDefinition{
				Type:    ValidationTypeOutputContains,
				Input:   "hello world",
				Pattern: "hello",
				Options: map[string]interface{}{
					"inverse": true,
				},
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewOutputValidator(tt.def)

			ctx := context.Background()
			result := v.Validate(ctx)

			if result.Success != tt.wantSuccess {
				t.Errorf("OutputValidator.Validate() success = %v, want %v, message: %s",
					result.Success, tt.wantSuccess, result.Message)
			}
		})
	}
}

func TestValidationRunner(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	runner := NewValidationRunner()

	// Add multiple validators
	defs := []ValidationDefinition{
		{
			Type:           ValidationTypeHTTPGet,
			URL:            server.URL,
			ExpectedStatus: 200,
			Description:    "Health check",
		},
		{
			Type:        ValidationTypeFileExists,
			Path:        testFile,
			Description: "Config file exists",
		},
		{
			Type:    ValidationTypeCLI,
			Command: "echo",
			Args:    []string{"test"},
		},
	}

	err := runner.AddFromDefinitions(defs)
	if err != nil {
		t.Fatalf("AddFromDefinitions() error = %v", err)
	}

	if len(runner.Validators) != 3 {
		t.Errorf("Expected 3 validators, got %d", len(runner.Validators))
	}

	ctx := context.Background()
	result := runner.Run(ctx)

	if !result.Success {
		t.Errorf("ValidationRunner.Run() expected success, got failure")
	}

	if result.TotalCount != 3 {
		t.Errorf("Expected TotalCount 3, got %d", result.TotalCount)
	}

	if result.PassedCount != 3 {
		t.Errorf("Expected PassedCount 3, got %d", result.PassedCount)
	}

	if result.FailedCount != 0 {
		t.Errorf("Expected FailedCount 0, got %d", result.FailedCount)
	}
}

func TestValidationRunnerWithFailures(t *testing.T) {
	runner := NewValidationRunner()

	defs := []ValidationDefinition{
		{
			Type:    ValidationTypeCLI,
			Command: "true", // Will succeed
		},
		{
			Type:    ValidationTypeCLI,
			Command: "false", // Will fail
		},
	}

	err := runner.AddFromDefinitions(defs)
	if err != nil {
		t.Fatalf("AddFromDefinitions() error = %v", err)
	}

	ctx := context.Background()
	result := runner.Run(ctx)

	if result.Success {
		t.Error("ValidationRunner.Run() expected failure, got success")
	}

	if result.PassedCount != 1 {
		t.Errorf("Expected PassedCount 1, got %d", result.PassedCount)
	}

	if result.FailedCount != 1 {
		t.Errorf("Expected FailedCount 1, got %d", result.FailedCount)
	}
}

func TestValidationResultSummary(t *testing.T) {
	result := ValidationRunResult{
		Success:     true,
		TotalCount:  3,
		PassedCount: 3,
		FailedCount: 0,
		Duration:    time.Second,
		FeatureID:   1,
		FeatureName: "Test feature",
		Results: []ValidationResult{
			{Success: true, Message: "Test 1 passed"},
			{Success: true, Message: "Test 2 passed"},
			{Success: true, Message: "Test 3 passed"},
		},
	}

	summary := result.Summary()

	if summary == "" {
		t.Error("Summary() returned empty string")
	}

	// Check that summary contains key information
	if !contains(summary, "PASSED") {
		t.Error("Summary() should contain 'PASSED'")
	}

	if !contains(summary, "3/3") {
		t.Error("Summary() should contain '3/3'")
	}
}

func TestValidationResultSummaryWithFailures(t *testing.T) {
	result := ValidationRunResult{
		Success:     false,
		TotalCount:  3,
		PassedCount: 2,
		FailedCount: 1,
		Duration:    time.Second,
		Results: []ValidationResult{
			{Success: true, Message: "Test 1 passed"},
			{Success: false, Message: "Test 2 failed", Error: "some error"},
			{Success: true, Message: "Test 3 passed"},
		},
	}

	summary := result.Summary()

	if !contains(summary, "FAILED") {
		t.Error("Summary() should contain 'FAILED'")
	}

	if !contains(summary, "2/3") {
		t.Error("Summary() should contain '2/3'")
	}
}

func TestEndpointValidatorDefaults(t *testing.T) {
	// Test that defaults are applied correctly
	defGet := ValidationDefinition{
		Type: ValidationTypeHTTPGet,
		URL:  "http://example.com",
	}
	vGet := NewEndpointValidator(defGet)

	if vGet.Method != "GET" {
		t.Errorf("Expected GET method for http_get, got %s", vGet.Method)
	}

	if vGet.ExpectedStatus != 200 {
		t.Errorf("Expected default status 200, got %d", vGet.ExpectedStatus)
	}

	defPost := ValidationDefinition{
		Type: ValidationTypeHTTPPost,
		URL:  "http://example.com",
	}
	vPost := NewEndpointValidator(defPost)

	if vPost.Method != "POST" {
		t.Errorf("Expected POST method for http_post, got %s", vPost.Method)
	}
}

func TestValidatorTimeout(t *testing.T) {
	def := ValidationDefinition{
		Type:    ValidationTypeHTTPGet,
		URL:     "http://example.com",
		Timeout: "5s",
	}
	v := NewEndpointValidator(def)

	if v.Config.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", v.Config.Timeout)
	}
}

func TestValidatorRetries(t *testing.T) {
	def := ValidationDefinition{
		Type:    ValidationTypeHTTPGet,
		URL:     "http://example.com",
		Retries: 5,
	}
	v := NewEndpointValidator(def)

	if v.Config.MaxRetries != 5 {
		t.Errorf("Expected 5 retries, got %d", v.Config.MaxRetries)
	}
}

func TestValidatorDescription(t *testing.T) {
	def := ValidationDefinition{
		Type:        ValidationTypeHTTPGet,
		URL:         "http://example.com/health",
		Description: "Health check endpoint",
	}
	v := NewEndpointValidator(def)

	if v.Description() != "Health check endpoint" {
		t.Errorf("Expected custom description, got %s", v.Description())
	}

	// Test default description
	defNoDesc := ValidationDefinition{
		Type: ValidationTypeHTTPGet,
		URL:  "http://example.com/health",
	}
	vNoDesc := NewEndpointValidator(defNoDesc)

	if vNoDesc.Description() == "" {
		t.Error("Expected default description, got empty string")
	}
}

func TestValidatorTypes(t *testing.T) {
	tests := []struct {
		def          ValidationDefinition
		expectedType ValidationType
	}{
		{
			def:          ValidationDefinition{Type: ValidationTypeHTTPGet, URL: "http://example.com"},
			expectedType: ValidationTypeHTTPGet,
		},
		{
			def:          ValidationDefinition{Type: ValidationTypeHTTPPost, URL: "http://example.com"},
			expectedType: ValidationTypeHTTPPost,
		},
		{
			def:          ValidationDefinition{Type: ValidationTypeCLI, Command: "echo"},
			expectedType: ValidationTypeCLI,
		},
		{
			def:          ValidationDefinition{Type: ValidationTypeFileExists, Path: "/tmp/test"},
			expectedType: ValidationTypeFileExists,
		},
		{
			def:          ValidationDefinition{Type: ValidationTypeOutputContains, Pattern: "test", Input: "test"},
			expectedType: ValidationTypeOutputContains,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.expectedType), func(t *testing.T) {
			v, err := CreateValidator(tt.def)
			if err != nil {
				t.Fatalf("CreateValidator() error = %v", err)
			}

			if v.Type() != tt.expectedType {
				t.Errorf("Type() = %v, want %v", v.Type(), tt.expectedType)
			}
		})
	}
}

func TestValidationRunResultFormatJSON(t *testing.T) {
	result := ValidationRunResult{
		Success:     true,
		TotalCount:  2,
		PassedCount: 2,
		FailedCount: 0,
		Duration:    time.Second,
		Results: []ValidationResult{
			{Success: true, Message: "Test passed"},
		},
	}

	json := result.FormatJSON()

	if json == "" {
		t.Error("FormatJSON() returned empty string")
	}

	if !contains(json, `"success": true`) {
		t.Error("FormatJSON() should contain success field")
	}

	if !contains(json, `"total_count": 2`) {
		t.Error("FormatJSON() should contain total_count field")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
