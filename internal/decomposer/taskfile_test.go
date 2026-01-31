package decomposer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileTaskPersistence(t *testing.T) {
	t.Parallel()

	persistence := NewFileTaskPersistence()
	assert.NotNil(t, persistence)
}

func TestFileTaskPersistence_SaveTask_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	task := &Task{
		Task: planner.Task{
			ID:           "T001",
			Title:        "Test Task",
			Description:  "This is a test description",
			Status:       planner.TaskStatusPending,
			Dependencies: []string{"T000"},
			Files:        []string{"file1.go", "file2.go"},
		},
		Complexity:      "medium",
		SuccessCriteria: []string{"Tests pass"},
	}

	err := persistence.SaveTask(task, tmpDir)
	require.NoError(t, err)

	// Verify file exists
	path := persistence.TaskPath("T001", tmpDir)
	_, err = os.Stat(path)
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "id: T001")
	assert.Contains(t, string(content), "title: \"Test Task\"")
	assert.Contains(t, string(content), "status: pending")
	assert.Contains(t, string(content), "complexity: medium")
	assert.Contains(t, string(content), "This is a test description")
}

func TestFileTaskPersistence_SaveTask_NilTask(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	err := persistence.SaveTask(nil, tmpDir)
	require.Error(t, err)

	var de *DecomposerError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, "persistence", de.Phase)
}

func TestFileTaskPersistence_SaveTask_WithTimestamps(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	startTime := time.Now().Add(-1 * time.Hour)
	completeTime := time.Now()

	task := &Task{
		Task: planner.Task{
			ID:          "T001",
			Title:       "Test Task",
			Description: "Description",
			Status:      planner.TaskStatusCompleted,
		},
		StartedAt:   &startTime,
		CompletedAt: &completeTime,
	}

	err := persistence.SaveTask(task, tmpDir)
	require.NoError(t, err)

	// Verify content includes timestamps
	path := persistence.TaskPath("T001", tmpDir)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "started_at:")
	assert.Contains(t, string(content), "completed_at:")
}

func TestFileTaskPersistence_LoadTask_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	// Save a task first
	originalTask := &Task{
		Task: planner.Task{
			ID:           "T001",
			Title:        "Test Task",
			Description:  "This is the description",
			Status:       planner.TaskStatusPending,
			Dependencies: []string{"T000"},
			Files:        []string{"file.go"},
		},
		Complexity: "small",
	}

	err := persistence.SaveTask(originalTask, tmpDir)
	require.NoError(t, err)

	// Load it back
	loadedTask, err := persistence.LoadTask("T001", tmpDir)
	require.NoError(t, err)

	assert.Equal(t, originalTask.ID, loadedTask.ID)
	assert.Equal(t, originalTask.Title, loadedTask.Title)
	assert.Equal(t, originalTask.Description, loadedTask.Description)
	assert.Equal(t, originalTask.Status, loadedTask.Status)
	assert.Equal(t, originalTask.Dependencies, loadedTask.Dependencies)
	assert.Equal(t, originalTask.Files, loadedTask.Files)
	assert.Equal(t, originalTask.Complexity, loadedTask.Complexity)
}

func TestFileTaskPersistence_LoadTask_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	_, err := persistence.LoadTask("T999", tmpDir)
	require.Error(t, err)

	var de *DecomposerError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, "persistence", de.Phase)
	assert.Contains(t, de.Message, "not found")
}

func TestFileTaskPersistence_SaveTaskGraph_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	graph := &TaskGraph{
		TaskGraph: planner.TaskGraph{
			Tasks: []planner.Task{
				{ID: "T001", Title: "First", Description: "Desc", Status: planner.TaskStatusPending},
				{ID: "T002", Title: "Second", Description: "Desc", Status: planner.TaskStatusPending},
			},
			ExecutionOrder: []string{"T001", "T002"},
			RawOutput:      "raw output",
		},
		ID:        "graph-001",
		CreatedAt: time.Now(),
		Cost:      0.05,
	}

	path := filepath.Join(tmpDir, "graph.yaml")
	err := persistence.SaveTaskGraph(graph, path)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestFileTaskPersistence_SaveTaskGraph_NilGraph(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	path := filepath.Join(tmpDir, "graph.yaml")
	err := persistence.SaveTaskGraph(nil, path)
	require.Error(t, err)
}

func TestFileTaskPersistence_LoadTaskGraph_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	originalGraph := &TaskGraph{
		TaskGraph: planner.TaskGraph{
			Tasks: []planner.Task{
				{ID: "T001", Title: "First", Description: "Desc", Status: planner.TaskStatusPending},
			},
			ExecutionOrder: []string{"T001"},
		},
		ID:   "graph-001",
		Cost: 0.05,
	}

	path := filepath.Join(tmpDir, "graph.yaml")
	err := persistence.SaveTaskGraph(originalGraph, path)
	require.NoError(t, err)

	// Load it back
	loadedGraph, err := persistence.LoadTaskGraph(path)
	require.NoError(t, err)

	assert.Equal(t, originalGraph.ID, loadedGraph.ID)
	assert.Equal(t, originalGraph.Cost, loadedGraph.Cost)
	assert.Equal(t, originalGraph.ExecutionOrder, loadedGraph.ExecutionOrder)
	require.Len(t, loadedGraph.Tasks, 1)
	assert.Equal(t, "T001", loadedGraph.Tasks[0].ID)
}

