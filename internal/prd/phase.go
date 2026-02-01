package prd

import (
	"context"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// Phase implements planner.PlanningPhase for PRD generation.
type Phase struct {
	generator Generator
}

// NewPhase creates a new PRD phase with the given generator.
// Returns nil if generator is nil.
func NewPhase(generator Generator) *Phase {
	if generator == nil {
		return nil
	}
	return &Phase{generator: generator}
}

// Name returns the phase name: "prd".
func (p *Phase) Name() string {
	return "prd"
}

// ShouldRun determines if PRD phase should execute.
// Returns false if PRD is already completed.
func (p *Phase) ShouldRun(config *planner.Config, plan *planner.Plan) bool {
	if plan == nil {
		return false
	}
	return !plan.IsPhaseCompleted("prd")
}

// Run executes the PRD generation phase.
// It generates a PRD from the user prompt and updates the plan.
func (p *Phase) Run(ctx context.Context, plan *planner.Plan) error {
	if plan == nil {
		return &PRDError{
			Phase:   "run",
			Message: "plan is nil",
		}
	}

	if plan.UserPrompt == "" {
		return &PRDError{
			Phase:   "run",
			Message: "user prompt is empty",
		}
	}

	// Generate PRD using the generator
	generatedPRD, err := p.generator.Generate(ctx, plan.UserPrompt)
	if err != nil {
		return &PRDError{
			Phase:   "run",
			Message: "PRD generation failed",
			Err:     err,
		}
	}

	// Copy generated PRD to plan (embedded planner.PRD)
	plan.PRD = &generatedPRD.PRD

	// Track cost
	plan.AddCost(generatedPRD.Cost)

	return nil
}

// Compile-time interface compliance check.
var _ planner.PlanningPhase = (*Phase)(nil)
