package architecture

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
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
	// Default valid response
	return &IterationResult{
		Output: `### Components
- **Parser**: Parses architecture output
  - Files: parser.go, parser_test.go

### Dependencies
- github.com/stretchr/testify

### File Structure
- internal/architecture/parser.go
- internal/architecture/types.go

### Technical Decisions
- Use regex for parsing`,
		Cost:     0.05,
		Duration: time.Second,
	}, nil
}

func TestNewGenerator(t *testing.T) {
	t.Parallel()

	t.Run("nil client returns nil", func(t *testing.T) {
		t.Parallel()
		gen := NewGenerator(nil, nil)
		assert.Nil(t, gen)
	})

	t.Run("nil config uses default", func(t *testing.T) {
		t.Parallel()
		client := &MockClaudeClient{}
		gen := NewGenerator(nil, client)
		require.NotNil(t, gen)
		assert.Equal(t, 3, gen.config.MaxRetries)
		assert.True(t, gen.config.ValidateOutput)
	})

	t.Run("custom config is used", func(t *testing.T) {
		t.Parallel()
		client := &MockClaudeClient{}
		config := &Config{MaxRetries: 5, ValidateOutput: false}
		gen := NewGenerator(config, client)
		require.NotNil(t, gen)
		assert.Equal(t, 5, gen.config.MaxRetries)
		assert.False(t, gen.config.ValidateOutput)
	})
}

func TestDefaultGenerator_Generate(t *testing.T) {
	t.Parallel()

	t.Run("successful generation", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{}
		gen := NewGenerator(nil, client)
		require.NotNil(t, gen)

		prd := &planner.PRD{
			Goals:           []string{"Build authentication system"},
			Requirements:    []string{"Support JWT tokens"},
			SuccessCriteria: []string{"All tests pass"},
		}

		before := time.Now()
		arch, err := gen.Generate(context.Background(), prd)
		after := time.Now()

		require.NoError(t, err)
		require.NotNil(t, arch)

		// Verify components parsed
		assert.Len(t, arch.Components, 1)
		assert.Equal(t, "Parser", arch.Components[0].Name)

		// Verify metadata
		assert.True(t, arch.CreatedAt.After(before) || arch.CreatedAt.Equal(before))
		assert.True(t, arch.CreatedAt.Before(after) || arch.CreatedAt.Equal(after))
		assert.Equal(t, 0.05, arch.Cost)
		assert.Greater(t, arch.Duration, time.Duration(0))

		// Verify client was called
		assert.Len(t, client.calls, 1)
	})

	t.Run("nil PRD returns error", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{}
		gen := NewGenerator(nil, client)

		arch, err := gen.Generate(context.Background(), nil)

		require.Error(t, err)
		assert.Nil(t, arch)
		assert.Equal(t, ErrNilPRD, err)
	})

	t.Run("client error wrapped", func(t *testing.T) {
		t.Parallel()

		clientErr := errors.New("network error")
		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return nil, clientErr
			},
		}
		gen := NewGenerator(nil, client)

		prd := &planner.PRD{Goals: []string{"Test"}}
		arch, err := gen.Generate(context.Background(), prd)

		require.Error(t, err)
		assert.Nil(t, arch)

		var ae *ArchitectureError
		require.ErrorAs(t, err, &ae)
		assert.Equal(t, "generate", ae.Phase)
		assert.True(t, errors.Is(err, clientErr))
	})

	t.Run("parse error returned", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return &IterationResult{
					Output: "No components here",
					Cost:   0.01,
				}, nil
			},
		}
		gen := NewGenerator(nil, client)

		prd := &planner.PRD{Goals: []string{"Test"}}
		arch, err := gen.Generate(context.Background(), prd)

		require.Error(t, err)
		assert.Nil(t, arch)
		assert.Equal(t, ErrParseNoComponents, err)
	})

	t.Run("validation error when enabled", func(t *testing.T) {
		t.Parallel()

		// Output with components but missing file structure and tech decisions
		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return &IterationResult{
					Output: `### Components
- **Test**: Test component
  - Files: test.go
`,
					Cost: 0.01,
				}, nil
			},
		}
		config := &Config{ValidateOutput: true}
		gen := NewGenerator(config, client)

		prd := &planner.PRD{Goals: []string{"Test"}}
		arch, err := gen.Generate(context.Background(), prd)

		require.Error(t, err)
		assert.Nil(t, arch)

		var ae *ArchitectureError
		require.ErrorAs(t, err, &ae)
		assert.Equal(t, "validate", ae.Phase)
	})

	t.Run("validation skipped when disabled", func(t *testing.T) {
		t.Parallel()

		// Output with components but missing file structure and tech decisions
		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return &IterationResult{
					Output: `### Components
- **Test**: Test component
  - Files: test.go
`,
					Cost: 0.01,
				}, nil
			},
		}
		config := &Config{ValidateOutput: false}
		gen := NewGenerator(config, client)

		prd := &planner.PRD{Goals: []string{"Test"}}
		arch, err := gen.Generate(context.Background(), prd)

		require.NoError(t, err)
		require.NotNil(t, arch)
		assert.Len(t, arch.Components, 1)
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		client := &MockClaudeClient{
			ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
				return nil, ctx.Err()
			},
		}
		gen := NewGenerator(nil, client)

		prd := &planner.PRD{Goals: []string{"Test"}}
		arch, err := gen.Generate(ctx, prd)

		require.Error(t, err)
		assert.Nil(t, arch)
	})

	t.Run("prompt contains PRD info", func(t *testing.T) {
		t.Parallel()

		client := &MockClaudeClient{}
		gen := NewGenerator(nil, client)

		prd := &planner.PRD{
			Goals:        []string{"Build user authentication"},
			Requirements: []string{"Support OAuth 2.0"},
		}
		_, err := gen.Generate(context.Background(), prd)
		require.NoError(t, err)

		require.Len(t, client.calls, 1)
		prompt := client.calls[0]
		assert.Contains(t, prompt, "Build user authentication")
		assert.Contains(t, prompt, "Support OAuth 2.0")
	})
}

func TestDefaultGenerator_Config(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{}
	config := &Config{MaxRetries: 10, ValidateOutput: false}
	gen := NewGenerator(config, client)

	assert.Equal(t, config, gen.Config())
}

func TestDefaultGenerator_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Compile-time check that DefaultGenerator implements Generator
	var _ Generator = (*DefaultGenerator)(nil)

	// Runtime check
	client := &MockClaudeClient{}
	gen := NewGenerator(nil, client)
	var iface Generator = gen
	assert.NotNil(t, iface)
}
