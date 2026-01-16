// Package multiagent provides multi-agent collaboration and coordination for Ralph.
// It enables parallel AI coordination with different agent roles working together.
package multiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// AgentRole represents the role an agent plays in the collaboration
type AgentRole string

const (
	// RoleImplementer creates code and implements features
	RoleImplementer AgentRole = "implementer"
	// RoleTester validates code through tests
	RoleTester AgentRole = "tester"
	// RoleReviewer checks code quality and suggests improvements
	RoleReviewer AgentRole = "reviewer"
	// RoleRefactorer improves existing code structure
	RoleRefactorer AgentRole = "refactorer"
)

// ValidRoles lists all valid agent roles
var ValidRoles = []AgentRole{RoleImplementer, RoleTester, RoleReviewer, RoleRefactorer}

// ParseAgentRole parses a string into an AgentRole
func ParseAgentRole(s string) (AgentRole, error) {
	role := AgentRole(strings.ToLower(strings.TrimSpace(s)))
	for _, valid := range ValidRoles {
		if role == valid {
			return role, nil
		}
	}
	return "", fmt.Errorf("invalid agent role %q: must be one of implementer, tester, reviewer, refactorer", s)
}

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	StatusIdle     AgentStatus = "idle"
	StatusRunning  AgentStatus = "running"
	StatusComplete AgentStatus = "complete"
	StatusFailed   AgentStatus = "failed"
	StatusTimeout  AgentStatus = "timeout"
)

// AgentConfig defines configuration for a single agent
type AgentConfig struct {
	// ID is a unique identifier for this agent configuration
	ID string `json:"id" yaml:"id"`

	// Role defines the agent's purpose (implementer, tester, reviewer, refactorer)
	Role AgentRole `json:"role" yaml:"role"`

	// Command is the CLI command to execute the agent (e.g., "cursor-agent", "claude")
	Command string `json:"command" yaml:"command"`

	// Specialization describes what this agent specializes in (e.g., "frontend", "backend", "testing")
	Specialization string `json:"specialization,omitempty" yaml:"specialization,omitempty"`

	// Priority determines execution order when roles have dependencies (higher = earlier)
	Priority int `json:"priority,omitempty" yaml:"priority,omitempty"`

	// Timeout is the maximum duration for this agent's execution
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// Enabled indicates if this agent should be used
	Enabled bool `json:"enabled" yaml:"enabled"`

	// PromptPrefix is prepended to all prompts sent to this agent
	PromptPrefix string `json:"prompt_prefix,omitempty" yaml:"prompt_prefix,omitempty"`

	// PromptSuffix is appended to all prompts sent to this agent
	PromptSuffix string `json:"prompt_suffix,omitempty" yaml:"prompt_suffix,omitempty"`
}

// MultiAgentConfig holds the configuration for multiple agents
type MultiAgentConfig struct {
	// Agents is the list of agent configurations
	Agents []AgentConfig `json:"agents" yaml:"agents"`

	// MaxParallel is the maximum number of agents to run in parallel
	MaxParallel int `json:"max_parallel,omitempty" yaml:"max_parallel,omitempty"`

	// ContextFile is the shared context file path for agent communication
	ContextFile string `json:"context_file,omitempty" yaml:"context_file,omitempty"`

	// ConflictResolution determines how to resolve conflicts between agents
	// Options: "priority" (use highest priority), "merge" (attempt to merge), "vote" (majority wins)
	ConflictResolution string `json:"conflict_resolution,omitempty" yaml:"conflict_resolution,omitempty"`
}

