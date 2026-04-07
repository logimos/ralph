package prompt

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/logimos/ralph/internal/config"
)

func TestBuildIterationPrompt_planUsesPriority(t *testing.T) {
	cfg := &config.Config{
		PlanFile:     "plan.json",
		ProgressFile: "progress.txt",
		TypeCheckCmd: "t",
		TestCmd:      "u",
	}
	p := BuildIterationPrompt(cfg, "", "", true)
	if !strings.Contains(p, `"priority"`) {
		t.Fatal("expected priority wording when planUsesPriority")
	}
	if !strings.Contains(p, "first untested, non-deferred") {
		t.Fatal("expected tested/deferred constraint in priority branch")
	}
	p2 := BuildIterationPrompt(cfg, "", "", false)
	if strings.Contains(p2, `"priority"`) {
		t.Fatal("did not expect priority JSON wording when planUsesPriority false")
	}
	if !strings.Contains(p2, "file order") {
		t.Fatal("expected file order wording")
	}
}

func TestBuildIterationPrompt_withoutSnapshot(t *testing.T) {
	cfg := &config.Config{
		PlanFile:     "plan.json",
		ProgressFile: "progress.txt",
		TypeCheckCmd: "t",
		TestCmd:      "u",
	}
	p := BuildIterationPrompt(cfg, "", "", false)
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
	p := BuildIterationPrompt(cfg, snap, "", false)
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
	p := BuildIterationPrompt(cfg, "", ctx, false)
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
	p := BuildIterationPrompt(cfg, "  "+snap+"  ", "", false)
	wantAt := "@" + filepath.Clean(strings.TrimSpace(snap))
	if !strings.Contains(p, wantAt) {
		t.Fatalf("expected trimmed snapshot ref %q in: %s", wantAt, p)
	}
}
