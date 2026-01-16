package integration

import (
	"bytes"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_FlagParsing_AllFlags(t *testing.T) {
	// Note: All tests include --dry-run and --disable-updates to prevent the loop
	// from actually executing during tests.
	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, flags *cli.Flags)
		wantErr  bool
	}{
		{
			name: "minimal valid flags with max-runs",
			args: []string{"-p", "test prompt", "-m", "5", "--dry-run", "--disable-updates"},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "test prompt", f.Prompt)
				assert.Equal(t, 5, f.MaxRuns)
			},
		},
		{
			name: "minimal valid flags with max-cost",
			args: []string{"-p", "test prompt", "--max-cost", "10.50", "--dry-run", "--disable-updates"},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "test prompt", f.Prompt)
				assert.Equal(t, 10.50, f.MaxCost)
			},
		},
		{
			name: "minimal valid flags with max-duration",
			args: []string{"-p", "test prompt", "--max-duration", "2h30m", "--dry-run", "--disable-updates"},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "test prompt", f.Prompt)
				assert.Equal(t, 2*time.Hour+30*time.Minute, f.MaxDuration)
			},
		},
		{
			name: "all limit types combined",
			args: []string{
				"-p", "prompt",
				"-m", "10",
				"--max-cost", "5.50",
				"--max-duration", "1h",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, 10, f.MaxRuns)
				assert.Equal(t, 5.50, f.MaxCost)
				assert.Equal(t, time.Hour, f.MaxDuration)
			},
		},
		{
			name: "github options",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--owner", "myorg",
				"--repo", "myrepo",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "myorg", f.Owner)
				assert.Equal(t, "myrepo", f.Repo)
			},
		},
		{
			name: "commit management flags",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--disable-commits",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.True(t, f.DisableCommits)
			},
		},
		{
			name: "branch management flags",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--disable-branches",
				"--git-branch-prefix", "ai/",
				"--merge-strategy", "rebase",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.True(t, f.DisableBranches)
				assert.Equal(t, "ai/", f.GitBranchPrefix)
				assert.Equal(t, "rebase", f.MergeStrategy)
			},
		},
		{
			name: "iteration control flags",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--completion-signal", "DONE",
				"--completion-threshold", "5",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "DONE", f.CompletionSignal)
				assert.Equal(t, 5, f.CompletionThreshold)
				assert.True(t, f.DryRun)
			},
		},
		{
			name: "reviewer flag",
			args: []string{
				"-p", "prompt", "-m", "1",
				"-r", "run tests",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "run tests", f.ReviewPrompt)
			},
		},
		{
			name: "reviewer flag long form",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--review-prompt", "npm test && npm run lint",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "npm test && npm run lint", f.ReviewPrompt)
			},
		},
		{
			name: "CI retry flags",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--ci-retry-max", "3",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, 3, f.CIRetryMax)
			},
		},
		{
			name: "disable CI retry",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--disable-ci-retry",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.True(t, f.DisableCIRetry)
			},
		},
		{
			name: "notes file flag",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--notes-file", "NOTES.md",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "NOTES.md", f.NotesFile)
			},
		},
		{
			name: "worktree flags",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--worktree", "test-worktree",
				"--worktree-base-dir", "/custom/path",
				"--cleanup-worktree",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.Equal(t, "test-worktree", f.Worktree)
				assert.Equal(t, "/custom/path", f.WorktreeBaseDir)
				assert.True(t, f.CleanupWorktree)
			},
		},
		{
			name: "principles flags",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--reset-principles",
				"--principles-file", "custom.yaml",
				"--log-decisions",
				"--dry-run", "--disable-updates",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.True(t, f.ResetPrinciples)
				assert.Equal(t, "custom.yaml", f.PrinciplesFile)
				assert.True(t, f.LogDecisions)
			},
		},
		{
			name: "update flags",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--auto-update",
				"--dry-run",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.True(t, f.AutoUpdate)
			},
		},
		{
			name: "disable updates flag",
			args: []string{
				"-p", "prompt", "-m", "1",
				"--disable-updates",
				"--dry-run",
			},
			validate: func(t *testing.T, f *cli.Flags) {
				assert.True(t, f.DisableUpdates)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use NewRootCmdForFlagParsing to avoid actually running the loop
			cmd := cli.NewRootCmdForFlagParsing()
			cmd.SetArgs(tt.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, cli.GetFlags())
			}
		})
	}
}

