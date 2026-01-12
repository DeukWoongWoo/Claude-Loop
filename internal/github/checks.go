package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CheckMonitor monitors PR check statuses.
type CheckMonitor struct {
	executor CommandExecutor
	repo     *RepoInfo
}

// NewCheckMonitor creates a new CheckMonitor.
func NewCheckMonitor(executor CommandExecutor, repo *RepoInfo) *CheckMonitor {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &CheckMonitor{
		executor: executor,
		repo:     repo,
	}
}

// checkResponse represents the JSON response from gh pr checks.
type checkResponse struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Bucket string `json:"bucket"`
}

// GetCheckStatus retrieves current check status for a PR.
func (m *CheckMonitor) GetCheckStatus(ctx context.Context, prNumber int) (*CheckSummary, error) {
	cmd := m.executor.CommandContext(ctx, "gh", "pr", "checks",
		fmt.Sprintf("%d", prNumber),
		"--repo", m.repo.RepoString(),
		"--json", "name,state,bucket",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		// Only treat as "no checks" if stderr explicitly says so
		if strings.Contains(stderrStr, "no checks") {
			return &CheckSummary{
				NoChecks:     true,
				AllCompleted: true,
				AllPassed:    true,
			}, nil
		}
		// Empty stdout with error should NOT be treated as "no checks"
		return nil, &GitHubError{
			Operation: "checks",
			Message:   "failed to get check status",
			Stderr:    stderrStr,
			Err:       err,
		}
	}

	// Parse JSON response
	var checks []checkResponse
	if err := json.Unmarshal(stdout.Bytes(), &checks); err != nil {
		return nil, &GitHubError{
			Operation: "checks",
			Message:   "failed to parse check response",
			Err:       err,
		}
	}

	// Handle empty checks array
	if len(checks) == 0 {
		return &CheckSummary{
			NoChecks:     true,
			AllCompleted: true,
			AllPassed:    true,
		}, nil
	}

	summary := &CheckSummary{
		Checks: make([]CheckStatus, len(checks)),
		Total:  len(checks),
	}

	for i, check := range checks {
		bucket := CheckBucket(check.Bucket)
		summary.Checks[i] = CheckStatus{
			Name:   check.Name,
			State:  check.State,
			Bucket: bucket,
		}

		switch bucket {
		case CheckBucketPending:
			summary.Pending++
		case CheckBucketPass:
			summary.Success++
		case CheckBucketFail:
			summary.Failed++
		}
	}

	summary.AllCompleted = summary.Pending == 0
	summary.AllPassed = summary.AllCompleted && summary.Failed == 0

	return summary, nil
}

// reviewResponse represents the JSON response for review status.
type reviewResponse struct {
	ReviewDecision string `json:"reviewDecision"`
	ReviewRequests []struct {
		RequestedReviewer struct {
			Login string `json:"login"`
		} `json:"requestedReviewer"`
	} `json:"reviewRequests"`
}

// GetReviewStatus retrieves review status for a PR.
// Returns: reviewDecision (APPROVED, CHANGES_REQUESTED, REVIEW_REQUIRED, or ""), pendingReviewCount, error
func (m *CheckMonitor) GetReviewStatus(ctx context.Context, prNumber int) (string, int, error) {
	cmd := m.executor.CommandContext(ctx, "gh", "pr", "view",
		fmt.Sprintf("%d", prNumber),
		"--repo", m.repo.RepoString(),
		"--json", "reviewDecision,reviewRequests",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", 0, &GitHubError{
			Operation: "review",
			Message:   "failed to get review status",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	var review reviewResponse
	if err := json.Unmarshal(stdout.Bytes(), &review); err != nil {
		return "", 0, &GitHubError{
			Operation: "review",
			Message:   "failed to parse review response",
			Err:       err,
		}
	}

	return review.ReviewDecision, len(review.ReviewRequests), nil
}

// WaitForChecks polls until all checks pass or timeout.
func (m *CheckMonitor) WaitForChecks(ctx context.Context, prNumber int, opts *WaitOptions) error {
	if opts == nil {
		opts = DefaultWaitOptions()
	}

	// Initial wait for checks to start
	if opts.InitialWait > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(opts.InitialWait):
		}
	}

	var lastSummary *CheckSummary
	var lastReviewStatus string

	for i := 0; i < opts.MaxIterations; i++ {
		summary, err := m.GetCheckStatus(ctx, prNumber)
		if err != nil {
			return err
		}

		var reviewStatus string
		if opts.RequireApproval {
			reviewStatus, _, err = m.GetReviewStatus(ctx, prNumber)
			if err != nil {
				return err
			}
		}

		// Call status change callback if status changed
		if opts.OnStatusChange != nil && (!checksEqual(lastSummary, summary) || lastReviewStatus != reviewStatus) {
			opts.OnStatusChange(summary, reviewStatus)
		}
		lastSummary = summary
		lastReviewStatus = reviewStatus

		// Check for failure
		if summary.Failed > 0 {
			return &CheckError{
				PRNumber: prNumber,
				Message:  "checks failed",
				Summary:  summary,
			}
		}

		// Check review status if required
		if opts.RequireApproval && reviewStatus == "CHANGES_REQUESTED" {
			return ErrChangesRequested
		}

		// Success condition: checks passed (or no checks) and approval obtained (if required)
		checksPassed := summary.AllPassed || summary.NoChecks
		approvalObtained := !opts.RequireApproval || reviewStatus == "APPROVED"
		if checksPassed && approvalObtained {
			return nil
		}

		// Wait before next poll
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(opts.PollInterval):
		}
	}

	return &CheckError{
		PRNumber: prNumber,
		Message:  "timeout waiting for checks",
		Summary:  lastSummary,
	}
}

// checksEqual compares two CheckSummary for equality.
func checksEqual(a, b *CheckSummary) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Total == b.Total &&
		a.Pending == b.Pending &&
		a.Success == b.Success &&
		a.Failed == b.Failed
}

// failedRunResponse represents the JSON response for workflow runs.
type failedRunResponse struct {
	DatabaseID int    `json:"databaseId"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// GetFailedRunID returns the most recent failed workflow run ID for a PR.
func (m *CheckMonitor) GetFailedRunID(ctx context.Context, prNumber int) (string, error) {
	// First get the PR's head SHA
	cmd := m.executor.CommandContext(ctx, "gh", "pr", "view",
		fmt.Sprintf("%d", prNumber),
		"--repo", m.repo.RepoString(),
		"--json", "headRefOid",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", &GitHubError{
			Operation: "pr",
			Message:   "failed to get PR head SHA",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	var prInfo struct {
		HeadRefOid string `json:"headRefOid"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &prInfo); err != nil {
		return "", &GitHubError{
			Operation: "pr",
			Message:   "failed to parse PR info",
			Err:       err,
		}
	}

	// Get failed runs for this commit
	cmd = m.executor.CommandContext(ctx, "gh", "run", "list",
		"--repo", m.repo.RepoString(),
		"--commit", prInfo.HeadRefOid,
		"--status", "failure",
		"--limit", "1",
		"--json", "databaseId,status,conclusion",
	)
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", &GitHubError{
			Operation: "run",
			Message:   "failed to list workflow runs",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	var runs []failedRunResponse
	if err := json.Unmarshal(stdout.Bytes(), &runs); err != nil {
		return "", &GitHubError{
			Operation: "run",
			Message:   "failed to parse run list",
			Err:       err,
		}
	}

	if len(runs) == 0 {
		return "", nil // No failed runs
	}

	return fmt.Sprintf("%d", runs[0].DatabaseID), nil
}