// AgentResult represents the result of an agent's execution
type AgentResult struct {
	// AgentID is the ID of the agent that produced this result
	AgentID string `json:"agent_id"`

	// Role is the role of the agent
	Role AgentRole `json:"role"`

	// Status indicates whether the execution was successful
	Status AgentStatus `json:"status"`

	// Output is the raw output from the agent
	Output string `json:"output"`

	// Error contains error information if the execution failed
	Error string `json:"error,omitempty"`

	// StartTime is when the agent started
	StartTime time.Time `json:"start_time"`

	// EndTime is when the agent finished
	EndTime time.Time `json:"end_time"`

	// Duration is how long the agent ran
	Duration time.Duration `json:"duration"`

	// Suggestions contains extracted suggestions or recommendations
	Suggestions []string `json:"suggestions,omitempty"`

	// Issues contains extracted issues or problems found
	Issues []string `json:"issues,omitempty"`

	// Approved indicates if a reviewer approved the changes
	Approved bool `json:"approved,omitempty"`
}

// SharedContext represents the shared context file for inter-agent communication
type SharedContext struct {
	mu sync.RWMutex

	// Path to the context file
	Path string `json:"-"`

	// FeatureID is the current feature being worked on
	FeatureID int `json:"feature_id"`

	// FeatureDescription describes the current feature
	FeatureDescription string `json:"feature_description"`

	// Iteration is the current iteration number
	Iteration int `json:"iteration"`

	// Results contains results from all agents
	Results []AgentResult `json:"results"`

	// Messages contains inter-agent messages
	Messages []ContextMessage `json:"messages"`

	// Decisions contains agreed-upon decisions
	Decisions []ContextDecision `json:"decisions"`

	// LastUpdated is when the context was last modified
	LastUpdated time.Time `json:"last_updated"`
}

// ContextMessage represents a message between agents
type ContextMessage struct {
	// FromAgent is the ID of the sending agent
	FromAgent string `json:"from_agent"`

	// ToAgent is the ID of the receiving agent ("all" for broadcast)
	ToAgent string `json:"to_agent"`

	// Type is the message type (info, warning, request, response)
	Type string `json:"type"`

	// Content is the message content
	Content string `json:"content"`

	// Timestamp is when the message was created
	Timestamp time.Time `json:"timestamp"`
}

// ContextDecision represents a decision made by agents
type ContextDecision struct {
	// Topic is what the decision is about
	Topic string `json:"topic"`

	// Decision is the actual decision made
	Decision string `json:"decision"`

	// Agents lists which agents agreed to this decision
	Agents []string `json:"agents"`

	// Timestamp is when the decision was made
	Timestamp time.Time `json:"timestamp"`
}

// NewSharedContext creates a new shared context
func NewSharedContext(path string) *SharedContext {
	return &SharedContext{
		Path:        path,
		Results:     make([]AgentResult, 0),
		Messages:    make([]ContextMessage, 0),
		Decisions:   make([]ContextDecision, 0),
		LastUpdated: time.Now(),
	}
}

// Load loads the shared context from file
func (sc *SharedContext) Load() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.Path == "" {
		return nil
	}

	data, err := os.ReadFile(sc.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, start fresh
		}
		return fmt.Errorf("failed to read context file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, sc); err != nil {
		return fmt.Errorf("failed to parse context file: %w", err)
	}

	return nil
}

// Save saves the shared context to file
func (sc *SharedContext) Save() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.Path == "" {
		return nil
	}

	sc.LastUpdated = time.Now()

	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(sc.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	if err := os.WriteFile(sc.Path, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

// AddResult adds an agent result to the context
func (sc *SharedContext) AddResult(result AgentResult) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.Results = append(sc.Results, result)
	sc.LastUpdated = time.Now()
}

// AddMessage adds a message to the context
func (sc *SharedContext) AddMessage(msg ContextMessage) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	msg.Timestamp = time.Now()
	sc.Messages = append(sc.Messages, msg)
	sc.LastUpdated = time.Now()
}

// AddDecision adds a decision to the context
func (sc *SharedContext) AddDecision(decision ContextDecision) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	decision.Timestamp = time.Now()
	sc.Decisions = append(sc.Decisions, decision)
	sc.LastUpdated = time.Now()
}

// GetResultsByRole returns all results from agents with the specified role
func (sc *SharedContext) GetResultsByRole(role AgentRole) []AgentResult {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var results []AgentResult
	for _, r := range sc.Results {
		if r.Role == role {
			results = append(results, r)
		}
	}
	return results
}

