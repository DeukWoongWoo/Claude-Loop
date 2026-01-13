package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Downloader handles binary downloads and verification.
type Downloader struct {
	opts *DownloaderOptions
}

// NewDownloader creates a new Downloader with the given options.
func NewDownloader(opts *DownloaderOptions) *Downloader {
	if opts == nil {
		opts = DefaultDownloaderOptions()
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}
	return &Downloader{opts: opts}
}

// Download downloads and extracts the binary from the given asset.
func (d *Downloader) Download(ctx context.Context, asset *Asset) (*DownloadResult, error) {
	downloadDir, err := os.MkdirTemp(d.opts.TempDir, "claude-loop-update-*")
	if err != nil {
		return nil, NewUpdateError("download", "failed to create temp directory", err)
	}

	var success bool
	defer func() {
		if !success {
			os.RemoveAll(downloadDir)
		}
	}()

	// Sanitize asset name to prevent path traversal attacks
	safeName := filepath.Base(asset.Name)
	archivePath := filepath.Join(downloadDir, safeName)
	if err := d.downloadFile(ctx, asset.DownloadURL, archivePath, asset.Size); err != nil {
		return nil, err
	}

	binaryPath, err := d.extractBinary(archivePath, downloadDir)
	if err != nil {
		return nil, err
	}

	checksum, err := d.calculateChecksum(binaryPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		return nil, NewUpdateError("download", "failed to stat binary", err)
	}

	success = true
	return &DownloadResult{
		FilePath: binaryPath,
		Checksum: checksum,
		Size:     info.Size(),
	}, nil
}

// downloadFile downloads a file with progress reporting.
func (d *Downloader) downloadFile(ctx context.Context, url, destPath string, expectedSize int64) error {
	// Validate URL scheme for security (skip for testing with AllowInsecure)
	if !d.opts.AllowInsecure && !strings.HasPrefix(url, "https://") {
		return NewUpdateError("download", "only HTTPS URLs are allowed for security", nil)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return NewUpdateError("download", "failed to create request", err)
	}

	resp, err := d.opts.HTTPClient.Do(req)
	if err != nil {
		return NewUpdateError("download", "failed to download file", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewUpdateError("download",
			fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status), nil)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return NewUpdateError("download", "failed to create file", err)
	}
	defer out.Close()

	var reader io.Reader = resp.Body
	if d.opts.OnProgress != nil {
		reader = &progressReader{
			reader:     resp.Body,
			total:      expectedSize,
			onProgress: d.opts.OnProgress,
		}
	}

	_, err = io.Copy(out, reader)
	if err != nil {
		if ctx.Err() != nil {
			return NewUpdateError("download", "download interrupted", ErrDownloadInterrupted)
		}
		return NewUpdateError("download", "failed to write file", err)
	}

	return nil
}

// progressReader wraps a reader and reports progress.
type progressReader struct {
	reader     io.Reader
	downloaded int64
	total      int64
	onProgress func(downloaded, total int64)
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.downloaded += int64(n)
	if r.onProgress != nil {
		r.onProgress(r.downloaded, r.total)
	}
	return n, err
}

// extractBinary extracts the binary from an archive.
func (d *Downloader) extractBinary(archivePath, destDir string) (string, error) {
	lowerPath := strings.ToLower(archivePath)

	if strings.HasSuffix(lowerPath, ".tar.gz") {
		return d.extractTarGz(archivePath, destDir)
	}

	switch strings.ToLower(filepath.Ext(archivePath)) {
	case ".zip":
		return d.extractZip(archivePath, destDir)
	case ".gz":
		return d.extractGzip(archivePath, destDir)
	default:
		return archivePath, nil
	}
}

