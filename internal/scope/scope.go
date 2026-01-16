// Package scope provides smart scope control for Ralph.
// It manages iteration budgets, time limits, and feature deferral
// to prevent over-building and ensure timely completion.
package scope

import (
	"fmt"
	"strings"
	"time"
)

// Complexity represents the estimated complexity of a feature
type Complexity string

const (
	// ComplexityLow indicates a simple feature (1-2 steps, short description)
	ComplexityLow Complexity = "low"
	// ComplexityMedium indicates a moderate feature (3-5 steps)
	ComplexityMedium Complexity = "medium"
	// ComplexityHigh indicates a complex feature (6+ steps)
	ComplexityHigh Complexity = "high"
)

// DeferReason describes why a feature was deferred
type DeferReason string

const (
	// DeferReasonIterationLimit indicates the feature exceeded its iteration budget
	DeferReasonIterationLimit DeferReason = "iteration_limit"
	// DeferReasonDeadline indicates the deadline was reached
	DeferReasonDeadline DeferReason = "deadline"
	// DeferReasonComplexity indicates the feature was deemed too complex for current scope
	DeferReasonComplexity DeferReason = "complexity"
	// DeferReasonManual indicates the feature was manually deferred
	DeferReasonManual DeferReason = "manual"
)

// Constraints defines the scope limits for execution
type Constraints struct {
	// MaxIterationsPerFeature is the maximum iterations allowed per feature (0 = unlimited)
	MaxIterationsPerFeature int
	// Deadline is the optional time limit for the entire run
	Deadline time.Time
	// QualityThreshold is the minimum test pass rate required (0-100, 0 = no requirement)
	QualityThreshold int
	// AutoDefer controls whether features are automatically deferred when limits are hit
	AutoDefer bool
}

// DefaultConstraints returns default scope constraints
func DefaultConstraints() *Constraints {
	return &Constraints{
		MaxIterationsPerFeature: 0, // unlimited by default
		QualityThreshold:        0, // no requirement by default
		AutoDefer:               true,
	}
}

// FeatureScope tracks the scope status for a single feature
type FeatureScope struct {
	FeatureID         int
	IterationsUsed    int
	StartTime         time.Time
	EndTime           time.Time
	EstimatedComplexity Complexity
	Deferred          bool
	DeferReason       DeferReason
	SimplificationSuggested bool
}

// Manager manages scope constraints and tracking for a run
type Manager struct {
	constraints  *Constraints
	startTime    time.Time
	featureScope map[int]*FeatureScope
	totalIterations int
	deferredFeatures []int
}

// NewManager creates a new scope manager with the given constraints
func NewManager(constraints *Constraints) *Manager {
	if constraints == nil {
		constraints = DefaultConstraints()
	}
	return &Manager{
		constraints:  constraints,
		startTime:    time.Now(),
		featureScope: make(map[int]*FeatureScope),
	}
}

// SetDeadline sets the deadline for the run
func (m *Manager) SetDeadline(deadline time.Time) {
	m.constraints.Deadline = deadline
}

// SetDeadlineDuration sets the deadline as a duration from now
func (m *Manager) SetDeadlineDuration(d time.Duration) {
	m.constraints.Deadline = time.Now().Add(d)
}

// GetConstraints returns the current scope constraints
func (m *Manager) GetConstraints() *Constraints {
	return m.constraints
}

// StartFeature begins scope tracking for a feature
func (m *Manager) StartFeature(featureID int, stepCount int, description string) *FeatureScope {
	scope := &FeatureScope{
		FeatureID:         featureID,
		StartTime:         time.Now(),
		EstimatedComplexity: EstimateComplexity(stepCount, description),
	}
	m.featureScope[featureID] = scope
	return scope
}

// RecordIteration records an iteration for a feature
func (m *Manager) RecordIteration(featureID int) {
	m.totalIterations++
	if scope, ok := m.featureScope[featureID]; ok {
		scope.IterationsUsed++
	}
}

// GetFeatureScope returns the scope status for a feature
func (m *Manager) GetFeatureScope(featureID int) *FeatureScope {
	return m.featureScope[featureID]
}