// GetMessagesFor returns messages addressed to a specific agent or "all"
func (sc *SharedContext) GetMessagesFor(agentID string) []ContextMessage {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var messages []ContextMessage
	for _, m := range sc.Messages {
		if m.ToAgent == agentID || m.ToAgent == "all" {
			messages = append(messages, m)
		}
	}
	return messages
}

// Clear clears the shared context
func (sc *SharedContext) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.Results = make([]AgentResult, 0)
	sc.Messages = make([]ContextMessage, 0)
	sc.Decisions = make([]ContextDecision, 0)
	sc.LastUpdated = time.Now()
}

// SetFeature sets the current feature information
func (sc *SharedContext) SetFeature(id int, description string, iteration int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.FeatureID = id
	sc.FeatureDescription = description
	sc.Iteration = iteration
	sc.LastUpdated = time.Now()
}

// Orchestrator coordinates multiple agents working together
type Orchestrator struct {
	config       *MultiAgentConfig
	context      *SharedContext
	executor     AgentExecutor
	mu           sync.Mutex
	agentStatus  map[string]AgentStatus
	healthChecks map[string]time.Time
}

// AgentExecutor is an interface for executing agent commands
// This allows for mocking in tests
type AgentExecutor interface {
	Execute(ctx context.Context, agentConfig *AgentConfig, prompt string) (string, error)
}

// DefaultAgentExecutor is the default implementation that executes real agent commands
type DefaultAgentExecutor struct {
	Verbose bool
}

// Execute runs an agent command and returns the output
func (e *DefaultAgentExecutor) Execute(ctx context.Context, agentConfig *AgentConfig, prompt string) (string, error) {
	// This will be implemented to call the actual agent command
	// For now, we'll use the existing agent.Execute function pattern
	return "", fmt.Errorf("default executor not fully implemented - use with actual agent package")
}

// NewOrchestrator creates a new multi-agent orchestrator
func NewOrchestrator(config *MultiAgentConfig, contextPath string) *Orchestrator {
	// Set defaults
	if config.MaxParallel <= 0 {
		config.MaxParallel = 2
	}
	if config.ConflictResolution == "" {
		config.ConflictResolution = "priority"
	}
	if contextPath == "" {
		contextPath = ".ralph-multiagent-context.json"
	}

	return &Orchestrator{
		config:       config,
		context:      NewSharedContext(contextPath),
		executor:     &DefaultAgentExecutor{},
		agentStatus:  make(map[string]AgentStatus),
		healthChecks: make(map[string]time.Time),
	}
}

// SetExecutor sets a custom agent executor (useful for testing)
func (o *Orchestrator) SetExecutor(executor AgentExecutor) {
	o.executor = executor
}

// GetEnabledAgents returns all enabled agents sorted by priority (high to low)
func (o *Orchestrator) GetEnabledAgents() []AgentConfig {
	var enabled []AgentConfig
	for _, agent := range o.config.Agents {
		if agent.Enabled {
			enabled = append(enabled, agent)
		}
	}

	// Sort by priority (higher priority first)
	sort.Slice(enabled, func(i, j int) bool {
		return enabled[i].Priority > enabled[j].Priority
	})

	return enabled
}

// GetAgentsByRole returns all enabled agents with the specified role
func (o *Orchestrator) GetAgentsByRole(role AgentRole) []AgentConfig {
	var agents []AgentConfig
	for _, agent := range o.config.Agents {
		if agent.Enabled && agent.Role == role {
			agents = append(agents, agent)
		}
	}
	return agents
}

