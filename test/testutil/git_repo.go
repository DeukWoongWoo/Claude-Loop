package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// GitRepoOptions configures the test git repository.
type GitRepoOptions struct {
	InitialCommit bool
	RemoteURL     string
	BranchName    string
}

// DefaultGitRepoOptions returns default options with an initial commit on main.
func DefaultGitRepoOptions() GitRepoOptions {
	return GitRepoOptions{
		InitialCommit: true,
		BranchName:    "main",
	}
}

// SetupGitRepo creates a temporary git repository and returns its path.
func SetupGitRepo(t *testing.T, opts GitRepoOptions) string {
	t.Helper()

	dir := t.TempDir()
	initGitRepo(t, dir, opts)
	return dir
}

// SetupGitRepoWithGitHub creates a test repo with a GitHub remote.
func SetupGitRepoWithGitHub(t *testing.T, owner, repo string) string {
	t.Helper()

	opts := DefaultGitRepoOptions()
	opts.RemoteURL = "https://github.com/" + owner + "/" + repo + ".git"
	return SetupGitRepo(t, opts)
}

func initGitRepo(t *testing.T, dir string, opts GitRepoOptions) {
	t.Helper()

	runGit(t, dir, "init")

	if opts.BranchName != "" {
		runGit(t, dir, "checkout", "-b", opts.BranchName)
	}

	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	if opts.InitialCommit {
		path := filepath.Join(dir, "README.md")
		if err := os.WriteFile(path, []byte("# Test Repository\n"), 0644); err != nil {
			t.Fatalf("failed to create README.md: %v", err)
		}
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-m", "Initial commit")
	}

	if opts.RemoteURL != "" {
		runGit(t, dir, "remote", "add", "origin", opts.RemoteURL)
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\noutput: %s", args, err, output)
	}
	return string(output)
}

// RunGitSafe runs a git command and returns an error instead of failing.
func RunGitSafe(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// GitCommit creates an empty commit with the given message.
func GitCommit(t *testing.T, dir, message string) {
	t.Helper()
	runGit(t, dir, "commit", "--allow-empty", "-m", message)
}

// GitBranch creates and checks out a new branch.
func GitBranch(t *testing.T, dir, branchName string) {
	t.Helper()
	runGit(t, dir, "checkout", "-b", branchName)
}

// GitCheckout switches to an existing branch.
func GitCheckout(t *testing.T, dir, branchName string) {
	t.Helper()
	runGit(t, dir, "checkout", branchName)
}

// GitCurrentBranch returns the current branch name.
func GitCurrentBranch(t *testing.T, dir string) string {
	t.Helper()
	output := runGit(t, dir, "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(output)
}

// GitAddFile creates a file and stages it.
func GitAddFile(t *testing.T, dir, filename, content string) {
	t.Helper()

	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	runGit(t, dir, "add", filename)
}

// GitStatus returns the porcelain status output.
func GitStatus(t *testing.T, dir string) string {
	t.Helper()
	return runGit(t, dir, "status", "--porcelain")
}

// GitLog returns the last n commits in oneline format.
func GitLog(t *testing.T, dir string, n int) string {
	t.Helper()
	return runGit(t, dir, "log", "--oneline", "-n", strconv.Itoa(n))
}
