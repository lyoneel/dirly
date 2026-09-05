package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// BatchRemoveFromDir Tests - Success Cases
// ============================================================================

func TestDirectory_BatchRemoveFromDir_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "file",
			path: "test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := filepath.Join(tmpDir, tt.path)
			if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			err := dir.BatchRemoveAllFromDir(tt.path)
			if err != nil {
				t.Errorf("BatchRemoveFromDir() unexpected error: %v", err)
			}

			if _, err := os.Stat(testPath); !os.IsNotExist(err) {
				t.Error("path should be removed")
			}
		})
	}
}

// ============================================================================
// BatchRemoveFile Tests - Success Cases
// ============================================================================

func TestDirectory_BatchRemoveFile_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files []string
	}{
		{
			name:  "multiple files",
			files: []string{"test1.txt", "test2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, file := range tt.files {
				testPath := filepath.Join(tmpDir, file)
				if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			dir := NewDirectory(tmpDir)

			err := dir.BatchRemove(tt.files)
			if err != nil {
				t.Errorf("BatchRemoveFile() unexpected error: %v", err)
			}

			for _, file := range tt.files {
				testPath := filepath.Join(tmpDir, file)
				if _, err := os.Stat(testPath); !os.IsNotExist(err) {
					t.Error("file should be removed")
				}
			}
		})
	}
}

// ============================================================================
// BatchRemoveFile Tests - Error Cases
// ============================================================================

func TestDirectory_BatchRemoveFile_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files []string
	}{
		{
			name:  "empty file list",
			files: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchRemove(tt.files)
			if err != nil {
				t.Errorf("BatchRemoveFile() should not error for empty list: %v", err)
			}
		})
	}
}
