package prd

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
		Output: `### Goals
- Test goal

### Requirements
- Test requirement

### Success Criteria
- Test criterion`,
		Cost:     0.05,
		Duration: time.Second,
	}, nil
}

func TestNewGenerator(t *testing.T) {
	t.Parallel()

	t.Run("with nil config", func(t *testing.T) {
		t.Parallel()
		client := &MockClaudeClient{}
		gen := NewGenerator(nil, client)

		assert.NotNil(t, gen)
		assert.NotNil(t, gen.config)
		assert.Equal(t, 3, gen.config.MaxRetries)
		assert.True(t, gen.config.ValidateOutput)
	})

	t.Run("with custom config", func(t *testing.T) {
		t.Parallel()
		config := &Config{MaxRetries: 5, ValidateOutput: false}
		gen := NewGenerator(config, &MockClaudeClient{})

		assert.Equal(t, 5, gen.config.MaxRetries)
		assert.False(t, gen.config.ValidateOutput)
	})

	t.Run("with nil client", func(t *testing.T) {
		t.Parallel()
		gen := NewGenerator(nil, nil)
		assert.Nil(t, gen)
	})
}

func TestDefaultGenerator_Generate(t *testing.T) {
	t.Parallel()

	t.Run("successful generation", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return &IterationResult{
					Output: `### Goals
- Build authentication

### Requirements
- Support OAuth

### Success Criteria
- All tests pass`,
					Cost:     0.03,
					Duration: 2 * time.Second,
				}, nil
			},
		}

		gen := NewGenerator(nil, client)
		prd, err := gen.Generate(context.Background(), "Build a login system")

		require.NoError(t, err)
		require.NotNil(t, prd)
		assert.Equal(t, []string{"Build authentication"}, prd.Goals)
		assert.Equal(t, []string{"Support OAuth"}, prd.Requirements)
		assert.Equal(t, []string{"All tests pass"}, prd.SuccessCriteria)
		assert.Equal(t, 0.03, prd.Cost)
		assert.NotZero(t, prd.CreatedAt)
		assert.NotZero(t, prd.Duration)
	})

	t.Run("empty prompt error", func(t *testing.T) {
		t.Parallel()

		gen := NewGenerator(nil, &MockClaudeClient{})
		prd, err := gen.Generate(context.Background(), "")

		require.Error(t, err)
		assert.Nil(t, prd)
		assert.Equal(t, ErrEmptyPrompt, err)
	})

	t.Run("client execution error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("network timeout")
		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return nil, expectedErr
			},
		}

		gen := NewGenerator(nil, client)
		prd, err := gen.Generate(context.Background(), "test prompt")

		require.Error(t, err)
		assert.Nil(t, prd)
		assert.True(t, IsPRDError(err))

		var prdErr *PRDError
		require.ErrorAs(t, err, &prdErr)
		assert.Equal(t, "generate", prdErr.Phase)
		assert.Equal(t, expectedErr, prdErr.Err)
	})

	t.Run("parse error - no goals", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return &IterationResult{
					Output: `### Requirements
- Some requirement

### Success Criteria
- Some criterion`,
					Cost: 0.01,
				}, nil
			},
		}

		gen := NewGenerator(nil, client)
		prd, err := gen.Generate(context.Background(), "test")

		require.Error(t, err)
		assert.Nil(t, prd)
		assert.Equal(t, ErrParseNoGoals, err)
	})

	t.Run("parse error - no requirements", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return &IterationResult{
					Output: `### Goals
- Some goal

### Success Criteria
- Some criterion`,
					Cost: 0.01,
				}, nil
			},
		}

		gen := NewGenerator(nil, client)
		prd, err := gen.Generate(context.Background(), "test")

		require.Error(t, err)
		assert.Nil(t, prd)
		assert.Equal(t, ErrParseNoRequirements, err)
	})

	t.Run("validation error when enabled", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return &IterationResult{
					Output: `### Goals
- Goal

### Requirements
- Req

### Success Criteria
`,
					Cost: 0.01,
				}, nil
			},
		}

		// Validation enabled by default
		gen := NewGenerator(nil, client)
		prd, err := gen.Generate(context.Background(), "test")

		require.Error(t, err)
		assert.Nil(t, prd)
		assert.True(t, IsPRDError(err))

		var prdErr *PRDError
		require.ErrorAs(t, err, &prdErr)
		assert.Equal(t, "validate", prdErr.Phase)
	})

	t.Run("validation disabled skips validation", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return &IterationResult{
					Output: `### Goals
- Goal

### Requirements
- Req

### Success Criteria
- Criterion`,
					Cost: 0.01,
				}, nil
			},
		}

		config := &Config{ValidateOutput: false}
		gen := NewGenerator(config, client)
		prd, err := gen.Generate(context.Background(), "test")

		require.NoError(t, err)
		require.NotNil(t, prd)
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return nil, ctx.Err()
			},
		}

		gen := NewGenerator(nil, client)
		_, err := gen.Generate(ctx, "test")

		require.Error(t, err)
	})

	t.Run("prompt contains user input", func(t *testing.T) {
		t.Parallel()

		var capturedPrompt string
		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				capturedPrompt = prompt
				return &IterationResult{
					Output: `### Goals
- Goal

### Requirements
- Req

### Success Criteria
- Criterion`,
				}, nil
			},
		}

		gen := NewGenerator(nil, client)
		_, err := gen.Generate(context.Background(), "My custom user prompt")

		require.NoError(t, err)
		assert.Contains(t, capturedPrompt, "My custom user prompt")
		assert.Contains(t, capturedPrompt, "PRD")
	})
}

func TestDefaultGenerator_Config(t *testing.T) {
	t.Parallel()

	config := &Config{MaxRetries: 7}
	gen := NewGenerator(config, &MockClaudeClient{})

	assert.Equal(t, config, gen.Config())
}

func TestDefaultGenerator_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Verify that DefaultGenerator implements Generator interface
	var _ Generator = (*DefaultGenerator)(nil)
}

func TestMockClaudeClient_DefaultBehavior(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{}
	result, err := client.Execute(context.Background(), "test prompt")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0.05, result.Cost)
	assert.Contains(t, result.Output, "### Goals")
	assert.Len(t, client.calls, 1)
	assert.Equal(t, "test prompt", client.calls[0])
}

func TestMockClaudeClient_RecordsCalls(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{}

	_, _ = client.Execute(context.Background(), "first")
	_, _ = client.Execute(context.Background(), "second")
	_, _ = client.Execute(context.Background(), "third")

	assert.Len(t, client.calls, 3)
	assert.Equal(t, []string{"first", "second", "third"}, client.calls)
}
