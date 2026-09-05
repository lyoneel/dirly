package dirly

import (
	"path/filepath"
	"strings"
)

// FilterConfig holds file filtering rules for Directory.
// It is immutable after construction via the builder pattern.
type FilterConfig struct {
	IncludePatterns   []string // e.g., ["*.yaml", "*.json"] - files to include
	ExcludePatterns   []string // e.g., ["something.yaml", "*.tmp.yaml"] - files to exclude
	AllowedExtensions []string // e.g., ["yaml", "json"]
}

// GetIncludePatterns returns a copy of the include patterns slice.
func (fc FilterConfig) GetIncludePatterns() []string {
	result := make([]string, len(fc.IncludePatterns))
	copy(result, fc.IncludePatterns)
	return result
}

// GetExcludePatterns returns a copy of the exclude patterns slice.
func (fc FilterConfig) GetExcludePatterns() []string {
	result := make([]string, len(fc.ExcludePatterns))
	copy(result, fc.ExcludePatterns)
	return result
}

// GetAllowedExtensions returns a copy of the allowed extensions slice.
func (fc FilterConfig) GetAllowedExtensions() []string {
	result := make([]string, len(fc.AllowedExtensions))
	copy(result, fc.AllowedExtensions)
	return result
}

// isEmpty returns true if no filtering is configured.
func (fc FilterConfig) isEmpty() bool {
	return len(fc.IncludePatterns) == 0 &&
		len(fc.ExcludePatterns) == 0 &&
		len(fc.AllowedExtensions) == 0
}

// validatePattern checks for invalid ! prefix in patterns.
// Returns an error if the pattern starts with "!".
func validatePattern(pattern string) error {
	if strings.HasPrefix(pattern, "!") {
		return &InvalidPatternError{Pattern: pattern, Reason: "patterns cannot start with '!'"}
	}
	return nil
}

// InvalidPatternError is returned when a pattern contains invalid syntax.
type InvalidPatternError struct {
	Pattern string
	Reason  string
}

func (e *InvalidPatternError) Error() string {
	return "invalid pattern: " + e.Pattern + " - " + e.Reason
}

// matchFile checks if a file matches any of the given patterns.
// If matchNested is false, only the filename is checked (unless pattern contains /).
// If matchNested is true, both filename and full relative path are checked.
func matchFile(filePath string, patterns []string, basePath string, matchNested bool) bool {
	fileName := filepath.Base(filePath)

	if !matchNested {
		for _, pattern := range patterns {
			matched, _ := filepath.Match(pattern, fileName)
			if matched {
				return true
			}
			// Support patterns with / even when matchNested=false (e.g., "config/*.yaml")
			if strings.Contains(pattern, string(filepath.Separator)) {
				relPath := strings.TrimPrefix(filePath, basePath+"/")
				matched = matchPatternInPath(relPath, pattern)
				if matched {
					return true
				}
			}
		}
		return false
	}

	// Match against full relative path
	relPath := strings.TrimPrefix(filePath, basePath+"/")
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, fileName)
		if matched {
			return true
		}

		// Also match against full relative path for nested matching
		matched, _ = filepath.Match(pattern, relPath)
		if matched {
			return true
		}

		// Support **/ prefix pattern (recursive matching)
		if strings.HasPrefix(pattern, "**/") {
			suffix := strings.TrimPrefix(pattern, "**/")
			// Match suffix against any path component
			matched = matchPatternInPath(relPath, suffix)
			if matched {
				return true
			}

			// Also match the pattern directly (for files in root directory)
			// e.g., "**/*.yaml" should match "root.yaml"
			matched, _ = filepath.Match(pattern, relPath)
			if matched {
				return true
			}
		}

		// Support pattern without ** but with / (e.g., "subdir/*.yaml")
		if strings.Contains(pattern, "/") && !strings.HasPrefix(pattern, "**/") {
			matched = matchPatternInPath(relPath, pattern)
			if matched {
				return true
			}
		}

		// For patterns like "*.yaml", also check if any path component matches
		// This allows subdir/nested.yaml to match *.yaml when matchNested=true
		if !strings.Contains(pattern, "/") && strings.Contains(relPath, string(filepath.Separator)) {
			parts := strings.Split(relPath, string(filepath.Separator))
			for _, part := range parts {
				matched, _ = filepath.Match(pattern, part)
				if matched {
					return true
				}
			}
		}

		// Also check if the pattern matches any suffix of the relative path
		// e.g., "subdir/nested.yaml" should match "*.yaml" by checking the filename at the end
		if !strings.Contains(pattern, "/") {
			lastSlash := strings.LastIndex(relPath, string(filepath.Separator))
			if lastSlash >= 0 {
				filenameOnly := relPath[lastSlash+1:]
				matched, _ = filepath.Match(pattern, filenameOnly)
				if matched {
					return true
				}
			}
		}
	}

	return false
}

