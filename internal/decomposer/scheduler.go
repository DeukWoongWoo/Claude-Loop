package decomposer

// Scheduler determines execution order for tasks.
type Scheduler interface {
	Schedule(tasks []Task) ([]string, error)
}

// DefaultScheduler implements Scheduler using topological sort.
type DefaultScheduler struct{}

// NewScheduler creates a new DefaultScheduler.
func NewScheduler() *DefaultScheduler {
	return &DefaultScheduler{}
}

// Schedule returns task IDs in execution order.
func (s *DefaultScheduler) Schedule(tasks []Task) ([]string, error) {
	if len(tasks) == 0 {
		return []string{}, nil
	}

	graph := NewDependencyGraph(tasks)
	return graph.TopologicalSort()
}

// Compile-time interface compliance check.
var _ Scheduler = (*DefaultScheduler)(nil)
