package environment

import (
	"os"
	"runtime"
	"testing"
	"time"
)

func TestDetectCIEnvironment_GitHubActions(t *testing.T) {
	// Save and restore environment
	oldVal := os.Getenv("GITHUB_ACTIONS")
	defer os.Setenv("GITHUB_ACTIONS", oldVal)

	os.Setenv("GITHUB_ACTIONS", "true")

	profile := &EnvironmentProfile{}
	detectCIEnvironment(profile)

	if profile.Type != EnvGitHubActions {
		t.Errorf("expected type %s, got %s", EnvGitHubActions, profile.Type)
	}
	if !profile.CIEnvironment {
		t.Error("expected CIEnvironment to be true")
	}
	if profile.CIProvider != "GitHub Actions" {
		t.Errorf("expected provider 'GitHub Actions', got '%s'", profile.CIProvider)
	}
}

func TestDetectCIEnvironment_GitLabCI(t *testing.T) {
	// Save and restore environment
	oldVal := os.Getenv("GITLAB_CI")
	defer os.Setenv("GITLAB_CI", oldVal)

	// Clear other CI vars
	oldGH := os.Getenv("GITHUB_ACTIONS")
	os.Unsetenv("GITHUB_ACTIONS")
	defer os.Setenv("GITHUB_ACTIONS", oldGH)

	os.Setenv("GITLAB_CI", "true")

	profile := &EnvironmentProfile{}
	detectCIEnvironment(profile)

	if profile.Type != EnvGitLabCI {
		t.Errorf("expected type %s, got %s", EnvGitLabCI, profile.Type)
	}
	if profile.CIProvider != "GitLab CI" {
		t.Errorf("expected provider 'GitLab CI', got '%s'", profile.CIProvider)
	}
}

func TestDetectCIEnvironment_Jenkins(t *testing.T) {
	// Save and restore environment
	oldVal := os.Getenv("JENKINS_URL")
	defer os.Setenv("JENKINS_URL", oldVal)

	// Clear other CI vars
	oldGH := os.Getenv("GITHUB_ACTIONS")
	oldGL := os.Getenv("GITLAB_CI")
	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("GITLAB_CI")
	defer func() {
		os.Setenv("GITHUB_ACTIONS", oldGH)
		os.Setenv("GITLAB_CI", oldGL)
	}()

	os.Setenv("JENKINS_URL", "http://jenkins.example.com")

	profile := &EnvironmentProfile{}
	detectCIEnvironment(profile)

	if profile.Type != EnvJenkins {
		t.Errorf("expected type %s, got %s", EnvJenkins, profile.Type)
	}
	if profile.CIProvider != "Jenkins" {
		t.Errorf("expected provider 'Jenkins', got '%s'", profile.CIProvider)
	}
}

func TestDetectCIEnvironment_CircleCI(t *testing.T) {
	// Save and restore environment
	oldVal := os.Getenv("CIRCLECI")
	defer os.Setenv("CIRCLECI", oldVal)

	// Clear other CI vars
	oldVars := map[string]string{
		"GITHUB_ACTIONS": os.Getenv("GITHUB_ACTIONS"),
		"GITLAB_CI":      os.Getenv("GITLAB_CI"),
		"JENKINS_URL":    os.Getenv("JENKINS_URL"),
	}
	for k := range oldVars {
		os.Unsetenv(k)
	}
	defer func() {
		for k, v := range oldVars {
			os.Setenv(k, v)
		}
	}()

	os.Setenv("CIRCLECI", "true")

	profile := &EnvironmentProfile{}
	detectCIEnvironment(profile)

	if profile.Type != EnvCircleCI {
		t.Errorf("expected type %s, got %s", EnvCircleCI, profile.Type)
	}
	if profile.CIProvider != "CircleCI" {
		t.Errorf("expected provider 'CircleCI', got '%s'", profile.CIProvider)
	}
}

