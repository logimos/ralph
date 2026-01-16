// Package memory provides persistent memory management for cross-session continuity.
// It allows Ralph to remember architectural decisions, coding conventions, tradeoffs,
// and context across multiple runs, reducing repetitive guidance and maintaining consistency.
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	// DefaultMemoryFile is the default memory file name
	DefaultMemoryFile = ".ralph-memory.json"

	// MemoryMarkerPattern is the regex pattern for extracting memories from agent output
	MemoryMarkerPattern = `\[REMEMBER:([A-Z]+)\](.*?)\[/REMEMBER\]`

	// DefaultRetentionDays is the default number of days to retain memories
	DefaultRetentionDays = 90
)

// EntryType represents the type of memory entry
type EntryType string

const (
	// EntryTypeDecision represents an architectural choice
	EntryTypeDecision EntryType = "decision"
	// EntryTypeConvention represents a coding standard or pattern
	EntryTypeConvention EntryType = "convention"
	// EntryTypeTradeoff represents an accepted compromise
	EntryTypeTradeoff EntryType = "tradeoff"
	// EntryTypeContext represents project-specific knowledge
	EntryTypeContext EntryType = "context"
)

// Entry represents a single memory entry
type Entry struct {
	ID        string    `json:"id"`
	Type      EntryType `json:"type"`
	Content   string    `json:"content"`
	Category  string    `json:"category,omitempty"`  // Related feature category (e.g., "infra", "ui")
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Source    string    `json:"source,omitempty"` // "agent", "user", or feature ID
}

// Memory represents the complete memory state
type Memory struct {
	Entries       []Entry `json:"entries"`
	LastUpdated   time.Time `json:"last_updated"`
	RetentionDays int       `json:"retention_days,omitempty"`
}

// Store handles memory persistence and operations
type Store struct {
	path          string
	memory        *Memory
	retentionDays int
}

// NewStore creates a new memory store for the given path
func NewStore(path string) *Store {
	if path == "" {
		path = DefaultMemoryFile
	}
	return &Store{
		path:          path,
		retentionDays: DefaultRetentionDays,
	}
}

// SetRetentionDays sets the number of days to retain memories
func (s *Store) SetRetentionDays(days int) {
	if days > 0 {
		s.retentionDays = days
	}
}

// Load reads the memory file from disk
func (s *Store) Load() error {
	// Initialize empty memory if file doesn't exist
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		s.memory = &Memory{
			Entries:       []Entry{},
			LastUpdated:   time.Now(),
			RetentionDays: s.retentionDays,
		}
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("failed to read memory file: %w", err)
	}

	var mem Memory
	if err := json.Unmarshal(data, &mem); err != nil {
		return fmt.Errorf("failed to parse memory file: %w", err)
	}

	s.memory = &mem
	if s.memory.RetentionDays > 0 {
		s.retentionDays = s.memory.RetentionDays
	}

	return nil
}

// Save writes the memory to disk
func (s *Store) Save() error {
	if s.memory == nil {
		s.memory = &Memory{
			Entries:       []Entry{},
			LastUpdated:   time.Now(),
			RetentionDays: s.retentionDays,
		}
	}

	s.memory.LastUpdated = time.Now()
	s.memory.RetentionDays = s.retentionDays

	data, err := json.MarshalIndent(s.memory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}

	return nil
}

// Clear removes all entries from memory and saves
func (s *Store) Clear() error {
	s.memory = &Memory{
		Entries:       []Entry{},
		LastUpdated:   time.Now(),
		RetentionDays: s.retentionDays,
	}
	return s.Save()
}

// Add adds a new memory entry
func (s *Store) Add(entryType EntryType, content, category, source string) (*Entry, error) {
	if s.memory == nil {
		if err := s.Load(); err != nil {
			return nil, err
		}
	}

	entry := Entry{
		ID:        generateID(),
		Type:      entryType,
		Content:   strings.TrimSpace(content),
		Category:  category,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Source:    source,
	}

	s.memory.Entries = append(s.memory.Entries, entry)

	if err := s.Save(); err != nil {
		return nil, err
	}

	return &entry, nil
}

// GetAll returns all memory entries
func (s *Store) GetAll() []Entry {
	if s.memory == nil {
		return []Entry{}
	}
	return s.memory.Entries
}

// GetByType returns entries of a specific type
func (s *Store) GetByType(entryType EntryType) []Entry {
	if s.memory == nil {
		return []Entry{}
	}

	var entries []Entry
	for _, e := range s.memory.Entries {
		if e.Type == entryType {
			entries = append(entries, e)
		}
	}
	return entries
}

// GetByCategory returns entries related to a specific category
func (s *Store) GetByCategory(category string) []Entry {
	if s.memory == nil {
		return []Entry{}
	}

	var entries []Entry
	categoryLower := strings.ToLower(category)
	for _, e := range s.memory.Entries {
		if strings.ToLower(e.Category) == categoryLower || e.Category == "" {
			entries = append(entries, e)
		}
	}
	return entries
}

// GetRelevant returns entries relevant to the given category, sorted by relevance
func (s *Store) GetRelevant(category string, maxEntries int) []Entry {
	if s.memory == nil || len(s.memory.Entries) == 0 {
		return []Entry{}
	}

	type scoredEntry struct {
		entry Entry
		score int
	}

	var scored []scoredEntry
	categoryLower := strings.ToLower(category)

	for _, e := range s.memory.Entries {
		score := calculateRelevanceScore(e, categoryLower)
		scored = append(scored, scoredEntry{entry: e, score: score})
	}

	// Sort by score (descending) then by updated time (most recent first)
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].entry.UpdatedAt.After(scored[j].entry.UpdatedAt)
	})

	// Return top entries
	var result []Entry
	for i, se := range scored {
		if maxEntries > 0 && i >= maxEntries {
			break
		}
		result = append(result, se.entry)
	}

	return result
}

