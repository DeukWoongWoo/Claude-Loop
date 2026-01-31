package planner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanStatus_IsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   PlanStatus
		terminal bool
	}{
		{PlanStatusPending, false},
		{PlanStatusInProgress, false},
		{PlanStatusCompleted, true},
		{PlanStatusFailed, true},
		{PlanStatusCancelled, true},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.terminal, tt.status.IsTerminal())
		})
	}
}

func TestNewPlan(t *testing.T) {
	t.Parallel()

	before := time.Now()
	plan := NewPlan("test-id", "build a feature")
	after := time.Now()

	assert.Equal(t, "test-id", plan.ID)
	assert.Equal(t, "build a feature", plan.UserPrompt)
	assert.Equal(t, PlanStatusPending, plan.Status)
	assert.NotNil(t, plan.CompletedPhases)
	assert.Empty(t, plan.CompletedPhases)
	assert.True(t, plan.CreatedAt.After(before) || plan.CreatedAt.Equal(before))
	assert.True(t, plan.CreatedAt.Before(after) || plan.CreatedAt.Equal(after))
	assert.Equal(t, plan.CreatedAt, plan.UpdatedAt)
	assert.Nil(t, plan.PRD)
	assert.Nil(t, plan.Architecture)
	assert.Nil(t, plan.TaskGraph)
	assert.Equal(t, float64(0), plan.TotalCost)
}

func TestPlan_IsPhaseCompleted(t *testing.T) {
	t.Parallel()

	t.Run("empty completed phases", func(t *testing.T) {
		t.Parallel()
		plan := NewPlan("test", "prompt")

		assert.False(t, plan.IsPhaseCompleted("prd"))
		assert.False(t, plan.IsPhaseCompleted("architecture"))
	})

	t.Run("with completed phases", func(t *testing.T) {
		t.Parallel()
		plan := NewPlan("test", "prompt")
		plan.CompletedPhases = []string{"prd", "architecture"}

		assert.True(t, plan.IsPhaseCompleted("prd"))
		assert.True(t, plan.IsPhaseCompleted("architecture"))
		assert.False(t, plan.IsPhaseCompleted("tasks"))
	})
}

func TestPlan_MarkPhaseCompleted(t *testing.T) {
	t.Parallel()

	t.Run("mark new phase", func(t *testing.T) {
		t.Parallel()
		plan := NewPlan("test", "prompt")
		initialTime := plan.UpdatedAt

		time.Sleep(time.Millisecond) // Ensure time difference
		plan.MarkPhaseCompleted("prd")

		assert.True(t, plan.IsPhaseCompleted("prd"))
		assert.Len(t, plan.CompletedPhases, 1)
		assert.True(t, plan.UpdatedAt.After(initialTime))
	})

	t.Run("mark already completed phase (no duplicate)", func(t *testing.T) {
		t.Parallel()
		plan := NewPlan("test", "prompt")
		plan.MarkPhaseCompleted("prd")
		plan.MarkPhaseCompleted("prd") // Mark again

		assert.Len(t, plan.CompletedPhases, 1) // Should not duplicate
	})

	t.Run("mark multiple phases", func(t *testing.T) {
		t.Parallel()
		plan := NewPlan("test", "prompt")
		plan.MarkPhaseCompleted("prd")
		plan.MarkPhaseCompleted("architecture")
		plan.MarkPhaseCompleted("tasks")

		assert.Len(t, plan.CompletedPhases, 3)
		assert.Equal(t, []string{"prd", "architecture", "tasks"}, plan.CompletedPhases)
	})
}

func TestPlan_AddCost(t *testing.T) {
	t.Parallel()

	plan := NewPlan("test", "prompt")
	initialTime := plan.UpdatedAt

	time.Sleep(time.Millisecond)
	plan.AddCost(0.05)

	assert.InDelta(t, 0.05, plan.TotalCost, 0.001)
	assert.True(t, plan.UpdatedAt.After(initialTime))

	plan.AddCost(0.03)
	assert.InDelta(t, 0.08, plan.TotalCost, 0.001)
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	require.NotNil(t, config)
	assert.Equal(t, ".claude/plans", config.PlanDir)
	assert.Equal(t, 3, config.MaxPhaseRetries)
	assert.Nil(t, config.OnProgress)
}

func TestConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	t.Run("with config", func(t *testing.T) {
		t.Parallel()
		config := DefaultConfig()
		assert.True(t, config.IsEnabled())
	})

	t.Run("with nil config", func(t *testing.T) {
		t.Parallel()
		var config *Config
		assert.False(t, config.IsEnabled())
	})
}

func TestPRD_Structure(t *testing.T) {
	t.Parallel()

	prd := &PRD{
		Goals:           []string{"Goal 1", "Goal 2"},
		Requirements:    []string{"Req 1"},
		Constraints:     []string{"Constraint 1"},
		SuccessCriteria: []string{"Criteria 1"},
		RawOutput:       "raw output",
	}

	assert.Len(t, prd.Goals, 2)
	assert.Len(t, prd.Requirements, 1)
	assert.Len(t, prd.Constraints, 1)
	assert.Len(t, prd.SuccessCriteria, 1)
	assert.Equal(t, "raw output", prd.RawOutput)
}

func TestArchitecture_Structure(t *testing.T) {
	t.Parallel()

	arch := &Architecture{
		Components: []Component{
			{Name: "Component1", Description: "Desc1", Files: []string{"file1.go"}},
		},
		Dependencies:  []string{"dep1"},
		FileStructure: []string{"internal/pkg/"},
		TechDecisions: []string{"Decision 1"},
		RawOutput:     "raw",
	}

	assert.Len(t, arch.Components, 1)
	assert.Equal(t, "Component1", arch.Components[0].Name)
	assert.Len(t, arch.Dependencies, 1)
	assert.Len(t, arch.FileStructure, 1)
	assert.Len(t, arch.TechDecisions, 1)
}

func TestTaskGraph_Structure(t *testing.T) {
	t.Parallel()

	graph := &TaskGraph{
		Tasks: []Task{
			{
				ID:           "T001",
				Title:        "Task 1",
				Description:  "Description",
				Dependencies: []string{},
				Status:       TaskStatusPending,
				Files:        []string{"file.go"},
			},
			{
				ID:           "T002",
				Title:        "Task 2",
				Description:  "Description 2",
				Dependencies: []string{"T001"},
				Status:       TaskStatusPending,
				Files:        []string{"file2.go"},
			},
		},
		ExecutionOrder: []string{"T001", "T002"},
		RawOutput:      "raw",
	}

	assert.Len(t, graph.Tasks, 2)
	assert.Equal(t, []string{"T001", "T002"}, graph.ExecutionOrder)
	assert.Equal(t, "T001", graph.Tasks[0].ID)
	assert.Empty(t, graph.Tasks[0].Dependencies)
	assert.Equal(t, []string{"T001"}, graph.Tasks[1].Dependencies)
}

func TestTask_Structure(t *testing.T) {
	t.Parallel()

	task := Task{
		ID:           "T001",
		Title:        "Implement feature",
		Description:  "Full description",
		Dependencies: []string{"T000"},
		Status:       TaskStatusInProgress,
		Files:        []string{"main.go", "handler.go"},
	}

	assert.Equal(t, "T001", task.ID)
	assert.Equal(t, "Implement feature", task.Title)
	assert.Equal(t, "Full description", task.Description)
	assert.Equal(t, []string{"T000"}, task.Dependencies)
	assert.Equal(t, TaskStatusInProgress, task.Status)
	assert.Len(t, task.Files, 2)
}

func TestTaskStatusConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "pending", TaskStatusPending)
	assert.Equal(t, "in_progress", TaskStatusInProgress)
	assert.Equal(t, "completed", TaskStatusCompleted)
	assert.Equal(t, "failed", TaskStatusFailed)
}

func TestComponent_Structure(t *testing.T) {
	t.Parallel()

	comp := Component{
		Name:        "PaymentService",
		Description: "Handles payments",
		Files:       []string{"payment.go", "stripe.go"},
	}

	assert.Equal(t, "PaymentService", comp.Name)
	assert.Equal(t, "Handles payments", comp.Description)
	assert.Len(t, comp.Files, 2)
}

func TestIterationResult_Structure(t *testing.T) {
	t.Parallel()

	result := &IterationResult{
		Output:                "output text",
		Cost:                  0.05,
		Duration:              time.Second,
		CompletionSignalFound: true,
	}

	assert.Equal(t, "output text", result.Output)
	assert.InDelta(t, 0.05, result.Cost, 0.001)
	assert.Equal(t, time.Second, result.Duration)
	assert.True(t, result.CompletionSignalFound)
}
