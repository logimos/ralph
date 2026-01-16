// Package nudge provides a system for lightweight mid-run guidance during Ralph execution.
// Users can create/edit nudges.json during a run to steer Ralph without stopping,
// and nudges are incorporated into subsequent iterations.
package nudge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultNudgeFile is the default nudge file name
	DefaultNudgeFile = "nudges.json"
)

// NudgeType represents the type of nudge
type NudgeType string

const (
	// NudgeTypeFocus prioritizes a specific feature or approach
	NudgeTypeFocus NudgeType = "focus"
	// NudgeTypeSkip defers a feature or skips certain work
	NudgeTypeSkip NudgeType = "skip"
	// NudgeTypeConstraint adds a requirement or limitation
	NudgeTypeConstraint NudgeType = "constraint"
	// NudgeTypeStyle specifies coding style preferences
	NudgeTypeStyle NudgeType = "style"
)

// Nudge represents a single nudge entry that provides guidance to the AI agent
type Nudge struct {
	ID           string    `json:"id"`
	Type         NudgeType `json:"type"`
	Content      string    `json:"content"`
	Priority     int       `json:"priority,omitempty"` // Higher = more important (default 0)
	CreatedAt    time.Time `json:"created_at"`
	Acknowledged bool      `json:"acknowledged,omitempty"` // Set to true when processed
	AckedAt      time.Time `json:"acked_at,omitempty"`     // When it was acknowledged
}

// NudgeFile represents the complete nudges file structure
type NudgeFile struct {
	Nudges      []Nudge   `json:"nudges"`
	LastUpdated time.Time `json:"last_updated"`
}

// Store handles nudge persistence and operations
type Store struct {
	path        string
	nudgeFile   *NudgeFile
	lastModTime time.Time
	mu          sync.RWMutex
}

// NewStore creates a new nudge store for the given path
func NewStore(path string) *Store {
	if path == "" {
		path = DefaultNudgeFile
	}
	return &Store{
		path: path,
	}
}

// Path returns the path to the nudge file
func (s *Store) Path() string {
	return s.path
}

// Load reads the nudge file from disk
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize empty nudge file if it doesn't exist
	info, err := os.Stat(s.path)
	if os.IsNotExist(err) {
		s.nudgeFile = &NudgeFile{
			Nudges:      []Nudge{},
			LastUpdated: time.Now(),
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to stat nudge file: %w", err)
	}

	s.lastModTime = info.ModTime()

	data, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("failed to read nudge file: %w", err)
	}

	// Handle empty file
	if len(data) == 0 {
		s.nudgeFile = &NudgeFile{
			Nudges:      []Nudge{},
			LastUpdated: time.Now(),
		}
		return nil
	}

	var nf NudgeFile
	if err := json.Unmarshal(data, &nf); err != nil {
		return fmt.Errorf("failed to parse nudge file: %w", err)
	}

	s.nudgeFile = &nf
	return nil
}

// Save writes the nudges to disk
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveUnsafe()
}

// saveUnsafe writes without acquiring the lock (must be called with lock held)
func (s *Store) saveUnsafe() error {
	if s.nudgeFile == nil {
		s.nudgeFile = &NudgeFile{
			Nudges:      []Nudge{},
			LastUpdated: time.Now(),
		}
	}

	s.nudgeFile.LastUpdated = time.Now()

	data, err := json.MarshalIndent(s.nudgeFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal nudges: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write nudge file: %w", err)
	}

	// Update mod time after save
	if info, err := os.Stat(s.path); err == nil {
		s.lastModTime = info.ModTime()
	}

	return nil
}

// Clear removes all nudges from the store and saves
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nudgeFile = &NudgeFile{
		Nudges:      []Nudge{},
		LastUpdated: time.Now(),
	}
	return s.saveUnsafe()
}

