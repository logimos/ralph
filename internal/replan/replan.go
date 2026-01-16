// Package replan provides adaptive plan replanning functionality for Ralph.
// It enables dynamic plan adjustment when tests fail repeatedly, requirements change,
// or features become blocked.
package replan

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/logimos/ralph/internal/plan"
)

// TriggerType represents the type of condition that triggered replanning
type TriggerType string

const (
	// TriggerNone indicates no trigger was activated
	TriggerNone TriggerType = "none"
	// TriggerTestFailure indicates replanning due to repeated test failures
	TriggerTestFailure TriggerType = "test_failure"
	// TriggerRequirementChange indicates replanning due to plan.json changes
	TriggerRequirementChange TriggerType = "requirement_change"
	// TriggerBlockedFeature indicates replanning due to a blocked feature
	TriggerBlockedFeature TriggerType = "blocked_feature"
	// TriggerManual indicates manually triggered replanning
	TriggerManual TriggerType = "manual"
)

// ReplanTrigger defines the interface for conditions that trigger replanning
type ReplanTrigger interface {
	// Name returns the trigger name
	Name() TriggerType
	// Check evaluates if the trigger condition is met
	Check(state *ReplanState) bool
	// Description returns a human-readable description
	Description() string
}

// ReplanState contains the current state needed for trigger evaluation
type ReplanState struct {
	// FeatureID is the current feature being worked on
	FeatureID int
	// ConsecutiveFailures is the number of consecutive test failures
	ConsecutiveFailures int
	// FailureTypes contains the types of recent failures
	FailureTypes []string
	// PlanHash is the hash of the current plan.json content
	PlanHash string
	// LastPlanHash is the hash from the previous check
	LastPlanHash string
	// BlockedFeatures contains IDs of features marked as blocked
	BlockedFeatures []int
	// TotalIterations is the total iterations run so far
	TotalIterations int
	// Plans contains the current plan data
	Plans []plan.Plan
}

// TestFailureTrigger triggers replanning when tests fail repeatedly
type TestFailureTrigger struct {
	// Threshold is the number of consecutive failures before triggering
	Threshold int
}

// NewTestFailureTrigger creates a new test failure trigger
func NewTestFailureTrigger(threshold int) *TestFailureTrigger {
	if threshold <= 0 {
		threshold = 3 // default
	}
	return &TestFailureTrigger{Threshold: threshold}
}

// Name returns the trigger name
func (t *TestFailureTrigger) Name() TriggerType {
	return TriggerTestFailure
}

// Description returns a human-readable description
func (t *TestFailureTrigger) Description() string {
	return fmt.Sprintf("Trigger replanning after %d consecutive test failures", t.Threshold)
}

// Check evaluates if the trigger condition is met
func (t *TestFailureTrigger) Check(state *ReplanState) bool {
	return state.ConsecutiveFailures >= t.Threshold
}

// RequirementChangeTrigger detects when plan.json has been manually edited
type RequirementChangeTrigger struct{}

// NewRequirementChangeTrigger creates a new requirement change trigger
func NewRequirementChangeTrigger() *RequirementChangeTrigger {
	return &RequirementChangeTrigger{}
}

// Name returns the trigger name
func (t *RequirementChangeTrigger) Name() TriggerType {
	return TriggerRequirementChange
}

// Description returns a human-readable description
func (t *RequirementChangeTrigger) Description() string {
	return "Trigger replanning when plan.json is externally modified"
}

// Check evaluates if the trigger condition is met
func (t *RequirementChangeTrigger) Check(state *ReplanState) bool {
	return state.PlanHash != "" && state.LastPlanHash != "" && state.PlanHash != state.LastPlanHash
}

// BlockedFeatureTrigger triggers when a feature becomes blocked
type BlockedFeatureTrigger struct {
	// MinBlocked is the minimum number of blocked features to trigger
	MinBlocked int
}

// NewBlockedFeatureTrigger creates a new blocked feature trigger
func NewBlockedFeatureTrigger(minBlocked int) *BlockedFeatureTrigger {
	if minBlocked <= 0 {
		minBlocked = 1 // default
	}
	return &BlockedFeatureTrigger{MinBlocked: minBlocked}
}

// Name returns the trigger name
func (t *BlockedFeatureTrigger) Name() TriggerType {
	return TriggerBlockedFeature
}

// Description returns a human-readable description
func (t *BlockedFeatureTrigger) Description() string {
	return fmt.Sprintf("Trigger replanning when %d or more features are blocked", t.MinBlocked)
}

// Check evaluates if the trigger condition is met
func (t *BlockedFeatureTrigger) Check(state *ReplanState) bool {
	return len(state.BlockedFeatures) >= t.MinBlocked
}