// Prune removes entries older than the retention period
func (s *Store) Prune() (int, error) {
	if s.memory == nil {
		return 0, nil
	}

	cutoff := time.Now().AddDate(0, 0, -s.retentionDays)
	originalCount := len(s.memory.Entries)

	var retained []Entry
	for _, e := range s.memory.Entries {
		if e.UpdatedAt.After(cutoff) {
			retained = append(retained, e)
		}
	}

	removed := originalCount - len(retained)
	if removed > 0 {
		s.memory.Entries = retained
		if err := s.Save(); err != nil {
			return 0, err
		}
	}

	return removed, nil
}

// Count returns the total number of memory entries
func (s *Store) Count() int {
	if s.memory == nil {
		return 0
	}
	return len(s.memory.Entries)
}

// Summary returns a formatted summary of all memories
func (s *Store) Summary() string {
	if s.memory == nil || len(s.memory.Entries) == 0 {
		return "No memories stored"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Memory Store: %d entries (retention: %d days)\n", len(s.memory.Entries), s.retentionDays))
	b.WriteString(fmt.Sprintf("Last updated: %s\n\n", s.memory.LastUpdated.Format(time.RFC3339)))

	// Group by type
	typeGroups := make(map[EntryType][]Entry)
	for _, e := range s.memory.Entries {
		typeGroups[e.Type] = append(typeGroups[e.Type], e)
	}

	typeOrder := []EntryType{EntryTypeDecision, EntryTypeConvention, EntryTypeTradeoff, EntryTypeContext}
	for _, t := range typeOrder {
		entries := typeGroups[t]
		if len(entries) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("=== %s (%d) ===\n", strings.Title(string(t)), len(entries)))
		for _, e := range entries {
			categoryStr := ""
			if e.Category != "" {
				categoryStr = fmt.Sprintf(" [%s]", e.Category)
			}
			b.WriteString(fmt.Sprintf("  - %s%s\n", e.Content, categoryStr))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ExtractFromOutput parses agent output for [REMEMBER:TYPE]...[/REMEMBER] markers
// and returns the extracted entries without saving them
func ExtractFromOutput(output string) []Entry {
	re := regexp.MustCompile(`(?s)\[REMEMBER:([A-Z]+)\](.*?)\[/REMEMBER\]`)
	matches := re.FindAllStringSubmatch(output, -1)

	var entries []Entry
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		typeStr := strings.ToLower(match[1])
		content := strings.TrimSpace(match[2])

		if content == "" {
			continue
		}

		entryType := parseEntryType(typeStr)

		entries = append(entries, Entry{
			ID:        generateID(),
			Type:      entryType,
			Content:   content,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "agent",
		})
	}

	return entries
}

// BuildPromptContext creates a formatted string of memories to inject into prompts
func (s *Store) BuildPromptContext(category string, maxEntries int) string {
	entries := s.GetRelevant(category, maxEntries)
	if len(entries) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n[MEMORY CONTEXT - Previous decisions and conventions to follow:]\n")

	for _, e := range entries {
		typeLabel := strings.ToUpper(string(e.Type))
		b.WriteString(fmt.Sprintf("- [%s] %s\n", typeLabel, e.Content))
	}

	b.WriteString("[END MEMORY CONTEXT]\n\n")
	b.WriteString("If you make any important architectural decisions, conventions, or tradeoffs, ")
	b.WriteString("wrap them in [REMEMBER:TYPE]...[/REMEMBER] markers where TYPE is ")
	b.WriteString("DECISION, CONVENTION, TRADEOFF, or CONTEXT.\n")

	return b.String()
}

// parseEntryType converts a string to EntryType
func parseEntryType(s string) EntryType {
	switch strings.ToLower(s) {
	case "decision":
		return EntryTypeDecision
	case "convention":
		return EntryTypeConvention
	case "tradeoff":
		return EntryTypeTradeoff
	case "context":
		return EntryTypeContext
	default:
		return EntryTypeContext // Default to context for unknown types
	}
}

// ParseEntryType is the exported version for use by other packages
func ParseEntryType(s string) (EntryType, error) {
	switch strings.ToLower(s) {
	case "decision":
		return EntryTypeDecision, nil
	case "convention":
		return EntryTypeConvention, nil
	case "tradeoff":
		return EntryTypeTradeoff, nil
	case "context":
		return EntryTypeContext, nil
	default:
		return "", fmt.Errorf("invalid memory type: %s (must be decision, convention, tradeoff, or context)", s)
	}
}

// calculateRelevanceScore calculates how relevant an entry is to a category
func calculateRelevanceScore(entry Entry, category string) int {
	score := 0

	// Base score by type (decisions and conventions are generally more important)
	switch entry.Type {
	case EntryTypeDecision:
		score += 3
	case EntryTypeConvention:
		score += 3
	case EntryTypeTradeoff:
		score += 2
	case EntryTypeContext:
		score += 1
	}

	// Category match boost
	if entry.Category != "" && strings.ToLower(entry.Category) == category {
		score += 5
	}

	// Recency boost (entries updated in last 7 days get a boost)
	if time.Since(entry.UpdatedAt) < 7*24*time.Hour {
		score += 2
	}

	return score
}

// generateID creates a unique ID for a memory entry
func generateID() string {
	return fmt.Sprintf("mem_%d", time.Now().UnixNano())
}

// ValidEntryTypes returns all valid entry type strings
func ValidEntryTypes() []string {
	return []string{"decision", "convention", "tradeoff", "context"}
}
