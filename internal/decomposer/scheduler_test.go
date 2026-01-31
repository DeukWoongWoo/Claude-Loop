package decomposer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScheduler(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler()
	assert.NotNil(t, scheduler)
}

func TestDefaultScheduler_Schedule_EmptyTasks(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler()
	order, err := scheduler.Schedule([]Task{})

	require.NoError(t, err)
	assert.Empty(t, order)
}

func TestDefaultScheduler_Schedule_SingleTask(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler()
	tasks := []Task{
		createTestTask("T001", "First", "Desc", nil),
	}

	order, err := scheduler.Schedule(tasks)

	require.NoError(t, err)
	assert.Equal(t, []string{"T001"}, order)
}

func TestDefaultScheduler_Schedule_LinearChain(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler()
	tasks := []Task{
		createTestTask("T001", "First", "Desc", nil),
		createTestTask("T002", "Second", "Desc", []string{"T001"}),
		createTestTask("T003", "Third", "Desc", []string{"T002"}),
	}

	order, err := scheduler.Schedule(tasks)

	require.NoError(t, err)
	assert.Equal(t, []string{"T001", "T002", "T003"}, order)
}

func TestDefaultScheduler_Schedule_DiamondPattern(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler()
	tasks := []Task{
		createTestTask("T001", "First", "Desc", nil),
		createTestTask("T002", "Second", "Desc", []string{"T001"}),
		createTestTask("T003", "Third", "Desc", []string{"T001"}),
		createTestTask("T004", "Fourth", "Desc", []string{"T002", "T003"}),
	}

	order, err := scheduler.Schedule(tasks)

	require.NoError(t, err)
	// T001 must come before T002 and T003
	// T002 and T003 must come before T004
	assert.Equal(t, "T001", order[0])
	assert.Equal(t, "T004", order[3])
	// T002 and T003 can be in either order (sorted alphabetically)
	assert.Equal(t, []string{"T001", "T002", "T003", "T004"}, order)
}

func TestDefaultScheduler_Schedule_CyclicDependency(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler()
	tasks := []Task{
		createTestTask("T001", "First", "Desc", []string{"T002"}),
		createTestTask("T002", "Second", "Desc", []string{"T001"}),
	}

	_, err := scheduler.Schedule(tasks)

	require.Error(t, err)
	var ge *GraphError
	require.ErrorAs(t, err, &ge)
	assert.Equal(t, "cycle", ge.Type)
}

func TestDefaultScheduler_Schedule_MultipleRoots(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler()
	tasks := []Task{
		createTestTask("T003", "Third", "Desc", nil),
		createTestTask("T001", "First", "Desc", nil),
		createTestTask("T002", "Second", "Desc", nil),
	}

	order, err := scheduler.Schedule(tasks)

	require.NoError(t, err)
	// Should be sorted alphabetically since all are independent
	assert.Equal(t, []string{"T001", "T002", "T003"}, order)
}

func TestDefaultScheduler_Schedule_ComplexGraph(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler()
	tasks := []Task{
		createTestTask("T001", "Types", "Desc", nil),
		createTestTask("T002", "Errors", "Desc", nil),
		createTestTask("T003", "Parser", "Desc", []string{"T001", "T002"}),
		createTestTask("T004", "Validator", "Desc", []string{"T001", "T002"}),
		createTestTask("T005", "Graph", "Desc", []string{"T001"}),
		createTestTask("T006", "Scheduler", "Desc", []string{"T005"}),
		createTestTask("T007", "Decomposer", "Desc", []string{"T003", "T004", "T006"}),
	}

	order, err := scheduler.Schedule(tasks)

	require.NoError(t, err)
	require.Len(t, order, 7)

	// Verify dependencies are respected
	indexOf := func(id string) int {
		for i, o := range order {
			if o == id {
				return i
			}
		}
		return -1
	}

	// T001 and T002 must come first (in alphabetical order)
	assert.True(t, indexOf("T001") < indexOf("T003"))
	assert.True(t, indexOf("T002") < indexOf("T003"))
	assert.True(t, indexOf("T001") < indexOf("T004"))
	assert.True(t, indexOf("T002") < indexOf("T004"))
	assert.True(t, indexOf("T001") < indexOf("T005"))
	assert.True(t, indexOf("T005") < indexOf("T006"))
	assert.True(t, indexOf("T003") < indexOf("T007"))
	assert.True(t, indexOf("T004") < indexOf("T007"))
	assert.True(t, indexOf("T006") < indexOf("T007"))
}

func TestDefaultScheduler_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ Scheduler = (*DefaultScheduler)(nil)
}
