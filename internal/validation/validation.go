// Package validation provides outcome-focused validation beyond tests and type checks.
// It supports validating API endpoints, CLI commands, file existence, and output patterns.
package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ValidationType represents the type of validation to perform
type ValidationType string

const (
	// ValidationTypeHTTPGet validates an HTTP GET request
	ValidationTypeHTTPGet ValidationType = "http_get"
	// ValidationTypeHTTPPost validates an HTTP POST request
	ValidationTypeHTTPPost ValidationType = "http_post"
	// ValidationTypeCLI validates a CLI command execution
	ValidationTypeCLI ValidationType = "cli_command"
	// ValidationTypeFileExists validates that a file exists
	ValidationTypeFileExists ValidationType = "file_exists"
	// ValidationTypeOutputContains validates that output contains a pattern
	ValidationTypeOutputContains ValidationType = "output_contains"
)

// DefaultTimeout is the default timeout for validation operations
const DefaultTimeout = 30 * time.Second

// DefaultMaxRetries is the default number of retries for validation
const DefaultMaxRetries = 3

// ValidationDefinition represents a validation rule defined in plan.json
type ValidationDefinition struct {
	Type           ValidationType         `json:"type"`
	URL            string                 `json:"url,omitempty"`             // For HTTP validations
	Method         string                 `json:"method,omitempty"`          // HTTP method (defaults based on type)
	Body           string                 `json:"body,omitempty"`            // Request body for POST
	Headers        map[string]string      `json:"headers,omitempty"`         // HTTP headers
	ExpectedStatus int                    `json:"expected_status,omitempty"` // Expected HTTP status code
	ExpectedBody   string                 `json:"expected_body,omitempty"`   // Expected response body pattern (regex)
	Command        string                 `json:"command,omitempty"`         // For CLI validations
	Args           []string               `json:"args,omitempty"`            // Command arguments
	Path           string                 `json:"path,omitempty"`            // For file_exists validation
	Pattern        string                 `json:"pattern,omitempty"`         // For output_contains validation
	Input          string                 `json:"input,omitempty"`           // Input to check for pattern
	Timeout        string                 `json:"timeout,omitempty"`         // Timeout duration (e.g., "30s")
	Retries        int                    `json:"retries,omitempty"`         // Number of retries
	Description    string                 `json:"description,omitempty"`     // Human-readable description
	Options        map[string]interface{} `json:"options,omitempty"`         // Additional options
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Success     bool          `json:"success"`
	Message     string        `json:"message"`
	Duration    time.Duration `json:"duration"`
	Retries     int           `json:"retries"`
	Output      string        `json:"output,omitempty"`      // Captured output for debugging
	StatusCode  int           `json:"status_code,omitempty"` // For HTTP validations
	Error       string        `json:"error,omitempty"`       // Error message if failed
	ValidatorID string        `json:"validator_id,omitempty"`
}

// Validator is the interface for all validation types
type Validator interface {
	// Validate performs the validation and returns the result
	Validate(ctx context.Context) ValidationResult
	// Type returns the validation type
	Type() ValidationType
	// Description returns a human-readable description
	Description() string
}

// ValidatorConfig holds common configuration for validators
type ValidatorConfig struct {
	Timeout    time.Duration
	MaxRetries int
}

// DefaultValidatorConfig returns the default validator configuration
func DefaultValidatorConfig() ValidatorConfig {
	return ValidatorConfig{
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
	}
}

// EndpointValidator validates HTTP endpoints
type EndpointValidator struct {
	URL            string
	Method         string
	Body           string
	Headers        map[string]string
	ExpectedStatus int
	ExpectedBody   string // Regex pattern
	Config         ValidatorConfig
	Desc           string
}

// NewEndpointValidator creates a new endpoint validator from a definition
func NewEndpointValidator(def ValidationDefinition) *EndpointValidator {
	method := def.Method
	if method == "" {
		if def.Type == ValidationTypeHTTPPost {
			method = "POST"
		} else {
			method = "GET"
		}
	}

	expectedStatus := def.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	timeout := DefaultTimeout
	if def.Timeout != "" {
		if d, err := time.ParseDuration(def.Timeout); err == nil {
			timeout = d
		}
	}

	retries := def.Retries
	if retries <= 0 {
		retries = DefaultMaxRetries
	}

	return &EndpointValidator{
		URL:            def.URL,
		Method:         strings.ToUpper(method),
		Body:           def.Body,
		Headers:        def.Headers,
		ExpectedStatus: expectedStatus,
		ExpectedBody:   def.ExpectedBody,
		Config: ValidatorConfig{
			Timeout:    timeout,
			MaxRetries: retries,
		},
		Desc: def.Description,
	}
}

