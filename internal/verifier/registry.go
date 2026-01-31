package verifier

import (
	"sync"
)

// CheckerRegistry manages available checkers.
type CheckerRegistry struct {
	mu       sync.RWMutex
	checkers []Checker
}

// NewCheckerRegistry creates a registry with default checkers.
func NewCheckerRegistry(executor CommandExecutor) *CheckerRegistry {
	if executor == nil {
		executor = &DefaultExecutor{}
	}

	return &CheckerRegistry{
		checkers: []Checker{
			NewFileExistsChecker(),
			NewBuildChecker(executor),
			NewTestChecker(executor),
			NewContentMatchChecker(),
		},
	}
}

// NewEmptyRegistry creates an empty registry without default checkers.
func NewEmptyRegistry() *CheckerRegistry {
	return &CheckerRegistry{
		checkers: []Checker{},
	}
}

// Register adds a checker to the registry.
func (r *CheckerRegistry) Register(checker Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers = append(r.checkers, checker)
}

// FindChecker returns the first checker that can handle the criterion.
// Returns nil if no checker can handle it.
func (r *CheckerRegistry) FindChecker(criterion string) Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, checker := range r.checkers {
		if checker.CanHandle(criterion) {
			return checker
		}
	}
	return nil
}

// AllCheckers returns a copy of all registered checkers.
func (r *CheckerRegistry) AllCheckers() []Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Checker, len(r.checkers))
	copy(result, r.checkers)
	return result
}

// Count returns the number of registered checkers.
func (r *CheckerRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.checkers)
}
