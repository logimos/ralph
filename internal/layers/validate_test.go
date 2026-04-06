package layers

import "testing"

func TestValidateCLI_httpSkipsLookPath(t *testing.T) {
	if err := ValidateCLI("nonexistent-binary-xyz", true); err != nil {
		t.Fatalf("expected nil for HTTP mode, got %v", err)
	}
}
