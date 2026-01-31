package decomposer

import (
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestTask(id, title, description string, deps []string) Task {
	return Task{
		Task: planner.Task{
			ID:           id,
			Title:        title,
			Description:  description,
			Dependencies: deps,
			Status:       planner.TaskStatusPending,
			Files:        []string{},
		},
	}
}

func TestValidator_Validate_ValidTasks(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	tasks := []Task{
		createTestTask("T001", "First Task", "Description 1", nil),
		createTestTask("T002", "Second Task", "Description 2", []string{"T001"}),
		createTestTask("T003", "Third Task", "Description 3", []string{"T001", "T002"}),
	}

	err := validator.Validate(tasks)
	assert.NoError(t, err)
}

func TestValidator_Validate_EmptyTasks(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	err := validator.Validate([]Task{})
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "tasks", ve.Field)
	assert.Contains(t, ve.Message, "at least one task is required")
}

func TestValidator_Validate_InvalidTaskID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		taskID string
	}{
		{"too few digits", "T01"},
		{"too many digits", "T0001"},
		{"lowercase t", "t001"},
		{"no T prefix", "001"},
		{"wrong prefix", "X001"},
		{"letters in digits", "T00A"},
	}

	validator := NewValidator()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tasks := []Task{
				createTestTask(tt.taskID, "Test", "Description", nil),
			}

			err := validator.Validate(tasks)
			require.Error(t, err)

			var ve *ValidationError
			require.ErrorAs(t, err, &ve)
			assert.Equal(t, "id", ve.Field)
		})
	}
}

func TestValidator_Validate_ValidTaskIDs(t *testing.T) {
	t.Parallel()

	tests := []string{"T001", "T010", "T100", "T999", "T000"}

	validator := NewValidator()

	for _, id := range tests {
		id := id
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			tasks := []Task{
				createTestTask(id, "Test", "Description", nil),
			}

			err := validator.Validate(tasks)
			assert.NoError(t, err)
		})
	}
}

func TestValidator_Validate_DuplicateTaskID(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	tasks := []Task{
		createTestTask("T001", "First", "Description", nil),
		createTestTask("T001", "Second", "Description", nil),
	}

	err := validator.Validate(tasks)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "id", ve.Field)
	assert.Equal(t, "T001", ve.TaskID)
	assert.Contains(t, ve.Message, "duplicate")
}

func TestValidator_Validate_NonExistentDependency(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	tasks := []Task{
		createTestTask("T001", "First", "Description", []string{"T999"}),
	}

	err := validator.Validate(tasks)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "dependencies", ve.Field)
	assert.Equal(t, "T001", ve.TaskID)
	assert.Contains(t, ve.Message, "T999")
}

func TestValidator_Validate_SelfReference(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	tasks := []Task{
		createTestTask("T001", "First", "Description", []string{"T001"}),
	}

	err := validator.Validate(tasks)
	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "dependencies", ve.Field)
	assert.Equal(t, "T001", ve.TaskID)
	assert.Contains(t, ve.Message, "cannot depend on itself")
}

func TestValidator_ValidateAll_CollectsAllErrors(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	tasks := []Task{
		{
			Task: planner.Task{
				ID:           "invalid",
				Title:        "",
				Description:  "",
				Dependencies: []string{"T999"},
				Status:       planner.TaskStatusPending,
			},
		},
		{
			Task: planner.Task{
				ID:           "invalid",
				Title:        "Second",
				Description:  "",
				Dependencies: nil,
				Status:       planner.TaskStatusPending,
			},
		},
	}

	errs := validator.ValidateAll(tasks)

	// Should have errors for:
	// - Invalid task ID (x2)
	// - Duplicate task ID
	// - Non-existent dependency
	// - Missing title
	// - Missing description (x2)
	assert.True(t, len(errs) >= 4, "expected at least 4 errors, got %d", len(errs))
}

func TestValidator_ValidateAll_EmptyTasks(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	errs := validator.ValidateAll([]Task{})
	require.Len(t, errs, 1)

	var ve *ValidationError
	require.ErrorAs(t, errs[0], &ve)
	assert.Equal(t, "tasks", ve.Field)
}

func TestValidator_validateTaskCompleteness_MissingTitle(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	tasks := []Task{
		{
			Task: planner.Task{
				ID:          "T001",
				Title:       "",
				Description: "Has description",
				Status:      planner.TaskStatusPending,
			},
		},
	}

	errs := validator.validateTaskCompleteness(tasks)
	require.Len(t, errs, 1)

	var ve *ValidationError
	require.ErrorAs(t, errs[0], &ve)
	assert.Equal(t, "title", ve.Field)
}

func TestValidator_validateTaskCompleteness_MissingDescription(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	tasks := []Task{
		{
			Task: planner.Task{
				ID:          "T001",
				Title:       "Has title",
				Description: "",
				Status:      planner.TaskStatusPending,
			},
		},
	}

	errs := validator.validateTaskCompleteness(tasks)
	require.Len(t, errs, 1)

	var ve *ValidationError
	require.ErrorAs(t, errs[0], &ve)
	assert.Equal(t, "description", ve.Field)
}

func TestNewValidator(t *testing.T) {
	t.Parallel()

	validator := NewValidator()
	assert.NotNil(t, validator)
}
