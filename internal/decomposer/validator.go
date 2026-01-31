package decomposer

import (
	"fmt"
	"regexp"
)

// Task ID pattern: T followed by exactly 3 digits.
var taskIDPattern = regexp.MustCompile(`^T\d{3}$`)

// Validator validates Task structures.
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks tasks for completeness and validity.
// Returns nil if valid, or the first error encountered.
func (v *Validator) Validate(tasks []Task) error {
	if len(tasks) == 0 {
		return &ValidationError{Field: "tasks", Message: "at least one task is required"}
	}

	if err := v.validateTaskIDs(tasks); err != nil {
		return err
	}

	if err := v.validateUniqueness(tasks); err != nil {
		return err
	}

	return v.validateDependencyReferences(tasks)
}

// ValidateAll returns all validation errors found.
func (v *Validator) ValidateAll(tasks []Task) []error {
	if len(tasks) == 0 {
		return []error{&ValidationError{Field: "tasks", Message: "at least one task is required"}}
	}

	var errs []error

	errs = append(errs, v.validateAllTaskIDs(tasks)...)
	errs = append(errs, v.validateAllUniqueness(tasks)...)
	errs = append(errs, v.validateAllDependencyReferences(tasks)...)
	errs = append(errs, v.validateTaskCompleteness(tasks)...)

	return errs
}

// validateTaskIDs checks that all task IDs match the T### pattern.
func (v *Validator) validateTaskIDs(tasks []Task) error {
	if errs := v.validateAllTaskIDs(tasks); len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// validateAllTaskIDs returns all task ID validation errors.
func (v *Validator) validateAllTaskIDs(tasks []Task) []error {
	var errs []error
	for _, task := range tasks {
		if !taskIDPattern.MatchString(task.ID) {
			errs = append(errs, &ValidationError{
				Field:   "id",
				TaskID:  task.ID,
				Message: "task ID must match pattern T### (e.g., T001)",
			})
		}
	}
	return errs
}

// validateUniqueness checks for duplicate task IDs.
func (v *Validator) validateUniqueness(tasks []Task) error {
	seen := make(map[string]bool)
	for _, task := range tasks {
		if seen[task.ID] {
			return &ValidationError{
				Field:   "id",
				TaskID:  task.ID,
				Message: "duplicate task ID",
			}
		}
		seen[task.ID] = true
	}
	return nil
}

// validateAllUniqueness returns all uniqueness errors.
func (v *Validator) validateAllUniqueness(tasks []Task) []error {
	var errs []error
	counts := make(map[string]int)
	for _, task := range tasks {
		counts[task.ID]++
	}
	for id, count := range counts {
		if count > 1 {
			errs = append(errs, &ValidationError{
				Field:   "id",
				TaskID:  id,
				Message: fmt.Sprintf("duplicate task ID (appears %d times)", count),
			})
		}
	}
	return errs
}

// validateDependencyReferences checks that all dependencies reference existing tasks.
func (v *Validator) validateDependencyReferences(tasks []Task) error {
	if errs := v.validateAllDependencyReferences(tasks); len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// validateAllDependencyReferences returns all dependency reference errors.
func (v *Validator) validateAllDependencyReferences(tasks []Task) []error {
	var errs []error
	taskIDs := make(map[string]bool)
	for _, task := range tasks {
		taskIDs[task.ID] = true
	}

	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if !taskIDs[dep] {
				errs = append(errs, &ValidationError{
					Field:   "dependencies",
					TaskID:  task.ID,
					Message: fmt.Sprintf("references non-existent task %s", dep),
				})
			}
			if dep == task.ID {
				errs = append(errs, &ValidationError{
					Field:   "dependencies",
					TaskID:  task.ID,
					Message: "task cannot depend on itself",
				})
			}
		}
	}
	return errs
}

// validateTaskCompleteness checks that each task has required fields.
func (v *Validator) validateTaskCompleteness(tasks []Task) []error {
	var errs []error
	for _, task := range tasks {
		if task.Title == "" {
			errs = append(errs, &ValidationError{
				Field:   "title",
				TaskID:  task.ID,
				Message: "task title is required",
			})
		}
		if task.Description == "" {
			errs = append(errs, &ValidationError{
				Field:   "description",
				TaskID:  task.ID,
				Message: "task description is required",
			})
		}
	}
	return errs
}
