package loop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_Run_StopsOnMaxRuns(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxRuns:              3,
		MaxConsecutiveErrors: 3,
	}

	mock := NewMockClient()
	executor := NewExecutor(config, mock)

	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonMaxRuns, result.StopReason)
	assert.Equal(t, 3, result.State.SuccessfulIterations)
	assert.Equal(t, 3, mock.CallCount)
}

func TestExecutor_Run_StopsOnMaxCost(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxCost:              0.05,
		MaxConsecutiveErrors: 3,
	}

	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Cost: 0.02},
			{Cost: 0.02},
			{Cost: 0.02}, // Would bring total to 0.06, exceeding 0.05
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonMaxCost, result.StopReason)
	// Should stop after 2 iterations (cost = 0.04), before 3rd would exceed
	// Actually stops after cost reaches/exceeds limit
	assert.True(t, result.State.TotalCost >= 0.04)
}

func TestExecutor_Run_StopsOnMaxDuration(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxDuration:          50 * time.Millisecond,
		MaxConsecutiveErrors: 3,
	}

	// Mock that takes some time
	mock := &MockClaudeClient{
		Results: make([]*IterationResult, 100), // More than needed
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonMaxDuration, result.StopReason)
	assert.True(t, result.State.Elapsed() >= 50*time.Millisecond)
}

func TestExecutor_Run_StopsOnCompletionSignal(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxRuns:              100, // High limit
		CompletionSignal:     "DONE",
		CompletionThreshold:  3,
		MaxConsecutiveErrors: 3,
	}

	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Output: "Working..."},
			{Output: "DONE"},
			{Output: "DONE"},
			{Output: "DONE"}, // Third consecutive signal
			{Output: "more work"},
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonCompletionSignal, result.StopReason)
	assert.Equal(t, 3, result.State.CompletionSignalCount)
}

func TestExecutor_Run_StopsOnConsecutiveErrors(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxRuns:              100,
		MaxConsecutiveErrors: 3,
	}

	mock := &MockClaudeClient{
		Errors: []error{
			errors.New("error 1"),
			errors.New("error 2"),
			errors.New("error 3"),
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonConsecutiveErrors, result.StopReason)
	assert.NotNil(t, result.LastError)
	assert.Equal(t, 3, result.State.ErrorCount)
	assert.Equal(t, 0, result.State.SuccessfulIterations)
}

func TestExecutor_Run_ResetsErrorCountOnSuccess(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxRuns:              5,
		MaxConsecutiveErrors: 3,
	}

	mock := &MockClaudeClient{
		Errors: []error{
			errors.New("error 1"),
			errors.New("error 2"),
			nil, // success - resets counter
			errors.New("error 3"),
			nil, // success
		},
		Results: []*IterationResult{
			nil, nil,
			{Output: "ok", Cost: 0.01},
			nil,
			{Output: "ok", Cost: 0.01},
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	// Should complete all 5 runs without hitting consecutive error limit
	// But we need 5 *successful* runs
	assert.Equal(t, StopReasonMaxRuns, result.StopReason)
}

func TestExecutor_Run_ContextCancellation(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxRuns:              1000, // High limit
		MaxConsecutiveErrors: 3,
	}

	// Use a slow mock client that simulates work
	slowMock := &SlowMockClaudeClient{
		Delay: 5 * time.Millisecond,
	}
	executor := NewExecutor(config, slowMock)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	result, err := executor.Run(ctx)

	require.NoError(t, err)
	assert.Equal(t, StopReasonContextCancelled, result.StopReason)
	assert.ErrorIs(t, result.LastError, context.DeadlineExceeded)
}

func TestExecutor_Run_DryRun(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxRuns:              3,
		DryRun:               true,
		MaxConsecutiveErrors: 3,
	}

	mock := NewMockClient()
	executor := NewExecutor(config, mock)

	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonMaxRuns, result.StopReason)
	assert.Equal(t, 3, result.State.SuccessfulIterations)
	assert.Equal(t, 0, mock.CallCount) // Client not called in dry run
	assert.Equal(t, 0.0, result.State.TotalCost)
}

