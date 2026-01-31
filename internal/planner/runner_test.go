package planner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPlanningPhase implements PlanningPhase for testing.
type MockPlanningPhase struct {
	name        string
	shouldRun   bool
	runErr      error
	runFunc     func(ctx context.Context, plan *Plan) error
	runCalled   bool
	runCount    int
}

func (m *MockPlanningPhase) Name() string {
	return m.name
}

func (m *MockPlanningPhase) ShouldRun(_ *Config, _ *Plan) bool {
	return m.shouldRun
}

func (m *MockPlanningPhase) Run(ctx context.Context, plan *Plan) error {
	m.runCalled = true
	m.runCount++
	if m.runFunc != nil {
		return m.runFunc(ctx, plan)
	}
	return m.runErr
}

// MockPersistence implements Persistence for testing.
type MockPersistence struct {
	plans      map[string]*Plan
	saveErr    error
	loadErr    error
	saveCalled bool
	loadCalled bool
	planDir    string
}

func NewMockPersistence() *MockPersistence {
	return &MockPersistence{
		plans:   make(map[string]*Plan),
		planDir: "/mock/plans",
	}
}

func (m *MockPersistence) Save(plan *Plan, path string) error {
	m.saveCalled = true
	if m.saveErr != nil {
		return m.saveErr
	}
	m.plans[path] = plan
	return nil
}

func (m *MockPersistence) Load(path string) (*Plan, error) {
	m.loadCalled = true
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	if plan, ok := m.plans[path]; ok {
		return plan, nil
	}
	return nil, ErrPlanNotFound
}

func (m *MockPersistence) Exists(path string) bool {
	_, ok := m.plans[path]
	return ok
}

func (m *MockPersistence) Delete(path string) error {
	delete(m.plans, path)
	return nil
}

func (m *MockPersistence) DefaultPlanPath(planID string) string {
	return m.planDir + "/" + planID + ".yaml"
}

func TestNewPhaseRunner(t *testing.T) {
	t.Parallel()

	t.Run("with nil config", func(t *testing.T) {
		t.Parallel()
		runner := NewPhaseRunner(nil, NewMockPersistence())

		require.NotNil(t, runner)
		assert.NotNil(t, runner.config)
		assert.Equal(t, DefaultConfig().PlanDir, runner.config.PlanDir)
	})

	t.Run("with custom config", func(t *testing.T) {
		t.Parallel()
		config := &Config{PlanDir: "/custom/path", MaxPhaseRetries: 5}
		runner := NewPhaseRunner(config, NewMockPersistence())

		assert.Equal(t, "/custom/path", runner.config.PlanDir)
		assert.Equal(t, 5, runner.config.MaxPhaseRetries)
	})

	t.Run("with phases", func(t *testing.T) {
		t.Parallel()
		phase1 := &MockPlanningPhase{name: "phase1"}
		phase2 := &MockPlanningPhase{name: "phase2"}
		runner := NewPhaseRunner(nil, NewMockPersistence(), phase1, phase2)

		assert.Len(t, runner.phases, 2)
		assert.Equal(t, "phase1", runner.phases[0].Name())
		assert.Equal(t, "phase2", runner.phases[1].Name())
	})

	t.Run("with nil persistence creates default", func(t *testing.T) {
		t.Parallel()
		config := &Config{PlanDir: "/test/dir"}
		runner := NewPhaseRunner(config, nil)

		require.NotNil(t, runner.persistence)
	})
}

