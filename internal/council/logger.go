package council

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DecisionLogger writes decisions to the log file.
type DecisionLogger struct {
	logFile string
	enabled bool
}

// NewDecisionLogger creates a new DecisionLogger.
func NewDecisionLogger(logFile string, enabled bool) *DecisionLogger {
	return &DecisionLogger{
		logFile: logFile,
		enabled: enabled,
	}
}

// Log writes a decision entry to the log file.
// Returns nil if logging is disabled, decision is nil, or decision is empty.
func (l *DecisionLogger) Log(decision *Decision) error {
	if !l.enabled {
		return nil
	}

	if decision == nil {
		return nil
	}

	if decision.Decision == "" && decision.Rationale == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(l.logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &CouncilError{
			Phase:   "log",
			Message: "failed to create log directory",
			Err:     err,
		}
	}

	// Open file in append mode
	f, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &CouncilError{
			Phase:   "log",
			Message: "failed to open log file",
			Err:     err,
		}
	}
	defer f.Close()

	// Format matches bash implementation (claude_loop.sh lines 1185-1193)
	entry := fmt.Sprintf(`---
timestamp: "%s"
iteration: %d
decision: "%s"
rationale: "%s"
preset: "%s"
council_invoked: %t
`, decision.Timestamp.UTC().Format(time.RFC3339),
		decision.Iteration,
		escapeYAMLString(decision.Decision),
		escapeYAMLString(decision.Rationale),
		decision.Preset,
		decision.CouncilInvoked)

	if _, err := f.WriteString(entry); err != nil {
		return &CouncilError{
			Phase:   "log",
			Message: "failed to write log entry",
			Err:     err,
		}
	}

	return nil
}

// IsEnabled returns true if logging is enabled.
func (l *DecisionLogger) IsEnabled() bool {
	return l.enabled
}

// escapeYAMLString escapes special characters for YAML string values.
func escapeYAMLString(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
