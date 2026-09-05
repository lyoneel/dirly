package dirly

import (
	"testing"
)

// ============================================================================
// Directory.applyFilters Direct Unit Tests - Edge Cases and Complex Scenarios
// ============================================================================

func Test_applyFilters_EmptyAndNil(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		include       []string
		exclude       []string
		extensions    []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "empty files slice",
			files:         []string{},
			include:       nil,
			exclude:       nil,
			extensions:    nil,
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   0,
		},
		{
			name:          "nil files slice",
			files:         nil,
			include:       nil,
			exclude:       nil,
			extensions:    nil,
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   0,
		},
		{
			name:          "no filters configured",
			files:         []string{"a.yaml", "b.json"},
			include:       nil,
			exclude:       nil,
			extensions:    nil,
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   2, // all files pass through
		},
		{
			name:          "empty filter configs",
			files:         []string{"a.yaml", "b.json"},
			include:       []string{},
			exclude:       []string{},
			extensions:    []string{},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   2, // all files pass through
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
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

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_IncludeOnly(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		include       []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "include single pattern",
			files:         []string{"a.yaml", "b.json", "c.txt"},
			include:       []string{"*.yaml"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   1, // only a.yaml matches
		},
		{
			name:          "include multiple patterns",
			files:         []string{"a.yaml", "b.json", "c.txt", "d.yaml"},
			include:       []string{"*.yaml", "*.json"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   3, // a.yaml, b.json, d.yaml match
		},
		{
			name:          "include with nested matching",
			files:         []string{"config/a.yaml", "src/b.go"},
			include:       []string{"*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   1, // config/a.yaml matches by filename
		},
		{
			name:          "include path-specific pattern",
			files:         []string{"config/a.yaml", "src/b.yaml"},
			include:       []string{"config/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   1, // only config/a.yaml matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.include) > 0 {
				builder = builder.Include(tt.include...)
			}
			if !tt.matchNested {
				builder = builder.MatchNested(false)
			}
			if tt.caseSensitive {
				builder = builder.CaseSensitive(true)
			}
			dir := builder.Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_ExcludeOnly(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		exclude       []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "exclude single pattern",
			files:         []string{"a.yaml", "b.tmp", "c.json"},
			exclude:       []string{"*.tmp"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   2, // b.tmp is excluded
		},
		{
			name:          "exclude multiple patterns",
			files:         []string{"a.yaml", "b.tmp", "c.json", "d.tmp"},
			exclude:       []string{"*.tmp", "*.bak"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   2, // b.tmp and d.tmp are excluded
		},
		{
			name:          "exclude with nested matching",
			files:         []string{"config/a.yaml", "temp/b.tmp"},
			exclude:       []string{"*.tmp"},
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   1, // temp/b.tmp is excluded
		},
		{
			name:          "exclude path-specific pattern",
			files:         []string{"config/a.yaml", "temp/b.yaml"},
			exclude:       []string{"temp/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   1, // temp/b.yaml is excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.exclude) > 0 {
				builder = builder.Exclude(tt.exclude...)
			}
			if !tt.matchNested {
				builder = builder.MatchNested(false)
			}
			if tt.caseSensitive {
				builder = builder.CaseSensitive(true)
			}
			dir := builder.Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_IncludeAndExclude(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		include       []string
		exclude       []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "include then exclude",
			files:         []string{"a.yaml", "b.tmp", "c.json"},
			include:       []string{"*.yaml", "*.tmp"},
			exclude:       []string{"*.tmp"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   1, // include *.yaml and *.tmp, then exclude *.tmp = only a.yaml
		},
		{
			name:          "include specific files exclude some",
			files:         []string{"a.yaml", "b.yaml", "c.tmp"},
			include:       []string{"*.yaml"},
			exclude:       []string{"b.yaml"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   1, // only a.yaml remains
		},
		{
			name:          "include recursive exclude specific",
			files:         []string{"config/a.yaml", "src/b.yaml", "temp/c.tmp"},
			include:       []string{"**/*.yaml"},
			exclude:       []string{"**/src/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   1, // include all yaml, exclude src/*.yaml = only config/a.yaml
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.include) > 0 {
				builder = builder.Include(tt.include...)
			}
			if len(tt.exclude) > 0 {
				builder = builder.Exclude(tt.exclude...)
			}
			if !tt.matchNested {
				builder = builder.MatchNested(false)
			}
			if tt.caseSensitive {
				builder = builder.CaseSensitive(true)
			}
			dir := builder.Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_ExtensionsOnly(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		extensions    []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "single extension filter",
			files:         []string{"a.yaml", "b.json", "c.txt"},
			extensions:    []string{"yaml"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   1, // only a.yaml has .yaml extension
		},
		{
			name:          "multiple extensions filter",
			files:         []string{"a.yaml", "b.json", "c.txt"},
			extensions:    []string{"yaml", "json"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   2, // a.yaml and b.json match
		},
		{
			name:          "case insensitive extension match",
			files:         []string{"a.YAML", "b.Json", "c.txt"},
			extensions:    []string{"yaml", "json"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   2, // a.YAML and b.Json match (case insensitive)
		},
		{
			name:          "file without extension excluded",
			files:         []string{"a.yaml", "b.json", "readme"},
			extensions:    []string{"yaml", "json"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   2, // readme has no extension so it's excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.extensions) > 0 {
				builder = builder.WithExtensions(tt.extensions...)
			}
			if !tt.matchNested {
				builder = builder.MatchNested(false)
			}
			if tt.caseSensitive {
				builder = builder.CaseSensitive(true)
			}
			dir := builder.Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_AllFiltersCombined(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		include       []string
		exclude       []string
		extensions    []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "all three filters",
			files:         []string{"a.yaml", "b.tmp", "c.json", "d.yaml"},
			include:       []string{"*.yaml", "*.json"},
			exclude:       []string{"b.tmp"}, // b.tmp doesn't match include anyway
			extensions:    []string{"yaml"},
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   2, // include *.yaml and *.json = a.yaml, c.json, d.yaml; exclude *.tmp (no effect); filter extensions to yaml = a.yaml, d.yaml
		},
		{
			name:          "include path with extension",
			files:         []string{"config/a.yaml", "src/b.go", "temp/c.tmp"},
			include:       []string{"config/*.yaml"},
			exclude:       nil,
			extensions:    []string{"yaml"},
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   1, // only config/a.yaml matches include and extension filter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.include) > 0 {
				builder = builder.Include(tt.include...)
			}
			if len(tt.exclude) > 0 {
				builder = builder.Exclude(tt.exclude...)
			}
			if len(tt.extensions) > 0 {
				builder = builder.WithExtensions(tt.extensions...)
			}
			if !tt.matchNested {
				builder = builder.MatchNested(false)
			}
			if tt.caseSensitive {
				builder = builder.CaseSensitive(true)
			}
			dir := builder.Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_RecursivePatterns(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		include       []string
		exclude       []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "include recursive pattern",
			files:         []string{"a.yaml", "config/b.yaml", "src/c.go"},
			include:       []string{"**/*.yaml"},
			exclude:       nil,
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   2, // a.yaml and config/b.yaml match **/*.yaml
		},
		{
			name:          "exclude recursive pattern",
			files:         []string{"a.yaml", "config/b.yaml", "src/c.go"},
			include:       nil,
			exclude:       []string{"**/config/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   2, // exclude config/b.yaml = a.yaml and src/c.go remain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.include) > 0 {
				builder = builder.Include(tt.include...)
			}
			if len(tt.exclude) > 0 {
				builder = builder.Exclude(tt.exclude...)
			}
			dir := builder.Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_CaseSensitive(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		include       []string
		exclude       []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "case sensitive include",
			files:         []string{"Config.yaml", "config.yaml"},
			include:       []string{"Config.yaml"},
			exclude:       nil,
			matchNested:   false,
			caseSensitive: true,
			expectedLen:   1, // only Config.yaml matches (case sensitive)
		},
		{
			name:          "case sensitive exclude",
			files:         []string{"Config.yaml", "config.yaml"},
			include:       nil,
			exclude:       []string{"Config.yaml"},
			matchNested:   false,
			caseSensitive: true,
			expectedLen:   1, // only config.yaml remains (Config.yaml excluded)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.include) > 0 {
				builder = builder.Include(tt.include...)
			}
			if len(tt.exclude) > 0 {
				builder = builder.Exclude(tt.exclude...)
			}
			dir := builder.CaseSensitive(true).Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_PathSpecificPatterns(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		include       []string
		exclude       []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "include path-specific pattern",
			files:         []string{"config/a.yaml", "src/b.yaml"},
			include:       []string{"config/*.yaml"},
			exclude:       nil,
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   1, // only config/a.yaml matches path-specific pattern
		},
		{
			name:          "exclude path-specific pattern",
			files:         []string{"config/a.yaml", "src/b.yaml"},
			include:       nil,
			exclude:       []string{"src/*.yaml"},
			matchNested:   true,
			caseSensitive: false,
			expectedLen:   1, // exclude src/b.yaml = only config/a.yaml remains
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.include) > 0 {
				builder = builder.Include(tt.include...)
			}
			if len(tt.exclude) > 0 {
				builder = builder.Exclude(tt.exclude...)
			}
			dir := builder.Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func Test_applyFilters_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		include       []string
		exclude       []string
		matchNested   bool
		caseSensitive bool
		expectedLen   int
	}{
		{
			name:          "include hidden files",
			files:         []string{".gitignore", "config.yaml"},
			include:       []string{".git*"},
			exclude:       nil,
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   1, // only .gitignore matches .git*
		},
		{
			name:          "exclude files with special chars",
			files:         []string{"config.yaml.bak", "config.yaml"},
			exclude:       []string{"*.bak"},
			include:       nil,
			matchNested:   false,
			caseSensitive: false,
			expectedLen:   1, // exclude config.yaml.bak = only config.yaml remains
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test")
			if len(tt.include) > 0 {
				builder = builder.Include(tt.include...)
			}
			if len(tt.exclude) > 0 {
				builder = builder.Exclude(tt.exclude...)
			}
			dir := builder.Build()

			result := dir.applyFilters(tt.files)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(result))
			}
		})
	}
}
