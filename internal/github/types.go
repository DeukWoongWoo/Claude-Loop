// Package github provides GitHub PR workflow operations via gh CLI.
package github

import (
	"context"
	"os/exec"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/loop"
)

// CommandExecutor abstracts exec.Command for testing.
type CommandExecutor interface {
	CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
}

// DefaultExecutor uses the real exec.CommandContext.
type DefaultExecutor struct{}

// CommandContext creates a new exec.Cmd with the given context.
func (e *DefaultExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// RepoInfo contains GitHub repository information.
type RepoInfo struct {
	Owner string // Repository owner (user or organization)
	Repo  string // Repository name
}

// PRInfo contains pull request information.
type PRInfo struct {
	Number         int       // PR number
	Title          string    // PR title
	Body           string    // PR description
	State          PRState   // open, closed, merged
	HeadBranch     string    // Source branch
	BaseBranch     string    // Target branch
	HeadRefOid     string    // HEAD commit SHA
	ReviewDecision string    // APPROVED, CHANGES_REQUESTED, REVIEW_REQUIRED, or empty
	ReviewRequests int       // Number of pending review requests
	IsMergeable    bool      // Whether PR can be merged
	CreatedAt      time.Time // Creation timestamp
	URL            string    // PR URL
}

// PRState represents the state of a pull request.
type PRState string

const (
	PRStateOpen   PRState = "OPEN"
	PRStateClosed PRState = "CLOSED"
	PRStateMerged PRState = "MERGED"
)

// CheckStatus represents the status of a single CI check.
type CheckStatus struct {
	Name   string      // Check name
	State  string      // Raw state from GitHub
	Bucket CheckBucket // Categorized bucket (pending, pass, fail)
}

// CheckBucket categorizes check results.
type CheckBucket string

const (
	CheckBucketPending CheckBucket = "pending"
	CheckBucketPass    CheckBucket = "pass"
	CheckBucketFail    CheckBucket = "fail"
)

// CheckSummary aggregates check statuses.
type CheckSummary struct {
	Checks       []CheckStatus // Individual check statuses
	Total        int           // Total number of checks
	Pending      int           // Checks still running
	Success      int           // Passed checks
	Failed       int           // Failed checks
	AllCompleted bool          // All checks have finished
	AllPassed    bool          // All checks passed (requires AllCompleted)
	NoChecks     bool          // No checks configured for this PR
}

// MergeStrategy defines how PRs should be merged.
type MergeStrategy string

const (
	MergeStrategySquash MergeStrategy = "squash"
	MergeStrategyMerge  MergeStrategy = "merge"
	MergeStrategyRebase MergeStrategy = "rebase"
)

// WaitOptions configures PR check waiting behavior.
type WaitOptions struct {
	MaxIterations   int                                              // Max polling iterations (default: 180 = 30 min at 10s intervals)
	PollInterval    time.Duration                                    // Time between polls (default: 10s)
	InitialWait     time.Duration                                    // Wait for checks to start (default: 3 min)
	OnStatusChange  func(summary *CheckSummary, reviewStatus string) // Callback on status change
	RequireApproval bool                                             // Whether to require review approval
}

// DefaultWaitOptions returns WaitOptions with default values.
func DefaultWaitOptions() *WaitOptions {
	return &WaitOptions{
		MaxIterations:   180,              // 30 minutes at 10s intervals
		PollInterval:    10 * time.Second, // 10 seconds between polls
		InitialWait:     3 * time.Minute,  // 3 minutes initial wait
		RequireApproval: false,
	}
}

// PRCreateOptions configures PR creation.
type PRCreateOptions struct {
	Title string // PR title
	Body  string // PR description
	Base  string // Base branch (default: main)
	Draft bool   // Create as draft PR
}

// WorkflowConfig configures the PR workflow.
type WorkflowConfig struct {
	MergeStrategy MergeStrategy               // How to merge (squash, merge, rebase)
	WaitOptions   *WaitOptions                // Check waiting configuration
	DryRun        bool                        // Don't actually create/merge PRs
	DeleteBranch  bool                        // Delete branch after merge
	OnProgress    func(status string)         // Progress callback
	OnCheckStatus func(summary *CheckSummary) // Check status callback

	// CIFixConfig enables CI failure auto-fix (nil disables)
	CIFixConfig *CIFixConfig

	// ClaudeClient is required if CIFixConfig is set
	ClaudeClient loop.ClaudeClient
}

// DefaultWorkflowConfig returns WorkflowConfig with default values.
func DefaultWorkflowConfig() *WorkflowConfig {
	return &WorkflowConfig{
		MergeStrategy: MergeStrategySquash,
		WaitOptions:   DefaultWaitOptions(),
		DryRun:        false,
		DeleteBranch:  true,
	}
}

// WorkflowResult represents the outcome of a PR workflow.
type WorkflowResult struct {
	PRNumber     int           // PR number
	PRTitle      string        // PR title
	PRURL        string        // PR URL
	Merged       bool          // Whether PR was merged
	MergeError   error         // Error during merge (if any)
	CheckSummary *CheckSummary // Final check status
}
