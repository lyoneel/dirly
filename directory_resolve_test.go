package dirly

import (
	"path/filepath"
	"testing"
)

// ============================================================================
// resolvePath Tests - Valid Paths
// ============================================================================

func TestDirectory_resolvePath_Valid(t *testing.T) {
	tests := []struct {
		name       string
		basePath   string
		relative   string
		wantExists bool
	}{
		{
			name:       "simple relative path",
			basePath:   "/tmp/test",
			relative:   "config.yaml",
			wantExists: true,
		},
		{
			name:       "path with subdirectory",
			basePath:   "/tmp/test",
			relative:   "subdir/config.yaml",
			wantExists: true,
		},
		{
			name:       "current directory reference",
			basePath:   "/tmp/test",
			relative:   "./config.yaml",
			wantExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tt.basePath)

			result, err := dir.resolvePath(tt.relative)
			if err != nil {
				t.Errorf("resolvePath() unexpected error: %v", err)
			}

			if !tt.wantExists && result == "" {
				return
			}

			expected := filepath.Join(tt.basePath, tt.relative)
			if result != expected {
				t.Errorf("expected resolved path to be %q, got %q", expected, result)
			}
		})
	}
}

// ============================================================================
// resolvePath Tests - Empty Path
// ============================================================================

func TestDirectory_resolvePath_Empty(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		relative string
	}{
		{
			name:     "empty relative path",
			basePath: "/tmp/test",
			relative: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tt.basePath)

			result, err := dir.resolvePath(tt.relative)
			if err == nil {
				t.Error("resolvePath() should return error for empty path")
			}

			if result != "" {
				t.Errorf("expected empty result, got %q", result)
			}
		})
	}
}

// ============================================================================
// resolvePath Tests - Absolute Paths (Should Fail)
// ============================================================================

func TestDirectory_resolvePath_Absolute(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		relative string
	}{
		{
			name:     "absolute path",
			basePath: "/tmp/test",
			relative: "/etc/passwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tt.basePath)

			result, err := dir.resolvePath(tt.relative)
			if err == nil {
				t.Error("resolvePath() should reject absolute paths")
			}

			if result != "" {
				t.Errorf("expected empty result for absolute path, got %q", result)
			}
		})
	}
}

// ============================================================================
// resolvePath Tests - Path Traversal (Should Fail)
// ============================================================================

func TestDirectory_resolvePath_PathTraversal(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		relative string
	}{
		{
			name:     "parent directory traversal",
			basePath: "/tmp/test",
			relative: "../etc/passwd",
		},
		{
			name:     "multiple parent traversals",
			basePath: "/tmp/test",
			relative: "../../etc/passwd",
		},
		{
			name:     "traversal in middle of path",
			basePath: "/tmp/test",
			relative: "config/../../etc/passwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tt.basePath)

			result, err := dir.resolvePath(tt.relative)
			if err == nil {
				t.Error("resolvePath() should reject path traversal attempts")
			}

			if result != "" {
				t.Errorf("expected empty result for path traversal, got %q", result)
			}
		})
	}
}
