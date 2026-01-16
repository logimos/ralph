package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	t.Run("default path", func(t *testing.T) {
		store := NewStore("")
		if store.path != DefaultMemoryFile {
			t.Errorf("expected default path %q, got %q", DefaultMemoryFile, store.path)
		}
	})

	t.Run("custom path", func(t *testing.T) {
		store := NewStore("custom.json")
		if store.path != "custom.json" {
			t.Errorf("expected path %q, got %q", "custom.json", store.path)
		}
	})

	t.Run("default retention", func(t *testing.T) {
		store := NewStore("")
		if store.retentionDays != DefaultRetentionDays {
			t.Errorf("expected retention %d, got %d", DefaultRetentionDays, store.retentionDays)
		}
	})
}

func TestStore_SetRetentionDays(t *testing.T) {
	store := NewStore("")

	// Valid retention
	store.SetRetentionDays(30)
	if store.retentionDays != 30 {
		t.Errorf("expected retention 30, got %d", store.retentionDays)
	}

	// Zero retention should not change
	store.SetRetentionDays(0)
	if store.retentionDays != 30 {
		t.Errorf("expected retention to remain 30, got %d", store.retentionDays)
	}

	// Negative retention should not change
	store.SetRetentionDays(-1)
	if store.retentionDays != 30 {
		t.Errorf("expected retention to remain 30, got %d", store.retentionDays)
	}
}

func TestStore_LoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	t.Run("load nonexistent file creates empty memory", func(t *testing.T) {
		store := NewStore(memFile)
		err := store.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if store.memory == nil {
			t.Fatal("memory should not be nil")
		}
		if len(store.memory.Entries) != 0 {
			t.Errorf("expected empty entries, got %d", len(store.memory.Entries))
		}
	})

	t.Run("save and load roundtrip", func(t *testing.T) {
		store := NewStore(memFile)
		store.Load()

		// Add an entry
		_, err := store.Add(EntryTypeDecision, "Use PostgreSQL", "infra", "user")
		if err != nil {
			t.Fatalf("failed to add entry: %v", err)
		}

		// Create new store and load
		store2 := NewStore(memFile)
		err = store2.Load()
		if err != nil {
			t.Fatalf("failed to load: %v", err)
		}

		if len(store2.memory.Entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(store2.memory.Entries))
		}

		entry := store2.memory.Entries[0]
		if entry.Type != EntryTypeDecision {
			t.Errorf("expected type decision, got %s", entry.Type)
		}
		if entry.Content != "Use PostgreSQL" {
			t.Errorf("expected content 'Use PostgreSQL', got %q", entry.Content)
		}
		if entry.Category != "infra" {
			t.Errorf("expected category 'infra', got %q", entry.Category)
		}
	})

	t.Run("load invalid JSON returns error", func(t *testing.T) {
		badFile := filepath.Join(tmpDir, "bad.json")
		os.WriteFile(badFile, []byte("not valid json"), 0644)

		store := NewStore(badFile)
		err := store.Load()
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestStore_Add(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	store := NewStore(memFile)
	store.Load()

	entry, err := store.Add(EntryTypeConvention, "Use snake_case", "db", "user")
	if err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	if entry == nil {
		t.Fatal("entry should not be nil")
	}

	if !strings.HasPrefix(entry.ID, "mem_") {
		t.Errorf("expected ID to start with 'mem_', got %q", entry.ID)
	}

	if entry.Type != EntryTypeConvention {
		t.Errorf("expected type convention, got %s", entry.Type)
	}

	if entry.Content != "Use snake_case" {
		t.Errorf("expected content 'Use snake_case', got %q", entry.Content)
	}

	if entry.Category != "db" {
		t.Errorf("expected category 'db', got %q", entry.Category)
	}

	if entry.Source != "user" {
		t.Errorf("expected source 'user', got %q", entry.Source)
	}

	// Check entry was persisted
	if store.Count() != 1 {
		t.Errorf("expected count 1, got %d", store.Count())
	}
}

func TestStore_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	store := NewStore(memFile)
	store.Load()

	// Add entries
	store.Add(EntryTypeDecision, "Decision 1", "", "user")
	store.Add(EntryTypeContext, "Context 1", "", "agent")

	if store.Count() != 2 {
		t.Fatalf("expected 2 entries, got %d", store.Count())
	}

	// Clear
	err := store.Clear()
	if err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("expected 0 entries after clear, got %d", store.Count())
	}

	// Verify persistence
	store2 := NewStore(memFile)
	store2.Load()
	if store2.Count() != 0 {
		t.Errorf("expected 0 entries after reload, got %d", store2.Count())
	}
}

