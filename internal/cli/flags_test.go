package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultFlags(t *testing.T) {
	f := DefaultFlags()

	// Verify default values match CLI_CONTRACT.md
	assert.Equal(t, "claude-loop/", f.GitBranchPrefix)
	assert.Equal(t, "squash", f.MergeStrategy)
	assert.Equal(t, "CONTINUOUS_CLAUDE_PROJECT_COMPLETE", f.CompletionSignal)
	assert.Equal(t, 3, f.CompletionThreshold)
	assert.Equal(t, 1, f.CIRetryMax)
	assert.Equal(t, "SHARED_TASK_NOTES.md", f.NotesFile)
	assert.Equal(t, "../claude-loop-worktrees", f.WorktreeBaseDir)
	assert.Equal(t, ".claude/principles.yaml", f.PrinciplesFile)

	// Boolean defaults should be false
	assert.False(t, f.DisableCommits)
	assert.False(t, f.DisableBranches)
	assert.False(t, f.DryRun)
	assert.False(t, f.DisableCIRetry)
	assert.False(t, f.CleanupWorktree)
	assert.False(t, f.ListWorktrees)
	assert.False(t, f.ResetPrinciples)
	assert.False(t, f.LogDecisions)
	assert.False(t, f.Verbose)
	assert.False(t, f.Stream)
	assert.False(t, f.AutoUpdate)
	assert.False(t, f.DisableUpdates)

	// Planning mode defaults should be false/empty
	assert.False(t, f.Plan)
	assert.False(t, f.PlanOnly)
	assert.Empty(t, f.Resume)
}

func TestFlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T)
	}{
		{
			name: "prompt short flag",
			args: []string{"-p", "test prompt", "-m", "1"},
			validate: func(t *testing.T) {
				assert.Equal(t, "test prompt", globalFlags.Prompt)
			},
		},
		{
			name: "prompt long flag",
			args: []string{"--prompt", "long prompt test", "-m", "1"},
			validate: func(t *testing.T) {
				assert.Equal(t, "long prompt test", globalFlags.Prompt)
			},
		},
		{
			name: "max-runs short flag",
			args: []string{"-p", "x", "-m", "10"},
			validate: func(t *testing.T) {
				assert.Equal(t, 10, globalFlags.MaxRuns)
			},
		},
		{
			name: "max-runs long flag",
			args: []string{"-p", "x", "--max-runs", "25"},
			validate: func(t *testing.T) {
				assert.Equal(t, 25, globalFlags.MaxRuns)
			},
		},
		{
			name: "max-cost flag",
			args: []string{"-p", "x", "--max-cost", "5.50"},
			validate: func(t *testing.T) {
				assert.Equal(t, 5.50, globalFlags.MaxCost)
			},
		},
		{
			name: "max-duration flag",
			args: []string{"-p", "x", "--max-duration", "2h30m"},
			validate: func(t *testing.T) {
				// Duration is parsed via parseDuration after Execute
				err := parseDuration()
				require.NoError(t, err)
				assert.Equal(t, 2*time.Hour+30*time.Minute, globalFlags.MaxDuration)
			},
		},
		{
			name: "review-prompt short flag",
			args: []string{"-p", "x", "-m", "1", "-r", "run tests"},
			validate: func(t *testing.T) {
				assert.Equal(t, "run tests", globalFlags.ReviewPrompt)
			},
		},
		{
			name: "boolean flags",
			args: []string{"-p", "x", "-m", "1", "--dry-run", "--disable-commits", "--disable-branches"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.DryRun)
				assert.True(t, globalFlags.DisableCommits)
				assert.True(t, globalFlags.DisableBranches)
			},
		},
		{
			name: "github flags",
			args: []string{"-p", "x", "-m", "1", "--owner", "myuser", "--repo", "myrepo"},
			validate: func(t *testing.T) {
				assert.Equal(t, "myuser", globalFlags.Owner)
				assert.Equal(t, "myrepo", globalFlags.Repo)
			},
		},
		{
			name: "merge-strategy flag",
			args: []string{"-p", "x", "-m", "1", "--merge-strategy", "rebase"},
			validate: func(t *testing.T) {
				assert.Equal(t, "rebase", globalFlags.MergeStrategy)
			},
		},
		{
			name: "worktree flags",
			args: []string{"-p", "x", "-m", "1", "--worktree", "instance-1", "--worktree-base-dir", "/tmp/wt", "--cleanup-worktree"},
			validate: func(t *testing.T) {
				assert.Equal(t, "instance-1", globalFlags.Worktree)
				assert.Equal(t, "/tmp/wt", globalFlags.WorktreeBaseDir)
				assert.True(t, globalFlags.CleanupWorktree)
			},
		},
		{
			name: "list-worktrees standalone",
			args: []string{"--list-worktrees"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.ListWorktrees)
			},
		},
		{
			name: "principles flags",
			args: []string{"-p", "x", "-m", "1", "--reset-principles", "--principles-file", "custom.yaml", "--log-decisions"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.ResetPrinciples)
				assert.Equal(t, "custom.yaml", globalFlags.PrinciplesFile)
				assert.True(t, globalFlags.LogDecisions)
			},
		},
		{
			name: "ci-retry flags",
			args: []string{"-p", "x", "-m", "1", "--disable-ci-retry"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.DisableCIRetry)
			},
		},
		{
			name: "ci-retry-max flag",
			args: []string{"-p", "x", "-m", "1", "--ci-retry-max", "3"},
			validate: func(t *testing.T) {
				assert.Equal(t, 3, globalFlags.CIRetryMax)
			},
		},
		{
			name: "completion flags",
			args: []string{"-p", "x", "-m", "1", "--completion-signal", "DONE", "--completion-threshold", "5"},
			validate: func(t *testing.T) {
				assert.Equal(t, "DONE", globalFlags.CompletionSignal)
				assert.Equal(t, 5, globalFlags.CompletionThreshold)
			},
		},
		{
			name: "notes-file flag",
			args: []string{"-p", "x", "-m", "1", "--notes-file", "NOTES.md"},
			validate: func(t *testing.T) {
				assert.Equal(t, "NOTES.md", globalFlags.NotesFile)
			},
		},
		{
			name: "update flags",
			args: []string{"-p", "x", "-m", "1", "--auto-update"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.AutoUpdate)
			},
		},
		{
			name: "disable-updates flag",
			args: []string{"-p", "x", "-m", "1", "--disable-updates"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.DisableUpdates)
			},
		},
		{
			name: "output control flags",
			args: []string{"-p", "x", "-m", "1", "--verbose", "--stream"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.Verbose)
				assert.True(t, globalFlags.Stream)
			},
		},
		{
			name: "plan flag",
			args: []string{"-p", "test", "--plan", "-m", "1"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.Plan)
				assert.False(t, globalFlags.PlanOnly)
			},
		},
		{
			name: "plan-only flag",
			args: []string{"-p", "test", "--plan-only", "-m", "1"},
			validate: func(t *testing.T) {
				assert.True(t, globalFlags.PlanOnly)
			},
		},
		{
			name: "resume flag",
			args: []string{"--resume", "plan-123456789", "-m", "1"},
			validate: func(t *testing.T) {
				assert.Equal(t, "plan-123456789", globalFlags.Resume)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for testing (NewRootCmd resets global state)
			cmd := NewRootCmd()
			// Override Run to prevent actual execution - we only test flag parsing
			cmd.Run = func(cmd *cobra.Command, args []string) {}
			cmd.SetArgs(tt.args)

			// Suppress output during tests
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			// Execute (ignore error for flag parsing tests, validation is separate)
			_ = cmd.Execute()

			// Validate the parsed values
			tt.validate(t)
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "hours only",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "minutes only",
			input:    "30m",
			expected: 30 * time.Minute,
		},
		{
			name:     "seconds only",
			input:    "90s",
			expected: 90 * time.Second,
		},
		{
			name:     "combined hours and minutes",
			input:    "1h30m",
			expected: 1*time.Hour + 30*time.Minute,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetFlags()
			maxDurationStr = tt.input

			err := parseDuration()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid duration format")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, globalFlags.MaxDuration)
			}
		})
	}
}

