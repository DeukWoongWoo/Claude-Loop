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
