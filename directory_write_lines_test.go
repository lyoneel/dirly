package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// WriteFromLines Tests - Success Cases
// ============================================================================

func TestDirectory_WriteFromLines_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		lines    []string
	}{
		{
			name:     "simple file",
			filename: "test.txt",
			lines:    []string{"line1", "line2", "line3"},
		},
		{
			name:     "file in subdirectory",
			filename: "subdir/test.txt",
			lines:    []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.WriteFromLines(tt.filename, tt.lines, 0644)
			if err != nil {
				t.Errorf("WriteFromLines() unexpected error: %v", err)
			}

			testFile := filepath.Join(tmpDir, tt.filename)
			result, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("failed to read written file: %v", err)
			}

			expectedContent := ""
			for _, line := range tt.lines {
				expectedContent += line + "\n"
			}
			if string(result) != expectedContent {
				t.Errorf("expected content %q, got %q", expectedContent, string(result))
			}
		})
	}
}

// ============================================================================
// WriteFromLines Tests - Error Cases
// ============================================================================

func TestDirectory_WriteFromLines_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		lines    []string
	}{
		{
			name:     "empty filename",
			filename: "",
			lines:    []string{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.WriteFromLines(tt.filename, tt.lines, 0644)
			if err == nil {
				t.Error("WriteFromLines() should return error for empty filename")
			}
		})
	}
}

// ============================================================================
// WriteFromLines Tests - Default Permissions from Directory
// ============================================================================

func TestDirectory_WriteFromLines_UsesDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		lines    []string
	}{
		{
			name:     "file written with default permissions",
			filename: "test.txt",
			lines:    []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.WriteFromLines(tt.filename, tt.lines, 0)
			if err != nil {
				t.Errorf("WriteFromLines() unexpected error: %v", err)
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
// WriteFromLines Tests - Custom Default Permissions from Directory
// ============================================================================

func TestDirectory_WriteFromLines_UsesCustomDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		filename    string
		lines       []string
		defaultPerm os.FileMode
	}{
		{
			name:        "file written with custom default permissions",
			filename:    "test.txt",
			lines:       []string{"line1"},
			defaultPerm: 0600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)
			dir.defaultPerm = tt.defaultPerm

			err := dir.WriteFromLines(tt.filename, tt.lines, 0)
			if err != nil {
				t.Errorf("WriteFromLines() unexpected error: %v", err)
			}

			testFile := filepath.Join(tmpDir, tt.filename)
			info, err := os.Stat(testFile)
			if err != nil {
				t.Fatalf("failed to stat written file %q: %v", tt.filename, err)
			}

			expectedPerm := tt.defaultPerm
			if info.Mode().Perm() != expectedPerm {
				t.Errorf("expected permissions %v for %q, got %v", expectedPerm, tt.filename, info.Mode().Perm())
			}
		})
	}
}
