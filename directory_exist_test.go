package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// Directory.Exists Tests - Basic Functionality
// ============================================================================

func Test_Exist_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func(dir string) string
		filename string
		expected bool
	}{
		{
			name: "file exists",
			setup: func(dir string) string {
				filePath := filepath.Join(dir, "test.txt")
				os.WriteFile(filePath, []byte("content"), 0644)
				return filePath
			},
			filename: "test.txt",
			expected: true,
		},
		{
			name: "file does not exist",
			setup: func(dir string) string {
				return ""
			},
			filename: "nonexistent.txt",
			expected: false,
		},
		{
			name:     "empty filename",
			setup:    func(dir string) string { return "" },
			filename: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			if tt.setup != nil {
				tt.setup(tmpDir)
			}

			result := dir.Exists(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Directory.Exists Tests - With Filters
// ============================================================================

func Test_Exist_WithFilters(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		files      []string
		include    []string
		exclude    []string
		extensions []string
		filename   string
		expected   bool
	}{
		{
			name:     "file exists without filters",
			files:    []string{"test.yaml"},
			filename: "test.yaml",
			expected: true,
		},
		{
			name:     "file excluded by exclude pattern",
			files:    []string{"temp.tmp"},
			exclude:  []string{"*.tmp"},
			filename: "temp.tmp",
			expected: false, // Excluded
		},
		{
			name:     "file included by include pattern",
			files:    []string{"test.yaml", "data.json"},
			include:  []string{"*.yaml"},
			filename: "test.yaml",
			expected: true,
		},
		{
			name:     "file not matching include pattern",
			files:    []string{"test.yaml", "data.json"},
			include:  []string{"*.json"},
			filename: "test.yaml",
			expected: false, // Not in *.json
		},
		{
			name:       "file with correct extension",
			files:      []string{"config.yaml", "readme.txt"},
			extensions: []string{"yaml"},
			filename:   "config.yaml",
			expected:   true,
		},
		{
			name:       "file with wrong extension filtered out",
			files:      []string{"config.yaml", "readme.txt"},
			extensions: []string{"yaml"},
			filename:   "readme.txt",
			expected:   false, // Not .yaml
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory(tmpDir)

			if len(tt.files) > 0 {
				for _, f := range tt.files {
					filePath := filepath.Join(tmpDir, f)
					os.WriteFile(filePath, []byte("content"), 0644)
				}
			}

			if len(tt.include) > 0 {
				builder = builder.Include(tt.include...)
			}
			if len(tt.exclude) > 0 {
				builder = builder.Exclude(tt.exclude...)
			}
			if len(tt.extensions) > 0 {
				builder = builder.WithExtensions(tt.extensions...)
			}

			dir := builder.Build()
			result := dir.Exists(tt.filename)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Directory.Exists Tests - Edge Cases
// ============================================================================

func Test_Exist_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func(dir string)
		filename string
		expected bool
	}{
		{
			name: "directory instead of file",
			setup: func(dir string) {
				os.Mkdir(filepath.Join(dir, "subdir"), 0755)
			},
			filename: "subdir",
			expected: false, // Exists should return false for directories
		},
		{
			name: "file with special characters in name",
			setup: func(dir string) {
				filePath := filepath.Join(dir, ".gitignore")
				os.WriteFile(filePath, []byte("*.tmp"), 0644)
			},
			filename: ".gitignore",
			expected: true,
		},
		{
			name: "file with spaces in name",
			setup: func(dir string) {
				filePath := filepath.Join(dir, "my file.txt")
				os.WriteFile(filePath, []byte("content"), 0644)
			},
			filename: "my file.txt",
			expected: true,
		},
		{
			name: "path traversal attempt",
			setup: func(dir string) {
				filePath := filepath.Join(tmpDir, "test.txt")
				os.WriteFile(filePath, []byte("content"), 0644)
			},
			filename: "../etc/passwd",
			expected: false, // Path traversal should be blocked
		},
		{
			name: "absolute path attempt",
			setup: func(dir string) {
				filePath := filepath.Join(tmpDir, "test.txt")
				os.WriteFile(filePath, []byte("content"), 0644)
			},
			filename: "/etc/passwd",
			expected: false, // Absolute paths should be blocked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			if tt.setup != nil {
				tt.setup(tmpDir)
			}

			result := dir.Exists(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Directory.Exists Tests - Case Sensitivity
// ============================================================================

func Test_Exist_CaseSensitivity(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		files         []string
		caseSensitive bool
		filename      string
		expected      bool
	}{
		{
			name:          "case sensitive exact match",
			files:         []string{"Config.yaml"},
			caseSensitive: true,
			filename:      "Config.yaml",
			expected:      true,
		},
		{
			name:          "case sensitive mismatch",
			files:         []string{"Config.yaml"},
			caseSensitive: true,
			filename:      "config.yaml",
			expected:      false, // Case-sensitive: Config != config
		},
		{
			name:          "case insensitive match (default, OS-dependent)",
			files:         []string{"Config.yaml"},
			caseSensitive: false,
			filename:      "config.yaml",
			expected:      false, // On Linux, filepath.Match is case-sensitive by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory(tmpDir).CaseSensitive(tt.caseSensitive)

			for _, f := range tt.files {
				filePath := filepath.Join(tmpDir, f)
				os.WriteFile(filePath, []byte("content"), 0644)
			}

			dir := builder.Build()
			result := dir.Exists(tt.filename)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Directory.Exists Tests - Nested Paths
// ============================================================================

func Test_Exist_NestedPaths(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(dir string)
		filename string
		expected bool
	}{
		{
			name: "file in subdirectory exists",
			setup: func(dir string) {
				subdir := filepath.Join(dir, "subdir")
				os.Mkdir(subdir, 0755)
				filePath := filepath.Join(subdir, "data.yaml")
				os.WriteFile(filePath, []byte("content"), 0644)
			},
			filename: "subdir/data.yaml",
			expected: true,
		},
		{
			name: "file in nested subdirectory exists",
			setup: func(dir string) {
				subdir := filepath.Join(dir, "a", "b", "c")
				os.MkdirAll(subdir, 0755)
				filePath := filepath.Join(subdir, "deep.yaml")
				os.WriteFile(filePath, []byte("content"), 0644)
			},
			filename: "a/b/c/deep.yaml",
			expected: true,
		},
		{
			name: "file in subdirectory does not exist",
			setup: func(dir string) {
				subdir := filepath.Join(dir, "subdir")
				os.Mkdir(subdir, 0755)
			},
			filename: "subdir/data.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dir := NewDirectory(tmpDir)

			if tt.setup != nil {
				tt.setup(tmpDir)
			}

			result := dir.Exists(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
