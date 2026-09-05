package dirly

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// BatchReadToBuff Tests - Success Cases
// ============================================================================

func TestDirectory_BatchReadToBuff_Success(t *testing.T) {
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

	result, err := dir.BatchReadToBuff([]string{"test1.txt", "test2.txt"})
	if err != nil {
		t.Errorf("BatchReadToBuff() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}

	for filename, expectedContent := range files {
		reader, ok := result[filename]
		if !ok {
			t.Errorf("missing key %q in result", filename)
			continue
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read from buffer for %q: %v", filename, err)
		}

		if string(data) != expectedContent {
			t.Errorf("expected content %q for %q, got %q", expectedContent, filename, string(data))
		}
	}
}

// ============================================================================
// BatchReadToBuff Tests - Error Cases
// ============================================================================

func TestDirectory_BatchReadToBuff_Error(t *testing.T) {
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

			result, err := dir.BatchReadToBuff(tt.files)
			if err == nil {
				t.Error("BatchReadToBuff() should return error for non-existent file")
			}

			if result != nil {
				t.Errorf("expected nil result on error, got %v", result)
			}
		})
	}
}

// ============================================================================
// BatchWriteFromBuff Tests - Success Cases
// ============================================================================

func TestDirectory_BatchWriteFromBuff_Success(t *testing.T) {
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

			files := make(map[string]*bufio.Reader)
			for filename, content := range tt.files {
				files[filename] = bufio.NewReader(strings.NewReader(content))
			}

			err := dir.BatchWriteFromBuff(files)
			if err != nil {
				t.Errorf("BatchWriteFromBuff() unexpected error: %v", err)
			}

			for filename, expectedContent := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
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
// BatchWriteFromBuff Tests - Error Cases
// ============================================================================

func TestDirectory_BatchWriteFromBuff_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string]*bufio.Reader
	}{
		{
			name:  "empty file list",
			files: map[string]*bufio.Reader{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromBuff(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromBuff() should not error for empty map: %v", err)
			}
		})
	}
}

// ============================================================================
// BatchWriteFromBuff Tests - Default Permissions from Directory
// ============================================================================

func TestDirectory_BatchWriteFromBuff_UsesDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string]*bufio.Reader
	}{
		{
			name: "files written with default permissions",
			files: map[string]*bufio.Reader{
				"test1.txt": bufio.NewReader(strings.NewReader("content1")),
				"test2.txt": bufio.NewReader(strings.NewReader("content2")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromBuff(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromBuff() unexpected error: %v", err)
			}

			for filename := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				info, err := os.Stat(testPath)
				if err != nil {
					t.Fatalf("failed to stat written file %q: %v", filename, err)
				}

				expectedPerm := os.FileMode(0644) // default from WriteFromBuff when perm=0
				if info.Mode().Perm() != expectedPerm {
					t.Errorf("expected permissions %v for %q, got %v", expectedPerm, filename, info.Mode().Perm())
				}
			}
		})
	}
}

// ============================================================================
// BatchWriteFromBuff Tests - Default Permissions from Directory with Custom Perm
// ============================================================================

func TestDirectory_BatchWriteFromBuff_UsesCustomDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		defaultPerm os.FileMode
		files       map[string]*bufio.Reader
	}{
		{
			name:        "files written with custom default permissions",
			defaultPerm: 0600,
			files: map[string]*bufio.Reader{
				"test1.txt": bufio.NewReader(strings.NewReader("content1")),
				"test2.txt": bufio.NewReader(strings.NewReader("content2")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)
			dir.defaultPerm = tt.defaultPerm

			err := dir.BatchWriteFromBuff(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromBuff() unexpected error: %v", err)
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