// ExecuteWorkflow runs the standard multi-agent workflow for a feature
// The workflow is: Implementer -> Tester -> Reviewer -> (optional) Refactorer
func (o *Orchestrator) ExecuteWorkflow(ctx context.Context, featureID int, featureDesc string, iteration int, basePrompt string) (*WorkflowResult, error) {
	o.mu.Lock()
	o.context.SetFeature(featureID, featureDesc, iteration)
	if err := o.context.Load(); err != nil {
		o.mu.Unlock()
		return nil, fmt.Errorf("failed to load context: %w", err)
	}
	o.mu.Unlock()

	result := &WorkflowResult{
		FeatureID:   featureID,
		FeatureDesc: featureDesc,
		Iteration:   iteration,
		StartTime:   time.Now(),
		Stages:      make([]StageResult, 0),
	}

	// Stage 1: Implementation
	implementers := o.GetAgentsByRole(RoleImplementer)
	if len(implementers) > 0 {
		stageResult := o.executeStage(ctx, "implementation", implementers, basePrompt, nil)
		result.Stages = append(result.Stages, stageResult)
		if !stageResult.Success {
			result.EndTime = time.Now()
			result.Success = false
			result.Error = "implementation stage failed"
			return result, nil
		}
	}

	// Stage 2: Testing (depends on implementation)
	testers := o.GetAgentsByRole(RoleTester)
	if len(testers) > 0 {
		testPrompt := o.buildDependentPrompt(basePrompt, RoleImplementer, "Validate the implementation by running and writing tests.")
		stageResult := o.executeStage(ctx, "testing", testers, testPrompt, nil)
		result.Stages = append(result.Stages, stageResult)
		// Testing failures don't stop the workflow, but are recorded
	}

	// Stage 3: Review (depends on implementation)
	reviewers := o.GetAgentsByRole(RoleReviewer)
	if len(reviewers) > 0 {
		reviewPrompt := o.buildDependentPrompt(basePrompt, RoleImplementer, "Review the implementation for code quality, best practices, and potential issues.")
		stageResult := o.executeStage(ctx, "review", reviewers, reviewPrompt, nil)
		result.Stages = append(result.Stages, stageResult)
	}

	// Stage 4: Refactoring (optional, based on review feedback)
	refactorers := o.GetAgentsByRole(RoleRefactorer)
	if len(refactorers) > 0 && o.shouldRefactor(result) {
		refactorPrompt := o.buildDependentPrompt(basePrompt, RoleReviewer, "Refactor the code based on review feedback to improve quality.")
		stageResult := o.executeStage(ctx, "refactoring", refactorers, refactorPrompt, nil)
		result.Stages = append(result.Stages, stageResult)
	}

	result.EndTime = time.Now()
	result.Success = o.evaluateWorkflowSuccess(result)

	// Save context
	o.mu.Lock()
	if err := o.context.Save(); err != nil {
		o.mu.Unlock()
		return result, fmt.Errorf("failed to save context: %w", err)
	}
	o.mu.Unlock()

	return result, nil
}

// executeStage runs a group of agents for a workflow stage
func (o *Orchestrator) executeStage(ctx context.Context, stageName string, agents []AgentConfig, prompt string, previousResults []AgentResult) StageResult {
	result := StageResult{
		Name:      stageName,
		StartTime: time.Now(),
		Results:   make([]AgentResult, 0),
	}

	// Execute agents in parallel (up to MaxParallel)
	results := o.executeParallel(ctx, agents, prompt)
	result.Results = results

	// Aggregate results
	result.Success = o.evaluateStageSuccess(results)
	result.EndTime = time.Now()

	// Add results to shared context
	for _, r := range results {
		o.context.AddResult(r)
	}

	return result
}

// executeParallel runs multiple agents in parallel with concurrency limit
func (o *Orchestrator) executeParallel(ctx context.Context, agents []AgentConfig, prompt string) []AgentResult {
	if len(agents) == 0 {
		return nil
	}

	results := make([]AgentResult, len(agents))
	sem := make(chan struct{}, o.config.MaxParallel)
	var wg sync.WaitGroup

	for i, agent := range agents {
		wg.Add(1)
		go func(idx int, ag AgentConfig) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = o.executeAgent(ctx, &ag, prompt)
		}(i, agent)
	}

	wg.Wait()
	return results
}