// ShouldDefer checks if a feature should be deferred based on scope constraints
func (m *Manager) ShouldDefer(featureID int) (bool, DeferReason) {
	scope := m.featureScope[featureID]
	if scope == nil {
		return false, ""
	}

	// Check iteration limit
	if m.constraints.MaxIterationsPerFeature > 0 {
		if scope.IterationsUsed >= m.constraints.MaxIterationsPerFeature {
			return true, DeferReasonIterationLimit
		}
	}

	// Check deadline
	if !m.constraints.Deadline.IsZero() {
		if time.Now().After(m.constraints.Deadline) {
			return true, DeferReasonDeadline
		}
	}

	return false, ""
}

// DeferFeature marks a feature as deferred
func (m *Manager) DeferFeature(featureID int, reason DeferReason) {
	if scope, ok := m.featureScope[featureID]; ok {
		scope.Deferred = true
		scope.DeferReason = reason
		scope.EndTime = time.Now()
	}
	m.deferredFeatures = append(m.deferredFeatures, featureID)
}

// CompleteFeature marks a feature as complete
func (m *Manager) CompleteFeature(featureID int) {
	if scope, ok := m.featureScope[featureID]; ok {
		scope.EndTime = time.Now()
	}
}

// GetDeferredFeatures returns the list of deferred feature IDs
func (m *Manager) GetDeferredFeatures() []int {
	return m.deferredFeatures
}

