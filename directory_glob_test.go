package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// getByGlob Tests - Success Cases
// ============================================================================

func TestDirectory_getByGlob_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		pattern  string
		files    []string
		expected int
	}{
		{
			name:     "wildcard pattern",
			pattern:  "*.txt",
			files:    []string{"test1.txt", "test2.txt", "test3.yaml"},
			expected: 2,
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

			result, err := dir.getByGlob(tt.pattern)
			if err != nil {
				t.Errorf("getByGlob() unexpected error: %v", err)
			}

			if len(result) != tt.expected {
				t.Errorf("expected %d matches, got %d", tt.expected, len(result))
			}
		})
	}
}

// ============================================================================
// getByGlob Tests - Error Cases
// ============================================================================

func TestDirectory_getByGlob_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		pattern string
	}{
		{
			name:    "empty pattern",
			pattern: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			result, err := dir.getByGlob(tt.pattern)
			if err == nil {
				t.Error("getByGlob() should return error for empty pattern")
			}

			if result != nil {
				t.Errorf("expected nil result on error, got %v", result)
			}
		})
	}
}

// ============================================================================
// GetAllByGlobAbs Tests
// ============================================================================

func TestDirectory_GetAbsByGlob(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	dir := NewDirectory(tmpDir)

	result, err := dir.GetAllByGlobAbs("*.txt")
	if err != nil {
		t.Errorf("GetAllByGlobAbs() unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 match, got %d", len(result))
	}

	if result[0] != testFile {
		t.Errorf("expected absolute path %q, got %q", testFile, result[0])
	}
}

// ============================================================================
// GetAllByGlobRel Tests
// ============================================================================

func TestDirectory_GetRelByGlob(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	dir := NewDirectory(tmpDir)

	result, err := dir.GetAllByGlobRel("*.txt")
	if err != nil {
		t.Errorf("GetAllByGlobRel() unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 match, got %d", len(result))
	}

	if result[0] != "test.txt" {
		t.Errorf("expected relative path 'test.txt', got %q", result[0])
	}
}

// ============================================================================
// GetAllAbs Tests
// ============================================================================

