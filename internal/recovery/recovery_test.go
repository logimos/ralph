package recovery

import (
	"strings"
	"testing"
	"time"
)

func TestFailureType_String(t *testing.T) {
	tests := []struct {
		ft   FailureType
		want string
	}{
		{FailureTypeNone, "none"},
		{FailureTypeTest, "test_failure"},
		{FailureTypeTypeCheck, "typecheck_failure"},
		{FailureTypeAgentError, "agent_error"},
		{FailureTypeTimeout, "timeout"},
	}

	for _, tt := range tests {
		if got := string(tt.ft); got != tt.want {
			t.Errorf("FailureType = %v, want %v", got, tt.want)
		}
	}
}

func TestFailure_String(t *testing.T) {
	f := &Failure{
		Type:       FailureTypeTest,
		Message:    "Test assertion failed",
		FeatureID:  5,
		Iteration:  2,
		RetryCount: 1,
	}

	result := f.String()
	if !strings.Contains(result, "test_failure") {
		t.Errorf("Failure.String() should contain failure type")
	}
	if !strings.Contains(result, "feature #5") {
		t.Errorf("Failure.String() should contain feature ID")
	}
	if !strings.Contains(result, "iteration 2") {
		t.Errorf("Failure.String() should contain iteration")
	}
}

func TestNewFailureTracker(t *testing.T) {
	ft := NewFailureTracker(5)
	if ft == nil {
		t.Fatal("NewFailureTracker returned nil")
	}
	if ft.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", ft.maxRetries)
	}
}

func TestFailureTracker_RecordFailure(t *testing.T) {
	ft := NewFailureTracker(3)

	failure := &Failure{
		Type:      FailureTypeTest,
		Message:   "Test failed",
		FeatureID: 1,
		Iteration: 1,
	}

	ft.RecordFailure(failure)

	if count := ft.GetRetryCount(1); count != 1 {
		t.Errorf("GetRetryCount(1) = %d, want 1", count)
	}
	if failure.RetryCount != 1 {
		t.Errorf("failure.RetryCount = %d, want 1", failure.RetryCount)
	}

	failures := ft.GetFailures(1)
	if len(failures) != 1 {
		t.Errorf("GetFailures(1) len = %d, want 1", len(failures))
	}
}

func TestFailureTracker_CanRetry(t *testing.T) {
	ft := NewFailureTracker(2)

	// Should be able to retry initially
	if !ft.CanRetry(1) {
		t.Error("CanRetry(1) should return true initially")
	}

	// Record first failure
	ft.RecordFailure(&Failure{FeatureID: 1})
	if !ft.CanRetry(1) {
		t.Error("CanRetry(1) should return true after 1 failure")
	}

	// Record second failure (reaches max)
	ft.RecordFailure(&Failure{FeatureID: 1})
	if ft.CanRetry(1) {
		t.Error("CanRetry(1) should return false after reaching max retries")
	}
}

func TestFailureTracker_ResetFeature(t *testing.T) {
	ft := NewFailureTracker(3)

	ft.RecordFailure(&Failure{FeatureID: 1})
	ft.RecordFailure(&Failure{FeatureID: 1})

	if count := ft.GetRetryCount(1); count != 2 {
		t.Errorf("GetRetryCount before reset = %d, want 2", count)
	}

	ft.ResetFeature(1)

	if count := ft.GetRetryCount(1); count != 0 {
		t.Errorf("GetRetryCount after reset = %d, want 0", count)
	}
}

func TestFailureTracker_GetSummary(t *testing.T) {
	ft := NewFailureTracker(3)

	// Empty tracker
	summary := ft.GetSummary()
	if summary != "No failures recorded" {
		t.Errorf("Empty tracker summary = %q, want 'No failures recorded'", summary)
	}

	// With failures
	ft.RecordFailure(&Failure{Type: FailureTypeTest, FeatureID: 1})
	ft.RecordFailure(&Failure{Type: FailureTypeTypeCheck, FeatureID: 1})
	ft.RecordFailure(&Failure{Type: FailureTypeTest, FeatureID: 2})

	summary = ft.GetSummary()
	if !strings.Contains(summary, "Feature #1") {
		t.Error("Summary should contain Feature #1")
	}
	if !strings.Contains(summary, "Feature #2") {
		t.Error("Summary should contain Feature #2")
	}
	if !strings.Contains(summary, "Total failures: 3") {
		t.Errorf("Summary should contain total failures, got: %s", summary)
	}
}

func TestDetectFailure_NoFailure(t *testing.T) {
	output := "Build successful\nAll tests passed"
	failure := DetectFailure(output, 0, 1, 1)
	if failure != nil {
		t.Errorf("DetectFailure should return nil for successful output, got: %v", failure)
	}
}

