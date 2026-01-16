// Package environment provides environment detection and adaptation for Ralph.
// It detects CI environments, system resources, and project complexity to
// automatically adjust execution parameters.
package environment

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// EnvironmentType represents the detected execution environment.
type EnvironmentType string

const (
	// EnvLocal represents a local development environment
	EnvLocal EnvironmentType = "local"
	// EnvGitHubActions represents GitHub Actions CI
	EnvGitHubActions EnvironmentType = "github-actions"
	// EnvGitLabCI represents GitLab CI
	EnvGitLabCI EnvironmentType = "gitlab-ci"
	// EnvJenkins represents Jenkins CI
	EnvJenkins EnvironmentType = "jenkins"
	// EnvCircleCI represents CircleCI
	EnvCircleCI EnvironmentType = "circleci"
	// EnvTravisCI represents Travis CI
	EnvTravisCI EnvironmentType = "travis-ci"
	// EnvAzureDevOps represents Azure DevOps Pipelines
	EnvAzureDevOps EnvironmentType = "azure-devops"
	// EnvGenericCI represents an unrecognized CI environment
	EnvGenericCI EnvironmentType = "ci"
)

// ProjectComplexity represents estimated project complexity.
type ProjectComplexity string

const (
	// ComplexitySmall represents a small project (<100 files)
	ComplexitySmall ProjectComplexity = "small"
	// ComplexityMedium represents a medium project (100-1000 files)
	ComplexityMedium ProjectComplexity = "medium"
	// ComplexityLarge represents a large project (>1000 files)
	ComplexityLarge ProjectComplexity = "large"
)

// Default timeout durations by environment
const (
	DefaultLocalTimeout  = 30 * time.Second
	DefaultCITimeout     = 120 * time.Second
	DefaultLargeTimeout  = 180 * time.Second
)

// EnvironmentProfile contains detected environment attributes.
type EnvironmentProfile struct {
	// Type is the detected environment type
	Type EnvironmentType `json:"type"`

	// CIEnvironment indicates whether running in a CI environment
	CIEnvironment bool `json:"is_ci"`

	// CIProvider is the specific CI provider name (if detected)
	CIProvider string `json:"ci_provider,omitempty"`

	// CPUCores is the number of available CPU cores
	CPUCores int `json:"cpu_cores"`

	// MemoryMB is the estimated available memory in megabytes
	MemoryMB int64 `json:"memory_mb"`

	// FileCount is the number of files in the project
	FileCount int `json:"file_count"`

	// RepoSizeBytes is the estimated repository size in bytes
	RepoSizeBytes int64 `json:"repo_size_bytes"`

	// Complexity is the estimated project complexity
	Complexity ProjectComplexity `json:"complexity"`

	// RecommendedTimeout is the suggested operation timeout
	RecommendedTimeout time.Duration `json:"recommended_timeout"`

	// RecommendedVerbose indicates if verbose output is recommended
	RecommendedVerbose bool `json:"recommended_verbose"`

	// ParallelHint is the suggested number of parallel operations
	ParallelHint int `json:"parallel_hint"`
}

// Detect analyzes the current environment and returns an EnvironmentProfile.
func Detect() *EnvironmentProfile {
	profile := &EnvironmentProfile{
		CPUCores: runtime.NumCPU(),
	}

	// Detect CI environment
	detectCIEnvironment(profile)

	// Detect system resources
	detectSystemResources(profile)

	// Detect project complexity
	detectProjectComplexity(profile)

	// Calculate recommendations based on detected attributes
	calculateRecommendations(profile)

	return profile
}

// detectCIEnvironment detects CI environment from environment variables.
func detectCIEnvironment(profile *EnvironmentProfile) {
	// Check for specific CI providers (order matters for precedence)
	ciChecks := []struct {
		envVar    string
		envType   EnvironmentType
		provider  string
	}{
		{"GITHUB_ACTIONS", EnvGitHubActions, "GitHub Actions"},
		{"GITLAB_CI", EnvGitLabCI, "GitLab CI"},
		{"JENKINS_URL", EnvJenkins, "Jenkins"},
		{"CIRCLECI", EnvCircleCI, "CircleCI"},
		{"TRAVIS", EnvTravisCI, "Travis CI"},
		{"TF_BUILD", EnvAzureDevOps, "Azure DevOps"},
	}

	for _, check := range ciChecks {
		if os.Getenv(check.envVar) != "" {
			profile.Type = check.envType
			profile.CIEnvironment = true
			profile.CIProvider = check.provider
			return
		}
	}

	// Check for generic CI indicators
	genericCIVars := []string{"CI", "CONTINUOUS_INTEGRATION", "BUILD_NUMBER"}
	for _, envVar := range genericCIVars {
		if os.Getenv(envVar) != "" {
			profile.Type = EnvGenericCI
			profile.CIEnvironment = true
			profile.CIProvider = "Unknown CI"
			return
		}
	}

	// Default to local environment
	profile.Type = EnvLocal
	profile.CIEnvironment = false
}

// detectSystemResources detects available system resources.
func detectSystemResources(profile *EnvironmentProfile) {
	// CPU cores already set via runtime.NumCPU()

	// Try to detect memory (platform-specific)
	profile.MemoryMB = detectMemory()
}

