// Package goals provides high-level goal management and automatic plan decomposition for Ralph.
package goals

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/logimos/ralph/internal/plan"
)

// GoalStatus represents the current status of a goal
type GoalStatus string

const (
	// StatusPending means the goal hasn't been started yet
	StatusPending GoalStatus = "pending"
	// StatusInProgress means work on the goal has started
	StatusInProgress GoalStatus = "in_progress"
	// StatusComplete means all generated plan items are complete
	StatusComplete GoalStatus = "complete"
	// StatusBlocked means the goal is blocked and cannot progress
	StatusBlocked GoalStatus = "blocked"
)

// Goal represents a high-level project goal that can be decomposed into plan items
type Goal struct {
	ID              string            `json:"id"`                          // Unique identifier for the goal
	Description     string            `json:"description"`                 // High-level goal description
	SuccessCriteria []string          `json:"success_criteria,omitempty"`  // What success looks like
	Priority        int               `json:"priority,omitempty"`          // Priority for ordering (higher = more important)
	Category        string            `json:"category,omitempty"`          // Category for grouping (e.g., "feature", "infrastructure")
	Tags            []string          `json:"tags,omitempty"`              // Tags for filtering and organization
	Dependencies    []string          `json:"dependencies,omitempty"`      // IDs of goals this depends on
	GeneratedPlanIDs []int            `json:"generated_plan_ids,omitempty"` // IDs of plan items generated from this goal
	Metadata        map[string]string `json:"metadata,omitempty"`          // Additional metadata
	Status          GoalStatus        `json:"status,omitempty"`            // Current goal status
	CreatedAt       time.Time         `json:"created_at,omitempty"`        // When the goal was created
	UpdatedAt       time.Time         `json:"updated_at,omitempty"`        // When the goal was last updated
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`      // When the goal was completed (if complete)
}

// GoalFile represents the structure of a goals.json file
type GoalFile struct {
	Goals       []Goal    `json:"goals"`
	LastUpdated time.Time `json:"last_updated,omitempty"`
	Version     string    `json:"version,omitempty"` // File format version
}

// GoalProgress represents the progress of a goal toward completion
type GoalProgress struct {
	Goal              *Goal
	TotalPlanItems    int
	CompletedItems    int
	DeferredItems     int
	RemainingItems    int
	PercentComplete   float64
	Status            GoalStatus
	BlockedByGoals    []string // Goal IDs that are blocking this goal
	EstimatedRemaining int     // Estimated remaining iterations (based on steps)
}

// Manager manages goals and their relationship to plan items
type Manager struct {
	goals     []Goal
	plans     []plan.Plan
	goalsFile string
}

// NewManager creates a new goal manager
func NewManager(plans []plan.Plan) *Manager {
	return &Manager{
		goals: []Goal{},
		plans: plans,
	}
}

// SetGoalsFile sets the path to the goals file
func (m *Manager) SetGoalsFile(path string) {
	m.goalsFile = path
}

// LoadGoals loads goals from a file
func (m *Manager) LoadGoals(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No goals file is fine - just use empty list
			m.goals = []Goal{}
			return nil
		}
		return fmt.Errorf("failed to read goals file: %w", err)
	}

	var goalFile GoalFile
	if err := json.Unmarshal(data, &goalFile); err != nil {
		// Try parsing as array of goals directly (simpler format)
		var goals []Goal
		if err2 := json.Unmarshal(data, &goals); err2 != nil {
			return fmt.Errorf("failed to parse goals file: %w", err)
		}
		m.goals = goals
	} else {
		m.goals = goalFile.Goals
	}

	m.goalsFile = path
	return nil
}

// SaveGoals saves goals to the configured file
func (m *Manager) SaveGoals() error {
	if m.goalsFile == "" {
		return fmt.Errorf("no goals file configured")
	}
	return m.SaveGoalsTo(m.goalsFile)
}

// SaveGoalsTo saves goals to a specific file
func (m *Manager) SaveGoalsTo(path string) error {
	goalFile := GoalFile{
		Goals:       m.goals,
		LastUpdated: time.Now(),
		Version:     "1.0",
	}

	data, err := json.MarshalIndent(goalFile, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal goals: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write goals file: %w", err)
	}

	return nil
}

// GetGoals returns all goals
func (m *Manager) GetGoals() []Goal {
	return m.goals
}

// GetGoalByID returns a goal by its ID
func (m *Manager) GetGoalByID(id string) *Goal {
	for i := range m.goals {
		if m.goals[i].ID == id {
			return &m.goals[i]
		}
	}
	return nil
}

// AddGoal adds a new goal
func (m *Manager) AddGoal(goal Goal) error {
	// Validate goal
	if goal.Description == "" {
		return fmt.Errorf("goal description cannot be empty")
	}

	// Generate ID if not provided
	if goal.ID == "" {
		goal.ID = generateGoalID()
	}

	// Check for duplicate ID
	if m.GetGoalByID(goal.ID) != nil {
		return fmt.Errorf("goal with ID %q already exists", goal.ID)
	}

	// Set timestamps
	now := time.Now()
	goal.CreatedAt = now
	goal.UpdatedAt = now

	// Set default status
	if goal.Status == "" {
		goal.Status = StatusPending
	}

	m.goals = append(m.goals, goal)
	return nil
}

// AddGoalFromDescription creates a goal from a simple description string
func (m *Manager) AddGoalFromDescription(description string, priority int) (*Goal, error) {
	goal := Goal{
		ID:          generateGoalID(),
		Description: description,
		Priority:    priority,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Try to infer category from description
	goal.Category = inferCategory(description)

	if err := m.AddGoal(goal); err != nil {
		return nil, err
	}

	return &goal, nil
}

// RemoveGoal removes a goal by ID
func (m *Manager) RemoveGoal(id string) bool {
	for i, g := range m.goals {
		if g.ID == id {
			m.goals = append(m.goals[:i], m.goals[i+1:]...)
			return true
		}
	}
	return false
}

// UpdateGoal updates an existing goal
func (m *Manager) UpdateGoal(goal Goal) error {
	for i, g := range m.goals {
		if g.ID == goal.ID {
			goal.UpdatedAt = time.Now()
			m.goals[i] = goal
			return nil
		}
	}
	return fmt.Errorf("goal with ID %q not found", goal.ID)
}

// SetPlans updates the plan list used for progress calculation
func (m *Manager) SetPlans(plans []plan.Plan) {
	m.plans = plans
}

// CalculateProgress calculates the progress for a specific goal
func (m *Manager) CalculateProgress(goalID string) *GoalProgress {
	goal := m.GetGoalByID(goalID)
	if goal == nil {
		return nil
	}

	progress := &GoalProgress{
		Goal:   goal,
		Status: goal.Status,
	}

	// Check dependencies
	progress.BlockedByGoals = m.getBlockingGoals(goal)
	if len(progress.BlockedByGoals) > 0 && goal.Status != StatusComplete {
		progress.Status = StatusBlocked
	}

	// If no generated plan IDs, check by plan's goal field (if we add that later)
	if len(goal.GeneratedPlanIDs) == 0 {
		return progress
	}

	// Calculate progress from generated plan items
	for _, planID := range goal.GeneratedPlanIDs {
		p := plan.GetByID(m.plans, planID)
		if p == nil {
			continue
		}

		progress.TotalPlanItems++
		if p.Tested {
			progress.CompletedItems++
		} else if p.Deferred {
			progress.DeferredItems++
		} else {
			progress.RemainingItems++
			progress.EstimatedRemaining += len(p.Steps)
		}
	}

	// Calculate percentage
	if progress.TotalPlanItems > 0 {
		progress.PercentComplete = float64(progress.CompletedItems) / float64(progress.TotalPlanItems) * 100
	}

	// Update status based on progress
	if progress.TotalPlanItems > 0 {
		if progress.CompletedItems == progress.TotalPlanItems {
			progress.Status = StatusComplete
		} else if progress.CompletedItems > 0 || progress.Status == StatusPending {
			progress.Status = StatusInProgress
		}
	}

	return progress
}

// CalculateAllProgress calculates progress for all goals
func (m *Manager) CalculateAllProgress() []*GoalProgress {
	var results []*GoalProgress
	for _, goal := range m.goals {
		progress := m.CalculateProgress(goal.ID)
		if progress != nil {
			results = append(results, progress)
		}
	}
	return results
}

// GetPendingGoals returns goals that haven't been started
func (m *Manager) GetPendingGoals() []Goal {
	var pending []Goal
	for _, g := range m.goals {
		if g.Status == StatusPending {
			pending = append(pending, g)
		}
	}
	return pending
}

// GetActiveGoals returns goals that are in progress
func (m *Manager) GetActiveGoals() []Goal {
	var active []Goal
	for _, g := range m.goals {
		if g.Status == StatusInProgress {
			active = append(active, g)
		}
	}
	return active
}

// GetCompletedGoals returns goals that are complete
func (m *Manager) GetCompletedGoals() []Goal {
	var completed []Goal
	for _, g := range m.goals {
		if g.Status == StatusComplete {
			completed = append(completed, g)
		}
	}
	return completed
}

// GetGoalsByPriority returns goals sorted by priority (highest first)
func (m *Manager) GetGoalsByPriority() []Goal {
	sorted := make([]Goal, len(m.goals))
	copy(sorted, m.goals)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})
	return sorted
}

// GetNextGoalToWork returns the highest priority goal that is ready to work on
func (m *Manager) GetNextGoalToWork() *Goal {
	sorted := m.GetGoalsByPriority()
	for i := range sorted {
		goal := &sorted[i]
		if goal.Status == StatusComplete {
			continue
		}

		// Check if blocked by dependencies
		blocking := m.getBlockingGoals(goal)
		if len(blocking) == 0 {
			return goal
		}
	}
	return nil
}

// LinkPlanToGoal associates a plan item with a goal
func (m *Manager) LinkPlanToGoal(goalID string, planID int) error {
	goal := m.GetGoalByID(goalID)
	if goal == nil {
		return fmt.Errorf("goal %q not found", goalID)
	}

	// Check if already linked
	for _, id := range goal.GeneratedPlanIDs {
		if id == planID {
			return nil // Already linked
		}
	}

	goal.GeneratedPlanIDs = append(goal.GeneratedPlanIDs, planID)
	goal.UpdatedAt = time.Now()

	// Update status if this is the first plan item
	if goal.Status == StatusPending {
		goal.Status = StatusInProgress
	}

	return m.UpdateGoal(*goal)
}

// MarkGoalComplete marks a goal as complete
func (m *Manager) MarkGoalComplete(goalID string) error {
	goal := m.GetGoalByID(goalID)
	if goal == nil {
		return fmt.Errorf("goal %q not found", goalID)
	}

	now := time.Now()
	goal.Status = StatusComplete
	goal.CompletedAt = &now
	goal.UpdatedAt = now

	return m.UpdateGoal(*goal)
}

// Summary returns a formatted summary of all goals
func (m *Manager) Summary() string {
	if len(m.goals) == 0 {
		return "No goals defined."
	}

	var sb strings.Builder
	sb.WriteString("=== Goals Summary ===\n\n")

	// Group by status
	pending := m.GetPendingGoals()
	active := m.GetActiveGoals()
	completed := m.GetCompletedGoals()
	blocked := m.getBlockedGoals()

	if len(active) > 0 {
		sb.WriteString("Active Goals:\n")
		for _, g := range active {
			progress := m.CalculateProgress(g.ID)
			sb.WriteString(formatGoalLine(g, progress))
		}
		sb.WriteString("\n")
	}

	if len(blocked) > 0 {
		sb.WriteString("Blocked Goals:\n")
		for _, g := range blocked {
			progress := m.CalculateProgress(g.ID)
			sb.WriteString(formatGoalLine(g, progress))
			if len(progress.BlockedByGoals) > 0 {
				sb.WriteString(fmt.Sprintf("      Blocked by: %s\n", strings.Join(progress.BlockedByGoals, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	if len(pending) > 0 {
		sb.WriteString("Pending Goals:\n")
		for _, g := range pending {
			sb.WriteString(fmt.Sprintf("  ○ [%d] %s\n", g.Priority, g.Description))
		}
		sb.WriteString("\n")
	}

	if len(completed) > 0 {
		sb.WriteString("Completed Goals:\n")
		for _, g := range completed {
			sb.WriteString(fmt.Sprintf("  ● %s\n", g.Description))
		}
		sb.WriteString("\n")
	}

	// Overall stats
	total := len(m.goals)
	completedCount := len(completed)
	sb.WriteString(fmt.Sprintf("Total: %d goals (%d complete, %d active, %d pending)\n",
		total, completedCount, len(active), len(pending)))

	return sb.String()
}

// HasGoals returns true if there are any goals defined
func (m *Manager) HasGoals() bool {
	return len(m.goals) > 0
}

// Count returns the number of goals
func (m *Manager) Count() int {
	return len(m.goals)
}

// getBlockingGoals returns IDs of goals that block the given goal
func (m *Manager) getBlockingGoals(goal *Goal) []string {
	var blocking []string
	for _, depID := range goal.Dependencies {
		dep := m.GetGoalByID(depID)
		if dep != nil && dep.Status != StatusComplete {
			blocking = append(blocking, depID)
		}
	}
	return blocking
}

// getBlockedGoals returns goals that are blocked by dependencies
func (m *Manager) getBlockedGoals() []Goal {
	var blocked []Goal
	for _, g := range m.goals {
		if g.Status == StatusComplete {
			continue
		}
		if len(m.getBlockingGoals(&g)) > 0 {
			blocked = append(blocked, g)
		}
	}
	return blocked
}

// generateGoalID generates a unique goal ID
func generateGoalID() string {
	return fmt.Sprintf("goal_%d", time.Now().UnixNano())
}

// inferCategory tries to infer a category from the goal description
func inferCategory(description string) string {
	lower := strings.ToLower(description)

	categoryKeywords := map[string][]string{
		"feature":        {"add", "implement", "create", "build", "develop"},
		"infrastructure": {"setup", "configure", "deploy", "infrastructure", "ci", "cd", "docker", "kubernetes"},
		"database":       {"database", "db", "migration", "schema", "sql", "postgres", "mysql", "mongo"},
		"ui":             {"ui", "frontend", "component", "page", "style", "css", "design"},
		"api":            {"api", "endpoint", "rest", "graphql", "service"},
		"security":       {"security", "auth", "authentication", "authorization", "permission"},
		"testing":        {"test", "testing", "coverage", "e2e", "integration test"},
		"performance":    {"performance", "optimize", "speed", "cache", "faster"},
		"documentation":  {"document", "docs", "readme", "guide"},
		"refactor":       {"refactor", "clean", "reorganize", "restructure"},
	}

	for category, keywords := range categoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return category
			}
		}
	}

	return "other"
}

// formatGoalLine formats a single goal line for display
func formatGoalLine(g Goal, progress *GoalProgress) string {
	var sb strings.Builder
	
	// Status indicator
	switch progress.Status {
	case StatusComplete:
		sb.WriteString("  ● ")
	case StatusInProgress:
		sb.WriteString("  ◐ ")
	case StatusBlocked:
		sb.WriteString("  ✕ ")
	default:
		sb.WriteString("  ○ ")
	}

	// Priority and description
	sb.WriteString(fmt.Sprintf("[%d] %s", g.Priority, g.Description))

	// Progress info if available
	if progress.TotalPlanItems > 0 {
		sb.WriteString(fmt.Sprintf(" (%d/%d items, %.0f%%)",
			progress.CompletedItems, progress.TotalPlanItems, progress.PercentComplete))
	}

	sb.WriteString("\n")
	return sb.String()
}

// FormatProgressBar creates a visual progress bar for a goal
func FormatProgressBar(progress *GoalProgress, width int) string {
	if progress.TotalPlanItems == 0 {
		return "[" + strings.Repeat("-", width) + "] 0%"
	}

	filled := int(progress.PercentComplete / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %.0f%%", bar, progress.PercentComplete)
}
