package loop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClaudeClient is a mock implementation for testing.
type MockClaudeClient struct {
	Results    []*IterationResult // Results to return in sequence
	Errors     []error            // Errors to return (nil for success)
	CallCount  int                // Number of times Execute was called
	LastPrompt string             // Last prompt received
}

func (m *MockClaudeClient) Execute(ctx context.Context, prompt string) (*IterationResult, error) {
	m.LastPrompt = prompt
	idx := m.CallCount
	m.CallCount++

	if idx < len(m.Errors) && m.Errors[idx] != nil {
		return nil, m.Errors[idx]
	}

	if idx < len(m.Results) && m.Results[idx] != nil {
		return m.Results[idx], nil
	}

	// Default result
	return &IterationResult{
		Output:   "default output",
		Cost:     0.01,
		Duration: 100 * time.Millisecond,
	}, nil
}

func NewMockClient() *MockClaudeClient {
	return &MockClaudeClient{}
}

// SlowMockClaudeClient is a mock that adds delay to simulate real execution time.
type SlowMockClaudeClient struct {
	Delay     time.Duration
	CallCount int
}

func (m *SlowMockClaudeClient) Execute(ctx context.Context, prompt string) (*IterationResult, error) {
	m.CallCount++

	select {
	case <-time.After(m.Delay):
		return &IterationResult{
			Output:   "slow output",
			Cost:     0.01,
			Duration: m.Delay,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestIterationHandler_Execute_Success(t *testing.T) {
	config := &Config{
		Prompt:           "test prompt",
		CompletionSignal: "DONE",
	}

	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Output: "completed", Cost: 0.05, Duration: 100 * time.Millisecond},
		},
	}

	handler := NewIterationHandler(config, mock)
	state := NewState()

	result, err := handler.Execute(context.Background(), state)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "completed", result.Output)
	assert.Equal(t, 0.05, result.Cost)
	assert.Equal(t, 1, state.SuccessfulIterations)
	assert.Equal(t, 1, state.TotalIterations)
	assert.Equal(t, 0.05, state.TotalCost)
	assert.Equal(t, 0, state.ErrorCount)
	assert.Equal(t, "test prompt", mock.LastPrompt)
}

func TestIterationHandler_Execute_WithCompletionSignal(t *testing.T) {
	config := &Config{
		Prompt:           "test prompt",
		CompletionSignal: "DONE",
	}

	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Output: "Task is DONE", Cost: 0.05},
		},
	}

	handler := NewIterationHandler(config, mock)
	state := NewState()

	result, err := handler.Execute(context.Background(), state)

	require.NoError(t, err)
	assert.True(t, result.CompletionSignalFound)
	assert.Equal(t, 1, state.CompletionSignalCount)
}

func TestIterationHandler_Execute_Error(t *testing.T) {
	config := &Config{Prompt: "test prompt"}

	mock := &MockClaudeClient{
		Errors: []error{errors.New("claude failed")},
	}

	handler := NewIterationHandler(config, mock)
	state := NewState()

	result, err := handler.Execute(context.Background(), state)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, IsIterationError(err))

	var iterErr *IterationError
	require.ErrorAs(t, err, &iterErr)
	assert.Equal(t, 1, iterErr.Iteration)
	assert.Contains(t, iterErr.Message, "claude execution failed")

	// State should reflect attempted iteration
	assert.Equal(t, 1, state.TotalIterations)
	assert.Equal(t, 0, state.SuccessfulIterations) // Not incremented on error
}

func TestIterationHandler_Execute_DryRun(t *testing.T) {
	config := &Config{
		Prompt: "test prompt",
		DryRun: true,
	}

	mock := NewMockClient()
	handler := NewIterationHandler(config, mock)
	state := NewState()

	result, err := handler.Execute(context.Background(), state)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Output, "[dry-run]")
	assert.Equal(t, 0.0, result.Cost)
	assert.Equal(t, 0, mock.CallCount) // Client should not be called
	assert.Equal(t, 1, state.SuccessfulIterations)
}