func TestFileTaskPersistence_LoadTaskGraph_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	path := filepath.Join(tmpDir, "nonexistent.yaml")
	_, err := persistence.LoadTaskGraph(path)
	require.Error(t, err)

	var de *DecomposerError
	require.ErrorAs(t, err, &de)
	assert.Contains(t, de.Message, "not found")
}

func TestFileTaskPersistence_ListTasks(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	// Create some task files
	tasks := []*Task{
		{Task: planner.Task{ID: "T001", Title: "First", Description: "Desc", Status: planner.TaskStatusPending}},
		{Task: planner.Task{ID: "T002", Title: "Second", Description: "Desc", Status: planner.TaskStatusPending}},
		{Task: planner.Task{ID: "T003", Title: "Third", Description: "Desc", Status: planner.TaskStatusPending}},
	}

	for _, task := range tasks {
		err := persistence.SaveTask(task, tmpDir)
		require.NoError(t, err)
	}

	// List tasks
	taskIDs, err := persistence.ListTasks(tmpDir)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"T001", "T002", "T003"}, taskIDs)
}

func TestFileTaskPersistence_ListTasks_EmptyDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	taskIDs, err := persistence.ListTasks(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, taskIDs)
}

func TestFileTaskPersistence_ListTasks_NonExistentDir(t *testing.T) {
	t.Parallel()

	persistence := NewFileTaskPersistence()

	taskIDs, err := persistence.ListTasks("/nonexistent/path")
	require.NoError(t, err)
	assert.Empty(t, taskIDs)
}

func TestFileTaskPersistence_ListTasks_IgnoresTmpFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	// Create a task file and a tmp file
	task := &Task{Task: planner.Task{ID: "T001", Title: "First", Description: "Desc", Status: planner.TaskStatusPending}}
	err := persistence.SaveTask(task, tmpDir)
	require.NoError(t, err)

	// Create a tmp file
	tmpFile := filepath.Join(tmpDir, "T002.md.tmp")
	err = os.WriteFile(tmpFile, []byte("tmp content"), 0600)
	require.NoError(t, err)

	// List should only return T001
	taskIDs, err := persistence.ListTasks(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"T001"}, taskIDs)
}

func TestFileTaskPersistence_TaskPath(t *testing.T) {
	t.Parallel()

	persistence := NewFileTaskPersistence()

	path := persistence.TaskPath("T001", "/tasks")
	assert.Equal(t, "/tasks/T001.md", path)
}

func TestFileTaskPersistence_TaskPath_PreventTraversal(t *testing.T) {
	t.Parallel()

	persistence := NewFileTaskPersistence()

	tests := []struct {
		name   string
		taskID string
	}{
		{"slash", "../secret"},
		{"backslash", "..\\secret"},
		{"dot", "."},
		{"dotdot", ".."},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := persistence.TaskPath(tt.taskID, "/tasks")
			assert.Equal(t, "/tasks/invalid-task-id.md", path)
		})
	}
}

func TestFileTaskPersistence_RoundTrip(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	persistence := NewFileTaskPersistence()

	// Create a task with all fields
	startTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	originalTask := &Task{
		Task: planner.Task{
			ID:           "T001",
			Title:        "Complete Task",
			Description:  "A complete task with all fields",
			Status:       planner.TaskStatusInProgress,
			Dependencies: []string{"T000"},
			Files:        []string{"file1.go", "file2.go"},
		},
		Complexity:      "large",
		SuccessCriteria: []string{"Tests pass", "Coverage > 90%"},
		StartedAt:       &startTime,
	}

	// Save
	err := persistence.SaveTask(originalTask, tmpDir)
	require.NoError(t, err)

	// Load
	loadedTask, err := persistence.LoadTask("T001", tmpDir)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, originalTask.ID, loadedTask.ID)
	assert.Equal(t, originalTask.Title, loadedTask.Title)
	assert.Equal(t, originalTask.Description, loadedTask.Description)
	assert.Equal(t, originalTask.Status, loadedTask.Status)
	assert.Equal(t, originalTask.Dependencies, loadedTask.Dependencies)
	assert.Equal(t, originalTask.Files, loadedTask.Files)
	assert.Equal(t, originalTask.Complexity, loadedTask.Complexity)
	require.NotNil(t, loadedTask.StartedAt)
	assert.True(t, originalTask.StartedAt.Equal(*loadedTask.StartedAt))
}

func TestFileTaskPersistence_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ TaskPersistence = (*FileTaskPersistence)(nil)
}