func TestPhaseRunner_Run(t *testing.T) {
	t.Parallel()

	t.Run("run all phases successfully", func(t *testing.T) {
		t.Parallel()
		phase1 := &MockPlanningPhase{name: "prd", shouldRun: true}
		phase2 := &MockPlanningPhase{name: "architecture", shouldRun: true}
		phase3 := &MockPlanningPhase{name: "tasks", shouldRun: true}

		persistence := NewMockPersistence()
		runner := NewPhaseRunner(nil, persistence, phase1, phase2, phase3)
		plan := NewPlan("test", "prompt")

		result, err := runner.Run(context.Background(), plan)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, PlanStatusCompleted, plan.Status)
		assert.Len(t, result.PhaseResults, 3)
		assert.True(t, phase1.runCalled)
		assert.True(t, phase2.runCalled)
		assert.True(t, phase3.runCalled)
		assert.Equal(t, []string{"prd", "architecture", "tasks"}, plan.CompletedPhases)
		assert.True(t, persistence.saveCalled)
	})

	t.Run("skip already completed phases", func(t *testing.T) {
		t.Parallel()
		phase1 := &MockPlanningPhase{name: "prd", shouldRun: false} // Already completed
		phase2 := &MockPlanningPhase{name: "architecture", shouldRun: true}

		runner := NewPhaseRunner(nil, NewMockPersistence(), phase1, phase2)
		plan := NewPlan("test", "prompt")
		plan.CompletedPhases = []string{"prd"}

		result, err := runner.Run(context.Background(), plan)

		require.NoError(t, err)
		assert.False(t, phase1.runCalled)
		assert.True(t, phase2.runCalled)
		assert.True(t, result.PhaseResults[0].Skipped)
		assert.False(t, result.PhaseResults[1].Skipped)
	})

	t.Run("stop on phase error", func(t *testing.T) {
		t.Parallel()
		phaseErr := errors.New("phase failed")
		phase1 := &MockPlanningPhase{name: "prd", shouldRun: true}
		phase2 := &MockPlanningPhase{name: "architecture", shouldRun: true, runErr: phaseErr}
		phase3 := &MockPlanningPhase{name: "tasks", shouldRun: true}

		runner := NewPhaseRunner(nil, NewMockPersistence(), phase1, phase2, phase3)
		plan := NewPlan("test", "prompt")

		result, err := runner.Run(context.Background(), plan)

		require.Error(t, err)
		assert.True(t, IsPlannerError(err))
		assert.Equal(t, PlanStatusFailed, plan.Status)
		assert.True(t, phase1.runCalled)
		assert.True(t, phase2.runCalled)
		assert.False(t, phase3.runCalled) // Should not run after error
		assert.Len(t, result.PhaseResults, 2)
		assert.NotNil(t, result.PhaseResults[1].Error)
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())

		phase1 := &MockPlanningPhase{
			name:      "prd",
			shouldRun: true,
			runFunc: func(ctx context.Context, plan *Plan) error {
				cancel() // Cancel during first phase
				return nil
			},
		}
		phase2 := &MockPlanningPhase{name: "architecture", shouldRun: true}

		runner := NewPhaseRunner(nil, NewMockPersistence(), phase1, phase2)
		plan := NewPlan("test", "prompt")

		result, err := runner.Run(ctx, plan)

		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.True(t, phase1.runCalled)
		// phase2 may or may not be called depending on timing
		assert.NotNil(t, result)
	})

	t.Run("nil plan", func(t *testing.T) {
		t.Parallel()
		runner := NewPhaseRunner(nil, NewMockPersistence())

		result, err := runner.Run(context.Background(), nil)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsPlannerError(err))
	})

	t.Run("progress callback called", func(t *testing.T) {
		t.Parallel()
		var progressCalls []string

		config := &Config{
			PlanDir: ".claude/plans",
			OnProgress: func(phase, status string) {
				progressCalls = append(progressCalls, phase+":"+status)
			},
		}

		phase1 := &MockPlanningPhase{name: "prd", shouldRun: true}
		phase2 := &MockPlanningPhase{name: "arch", shouldRun: false} // Skipped

		runner := NewPhaseRunner(config, NewMockPersistence(), phase1, phase2)
		plan := NewPlan("test", "prompt")

		_, err := runner.Run(context.Background(), plan)

		require.NoError(t, err)
		assert.Contains(t, progressCalls, "prd:running")
		assert.Contains(t, progressCalls, "prd:completed")
		assert.Contains(t, progressCalls, "arch:skipped")
	})

	t.Run("save error propagates", func(t *testing.T) {
		t.Parallel()
		persistence := NewMockPersistence()
		persistence.saveErr = errors.New("save failed")

		phase1 := &MockPlanningPhase{name: "prd", shouldRun: true}
		runner := NewPhaseRunner(nil, persistence, phase1)
		plan := NewPlan("test", "prompt")

		result, err := runner.Run(context.Background(), plan)

		require.Error(t, err)
		assert.NotNil(t, result)
	})
}

