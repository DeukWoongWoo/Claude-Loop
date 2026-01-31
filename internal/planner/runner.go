package planner

import (
	"context"
	"fmt"
	"time"
)

// PhaseRunner orchestrates the execution of planning phases.
type PhaseRunner struct {
	config      *Config
	phases      []PlanningPhase
	persistence Persistence
}

// NewPhaseRunner creates a new PhaseRunner.
// Phases are executed in the order provided.
func NewPhaseRunner(config *Config, persistence Persistence, phases ...PlanningPhase) *PhaseRunner {
	if config == nil {
		config = DefaultConfig()
	}
	if persistence == nil {
		persistence = NewFilePersistence(config.PlanDir)
	}
	return &PhaseRunner{
		config:      config,
		phases:      phases,
		persistence: persistence,
	}
}

// RunResult contains the outcome of running all phases.
type RunResult struct {
	Plan          *Plan          // The plan after execution
	PhaseResults  []PhaseResult  // Results for each phase
	TotalCost     float64        // Total cost across all phases
	TotalDuration time.Duration  // Total duration across all phases
	Error         error          // First error encountered (if any)
}

// Run executes all phases in order.
// - Skips already-completed phases
// - Saves plan after each phase
// - Stops on first error
func (r *PhaseRunner) Run(ctx context.Context, plan *Plan) (*RunResult, error) {
	if plan == nil {
		return nil, &PlannerError{
			Phase:   "run",
			Message: "plan is nil",
		}
	}

	result := &RunResult{
		Plan:         plan,
		PhaseResults: make([]PhaseResult, 0, len(r.phases)),
	}
	startTime := time.Now()

	for _, phase := range r.phases {
		// Check context cancellation
		select {
		case <-ctx.Done():
			plan.Status = PlanStatusCancelled
			plan.UpdatedAt = time.Now()
			_ = r.save(plan)
			result.TotalDuration = time.Since(startTime)
			result.TotalCost = plan.TotalCost
			result.Error = ctx.Err()
			r.progress(phase.Name(), "cancelled")
			return result, ctx.Err()
		default:
		}

		phaseResult := PhaseResult{PhaseName: phase.Name()}
		phaseStart := time.Now()

		// Check if phase should run
		if !phase.ShouldRun(r.config, plan) {
			phaseResult.Skipped = true
			result.PhaseResults = append(result.PhaseResults, phaseResult)
			r.progress(phase.Name(), "skipped")
			continue
		}

		// Report progress
		r.progress(phase.Name(), "running")

		// Execute phase
		plan.Status = PlanStatusInProgress
		err := phase.Run(ctx, plan)
		phaseResult.Duration = time.Since(phaseStart)

		if err != nil {
			phaseResult.Error = err
			plan.Status = PlanStatusFailed
			result.PhaseResults = append(result.PhaseResults, phaseResult)
			result.TotalDuration = time.Since(startTime)
			result.TotalCost = plan.TotalCost
			result.Error = &PlannerError{
				Phase:   "phase",
				Message: fmt.Sprintf("phase %s failed", phase.Name()),
				Err:     err,
			}

			// Save failed state
			_ = r.save(plan)
			r.progress(phase.Name(), "failed")
			return result, result.Error
		}

		// Mark phase completed and save
		plan.MarkPhaseCompleted(phase.Name())
		plan.UpdatedAt = time.Now()

		if saveErr := r.save(plan); saveErr != nil {
			result.TotalDuration = time.Since(startTime)
			result.TotalCost = plan.TotalCost
			result.Error = saveErr
			return result, saveErr
		}

		result.PhaseResults = append(result.PhaseResults, phaseResult)
		r.progress(phase.Name(), "completed")
	}

	plan.Status = PlanStatusCompleted
	plan.UpdatedAt = time.Now()
	_ = r.save(plan)

	result.TotalDuration = time.Since(startTime)
	result.TotalCost = plan.TotalCost
	return result, nil
}

// RunPhase executes a single phase by name.
func (r *PhaseRunner) RunPhase(ctx context.Context, plan *Plan, phaseName string) (*PhaseResult, error) {
	if plan == nil {
		return nil, &PlannerError{
			Phase:   "run",
			Message: "plan is nil",
		}
	}

	// Find the phase
	var targetPhase PlanningPhase
	for _, phase := range r.phases {
		if phase.Name() == phaseName {
			targetPhase = phase
			break
		}
	}

	if targetPhase == nil {
		return nil, &PlannerError{
			Phase:   "phase",
			Message: fmt.Sprintf("phase %s not found", phaseName),
		}
	}

	result := &PhaseResult{PhaseName: phaseName}
	phaseStart := time.Now()

	// Check if phase should run
	if !targetPhase.ShouldRun(r.config, plan) {
		result.Skipped = true
		return result, nil
	}

	// Execute phase
	plan.Status = PlanStatusInProgress
	err := targetPhase.Run(ctx, plan)
	result.Duration = time.Since(phaseStart)

	if err != nil {
		result.Error = err
		plan.Status = PlanStatusFailed
		_ = r.save(plan)
		return result, err
	}

	// Mark completed and save
	plan.MarkPhaseCompleted(phaseName)
	plan.UpdatedAt = time.Now()
	if saveErr := r.save(plan); saveErr != nil {
		return result, saveErr
	}

	return result, nil
}

// ResumePlan loads a plan and resumes from the last completed phase.
func (r *PhaseRunner) ResumePlan(ctx context.Context, planPath string) (*RunResult, error) {
	plan, err := r.persistence.Load(planPath)
	if err != nil {
		return nil, err
	}

	// Check if plan is in a terminal state
	if plan.Status.IsTerminal() && plan.Status != PlanStatusFailed {
		return &RunResult{
			Plan:         plan,
			PhaseResults: []PhaseResult{},
			TotalCost:    plan.TotalCost,
			Error:        nil,
		}, nil
	}

	// Reset status if it was failed (to allow retry)
	if plan.Status == PlanStatusFailed {
		plan.Status = PlanStatusInProgress
	}

	return r.Run(ctx, plan)
}

// save persists the plan using the default path.
func (r *PhaseRunner) save(plan *Plan) error {
	path := r.persistence.DefaultPlanPath(plan.ID)
	return r.persistence.Save(plan, path)
}

// progress reports progress if callback is set.
func (r *PhaseRunner) progress(phase, status string) {
	if r.config.OnProgress != nil {
		r.config.OnProgress(phase, status)
	}
}

// Config returns the runner's configuration.
func (r *PhaseRunner) Config() *Config {
	return r.config
}

// Phases returns the registered phases.
func (r *PhaseRunner) Phases() []PlanningPhase {
	return r.phases
}

// Persistence returns the persistence handler.
func (r *PhaseRunner) Persistence() Persistence {
	return r.persistence
}
