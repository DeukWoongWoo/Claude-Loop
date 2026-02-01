package planner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/loop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLoopClient implements loop.ClaudeClient for testing.
type mockLoopClient struct {
	result *loop.IterationResult
	err    error
}

func (m *mockLoopClient) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestNewClaudeClientAdapter(t *testing.T) {
	tests := []struct {
		name   string
		client loop.ClaudeClient
		isNil  bool
	}{
		{
			name:   "with valid client",
			client: &mockLoopClient{},
			isNil:  false,
		},
		{
			name:   "with nil client",
			client: nil,
			isNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewClaudeClientAdapter(tt.client)
			if tt.isNil {
				assert.Nil(t, adapter)
			} else {
				assert.NotNil(t, adapter)
			}
		})
	}
}

func TestClaudeClientAdapter_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		mockResult *loop.IterationResult
		mockErr    error
		prompt     string
		wantErr    bool
		errMsg     string
	}{
		{
			name: "successful execution",
			mockResult: &loop.IterationResult{
				Output:                "test output",
				Cost:                  0.05,
				Duration:              5 * time.Second,
				CompletionSignalFound: false,
			},
			prompt:  "test prompt",
			wantErr: false,
		},
		{
			name: "with completion signal",
			mockResult: &loop.IterationResult{
				Output:                "task complete CONTINUOUS_CLAUDE_PROJECT_COMPLETE",
				Cost:                  0.10,
				Duration:              10 * time.Second,
				CompletionSignalFound: true,
			},
			prompt:  "complete task",
			wantErr: false,
		},
		{
			name:    "client error",
			mockErr: errors.New("execution failed"),
			prompt:  "test prompt",
			wantErr: true,
			errMsg:  "execution failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLoopClient{
				result: tt.mockResult,
				err:    tt.mockErr,
			}
			adapter := NewClaudeClientAdapter(mock)
			require.NotNil(t, adapter)

			result, err := adapter.Execute(ctx, tt.prompt)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.mockResult.Output, result.Output)
			assert.Equal(t, tt.mockResult.Cost, result.Cost)
			assert.Equal(t, tt.mockResult.Duration, result.Duration)
			assert.Equal(t, tt.mockResult.CompletionSignalFound, result.CompletionSignalFound)
		})
	}
}

func TestClaudeClientAdapter_Execute_NilClient(t *testing.T) {
	// Create adapter with valid client, then test internal nil check
	adapter := &ClaudeClientAdapter{client: nil}

	_, err := adapter.Execute(context.Background(), "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude client is nil")
}

func TestClaudeClientAdapter_InterfaceCompliance(t *testing.T) {
	// Verify that adapter implements ClaudeClient interface
	var _ ClaudeClient = (*ClaudeClientAdapter)(nil)
}
