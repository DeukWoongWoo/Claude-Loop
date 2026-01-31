// Package prd provides PRD (Product Requirements Document) generation capabilities.
// It parses Claude's output into structured PRD data and validates the result.
package prd

import (
	"errors"
	"fmt"
)

// PRDError represents an error during PRD operations.
type PRDError struct {
	Phase   string // "config", "generate", "parse", "validate"
	Message string
	Err     error
}

func (e *PRDError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("prd %s: %s: %v", e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("prd %s: %s", e.Phase, e.Message)
}

func (e *PRDError) Unwrap() error {
	return e.Err
}

// IsPRDError checks if an error is a PRDError.
func IsPRDError(err error) bool {
	var pe *PRDError
	return errors.As(err, &pe)
}

// Predefined errors.
var (
	ErrNilClient           = &PRDError{Phase: "config", Message: "claude client is nil"}
	ErrEmptyPrompt         = &PRDError{Phase: "generate", Message: "user prompt is empty"}
	ErrParseNoGoals        = &PRDError{Phase: "parse", Message: "no goals found in output"}
	ErrParseNoRequirements = &PRDError{Phase: "parse", Message: "no requirements found in output"}
)

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