// ManualTrigger is activated explicitly by user request
type ManualTrigger struct {
	triggered bool
}

// NewManualTrigger creates a new manual trigger
func NewManualTrigger() *ManualTrigger {
	return &ManualTrigger{}
}

// Name returns the trigger name
func (t *ManualTrigger) Name() TriggerType {
	return TriggerManual
}

// Description returns a human-readable description
func (t *ManualTrigger) Description() string {
	return "Manually triggered replanning"
}

// Check evaluates if the trigger condition is met
func (t *ManualTrigger) Check(state *ReplanState) bool {
	return t.triggered
}

// Activate sets the manual trigger
func (t *ManualTrigger) Activate() {
	t.triggered = true
}

// Reset clears the manual trigger
func (t *ManualTrigger) Reset() {
	t.triggered = false
}

// PlanVersion represents a versioned backup of a plan
type PlanVersion struct {
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Trigger   string    `json:"trigger"`
	Path      string    `json:"path"`
	Hash      string    `json:"hash"`
}

// PlanVersioner manages plan file versioning
type PlanVersioner struct {
	basePath string
	versions []PlanVersion
}

// NewPlanVersioner creates a new plan versioner
func NewPlanVersioner(planPath string) *PlanVersioner {
	return &PlanVersioner{
		basePath: planPath,
		versions: make([]PlanVersion, 0),
	}
}

// CreateBackup creates a versioned backup of the current plan file
func (pv *PlanVersioner) CreateBackup(trigger TriggerType) (string, error) {
	// Read current plan file
	data, err := os.ReadFile(pv.basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read plan file: %w", err)
	}

	// Calculate hash for deduplication
	hash := fmt.Sprintf("%x", md5.Sum(data))

	// Check if this exact version already exists
	for _, v := range pv.versions {
		if v.Hash == hash {
			return v.Path, nil // Already backed up
		}
	}

	// Determine next version number
	nextVersion := len(pv.versions) + 1

	// Create backup path
	ext := filepath.Ext(pv.basePath)
	base := strings.TrimSuffix(pv.basePath, ext)
	backupPath := fmt.Sprintf("%s.bak.%d%s", base, nextVersion, ext)

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	// Record version
	version := PlanVersion{
		Version:   nextVersion,
		Timestamp: time.Now(),
		Trigger:   string(trigger),
		Path:      backupPath,
		Hash:      hash,
	}
	pv.versions = append(pv.versions, version)

	return backupPath, nil
}

// GetVersions returns all recorded versions
func (pv *PlanVersioner) GetVersions() []PlanVersion {
	return pv.versions
}

// GetLatestVersion returns the most recent version
func (pv *PlanVersioner) GetLatestVersion() *PlanVersion {
	if len(pv.versions) == 0 {
		return nil
	}
	return &pv.versions[len(pv.versions)-1]
}

// RestoreVersion restores a specific version
func (pv *PlanVersioner) RestoreVersion(version int) error {
	if version < 1 || version > len(pv.versions) {
		return fmt.Errorf("invalid version number: %d", version)
	}

	v := pv.versions[version-1]
	data, err := os.ReadFile(v.Path)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if err := os.WriteFile(pv.basePath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore plan: %w", err)
	}

	return nil
}

// DiscoverBackups finds existing backup files
func (pv *PlanVersioner) DiscoverBackups() error {
	ext := filepath.Ext(pv.basePath)
	base := strings.TrimSuffix(pv.basePath, ext)

	// Find all matching backup files
	pattern := fmt.Sprintf("%s.bak.*%s", base, ext)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find backups: %w", err)
	}

	for _, match := range matches {
		// Extract version number from filename
		var version int
		_, err := fmt.Sscanf(filepath.Base(match), filepath.Base(base)+".bak.%d"+ext, &version)
		if err != nil {
			continue
		}

		// Read file for hash
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		pv.versions = append(pv.versions, PlanVersion{
			Version:   version,
			Timestamp: info.ModTime(),
			Path:      match,
			Hash:      fmt.Sprintf("%x", md5.Sum(data)),
		})
	}

	return nil
}

// PlanDiff represents changes between two plan versions
type PlanDiff struct {
	Added    []plan.Plan `json:"added"`
	Removed  []plan.Plan `json:"removed"`
	Modified []PlanChange `json:"modified"`
}

// PlanChange represents a modification to a plan
type PlanChange struct {
	ID          int    `json:"id"`
	Field       string `json:"field"`
	OldValue    string `json:"old_value"`
	NewValue    string `json:"new_value"`
	Description string `json:"description,omitempty"`
}

