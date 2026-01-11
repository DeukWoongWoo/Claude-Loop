package cli

import (
	"errors"
	"fmt"
)

// ValidationError represents a CLI validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// validatePrompt checks if prompt is provided when required.
func (f *Flags) validatePrompt() *ValidationError {
	if f.Prompt == "" {
		return &ValidationError{
			Field:   "prompt",
			Message: "prompt is required: use -p or --prompt",
		}
	}
	return nil
}

// validateLimit checks if at least one execution limit is provided.
func (f *Flags) validateLimit() *ValidationError {
	hasLimit := f.MaxRuns > 0 || f.MaxCost > 0 || f.MaxDuration > 0
	if !hasLimit {
		return &ValidationError{
			Field:   "limit",
			Message: "at least one limit required: use -m/--max-runs, --max-cost, or --max-duration",
		}
	}
	return nil
}

// validateMergeStrategy checks if merge strategy is valid.
func (f *Flags) validateMergeStrategy() *ValidationError {
	if f.MergeStrategy == "" {
		return nil
	}

	switch f.MergeStrategy {
	case "squash", "merge", "rebase":
		return nil
	default:
		return &ValidationError{
			Field:   "merge-strategy",
			Message: fmt.Sprintf("merge-strategy must be squash, merge, or rebase (got %q)", f.MergeStrategy),
		}
	}
}

// validateNonNegative checks that numeric values are not negative.
func (f *Flags) validateNonNegative() *ValidationError {
	if f.MaxRuns < 0 {
		return &ValidationError{
			Field:   "max-runs",
			Message: "max-runs cannot be negative",
		}
	}
	if f.MaxCost < 0 {
		return &ValidationError{
			Field:   "max-cost",
			Message: "max-cost cannot be negative",
		}
	}
	if f.MaxDuration < 0 {
		return &ValidationError{
			Field:   "max-duration",
			Message: "max-duration cannot be negative",
		}
	}
	if f.CIRetryMax < 0 {
		return &ValidationError{
			Field:   "ci-retry-max",
			Message: "ci-retry-max cannot be negative",
		}
	}
	if f.CompletionThreshold < 0 {
		return &ValidationError{
			Field:   "completion-threshold",
			Message: "completion-threshold cannot be negative",
		}
	}
	return nil
}

// Validate checks the Flags for validity according to CLI_CONTRACT.md rules.
// Returns nil if valid, or the first error encountered.
func (f *Flags) Validate() error {
	if f.ListWorktrees {
		return nil
	}

	if err := f.validatePrompt(); err != nil {
		return err
	}
	if err := f.validateNonNegative(); err != nil {
		return err
	}
	if err := f.validateLimit(); err != nil {
		return err
	}
	if err := f.validateMergeStrategy(); err != nil {
		return err
	}

	return nil
}

// ValidateAll runs all validation checks and returns all errors found.
func (f *Flags) ValidateAll() []error {
	if f.ListWorktrees {
		return nil
	}

	var errs []error

	if err := f.validatePrompt(); err != nil {
		errs = append(errs, err)
	}
	if err := f.validateNonNegative(); err != nil {
		errs = append(errs, err)
	}
	if err := f.validateLimit(); err != nil {
		errs = append(errs, err)
	}
	if err := f.validateMergeStrategy(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// IsValidationError checks if an error is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
