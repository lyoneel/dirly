package dirly

import (
	"testing"
)

// ============================================================================
// InvalidPatternError Tests
// ============================================================================

func TestInvalidPatternError_Error(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		reason   string
		expected string
	}{
		{
			name:     "simple pattern",
			pattern:  "*.yaml",
			reason:   "patterns cannot start with '!'",
			expected: "invalid pattern: *.yaml - patterns cannot start with '!'",
		},
		{
			name:     "complex pattern",
			pattern:  "!config/*.yaml",
			reason:   "invalid syntax",
			expected: "invalid pattern: !config/*.yaml - invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &InvalidPatternError{
				Pattern: tt.pattern,
				Reason:  tt.reason,
			}

			result := err.Error()
			if result != tt.expected {
				t.Errorf("expected Error() to return %q, got %q", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// validatePattern Tests
// ============================================================================

func Test_validatePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{
			name:    "valid pattern",
			pattern: "*.yaml",
			wantErr: false,
		},
		{
			name:    "valid pattern with path",
			pattern: "config/*.yaml",
			wantErr: false,
		},
		{
			name:    "pattern starting with !",
			pattern: "!*.yaml",
			wantErr: true,
		},
		{
			name:    "complex pattern starting with !",
			pattern: "!config/*.yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePattern(tt.pattern)

			if (err != nil) != tt.wantErr {
				t.Errorf("validatePattern() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				if _, ok := err.(*InvalidPatternError); !ok {
					t.Errorf("expected InvalidPatternError, got %T", err)
				}
			}
		})
	}
}

// ============================================================================
// matchFile Tests - Basic Matching
// ============================================================================

func Test_matchFile_Basic(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "exact filename match",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "wildcard match",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "no match",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"*.json"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
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
// matchFile Tests - Nested Matching
// ============================================================================

func Test_matchFile_Nested(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "nested path match",
			filePath:    "/tmp/test/config/subdir/config.yaml",
			patterns:    []string{"config.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true,
		},
		{
			name:        "nested wildcard match",
			filePath:    "/tmp/test/config/subdir/data.yaml",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true,
		},
		{
			name:        "full path match",
			filePath:    "/tmp/test/config/subdir/data.yaml",
			patterns:    []string{"config/subdir/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
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
// matchFile Tests - Recursive Matching (**/)
// ============================================================================

func Test_matchFile_Recursive(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "recursive match root file",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"**/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true,
		},
		{
			name:        "recursive match nested file",
			filePath:    "/tmp/test/config/subdir/data.yaml",
			patterns:    []string{"**/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true,
		},
		{
			name:        "recursive match specific path",
			filePath:    "/tmp/test/config/data.yaml",
			patterns:    []string{"**/config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
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
// matchFile Tests - Edge Cases
// ============================================================================

func Test_matchFile_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "empty patterns",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "nil patterns",
			filePath:    "/tmp/test/config.yaml",
			patterns:    nil,
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "empty filePath",
			filePath:    "",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "pattern with special chars",
			filePath:    "/tmp/test/config.yaml.bak",
			patterns:    []string{"*.yaml*"},
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
// matchFiles Tests
// ============================================================================

func Test_matchFiles(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		patterns    []string
		basePath    string
		matchNested bool
		expectedLen int
	}{
		{
			name:        "match multiple files",
			files:       []string{"config.yaml", "data.json", "readme.txt"},
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 1,
		},
		{
			name:        "no matches",
			files:       []string{"config.yaml", "data.json", "readme.txt"},
			patterns:    []string{"*.xml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 0,
		},
		{
			name:        "all match",
			files:       []string{"config.yaml", "data.yaml", "test.yaml"},
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 3,
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
// filterOutFiles Tests
// ============================================================================

func Test_filterOutFiles(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		patterns    []string
		basePath    string
		matchNested bool
		expectedLen int
	}{
		{
			name:        "filter out matching files",
			files:       []string{"config.yaml", "temp.yaml", "data.json"},
			patterns:    []string{"*.tmp"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 3, // none match *.tmp
		},
		{
			name:        "filter out some files",
			files:       []string{"config.yaml", "temp.tmp", "data.json"},
			patterns:    []string{"*.tmp"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 2, // temp.tmp is filtered out
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

// ============================================================================
// matchPatternInPath Tests
// ============================================================================

func Test_matchPatternInPath(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		pattern  string
		expected bool
	}{
		{
			name:     "exact path match",
			relPath:  "config/data.yaml",
			pattern:  "data.yaml",
			expected: true,
		},
		{
			name:     "wildcard in path",
			relPath:  "config/data.yaml",
			pattern:  "*.yaml",
			expected: true,
		},
		{
			name:     "no match",
			relPath:  "config/data.json",
			pattern:  "*.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPatternInPath(tt.relPath, tt.pattern)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// shouldMatch Tests
// ============================================================================

func Test_shouldMatch(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "simple pattern match",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{"*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive pattern match",
			filePath:      "/tmp/test/config/subdir/data.yaml",
			patterns:      []string{"**/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldMatch(tt.filePath, tt.patterns, tt.matchNested, tt.caseSensitive)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// matchFileCaseSensitive Tests
// ============================================================================

func Test_matchFileCaseSensitive(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "case sensitive exact match",
			filePath:    "/tmp/test/Config.yaml",
			patterns:    []string{"Config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "case sensitive no match",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"Config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchFileCaseSensitive(tt.filePath, tt.patterns, tt.basePath, tt.matchNested)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// filterOutFilesCaseSensitive Tests
// ============================================================================

func Test_filterOutFilesCaseSensitive(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		patterns    []string
		basePath    string
		matchNested bool
		expectedLen int
	}{
		{
			name:        "case sensitive filter",
			files:       []string{"Config.yaml", "config.yaml"},
			patterns:    []string{"Config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 1, // only Config.yaml is filtered out
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterOutFilesCaseSensitive(tt.files, tt.patterns, tt.basePath, tt.matchNested)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files after filtering, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

// ============================================================================
// matchFilesCaseSensitive Tests
// ============================================================================

func Test_matchFilesCaseSensitive(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		patterns    []string
		basePath    string
		matchNested bool
		expectedLen int
	}{
		{
			name:        "case sensitive match",
			files:       []string{"Config.yaml", "config.yaml"},
			patterns:    []string{"Config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 1, // only Config.yaml matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchFilesCaseSensitive(tt.files, tt.patterns, tt.basePath, tt.matchNested)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d matches, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

// ============================================================================
// addFiles Tests
// ============================================================================

func Test_addFiles(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		pattern     string
		basePath    string
		matchNested bool
		expectedLen int
	}{
		{
			name:        "add matching file",
			files:       []string{"config.yaml"},
			pattern:     "*.yaml",
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 1, // no duplicates added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addFiles(tt.files, tt.pattern, tt.basePath, tt.matchNested)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

// ============================================================================
// filterByExtension Tests
// ============================================================================

func Test_filterByExtension(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		extensions  []string
		basePath    string
		matchNested bool
		expectedLen int
	}{
		{
			name:        "filter by single extension",
			files:       []string{"config.yaml", "data.json", "readme.txt"},
			extensions:  []string{"yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 1,
		},
		{
			name:        "filter by multiple extensions",
			files:       []string{"config.yaml", "data.json", "readme.txt"},
			extensions:  []string{"yaml", "json"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 2,
		},
		{
			name:        "case insensitive extension match",
			files:       []string{"config.YAML", "data.json"},
			extensions:  []string{"yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByExtension(tt.files, tt.extensions, tt.basePath, tt.matchNested)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files after filtering, got %d", tt.expectedLen, len(result))
			}
		})
	}
}
