// Package recovery provides failure handling and recovery strategies for Ralph.
package recovery

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// FailureType represents the type of failure encountered
type FailureType string

const (
	// FailureTypeNone indicates no failure
	FailureTypeNone FailureType = "none"
	// FailureTypeTest indicates test failures
	FailureTypeTest FailureType = "test_failure"
	// FailureTypeTypeCheck indicates type check/compilation failures
	FailureTypeTypeCheck FailureType = "typecheck_failure"
	// FailureTypeAgentError indicates agent execution errors
	FailureTypeAgentError FailureType = "agent_error"
	// FailureTypeTimeout indicates timeout failures
	FailureTypeTimeout FailureType = "timeout"
)

// Failure represents a detected failure with context
type Failure struct {
	Type        FailureType
	Message     string
	Output      string
	FeatureID   int
	Iteration   int
	Timestamp   time.Time
	RetryCount  int
}

// String returns a human-readable representation of the failure
func (f *Failure) String() string {
	return fmt.Sprintf("[%s] %s (feature #%d, iteration %d, retries: %d)",
		f.Type, f.Message, f.FeatureID, f.Iteration, f.RetryCount)
}

// FailureTracker tracks failures per feature and manages retry counts
type FailureTracker struct {
	failures   map[int][]*Failure // featureID -> list of failures
	retryCounts map[int]int       // featureID -> current retry count
	maxRetries int
}

// NewFailureTracker creates a new failure tracker with the specified max retries
func NewFailureTracker(maxRetries int) *FailureTracker {
	return &FailureTracker{
		failures:    make(map[int][]*Failure),
		retryCounts: make(map[int]int),
		maxRetries:  maxRetries,
	}
}

// RecordFailure records a failure for a feature
func (ft *FailureTracker) RecordFailure(failure *Failure) {
	featureID := failure.FeatureID
	ft.failures[featureID] = append(ft.failures[featureID], failure)
	ft.retryCounts[featureID]++
	failure.RetryCount = ft.retryCounts[featureID]
}

// GetRetryCount returns the current retry count for a feature
func (ft *FailureTracker) GetRetryCount(featureID int) int {
	return ft.retryCounts[featureID]
}

// CanRetry returns true if the feature can be retried (hasn't exceeded max retries)
func (ft *FailureTracker) CanRetry(featureID int) bool {
	return ft.retryCounts[featureID] < ft.maxRetries
}

// GetFailures returns all failures for a feature
func (ft *FailureTracker) GetFailures(featureID int) []*Failure {
	return ft.failures[featureID]
}

// ResetFeature resets the retry count for a feature (e.g., after successful completion)
func (ft *FailureTracker) ResetFeature(featureID int) {
	ft.retryCounts[featureID] = 0
}

// GetSummary returns a summary of all tracked failures
func (ft *FailureTracker) GetSummary() string {
	if len(ft.failures) == 0 {
		return "No failures recorded"
	}
	
	var sb strings.Builder
	sb.WriteString("Failure Summary:\n")
	
	totalFailures := 0
	for featureID, failures := range ft.failures {
		sb.WriteString(fmt.Sprintf("  Feature #%d: %d failure(s)\n", featureID, len(failures)))
		totalFailures += len(failures)
		
		// Group by type
		byType := make(map[FailureType]int)
		for _, f := range failures {
			byType[f.Type]++
		}
		for t, count := range byType {
			sb.WriteString(fmt.Sprintf("    - %s: %d\n", t, count))
		}
	}
	
	sb.WriteString(fmt.Sprintf("Total failures: %d\n", totalFailures))
	return sb.String()
}

// DetectFailure analyzes agent output and command results to detect failures
func DetectFailure(output string, exitCode int, featureID, iteration int) *Failure {
	// Check for explicit failure indicators
	failure := &Failure{
		FeatureID: featureID,
		Iteration: iteration,
		Timestamp: time.Now(),
		Output:    output,
	}

	// Check exit code first
	if exitCode != 0 {
		failure.Type = FailureTypeAgentError
		failure.Message = fmt.Sprintf("Command exited with code %d", exitCode)
		
		// Try to determine more specific failure type from output
		specificType := detectFailureFromOutput(output)
		if specificType != FailureTypeNone {
			failure.Type = specificType
			failure.Message = getFailureMessage(specificType, output)
		}
		
		return failure
	}

	// Check for failures even with exit code 0 (some tools don't propagate exit codes)
	failureType := detectFailureFromOutput(output)
	if failureType != FailureTypeNone {
		failure.Type = failureType
		failure.Message = getFailureMessage(failureType, output)
		return failure
	}

	return nil // No failure detected
}

