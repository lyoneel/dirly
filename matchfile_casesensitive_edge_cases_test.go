package dirly

import (
	"testing"
)

// ============================================================================
// matchFileCaseSensitive Edge Cases - Mixed Case Files and Paths
// ============================================================================

func Test_matchFileCaseSensitive_MixedCaseFiles(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "exact case match",
			filePath:    "/tmp/test/Config.yaml",
			patterns:    []string{"Config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "case mismatch - uppercase pattern lowercase file",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"Config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "case mismatch - lowercase pattern uppercase file",
			filePath:    "/tmp/test/CONFIG.YAML",
			patterns:    []string{"config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "case mismatch - mixed case pattern",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"Config.YAML"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "exact mixed case match",
			filePath:    "/tmp/test/MyConfig.yaml",
			patterns:    []string{"MyConfig.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "case mismatch in middle of filename",
			filePath:    "/tmp/test/myconfig.yaml",
			patterns:    []string{"MyConfig.yaml"},
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

func Test_matchFileCaseSensitive_MixedCasePaths(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "exact case path match nested",
			filePath:    "/tmp/test/Config/Subdir/Data.yaml",
			patterns:    []string{"Config/Subdir/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true,
		},
		{
			name:        "case mismatch in path component",
			filePath:    "/tmp/test/config/subdir/data.yaml",
			patterns:    []string{"Config/Subdir/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    false,
		},
		{
			name:        "case mismatch in nested filename",
			filePath:    "/tmp/test/Config/subdir/data.yaml",
			patterns:    []string{"Config/SubDir/Data.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    false,
		},
		{
			name:        "exact case nested path match",
			filePath:    "/tmp/test/MyProject/src/main.go",
			patterns:    []string{"MyProject/src/*.go"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true,
		},
		{
			name:        "case mismatch in deep nested path",
			filePath:    "/tmp/test/myproject/src/main.go",
			patterns:    []string{"MyProject/Src/Main.go"},
			basePath:    "/tmp/test",
			matchNested: true,
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

func Test_matchFileCaseSensitive_WildcardsWithMixedCase(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "wildcard with mixed case file",
			filePath:    "/tmp/test/Config.yaml",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true, // wildcard matches regardless of case in filename
		},
		{
			name:        "wildcard with mixed case extension (Linux case-sensitive)",
			filePath:    "/tmp/test/config.YAML",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // filepath.Match is case-sensitive on Linux
		},
		{
			name:        "wildcard with exact case extension match",
			filePath:    "/tmp/test/config.yaml",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "wildcard with wrong case extension (Linux)",
			filePath:    "/tmp/test/config.YAML",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // case-sensitive on Linux
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

func Test_matchFileCaseSensitive_RecursivePatterns(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "recursive pattern exact case match",
			filePath:    "/tmp/test/Config/subdir/Data.yaml",
			patterns:    []string{"**/Config/**/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true,
		},
		{
			name:        "recursive pattern case mismatch in path",
			filePath:    "/tmp/test/config/subdir/data.yaml",
			patterns:    []string{"**/Config/**/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    false,
		},
		{
			name:        "recursive pattern with mixed case filename",
			filePath:    "/tmp/test/config/SubDir/Data.yaml",
			patterns:    []string{"**/subdir/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    false, // subdir != SubDir in case-sensitive mode
		},
		{
			name:        "recursive pattern exact match root level",
			filePath:    "/tmp/test/Config.yaml",
			patterns:    []string{"**/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true,
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

func Test_matchFileCaseSensitive_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "hidden file exact case match",
			filePath:    "/tmp/test/.Gitignore",
			patterns:    []string{".gitignore"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // .Gitignore != .gitignore
		},
		{
			name:        "hidden file exact case match correct",
			filePath:    "/tmp/test/.gitignore",
			patterns:    []string{".gitignore"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "file with dots exact case match",
			filePath:    "/tmp/test/My.Config.yaml",
			patterns:    []string{"my.config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "file with dots exact case match correct",
			filePath:    "/tmp/test/My.Config.yaml",
			patterns:    []string{"My.Config.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
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

func Test_matchFileCaseSensitive_CharacterClasses(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "character class uppercase range match",
			filePath:    "/tmp/test/fileA.txt",
			patterns:    []string{"file[A-Z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "character class lowercase range no match uppercase file",
			filePath:    "/tmp/test/fileA.txt",
			patterns:    []string{"file[a-z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "character class lowercase range match",
			filePath:    "/tmp/test/filea.txt",
			patterns:    []string{"file[a-z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "character class mixed case range match uppercase",
			filePath:    "/tmp/test/fileA.txt",
			patterns:    []string{"file[A-Za-z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "character class mixed case range match lowercase",
			filePath:    "/tmp/test/filea.txt",
			patterns:    []string{"file[A-Za-z].txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
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

func Test_matchFileCaseSensitive_QuestionMarkWildcards(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "question mark single char match",
			filePath:    "/tmp/test/file1.txt",
			patterns:    []string{"file?.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
		},
		{
			name:        "question mark case sensitive match",
			filePath:    "/tmp/test/File1.txt",
			patterns:    []string{"file?.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // file != File
		},
		{
			name:        "question mark case sensitive match correct",
			filePath:    "/tmp/test/file1.txt",
			patterns:    []string{"file?.txt"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true,
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

func Test_matchFileCaseSensitive_EmptyAndNilInputs(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "nil patterns slice",
			filePath:    "/tmp/test/Config.yaml",
			patterns:    nil,
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false,
		},
		{
			name:        "empty patterns slice",
			filePath:    "/tmp/test/Config.yaml",
			patterns:    []string{},
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
// Additional Edge Cases for matchFileCaseSensitive - Path Patterns and Component Matching
// ============================================================================

func Test_matchFileCaseSensitive_NonNestedWithPathPatterns(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "non-nested with path pattern match",
			filePath:    "/tmp/test/config/data.yaml",
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true, // Path pattern should still work even when matchNested=false
		},
		{
			name:        "non-nested with path pattern no match wrong dir",
			filePath:    "/tmp/test/src/data.yaml",
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // Wrong directory
		},
		{
			name:        "non-nested with path pattern case mismatch",
			filePath:    "/tmp/test/Config/data.yaml",
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // Case-sensitive: Config != config
		},
		{
			name:        "non-nested with path pattern exact case match",
			filePath:    "/tmp/test/config/data.yaml",
			patterns:    []string{"config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    true, // Exact case match
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

func Test_matchFileCaseSensitive_RecursivePatternsWithMatchNestedFalse(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "recursive pattern with matchNested=false (not supported)",
			filePath:    "/tmp/test/config/data.yaml",
			patterns:    []string{"**/config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // **/ patterns require matchNested=true to work properly
		},
		{
			name:        "recursive pattern case mismatch with matchNested=false",
			filePath:    "/tmp/test/Config/data.yaml",
			patterns:    []string{"**/config/*.yaml"},
			basePath:    "/tmp/test",
			matchNested: false,
			expected:    false, // Case-sensitive: Config != config
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

func Test_matchFileCaseSensitive_ComponentMatching(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "component matching exact case",
			filePath:    "/tmp/test/Config/subdir/data.yaml",
			patterns:    []string{"data.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true, // Matches filename at end of path
		},
		{
			name:        "component matching case mismatch",
			filePath:    "/tmp/test/config/subdir/data.yaml",
			patterns:    []string{"Data.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    false, // Case-sensitive: Data != data
		},
		{
			name:        "component matching path component case mismatch",
			filePath:    "/tmp/test/Config/subdir/data.yaml",
			patterns:    []string{"config"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    false, // Case-sensitive: Config != config
		},
		{
			name:        "component matching path component exact case",
			filePath:    "/tmp/test/Config/subdir/data.yaml",
			patterns:    []string{"Config"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true, // Exact match on path component
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

func Test_matchFileCaseSensitive_FilenameExtraction(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		patterns    []string
		basePath    string
		matchNested bool
		expected    bool
	}{
		{
			name:        "filename extraction exact case",
			filePath:    "/tmp/test/Config/subdir/Data.yaml",
			patterns:    []string{"Data.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true, // Extracted filename matches exactly
		},
		{
			name:        "filename extraction case mismatch",
			filePath:    "/tmp/test/Config/subdir/data.yaml",
			patterns:    []string{"Data.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    false, // Case-sensitive: Data != data
		},
		{
			name:        "filename with wildcard extraction",
			filePath:    "/tmp/test/Config/subdir/Data.yaml",
			patterns:    []string{"*.yaml"},
			basePath:    "/tmp/test",
			matchNested: true,
			expected:    true, // Wildcard matches extracted filename
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