func TestStore_GetByType(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	store := NewStore(memFile)
	store.Load()

	store.Add(EntryTypeDecision, "Decision 1", "", "user")
	store.Add(EntryTypeDecision, "Decision 2", "", "user")
	store.Add(EntryTypeConvention, "Convention 1", "", "user")
	store.Add(EntryTypeContext, "Context 1", "", "agent")

	decisions := store.GetByType(EntryTypeDecision)
	if len(decisions) != 2 {
		t.Errorf("expected 2 decisions, got %d", len(decisions))
	}

	conventions := store.GetByType(EntryTypeConvention)
	if len(conventions) != 1 {
		t.Errorf("expected 1 convention, got %d", len(conventions))
	}

	tradeoffs := store.GetByType(EntryTypeTradeoff)
	if len(tradeoffs) != 0 {
		t.Errorf("expected 0 tradeoffs, got %d", len(tradeoffs))
	}
}

func TestStore_GetByCategory(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	store := NewStore(memFile)
	store.Load()

	store.Add(EntryTypeDecision, "Infra Decision", "infra", "user")
	store.Add(EntryTypeConvention, "UI Convention", "ui", "user")
	store.Add(EntryTypeContext, "General Context", "", "agent")

	infraEntries := store.GetByCategory("infra")
	// Should include infra + entries with no category
	if len(infraEntries) != 2 {
		t.Errorf("expected 2 infra entries (including general), got %d", len(infraEntries))
	}

	uiEntries := store.GetByCategory("UI") // Case insensitive
	if len(uiEntries) != 2 {
		t.Errorf("expected 2 UI entries (including general), got %d", len(uiEntries))
	}
}

func TestStore_GetRelevant(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	store := NewStore(memFile)
	store.Load()

	// Add various entries
	store.Add(EntryTypeDecision, "Infra Decision", "infra", "user")
	store.Add(EntryTypeConvention, "Infra Convention", "infra", "user")
	store.Add(EntryTypeContext, "UI Context", "ui", "agent")
	store.Add(EntryTypeTradeoff, "General Tradeoff", "", "user")

	// Get relevant entries for infra
	relevant := store.GetRelevant("infra", 10)
	if len(relevant) != 4 {
		t.Errorf("expected 4 entries, got %d", len(relevant))
	}

	// Infra entries should be first (higher relevance score)
	if relevant[0].Category != "infra" && relevant[1].Category != "infra" {
		t.Error("expected infra entries to be ranked higher")
	}

	// Test max entries limit
	limited := store.GetRelevant("infra", 2)
	if len(limited) != 2 {
		t.Errorf("expected 2 entries with limit, got %d", len(limited))
	}
}

func TestStore_Prune(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	store := NewStore(memFile)
	store.SetRetentionDays(7)
	store.Load()

	// Add a fresh entry
	store.Add(EntryTypeDecision, "Fresh", "", "user")

	// Manually add an old entry
	oldEntry := Entry{
		ID:        "old_entry",
		Type:      EntryTypeContext,
		Content:   "Old content",
		CreatedAt: time.Now().AddDate(0, 0, -30),
		UpdatedAt: time.Now().AddDate(0, 0, -30),
	}
	store.memory.Entries = append(store.memory.Entries, oldEntry)
	store.Save()

	if store.Count() != 2 {
		t.Fatalf("expected 2 entries before prune, got %d", store.Count())
	}

	// Prune
	removed, err := store.Prune()
	if err != nil {
		t.Fatalf("prune failed: %v", err)
	}

	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	if store.Count() != 1 {
		t.Errorf("expected 1 entry after prune, got %d", store.Count())
	}
}

