package planner

import (
	"context"
	"time"
)

// PlanningPhase defines the interface for each planning step.
// PRD, Architecture, and Decomposition phases implement this interface.
type PlanningPhase interface {
	// Name returns the unique name of this phase (e.g., "prd", "architecture", "tasks").
	Name() string

	// ShouldRun determines if this phase should execute given current state.
	// Returns false if phase is already completed or conditions not met.
	ShouldRun(config *Config, plan *Plan) bool

	// Run executes the phase and updates the plan in-place.
	// Returns error if phase fails (plan may be partially updated).
	Run(ctx context.Context, plan *Plan) error
}

// PhaseResult contains the outcome of a phase execution.
type PhaseResult struct {
	PhaseName string        // Name of the executed phase
	Skipped   bool          // True if phase was skipped (already completed)
	Cost      float64       // Cost in USD for this phase execution
	Duration  time.Duration // How long this phase took
	Error     error         // Error if phase failed
}

// ValidPhaseNames lists all valid phase names in execution order.
var ValidPhaseNames = []string{"prd", "architecture", "tasks"}

// IsValidPhaseName checks if a phase name is valid.
func IsValidPhaseName(name string) bool {
	for _, valid := range ValidPhaseNames {
		if valid == name {
			return true
		}
	}
	return false
}

// GetPhaseIndex returns the index of a phase in the execution order.
// Returns -1 if not found.
func GetPhaseIndex(name string) int {
	for i, valid := range ValidPhaseNames {
		if valid == name {
			return i
		}
	}
	return -1
}
