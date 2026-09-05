package dirly

import (
	"testing"
)

// ============================================================================
// NewFilteredDirectory Tests
// ============================================================================

func TestNewFilteredDirectory(t *testing.T) {
	tests := []struct {
		name           string
		basePath       string
		expectedConfig FilterConfig
	}{
		{
			name:     "empty base path",
			basePath: "",
			expectedConfig: FilterConfig{
				IncludePatterns:   nil,
				ExcludePatterns:   nil,
				AllowedExtensions: nil,
			},
		},
		{
			name:     "normal base path",
			basePath: "/tmp/test",
			expectedConfig: FilterConfig{
				IncludePatterns:   nil,
				ExcludePatterns:   nil,
				AllowedExtensions: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory(tt.basePath)

			if builder.basePath != tt.basePath {
				t.Errorf("expected basePath to be %q, got %q", tt.basePath, builder.basePath)
			}

			resultConfig := builder.config
			if len(resultConfig.IncludePatterns) != 0 {
				t.Errorf("expected IncludePatterns to be empty, got %v", resultConfig.IncludePatterns)
			}
			if len(resultConfig.ExcludePatterns) != 0 {
				t.Errorf("expected ExcludePatterns to be empty, got %v", resultConfig.ExcludePatterns)
			}
			if len(resultConfig.AllowedExtensions) != 0 {
				t.Errorf("expected AllowedExtensions to be empty, got %v", resultConfig.AllowedExtensions)
			}

			// Verify defaults
			if builder.matchNested != true {
				t.Errorf("expected matchNested to be true by default, got %v", builder.matchNested)
			}
			if builder.caseSensitive != false {
				t.Errorf("expected caseSensitive to be false by default, got %v", builder.caseSensitive)
			}
		})
	}
}

// ============================================================================
// WithExtensions Tests
// ============================================================================

func TestDirectoryBuilder_WithExtensions(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  FilterConfig
		extensions     []string
		expectedLength int
	}{
		{
			name:           "add single extension",
			initialConfig:  FilterConfig{},
			extensions:     []string{"yaml"},
			expectedLength: 1,
		},
		{
			name:           "add multiple extensions",
			initialConfig:  FilterConfig{},
			extensions:     []string{"yaml", "json", "yml"},
			expectedLength: 3,
		},
		{
			name: "append to existing extensions",
			initialConfig: FilterConfig{
				AllowedExtensions: []string{"txt"},
			},
			extensions:     []string{"yaml", "json"},
			expectedLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DirectoryBuilder{
				basePath: "/tmp/test",
				config:   tt.initialConfig,
			}

			result := builder.WithExtensions(tt.extensions...)

			if result != builder {
				t.Error("WithExtensions should return the same builder instance for chaining")
			}

			if len(builder.config.AllowedExtensions) != tt.expectedLength {
				t.Errorf("expected AllowedExtensions length %d, got %d", tt.expectedLength, len(builder.config.AllowedExtensions))
			}
		})
	}
}

// ============================================================================
// Include Tests
// ============================================================================

func TestDirectoryBuilder_Include(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  FilterConfig
		patterns       []string
		expectedLength int
	}{
		{
			name:           "add single pattern",
			initialConfig:  FilterConfig{},
			patterns:       []string{"*.yaml"},
			expectedLength: 1,
		},
		{
			name:           "add multiple patterns",
			initialConfig:  FilterConfig{},
			patterns:       []string{"*.yaml", "*.json", "*.yml"},
			expectedLength: 3,
		},
		{
			name: "append to existing include patterns",
			initialConfig: FilterConfig{
				IncludePatterns: []string{"*.txt"},
			},
			patterns:       []string{"*.yaml", "*.json"},
			expectedLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DirectoryBuilder{
				basePath: "/tmp/test",
				config:   tt.initialConfig,
			}

			result := builder.Include(tt.patterns...)

			if result != builder {
				t.Error("Include should return the same builder instance for chaining")
			}

			if len(builder.config.IncludePatterns) != tt.expectedLength {
				t.Errorf("expected IncludePatterns length %d, got %d", tt.expectedLength, len(builder.config.IncludePatterns))
			}
		})
	}
}

func TestDirectoryBuilder_Include_Panic(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
	}{
		{
			name:     "pattern starting with !",
			patterns: []string{"!*.yaml"},
		},
		{
			name:     "multiple patterns including one with !",
			patterns: []string{"*.yaml", "!*.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Include() should panic for patterns starting with '!'")
				}
			}()

			builder := &DirectoryBuilder{
				basePath: "/tmp/test",
				config:   FilterConfig{},
			}

			builder.Include(tt.patterns...)
		})
	}
}

