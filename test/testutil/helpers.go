// Package testutil provides helper functions for integration and E2E testing.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// WriteFile writes content to a file, creating parent directories as needed.
func WriteFile(t *testing.T, dir, filename, content string) string {
	t.Helper()

	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
	return path
}

// ReadFile reads and returns the content of a file.
func ReadFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	return string(content)
}

// FileExists returns true if the file exists.
func FileExists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	return err == nil
}

// AssertFileExists fails if the file does not exist.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if !FileExists(t, path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

// AssertFileNotExists fails if the file exists.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if FileExists(t, path) {
		t.Errorf("expected file to not exist: %s", path)
	}
}

// GetFixturePath returns the path to a fixture file in test/fixtures.
func GetFixturePath(relativePath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Join("test", "fixtures", relativePath)
	}

	// Walk up to find the project root (where go.mod is)
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "test", "fixtures", relativePath)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return filepath.Join("test", "fixtures", relativePath)
}

// MustGetEnv returns an environment variable or skips the test if not set.
func MustGetEnv(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	if value == "" {
		t.Skipf("skipping: environment variable %s not set", key)
	}
	return value
}

// SetEnv sets an environment variable for the duration of the test.
func SetEnv(t *testing.T, key, value string) {
	t.Helper()

	oldValue := os.Getenv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set environment variable %s: %v", key, err)
	}

	t.Cleanup(func() {
		if oldValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, oldValue)
		}
	})
}

// Chdir changes the working directory for the duration of the test.
func Chdir(t *testing.T, dir string) {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to directory %s: %v", dir, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("warning: failed to restore directory to %s: %v", oldDir, err)
		}
	})
}

// TempDir is a convenience wrapper for t.TempDir().
// Deprecated: Use t.TempDir() directly.
func TempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// GetTestdataPath is an alias for GetFixturePath.
// Deprecated: Use GetFixturePath directly.
func GetTestdataPath(t *testing.T, relativePath string) string {
	t.Helper()
	return GetFixturePath(relativePath)
}
