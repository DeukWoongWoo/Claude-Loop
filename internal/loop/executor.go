package loop

import (
	"context"
)

// Executor is the main loop executor that orchestrates iteration execution.
type Executor struct {
	config             *Config
	limitChecker       *LimitChecker
	completionDetector *CompletionDetector
	iterationHandler   *IterationHandler
}

// NewExecutor creates a new Executor with the given configuration and client.
func NewExecutor(config *Config, client ClaudeClient) *Executor {
	return &Executor{
		config:             config,
		limitChecker:       NewLimitChecker(config),
		completionDetector: NewCompletionDetector(config),
		iterationHandler:   NewIterationHandler(config, client),
	}
}

// Run executes the main loop until a stop condition is met.
// Returns the final state and the reason for stopping.
func (e *Executor) Run(ctx context.Context) (*LoopResult, error) {
	state := NewState()

	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return &LoopResult{
				State:      state,
				StopReason: StopReasonContextCancelled,
				LastError:  ctx.Err(),
			}, nil
		default:
			// Continue with iteration
		}

		// Check if any limits have been reached BEFORE starting iteration
		if result := e.limitChecker.Check(state); result.LimitReached {
			return &LoopResult{
				State:      state,
				StopReason: result.Reason,
			}, nil
		}

		// Check completion threshold
		if result := e.completionDetector.CheckThreshold(state); result.LimitReached {
			return &LoopResult{
				State:      state,
				StopReason: result.Reason,
			}, nil
		}

		// Execute single iteration
		_, err := e.iterationHandler.Execute(ctx, state)

		if err != nil {
			shouldContinue := e.iterationHandler.HandleError(state, err)

			// Call progress callback after error handling (so ErrorCount is updated)
			if e.config.OnProgress != nil {
				e.config.OnProgress(state)
			}

			if !shouldContinue {
				return &LoopResult{
					State:      state,
					StopReason: StopReasonConsecutiveErrors,
					LastError:  err,
				}, nil
			}
			// Continue to next iteration after error
			continue
		}

		// Call progress callback after successful iteration
		if e.config.OnProgress != nil {
			e.config.OnProgress(state)
		}

		// Check limits again after iteration (cost may have changed)
		if result := e.limitChecker.Check(state); result.LimitReached {
			return &LoopResult{
				State:      state,
				StopReason: result.Reason,
			}, nil
		}

		// Check completion threshold after iteration
		if result := e.completionDetector.CheckThreshold(state); result.LimitReached {
			return &LoopResult{
				State:      state,
				StopReason: result.Reason,
			}, nil
		}
	}
}

// RunOnce executes a single iteration (useful for testing).
func (e *Executor) RunOnce(ctx context.Context, state *State) (*IterationResult, error) {
	return e.iterationHandler.Execute(ctx, state)
}
