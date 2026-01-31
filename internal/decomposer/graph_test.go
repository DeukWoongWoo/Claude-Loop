package decomposer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyGraph(t *testing.T) {
	t.Parallel()

	tasks := []Task{
		createTestTask("T001", "First", "Desc", nil),
		createTestTask("T002", "Second", "Desc", []string{"T001"}),
	}

	graph := NewDependencyGraph(tasks)

	assert.Equal(t, 2, graph.Size())
	assert.NotNil(t, graph.nodes["T001"])
	assert.NotNil(t, graph.nodes["T002"])
}

func TestDependencyGraph_DetectCycle_NoCycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tasks []Task
	}{
		{
			name: "single node",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", nil),
			},
		},
		{
			name: "linear chain",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", nil),
				createTestTask("T002", "Second", "Desc", []string{"T001"}),
				createTestTask("T003", "Third", "Desc", []string{"T002"}),
			},
		},
		{
			name: "diamond pattern",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", nil),
				createTestTask("T002", "Second", "Desc", []string{"T001"}),
				createTestTask("T003", "Third", "Desc", []string{"T001"}),
				createTestTask("T004", "Fourth", "Desc", []string{"T002", "T003"}),
			},
		},
		{
			name: "multiple roots",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", nil),
				createTestTask("T002", "Second", "Desc", nil),
				createTestTask("T003", "Third", "Desc", []string{"T001", "T002"}),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			graph := NewDependencyGraph(tt.tasks)
			cycle := graph.DetectCycle()
			assert.Nil(t, cycle)
		})
	}
}

func TestDependencyGraph_DetectCycle_WithCycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tasks         []Task
		expectedCycle []string
	}{
		{
			name: "simple cycle A -> B -> A",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", []string{"T002"}),
				createTestTask("T002", "Second", "Desc", []string{"T001"}),
			},
			expectedCycle: []string{"T001", "T002", "T001"},
		},
		{
			name: "three node cycle A -> B -> C -> A",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", []string{"T003"}),
				createTestTask("T002", "Second", "Desc", []string{"T001"}),
				createTestTask("T003", "Third", "Desc", []string{"T002"}),
			},
			expectedCycle: []string{"T001", "T003", "T002", "T001"},
		},
		{
			name: "self reference",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", []string{"T001"}),
			},
			expectedCycle: []string{"T001", "T001"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			graph := NewDependencyGraph(tt.tasks)
			cycle := graph.DetectCycle()

			require.NotNil(t, cycle, "expected cycle but got nil")
			// Verify cycle starts and ends with same element
			assert.Equal(t, cycle[0], cycle[len(cycle)-1], "cycle should start and end with same element")
		})
	}
}

func TestDependencyGraph_TopologicalSort_ValidGraph(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tasks         []Task
		expectedOrder []string
	}{
		{
			name: "single node",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", nil),
			},
			expectedOrder: []string{"T001"},
		},
		{
			name: "linear chain",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", nil),
				createTestTask("T002", "Second", "Desc", []string{"T001"}),
				createTestTask("T003", "Third", "Desc", []string{"T002"}),
			},
			expectedOrder: []string{"T001", "T002", "T003"},
		},
		{
			name: "diamond pattern",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", nil),
				createTestTask("T002", "Second", "Desc", []string{"T001"}),
				createTestTask("T003", "Third", "Desc", []string{"T001"}),
				createTestTask("T004", "Fourth", "Desc", []string{"T002", "T003"}),
			},
			expectedOrder: []string{"T001", "T002", "T003", "T004"},
		},
		{
			name: "multiple independent chains",
			tasks: []Task{
				createTestTask("T001", "First", "Desc", nil),
				createTestTask("T002", "Second", "Desc", nil),
				createTestTask("T003", "Third", "Desc", []string{"T001"}),
				createTestTask("T004", "Fourth", "Desc", []string{"T002"}),
			},
			expectedOrder: []string{"T001", "T002", "T003", "T004"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			graph := NewDependencyGraph(tt.tasks)
			order, err := graph.TopologicalSort()

			require.NoError(t, err)
			assert.Equal(t, tt.expectedOrder, order)
		})
	}
}

