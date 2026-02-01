package prd

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
	prd *PRD
	err error
}

func (m *mockGenerator) Generate(ctx context.Context, userPrompt string) (*PRD, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.prd, nil
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
	assert.Equal(t, "prd", phase.Name())
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
			name: "phase not completed",
			plan: &planner.Plan{
				CompletedPhases: []string{},
			},
			expected: true,
		},
		{
			name: "phase already completed",
			plan: &planner.Plan{
				CompletedPhases: []string{"prd"},
			},
			expected: false,
		},
		{
			name: "other phases completed",
			plan: &planner.Plan{
				CompletedPhases: []string{"architecture"},
			},
			expected: true,
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
		prd     *PRD
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
			name: "empty user prompt",
			plan: &planner.Plan{
				UserPrompt:      "",
				CompletedPhases: []string{},
			},
			wantErr: true,
			errMsg:  "user prompt is empty",
		},
		{
			name: "successful generation",
			plan: &planner.Plan{
				UserPrompt:      "Add payment feature",
				CompletedPhases: []string{},
			},
			prd: &PRD{
				PRD: planner.PRD{
					Goals:           []string{"Implement payment"},
					Requirements:    []string{"Stripe integration"},
					Constraints:     []string{"PCI compliance"},
					SuccessCriteria: []string{"Payment works"},
				},
				Cost:     0.05,
				Duration: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "generator error",
			plan: &planner.Plan{
				UserPrompt:      "Add feature",
				CompletedPhases: []string{},
			},
			genErr:  errors.New("generation failed"),
			wantErr: true,
			errMsg:  "PRD generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := &mockGenerator{
				prd: tt.prd,
				err: tt.genErr,
			}
			phase := NewPhase(generator)
			require.NotNil(t, phase)

			err := phase.Run(ctx, tt.plan)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tt.plan.PRD)
			assert.Equal(t, tt.prd.Goals, tt.plan.PRD.Goals)
			assert.Equal(t, tt.prd.Requirements, tt.plan.PRD.Requirements)
			assert.Equal(t, tt.prd.Cost, tt.plan.TotalCost)
		})
	}
}

func TestPhase_InterfaceCompliance(t *testing.T) {
	// Verify that Phase implements planner.PlanningPhase interface
	var _ planner.PlanningPhase = (*Phase)(nil)
}
