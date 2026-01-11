// Package loop provides the main execution loop for Claude Code iterations.
package loop

import (
	"context"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/cli"
)

// ClaudeClient is the interface for executing Claude Code iterations.
// This abstraction allows testing the loop without a real Claude client.
type ClaudeClient interface {
	// Execute runs a single Claude Code iteration with the given prompt.
	// Returns the result including cost, output, and any completion signal detected.
	Execute(ctx context.Context, prompt string) (*IterationResult, error)
}

// IterationResult represents the outcome of a single Claude execution.
type IterationResult struct {
	Output                string        // Claude's output text
	Cost                  float64       // Cost in USD for this iteration
	Duration              time.Duration // How long this iteration took
	CompletionSignalFound bool          // Whether completion signal was detected in output
}

// StopReason indicates why the loop stopped.
type StopReason string

const (
	StopReasonNone              StopReason = ""
	StopReasonMaxRuns           StopReason = "max_runs_reached"
	StopReasonMaxCost           StopReason = "max_cost_reached"
	StopReasonMaxDuration       StopReason = "max_duration_reached"
	StopReasonCompletionSignal  StopReason = "completion_signal"
	StopReasonConsecutiveErrors StopReason = "consecutive_errors"
	StopReasonContextCancelled  StopReason = "context_cancelled"
)

// State tracks the internal state of the loop during execution.
type State struct {
	SuccessfulIterations  int           // Count of completed iterations
	TotalIterations       int           // All iterations including errors
	ErrorCount            int           // Consecutive error counter (reset on success)
	CompletionSignalCount int           // Consecutive completion signals
	TotalCost             float64       // Accumulated USD cost
	StartTime             time.Time     // Loop start time
	LastIterationTime     time.Time     // When last iteration completed
}

// NewState creates a new State with initialized start time.
func NewState() *State {
	return &State{
		StartTime: time.Now(),
	}
}

// Elapsed returns the duration since the loop started.
func (s *State) Elapsed() time.Duration {
	return time.Since(s.StartTime)
}

// Config holds the loop configuration derived from CLI flags.
type Config struct {
	Prompt               string
	MaxRuns              int           // 0 means unlimited
	MaxCost              float64       // 0 means unlimited
	MaxDuration          time.Duration // 0 means unlimited
	CompletionSignal     string
	CompletionThreshold  int
	MaxConsecutiveErrors int // Default: 3
	DryRun               bool
	OnProgress           func(state *State) // Optional progress callback (nil allowed)
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		CompletionSignal:     "CONTINUOUS_CLAUDE_PROJECT_COMPLETE",
		CompletionThreshold:  3,
		MaxConsecutiveErrors: 3,
	}
}

// ConfigFromFlags creates a Config from CLI Flags.
func ConfigFromFlags(f *cli.Flags) *Config {
	return &Config{
		Prompt:               f.Prompt,
		MaxRuns:              f.MaxRuns,
		MaxCost:              f.MaxCost,
		MaxDuration:          f.MaxDuration,
		CompletionSignal:     f.CompletionSignal,
		CompletionThreshold:  f.CompletionThreshold,
		MaxConsecutiveErrors: 3,
		DryRun:               f.DryRun,
	}
}

// LoopResult represents the final result of the loop execution.
type LoopResult struct {
	State      *State
	StopReason StopReason
	LastError  error
}
