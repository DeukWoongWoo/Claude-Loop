package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/git"
	"github.com/DeukWoongWoo/claude-loop/internal/loop"
	"github.com/DeukWoongWoo/claude-loop/internal/prompt"
)

// CIFixConfig configures the CI fix behavior.
type CIFixConfig struct {
	// MaxRetries is the maximum number of fix attempts (from --ci-retry-max flag).
	MaxRetries int

	// DisableRetry disables CI fix attempts (from --disable-ci-retry flag).
	DisableRetry bool

	// PRNumber is the PR being fixed.
	PRNumber int

	// BranchName is the branch to push fixes to.
	BranchName string

	// WaitOptions configures check waiting behavior.
	WaitOptions *WaitOptions

	// OnProgress is called with status updates during fix attempts.
	OnProgress func(status string)

	// OnAttempt is called at the start of each fix attempt.
	OnAttempt func(attempt int, max int)
}

// DefaultCIFixConfig returns CIFixConfig with default values.
func DefaultCIFixConfig() *CIFixConfig {
	return &CIFixConfig{
		MaxRetries:  1,
		WaitOptions: DefaultWaitOptions(),
	}
}

// CIFixResult represents the outcome of CI fix attempts.
type CIFixResult struct {
	// Success indicates whether the fix was successful.
	Success bool

	// Attempts is the number of attempts made.
	Attempts int

	// FixedOnAttempt is which attempt succeeded (0 if failed).
	FixedOnAttempt int

	// LastError is the last error encountered.
	LastError error

	// FailureInfo contains info about the original failure.
	FailureInfo *CIFailureInfo
}

// CIFixManager orchestrates the CI failure fix workflow.
type CIFixManager struct {
	analyzer      *CIAnalyzer
	claudeClient  loop.ClaudeClient
	commitMgr     *git.CommitManager
	checkMonitor  *CheckMonitor
	promptBuilder *prompt.CIFixBuilder
	config        *CIFixConfig
	repo          *RepoInfo
}

// NewCIFixManager creates a new CIFixManager.
func NewCIFixManager(
	executor CommandExecutor,
	gitExecutor git.CommandExecutor,
	repo *RepoInfo,
	claudeClient loop.ClaudeClient,
	config *CIFixConfig,
) *CIFixManager {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	if config == nil {
		config = DefaultCIFixConfig()
	}
	return &CIFixManager{
		analyzer:      NewCIAnalyzer(executor, repo),
		claudeClient:  claudeClient,
		commitMgr:     git.NewCommitManager(gitExecutor),
		checkMonitor:  NewCheckMonitor(executor, repo),
		promptBuilder: prompt.NewCIFixBuilder(),
		config:        config,
		repo:          repo,
	}
}

// AttemptFix tries to fix CI failures for a PR.
// Returns after success or max retries exhausted.
func (m *CIFixManager) AttemptFix(ctx context.Context) (*CIFixResult, error) {
	if m.config.DisableRetry {
		return &CIFixResult{
			Success:   false,
			Attempts:  0,
			LastError: ErrCIFixDisabled,
		}, nil
	}

	result := &CIFixResult{}
	maxAttempts := m.config.MaxRetries
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.Attempts = attempt

		if m.config.OnAttempt != nil {
			m.config.OnAttempt(attempt, maxAttempts)
		}

		err := m.runSingleAttempt(ctx, attempt)
		if err == nil {
			result.Success = true
			result.FixedOnAttempt = attempt
			return result, nil
		}

		result.LastError = err

		// Check if context was cancelled
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		// Check if we should continue
		if attempt < maxAttempts {
			if m.config.OnProgress != nil {
				m.config.OnProgress(fmt.Sprintf("Attempt %d failed, retrying...", attempt))
			}
		}
	}

	return result, result.LastError
}

