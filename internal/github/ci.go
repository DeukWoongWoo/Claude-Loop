package github

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"time"
)

// CIAnalyzer retrieves and parses CI failure information.
type CIAnalyzer struct {
	executor     CommandExecutor
	repo         *RepoInfo
	checkMonitor *CheckMonitor
}

// NewCIAnalyzer creates a new CIAnalyzer.
// If executor is nil, DefaultExecutor is used.
func NewCIAnalyzer(executor CommandExecutor, repo *RepoInfo) *CIAnalyzer {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &CIAnalyzer{
		executor:     executor,
		repo:         repo,
		checkMonitor: NewCheckMonitor(executor, repo),
	}
}

// CIFailureInfo contains information about a CI failure.
type CIFailureInfo struct {
	RunID        string    // Workflow run ID
	WorkflowName string    // Name of the workflow
	JobName      string    // Failed job name
	FailedSteps  []string  // Names of failed steps
	ErrorLogs    string    // Truncated error log output
	URL          string    // Link to the failed run
	Timestamp    time.Time // When the failure occurred
}

// runViewResponse represents JSON from gh run view.
type runViewResponse struct {
	DatabaseID   int           `json:"databaseId"`
	DisplayTitle string        `json:"displayTitle"`
	Name         string        `json:"name"`
	Conclusion   string        `json:"conclusion"`
	URL          string        `json:"url"`
	CreatedAt    time.Time     `json:"createdAt"`
	Jobs         []runViewJob  `json:"jobs"`
}

// runViewJob represents a job in the run view response.
type runViewJob struct {
	Name       string        `json:"name"`
	Conclusion string        `json:"conclusion"`
	Steps      []runViewStep `json:"steps"`
}

// runViewStep represents a step in a job.
type runViewStep struct {
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
}

// GetFailureLogs retrieves logs for a failed workflow run.
func (a *CIAnalyzer) GetFailureLogs(ctx context.Context, runID string) (*CIFailureInfo, error) {
	if runID == "" {
		return nil, &GitHubError{
			Operation: "ci",
			Message:   "runID cannot be empty",
		}
	}

	// Get run metadata
	cmd := a.executor.CommandContext(ctx, "gh", "run", "view", runID,
		"--repo", a.repo.RepoString(),
		"--json", "databaseId,displayTitle,name,conclusion,url,createdAt,jobs",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, &GitHubError{
			Operation: "ci",
			Message:   "failed to get run info",
			Stderr:    strings.TrimSpace(stderr.String()),
			Err:       err,
		}
	}

	var runInfo runViewResponse
	if err := json.Unmarshal(stdout.Bytes(), &runInfo); err != nil {
		return nil, &GitHubError{
			Operation: "ci",
			Message:   "failed to parse run info",
			Err:       err,
		}
	}

	info := &CIFailureInfo{
		RunID:        runID,
		WorkflowName: runInfo.Name,
		URL:          runInfo.URL,
		Timestamp:    runInfo.CreatedAt,
	}

	// Find failed job and steps
	for _, job := range runInfo.Jobs {
		if job.Conclusion == "failure" {
			info.JobName = job.Name
			for _, step := range job.Steps {
				if step.Conclusion == "failure" {
					info.FailedSteps = append(info.FailedSteps, step.Name)
				}
			}
			break // Use first failed job
		}
	}

	// Get failed logs (truncated output)
	info.ErrorLogs = a.getFailedLogs(ctx, runID)

	return info, nil
}

// maxLogSize is the maximum log size to avoid token limits.
const maxLogSize = 5000

// getFailedLogs retrieves and truncates failed log output.
// Returns a placeholder message if log retrieval fails.
func (a *CIAnalyzer) getFailedLogs(ctx context.Context, runID string) string {
	cmd := a.executor.CommandContext(ctx, "gh", "run", "view", runID,
		"--repo", a.repo.RepoString(),
		"--log-failed",
	)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "(failed to retrieve logs)"
	}

	logs := stdout.String()
	if len(logs) > maxLogSize {
		return "...(truncated)...\n" + logs[len(logs)-maxLogSize:]
	}
	return logs
}

// GetLatestFailure gets the most recent failure for a PR.
// Returns nil, nil if no failures are found.
func (a *CIAnalyzer) GetLatestFailure(ctx context.Context, prNumber int) (*CIFailureInfo, error) {
	runID, err := a.checkMonitor.GetFailedRunID(ctx, prNumber)
	if err != nil {
		return nil, err
	}
	if runID == "" {
		return nil, nil // No failed runs
	}
	return a.GetFailureLogs(ctx, runID)
}