func TestCLI_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{
			name:        "missing prompt",
			args:        []string{"-m", "5"},
			errContains: "prompt is required",
		},
		{
			name:        "missing limit",
			args:        []string{"-p", "test prompt"},
			errContains: "at least one limit",
		},
		{
			name:        "invalid merge strategy",
			args:        []string{"-p", "prompt", "-m", "1", "--merge-strategy", "invalid"},
			errContains: "merge-strategy must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewRootCmdForFlagParsing()
			cmd.SetArgs(tt.args)

			var stderr bytes.Buffer
			cmd.SetErr(&stderr)

			err := cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestCLI_DefaultValues(t *testing.T) {
	cmd := cli.NewRootCmdForFlagParsing()
	cmd.SetArgs([]string{"-p", "test", "-m", "1"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err)

	f := cli.GetFlags()
	assert.Equal(t, "claude-loop/", f.GitBranchPrefix)
	assert.Equal(t, "squash", f.MergeStrategy)
	assert.Equal(t, "CONTINUOUS_CLAUDE_PROJECT_COMPLETE", f.CompletionSignal)
	assert.Equal(t, 3, f.CompletionThreshold)
	assert.Equal(t, 1, f.CIRetryMax)
	assert.Equal(t, "SHARED_TASK_NOTES.md", f.NotesFile)
	assert.Equal(t, "../claude-loop-worktrees", f.WorktreeBaseDir)
	assert.Equal(t, ".claude/principles.yaml", f.PrinciplesFile)
}

func TestCLI_HelpAndVersion(t *testing.T) {
	t.Run("help flag shows usage", func(t *testing.T) {
		cmd := cli.NewRootCmdForFlagParsing()
		cmd.SetArgs([]string{"--help"})

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "claude-loop")
		assert.Contains(t, output, "--prompt")
		assert.Contains(t, output, "--max-runs")
	})

	t.Run("version flag shows version", func(t *testing.T) {
		cmd := cli.NewRootCmdForFlagParsing()
		cmd.SetArgs([]string{"--version"})

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "claude-loop version")
	})
}

func TestCLI_ListWorktrees_SkipsValidation(t *testing.T) {
	// --list-worktrees should work without prompt and limits
	// Note: This test uses NewRootCmd since it actually needs to execute
	// the list-worktrees functionality (which exits early without running the loop)
	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"--list-worktrees"})

	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	// Should not error due to missing prompt/limits
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestCLI_DurationParsing(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		expected time.Duration
	}{
		{"hours", "2h", 2 * time.Hour},
		{"minutes", "30m", 30 * time.Minute},
		{"seconds", "90s", 90 * time.Second},
		{"combined", "1h30m", 90 * time.Minute},
		{"complex", "2h30m15s", 2*time.Hour + 30*time.Minute + 15*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewRootCmdForFlagParsing()
			cmd.SetArgs([]string{"-p", "test", "--max-duration", tt.duration})

			err := cmd.Execute()
			require.NoError(t, err)

			f := cli.GetFlags()
			assert.Equal(t, tt.expected, f.MaxDuration)
		})
	}
}

func TestCLI_InvalidDuration(t *testing.T) {
	cmd := cli.NewRootCmdForFlagParsing()
	cmd.SetArgs([]string{"-p", "test", "--max-duration", "invalid"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration")
}

func TestCLI_MergeStrategyValidation(t *testing.T) {
	validStrategies := []string{"squash", "merge", "rebase"}

	for _, strategy := range validStrategies {
		t.Run(strategy, func(t *testing.T) {
			cmd := cli.NewRootCmdForFlagParsing()
			cmd.SetArgs([]string{"-p", "test", "-m", "1", "--merge-strategy", strategy})

			err := cmd.Execute()
			require.NoError(t, err)

			f := cli.GetFlags()
			assert.Equal(t, strategy, f.MergeStrategy)
		})
	}
}
