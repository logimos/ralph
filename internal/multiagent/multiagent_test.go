package multiagent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// MockAgentExecutor is a mock implementation of AgentExecutor for testing
type MockAgentExecutor struct {
	mu       sync.Mutex
	results  map[string]string // AgentID -> Output
	errors   map[string]error  // AgentID -> Error
	delays   map[string]time.Duration // AgentID -> Delay
	calls    []ExecutorCall
}

type ExecutorCall struct {
	AgentID string
	Prompt  string
}

func NewMockExecutor() *MockAgentExecutor {
	return &MockAgentExecutor{
		results: make(map[string]string),
		errors:  make(map[string]error),
		delays:  make(map[string]time.Duration),
		calls:   make([]ExecutorCall, 0),
	}
}

func (m *MockAgentExecutor) SetResult(agentID, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[agentID] = output
}

func (m *MockAgentExecutor) SetError(agentID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[agentID] = err
}

func (m *MockAgentExecutor) SetDelay(agentID string, delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delays[agentID] = delay
}

func (m *MockAgentExecutor) Execute(ctx context.Context, agentConfig *AgentConfig, prompt string) (string, error) {
	m.mu.Lock()
	m.calls = append(m.calls, ExecutorCall{AgentID: agentConfig.ID, Prompt: prompt})
	
	delay := m.delays[agentConfig.ID]
	result := m.results[agentConfig.ID]
	err := m.errors[agentConfig.ID]
	m.mu.Unlock()

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return result, err
}

func (m *MockAgentExecutor) GetCalls() []ExecutorCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]ExecutorCall, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// Tests