// Validate performs the HTTP endpoint validation
func (v *EndpointValidator) Validate(ctx context.Context) ValidationResult {
	start := time.Now()
	result := ValidationResult{
		ValidatorID: fmt.Sprintf("http_%s_%s", strings.ToLower(v.Method), sanitizeURL(v.URL)),
	}

	var lastErr error
	for attempt := 0; attempt <= v.Config.MaxRetries; attempt++ {
		result.Retries = attempt

		// Create request
		var body io.Reader
		if v.Body != "" {
			body = strings.NewReader(v.Body)
		}

		req, err := http.NewRequestWithContext(ctx, v.Method, v.URL, body)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Set headers
		for k, val := range v.Headers {
			req.Header.Set(k, val)
		}
		if v.Body != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// Create client with timeout
		client := &http.Client{
			Timeout: v.Config.Timeout,
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
			continue
		}

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		result.StatusCode = resp.StatusCode
		result.Output = string(respBody)

		// Check status code
		if resp.StatusCode != v.ExpectedStatus {
			lastErr = fmt.Errorf("expected status %d, got %d", v.ExpectedStatus, resp.StatusCode)
			continue
		}

		// Check body pattern if specified
		if v.ExpectedBody != "" {
			matched, err := regexp.MatchString(v.ExpectedBody, string(respBody))
			if err != nil {
				lastErr = fmt.Errorf("invalid body pattern: %w", err)
				continue
			}
			if !matched {
				lastErr = fmt.Errorf("response body does not match pattern %q", v.ExpectedBody)
				continue
			}
		}

		// Success
		result.Success = true
		result.Message = fmt.Sprintf("%s %s returned %d", v.Method, v.URL, resp.StatusCode)
		result.Duration = time.Since(start)
		return result
	}

	// All retries failed
	result.Success = false
	result.Duration = time.Since(start)
	if lastErr != nil {
		result.Error = lastErr.Error()
		result.Message = fmt.Sprintf("validation failed after %d retries: %s", result.Retries+1, lastErr)
	}
	return result
}

// Type returns the validation type
func (v *EndpointValidator) Type() ValidationType {
	if v.Method == "POST" {
		return ValidationTypeHTTPPost
	}
	return ValidationTypeHTTPGet
}

// Description returns a human-readable description
func (v *EndpointValidator) Description() string {
	if v.Desc != "" {
		return v.Desc
	}
	return fmt.Sprintf("%s %s", v.Method, v.URL)
}

// CLIValidator validates CLI command execution
type CLIValidator struct {
	Command        string
	Args           []string
	ExpectedOutput string // Regex pattern for stdout
	ExpectedExitCode int
	Config         ValidatorConfig
	Desc           string
}

// NewCLIValidator creates a new CLI validator from a definition
func NewCLIValidator(def ValidationDefinition) *CLIValidator {
	timeout := DefaultTimeout
	if def.Timeout != "" {
		if d, err := time.ParseDuration(def.Timeout); err == nil {
			timeout = d
		}
	}

	retries := def.Retries
	if retries <= 0 {
		retries = DefaultMaxRetries
	}

	expectedExitCode := 0
	if exitCode, ok := def.Options["expected_exit_code"].(float64); ok {
		expectedExitCode = int(exitCode)
	}

	return &CLIValidator{
		Command:        def.Command,
		Args:           def.Args,
		ExpectedOutput: def.ExpectedBody, // Reuse expected_body for output pattern
		ExpectedExitCode: expectedExitCode,
		Config: ValidatorConfig{
			Timeout:    timeout,
			MaxRetries: retries,
		},
		Desc: def.Description,
	}
}

