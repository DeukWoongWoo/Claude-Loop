package planner

import (
	"context"
	"time"
)

// ClaudeClient executes Claude Code iterations.
// Defined here to avoid import cycles with the loop package.
type ClaudeClient interface {
	Execute(ctx context.Context, prompt string) (*IterationResult, error)
}

// IterationResult mirrors loop.IterationResult to avoid import cycles.
type IterationResult struct {
	Output                string
	Cost                  float64
	Duration              time.Duration
	CompletionSignalFound bool
}

// PlanStatus represents the current state of a plan.
type PlanStatus string

const (
	PlanStatusPending    PlanStatus = "pending"
	PlanStatusInProgress PlanStatus = "in_progress"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusFailed     PlanStatus = "failed"
	PlanStatusCancelled  PlanStatus = "cancelled"
)

// IsTerminal returns true if the status is a terminal state.
func (s PlanStatus) IsTerminal() bool {
	return s == PlanStatusCompleted || s == PlanStatusFailed || s == PlanStatusCancelled
}

// PRD represents the Product Requirements Document output.
type PRD struct {
	Goals           []string `yaml:"goals"`
	Requirements    []string `yaml:"requirements"`
	Constraints     []string `yaml:"constraints"`
	SuccessCriteria []string `yaml:"success_criteria"`
	RawOutput       string   `yaml:"raw_output"`
}

// Architecture represents the architecture design output.
type Architecture struct {
	Components    []Component `yaml:"components"`
	Dependencies  []string    `yaml:"dependencies"`
	FileStructure []string    `yaml:"file_structure"`
	TechDecisions []string    `yaml:"tech_decisions"`
	RawOutput     string      `yaml:"raw_output"`
}

// Component represents an architectural component.
type Component struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Files       []string `yaml:"files"`
}

// TaskGraph represents the task decomposition output.
type TaskGraph struct {
	Tasks          []Task   `yaml:"tasks"`
	ExecutionOrder []string `yaml:"execution_order"` // Task IDs in order
	RawOutput      string   `yaml:"raw_output"`
}

// Task represents a single decomposed task.
type Task struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	Dependencies []string `yaml:"dependencies"` // Other task IDs
	Status       string   `yaml:"status"`       // pending, in_progress, completed
	Files        []string `yaml:"files"`        // Files to modify
}

// TaskStatus constants.
const (
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// Plan is the main plan structure that persists across sessions.
type Plan struct {
	ID        string     `yaml:"id"`
	UserPrompt string    `yaml:"user_prompt"`
	CreatedAt  time.Time `yaml:"created_at"`
	UpdatedAt  time.Time `yaml:"updated_at"`
	Status     PlanStatus `yaml:"status"`

	// Phase outputs
	PRD          *PRD          `yaml:"prd,omitempty"`
	Architecture *Architecture `yaml:"architecture,omitempty"`
	TaskGraph    *TaskGraph    `yaml:"task_graph,omitempty"`

	// Execution state
	CurrentTaskID   string   `yaml:"current_task_id,omitempty"`
	CompletedPhases []string `yaml:"completed_phases"`

	// Cost tracking
	TotalCost float64 `yaml:"total_cost"`
}

// NewPlan creates a new Plan with initialized fields.
func NewPlan(id, userPrompt string) *Plan {
	now := time.Now()
	return &Plan{
		ID:              id,
		UserPrompt:      userPrompt,
		CreatedAt:       now,
		UpdatedAt:       now,
		Status:          PlanStatusPending,
		CompletedPhases: make([]string, 0),
	}
}

// IsPhaseCompleted checks if a phase is in CompletedPhases.
func (p *Plan) IsPhaseCompleted(phaseName string) bool {
	for _, completed := range p.CompletedPhases {
		if completed == phaseName {
			return true
		}
	}
	return false
}

// MarkPhaseCompleted adds a phase to CompletedPhases if not already present.
func (p *Plan) MarkPhaseCompleted(phaseName string) {
	if !p.IsPhaseCompleted(phaseName) {
		p.CompletedPhases = append(p.CompletedPhases, phaseName)
		p.UpdatedAt = time.Now()
	}
}

// AddCost adds to the total cost and updates the timestamp.
func (p *Plan) AddCost(cost float64) {
	p.TotalCost += cost
	p.UpdatedAt = time.Now()
}

// Config holds planner configuration.
type Config struct {
	PlanDir         string                     // Directory for plan files (default: .claude/plans)
	MaxPhaseRetries int                        // Max retries per phase (default: 3) - reserved for future use
	OnProgress      func(phase, status string) // Progress callback
}

// DefaultConfig returns Config with default values.
func DefaultConfig() *Config {
	return &Config{
		PlanDir:         ".claude/plans",
		MaxPhaseRetries: 3,
	}
}

// IsEnabled always returns true for planner config.
// Planner is enabled when a plan exists or is being created.
func (c *Config) IsEnabled() bool {
	return c != nil
}
