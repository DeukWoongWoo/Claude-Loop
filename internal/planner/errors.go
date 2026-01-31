// Package planner provides planning infrastructure for multi-phase task execution.
package planner

import (
	"errors"
	"fmt"
)

// PlannerError represents an error during planner operation.
type PlannerError struct {
	Phase   string // "config", "persistence", "phase", "run"
	Message string
	Err     error
}

func (e *PlannerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("planner %s: %s: %v", e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("planner %s: %s", e.Phase, e.Message)
}

func (e *PlannerError) Unwrap() error {
	return e.Err
}

// IsPlannerError checks if an error is a PlannerError.
func IsPlannerError(err error) bool {
	var pe *PlannerError
	return errors.As(err, &pe)
}

// Predefined errors.
var (
	ErrPlanNotFound     = &PlannerError{Phase: "persistence", Message: "plan not found"}
	ErrInvalidPlanID    = &PlannerError{Phase: "persistence", Message: "invalid plan ID"}
	ErrInvalidPlanState = &PlannerError{Phase: "run", Message: "invalid plan state"}
	ErrPhaseNotFound    = &PlannerError{Phase: "phase", Message: "phase not found"}
	ErrPhaseFailed      = &PlannerError{Phase: "phase", Message: "phase execution failed"}
)
