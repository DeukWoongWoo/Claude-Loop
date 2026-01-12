package github

import (
	"context"
	"fmt"
)

// WorkflowManager orchestrates the complete PR workflow.
type WorkflowManager struct {
	executor  CommandExecutor
	repo      *RepoInfo
	prManager *PRManager
	monitor   *CheckMonitor
}

// NewWorkflowManager creates a new WorkflowManager.
func NewWorkflowManager(executor CommandExecutor, repo *RepoInfo) *WorkflowManager {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	prManager := NewPRManager(executor, repo)
	return &WorkflowManager{
		executor:  executor,
		repo:      repo,
		prManager: prManager,
		monitor:   prManager.GetCheckMonitor(),
	}
}

// RunPRWorkflow executes the complete PR workflow:
// 1. Create PR
// 2. Wait for checks
// 3. Merge PR
func (w *WorkflowManager) RunPRWorkflow(ctx context.Context, opts *PRCreateOptions, cfg *WorkflowConfig) (*WorkflowResult, error) {
	if cfg == nil {
		cfg = DefaultWorkflowConfig()
	}

	result := &WorkflowResult{}

	// Step 1: Create PR
	if cfg.OnProgress != nil {
		cfg.OnProgress("Creating pull request...")
	}

	if cfg.DryRun {
		if opts == nil {
			return nil, &GitHubError{Operation: "pr", Message: "create options required for dry run"}
		}
		if cfg.OnProgress != nil {
			cfg.OnProgress("[DRY RUN] Would create PR: " + opts.Title)
		}
		result.PRTitle = opts.Title
		return result, nil
	}

	prNum, url, err := w.prManager.Create(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("create PR: %w", err)
	}

	result.PRNumber = prNum
	result.PRURL = url
	result.PRTitle = opts.Title

	if cfg.OnProgress != nil {
		cfg.OnProgress(fmt.Sprintf("Created PR #%d: %s", prNum, url))
	}

	// Step 2: Wait for checks
	if cfg.OnProgress != nil {
		cfg.OnProgress("Waiting for CI checks...")
	}

	waitOpts := cfg.WaitOptions
	if waitOpts == nil {
		waitOpts = DefaultWaitOptions()
	}

	// Wire up status callback
	if cfg.OnCheckStatus != nil {
		originalCallback := waitOpts.OnStatusChange
		waitOpts.OnStatusChange = func(summary *CheckSummary, reviewStatus string) {
			cfg.OnCheckStatus(summary)
			if originalCallback != nil {
				originalCallback(summary, reviewStatus)
			}
		}
	}

	err = w.monitor.WaitForChecks(ctx, prNum, waitOpts)
	if err != nil {
		// Get final check summary for result
		summary, _ := w.monitor.GetCheckStatus(ctx, prNum)
		result.CheckSummary = summary
		result.MergeError = err
		return result, err
	}

	// Get final check summary
	summary, _ := w.monitor.GetCheckStatus(ctx, prNum)
	result.CheckSummary = summary

	if cfg.OnProgress != nil {
		cfg.OnProgress("All checks passed!")
	}

	// Step 3: Merge PR
	if cfg.OnProgress != nil {
		cfg.OnProgress(fmt.Sprintf("Merging PR with %s strategy...", cfg.MergeStrategy))
	}

	err = w.prManager.Merge(ctx, prNum, cfg.MergeStrategy, cfg.DeleteBranch)
	if err != nil {
		result.MergeError = err
		return result, fmt.Errorf("merge PR: %w", err)
	}

	result.Merged = true

	if cfg.OnProgress != nil {
		cfg.OnProgress(fmt.Sprintf("PR #%d merged successfully!", prNum))
	}

	return result, nil
}

// MergeAndCleanup merges an existing PR and cleans up.
// This is equivalent to merge_pr_and_cleanup() in the bash script.
func (w *WorkflowManager) MergeAndCleanup(ctx context.Context, prNumber int, strategy MergeStrategy, deleteBranch bool) error {
	// First verify the PR exists and is mergeable
	info, err := w.prManager.GetInfo(ctx, prNumber)
	if err != nil {
		return fmt.Errorf("get PR info: %w", err)
	}

	if info.State != PRStateOpen {
		return &PRError{
			PRNumber: prNumber,
			Message:  fmt.Sprintf("PR is not open (state: %s)", info.State),
		}
	}

	if !info.IsMergeable {
		return &PRError{
			PRNumber: prNumber,
			Message:  "PR is not mergeable",
		}
	}

	// Merge the PR
	if err := w.prManager.Merge(ctx, prNumber, strategy, deleteBranch); err != nil {
		return fmt.Errorf("merge PR: %w", err)
	}

	return nil
}

// HandleFailedChecks handles PR check failures by closing the PR.
func (w *WorkflowManager) HandleFailedChecks(ctx context.Context, prNumber int, deleteBranch bool) error {
	return w.prManager.Close(ctx, prNumber, deleteBranch)
}

// TryUpdateBranch attempts to update the PR branch with the latest from base.
// Returns nil if successful or already up-to-date, error otherwise.
func (w *WorkflowManager) TryUpdateBranch(ctx context.Context, prNumber int) error {
	err := w.prManager.UpdateBranch(ctx, prNumber)
	if err == ErrPRAlreadyUpToDate {
		return nil // Already up-to-date is not an error
	}
	return err
}

// GetPRManager returns the underlying PRManager.
func (w *WorkflowManager) GetPRManager() *PRManager {
	return w.prManager
}

// GetCheckMonitor returns the underlying CheckMonitor.
func (w *WorkflowManager) GetCheckMonitor() *CheckMonitor {
	return w.monitor
}