func TestStore_Summary(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	t.Run("empty memory", func(t *testing.T) {
		store := NewStore(memFile)
		store.Load()
		summary := store.Summary()
		if summary != "No memories stored" {
			t.Errorf("expected 'No memories stored', got %q", summary)
		}
	})

	t.Run("with entries", func(t *testing.T) {
		store := NewStore(memFile)
		store.Load()
		store.Add(EntryTypeDecision, "Use PostgreSQL", "infra", "user")
		store.Add(EntryTypeConvention, "Use snake_case", "db", "agent")

		summary := store.Summary()
		if !strings.Contains(summary, "2 entries") {
			t.Error("summary should contain entry count")
		}
		if !strings.Contains(summary, "Decision") {
			t.Error("summary should contain Decision section")
		}
		if !strings.Contains(summary, "Convention") {
			t.Error("summary should contain Convention section")
		}
		if !strings.Contains(summary, "Use PostgreSQL") {
			t.Error("summary should contain entry content")
		}
	})
}

func TestExtractFromOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int
		types    []EntryType
	}{
		{
			name:     "no markers",
			output:   "Just some regular output without markers",
			expected: 0,
			types:    nil,
		},
		{
			name:     "single decision marker",
			output:   "Some text [REMEMBER:DECISION]Use PostgreSQL for persistence[/REMEMBER] more text",
			expected: 1,
			types:    []EntryType{EntryTypeDecision},
		},
		{
			name:     "multiple markers",
			output:   "[REMEMBER:DECISION]Decision 1[/REMEMBER] text [REMEMBER:CONVENTION]Convention 1[/REMEMBER]",
			expected: 2,
			types:    []EntryType{EntryTypeDecision, EntryTypeConvention},
		},
		{
			name:     "multiline content",
			output:   "[REMEMBER:CONTEXT]\nThis is a multiline\ncontext entry\n[/REMEMBER]",
			expected: 1,
			types:    []EntryType{EntryTypeContext},
		},
		{
			name:     "tradeoff marker",
			output:   "[REMEMBER:TRADEOFF]Sacrificed type safety for performance[/REMEMBER]",
			expected: 1,
			types:    []EntryType{EntryTypeTradeoff},
		},
		{
			name:     "empty content ignored",
			output:   "[REMEMBER:DECISION][/REMEMBER]",
			expected: 0,
			types:    nil,
		},
		{
			name:     "whitespace only content ignored",
			output:   "[REMEMBER:DECISION]   [/REMEMBER]",
			expected: 0,
			types:    nil,
		},
		{
			name:     "unknown type defaults to context",
			output:   "[REMEMBER:UNKNOWN]Some content[/REMEMBER]",
			expected: 1,
			types:    []EntryType{EntryTypeContext},
		},
		{
			name:     "lowercase type",
			output:   "[REMEMBER:decision]Lowercase decision[/REMEMBER]",
			expected: 0, // Pattern requires uppercase
			types:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := ExtractFromOutput(tt.output)
			if len(entries) != tt.expected {
				t.Errorf("expected %d entries, got %d", tt.expected, len(entries))
			}

			for i, expectedType := range tt.types {
				if i < len(entries) && entries[i].Type != expectedType {
					t.Errorf("entry %d: expected type %s, got %s", i, expectedType, entries[i].Type)
				}
			}

			// Verify all entries have required fields
			for i, e := range entries {
				if e.ID == "" {
					t.Errorf("entry %d: ID should not be empty", i)
				}
				if e.Content == "" {
					t.Errorf("entry %d: Content should not be empty", i)
				}
				if e.Source != "agent" {
					t.Errorf("entry %d: Source should be 'agent', got %q", i, e.Source)
				}
			}
		})
	}
}

func TestStore_BuildPromptContext(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	t.Run("empty memory returns empty string", func(t *testing.T) {
		store := NewStore(memFile)
		store.Load()

		context := store.BuildPromptContext("infra", 10)
		if context != "" {
			t.Errorf("expected empty context, got %q", context)
		}
	})

	t.Run("with entries includes content", func(t *testing.T) {
		store := NewStore(memFile)
		store.Load()
		store.Add(EntryTypeDecision, "Use PostgreSQL", "infra", "user")
		store.Add(EntryTypeConvention, "Use snake_case", "db", "agent")

		context := store.BuildPromptContext("infra", 10)

		if !strings.Contains(context, "MEMORY CONTEXT") {
			t.Error("context should contain MEMORY CONTEXT header")
		}
		if !strings.Contains(context, "Use PostgreSQL") {
			t.Error("context should contain entry content")
		}
		if !strings.Contains(context, "[DECISION]") {
			t.Error("context should contain type label")
		}
		if !strings.Contains(context, "[REMEMBER:") {
			t.Error("context should contain REMEMBER marker instructions")
		}
	})

	t.Run("respects max entries", func(t *testing.T) {
		tmpDir2 := t.TempDir()
		memFile2 := filepath.Join(tmpDir2, "test-memory-max.json")
		store := NewStore(memFile2)
		store.Load()

		// Add many entries
		for i := 0; i < 20; i++ {
			store.Add(EntryTypeContext, "Entry content", "", "user")
		}

		context := store.BuildPromptContext("", 5)
		entryCount := strings.Count(context, "[CONTEXT]")
		if entryCount != 5 {
			t.Errorf("expected 5 entries in context, got %d", entryCount)
		}
	})
}