// Validate performs the CLI command validation
func (v *CLIValidator) Validate(ctx context.Context) ValidationResult {
	start := time.Now()
	result := ValidationResult{
		ValidatorID: fmt.Sprintf("cli_%s", sanitizeCommand(v.Command)),
	}

	var lastErr error
	for attempt := 0; attempt <= v.Config.MaxRetries; attempt++ {
		result.Retries = attempt

		// Create command with context for timeout
		cmdCtx, cancel := context.WithTimeout(ctx, v.Config.Timeout)
		cmd := exec.CommandContext(cmdCtx, v.Command, v.Args...)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		cancel()

		result.Output = stdout.String()
		if stderr.Len() > 0 {
			result.Output += "\n[stderr]: " + stderr.String()
		}

		// Get exit code
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				lastErr = fmt.Errorf("command failed: %w", err)
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
		}

		result.StatusCode = exitCode

		// Check exit code
		if exitCode != v.ExpectedExitCode {
			lastErr = fmt.Errorf("expected exit code %d, got %d", v.ExpectedExitCode, exitCode)
			continue
		}

		// Check output pattern if specified
		if v.ExpectedOutput != "" {
			matched, err := regexp.MatchString(v.ExpectedOutput, stdout.String())
			if err != nil {
				lastErr = fmt.Errorf("invalid output pattern: %w", err)
				continue
			}
			if !matched {
				lastErr = fmt.Errorf("output does not match pattern %q", v.ExpectedOutput)
				continue
			}
		}

		// Success
		result.Success = true
		result.Message = fmt.Sprintf("command %q completed with exit code %d", v.Command, exitCode)
		result.Duration = time.Since(start)
		return result
	}

	// All retries failed
	result.Success = false
	result.Duration = time.Since(start)
	if lastErr != nil {
		result.Error = lastErr.Error()
		result.Message = fmt.Sprintf("validation failed after %d retries: %s", result.Retries+1, lastErr)
	}
	return result
}

// Type returns the validation type
func (v *CLIValidator) Type() ValidationType {
	return ValidationTypeCLI
}

// Description returns a human-readable description
func (v *CLIValidator) Description() string {
	if v.Desc != "" {
		return v.Desc
	}
	return fmt.Sprintf("cli: %s %s", v.Command, strings.Join(v.Args, " "))
}

// FileExistsValidator validates that a file exists
type FileExistsValidator struct {
	Path           string
	ShouldExist    bool
	MinSize        int64 // Minimum file size in bytes (0 = no check)
	ContentPattern string // Regex pattern to match file content
	Config         ValidatorConfig
	Desc           string
}

// NewFileExistsValidator creates a new file exists validator from a definition
func NewFileExistsValidator(def ValidationDefinition) *FileExistsValidator {
	shouldExist := true
	if exists, ok := def.Options["should_exist"].(bool); ok {
		shouldExist = exists
	}

	var minSize int64
	if size, ok := def.Options["min_size"].(float64); ok {
		minSize = int64(size)
	}

	return &FileExistsValidator{
		Path:           def.Path,
		ShouldExist:    shouldExist,
		MinSize:        minSize,
		ContentPattern: def.Pattern,
		Config:         DefaultValidatorConfig(),
		Desc:           def.Description,
	}
}

// Validate performs the file existence validation
func (v *FileExistsValidator) Validate(ctx context.Context) ValidationResult {
	start := time.Now()
	result := ValidationResult{
		ValidatorID: fmt.Sprintf("file_%s", sanitizePath(v.Path)),
	}

	// Check if file exists
	info, err := os.Stat(v.Path)
	exists := err == nil

	if v.ShouldExist {
		if !exists {
			result.Success = false
			result.Message = fmt.Sprintf("file does not exist: %s", v.Path)
			result.Error = "file not found"
			result.Duration = time.Since(start)
			return result
		}

		// Check minimum size
		if v.MinSize > 0 && info.Size() < v.MinSize {
			result.Success = false
			result.Message = fmt.Sprintf("file %s is smaller than expected (got %d bytes, expected >= %d)", v.Path, info.Size(), v.MinSize)
			result.Error = "file too small"
			result.Duration = time.Since(start)
			return result
		}

		// Check content pattern if specified
		if v.ContentPattern != "" {
			content, err := os.ReadFile(v.Path)
			if err != nil {
				result.Success = false
				result.Message = fmt.Sprintf("failed to read file: %s", v.Path)
				result.Error = err.Error()
				result.Duration = time.Since(start)
				return result
			}

			matched, err := regexp.MatchString(v.ContentPattern, string(content))
			if err != nil {
				result.Success = false
				result.Message = "invalid content pattern"
				result.Error = err.Error()
				result.Duration = time.Since(start)
				return result
			}
			if !matched {
				result.Success = false
				result.Message = fmt.Sprintf("file content does not match pattern %q", v.ContentPattern)
				result.Error = "content mismatch"
				result.Duration = time.Since(start)
				return result
			}
		}

		result.Success = true
		result.Message = fmt.Sprintf("file exists: %s (%d bytes)", v.Path, info.Size())
	} else {
		// File should NOT exist
		if exists {
			result.Success = false
			result.Message = fmt.Sprintf("file should not exist: %s", v.Path)
			result.Error = "file exists unexpectedly"
		} else {
			result.Success = true
			result.Message = fmt.Sprintf("file correctly does not exist: %s", v.Path)
		}
	}

	result.Duration = time.Since(start)
	return result
}