// RemainingTime returns the time remaining until deadline, or zero if no deadline
func (m *Manager) RemainingTime() time.Duration {
	if m.constraints.Deadline.IsZero() {
		return 0
	}
	remaining := time.Until(m.constraints.Deadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsDeadlineExceeded checks if the deadline has passed
func (m *Manager) IsDeadlineExceeded() bool {
	if m.constraints.Deadline.IsZero() {
		return false
	}
	return time.Now().After(m.constraints.Deadline)
}

// RemainingIterations returns remaining iterations for a feature, or -1 if unlimited
func (m *Manager) RemainingIterations(featureID int) int {
	if m.constraints.MaxIterationsPerFeature <= 0 {
		return -1
	}
	scope := m.featureScope[featureID]
	if scope == nil {
		return m.constraints.MaxIterationsPerFeature
	}
	remaining := m.constraints.MaxIterationsPerFeature - scope.IterationsUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetTotalIterations returns the total iterations run so far
func (m *Manager) GetTotalIterations() int {
	return m.totalIterations
}

// GetElapsedTime returns the elapsed time since the run started
func (m *Manager) GetElapsedTime() time.Duration {
	return time.Since(m.startTime)
}

// EstimateComplexity estimates the complexity of a feature based on its steps and description
func EstimateComplexity(stepCount int, description string) Complexity {
	// Base complexity from step count
	var complexity Complexity
	switch {
	case stepCount <= 2:
		complexity = ComplexityLow
	case stepCount <= 5:
		complexity = ComplexityMedium
	default:
		complexity = ComplexityHigh
	}

	// Adjust based on description keywords
	descLower := strings.ToLower(description)
	complexIndicators := []string{
		"refactor", "migrate", "integration", "comprehensive",
		"multi", "parallel", "concurrent", "distributed",
		"security", "authentication", "authorization",
	}

	for _, indicator := range complexIndicators {
		if strings.Contains(descLower, indicator) {
			// Bump up complexity
			if complexity == ComplexityLow {
				complexity = ComplexityMedium
			} else if complexity == ComplexityMedium {
				complexity = ComplexityHigh
			}
			break
		}
	}

	return complexity
}

// ComplexityToIterations returns suggested max iterations for a complexity level
func ComplexityToIterations(complexity Complexity) int {
	switch complexity {
	case ComplexityLow:
		return 3
	case ComplexityMedium:
		return 5
	case ComplexityHigh:
		return 10
	default:
		return 5
	}
}

// SuggestSimplification returns simplification suggestions for a feature
func SuggestSimplification(stepCount int, description string) []string {
	suggestions := []string{}

	if stepCount > 5 {
		suggestions = append(suggestions,
			fmt.Sprintf("Feature has %d steps - consider breaking into smaller features", stepCount))
	}

	descLower := strings.ToLower(description)

	if strings.Contains(descLower, " and ") {
		suggestions = append(suggestions,
			"Description contains 'and' - may indicate multiple features that could be split")
	}

	if strings.Contains(descLower, "comprehensive") || strings.Contains(descLower, "complete") {
		suggestions = append(suggestions,
			"Consider implementing a minimal version first, then enhancing")
	}

	if strings.Contains(descLower, "all ") {
		suggestions = append(suggestions,
			"'All' may be ambitious - consider implementing a subset first")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions,
			"Focus on core functionality, defer edge cases")
	}

	return suggestions
}

// Status represents the current scope status for display
type Status struct {
	TotalIterations     int
	ElapsedTime         time.Duration
	RemainingTime       time.Duration
	DeadlineSet         bool
	DeadlineExceeded    bool
	DeferredCount       int
	DeferredFeatureIDs  []int
	IterationsPerFeature map[int]int
	MaxIterationsPerFeature int
}

// GetStatus returns the current scope status
func (m *Manager) GetStatus() *Status {
	iterationsPerFeature := make(map[int]int)
	for id, scope := range m.featureScope {
		iterationsPerFeature[id] = scope.IterationsUsed
	}

	return &Status{
		TotalIterations:       m.totalIterations,
		ElapsedTime:           m.GetElapsedTime(),
		RemainingTime:         m.RemainingTime(),
		DeadlineSet:           !m.constraints.Deadline.IsZero(),
		DeadlineExceeded:      m.IsDeadlineExceeded(),
		DeferredCount:         len(m.deferredFeatures),
		DeferredFeatureIDs:    m.deferredFeatures,
		IterationsPerFeature:  iterationsPerFeature,
		MaxIterationsPerFeature: m.constraints.MaxIterationsPerFeature,
	}
}

// FormatStatus returns a formatted string of the scope status
func (m *Manager) FormatStatus() string {
	status := m.GetStatus()
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Elapsed time: %s\n", status.ElapsedTime.Round(time.Second)))

	if status.DeadlineSet {
		if status.DeadlineExceeded {
			sb.WriteString("Deadline: EXCEEDED\n")
		} else {
			sb.WriteString(fmt.Sprintf("Time remaining: %s\n", status.RemainingTime.Round(time.Second)))
		}
	}

	if status.MaxIterationsPerFeature > 0 {
		sb.WriteString(fmt.Sprintf("Max iterations per feature: %d\n", status.MaxIterationsPerFeature))
	}

	if status.DeferredCount > 0 {
		sb.WriteString(fmt.Sprintf("Deferred features: %d (IDs: %v)\n", status.DeferredCount, status.DeferredFeatureIDs))
	}

	return sb.String()
}

// DeferralInfo contains information about a deferred feature
type DeferralInfo struct {
	FeatureID     int
	Reason        DeferReason
	IterationsUsed int
	Suggestions   []string
}

// GetDeferralInfo returns detailed information about deferred features
func (m *Manager) GetDeferralInfo() []DeferralInfo {
	var info []DeferralInfo
	for _, id := range m.deferredFeatures {
		scope := m.featureScope[id]
		if scope != nil {
			info = append(info, DeferralInfo{
				FeatureID:      id,
				Reason:         scope.DeferReason,
				IterationsUsed: scope.IterationsUsed,
			})
		}
	}
	return info
}

// FormatDeferralReason returns a human-readable string for the deferral reason
func FormatDeferralReason(reason DeferReason) string {
	switch reason {
	case DeferReasonIterationLimit:
		return "exceeded iteration limit"
	case DeferReasonDeadline:
		return "deadline reached"
	case DeferReasonComplexity:
		return "too complex for current scope"
	case DeferReasonManual:
		return "manually deferred"
	default:
		return string(reason)
	}
}

// ShouldSuggestSimplification checks if simplification should be suggested
func (m *Manager) ShouldSuggestSimplification(featureID int) bool {
	scope := m.featureScope[featureID]
	if scope == nil {
		return false
	}

	// Suggest simplification if:
	// 1. Feature is high complexity
	// 2. Or iterations used is >= 50% of limit
	if scope.EstimatedComplexity == ComplexityHigh {
		return true
	}

	if m.constraints.MaxIterationsPerFeature > 0 {
		halfLimit := m.constraints.MaxIterationsPerFeature / 2
		if halfLimit == 0 {
			halfLimit = 1
		}
		if scope.IterationsUsed >= halfLimit {
			return true
		}
	}

	return false
}

// MarkSimplificationSuggested marks that simplification was suggested for a feature
func (m *Manager) MarkSimplificationSuggested(featureID int) {
	if scope, ok := m.featureScope[featureID]; ok {
		scope.SimplificationSuggested = true
	}
}

// WasSimplificationSuggested checks if simplification was already suggested
func (m *Manager) WasSimplificationSuggested(featureID int) bool {
	if scope, ok := m.featureScope[featureID]; ok {
		return scope.SimplificationSuggested
	}
	return false
}
