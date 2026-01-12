package git

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
)

// Repository provides Git repository operations.
type Repository struct {
	executor CommandExecutor
	rootPath string // cached root path
}

// NewRepository creates a new Repository.
// If executor is nil, DefaultExecutor is used.
func NewRepository(executor CommandExecutor) *Repository {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &Repository{executor: executor}
}

// IsGitRepository checks if the current directory is inside a Git repository.
func (r *Repository) IsGitRepository(ctx context.Context) (bool, error) {
	cmd := r.executor.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	if err != nil {
		// exit code != 0 means not a repository
		return false, nil
	}
	return strings.TrimSpace(string(output)) == "true", nil
}

// GetRootPath returns the root path of the Git repository.
func (r *Repository) GetRootPath(ctx context.Context) (string, error) {
	if r.rootPath != "" {
		return r.rootPath, nil
	}

	cmd := r.executor.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return "", &GitError{
			Operation: "repo",
			Message:   "failed to get repository root",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	r.rootPath = strings.TrimSpace(string(output))
	return r.rootPath, nil
}

// GetCurrentBranch returns the current branch name.
func (r *Repository) GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := r.executor.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return "", &GitError{
			Operation: "branch",
			Message:   "failed to get current branch",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL returns the URL of the origin remote.
// Returns empty string if origin remote doesn't exist.
func (r *Repository) GetRemoteURL(ctx context.Context) (string, error) {
	cmd := r.executor.CommandContext(ctx, "git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		// origin might not exist - return empty string, not an error
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}

// IsClean checks if the working directory is clean (no uncommitted changes).
func (r *Repository) IsClean(ctx context.Context) (bool, error) {
	cmd := r.executor.CommandContext(ctx, "git", "status", "--porcelain")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return false, &GitError{
			Operation: "repo",
			Message:   "failed to check working tree status",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	return len(strings.TrimSpace(string(output))) == 0, nil
}

// GetInfo returns comprehensive repository information.
func (r *Repository) GetInfo(ctx context.Context) (*RepoInfo, error) {
	isRepo, err := r.IsGitRepository(ctx)
	if err != nil {
		return nil, err
	}
	if !isRepo {
		return nil, ErrNotGitRepository
	}

	rootPath, err := r.GetRootPath(ctx)
	if err != nil {
		return nil, err
	}

	currentBranch, err := r.GetCurrentBranch(ctx)
	if err != nil {
		return nil, err
	}

	remoteURL, _ := r.GetRemoteURL(ctx) // ignore error (remote might not exist)

	isClean, err := r.IsClean(ctx)
	if err != nil {
		return nil, err
	}

	return &RepoInfo{
		RootPath:      rootPath,
		CurrentBranch: currentBranch,
		RemoteURL:     remoteURL,
		IsClean:       isClean,
	}, nil
}

// GetIterationDisplay returns a display string for iteration context.
// Returns worktreeName if provided, otherwise returns the current branch base name.
func (r *Repository) GetIterationDisplay(ctx context.Context, worktreeName string) string {
	if worktreeName != "" {
		return worktreeName
	}

	branch, err := r.GetCurrentBranch(ctx)
	if err != nil {
		return "unknown"
	}
	return filepath.Base(branch)
}
