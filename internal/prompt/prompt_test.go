package prompt

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/logimos/ralph/internal/config"
)

func TestBuildIterationPrompt_withoutSnapshot(t *testing.T) {
	cfg := &config.Config{
		PlanFile:     "plan.json",
		ProgressFile: "progress.txt",
		TypeCheckCmd: "t",
		TestCmd:      "u",
	}
	p := BuildIterationPrompt(cfg, "", "")
	if strings.Count(p, "@") < 2 {
		t.Fatalf("expected at least two @ refs, got: %s", p)
	}
	if strings.Contains(p, "context-snapshot") {
		t.Fatal("should not mention snapshot when path empty")
	}
}

func TestBuildIterationPrompt_withSnapshot(t *testing.T) {
	cfg := &config.Config{
		PlanFile:     "plan.json",
		ProgressFile: "progress.txt",
		TypeCheckCmd: "t",
		TestCmd:      "u",
	}
	snap := filepath.Join(t.TempDir(), ".layers", "context-snapshot.md")
	p := BuildIterationPrompt(cfg, snap, "")
	wantAt := "@" + filepath.Clean(snap)
	if !strings.Contains(p, wantAt) {
		t.Fatalf("expected snapshot @ ref containing %q, got: %s", wantAt, p)
	}
	if !strings.Contains(p, "bounded") {
		t.Fatal("expected hint about bounded snapshot")
	}
}

func TestBuildIterationPrompt_boundedProgressReadPath(t *testing.T) {
	cfg := &config.Config{
		PlanFile:     "plan.json",
		ProgressFile: "progress.txt",
		TypeCheckCmd: "t",
		TestCmd:      "u",
	}
	ctx := filepath.Join(t.TempDir(), "progress-context.txt")
	p := BuildIterationPrompt(cfg, "", ctx)
	wantAt := "@" + filepath.Clean(ctx)
	if !strings.Contains(p, wantAt) {
		t.Fatalf("expected @ ref %q in: %s", wantAt, p)
	}
	if !strings.Contains(p, "bounded UTF-8 tail") {
		t.Fatal("expected bounded progress hint")
	}
}

func TestBuildIterationPrompt_snapshotTrimsWhitespace(t *testing.T) {
	cfg := &config.Config{
		PlanFile:     "plan.json",
		ProgressFile: "progress.txt",
		TypeCheckCmd: "t",
		TestCmd:      "u",
	}
	snap := filepath.Join(t.TempDir(), "context-snapshot.md")
	p := BuildIterationPrompt(cfg, "  "+snap+"  ", "")
	wantAt := "@" + filepath.Clean(strings.TrimSpace(snap))
	if !strings.Contains(p, wantAt) {
		t.Fatalf("expected trimmed snapshot ref %q in: %s", wantAt, p)
	}
}
