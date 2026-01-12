package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorktreeManager provides worktree operations.
type WorktreeManager struct {
	executor CommandExecutor
	repo     *Repository
	branch   *BranchManager
}

// NewWorktreeManager creates a new WorktreeManager.
// If executor is nil, DefaultExecutor is used.
func NewWorktreeManager(executor CommandExecutor) *WorktreeManager {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &WorktreeManager{
		executor: executor,
		repo:     NewRepository(executor),
		branch:   NewBranchManager(executor),
	}
}

// List returns all worktrees for the repository.
// Implements --list-worktrees CLI flag behavior.
func (w *WorktreeManager) List(ctx context.Context) ([]WorktreeInfo, error) {
	cmd := w.executor.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, &GitError{
			Operation: "worktree",
			Message:   "failed to list worktrees",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	return w.parseWorktreeList(string(output))
}

// parseWorktreeList parses the porcelain format output of git worktree list.
func (w *WorktreeManager) parseWorktreeList(output string) ([]WorktreeInfo, error) {
	var worktrees []WorktreeInfo
	var current *WorktreeInfo

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Empty line marks end of a worktree entry
		if line == "" {
			if current != nil {
				worktrees = append(worktrees, *current)
				current = nil
			}
			continue
		}

		// Start of a new worktree entry
		if strings.HasPrefix(line, "worktree ") {
			current = &WorktreeInfo{
				Path: strings.TrimPrefix(line, "worktree "),
			}
			continue
		}

		// Skip if no current worktree being parsed
		if current == nil {
			continue
		}

		// Parse worktree attributes
		switch {
		case strings.HasPrefix(line, "HEAD "):
			hash := strings.TrimPrefix(line, "HEAD ")
			if len(hash) >= 7 {
				current.CommitHash = hash[:7]
			} else {
				current.CommitHash = hash
			}
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "bare":
			current.IsBare = true
		case line == "detached":
			current.Branch = "(detached)"
		}
	}

	// Handle last entry if output doesn't end with empty line
	if current != nil {
		worktrees = append(worktrees, *current)
	}

	// Mark first worktree as main
	if len(worktrees) > 0 {
		worktrees[0].IsMain = true
	}

	return worktrees, nil
}

// Setup creates or reuses a worktree.
// Returns the worktree path.
func (w *WorktreeManager) Setup(ctx context.Context, name string, opts *WorktreeOptions) (string, error) {
	if opts == nil {
		opts = DefaultWorktreeOptions()
	}

	// Validate worktree name to prevent path traversal
	if name == "" {
		return "", &WorktreeError{
			Path:    name,
			Message: "worktree name cannot be empty",
		}
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return "", &WorktreeError{
			Path:    name,
			Message: "worktree name contains invalid characters",
		}
	}

	// Determine worktree path
	baseDir := opts.BaseDir
	if !filepath.IsAbs(baseDir) {
		repoRoot, err := w.repo.GetRootPath(ctx)
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(filepath.Dir(repoRoot), filepath.Base(baseDir))
	}
	worktreePath := filepath.Join(baseDir, name)

	// Check for existing worktree
	existing, err := w.List(ctx)
	if err != nil {
		return "", err
	}

	for _, wt := range existing {
		if wt.Path == worktreePath {
			// Already exists, return path
			return worktreePath, nil
		}
	}

	// Create base directory
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", &WorktreeError{
			Path:    worktreePath,
			Message: "failed to create base directory",
			Err:     err,
		}
	}

	// Build git worktree add command
	args := []string{"worktree", "add"}
	if opts.CreateBranch {
		args = append(args, "-b", name)
	}
	args = append(args, worktreePath)
	if opts.BaseBranch != "" {
		args = append(args, opts.BaseBranch)
	}

	cmd := w.executor.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", &WorktreeError{
			Path:    worktreePath,
			Message: "failed to create worktree",
			Err:     fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String())),
		}
	}

	return worktreePath, nil
}

// Remove deletes a worktree.
// Implements --cleanup-worktree CLI flag behavior.
func (w *WorktreeManager) Remove(ctx context.Context, pathOrName string, force bool) error {
	// If not absolute path, try to find matching worktree
	worktreePath := pathOrName
	if !filepath.IsAbs(pathOrName) {
		existing, err := w.List(ctx)
		if err != nil {
			return err
		}

		found := false
		for _, wt := range existing {
			if filepath.Base(wt.Path) == pathOrName || wt.Path == pathOrName {
				worktreePath = wt.Path
				found = true
				break
			}
		}

		if !found {
			return &WorktreeError{
				Path:    pathOrName,
				Message: "worktree not found",
			}
		}
	}

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)

	cmd := w.executor.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &WorktreeError{
			Path:    worktreePath,
			Message: "failed to remove worktree",
			Err:     fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String())),
		}
	}

	return nil
}

// Prune removes stale worktree entries.
func (w *WorktreeManager) Prune(ctx context.Context) error {
	cmd := w.executor.CommandContext(ctx, "git", "worktree", "prune")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &GitError{
			Operation: "worktree",
			Message:   "failed to prune worktrees",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	return nil
}

// FormatList returns a formatted string of worktrees for display.
// Used by --list-worktrees CLI command.
func (w *WorktreeManager) FormatList(worktrees []WorktreeInfo) string {
	if len(worktrees) == 0 {
		return "No worktrees found."
	}

	var sb strings.Builder
	sb.WriteString("Active worktrees:\n")

	for _, wt := range worktrees {
		marker := "  "
		if wt.IsMain {
			marker = "* "
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", marker, wt.Path))
		sb.WriteString(fmt.Sprintf("    Branch: %s\n", wt.Branch))
		if wt.CommitHash != "" {
			sb.WriteString(fmt.Sprintf("    Commit: %s\n", wt.CommitHash))
		}
	}

	return sb.String()
}
