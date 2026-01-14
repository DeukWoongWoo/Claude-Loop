// Package council provides principle conflict detection and resolution.
package council

import (
	"context"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
)

// ClaudeClient executes Claude Code iterations.
// Defined here to avoid import cycles with the loop package.
type ClaudeClient interface {
	Execute(ctx context.Context, prompt string) (*IterationResult, error)
}

// IterationResult mirrors loop.IterationResult to avoid import cycles.
type IterationResult struct {
	Output                string
	Cost                  float64
	Duration              time.Duration
	CompletionSignalFound bool
}

// Council handles principle conflict detection and resolution.
type Council interface {
	// DetectConflict checks if output contains unresolved principle conflicts.
	DetectConflict(output string) bool

	// Resolve invokes the LLM council for conflict resolution.
	Resolve(ctx context.Context, conflictContext string) (*Result, error)

	// LogDecision logs a decision to the decision log file.
	LogDecision(decision *Decision) error
}

// Config holds council configuration.
type Config struct {
	Principles   *config.Principles // Loaded principles (may be nil)
	Preset       config.Preset      // Current principle preset
	LogDecisions bool               // Whether to log decisions
	LogFile      string             // Path to decision log file
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		LogFile: ".claude/principles-decisions.log",
	}
}

// IsEnabled returns true if council is configured with principles.
func (c *Config) IsEnabled() bool {
	return c != nil && c.Principles != nil
}

// Result represents the outcome of a council resolution.
type Result struct {
	Output     string        // Council's resolution recommendation
	Cost       float64       // Cost in USD for council invocation
	Duration   time.Duration // How long the council took
	Resolution string        // Extracted resolution recommendation
	Rationale  string        // Extracted rationale
}

// Decision represents a logged decision entry.
type Decision struct {
	Timestamp      time.Time     // When the decision was made
	Iteration      int           // Which iteration this decision occurred
	Decision       string        // The decision made
	Rationale      string        // The rationale for the decision
	Preset         config.Preset // The active preset
	CouncilInvoked bool          // Whether council was invoked for this decision
}
