package dirly

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// BatchReadToScanner Tests - Success Cases
// ============================================================================

func TestDirectory_BatchReadToScanner_Success(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"test1.txt": "line1\nline2\n",
		"test2.txt": "hello\nworld\n",
	}

	for filename, content := range files {
		testPath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(testPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	dir := NewDirectory(tmpDir)

	result, err := dir.BatchReadToScanner([]string{"test1.txt", "test2.txt"})
	if err != nil {
		t.Errorf("BatchReadToScanner() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}

	for filename, expectedContent := range files {
		scanner, ok := result[filename]
		if !ok {
			t.Errorf("missing key %q in result", filename)
			continue
		}

		var data []byte
		for scanner.Scan() {
			data = append(data, scanner.Bytes()...)
			data = append(data, '\n')
		}

		if err := scanner.Err(); err != nil {
			t.Fatalf("scanner error for %q: %v", filename, err)
		}

		if string(data) != expectedContent {
			t.Errorf("expected content %q for %q, got %q", expectedContent, filename, string(data))
		}
	}
}

// ============================================================================
// BatchReadToScanner Tests - Error Cases
// ============================================================================

func TestDirectory_BatchReadToScanner_Error(t *testing.T) {
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

			result, err := dir.BatchReadToScanner(tt.files)
			if err == nil {
				t.Error("BatchReadToScanner() should return error for non-existent file")
			}

			if result != nil {
				t.Errorf("expected nil result on error, got %v", result)
			}
		})
	}
}

// ============================================================================
// BatchWriteFromScanner Tests - Success Cases
// ============================================================================

func TestDirectory_BatchWriteFromScanner_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name: "multiple files",
			files: map[string]string{
				"test1.txt": "line1\nline2\n",
				"test2.txt": "hello\nworld\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			files := make(map[string]*bufio.Scanner)
			for filename, content := range tt.files {
				files[filename] = bufio.NewScanner(strings.NewReader(content))
			}

			err := dir.BatchWriteFromScanner(files)
			if err != nil {
				t.Errorf("BatchWriteFromScanner() unexpected error: %v", err)
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
// BatchWriteFromScanner Tests - Error Cases
// ============================================================================

func TestDirectory_BatchWriteFromScanner_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string]*bufio.Scanner
	}{
		{
			name:  "empty file list",
			files: map[string]*bufio.Scanner{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromScanner(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromScanner() should not error for empty map: %v", err)
			}
		})
	}
}

// ============================================================================
// BatchWriteFromScanner Tests - Default Permissions from Directory
// ============================================================================

func TestDirectory_BatchWriteFromScanner_UsesDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string]*bufio.Scanner
	}{
		{
			name: "files written with default permissions",
			files: map[string]*bufio.Scanner{
				"test1.txt": bufio.NewScanner(strings.NewReader("line1\n")),
				"test2.txt": bufio.NewScanner(strings.NewReader("line2\n")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromScanner(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromScanner() unexpected error: %v", err)
			}

			for filename := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				info, err := os.Stat(testPath)
				if err != nil {
					t.Fatalf("failed to stat written file %q: %v", filename, err)
				}

				expectedPerm := os.FileMode(0644) // default from WriteFromScanner when perm=0
				if info.Mode().Perm() != expectedPerm {
					t.Errorf("expected permissions %v for %q, got %v", expectedPerm, filename, info.Mode().Perm())
				}
			}
		})
	}
}

// ============================================================================
// BatchWriteFromScanner Tests - Default Permissions from Directory with Custom Perm
// ============================================================================

func TestDirectory_BatchWriteFromScanner_UsesCustomDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		defaultPerm os.FileMode
		files       map[string]*bufio.Scanner
	}{
		{
			name:        "files written with custom default permissions",
			defaultPerm: 0600,
			files: map[string]*bufio.Scanner{
				"test1.txt": bufio.NewScanner(strings.NewReader("line1\n")),
				"test2.txt": bufio.NewScanner(strings.NewReader("line2\n")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)
			dir.defaultPerm = tt.defaultPerm

			err := dir.BatchWriteFromScanner(tt.files)
			if err != nil {
				t.Errorf("BatchWriteFromScanner() unexpected error: %v", err)
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
