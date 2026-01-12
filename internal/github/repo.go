package github

import (
	"bytes"
	"context"
	"regexp"
	"strings"
)

// Pre-compiled regexes for parsing remote URLs.
var (
	httpsRemoteRegex = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
	sshRemoteRegex   = regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
)

// RepoDetector detects GitHub repository from git remote.
type RepoDetector struct {
	executor CommandExecutor
}

// NewRepoDetector creates a new RepoDetector.
func NewRepoDetector(executor CommandExecutor) *RepoDetector {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &RepoDetector{executor: executor}
}

// DetectRepo detects owner/repo from the git remote URL.
func (d *RepoDetector) DetectRepo(ctx context.Context) (*RepoInfo, error) {
	cmd := d.executor.CommandContext(ctx, "git", "remote", "get-url", "origin")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, &GitHubError{
			Operation: "repo",
			Message:   "failed to get remote URL",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	url := strings.TrimSpace(stdout.String())
	return parseRemoteURL(url)
}

// parseRemoteURL parses GitHub owner/repo from remote URL.
// Supports both HTTPS (https://github.com/owner/repo.git) and SSH (git@github.com:owner/repo.git) formats.
func parseRemoteURL(url string) (*RepoInfo, error) {
	if matches := httpsRemoteRegex.FindStringSubmatch(url); matches != nil {
		return &RepoInfo{Owner: matches[1], Repo: matches[2]}, nil
	}
	if matches := sshRemoteRegex.FindStringSubmatch(url); matches != nil {
		return &RepoInfo{Owner: matches[1], Repo: matches[2]}, nil
	}
	return nil, ErrNotGitHubRepo
}

// ValidateGHCLI checks if gh CLI is available and authenticated.
func (d *RepoDetector) ValidateGHCLI(ctx context.Context) error {
	// Check if gh is installed
	cmd := d.executor.CommandContext(ctx, "gh", "--version")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return ErrGHCLINotFound
	}

	// Check if authenticated
	cmd = d.executor.CommandContext(ctx, "gh", "auth", "status")
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return ErrGHNotAuthenticated
	}

	return nil
}

// RepoString returns the "owner/repo" format string.
func (r *RepoInfo) RepoString() string {
	return r.Owner + "/" + r.Repo
}
