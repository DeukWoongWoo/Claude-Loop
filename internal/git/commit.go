package git

import (
	"bytes"
	"context"
	"strings"
)

// CommitManager provides Git commit and push operations.
type CommitManager struct {
	executor CommandExecutor
}

// NewCommitManager creates a new CommitManager.
// If executor is nil, DefaultExecutor is used.
func NewCommitManager(executor CommandExecutor) *CommitManager {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &CommitManager{executor: executor}
}

// StageAll stages all changes (git add -A).
func (c *CommitManager) StageAll(ctx context.Context) error {
	cmd := c.executor.CommandContext(ctx, "git", "add", "-A")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &GitError{
			Operation: "commit",
			Message:   "failed to stage changes",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}
	return nil
}

// HasStagedChanges checks if there are staged changes to commit.
// Returns true if there are staged changes, false otherwise.
func (c *CommitManager) HasStagedChanges(ctx context.Context) (bool, error) {
	cmd := c.executor.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	err := cmd.Run()
	if err != nil {
		// Exit code != 0 means there are changes
		return true, nil
	}
	return false, nil
}

// Commit creates a commit with the given message.
func (c *CommitManager) Commit(ctx context.Context, message string) error {
	cmd := c.executor.CommandContext(ctx, "git", "commit", "-m", message)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if strings.Contains(stderrStr, "nothing to commit") {
			return ErrNothingToCommit
		}
		return &GitError{
			Operation: "commit",
			Message:   "failed to create commit",
			Stderr:    stderrStr,
			Err:       err,
		}
	}
	return nil
}

// Push pushes to remote.
// If remote is empty, defaults to "origin".
// If branch is empty, pushes the current branch.
func (c *CommitManager) Push(ctx context.Context, remote, branch string) error {
	if remote == "" {
		remote = "origin"
	}

	args := []string{"push", remote}
	if branch != "" {
		args = append(args, branch)
	}

	cmd := c.executor.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &GitError{
			Operation: "commit",
			Message:   "failed to push changes",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}
	return nil
}

// CommitAndPush stages, commits, and pushes changes in one operation.
// Returns ErrNothingToCommit if there are no changes to commit.
func (c *CommitManager) CommitAndPush(ctx context.Context, message string) error {
	if err := c.StageAll(ctx); err != nil {
		return err
	}

	hasChanges, err := c.HasStagedChanges(ctx)
	if err != nil {
		return err
	}
	if !hasChanges {
		return ErrNothingToCommit
	}

	if err := c.Commit(ctx, message); err != nil {
		return err
	}

	return c.Push(ctx, "", "")
}