// matchFileCaseSensitive checks if a file matches any of the given patterns with explicit case sensitivity.
func matchFileCaseSensitive(filePath string, patterns []string, basePath string, matchNested bool) bool {
	fileName := filepath.Base(filePath)

	if !matchNested {
		for _, pattern := range patterns {
			matched, _ := filepath.Match(pattern, fileName)
			if matched {
				return true
			}
			// Support patterns with / even when matchNested=false (e.g., "config/*.yaml")
			if strings.Contains(pattern, string(filepath.Separator)) {
				relPath := strings.TrimPrefix(filePath, basePath+"/")
				matched = matchPatternInPath(relPath, pattern)
				if matched {
					return true
				}
			}
		}
		return false
	}

	relPath := strings.TrimPrefix(filePath, basePath+"/")
	for _, pattern := range patterns {
		// Force case-sensitive matching by comparing exact strings
		matched, _ := filepath.Match(pattern, fileName)
		if matched {
			return true
		}

		matched, _ = filepath.Match(pattern, relPath)
		if matched {
			return true
		}

		// For patterns without wildcards, do exact case-sensitive comparison
		if !strings.ContainsAny(pattern, "*?[]") {
			if fileName == pattern || relPath == pattern {
				return true
			}
		}

		// Support **/ prefix pattern (recursive matching)
		if strings.HasPrefix(pattern, "**/") {
			suffix := strings.TrimPrefix(pattern, "**/")
			matched = matchPatternInPath(relPath, suffix)
			if matched {
				return true
			}

			matched, _ = filepath.Match(pattern, relPath)
			if matched {
				return true
			}
		}

		// Support pattern without ** but with / (e.g., "subdir/*.yaml")
		if strings.Contains(pattern, "/") && !strings.HasPrefix(pattern, "**/") {
			matched = matchPatternInPath(relPath, pattern)
			if matched {
				return true
			}
		}

		// For patterns like "*.yaml", also check if any path component matches
		if !strings.Contains(pattern, "/") && strings.Contains(relPath, string(filepath.Separator)) {
			parts := strings.Split(relPath, string(filepath.Separator))
			for _, part := range parts {
				matched, _ = filepath.Match(pattern, part)
				if matched {
					return true
				}
			}

			// Also check exact case-sensitive match for each component
			if !strings.ContainsAny(pattern, "*?[]") {
				for _, part := range parts {
					if part == pattern {
						return true
					}
				}
			}
		}

		// Check filename at end of path with case sensitivity
		if !strings.Contains(pattern, "/") {
			lastSlash := strings.LastIndex(relPath, string(filepath.Separator))
			if lastSlash >= 0 {
				filenameOnly := relPath[lastSlash+1:]
				matched, _ = filepath.Match(pattern, filenameOnly)
				if matched {
					return true
				}

				// Exact case-sensitive match for filename
				if !strings.ContainsAny(pattern, "*?[]") && fileName == pattern {
					return true
				}
			}
		}
	}

	return false
}

