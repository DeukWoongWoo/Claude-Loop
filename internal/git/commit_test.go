package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitManager_StageAll(t *testing.T) {
	t.Run("stages all changes successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.StageAll(context.Background())
		require.NoError(t, err)
	})

	t.Run("error on staging failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "fatal: not a git repository"},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.StageAll(context.Background())
		assert.Error(t, err)
		assert.True(t, IsGitError(err))
		assert.Contains(t, err.Error(), "failed to stage changes")
	})
}

func TestCommitManager_HasStagedChanges(t *testing.T) {
	t.Run("has staged changes", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1}, // git diff --cached --quiet exits 1 when there are changes
			},
		}
		cm := NewCommitManager(mock)

		hasChanges, err := cm.HasStagedChanges(context.Background())
		require.NoError(t, err)
		assert.True(t, hasChanges)
	})

	t.Run("no staged changes", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 0}, // git diff --cached --quiet exits 0 when clean
			},
		}
		cm := NewCommitManager(mock)

		hasChanges, err := cm.HasStagedChanges(context.Background())
		require.NoError(t, err)
		assert.False(t, hasChanges)
	})
}

func TestCommitManager_Commit(t *testing.T) {
	t.Run("creates commit successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "[main abc1234] Test commit"},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.Commit(context.Background(), "Test commit")
		require.NoError(t, err)
	})

	t.Run("returns ErrNothingToCommit when nothing to commit", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "nothing to commit, working tree clean"},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.Commit(context.Background(), "Test commit")
		assert.ErrorIs(t, err, ErrNothingToCommit)
	})

	t.Run("error on commit failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "error: commit failed"},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.Commit(context.Background(), "Test commit")
		assert.Error(t, err)
		assert.True(t, IsGitError(err))
		assert.Contains(t, err.Error(), "failed to create commit")
	})
}

func TestCommitManager_Push(t *testing.T) {
	t.Run("pushes to origin successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.Push(context.Background(), "", "")
		require.NoError(t, err)
	})

	t.Run("pushes to specific remote", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.Push(context.Background(), "upstream", "main")
		require.NoError(t, err)
	})

	t.Run("error on push failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "error: failed to push some refs"},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.Push(context.Background(), "", "")
		assert.Error(t, err)
		assert.True(t, IsGitError(err))
		assert.Contains(t, err.Error(), "failed to push changes")
	})
}

func TestCommitManager_CommitAndPush(t *testing.T) {
	t.Run("commits and pushes successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},                           // git add -A
				{ExitCode: 1},                          // git diff --cached --quiet (has changes)
				{Stdout: "[main abc1234] Test commit"}, // git commit
				{Stdout: ""},                           // git push
			},
		}
		cm := NewCommitManager(mock)

		err := cm.CommitAndPush(context.Background(), "Test commit")
		require.NoError(t, err)
	})

	t.Run("returns ErrNothingToCommit when no changes", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},  // git add -A
				{ExitCode: 0}, // git diff --cached --quiet (no changes)
			},
		}
		cm := NewCommitManager(mock)

		err := cm.CommitAndPush(context.Background(), "Test commit")
		assert.ErrorIs(t, err, ErrNothingToCommit)
	})

	t.Run("error on stage failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "fatal: not a git repository"},
			},
		}
		cm := NewCommitManager(mock)

		err := cm.CommitAndPush(context.Background(), "Test commit")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to stage changes")
	})

	t.Run("error on commit failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},                         // git add -A
				{ExitCode: 1},                        // git diff --cached --quiet (has changes)
				{ExitCode: 1, Stderr: "commit error"}, // git commit fails
			},
		}
		cm := NewCommitManager(mock)

		err := cm.CommitAndPush(context.Background(), "Test commit")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create commit")
	})

	t.Run("error on push failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},                           // git add -A
				{ExitCode: 1},                          // git diff --cached --quiet (has changes)
				{Stdout: "[main abc1234] Test commit"}, // git commit
				{ExitCode: 1, Stderr: "push error"},    // git push fails
			},
		}
		cm := NewCommitManager(mock)

		err := cm.CommitAndPush(context.Background(), "Test commit")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to push changes")
	})
}

func TestNewCommitManager(t *testing.T) {
	t.Run("uses default executor when nil", func(t *testing.T) {
		cm := NewCommitManager(nil)
		assert.NotNil(t, cm)
		assert.NotNil(t, cm.executor)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		cm := NewCommitManager(mock)
		assert.Equal(t, mock, cm.executor)
	})
}