func TestParseAgentRole(t *testing.T) {
	tests := []struct {
		input    string
		expected AgentRole
		wantErr  bool
	}{
		{"implementer", RoleImplementer, false},
		{"IMPLEMENTER", RoleImplementer, false},
		{"Implementer", RoleImplementer, false},
		{"tester", RoleTester, false},
		{"reviewer", RoleReviewer, false},
		{"refactorer", RoleRefactorer, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAgentRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAgentRole(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseAgentRole(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSharedContext(t *testing.T) {
	// Create temp file for context
	tmpDir := t.TempDir()
	contextPath := filepath.Join(tmpDir, "context.json")

	t.Run("NewSharedContext", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		if sc.Path != contextPath {
			t.Errorf("Path = %q, want %q", sc.Path, contextPath)
		}
		if len(sc.Results) != 0 {
			t.Errorf("Results should be empty initially")
		}
	})

	t.Run("SetFeature", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		sc.SetFeature(42, "Test feature", 3)
		
		if sc.FeatureID != 42 {
			t.Errorf("FeatureID = %d, want 42", sc.FeatureID)
		}
		if sc.FeatureDescription != "Test feature" {
			t.Errorf("FeatureDescription = %q, want %q", sc.FeatureDescription, "Test feature")
		}
		if sc.Iteration != 3 {
			t.Errorf("Iteration = %d, want 3", sc.Iteration)
		}
	})

	t.Run("AddResult", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		result := AgentResult{
			AgentID: "test-agent",
			Role:    RoleImplementer,
			Status:  StatusComplete,
			Output:  "Test output",
		}
		sc.AddResult(result)
		
		if len(sc.Results) != 1 {
			t.Errorf("Results length = %d, want 1", len(sc.Results))
		}
		if sc.Results[0].AgentID != "test-agent" {
			t.Errorf("AgentID = %q, want %q", sc.Results[0].AgentID, "test-agent")
		}
	})

	t.Run("AddMessage", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		msg := ContextMessage{
			FromAgent: "agent-1",
			ToAgent:   "agent-2",
			Type:      "info",
			Content:   "Test message",
		}
		sc.AddMessage(msg)
		
		if len(sc.Messages) != 1 {
			t.Errorf("Messages length = %d, want 1", len(sc.Messages))
		}
		if sc.Messages[0].Content != "Test message" {
			t.Errorf("Content = %q, want %q", sc.Messages[0].Content, "Test message")
		}
	})

	t.Run("AddDecision", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		decision := ContextDecision{
			Topic:    "Architecture",
			Decision: "Use microservices",
			Agents:   []string{"agent-1", "agent-2"},
		}
		sc.AddDecision(decision)
		
		if len(sc.Decisions) != 1 {
			t.Errorf("Decisions length = %d, want 1", len(sc.Decisions))
		}
	})

	t.Run("GetResultsByRole", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		sc.AddResult(AgentResult{AgentID: "impl-1", Role: RoleImplementer, Status: StatusComplete})
		sc.AddResult(AgentResult{AgentID: "test-1", Role: RoleTester, Status: StatusComplete})
		sc.AddResult(AgentResult{AgentID: "impl-2", Role: RoleImplementer, Status: StatusComplete})
		
		results := sc.GetResultsByRole(RoleImplementer)
		if len(results) != 2 {
			t.Errorf("GetResultsByRole(Implementer) returned %d results, want 2", len(results))
		}
	})

	t.Run("GetMessagesFor", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		sc.AddMessage(ContextMessage{FromAgent: "a", ToAgent: "b", Content: "msg1"})
		sc.AddMessage(ContextMessage{FromAgent: "a", ToAgent: "all", Content: "msg2"})
		sc.AddMessage(ContextMessage{FromAgent: "a", ToAgent: "c", Content: "msg3"})
		
		messagesForB := sc.GetMessagesFor("b")
		if len(messagesForB) != 2 { // "b" + "all"
			t.Errorf("GetMessagesFor(b) returned %d messages, want 2", len(messagesForB))
		}
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		sc.SetFeature(1, "Feature 1", 1)
		sc.AddResult(AgentResult{AgentID: "test", Role: RoleTester, Status: StatusComplete})
		
		err := sc.Save()
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		
		// Load into new context
		sc2 := NewSharedContext(contextPath)
		err = sc2.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		
		if sc2.FeatureID != 1 {
			t.Errorf("After load, FeatureID = %d, want 1", sc2.FeatureID)
		}
		if len(sc2.Results) != 1 {
			t.Errorf("After load, Results length = %d, want 1", len(sc2.Results))
		}
	})

	t.Run("Clear", func(t *testing.T) {
		sc := NewSharedContext(contextPath)
		sc.AddResult(AgentResult{AgentID: "test", Role: RoleTester})
		sc.AddMessage(ContextMessage{Content: "test"})
		sc.Clear()
		
		if len(sc.Results) != 0 {
			t.Errorf("After Clear(), Results should be empty")
		}
		if len(sc.Messages) != 0 {
			t.Errorf("After Clear(), Messages should be empty")
		}
	})
}

func TestOrchestrator(t *testing.T) {
	config := &MultiAgentConfig{
		Agents: []AgentConfig{
			{ID: "impl-1", Role: RoleImplementer, Command: "test-cmd", Priority: 10, Enabled: true},
			{ID: "test-1", Role: RoleTester, Command: "test-cmd", Priority: 8, Enabled: true},
			{ID: "review-1", Role: RoleReviewer, Command: "test-cmd", Priority: 6, Enabled: true},
			{ID: "disabled", Role: RoleImplementer, Command: "test-cmd", Priority: 5, Enabled: false},
		},
		MaxParallel:        2,
		ConflictResolution: "priority",
	}

	t.Run("NewOrchestrator", func(t *testing.T) {
		orch := NewOrchestrator(config, "")
		if orch == nil {
			t.Fatal("NewOrchestrator returned nil")
		}
		if orch.config.MaxParallel != 2 {
			t.Errorf("MaxParallel = %d, want 2", orch.config.MaxParallel)
		}
	})

	t.Run("GetEnabledAgents", func(t *testing.T) {
		orch := NewOrchestrator(config, "")
		enabled := orch.GetEnabledAgents()
		
		if len(enabled) != 3 {
			t.Errorf("GetEnabledAgents() returned %d agents, want 3", len(enabled))
		}
		
		// Check sorting by priority
		if enabled[0].ID != "impl-1" {
			t.Errorf("First agent should be impl-1 (highest priority), got %s", enabled[0].ID)
		}
	})

	t.Run("GetAgentsByRole", func(t *testing.T) {
		orch := NewOrchestrator(config, "")
		implementers := orch.GetAgentsByRole(RoleImplementer)
		
		if len(implementers) != 1 { // disabled one shouldn't be included
			t.Errorf("GetAgentsByRole(Implementer) returned %d agents, want 1", len(implementers))
		}
	})

	t.Run("GetAgentStatus", func(t *testing.T) {
		orch := NewOrchestrator(config, "")
		status := orch.GetAgentStatus("impl-1")
		
		if status != StatusIdle {
			t.Errorf("Initial status should be Idle, got %s", status)
		}
	})
}