// ============================================================================
// Exclude Tests
// ============================================================================

func TestDirectoryBuilder_Exclude(t *testing.T) {
	tests := []struct {
		name           string
		initialConfig  FilterConfig
		patterns       []string
		expectedLength int
	}{
		{
			name:           "add single pattern",
			initialConfig:  FilterConfig{},
			patterns:       []string{"*.tmp"},
			expectedLength: 1,
		},
		{
			name:           "add multiple patterns",
			initialConfig:  FilterConfig{},
			patterns:       []string{"*.tmp", "*.bak", ".gitignore"},
			expectedLength: 3,
		},
		{
			name: "append to existing exclude patterns",
			initialConfig: FilterConfig{
				ExcludePatterns: []string{"temp.yaml"},
			},
			patterns:       []string{"*.tmp", "*.bak"},
			expectedLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DirectoryBuilder{
				basePath: "/tmp/test",
				config:   tt.initialConfig,
			}

			result := builder.Exclude(tt.patterns...)

			if result != builder {
				t.Error("Exclude should return the same builder instance for chaining")
			}

			if len(builder.config.ExcludePatterns) != tt.expectedLength {
				t.Errorf("expected ExcludePatterns length %d, got %d", tt.expectedLength, len(builder.config.ExcludePatterns))
			}
		})
	}
}

func TestDirectoryBuilder_Exclude_Panic(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
	}{
		{
			name:     "pattern starting with !",
			patterns: []string{"!*.yaml"},
		},
		{
			name:     "multiple patterns including one with !",
			patterns: []string{"*.yaml", "!*.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Exclude() should panic for patterns starting with '!'")
				}
			}()

			builder := &DirectoryBuilder{
				basePath: "/tmp/test",
				config:   FilterConfig{},
			}

			builder.Exclude(tt.patterns...)
		})
	}
}

// ============================================================================
// Match Tests
// ============================================================================

func TestDirectoryBuilder_Match(t *testing.T) {
	tests := []struct {
		name            string
		initialConfig   FilterConfig
		patterns        []string
		include         bool
		expectedInclude int
		expectedExclude int
	}{
		{
			name:            "add include pattern",
			initialConfig:   FilterConfig{},
			patterns:        []string{"*.yaml"},
			include:         true,
			expectedInclude: 1,
			expectedExclude: 0,
		},
		{
			name:            "add exclude pattern",
			initialConfig:   FilterConfig{},
			patterns:        []string{"*.tmp"},
			include:         false,
			expectedInclude: 0,
			expectedExclude: 1,
		},
		{
			name: "append to existing include patterns",
			initialConfig: FilterConfig{
				IncludePatterns: []string{"*.txt"},
			},
			patterns:        []string{"*.yaml", "*.json"},
			include:         true,
			expectedInclude: 3,
			expectedExclude: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DirectoryBuilder{
				basePath: "/tmp/test",
				config:   tt.initialConfig,
			}

			result := builder.Match(tt.patterns, tt.include)

			if result != builder {
				t.Error("Match should return the same builder instance for chaining")
			}

			if len(builder.config.IncludePatterns) != tt.expectedInclude {
				t.Errorf("expected IncludePatterns length %d, got %d", tt.expectedInclude, len(builder.config.IncludePatterns))
			}
			if len(builder.config.ExcludePatterns) != tt.expectedExclude {
				t.Errorf("expected ExcludePatterns length %d, got %d", tt.expectedExclude, len(builder.config.ExcludePatterns))
			}
		})
	}
}

// ============================================================================
// MatchNested Tests
// ============================================================================

func TestDirectoryBuilder_MatchNested(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		setTo    bool
		expected bool
	}{
		{
			name:     "set to true",
			initial:  false,
			setTo:    true,
			expected: true,
		},
		{
			name:     "set to false",
			initial:  true,
			setTo:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DirectoryBuilder{
				basePath:    "/tmp/test",
				config:      FilterConfig{},
				matchNested: tt.initial,
			}

			result := builder.MatchNested(tt.setTo)

			if result != builder {
				t.Error("MatchNested should return the same builder instance for chaining")
			}

			if builder.matchNested != tt.expected {
				t.Errorf("expected matchNested to be %v, got %v", tt.expected, builder.matchNested)
			}
		})
	}
}

// ============================================================================
// CaseSensitive Tests
// ============================================================================

