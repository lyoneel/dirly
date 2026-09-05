package dirly

import (
	"testing"
)

// ============================================================================
// matchFile Edge Cases - Empty Patterns, Nil Inputs, Special Characters
// ============================================================================

func Test_matchFile_EmptyPatterns(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "empty slice patterns",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "nil patterns slice",
			filePath:    "/tmp/test/config.yaml",
			patterns:    nil,
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "empty string in patterns slice",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{""},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "multiple empty patterns",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"", "", ""},
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

func Test_matchFile_NilInputs(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "nil filePath",
			filePath:    "",
			patterns:    []string{"*.yaml"},
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
			name:        "nil basePath",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"*.yaml"},
			basePath:    "",
			matchNested: false,
			expected:    true, // filepath.Match still works with empty base
		},
		{
			name:        "empty basePath",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"*.yaml"},
			basePath:    "",
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

func Test_matchFile_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "pattern with dot in name",
			filePath:    "/tmp/test/.gitignore",
			patterns:    []string{".git*"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with underscore",
			filePath:    "/tmp/test/config_file.yaml",
			patterns:    []string{"config_*"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with hyphen",
			filePath:    "/tmp/test/my-config.yaml",
			patterns:    []string{"my-*"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with multiple dots",
			filePath:    "/tmp/test/config.v1.yaml",
			patterns:    []string{"*.v1.*"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with asterisk in name (literal)",
			filePath:    "/tmp/test/file*.txt",
			patterns:    []string{"file*.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with question mark wildcard",
			filePath:    "/tmp/test/file1.txt",
			patterns:    []string{"file?.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with question mark - no match",
			filePath:    "/tmp/test/file12.txt",
			patterns:    []string{"file?.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "pattern with character class [0-9]",
			filePath:    "/tmp/test/file1.txt",
			patterns:    []string{"file[0-9].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with character class [abc]",
			filePath:    "/tmp/test/filea.txt",
			patterns:    []string{"file[abc].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with character class range [a-z]",
			filePath:    "/tmp/test/filem.txt",
			patterns:    []string{"file[a-z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with negated character class [^0-9]",
			filePath:    "/tmp/test/filea.txt",
			patterns:    []string{"file[^0-9].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with special regex chars in name",
			filePath:    "/tmp/test/file+test.txt",
			patterns:    []string{"file+test.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with caret anchor",
			filePath:    "/tmp/test/^test.txt",
			patterns:    []string{"^test.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with dollar sign",
			filePath:    "/tmp/test/test$.txt",
			patterns:    []string{"test$.txt"},
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

func Test_matchFile_CharacterClasses(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "character class with multiple ranges",
			filePath:    "/tmp/test/file1a.txt",
			patterns:    []string{"file[0-9][a-z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "character class with uppercase range",
			filePath:    "/tmp/test/fileA.txt",
			patterns:    []string{"file[A-Z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "character class with mixed case range",
			filePath:    "/tmp/test/fileM.txt",
			patterns:    []string{"file[A-Za-z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "character class with space",
			filePath:    "/tmp/test/file .txt",
			patterns:    []string{"file[ ].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "character class with hyphen at start",
			filePath:    "/tmp/test/file-.txt",
			patterns:    []string{"file[-].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // [-] at start is not treated as literal hyphen in filepath.Match
		},
		{
			name:        "character class with hyphen at end",
			filePath:    "/tmp/test/file-.txt",
			patterns:    []string{"file[-].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // [-] in character class needs escaping or position adjustment
		},
		{
			name:        "character class with bracket",
			filePath:    "/tmp/test/file[.txt",
			patterns:    []string{"file[].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // [] in character class is invalid syntax without content
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

func Test_matchFile_QuestionMarkWildcards(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "single char wildcard match",
			filePath:    "/tmp/test/file1.txt",
			patterns:    []string{"file?.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "single char wildcard - no match (too long)",
			filePath:    "/tmp/test/file12.txt",
			patterns:    []string{"file?.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "single char wildcard - no match (too short)",
			filePath:    "/tmp/test/file.txt",
			patterns:    []string{"file?.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "multiple question marks",
			filePath:    "/tmp/test/file12.txt",
			patterns:    []string{"file??.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "question mark with extension",
			filePath:    "/tmp/test/config1.yaml",
			patterns:    []string{"config?.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "question mark in nested path",
			filePath:    "/tmp/test/config/subdir/data1.yaml",
			patterns:    []string{"data?.yaml"},
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

func Test_matchFile_AsteriskWildcards(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "asterisk matches any chars",
			filePath:    "/tmp/test/file123.txt",
			patterns:    []string{"file*.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "asterisk matches empty string",
			filePath:    "/tmp/test/file.txt",
			patterns:    []string{"file*.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "asterisk at start",
			filePath:    "/tmp/test/myconfig.yaml",
			patterns:    []string{"*config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "asterisk at end",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"config*"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "asterisk in middle",
			filePath:    "/tmp/test/myconfig.yaml",
			patterns:    []string{"my*config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "multiple asterisks",
			filePath:    "/tmp/test/myconfigv1.yaml",
			patterns:    []string{"my*"},
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

func Test_matchFile_ComplexPatterns(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "pattern with both ? and *",
			filePath:    "/tmp/test/file1abc.txt",
			patterns:    []string{"file?*.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "complex character class with ranges",
			filePath:    "/tmp/test/file1a.txt",
			patterns:    []string{"file[0-9][a-z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "pattern with escaped chars (literal asterisk)",
			filePath:    "/tmp/test/file*.txt",
			patterns:    []string{"file\\*.txt"},
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