func TestPhaseRunner_RunPhase(t *testing.T) {
	t.Parallel()

	t.Run("run specific phase", func(t *testing.T) {
		t.Parallel()
		phase1 := &MockPlanningPhase{name: "prd", shouldRun: true}
		phase2 := &MockPlanningPhase{name: "architecture", shouldRun: true}

		runner := NewPhaseRunner(nil, NewMockPersistence(), phase1, phase2)
		plan := NewPlan("test", "prompt")

		result, err := runner.RunPhase(context.Background(), plan, "architecture")

		require.NoError(t, err)
		assert.False(t, phase1.runCalled)
		assert.True(t, phase2.runCalled)
		assert.Equal(t, "architecture", result.PhaseName)
		assert.False(t, result.Skipped)
	})

	t.Run("phase not found", func(t *testing.T) {
		t.Parallel()
		runner := NewPhaseRunner(nil, NewMockPersistence())
		plan := NewPlan("test", "prompt")

		result, err := runner.RunPhase(context.Background(), plan, "nonexistent")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsPlannerError(err))
	})

	t.Run("nil plan", func(t *testing.T) {
		t.Parallel()
		runner := NewPhaseRunner(nil, NewMockPersistence())

		result, err := runner.RunPhase(context.Background(), nil, "prd")

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("phase skipped", func(t *testing.T) {
		t.Parallel()
		phase := &MockPlanningPhase{name: "prd", shouldRun: false}
		runner := NewPhaseRunner(nil, NewMockPersistence(), phase)
		plan := NewPlan("test", "prompt")

		result, err := runner.RunPhase(context.Background(), plan, "prd")

		require.NoError(t, err)
		assert.True(t, result.Skipped)
		assert.False(t, phase.runCalled)
	})

	t.Run("phase error", func(t *testing.T) {
		t.Parallel()
		phaseErr := errors.New("execution failed")
		phase := &MockPlanningPhase{name: "prd", shouldRun: true, runErr: phaseErr}
		runner := NewPhaseRunner(nil, NewMockPersistence(), phase)
		plan := NewPlan("test", "prompt")

		result, err := runner.RunPhase(context.Background(), plan, "prd")

		require.Error(t, err)
		assert.Equal(t, phaseErr, err)
		assert.NotNil(t, result)
		assert.Equal(t, phaseErr, result.Error)
		assert.Equal(t, PlanStatusFailed, plan.Status)
	})
}

func TestPhaseRunner_ResumePlan(t *testing.T) {
	t.Parallel()

	t.Run("resume from saved plan", func(t *testing.T) {
		t.Parallel()
		persistence := NewMockPersistence()

		// Pre-save a plan
		savedPlan := NewPlan("resume-test", "prompt")
		savedPlan.CompletedPhases = []string{"prd"}
		savedPlan.Status = PlanStatusInProgress
		path := persistence.DefaultPlanPath(savedPlan.ID)
		persistence.plans[path] = savedPlan

		phase1 := &MockPlanningPhase{name: "prd", shouldRun: false}     // Already completed
		phase2 := &MockPlanningPhase{name: "architecture", shouldRun: true}

		runner := NewPhaseRunner(nil, persistence, phase1, phase2)

		result, err := runner.ResumePlan(context.Background(), path)

		require.NoError(t, err)
		assert.True(t, persistence.loadCalled)
		assert.False(t, phase1.runCalled)
		assert.True(t, phase2.runCalled)
		assert.Equal(t, PlanStatusCompleted, result.Plan.Status)
	})

	t.Run("plan not found", func(t *testing.T) {
		t.Parallel()
		persistence := NewMockPersistence()
		runner := NewPhaseRunner(nil, persistence)

		result, err := runner.ResumePlan(context.Background(), "/nonexistent.yaml")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrPlanNotFound, err)
	})

	t.Run("resume completed plan (no-op)", func(t *testing.T) {
		t.Parallel()
		persistence := NewMockPersistence()

		// Pre-save a completed plan
		savedPlan := NewPlan("completed", "prompt")
		savedPlan.Status = PlanStatusCompleted
		savedPlan.CompletedPhases = []string{"prd", "architecture", "tasks"}
		path := persistence.DefaultPlanPath(savedPlan.ID)
		persistence.plans[path] = savedPlan

		phase := &MockPlanningPhase{name: "prd", shouldRun: true}
		runner := NewPhaseRunner(nil, persistence, phase)

		result, err := runner.ResumePlan(context.Background(), path)

		require.NoError(t, err)
		assert.False(t, phase.runCalled) // Should not run any phase
		assert.Equal(t, PlanStatusCompleted, result.Plan.Status)
	})

	t.Run("resume failed plan (retry)", func(t *testing.T) {
		t.Parallel()
		persistence := NewMockPersistence()

		// Pre-save a failed plan
		savedPlan := NewPlan("failed", "prompt")
		savedPlan.Status = PlanStatusFailed
		savedPlan.CompletedPhases = []string{"prd"}
		path := persistence.DefaultPlanPath(savedPlan.ID)
		persistence.plans[path] = savedPlan

		phase1 := &MockPlanningPhase{name: "prd", shouldRun: false}
		phase2 := &MockPlanningPhase{name: "architecture", shouldRun: true}
		runner := NewPhaseRunner(nil, persistence, phase1, phase2)

		result, err := runner.ResumePlan(context.Background(), path)

		require.NoError(t, err)
		assert.True(t, phase2.runCalled)
		assert.Equal(t, PlanStatusCompleted, result.Plan.Status)
	})
}

func TestPhaseRunner_Accessors(t *testing.T) {
	t.Parallel()

	config := &Config{PlanDir: "/test"}
	persistence := NewMockPersistence()
	phase := &MockPlanningPhase{name: "test"}
	runner := NewPhaseRunner(config, persistence, phase)

	t.Run("Config", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, config, runner.Config())
	})

	t.Run("Phases", func(t *testing.T) {
		t.Parallel()
		phases := runner.Phases()
		require.Len(t, phases, 1)
		assert.Equal(t, "test", phases[0].Name())
	})

	t.Run("Persistence", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, persistence, runner.Persistence())
	})
}

func TestRunResult_Structure(t *testing.T) {
	t.Parallel()

	plan := NewPlan("test", "prompt")
	result := &RunResult{
		Plan: plan,
		PhaseResults: []PhaseResult{
			{PhaseName: "prd", Skipped: false, Cost: 0.05, Duration: time.Second},
			{PhaseName: "arch", Skipped: true},
		},
		TotalCost:     0.05,
		TotalDuration: 2 * time.Second,
		Error:         nil,
	}

	assert.Equal(t, plan, result.Plan)
	assert.Len(t, result.PhaseResults, 2)
	assert.InDelta(t, 0.05, result.TotalCost, 0.001)
	assert.Equal(t, 2*time.Second, result.TotalDuration)
	assert.Nil(t, result.Error)
}