// executeAgent runs a single agent and returns the result
func (o *Orchestrator) executeAgent(ctx context.Context, agent *AgentConfig, prompt string) AgentResult {
	result := AgentResult{
		AgentID:   agent.ID,
		Role:      agent.Role,
		StartTime: time.Now(),
	}

	// Update status
	o.mu.Lock()
	o.agentStatus[agent.ID] = StatusRunning
	o.healthChecks[agent.ID] = time.Now()
	o.mu.Unlock()

	// Build the full prompt with prefix/suffix
	fullPrompt := prompt
	if agent.PromptPrefix != "" {
		fullPrompt = agent.PromptPrefix + "\n\n" + fullPrompt
	}
	if agent.PromptSuffix != "" {
		fullPrompt = fullPrompt + "\n\n" + agent.PromptSuffix
	}

	// Create context with timeout if specified
	execCtx := ctx
	if agent.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, agent.Timeout)
		defer cancel()
	}

	// Execute the agent
	output, err := o.executor.Execute(execCtx, agent, fullPrompt)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = StatusTimeout
			result.Error = "execution timed out"
		} else {
			result.Status = StatusFailed
			result.Error = err.Error()
		}
	} else {
		result.Status = StatusComplete
		result.Output = output

		// Extract suggestions and issues from output
		result.Suggestions = extractSuggestions(output)
		result.Issues = extractIssues(output)

		// Check for approval (for reviewers)
		if agent.Role == RoleReviewer {
			result.Approved = isApproved(output)
		}
	}

	// Update status
	o.mu.Lock()
	o.agentStatus[agent.ID] = result.Status
	o.mu.Unlock()

	return result
}

