// Package verifier provides task completion verification capabilities.
// It verifies success criteria through file existence checks, build validation,
// test execution, and optional AI-based semantic verification.
package verifier

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

// Verifier verifies task completion against success criteria.
type Verifier interface {
	Verify(ctx context.Context, task *VerificationTask) (*VerificationResult, error)
}

// VerificationTask contains the task and criteria to verify.
type VerificationTask struct {
	TaskID          string
	Title           string
	Description     string
	SuccessCriteria []string
	Files           []string // Files associated with the task
	WorkDir         string   // Working directory for commands
}

// VerificationResult represents the outcome of verification.
type VerificationResult struct {
	TaskID    string
	Passed    bool
	Checks    []CheckResult
	Cost      float64       // AI verification cost (if any)
	Duration  time.Duration
	Timestamp time.Time
}

// AllPassed returns true if all checks passed.
func (r *VerificationResult) AllPassed() bool {
	for _, check := range r.Checks {
		if !check.Passed {
			return false
		}
	}
	return len(r.Checks) > 0
}

// FailedChecks returns only the checks that failed.
func (r *VerificationResult) FailedChecks() []CheckResult {
	var failed []CheckResult
	for _, check := range r.Checks {
		if !check.Passed {
			failed = append(failed, check)
		}
	}
	return failed
}

// CheckResult represents the outcome of a single criterion check.
type CheckResult struct {
	Criterion   string        // Original success criterion string
	CheckerType string        // Type of checker used (file_exists, build, test, content, ai)
	Passed      bool
	Evidence    *Evidence     // Evidence of verification (output, screenshot, etc.)
	Duration    time.Duration
	Error       string // Error message if check failed
}

// Evidence contains proof of verification outcome.
type Evidence struct {
	Type        EvidenceType
	Content     string    // Actual output or content
	Expected    string    // Expected value (if applicable)
	Timestamp   time.Time
	CommandRun  string // Command that was executed (if applicable)
	ExitCode    int    // Exit code (if command was run)
}

// EvidenceType classifies the type of evidence collected.
type EvidenceType string

const (
	EvidenceTypeCommandOutput EvidenceType = "command_output"
	EvidenceTypeFileContent   EvidenceType = "file_content"
	EvidenceTypeFileExists    EvidenceType = "file_exists"
	EvidenceTypeAIAnalysis    EvidenceType = "ai_analysis"
)

// VerificationLevel defines the depth of verification.
type VerificationLevel string

const (
	VerificationLevelBasic    VerificationLevel = "basic"    // File existence only
	VerificationLevelStandard VerificationLevel = "standard" // Build + Lint
	VerificationLevelStrict   VerificationLevel = "strict"   // Build + Lint + Test
)

// Config holds verifier configuration.
type Config struct {
	Level         VerificationLevel
	MaxRetries    int           // Max retries for flaky tests (default: 1)
	Timeout       time.Duration // Per-check timeout (default: 5m)
	EnableAI      bool          // Enable AI-based verification (default: false)
	WorkDir       string        // Working directory for commands
	CaptureOutput bool          // Capture command output as evidence (default: true)
}

// DefaultConfig returns Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Level:         VerificationLevelStandard,
		MaxRetries:    1,
		Timeout:       5 * time.Minute,
		EnableAI:      false,
		CaptureOutput: true,
	}
}

// IsEnabled returns true if verifier config is valid.
func (c *Config) IsEnabled() bool {
	return c != nil
}
