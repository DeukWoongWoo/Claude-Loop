package planner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidPhaseNames(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"prd", "architecture", "tasks"}, ValidPhaseNames)
	assert.Len(t, ValidPhaseNames, 3)
}

func TestIsValidPhaseName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected bool
	}{
		{"prd", true},
		{"architecture", true},
		{"tasks", true},
		{"PRD", false},         // Case sensitive
		{"invalid", false},
		{"", false},
		{"task", false},        // Close but not exact
		{"prdd", false},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, IsValidPhaseName(tt.name))
		})
	}
}

func TestGetPhaseIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected int
	}{
		{"prd", 0},
		{"architecture", 1},
		{"tasks", 2},
		{"invalid", -1},
		{"", -1},
		{"PRD", -1},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, GetPhaseIndex(tt.name))
		})
	}
}

func TestPhaseResult_Structure(t *testing.T) {
	t.Parallel()

	t.Run("successful result", func(t *testing.T) {
		t.Parallel()
		result := PhaseResult{
			PhaseName: "prd",
			Skipped:   false,
			Cost:      0.05,
			Duration:  1000000000, // 1 second
			Error:     nil,
		}

		assert.Equal(t, "prd", result.PhaseName)
		assert.False(t, result.Skipped)
		assert.InDelta(t, 0.05, result.Cost, 0.001)
		assert.Nil(t, result.Error)
	})

	t.Run("skipped result", func(t *testing.T) {
		t.Parallel()
		result := PhaseResult{
			PhaseName: "prd",
			Skipped:   true,
			Cost:      0,
			Duration:  0,
			Error:     nil,
		}

		assert.True(t, result.Skipped)
		assert.Equal(t, float64(0), result.Cost)
	})

	t.Run("failed result", func(t *testing.T) {
		t.Parallel()
		err := &PlannerError{Phase: "phase", Message: "execution failed"}
		result := PhaseResult{
			PhaseName: "architecture",
			Skipped:   false,
			Cost:      0.02,
			Duration:  500000000,
			Error:     err,
		}

		assert.NotNil(t, result.Error)
		assert.True(t, IsPlannerError(result.Error))
	})
}

// Verify PlanningPhase interface can be implemented
type mockPhase struct {
	name      string
	shouldRun bool
	runErr    error
}

func (m *mockPhase) Name() string                                { return m.name }
func (m *mockPhase) ShouldRun(config *Config, plan *Plan) bool   { return m.shouldRun }
func (m *mockPhase) Run(_ context.Context, _ *Plan) error        { return m.runErr }

func TestPlanningPhase_Interface(t *testing.T) {
	t.Parallel()

	// Verify the mock implements the interface
	var _ PlanningPhase = &mockPhase{}

	phase := &mockPhase{
		name:      "test-phase",
		shouldRun: true,
		runErr:    nil,
	}

	assert.Equal(t, "test-phase", phase.Name())
	assert.True(t, phase.ShouldRun(nil, nil))
	assert.Nil(t, phase.Run(nil, nil))
}
