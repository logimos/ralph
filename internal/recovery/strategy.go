package recovery

import (
	"fmt"
	"os/exec"
	"strings"
)

// StrategyType represents the type of recovery strategy
type StrategyType string

const (
	// StrategyRetry retries the same feature with modified prompt
	StrategyRetry StrategyType = "retry"
	// StrategySkip marks the feature as blocked and moves to the next
	StrategySkip StrategyType = "skip"
	// StrategyRollback reverts to the last known good state via git
	StrategyRollback StrategyType = "rollback"
)

// ParseStrategyType parses a string into a StrategyType
func ParseStrategyType(s string) (StrategyType, error) {
	switch strings.ToLower(s) {
	case "retry":
		return StrategyRetry, nil
	case "skip":
		return StrategySkip, nil
	case "rollback":
		return StrategyRollback, nil
	default:
		return "", fmt.Errorf("unknown recovery strategy: %s (valid: retry, skip, rollback)", s)
	}
}

// RecoveryResult represents the result of applying a recovery strategy
type RecoveryResult struct {
	Success       bool
	Message       string
	ShouldRetry   bool   // Should the feature be retried
	ShouldSkip    bool   // Should the feature be skipped
	ModifiedPrompt string // Optional modified prompt for retry
}

// RecoveryStrategy defines the interface for recovery strategies
type RecoveryStrategy interface {
	// Name returns the strategy name
	Name() StrategyType
	// Apply applies the recovery strategy for the given failure
	Apply(failure *Failure) RecoveryResult
	// Description returns a human-readable description of the strategy
	Description() string
}

// RetryStrategy implements retry with modified prompt emphasis
type RetryStrategy struct {
	maxRetries int
	tracker    *FailureTracker
}

// NewRetryStrategy creates a new retry strategy
func NewRetryStrategy(maxRetries int, tracker *FailureTracker) *RetryStrategy {
	return &RetryStrategy{
		maxRetries: maxRetries,
		tracker:    tracker,
	}
}

// Name returns the strategy name
func (s *RetryStrategy) Name() StrategyType {
	return StrategyRetry
}

// Description returns a human-readable description
func (s *RetryStrategy) Description() string {
	return "Retry the feature with enhanced prompt guidance based on the failure type"
}

// Apply applies the retry strategy
func (s *RetryStrategy) Apply(failure *Failure) RecoveryResult {
	if !s.tracker.CanRetry(failure.FeatureID) {
		return RecoveryResult{
			Success:     false,
			Message:     fmt.Sprintf("Max retries (%d) exceeded for feature #%d", s.maxRetries, failure.FeatureID),
			ShouldRetry: false,
			ShouldSkip:  true, // Escalate to skip
		}
	}

	// Generate modified prompt based on failure type
	modifiedPrompt := s.generateRetryPrompt(failure)

	retryCount := s.tracker.GetRetryCount(failure.FeatureID)
	return RecoveryResult{
		Success:        true,
		Message:        fmt.Sprintf("Retrying feature #%d (attempt %d/%d)", failure.FeatureID, retryCount+1, s.maxRetries),
		ShouldRetry:    true,
		ShouldSkip:     false,
		ModifiedPrompt: modifiedPrompt,
	}
}

// generateRetryPrompt creates a modified prompt that addresses the specific failure
func (s *RetryStrategy) generateRetryPrompt(failure *Failure) string {
	var emphasis string

	switch failure.Type {
	case FailureTypeTest:
		emphasis = fmt.Sprintf(`IMPORTANT: The previous attempt failed due to test failures.
Error: %s

Please focus on:
1. Fix the failing tests before making other changes
2. Ensure all test assertions pass
3. Run tests locally before completing`, failure.Message)

	case FailureTypeTypeCheck:
		emphasis = fmt.Sprintf(`IMPORTANT: The previous attempt failed due to type/compilation errors.
Error: %s

Please focus on:
1. Fix all type errors and compilation issues first
2. Ensure the code compiles cleanly
3. Check imports and dependencies`, failure.Message)

	case FailureTypeTimeout:
		emphasis = fmt.Sprintf(`IMPORTANT: The previous attempt timed out.
Error: %s

Please focus on:
1. Simplify the implementation if possible
2. Break down into smaller steps
3. Avoid long-running operations`, failure.Message)

	case FailureTypeAgentError:
		emphasis = fmt.Sprintf(`IMPORTANT: The previous attempt encountered an error.
Error: %s

Please focus on:
1. Review the error message carefully
2. Address the root cause
3. Verify the approach is correct`, failure.Message)

	default:
		emphasis = fmt.Sprintf(`IMPORTANT: The previous attempt failed.
Error: %s

Please review the error and try a different approach.`, failure.Message)
	}

	return emphasis
}

// SkipStrategy marks features as blocked and moves to the next
type SkipStrategy struct {
	tracker *FailureTracker
}

// NewSkipStrategy creates a new skip strategy
func NewSkipStrategy(tracker *FailureTracker) *SkipStrategy {
	return &SkipStrategy{
		tracker: tracker,
	}
}

// Name returns the strategy name
func (s *SkipStrategy) Name() StrategyType {
	return StrategySkip
}

// Description returns a human-readable description
func (s *SkipStrategy) Description() string {
	return "Mark the feature as blocked and proceed to the next feature"
}

// Apply applies the skip strategy
func (s *SkipStrategy) Apply(failure *Failure) RecoveryResult {
	failures := s.tracker.GetFailures(failure.FeatureID)
	
	return RecoveryResult{
		Success:     true,
		Message:     fmt.Sprintf("Skipping feature #%d after %d failure(s). Moving to next feature.", failure.FeatureID, len(failures)),
		ShouldRetry: false,
		ShouldSkip:  true,
	}
}

