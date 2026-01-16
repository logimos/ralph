package ui

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"DEBUG", LogLevelDebug},
		{"info", LogLevelInfo},
		{"INFO", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"warning", LogLevelWarn},
		{"WARN", LogLevelWarn},
		{"error", LogLevelError},
		{"ERROR", LogLevelError},
		{"quiet", LogLevelQuiet},
		{"QUIET", LogLevelQuiet},
		{"", LogLevelInfo},
		{"invalid", LogLevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "debug"},
		{LogLevelInfo, "info"},
		{LogLevelWarn, "warn"},
		{LogLevelError, "error"},
		{LogLevelQuiet, "quiet"},
		{LogLevel(99), "info"}, // unknown defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("LogLevel(%d).String() = %q, want %q", tt.level, result, tt.expected)
			}
		})
	}
}

func TestUISuccess(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	ui.Success("test message %d", 42)

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Error("Success output should contain checkmark")
	}
	if !strings.Contains(output, "test message 42") {
		t.Error("Success output should contain formatted message")
	}
}

func TestUIError(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	ui.Error("error message %s", "here")

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Error("Error output should contain X mark")
	}
	if !strings.Contains(output, "error message here") {
		t.Error("Error output should contain formatted message")
	}
}

func TestUIWarn(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:   &buf,
		NoColor:  true,
		LogLevel: LogLevelWarn,
	})

	ui.Warn("warning message")

	output := buf.String()
	if !strings.Contains(output, "⚠") {
		t.Error("Warn output should contain warning symbol")
	}
	if !strings.Contains(output, "warning message") {
		t.Error("Warn output should contain message")
	}
}

func TestUIWarnSuppressedByLogLevel(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:   &buf,
		NoColor:  true,
		LogLevel: LogLevelError,
	})

	ui.Warn("warning message")

	if buf.Len() > 0 {
		t.Error("Warn should be suppressed when LogLevel > LogLevelWarn")
	}
}

func TestUIInfo(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	ui.Info("info message")

	output := buf.String()
	if !strings.Contains(output, "ℹ") {
		t.Error("Info output should contain info symbol")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info output should contain message")
	}
}

func TestUIInfoSuppressedByQuiet(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
		Quiet:   true,
	})

	ui.Info("info message")

	if buf.Len() > 0 {
		t.Error("Info should be suppressed in quiet mode")
	}
}

func TestUIDebug(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:   &buf,
		NoColor:  true,
		LogLevel: LogLevelDebug,
	})

	ui.Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "⋯") {
		t.Error("Debug output should contain ellipsis symbol")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("Debug output should contain message")
	}
}

func TestUIDebugSuppressedByDefault(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:   &buf,
		NoColor:  true,
		LogLevel: LogLevelInfo, // Default is Info
	})

	ui.Debug("debug message")

	if buf.Len() > 0 {
		t.Error("Debug should be suppressed when LogLevel > LogLevelDebug")
	}
}

func TestUIHeader(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	ui.Header("Section Title")

	output := buf.String()
	if !strings.Contains(output, "=== Section Title ===") {
		t.Errorf("Header output should contain formatted title, got: %s", output)
	}
}

func TestUIPrint(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	ui.Print("plain message")

	output := buf.String()
	if strings.TrimSpace(output) != "plain message" {
		t.Errorf("Print output = %q, want %q", output, "plain message")
	}
}

func TestUIJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:     &buf,
		NoColor:    true,
		JSONOutput: true,
	})

	ui.Info("test message")

	output := buf.String()
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry["level"] != "info" {
		t.Errorf("JSON level = %v, want 'info'", entry["level"])
	}
	if entry["message"] != "test message" {
		t.Errorf("JSON message = %v, want 'test message'", entry["message"])
	}
	if _, ok := entry["timestamp"]; !ok {
		t.Error("JSON output should contain timestamp")
	}
}

func TestUIJSONOutputError(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:     &buf,
		NoColor:    true,
		JSONOutput: true,
	})

	ui.Error("error test")

	output := buf.String()
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry["level"] != "error" {
		t.Errorf("JSON level = %v, want 'error'", entry["level"])
	}
}

func TestProgressBar(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	pb := ui.NewProgressBar(10, "Testing")
	
	if pb.total != 10 {
		t.Errorf("ProgressBar total = %d, want 10", pb.total)
	}
	if pb.current != 0 {
		t.Errorf("ProgressBar current = %d, want 0", pb.current)
	}
	
	pb.Update(5)
	if pb.current != 5 {
		t.Errorf("After Update(5), current = %d, want 5", pb.current)
	}
	
	pb.Increment()
	if pb.current != 6 {
		t.Errorf("After Increment(), current = %d, want 6", pb.current)
	}
	
	pb.SetMessage("New message")
	if pb.message != "New message" {
		t.Errorf("After SetMessage, message = %q, want 'New message'", pb.message)
	}
}

func TestSpinner(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	spinner := ui.NewSpinner("Loading")
	
	if spinner.running {
		t.Error("Spinner should not be running initially")
	}
	
	spinner.Start()
	
	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)
	
	spinner.SetMessage("Still loading")
	
	spinner.Stop()
	
	if spinner.running {
		t.Error("Spinner should not be running after Stop()")
	}
}

