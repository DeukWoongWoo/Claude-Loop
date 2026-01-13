package update

import (
	"bytes"
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCommand represents a mock command response.
type MockCommand struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// MockExecutor is a test double for CommandExecutor.
type MockExecutor struct {
	Commands []MockCommand
	Index    int
	Called   [][]string // Records all command calls
}

func (m *MockExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	m.Called = append(m.Called, append([]string{name}, args...))

	var mockCmd MockCommand
	if m.Index < len(m.Commands) {
		mockCmd = m.Commands[m.Index]
		m.Index++
	}

	// Create a command that echoes our mock output
	if mockCmd.ExitCode != 0 {
		cmd := exec.CommandContext(ctx, "sh", "-c", "echo '"+mockCmd.Stderr+"' >&2; exit "+string(rune('0'+mockCmd.ExitCode)))
		return cmd
	}

	cmd := exec.CommandContext(ctx, "echo", "-n", mockCmd.Stdout)
	return cmd
}

// MockCmd is a simpler mock that just returns pre-set stdout/stderr.
type MockCmd struct {
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func TestChecker_GetLatestRelease(t *testing.T) {
	releaseJSON := `{
		"tagName": "v0.19.0",
		"publishedAt": "2026-01-12T10:00:00Z",
		"body": "Release notes here",
		"url": "https://github.com/DeukWoongWoo/claude-loop/releases/tag/v0.19.0",
		"assets": [
			{
				"name": "claude-loop_darwin_amd64.tar.gz",
				"url": "https://github.com/download/darwin_amd64.tar.gz",
				"size": 1024000,
				"contentType": "application/gzip"
			},
			{
				"name": "claude-loop_darwin_arm64.tar.gz",
				"url": "https://github.com/download/darwin_arm64.tar.gz",
				"size": 1024000,
				"contentType": "application/gzip"
			},
			{
				"name": "checksums.txt",
				"url": "https://github.com/download/checksums.txt",
				"size": 256,
				"contentType": "text/plain"
			}
		]
	}`

	mock := &MockExecutor{
		Commands: []MockCommand{
			{Stdout: releaseJSON},
		},
	}

	checker := NewChecker(&CheckerOptions{
		RepoOwner:      "DeukWoongWoo",
		RepoName:       "claude-loop",
		CurrentVersion: "v0.18.0",
		Executor:       mock,
	})

	ctx := context.Background()
	release, err := checker.GetLatestRelease(ctx)
	require.NoError(t, err)

	assert.Equal(t, "v0.19.0", release.Version)
	assert.Equal(t, "v0.19.0", release.TagName)
	assert.Len(t, release.Assets, 3)
	assert.Equal(t, "claude-loop_darwin_amd64.tar.gz", release.Assets[0].Name)
}

func TestChecker_CheckForUpdate(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		releaseVersion string
		wantNewer      bool
	}{
		{"newer version available", "v0.18.0", "v0.19.0", true},
		{"same version", "v0.18.0", "v0.18.0", false},
		{"older version on server", "v0.19.0", "v0.18.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseJSON := `{
				"tagName": "` + tt.releaseVersion + `",
				"publishedAt": "2026-01-12T10:00:00Z",
				"body": "Release notes",
				"url": "https://github.com/test",
				"assets": []
			}`

			mock := &MockExecutor{
				Commands: []MockCommand{
					{Stdout: releaseJSON},
				},
			}

			checker := NewChecker(&CheckerOptions{
				RepoOwner:      "DeukWoongWoo",
				RepoName:       "claude-loop",
				CurrentVersion: tt.currentVersion,
				Executor:       mock,
			})

			ctx := context.Background()
			release, isNewer, err := checker.CheckForUpdate(ctx)
			require.NoError(t, err)

			assert.Equal(t, tt.releaseVersion, release.Version)
			assert.Equal(t, tt.wantNewer, isNewer)
		})
	}
}

func TestChecker_FindAssetForPlatform(t *testing.T) {
	release := &ReleaseInfo{
		Version: "v0.19.0",
		Assets: []Asset{
			{Name: "claude-loop_darwin_amd64.tar.gz", DownloadURL: "https://example.com/darwin_amd64.tar.gz"},
			{Name: "claude-loop_darwin_arm64.tar.gz", DownloadURL: "https://example.com/darwin_arm64.tar.gz"},
			{Name: "claude-loop_linux_amd64.tar.gz", DownloadURL: "https://example.com/linux_amd64.tar.gz"},
			{Name: "claude-loop_linux_arm64.tar.gz", DownloadURL: "https://example.com/linux_arm64.tar.gz"},
			{Name: "claude-loop_windows_amd64.zip", DownloadURL: "https://example.com/windows_amd64.zip"},
			{Name: "checksums.txt", DownloadURL: "https://example.com/checksums.txt"},
		},
	}

	checker := NewChecker(nil)

	tests := []struct {
		name     string
		platform Platform
		wantName string
		wantErr  bool
	}{
		{"darwin amd64", Platform{OS: "darwin", Arch: "amd64"}, "claude-loop_darwin_amd64.tar.gz", false},
		{"darwin arm64", Platform{OS: "darwin", Arch: "arm64"}, "claude-loop_darwin_arm64.tar.gz", false},
		{"linux amd64", Platform{OS: "linux", Arch: "amd64"}, "claude-loop_linux_amd64.tar.gz", false},
		{"linux arm64", Platform{OS: "linux", Arch: "arm64"}, "claude-loop_linux_arm64.tar.gz", false},
		{"windows amd64", Platform{OS: "windows", Arch: "amd64"}, "claude-loop_windows_amd64.zip", false},
		{"unsupported platform", Platform{OS: "freebsd", Arch: "386"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, err := checker.FindAssetForPlatform(release, tt.platform)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrAssetNotFound)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, asset.Name)
		})
	}
}

func TestChecker_FindChecksumAsset(t *testing.T) {
	tests := []struct {
		name     string
		assets   []Asset
		wantName string
		wantErr  bool
	}{
		{
			name: "find checksums.txt",
			assets: []Asset{
				{Name: "binary.tar.gz"},
				{Name: "checksums.txt"},
			},
			wantName: "checksums.txt",
		},
		{
			name: "find SHA256SUMS",
			assets: []Asset{
				{Name: "binary.tar.gz"},
				{Name: "SHA256SUMS"},
			},
			wantName: "SHA256SUMS",
		},
		{
			name: "no checksum file",
			assets: []Asset{
				{Name: "binary.tar.gz"},
				{Name: "readme.md"},
			},
			wantErr: true,
		},
	}

	checker := NewChecker(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release := &ReleaseInfo{Assets: tt.assets}
			asset, err := checker.FindChecksumAsset(release)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrChecksumNotFound)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, asset.Name)
		})
	}
}

func TestNewChecker_Defaults(t *testing.T) {
	checker := NewChecker(nil)
	assert.NotNil(t, checker.opts)
	assert.Equal(t, "DeukWoongWoo", checker.opts.RepoOwner)
	assert.Equal(t, "claude-loop", checker.opts.RepoName)
	assert.NotNil(t, checker.opts.Executor)
}