// ComputeDiff computes the difference between two plan lists
func ComputeDiff(oldPlans, newPlans []plan.Plan) *PlanDiff {
	diff := &PlanDiff{
		Added:    make([]plan.Plan, 0),
		Removed:  make([]plan.Plan, 0),
		Modified: make([]PlanChange, 0),
	}

	// Create maps for lookup
	oldMap := make(map[int]plan.Plan)
	newMap := make(map[int]plan.Plan)

	for _, p := range oldPlans {
		oldMap[p.ID] = p
	}
	for _, p := range newPlans {
		newMap[p.ID] = p
	}

	// Find added and modified
	for _, newP := range newPlans {
		oldP, exists := oldMap[newP.ID]
		if !exists {
			diff.Added = append(diff.Added, newP)
		} else {
			// Check for modifications
			changes := comparePlans(oldP, newP)
			diff.Modified = append(diff.Modified, changes...)
		}
	}

	// Find removed
	for _, oldP := range oldPlans {
		if _, exists := newMap[oldP.ID]; !exists {
			diff.Removed = append(diff.Removed, oldP)
		}
	}

	return diff
}

// comparePlans compares two plans and returns the changes
func comparePlans(old, new plan.Plan) []PlanChange {
	var changes []PlanChange

	if old.Description != new.Description {
		changes = append(changes, PlanChange{
			ID:       old.ID,
			Field:    "description",
			OldValue: old.Description,
			NewValue: new.Description,
		})
	}

	if old.Category != new.Category {
		changes = append(changes, PlanChange{
			ID:       old.ID,
			Field:    "category",
			OldValue: old.Category,
			NewValue: new.Category,
		})
	}

	if old.Tested != new.Tested {
		changes = append(changes, PlanChange{
			ID:       old.ID,
			Field:    "tested",
			OldValue: fmt.Sprintf("%v", old.Tested),
			NewValue: fmt.Sprintf("%v", new.Tested),
		})
	}

	if old.Deferred != new.Deferred {
		changes = append(changes, PlanChange{
			ID:       old.ID,
			Field:    "deferred",
			OldValue: fmt.Sprintf("%v", old.Deferred),
			NewValue: fmt.Sprintf("%v", new.Deferred),
		})
	}

	// Compare steps
	oldSteps := strings.Join(old.Steps, "|")
	newSteps := strings.Join(new.Steps, "|")
	if oldSteps != newSteps {
		changes = append(changes, PlanChange{
			ID:       old.ID,
			Field:    "steps",
			OldValue: fmt.Sprintf("%d steps", len(old.Steps)),
			NewValue: fmt.Sprintf("%d steps", len(new.Steps)),
		})
	}

	if old.ExpectedOutput != new.ExpectedOutput {
		changes = append(changes, PlanChange{
			ID:       old.ID,
			Field:    "expected_output",
			OldValue: truncate(old.ExpectedOutput, 50),
			NewValue: truncate(new.ExpectedOutput, 50),
		})
	}

	if old.Milestone != new.Milestone {
		changes = append(changes, PlanChange{
			ID:       old.ID,
			Field:    "milestone",
			OldValue: old.Milestone,
			NewValue: new.Milestone,
		})
	}

	return changes
}

// truncate shortens a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// IsEmpty returns true if there are no changes
func (d *PlanDiff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Modified) == 0
}

// Summary returns a human-readable summary of the diff
func (d *PlanDiff) Summary() string {
	if d.IsEmpty() {
		return "No changes detected"
	}

	var sb strings.Builder
	sb.WriteString("Plan Changes:\n")

	if len(d.Added) > 0 {
		sb.WriteString(fmt.Sprintf("  + Added: %d feature(s)\n", len(d.Added)))
		for _, p := range d.Added {
			sb.WriteString(fmt.Sprintf("    - #%d: %s\n", p.ID, truncate(p.Description, 60)))
		}
	}

	if len(d.Removed) > 0 {
		sb.WriteString(fmt.Sprintf("  - Removed: %d feature(s)\n", len(d.Removed)))
		for _, p := range d.Removed {
			sb.WriteString(fmt.Sprintf("    - #%d: %s\n", p.ID, truncate(p.Description, 60)))
		}
	}

	if len(d.Modified) > 0 {
		sb.WriteString(fmt.Sprintf("  ~ Modified: %d change(s)\n", len(d.Modified)))
		for _, c := range d.Modified {
			sb.WriteString(fmt.Sprintf("    - #%d.%s: %s -> %s\n", c.ID, c.Field, c.OldValue, c.NewValue))
		}
	}

	return sb.String()
}

// CalculatePlanHash computes a hash of the plan file content
func CalculatePlanHash(planPath string) (string, error) {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), nil
}

// CalculatePlansHash computes a hash of plans structure
func CalculatePlansHash(plans []plan.Plan) string {
	data, err := json.Marshal(plans)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", md5.Sum(data))
}
