package planner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFilePersistence(t *testing.T) {
	t.Parallel()

	p := NewFilePersistence("/tmp/plans")

	require.NotNil(t, p)
	assert.Equal(t, "/tmp/plans", p.planDir)
}

func TestFilePersistence_Save(t *testing.T) {
	t.Parallel()

	t.Run("save valid plan", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		plan := NewPlan("test-plan", "test prompt")
		plan.PRD = &PRD{
			Goals:        []string{"Goal 1"},
			Requirements: []string{"Req 1"},
		}

		path := p.DefaultPlanPath(plan.ID)
		err := p.Save(plan, path)

		require.NoError(t, err)
		assert.True(t, p.Exists(path))

		// Verify file content
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(data), "test-plan")
		assert.Contains(t, string(data), "Goal 1")
	})

	t.Run("save nil plan", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		err := p.Save(nil, filepath.Join(tmpDir, "nil.yaml"))

		require.Error(t, err)
		assert.True(t, IsPlannerError(err))
	})

	t.Run("save creates directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "nested", "dir")
		p := NewFilePersistence(nestedDir)

		plan := NewPlan("test", "prompt")
		path := p.DefaultPlanPath(plan.ID)
		err := p.Save(plan, path)

		require.NoError(t, err)
		assert.True(t, p.Exists(path))
	})

	t.Run("atomic write (temp file removed on success)", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		plan := NewPlan("atomic-test", "prompt")
		path := p.DefaultPlanPath(plan.ID)
		err := p.Save(plan, path)

		require.NoError(t, err)

		// Temp file should not exist
		tmpPath := path + ".tmp"
		assert.False(t, p.Exists(tmpPath))
	})
}

func TestFilePersistence_Load(t *testing.T) {
	t.Parallel()

	t.Run("load existing plan", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		// Save a plan first
		original := NewPlan("load-test", "original prompt")
		original.PRD = &PRD{Goals: []string{"Goal 1"}}
		original.Status = PlanStatusInProgress
		original.CompletedPhases = []string{"prd"}
		original.TotalCost = 0.05

		path := p.DefaultPlanPath(original.ID)
		require.NoError(t, p.Save(original, path))

		// Load it back
		loaded, err := p.Load(path)

		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "load-test", loaded.ID)
		assert.Equal(t, "original prompt", loaded.UserPrompt)
		assert.Equal(t, PlanStatusInProgress, loaded.Status)
		assert.Equal(t, []string{"prd"}, loaded.CompletedPhases)
		assert.InDelta(t, 0.05, loaded.TotalCost, 0.001)
		require.NotNil(t, loaded.PRD)
		assert.Equal(t, []string{"Goal 1"}, loaded.PRD.Goals)
	})

	t.Run("load non-existent plan", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		path := filepath.Join(tmpDir, "nonexistent.yaml")
		loaded, err := p.Load(path)

		require.Error(t, err)
		assert.Nil(t, loaded)
		assert.Equal(t, ErrPlanNotFound, err)
	})

	t.Run("load invalid yaml", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		// Write invalid YAML
		path := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(path, []byte("invalid: yaml: content: ["), 0644)
		require.NoError(t, err)

		loaded, err := p.Load(path)

		require.Error(t, err)
		assert.Nil(t, loaded)
		assert.True(t, IsPlannerError(err))
	})
}

func TestFilePersistence_Exists(t *testing.T) {
	t.Parallel()

	t.Run("existing file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		path := filepath.Join(tmpDir, "exists.yaml")
		require.NoError(t, os.WriteFile(path, []byte("test"), 0644))

		assert.True(t, p.Exists(path))
	})

	t.Run("non-existing file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		path := filepath.Join(tmpDir, "notexists.yaml")

		assert.False(t, p.Exists(path))
	})
}

func TestFilePersistence_Delete(t *testing.T) {
	t.Parallel()

	t.Run("delete existing file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		// Create a file
		plan := NewPlan("delete-test", "prompt")
		path := p.DefaultPlanPath(plan.ID)
		require.NoError(t, p.Save(plan, path))
		assert.True(t, p.Exists(path))

		// Delete it
		err := p.Delete(path)

		require.NoError(t, err)
		assert.False(t, p.Exists(path))
	})

	t.Run("delete non-existing file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		path := filepath.Join(tmpDir, "notexists.yaml")
		err := p.Delete(path)

		require.Error(t, err)
		assert.Equal(t, ErrPlanNotFound, err)
	})
}

func TestFilePersistence_DefaultPlanPath(t *testing.T) {
	t.Parallel()

	t.Run("valid plan ID", func(t *testing.T) {
		t.Parallel()
		p := NewFilePersistence("/home/user/.claude/plans")

		path := p.DefaultPlanPath("my-plan-id")

		assert.Equal(t, "/home/user/.claude/plans/my-plan-id.yaml", path)
	})

	t.Run("path traversal attack prevention", func(t *testing.T) {
		t.Parallel()
		p := NewFilePersistence("/home/user/.claude/plans")

		// These should all return the safe fallback path
		testCases := []string{
			"../../../etc/passwd",
			"..\\..\\windows\\system32",
			"foo/bar",
			"foo\\bar",
			".",
			"..",
			"",
		}

		for _, maliciousID := range testCases {
			path := p.DefaultPlanPath(maliciousID)
			assert.Equal(t, "/home/user/.claude/plans/invalid-plan-id.yaml", path,
				"Should return safe fallback for ID: %q", maliciousID)
		}
	})

	t.Run("valid IDs with special characters", func(t *testing.T) {
		t.Parallel()
		p := NewFilePersistence("/tmp/plans")

		// These are valid plan IDs (no path separators)
		validIDs := []string{
			"plan-123",
			"my_plan",
			"plan.v2",
			"UPPERCASE",
			"123-456-789",
		}

		for _, validID := range validIDs {
			path := p.DefaultPlanPath(validID)
			expected := "/tmp/plans/" + validID + ".yaml"
			assert.Equal(t, expected, path, "Should accept valid ID: %q", validID)
		}
	})
}

