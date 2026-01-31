// Package decomposer provides task decomposition and scheduling capabilities.
// It transforms Architecture designs into executable TaskGraphs with validated
// dependencies and topologically-sorted execution order.
package decomposer

import (
	"context"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
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

// Decomposer creates TaskGraphs from Architecture designs.
type Decomposer interface {
	Decompose(ctx context.Context, arch *planner.Architecture) (*TaskGraph, error)
}

// TaskGraph extends planner.TaskGraph with additional metadata fields.
type TaskGraph struct {
	planner.TaskGraph `yaml:",inline"`

	// Extended fields
	ID              string `yaml:"id,omitempty"`
	ArchitectureRef string `yaml:"architecture_ref,omitempty"` // Reference to source architecture

	// Metadata
	CreatedAt time.Time     `yaml:"created_at,omitempty"`
	Cost      float64       `yaml:"cost,omitempty"`
	Duration  time.Duration `yaml:"duration,omitempty"`
}

// Task extends planner.Task with scheduling and execution fields.
type Task struct {
	planner.Task `yaml:",inline"`

	// Extended fields
	SuccessCriteria []string `yaml:"success_criteria,omitempty"`
	Complexity      string   `yaml:"complexity,omitempty"` // small, medium, large

	// Timing fields (for execution tracking)
	StartedAt   *time.Time `yaml:"started_at,omitempty"`
	CompletedAt *time.Time `yaml:"completed_at,omitempty"`
}

// Complexity constants.
const (
	ComplexitySmall  = "small"
	ComplexityMedium = "medium"
	ComplexityLarge  = "large"
)

// Config holds decomposer configuration.
type Config struct {
	TaskDir        string // Directory for task files (default: .claude/tasks)
	MaxRetries     int    // Max generation attempts (default: 3)
	ValidateOutput bool   // Enable validation after generation (default: true)
}

// DefaultConfig returns Config with default values.
func DefaultConfig() *Config {
	return &Config{
		TaskDir:        ".claude/tasks",
		MaxRetries:     3,
		ValidateOutput: true,
	}
}

// IsEnabled returns true if decomposer config is valid.
func (c *Config) IsEnabled() bool {
	return c != nil
}

// Result contains the outcome of decomposition.
type Result struct {
	TaskGraph *TaskGraph
	Cost      float64
	Duration  time.Duration
}
