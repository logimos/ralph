package layers

import (
	"path/filepath"
	"strings"
)

// Default snapshot file name under the Layers data directory (matches layers/src/run/resolvePaths.ts).
const DefaultSnapshotFileName = "context-snapshot.md"

// ProjectRoot returns the absolute directory containing the plan file (repository root when plan is at repo root).
func ProjectRoot(planFile string) (string, error) {
	abs, err := filepath.Abs(planFile)
	if err != nil {
		return "", err
	}
	return filepath.Dir(abs), nil
}

// ResolveDataDir returns the Layers data directory: <projectRoot>/.layers or an override
// (absolute path, or relative to project root), matching the TS resolveDataDir behavior.
func ResolveDataDir(projectRoot string, dataDirOverride string) string {
	root := filepath.Clean(projectRoot)
	o := strings.TrimSpace(dataDirOverride)
	if o == "" {
		return filepath.Join(root, ".layers")
	}
	if filepath.IsAbs(o) {
		return filepath.Clean(o)
	}
	return filepath.Join(root, o)
}

// ContextSnapshotPath returns the absolute path to context-snapshot.md under the Layers data dir.
func ContextSnapshotPath(projectRoot string, dataDirOverride string) (string, error) {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", err
	}
	dd := ResolveDataDir(root, dataDirOverride)
	absDD, err := filepath.Abs(dd)
	if err != nil {
		return "", err
	}
	return filepath.Join(absDD, DefaultSnapshotFileName), nil
}
