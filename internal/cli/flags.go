// Package cli provides the command-line interface for claude-loop.
package cli

import "time"

// Flags holds all CLI flag values for claude-loop.
type Flags struct {
	// Required options
	Prompt      string        // -p, --prompt: The prompt/goal for Claude Code
	MaxRuns     int           // -m, --max-runs: Maximum iterations (0 = unlimited with cost/duration)
	MaxCost     float64       // --max-cost: Maximum cost in USD
	MaxDuration time.Duration // --max-duration: Maximum duration

	// GitHub configuration
	Owner string // --owner: GitHub repository owner
	Repo  string // --repo: GitHub repository name

	// Commit & Branch management
	DisableCommits  bool   // --disable-commits: Disable automatic commits and PR creation
	DisableBranches bool   // --disable-branches: Commit on current branch only
	GitBranchPrefix string // --git-branch-prefix: Branch prefix (default: "claude-loop/")
	MergeStrategy   string // --merge-strategy: PR merge strategy (squash, merge, rebase)

	// Iteration control
	CompletionSignal    string // --completion-signal: Phrase indicating project complete
	CompletionThreshold int    // --completion-threshold: Consecutive signals to stop
	DryRun              bool   // --dry-run: Simulate without changes

	// Review & CI
	ReviewPrompt   string // -r, --review-prompt: Reviewer pass prompt
	DisableCIRetry bool   // --disable-ci-retry: Disable CI failure retry
	CIRetryMax     int    // --ci-retry-max: Maximum CI fix attempts

	// Shared state
	NotesFile string // --notes-file: Shared notes file path

	// Worktree support
	Worktree        string // --worktree: Git worktree name
	WorktreeBaseDir string // --worktree-base-dir: Base directory for worktrees
	CleanupWorktree bool   // --cleanup-worktree: Remove worktree after completion
	ListWorktrees   bool   // --list-worktrees: List worktrees and exit

	// Principles framework
	ResetPrinciples bool   // --reset-principles: Force re-collection of principles
	PrinciplesFile  string // --principles-file: Custom principles file path
	LogDecisions    bool   // --log-decisions: Enable decision logging

	// Output control
	Verbose bool // --verbose: Show detailed iteration summaries
	Stream  bool // --stream: Stream Claude output in real-time

	// Update management
	AutoUpdate     bool // --auto-update: Auto-install updates
	DisableUpdates bool // --disable-updates: Skip update checks

	// Planning mode
	Plan     bool   // --plan: Enable planning mode (PRD → Architecture → Tasks)
	PlanOnly bool   // --plan-only: Generate plan without execution
	Resume   string // --resume: Resume from saved plan ID
}

// DefaultFlags returns a Flags struct with default values as defined in CLI_CONTRACT.md.
func DefaultFlags() *Flags {
	return &Flags{
		// Commit & Branch defaults
		GitBranchPrefix: "claude-loop/",
		MergeStrategy:   "squash",

		// Iteration control defaults
		CompletionSignal:    "CONTINUOUS_CLAUDE_PROJECT_COMPLETE",
		CompletionThreshold: 3,

		// Review & CI defaults
		CIRetryMax: 1,

		// Shared state defaults
		NotesFile: "SHARED_TASK_NOTES.md",

		// Worktree defaults
		WorktreeBaseDir: "../claude-loop-worktrees",

		// Principles defaults
		PrinciplesFile: ".claude/principles.yaml",
	}
}

// globalFlags is the singleton instance used by the root command.
var globalFlags = DefaultFlags()

// GetFlags returns the current flag values.
func GetFlags() *Flags {
	return globalFlags
}

// ResetFlags resets flags to defaults (useful for testing).
func ResetFlags() {
	globalFlags = DefaultFlags()
}