func TestDependencyGraph_TopologicalSort_CycleError(t *testing.T) {
	t.Parallel()

	tasks := []Task{
		createTestTask("T001", "First", "Desc", []string{"T002"}),
		createTestTask("T002", "Second", "Desc", []string{"T001"}),
	}

	graph := NewDependencyGraph(tasks)
	_, err := graph.TopologicalSort()

	require.Error(t, err)

	var ge *GraphError
	require.ErrorAs(t, err, &ge)
	assert.Equal(t, "cycle", ge.Type)
}

func TestDependencyGraph_TopologicalSort_EmptyGraph(t *testing.T) {
	t.Parallel()

	graph := NewDependencyGraph([]Task{})
	order, err := graph.TopologicalSort()

	require.NoError(t, err)
	assert.Empty(t, order)
}

func TestDependencyGraph_TopologicalSort_DeterministicOrder(t *testing.T) {
	t.Parallel()

	// Run multiple times to verify determinism
	tasks := []Task{
		createTestTask("T003", "Third", "Desc", nil),
		createTestTask("T001", "First", "Desc", nil),
		createTestTask("T002", "Second", "Desc", nil),
	}

	var firstResult []string
	for i := 0; i < 10; i++ {
		graph := NewDependencyGraph(tasks)
		order, err := graph.TopologicalSort()
		require.NoError(t, err)

		if firstResult == nil {
			firstResult = order
		} else {
			assert.Equal(t, firstResult, order, "topological sort should be deterministic")
		}
	}

	// All independent nodes should be sorted alphabetically
	assert.Equal(t, []string{"T001", "T002", "T003"}, firstResult)
}

func TestDependencyGraph_GetTask(t *testing.T) {
	t.Parallel()

	tasks := []Task{
		createTestTask("T001", "First", "Desc", nil),
	}

	graph := NewDependencyGraph(tasks)

	task, exists := graph.GetTask("T001")
	assert.True(t, exists)
	assert.Equal(t, "T001", task.ID)

	_, exists = graph.GetTask("T999")
	assert.False(t, exists)
}

func TestDependencyGraph_GetDependents(t *testing.T) {
	t.Parallel()

	tasks := []Task{
		createTestTask("T001", "First", "Desc", nil),
		createTestTask("T002", "Second", "Desc", []string{"T001"}),
		createTestTask("T003", "Third", "Desc", []string{"T001"}),
	}

	graph := NewDependencyGraph(tasks)

	dependents := graph.GetDependents("T001")
	assert.ElementsMatch(t, []string{"T002", "T003"}, dependents)

	dependents = graph.GetDependents("T002")
	assert.Empty(t, dependents)

	dependents = graph.GetDependents("T999")
	assert.Nil(t, dependents)
}

func TestDependencyGraph_Size(t *testing.T) {
	t.Parallel()

	tasks := []Task{
		createTestTask("T001", "First", "Desc", nil),
		createTestTask("T002", "Second", "Desc", nil),
		createTestTask("T003", "Third", "Desc", nil),
	}

	graph := NewDependencyGraph(tasks)
	assert.Equal(t, 3, graph.Size())

	emptyGraph := NewDependencyGraph([]Task{})
	assert.Equal(t, 0, emptyGraph.Size())
}

func TestDependencyGraph_HandlesNonExistentDependency(t *testing.T) {
	t.Parallel()

	// Task depends on non-existent task (validator should catch this,
	// but graph should handle gracefully)
	tasks := []Task{
		createTestTask("T001", "First", "Desc", []string{"T999"}),
	}

	graph := NewDependencyGraph(tasks)

	// Should not detect cycle
	cycle := graph.DetectCycle()
	assert.Nil(t, cycle)

	// Should still produce valid order
	order, err := graph.TopologicalSort()
	require.NoError(t, err)
	assert.Equal(t, []string{"T001"}, order)
}
