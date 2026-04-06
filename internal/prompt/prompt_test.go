package prompt

import (
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
	p := BuildIterationPrompt(cfg, "")
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
	p := BuildIterationPrompt(cfg, "/tmp/.layers/context-snapshot.md")
	if !strings.Contains(p, "@/tmp/.layers/context-snapshot.md") {
		t.Fatalf("expected snapshot @ ref: %s", p)
	}
	if !strings.Contains(p, "bounded") {
		t.Fatal("expected hint about bounded snapshot")
	}
}