// detectFailureFromOutput analyzes output text to detect failure type
func detectFailureFromOutput(output string) FailureType {
	outputLower := strings.ToLower(output)

	// Check timeout first (most specific) - but not if it's in test output context
	if !isTestRelated(outputLower) {
		timeoutPatterns := []string{
			"timeout",
			"timed out",
			"deadline exceeded",
			"context deadline",
		}
		for _, pattern := range timeoutPatterns {
			if strings.Contains(outputLower, pattern) {
				return FailureTypeTimeout
			}
		}
	}

	// Type check / compilation failure patterns (check before test failures)
	// These are more specific compilation/type errors
	typeCheckPatterns := []string{
		"cannot find module",
		"cannot find package",
		"cannot find",
		"undefined:",
		"type error",
		"syntax error",
		"compilation failed",
		"build failed",
		"could not compile",
		"cannot compile",
		"does not exist",
		"no such file",
		"undeclared name",
		"not declared",
		"import cycle",
	}
	for _, pattern := range typeCheckPatterns {
		if strings.Contains(outputLower, pattern) {
			return FailureTypeTypeCheck
		}
	}

	// Test failure patterns - explicit test failure markers
	testFailurePatterns := []string{
		"test failed",
		"tests failed",
		"assertion failed",
		"--- fail:",
		"=== fail",
	}
	for _, pattern := range testFailurePatterns {
		if strings.Contains(outputLower, pattern) {
			return FailureTypeTest
		}
	}

	// Check for Go test output pattern like "FAIL github.com/..." or "FAIL\t"
	if matched, _ := regexp.MatchString(`(?i)FAIL\s+(github\.com|gitlab\.com|bitbucket\.org|[a-z]+/[a-z]+)`, output); matched {
		return FailureTypeTest
	}

	// Check for general FAIL pattern with test context
	if matched, _ := regexp.MatchString(`(?i)\bFAIL\b`, output); matched {
		if isTestRelated(outputLower) {
			return FailureTypeTest
		}
	}

	// Check for panic in test context
	if strings.Contains(outputLower, "panic:") && isTestRelated(outputLower) {
		return FailureTypeTest
	}

	// Generic "error:" or "failed" - need test context to be test failure
	if (strings.Contains(outputLower, "error:") || strings.Contains(outputLower, "failed")) {
		if isTestRelated(outputLower) {
			return FailureTypeTest
		}
	}

	// Timeout patterns at lower priority (if we get here, not in test context)
	timeoutPatterns := []string{
		"timeout",
		"timed out",
		"deadline exceeded",
		"context deadline",
	}
	for _, pattern := range timeoutPatterns {
		if strings.Contains(outputLower, pattern) {
			return FailureTypeTimeout
		}
	}

	return FailureTypeNone
}

// isTestRelated checks if the output is related to test execution
func isTestRelated(output string) bool {
	testIndicators := []string{
		"test",
		"spec",
		"assert",
		"expect",
		"should",
		"describe",
		"it(",
		"--- fail",
		"--- pass",
		"=== run",
		"pytest",
		"jest",
		"mocha",
		"junit",
		"testng",
	}
	
	for _, indicator := range testIndicators {
		if strings.Contains(output, indicator) {
			return true
		}
	}
	return false
}

// getFailureMessage extracts a meaningful message from the output
func getFailureMessage(failureType FailureType, output string) string {
	lines := strings.Split(output, "\n")
	
	// Find the first error line
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		
		switch failureType {
		case FailureTypeTest:
			if strings.Contains(lineLower, "fail") || 
			   strings.Contains(lineLower, "error") ||
			   strings.Contains(lineLower, "panic") {
				return strings.TrimSpace(line)
			}
		case FailureTypeTypeCheck:
			if strings.Contains(lineLower, "error") ||
			   strings.Contains(lineLower, "cannot") ||
			   strings.Contains(lineLower, "undefined") {
				return strings.TrimSpace(line)
			}
		case FailureTypeTimeout:
			if strings.Contains(lineLower, "timeout") ||
			   strings.Contains(lineLower, "deadline") {
				return strings.TrimSpace(line)
			}
		case FailureTypeAgentError:
			if strings.Contains(lineLower, "error") ||
			   strings.Contains(lineLower, "failed") {
				return strings.TrimSpace(line)
			}
		}
	}

	// Default messages
	switch failureType {
	case FailureTypeTest:
		return "Test execution failed"
	case FailureTypeTypeCheck:
		return "Type check/compilation failed"
	case FailureTypeTimeout:
		return "Operation timed out"
	case FailureTypeAgentError:
		return "Agent execution error"
	default:
		return "Unknown failure"
	}
}
