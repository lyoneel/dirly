package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGlobWithMatchNested tests that glob patterns respect matchNested setting
// when filters are applied. This ensures glob operates on filtered files, not raw filesystem.
func TestGlobWithMatchNested(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure:
	// tmpDir/
	//   root.yaml          <- should match *.yaml in both modes
	//   config/
	//     settings.yaml    <- should match *.yaml with matchNested=true, NOT with false
	//     backup/
	//       data.yaml      <- should match **/*.yaml in both modes

	files := []string{
		"root.yaml",
		filepath.Join("config", "settings.yaml"),
		filepath.Join("config", "backup", "data.yaml"),
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	tests := []struct {
		name          string
		matchNested   bool
		pattern       string
		expectedCount int
		expectedFiles []string // base names only
	}{
		{
			name:          "matchNested=true with *.yaml should match all yaml files",
			matchNested:   true,
			pattern:       "*.yaml",
			expectedCount: 3,
			expectedFiles: []string{"root.yaml", "settings.yaml", "data.yaml"},
		},
		{
			name:          "matchNested=false with *.yaml should match only root yaml files (by filename)",
			matchNested:   false,
			pattern:       "*.yaml",
			expectedCount: 1, // Only root.yaml, not nested ones
			expectedFiles: []string{"root.yaml"},
		},
		{
			name:          "matchNested=true with **/*.yaml should match all yaml files recursively",
			matchNested:   true,
			pattern:       "**/*.yaml",
			expectedCount: 3,
			expectedFiles: []string{"root.yaml", "settings.yaml", "data.yaml"},
		},
		{
			name:          "matchNested=false with **/*.yaml should match only root yaml files (no recursive)",
			matchNested:   false,
			pattern:       "**/*.yaml",
			expectedCount: 1, // Only root.yaml
			expectedFiles: []string{"root.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewFilteredDirectory(tmpDir).
				MatchNested(tt.matchNested).
				Build()

			result, err := dir.GetAllByGlobRel(tt.pattern)
			if err != nil {
				t.Fatalf("GetAllByGlobRel() unexpected error: %v", err)
			}

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d files, got %d: %v", tt.expectedCount, len(result), result)
			}

			// Check that expected files are present
			for _, expectedFile := range tt.expectedFiles {
				found := false
				for _, path := range result {
					if filepath.Base(path) == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find %q in results, got: %v", expectedFile, result)
				}
			}

			// Check that unexpected files are NOT present
			unexpectedFiles := []string{"settings.yaml", "data.yaml"}
			for _, unexpectedFile := range unexpectedFiles {
				if !contains(tt.expectedFiles, unexpectedFile) {
					found := false
					for _, path := range result {
						if filepath.Base(path) == unexpectedFile {
							found = true
							break
						}
					}
					if found {
						t.Errorf("should NOT contain %q, but it was in results: %v", unexpectedFile, result)
					}
				}
			}
		})
	}
}

// TestGlobWithIncludeFilterAndMatchNested tests that Include filters work correctly with matchNested
func TestGlobWithIncludeFilterAndMatchNested(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure:
	// tmpDir/
	//   config.yaml          <- should match *.yaml in both modes
	//   settings.yaml        <- should match *.yaml in both modes
	//   config/
	//     database.yaml      <- should match with matchNested=true, NOT with false

	files := []string{
		"config.yaml",
		"settings.yaml",
		filepath.Join("config", "database.yaml"),
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	tests := []struct {
		name          string
		matchNested   bool
		pattern       string
		includeFilter string
		expectedCount int
	}{
		{
			name:          "matchNested=true with Include filter should match nested files",
			matchNested:   true,
			pattern:       "*.yaml",
			includeFilter: "*.yaml",
			expectedCount: 3, // All yaml files
		},
		{
			name:          "matchNested=false with Include filter should NOT match nested files",
			matchNested:   false,
			pattern:       "*.yaml",
			includeFilter: "*.yaml",
			expectedCount: 2, // Only root-level yaml files (config.yaml, settings.yaml)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewFilteredDirectory(tmpDir).
				MatchNested(tt.matchNested).
				Include(tt.includeFilter).
				Build()

			result, err := dir.GetAllByGlobRel(tt.pattern)
			if err != nil {
				t.Fatalf("GetAllByGlobRel() unexpected error: %v", err)
			}

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d files, got %d: %v", tt.expectedCount, len(result), result)
			}
		})
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