// detectMemory attempts to detect available system memory.
// Returns the memory in megabytes, or 0 if detection fails.
func detectMemory() int64 {
	// Try reading from /proc/meminfo on Linux
	data, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					kb, err := strconv.ParseInt(fields[1], 10, 64)
					if err == nil {
						return kb / 1024 // Convert KB to MB
					}
				}
			}
		}
	}

	// Try using sysctl on macOS
	if runtime.GOOS == "darwin" {
		output, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			bytes, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
			if err == nil {
				return bytes / (1024 * 1024) // Convert bytes to MB
			}
		}
	}

	// Default to 0 if detection fails
	return 0
}

// detectProjectComplexity estimates project complexity based on file count and repo size.
func detectProjectComplexity(profile *EnvironmentProfile) {
	cwd, err := os.Getwd()
	if err != nil {
		profile.Complexity = ComplexitySmall
		return
	}

	fileCount := 0
	var totalSize int64

	// Walk the directory tree, ignoring common non-source directories
	ignoreDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		".venv":        true,
		"venv":         true,
		"__pycache__":  true,
		"dist":         true,
		"build":        true,
		"target":       true,
		".idea":        true,
		".vscode":      true,
	}

	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip ignored directories
		if info.IsDir() {
			if ignoreDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		fileCount++
		totalSize += info.Size()

		// Early exit for large projects to avoid slowdown
		if fileCount > 5000 {
			return filepath.SkipAll
		}

		return nil
	})

	profile.FileCount = fileCount
	profile.RepoSizeBytes = totalSize

	// Determine complexity based on file count
	switch {
	case fileCount < 100:
		profile.Complexity = ComplexitySmall
	case fileCount < 1000:
		profile.Complexity = ComplexityMedium
	default:
		profile.Complexity = ComplexityLarge
	}
}

// calculateRecommendations calculates execution recommendations based on profile.
func calculateRecommendations(profile *EnvironmentProfile) {
	// Recommend verbose output in CI environments (for logging)
	profile.RecommendedVerbose = profile.CIEnvironment

	// Calculate recommended timeout
	baseTimeout := DefaultLocalTimeout
	if profile.CIEnvironment {
		baseTimeout = DefaultCITimeout
	}

	// Adjust for project complexity
	switch profile.Complexity {
	case ComplexityLarge:
		baseTimeout = baseTimeout * 3
	case ComplexityMedium:
		baseTimeout = baseTimeout * 2
	}

	profile.RecommendedTimeout = baseTimeout

	// Calculate parallel hint based on CPU cores
	// Leave some cores free for system operations
	parallelHint := profile.CPUCores
	if parallelHint > 2 {
		parallelHint = parallelHint - 1
	}
	if parallelHint > 8 {
		parallelHint = 8 // Cap at 8 for reasonable resource usage
	}
	if parallelHint < 1 {
		parallelHint = 1
	}

	profile.ParallelHint = parallelHint
}

// IsCI returns true if running in any CI environment.
func (p *EnvironmentProfile) IsCI() bool {
	return p.CIEnvironment
}

// Summary returns a human-readable summary of the environment profile.
func (p *EnvironmentProfile) Summary() string {
	var sb strings.Builder

	sb.WriteString("Environment: ")
	if p.CIEnvironment {
		sb.WriteString(p.CIProvider)
	} else {
		sb.WriteString("Local development")
	}
	sb.WriteString("\n")

	sb.WriteString("  CPU cores: ")
	sb.WriteString(strconv.Itoa(p.CPUCores))
	sb.WriteString("\n")

	if p.MemoryMB > 0 {
		sb.WriteString("  Memory: ")
		if p.MemoryMB > 1024 {
			sb.WriteString(strconv.FormatFloat(float64(p.MemoryMB)/1024, 'f', 1, 64))
			sb.WriteString(" GB\n")
		} else {
			sb.WriteString(strconv.FormatInt(p.MemoryMB, 10))
			sb.WriteString(" MB\n")
		}
	}

	sb.WriteString("  Project complexity: ")
	sb.WriteString(string(p.Complexity))
	sb.WriteString(" (")
	sb.WriteString(strconv.Itoa(p.FileCount))
	sb.WriteString(" files)\n")

	sb.WriteString("  Recommended timeout: ")
	sb.WriteString(p.RecommendedTimeout.String())
	sb.WriteString("\n")

	sb.WriteString("  Parallel hint: ")
	sb.WriteString(strconv.Itoa(p.ParallelHint))
	sb.WriteString(" workers")

	return sb.String()
}

// ParseEnvironmentType parses a string into an EnvironmentType.
// Returns EnvLocal if the string is empty or unrecognized.
func ParseEnvironmentType(s string) EnvironmentType {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "github-actions", "github", "gh":
		return EnvGitHubActions
	case "gitlab-ci", "gitlab", "gl":
		return EnvGitLabCI
	case "jenkins":
		return EnvJenkins
	case "circleci", "circle":
		return EnvCircleCI
	case "travis-ci", "travis":
		return EnvTravisCI
	case "azure-devops", "azure", "azdo":
		return EnvAzureDevOps
	case "ci", "generic-ci":
		return EnvGenericCI
	case "local", "":
		return EnvLocal
	default:
		return EnvLocal
	}
}

// ForceEnvironment creates an EnvironmentProfile with a forced environment type.
// This is used when the user overrides environment detection via the -environment flag.
func ForceEnvironment(envType EnvironmentType) *EnvironmentProfile {
	// Start with detection to get accurate resource information
	profile := Detect()

	// Override the environment type
	profile.Type = envType
	profile.CIEnvironment = envType != EnvLocal

	if profile.CIEnvironment {
		profile.CIProvider = string(envType)
	} else {
		profile.CIProvider = ""
	}

	// Recalculate recommendations with forced environment
	calculateRecommendations(profile)

	return profile
}
