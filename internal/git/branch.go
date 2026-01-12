package git

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// BranchManager provides branch operations.
type BranchManager struct {
	executor CommandExecutor
	repo     *Repository
}

// NewBranchManager creates a new BranchManager.
// If executor is nil, DefaultExecutor is used.
func NewBranchManager(executor CommandExecutor) *BranchManager {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &BranchManager{
		executor: executor,
		repo:     NewRepository(executor),
	}
}

// GenerateBranchName creates a branch name with date and random hash.
// Format: {prefix}YYYY-MM-DD-{random_hex}
// Example: claude-loop/2024-01-15-a3f2b1
func (b *BranchManager) GenerateBranchName(prefix string) (string, error) {
	if prefix == "" {
		prefix = DefaultBranchOptions().Prefix
	}

	// Date part
	dateStr := time.Now().Format("2006-01-02")

	// Random hash part (6 hex characters)
	hashBytes := make([]byte, 3)
	if _, err := rand.Read(hashBytes); err != nil {
		return "", &GitError{
			Operation: "branch",
			Message:   "failed to generate random hash",
			Err:       err,
		}
	}
	hashStr := hex.EncodeToString(hashBytes)

	return fmt.Sprintf("%s%s-%s", prefix, dateStr, hashStr), nil
}

// CreateBranch creates a new branch from the base branch.
// If opts is nil, defaults are used.
func (b *BranchManager) CreateBranch(ctx context.Context, name string, opts *BranchOptions) error {
	if opts == nil {
		opts = DefaultBranchOptions()
	}

	// Check if branch already exists
	exists, err := b.BranchExists(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return &BranchError{
			Branch:  name,
			Message: "branch already exists",
		}
	}

	// Determine base branch
	baseBranch := opts.BaseBranch
	if baseBranch == "" {
		baseBranch, err = b.repo.GetCurrentBranch(ctx)
		if err != nil {
			return err
		}
	}

	// Create branch
	args := []string{"branch", name, baseBranch}
	cmd := b.executor.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &BranchError{
			Branch:  name,
			Message: "failed to create branch",
			Err:     fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String())),
		}
	}

	return nil
}

// CreateIterationBranch creates a branch with auto-generated name.
// Returns the created branch name.
func (b *BranchManager) CreateIterationBranch(ctx context.Context, opts *BranchOptions) (string, error) {
	if opts == nil {
		opts = DefaultBranchOptions()
	}

	name, err := b.GenerateBranchName(opts.Prefix)
	if err != nil {
		return "", err
	}

	if err := b.CreateBranch(ctx, name, opts); err != nil {
		return "", err
	}

	return name, nil
}

// BranchExists checks if a branch exists.
func (b *BranchManager) BranchExists(ctx context.Context, name string) (bool, error) {
	cmd := b.executor.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	err := cmd.Run()
	if err != nil {
		// exit code != 0 means branch doesn't exist
		return false, nil
	}
	return true, nil
}

// DeleteBranch deletes a local branch.
// If force is true, uses -D (force delete), otherwise uses -d.
func (b *BranchManager) DeleteBranch(ctx context.Context, name string, force bool) error {
	exists, err := b.BranchExists(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		return &BranchError{
			Branch:  name,
			Message: "branch not found",
		}
	}

	flag := "-d"
	if force {
		flag = "-D"
	}

	cmd := b.executor.CommandContext(ctx, "git", "branch", flag, name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &BranchError{
			Branch:  name,
			Message: "failed to delete branch",
			Err:     fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String())),
		}
	}

	return nil
}

// Checkout switches to the specified branch.
func (b *BranchManager) Checkout(ctx context.Context, name string) error {
	cmd := b.executor.CommandContext(ctx, "git", "checkout", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &BranchError{
			Branch:  name,
			Message: "failed to checkout branch",
			Err:     fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String())),
		}
	}

	return nil
}

// ListBranches returns all local branch names.
func (b *BranchManager) ListBranches(ctx context.Context) ([]string, error) {
	cmd := b.executor.CommandContext(ctx, "git", "branch", "--format=%(refname:short)")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, &GitError{
			Operation: "branch",
			Message:   "failed to list branches",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return []string{}, nil
	}

	lines := strings.Split(trimmed, "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		if branch := strings.TrimSpace(line); branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}
