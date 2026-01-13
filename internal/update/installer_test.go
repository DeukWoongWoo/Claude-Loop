package update

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockInstallerExecutor is a test double for CommandExecutor in installer tests.
type MockInstallerExecutor struct {
	VersionCheckSuccess bool
	Called              [][]string
}

func (m *MockInstallerExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	m.Called = append(m.Called, append([]string{name}, args...))

	if m.VersionCheckSuccess {
		return exec.CommandContext(ctx, "true")
	}
	return exec.CommandContext(ctx, "false")
}

func TestInstaller_Install(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Install test not supported on Windows")
	}

	tmpDir := t.TempDir()

	// Create mock binaries
	currentBinary := filepath.Join(tmpDir, "claude-loop")
	newBinary := filepath.Join(tmpDir, "new-claude-loop")

	err := os.WriteFile(currentBinary, []byte("#!/bin/bash\necho 'old version'"), 0755)
	require.NoError(t, err)

	err = os.WriteFile(newBinary, []byte("#!/bin/bash\necho 'new version'"), 0755)
	require.NoError(t, err)

	installer := NewInstaller(&InstallerOptions{
		BinaryPath:   currentBinary,
		BackupSuffix: ".bak",
		Executor:     &MockInstallerExecutor{VersionCheckSuccess: true},
	})

	ctx := context.Background()
	err = installer.Install(ctx, newBinary)
	require.NoError(t, err)

	// Verify the binary was replaced
	content, err := os.ReadFile(currentBinary)
	require.NoError(t, err)
	assert.Contains(t, string(content), "new version")
}

func TestInstaller_Install_BackupCreated(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Install test not supported on Windows")
	}

	tmpDir := t.TempDir()

	currentBinary := filepath.Join(tmpDir, "claude-loop")
	newBinary := filepath.Join(tmpDir, "new-claude-loop")
	backupPath := currentBinary + ".bak"

	// Create binaries
	err := os.WriteFile(currentBinary, []byte("old"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(newBinary, []byte("new"), 0755)
	require.NoError(t, err)

	installer := NewInstaller(&InstallerOptions{
		BinaryPath:   currentBinary,
		BackupSuffix: ".bak",
		Executor:     &MockInstallerExecutor{VersionCheckSuccess: true},
	})

	ctx := context.Background()
	err = installer.Install(ctx, newBinary)
	require.NoError(t, err)

	// Backup should be removed after successful install
	_, err = os.Stat(backupPath)
	assert.True(t, os.IsNotExist(err))
}

func TestInstaller_Install_VerifyFails(t *testing.T) {
	tmpDir := t.TempDir()

	currentBinary := filepath.Join(tmpDir, "claude-loop")
	newBinary := filepath.Join(tmpDir, "new-claude-loop")

	// Create current binary
	err := os.WriteFile(currentBinary, []byte("old"), 0755)
	require.NoError(t, err)

	// Create new binary
	err = os.WriteFile(newBinary, []byte("new"), 0755)
	require.NoError(t, err)

	installer := NewInstaller(&InstallerOptions{
		BinaryPath:   currentBinary,
		BackupSuffix: ".bak",
		Executor:     &MockInstallerExecutor{VersionCheckSuccess: false}, // Version check fails
	})

	ctx := context.Background()
	err = installer.Install(ctx, newBinary)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version check")
}

func TestInstaller_Install_BinaryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	currentBinary := filepath.Join(tmpDir, "claude-loop")

	err := os.WriteFile(currentBinary, []byte("old"), 0755)
	require.NoError(t, err)

	installer := NewInstaller(&InstallerOptions{
		BinaryPath: currentBinary,
	})

	ctx := context.Background()
	err = installer.Install(ctx, filepath.Join(tmpDir, "nonexistent"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInstaller_Install_PathIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	currentBinary := filepath.Join(tmpDir, "claude-loop")
	newDir := filepath.Join(tmpDir, "new-dir")

	err := os.WriteFile(currentBinary, []byte("old"), 0755)
	require.NoError(t, err)
	err = os.Mkdir(newDir, 0755)
	require.NoError(t, err)

	installer := NewInstaller(&InstallerOptions{
		BinaryPath: currentBinary,
	})

	ctx := context.Background()
	err = installer.Install(ctx, newDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "directory")
}

func TestInstaller_Rollback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Rollback test not supported on Windows")
	}

	tmpDir := t.TempDir()
	currentBinary := filepath.Join(tmpDir, "claude-loop")
	backupPath := currentBinary + ".bak"

	// Create current (new) and backup (old) binaries
	err := os.WriteFile(currentBinary, []byte("new"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(backupPath, []byte("old"), 0755)
	require.NoError(t, err)

	installer := NewInstaller(&InstallerOptions{
		BinaryPath:   currentBinary,
		BackupSuffix: ".bak",
	})

	ctx := context.Background()
	err = installer.Rollback(ctx)
	require.NoError(t, err)

	// Verify the old binary was restored
	content, err := os.ReadFile(currentBinary)
	require.NoError(t, err)
	assert.Equal(t, "old", string(content))

	// Backup should be removed
	_, err = os.Stat(backupPath)
	assert.True(t, os.IsNotExist(err))
}

func TestInstaller_Rollback_NoBackup(t *testing.T) {
	tmpDir := t.TempDir()
	currentBinary := filepath.Join(tmpDir, "claude-loop")

	err := os.WriteFile(currentBinary, []byte("current"), 0755)
	require.NoError(t, err)

	installer := NewInstaller(&InstallerOptions{
		BinaryPath:   currentBinary,
		BackupSuffix: ".bak",
	})

	ctx := context.Background()
	err = installer.Rollback(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "backup not found")
}

func TestInstaller_GetCurrentBinaryPath(t *testing.T) {
	t.Run("with custom path", func(t *testing.T) {
		customPath := "/custom/path/claude-loop"
		installer := NewInstaller(&InstallerOptions{
			BinaryPath: customPath,
		})

		path, err := installer.GetCurrentBinaryPath()
		require.NoError(t, err)
		assert.Equal(t, customPath, path)
	})

	t.Run("without custom path", func(t *testing.T) {
		installer := NewInstaller(nil)

		path, err := installer.GetCurrentBinaryPath()
		require.NoError(t, err)
		assert.NotEmpty(t, path)
	})
}

func TestNewInstaller_Defaults(t *testing.T) {
	installer := NewInstaller(nil)
	assert.NotNil(t, installer.opts)
	assert.Equal(t, ".bak", installer.opts.BackupSuffix)
	assert.NotNil(t, installer.opts.Executor)
}

func TestInstaller_verifyBinary_NonExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Executable permission test not supported on Windows")
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "claude-loop")

	// Create non-executable file
	err := os.WriteFile(binaryPath, []byte("test"), 0644)
	require.NoError(t, err)

	installer := NewInstaller(&InstallerOptions{
		Executor: &MockInstallerExecutor{VersionCheckSuccess: true},
	})

	// verifyBinary should make it executable
	err = installer.verifyBinary(binaryPath)
	require.NoError(t, err)

	// Check it's now executable
	info, err := os.Stat(binaryPath)
	require.NoError(t, err)
	assert.True(t, info.Mode()&0111 != 0)
}
