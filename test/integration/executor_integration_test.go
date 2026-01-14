package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/loop"
	"github.com/DeukWoongWoo/claude-loop/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_BasicLoop_Integration(t *testing.T) {
	t.Run("runs until max runs reached", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "work 1", Cost: 0.10},
			{Output: "work 2", Cost: 0.10},
			{Output: "work 3", Cost: 0.10},
		}

		config := &loop.Config{
			Prompt:               "test prompt",
			MaxRuns:              3,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonMaxRuns, result.StopReason)
		assert.Equal(t, 3, result.State.SuccessfulIterations)
		assert.Equal(t, 3, mockClient.CallCount())
		assert.InDelta(t, 0.30, result.State.TotalCost, 0.001)
	})

	t.Run("stops on max cost reached", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "work 1", Cost: 0.40},
			{Output: "work 2", Cost: 0.40},
			{Output: "work 3", Cost: 0.40},
		}

		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              10,
			MaxCost:              1.00,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonMaxCost, result.StopReason)
		// Cost check happens after iteration, so may exceed slightly
		assert.GreaterOrEqual(t, result.State.TotalCost, 1.00)
	})
}

func TestExecutor_WithReviewer_Integration(t *testing.T) {
	t.Run("reviewer runs after each successful iteration", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "main work 1", Cost: 0.10}, // main 1
			{Output: "tests pass", Cost: 0.02},  // reviewer 1
			{Output: "main work 2", Cost: 0.10}, // main 2
			{Output: "tests pass", Cost: 0.02},  // reviewer 2
		}

		config := &loop.Config{
			Prompt:               "implement feature",
			MaxRuns:              2,
			ReviewPrompt:         "run npm test",
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonMaxRuns, result.StopReason)
		assert.Equal(t, 2, result.State.SuccessfulIterations)
		assert.Equal(t, 4, mockClient.CallCount()) // 2 main + 2 reviewer

		assert.InDelta(t, 0.24, result.State.TotalCost, 0.001)
		assert.InDelta(t, 0.04, result.State.ReviewerCost, 0.001)
	})

	t.Run("reviewer errors tracked separately", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "main work 1", Cost: 0.10},
			{Err: errors.New("reviewer failed")},
			{Output: "main work 2", Cost: 0.10},
			{Err: errors.New("reviewer failed")},
			{Output: "main work 3", Cost: 0.10},
			{Err: errors.New("reviewer failed")}, // 3rd consecutive error
		}

		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              10,
			ReviewPrompt:         "check",
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonConsecutiveErrors, result.StopReason)
		assert.Equal(t, 3, result.State.SuccessfulIterations)
		assert.Equal(t, 3, result.State.ReviewerErrorCount)
	})
}

func TestExecutor_CompletionSignal_Integration(t *testing.T) {
	t.Run("stops after completion threshold reached", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "working...", Cost: 0.10, HasSignal: false},
			{Output: "DONE", Cost: 0.10, HasSignal: true},
			{Output: "DONE", Cost: 0.10, HasSignal: true},
			{Output: "DONE", Cost: 0.10, HasSignal: true},
		}

		config := &loop.Config{
			Prompt:               "work",
			MaxRuns:              100,
			CompletionSignal:     "DONE",
			CompletionThreshold:  3,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonCompletionSignal, result.StopReason)
		assert.GreaterOrEqual(t, result.State.CompletionSignalCount, 3)
	})

	t.Run("completion signal detection in main and reviewer", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "DONE", Cost: 0.10, HasSignal: true}, // main 1
			{Output: "DONE", Cost: 0.02, HasSignal: true}, // review 1
			{Output: "DONE", Cost: 0.10, HasSignal: true}, // main 2 - threshold reached
		}

		config := &loop.Config{
			Prompt:               "work",
			MaxRuns:              100,
			ReviewPrompt:         "check",
			CompletionSignal:     "DONE",
			CompletionThreshold:  3,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonCompletionSignal, result.StopReason)
	})
}

func TestExecutor_DryRun_Integration(t *testing.T) {
	t.Run("dry run does not call Claude", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()

		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              5,
			DryRun:               true,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonMaxRuns, result.StopReason)
		assert.Equal(t, 5, result.State.SuccessfulIterations)
		assert.Equal(t, 0, mockClient.CallCount()) // No Claude calls in dry run
		assert.Equal(t, 0.0, result.State.TotalCost)
	})
}

func TestExecutor_ErrorHandling_Integration(t *testing.T) {
	t.Run("recovers from transient errors", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "work 1", Cost: 0.10},
			{Err: errors.New("transient error")},
			{Output: "work 2", Cost: 0.10},
			{Output: "work 3", Cost: 0.10},
		}

		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              3,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonMaxRuns, result.StopReason)
		assert.Equal(t, 3, result.State.SuccessfulIterations)
	})

	t.Run("stops after consecutive errors", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "work 1", Cost: 0.10},
			{Err: errors.New("error 1")},
			{Err: errors.New("error 2")},
			{Err: errors.New("error 3")},
		}

		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              10,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonConsecutiveErrors, result.StopReason)
		assert.Equal(t, 1, result.State.SuccessfulIterations)
		assert.Equal(t, 3, result.State.ErrorCount)
	})
}

func TestExecutor_ContextCancellation_Integration(t *testing.T) {
	t.Run("respects already cancelled context", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "work 1", Cost: 0.10},
		}

		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              100,
			MaxConsecutiveErrors: 3,
		}

		// Create already cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(ctx)

		require.NoError(t, err)
		assert.Equal(t, loop.StopReasonContextCancelled, result.StopReason)
		assert.Equal(t, 0, mockClient.CallCount()) // Should not have called Claude
	})
}

func TestExecutor_ProgressCallback_Integration(t *testing.T) {
	t.Run("progress callback called after each iteration", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "work 1", Cost: 0.10},
			{Output: "work 2", Cost: 0.10},
			{Output: "work 3", Cost: 0.10},
		}

		var progressCalls []*loop.State
		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              3,
			MaxConsecutiveErrors: 3,
			OnProgress: func(state *loop.State) {
				// Create a copy to track state at each call
				stateCopy := *state
				progressCalls = append(progressCalls, &stateCopy)
			},
		}

		executor := loop.NewExecutor(config, mockClient)
		_, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 3, len(progressCalls))
		assert.Equal(t, 1, progressCalls[0].SuccessfulIterations)
		assert.Equal(t, 2, progressCalls[1].SuccessfulIterations)
		assert.Equal(t, 3, progressCalls[2].SuccessfulIterations)
	})
}

func TestExecutor_StateTracking_Integration(t *testing.T) {
	t.Run("accumulates cost correctly", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "work 1", Cost: 0.10},
			{Output: "work 2", Cost: 0.20},
			{Output: "work 3", Cost: 0.30},
		}

		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              3,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.InDelta(t, 0.60, result.State.TotalCost, 0.001)
	})

	t.Run("tracks total iterations including errors", func(t *testing.T) {
		mockClient := mocks.NewConfigurableClaudeClient()
		mockClient.Responses = []mocks.ClientResponse{
			{Output: "work 1", Cost: 0.10},
			{Err: errors.New("error")},
			{Output: "work 2", Cost: 0.10},
		}

		config := &loop.Config{
			Prompt:               "test",
			MaxRuns:              2,
			MaxConsecutiveErrors: 3,
		}

		executor := loop.NewExecutor(config, mockClient)
		result, err := executor.Run(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 2, result.State.SuccessfulIterations)
		assert.Equal(t, 3, result.State.TotalIterations) // 2 success + 1 error
	})
}
