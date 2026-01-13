// Package reviewer provides the review pass functionality for validating
// changes made during main iterations.
package reviewer

import (
	"context"
	"time"
)

// ClaudeClient executes Claude Code iterations.
// Defined here to avoid import cycles with the loop package.
type ClaudeClient interface {
	Execute(ctx context.Context, prompt string) (*IterationResult, error)
}

// IterationResult represents the outcome of a single Claude execution.
// Mirrors loop.IterationResult to avoid import cycles.
type IterationResult struct {
	Output                string
	Cost                  float64
	Duration              time.Duration
	CompletionSignalFound bool
}

// Reviewer runs review passes.
type Reviewer interface {
	Run(ctx context.Context) (*Result, error)
}

// Config holds reviewer configuration.
type Config struct {
	ReviewPrompt         string // User's review instructions from -r flag
	MaxConsecutiveErrors int    // Threshold for aborting on repeated failures
}

// Result represents the outcome of a reviewer pass.
type Result struct {
	Output                string
	Cost                  float64
	Duration              time.Duration
	CompletionSignalFound bool
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		MaxConsecutiveErrors: 3,
	}
}

// IsEnabled returns true if reviewer is configured with a prompt.
func (c *Config) IsEnabled() bool {
	return c != nil && c.ReviewPrompt != ""
}
