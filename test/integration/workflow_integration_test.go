package integration

import (
	"context"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/github"
	"github.com/DeukWoongWoo/claude-loop/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRManager_CreatePR_Integration(t *testing.T) {
	t.Run("creates PR successfully", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.SuccessCommand("https://github.com/testowner/testrepo/pull/42"),
		)

		repo := &github.RepoInfo{Owner: "testowner", Repo: "testrepo"}
		prManager := github.NewPRManager(mockExec, repo)

		prNum, url, err := prManager.Create(context.Background(), &github.PRCreateOptions{
			Title: "Test PR",
			Body:  "Description",
			Base:  "main",
		})

		require.NoError(t, err)
		assert.Equal(t, 42, prNum)
		assert.Contains(t, url, "github.com")
		assert.True(t, mockExec.AllCommandsUsed())
	})

	t.Run("handles creation error", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.FailCommand("error: could not create PR", 1),
		)

		repo := &github.RepoInfo{Owner: "testowner", Repo: "testrepo"}
		prManager := github.NewPRManager(mockExec, repo)

		_, _, err := prManager.Create(context.Background(), &github.PRCreateOptions{
			Title: "Test PR",
		})

		assert.Error(t, err)
	})
}

func TestPRManager_MergePR_Integration(t *testing.T) {
	t.Run("merges PR with squash strategy", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.SuccessCommand(""),
		)

		repo := &github.RepoInfo{Owner: "testowner", Repo: "testrepo"}
		prManager := github.NewPRManager(mockExec, repo)

		err := prManager.Merge(context.Background(), 42, github.MergeStrategySquash, true)

		require.NoError(t, err)
		assert.True(t, mockExec.AllCommandsUsed())
	})
}

func TestCheckMonitor_WaitForChecks_Integration(t *testing.T) {
	t.Run("returns when all checks pass", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			// First poll - checks pending
			mocks.SuccessCommand(`[{"name":"ci","state":"IN_PROGRESS","bucket":"pending"}]`),
			// Second poll - checks passed
			mocks.SuccessCommand(`[{"name":"ci","state":"SUCCESS","bucket":"pass"}]`),
		)

		repo := &github.RepoInfo{Owner: "testowner", Repo: "testrepo"}
		prManager := github.NewPRManager(mockExec, repo)
		monitor := prManager.GetCheckMonitor()

		opts := &github.WaitOptions{
			MaxIterations: 5,
			PollInterval:  10 * time.Millisecond,
			InitialWait:   0,
		}

		err := monitor.WaitForChecks(context.Background(), 42, opts)

		require.NoError(t, err)
	})

	t.Run("returns error when checks fail", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			// Checks failed
			mocks.SuccessCommand(`[{"name":"ci","state":"FAILURE","bucket":"fail"}]`),
		)

		repo := &github.RepoInfo{Owner: "testowner", Repo: "testrepo"}
		prManager := github.NewPRManager(mockExec, repo)
		monitor := prManager.GetCheckMonitor()

		opts := &github.WaitOptions{
			MaxIterations: 1,
			PollInterval:  10 * time.Millisecond,
			InitialWait:   0,
		}

		err := monitor.WaitForChecks(context.Background(), 42, opts)

		assert.Error(t, err)
		assert.True(t, github.IsCheckError(err))
	})
}

func TestRepoDetector_Integration(t *testing.T) {
	t.Run("detects GitHub repo from HTTPS URL", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.SuccessCommand("https://github.com/owner/repo.git"),
		)

		detector := github.NewRepoDetector(mockExec)
		info, err := detector.DetectRepo(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "owner", info.Owner)
		assert.Equal(t, "repo", info.Repo)
	})

	t.Run("detects GitHub repo from SSH URL", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.SuccessCommand("git@github.com:owner/repo.git"),
		)

		detector := github.NewRepoDetector(mockExec)
		info, err := detector.DetectRepo(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "owner", info.Owner)
		assert.Equal(t, "repo", info.Repo)
	})

	t.Run("returns error for non-GitHub repo", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.SuccessCommand("https://gitlab.com/owner/repo.git"),
		)

		detector := github.NewRepoDetector(mockExec)
		_, err := detector.DetectRepo(context.Background())

		assert.Error(t, err)
		assert.Equal(t, github.ErrNotGitHubRepo, err)
	})

	t.Run("validates gh CLI installed", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.SuccessCommand("gh version 2.40.0"),
			mocks.SuccessCommand("Logged in to github.com"),
		)

		detector := github.NewRepoDetector(mockExec)
		err := detector.ValidateGHCLI(context.Background())

		require.NoError(t, err)
	})

	t.Run("returns error when gh CLI not found", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.FailCommand("command not found: gh", 127),
		)

		detector := github.NewRepoDetector(mockExec)
		err := detector.ValidateGHCLI(context.Background())

		assert.Equal(t, github.ErrGHCLINotFound, err)
	})

	t.Run("returns error when gh not authenticated", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence(
			mocks.SuccessCommand("gh version 2.40.0"),
			mocks.FailCommand("not logged in", 1),
		)

		detector := github.NewRepoDetector(mockExec)
		err := detector.ValidateGHCLI(context.Background())

		assert.Equal(t, github.ErrGHNotAuthenticated, err)
	})
}

func TestWorkflowManager_DryRun_Integration(t *testing.T) {
	t.Run("dry run does not create PR", func(t *testing.T) {
		mockExec := mocks.NewCommandSequence()

		repo := &github.RepoInfo{Owner: "testowner", Repo: "testrepo"}
		manager := github.NewWorkflowManager(mockExec, repo)

		var progressMessages []string
		cfg := &github.WorkflowConfig{
			DryRun: true,
			OnProgress: func(status string) {
				progressMessages = append(progressMessages, status)
			},
		}

		result, err := manager.RunPRWorkflow(context.Background(), &github.PRCreateOptions{
			Title: "Test PR",
		}, cfg)

		require.NoError(t, err)
		assert.Equal(t, "Test PR", result.PRTitle)
		assert.Equal(t, 0, mockExec.CallCount())
		assert.Contains(t, progressMessages, "[DRY RUN] Would create PR: Test PR")
	})
}