func TestDirectoryBuilder_CaseSensitive(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		setTo    bool
		expected bool
	}{
		{
			name:     "set to true",
			initial:  false,
			setTo:    true,
			expected: true,
		},
		{
			name:     "set to false",
			initial:  true,
			setTo:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DirectoryBuilder{
				basePath:      "/tmp/test",
				config:        FilterConfig{},
				caseSensitive: tt.initial,
			}

			result := builder.CaseSensitive(tt.setTo)

			if result != builder {
				t.Error("CaseSensitive should return the same builder instance for chaining")
			}

			if builder.caseSensitive != tt.expected {
				t.Errorf("expected caseSensitive to be %v, got %v", tt.expected, builder.caseSensitive)
			}
		})
	}
}

// ============================================================================
// Build Tests
// ============================================================================

func TestDirectoryBuilder_Build(t *testing.T) {
	tests := []struct {
		name                string
		basePath            string
		config              FilterConfig
		matchNested         bool
		caseSensitive       bool
		expectedUseFastPath bool
	}{
		{
			name:                "empty config creates fast path",
			basePath:            "/tmp/test",
			config:              FilterConfig{},
			matchNested:         true,
			caseSensitive:       false,
			expectedUseFastPath: true,
		},
		{
			name:     "non-empty config disables fast path",
			basePath: "/tmp/test",
			config: FilterConfig{
				IncludePatterns: []string{"*.yaml"},
			},
			matchNested:         true,
			caseSensitive:       false,
			expectedUseFastPath: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DirectoryBuilder{
				basePath:      tt.basePath,
				config:        tt.config,
				matchNested:   tt.matchNested,
				caseSensitive: tt.caseSensitive,
			}

			result := builder.Build()

			if result == nil {
				t.Fatal("Build() should not return nil")
			}

			if result.basePath != tt.basePath {
				t.Errorf("expected basePath to be %q, got %q", tt.basePath, result.basePath)
			}

			resultConfig := result.FilterConfig()
			if len(resultConfig.IncludePatterns) != len(tt.config.IncludePatterns) {
				t.Errorf("expected IncludePatterns length %d, got %d", len(tt.config.IncludePatterns), len(resultConfig.IncludePatterns))
			}
			if len(resultConfig.ExcludePatterns) != len(tt.config.ExcludePatterns) {
				t.Errorf("expected ExcludePatterns length %d, got %d", len(tt.config.ExcludePatterns), len(resultConfig.ExcludePatterns))
			}
			if len(resultConfig.AllowedExtensions) != len(tt.config.AllowedExtensions) {
				t.Errorf("expected AllowedExtensions length %d, got %d", len(tt.config.AllowedExtensions), len(resultConfig.AllowedExtensions))
			}

			if result.matchNested != tt.matchNested {
				t.Errorf("expected matchNested to be %v, got %v", tt.matchNested, result.matchNested)
			}

			if result.caseSensitive != tt.caseSensitive {
				t.Errorf("expected caseSensitive to be %v, got %v", tt.caseSensitive, result.caseSensitive)
			}

			if result.useFastPath != tt.expectedUseFastPath {
				t.Errorf("expected useFastPath to be %v, got %v", tt.expectedUseFastPath, result.useFastPath)
			}
		})
	}
}

// ============================================================================
// Builder Chaining Tests
// ============================================================================

func TestDirectoryBuilder_Chaining(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "full chain with all methods",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFilteredDirectory("/tmp/test").
				WithExtensions("yaml", "json").
				Include("*.yaml", "*.yml").
				Exclude("*.tmp").
				Match([]string{"*.bak"}, false).
				MatchNested(true).
				CaseSensitive(false)

			if builder == nil {
				t.Fatal("builder should not be nil")
			}

			dir := builder.Build()
			if dir == nil {
				t.Fatal("Build() should not return nil")
			}

			config := dir.FilterConfig()
			if len(config.IncludePatterns) != 2 {
				t.Errorf("expected 2 include patterns, got %d", len(config.IncludePatterns))
			}
			if len(config.ExcludePatterns) != 2 { // *.tmp + *.bak
				t.Errorf("expected 2 exclude patterns, got %d", len(config.ExcludePatterns))
			}
			if len(config.AllowedExtensions) != 2 {
				t.Errorf("expected 2 extensions, got %d", len(config.AllowedExtensions))
			}

			if !dir.matchNested {
				t.Error("matchNested should be true")
			}
			if dir.caseSensitive {
				t.Error("caseSensitive should be false")
			}
		})
	}
}
