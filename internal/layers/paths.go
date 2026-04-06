package layers

import (
	"path/filepath"
)

// ProjectRoot returns the absolute directory containing the plan file (repository root when plan is at repo root).
func ProjectRoot(planFile string) (string, error) {
	abs, err := filepath.Abs(planFile)
	if err != nil {
		return "", err
	}
	return filepath.Dir(abs), nil
}
