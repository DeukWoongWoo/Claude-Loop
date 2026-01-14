package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func skipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (go.mod)")
		}
		dir = parent
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "claude-loop-test")
	projectRoot := findProjectRoot(t)

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/claude-loop")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\noutput: %s", err, output)
	}

	return binPath
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
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		combinedOutput := stdout.String() + stderr.String()
		// Skip only if the main loop is not yet implemented
		if strings.Contains(combinedOutput, "not yet implemented") {
			t.Skip("Main loop not yet fully implemented")
		}
		// Fail on unexpected errors to catch regressions
		t.Fatalf("dry run failed with unexpected error: %v\noutput: %s", err, combinedOutput)
	}
}

func TestE2E_ListWorktrees(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	output, err := exec.Command(binPath, "--list-worktrees").CombinedOutput()
	outputStr := string(output)

	// Log error for debugging but don't fail - command may fail for valid reasons
	if err != nil {
		t.Logf("--list-worktrees exited with error (may be expected): %v", err)
	}

	if strings.Contains(outputStr, "prompt is required") {
		t.Error("--list-worktrees should not require --prompt")
	}
	if strings.Contains(outputStr, "at least one limit") {
		t.Error("--list-worktrees should not require limit flags")
	}
}

func TestE2E_FlagCombinations(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	tests := []struct {
		name string
		args []string
	}{
		{"max-runs only", []string{"-p", "test", "-m", "5"}},
		{"max-cost only", []string{"-p", "test", "--max-cost", "10.00"}},
		{"max-duration only", []string{"-p", "test", "--max-duration", "1h"}},
		{"all limits combined", []string{"-p", "test", "-m", "10", "--max-cost", "5.00", "--max-duration", "30m"}},
		{"with reviewer", []string{"-p", "test", "-m", "5", "-r", "run tests"}},
		{"with github options", []string{"-p", "test", "-m", "5", "--owner", "org", "--repo", "repo"}},
		{"with dry-run", []string{"-p", "test", "-m", "5", "--dry-run"}},
		{"with disable-commits", []string{"-p", "test", "-m", "5", "--disable-commits"}},
		{"with worktree options", []string{"-p", "test", "-m", "5", "--worktree", "test-wt", "--cleanup-worktree"}},
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
			output, err := exec.Command(binPath, "-p", "test", "-m", "1", "--merge-strategy", strategy).CombinedOutput()
			if err != nil && strings.Contains(string(output), "merge-strategy must be") {
				t.Errorf("valid merge strategy %q was rejected", strategy)
			}
		})
	}
}

func TestE2E_DurationFormats(t *testing.T) {
	skipIfShort(t)
	binPath := buildBinary(t)

	for _, duration := range []string{"30s", "5m", "2h", "1h30m", "2h30m15s"} {
		t.Run(duration, func(t *testing.T) {
			output, err := exec.Command(binPath, "-p", "test", "--max-duration", duration).CombinedOutput()
			if err != nil && strings.Contains(string(output), "invalid duration") {
				t.Errorf("valid duration %q was rejected", duration)
			}
		})
	}
}
