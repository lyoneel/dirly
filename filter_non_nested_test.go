package dirly

import (
	"testing"
)

// ============================================================================
// matchFile Tests - Non-Nested with Path Patterns
// ============================================================================

func Test_matchFile_NonNestedWithPathPattern(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "path pattern matches subdirectory",
			filePath:    "/tmp/test/config/data.yaml",
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "path pattern does not match deeper subdirectory",
			filePath:    "/tmp/test/config/subdir/data.yaml",
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // only direct children of config/ match
		},
		{
			name:        "path pattern does not match root directory file",
			filePath:    "/tmp/test/data.yaml",
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "simple pattern matches filename only",
			filePath:    "/tmp/test/config/data.yaml",
			patterns:    []string{"data.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "wildcard pattern matches filename only",
			filePath:    "/tmp/test/config/data.yaml",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchFile(tt.filePath, tt.patterns, tt.basePath, tt.matchNested)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// matchFiles Non-Nested Tests with Path Patterns
// ============================================================================

func Test_matchFiles_NonNestedWithPathPattern(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		patterns    []string
		basePath    string
		matchNested bool
		expectedLen int
	}{
		{
			name:        "path pattern filters to specific subdirectory",
			files:       []string{"config/data.yaml", "config/settings.yaml", "data.json"},
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 2, // only config/ files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchFiles(tt.files, tt.patterns, tt.basePath, tt.matchNested)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d matches, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

// ============================================================================
// filterOutFiles Non-Nested Tests with Path Patterns
// ============================================================================

func Test_filterOutFiles_NonNestedWithPathPattern(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		patterns    []string
		basePath    string
		matchNested bool
		expectedLen int
	}{
		{
			name:        "exclude specific subdirectory",
			files:       []string{"config/data.yaml", "data.json"},
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 1, // config/ files excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterOutFiles(tt.files, tt.patterns, tt.basePath, tt.matchNested)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files after filtering, got %d", tt.expectedLen, len(result))
			}
		})
	}
}
