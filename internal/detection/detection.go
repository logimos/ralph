// Package detection provides build system detection and preset management.
package detection

import (
	"fmt"
	"os"

	"github.com/logimos/ralph/internal/config"
)

// BuildSystemPreset defines commands for a build system
type BuildSystemPreset struct {
	TypeCheck string
	Test      string
}

// BuildSystemPresets defines commands for common build systems
var BuildSystemPresets = map[string]BuildSystemPreset{
	"pnpm": {
		TypeCheck: "pnpm typecheck",
		Test:      "pnpm test",
	},
	"npm": {
		TypeCheck: "npm run typecheck",
		Test:      "npm test",
	},
	"yarn": {
		TypeCheck: "yarn typecheck",
		Test:      "yarn test",
	},
	"gradle": {
		TypeCheck: "./gradlew check",
		Test:      "./gradlew test",
	},
	"maven": {
		TypeCheck: "mvn compile",
		Test:      "mvn test",
	},
	"cargo": {
		TypeCheck: "cargo check",
		Test:      "cargo test",
	},
	"go": {
		TypeCheck: "go build ./...",
		Test:      "go test ./...",
	},
	"python": {
		TypeCheck: "mypy .",
		Test:      "pytest",
	},
}

// DetectBuildSystem attempts to detect the build system from project files
func DetectBuildSystem() string {
	// Check for Gradle
	if _, err := os.Stat("build.gradle"); err == nil {
		return "gradle"
	}
	if _, err := os.Stat("build.gradle.kts"); err == nil {
		return "gradle"
	}
	if _, err := os.Stat("gradlew"); err == nil {
		return "gradle"
	}

	// Check for Maven
	if _, err := os.Stat("pom.xml"); err == nil {
		return "maven"
	}

	// Check for Cargo (Rust)
	if _, err := os.Stat("Cargo.toml"); err == nil {
		return "cargo"
	}

	// Check for Go modules
	if _, err := os.Stat("go.mod"); err == nil {
		return "go"
	}

	// Check for Python (common indicators)
	if _, err := os.Stat("setup.py"); err == nil {
		return "python"
	}
	if _, err := os.Stat("pyproject.toml"); err == nil {
		return "python"
	}
	if _, err := os.Stat("requirements.txt"); err == nil {
		return "python"
	}

	// Check for pnpm (has pnpm-lock.yaml)
	if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		return "pnpm"
	}

	// Check for yarn (has yarn.lock)
	if _, err := os.Stat("yarn.lock"); err == nil {
		return "yarn"
	}

	// Check for npm (has package.json, but no lock file means npm)
	if _, err := os.Stat("package.json"); err == nil {
		return "npm"
	}

	// Default to pnpm for backward compatibility
	return "pnpm"
}

// ApplyBuildSystemConfig applies build system presets or auto-detection
func ApplyBuildSystemConfig(cfg *config.Config) {
	// If both typecheck and test are explicitly set, don't override
	if cfg.TypeCheckCmd != "" && cfg.TestCmd != "" {
		return
	}

	var buildSystem string

	// Determine which build system to use
	if cfg.BuildSystem != "" {
		if cfg.BuildSystem == "auto" {
			buildSystem = DetectBuildSystem()
			if cfg.Verbose {
				fmt.Printf("Auto-detected build system: %s\n", buildSystem)
			}
		} else {
			buildSystem = cfg.BuildSystem
		}
	} else {
		// Auto-detect if neither build-system nor individual commands are set
		if cfg.TypeCheckCmd == "" && cfg.TestCmd == "" {
			buildSystem = DetectBuildSystem()
			if cfg.Verbose {
				fmt.Printf("Auto-detected build system: %s\n", buildSystem)
			}
		} else {
			// Use defaults if only one command is set
			buildSystem = "pnpm"
		}
	}

	// Apply preset if available
	if preset, ok := BuildSystemPresets[buildSystem]; ok {
		if cfg.TypeCheckCmd == "" {
			cfg.TypeCheckCmd = preset.TypeCheck
		}
		if cfg.TestCmd == "" {
			cfg.TestCmd = preset.Test
		}
	} else {
		// Unknown build system, use defaults
		if cfg.TypeCheckCmd == "" {
			cfg.TypeCheckCmd = BuildSystemPresets["pnpm"].TypeCheck
		}
		if cfg.TestCmd == "" {
			cfg.TestCmd = BuildSystemPresets["pnpm"].Test
		}
		if cfg.Verbose {
			fmt.Printf("Warning: Unknown build system '%s', using pnpm defaults\n", buildSystem)
		}
	}
}
