package dirly

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// WriteFromScanner Tests - Success Cases
// ============================================================================

func TestDirectory_WriteFromScanner_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		filename  string
		content   string
		wantLines int
	}{
		{
			name:      "simple file",
			filename:  "test.txt",
			content:   "line1\nline2\nline3\n",
			wantLines: 3,
		},
		{
			name:      "file in subdirectory",
			filename:  "subdir/test.txt",
			content:   "hello\nworld\n",
			wantLines: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			scanner := bufio.NewScanner(strings.NewReader(tt.content))
			err := dir.WriteFromScanner(tt.filename, scanner, 0644)
			if err != nil {
				t.Errorf("WriteFromScanner() unexpected error: %v", err)
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
// WriteFromScanner Tests - Error Cases
// ============================================================================

func TestDirectory_WriteFromScanner_Error(t *testing.T) {
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

			scanner := bufio.NewScanner(strings.NewReader(tt.content))
			err := dir.WriteFromScanner(tt.filename, scanner, 0644)
			if err == nil {
				t.Error("WriteFromScanner() should return error for empty filename")
			}
		})
	}
}

// ============================================================================
// WriteFromScanner Tests - Default Permissions from Directory
// ============================================================================

func TestDirectory_WriteFromScanner_UsesDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "file written with default permissions",
			filename: "test.txt",
			content:  "line1\nline2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			scanner := bufio.NewScanner(strings.NewReader(tt.content))
			err := dir.WriteFromScanner(tt.filename, scanner, 0)
			if err != nil {
				t.Errorf("WriteFromScanner() unexpected error: %v", err)
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
// WriteFromScanner Tests - Custom Default Permissions from Directory
// ============================================================================

func TestDirectory_WriteFromScanner_UsesCustomDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		filename    string
		content     string
		defaultPerm os.FileMode
	}{
		{
			name:        "file written with custom default permissions",
			filename:    "test.txt",
			content:     "line1\n",
			defaultPerm: 0600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)
			dir.defaultPerm = tt.defaultPerm

			scanner := bufio.NewScanner(strings.NewReader(tt.content))
			err := dir.WriteFromScanner(tt.filename, scanner, 0)
			if err != nil {
				t.Errorf("WriteFromScanner() unexpected error: %v", err)
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
