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

// mockDecomposer implements Decomposer for testing.
type mockDecomposer struct {
	taskGraph *TaskGraph
	err       error
}

func (m *mockDecomposer) Decompose(ctx context.Context, arch *planner.Architecture) (*TaskGraph, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.taskGraph, nil
}

func TestNewPhase(t *testing.T) {
	tests := []struct {
		name       string
		decomposer Decomposer
		isNil      bool
	}{
		{
			name:       "with valid decomposer",
			decomposer: &mockDecomposer{},
			isNil:      false,
		},
		{
			name:       "with nil decomposer",
			decomposer: nil,
			isNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase := NewPhase(tt.decomposer)
			if tt.isNil {
				assert.Nil(t, phase)
			} else {
				assert.NotNil(t, phase)
			}
		})
	}
}

func TestPhase_Name(t *testing.T) {
	phase := NewPhase(&mockDecomposer{})
	require.NotNil(t, phase)
	assert.Equal(t, "tasks", phase.Name())
}

func TestPhase_ShouldRun(t *testing.T) {
	tests := []struct {
		name     string
		plan     *planner.Plan
		expected bool
	}{
		{
			name:     "nil plan",
			plan:     nil,
			expected: false,
		},
		{
			name: "nil Architecture",
			plan: &planner.Plan{
				Architecture:    nil,
				CompletedPhases: []string{},
			},
			expected: false,
		},
		{
			name: "phase not completed with Architecture",
			plan: &planner.Plan{
				Architecture: &planner.Architecture{
					Components: []planner.Component{{Name: "test"}},
				},
				CompletedPhases: []string{"prd", "architecture"},
			},
			expected: true,
		},
		{
			name: "phase already completed",
			plan: &planner.Plan{
				Architecture: &planner.Architecture{
					Components: []planner.Component{{Name: "test"}},
				},
				CompletedPhases: []string{"prd", "architecture", "tasks"},
			},
			expected: false,
		},
	}

	phase := NewPhase(&mockDecomposer{})
	require.NotNil(t, phase)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := phase.ShouldRun(nil, tt.plan)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPhase_Run(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		plan      *planner.Plan
		taskGraph *TaskGraph
		decErr    error
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "nil plan",
			plan:    nil,
			wantErr: true,
			errMsg:  "plan is nil",
		},
		{
			name: "nil Architecture",
			plan: &planner.Plan{
				Architecture:    nil,
				CompletedPhases: []string{},
			},
			wantErr: true,
			errMsg:  "Architecture is nil",
		},
		{
			name: "successful decomposition",
			plan: &planner.Plan{
				Architecture: &planner.Architecture{
					Components: []planner.Component{
						{Name: "PaymentService", Description: "Handles payments"},
					},
				},
				CompletedPhases: []string{"prd", "architecture"},
			},
			taskGraph: &TaskGraph{
				TaskGraph: planner.TaskGraph{
					Tasks: []planner.Task{
						{ID: "T001", Title: "Create payment service", Status: "pending"},
						{ID: "T002", Title: "Add Stripe integration", Status: "pending", Dependencies: []string{"T001"}},
					},
					ExecutionOrder: []string{"T001", "T002"},
				},
				Cost:     0.10,
				Duration: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "decomposer error",
			plan: &planner.Plan{
				Architecture: &planner.Architecture{
					Components: []planner.Component{{Name: "test"}},
				},
				CompletedPhases: []string{"prd", "architecture"},
			},
			decErr:  errors.New("decomposition failed"),
			wantErr: true,
			errMsg:  "Task decomposition failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decomposer := &mockDecomposer{
				taskGraph: tt.taskGraph,
				err:       tt.decErr,
			}
			phase := NewPhase(decomposer)
			require.NotNil(t, phase)

			initialCost := float64(0)
			if tt.plan != nil {
				initialCost = tt.plan.TotalCost
			}

			err := phase.Run(ctx, tt.plan)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tt.plan.TaskGraph)
			assert.Equal(t, tt.taskGraph.Tasks, tt.plan.TaskGraph.Tasks)
			assert.Equal(t, tt.taskGraph.ExecutionOrder, tt.plan.TaskGraph.ExecutionOrder)
			assert.Equal(t, initialCost+tt.taskGraph.Cost, tt.plan.TotalCost)
		})
	}
}

func TestPhase_InterfaceCompliance(t *testing.T) {
	// Verify that Phase implements planner.PlanningPhase interface
	var _ planner.PlanningPhase = (*Phase)(nil)
}
