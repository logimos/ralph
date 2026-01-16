package recovery

import (
	"strings"
	"testing"
)

func TestParseStrategyType(t *testing.T) {
	tests := []struct {
		input   string
		want    StrategyType
		wantErr bool
	}{
		{"retry", StrategyRetry, false},
		{"RETRY", StrategyRetry, false},
		{"Retry", StrategyRetry, false},
		{"skip", StrategySkip, false},
		{"SKIP", StrategySkip, false},
		{"rollback", StrategyRollback, false},
		{"ROLLBACK", StrategyRollback, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		got, err := ParseStrategyType(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseStrategyType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseStrategyType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRetryStrategy_Name(t *testing.T) {
	tracker := NewFailureTracker(3)
	s := NewRetryStrategy(3, tracker)
	if s.Name() != StrategyRetry {
		t.Errorf("RetryStrategy.Name() = %v, want %v", s.Name(), StrategyRetry)
	}
}

func TestRetryStrategy_Description(t *testing.T) {
	tracker := NewFailureTracker(3)
	s := NewRetryStrategy(3, tracker)
	desc := s.Description()
	if desc == "" {
		t.Error("RetryStrategy.Description() should not be empty")
	}
}

func TestRetryStrategy_Apply_CanRetry(t *testing.T) {
	tracker := NewFailureTracker(3)
	s := NewRetryStrategy(3, tracker)

	failure := &Failure{
		Type:      FailureTypeTest,
		Message:   "Test failed",
		FeatureID: 1,
		Iteration: 1,
	}
	tracker.RecordFailure(failure)

	result := s.Apply(failure)

	if !result.Success {
		t.Error("RetryStrategy.Apply() should succeed when retries available")
	}
	if !result.ShouldRetry {
		t.Error("RetryStrategy.Apply() should set ShouldRetry=true")
	}
	if result.ShouldSkip {
		t.Error("RetryStrategy.Apply() should set ShouldSkip=false")
	}
	if result.ModifiedPrompt == "" {
		t.Error("RetryStrategy.Apply() should generate modified prompt")
	}
}

func TestRetryStrategy_Apply_ExceedsMaxRetries(t *testing.T) {
	tracker := NewFailureTracker(2)
	s := NewRetryStrategy(2, tracker)

	// Record max retries
	tracker.RecordFailure(&Failure{FeatureID: 1})
	tracker.RecordFailure(&Failure{FeatureID: 1})

	failure := &Failure{
		Type:      FailureTypeTest,
		FeatureID: 1,
	}

	result := s.Apply(failure)

	if result.Success {
		t.Error("RetryStrategy.Apply() should fail when max retries exceeded")
	}
	if result.ShouldRetry {
		t.Error("RetryStrategy.Apply() should set ShouldRetry=false when exceeded")
	}
	if !result.ShouldSkip {
		t.Error("RetryStrategy.Apply() should escalate to skip when exceeded")
	}
}

func TestRetryStrategy_GenerateRetryPrompt(t *testing.T) {
	tracker := NewFailureTracker(3)
	s := NewRetryStrategy(3, tracker)

	testCases := []struct {
		failureType FailureType
		expectWord  string
	}{
		{FailureTypeTest, "test"},
		{FailureTypeTypeCheck, "type"},
		{FailureTypeTimeout, "timed out"},
		{FailureTypeAgentError, "error"},
	}

	for _, tc := range testCases {
		failure := &Failure{
			Type:    tc.failureType,
			Message: "Some error message",
		}
		prompt := s.generateRetryPrompt(failure)
		if !strings.Contains(strings.ToLower(prompt), tc.expectWord) {
			t.Errorf("generateRetryPrompt(%v) should contain %q, got: %s", tc.failureType, tc.expectWord, prompt)
		}
		if !strings.Contains(prompt, "IMPORTANT") {
			t.Errorf("generateRetryPrompt should contain IMPORTANT marker")
		}
	}
}

func TestSkipStrategy_Name(t *testing.T) {
	tracker := NewFailureTracker(3)
	s := NewSkipStrategy(tracker)
	if s.Name() != StrategySkip {
		t.Errorf("SkipStrategy.Name() = %v, want %v", s.Name(), StrategySkip)
	}
}

func TestSkipStrategy_Apply(t *testing.T) {
	tracker := NewFailureTracker(3)
	s := NewSkipStrategy(tracker)

	// Record some failures
	tracker.RecordFailure(&Failure{FeatureID: 1})
	tracker.RecordFailure(&Failure{FeatureID: 1})

	failure := &Failure{
		Type:      FailureTypeTest,
		FeatureID: 1,
	}

	result := s.Apply(failure)

	if !result.Success {
		t.Error("SkipStrategy.Apply() should always succeed")
	}
	if result.ShouldRetry {
		t.Error("SkipStrategy.Apply() should set ShouldRetry=false")
	}
	if !result.ShouldSkip {
		t.Error("SkipStrategy.Apply() should set ShouldSkip=true")
	}
	if !strings.Contains(result.Message, "2 failure(s)") {
		t.Errorf("SkipStrategy.Apply() message should mention failure count, got: %s", result.Message)
	}
}

func TestRollbackStrategy_Name(t *testing.T) {
	tracker := NewFailureTracker(3)
	s := NewRollbackStrategy(tracker)
	if s.Name() != StrategyRollback {
		t.Errorf("RollbackStrategy.Name() = %v, want %v", s.Name(), StrategyRollback)
	}
}

// Note: RollbackStrategy.Apply() requires git operations which are difficult to test
// without a real git repository. These would be better as integration tests.

func TestNewRecoveryManager(t *testing.T) {
	rm := NewRecoveryManager(5, StrategyRetry)
	if rm == nil {
		t.Fatal("NewRecoveryManager returned nil")
	}
	if rm.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", rm.maxRetries)
	}
	if rm.defaultStrategy != StrategyRetry {
		t.Errorf("defaultStrategy = %v, want %v", rm.defaultStrategy, StrategyRetry)
	}
	if len(rm.strategies) != 3 {
		t.Errorf("Expected 3 strategies, got %d", len(rm.strategies))
	}
}

func TestRecoveryManager_GetTracker(t *testing.T) {
	rm := NewRecoveryManager(3, StrategyRetry)
	tracker := rm.GetTracker()
	if tracker == nil {
		t.Error("GetTracker() should not return nil")
	}
}

func TestRecoveryManager_HandleFailure_NoFailure(t *testing.T) {
	rm := NewRecoveryManager(3, StrategyRetry)

	failure, result := rm.HandleFailure("All tests passed", 0, 1, 1)

	if failure != nil {
		t.Error("HandleFailure should return nil failure for successful output")
	}
	if !result.Success {
		t.Error("HandleFailure result should be successful for no failure")
	}
}

func TestRecoveryManager_HandleFailure_WithFailure(t *testing.T) {
	rm := NewRecoveryManager(3, StrategyRetry)

	failure, result := rm.HandleFailure("FAIL: TestSomething", 1, 1, 1)

	if failure == nil {
		t.Fatal("HandleFailure should detect failure")
	}
	if failure.Type != FailureTypeTest {
		t.Errorf("failure.Type = %v, want test_failure", failure.Type)
	}
	if !result.ShouldRetry {
		t.Error("With retry strategy, result should indicate retry")
	}
}

func TestRecoveryManager_HandleFailure_SkipStrategy(t *testing.T) {
	rm := NewRecoveryManager(3, StrategySkip)

	failure, result := rm.HandleFailure("FAIL: TestSomething", 1, 1, 1)

	if failure == nil {
		t.Fatal("HandleFailure should detect failure")
	}
	if !result.ShouldSkip {
		t.Error("With skip strategy, result should indicate skip")
	}
}

func TestRecoveryManager_SelectStrategy_ExceedsMaxRetries(t *testing.T) {
	// Use maxRetries=2 so first failure allows retry, second exhausts budget
	rm := NewRecoveryManager(2, StrategyRetry)

	// First failure - should retry (use explicit test failure message)
	failure1, result1 := rm.HandleFailure("--- FAIL: TestSomething\ntest failed", 1, 1, 1)
	if failure1 == nil {
		t.Fatal("First failure should be detected")
	}
	if !result1.ShouldRetry {
		t.Error("First failure should allow retry")
	}

	// Second failure - still within retry budget
	failure2, result2 := rm.HandleFailure("--- FAIL: TestSomething\ntest failed", 1, 1, 2)
	if failure2 == nil {
		t.Fatal("Second failure should be detected")
	}
	if result2.ShouldRetry {
		t.Error("After max retries, should not allow retry")
	}
	if !result2.ShouldSkip {
		t.Error("After max retries, should escalate to skip")
	}
}

func TestRecoveryManager_GetFailureSummary(t *testing.T) {
	rm := NewRecoveryManager(3, StrategyRetry)

	// No failures
	summary := rm.GetFailureSummary()
	if summary != "No failures recorded" {
		t.Errorf("Empty manager summary = %q, want 'No failures recorded'", summary)
	}

	// With failures
	rm.HandleFailure("FAIL: test1", 1, 1, 1)
	rm.HandleFailure("build failed", 1, 2, 1)

	summary = rm.GetFailureSummary()
	if !strings.Contains(summary, "Failure Summary") {
		t.Error("Summary should contain 'Failure Summary'")
	}
}

func TestRecoveryManager_ShouldEscalate(t *testing.T) {
	rm := NewRecoveryManager(2, StrategyRetry)

	if rm.ShouldEscalate(1) {
		t.Error("ShouldEscalate should return false initially")
	}

	rm.HandleFailure("FAIL", 1, 1, 1)
	if rm.ShouldEscalate(1) {
		t.Error("ShouldEscalate should return false after 1 failure")
	}

	rm.HandleFailure("FAIL", 1, 1, 2)
	if !rm.ShouldEscalate(1) {
		t.Error("ShouldEscalate should return true after max retries")
	}
}

func TestRecoveryResult_Fields(t *testing.T) {
	result := RecoveryResult{
		Success:        true,
		Message:        "Recovery applied",
		ShouldRetry:    true,
		ShouldSkip:     false,
		ModifiedPrompt: "Please fix the error",
	}

	if !result.Success {
		t.Error("Success should be true")
	}
	if result.ShouldSkip {
		t.Error("ShouldSkip should be false")
	}
	if result.ModifiedPrompt != "Please fix the error" {
		t.Error("ModifiedPrompt not set correctly")
	}
}
