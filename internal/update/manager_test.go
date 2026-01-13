package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_CheckAndUpdate_DisabledUpdates(t *testing.T) {
	manager := NewManager(&ManagerOptions{
		DisableUpdates: true,
	})

	ctx := context.Background()
	result, err := manager.CheckAndUpdate(ctx)
	require.NoError(t, err)
	assert.False(t, result.Updated)
	assert.Contains(t, result.Message, "disabled")
}

func TestManager_CheckAndUpdate_AlreadyLatest(t *testing.T) {
	releaseJSON := `{
		"tagName": "v0.18.0",
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

	manager := NewManager(&ManagerOptions{
		CheckerOptions: &CheckerOptions{
			RepoOwner:      "DeukWoongWoo",
			RepoName:       "claude-loop",
			CurrentVersion: "v0.18.0", // Same as release version
			Executor:       mock,
		},
	})

	ctx := context.Background()
	result, err := manager.CheckAndUpdate(ctx)
	require.NoError(t, err)
	assert.False(t, result.Updated)
	assert.Contains(t, result.Message, "latest")
}

func TestManager_CheckAndUpdate_UserDeclines(t *testing.T) {
	releaseJSON := `{
		"tagName": "v0.19.0",
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

	manager := NewManager(&ManagerOptions{
		AutoUpdate: false,
		OnPrompt: func(current, latest string) bool {
			return false // User declines
		},
		CheckerOptions: &CheckerOptions{
			RepoOwner:      "DeukWoongWoo",
			RepoName:       "claude-loop",
			CurrentVersion: "v0.18.0",
			Executor:       mock,
		},
	})

	ctx := context.Background()
	result, err := manager.CheckAndUpdate(ctx)
	require.NoError(t, err)
	assert.False(t, result.Updated)
	assert.Contains(t, result.Message, "declined")
}

func TestManager_CheckAndUpdate_FullFlow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Full flow test not supported on Windows")
	}

	// Create test binary content
	binaryContent := []byte("#!/bin/bash\necho 'v0.19.0'")
	tarGzContent := createTarGz(t, "claude-loop", binaryContent)

	// Create HTTP server for binary download
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.Write(tarGzContent)
	}))
	defer server.Close()

	releaseJSON := fmt.Sprintf(`{
		"tagName": "v0.19.0",
		"publishedAt": "2026-01-12T10:00:00Z",
		"body": "Release notes",
		"url": "https://github.com/test",
		"assets": [
			{
				"name": "claude-loop_%s_%s.tar.gz",
				"url": "%s",
				"size": %d,
				"contentType": "application/gzip"
			}
		]
	}`, runtime.GOOS, runtime.GOARCH, server.URL, len(tarGzContent))

	// Create current binary
	tmpDir := t.TempDir()
	currentBinary := filepath.Join(tmpDir, "claude-loop")
	err := os.WriteFile(currentBinary, []byte("#!/bin/bash\necho 'v0.18.0'"), 0755)
	require.NoError(t, err)

	ghMock := &MockExecutor{
		Commands: []MockCommand{
			{Stdout: releaseJSON},
		},
	}

	var progressMessages []string
	manager := NewManager(&ManagerOptions{
		AutoUpdate: true,
		OnProgress: func(status string) {
			progressMessages = append(progressMessages, status)
		},
		CheckerOptions: &CheckerOptions{
			RepoOwner:      "DeukWoongWoo",
			RepoName:       "claude-loop",
			CurrentVersion: "v0.18.0",
			Executor:       ghMock,
		},
		DownloaderOptions: &DownloaderOptions{
			TempDir:        tmpDir,
			HTTPClient:     server.Client(),
			VerifyChecksum: false, // No checksum in this test
			AllowInsecure:  true,  // Allow HTTP for testing
		},
		InstallerOptions: &InstallerOptions{
			BinaryPath:   currentBinary,
			BackupSuffix: ".bak",
			Executor:     &MockInstallerExecutor{VersionCheckSuccess: true},
		},
	})

	ctx := context.Background()
	result, err := manager.CheckAndUpdate(ctx)
	require.NoError(t, err)
	assert.True(t, result.Updated)
	assert.Equal(t, "v0.18.0", result.PreviousVersion)
	assert.Equal(t, "v0.19.0", result.NewVersion)
	assert.True(t, result.RestartRequired)

	// Verify progress was reported
	assert.NotEmpty(t, progressMessages)
	assert.Contains(t, progressMessages[0], "Checking")
}

func TestManager_CheckOnly(t *testing.T) {
	releaseJSON := `{
		"tagName": "v0.19.0",
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

	manager := NewManager(&ManagerOptions{
		CheckerOptions: &CheckerOptions{
			RepoOwner:      "DeukWoongWoo",
			RepoName:       "claude-loop",
			CurrentVersion: "v0.18.0",
			Executor:       mock,
		},
	})

	ctx := context.Background()
	release, isNewer, err := manager.CheckOnly(ctx)
	require.NoError(t, err)
	assert.True(t, isNewer)
	assert.Equal(t, "v0.19.0", release.Version)
}

func TestManager_CheckOnly_Disabled(t *testing.T) {
	manager := NewManager(&ManagerOptions{
		DisableUpdates: true,
	})

	ctx := context.Background()
	release, isNewer, err := manager.CheckOnly(ctx)
	require.NoError(t, err)
	assert.False(t, isNewer)
	assert.Nil(t, release)
}

func TestManager_GetCurrentVersion(t *testing.T) {
	manager := NewManager(&ManagerOptions{
		CheckerOptions: &CheckerOptions{
			CurrentVersion: "v1.2.3",
		},
	})

	assert.Equal(t, "v1.2.3", manager.GetCurrentVersion())
}

func TestManager_IsUpdateDisabled(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		manager := NewManager(&ManagerOptions{
			DisableUpdates: false,
		})
		assert.False(t, manager.IsUpdateDisabled())
	})

	t.Run("disabled", func(t *testing.T) {
		manager := NewManager(&ManagerOptions{
			DisableUpdates: true,
		})
		assert.True(t, manager.IsUpdateDisabled())
	})
}

func TestNewManager_Defaults(t *testing.T) {
	manager := NewManager(nil)
	assert.NotNil(t, manager.opts)
	assert.NotNil(t, manager.checker)
	assert.NotNil(t, manager.downloader)
	assert.NotNil(t, manager.installer)
}

func TestManager_Update_ForcesAutoUpdate(t *testing.T) {
	releaseJSON := `{
		"tagName": "v0.18.0",
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

	promptCalled := false
	manager := NewManager(&ManagerOptions{
		AutoUpdate: false,
		OnPrompt: func(current, latest string) bool {
			promptCalled = true
			return false
		},
		CheckerOptions: &CheckerOptions{
			RepoOwner:      "DeukWoongWoo",
			RepoName:       "claude-loop",
			CurrentVersion: "v0.18.0", // Same version - no update needed
			Executor:       mock,
		},
	})

	ctx := context.Background()
	result, err := manager.Update(ctx)
	require.NoError(t, err)

	// Should not prompt because Update forces auto-update
	assert.False(t, promptCalled)
	// But no update available anyway
	assert.False(t, result.Updated)
}