func TestExecutor_Run_OnProgress(t *testing.T) {
	var progressCalls []*State
	config := &Config{
		Prompt:               "test",
		MaxRuns:              3,
		MaxConsecutiveErrors: 3,
		OnProgress: func(state *State) {
			// Make a copy to record state at each call
			stateCopy := *state
			progressCalls = append(progressCalls, &stateCopy)
		},
	}

	mock := NewMockClient()
	executor := NewExecutor(config, mock)

	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonMaxRuns, result.StopReason)
	assert.Len(t, progressCalls, 3)

	// Verify progressive state
	assert.Equal(t, 1, progressCalls[0].SuccessfulIterations)
	assert.Equal(t, 2, progressCalls[1].SuccessfulIterations)
	assert.Equal(t, 3, progressCalls[2].SuccessfulIterations)
}

func TestExecutor_Run_OnProgressWithErrors(t *testing.T) {
	var progressCalls int
	config := &Config{
		Prompt:               "test",
		MaxRuns:              100,
		MaxConsecutiveErrors: 3,
		OnProgress: func(state *State) {
			progressCalls++
		},
	}

	mock := &MockClaudeClient{
		Errors: []error{
			errors.New("error 1"),
			errors.New("error 2"),
			errors.New("error 3"),
		},
	}

	executor := NewExecutor(config, mock)
	_, err := executor.Run(context.Background())

	require.NoError(t, err)
	// Progress should be called even for errors
	assert.Equal(t, 3, progressCalls)
}

func TestExecutor_RunOnce(t *testing.T) {
	config := &Config{
		Prompt: "test prompt",
	}

	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Output: "result", Cost: 0.05},
		},
	}

	executor := NewExecutor(config, mock)
	state := NewState()

	result, err := executor.RunOnce(context.Background(), state)

	require.NoError(t, err)
	assert.Equal(t, "result", result.Output)
	assert.Equal(t, 1, state.SuccessfulIterations)
}

func TestNewExecutor(t *testing.T) {
	config := &Config{Prompt: "test"}
	client := NewMockClient()

	executor := NewExecutor(config, client)

	assert.NotNil(t, executor)
	assert.Equal(t, config, executor.config)
	assert.NotNil(t, executor.limitChecker)
	assert.NotNil(t, executor.completionDetector)
	assert.NotNil(t, executor.iterationHandler)
}

func TestExecutor_Run_CompletionSignalResets(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxRuns:              10,
		CompletionSignal:     "DONE",
		CompletionThreshold:  3,
		MaxConsecutiveErrors: 3,
	}

	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Output: "DONE"},
			{Output: "DONE"},
			{Output: "working"}, // Reset
			{Output: "DONE"},
			{Output: "DONE"},
			{Output: "DONE"}, // Third consecutive
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonCompletionSignal, result.StopReason)
	assert.Equal(t, 6, result.State.SuccessfulIterations)
}

func TestExecutor_Run_NoLimits(t *testing.T) {
	// When no limits are set, should run until context cancellation or error
	config := &Config{
		Prompt:               "test",
		MaxConsecutiveErrors: 3,
		// No MaxRuns, MaxCost, MaxDuration
	}

	mock := &MockClaudeClient{
		// Simulate context cancellation after a few iterations
	}

	executor := NewExecutor(config, mock)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	result, err := executor.Run(ctx)

	require.NoError(t, err)
	assert.Equal(t, StopReasonContextCancelled, result.StopReason)
	assert.True(t, result.State.SuccessfulIterations > 0)
}

// --- Reviewer Integration Tests ---

func TestExecutor_WithReviewer_Success(t *testing.T) {
	config := &Config{
		Prompt:               "main task",
		MaxRuns:              2,
		MaxConsecutiveErrors: 3,
		ReviewPrompt:         "run tests",
	}

	// Mock client that returns different results for main vs reviewer iterations
	mock := &MockClaudeClient{
		Results: []*IterationResult{
			// Iteration 1: main
			{Output: "main output 1", Cost: 0.10},
			// Iteration 1: reviewer
			{Output: "review output 1", Cost: 0.02},
			// Iteration 2: main
			{Output: "main output 2", Cost: 0.10},
			// Iteration 2: reviewer
			{Output: "review output 2", Cost: 0.02},
		},
	}

	executor := NewExecutor(config, mock)

	// Verify reviewer is initialized
	assert.NotNil(t, executor.reviewer)

	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonMaxRuns, result.StopReason)
	assert.Equal(t, 2, result.State.SuccessfulIterations)

	// Each successful iteration calls main + reviewer
	assert.Equal(t, 4, mock.CallCount)

	// Total cost should include both main and reviewer
	// Main: 0.10 + 0.10 = 0.20, Reviewer: 0.02 + 0.02 = 0.04
	assert.InDelta(t, 0.24, result.State.TotalCost, 0.01)

	// Reviewer cost tracked separately
	assert.InDelta(t, 0.04, result.State.ReviewerCost, 0.01)
}

