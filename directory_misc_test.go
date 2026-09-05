package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// isPathWithinBase Tests
// ============================================================================

func TestDirectory_isPathWithinBase(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		path     string
		expected bool
	}{
		{
			name:     "path within base",
			basePath: "/tmp/test",
			path:     "/tmp/test/config.yaml",
			expected: true,
		},
		{
			name:     "path is base itself",
			basePath: "/tmp/test",
			path:     "/tmp/test",
			expected: true,
		},
		{
			name:     "path outside base",
			basePath: "/tmp/test",
			path:     "/etc/passwd",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tt.basePath)

			result := dir.isPathWithinBase(tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// AbsPath Tests
// ============================================================================

func TestDirectory_AbsPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "simple filename",
			filename: "test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			result, err := dir.AbsPath(tt.filename)
			if err != nil {
				t.Errorf("AbsPath() unexpected error: %v", err)
			}

			expected := filepath.Join(tmpDir, tt.filename)
			if result != expected {
				t.Errorf("expected absolute path %q, got %q", expected, result)
			}
		})
	}
}

// ============================================================================
// BatchRead Tests - Success Cases
// ============================================================================

func TestDirectory_BatchRead_Success(t *testing.T) {
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

	result, err := dir.BatchReadToBytes([]string{"test1.txt", "test2.txt"})
	if err != nil {
		t.Errorf("BatchRead() unexpected error: %v", err)
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
		if string(content) != expectedContent {
			t.Errorf("expected content %q for %q, got %q", expectedContent, filename, string(content))
		}
	}
}

// ============================================================================
// BatchRead Tests - Error Cases
// ============================================================================

func TestDirectory_BatchRead_Error(t *testing.T) {
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

			result, err := dir.BatchReadToBytes(tt.files)
			if err == nil {
				t.Error("BatchRead() should return error for non-existent file")
			}

			if result != nil {
				t.Errorf("expected nil result on error, got %v", result)
			}
		})
	}
}

// ============================================================================
// BatchWrite Tests - Success Cases
// ============================================================================

func TestDirectory_BatchWrite_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string][]byte
	}{
		{
			name: "multiple files",
			files: map[string][]byte{
				"test1.txt": []byte("content1"),
				"test2.txt": []byte("content2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromBytes(tt.files)
			if err != nil {
				t.Errorf("BatchWrite() unexpected error: %v", err)
			}

			for filename, content := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				result, err := os.ReadFile(testPath)
				if err != nil {
					t.Fatalf("failed to read written file %q: %v", filename, err)
				}

				if string(result) != string(content) {
					t.Errorf("expected content %q for %q, got %q", content, filename, result)
				}
			}
		})
	}
}

// ============================================================================
// BatchWrite Tests - Error Cases
// ============================================================================

func TestDirectory_BatchWrite_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string][]byte
	}{
		{
			name:  "empty file list",
			files: map[string][]byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromBytes(tt.files)
			if err != nil {
				t.Errorf("BatchWrite() should not error for empty map: %v", err)
			}
		})
	}
}

// ============================================================================
// BatchWrite Tests - Default Permissions from Directory
// ============================================================================

func TestDirectory_BatchWrite_UsesDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		files map[string][]byte
	}{
		{
			name: "files written with default permissions",
			files: map[string][]byte{
				"test1.txt": []byte("content1"),
				"test2.txt": []byte("content2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			err := dir.BatchWriteFromBytes(tt.files)
			if err != nil {
				t.Errorf("BatchWrite() unexpected error: %v", err)
			}

			for filename := range tt.files {
				testPath := filepath.Join(tmpDir, filename)
				info, err := os.Stat(testPath)
				if err != nil {
					t.Fatalf("failed to stat written file %q: %v", filename, err)
				}

				expectedPerm := os.FileMode(0644) // default from WriteFromBytes when perm=0
				if info.Mode().Perm() != expectedPerm {
					t.Errorf("expected permissions %v for %q, got %v", expectedPerm, filename, info.Mode().Perm())
				}
			}
		})
	}
}

// ============================================================================
// BatchWrite Tests - Default Permissions from Directory with Custom Perm
// ============================================================================

func TestDirectory_BatchWrite_UsesCustomDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		defaultPerm os.FileMode
		files       map[string][]byte
	}{
		{
			name:        "files written with custom default permissions",
			defaultPerm: 0600,
			files: map[string][]byte{
				"test1.txt": []byte("content1"),
				"test2.txt": []byte("content2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)
			dir.defaultPerm = tt.defaultPerm

			err := dir.BatchWriteFromBytes(tt.files)
			if err != nil {
				t.Errorf("BatchWrite() unexpected error: %v", err)
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
