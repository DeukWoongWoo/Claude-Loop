package update

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Checker queries GitHub for the latest release.
type Checker struct {
	opts *CheckerOptions
}

// NewChecker creates a new Checker with the given options.
func NewChecker(opts *CheckerOptions) *Checker {
	if opts == nil {
		opts = DefaultCheckerOptions("v0.0.0")
	}
	if opts.Executor == nil {
		opts.Executor = &DefaultExecutor{}
	}
	return &Checker{opts: opts}
}

// ghReleaseResponse represents the JSON from gh release view.
type ghReleaseResponse struct {
	TagName     string    `json:"tagName"`
	PublishedAt time.Time `json:"publishedAt"`
	Body        string    `json:"body"`
	URL         string    `json:"url"`
	Assets      []struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Size        int64  `json:"size"`
		ContentType string `json:"contentType"`
	} `json:"assets"`
}

// GetLatestRelease retrieves the latest release from GitHub.
func (c *Checker) GetLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	repoSlug := c.opts.RepoOwner + "/" + c.opts.RepoName

	// Use gh CLI to get latest release
	cmd := c.opts.Executor.CommandContext(ctx, "gh", "release", "view",
		"--repo", repoSlug,
		"--json", "tagName,publishedAt,body,url,assets",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if strings.Contains(stderrStr, "release not found") ||
			strings.Contains(stderrStr, "no releases") {
			return nil, NewUpdateError("check", "no releases found", ErrNoReleasesFound)
		}
		return nil, NewUpdateError("check", "failed to query GitHub releases", err)
	}

	var resp ghReleaseResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return nil, NewUpdateError("check", "failed to parse release response", err)
	}

	release := &ReleaseInfo{
		Version:     resp.TagName,
		TagName:     resp.TagName,
		PublishedAt: resp.PublishedAt,
		HTMLURL:     resp.URL,
		Body:        resp.Body,
		Assets:      make([]Asset, 0, len(resp.Assets)),
	}

	for _, a := range resp.Assets {
		release.Assets = append(release.Assets, Asset{
			Name:        a.Name,
			DownloadURL: a.URL,
			Size:        a.Size,
			ContentType: a.ContentType,
		})
	}

	return release, nil
}

// CheckForUpdate checks if a newer version is available.
func (c *Checker) CheckForUpdate(ctx context.Context) (*ReleaseInfo, bool, error) {
	release, err := c.GetLatestRelease(ctx)
	if err != nil {
		return nil, false, err
	}

	isNewer, err := IsNewer(c.opts.CurrentVersion, release.Version)
	if err != nil {
		return nil, false, err
	}

	return release, isNewer, nil
}

// FindAssetForPlatform finds the appropriate binary asset for the platform.
func (c *Checker) FindAssetForPlatform(release *ReleaseInfo, platform Platform) (*Asset, error) {
	osName := strings.ToLower(platform.OS)
	arch := strings.ToLower(platform.Arch)
	prefix := fmt.Sprintf("claude-loop_%s_%s", osName, arch)

	for _, asset := range release.Assets {
		assetName := strings.ToLower(asset.Name)
		if strings.HasPrefix(assetName, prefix) {
			return &asset, nil
		}
	}

	return nil, NewUpdateError("check",
		fmt.Sprintf("no binary found for %s/%s", platform.OS, platform.Arch),
		ErrAssetNotFound)
}

// FindChecksumAsset finds the checksum file (SHA256SUMS or similar).
func (c *Checker) FindChecksumAsset(release *ReleaseInfo) (*Asset, error) {
	checksumPatterns := []string{
		"checksums.txt",
		"sha256sums",
		"sha256sums.txt",
	}

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		for _, pattern := range checksumPatterns {
			if strings.Contains(name, pattern) {
				return &asset, nil
			}
		}
	}

	return nil, NewUpdateError("check", "checksum file not found", ErrChecksumNotFound)
}
