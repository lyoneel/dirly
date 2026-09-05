package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// WriteFile Tests - Success Cases
// ============================================================================

func TestDirectory_WriteFile_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
		perm     os.FileMode
	}{
		{
			name:     "simple file",
			filename: "test.txt",
			content:  "hello world",
			perm:     0644,
		},
		{
			name:     "file in subdirectory",
			filename: "subdir/test.txt",
			content:  "hello world",
			perm:     0644,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.WriteFromBytes(tt.filename, []byte(tt.content), tt.perm)
			if err != nil {
				t.Errorf("WriteFile() unexpected error: %v", err)
			}

			testFile := filepath.Join(tmpDir, tt.filename)
			result, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("failed to read written file: %v", err)
			}

			if string(result) != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, string(result))
			}
		})
	}
}

// ============================================================================
// WriteFile Tests - Error Cases
// ============================================================================

func TestDirectory_WriteFile_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "empty filename",
			filename: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.WriteFromBytes(tt.filename, []byte("test"), 0644)
			if err == nil {
				t.Error("WriteFile() should return error for empty filename")
			}
		})
	}
}
