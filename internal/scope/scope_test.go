package scope

import (
	"testing"
	"time"
)

func TestDefaultConstraints(t *testing.T) {
	c := DefaultConstraints()

	if c.MaxIterationsPerFeature != 0 {
		t.Errorf("expected MaxIterationsPerFeature=0, got %d", c.MaxIterationsPerFeature)
	}
	if c.QualityThreshold != 0 {
		t.Errorf("expected QualityThreshold=0, got %d", c.QualityThreshold)
	}
	if !c.AutoDefer {
		t.Error("expected AutoDefer=true")
	}
}

func TestNewManager(t *testing.T) {
	t.Run("with nil constraints", func(t *testing.T) {
		m := NewManager(nil)
		if m.constraints == nil {
			t.Error("expected default constraints, got nil")
		}
		if m.constraints.MaxIterationsPerFeature != 0 {
			t.Errorf("expected MaxIterationsPerFeature=0, got %d", m.constraints.MaxIterationsPerFeature)
		}
	})

	t.Run("with custom constraints", func(t *testing.T) {
		c := &Constraints{MaxIterationsPerFeature: 5}
		m := NewManager(c)
		if m.constraints.MaxIterationsPerFeature != 5 {
			t.Errorf("expected MaxIterationsPerFeature=5, got %d", m.constraints.MaxIterationsPerFeature)
		}
	})
}

func TestSetDeadline(t *testing.T) {
	m := NewManager(nil)
	deadline := time.Now().Add(1 * time.Hour)

	m.SetDeadline(deadline)

	if !m.constraints.Deadline.Equal(deadline) {
		t.Errorf("expected deadline %v, got %v", deadline, m.constraints.Deadline)
	}
}

func TestSetDeadlineDuration(t *testing.T) {
	m := NewManager(nil)
	before := time.Now()

	m.SetDeadlineDuration(1 * time.Hour)

	after := time.Now()
	expected := before.Add(1 * time.Hour)

	if m.constraints.Deadline.Before(expected.Add(-1*time.Second)) || m.constraints.Deadline.After(after.Add(1*time.Hour+time.Second)) {
		t.Errorf("deadline %v not within expected range", m.constraints.Deadline)
	}
}

func TestStartFeature(t *testing.T) {
	m := NewManager(nil)
	scope := m.StartFeature(1, 3, "Test feature")

	if scope.FeatureID != 1 {
		t.Errorf("expected FeatureID=1, got %d", scope.FeatureID)
	}
	if scope.IterationsUsed != 0 {
		t.Errorf("expected IterationsUsed=0, got %d", scope.IterationsUsed)
	}
	if scope.EstimatedComplexity != ComplexityMedium {
		t.Errorf("expected complexity=medium for 3 steps, got %s", scope.EstimatedComplexity)
	}

	// Verify scope is stored in manager
	storedScope := m.GetFeatureScope(1)
	if storedScope == nil {
		t.Error("expected scope to be stored in manager")
	}
}

func TestRecordIteration(t *testing.T) {
	m := NewManager(nil)
	m.StartFeature(1, 3, "Test feature")

	m.RecordIteration(1)
	m.RecordIteration(1)

	scope := m.GetFeatureScope(1)
	if scope.IterationsUsed != 2 {
		t.Errorf("expected IterationsUsed=2, got %d", scope.IterationsUsed)
	}
	if m.GetTotalIterations() != 2 {
		t.Errorf("expected total iterations=2, got %d", m.GetTotalIterations())
	}
}

func TestShouldDefer_IterationLimit(t *testing.T) {
	c := &Constraints{MaxIterationsPerFeature: 3}
	m := NewManager(c)
	m.StartFeature(1, 3, "Test feature")

	// Should not defer initially
	shouldDefer, reason := m.ShouldDefer(1)
	if shouldDefer {
		t.Error("should not defer before limit reached")
	}

	// Record iterations up to limit
	m.RecordIteration(1)
	m.RecordIteration(1)
	shouldDefer, reason = m.ShouldDefer(1)
	if shouldDefer {
		t.Error("should not defer at 2/3 iterations")
	}

	// Hit the limit
	m.RecordIteration(1)
	shouldDefer, reason = m.ShouldDefer(1)
	if !shouldDefer {
		t.Error("should defer after limit reached")
	}
	if reason != DeferReasonIterationLimit {
		t.Errorf("expected reason=%s, got %s", DeferReasonIterationLimit, reason)
	}
}

