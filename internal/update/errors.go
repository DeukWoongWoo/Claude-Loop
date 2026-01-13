package update

import (
	"errors"
	"fmt"
)

// UpdateError represents an update operation error.
type UpdateError struct {
	Operation string // Failed operation (e.g., "check", "download", "install")
	Message   string // Error message
	Err       error  // Underlying error
}

func (e *UpdateError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("update %s: %s: %v", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("update %s: %s", e.Operation, e.Message)
}

func (e *UpdateError) Unwrap() error {
	return e.Err
}

// VersionError represents a version comparison/parsing error.
type VersionError struct {
	Version string
	Message string
	Err     error
}

func (e *VersionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("version %q: %s: %v", e.Version, e.Message, e.Err)
	}
	return fmt.Sprintf("version %q: %s", e.Version, e.Message)
}

func (e *VersionError) Unwrap() error {
	return e.Err
}

// ChecksumError represents a checksum verification failure.
type ChecksumError struct {
	Expected string
	Actual   string
	File     string
}

func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum mismatch for %s: expected %s, got %s", e.File, e.Expected, e.Actual)
}

// Predefined errors for common conditions.
var (
	ErrNoUpdateAvailable   = errors.New("already at latest version")
	ErrAssetNotFound       = errors.New("no compatible binary found for platform")
	ErrChecksumNotFound    = errors.New("checksum file not found")
	ErrChecksumMismatch    = errors.New("checksum verification failed")
	ErrDownloadFailed      = errors.New("failed to download binary")
	ErrDownloadInterrupted = errors.New("download interrupted")
	ErrInstallFailed       = errors.New("failed to install binary")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrRestartFailed       = errors.New("failed to restart")
	ErrBinaryNotFound      = errors.New("binary not found in archive")
	ErrNoReleasesFound     = errors.New("no releases found")
)

// IsUpdateError checks if an error is an UpdateError.
func IsUpdateError(err error) bool {
	var ue *UpdateError
	return errors.As(err, &ue)
}

// IsVersionError checks if an error is a VersionError.
func IsVersionError(err error) bool {
	var ve *VersionError
	return errors.As(err, &ve)
}

// IsChecksumError checks if an error is a ChecksumError.
func IsChecksumError(err error) bool {
	var ce *ChecksumError
	return errors.As(err, &ce)
}

// NewUpdateError creates a new UpdateError.
func NewUpdateError(operation, message string, err error) *UpdateError {
	return &UpdateError{
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// NewVersionError creates a new VersionError.
func NewVersionError(version, message string, err error) *VersionError {
	return &VersionError{
		Version: version,
		Message: message,
		Err:     err,
	}
}
