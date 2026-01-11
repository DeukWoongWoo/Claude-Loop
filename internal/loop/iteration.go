package loop

import (
	"context"
	"time"
)

// IterationHandler handles the execution of a single iteration.
type IterationHandler struct {
	config             *Config
	client             ClaudeClient
	completionDetector *CompletionDetector
}

// NewIterationHandler creates a new IterationHandler.
func NewIterationHandler(config *Config, client ClaudeClient) *IterationHandler {
	return &IterationHandler{
		config:             config,
		client:             client,
		completionDetector: NewCompletionDetector(config),
	}
}

// Execute runs a single iteration and updates the state.
// Returns the result and any error that occurred.
func (ih *IterationHandler) Execute(ctx context.Context, state *State) (*IterationResult, error) {
	state.TotalIterations++

	// Dry run mode - simulate successful execution
	if ih.config.DryRun {
		return ih.executeDryRun(state)
	}

	// Execute Claude iteration
	result, err := ih.client.Execute(ctx, ih.config.Prompt)
	if err != nil {
		return nil, &IterationError{
			Iteration: state.TotalIterations,
			Message:   "claude execution failed",
			Err:       err,
		}
	}

	// Update state based on result
	ih.updateStateOnSuccess(state, result)

	return result, nil
}

func (ih *IterationHandler) executeDryRun(state *State) (*IterationResult, error) {
	result := &IterationResult{
		Output:                "[dry-run] Simulated execution",
		Cost:                  0,
		Duration:              0,
		CompletionSignalFound: false,
	}
	ih.updateStateOnSuccess(state, result)
	return result, nil
}

func (ih *IterationHandler) updateStateOnSuccess(state *State, result *IterationResult) {
	state.SuccessfulIterations++
	state.TotalCost += result.Cost
	state.LastIterationTime = time.Now()
	state.ErrorCount = 0 // Reset consecutive error count on success

	// Check for completion signal
	signalFound := ih.completionDetector.Detect(result.Output)
	result.CompletionSignalFound = signalFound
	ih.completionDetector.UpdateState(state, signalFound)
}

// HandleError processes an iteration error and updates the state.
// Returns true if the loop should continue, false if it should stop.
func (ih *IterationHandler) HandleError(state *State, err error) bool {
	state.ErrorCount++

	// Stop if we've hit max consecutive errors
	if state.ErrorCount >= ih.config.MaxConsecutiveErrors {
		return false
	}

	return true
}
