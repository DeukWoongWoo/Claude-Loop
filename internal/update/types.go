// Package update provides CLI auto-update functionality.
package update

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// CommandExecutor abstracts exec.Command for testing.
type CommandExecutor interface {
	CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
}

// DefaultExecutor uses the real exec.CommandContext.
type DefaultExecutor struct{}

// CommandContext creates a new exec.Cmd with context.
func (e *DefaultExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// HTTPClient abstracts HTTP operations for testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ReleaseInfo contains information about a GitHub release.
type ReleaseInfo struct {
	Version     string
	TagName     string
	PublishedAt time.Time
	Assets      []Asset
	HTMLURL     string
	Body        string
}

// Asset represents a downloadable binary from a release.
type Asset struct {
	Name        string
	DownloadURL string
	Size        int64
	ContentType string
}

// Platform represents the current OS/architecture.
type Platform struct {
	OS   string
	Arch string
}

// CurrentPlatform returns the current runtime platform.
func CurrentPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// UpdateResult represents the outcome of an update operation.
type UpdateResult struct {
	Updated         bool
	PreviousVersion string
	NewVersion      string
	RestartRequired bool
	Message         string
}

// CheckerOptions configures the update checker.
type CheckerOptions struct {
	RepoOwner      string
	RepoName       string
	CurrentVersion string
	Executor       CommandExecutor
	Timeout        time.Duration // Timeout for GitHub API queries (default: 10s)
}

// DefaultCheckerOptions returns CheckerOptions with defaults.
func DefaultCheckerOptions(currentVersion string) *CheckerOptions {
	return &CheckerOptions{
		RepoOwner:      "DeukWoongWoo",
		RepoName:       "claude-loop",
		CurrentVersion: currentVersion,
		Executor:       &DefaultExecutor{},
		Timeout:        10 * time.Second,
	}
}

// DownloaderOptions configures the binary downloader.
type DownloaderOptions struct {
	TempDir         string
	HTTPClient      HTTPClient
	VerifyChecksum  bool
	OnProgress      func(downloaded, total int64)
	AllowInsecure   bool // For testing only - allows HTTP URLs
}

// defaultHTTPClient is an HTTP client with reasonable timeouts for downloads.
var defaultHTTPClient = &http.Client{
	Timeout: 5 * time.Minute, // 5 minutes for large binary downloads
}

// DefaultDownloaderOptions returns DownloaderOptions with defaults.
func DefaultDownloaderOptions() *DownloaderOptions {
	return &DownloaderOptions{
		TempDir:        os.TempDir(),
		HTTPClient:     defaultHTTPClient,
		VerifyChecksum: true,
	}
}

// InstallerOptions configures the binary installer.
type InstallerOptions struct {
	BinaryPath   string
	BackupSuffix string
	Executor     CommandExecutor
}

// DefaultInstallerOptions returns InstallerOptions with defaults.
func DefaultInstallerOptions() *InstallerOptions {
	return &InstallerOptions{
		BackupSuffix: ".bak",
		Executor:     &DefaultExecutor{},
	}
}

// ManagerOptions configures the update manager.
type ManagerOptions struct {
	AutoUpdate        bool
	DisableUpdates    bool
	CheckerOptions    *CheckerOptions
	DownloaderOptions *DownloaderOptions
	InstallerOptions  *InstallerOptions
	OnPrompt          func(current, latest string) bool
	OnProgress        func(status string)
}

// DefaultManagerOptions returns ManagerOptions with defaults.
func DefaultManagerOptions(currentVersion string) *ManagerOptions {
	return &ManagerOptions{
		CheckerOptions:    DefaultCheckerOptions(currentVersion),
		DownloaderOptions: DefaultDownloaderOptions(),
		InstallerOptions:  DefaultInstallerOptions(),
	}
}

// DownloadResult contains the result of a download operation.
type DownloadResult struct {
	FilePath string
	Checksum string
	Size     int64
}