func TestFilePersistence_ListPlans(t *testing.T) {
	t.Parallel()

	t.Run("list plans in directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		// Create some plan files
		require.NoError(t, p.Save(NewPlan("plan1", "p1"), p.DefaultPlanPath("plan1")))
		require.NoError(t, p.Save(NewPlan("plan2", "p2"), p.DefaultPlanPath("plan2")))
		require.NoError(t, p.Save(NewPlan("plan3", "p3"), p.DefaultPlanPath("plan3")))

		// Create a non-yaml file (should be ignored)
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0644))

		// Create a subdirectory (should be ignored)
		require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755))

		plans, err := p.ListPlans()

		require.NoError(t, err)
		assert.Len(t, plans, 3)
		assert.Contains(t, plans, "plan1")
		assert.Contains(t, plans, "plan2")
		assert.Contains(t, plans, "plan3")
	})

	t.Run("list empty directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		plans, err := p.ListPlans()

		require.NoError(t, err)
		assert.Empty(t, plans)
	})

	t.Run("list non-existing directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		nonExistDir := filepath.Join(tmpDir, "nonexistent")
		p := NewFilePersistence(nonExistDir)

		plans, err := p.ListPlans()

		require.NoError(t, err) // Should not error, just return empty
		assert.Empty(t, plans)
	})

	t.Run("ignores temp files", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		p := NewFilePersistence(tmpDir)

		// Create a plan and a temp file
		require.NoError(t, p.Save(NewPlan("plan1", "p1"), p.DefaultPlanPath("plan1")))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "plan2.yaml.tmp"), []byte("temp"), 0644))

		plans, err := p.ListPlans()

		require.NoError(t, err)
		assert.Len(t, plans, 1)
		assert.Equal(t, []string{"plan1"}, plans)
	})
}

func TestFilePersistence_PlanDir(t *testing.T) {
	t.Parallel()

	p := NewFilePersistence("/custom/path")

	assert.Equal(t, "/custom/path", p.PlanDir())
}

func TestFilePersistence_SaveLoadRoundtrip(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	p := NewFilePersistence(tmpDir)

	// Create a complex plan
	now := time.Now().Truncate(time.Second) // Truncate for YAML round-trip
	original := &Plan{
		ID:         "roundtrip-test",
		UserPrompt: "Build a payment system",
		CreatedAt:  now,
		UpdatedAt:  now,
		Status:     PlanStatusInProgress,
		PRD: &PRD{
			Goals:           []string{"Process payments", "Handle refunds"},
			Requirements:    []string{"Stripe integration"},
			Constraints:     []string{"No breaking changes"},
			SuccessCriteria: []string{"95% test coverage"},
			RawOutput:       "raw prd output",
		},
		Architecture: &Architecture{
			Components: []Component{
				{Name: "PaymentService", Description: "Handles payments", Files: []string{"payment.go"}},
			},
			Dependencies:  []string{"stripe-go"},
			FileStructure: []string{"internal/payment/"},
			TechDecisions: []string{"Use Stripe API v2"},
			RawOutput:     "raw arch output",
		},
		TaskGraph: &TaskGraph{
			Tasks: []Task{
				{ID: "T001", Title: "Setup", Description: "Setup project", Status: TaskStatusCompleted},
				{ID: "T002", Title: "Implement", Description: "Implement feature", Dependencies: []string{"T001"}, Status: TaskStatusPending},
			},
			ExecutionOrder: []string{"T001", "T002"},
			RawOutput:      "raw tasks output",
		},
		CurrentTaskID:   "T002",
		CompletedPhases: []string{"prd", "architecture"},
		TotalCost:       0.15,
	}

	path := p.DefaultPlanPath(original.ID)

	// Save
	err := p.Save(original, path)
	require.NoError(t, err)

	// Load
	loaded, err := p.Load(path)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, original.ID, loaded.ID)
	assert.Equal(t, original.UserPrompt, loaded.UserPrompt)
	assert.Equal(t, original.Status, loaded.Status)
	assert.Equal(t, original.CurrentTaskID, loaded.CurrentTaskID)
	assert.Equal(t, original.CompletedPhases, loaded.CompletedPhases)
	assert.InDelta(t, original.TotalCost, loaded.TotalCost, 0.001)

	// Verify PRD
	require.NotNil(t, loaded.PRD)
	assert.Equal(t, original.PRD.Goals, loaded.PRD.Goals)
	assert.Equal(t, original.PRD.Requirements, loaded.PRD.Requirements)
	assert.Equal(t, original.PRD.Constraints, loaded.PRD.Constraints)
	assert.Equal(t, original.PRD.SuccessCriteria, loaded.PRD.SuccessCriteria)

	// Verify Architecture
	require.NotNil(t, loaded.Architecture)
	assert.Len(t, loaded.Architecture.Components, 1)
	assert.Equal(t, "PaymentService", loaded.Architecture.Components[0].Name)

	// Verify TaskGraph
	require.NotNil(t, loaded.TaskGraph)
	assert.Len(t, loaded.TaskGraph.Tasks, 2)
	assert.Equal(t, []string{"T001", "T002"}, loaded.TaskGraph.ExecutionOrder)
}

// Test Persistence interface compliance
var _ Persistence = (*FilePersistence)(nil)