func TestShouldDefer_Deadline(t *testing.T) {
	m := NewManager(nil)
	m.StartFeature(1, 3, "Test feature")

	// Set deadline in the past
	m.SetDeadline(time.Now().Add(-1 * time.Hour))

	shouldDefer, reason := m.ShouldDefer(1)
	if !shouldDefer {
		t.Error("should defer when deadline exceeded")
	}
	if reason != DeferReasonDeadline {
		t.Errorf("expected reason=%s, got %s", DeferReasonDeadline, reason)
	}
}

func TestShouldDefer_NoLimit(t *testing.T) {
	m := NewManager(nil) // No limits set
	m.StartFeature(1, 3, "Test feature")

	// Record many iterations
	for i := 0; i < 100; i++ {
		m.RecordIteration(1)
	}

	shouldDefer, _ := m.ShouldDefer(1)
	if shouldDefer {
		t.Error("should never defer when no limits set")
	}
}

func TestDeferFeature(t *testing.T) {
	m := NewManager(nil)
	m.StartFeature(1, 3, "Test feature")

	m.DeferFeature(1, DeferReasonIterationLimit)

	scope := m.GetFeatureScope(1)
	if !scope.Deferred {
		t.Error("expected feature to be marked as deferred")
	}
	if scope.DeferReason != DeferReasonIterationLimit {
		t.Errorf("expected reason=%s, got %s", DeferReasonIterationLimit, scope.DeferReason)
	}

	deferred := m.GetDeferredFeatures()
	if len(deferred) != 1 || deferred[0] != 1 {
		t.Errorf("expected deferred=[1], got %v", deferred)
	}
}

func TestCompleteFeature(t *testing.T) {
	m := NewManager(nil)
	m.StartFeature(1, 3, "Test feature")

	m.CompleteFeature(1)

	scope := m.GetFeatureScope(1)
	if scope.EndTime.IsZero() {
		t.Error("expected EndTime to be set")
	}
}

func TestRemainingTime(t *testing.T) {
	t.Run("no deadline", func(t *testing.T) {
		m := NewManager(nil)
		if m.RemainingTime() != 0 {
			t.Error("expected 0 remaining time when no deadline")
		}
	})

	t.Run("with deadline in future", func(t *testing.T) {
		m := NewManager(nil)
		m.SetDeadlineDuration(1 * time.Hour)

		remaining := m.RemainingTime()
		if remaining < 59*time.Minute || remaining > 61*time.Minute {
			t.Errorf("expected ~1 hour remaining, got %v", remaining)
		}
	})

	t.Run("with deadline in past", func(t *testing.T) {
		m := NewManager(nil)
		m.SetDeadline(time.Now().Add(-1 * time.Hour))

		if m.RemainingTime() != 0 {
			t.Error("expected 0 remaining time when deadline passed")
		}
	})
}

func TestIsDeadlineExceeded(t *testing.T) {
	t.Run("no deadline", func(t *testing.T) {
		m := NewManager(nil)
		if m.IsDeadlineExceeded() {
			t.Error("no deadline should not be exceeded")
		}
	})

	t.Run("deadline in past", func(t *testing.T) {
		m := NewManager(nil)
		m.SetDeadline(time.Now().Add(-1 * time.Hour))
		if !m.IsDeadlineExceeded() {
			t.Error("past deadline should be exceeded")
		}
	})

	t.Run("deadline in future", func(t *testing.T) {
		m := NewManager(nil)
		m.SetDeadline(time.Now().Add(1 * time.Hour))
		if m.IsDeadlineExceeded() {
			t.Error("future deadline should not be exceeded")
		}
	})
}

