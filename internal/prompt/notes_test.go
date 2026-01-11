package prompt

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileNotesLoader(t *testing.T) {
	t.Parallel()

	loader := NewFileNotesLoader()
	assert.NotNil(t, loader)
}

func TestFileNotesLoader_Load_ExistingFile(t *testing.T) {
	t.Parallel()

	// Create a temporary file with content
	tmpDir := t.TempDir()
	notesPath := filepath.Join(tmpDir, "SHARED_TASK_NOTES.md")
	expectedContent := "# Notes\n\n- Task 1 completed\n- Task 2 in progress"
	err := os.WriteFile(notesPath, []byte(expectedContent), 0644)
	require.NoError(t, err)

	loader := NewFileNotesLoader()
	content, exists, err := loader.Load(notesPath)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, expectedContent, content)
}

func TestFileNotesLoader_Load_NonExistentFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	notesPath := filepath.Join(tmpDir, "non_existent_file.md")

	loader := NewFileNotesLoader()
	content, exists, err := loader.Load(notesPath)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Empty(t, content)
}

func TestFileNotesLoader_Load_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	notesPath := filepath.Join(tmpDir, "empty_notes.md")
	err := os.WriteFile(notesPath, []byte(""), 0644)
	require.NoError(t, err)

	loader := NewFileNotesLoader()
	content, exists, err := loader.Load(notesPath)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Empty(t, content)
}

func TestFileNotesLoader_Load_EmptyPath(t *testing.T) {
	t.Parallel()

	loader := NewFileNotesLoader()
	content, exists, err := loader.Load("")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Empty(t, content)
}

func TestFileNotesLoader_Load_DirectoryInsteadOfFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	loader := NewFileNotesLoader()
	_, _, err := loader.Load(tmpDir)

	// Reading a directory should return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read notes file")
}

func TestFileNotesLoader_Load_FileWithUnicodeContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	notesPath := filepath.Join(tmpDir, "unicode_notes.md")
	expectedContent := "# 노트\n\n- 작업 1 완료\n- 작업 2 진행 중\n- 이모지 테스트: 🚀"
	err := os.WriteFile(notesPath, []byte(expectedContent), 0644)
	require.NoError(t, err)

	loader := NewFileNotesLoader()
	content, exists, err := loader.Load(notesPath)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, expectedContent, content)
}

func TestFileNotesLoader_Load_LargeFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	notesPath := filepath.Join(tmpDir, "large_notes.md")

	// Create a large file (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('a' + (i % 26))
	}
	err := os.WriteFile(notesPath, largeContent, 0644)
	require.NoError(t, err)

	loader := NewFileNotesLoader()
	content, exists, err := loader.Load(notesPath)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Len(t, content, 1024*1024)
}

func TestMockNotesLoader_Load(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		mockContent     string
		mockExists      bool
		mockErr         error
		expectedContent string
		expectedExists  bool
		expectedErr     bool
	}{
		{
			name:            "returns configured content",
			mockContent:     "mock notes content",
			mockExists:      true,
			mockErr:         nil,
			expectedContent: "mock notes content",
			expectedExists:  true,
			expectedErr:     false,
		},
		{
			name:            "returns not exists",
			mockContent:     "",
			mockExists:      false,
			mockErr:         nil,
			expectedContent: "",
			expectedExists:  false,
			expectedErr:     false,
		},
		{
			name:            "returns error",
			mockContent:     "",
			mockExists:      false,
			mockErr:         errors.New("mock error"),
			expectedContent: "",
			expectedExists:  false,
			expectedErr:     true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := &MockNotesLoader{
				Content: tt.mockContent,
				Exists:  tt.mockExists,
				Err:     tt.mockErr,
			}

			content, exists, err := loader.Load("any/path.md")

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedExists, exists)
			assert.Equal(t, tt.expectedContent, content)
		})
	}
}

func TestMockNotesLoader_IgnoresPath(t *testing.T) {
	t.Parallel()

	loader := &MockNotesLoader{
		Content: "fixed content",
		Exists:  true,
		Err:     nil,
	}

	// Call with different paths - should return same result
	content1, exists1, err1 := loader.Load("path1.md")
	content2, exists2, err2 := loader.Load("path2.md")
	content3, exists3, err3 := loader.Load("")

	assert.Equal(t, content1, content2)
	assert.Equal(t, content2, content3)
	assert.Equal(t, exists1, exists2)
	assert.Equal(t, exists2, exists3)
	assert.Equal(t, err1, err2)
	assert.Equal(t, err2, err3)
}

func TestNotesLoaderInterface(t *testing.T) {
	t.Parallel()

	// Verify that both loaders implement the interface
	var _ NotesLoader = &FileNotesLoader{}
	var _ NotesLoader = &MockNotesLoader{}
}
