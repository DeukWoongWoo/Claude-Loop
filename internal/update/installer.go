package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

// Installer handles binary replacement and restart.
type Installer struct {
	opts *InstallerOptions
}

// NewInstaller creates a new Installer with the given options.
func NewInstaller(opts *InstallerOptions) *Installer {
	if opts == nil {
		opts = DefaultInstallerOptions()
	}
	if opts.Executor == nil {
		opts.Executor = &DefaultExecutor{}
	}
	return &Installer{opts: opts}
}

// resolveBinaryPath returns the current binary path, resolving symlinks.
func (i *Installer) resolveBinaryPath(operation string) (string, error) {
	if i.opts.BinaryPath != "" {
		return i.opts.BinaryPath, nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", NewUpdateError(operation, "failed to get current executable path", err)
	}

	resolved, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", NewUpdateError(operation, "failed to resolve symlinks", err)
	}

	return resolved, nil
}

// Install replaces the current binary with the new one.
func (i *Installer) Install(ctx context.Context, newBinaryPath string) error {
	currentPath, err := i.resolveBinaryPath("install")
	if err != nil {
		return err
	}

	if err := i.verifyBinary(newBinaryPath); err != nil {
		return err
	}

	backupPath := currentPath + i.opts.BackupSuffix
	os.Remove(backupPath)

	if runtime.GOOS == "windows" {
		return i.installWindows(ctx, newBinaryPath, currentPath, backupPath)
	}

	if err := os.Rename(currentPath, backupPath); err != nil {
		return NewUpdateError("install", "failed to backup current binary", err)
	}

	if err := os.Rename(newBinaryPath, currentPath); err != nil {
		// Attempt rollback - if this fails, we have a critical error
		if rollbackErr := os.Rename(backupPath, currentPath); rollbackErr != nil {
			return NewUpdateError("install",
				fmt.Sprintf("failed to install new binary and rollback failed: install=%v, rollback=%v", err, rollbackErr),
				err)
		}
		return NewUpdateError("install", "failed to install new binary", err)
	}

	if err := os.Chmod(currentPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set executable permission: %v\n", err)
	}

	os.Remove(backupPath)
	return nil
}

// verifyBinary checks that the new binary is valid and executable.
func (i *Installer) verifyBinary(binaryPath string) error {
	info, err := os.Stat(binaryPath)
	if err != nil {
		return NewUpdateError("install", "new binary not found", err)
	}

	if info.IsDir() {
		return NewUpdateError("install", "new binary path is a directory", nil)
	}

	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return NewUpdateError("install", "new binary is not executable", err)
		}
	}

	cmd := i.opts.Executor.CommandContext(context.Background(), binaryPath, "--version")
	if err := cmd.Run(); err != nil {
		return NewUpdateError("install", "new binary failed version check", err)
	}

	return nil
}

// installWindows handles Windows-specific installation where running binary is locked.
func (i *Installer) installWindows(ctx context.Context, newBinaryPath, currentPath, backupPath string) error {
	scriptPath := filepath.Join(os.TempDir(), "claude-loop-update.bat")
	script := fmt.Sprintf(`@echo off
:loop
timeout /t 1 /nobreak >nul
move /y "%s" "%s" >nul 2>&1
if errorlevel 1 goto loop
move /y "%s" "%s" >nul
del "%s"
`, currentPath, backupPath, newBinaryPath, currentPath, scriptPath)

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return NewUpdateError("install", "failed to create update script", err)
	}

	cmd := exec.Command("cmd", "/c", "start", "/b", scriptPath)
	if err := cmd.Start(); err != nil {
		return NewUpdateError("install", "failed to start update script", err)
	}

	return nil
}

// Restart restarts the application with the same arguments.
func (i *Installer) Restart(ctx context.Context, args []string) error {
	execPath, err := os.Executable()
	if err != nil {
		return NewUpdateError("restart", "failed to get executable path", err)
	}

	if runtime.GOOS != "windows" {
		return syscall.Exec(execPath, append([]string{execPath}, args...), os.Environ())
	}

	cmd := exec.Command(execPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return NewUpdateError("restart", "failed to start new process", err)
	}

	os.Exit(0)
	return nil
}

// Rollback restores the backup binary.
func (i *Installer) Rollback(ctx context.Context) error {
	currentPath, err := i.resolveBinaryPath("rollback")
	if err != nil {
		return err
	}

	backupPath := currentPath + i.opts.BackupSuffix

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return NewUpdateError("rollback", "backup not found", nil)
	}

	if err := os.Remove(currentPath); err != nil && !os.IsNotExist(err) {
		return NewUpdateError("rollback", "failed to remove current binary", err)
	}

	if err := os.Rename(backupPath, currentPath); err != nil {
		return NewUpdateError("rollback", "failed to restore backup", err)
	}

	return nil
}

// GetCurrentBinaryPath returns the path to the currently running binary.
func (i *Installer) GetCurrentBinaryPath() (string, error) {
	path, err := i.resolveBinaryPath("install")
	if err != nil {
		// Fall back to unresolved path if symlink resolution fails
		if i.opts.BinaryPath != "" {
			return i.opts.BinaryPath, nil
		}
		execPath, execErr := os.Executable()
		if execErr != nil {
			return "", err
		}
		return execPath, nil
	}
	return path, nil
}