// extractTarGz extracts a .tar.gz archive.
func (d *Downloader) extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", NewUpdateError("extract", "failed to open archive", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", NewUpdateError("extract", "failed to create gzip reader", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var binaryPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", NewUpdateError("extract", "failed to read tar header", err)
		}

		// Look for the binary (executable file, not a directory)
		if header.Typeflag == tar.TypeReg {
			name := filepath.Base(header.Name)
			if isBinaryFile(name) {
				binaryPath = filepath.Join(destDir, name)
				outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
				if err != nil {
					return "", NewUpdateError("extract", "failed to create binary file", err)
				}
				if _, err := io.Copy(outFile, tr); err != nil {
					outFile.Close()
					return "", NewUpdateError("extract", "failed to write binary", err)
				}
				outFile.Close()
			}
		}
	}

	if binaryPath == "" {
		return "", NewUpdateError("extract", "binary not found in archive", ErrBinaryNotFound)
	}

	return binaryPath, nil
}

// extractZip extracts a .zip archive (for Windows).
func (d *Downloader) extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", NewUpdateError("extract", "failed to open zip archive", err)
	}
	defer r.Close()

	var binaryPath string

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if isBinaryFile(name) {
			binaryPath = filepath.Join(destDir, name)

			rc, err := f.Open()
			if err != nil {
				return "", NewUpdateError("extract", "failed to open file in archive", err)
			}

			outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY, f.Mode())
			if err != nil {
				rc.Close()
				return "", NewUpdateError("extract", "failed to create binary file", err)
			}

			_, err = io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()

			if err != nil {
				return "", NewUpdateError("extract", "failed to write binary", err)
			}
			break
		}
	}

	if binaryPath == "" {
		return "", NewUpdateError("extract", "binary not found in archive", ErrBinaryNotFound)
	}

	return binaryPath, nil
}

// extractGzip extracts a .gz file.
func (d *Downloader) extractGzip(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", NewUpdateError("extract", "failed to open gzip file", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", NewUpdateError("extract", "failed to create gzip reader", err)
	}
	defer gzr.Close()

	binaryPath := filepath.Join(destDir, "claude-loop")
	outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return "", NewUpdateError("extract", "failed to create binary file", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, gzr); err != nil {
		return "", NewUpdateError("extract", "failed to write binary", err)
	}

	return binaryPath, nil
}

// isBinaryFile checks if a filename looks like our binary.
func isBinaryFile(name string) bool {
	// Match claude-loop or claude-loop.exe
	return name == "claude-loop" || name == "claude-loop.exe" ||
		strings.HasPrefix(name, "claude-loop_") && !strings.Contains(name, ".")
}

// calculateChecksum calculates SHA256 checksum of a file.
func (d *Downloader) calculateChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", NewUpdateError("verify", "failed to open file for checksum", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", NewUpdateError("verify", "failed to calculate checksum", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyChecksum verifies a file against a checksum.
func (d *Downloader) VerifyChecksum(filePath, expectedChecksum string) error {
	actualChecksum, err := d.calculateChecksum(filePath)
	if err != nil {
		return err
	}

	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return &ChecksumError{
			Expected: expectedChecksum,
			Actual:   actualChecksum,
			File:     filePath,
		}
	}

	return nil
}

// DownloadChecksums downloads and parses the checksums file.
func (d *Downloader) DownloadChecksums(ctx context.Context, asset *Asset) (map[string]string, error) {
	// Create temp file
	tmpFile, err := os.CreateTemp(d.opts.TempDir, "checksums-*")
	if err != nil {
		return nil, NewUpdateError("download", "failed to create temp file", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := d.downloadFile(ctx, asset.DownloadURL, tmpPath, asset.Size); err != nil {
		return nil, err
	}

	// Parse checksums file
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, NewUpdateError("download", "failed to read checksums file", err)
	}

	checksums := make(map[string]string)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			filename := strings.TrimPrefix(parts[1], "*") // Remove leading * if present
			checksums[filename] = checksum
		}
	}

	return checksums, nil
}

// Cleanup removes downloaded files.
func (d *Downloader) Cleanup(result *DownloadResult) error {
	if result == nil || result.FilePath == "" {
		return nil
	}
	// Remove the parent temp directory
	dir := filepath.Dir(result.FilePath)
	return os.RemoveAll(dir)
}
