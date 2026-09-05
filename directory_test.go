package dirly

import (
	"testing"
)

// ============================================================================
// FilterConfig Getter Tests
// ============================================================================

// TestFilterConfig_GetIncludePatterns tests that GetIncludePatterns returns a copy of the slice.
func TestFilterConfig_GetIncludePatterns(t *testing.T) {
	tests := []struct {
		name     string
		config   FilterConfig
		expected int
	}{
		{
			name: "empty config",
			config: FilterConfig{
				IncludePatterns: nil,
			},
			expected: 0,
		},
		{
			name: "single pattern",
			config: FilterConfig{
				IncludePatterns: []string{"*.yaml"},
			},
			expected: 1,
		},
		{
			name: "multiple patterns",
			config: FilterConfig{
				IncludePatterns: []string{"*.yaml", "*.json", "*.yml"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetIncludePatterns()

			if len(result) != tt.expected {
				t.Errorf("expected length %d, got %d", tt.expected, len(result))
			}

			// Verify it's a copy by modifying the result
			if len(result) > 0 {
				result[0] = "modified"
				if tt.config.IncludePatterns[0] == "modified" {
					t.Error("GetIncludePatterns should return a copy, not the original slice")
				}
			}
		})
	}
}

// TestFilterConfig_GetExcludePatterns tests that GetExcludePatterns returns a copy of the slice.
func TestFilterConfig_GetExcludePatterns(t *testing.T) {
	tests := []struct {
		name     string
		config   FilterConfig
		expected int
	}{
		{
			name: "empty config",
			config: FilterConfig{
				ExcludePatterns: nil,
			},
			expected: 0,
		},
		{
			name: "single pattern",
			config: FilterConfig{
				ExcludePatterns: []string{"temp.yaml"},
			},
			expected: 1,
		},
		{
			name: "multiple patterns",
			config: FilterConfig{
				ExcludePatterns: []string{"*.tmp", "*.bak", ".gitignore"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetExcludePatterns()

			if len(result) != tt.expected {
				t.Errorf("expected length %d, got %d", tt.expected, len(result))
			}

			// Verify it's a copy by modifying the result
			if len(result) > 0 {
				result[0] = "modified"
				if tt.config.ExcludePatterns[0] == "modified" {
					t.Error("GetExcludePatterns should return a copy, not the original slice")
				}
			}
		})
	}
}

// TestFilterConfig_GetAllowedExtensions tests that GetAllowedExtensions returns a copy of the slice.
func TestFilterConfig_GetAllowedExtensions(t *testing.T) {
	tests := []struct {
		name     string
		config   FilterConfig
		expected int
	}{
		{
			name: "empty config",
			config: FilterConfig{
				AllowedExtensions: nil,
			},
			expected: 0,
		},
		{
			name: "single extension",
			config: FilterConfig{
				AllowedExtensions: []string{"yaml"},
			},
			expected: 1,
		},
		{
			name: "multiple extensions",
			config: FilterConfig{
				AllowedExtensions: []string{"yaml", "json", "yml"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetAllowedExtensions()

			if len(result) != tt.expected {
				t.Errorf("expected length %d, got %d", tt.expected, len(result))
			}

			// Verify it's a copy by modifying the result
			if len(result) > 0 {
				result[0] = "modified"
				if tt.config.AllowedExtensions[0] == "modified" {
					t.Error("GetAllowedExtensions should return a copy, not the original slice")
				}
			}
		})
	}
}

// TestFilterConfig_isEmpty tests the isEmpty method.
func TestFilterConfig_isEmpty(t *testing.T) {
	tests := []struct {
		name     string
		config   FilterConfig
		expected bool
	}{
		{
			name: "empty config",
			config: FilterConfig{
				IncludePatterns:   nil,
				ExcludePatterns:   nil,
				AllowedExtensions: nil,
			},
			expected: true,
		},
		{
			name: "with include patterns only",
			config: FilterConfig{
				IncludePatterns: []string{"*.yaml"},
			},
			expected: false,
		},
		{
			name: "with exclude patterns only",
			config: FilterConfig{
				ExcludePatterns: []string{"*.tmp"},
			},
			expected: false,
		},
		{
			name: "with allowed extensions only",
			config: FilterConfig{
				AllowedExtensions: []string{"yaml"},
			},
			expected: false,
		},
		{
			name: "all filters set",
			config: FilterConfig{
				IncludePatterns:   []string{"*.yaml"},
				ExcludePatterns:   []string{"*.tmp"},
				AllowedExtensions: []string{"yaml"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.isEmpty()
			if result != tt.expected {
				t.Errorf("expected isEmpty() to be %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Directory Getter Tests
// ============================================================================

// TestDirectory_BasePath tests that BasePath returns the correct base directory.
func TestDirectory_BasePath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
	}{
		{
			name:     "simple path",
			basePath: "/tmp/test",
		},
		{
			name:     "path with trailing slash",
			basePath: "/tmp/test/",
		},
		{
			name:     "relative path",
			basePath: "./data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tt.basePath)
			result := dir.BasePath()

			if result != tt.basePath {
				t.Errorf("expected BasePath() to return %q, got %q", tt.basePath, result)
			}
		})
	}
}

// TestDirectory_FilterConfig tests that FilterConfig returns a copy of the filter configuration.
func TestDirectory_FilterConfig(t *testing.T) {
	tests := []struct {
		name       string
		basePath   string
		include    []string
		exclude    []string
		extensions []string
	}{
		{
			name:       "empty config",
			basePath:   "/tmp/test",
			include:    nil,
			exclude:    nil,
			extensions: nil,
		},
		{
			name:       "with filters",
			basePath:   "/tmp/test",
			include:    []string{"*.yaml"},
			exclude:    []string{"*.tmp"},
			extensions: []string{"yaml", "json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewFilteredDirectory(tt.basePath).
				Include(tt.include...).
				Exclude(tt.exclude...).
				WithExtensions(tt.extensions...).
				Build()

			result := dir.FilterConfig()

			if len(result.IncludePatterns) != len(tt.include) {
				t.Errorf("expected IncludePatterns length %d, got %d", len(tt.include), len(result.IncludePatterns))
			}
			if len(result.ExcludePatterns) != len(tt.exclude) {
				t.Errorf("expected ExcludePatterns length %d, got %d", len(tt.exclude), len(result.ExcludePatterns))
			}
			if len(result.AllowedExtensions) != len(tt.extensions) {
				t.Errorf("expected AllowedExtensions length %d, got %d", len(tt.extensions), len(result.AllowedExtensions))
			}

			// Verify it's a copy by modifying the result
			if len(result.IncludePatterns) > 0 {
				result.IncludePatterns[0] = "modified"
				if dir.FilterConfig().IncludePatterns[0] == "modified" {
					t.Error("FilterConfig should return a copy, not the original slice")
				}
			}
		})
	}
}
