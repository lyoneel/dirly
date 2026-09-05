package dirly

import (
	"testing"
)

// ============================================================================
// shouldMatch Edge Cases - Recursive Patterns, Empty Inputs, Multiple Combinations
// ============================================================================

func Test_shouldMatch_EmptyInputs(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "nil patterns slice",
			filePath:      "/tmp/test/config.yaml",
			patterns:      nil,
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "empty patterns slice",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "empty filePath with nil patterns",
			filePath:      "",
			patterns:      nil,
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "empty filePath with empty patterns",
			filePath:      "",
			patterns:      []string{},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "empty string in patterns slice (filepath.Match matches empty pattern)",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{""},
			matchNested:   true,
			caseSensitive: false,
			expected:      true, // filepath.Match("", ...) returns true - empty pattern matches everything
		},
		{
			name:          "multiple empty strings in patterns (filepath.Match matches empty pattern)",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{"", "", ""},
			matchNested:   true,
			caseSensitive: false,
			expected:      true, // filepath.Match("", ...) returns true - empty pattern matches everything
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

func Test_shouldMatch_RecursivePatterns(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "recursive pattern root level",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{"**/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive pattern one level deep",
			filePath:      "/tmp/test/config/data.yaml",
			patterns:      []string{"**/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive pattern multiple levels deep",
			filePath:      "/tmp/test/config/subdir/deep/data.yaml",
			patterns:      []string{"**/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive pattern with specific path component",
			filePath:      "/tmp/test/config/data.yaml",
			patterns:      []string{"**/config/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive pattern with specific path component deep",
			filePath:      "/tmp/test/config/subdir/data.yaml",
			patterns:      []string{"**/subdir/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive pattern no match wrong extension",
			filePath:      "/tmp/test/config/data.json",
			patterns:      []string{"**/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "recursive pattern exact path match",
			filePath:      "/tmp/test/config/subdir/data.yaml",
			patterns:      []string{"**/config/subdir/*.yaml"},
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

func Test_shouldMatch_MultiplePatternCombinations(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "multiple patterns first matches",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{"*.yaml", "*.json"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "multiple patterns second matches",
			filePath:      "/tmp/test/data.json",
			patterns:      []string{"*.yaml", "*.json"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "multiple patterns none match",
			filePath:      "/tmp/test/data.txt",
			patterns:      []string{"*.yaml", "*.json"},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "multiple patterns with recursive and simple",
			filePath:      "/tmp/test/config/data.yaml",
			patterns:      []string{"**/*.yaml", "*.json"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "multiple patterns with path-specific and wildcard",
			filePath:      "/tmp/test/config/data.yaml",
			patterns:      []string{"config/*.yaml", "*.json"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "multiple patterns all recursive different paths",
			filePath:      "/tmp/test/src/main.go",
			patterns:      []string{"**/*.go", "**/*.rs"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "multiple patterns complex combination",
			filePath:      "/tmp/test/config/settings.yaml",
			patterns:      []string{"**/config/*.yaml", "*.json", "**/*.toml"},
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

func Test_shouldMatch_PathSpecificPatterns(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "path-specific pattern exact match",
			filePath:      "/tmp/test/config/data.yaml",
			patterns:      []string{"config/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "path-specific pattern no match wrong dir",
			filePath:      "/tmp/test/src/data.yaml",
			patterns:      []string{"config/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "path-specific pattern with wildcard in path",
			filePath:      "/tmp/test/config-prod/data.yaml",
			patterns:      []string{"config-*/data.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "path-specific pattern deep nesting",
			filePath:      "/tmp/test/a/b/c/data.yaml",
			patterns:      []string{"a/b/c/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "path-specific pattern too deep no match",
			filePath:      "/tmp/test/a/b/c/d/data.yaml",
			patterns:      []string{"a/b/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
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

func Test_shouldMatch_SimplePatterns(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "simple wildcard match",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{"*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "simple wildcard no match",
			filePath:      "/tmp/test/config.json",
			patterns:      []string{"*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "simple exact match",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{"config.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "simple exact no match",
			filePath:      "/tmp/test/config.json",
			patterns:      []string{"config.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "simple pattern matches any path component",
			filePath:      "/tmp/test/config/data.yaml",
			patterns:      []string{"data.yaml"},
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

func Test_shouldMatch_Wildcards(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "asterisk wildcard match",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{"*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "question mark single char match",
			filePath:      "/tmp/test/file1.yaml",
			patterns:      []string{"file?.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "question mark no match too long",
			filePath:      "/tmp/test/file12.yaml",
			patterns:      []string{"file?.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
		},
		{
			name:          "character class match",
			filePath:      "/tmp/test/file1.yaml",
			patterns:      []string{"file[0-9].yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "character class no match",
			filePath:      "/tmp/test/filea.yaml",
			patterns:      []string{"file[0-9].yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      false,
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

func Test_shouldMatch_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "case insensitive match",
			filePath:      "/tmp/test/Config.yaml",
			patterns:      []string{"*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "case sensitive exact match",
			filePath:      "/tmp/test/Config.yaml",
			patterns:      []string{"Config.yaml"},
			matchNested:   true,
			caseSensitive: true,
			expected:      true,
		},
		{
			name:          "case sensitive no match different case",
			filePath:      "/tmp/test/config.yaml",
			patterns:      []string{"Config.yaml"},
			matchNested:   true,
			caseSensitive: true,
			expected:      false,
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

func Test_shouldMatch_ComplexRecursivePatterns(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		patterns      []string
		matchNested   bool
		caseSensitive bool
		expected      bool
	}{
		{
			name:          "recursive with path and extension",
			filePath:      "/tmp/test/src/main.go",
			patterns:      []string{"**/src/*.go"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive with multiple path components",
			filePath:      "/tmp/test/src/components/button.go",
			patterns:      []string{"**/src/components/*.go"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive pattern matches any depth",
			filePath:      "/tmp/test/a/b/c/d/e/file.yaml",
			patterns:      []string{"**/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expected:      true,
		},
		{
			name:          "recursive pattern specific component anywhere",
			filePath:      "/tmp/test/a/config/data.yaml",
			patterns:      []string{"**/config/*.yaml"},
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
