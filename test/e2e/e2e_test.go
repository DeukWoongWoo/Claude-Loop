package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce    sync.Once
	cachedBinary string
	buildError   error
)

func skipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		projectRoot, err := findProjectRoot()
		if err != nil {
			buildError = err
			return
		}

		// Use a shared temp directory for the binary
		tempDir := os.TempDir()
		cachedBinary = filepath.Join(tempDir, "claude-loop-e2e-test")

		cmd := exec.Command("go", "build", "-o", cachedBinary, "./cmd/claude-loop")
		cmd.Dir = projectRoot

		output, err := cmd.CombinedOutput()
		if err != nil {
			buildError = &exec.ExitError{Stderr: output}
			cachedBinary = ""
			return
		}
	})

	if buildError != nil {
		t.Fatalf("build failed: %v", buildError)
	}

	return cachedBinary
}

func TestE2E_HelpOutput(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	output, err := exec.Command(binPath, "--help").Output()
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}

	requiredStrings := []string{
		"claude-loop",
		"--prompt",
		"--max-runs",
		"--max-cost",
		"--max-duration",
		"--dry-run",
		"--review-prompt",
		"--worktree",
	}

	outputStr := string(output)
	for _, s := range requiredStrings {
		if !strings.Contains(outputStr, s) {
			t.Errorf("--help output missing expected string: %q", s)
		}
	}
}

func TestE2E_VersionOutput(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	output, err := exec.Command(binPath, "--version").Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	if !strings.Contains(string(output), "claude-loop version") {
		t.Errorf("--version output missing expected version string")
	}
}

func TestE2E_ValidationErrors(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	tests := []struct {
		name        string
		args        []string
		expectError string
	}{
		{"missing prompt", []string{"-m", "5"}, "prompt is required"},
		{"missing limit", []string{"-p", "test"}, "at least one limit"},
		{"invalid merge strategy", []string{"-p", "test", "-m", "1", "--merge-strategy", "invalid"}, "merge-strategy must be"},
		{"invalid duration format", []string{"-p", "test", "--max-duration", "invalid"}, "invalid duration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := exec.Command(binPath, tt.args...).CombinedOutput()
			if err == nil {
				t.Error("expected error but got success")
				return
			}
			if !strings.Contains(string(output), tt.expectError) {
				t.Errorf("expected error containing %q, got: %q", tt.expectError, output)
			}
		})
	}
}

func TestE2E_DryRunMode(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	cmd := exec.Command(binPath,
		"-p", "Test prompt for dry run",
		"-m", "3",
		"--dry-run",
		"--disable-commits",
		"--disable-updates",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	combinedOutput := stdout.String() + stderr.String()

	if err != nil {
		t.Fatalf("dry run failed: %v\noutput: %s", err, combinedOutput)
	}

	// Verify dry-run output format
	if !strings.Contains(combinedOutput, "Loop Complete") {
		t.Error("expected 'Loop Complete' in output")
	}
	if !strings.Contains(combinedOutput, "Stop reason:") {
		t.Error("expected 'Stop reason:' in output")
	}
}

func TestE2E_ListWorktrees(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	output, err := exec.Command(binPath, "--list-worktrees").CombinedOutput()
	outputStr := string(output)

	// Should not require other flags
	if strings.Contains(outputStr, "prompt is required") {
		t.Error("--list-worktrees should not require --prompt")
	}
	if strings.Contains(outputStr, "at least one limit") {
		t.Error("--list-worktrees should not require limit flags")
	}

	// Should show worktree output (if in git repo)
	if err == nil {
		// Success case: should show "Active worktrees" or "No worktrees"
		if !strings.Contains(outputStr, "worktrees") && !strings.Contains(outputStr, "Worktrees") {
			t.Error("expected worktree output")
		}
	}
}

func TestE2E_FlagCombinations(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	// Note: All tests use --dry-run --disable-updates for fast execution.
	// Tests with --max-cost or --max-duration also need -m to prevent infinite loops
	// (in dry-run mode, cost is always $0 and duration limits alone would still run many iterations)
	tests := []struct {
		name string
		args []string
	}{
		{"max-runs only", []string{"-p", "test", "-m", "5", "--dry-run", "--disable-updates"}},
		{"max-cost only", []string{"-p", "test", "--max-cost", "10.00", "-m", "5", "--dry-run", "--disable-updates"}},
		{"max-duration only", []string{"-p", "test", "--max-duration", "1h", "-m", "5", "--dry-run", "--disable-updates"}},
		{"all limits combined", []string{"-p", "test", "-m", "10", "--max-cost", "5.00", "--max-duration", "30m", "--dry-run", "--disable-updates"}},
		{"with reviewer", []string{"-p", "test", "-m", "5", "-r", "run tests", "--dry-run", "--disable-updates"}},
		{"with github options", []string{"-p", "test", "-m", "5", "--owner", "org", "--repo", "repo", "--dry-run", "--disable-updates"}},
		{"with dry-run", []string{"-p", "test", "-m", "5", "--dry-run", "--disable-updates"}},
		{"with disable-commits", []string{"-p", "test", "-m", "5", "--disable-commits", "--dry-run", "--disable-updates"}},
		{"with worktree options", []string{"-p", "test", "-m", "5", "--worktree", "test-wt", "--cleanup-worktree", "--dry-run", "--disable-updates"}},
	}

	validationErrors := []string{
		"prompt is required",
		"at least one limit",
		"merge-strategy must be",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := exec.Command(binPath, tt.args...).CombinedOutput()
			if err != nil {
				outputStr := string(output)
				for _, ve := range validationErrors {
					if strings.Contains(outputStr, ve) {
						t.Errorf("unexpected validation error: %s", ve)
					}
				}
			}
		})
	}
}

func TestE2E_MergeStrategies(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	for _, strategy := range []string{"squash", "merge", "rebase"} {
		t.Run(strategy, func(t *testing.T) {
			output, err := exec.Command(binPath, "-p", "test", "-m", "1", "--merge-strategy", strategy, "--dry-run", "--disable-updates").CombinedOutput()
			if err != nil && strings.Contains(string(output), "merge-strategy must be") {
				t.Errorf("valid merge strategy %q was rejected", strategy)
			}
		})
	}
}

func TestE2E_DurationFormats(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	// Note: -m 3 is used as a safety limit to prevent infinite loops in dry-run mode
	for _, duration := range []string{"30s", "5m", "2h", "1h30m", "2h30m15s"} {
		t.Run(duration, func(t *testing.T) {
			output, err := exec.Command(binPath, "-p", "test", "--max-duration", duration, "-m", "3", "--dry-run", "--disable-updates").CombinedOutput()
			if err != nil && strings.Contains(string(output), "invalid duration") {
				t.Errorf("valid duration %q was rejected", duration)
			}
		})
	}
}

func TestE2E_DryRunStopReasons(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	tests := []struct {
		name       string
		args       []string
		stopReason string
	}{
		{"max_runs", []string{"-p", "test", "-m", "3", "--dry-run", "--disable-updates"}, "max_runs_reached"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := exec.Command(binPath, tt.args...).CombinedOutput()
			if err != nil {
				t.Fatalf("unexpected error: %v\noutput: %s", err, output)
			}
			if !strings.Contains(string(output), tt.stopReason) {
				t.Errorf("expected stop reason %q in output, got: %s", tt.stopReason, output)
			}
		})
	}
}