func TestExecutor_WithReviewer_ReviewerError_Continues(t *testing.T) {
	config := &Config{
		Prompt:               "main task",
		MaxRuns:              3,
		MaxConsecutiveErrors: 3,
		ReviewPrompt:         "run tests",
	}

	// First reviewer call fails, second succeeds
	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Output: "main 1", Cost: 0.10},
			nil, // Reviewer 1 will error
			{Output: "main 2", Cost: 0.10},
			{Output: "review 2", Cost: 0.02},
			{Output: "main 3", Cost: 0.10},
			{Output: "review 3", Cost: 0.02},
		},
		Errors: []error{
			nil,                           // main 1 success
			errors.New("reviewer failed"), // reviewer 1 fails
			nil,                           // main 2 success
			nil,                           // reviewer 2 success
			nil,                           // main 3 success
			nil,                           // reviewer 3 success
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonMaxRuns, result.StopReason)

	// Should have 3 successful iterations (even though 1 reviewer failed)
	assert.Equal(t, 3, result.State.SuccessfulIterations)

	// Reviewer error count should be reset after success
	assert.Equal(t, 0, result.State.ReviewerErrorCount)
}

func TestExecutor_WithReviewer_ConsecutiveErrors_Aborts(t *testing.T) {
	config := &Config{
		Prompt:               "main task",
		MaxRuns:              10,
		MaxConsecutiveErrors: 3,
		ReviewPrompt:         "run tests",
	}

	// Reviewer fails 3 times consecutively
	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Output: "main 1", Cost: 0.10},
			nil, // Reviewer 1 errors
			{Output: "main 2", Cost: 0.10},
			nil, // Reviewer 2 errors
			{Output: "main 3", Cost: 0.10},
			nil, // Reviewer 3 errors - triggers abort
		},
		Errors: []error{
			nil, errors.New("reviewer error 1"),
			nil, errors.New("reviewer error 2"),
			nil, errors.New("reviewer error 3"),
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonConsecutiveErrors, result.StopReason)
	assert.NotNil(t, result.LastError)
	assert.Equal(t, 3, result.State.ReviewerErrorCount)
}

func TestExecutor_WithReviewer_CompletionSignal(t *testing.T) {
	config := &Config{
		Prompt:               "main task",
		MaxRuns:              10,
		MaxConsecutiveErrors: 3,
		ReviewPrompt:         "check completion",
		CompletionSignal:     "DONE",
		CompletionThreshold:  3, // Need 3 consecutive signals
	}

	// Both main iteration and reviewer output contribute to signal count
	// Sequence: main1 (1), review1 (2), main2 (3) -> threshold reached after main2
	// But reviewer2 still runs, making count 4
	mock := &MockClaudeClient{
		Results: []*IterationResult{
			// Iteration 1: main has DONE (count=1), reviewer has DONE (count=2)
			{Output: "main DONE", Cost: 0.10},
			{Output: "review DONE", Cost: 0.02},
			// Iteration 2: main has DONE (count=3), reviewer has DONE (count=4)
			{Output: "main DONE", Cost: 0.10},
			{Output: "review DONE", Cost: 0.02},
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonCompletionSignal, result.StopReason)
	// Threshold is 3, but reviewer also runs, so final count is 4
	assert.True(t, result.State.CompletionSignalCount >= 3)
}

func TestExecutor_WithoutReviewer(t *testing.T) {
	config := &Config{
		Prompt:               "main task",
		MaxRuns:              2,
		MaxConsecutiveErrors: 3,
		ReviewPrompt:         "", // No reviewer
	}

	mock := NewMockClient()
	executor := NewExecutor(config, mock)

	// Verify reviewer is not initialized
	assert.Nil(t, executor.reviewer)

	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonMaxRuns, result.StopReason)
	assert.Equal(t, 2, mock.CallCount) // Only main iterations, no reviewer
	assert.Equal(t, 0.0, result.State.ReviewerCost)
}

func TestNewExecutor_WithReviewPrompt(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		ReviewPrompt:         "run tests",
		MaxConsecutiveErrors: 5,
	}
	client := NewMockClient()

	executor := NewExecutor(config, client)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.reviewer)

	// Verify reviewer config is properly set
	reviewerConfig := executor.reviewer.Config()
	assert.Equal(t, "run tests", reviewerConfig.ReviewPrompt)
	assert.Equal(t, 5, reviewerConfig.MaxConsecutiveErrors)
}

