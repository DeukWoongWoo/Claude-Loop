package council

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClaudeClient is a mock implementation of ClaudeClient for testing.
type mockClaudeClient struct {
	response *IterationResult
	err      error
	calls    []string
}

func (m *mockClaudeClient) Execute(ctx context.Context, prompt string) (*IterationResult, error) {
	m.calls = append(m.calls, prompt)
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func TestNewCouncil(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		cfg := &Config{
			LogFile:      "/custom/path.log",
			LogDecisions: true,
			Principles:   &config.Principles{},
		}
		client := &mockClaudeClient{}

		council := NewCouncil(cfg, client)

		assert.NotNil(t, council)
		assert.Equal(t, cfg, council.config)
		assert.NotNil(t, council.detector)
		assert.NotNil(t, council.promptBuilder)
		assert.NotNil(t, council.logger)
	})

	t.Run("with nil config uses defaults", func(t *testing.T) {
		client := &mockClaudeClient{}

		council := NewCouncil(nil, client)

		assert.NotNil(t, council)
		assert.NotNil(t, council.config)
		assert.Equal(t, ".claude/principles-decisions.log", council.config.LogFile)
	})
}

func TestDefaultCouncil_DetectConflict(t *testing.T) {
	client := &mockClaudeClient{}
	council := NewCouncil(DefaultConfig(), client)

	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "detects conflict",
			output:   "PRINCIPLE_CONFLICT_UNRESOLVED in the response",
			expected: true,
		},
		{
			name:     "no conflict",
			output:   "Everything completed successfully",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := council.DetectConflict(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultCouncil_Resolve(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		client := &mockClaudeClient{
			response: &IterationResult{
				Output:   "**Decision**: Use startup\n**Rationale**: Speed priority",
				Cost:     0.05,
				Duration: 2 * time.Second,
			},
		}
		cfg := &Config{
			Principles: &config.Principles{
				Version: "2.3",
				Preset:  config.PresetStartup,
			},
		}
		council := NewCouncil(cfg, client)

		result, err := council.Resolve(ctx, "Conflict between speed and correctness")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Use startup", result.Resolution)
		assert.Equal(t, "Speed priority", result.Rationale)
		assert.Equal(t, 0.05, result.Cost)
		assert.Len(t, client.calls, 1)
		assert.Contains(t, client.calls[0], "Conflict between speed and correctness")
	})

	t.Run("error when no principles", func(t *testing.T) {
		client := &mockClaudeClient{}
		cfg := &Config{
			Principles: nil,
		}
		council := NewCouncil(cfg, client)

		result, err := council.Resolve(ctx, "Some conflict")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrNoPrinciples, err)
	})

	t.Run("error when client fails", func(t *testing.T) {
		client := &mockClaudeClient{
			err: errors.New("client error"),
		}
		cfg := &Config{
			Principles: &config.Principles{},
		}
		council := NewCouncil(cfg, client)

		result, err := council.Resolve(ctx, "Some conflict")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsCouncilError(err))

		var councilErr *CouncilError
		require.True(t, errors.As(err, &councilErr))
		assert.Equal(t, "resolve", councilErr.Phase)
	})
}

func TestDefaultCouncil_LogDecision(t *testing.T) {
	t.Run("delegates to logger", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &Config{
			LogFile:      tmpDir + "/test.log",
			LogDecisions: true,
		}
		council := NewCouncil(cfg, &mockClaudeClient{})

		decision := &Decision{
			Timestamp: time.Now(),
			Iteration: 1,
			Decision:  "Test decision",
			Rationale: "Test rationale",
		}

		err := council.LogDecision(decision)
		assert.NoError(t, err)
	})
}

func TestDefaultCouncil_Config(t *testing.T) {
	cfg := &Config{
		LogFile: "/test/path",
	}
	council := NewCouncil(cfg, &mockClaudeClient{})

	assert.Equal(t, cfg, council.Config())
}

func TestDefaultCouncil_ExtractDecisionFromOutput(t *testing.T) {
	council := NewCouncil(DefaultConfig(), &mockClaudeClient{})

	output := `Analysis complete.
**Decision**: Apply enterprise settings
**Rationale**: Security requirements mandate stricter controls`

	decision, rationale := council.ExtractDecisionFromOutput(output)

	assert.Equal(t, "Apply enterprise settings", decision)
	assert.Equal(t, "Security requirements mandate stricter controls", rationale)
}