func TestGetFlags(t *testing.T) {
	ResetFlags()
	f := GetFlags()

	assert.NotNil(t, f)
	assert.Same(t, globalFlags, f)
}

func TestResetFlags(t *testing.T) {
	// Modify some flags
	globalFlags.Prompt = "modified"
	globalFlags.MaxRuns = 100
	globalFlags.DryRun = true

	// Reset
	ResetFlags()

	// Verify reset to defaults
	assert.Empty(t, globalFlags.Prompt)
	assert.Equal(t, 0, globalFlags.MaxRuns)
	assert.False(t, globalFlags.DryRun)
	assert.Equal(t, "squash", globalFlags.MergeStrategy)
}

func TestConfigToLoopConfig(t *testing.T) {
	flags := &Flags{
		Prompt:              "test prompt",
		MaxRuns:             10,
		MaxCost:             5.50,
		MaxDuration:         2 * time.Hour,
		CompletionSignal:    "DONE",
		CompletionThreshold: 5,
		DryRun:              true,
		NotesFile:           "NOTES.md",
		ReviewPrompt:        "run tests",
		LogDecisions:        true,
	}

	cfg := ConfigToLoopConfig(flags)

	assert.Equal(t, "test prompt", cfg.Prompt)
	assert.Equal(t, 10, cfg.MaxRuns)
	assert.Equal(t, 5.50, cfg.MaxCost)
	assert.Equal(t, 2*time.Hour, cfg.MaxDuration)
	assert.Equal(t, "DONE", cfg.CompletionSignal)
	assert.Equal(t, 5, cfg.CompletionThreshold)
	assert.Equal(t, 3, cfg.MaxConsecutiveErrors) // hardcoded default
	assert.True(t, cfg.DryRun)
	assert.Equal(t, "NOTES.md", cfg.NotesFile)
	assert.Equal(t, "run tests", cfg.ReviewPrompt)
	assert.True(t, cfg.LogDecisions)
	// Principles should be nil (set separately after loading)
	assert.Nil(t, cfg.Principles)
}
