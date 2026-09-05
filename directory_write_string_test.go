package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// WriteFromString Tests - Success Cases
// ============================================================================

func TestDirectory_WriteFromString_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "simple file",
			filename: "test.txt",
			content:  "hello world",
		},
		{
			name:     "file in subdirectory",
			filename: "subdir/test.txt",
			content:  "nested content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.WriteFromString(tt.filename, tt.content, 0644)
			if err != nil {
				t.Errorf("WriteFromString() unexpected error: %v", err)
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
// WriteFromString Tests - Error Cases
// ============================================================================

func TestDirectory_WriteFromString_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "empty filename",
			filename: "",
			content:  "test content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.WriteFromString(tt.filename, tt.content, 0644)
			if err == nil {
				t.Error("WriteFromString() should return error for empty filename")
			}
		})
	}
}

// ============================================================================
// WriteFromString Tests - Default Permissions from Directory
// ============================================================================

func TestDirectory_WriteFromString_UsesDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "file written with default permissions",
			filename: "test.txt",
			content:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.WriteFromString(tt.filename, tt.content, 0)
			if err != nil {
				t.Errorf("WriteFromString() unexpected error: %v", err)
			}

			testFile := filepath.Join(tmpDir, tt.filename)
			info, err := os.Stat(testFile)
			if err != nil {
				t.Fatalf("failed to stat written file %q: %v", tt.filename, err)
			}

			expectedPerm := os.FileMode(0644) // default when perm=0 and no defaultPerm set
			if info.Mode().Perm() != expectedPerm {
				t.Errorf("expected permissions %v for %q, got %v", expectedPerm, tt.filename, info.Mode().Perm())
			}
		})
	}
}

// ============================================================================
// WriteFromString Tests - Ownership from Directory
// ============================================================================

func TestDirectory_WriteFromString_UsesOwnership(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		filename   string
		content    string
		defaultUID int
		defaultGID int
	}{
		{
			name:       "file written with ownership",
			filename:   "test.txt",
			content:    "hello world",
			defaultUID: 1000,
			defaultGID: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)
			dir.defaultUID = tt.defaultUID
			dir.defaultGID = tt.defaultGID

			err := dir.WriteFromString(tt.filename, tt.content, 0644)
			if err != nil {
				t.Errorf("WriteFromString() unexpected error: %v", err)
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
