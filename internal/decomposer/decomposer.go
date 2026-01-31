package decomposer

import (
	"context"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// DefaultDecomposer implements the Decomposer interface.
type DefaultDecomposer struct {
	config        *Config
	client        ClaudeClient
	promptBuilder *planner.PromptBuilder
	parser        *Parser
	validator     *Validator
	scheduler     *DefaultScheduler
}

// NewDecomposer creates a new DefaultDecomposer.
// If config is nil, DefaultConfig() is used.
// Returns nil if client is nil.
func NewDecomposer(config *Config, client ClaudeClient) *DefaultDecomposer {
	if client == nil {
		return nil
	}
	if config == nil {
		config = DefaultConfig()
	}
	return &DefaultDecomposer{
		config:        config,
		client:        client,
		promptBuilder: planner.NewPromptBuilder(),
		parser:        NewParser(),
		validator:     NewValidator(),
		scheduler:     NewScheduler(),
	}
}

// Decompose creates a TaskGraph from the given Architecture.
func (d *DefaultDecomposer) Decompose(ctx context.Context, arch *planner.Architecture) (*TaskGraph, error) {
	if arch == nil {
		return nil, ErrNilArchitecture
	}

	startTime := time.Now()

	// Build prompt using planner's PromptBuilder
	prompt, err := d.promptBuilder.BuildTasksPrompt(arch)
	if err != nil {
		return nil, &DecomposerError{
			Phase:   "generate",
			Message: "failed to build prompt",
			Err:     err,
		}
	}

	// Execute Claude
	result, err := d.client.Execute(ctx, prompt)
	if err != nil {
		return nil, &DecomposerError{
			Phase:   "generate",
			Message: "claude execution failed",
			Err:     err,
		}
	}

	// Parse output
	tasks, err := d.parser.Parse(result.Output)
	if err != nil {
		return nil, err
	}

	// Validate
	if d.config.ValidateOutput {
		if err := d.validator.Validate(tasks); err != nil {
			return nil, &DecomposerError{
				Phase:   "validate",
				Message: "task validation failed",
				Err:     err,
			}
		}
	}

	// Schedule (topological sort)
	executionOrder, err := d.scheduler.Schedule(tasks)
	if err != nil {
		return nil, err
	}

	// Build TaskGraph with planner.Task slice
	plannerTasks := make([]planner.Task, len(tasks))
	for i, task := range tasks {
		plannerTasks[i] = task.Task
	}

	taskGraph := &TaskGraph{
		TaskGraph: planner.TaskGraph{
			Tasks:          plannerTasks,
			ExecutionOrder: executionOrder,
			RawOutput:      result.Output,
		},
		CreatedAt: startTime,
		Cost:      result.Cost,
		Duration:  time.Since(startTime),
	}

	return taskGraph, nil
}

// Config returns the decomposer's configuration.
func (d *DefaultDecomposer) Config() *Config {
	return d.config
}

// Compile-time interface compliance check.
var _ Decomposer = (*DefaultDecomposer)(nil)
