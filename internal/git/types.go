// Package git provides Git operations for branch and worktree management.
package git

import (
	"context"
	"os/exec"
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

// RepoInfo contains Git repository information.
type RepoInfo struct {
	RootPath      string // Git repository root path
	CurrentBranch string // Currently checked out branch
	RemoteURL     string // origin remote URL
	IsClean       bool   // Working directory status (clean = no uncommitted changes)
}

// WorktreeInfo contains information about a Git worktree.
type WorktreeInfo struct {
	Path       string // Worktree absolute path
	Branch     string // Branch name in the worktree
	CommitHash string // HEAD commit hash (short form)
	IsBare     bool   // Whether this is a bare worktree
	IsMain     bool   // Whether this is the main worktree
}

// BranchOptions configures branch creation.
type BranchOptions struct {
	Prefix     string // Branch name prefix (default: "claude-loop/")
	BaseBranch string // Base branch to create from (default: current branch)
}

// DefaultBranchOptions returns BranchOptions with default values.
func DefaultBranchOptions() *BranchOptions {
	return &BranchOptions{
		Prefix: "claude-loop/",
	}
}

// WorktreeOptions configures worktree creation.
type WorktreeOptions struct {
	BaseDir      string // Worktree base directory (default: "../claude-loop-worktrees")
	CreateBranch bool   // Whether to create a new branch for the worktree
	BaseBranch   string // Base branch when CreateBranch is true
}

// DefaultWorktreeOptions returns WorktreeOptions with default values.
func DefaultWorktreeOptions() *WorktreeOptions {
	return &WorktreeOptions{
		BaseDir: "../claude-loop-worktrees",
	}
}
