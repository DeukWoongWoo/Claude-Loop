package decomposer

import (
	"context"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// Phase implements planner.PlanningPhase for task decomposition.
type Phase struct {
	decomposer Decomposer
}

// NewPhase creates a new Decomposer phase with the given decomposer.
// Returns nil if decomposer is nil.
func NewPhase(decomposer Decomposer) *Phase {
	if decomposer == nil {
		return nil
	}
	return &Phase{decomposer: decomposer}
}

// Name returns the phase name: "tasks".
func (p *Phase) Name() string {
	return "tasks"
}

// ShouldRun determines if Decomposer phase should execute.
// Returns false if TaskGraph is already completed or if Architecture is missing.
func (p *Phase) ShouldRun(config *planner.Config, plan *planner.Plan) bool {
	if plan == nil {
		return false
	}
	// Can't run without Architecture
	if plan.Architecture == nil {
		return false
	}
	return !plan.IsPhaseCompleted("tasks")
}

// Run executes the task decomposition phase.
// It generates a TaskGraph from the plan's Architecture and updates the plan.
func (p *Phase) Run(ctx context.Context, plan *planner.Plan) error {
	if plan == nil {
		return &DecomposerError{
			Phase:   "run",
			Message: "plan is nil",
		}
	}

	if plan.Architecture == nil {
		return &DecomposerError{
			Phase:   "run",
			Message: "Architecture is nil - tasks phase requires Architecture to be completed first",
		}
	}

	// Generate TaskGraph using the decomposer
	generatedGraph, err := p.decomposer.Decompose(ctx, plan.Architecture)
	if err != nil {
		return &DecomposerError{
			Phase:   "run",
			Message: "Task decomposition failed",
			Err:     err,
		}
	}

	// Copy generated TaskGraph to plan (embedded planner.TaskGraph)
	plan.TaskGraph = &generatedGraph.TaskGraph

	// Track cost
	plan.AddCost(generatedGraph.Cost)

	return nil
}

// Compile-time interface compliance check.
var _ planner.PlanningPhase = (*Phase)(nil)
