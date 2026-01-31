package decomposer

import "sort"

// DependencyGraph represents a directed graph of task dependencies.
type DependencyGraph struct {
	nodes    map[string]*graphNode
	taskByID map[string]Task
}

type graphNode struct {
	id           string
	dependencies []string // Tasks this node depends on (incoming edges)
	dependents   []string // Tasks that depend on this node (outgoing edges)
	inDegree     int      // Number of unresolved dependencies
}

// NewDependencyGraph creates a graph from tasks.
func NewDependencyGraph(tasks []Task) *DependencyGraph {
	g := &DependencyGraph{
		nodes:    make(map[string]*graphNode),
		taskByID: make(map[string]Task),
	}

	// First pass: create all nodes
	for _, task := range tasks {
		g.taskByID[task.ID] = task
		g.nodes[task.ID] = &graphNode{
			id:           task.ID,
			dependencies: task.Dependencies,
			dependents:   []string{},
			inDegree:     0, // Will be calculated in second pass
		}
	}

	// Second pass: calculate in-degree (only count existing dependencies)
	// and build reverse edges (dependents)
	for _, task := range tasks {
		validDeps := 0
		for _, dep := range task.Dependencies {
			if node, exists := g.nodes[dep]; exists {
				node.dependents = append(node.dependents, task.ID)
				validDeps++
			}
			// Non-existent dependencies are ignored (validator should catch these)
		}
		g.nodes[task.ID].inDegree = validDeps
	}

	return g
}

// DetectCycle detects circular dependencies using DFS.
// Returns the cycle path if found, nil otherwise.
func (g *DependencyGraph) DetectCycle() []string {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var dfs func(id string, path []string) []string
	dfs = func(id string, path []string) []string {
		visited[id] = true
		inStack[id] = true
		path = append(path, id)

		node := g.nodes[id]
		for _, dep := range node.dependencies {
			if _, exists := g.nodes[dep]; !exists {
				continue // Skip non-existent dependencies (handled by validator)
			}
			if !visited[dep] {
				if cycle := dfs(dep, path); cycle != nil {
					return cycle
				}
			} else if inStack[dep] {
				// Found cycle - extract the cycle from path
				cycleStart := -1
				for i, p := range path {
					if p == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := append([]string{}, path[cycleStart:]...)
					cycle = append(cycle, dep)
					return cycle
				}
			}
		}

		inStack[id] = false
		return nil
	}

	// Get sorted list of node IDs for deterministic order
	var nodeIDs []string
	for id := range g.nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)

	for _, id := range nodeIDs {
		if !visited[id] {
			if cycle := dfs(id, []string{}); cycle != nil {
				return cycle
			}
		}
	}

	return nil
}

// TopologicalSort returns tasks in dependency order using Kahn's algorithm.
// Tasks with no dependencies come first.
// Returns error if a cycle is detected.
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	// First check for cycles
	if cycle := g.DetectCycle(); cycle != nil {
		return nil, &GraphError{
			Type:    "cycle",
			TaskIDs: cycle,
			Message: "circular dependency detected",
		}
	}

	// Clone inDegree counts
	inDegree := make(map[string]int)
	for id, node := range g.nodes {
		inDegree[id] = node.inDegree
	}

	// Initialize queue with nodes having no dependencies (in-degree 0)
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// Sort queue for deterministic output
	sort.Strings(queue)

	var result []string

	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// For each dependent, reduce in-degree
		node := g.nodes[current]
		var newReady []string
		for _, dependent := range node.dependents {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				newReady = append(newReady, dependent)
			}
		}

		// Sort and append for deterministic order
		sort.Strings(newReady)
		queue = append(queue, newReady...)
	}

	// Check if all nodes were processed (shouldn't happen if cycle check passed)
	if len(result) != len(g.nodes) {
		return nil, &GraphError{
			Type:    "incomplete",
			Message: "topological sort incomplete, possible undetected cycle",
		}
	}

	return result, nil
}

// GetTask returns a task by ID.
func (g *DependencyGraph) GetTask(id string) (Task, bool) {
	task, exists := g.taskByID[id]
	return task, exists
}

// GetDependents returns tasks that depend on the given task.
func (g *DependencyGraph) GetDependents(id string) []string {
	if node, exists := g.nodes[id]; exists {
		return node.dependents
	}
	return nil
}

// Size returns the number of nodes in the graph.
func (g *DependencyGraph) Size() int {
	return len(g.nodes)
}