// runSingleAttempt executes one fix attempt.
func (m *CIFixManager) runSingleAttempt(ctx context.Context, attempt int) error {
	// Step 1: Get failure info
	if m.config.OnProgress != nil {
		m.config.OnProgress("Analyzing CI failure...")
	}

	failureInfo, err := m.analyzer.GetLatestFailure(ctx, m.config.PRNumber)
	if err != nil {
		return &CIFixError{
			Attempt: attempt,
			Phase:   "analyze",
			Message: "failed to get failure info",
			Err:     err,
		}
	}
	if failureInfo == nil {
		return &CIFixError{
			Attempt: attempt,
			Phase:   "analyze",
			Message: "no CI failures found",
		}
	}

	// Step 2: Build prompt
	if m.config.OnProgress != nil {
		m.config.OnProgress("Building fix prompt...")
	}

	promptCtx := prompt.CIFixContext{
		FailureInfo: toPromptFailureInfo(failureInfo),
		PRNumber:    m.config.PRNumber,
		BranchName:  m.config.BranchName,
		Attempt:     attempt,
		MaxAttempts: m.config.MaxRetries,
	}

	buildResult, err := m.promptBuilder.Build(promptCtx)
	if err != nil {
		return &CIFixError{
			Attempt: attempt,
			Phase:   "prompt",
			Message: "failed to build prompt",
			Err:     err,
		}
	}

	// Step 3: Execute Claude
	if m.config.OnProgress != nil {
		m.config.OnProgress("Running Claude to fix CI failure...")
	}

	_, err = m.claudeClient.Execute(ctx, buildResult.Prompt)
	if err != nil {
		return &CIFixError{
			Attempt: attempt,
			Phase:   "claude",
			Message: "Claude execution failed",
			Err:     err,
		}
	}

	// Step 4: Commit and push
	if m.config.OnProgress != nil {
		m.config.OnProgress("Committing and pushing fix...")
	}

	commitMsg := fmt.Sprintf("fix(ci): auto-fix CI failure (attempt %d)\n\nFailed workflow: %s\nFailed job: %s",
		attempt, failureInfo.WorkflowName, failureInfo.JobName)

	err = m.commitMgr.CommitAndPush(ctx, commitMsg)
	if err != nil {
		// Check if nothing to commit (Claude might not have made changes)
		if errors.Is(err, git.ErrNothingToCommit) {
			return &CIFixError{
				Attempt: attempt,
				Phase:   "commit",
				Message: "Claude made no changes",
			}
		}
		return &CIFixError{
			Attempt: attempt,
			Phase:   "commit",
			Message: "failed to commit/push",
			Err:     err,
		}
	}

	// Step 5: Wait for new checks
	if m.config.OnProgress != nil {
		m.config.OnProgress("Waiting for CI checks...")
	}

	err = m.checkMonitor.WaitForChecks(ctx, m.config.PRNumber, m.getWaitOptions())
	if err != nil {
		return &CIFixError{
			Attempt: attempt,
			Phase:   "verify",
			Message: "CI still failing after fix",
			Err:     err,
		}
	}

	return nil
}

// CIFixError represents a CI fix failure.
type CIFixError struct {
	Attempt int    // Which attempt this error occurred on
	Phase   string // "analyze", "prompt", "claude", "commit", "verify"
	Message string // Error message
	Err     error  // Underlying error
}

func (e *CIFixError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("CI fix attempt %d (%s): %s: %v", e.Attempt, e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("CI fix attempt %d (%s): %s", e.Attempt, e.Phase, e.Message)
}

func (e *CIFixError) Unwrap() error {
	return e.Err
}

// IsCIFixError checks if an error is a CIFixError.
func IsCIFixError(err error) bool {
	var cfe *CIFixError
	return errors.As(err, &cfe)
}

// Predefined CI fix errors
var (
	ErrCIFixDisabled   = &CIFixError{Message: "CI retry is disabled"}
	ErrNoFailuresFound = &CIFixError{Message: "no CI failures found"}
	ErrCIStillFailing  = &CIFixError{Message: "CI still failing after fix"}
)

// toPromptFailureInfo converts CIFailureInfo to prompt.CIFailureInfo.
// This avoids import cycles between github and prompt packages.
func toPromptFailureInfo(info *CIFailureInfo) *prompt.CIFailureInfo {
	return &prompt.CIFailureInfo{
		RunID:        info.RunID,
		WorkflowName: info.WorkflowName,
		JobName:      info.JobName,
		FailedSteps:  info.FailedSteps,
		ErrorLogs:    info.ErrorLogs,
		URL:          info.URL,
	}
}

// getWaitOptions returns the configured WaitOptions or defaults for CI re-runs.
func (m *CIFixManager) getWaitOptions() *WaitOptions {
	if m.config.WaitOptions != nil {
		return m.config.WaitOptions
	}
	// Use shorter initial wait for re-runs (30s vs default 3 min)
	defaults := DefaultWaitOptions()
	return &WaitOptions{
		MaxIterations: defaults.MaxIterations,
		PollInterval:  defaults.PollInterval,
		InitialWait:   30 * time.Second,
	}
}
