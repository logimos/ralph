package layers

import (
	"path/filepath"
	"testing"
)

func TestResolveDataDir_default(t *testing.T) {
	root := filepath.Join(t.TempDir(), "proj")
	d := ResolveDataDir(root, "")
	want := filepath.Join(root, ".layers")
	if filepath.Clean(d) != filepath.Clean(want) {
		t.Fatalf("got %q want %q", d, want)
	}
}

func TestContextSnapshotPath(t *testing.T) {
	root := t.TempDir()
	p, err := ContextSnapshotPath(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != DefaultSnapshotFileName {
		t.Fatalf("got %q", p)
	}
	if filepath.Dir(p) != filepath.Join(root, ".layers") {
		t.Fatalf("unexpected dir in %q", p)
	}
}
