package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Pre-compiled regex for parsing PR URL.
var prURLRegex = regexp.MustCompile(`/pull/(\d+)`)

// PRManager manages pull request operations.
type PRManager struct {
	executor CommandExecutor
	repo     *RepoInfo
	monitor  *CheckMonitor
}

// NewPRManager creates a new PRManager.
func NewPRManager(executor CommandExecutor, repo *RepoInfo) *PRManager {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &PRManager{
		executor: executor,
		repo:     repo,
		monitor:  NewCheckMonitor(executor, repo),
	}
}

// Create creates a new pull request.
// Returns the PR number, URL, and any error.
func (p *PRManager) Create(ctx context.Context, opts *PRCreateOptions) (int, string, error) {
	if opts == nil {
		return 0, "", &GitHubError{
			Operation: "pr",
			Message:   "create options required",
		}
	}

	args := []string{"pr", "create",
		"--repo", p.repo.RepoString(),
		"--title", opts.Title,
		"--body", opts.Body,
	}

	if opts.Base != "" {
		args = append(args, "--base", opts.Base)
	}

	if opts.Draft {
		args = append(args, "--draft")
	}

	cmd := p.executor.CommandContext(ctx, "gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, "", &GitHubError{
			Operation: "pr",
			Message:   "failed to create PR",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	// Output is the PR URL
	url := strings.TrimSpace(stdout.String())
	prNum, err := parsePRNumber(url)
	if err != nil {
		return 0, url, err
	}

	return prNum, url, nil
}

// parsePRNumber extracts PR number from a GitHub PR URL.
func parsePRNumber(url string) (int, error) {
	matches := prURLRegex.FindStringSubmatch(url)
	if matches == nil {
		return 0, &GitHubError{
			Operation: "pr",
			Message:   "failed to parse PR number from URL: " + url,
		}
	}
	return strconv.Atoi(matches[1])
}

// prViewResponse represents the JSON response from gh pr view.
type prViewResponse struct {
	Number         int       `json:"number"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	State          string    `json:"state"`
	HeadRefName    string    `json:"headRefName"`
	BaseRefName    string    `json:"baseRefName"`
	HeadRefOid     string    `json:"headRefOid"`
	ReviewDecision string    `json:"reviewDecision"`
	ReviewRequests []any     `json:"reviewRequests"`
	Mergeable      string    `json:"mergeable"`
	CreatedAt      time.Time `json:"createdAt"`
	URL            string    `json:"url"`
}

// GetInfo retrieves PR information.
func (p *PRManager) GetInfo(ctx context.Context, prNumber int) (*PRInfo, error) {
	cmd := p.executor.CommandContext(ctx, "gh", "pr", "view",
		fmt.Sprintf("%d", prNumber),
		"--repo", p.repo.RepoString(),
		"--json", "number,title,body,state,headRefName,baseRefName,headRefOid,reviewDecision,reviewRequests,mergeable,createdAt,url",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if strings.Contains(stderrStr, "not found") {
			return nil, ErrPRNotFound
		}
		return nil, &GitHubError{
			Operation: "pr",
			Message:   "failed to get PR info",
			Stderr:    stderrStr,
			Err:       err,
		}
	}

	var resp prViewResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return nil, &GitHubError{
			Operation: "pr",
			Message:   "failed to parse PR response",
			Err:       err,
		}
	}

	return &PRInfo{
		Number:         resp.Number,
		Title:          resp.Title,
		Body:           resp.Body,
		State:          PRState(resp.State),
		HeadBranch:     resp.HeadRefName,
		BaseBranch:     resp.BaseRefName,
		HeadRefOid:     resp.HeadRefOid,
		ReviewDecision: resp.ReviewDecision,
		ReviewRequests: len(resp.ReviewRequests),
		IsMergeable:    resp.Mergeable == "MERGEABLE",
		CreatedAt:      resp.CreatedAt,
		URL:            resp.URL,
	}, nil
}

// UpdateBranch updates the PR branch with the latest from base.
func (p *PRManager) UpdateBranch(ctx context.Context, prNumber int) error {
	cmd := p.executor.CommandContext(ctx, "gh", "pr", "update-branch",
		fmt.Sprintf("%d", prNumber),
		"--repo", p.repo.RepoString(),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if strings.Contains(stderrStr, "already up to date") ||
			strings.Contains(stderrStr, "already up-to-date") {
			return ErrPRAlreadyUpToDate
		}
		if strings.Contains(stderrStr, "conflict") {
			return ErrPRMergeConflict
		}
		return &GitHubError{
			Operation: "update",
			Message:   "failed to update PR branch",
			Stderr:    stderrStr,
			Err:       err,
		}
	}

	return nil
}

// Merge merges the PR with the specified strategy.
func (p *PRManager) Merge(ctx context.Context, prNumber int, strategy MergeStrategy, deleteBranch bool) error {
	args := []string{"pr", "merge",
		fmt.Sprintf("%d", prNumber),
		"--repo", p.repo.RepoString(),
	}

	switch strategy {
	case MergeStrategyMerge:
		args = append(args, "--merge")
	case MergeStrategyRebase:
		args = append(args, "--rebase")
	default:
		args = append(args, "--squash")
	}

	if deleteBranch {
		args = append(args, "--delete-branch")
	}

	cmd := p.executor.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if strings.Contains(stderrStr, "conflict") {
			return ErrPRMergeConflict
		}
		if strings.Contains(stderrStr, "not mergeable") {
			return ErrPRMergeFailed
		}
		return &GitHubError{
			Operation: "merge",
			Message:   "failed to merge PR",
			Stderr:    stderrStr,
			Err:       err,
		}
	}

	return nil
}

// Close closes the PR without merging.
func (p *PRManager) Close(ctx context.Context, prNumber int, deleteBranch bool) error {
	args := []string{"pr", "close",
		fmt.Sprintf("%d", prNumber),
		"--repo", p.repo.RepoString(),
	}

	if deleteBranch {
		args = append(args, "--delete-branch")
	}

	cmd := p.executor.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &GitHubError{
			Operation: "pr",
			Message:   "failed to close PR",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	return nil
}

// GetCheckMonitor returns the CheckMonitor for this PRManager.
func (p *PRManager) GetCheckMonitor() *CheckMonitor {
	return p.monitor
}
