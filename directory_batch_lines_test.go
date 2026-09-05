package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// BatchReadToLines Tests - Success Cases
// ============================================================================

func TestDirectory_BatchReadToLines_Success(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string][]string{
		"test1.txt": {"line1", "line2"},
		"test2.txt": {"hello", "world"},
	}

	for filename, lines := range files {
		testPath := filepath.Join(tmpDir, filename)
		content := ""
		for _, line := range lines {
			content += line + "\n"
		}
		if err := os.WriteFile(testPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	dir := NewDirectory(tmpDir)

	result, err := dir.BatchReadToLines([]string{"test1.txt", "test2.txt"})
	if err != nil {
		t.Errorf("BatchReadToLines() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}

	for filename, expectedLines := range files {
		lines, ok := result[filename]
		if !ok {
			t.Errorf("missing key %q in result", filename)
			continue
		}
		if len(lines) != len(expectedLines) {
			t.Errorf("expected %d lines for %q, got %d", len(expectedLines), filename, len(lines))
			continue
		}
		for i, line := range expectedLines {
			if lines[i] != line {
				t.Errorf("expected line %q at index %d for %q, got %q", line, i, filename, lines[i])
			}
		}
	}
}

// ============================================================================
// BatchReadToLines Tests - Error Cases
// ============================================================================

func TestDirectory_BatchReadToLines_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files []string
	}{
		{
			name:  "non-existent file",
			files: []string{"nonexistent.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			result, err := dir.BatchReadToLines(tt.files)
			if err == nil {
				t.Error("BatchReadToLines() should return error for non-existent file")
			}

			if result != nil {
				t.Errorf("expected nil result on error, got %v", result)
			}
		})
	}
}

// ============================================================================
// BatchWriteFromLines Tests - Success Cases
// ============================================================================

func TestDirectory_BatchWriteFromLines_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string][]string
	}{
		{
			name: "multiple files",
			files: map[string][]string{
				"test1.txt": {"line1", "line2"},
				"test2.txt": {"hello", "world"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromLines(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromLines() unexpected error: %v", err)
			}

			for filename, lines := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				expectedContent := ""
				for _, line := range lines {
					expectedContent += line + "\n"
				}
				result, err := os.ReadFile(testPath)
				if err != nil {
					t.Fatalf("failed to read written file %q: %v", filename, err)
				}

				if string(result) != expectedContent {
					t.Errorf("expected content %q for %q, got %q", expectedContent, filename, result)
				}
			}
		})
	}
}

// ============================================================================
// BatchWriteFromLines Tests - Error Cases
// ============================================================================

func TestDirectory_BatchWriteFromLines_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string][]string
	}{
		{
			name:  "empty file list",
			files: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromLines(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromLines() should not error for empty map: %v", err)
			}
		})
	}
}

// ============================================================================
// BatchWriteFromLines Tests - Default Permissions from Directory
// ============================================================================

func TestDirectory_BatchWriteFromLines_UsesDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string][]string
	}{
		{
			name: "files written with default permissions",
			files: map[string][]string{
				"test1.txt": {"line1"},
				"test2.txt": {"line2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromLines(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromLines() unexpected error: %v", err)
			}

			for filename := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				info, err := os.Stat(testPath)
				if err != nil {
					t.Fatalf("failed to stat written file %q: %v", filename, err)
				}

				expectedPerm := os.FileMode(0644) // default from WriteFromLines when perm=0
				if info.Mode().Perm() != expectedPerm {
					t.Errorf("expected permissions %v for %q, got %v", expectedPerm, filename, info.Mode().Perm())
				}
			}
		})
	}
}

// ============================================================================
// BatchWriteFromLines Tests - Default Permissions from Directory with Custom Perm
// ============================================================================

func TestDirectory_BatchWriteFromLines_UsesCustomDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		defaultPerm os.FileMode
		files       map[string][]string
	}{
		{
			name:        "files written with custom default permissions",
			defaultPerm: 0600,
			files: map[string][]string{
				"test1.txt": {"line1"},
				"test2.txt": {"line2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)
			dir.defaultPerm = tt.defaultPerm

			err := dir.BatchWriteFromLines(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromLines() unexpected error: %v", err)
			}

			for filename := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				info, err := os.Stat(testPath)
				if err != nil {
					t.Fatalf("failed to stat written file %q: %v", filename, err)
				}

				expectedPerm := tt.defaultPerm
				if info.Mode().Perm() != expectedPerm {
					t.Errorf("expected permissions %v for %q, got %v", expectedPerm, filename, info.Mode().Perm())
				}
			}
		})
	}
}
