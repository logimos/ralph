// Package ui provides enhanced CLI output formatting with colors, spinners,
// progress bars, and structured logging for Ralph.
package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// LogLevelDebug is for detailed debugging information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for general information
	LogLevelInfo
	// LogLevelWarn is for warning messages
	LogLevelWarn
	// LogLevelError is for error messages
	LogLevelError
	// LogLevelQuiet suppresses all output except errors
	LogLevelQuiet
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// Spinner animation frames
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// OutputConfig holds configuration for UI output
type OutputConfig struct {
	NoColor    bool
	Quiet      bool
	JSONOutput bool
	LogLevel   LogLevel
	Writer     io.Writer
}

// UI handles all formatted output for Ralph
type UI struct {
	config    OutputConfig
	mu        sync.Mutex
	isTTY     bool
	spinnerCh chan struct{}
}

// New creates a new UI instance with the given configuration
func New(cfg OutputConfig) *UI {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	// Detect if output is a TTY
	isTTY := false
	if f, ok := cfg.Writer.(*os.File); ok {
		isTTY = term.IsTerminal(int(f.Fd()))
	}

	// Disable colors if not a TTY or NoColor is set
	if !isTTY || cfg.NoColor {
		cfg.NoColor = true
	}

	return &UI{
		config: cfg,
		isTTY:  isTTY,
	}
}

// DefaultConfig returns the default UI configuration
func DefaultConfig() OutputConfig {
	return OutputConfig{
		NoColor:    false,
		Quiet:      false,
		JSONOutput: false,
		LogLevel:   LogLevelInfo,
		Writer:     os.Stdout,
	}
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	case "quiet":
		return LogLevelQuiet
	default:
		return LogLevelInfo
	}
}

// LogLevelString returns the string representation of a LogLevel
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelQuiet:
		return "quiet"
	default:
		return "info"
	}
}

// color wraps text with ANSI color codes if colors are enabled
func (u *UI) color(c, text string) string {
	if u.config.NoColor {
		return text
	}
	return c + text + colorReset
}