func TestDetectFailure_ExitCode(t *testing.T) {
	output := "Some output"
	failure := DetectFailure(output, 1, 1, 1)
	if failure == nil {
		t.Fatal("DetectFailure should detect non-zero exit code")
	}
	if failure.Type != FailureTypeAgentError {
		t.Errorf("failure.Type = %v, want agent_error", failure.Type)
	}
}

func TestDetectFailure_TestFailure(t *testing.T) {
	testCases := []string{
		"--- FAIL: TestSomething\ntest failed",
		"FAIL github.com/foo/bar 0.123s",
		"Test assertion failed: expected 1 got 2",
		"panic: test timed out",
	}

	for _, output := range testCases {
		failure := DetectFailure(output, 1, 1, 1)
		if failure == nil {
			t.Errorf("DetectFailure should detect test failure in: %q", output)
			continue
		}
		if failure.Type != FailureTypeTest {
			t.Errorf("For %q: failure.Type = %v, want test_failure", output, failure.Type)
		}
	}
}

func TestDetectFailure_TypeCheckFailure(t *testing.T) {
	testCases := []string{
		"cannot find module foo",
		"undefined: SomeFunction",
		"syntax error: unexpected }",
		"compilation failed",
		"build failed: cannot compile",
	}

	for _, output := range testCases {
		failure := DetectFailure(output, 1, 1, 1)
		if failure == nil {
			t.Errorf("DetectFailure should detect typecheck failure in: %q", output)
			continue
		}
		if failure.Type != FailureTypeTypeCheck {
			t.Errorf("For %q: failure.Type = %v, want typecheck_failure", output, failure.Type)
		}
	}
}

func TestDetectFailure_Timeout(t *testing.T) {
	testCases := []string{
		"context deadline exceeded",
		"operation timed out",
		"timeout waiting for response",
	}

	for _, output := range testCases {
		failure := DetectFailure(output, 1, 1, 1)
		if failure == nil {
			t.Errorf("DetectFailure should detect timeout in: %q", output)
			continue
		}
		if failure.Type != FailureTypeTimeout {
			t.Errorf("For %q: failure.Type = %v, want timeout", output, failure.Type)
		}
	}
}

func TestDetectFailure_SetsMetadata(t *testing.T) {
	before := time.Now()
	failure := DetectFailure("FAIL: test error", 1, 5, 3)
	after := time.Now()

	if failure == nil {
		t.Fatal("Expected failure to be detected")
	}

	if failure.FeatureID != 5 {
		t.Errorf("FeatureID = %d, want 5", failure.FeatureID)
	}
	if failure.Iteration != 3 {
		t.Errorf("Iteration = %d, want 3", failure.Iteration)
	}
	if failure.Timestamp.Before(before) || failure.Timestamp.After(after) {
		t.Error("Timestamp should be set to current time")
	}
	if failure.Output != "FAIL: test error" {
		t.Errorf("Output not preserved correctly")
	}
}

func TestDetectFailureFromOutput_Priority(t *testing.T) {
	// Test failures should have higher priority than compilation errors
	// when both indicators are present in test-related output
	output := "running tests...\n--- FAIL: TestX\nerror: assertion failed"
	failureType := detectFailureFromOutput(output)
	if failureType != FailureTypeTest {
		t.Errorf("detectFailureFromOutput should prioritize test failures, got %v", failureType)
	}
}

func TestIsTestRelated(t *testing.T) {
	testRelated := []string{
		"running test suite",
		"TestSomething",
		"describe('feature')",
		"it('should work')",
		"=== RUN TestX",
		"pytest collected 5 items",
	}

	for _, output := range testRelated {
		if !isTestRelated(strings.ToLower(output)) {
			t.Errorf("isTestRelated(%q) should return true", output)
		}
	}

	notTestRelated := []string{
		"building main.go",
		"linking binary",
	}

	for _, output := range notTestRelated {
		if isTestRelated(strings.ToLower(output)) {
			t.Errorf("isTestRelated(%q) should return false", output)
		}
	}
}

func TestGetFailureMessage(t *testing.T) {
	// Should extract error line
	output := "Building...\nerror: undefined variable x\nDone"
	msg := getFailureMessage(FailureTypeTypeCheck, output)
	if !strings.Contains(strings.ToLower(msg), "undefined") {
		t.Errorf("getFailureMessage should extract error line, got: %q", msg)
	}

	// Should return default message if no match
	msg = getFailureMessage(FailureTypeTest, "no clear error")
	if msg == "" {
		t.Error("getFailureMessage should return default message")
	}
}
