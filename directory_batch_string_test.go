package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// BatchReadToString Tests - Success Cases
// ============================================================================

func TestDirectory_BatchReadToString_Success(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"test1.txt": "content1",
		"test2.txt": "content2",
	}

	for filename, content := range files {
		testPath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(testPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	dir := NewDirectory(tmpDir)

	result, err := dir.BatchReadToString([]string{"test1.txt", "test2.txt"})
	if err != nil {
		t.Errorf("BatchReadToString() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}

	for filename, expectedContent := range files {
		content, ok := result[filename]
		if !ok {
			t.Errorf("missing key %q in result", filename)
			continue
		}
		if content != expectedContent {
			t.Errorf("expected content %q for %q, got %q", expectedContent, filename, content)
		}
	}
}

// ============================================================================
// BatchReadToString Tests - Error Cases
// ============================================================================

func TestDirectory_BatchReadToString_Error(t *testing.T) {
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

			result, err := dir.BatchReadToString(tt.files)
			if err == nil {
				t.Error("BatchReadToString() should return error for non-existent file")
			}

			if result != nil {
				t.Errorf("expected nil result on error, got %v", result)
			}
		})
	}
}

// ============================================================================
// BatchWriteFromString Tests - Success Cases
// ============================================================================

func TestDirectory_BatchWriteFromString_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name: "multiple files",
			files: map[string]string{
				"test1.txt": "content1",
				"test2.txt": "content2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromString(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromString() unexpected error: %v", err)
			}

			for filename, content := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				result, err := os.ReadFile(testPath)
				if err != nil {
					t.Fatalf("failed to read written file %q: %v", filename, err)
				}

				if string(result) != content {
					t.Errorf("expected content %q for %q, got %q", content, filename, result)
				}
			}
		})
	}
}

// ============================================================================
// BatchWriteFromString Tests - Error Cases
// ============================================================================

func TestDirectory_BatchWriteFromString_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name:  "empty file list",
			files: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromString(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromString() should not error for empty map: %v", err)
			}
		})
	}
}

// ============================================================================
// BatchWriteFromString Tests - Default Permissions from Directory
// ============================================================================

func TestDirectory_BatchWriteFromString_UsesDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name: "files written with default permissions",
			files: map[string]string{
				"test1.txt": "content1",
				"test2.txt": "content2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromString(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromString() unexpected error: %v", err)
			}

			for filename := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				info, err := os.Stat(testPath)
				if err != nil {
					t.Fatalf("failed to stat written file %q: %v", filename, err)
				}

				expectedPerm := os.FileMode(0644) // default from WriteFromString when perm=0
				if info.Mode().Perm() != expectedPerm {
					t.Errorf("expected permissions %v for %q, got %v", expectedPerm, filename, info.Mode().Perm())
				}
			}
		})
	}
}

// ============================================================================
// BatchWriteFromString Tests - Default Permissions from Directory with Custom Perm
// ============================================================================

func TestDirectory_BatchWriteFromString_UsesCustomDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		defaultPerm os.FileMode
		files       map[string]string
	}{
		{
			name:        "files written with custom default permissions",
			defaultPerm: 0600,
			files: map[string]string{
				"test1.txt": "content1",
				"test2.txt": "content2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)
			dir.defaultPerm = tt.defaultPerm

			err := dir.BatchWriteFromString(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromString() unexpected error: %v", err)
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