func TestOrchestratorExecution(t *testing.T) {
	config := &MultiAgentConfig{
		Agents: []AgentConfig{
			{ID: "impl-1", Role: RoleImplementer, Command: "test-cmd", Priority: 10, Enabled: true},
			{ID: "test-1", Role: RoleTester, Command: "test-cmd", Priority: 8, Enabled: true},
		},
		MaxParallel:        2,
		ConflictResolution: "priority",
	}

	t.Run("ExecuteWorkflow", func(t *testing.T) {
		tmpDir := t.TempDir()
		contextPath := filepath.Join(tmpDir, "context.json")
		
		orch := NewOrchestrator(config, contextPath)
		mock := NewMockExecutor()
		mock.SetResult("impl-1", "Implementation complete\n\nSuggestions:\n- Add error handling")
		mock.SetResult("test-1", "All tests passed")
		orch.SetExecutor(mock)

		ctx := context.Background()
		result, err := orch.ExecuteWorkflow(ctx, 1, "Test feature", 1, "Implement feature X")
		
		if err != nil {
			t.Fatalf("ExecuteWorkflow error = %v", err)
		}
		if !result.Success {
			t.Errorf("Workflow should have succeeded")
		}
		if len(result.Stages) < 1 {
			t.Errorf("Should have at least 1 stage")
		}
	})

	t.Run("ExecuteParallel", func(t *testing.T) {
		tmpDir := t.TempDir()
		contextPath := filepath.Join(tmpDir, "context.json")
		
		orch := NewOrchestrator(config, contextPath)
		mock := NewMockExecutor()
		mock.SetResult("impl-1", "Output 1")
		mock.SetResult("test-1", "Output 2")
		// Add small delay to test parallel execution
		mock.SetDelay("impl-1", 50*time.Millisecond)
		mock.SetDelay("test-1", 50*time.Millisecond)
		orch.SetExecutor(mock)

		start := time.Now()
		results := orch.executeParallel(context.Background(), orch.GetEnabledAgents(), "test prompt")
		elapsed := time.Since(start)

		if len(results) != 2 {
			t.Errorf("Should have 2 results, got %d", len(results))
		}
		
		// With MaxParallel=2, both agents should run in parallel
		// Total time should be ~50ms, not ~100ms
		if elapsed > 150*time.Millisecond {
			t.Logf("Elapsed time was %v, expected ~50ms for parallel execution", elapsed)
		}
	})
}

func TestExtractSuggestions(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int
	}{
		{
			name: "Basic suggestions",
			output: `
Some text here.

Suggestions:
- Add error handling
- Improve performance
- Add tests

More text.`,
			expected: 3,
		},
		{
			name: "Recommendations",
			output: `
I recommend the following:
* Use dependency injection
* Add logging

Done.`,
			expected: 2,
		},
		{
			name:     "No suggestions",
			output:   "Everything looks good.",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := extractSuggestions(tt.output)
			if len(suggestions) != tt.expected {
				t.Errorf("extractSuggestions() returned %d suggestions, want %d", len(suggestions), tt.expected)
			}
		})
	}
}

func TestExtractIssues(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int
	}{
		{
			name: "Issues found",
			output: `
Code review complete.

Issues:
- Missing error handling
- No input validation

Please fix these.`,
			expected: 2,
		},
		{
			name: "Problems found",
			output: `
Problems detected:
• Memory leak in function X
• Race condition in module Y`,
			expected: 2,
		},
		{
			name:     "No issues",
			output:   "Code looks great!",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := extractIssues(tt.output)
			if len(issues) != tt.expected {
				t.Errorf("extractIssues() returned %d issues, want %d", len(issues), tt.expected)
			}
		})
	}
}