// Add creates and saves a new nudge
func (s *Store) Add(nudgeType NudgeType, content string, priority int) (*Nudge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nudgeFile == nil {
		s.nudgeFile = &NudgeFile{
			Nudges:      []Nudge{},
			LastUpdated: time.Now(),
		}
	}

	nudge := Nudge{
		ID:        generateID(),
		Type:      nudgeType,
		Content:   strings.TrimSpace(content),
		Priority:  priority,
		CreatedAt: time.Now(),
	}

	s.nudgeFile.Nudges = append(s.nudgeFile.Nudges, nudge)

	if err := s.saveUnsafe(); err != nil {
		return nil, err
	}

	return &nudge, nil
}

// GetActive returns all non-acknowledged nudges sorted by priority (highest first)
func (s *Store) GetActive() []Nudge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.nudgeFile == nil {
		return []Nudge{}
	}

	var active []Nudge
	for _, n := range s.nudgeFile.Nudges {
		if !n.Acknowledged {
			active = append(active, n)
		}
	}

	// Sort by priority (descending) then by created time (oldest first)
	sort.Slice(active, func(i, j int) bool {
		if active[i].Priority != active[j].Priority {
			return active[i].Priority > active[j].Priority
		}
		return active[i].CreatedAt.Before(active[j].CreatedAt)
	})

	return active
}

// GetAll returns all nudges
func (s *Store) GetAll() []Nudge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.nudgeFile == nil {
		return []Nudge{}
	}
	return s.nudgeFile.Nudges
}

// GetByType returns nudges of a specific type (includes acknowledged)
func (s *Store) GetByType(nudgeType NudgeType) []Nudge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.nudgeFile == nil {
		return []Nudge{}
	}

	var result []Nudge
	for _, n := range s.nudgeFile.Nudges {
		if n.Type == nudgeType {
			result = append(result, n)
		}
	}
	return result
}

// Acknowledge marks a nudge as processed
func (s *Store) Acknowledge(nudgeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nudgeFile == nil {
		return fmt.Errorf("no nudge file loaded")
	}

	for i, n := range s.nudgeFile.Nudges {
		if n.ID == nudgeID {
			s.nudgeFile.Nudges[i].Acknowledged = true
			s.nudgeFile.Nudges[i].AckedAt = time.Now()
			return s.saveUnsafe()
		}
	}

	return fmt.Errorf("nudge not found: %s", nudgeID)
}

// AcknowledgeAll marks all active nudges as processed
func (s *Store) AcknowledgeAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nudgeFile == nil {
		return nil
	}

	now := time.Now()
	changed := false
	for i := range s.nudgeFile.Nudges {
		if !s.nudgeFile.Nudges[i].Acknowledged {
			s.nudgeFile.Nudges[i].Acknowledged = true
			s.nudgeFile.Nudges[i].AckedAt = now
			changed = true
		}
	}

	if changed {
		return s.saveUnsafe()
	}
	return nil
}

// HasChanged checks if the nudge file has been modified since last load
func (s *Store) HasChanged() bool {
	s.mu.RLock()
	lastMod := s.lastModTime
	s.mu.RUnlock()

	info, err := os.Stat(s.path)
	if err != nil {
		return false
	}

	return info.ModTime().After(lastMod)
}

// Reload reloads the nudge file if it has changed
func (s *Store) Reload() (bool, error) {
	if !s.HasChanged() {
		return false, nil
	}

	err := s.Load()
	return err == nil, err
}

// Count returns the total number of nudges
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.nudgeFile == nil {
		return 0
	}
	return len(s.nudgeFile.Nudges)
}

// ActiveCount returns the number of non-acknowledged nudges
func (s *Store) ActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.nudgeFile == nil {
		return 0
	}

	count := 0
	for _, n := range s.nudgeFile.Nudges {
		if !n.Acknowledged {
			count++
		}
	}
	return count
}