func TestDetectCIEnvironment_GenericCI(t *testing.T) {
	// Save and restore environment
	oldCI := os.Getenv("CI")
	defer os.Setenv("CI", oldCI)

	// Clear all specific CI vars
	ciVars := []string{"GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "CIRCLECI", "TRAVIS", "TF_BUILD"}
	oldVars := make(map[string]string)
	for _, v := range ciVars {
		oldVars[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	defer func() {
		for k, v := range oldVars {
			os.Setenv(k, v)
		}
	}()

	os.Setenv("CI", "true")

	profile := &EnvironmentProfile{}
	detectCIEnvironment(profile)

	if profile.Type != EnvGenericCI {
		t.Errorf("expected type %s, got %s", EnvGenericCI, profile.Type)
	}
	if !profile.CIEnvironment {
		t.Error("expected CIEnvironment to be true")
	}
}

func TestDetectCIEnvironment_Local(t *testing.T) {
	// Clear all CI vars
	ciVars := []string{"GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "CIRCLECI", "TRAVIS", "TF_BUILD", "CI", "CONTINUOUS_INTEGRATION", "BUILD_NUMBER"}
	oldVars := make(map[string]string)
	for _, v := range ciVars {
		oldVars[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	defer func() {
		for k, v := range oldVars {
			os.Setenv(k, v)
		}
	}()

	profile := &EnvironmentProfile{}
	detectCIEnvironment(profile)

	if profile.Type != EnvLocal {
		t.Errorf("expected type %s, got %s", EnvLocal, profile.Type)
	}
	if profile.CIEnvironment {
		t.Error("expected CIEnvironment to be false")
	}
}

func TestDetect_CPUCores(t *testing.T) {
	profile := Detect()

	if profile.CPUCores < 1 {
		t.Error("expected at least 1 CPU core")
	}
	if profile.CPUCores != runtime.NumCPU() {
		t.Errorf("expected %d CPU cores, got %d", runtime.NumCPU(), profile.CPUCores)
	}
}

func TestDetect_ParallelHint(t *testing.T) {
	profile := Detect()

	if profile.ParallelHint < 1 {
		t.Error("parallel hint should be at least 1")
	}
	if profile.ParallelHint > 8 {
		t.Error("parallel hint should be capped at 8")
	}
}

func TestDetect_Complexity(t *testing.T) {
	profile := Detect()

	// Complexity should be one of the valid values
	validComplexities := map[ProjectComplexity]bool{
		ComplexitySmall:  true,
		ComplexityMedium: true,
		ComplexityLarge:  true,
	}

	if !validComplexities[profile.Complexity] {
		t.Errorf("invalid complexity: %s", profile.Complexity)
	}
}

func TestDetect_RecommendedTimeout(t *testing.T) {
	profile := Detect()

	if profile.RecommendedTimeout < DefaultLocalTimeout {
		t.Errorf("recommended timeout should be at least %v, got %v", DefaultLocalTimeout, profile.RecommendedTimeout)
	}
}

func TestParseEnvironmentType(t *testing.T) {
	tests := []struct {
		input    string
		expected EnvironmentType
	}{
		{"github-actions", EnvGitHubActions},
		{"github", EnvGitHubActions},
		{"gh", EnvGitHubActions},
		{"GITHUB", EnvGitHubActions},
		{"gitlab-ci", EnvGitLabCI},
		{"gitlab", EnvGitLabCI},
		{"gl", EnvGitLabCI},
		{"jenkins", EnvJenkins},
		{"circleci", EnvCircleCI},
		{"circle", EnvCircleCI},
		{"travis-ci", EnvTravisCI},
		{"travis", EnvTravisCI},
		{"azure-devops", EnvAzureDevOps},
		{"azure", EnvAzureDevOps},
		{"ci", EnvGenericCI},
		{"generic-ci", EnvGenericCI},
		{"local", EnvLocal},
		{"", EnvLocal},
		{"unknown", EnvLocal},
		{"  local  ", EnvLocal},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseEnvironmentType(tt.input)
			if result != tt.expected {
				t.Errorf("ParseEnvironmentType(%q) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestForceEnvironment_GitHub(t *testing.T) {
	profile := ForceEnvironment(EnvGitHubActions)

	if profile.Type != EnvGitHubActions {
		t.Errorf("expected type %s, got %s", EnvGitHubActions, profile.Type)
	}
	if !profile.CIEnvironment {
		t.Error("expected CIEnvironment to be true for forced GitHub Actions")
	}
	// Recommendations should reflect CI environment
	if !profile.RecommendedVerbose {
		t.Error("expected RecommendedVerbose to be true for CI")
	}
}

func TestForceEnvironment_Local(t *testing.T) {
	profile := ForceEnvironment(EnvLocal)

	if profile.Type != EnvLocal {
		t.Errorf("expected type %s, got %s", EnvLocal, profile.Type)
	}
	if profile.CIEnvironment {
		t.Error("expected CIEnvironment to be false for forced local")
	}
}

func TestSummary(t *testing.T) {
	profile := &EnvironmentProfile{
		Type:               EnvGitHubActions,
		CIEnvironment:      true,
		CIProvider:         "GitHub Actions",
		CPUCores:           4,
		MemoryMB:           8192,
		FileCount:          150,
		Complexity:         ComplexityMedium,
		RecommendedTimeout: 60 * time.Second,
		ParallelHint:       3,
	}

	summary := profile.Summary()

	// Check that key information is present
	checks := []string{
		"GitHub Actions",
		"CPU cores: 4",
		"8.0 GB",
		"medium",
		"150 files",
		"1m0s",
		"3 workers",
	}

	for _, check := range checks {
		if !contains(summary, check) {
			t.Errorf("summary should contain %q, got:\n%s", check, summary)
		}
	}
}

func TestSummary_Local(t *testing.T) {
	profile := &EnvironmentProfile{
		Type:               EnvLocal,
		CIEnvironment:      false,
		CPUCores:           2,
		MemoryMB:           512,
		FileCount:          50,
		Complexity:         ComplexitySmall,
		RecommendedTimeout: 30 * time.Second,
		ParallelHint:       2,
	}

	summary := profile.Summary()

	if !contains(summary, "Local development") {
		t.Errorf("local summary should contain 'Local development', got:\n%s", summary)
	}
	if !contains(summary, "512 MB") {
		t.Errorf("local summary should contain '512 MB', got:\n%s", summary)
	}
}

func TestCalculateRecommendations_CILargeProject(t *testing.T) {
	profile := &EnvironmentProfile{
		CIEnvironment: true,
		Complexity:    ComplexityLarge,
		CPUCores:      4,
	}

	calculateRecommendations(profile)

	if !profile.RecommendedVerbose {
		t.Error("CI should have verbose output recommended")
	}

	// Large project in CI should have longer timeout
	expectedMinTimeout := DefaultCITimeout * 3
	if profile.RecommendedTimeout < expectedMinTimeout {
		t.Errorf("large CI project should have timeout >= %v, got %v", expectedMinTimeout, profile.RecommendedTimeout)
	}
}

func TestCalculateRecommendations_LocalSmallProject(t *testing.T) {
	profile := &EnvironmentProfile{
		CIEnvironment: false,
		Complexity:    ComplexitySmall,
		CPUCores:      2,
	}

	calculateRecommendations(profile)

	if profile.RecommendedVerbose {
		t.Error("local should not have verbose output recommended by default")
	}

	if profile.RecommendedTimeout != DefaultLocalTimeout {
		t.Errorf("small local project should have timeout %v, got %v", DefaultLocalTimeout, profile.RecommendedTimeout)
	}
}

// contains checks if str contains substr
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 || (len(str) > 0 && containsSubstring(str, substr)))
}

func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
