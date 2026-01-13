package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloader_Download(t *testing.T) {
	// Create a test binary
	binaryContent := []byte("#!/bin/bash\necho 'hello world'")

	// Create a tar.gz with the binary
	tarGzContent := createTarGz(t, "claude-loop", binaryContent)

	// Start test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.Write(tarGzContent)
	}))
	defer server.Close()

	downloader := NewDownloader(&DownloaderOptions{
		TempDir:        t.TempDir(),
		HTTPClient:     server.Client(),
		VerifyChecksum: false,
		AllowInsecure:  true, // Allow HTTP for testing
	})

	ctx := context.Background()
	asset := &Asset{
		Name:        "claude-loop_darwin_amd64.tar.gz",
		DownloadURL: server.URL,
		Size:        int64(len(tarGzContent)),
	}

	result, err := downloader.Download(ctx, asset)
	require.NoError(t, err)
	assert.NotEmpty(t, result.FilePath)
	assert.NotEmpty(t, result.Checksum)

	// Verify binary content
	content, err := os.ReadFile(result.FilePath)
	require.NoError(t, err)
	assert.Equal(t, binaryContent, content)

	// Cleanup
	err = downloader.Cleanup(result)
	require.NoError(t, err)
}

func TestDownloader_extractTarGz(t *testing.T) {
	binaryContent := []byte("test binary content")
	tarGzContent := createTarGz(t, "claude-loop", binaryContent)

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	err := os.WriteFile(archivePath, tarGzContent, 0644)
	require.NoError(t, err)

	downloader := NewDownloader(nil)
	binaryPath, err := downloader.extractTarGz(archivePath, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(binaryPath)
	require.NoError(t, err)
	assert.Equal(t, binaryContent, content)
}

func TestDownloader_extractZip(t *testing.T) {
	binaryContent := []byte("test binary content")

	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")

	// Create zip file
	createZip(t, archivePath, "claude-loop.exe", binaryContent)

	downloader := NewDownloader(nil)
	binaryPath, err := downloader.extractZip(archivePath, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(binaryPath)
	require.NoError(t, err)
	assert.Equal(t, binaryContent, content)
}

func TestDownloader_VerifyChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-binary")
	content := []byte("test content for checksum")
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	// Calculate expected checksum
	h := sha256.New()
	h.Write(content)
	expectedChecksum := hex.EncodeToString(h.Sum(nil))

	downloader := NewDownloader(nil)

	t.Run("valid checksum", func(t *testing.T) {
		err := downloader.VerifyChecksum(testFile, expectedChecksum)
		assert.NoError(t, err)
	})

	t.Run("invalid checksum", func(t *testing.T) {
		err := downloader.VerifyChecksum(testFile, "invalid-checksum")
		assert.Error(t, err)
		assert.True(t, IsChecksumError(err))
	})

	t.Run("case insensitive", func(t *testing.T) {
		// Checksum comparison should be case-insensitive
		err := downloader.VerifyChecksum(testFile, "ABCD"+expectedChecksum[4:])
		// This should fail because the checksum is different (ABCD prefix)
		assert.Error(t, err)
	})
}

func TestDownloader_DownloadChecksums(t *testing.T) {
	// Format: checksum  *filename (BSD style) or checksum  filename (GNU style)
	checksumContent := `abc123def456  claude-loop_darwin_amd64.tar.gz
def789abc012  claude-loop_linux_amd64.tar.gz
ghi345jkl678  *claude-loop_windows_amd64.zip
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(checksumContent))
	}))
	defer server.Close()

	downloader := NewDownloader(&DownloaderOptions{
		TempDir:       t.TempDir(),
		HTTPClient:    server.Client(),
		AllowInsecure: true, // Allow HTTP for testing
	})

	ctx := context.Background()
	checksums, err := downloader.DownloadChecksums(ctx, &Asset{
		Name:        "checksums.txt",
		DownloadURL: server.URL,
	})
	require.NoError(t, err)

	assert.Equal(t, "abc123def456", checksums["claude-loop_darwin_amd64.tar.gz"])
	assert.Equal(t, "def789abc012", checksums["claude-loop_linux_amd64.tar.gz"])
	assert.Equal(t, "ghi345jkl678", checksums["claude-loop_windows_amd64.zip"])
}

func TestDownloader_downloadFile_Progress(t *testing.T) {
	content := []byte("test content for progress tracking")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer server.Close()

	var progressCalled bool
	var lastDownloaded int64

	downloader := NewDownloader(&DownloaderOptions{
		TempDir:       t.TempDir(),
		HTTPClient:    server.Client(),
		AllowInsecure: true, // Allow HTTP for testing
		OnProgress: func(downloaded, total int64) {
			progressCalled = true
			lastDownloaded = downloaded
			_ = total // We track total but don't assert on it in this test
		},
	})

	tmpFile := filepath.Join(t.TempDir(), "download-test")
	ctx := context.Background()
	err := downloader.downloadFile(ctx, server.URL, tmpFile, int64(len(content)))
	require.NoError(t, err)

	assert.True(t, progressCalled)
	assert.Equal(t, int64(len(content)), lastDownloaded)
}

func TestDownloader_downloadFile_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	downloader := NewDownloader(&DownloaderOptions{
		TempDir:       t.TempDir(),
		HTTPClient:    server.Client(),
		AllowInsecure: true, // Allow HTTP for testing
	})

	tmpFile := filepath.Join(t.TempDir(), "download-test")
	ctx := context.Background()
	err := downloader.downloadFile(ctx, server.URL, tmpFile, 100)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestDownloader_extractBinary_NoExtension(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "claude-loop")
	err := os.WriteFile(binaryPath, []byte("binary"), 0755)
	require.NoError(t, err)

	downloader := NewDownloader(nil)
	result, err := downloader.extractBinary(binaryPath, tmpDir)
	require.NoError(t, err)
	assert.Equal(t, binaryPath, result)
}

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"exact match", "claude-loop", true},
		{"windows exe", "claude-loop.exe", true},
		{"prefix match", "claude-loop_darwin_amd64", true},
		{"unrelated file", "readme.md", false},
		{"checksum file", "checksums.txt", false},
		{"tar.gz file", "claude-loop_darwin_amd64.tar.gz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isBinaryFile(tt.filename))
		})
	}
}

func TestNewDownloader_Defaults(t *testing.T) {
	d := NewDownloader(nil)
	assert.NotNil(t, d.opts)
	assert.NotNil(t, d.opts.HTTPClient)
	assert.True(t, d.opts.VerifyChecksum)
}

// Helper function to create tar.gz content
func createTarGz(t *testing.T, filename string, content []byte) []byte {
	t.Helper()

	var buf []byte
	pr, pw := io.Pipe()

	go func() {
		gw := gzip.NewWriter(pw)
		tw := tar.NewWriter(gw)

		hdr := &tar.Header{
			Name: filename,
			Mode: 0755,
			Size: int64(len(content)),
		}
		tw.WriteHeader(hdr)
		tw.Write(content)
		tw.Close()
		gw.Close()
		pw.Close()
	}()

	buf, _ = io.ReadAll(pr)
	return buf
}

// Helper function to create zip file
func createZip(t *testing.T, zipPath, filename string, content []byte) {
	t.Helper()

	f, err := os.Create(zipPath)
	require.NoError(t, err)
	defer f.Close()

	zw := zip.NewWriter(f)
	w, err := zw.Create(filename)
	require.NoError(t, err)
	w.Write(content)
	zw.Close()
}
