package decomposer

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
	return &IterationResult{}, nil
}

func TestNewDecomposer_NilClient(t *testing.T) {
	t.Parallel()

	decomposer := NewDecomposer(nil, nil)
	assert.Nil(t, decomposer)
}

func TestNewDecomposer_DefaultConfig(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{}
	decomposer := NewDecomposer(nil, client)

	require.NotNil(t, decomposer)
	assert.NotNil(t, decomposer.Config())
	assert.Equal(t, ".claude/tasks", decomposer.Config().TaskDir)
	assert.Equal(t, 3, decomposer.Config().MaxRetries)
	assert.True(t, decomposer.Config().ValidateOutput)
}

func TestNewDecomposer_CustomConfig(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{}
	config := &Config{
		TaskDir:        ".custom/tasks",
		MaxRetries:     5,
		ValidateOutput: false,
	}
	decomposer := NewDecomposer(config, client)

	require.NotNil(t, decomposer)
	assert.Equal(t, ".custom/tasks", decomposer.Config().TaskDir)
	assert.Equal(t, 5, decomposer.Config().MaxRetries)
	assert.False(t, decomposer.Config().ValidateOutput)
}

func TestDefaultDecomposer_Decompose_Success(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return &IterationResult{
				Output: `### Task T001: First task
- **Description**: Do the first thing
- **Dependencies**: none
- **Files**: file1.go
- **Complexity**: small

### Task T002: Second task
- **Description**: Do the second thing
- **Dependencies**: [T001]
- **Files**: file2.go
- **Complexity**: medium
`,
				Cost:     0.05,
				Duration: 100 * time.Millisecond,
			}, nil
		},
	}

	decomposer := NewDecomposer(nil, client)
	arch := &planner.Architecture{
		Components: []planner.Component{
			{Name: "Test", Description: "Test component", Files: []string{"test.go"}},
		},
		Dependencies:  []string{},
		FileStructure: []string{"test.go"},
		TechDecisions: []string{"Use Go"},
	}

	taskGraph, err := decomposer.Decompose(context.Background(), arch)

	require.NoError(t, err)
	require.NotNil(t, taskGraph)
	assert.Len(t, taskGraph.Tasks, 2)
	assert.Equal(t, []string{"T001", "T002"}, taskGraph.ExecutionOrder)
	assert.Equal(t, 0.05, taskGraph.Cost)
	assert.NotZero(t, taskGraph.Duration)
}

func TestDefaultDecomposer_Decompose_NilArchitecture(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{}
	decomposer := NewDecomposer(nil, client)

	_, err := decomposer.Decompose(context.Background(), nil)

	require.Error(t, err)
	assert.Equal(t, ErrNilArchitecture, err)
}

func TestDefaultDecomposer_Decompose_ClientError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("client error")
	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return nil, expectedErr
		},
	}

	decomposer := NewDecomposer(nil, client)
	arch := &planner.Architecture{
		Components:    []planner.Component{{Name: "Test", Description: "Test"}},
		FileStructure: []string{"test.go"},
		TechDecisions: []string{"Go"},
	}

	_, err := decomposer.Decompose(context.Background(), arch)

	require.Error(t, err)
	var de *DecomposerError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, "generate", de.Phase)
	assert.ErrorIs(t, err, expectedErr)
}

func TestDefaultDecomposer_Decompose_ParseError(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return &IterationResult{
				Output: "No valid tasks here",
				Cost:   0.01,
			}, nil
		},
	}

	decomposer := NewDecomposer(nil, client)
	arch := &planner.Architecture{
		Components:    []planner.Component{{Name: "Test", Description: "Test"}},
		FileStructure: []string{"test.go"},
		TechDecisions: []string{"Go"},
	}

	_, err := decomposer.Decompose(context.Background(), arch)

	require.Error(t, err)
	assert.Equal(t, ErrParseNoTasks, err)
}

func TestDefaultDecomposer_Decompose_ValidationError(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			// Return invalid task ID
			return &IterationResult{
				Output: `### Task T001: First task
- **Description**: Do something
- **Dependencies**: T999
- **Files**: file.go
`,
				Cost: 0.01,
			}, nil
		},
	}

	decomposer := NewDecomposer(nil, client)
	arch := &planner.Architecture{
		Components:    []planner.Component{{Name: "Test", Description: "Test"}},
		FileStructure: []string{"test.go"},
		TechDecisions: []string{"Go"},
	}

	_, err := decomposer.Decompose(context.Background(), arch)

	require.Error(t, err)
	var de *DecomposerError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, "validate", de.Phase)
}

func TestDefaultDecomposer_Decompose_ValidationDisabled(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			// Return task with invalid dependency - would fail validation
			return &IterationResult{
				Output: `### Task T001: First task
- **Description**: Do something
- **Dependencies**: T999
- **Files**: file.go
`,
				Cost: 0.01,
			}, nil
		},
	}

	config := &Config{
		ValidateOutput: false,
	}
	decomposer := NewDecomposer(config, client)
	arch := &planner.Architecture{
		Components:    []planner.Component{{Name: "Test", Description: "Test"}},
		FileStructure: []string{"test.go"},
		TechDecisions: []string{"Go"},
	}

	// Should succeed because validation is disabled
	taskGraph, err := decomposer.Decompose(context.Background(), arch)

	require.NoError(t, err)
	require.NotNil(t, taskGraph)
}

func TestDefaultDecomposer_Decompose_CycleError(t *testing.T) {
	t.Parallel()

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return &IterationResult{
				Output: `### Task T001: First task
- **Description**: Do first thing
- **Dependencies**: [T002]
- **Files**: file1.go

### Task T002: Second task
- **Description**: Do second thing
- **Dependencies**: [T001]
- **Files**: file2.go
`,
				Cost: 0.01,
			}, nil
		},
	}

	decomposer := NewDecomposer(nil, client)
	arch := &planner.Architecture{
		Components:    []planner.Component{{Name: "Test", Description: "Test"}},
		FileStructure: []string{"test.go"},
		TechDecisions: []string{"Go"},
	}

	_, err := decomposer.Decompose(context.Background(), arch)

	require.Error(t, err)
	var ge *GraphError
	require.ErrorAs(t, err, &ge)
	assert.Equal(t, "cycle", ge.Type)
}

func TestDefaultDecomposer_Decompose_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return nil, ctx.Err()
		},
	}

	decomposer := NewDecomposer(nil, client)
	arch := &planner.Architecture{
		Components:    []planner.Component{{Name: "Test", Description: "Test"}},
		FileStructure: []string{"test.go"},
		TechDecisions: []string{"Go"},
	}

	_, err := decomposer.Decompose(ctx, arch)

	require.Error(t, err)
}

func TestDefaultDecomposer_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ Decomposer = (*DefaultDecomposer)(nil)
}
