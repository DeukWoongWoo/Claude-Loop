package reviewer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClaudeClient implements ClaudeClient for testing.
type MockClaudeClient struct {
	ExecuteFunc func(ctx context.Context, prompt string) (*IterationResult, error)
	calls       []string
}

func (m *MockClaudeClient) Execute(ctx context.Context, prompt string) (*IterationResult, error) {
	m.calls = append(m.calls, prompt)
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, prompt)
	}
	return &IterationResult{
		Output:                "review completed",
		Cost:                  0.05,
		Duration:              time.Second,
		CompletionSignalFound: false,
	}, nil
}

func TestNewReviewer(t *testing.T) {
	t.Parallel()

	t.Run("with nil config", func(t *testing.T) {
		t.Parallel()
		client := &MockClaudeClient{}
		reviewer := NewReviewer(nil, client)

		assert.NotNil(t, reviewer)
		assert.NotNil(t, reviewer.config)
		assert.Equal(t, 3, reviewer.config.MaxConsecutiveErrors)
	})

	t.Run("with provided config", func(t *testing.T) {
		t.Parallel()
		client := &MockClaudeClient{}
		config := &Config{
			ReviewPrompt:         "run tests",
			MaxConsecutiveErrors: 5,
		}
		reviewer := NewReviewer(config, client)

		assert.NotNil(t, reviewer)
		assert.Equal(t, config, reviewer.config)
		assert.Equal(t, "run tests", reviewer.config.ReviewPrompt)
	})
}

func TestDefaultReviewer_Run_Success(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return &IterationResult{
				Output:                "All tests passed",
				Cost:                  0.03,
				Duration:              2 * time.Second,
				CompletionSignalFound: false,
			}, nil
		},
	}

	config := &Config{ReviewPrompt: "run npm test"}
	reviewer := NewReviewer(config, client)

	result, err := reviewer.Run(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "All tests passed", result.Output)
	assert.Equal(t, 0.03, result.Cost)
	assert.Equal(t, 2*time.Second, result.Duration)
	assert.False(t, result.CompletionSignalFound)

	// Verify prompt was passed correctly
	require.Len(t, client.calls, 1)
	assert.Contains(t, client.calls[0], "CODE REVIEW CONTEXT")
	assert.Contains(t, client.calls[0], "run npm test")
}

func TestDefaultReviewer_Run_WithCompletionSignal(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return &IterationResult{
				Output:                "Project complete CONTINUOUS_CLAUDE_PROJECT_COMPLETE",
				Cost:                  0.02,
				Duration:              time.Second,
				CompletionSignalFound: true,
			}, nil
		},
	}

	config := &Config{ReviewPrompt: "check completion"}
	reviewer := NewReviewer(config, client)

	result, err := reviewer.Run(context.Background())

	require.NoError(t, err)
	assert.True(t, result.CompletionSignalFound)
}

func TestDefaultReviewer_Run_ClientError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection timeout")
	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return nil, expectedErr
		},
	}

	config := &Config{ReviewPrompt: "run tests"}
	reviewer := NewReviewer(config, client)

	result, err := reviewer.Run(context.Background())

	require.Error(t, err)
	assert.Nil(t, result)

	var reviewerErr *ReviewerError
	require.True(t, errors.As(err, &reviewerErr))
	assert.Equal(t, "execute", reviewerErr.Phase)
	assert.Equal(t, "claude execution failed", reviewerErr.Message)
	assert.True(t, errors.Is(err, expectedErr))
}

func TestDefaultReviewer_Run_NoPromptError(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{}
	config := &Config{ReviewPrompt: ""} // Empty prompt
	reviewer := NewReviewer(config, client)

	result, err := reviewer.Run(context.Background())

	require.Error(t, err)
	assert.Nil(t, result)

	var reviewerErr *ReviewerError
	require.True(t, errors.As(err, &reviewerErr))
	assert.Equal(t, "prompt", reviewerErr.Phase)
	assert.Contains(t, err.Error(), "no review prompt provided")

	// Client should not have been called
	assert.Empty(t, client.calls)
}

func TestDefaultReviewer_Run_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return nil, ctx.Err()
		},
	}

	config := &Config{ReviewPrompt: "run tests"}
	reviewer := NewReviewer(config, client)

	result, err := reviewer.Run(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestDefaultReviewer_Config(t *testing.T) {
	t.Parallel()

	config := &Config{
		ReviewPrompt:         "test prompt",
		MaxConsecutiveErrors: 5,
	}
	client := &MockClaudeClient{}
	reviewer := NewReviewer(config, client)

	assert.Equal(t, config, reviewer.Config())
}

func TestDefaultReviewer_Run_CostAccumulation(t *testing.T) {
	t.Parallel()

	costs := []float64{0.01, 0.02, 0.03}
	callIndex := 0

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			cost := costs[callIndex]
			callIndex++
			return &IterationResult{
				Output:   "ok",
				Cost:     cost,
				Duration: time.Second,
			}, nil
		},
	}

	config := &Config{ReviewPrompt: "run tests"}
	reviewer := NewReviewer(config, client)

	// Run multiple times and verify costs
	var totalCost float64
	for i := 0; i < 3; i++ {
		result, err := reviewer.Run(context.Background())
		require.NoError(t, err)
		totalCost += result.Cost
	}

	assert.InDelta(t, 0.06, totalCost, 0.001)
}
