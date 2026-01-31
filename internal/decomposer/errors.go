package decomposer

import (
	"errors"
	"fmt"
)

// DecomposerError represents an error during decomposer operations.
type DecomposerError struct {
	Phase   string // "config", "generate", "parse", "validate", "graph", "schedule"
	Message string
	Err     error
}

func (e *DecomposerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("decomposer %s: %s: %v", e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("decomposer %s: %s", e.Phase, e.Message)
}

func (e *DecomposerError) Unwrap() error {
	return e.Err
}

// IsDecomposerError checks if an error is a DecomposerError.
func IsDecomposerError(err error) bool {
	var de *DecomposerError
	return errors.As(err, &de)
}

// ValidationError represents a specific validation failure.
type ValidationError struct {
	Field   string
	TaskID  string // Optional: specific task that failed
	Message string
}

func (e *ValidationError) Error() string {
	if e.TaskID != "" {
		return fmt.Sprintf("validation error on %s (task %s): %s", e.Field, e.TaskID, e.Message)
	}
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// IsValidationError checks if an error is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// GraphError represents dependency graph errors.
type GraphError struct {
	Type    string   // "cycle", "missing_dependency", "self_reference"
	TaskIDs []string // Tasks involved in the error
	Message string
}

func (e *GraphError) Error() string {
	return fmt.Sprintf("graph error (%s): %s [tasks: %v]", e.Type, e.Message, e.TaskIDs)
}

// IsGraphError checks if an error is a GraphError.
func IsGraphError(err error) bool {
	var ge *GraphError
	return errors.As(err, &ge)
}

// Predefined errors.
var (
	ErrNilClient        = &DecomposerError{Phase: "config", Message: "claude client is nil"}
	ErrNilArchitecture  = &DecomposerError{Phase: "generate", Message: "architecture is nil"}
	ErrParseNoTasks     = &DecomposerError{Phase: "parse", Message: "no tasks found in output"}
	ErrCyclicDependency = &DecomposerError{Phase: "graph", Message: "cyclic dependency detected"}
)