func TestIterationHandler_Execute_UpdatesState(t *testing.T) {
	config := &Config{Prompt: "test"}

	mock := &MockClaudeClient{
		Results: []*IterationResult{
			{Output: "output1", Cost: 0.01},
			{Output: "output2", Cost: 0.02},
			{Output: "output3", Cost: 0.03},
		},
	}

	handler := NewIterationHandler(config, mock)
	state := NewState()
	state.ErrorCount = 5 // Should be reset on success

	// First iteration
	_, err := handler.Execute(context.Background(), state)
	require.NoError(t, err)
	assert.Equal(t, 0.01, state.TotalCost)
	assert.Equal(t, 0, state.ErrorCount)

	// Second iteration
	_, err = handler.Execute(context.Background(), state)
	require.NoError(t, err)
	assert.Equal(t, 0.03, state.TotalCost)
	assert.Equal(t, 2, state.SuccessfulIterations)

	// Third iteration
	_, err = handler.Execute(context.Background(), state)
	require.NoError(t, err)
	assert.Equal(t, 0.06, state.TotalCost)
	assert.Equal(t, 3, state.SuccessfulIterations)
}

func TestIterationHandler_HandleError(t *testing.T) {
	tests := []struct {
		name           string
		maxErrors      int
		initialErrors  int
		wantContinue   bool
		wantErrorCount int
	}{
		{
			name:           "first error - continue",
			maxErrors:      3,
			initialErrors:  0,
			wantContinue:   true,
			wantErrorCount: 1,
		},
		{
			name:           "second error - continue",
			maxErrors:      3,
			initialErrors:  1,
			wantContinue:   true,
			wantErrorCount: 2,
		},
		{
			name:           "third error - stop",
			maxErrors:      3,
			initialErrors:  2,
			wantContinue:   false,
			wantErrorCount: 3,
		},
		{
			name:           "single error allowed",
			maxErrors:      1,
			initialErrors:  0,
			wantContinue:   false,
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{MaxConsecutiveErrors: tt.maxErrors}
			handler := NewIterationHandler(config, NewMockClient())
			state := &State{ErrorCount: tt.initialErrors}

			shouldContinue := handler.HandleError(state, errors.New("test error"))

			assert.Equal(t, tt.wantContinue, shouldContinue)
			assert.Equal(t, tt.wantErrorCount, state.ErrorCount)
		})
	}
}

func TestIterationHandler_ErrorCountReset(t *testing.T) {
	config := &Config{
		Prompt:               "test",
		MaxConsecutiveErrors: 3,
	}

	mock := &MockClaudeClient{
		Errors: []error{
			errors.New("error1"),
			errors.New("error2"),
			nil, // success
			errors.New("error3"),
		},
		Results: []*IterationResult{
			nil, nil,
			{Output: "success", Cost: 0.01},
			nil,
		},
	}

	handler := NewIterationHandler(config, mock)
	state := NewState()

	// First error
	_, err := handler.Execute(context.Background(), state)
	assert.Error(t, err)
	handler.HandleError(state, err)
	assert.Equal(t, 1, state.ErrorCount)

	// Second error
	_, err = handler.Execute(context.Background(), state)
	assert.Error(t, err)
	handler.HandleError(state, err)
	assert.Equal(t, 2, state.ErrorCount)

	// Success - should reset error count
	_, err = handler.Execute(context.Background(), state)
	assert.NoError(t, err)
	assert.Equal(t, 0, state.ErrorCount)

	// New error after success
	_, err = handler.Execute(context.Background(), state)
	assert.Error(t, err)
	handler.HandleError(state, err)
	assert.Equal(t, 1, state.ErrorCount)
}

func TestNewIterationHandler(t *testing.T) {
	config := &Config{Prompt: "test"}
	client := NewMockClient()

	handler := NewIterationHandler(config, client)

	assert.NotNil(t, handler)
	assert.Equal(t, config, handler.config)
	assert.Equal(t, client, handler.client)
	assert.NotNil(t, handler.completionDetector)
}
