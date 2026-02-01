package architecture

import (
	"context"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// ClaudeClient is an alias for planner.ClaudeClient to enable centralized adapter usage.
type ClaudeClient = planner.ClaudeClient

// IterationResult is an alias for planner.IterationResult to enable centralized adapter usage.
type IterationResult = planner.IterationResult

// Generator creates Architecture designs from PRDs.
type Generator interface {
	Generate(ctx context.Context, prd *planner.PRD) (*Architecture, error)
}

// Architecture extends planner.Architecture with additional metadata fields.
type Architecture struct {
	planner.Architecture `yaml:",inline"`

	// Extended fields
	ID        string `yaml:"id,omitempty"`
	Title     string `yaml:"title,omitempty"`
	Summary   string `yaml:"summary,omitempty"`
	Rationale string `yaml:"rationale,omitempty"`

	// Metadata
	CreatedAt time.Time     `yaml:"created_at,omitempty"`
	Cost      float64       `yaml:"cost,omitempty"`
	Duration  time.Duration `yaml:"duration,omitempty"`
}

// Config holds generator configuration.
type Config struct {
	MaxRetries     int  // Max generation attempts (default: 3) - reserved for future retry logic
	ValidateOutput bool // Enable validation after generation (default: true)
}

// DefaultConfig returns Config with default values.
func DefaultConfig() *Config {
	return &Config{
		MaxRetries:     3,
		ValidateOutput: true,
	}
}

// IsEnabled returns true if generator config is valid.
func (c *Config) IsEnabled() bool {
	return c != nil
}

// Result contains the outcome of Architecture generation.
type Result struct {
	Architecture *Architecture
	Cost         float64
	Duration     time.Duration
}
