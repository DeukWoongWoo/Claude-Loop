package architecture

import (
	"context"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// Phase implements planner.PlanningPhase for Architecture generation.
type Phase struct {
	generator Generator
}

// NewPhase creates a new Architecture phase with the given generator.
// Returns nil if generator is nil.
func NewPhase(generator Generator) *Phase {
	if generator == nil {
		return nil
	}
	return &Phase{generator: generator}
}

// Name returns the phase name: "architecture".
func (p *Phase) Name() string {
	return "architecture"
}

// ShouldRun determines if Architecture phase should execute.
// Returns false if Architecture is already completed or if PRD is missing.
func (p *Phase) ShouldRun(config *planner.Config, plan *planner.Plan) bool {
	if plan == nil {
		return false
	}
	// Can't run without PRD
	if plan.PRD == nil {
		return false
	}
	return !plan.IsPhaseCompleted("architecture")
}

// Run executes the Architecture generation phase.
// It generates an Architecture from the plan's PRD and updates the plan.
func (p *Phase) Run(ctx context.Context, plan *planner.Plan) error {
	if plan == nil {
		return &ArchitectureError{
			Phase:   "run",
			Message: "plan is nil",
		}
	}

	if plan.PRD == nil {
		return &ArchitectureError{
			Phase:   "run",
			Message: "PRD is nil - architecture phase requires PRD to be completed first",
		}
	}

	// Generate Architecture using the generator
	generatedArch, err := p.generator.Generate(ctx, plan.PRD)
	if err != nil {
		return &ArchitectureError{
			Phase:   "run",
			Message: "Architecture generation failed",
			Err:     err,
		}
	}

	// Copy generated Architecture to plan (embedded planner.Architecture)
	plan.Architecture = &generatedArch.Architecture

	// Track cost
	plan.AddCost(generatedArch.Cost)

	return nil
}

// Compile-time interface compliance check.
var _ planner.PlanningPhase = (*Phase)(nil)
