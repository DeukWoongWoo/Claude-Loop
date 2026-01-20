package update

import (
	"context"
	"fmt"
)

// Manager orchestrates the update process.
type Manager struct {
	opts       *ManagerOptions
	checker    *Checker
	downloader *Downloader
	installer  *Installer
}

// NewManager creates a new Manager with the given options.
func NewManager(opts *ManagerOptions) *Manager {
	if opts == nil {
		opts = DefaultManagerOptions("v0.0.0")
	}

	return &Manager{
		opts:       opts,
		checker:    NewChecker(opts.CheckerOptions),
		downloader: NewDownloader(opts.DownloaderOptions),
		installer:  NewInstaller(opts.InstallerOptions),
	}
}

// CheckAndUpdate performs the full update flow.
func (m *Manager) CheckAndUpdate(ctx context.Context) (*UpdateResult, error) {
	if m.opts.DisableUpdates {
		return &UpdateResult{
			Updated: false,
			Message: "update checks disabled",
		}, nil
	}

	currentVersion := m.opts.CheckerOptions.CurrentVersion

	// Check for updates
	m.progress("Checking for updates...")
	m.progress(fmt.Sprintf("  Querying %s/%s...",
		m.opts.CheckerOptions.RepoOwner,
		m.opts.CheckerOptions.RepoName))

	release, isNewer, err := m.checker.CheckForUpdate(ctx)
	if err != nil {
		return nil, err
	}
	m.progress("  Done")

	if !isNewer {
		return &UpdateResult{
			Updated:         false,
			PreviousVersion: currentVersion,
			Message:         fmt.Sprintf("Already at latest version %s", currentVersion),
		}, nil
	}

	m.progress(fmt.Sprintf("New version available: %s (current: %s)", release.Version, currentVersion))

	// Require explicit user consent when not in auto-update mode
	if !m.opts.AutoUpdate {
		if m.opts.OnPrompt == nil {
			// No prompt handler and not auto-update - cannot proceed without consent
			return &UpdateResult{
				Updated:         false,
				PreviousVersion: currentVersion,
				Message:         fmt.Sprintf("New version %s available. Run 'claude-loop update' to install.", release.Version),
			}, nil
		}
		if !m.opts.OnPrompt(currentVersion, release.Version) {
			return &UpdateResult{
				Updated:         false,
				PreviousVersion: currentVersion,
				Message:         "Update declined by user",
			}, nil
		}
	}

	platform := CurrentPlatform()
	asset, err := m.checker.FindAssetForPlatform(release, platform)
	if err != nil {
		return nil, err
	}

	var checksums map[string]string
	checksumAsset, err := m.checker.FindChecksumAsset(release)
	if err == nil {
		m.progress("Downloading checksums...")
		checksums, err = m.downloader.DownloadChecksums(ctx, checksumAsset)
		if err != nil {
			// Non-fatal, continue without checksum verification
			m.progress("Warning: could not download checksums, skipping verification")
			checksums = nil
		}
	}

	m.progress(fmt.Sprintf("Downloading %s (%s/%s)...", release.Version, platform.OS, platform.Arch))
	result, err := m.downloader.Download(ctx, asset)
	if err != nil {
		return nil, err
	}

	if m.opts.DownloaderOptions.VerifyChecksum && checksums != nil {
		expectedChecksum, ok := checksums[asset.Name]
		if ok {
			m.progress("Verifying checksum...")
			if err := m.downloader.VerifyChecksum(result.FilePath, expectedChecksum); err != nil {
				m.downloader.Cleanup(result)
				return nil, err
			}
			m.progress("Checksum verified")
		}
	}

	m.progress("Installing update...")
	if err := m.installer.Install(ctx, result.FilePath); err != nil {
		m.downloader.Cleanup(result)
		return nil, err
	}

	m.downloader.Cleanup(result)

	m.progress(fmt.Sprintf("Successfully updated from %s to %s", currentVersion, release.Version))

	return &UpdateResult{
		Updated:         true,
		PreviousVersion: currentVersion,
		NewVersion:      release.Version,
		RestartRequired: true,
		Message:         fmt.Sprintf("Updated to %s. Please restart to use the new version.", release.Version),
	}, nil
}

// CheckOnly checks for updates without installing.
func (m *Manager) CheckOnly(ctx context.Context) (*ReleaseInfo, bool, error) {
	if m.opts.DisableUpdates {
		return nil, false, nil
	}
	return m.checker.CheckForUpdate(ctx)
}

// Update performs an immediate update without prompting.
func (m *Manager) Update(ctx context.Context) (*UpdateResult, error) {
	origAutoUpdate := m.opts.AutoUpdate
	m.opts.AutoUpdate = true
	defer func() { m.opts.AutoUpdate = origAutoUpdate }()

	return m.CheckAndUpdate(ctx)
}

// Restart restarts the application.
func (m *Manager) Restart(ctx context.Context, args []string) error {
	return m.installer.Restart(ctx, args)
}

// Rollback restores the previous version from backup.
func (m *Manager) Rollback(ctx context.Context) error {
	return m.installer.Rollback(ctx)
}

// progress reports progress if callback is set.
func (m *Manager) progress(status string) {
	if m.opts.OnProgress != nil {
		m.opts.OnProgress(status)
	}
}

// GetCurrentVersion returns the current version being managed.
func (m *Manager) GetCurrentVersion() string {
	return m.opts.CheckerOptions.CurrentVersion
}

// IsUpdateDisabled returns whether updates are disabled.
func (m *Manager) IsUpdateDisabled() bool {
	return m.opts.DisableUpdates
}