// RollbackStrategy reverts changes via git
type RollbackStrategy struct {
	tracker *FailureTracker
}

// NewRollbackStrategy creates a new rollback strategy
func NewRollbackStrategy(tracker *FailureTracker) *RollbackStrategy {
	return &RollbackStrategy{
		tracker: tracker,
	}
}

// Name returns the strategy name
func (s *RollbackStrategy) Name() StrategyType {
	return StrategyRollback
}

// Description returns a human-readable description
func (s *RollbackStrategy) Description() string {
	return "Revert to the last known good state using git, then retry"
}

// Apply applies the rollback strategy
func (s *RollbackStrategy) Apply(failure *Failure) RecoveryResult {
	// Check if we're in a git repository
	if !isGitRepo() {
		return RecoveryResult{
			Success:     false,
			Message:     "Cannot rollback: not in a git repository",
			ShouldRetry: false,
			ShouldSkip:  true, // Fall back to skip
		}
	}

	// Check for uncommitted changes
	if !hasUncommittedChanges() {
		return RecoveryResult{
			Success:     false,
			Message:     "Cannot rollback: no uncommitted changes to revert",
			ShouldRetry: true, // Just retry without rollback
			ShouldSkip:  false,
		}
	}

	// Perform git checkout to discard changes
	if err := gitCheckoutAll(); err != nil {
		return RecoveryResult{
			Success:     false,
			Message:     fmt.Sprintf("Rollback failed: %v", err),
			ShouldRetry: false,
			ShouldSkip:  true,
		}
	}

	return RecoveryResult{
		Success:     true,
		Message:     fmt.Sprintf("Rolled back changes for feature #%d. Clean state restored.", failure.FeatureID),
		ShouldRetry: true,
		ShouldSkip:  false,
	}
}

// isGitRepo checks if the current directory is a git repository
func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

// hasUncommittedChanges checks if there are uncommitted changes
func hasUncommittedChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// gitCheckoutAll discards all uncommitted changes
func gitCheckoutAll() error {
	// First, reset staged changes
	resetCmd := exec.Command("git", "reset", "HEAD", "--")
	if err := resetCmd.Run(); err != nil {
		// Ignore error, might not have anything staged
	}

	// Then checkout all tracked files to discard changes
	checkoutCmd := exec.Command("git", "checkout", "--", ".")
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	// Clean untracked files (only if they were created during this run)
	// Note: We don't use -f -d here to be safe - only tracked file changes are reverted
	
	return nil
}

// RecoveryManager orchestrates failure detection and recovery
type RecoveryManager struct {
	tracker          *FailureTracker
	defaultStrategy  StrategyType
	strategies       map[StrategyType]RecoveryStrategy
	maxRetries       int
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(maxRetries int, defaultStrategy StrategyType) *RecoveryManager {
	tracker := NewFailureTracker(maxRetries)
	
	return &RecoveryManager{
		tracker:         tracker,
		defaultStrategy: defaultStrategy,
		maxRetries:      maxRetries,
		strategies: map[StrategyType]RecoveryStrategy{
			StrategyRetry:    NewRetryStrategy(maxRetries, tracker),
			StrategySkip:     NewSkipStrategy(tracker),
			StrategyRollback: NewRollbackStrategy(tracker),
		},
	}
}

// GetTracker returns the failure tracker
func (rm *RecoveryManager) GetTracker() *FailureTracker {
	return rm.tracker
}

// HandleFailure processes a failure and applies the appropriate recovery strategy
func (rm *RecoveryManager) HandleFailure(output string, exitCode int, featureID, iteration int) (*Failure, RecoveryResult) {
	// Detect failure
	failure := DetectFailure(output, exitCode, featureID, iteration)
	if failure == nil {
		return nil, RecoveryResult{Success: true, Message: "No failure detected"}
	}

	// Record the failure
	rm.tracker.RecordFailure(failure)

	// Select strategy based on failure type and configuration
	strategy := rm.selectStrategy(failure)

	// Apply the strategy
	result := strategy.Apply(failure)

	return failure, result
}

// selectStrategy chooses the appropriate strategy based on failure and config
func (rm *RecoveryManager) selectStrategy(failure *Failure) RecoveryStrategy {
	// Check if we've exceeded max retries - force skip
	if !rm.tracker.CanRetry(failure.FeatureID) {
		return rm.strategies[StrategySkip]
	}

	// For rollback strategy, only use it for certain failure types
	if rm.defaultStrategy == StrategyRollback {
		// Rollback is most useful for type check and test failures
		// where reverting might help start fresh
		if failure.Type == FailureTypeTypeCheck || failure.Type == FailureTypeTest {
			return rm.strategies[StrategyRollback]
		}
		// Fall back to retry for other cases
		return rm.strategies[StrategyRetry]
	}

	// Use the configured default strategy
	if strategy, ok := rm.strategies[rm.defaultStrategy]; ok {
		return strategy
	}

	// Default to retry
	return rm.strategies[StrategyRetry]
}

// GetFailureSummary returns a summary of all tracked failures
func (rm *RecoveryManager) GetFailureSummary() string {
	return rm.tracker.GetSummary()
}

// ShouldEscalate determines if we should escalate from retry to skip
func (rm *RecoveryManager) ShouldEscalate(featureID int) bool {
	return !rm.tracker.CanRetry(featureID)
}