// Type returns the validation type
func (v *FileExistsValidator) Type() ValidationType {
	return ValidationTypeFileExists
}

// Description returns a human-readable description
func (v *FileExistsValidator) Description() string {
	if v.Desc != "" {
		return v.Desc
	}
	if v.ShouldExist {
		return fmt.Sprintf("file exists: %s", v.Path)
	}
	return fmt.Sprintf("file does not exist: %s", v.Path)
}

// OutputValidator validates that output contains expected patterns
type OutputValidator struct {
	Input   string
	Pattern string
	Inverse bool // If true, pattern should NOT match
	Config  ValidatorConfig
	Desc    string
}

// NewOutputValidator creates a new output validator from a definition
func NewOutputValidator(def ValidationDefinition) *OutputValidator {
	inverse := false
	if inv, ok := def.Options["inverse"].(bool); ok {
		inverse = inv
	}

	return &OutputValidator{
		Input:   def.Input,
		Pattern: def.Pattern,
		Inverse: inverse,
		Config:  DefaultValidatorConfig(),
		Desc:    def.Description,
	}
}

// Validate performs the output pattern validation
func (v *OutputValidator) Validate(ctx context.Context) ValidationResult {
	start := time.Now()
	result := ValidationResult{
		ValidatorID: "output_contains",
		Output:      v.Input,
	}

	matched, err := regexp.MatchString(v.Pattern, v.Input)
	if err != nil {
		result.Success = false
		result.Message = "invalid pattern"
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	if v.Inverse {
		// Pattern should NOT match
		if matched {
			result.Success = false
			result.Message = fmt.Sprintf("output unexpectedly matches pattern %q", v.Pattern)
			result.Error = "unexpected match"
		} else {
			result.Success = true
			result.Message = fmt.Sprintf("output correctly does not match pattern %q", v.Pattern)
		}
	} else {
		// Pattern should match
		if !matched {
			result.Success = false
			result.Message = fmt.Sprintf("output does not match pattern %q", v.Pattern)
			result.Error = "no match"
		} else {
			result.Success = true
			result.Message = fmt.Sprintf("output matches pattern %q", v.Pattern)
		}
	}

	result.Duration = time.Since(start)
	return result
}

// Type returns the validation type
func (v *OutputValidator) Type() ValidationType {
	return ValidationTypeOutputContains
}

// Description returns a human-readable description
func (v *OutputValidator) Description() string {
	if v.Desc != "" {
		return v.Desc
	}
	if v.Inverse {
		return fmt.Sprintf("output does not contain: %s", v.Pattern)
	}
	return fmt.Sprintf("output contains: %s", v.Pattern)
}

// CreateValidator creates a validator from a validation definition
func CreateValidator(def ValidationDefinition) (Validator, error) {
	switch def.Type {
	case ValidationTypeHTTPGet, ValidationTypeHTTPPost:
		if def.URL == "" {
			return nil, fmt.Errorf("URL is required for HTTP validation")
		}
		return NewEndpointValidator(def), nil

	case ValidationTypeCLI:
		if def.Command == "" {
			return nil, fmt.Errorf("command is required for CLI validation")
		}
		return NewCLIValidator(def), nil

	case ValidationTypeFileExists:
		if def.Path == "" {
			return nil, fmt.Errorf("path is required for file_exists validation")
		}
		return NewFileExistsValidator(def), nil

	case ValidationTypeOutputContains:
		if def.Pattern == "" {
			return nil, fmt.Errorf("pattern is required for output_contains validation")
		}
		return NewOutputValidator(def), nil

	default:
		return nil, fmt.Errorf("unknown validation type: %s", def.Type)
	}
}

// ParseValidationType parses a string into a ValidationType
func ParseValidationType(s string) (ValidationType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "http_get", "get", "http-get":
		return ValidationTypeHTTPGet, nil
	case "http_post", "post", "http-post":
		return ValidationTypeHTTPPost, nil
	case "cli_command", "cli", "command":
		return ValidationTypeCLI, nil
	case "file_exists", "file", "exists":
		return ValidationTypeFileExists, nil
	case "output_contains", "output", "contains":
		return ValidationTypeOutputContains, nil
	default:
		return "", fmt.Errorf("unknown validation type %q: must be one of http_get, http_post, cli_command, file_exists, output_contains", s)
	}
}