func TestParseEntryType(t *testing.T) {
	tests := []struct {
		input    string
		expected EntryType
		wantErr  bool
	}{
		{"decision", EntryTypeDecision, false},
		{"DECISION", EntryTypeDecision, false},
		{"Decision", EntryTypeDecision, false},
		{"convention", EntryTypeConvention, false},
		{"tradeoff", EntryTypeTradeoff, false},
		{"context", EntryTypeContext, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseEntryType(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestValidEntryTypes(t *testing.T) {
	types := ValidEntryTypes()
	if len(types) != 4 {
		t.Errorf("expected 4 valid types, got %d", len(types))
	}

	expected := map[string]bool{
		"decision":   true,
		"convention": true,
		"tradeoff":   true,
		"context":    true,
	}

	for _, typ := range types {
		if !expected[typ] {
			t.Errorf("unexpected type %q", typ)
		}
	}
}

func TestCalculateRelevanceScore(t *testing.T) {
	now := time.Now()
	oldTime := now.AddDate(0, 0, -30)

	tests := []struct {
		name     string
		entry    Entry
		category string
		minScore int
	}{
		{
			name: "decision with matching category",
			entry: Entry{
				Type:      EntryTypeDecision,
				Category:  "infra",
				UpdatedAt: now,
			},
			category: "infra",
			minScore: 8, // base(3) + category(5) + recency(2) = 10
		},
		{
			name: "context without category match",
			entry: Entry{
				Type:      EntryTypeContext,
				Category:  "ui",
				UpdatedAt: oldTime,
			},
			category: "infra",
			minScore: 1, // base(1) only
		},
		{
			name: "convention recent",
			entry: Entry{
				Type:      EntryTypeConvention,
				Category:  "",
				UpdatedAt: now,
			},
			category: "any",
			minScore: 5, // base(3) + recency(2) = 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateRelevanceScore(tt.entry, tt.category)
			if score < tt.minScore {
				t.Errorf("expected score >= %d, got %d", tt.minScore, score)
			}
		})
	}
}

func TestStore_Count(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	store := NewStore(memFile)

	// Before load
	if store.Count() != 0 {
		t.Errorf("expected count 0 before load, got %d", store.Count())
	}

	store.Load()

	// After load (empty)
	if store.Count() != 0 {
		t.Errorf("expected count 0 after load, got %d", store.Count())
	}

	// After adding
	store.Add(EntryTypeDecision, "Test", "", "user")
	if store.Count() != 1 {
		t.Errorf("expected count 1 after add, got %d", store.Count())
	}

	store.Add(EntryTypeContext, "Test 2", "", "agent")
	if store.Count() != 2 {
		t.Errorf("expected count 2 after second add, got %d", store.Count())
	}
}

func TestStore_GetAll(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "test-memory.json")

	store := NewStore(memFile)

	// Before load
	entries := store.GetAll()
	if len(entries) != 0 {
		t.Errorf("expected empty before load, got %d", len(entries))
	}

	store.Load()
	store.Add(EntryTypeDecision, "Decision 1", "", "user")
	store.Add(EntryTypeConvention, "Convention 1", "", "user")

	entries = store.GetAll()
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestStore_SaveWithDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "memory.json")

	store := NewStore(nestedPath)
	store.Load()
	store.Add(EntryTypeContext, "Test", "", "user")

	err := store.Save()
	if err != nil {
		t.Fatalf("failed to save to nested path: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("memory file should exist at nested path")
	}
}
