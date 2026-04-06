package layers

import (
	"fmt"
	"os/exec"
	"strings"
)

// ValidateCLI checks that the Layers CLI can be resolved when not using HTTP mode.
func ValidateCLI(command string, useHTTP bool) error {
	if useHTTP {
		return nil
	}
	cmd := strings.TrimSpace(command)
	if cmd == "" {
		_, err := exec.LookPath(defaultCLI)
		if err != nil {
			return fmt.Errorf("%q not found in PATH: %w", defaultCLI, err)
		}
		return nil
	}
	name, _ := parseCommandLine(cmd)
	if name == "" {
		return fmt.Errorf("empty layers-command after parsing")
	}
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%q not found in PATH: %w", name, err)
	}
	return nil
}