func TestRemainingIterations(t *testing.T) {
	t.Run("no limit", func(t *testing.T) {
		m := NewManager(nil)
		m.StartFeature(1, 3, "Test")

		if m.RemainingIterations(1) != -1 {
			t.Error("expected -1 for unlimited")
		}
	})

	t.Run("with limit", func(t *testing.T) {
		c := &Constraints{MaxIterationsPerFeature: 5}
		m := NewManager(c)
		m.StartFeature(1, 3, "Test")

		if m.RemainingIterations(1) != 5 {
			t.Errorf("expected 5 remaining, got %d", m.RemainingIterations(1))
		}

		m.RecordIteration(1)
		m.RecordIteration(1)

		if m.RemainingIterations(1) != 3 {
			t.Errorf("expected 3 remaining, got %d", m.RemainingIterations(1))
		}
	})

	t.Run("unknown feature with limit", func(t *testing.T) {
		c := &Constraints{MaxIterationsPerFeature: 5}
		m := NewManager(c)

		if m.RemainingIterations(99) != 5 {
			t.Errorf("expected max limit for unknown feature, got %d", m.RemainingIterations(99))
		}
	})
}

func TestEstimateComplexity(t *testing.T) {
	tests := []struct {
		name        string
		stepCount   int
		description string
		expected    Complexity
	}{
		{"low complexity - few steps", 2, "Simple task", ComplexityLow},
		{"low complexity - one step", 1, "Quick fix", ComplexityLow},
		{"medium complexity - moderate steps", 4, "Add feature", ComplexityMedium},
		{"high complexity - many steps", 7, "Complex work", ComplexityHigh},
		{"bumped up - refactor keyword", 2, "Refactor the codebase", ComplexityMedium},
		{"bumped up - integration keyword", 2, "Integration testing", ComplexityMedium},
		{"bumped up - security keyword", 3, "Add security features", ComplexityHigh},
		{"bumped up - comprehensive keyword", 3, "Comprehensive coverage", ComplexityHigh},
		{"bumped up - multi keyword", 2, "Multi-tenant support", ComplexityMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateComplexity(tt.stepCount, tt.description)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestComplexityToIterations(t *testing.T) {
	tests := []struct {
		complexity Complexity
		expected   int
	}{
		{ComplexityLow, 3},
		{ComplexityMedium, 5},
		{ComplexityHigh, 10},
		{Complexity("unknown"), 5},
	}

	for _, tt := range tests {
		t.Run(string(tt.complexity), func(t *testing.T) {
			result := ComplexityToIterations(tt.complexity)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestSuggestSimplification(t *testing.T) {
	t.Run("many steps", func(t *testing.T) {
		suggestions := SuggestSimplification(8, "Simple task")
		found := false
		for _, s := range suggestions {
			if contains(s, "8 steps") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected suggestion about step count")
		}
	})

	t.Run("contains 'and'", func(t *testing.T) {
		suggestions := SuggestSimplification(2, "Add login and registration")
		found := false
		for _, s := range suggestions {
			if contains(s, "'and'") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected suggestion about 'and' in description")
		}
	})

	t.Run("contains 'comprehensive'", func(t *testing.T) {
		suggestions := SuggestSimplification(2, "Comprehensive testing")
		found := false
		for _, s := range suggestions {
			if contains(s, "minimal version") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected suggestion about minimal version")
		}
	})

	t.Run("contains 'all'", func(t *testing.T) {
		suggestions := SuggestSimplification(2, "Fix all bugs")
		found := false
		for _, s := range suggestions {
			if contains(s, "subset") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected suggestion about subset")
		}
	})

	t.Run("default suggestion", func(t *testing.T) {
		suggestions := SuggestSimplification(2, "Simple task")
		if len(suggestions) == 0 {
			t.Error("expected at least one suggestion")
		}
	})
}

func TestGetStatus(t *testing.T) {
	c := &Constraints{MaxIterationsPerFeature: 5}
	m := NewManager(c)
	m.SetDeadlineDuration(1 * time.Hour)
	m.StartFeature(1, 3, "Test")
	m.RecordIteration(1)
	m.RecordIteration(1)
	m.DeferFeature(2, DeferReasonManual)

	status := m.GetStatus()

	if status.TotalIterations != 2 {
		t.Errorf("expected TotalIterations=2, got %d", status.TotalIterations)
	}
	if !status.DeadlineSet {
		t.Error("expected DeadlineSet=true")
	}
	if status.DeadlineExceeded {
		t.Error("expected DeadlineExceeded=false")
	}
	if status.DeferredCount != 1 {
		t.Errorf("expected DeferredCount=1, got %d", status.DeferredCount)
	}
	if status.MaxIterationsPerFeature != 5 {
		t.Errorf("expected MaxIterationsPerFeature=5, got %d", status.MaxIterationsPerFeature)
	}
	if status.IterationsPerFeature[1] != 2 {
		t.Errorf("expected iterations for feature 1=2, got %d", status.IterationsPerFeature[1])
	}
}

func TestFormatStatus(t *testing.T) {
	c := &Constraints{MaxIterationsPerFeature: 5}
	m := NewManager(c)
	m.SetDeadlineDuration(1 * time.Hour)
	m.StartFeature(1, 3, "Test")
	m.DeferFeature(1, DeferReasonIterationLimit)

	output := m.FormatStatus()

	if !contains(output, "Elapsed time") {
		t.Error("expected 'Elapsed time' in output")
	}
	if !contains(output, "Max iterations per feature: 5") {
		t.Error("expected max iterations info in output")
	}
	if !contains(output, "Deferred features: 1") {
		t.Error("expected deferred features info in output")
	}
}

func TestFormatDeferralReason(t *testing.T) {
	tests := []struct {
		reason   DeferReason
		expected string
	}{
		{DeferReasonIterationLimit, "exceeded iteration limit"},
		{DeferReasonDeadline, "deadline reached"},
		{DeferReasonComplexity, "too complex for current scope"},
		{DeferReasonManual, "manually deferred"},
		{DeferReason("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(string(tt.reason), func(t *testing.T) {
			result := FormatDeferralReason(tt.reason)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestShouldSuggestSimplification(t *testing.T) {
	t.Run("high complexity", func(t *testing.T) {
		m := NewManager(nil)
		m.StartFeature(1, 10, "Complex security refactoring")

		if !m.ShouldSuggestSimplification(1) {
			t.Error("should suggest simplification for high complexity")
		}
	})

	t.Run("at half iteration limit", func(t *testing.T) {
		c := &Constraints{MaxIterationsPerFeature: 6}
		m := NewManager(c)
		m.StartFeature(1, 2, "Simple task") // Low complexity

		// At 2 iterations (< half of 6), should not suggest
		m.RecordIteration(1)
		m.RecordIteration(1)
		if m.ShouldSuggestSimplification(1) {
			t.Error("should not suggest at 2/6 iterations")
		}

		// At 3 iterations (= half of 6), should suggest
		m.RecordIteration(1)
		if !m.ShouldSuggestSimplification(1) {
			t.Error("should suggest at half limit")
		}
	})

	t.Run("unknown feature", func(t *testing.T) {
		m := NewManager(nil)
		if m.ShouldSuggestSimplification(99) {
			t.Error("should not suggest for unknown feature")
		}
	})
}

func TestSimplificationSuggestedTracking(t *testing.T) {
	m := NewManager(nil)
	m.StartFeature(1, 3, "Test")

	if m.WasSimplificationSuggested(1) {
		t.Error("should not be suggested initially")
	}

	m.MarkSimplificationSuggested(1)

	if !m.WasSimplificationSuggested(1) {
		t.Error("should be marked as suggested")
	}
}

func TestGetDeferralInfo(t *testing.T) {
	m := NewManager(nil)
	m.StartFeature(1, 3, "Test 1")
	m.StartFeature(2, 5, "Test 2")
	m.RecordIteration(1)
	m.RecordIteration(1)

	m.DeferFeature(1, DeferReasonIterationLimit)
	m.DeferFeature(2, DeferReasonDeadline)

	info := m.GetDeferralInfo()
	if len(info) != 2 {
		t.Errorf("expected 2 deferral infos, got %d", len(info))
	}

	// Find feature 1
	var info1 *DeferralInfo
	for i := range info {
		if info[i].FeatureID == 1 {
			info1 = &info[i]
			break
		}
	}

	if info1 == nil {
		t.Fatal("could not find deferral info for feature 1")
	}

	if info1.Reason != DeferReasonIterationLimit {
		t.Errorf("expected reason=%s, got %s", DeferReasonIterationLimit, info1.Reason)
	}
	if info1.IterationsUsed != 2 {
		t.Errorf("expected IterationsUsed=2, got %d", info1.IterationsUsed)
	}
}

// Helper function for string containment check
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
