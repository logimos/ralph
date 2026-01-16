// Package milestone provides milestone-based progress tracking for Ralph.
package milestone

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/logimos/ralph/internal/plan"
)

// Status represents the current status of a milestone
type Status string

const (
	// StatusNotStarted indicates no features in the milestone have been completed
	StatusNotStarted Status = "not_started"
	// StatusInProgress indicates some features in the milestone are complete
	StatusInProgress Status = "in_progress"
	// StatusComplete indicates all features in the milestone are complete
	StatusComplete Status = "complete"
)

// Milestone represents a project milestone with associated features
type Milestone struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Criteria    string   `json:"criteria,omitempty"`    // Success criteria for the milestone
	Order       int      `json:"order,omitempty"`       // Display/priority order
	Features    []int    `json:"features,omitempty"`    // List of feature IDs (alternative to milestone field in Plan)
}

// MilestoneFile represents the structure of a plan.json file that includes milestones
type MilestoneFile struct {
	Milestones []Milestone  `json:"milestones,omitempty"`
	Plans      []plan.Plan  `json:"plans,omitempty"` // For files that use the new format
}

// Progress represents the progress of a milestone
type Progress struct {
	Milestone       *Milestone
	TotalFeatures   int
	CompletedFeatures int
	Percentage      float64
	Status          Status
	Features        []plan.Plan // Features belonging to this milestone
}

// Manager handles milestone operations
type Manager struct {
	milestones []Milestone
	plans      []plan.Plan
}

// NewManager creates a new milestone manager from plans
func NewManager(plans []plan.Plan) *Manager {
	return &Manager{
		plans: plans,
	}
}

// LoadMilestones loads milestone definitions from a JSON file.
// The file can be a separate milestones.json or embedded in plan.json.
func (m *Manager) LoadMilestones(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read milestones file: %w", err)
	}

	// First, try to parse as a file with milestones array
	var milestoneFile MilestoneFile
	if err := json.Unmarshal(data, &milestoneFile); err == nil && len(milestoneFile.Milestones) > 0 {
		m.milestones = milestoneFile.Milestones
		return nil
	}

	// Try to parse as a plain array of milestones
	var milestones []Milestone
	if err := json.Unmarshal(data, &milestones); err == nil && len(milestones) > 0 {
		m.milestones = milestones
		return nil
	}

	return nil // No milestones defined, which is valid
}

// SetMilestones sets the milestones manually (useful for testing)
func (m *Manager) SetMilestones(milestones []Milestone) {
	m.milestones = milestones
}

// ExtractMilestonesFromPlans extracts unique milestones from plan milestone fields
func (m *Manager) ExtractMilestonesFromPlans() {
	milestoneMap := make(map[string]*Milestone)
	
	for _, p := range m.plans {
		if p.Milestone == "" {
			continue
		}
		
		if _, exists := milestoneMap[p.Milestone]; !exists {
			milestoneMap[p.Milestone] = &Milestone{
				ID:   strings.ToLower(strings.ReplaceAll(p.Milestone, " ", "-")),
				Name: p.Milestone,
			}
		}
	}
	
	// Convert map to slice
	for _, milestone := range milestoneMap {
		m.milestones = append(m.milestones, *milestone)
	}
	
	// Sort by name for consistent ordering
	sort.Slice(m.milestones, func(i, j int) bool {
		return m.milestones[i].Name < m.milestones[j].Name
	})
}

// GetMilestones returns all defined milestones
func (m *Manager) GetMilestones() []Milestone {
	// If no milestones were loaded, try to extract from plans
	if len(m.milestones) == 0 {
		m.ExtractMilestonesFromPlans()
	}
	return m.milestones
}

// GetFeaturesForMilestone returns all features belonging to a milestone
func (m *Manager) GetFeaturesForMilestone(milestoneName string) []plan.Plan {
	var features []plan.Plan
	
	// Find the milestone definition to get any explicit feature IDs
	var milestoneFeatureIDs map[int]bool
	for _, ms := range m.milestones {
		if strings.EqualFold(ms.Name, milestoneName) || strings.EqualFold(ms.ID, milestoneName) {
			if len(ms.Features) > 0 {
				milestoneFeatureIDs = make(map[int]bool)
				for _, id := range ms.Features {
					milestoneFeatureIDs[id] = true
				}
			}
			break
		}
	}
	
	// Get features that match the milestone
	for _, p := range m.plans {
		// Check if feature is explicitly listed in milestone definition
		if milestoneFeatureIDs != nil && milestoneFeatureIDs[p.ID] {
			features = append(features, p)
			continue
		}
		
		// Check if feature has the milestone field set
		if strings.EqualFold(p.Milestone, milestoneName) {
			features = append(features, p)
		}
	}
	
	// Sort by milestone_order if set, then by ID
	sort.Slice(features, func(i, j int) bool {
		if features[i].MilestoneOrder != features[j].MilestoneOrder {
			return features[i].MilestoneOrder < features[j].MilestoneOrder
		}
		return features[i].ID < features[j].ID
	})
	
	return features
}