func TestIsApproved(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{"Approved", "LGTM! Ship it.", true},
		{"Looks good", "This looks good to me.", true},
		{"Ready to merge", "Code is ready to merge.", true},
		{"Rejected", "Rejected - needs more work.", false},
		{"Needs work", "This needs work before merging.", false},
		{"Request changes", "I request changes to the implementation.", false},
		{"Default approved", "The code is fine.", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isApproved(tt.output)
			if result != tt.expected {
				t.Errorf("isApproved(%q) = %v, want %v", tt.output, result, tt.expected)
			}
		})
	}
}

func TestConflictDetection(t *testing.T) {
	t.Run("Approval conflict", func(t *testing.T) {
		results := []AgentResult{
			{AgentID: "reviewer-1", Role: RoleReviewer, Approved: true},
			{AgentID: "reviewer-2", Role: RoleReviewer, Approved: false},
		}

		conflicts := detectConflicts(results)
		if len(conflicts) != 1 {
			t.Errorf("Should detect 1 conflict, got %d", len(conflicts))
		}
		if len(conflicts) > 0 && conflicts[0].Type != "approval" {
			t.Errorf("Conflict type should be 'approval', got %s", conflicts[0].Type)
		}
	})

	t.Run("No conflict", func(t *testing.T) {
		results := []AgentResult{
			{AgentID: "reviewer-1", Role: RoleReviewer, Approved: true},
			{AgentID: "reviewer-2", Role: RoleReviewer, Approved: true},
		}

		conflicts := detectConflicts(results)
		if len(conflicts) != 0 {
			t.Errorf("Should detect 0 conflicts, got %d", len(conflicts))
		}
	})
}

func TestConflictResolution(t *testing.T) {
	config := &MultiAgentConfig{
		Agents: []AgentConfig{
			{ID: "agent-1", Role: RoleReviewer, Command: "cmd", Priority: 10, Enabled: true},
			{ID: "agent-2", Role: RoleReviewer, Command: "cmd", Priority: 5, Enabled: true},
		},
		ConflictResolution: "priority",
	}

	t.Run("ResolveByPriority", func(t *testing.T) {
		orch := NewOrchestrator(config, "")
		results := []AgentResult{
			{AgentID: "agent-1", Role: RoleReviewer, Status: StatusComplete, Approved: true},
			{AgentID: "agent-2", Role: RoleReviewer, Status: StatusComplete, Approved: false},
		}
		conflicts := []Conflict{{Type: "approval"}}

		winning := orch.resolveByPriority(results, conflicts)
		if len(winning) != 2 {
			t.Errorf("Should return both results, got %d", len(winning))
		}
		// First should be highest priority
		if winning[0].AgentID != "agent-1" {
			t.Errorf("First result should be from agent-1 (highest priority), got %s", winning[0].AgentID)
		}
	})

	t.Run("ResolveByMerge", func(t *testing.T) {
		orch := NewOrchestrator(config, "")
		results := []AgentResult{
			{AgentID: "agent-1", Status: StatusComplete, Suggestions: []string{"A", "B"}},
			{AgentID: "agent-2", Status: StatusComplete, Suggestions: []string{"B", "C"}},
		}

		winning := orch.resolveByMerge(results, nil)
		if len(winning) != 2 {
			t.Errorf("Should return both results, got %d", len(winning))
		}
		// Check merged suggestions (deduplicated)
		if len(winning[0].Suggestions) != 3 {
			t.Errorf("Merged suggestions should have 3 items, got %d", len(winning[0].Suggestions))
		}
	})
}

