package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// RemoveFile Tests - Success Cases
// ============================================================================

func TestDirectory_RemoveFile_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "simple file",
			filename: "test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			err := dir.Remove(tt.filename)
			if err != nil {
				t.Errorf("RemoveFile() unexpected error: %v", err)
			}

			if _, err := os.Stat(testFile); !os.IsNotExist(err) {
				t.Error("file should be removed")
			}
		})
	}
}

// ============================================================================
// RemoveFile Tests - Error Cases
// ============================================================================

func TestDirectory_RemoveFile_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "empty filename",
			filename: "",
		},
		{
			name:     "non-existent file",
			filename: "nonexistent.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.Remove(tt.filename)
			if err == nil {
				t.Error("RemoveFile() should return error")
			}
		})
	}
}
