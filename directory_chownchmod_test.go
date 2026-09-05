package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// Chown Tests - Success Cases
// ============================================================================

func TestDirectory_Chown_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "simple file",
			filename: "test.txt",
		},
		{
			name:     "file in subdirectory",
			filename: "subdir/test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}

			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err := dir.Chown(tt.filename, -1, -1)
			if err != nil {
				t.Errorf("Chown() unexpected error: %v", err)
			}
		})
	}
}

// ============================================================================
// Chown Tests - Error Cases
// ============================================================================

func TestDirectory_Chown_Error(t *testing.T) {
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

			err := dir.Chown(tt.filename, -1, -1)
			if err == nil {
				t.Error("Chown() should return error")
			}
		})
	}
}

// ============================================================================
// Chmod Tests - Success Cases
// ============================================================================

func TestDirectory_Chmod_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		perm     os.FileMode
	}{
		{
			name:     "simple file",
			filename: "test.txt",
			perm:     0644,
		},
		{
			name:     "file in subdirectory",
			filename: "subdir/test.txt",
			perm:     0755,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}

			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err := dir.Chmod(tt.filename, tt.perm)
			if err != nil {
				t.Errorf("Chmod() unexpected error: %v", err)
			}

			info, err := os.Stat(testFile)
			if err != nil {
				t.Fatalf("failed to stat file after chmod: %v", err)
			}

			if info.Mode().Perm() != tt.perm {
				t.Errorf("expected permissions %o, got %o", tt.perm, info.Mode().Perm())
			}
		})
	}
}

// ============================================================================
// Chmod Tests - Error Cases
// ============================================================================

func TestDirectory_Chmod_Error(t *testing.T) {
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

			err := dir.Chmod(tt.filename, 0644)
			if err == nil {
				t.Error("Chmod() should return error")
			}
		})
	}
}