// Helper functions

func sanitizeURL(url string) string {
	url = strings.ReplaceAll(url, "://", "_")
	url = strings.ReplaceAll(url, "/", "_")
	url = strings.ReplaceAll(url, ":", "_")
	if len(url) > 30 {
		url = url[:30]
	}
	return url
}

func sanitizeCommand(cmd string) string {
	cmd = strings.ReplaceAll(cmd, "/", "_")
	cmd = strings.ReplaceAll(cmd, " ", "_")
	if len(cmd) > 20 {
		cmd = cmd[:20]
	}
	return cmd
}

func sanitizePath(path string) string {
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, "\\", "_")
	if len(path) > 30 {
		path = path[:30]
	}
	return path
}

// ValidationRunner runs multiple validations and aggregates results
type ValidationRunner struct {
	Validators []Validator
	Timeout    time.Duration
}

// NewValidationRunner creates a new validation runner
func NewValidationRunner() *ValidationRunner {
	return &ValidationRunner{
		Timeout: DefaultTimeout * 10, // Overall timeout for all validations
	}
}

// AddValidator adds a validator to the runner
func (r *ValidationRunner) AddValidator(v Validator) {
	r.Validators = append(r.Validators, v)
}

// AddFromDefinitions creates validators from definitions and adds them
func (r *ValidationRunner) AddFromDefinitions(defs []ValidationDefinition) error {
	for _, def := range defs {
		v, err := CreateValidator(def)
		if err != nil {
			return fmt.Errorf("failed to create validator: %w", err)
		}
		r.AddValidator(v)
	}
	return nil
}

// ValidationRunResult represents the results of running all validations
type ValidationRunResult struct {
	Success      bool               `json:"success"`
	TotalCount   int                `json:"total_count"`
	PassedCount  int                `json:"passed_count"`
	FailedCount  int                `json:"failed_count"`
	Results      []ValidationResult `json:"results"`
	Duration     time.Duration      `json:"duration"`
	FeatureID    int                `json:"feature_id,omitempty"`
	FeatureName  string             `json:"feature_name,omitempty"`
}

// Run executes all validators and returns aggregated results
func (r *ValidationRunner) Run(ctx context.Context) ValidationRunResult {
	start := time.Now()
	runResult := ValidationRunResult{
		TotalCount: len(r.Validators),
	}

	// Create context with overall timeout
	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	for _, v := range r.Validators {
		result := v.Validate(ctx)
		runResult.Results = append(runResult.Results, result)

		if result.Success {
			runResult.PassedCount++
		} else {
			runResult.FailedCount++
		}
	}

	runResult.Success = runResult.FailedCount == 0
	runResult.Duration = time.Since(start)
	return runResult
}

// Summary returns a human-readable summary of the run results
func (r *ValidationRunResult) Summary() string {
	var sb strings.Builder

	if r.FeatureName != "" {
		sb.WriteString(fmt.Sprintf("Feature #%d: %s\n", r.FeatureID, r.FeatureName))
	}

	status := "PASSED"
	if !r.Success {
		status = "FAILED"
	}

	sb.WriteString(fmt.Sprintf("Validation: %s (%d/%d passed, %s)\n",
		status, r.PassedCount, r.TotalCount, r.Duration.Round(time.Millisecond)))

	for _, result := range r.Results {
		icon := "✓"
		if !result.Success {
			icon = "✗"
		}
		sb.WriteString(fmt.Sprintf("  %s %s", icon, result.Message))
		if result.Error != "" {
			sb.WriteString(fmt.Sprintf(" [%s]", result.Error))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatJSON returns the run results as formatted JSON
func (r *ValidationRunResult) FormatJSON() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(data)
}