func TestDirectory_GetAllAbs(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{"test1.txt", "test2.yaml"}
	for _, file := range files {
		testPath := filepath.Join(tmpDir, file)
		if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	dir := NewDirectory(tmpDir)

	result, err := dir.GetAllAbs()
	if err != nil {
		t.Errorf("GetAllAbs() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result))
	}

	for _, path := range result {
		if !filepath.IsAbs(path) {
			t.Errorf("GetAllAbs() should return absolute paths, got %q", path)
		}
	}
}

// ============================================================================
// GetAllRel Tests
// ============================================================================

func TestDirectory_GetAllRel(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{"test1.txt", "test2.yaml"}
	for _, file := range files {
		testPath := filepath.Join(tmpDir, file)
		if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	dir := NewDirectory(tmpDir)

	result, err := dir.GetAllRel()
	if err != nil {
		t.Errorf("GetAllRel() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result))
	}

	for _, path := range result {
		if filepath.IsAbs(path) {
			t.Errorf("GetAllRel() should return relative paths, got %q", path)
		}
	}
}

// ============================================================================
// GetAll* Tests - Directory Ignoring
// ============================================================================

func TestDirectory_GetAll_IgnoresDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{"file1.txt", "file2.yaml"}
	dirs := []string{"subdir1", "subdir2/nested"}

	for _, file := range files {
		testPath := filepath.Join(tmpDir, file)
		if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	for _, dir := range dirs {
		testPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(testPath, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}
	}

	dir := NewDirectory(tmpDir)

	resultAbs, err := dir.GetAllAbs()
	if err != nil {
		t.Errorf("GetAllAbs() unexpected error: %v", err)
	}

	if len(resultAbs) != 2 {
		t.Errorf("expected 2 files (no dirs), got %d: %v", len(resultAbs), resultAbs)
	}

	for _, path := range resultAbs {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("failed to stat %q: %v", path, err)
		}
		if info.IsDir() {
			t.Errorf("GetAllAbs() should not return directories, got %q", path)
		}
	}

	resultRel, err := dir.GetAllRel()
	if err != nil {
		t.Errorf("GetAllRel() unexpected error: %v", err)
	}

	if len(resultRel) != 2 {
		t.Errorf("expected 2 files (no dirs), got %d: %v", len(resultRel), resultRel)
	}

	for _, path := range resultRel {
		if filepath.IsAbs(path) {
			t.Errorf("GetAllRel() should return relative paths, got %q", path)
		}
	}
}

// ============================================================================
// GetAll* Tests - Filter Application
// ============================================================================

func TestDirectory_GetAll_AppliesFilters(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{
		"allowed.txt", "allowed.yaml",
		"excluded.txt", "excluded.yaml",
		"other.md",
	}

	for _, file := range files {
		testPath := filepath.Join(tmpDir, file)
		if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	testDir := filepath.Join(tmpDir, "excluded")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		name             string
		setup            func() *Directory
		expectedCount    int
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "include pattern filter",
			setup: func() *Directory {
				return NewFilteredDirectory(tmpDir).Include("allowed.*").Build()
			},
			expectedCount:    2,
			shouldContain:    []string{"allowed.txt", "allowed.yaml"},
			shouldNotContain: []string{"excluded.txt", "other.md"},
		},
		{
			name: "exclude pattern filter",
			setup: func() *Directory {
				return NewFilteredDirectory(tmpDir).Exclude("excluded.*").Build()
			},
			expectedCount:    3,
			shouldContain:    []string{"allowed.txt", "other.md"},
			shouldNotContain: []string{"excluded.txt", "excluded.yaml"},
		},
		{
			name: "extension filter",
			setup: func() *Directory {
				return NewFilteredDirectory(tmpDir).WithExtensions("txt").Build()
			},
			expectedCount:    2,
			shouldContain:    []string{"allowed.txt", "excluded.txt"},
			shouldNotContain: []string{"allowed.yaml", "other.md"},
		},
		{
			name: "combined filters",
			setup: func() *Directory {
				return NewFilteredDirectory(tmpDir).Include("*.*").Exclude("excluded.*").WithExtensions("txt", "yaml").Build()
			},
			expectedCount:    2,
			shouldContain:    []string{"allowed.txt", "allowed.yaml"},
			shouldNotContain: []string{"excluded.txt", "other.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()

			resultAbs, err := d.GetAllAbs()
			if err != nil {
				t.Fatalf("GetAllAbs() unexpected error: %v", err)
			}

			if len(resultAbs) != tt.expectedCount {
				t.Errorf("expected %d files, got %d: %v", tt.expectedCount, len(resultAbs), resultAbs)
			}

			resultRel, err := d.GetAllRel()
			if err != nil {
				t.Fatalf("GetAllRel() unexpected error: %v", err)
			}

			if len(resultRel) != tt.expectedCount {
				t.Errorf("expected %d files (rel), got %d: %v", tt.expectedCount, len(resultRel), resultRel)
			}

			for _, expected := range tt.shouldContain {
				found := false
				for _, path := range resultRel {
					if filepath.Base(path) == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to contain %q, got: %v", expected, resultRel)
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				found := false
				for _, path := range resultRel {
					if filepath.Base(path) == notExpected {
						found = true
						break
					}
				}
				if found {
					t.Errorf("should not contain %q, got: %v", notExpected, resultRel)
				}
			}
		})
	}
}

// ============================================================================
// GetAll* Tests - Glob with Filters
// ============================================================================

func TestDirectory_GetAllByGlob_AppliesFilters(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{
		"allowed.txt", "allowed.yaml",
		"excluded.txt",
		"other.md",
	}

	for _, file := range files {
		testPath := filepath.Join(tmpDir, file)
		if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	tests := []struct {
		name          string
		pattern       string
		setup         func() *Directory
		expectedCount int
	}{
		{
			name:    "glob with include filter",
			pattern: "*.txt",
			setup: func() *Directory {
				return NewFilteredDirectory(tmpDir).Include("allowed.*").Build()
			},
			expectedCount: 1,
		},
		{
			name:    "glob with exclude filter",
			pattern: "*.txt",
			setup: func() *Directory {
				return NewFilteredDirectory(tmpDir).Exclude("excluded.*").Build()
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()

			resultAbs, err := d.GetAllByGlobAbs(tt.pattern)
			if err != nil {
				t.Fatalf("GetAllByGlobAbs() unexpected error: %v", err)
			}

			if len(resultAbs) != tt.expectedCount {
				t.Errorf("expected %d files, got %d: %v", tt.expectedCount, len(resultAbs), resultAbs)
			}

			resultRel, err := d.GetAllByGlobRel(tt.pattern)
			if err != nil {
				t.Fatalf("GetAllByGlobRel() unexpected error: %v", err)
			}

			if len(resultRel) != tt.expectedCount {
				t.Errorf("expected %d files (rel), got %d: %v", tt.expectedCount, len(resultRel), resultRel)
			}
		})
	}
}
