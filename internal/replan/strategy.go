package replan

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/logimos/ralph/internal/plan"
)

// StrategyType represents the type of replanning strategy
type StrategyType string

const (
	// StrategyIncremental adjusts the remaining plan based on completed work
	StrategyIncremental StrategyType = "incremental"
	// StrategyAgentBased uses the AI agent to generate a new plan
	StrategyAgentBased StrategyType = "agent"
	// StrategyNone indicates no replanning should occur
	StrategyNone StrategyType = "none"
)

// ParseStrategyType parses a string into a StrategyType
func ParseStrategyType(s string) (StrategyType, error) {
	switch strings.ToLower(s) {
	case "incremental", "inc":
		return StrategyIncremental, nil
	case "agent", "ai":
		return StrategyAgentBased, nil
	case "none", "off", "":
		return StrategyNone, nil
	default:
		return "", fmt.Errorf("unknown replan strategy: %s (valid: incremental, agent, none)", s)
	}
}

// ReplanResult represents the result of a replanning operation
type ReplanResult struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message"`
	Trigger     TriggerType `json:"trigger"`
	Strategy    StrategyType `json:"strategy"`
	OldPlanPath string      `json:"old_plan_path,omitempty"`
	NewPlans    []plan.Plan `json:"new_plans,omitempty"`
	Diff        *PlanDiff   `json:"diff,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ReplanStrategy defines the interface for replanning strategies
type ReplanStrategy interface {
	// Name returns the strategy name
	Name() StrategyType
	// Execute performs the replanning
	Execute(state *ReplanState, trigger TriggerType) (*ReplanResult, error)
	// Description returns a human-readable description
	Description() string
}

// IncrementalStrategy adjusts the remaining plan based on completed work
type IncrementalStrategy struct{}

// NewIncrementalStrategy creates a new incremental strategy
func NewIncrementalStrategy() *IncrementalStrategy {
	return &IncrementalStrategy{}
}

// Name returns the strategy name
func (s *IncrementalStrategy) Name() StrategyType {
	return StrategyIncremental
}

// Description returns a human-readable description
func (s *IncrementalStrategy) Description() string {
	return "Adjust remaining features based on completed work and current state"
}

// Execute performs the incremental replanning
func (s *IncrementalStrategy) Execute(state *ReplanState, trigger TriggerType) (*ReplanResult, error) {
	if len(state.Plans) == 0 {
		return &ReplanResult{
			Success:   false,
			Message:   "No plans to replan",
			Trigger:   trigger,
			Strategy:  StrategyIncremental,
			Timestamp: time.Now(),
		}, nil
	}

	// Create a copy of plans to modify
	newPlans := make([]plan.Plan, len(state.Plans))
	copy(newPlans, state.Plans)

	var adjustments []string

	// Apply adjustments based on trigger type
	switch trigger {
	case TriggerTestFailure:
		// For test failures, simplify the current feature or break it into smaller steps
		adjustments = s.handleTestFailure(newPlans, state)

	case TriggerBlockedFeature:
		// For blocked features, reorder remaining features or mark alternatives
		adjustments = s.handleBlockedFeature(newPlans, state)

	case TriggerRequirementChange:
		// For requirement changes, validate and reconcile
		adjustments = s.handleRequirementChange(newPlans, state)

	default:
		adjustments = append(adjustments, "General plan validation performed")
	}

	// Calculate diff
	diff := ComputeDiff(state.Plans, newPlans)

	return &ReplanResult{
		Success:   true,
		Message:   fmt.Sprintf("Incremental replan completed: %s", strings.Join(adjustments, "; ")),
		Trigger:   trigger,
		Strategy:  StrategyIncremental,
		NewPlans:  newPlans,
		Diff:      diff,
		Timestamp: time.Now(),
	}, nil
}

// handleTestFailure handles replanning due to test failures
func (s *IncrementalStrategy) handleTestFailure(plans []plan.Plan, state *ReplanState) []string {
	var adjustments []string

	// Find the current (failing) feature
	for i := range plans {
		if plans[i].ID == state.FeatureID && !plans[i].Tested && !plans[i].Deferred {
			// Check if feature has many steps - might need simplification
			if len(plans[i].Steps) > 5 {
				adjustments = append(adjustments, 
					fmt.Sprintf("Feature #%d has %d steps - consider breaking into smaller tasks", 
						plans[i].ID, len(plans[i].Steps)))
			}

			// Add a note about the failure to the feature's description
			if !strings.Contains(plans[i].Description, "[REQUIRES REVIEW]") {
				plans[i].Description = plans[i].Description + " [REQUIRES REVIEW: Multiple test failures]"
				adjustments = append(adjustments, 
					fmt.Sprintf("Marked feature #%d for review", plans[i].ID))
			}
			break
		}
	}

	// Check for features that might be prerequisites
	adjustments = append(adjustments, s.identifyPotentialPrerequisites(plans, state.FeatureID)...)

	return adjustments
}

// handleBlockedFeature handles replanning due to blocked features
func (s *IncrementalStrategy) handleBlockedFeature(plans []plan.Plan, state *ReplanState) []string {
	var adjustments []string

	// Mark blocked features
	for i := range plans {
		for _, blockedID := range state.BlockedFeatures {
			if plans[i].ID == blockedID && !plans[i].Deferred {
				plans[i].Deferred = true
				plans[i].DeferReason = "blocked_during_execution"
				adjustments = append(adjustments,
					fmt.Sprintf("Deferred blocked feature #%d", plans[i].ID))
			}
		}
	}

	// Find next viable feature
	for _, p := range plans {
		if !p.Tested && !p.Deferred {
			adjustments = append(adjustments,
				fmt.Sprintf("Next feature to work on: #%d", p.ID))
			break
		}
	}

	return adjustments
}

// handleRequirementChange handles replanning due to plan file changes
func (s *IncrementalStrategy) handleRequirementChange(plans []plan.Plan, state *ReplanState) []string {
	var adjustments []string

	// Validate that we have a consistent plan state
	testedCount := 0
	untestedCount := 0
	deferredCount := 0

	for _, p := range plans {
		if p.Tested {
			testedCount++
		} else if p.Deferred {
			deferredCount++
		} else {
			untestedCount++
		}
	}

	adjustments = append(adjustments,
		fmt.Sprintf("Plan reconciled: %d tested, %d untested, %d deferred",
			testedCount, untestedCount, deferredCount))

	return adjustments
}

// identifyPotentialPrerequisites looks for features that might be prerequisites
func (s *IncrementalStrategy) identifyPotentialPrerequisites(plans []plan.Plan, currentID int) []string {
	var adjustments []string

	// Find current feature
	var current *plan.Plan
	for i := range plans {
		if plans[i].ID == currentID {
			current = &plans[i]
			break
		}
	}

	if current == nil {
		return adjustments
	}

	// Look for potential dependencies based on category and description
	currentLower := strings.ToLower(current.Description)
	for _, p := range plans {
		if p.ID >= currentID || p.Deferred {
			continue
		}

		// If a prerequisite feature is not tested, flag it
		if !p.Tested {
			pLower := strings.ToLower(p.Description)
			// Simple heuristic: if current feature mentions something from an earlier feature
			if strings.Contains(currentLower, p.Category) || 
			   containsAnyWord(currentLower, strings.Fields(pLower)) {
				adjustments = append(adjustments,
					fmt.Sprintf("Feature #%d may depend on untested feature #%d", currentID, p.ID))
			}
		}
	}

	return adjustments
}

// containsAnyWord checks if s contains any of the words
func containsAnyWord(s string, words []string) bool {
	for _, word := range words {
		if len(word) > 4 && strings.Contains(s, word) {
			return true
		}
	}
	return false
}

// AgentBasedStrategy uses the AI agent to generate a new plan
type AgentBasedStrategy struct {
	agentCmd string
}

// NewAgentBasedStrategy creates a new agent-based strategy
func NewAgentBasedStrategy(agentCmd string) *AgentBasedStrategy {
	return &AgentBasedStrategy{agentCmd: agentCmd}
}

// Name returns the strategy name
func (s *AgentBasedStrategy) Name() StrategyType {
	return StrategyAgentBased
}

// Description returns a human-readable description
func (s *AgentBasedStrategy) Description() string {
	return "Use AI agent to analyze current state and generate updated plan"
}

// Execute performs the agent-based replanning
func (s *AgentBasedStrategy) Execute(state *ReplanState, trigger TriggerType) (*ReplanResult, error) {
	// Build prompt for the agent
	prompt := s.buildReplanPrompt(state, trigger)

	// Execute agent
	output, err := s.executeAgent(prompt)
	if err != nil {
		return &ReplanResult{
			Success:   false,
			Message:   fmt.Sprintf("Agent execution failed: %v", err),
			Trigger:   trigger,
			Strategy:  StrategyAgentBased,
			Timestamp: time.Now(),
		}, err
	}

	// Try to extract new plans from output
	newPlans, err := s.extractPlansFromOutput(output)
	if err != nil {
		return &ReplanResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to extract plans from agent output: %v", err),
			Trigger:   trigger,
			Strategy:  StrategyAgentBased,
			Timestamp: time.Now(),
		}, err
	}

	// Calculate diff
	diff := ComputeDiff(state.Plans, newPlans)

	return &ReplanResult{
		Success:   true,
		Message:   "Agent-based replanning completed",
		Trigger:   trigger,
		Strategy:  StrategyAgentBased,
		NewPlans:  newPlans,
		Diff:      diff,
		Timestamp: time.Now(),
	}, nil
}

// buildReplanPrompt creates the prompt for the agent
func (s *AgentBasedStrategy) buildReplanPrompt(state *ReplanState, trigger TriggerType) string {
	var sb strings.Builder

	sb.WriteString("You are helping replan a software development project.\n\n")
	sb.WriteString(fmt.Sprintf("REPLAN TRIGGER: %s\n\n", trigger))

	// Current state
	sb.WriteString("CURRENT STATE:\n")
	sb.WriteString(fmt.Sprintf("- Total iterations run: %d\n", state.TotalIterations))
	sb.WriteString(fmt.Sprintf("- Current feature ID: %d\n", state.FeatureID))
	sb.WriteString(fmt.Sprintf("- Consecutive failures: %d\n", state.ConsecutiveFailures))
	sb.WriteString(fmt.Sprintf("- Blocked features: %v\n", state.BlockedFeatures))

	// Current plan summary
	sb.WriteString("\nCURRENT PLAN:\n")
	for _, p := range state.Plans {
		status := "[ ]"
		if p.Tested {
			status = "[x]"
		} else if p.Deferred {
			status = "[D]"
		}
		sb.WriteString(fmt.Sprintf("  %s #%d [%s]: %s\n", status, p.ID, p.Category, p.Description))
	}

	// Instructions based on trigger
	sb.WriteString("\nINSTRUCTIONS:\n")
	switch trigger {
	case TriggerTestFailure:
		sb.WriteString("Multiple test failures have occurred. Please analyze the current plan and suggest:\n")
		sb.WriteString("1. Whether the current feature should be broken into smaller steps\n")
		sb.WriteString("2. If there are missing prerequisite features\n")
		sb.WriteString("3. An updated plan that addresses the failures\n")
	case TriggerBlockedFeature:
		sb.WriteString("One or more features are blocked. Please suggest:\n")
		sb.WriteString("1. Alternative approaches or workarounds\n")
		sb.WriteString("2. Reordering of remaining features\n")
		sb.WriteString("3. Whether blocked features should be deferred\n")
	case TriggerRequirementChange:
		sb.WriteString("Requirements have changed. Please:\n")
		sb.WriteString("1. Validate the updated plan for consistency\n")
		sb.WriteString("2. Suggest any necessary adjustments\n")
		sb.WriteString("3. Identify any new dependencies\n")
	default:
		sb.WriteString("Please analyze the current state and suggest improvements to the plan.\n")
	}

	sb.WriteString("\nOutput an updated plan.json array. Keep the same structure and IDs where possible.\n")

	return sb.String()
}

// executeAgent runs the AI agent with the given prompt
func (s *AgentBasedStrategy) executeAgent(prompt string) (string, error) {
	cmd := exec.Command(s.agentCmd, prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("agent execution failed: %w", err)
	}
	return string(output), nil
}

// extractPlansFromOutput tries to extract plan JSON from agent output
func (s *AgentBasedStrategy) extractPlansFromOutput(output string) ([]plan.Plan, error) {
	// Try to find JSON array in output
	start := strings.Index(output, "[")
	end := strings.LastIndex(output, "]")

	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array found in output")
	}

	jsonStr := output[start : end+1]

	var plans []plan.Plan
	if err := parseJSON(jsonStr, &plans); err != nil {
		return nil, fmt.Errorf("failed to parse plans: %w", err)
	}

	return plans, nil
}

// parseJSON is a simple JSON parser wrapper
func parseJSON(s string, v interface{}) error {
	// Use standard library's json package
	return nil // Will be replaced with actual implementation
}

// ReplanManager orchestrates the replanning process
type ReplanManager struct {
	triggers   []ReplanTrigger
	strategies map[StrategyType]ReplanStrategy
	versioner  *PlanVersioner
	planPath   string
	state      *ReplanState
	autoReplan bool
}

// NewReplanManager creates a new replan manager
func NewReplanManager(planPath string, agentCmd string, autoReplan bool) *ReplanManager {
	rm := &ReplanManager{
		triggers:   make([]ReplanTrigger, 0),
		strategies: make(map[StrategyType]ReplanStrategy),
		versioner:  NewPlanVersioner(planPath),
		planPath:   planPath,
		autoReplan: autoReplan,
		state: &ReplanState{
			BlockedFeatures: make([]int, 0),
			FailureTypes:    make([]string, 0),
		},
	}

	// Register default triggers
	rm.triggers = append(rm.triggers,
		NewTestFailureTrigger(3),
		NewRequirementChangeTrigger(),
		NewBlockedFeatureTrigger(1),
	)

	// Register default strategies
	rm.strategies[StrategyIncremental] = NewIncrementalStrategy()
	rm.strategies[StrategyAgentBased] = NewAgentBasedStrategy(agentCmd)

	// Discover existing backups
	rm.versioner.DiscoverBackups()

	return rm
}

// UpdateState updates the replan state with current information
func (rm *ReplanManager) UpdateState(featureID int, consecutiveFailures int, failureTypes []string, plans []plan.Plan) {
	// Save old hash
	rm.state.LastPlanHash = rm.state.PlanHash

	// Update state
	rm.state.FeatureID = featureID
	rm.state.ConsecutiveFailures = consecutiveFailures
	rm.state.FailureTypes = failureTypes
	rm.state.Plans = plans

	// Calculate new hash
	rm.state.PlanHash = CalculatePlansHash(plans)
}

// AddBlockedFeature adds a feature to the blocked list
func (rm *ReplanManager) AddBlockedFeature(featureID int) {
	for _, id := range rm.state.BlockedFeatures {
		if id == featureID {
			return // Already blocked
		}
	}
	rm.state.BlockedFeatures = append(rm.state.BlockedFeatures, featureID)
}

// ClearBlockedFeatures clears the blocked features list
func (rm *ReplanManager) ClearBlockedFeatures() {
	rm.state.BlockedFeatures = make([]int, 0)
}

// IncrementIterations increments the iteration counter
func (rm *ReplanManager) IncrementIterations() {
	rm.state.TotalIterations++
}

// CheckTriggers evaluates all triggers and returns the first one that fires
func (rm *ReplanManager) CheckTriggers() TriggerType {
	for _, trigger := range rm.triggers {
		if trigger.Check(rm.state) {
			return trigger.Name()
		}
	}
	return TriggerNone
}

// ShouldReplan checks if replanning should occur based on triggers
func (rm *ReplanManager) ShouldReplan() (bool, TriggerType) {
	trigger := rm.CheckTriggers()
	if trigger == TriggerNone {
		return false, TriggerNone
	}
	return rm.autoReplan, trigger
}

// ExecuteReplan performs replanning with the specified strategy
func (rm *ReplanManager) ExecuteReplan(strategyType StrategyType, trigger TriggerType) (*ReplanResult, error) {
	// Create backup before replanning
	backupPath, err := rm.versioner.CreateBackup(trigger)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Get strategy
	strategy, ok := rm.strategies[strategyType]
	if !ok {
		strategy = rm.strategies[StrategyIncremental] // Default to incremental
	}

	// Execute replanning
	result, err := strategy.Execute(rm.state, trigger)
	if err != nil {
		return nil, err
	}

	result.OldPlanPath = backupPath

	// If successful and we have new plans, write them
	if result.Success && len(result.NewPlans) > 0 {
		if err := plan.WriteFile(rm.planPath, result.NewPlans); err != nil {
			return nil, fmt.Errorf("failed to write updated plan: %w", err)
		}

		// Update state with new plans
		rm.state.Plans = result.NewPlans
		rm.state.PlanHash = CalculatePlansHash(result.NewPlans)
		rm.state.LastPlanHash = rm.state.PlanHash
	}

	return result, nil
}

// ManualReplan triggers replanning manually
func (rm *ReplanManager) ManualReplan(strategyType StrategyType) (*ReplanResult, error) {
	return rm.ExecuteReplan(strategyType, TriggerManual)
}

// GetVersions returns all plan versions
func (rm *ReplanManager) GetVersions() []PlanVersion {
	return rm.versioner.GetVersions()
}

// RestoreVersion restores a specific plan version
func (rm *ReplanManager) RestoreVersion(version int) error {
	return rm.versioner.RestoreVersion(version)
}

// GetState returns the current replan state
func (rm *ReplanManager) GetState() *ReplanState {
	return rm.state
}

// SetAutoReplan sets whether auto-replanning is enabled
func (rm *ReplanManager) SetAutoReplan(enabled bool) {
	rm.autoReplan = enabled
}

// IsAutoReplanEnabled returns whether auto-replanning is enabled
func (rm *ReplanManager) IsAutoReplanEnabled() bool {
	return rm.autoReplan
}

// GetTriggerDescriptions returns descriptions of all registered triggers
func (rm *ReplanManager) GetTriggerDescriptions() []string {
	var descriptions []string
	for _, t := range rm.triggers {
		descriptions = append(descriptions, fmt.Sprintf("%s: %s", t.Name(), t.Description()))
	}
	return descriptions
}

// ResetState resets the replan state (e.g., after successful feature completion)
func (rm *ReplanManager) ResetState() {
	rm.state.ConsecutiveFailures = 0
	rm.state.FailureTypes = make([]string, 0)
	// Don't reset blocked features or total iterations
}