func TestLoadMultiAgentConfig(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "agents.json")
		
		config := MultiAgentConfig{
			Agents: []AgentConfig{
				{ID: "impl-1", Role: RoleImplementer, Command: "cursor-agent", Enabled: true},
			},
			MaxParallel: 2,
		}
		
		data, _ := json.Marshal(config)
		os.WriteFile(configPath, data, 0644)
		
		loaded, err := LoadMultiAgentConfig(configPath)
		if err != nil {
			t.Fatalf("LoadMultiAgentConfig error = %v", err)
		}
		
		if len(loaded.Agents) != 1 {
			t.Errorf("Should have 1 agent, got %d", len(loaded.Agents))
		}
	})

	t.Run("Invalid config - no agents", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "agents.json")
		
		config := MultiAgentConfig{
			Agents: []AgentConfig{},
		}
		
		data, _ := json.Marshal(config)
		os.WriteFile(configPath, data, 0644)
		
		_, err := LoadMultiAgentConfig(configPath)
		if err == nil {
			t.Error("Should error on empty agents list")
		}
	})

	t.Run("Invalid config - duplicate ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "agents.json")
		
		config := MultiAgentConfig{
			Agents: []AgentConfig{
				{ID: "agent-1", Role: RoleImplementer, Command: "cmd", Enabled: true},
				{ID: "agent-1", Role: RoleTester, Command: "cmd", Enabled: true},
			},
		}
		
		data, _ := json.Marshal(config)
		os.WriteFile(configPath, data, 0644)
		
		_, err := LoadMultiAgentConfig(configPath)
		if err == nil {
			t.Error("Should error on duplicate agent ID")
		}
	})

	t.Run("Invalid config - invalid role", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "agents.json")
		
		configJSON := `{
			"agents": [
				{"id": "agent-1", "role": "invalid_role", "command": "cmd", "enabled": true}
			]
		}`
		
		os.WriteFile(configPath, []byte(configJSON), 0644)
		
		_, err := LoadMultiAgentConfig(configPath)
		if err == nil {
			t.Error("Should error on invalid role")
		}
	})

	t.Run("File not found", func(t *testing.T) {
		_, err := LoadMultiAgentConfig("/nonexistent/path/agents.json")
		if err == nil {
			t.Error("Should error on file not found")
		}
	})
}

func TestWorkflowResultSummary(t *testing.T) {
	result := &WorkflowResult{
		FeatureID:   1,
		FeatureDesc: "Test Feature",
		Iteration:   2,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(30 * time.Second),
		Success:     true,
		Stages: []StageResult{
			{
				Name:    "implementation",
				Success: true,
				Results: []AgentResult{
					{AgentID: "impl-1", Role: RoleImplementer, Status: StatusComplete},
				},
			},
		},
	}

	summary := result.Summary()
	
	if summary == "" {
		t.Error("Summary should not be empty")
	}
	
	// Check that summary contains key information
	if !containsSubstring(summary, "Feature #1") {
		t.Error("Summary should contain feature ID")
	}
	if !containsSubstring(summary, "SUCCESS") {
		t.Error("Summary should contain status")
	}
	if !containsSubstring(summary, "implementation") {
		t.Error("Summary should contain stage name")
	}
}

func TestHealthStatus(t *testing.T) {
	config := &MultiAgentConfig{
		Agents: []AgentConfig{
			{ID: "agent-1", Role: RoleImplementer, Command: "cmd", Enabled: true, Timeout: 1 * time.Second},
			{ID: "agent-2", Role: RoleTester, Command: "cmd", Enabled: true},
		},
	}

	orch := NewOrchestrator(config, "")
	
	health := orch.GetHealthStatus()
	
	if len(health) != 2 {
		t.Errorf("Should have 2 health entries, got %d", len(health))
	}
	
	// All agents should be healthy initially
	for id, info := range health {
		if !info.Healthy {
			t.Errorf("Agent %s should be healthy initially", id)
		}
	}
}

func TestDeduplicateStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected int
	}{
		{"With duplicates", []string{"a", "b", "a", "c", "b"}, 3},
		{"No duplicates", []string{"a", "b", "c"}, 3},
		{"Empty", []string{}, 0},
		{"All same", []string{"a", "a", "a"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateStrings(tt.input)
			if len(result) != tt.expected {
				t.Errorf("deduplicateStrings() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsSubstring(s[1:], substr)))
}
