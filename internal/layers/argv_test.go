package layers

import (
	"testing"
)

func TestParseCommandLine(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		n, r := parseCommandLine("layers")
		if n != "layers" || len(r) != 0 {
			t.Fatalf("got %q %v", n, r)
		}
	})
	t.Run("node and path", func(t *testing.T) {
		n, r := parseCommandLine(`node /tmp/foo/main.js`)
		if n != "node" || len(r) != 1 || r[0] != "/tmp/foo/main.js" {
			t.Fatalf("got %q %v", n, r)
		}
	})
	t.Run("quoted path with spaces", func(t *testing.T) {
		n, r := parseCommandLine(`node "/tmp/my path/main.js"`)
		if n != "node" || len(r) != 1 || r[0] != "/tmp/my path/main.js" {
			t.Fatalf("got %q %v", n, r)
		}
	})
}