// matchPatternInPath checks if a pattern matches any part of the relative path.
func matchPatternInPath(relPath string, pattern string) bool {
	parts := strings.Split(relPath, string(filepath.Separator))

	for i := range parts {
		// Try matching from this position onwards
		subPath := strings.Join(parts[i:], string(filepath.Separator))
		matched, _ := filepath.Match(pattern, subPath)
		if matched {
			return true
		}

		// Also try matching just the filename part
		matched, _ = filepath.Match(pattern, parts[i])
		if matched {
			return true
		}
	}

	return false
}

// matchFiles returns files that match any of the given patterns.
func matchFiles(files []string, patterns []string, basePath string, matchNested bool) []string {
	var matches []string
	for _, file := range files {
		if matchFile(file, patterns, basePath, matchNested) {
			matches = append(matches, file)
		}
	}
	return matches
}

// matchFilesCaseSensitive returns files that match any of the given patterns with explicit case sensitivity.
func matchFilesCaseSensitive(files []string, patterns []string, basePath string, matchNested bool) []string {
	var matches []string
	for _, file := range files {
		if matchFileCaseSensitive(file, patterns, basePath, matchNested) {
			matches = append(matches, file)
		}
	}
	return matches
}

// filterOutFiles removes files that match any of the given patterns.
func filterOutFiles(files []string, patterns []string, basePath string, matchNested bool) []string {
	var filtered []string
	for _, file := range files {
		if !matchFile(file, patterns, basePath, matchNested) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// filterOutFilesCaseSensitive removes files that match any of the given patterns with explicit case sensitivity.
func filterOutFilesCaseSensitive(files []string, patterns []string, basePath string, matchNested bool) []string {
	var filtered []string
	for _, file := range files {
		if !matchFileCaseSensitive(file, patterns, basePath, matchNested) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// addFiles adds files that match the pattern to the existing list (for overrides).
func addFiles(files []string, pattern string, basePath string, matchNested bool) []string {
	for _, file := range files {
		if !matchFile(file, []string{pattern}, basePath, matchNested) {
			continue
		}
		// Check if already in list to avoid duplicates
		exists := false
		for _, f := range files {
			if f == file {
				exists = true
				break
			}
		}
		if !exists {
			files = append(files, file)
		}
	}
	return files
}

// filterByExtension returns only files with the specified extensions.
func filterByExtension(files []string, extensions []string, basePath string, matchNested bool) []string {
	var filtered []string
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		if ext != "" {
			ext = ext[1:] // Remove leading dot
			for _, allowedExt := range extensions {
				if strings.EqualFold(ext, allowedExt) {
					filtered = append(filtered, file)
					break
				}
			}
		} else if len(extensions) == 0 {
			// No extension and no filter specified - include it
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// shouldMatch checks if a file path matches any of the given patterns.
// It handles **/ prefix for recursive matching.
func shouldMatch(filePath string, patterns []string, matchNested bool, caseSensitive bool) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "**/") {
			// Recursive pattern: match suffix against any path component
			suffix := strings.TrimPrefix(pattern, "**/")
			if matchPatternInPath(filePath, suffix) {
				return true
			}
			// Also try matching the full relative path
			matched, _ := filepath.Match(pattern, filePath)
			if matched {
				return true
			}
		} else if strings.Contains(pattern, "/") && !strings.HasPrefix(pattern, "**/") {
			// AbsPath-specific pattern (e.g., "config/*.yaml")
			if matchPatternInPath(filePath, pattern) {
				return true
			}
		} else {
			// Simple pattern (e.g., "*.yaml") - match against filename or any path component
			fileName := filepath.Base(filePath)
			matched, _ := filepath.Match(pattern, fileName)
			if matched {
				return true
			}
			// Also check if pattern matches any part of the relative path
			if strings.Contains(filePath, string(filepath.Separator)) {
				parts := strings.Split(filePath, string(filepath.Separator))
				for _, part := range parts {
					matched, _ = filepath.Match(pattern, part)
					if matched {
						return true
					}
				}
			}
		}
	}
	return false
}