func TestNewExecutor_WithoutReviewPrompt(t *testing.T) {
	config := &Config{
		Prompt:       "test",
		ReviewPrompt: "",
	}
	client := NewMockClient()

	executor := NewExecutor(config, client)

	assert.NotNil(t, executor)
	assert.Nil(t, executor.reviewer)
}

func TestExecutor_WithReviewer_CompletionSignal_OnlyMainHasDone(t *testing.T) {
	config := &Config{
		Prompt:               "main task",
		MaxRuns:              10,
		MaxConsecutiveErrors: 3,
		ReviewPrompt:         "run tests",
		CompletionSignal:     "DONE",
		CompletionThreshold:  3,
	}

	// Main iteration outputs DONE, but reviewer outputs "Tests passed" (no DONE)
	// The completion signal count should NOT be reset by reviewer
	mock := &MockClaudeClient{
		Results: []*IterationResult{
			// Iteration 1: main DONE (count=1), reviewer no DONE (count stays 1)
			{Output: "Task DONE", Cost: 0.10},
			{Output: "All tests passed", Cost: 0.02},
			// Iteration 2: main DONE (count=2), reviewer no DONE (count stays 2)
			{Output: "More work DONE", Cost: 0.10},
			{Output: "Tests still passing", Cost: 0.02},
			// Iteration 3: main DONE (count=3), threshold reached
			{Output: "Final DONE", Cost: 0.10},
			{Output: "Validated successfully", Cost: 0.02},
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonCompletionSignal, result.StopReason)
	assert.Equal(t, 3, result.State.CompletionSignalCount)
	assert.Equal(t, 3, result.State.SuccessfulIterations)
}

func TestExecutor_WithReviewer_CompletionSignal_ReviewerAlsoHasDone(t *testing.T) {
	config := &Config{
		Prompt:               "main task",
		MaxRuns:              10,
		MaxConsecutiveErrors: 3,
		ReviewPrompt:         "check completion",
		CompletionSignal:     "DONE",
		CompletionThreshold:  3,
	}

	// Both main and reviewer output DONE - count should increment for both
	mock := &MockClaudeClient{
		Results: []*IterationResult{
			// Iteration 1: main DONE (count=1), reviewer DONE (count=2)
			{Output: "Task DONE", Cost: 0.10},
			{Output: "Review DONE", Cost: 0.02},
			// Iteration 2: main DONE (count=3), threshold reached after main
			{Output: "More DONE", Cost: 0.10},
			{Output: "Also DONE", Cost: 0.02}, // count=4
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonCompletionSignal, result.StopReason)
	// Main contributes 2, reviewer contributes 2, total = 4
	assert.Equal(t, 4, result.State.CompletionSignalCount)
}

func TestExecutor_WithReviewer_CompletionSignal_ResetOnNoDone(t *testing.T) {
	config := &Config{
		Prompt:               "main task",
		MaxRuns:              10,
		MaxConsecutiveErrors: 3,
		ReviewPrompt:         "run tests",
		CompletionSignal:     "DONE",
		CompletionThreshold:  3,
	}

	// Main outputs DONE twice, then no DONE - should reset
	mock := &MockClaudeClient{
		Results: []*IterationResult{
			// Iteration 1: main DONE (count=1), reviewer no DONE (count=1)
			{Output: "DONE", Cost: 0.10},
			{Output: "ok", Cost: 0.02},
			// Iteration 2: main DONE (count=2), reviewer no DONE (count=2)
			{Output: "DONE", Cost: 0.10},
			{Output: "ok", Cost: 0.02},
			// Iteration 3: main no DONE (count resets to 0), reviewer no DONE
			{Output: "working", Cost: 0.10},
			{Output: "ok", Cost: 0.02},
			// Iteration 4: main DONE (count=1), reviewer DONE (count=2)
			{Output: "DONE", Cost: 0.10},
			{Output: "DONE", Cost: 0.02},
			// Iteration 5: main DONE (count=3), threshold reached
			{Output: "DONE", Cost: 0.10},
			{Output: "ok", Cost: 0.02},
		},
	}

	executor := NewExecutor(config, mock)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, StopReasonCompletionSignal, result.StopReason)
	assert.Equal(t, 5, result.State.SuccessfulIterations)
}