func TestSummary(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	summary := Summary{
		FeaturesCompleted: 5,
		FeaturesFailed:    1,
		FeaturesSkipped:   2,
		TotalIterations:   10,
		IterationsRun:     8,
		FailuresRecovered: 3,
		StartTime:         time.Now().Add(-5 * time.Minute),
		EndTime:           time.Now(),
		Errors:            []string{"Error 1", "Error 2"},
	}

	ui.PrintSummary(summary)

	output := buf.String()
	
	// Check that key information is present
	if !strings.Contains(output, "Execution Summary") {
		t.Error("Summary should contain header")
	}
	if !strings.Contains(output, "8/10 iterations") {
		t.Errorf("Summary should contain iteration progress, got: %s", output)
	}
	if !strings.Contains(output, "5") {
		t.Error("Summary should contain features completed count")
	}
	if !strings.Contains(output, "Error 1") {
		t.Error("Summary should list errors")
	}
}

func TestSummaryJSON(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:     &buf,
		NoColor:    true,
		JSONOutput: true,
	})

	summary := Summary{
		FeaturesCompleted: 5,
		TotalIterations:   10,
		IterationsRun:     8,
		StartTime:         time.Now().Add(-5 * time.Minute),
		EndTime:           time.Now(),
	}

	ui.PrintSummary(summary)

	output := buf.String()
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry["type"] != "summary" {
		t.Errorf("JSON type = %v, want 'summary'", entry["type"])
	}
	
	data, ok := entry["data"].(map[string]interface{})
	if !ok {
		t.Fatal("JSON summary should have data field")
	}
	
	if data["features_completed"].(float64) != 5 {
		t.Errorf("features_completed = %v, want 5", data["features_completed"])
	}
}

func TestSummaryQuiet(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
		Quiet:   true,
	})

	summary := Summary{
		FeaturesCompleted: 5,
		TotalIterations:   10,
	}

	ui.PrintSummary(summary)

	if buf.Len() > 0 {
		t.Error("Summary should not be printed in quiet mode")
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	table := ui.NewTable("ID", "Name", "Status")
	table.AddRow("1", "Feature A", "Complete")
	table.AddRow("2", "Feature B", "Pending")
	table.Render()

	output := buf.String()
	
	if !strings.Contains(output, "ID") {
		t.Error("Table should contain headers")
	}
	if !strings.Contains(output, "Feature A") {
		t.Error("Table should contain row data")
	}
	if !strings.Contains(output, "Complete") {
		t.Error("Table should contain status")
	}
}

func TestTableJSON(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:     &buf,
		NoColor:    true,
		JSONOutput: true,
	})

	table := ui.NewTable("ID", "Name")
	table.AddRow("1", "Test")
	table.Render()

	output := buf.String()
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry["type"] != "table" {
		t.Errorf("JSON type = %v, want 'table'", entry["type"])
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30.0s"},
		{90 * time.Second, "1m30s"},
		{2*time.Hour + 30*time.Minute, "2h30m"},
		{500 * time.Millisecond, "0.5s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.NoColor {
		t.Error("Default NoColor should be false")
	}
	if cfg.Quiet {
		t.Error("Default Quiet should be false")
	}
	if cfg.JSONOutput {
		t.Error("Default JSONOutput should be false")
	}
	if cfg.LogLevel != LogLevelInfo {
		t.Errorf("Default LogLevel = %v, want LogLevelInfo", cfg.LogLevel)
	}
}

func TestUIHelpers(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:     &buf,
		NoColor:    true,
		Quiet:      true,
		JSONOutput: true,
	})

	if !ui.IsQuiet() {
		t.Error("IsQuiet() should return true when Quiet is set")
	}
	if !ui.IsJSONOutput() {
		t.Error("IsJSONOutput() should return true when JSONOutput is set")
	}
	// Note: IsTTY() depends on the actual file descriptor, so we don't test it here
}

func TestColorDisabledForNonTTY(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer: &buf,
		// NoColor not explicitly set, but buffer is not a TTY
	})

	// Colors should be disabled for non-TTY
	if !ui.config.NoColor {
		t.Error("NoColor should be true for non-TTY writer")
	}
}

func TestStatusLine(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	ui.StatusLine(3, 10, "Working on feature X")

	output := buf.String()
	if !strings.Contains(output, "3/10") {
		t.Errorf("StatusLine should contain iteration info, got: %s", output)
	}
	if !strings.Contains(output, "feature X") {
		t.Error("StatusLine should contain feature name")
	}
}

func TestSubHeader(t *testing.T) {
	var buf bytes.Buffer
	ui := New(OutputConfig{
		Writer:  &buf,
		NoColor: true,
	})

	ui.SubHeader("Sub Section")

	output := buf.String()
	if !strings.Contains(output, "--- Sub Section ---") {
		t.Errorf("SubHeader output should contain formatted sub-title, got: %s", output)
	}
}