// BuildPromptContext creates a formatted string of nudges to inject into agent prompts
func (s *Store) BuildPromptContext() string {
	active := s.GetActive()
	if len(active) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n[USER GUIDANCE - Please follow these instructions carefully:]\n")

	// Group by type for clearer presentation
	typeOrder := []NudgeType{NudgeTypeFocus, NudgeTypeSkip, NudgeTypeConstraint, NudgeTypeStyle}
	typeLabels := map[NudgeType]string{
		NudgeTypeFocus:      "FOCUS",
		NudgeTypeSkip:       "SKIP",
		NudgeTypeConstraint: "CONSTRAINT",
		NudgeTypeStyle:      "STYLE",
	}

	for _, t := range typeOrder {
		var nudgesOfType []Nudge
		for _, n := range active {
			if n.Type == t {
				nudgesOfType = append(nudgesOfType, n)
			}
		}

		if len(nudgesOfType) == 0 {
			continue
		}

		for _, n := range nudgesOfType {
			label := typeLabels[n.Type]
			priorityStr := ""
			if n.Priority > 0 {
				priorityStr = fmt.Sprintf(" (priority: %d)", n.Priority)
			}
			b.WriteString(fmt.Sprintf("- [%s%s] %s\n", label, priorityStr, n.Content))
		}
	}

	b.WriteString("[END USER GUIDANCE]\n\n")

	return b.String()
}

// Summary returns a formatted summary of all nudges
func (s *Store) Summary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.nudgeFile == nil || len(s.nudgeFile.Nudges) == 0 {
		return "No nudges defined"
	}

	var b strings.Builder
	active := 0
	acknowledged := 0
	for _, n := range s.nudgeFile.Nudges {
		if n.Acknowledged {
			acknowledged++
		} else {
			active++
		}
	}

	b.WriteString(fmt.Sprintf("Nudge Store: %d total (%d active, %d acknowledged)\n",
		len(s.nudgeFile.Nudges), active, acknowledged))
	b.WriteString(fmt.Sprintf("File: %s\n", s.path))
	b.WriteString(fmt.Sprintf("Last updated: %s\n\n", s.nudgeFile.LastUpdated.Format(time.RFC3339)))

	// Group by type
	typeGroups := make(map[NudgeType][]Nudge)
	for _, n := range s.nudgeFile.Nudges {
		typeGroups[n.Type] = append(typeGroups[n.Type], n)
	}

	typeOrder := []NudgeType{NudgeTypeFocus, NudgeTypeSkip, NudgeTypeConstraint, NudgeTypeStyle}
	for _, t := range typeOrder {
		nudges := typeGroups[t]
		if len(nudges) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("=== %s ===\n", strings.ToUpper(string(t))))
		for _, n := range nudges {
			status := "[ ]"
			if n.Acknowledged {
				status = "[x]"
			}
			priorityStr := ""
			if n.Priority > 0 {
				priorityStr = fmt.Sprintf(" (p%d)", n.Priority)
			}
			b.WriteString(fmt.Sprintf("  %s%s %s\n", status, priorityStr, n.Content))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ParseNudgeType converts a string to NudgeType
func ParseNudgeType(s string) (NudgeType, error) {
	switch strings.ToLower(s) {
	case "focus":
		return NudgeTypeFocus, nil
	case "skip":
		return NudgeTypeSkip, nil
	case "constraint":
		return NudgeTypeConstraint, nil
	case "style":
		return NudgeTypeStyle, nil
	default:
		return "", fmt.Errorf("invalid nudge type: %s (must be focus, skip, constraint, or style)", s)
	}
}

// ValidNudgeTypes returns all valid nudge type strings
func ValidNudgeTypes() []string {
	return []string{"focus", "skip", "constraint", "style"}
}

// generateID creates a unique ID for a nudge
func generateID() string {
	return fmt.Sprintf("nudge_%d", time.Now().UnixNano())
}

// FormatAcknowledgment creates a message for logging nudge acknowledgment
func FormatAcknowledgment(nudges []Nudge) string {
	if len(nudges) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Acknowledged nudges:\n")
	for _, n := range nudges {
		b.WriteString(fmt.Sprintf("  - [%s] %s\n", strings.ToUpper(string(n.Type)), n.Content))
	}
	return b.String()
}
