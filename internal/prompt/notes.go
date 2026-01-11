package prompt

import (
	"fmt"
	"os"
)

// NotesLoader handles loading notes files.
type NotesLoader interface {
	// Load reads the notes file content.
	// Returns the content, whether the file exists, and any error.
	// If the file doesn't exist, returns ("", false, nil).
	Load(path string) (content string, exists bool, err error)
}

// FileNotesLoader loads notes from the filesystem.
type FileNotesLoader struct{}

// NewFileNotesLoader creates a new FileNotesLoader.
func NewFileNotesLoader() *FileNotesLoader {
	return &FileNotesLoader{}
}

// Load reads the notes file from the filesystem.
// Returns the content, whether the file exists, and any error.
func (l *FileNotesLoader) Load(path string) (string, bool, error) {
	if path == "" {
		return "", false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("failed to read notes file %q: %w", path, err)
	}

	return string(data), true, nil
}

// MockNotesLoader is a mock implementation of NotesLoader for testing.
type MockNotesLoader struct {
	Content string
	Exists  bool
	Err     error
}

// Load returns the preconfigured mock values.
func (m *MockNotesLoader) Load(_ string) (string, bool, error) {
	return m.Content, m.Exists, m.Err
}