// Success prints a success message in green
func (u *UI) Success(format string, args ...interface{}) {
	if u.config.Quiet && u.config.LogLevel > LogLevelInfo {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	if u.config.JSONOutput {
		u.writeJSON("success", msg)
	} else {
		fmt.Fprintf(u.config.Writer, "%s %s\n", u.color(colorGreen, "✓"), msg)
	}
}

// Error prints an error message in red
func (u *UI) Error(format string, args ...interface{}) {
	u.mu.Lock()
	defer u.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	if u.config.JSONOutput {
		u.writeJSON("error", msg)
	} else {
		fmt.Fprintf(u.config.Writer, "%s %s\n", u.color(colorRed, "✗"), msg)
	}
}

// Warn prints a warning message in yellow
func (u *UI) Warn(format string, args ...interface{}) {
	if u.config.LogLevel > LogLevelWarn {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	if u.config.JSONOutput {
		u.writeJSON("warning", msg)
	} else {
		fmt.Fprintf(u.config.Writer, "%s %s\n", u.color(colorYellow, "⚠"), msg)
	}
}

// Info prints an informational message
func (u *UI) Info(format string, args ...interface{}) {
	if u.config.Quiet || u.config.LogLevel > LogLevelInfo {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	if u.config.JSONOutput {
		u.writeJSON("info", msg)
	} else {
		fmt.Fprintf(u.config.Writer, "%s %s\n", u.color(colorBlue, "ℹ"), msg)
	}
}

// Debug prints a debug message in gray
func (u *UI) Debug(format string, args ...interface{}) {
	if u.config.LogLevel > LogLevelDebug {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	if u.config.JSONOutput {
		u.writeJSON("debug", msg)
	} else {
		fmt.Fprintf(u.config.Writer, "%s %s\n", u.color(colorGray, "⋯"), msg)
	}
}

// Print prints a plain message without any formatting
func (u *UI) Print(format string, args ...interface{}) {
	if u.config.Quiet {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	if u.config.JSONOutput {
		u.writeJSON("output", msg)
	} else {
		fmt.Fprintln(u.config.Writer, msg)
	}
}

// Printf prints a formatted message without newline
func (u *UI) Printf(format string, args ...interface{}) {
	if u.config.Quiet {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.config.JSONOutput {
		u.writeJSON("output", fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(u.config.Writer, format, args...)
	}
}

// Header prints a header/section title
func (u *UI) Header(format string, args ...interface{}) {
	if u.config.Quiet {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	if u.config.JSONOutput {
		u.writeJSON("header", msg)
	} else {
		fmt.Fprintf(u.config.Writer, "\n%s\n", u.color(colorBold+colorCyan, "=== "+msg+" ==="))
	}
}

// SubHeader prints a sub-header
func (u *UI) SubHeader(format string, args ...interface{}) {
	if u.config.Quiet {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	if u.config.JSONOutput {
		u.writeJSON("subheader", msg)
	} else {
		fmt.Fprintf(u.config.Writer, "%s\n", u.color(colorBold, "--- "+msg+" ---"))
	}
}

// writeJSON outputs a message in JSON format
func (u *UI) writeJSON(level, message string) {
	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level,
		"message":   message,
	}
	data, _ := json.Marshal(entry)
	fmt.Fprintln(u.config.Writer, string(data))
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	ui       *UI
	total    int
	current  int
	message  string
	width    int
	mu       sync.Mutex
	started  time.Time
}

// NewProgressBar creates a new progress bar
func (u *UI) NewProgressBar(total int, message string) *ProgressBar {
	return &ProgressBar{
		ui:      u,
		total:   total,
		message: message,
		width:   40,
		started: time.Now(),
	}
}

// Update updates the progress bar to a new value
func (pb *ProgressBar) Update(current int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current = current
	pb.render()
}

// Increment increments the progress by 1
func (pb *ProgressBar) Increment() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current++
	pb.render()
}

// SetMessage updates the progress bar message
func (pb *ProgressBar) SetMessage(msg string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.message = msg
	pb.render()
}

// Complete marks the progress bar as complete
func (pb *ProgressBar) Complete() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current = pb.total
	pb.render()
	fmt.Fprintln(pb.ui.config.Writer) // Move to next line
}

// render draws the progress bar
func (pb *ProgressBar) render() {
	if pb.ui.config.Quiet || pb.ui.config.JSONOutput {
		return
	}

	// Only show progress bar on TTY
	if !pb.ui.isTTY {
		return
	}

	percent := float64(pb.current) / float64(pb.total)
	if percent > 1 {
		percent = 1
	}

	filled := int(percent * float64(pb.width))
	empty := pb.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	elapsed := time.Since(pb.started)
	
	// Estimate remaining time
	var eta string
	if pb.current > 0 && pb.current < pb.total {
		perItem := elapsed / time.Duration(pb.current)
		remaining := perItem * time.Duration(pb.total-pb.current)
		eta = fmt.Sprintf(" ETA: %s", formatDuration(remaining))
	}

	// Clear line and render progress bar
	fmt.Fprintf(pb.ui.config.Writer, "\r\033[K%s [%s] %d/%d (%.0f%%)%s",
		pb.message, pb.ui.color(colorGreen, bar), pb.current, pb.total, percent*100, eta)
}

// Spinner represents a loading spinner
type Spinner struct {
	ui      *UI
	message string
	stopCh  chan struct{}
	doneCh  chan struct{}
	mu      sync.Mutex
	running bool
}

// NewSpinner creates a new spinner
func (u *UI) NewSpinner(message string) *Spinner {
	return &Spinner{
		ui:      u,
		message: message,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.doneCh)

		if s.ui.config.Quiet || s.ui.config.JSONOutput || !s.ui.isTTY {
			return
		}

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		frame := 0
		for {
			select {
			case <-s.stopCh:
				fmt.Fprintf(s.ui.config.Writer, "\r\033[K") // Clear the line
				return
			case <-ticker.C:
				fmt.Fprintf(s.ui.config.Writer, "\r\033[K%s %s",
					s.ui.color(colorCyan, spinnerFrames[frame]), s.message)
				frame = (frame + 1) % len(spinnerFrames)
			}
		}
	}()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh
}

// SetMessage updates the spinner message
func (s *Spinner) SetMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}

// Summary holds information for the summary dashboard
type Summary struct {
	FeaturesCompleted int
	FeaturesFailed    int
	FeaturesSkipped   int
	TotalIterations   int
	IterationsRun     int
	FailuresRecovered int
	StartTime         time.Time
	EndTime           time.Time
	Errors            []string
}

// PrintSummary displays a summary dashboard at the end of execution
func (u *UI) PrintSummary(s Summary) {
	if u.config.Quiet {
		return
	}

	duration := s.EndTime.Sub(s.StartTime)

	if u.config.JSONOutput {
		summaryJSON := map[string]interface{}{
			"features_completed":  s.FeaturesCompleted,
			"features_failed":     s.FeaturesFailed,
			"features_skipped":    s.FeaturesSkipped,
			"total_iterations":    s.TotalIterations,
			"iterations_run":      s.IterationsRun,
			"failures_recovered":  s.FailuresRecovered,
			"duration_seconds":    duration.Seconds(),
			"errors":              s.Errors,
		}
		data, _ := json.Marshal(map[string]interface{}{"type": "summary", "data": summaryJSON})
		fmt.Fprintln(u.config.Writer, string(data))
		return
	}

	u.Header("Execution Summary")

	// Create a simple box
	boxWidth := 45
	line := strings.Repeat("─", boxWidth-2)
	
	fmt.Fprintf(u.config.Writer, "┌%s┐\n", line)
	
	// Progress
	fmt.Fprintf(u.config.Writer, "│ %-20s %20s │\n", "Progress:", 
		fmt.Sprintf("%d/%d iterations", s.IterationsRun, s.TotalIterations))
	
	// Features
	fmt.Fprintf(u.config.Writer, "│ %-20s %20s │\n", "Features completed:",
		u.color(colorGreen, fmt.Sprintf("%d", s.FeaturesCompleted)))
	
	if s.FeaturesFailed > 0 {
		fmt.Fprintf(u.config.Writer, "│ %-20s %20s │\n", "Features failed:",
			u.color(colorRed, fmt.Sprintf("%d", s.FeaturesFailed)))
	}
	
	if s.FeaturesSkipped > 0 {
		fmt.Fprintf(u.config.Writer, "│ %-20s %20s │\n", "Features skipped:",
			u.color(colorYellow, fmt.Sprintf("%d", s.FeaturesSkipped)))
	}
	
	if s.FailuresRecovered > 0 {
		fmt.Fprintf(u.config.Writer, "│ %-20s %20s │\n", "Failures recovered:",
			fmt.Sprintf("%d", s.FailuresRecovered))
	}
	
	// Duration
	fmt.Fprintf(u.config.Writer, "│ %-20s %20s │\n", "Duration:",
		formatDuration(duration))
	
	fmt.Fprintf(u.config.Writer, "└%s┘\n", line)

	// List errors if any
	if len(s.Errors) > 0 {
		fmt.Fprintln(u.config.Writer)
		u.SubHeader("Errors Encountered")
		for _, err := range s.Errors {
			fmt.Fprintf(u.config.Writer, "  %s %s\n", u.color(colorRed, "•"), err)
		}
	}
}

// StatusLine displays a real-time status line
func (u *UI) StatusLine(iteration, total int, feature string) {
	if u.config.Quiet || u.config.JSONOutput {
		return
	}

	if !u.isTTY {
		// For non-TTY, just print a simple line
		u.Print("Iteration %d/%d: %s", iteration, total, feature)
		return
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	// Clear line and print status
	fmt.Fprintf(u.config.Writer, "\r\033[K%s Iteration %s: %s",
		u.color(colorBlue, "▶"),
		u.color(colorBold, fmt.Sprintf("%d/%d", iteration, total)),
		feature)
}

// ClearLine clears the current line (for use after StatusLine)
func (u *UI) ClearLine() {
	if u.config.Quiet || !u.isTTY {
		return
	}
	fmt.Fprintf(u.config.Writer, "\r\033[K")
}

// IsTTY returns whether the output is a terminal
func (u *UI) IsTTY() bool {
	return u.isTTY
}

// IsQuiet returns whether quiet mode is enabled
func (u *UI) IsQuiet() bool {
	return u.config.Quiet
}

// IsJSONOutput returns whether JSON output mode is enabled
func (u *UI) IsJSONOutput() bool {
	return u.config.JSONOutput
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

// Table represents a simple table for output
type Table struct {
	ui      *UI
	headers []string
	rows    [][]string
}

// NewTable creates a new table
func (u *UI) NewTable(headers ...string) *Table {
	return &Table{
		ui:      u,
		headers: headers,
		rows:    make([][]string, 0),
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells ...string) {
	t.rows = append(t.rows, cells)
}

// Render outputs the table
func (t *Table) Render() {
	if t.ui.config.Quiet {
		return
	}

	if t.ui.config.JSONOutput {
		tableData := make([]map[string]string, len(t.rows))
		for i, row := range t.rows {
			rowMap := make(map[string]string)
			for j, cell := range row {
				if j < len(t.headers) {
					rowMap[t.headers[j]] = cell
				}
			}
			tableData[i] = rowMap
		}
		data, _ := json.Marshal(map[string]interface{}{"type": "table", "data": tableData})
		fmt.Fprintln(t.ui.config.Writer, string(data))
		return
	}

	// Calculate column widths
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(h)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range t.headers {
		fmt.Fprintf(t.ui.config.Writer, "%-*s  ", widths[i], t.ui.color(colorBold, h))
	}
	fmt.Fprintln(t.ui.config.Writer)

	// Print separator
	for i := range t.headers {
		fmt.Fprintf(t.ui.config.Writer, "%s  ", strings.Repeat("-", widths[i]))
	}
	fmt.Fprintln(t.ui.config.Writer)

	// Print rows
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(t.ui.config.Writer, "%-*s  ", widths[i], cell)
			}
		}
		fmt.Fprintln(t.ui.config.Writer)
	}
}