// CalculateProgress calculates the progress for a specific milestone
func (m *Manager) CalculateProgress(milestoneName string) *Progress {
	features := m.GetFeaturesForMilestone(milestoneName)
	
	// Find the milestone definition
	var milestone *Milestone
	for i := range m.milestones {
		if strings.EqualFold(m.milestones[i].Name, milestoneName) || 
		   strings.EqualFold(m.milestones[i].ID, milestoneName) {
			milestone = &m.milestones[i]
			break
		}
	}
	
	if milestone == nil {
		// Create a temporary milestone for the name
		milestone = &Milestone{
			ID:   strings.ToLower(strings.ReplaceAll(milestoneName, " ", "-")),
			Name: milestoneName,
		}
	}
	
	total := len(features)
	completed := 0
	for _, f := range features {
		if f.Tested {
			completed++
		}
	}
	
	var percentage float64
	if total > 0 {
		percentage = float64(completed) / float64(total) * 100
	}
	
	status := StatusNotStarted
	if completed > 0 {
		status = StatusInProgress
	}
	if completed == total && total > 0 {
		status = StatusComplete
	}
	
	return &Progress{
		Milestone:         milestone,
		TotalFeatures:     total,
		CompletedFeatures: completed,
		Percentage:        percentage,
		Status:            status,
		Features:          features,
	}
}

// CalculateAllProgress calculates progress for all milestones
func (m *Manager) CalculateAllProgress() []*Progress {
	milestones := m.GetMilestones()
	var progress []*Progress
	
	for _, ms := range milestones {
		progress = append(progress, m.CalculateProgress(ms.Name))
	}
	
	// Sort by order, then by name
	sort.Slice(progress, func(i, j int) bool {
		if progress[i].Milestone.Order != progress[j].Milestone.Order {
			return progress[i].Milestone.Order < progress[j].Milestone.Order
		}
		return progress[i].Milestone.Name < progress[j].Milestone.Name
	})
	
	return progress
}

// GetCompletedMilestones returns milestones that are 100% complete
func (m *Manager) GetCompletedMilestones() []*Progress {
	var completed []*Progress
	for _, p := range m.CalculateAllProgress() {
		if p.Status == StatusComplete {
			completed = append(completed, p)
		}
	}
	return completed
}

// GetNextMilestoneToComplete returns the milestone closest to completion
// that isn't already complete
func (m *Manager) GetNextMilestoneToComplete() *Progress {
	var best *Progress
	
	for _, p := range m.CalculateAllProgress() {
		if p.Status == StatusComplete {
			continue
		}
		if best == nil || p.Percentage > best.Percentage {
			best = p
		}
	}
	
	return best
}

// FormatProgress returns a formatted string showing milestone progress
func FormatProgress(p *Progress) string {
	statusIcon := "○"
	switch p.Status {
	case StatusInProgress:
		statusIcon = "◐"
	case StatusComplete:
		statusIcon = "●"
	}
	
	return fmt.Sprintf("%s %s: %d/%d (%.0f%%)",
		statusIcon,
		p.Milestone.Name,
		p.CompletedFeatures,
		p.TotalFeatures,
		p.Percentage)
}

// FormatProgressBar returns a visual progress bar for a milestone
func FormatProgressBar(p *Progress, width int) string {
	if width < 10 {
		width = 10
	}
	
	filled := int(float64(width) * p.Percentage / 100)
	if filled > width {
		filled = width
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	
	return fmt.Sprintf("[%s] %.0f%%", bar, p.Percentage)
}

// CelebrationMessage returns a celebration message for a completed milestone
func CelebrationMessage(milestoneName string) string {
	messages := []string{
		"Milestone '%s' complete! Great progress!",
		"Congratulations! Milestone '%s' is done!",
		"Milestone '%s' achieved! Keep up the great work!",
		"You've reached milestone '%s'! Well done!",
	}
	// Use a simple deterministic selection based on name length
	idx := len(milestoneName) % len(messages)
	return fmt.Sprintf(messages[idx], milestoneName)
}

// Summary returns a summary string of all milestone progress
func (m *Manager) Summary() string {
	progress := m.CalculateAllProgress()
	
	if len(progress) == 0 {
		return "No milestones defined"
	}
	
	var sb strings.Builder
	sb.WriteString("Milestone Progress:\n")
	
	for _, p := range progress {
		sb.WriteString(fmt.Sprintf("  %s\n", FormatProgress(p)))
		if p.Milestone.Description != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", p.Milestone.Description))
		}
	}
	
	// Overall summary
	totalFeatures := 0
	completedFeatures := 0
	completedMilestones := 0
	for _, p := range progress {
		totalFeatures += p.TotalFeatures
		completedFeatures += p.CompletedFeatures
		if p.Status == StatusComplete {
			completedMilestones++
		}
	}
	
	overallPct := float64(0)
	if totalFeatures > 0 {
		overallPct = float64(completedFeatures) / float64(totalFeatures) * 100
	}
	
	sb.WriteString(fmt.Sprintf("\nOverall: %d/%d milestones complete, %d/%d features (%.0f%%)\n",
		completedMilestones, len(progress),
		completedFeatures, totalFeatures,
		overallPct))
	
	return sb.String()
}

// HasMilestones returns true if any milestones are defined (either explicitly or via plan fields)
func (m *Manager) HasMilestones() bool {
	if len(m.milestones) > 0 {
		return true
	}
	
	// Check if any plans have milestone field set
	for _, p := range m.plans {
		if p.Milestone != "" {
			return true
		}
	}
	
	return false
}
