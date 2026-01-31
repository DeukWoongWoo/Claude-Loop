// Package architecture provides architecture design generation capabilities.
// It parses Claude's output into structured architecture data and validates the result.
package architecture

import (
	"errors"
	"fmt"
)

// ArchitectureError represents an error during architecture operations.
type ArchitectureError struct {
	Phase   string // "config", "generate", "parse", "validate"
	Message string
	Err     error
}

func (e *ArchitectureError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("architecture %s: %s: %v", e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("architecture %s: %s", e.Phase, e.Message)
}

func (e *ArchitectureError) Unwrap() error {
	return e.Err
}

// IsArchitectureError checks if an error is an ArchitectureError.
func IsArchitectureError(err error) bool {
	var ae *ArchitectureError
	return errors.As(err, &ae)
}

// ValidationError represents a specific validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// IsValidationError checks if an error is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// Predefined errors.
var (
	ErrNilClient            = &ArchitectureError{Phase: "config", Message: "claude client is nil"}
	ErrNilPRD               = &ArchitectureError{Phase: "generate", Message: "PRD is nil"}
	ErrParseNoComponents    = &ArchitectureError{Phase: "parse", Message: "no components found in output"}
	ErrParseNoFileStructure = &ArchitectureError{Phase: "parse", Message: "no file structure found in output"}
)
