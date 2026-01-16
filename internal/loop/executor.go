package loop

import (
	"context"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/council"
	"github.com/DeukWoongWoo/claude-loop/internal/reviewer"
)

// reviewerClientAdapter adapts loop.ClaudeClient to reviewer.ClaudeClient
// to avoid import cycles between packages.
type reviewerClientAdapter struct {
	client ClaudeClient
}

func (a *reviewerClientAdapter) Execute(ctx context.Context, prompt string) (*reviewer.IterationResult, error) {
	result, err := a.client.Execute(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return &reviewer.IterationResult{
		Output:                result.Output,
		Cost:                  result.Cost,
		Duration:              result.Duration,
		CompletionSignalFound: result.CompletionSignalFound,
	}, nil
}

// councilClientAdapter adapts loop.ClaudeClient to council.ClaudeClient
// to avoid import cycles between packages.
type councilClientAdapter struct {
	client ClaudeClient
}

func (a *councilClientAdapter) Execute(ctx context.Context, prompt string) (*council.IterationResult, error) {
	result, err := a.client.Execute(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return &council.IterationResult{
		Output:                result.Output,
		Cost:                  result.Cost,
		Duration:              result.Duration,
		CompletionSignalFound: result.CompletionSignalFound,
	}, nil
}

// Executor is the main loop executor that orchestrates iteration execution.
type Executor struct {
	config             *Config
	limitChecker       *LimitChecker
	completionDetector *CompletionDetector
	iterationHandler   *IterationHandler
	reviewer           *reviewer.DefaultReviewer
	council            *council.DefaultCouncil
}

// NewExecutor creates a new Executor with the given configuration and client.
func NewExecutor(config *Config, client ClaudeClient) *Executor {
	e := &Executor{
		config:             config,
		limitChecker:       NewLimitChecker(config),
		completionDetector: NewCompletionDetector(config),
		iterationHandler:   NewIterationHandler(config, client),
	}

	// Initialize reviewer if review prompt is provided
	if config.ReviewPrompt != "" {
		e.reviewer = reviewer.NewReviewer(&reviewer.Config{
			ReviewPrompt:         config.ReviewPrompt,
			MaxConsecutiveErrors: config.MaxConsecutiveErrors,
		}, &reviewerClientAdapter{client: client})
	}

	// Initialize council if principles are loaded
	if config.Principles != nil {
		e.council = council.NewCouncil(&council.Config{
			Principles:   config.Principles,
			Preset:       config.Principles.Preset,
			LogDecisions: config.LogDecisions,
			LogFile:      ".claude/principles-decisions.log",
		}, &councilClientAdapter{client: client})
	}

	return e
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
		iterResult, err := e.iterationHandler.Execute(ctx, state)

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

		// Handle principle conflict detection and council invocation (skip in dry-run)
		if e.council != nil && !e.config.DryRun {
			e.handleCouncil(ctx, state, iterResult.Output)
		}

		// Run reviewer pass if configured (skip in dry-run)
		if e.reviewer != nil && !e.config.DryRun {
			if reviewErr := e.runReviewerPass(ctx, state); reviewErr != nil {
				return &LoopResult{
					State:      state,
					StopReason: StopReasonConsecutiveErrors,
					LastError:  reviewErr,
				}, nil
			}
		}

		// Call progress callback after successful iteration (and review)
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

// runReviewerPass executes a reviewer pass and updates state.
// Returns nil to continue the loop, or an error if the loop should stop due to consecutive errors.
func (e *Executor) runReviewerPass(ctx context.Context, state *State) error {
	reviewResult, err := e.reviewer.Run(ctx)
	if err != nil {
		state.ReviewerErrorCount++

		// Stop if too many consecutive reviewer errors
		if state.ReviewerErrorCount >= e.config.MaxConsecutiveErrors {
			return err
		}
		return nil
	}

	// Success: accumulate cost, reset error count
	state.ReviewerCost += reviewResult.Cost
	state.TotalCost += reviewResult.Cost
	state.ReviewerErrorCount = 0

	// Check for completion signal in reviewer output
	// Only increment if found; do NOT reset on absence (main iteration already updated state)
	if e.completionDetector.Detect(reviewResult.Output) {
		state.CompletionSignalCount++
	}

	return nil
}

// handleCouncil checks for principle conflicts and invokes council if needed.
// This is advisory and does not block the loop on failure (graceful degradation).
func (e *Executor) handleCouncil(ctx context.Context, state *State, output string) {
	// Check for unresolved conflicts first
	hasConflict := e.council.DetectConflict(output)

	if hasConflict {
		// Invoke council for resolution
		result, err := e.council.Resolve(ctx, output)
		if err != nil {
			// Log failure but don't block - council is advisory
			return
		}

		// Update state with council cost
		state.CouncilCost += result.Cost
		state.TotalCost += result.Cost
		state.CouncilInvocations++

		// Log the council decision (not the original conflicting decision)
		_ = e.council.LogDecision(&council.Decision{
			Timestamp:      time.Now(),
			Iteration:      state.TotalIterations,
			Decision:       result.Resolution,
			Rationale:      result.Rationale,
			Preset:         e.config.Principles.Preset,
			CouncilInvoked: true,
		})
		return
	}

	// No conflict - extract and log any decisions from normal output
	decision, rationale := e.council.ExtractDecisionFromOutput(output)
	if decision != "" || rationale != "" {
		_ = e.council.LogDecision(&council.Decision{
			Timestamp:      time.Now(),
			Iteration:      state.TotalIterations,
			Decision:       decision,
			Rationale:      rationale,
			Preset:         e.config.Principles.Preset,
			CouncilInvoked: false,
		})
	}
}