// buildDependentPrompt creates a prompt that includes context from previous stages
func (o *Orchestrator) buildDependentPrompt(basePrompt string, dependsOnRole AgentRole, instruction string) string {
	var sb strings.Builder
	sb.WriteString(instruction)
	sb.WriteString("\n\n")

	// Include relevant results from the dependent role
	results := o.context.GetResultsByRole(dependsOnRole)
	if len(results) > 0 {
		sb.WriteString(fmt.Sprintf("=== Previous %s Output ===\n", dependsOnRole))
		for _, r := range results {
			if r.Status == StatusComplete && r.Output != "" {
				sb.WriteString(r.Output)
				sb.WriteString("\n---\n")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("=== Original Task ===\n")
	sb.WriteString(basePrompt)

	return sb.String()
}

// shouldRefactor determines if refactoring should be triggered based on review results
func (o *Orchestrator) shouldRefactor(result *WorkflowResult) bool {
	// Check if there are review results with issues
	for _, stage := range result.Stages {
		if stage.Name == "review" {
			for _, r := range stage.Results {
				if len(r.Issues) > 0 || !r.Approved {
					return true
				}
			}
		}
	}
	return false
}

// evaluateStageSuccess determines if a stage was successful
func (o *Orchestrator) evaluateStageSuccess(results []AgentResult) bool {
	if len(results) == 0 {
		return true
	}

	successCount := 0
	for _, r := range results {
		if r.Status == StatusComplete {
			successCount++
		}
	}

	// At least one agent must succeed
	return successCount > 0
}

// evaluateWorkflowSuccess determines if the overall workflow was successful
func (o *Orchestrator) evaluateWorkflowSuccess(result *WorkflowResult) bool {
	// Implementation stage must succeed
	for _, stage := range result.Stages {
		if stage.Name == "implementation" && !stage.Success {
			return false
		}
	}
	return true
}

// ResolveConflicts resolves conflicts between agent results
func (o *Orchestrator) ResolveConflicts(results []AgentResult) (*ConflictResolution, error) {
	resolution := &ConflictResolution{
		Strategy:  o.config.ConflictResolution,
		Conflicts: make([]Conflict, 0),
	}

	// Detect conflicts (e.g., different suggestions for the same thing)
	conflicts := detectConflicts(results)
	resolution.Conflicts = conflicts

	if len(conflicts) == 0 {
		resolution.Resolved = true
		return resolution, nil
	}

	// Resolve based on strategy
	switch o.config.ConflictResolution {
	case "priority":
		resolution.WinningResults = o.resolveByPriority(results, conflicts)
	case "merge":
		resolution.WinningResults = o.resolveByMerge(results, conflicts)
	case "vote":
		resolution.WinningResults = o.resolveByVote(results, conflicts)
	default:
		resolution.WinningResults = o.resolveByPriority(results, conflicts)
	}

	resolution.Resolved = true
	return resolution, nil
}

// resolveByPriority resolves conflicts by using the highest priority agent's result
func (o *Orchestrator) resolveByPriority(results []AgentResult, conflicts []Conflict) []AgentResult {
	// Get agent priorities
	priorities := make(map[string]int)
	for _, agent := range o.config.Agents {
		priorities[agent.ID] = agent.Priority
	}

	// For each conflict, pick the result from the highest priority agent
	winning := make([]AgentResult, 0)
	usedAgents := make(map[string]bool)

	for _, r := range results {
		if r.Status == StatusComplete && !usedAgents[r.AgentID] {
			winning = append(winning, r)
			usedAgents[r.AgentID] = true
		}
	}

	// Sort by priority
	sort.Slice(winning, func(i, j int) bool {
		return priorities[winning[i].AgentID] > priorities[winning[j].AgentID]
	})

	return winning
}

// resolveByMerge attempts to merge non-conflicting suggestions from all agents
func (o *Orchestrator) resolveByMerge(results []AgentResult, conflicts []Conflict) []AgentResult {
	// For merge strategy, we keep all results but combine suggestions
	merged := make([]AgentResult, 0)
	allSuggestions := make([]string, 0)
	allIssues := make([]string, 0)

	for _, r := range results {
		if r.Status == StatusComplete {
			merged = append(merged, r)
			allSuggestions = append(allSuggestions, r.Suggestions...)
			allIssues = append(allIssues, r.Issues...)
		}
	}

	// Deduplicate suggestions and issues
	allSuggestions = deduplicateStrings(allSuggestions)
	allIssues = deduplicateStrings(allIssues)

	// Add merged suggestions/issues to the first result
	if len(merged) > 0 {
		merged[0].Suggestions = allSuggestions
		merged[0].Issues = allIssues
	}

	return merged
}

// resolveByVote resolves conflicts by majority vote (most common suggestion wins)
func (o *Orchestrator) resolveByVote(results []AgentResult, conflicts []Conflict) []AgentResult {
	// Count votes for each suggestion
	votes := make(map[string]int)
	for _, r := range results {
		if r.Status == StatusComplete {
			for _, s := range r.Suggestions {
				votes[s]++
			}
		}
	}

	// Keep suggestions that have majority support
	threshold := len(results) / 2
	winningSuggestions := make([]string, 0)
	for suggestion, count := range votes {
		if count > threshold {
			winningSuggestions = append(winningSuggestions, suggestion)
		}
	}

	// Return results with winning suggestions
	winning := make([]AgentResult, 0)
	for _, r := range results {
		if r.Status == StatusComplete {
			r.Suggestions = winningSuggestions
			winning = append(winning, r)
		}
	}

	return winning
}

// GetAgentStatus returns the current status of an agent
func (o *Orchestrator) GetAgentStatus(agentID string) AgentStatus {
	o.mu.Lock()
	defer o.mu.Unlock()
	if status, ok := o.agentStatus[agentID]; ok {
		return status
	}
	return StatusIdle
}

// GetHealthStatus returns the health status of all agents
func (o *Orchestrator) GetHealthStatus() map[string]HealthInfo {
	o.mu.Lock()
	defer o.mu.Unlock()

	health := make(map[string]HealthInfo)
	now := time.Now()

	for _, agent := range o.config.Agents {
		if !agent.Enabled {
			continue
		}

		info := HealthInfo{
			AgentID: agent.ID,
			Status:  o.agentStatus[agent.ID],
			Healthy: true,
		}

		// Check if agent has been running too long
		if lastCheck, ok := o.healthChecks[agent.ID]; ok {
			info.LastCheck = lastCheck
			if info.Status == StatusRunning {
				runningTime := now.Sub(lastCheck)
				timeout := agent.Timeout
				if timeout == 0 {
					timeout = 5 * time.Minute // Default timeout
				}
				if runningTime > timeout*2 {
					info.Healthy = false
					info.Message = "agent may be stuck"
				}
			}
		}

		health[agent.ID] = info
	}

	return health
}

// WorkflowResult contains the results of a multi-agent workflow execution
type WorkflowResult struct {
	FeatureID   int           `json:"feature_id"`
	FeatureDesc string        `json:"feature_desc"`
	Iteration   int           `json:"iteration"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Stages      []StageResult `json:"stages"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
}

// StageResult contains results for a single workflow stage
type StageResult struct {
	Name      string        `json:"name"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Results   []AgentResult `json:"results"`
	Success   bool          `json:"success"`
}

// ConflictResolution contains the results of conflict resolution
type ConflictResolution struct {
	Strategy       string        `json:"strategy"`
	Conflicts      []Conflict    `json:"conflicts"`
	WinningResults []AgentResult `json:"winning_results"`
	Resolved       bool          `json:"resolved"`
}

// Conflict represents a conflict between agent results
type Conflict struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	AgentIDs    []string `json:"agent_ids"`
}

// HealthInfo contains health information for an agent
type HealthInfo struct {
	AgentID   string      `json:"agent_id"`
	Status    AgentStatus `json:"status"`
	Healthy   bool        `json:"healthy"`
	LastCheck time.Time   `json:"last_check"`
	Message   string      `json:"message,omitempty"`
}

// Helper functions

// extractSuggestions extracts suggestions from agent output
func extractSuggestions(output string) []string {
	var suggestions []string
	lines := strings.Split(output, "\n")
	inSuggestions := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)

		// Check for suggestion markers
		if strings.Contains(lower, "suggestion") || strings.Contains(lower, "recommend") {
			inSuggestions = true
		}

		// Extract bullet points following suggestion markers
		if inSuggestions && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "•")) {
			suggestion := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(line, "-"), "*"), "•")
			suggestion = strings.TrimSpace(suggestion)
			if suggestion != "" {
				suggestions = append(suggestions, suggestion)
			}
		}

		// Reset if we hit a different section
		if len(line) > 0 && !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "•") && inSuggestions {
			if !strings.Contains(lower, "suggestion") && !strings.Contains(lower, "recommend") {
				inSuggestions = false
			}
		}
	}

	return suggestions
}

// extractIssues extracts issues from agent output
func extractIssues(output string) []string {
	var issues []string
	lines := strings.Split(output, "\n")
	inIssues := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)

		// Check for issue markers
		if strings.Contains(lower, "issue") || strings.Contains(lower, "problem") ||
			strings.Contains(lower, "error") || strings.Contains(lower, "concern") {
			inIssues = true
		}

		// Extract bullet points following issue markers
		if inIssues && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "•")) {
			issue := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(line, "-"), "*"), "•")
			issue = strings.TrimSpace(issue)
			if issue != "" {
				issues = append(issues, issue)
			}
		}

		// Reset if we hit a different section
		if len(line) > 0 && !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "•") && inIssues {
			if !strings.Contains(lower, "issue") && !strings.Contains(lower, "problem") &&
				!strings.Contains(lower, "error") && !strings.Contains(lower, "concern") {
				inIssues = false
			}
		}
	}

	return issues
}

// isApproved checks if a reviewer approved the changes
func isApproved(output string) bool {
	lower := strings.ToLower(output)
	approvalPatterns := []string{
		"approved",
		"lgtm",
		"looks good",
		"ship it",
		"ready to merge",
		"no issues found",
		"no concerns",
	}
	rejectionPatterns := []string{
		"rejected",
		"not approved",
		"needs work",
		"request changes",
		"do not merge",
		"major issues",
	}

	// Check for rejection first
	for _, pattern := range rejectionPatterns {
		if strings.Contains(lower, pattern) {
			return false
		}
	}

	// Check for approval
	for _, pattern := range approvalPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// Default to approved if no clear signals
	return true
}

// detectConflicts detects conflicts between agent results
func detectConflicts(results []AgentResult) []Conflict {
	var conflicts []Conflict

	// Detect approval conflicts (some approve, some reject)
	approved := make([]string, 0)
	rejected := make([]string, 0)
	for _, r := range results {
		if r.Role == RoleReviewer {
			if r.Approved {
				approved = append(approved, r.AgentID)
			} else {
				rejected = append(rejected, r.AgentID)
			}
		}
	}
	if len(approved) > 0 && len(rejected) > 0 {
		conflicts = append(conflicts, Conflict{
			Type:        "approval",
			Description: "Reviewers disagree on approval",
			AgentIDs:    append(approved, rejected...),
		})
	}

	return conflicts
}

// deduplicateStrings removes duplicate strings from a slice
func deduplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// LoadMultiAgentConfig loads multi-agent configuration from a file
func LoadMultiAgentConfig(path string) (*MultiAgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents config file: %w", err)
	}

	config := &MultiAgentConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse agents config file: %w", err)
	}

	// Validate configuration
	if err := validateMultiAgentConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// validateMultiAgentConfig validates the multi-agent configuration
func validateMultiAgentConfig(config *MultiAgentConfig) error {
	if len(config.Agents) == 0 {
		return fmt.Errorf("no agents configured")
	}

	seenIDs := make(map[string]bool)
	for i, agent := range config.Agents {
		if agent.ID == "" {
			return fmt.Errorf("agent at index %d has no ID", i)
		}
		if seenIDs[agent.ID] {
			return fmt.Errorf("duplicate agent ID: %s", agent.ID)
		}
		seenIDs[agent.ID] = true

		if agent.Command == "" {
			return fmt.Errorf("agent %s has no command", agent.ID)
		}

		if _, err := ParseAgentRole(string(agent.Role)); err != nil {
			return fmt.Errorf("agent %s: %w", agent.ID, err)
		}
	}

	if config.MaxParallel < 0 {
		return fmt.Errorf("max_parallel cannot be negative")
	}

	validResolutions := map[string]bool{
		"":         true,
		"priority": true,
		"merge":    true,
		"vote":     true,
	}
	if !validResolutions[config.ConflictResolution] {
		return fmt.Errorf("invalid conflict_resolution %q: must be priority, merge, or vote", config.ConflictResolution)
	}

	return nil
}

// Summary returns a human-readable summary of the workflow result
func (wr *WorkflowResult) Summary() string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Multi-Agent Workflow for Feature #%d\n", wr.FeatureID))
	sb.WriteString(fmt.Sprintf("Feature: %s\n", wr.FeatureDesc))
	sb.WriteString(fmt.Sprintf("Iteration: %d\n", wr.Iteration))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", wr.EndTime.Sub(wr.StartTime).Round(time.Second)))
	
	if wr.Success {
		sb.WriteString("Status: SUCCESS\n")
	} else {
		sb.WriteString(fmt.Sprintf("Status: FAILED (%s)\n", wr.Error))
	}
	
	sb.WriteString("\nStages:\n")
	for _, stage := range wr.Stages {
		status := "✓"
		if !stage.Success {
			status = "✗"
		}
		sb.WriteString(fmt.Sprintf("  %s %s (%d agents, %s)\n", 
			status, stage.Name, len(stage.Results), 
			stage.EndTime.Sub(stage.StartTime).Round(time.Second)))
		
		for _, r := range stage.Results {
			agentStatus := "✓"
			if r.Status != StatusComplete {
				agentStatus = "✗"
			}
			sb.WriteString(fmt.Sprintf("    %s %s [%s]: %s\n", 
				agentStatus, r.AgentID, r.Role, r.Status))
			
			if len(r.Suggestions) > 0 {
				sb.WriteString(fmt.Sprintf("      Suggestions: %d\n", len(r.Suggestions)))
			}
			if len(r.Issues) > 0 {
				sb.WriteString(fmt.Sprintf("      Issues: %d\n", len(r.Issues)))
			}
		}
	}
	
	return sb.String()
}
