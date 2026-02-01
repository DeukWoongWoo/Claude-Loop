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

// mockGenerator implements Generator for testing.
type mockGenerator struct {
	arch *Architecture
	err  error
}

func (m *mockGenerator) Generate(ctx context.Context, prd *planner.PRD) (*Architecture, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.arch, nil
}

func TestNewPhase(t *testing.T) {
	tests := []struct {
		name      string
		generator Generator
		isNil     bool
	}{
		{
			name:      "with valid generator",
			generator: &mockGenerator{},
			isNil:     false,
		},
		{
			name:      "with nil generator",
			generator: nil,
			isNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase := NewPhase(tt.generator)
			if tt.isNil {
				assert.Nil(t, phase)
			} else {
				assert.NotNil(t, phase)
			}
		})
	}
}

func TestPhase_Name(t *testing.T) {
	phase := NewPhase(&mockGenerator{})
	require.NotNil(t, phase)
	assert.Equal(t, "architecture", phase.Name())
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
			name: "nil PRD",
			plan: &planner.Plan{
				PRD:             nil,
				CompletedPhases: []string{},
			},
			expected: false,
		},
		{
			name: "phase not completed with PRD",
			plan: &planner.Plan{
				PRD:             &planner.PRD{Goals: []string{"test"}},
				CompletedPhases: []string{"prd"},
			},
			expected: true,
		},
		{
			name: "phase already completed",
			plan: &planner.Plan{
				PRD:             &planner.PRD{Goals: []string{"test"}},
				CompletedPhases: []string{"prd", "architecture"},
			},
			expected: false,
		},
	}

	phase := NewPhase(&mockGenerator{})
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
		name    string
		plan    *planner.Plan
		arch    *Architecture
		genErr  error
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil plan",
			plan:    nil,
			wantErr: true,
			errMsg:  "plan is nil",
		},
		{
			name: "nil PRD",
			plan: &planner.Plan{
				PRD:             nil,
				CompletedPhases: []string{},
			},
			wantErr: true,
			errMsg:  "PRD is nil",
		},
		{
			name: "successful generation",
			plan: &planner.Plan{
				PRD: &planner.PRD{
					Goals:        []string{"Implement payment"},
					Requirements: []string{"Stripe integration"},
				},
				CompletedPhases: []string{"prd"},
			},
			arch: &Architecture{
				Architecture: planner.Architecture{
					Components: []planner.Component{
						{Name: "PaymentService", Description: "Handles payments"},
					},
					Dependencies:  []string{"stripe-go"},
					FileStructure: []string{"internal/payment/service.go"},
					TechDecisions: []string{"Use Stripe SDK"},
				},
				Cost:     0.08,
				Duration: 8 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "generator error",
			plan: &planner.Plan{
				PRD:             &planner.PRD{Goals: []string{"test"}},
				CompletedPhases: []string{"prd"},
			},
			genErr:  errors.New("generation failed"),
			wantErr: true,
			errMsg:  "Architecture generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := &mockGenerator{
				arch: tt.arch,
				err:  tt.genErr,
			}
			phase := NewPhase(generator)
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
			require.NotNil(t, tt.plan.Architecture)
			assert.Equal(t, tt.arch.Components, tt.plan.Architecture.Components)
			assert.Equal(t, tt.arch.Dependencies, tt.plan.Architecture.Dependencies)
			assert.Equal(t, initialCost+tt.arch.Cost, tt.plan.TotalCost)
		})
	}
}

func TestPhase_InterfaceCompliance(t *testing.T) {
	// Verify that Phase implements planner.PlanningPhase interface
	var _ planner.PlanningPhase = (*Phase)(nil)
}
