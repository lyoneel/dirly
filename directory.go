// Package dirly provides cross-platform directory operations with file filtering.
//
// This package is designed to be independent and may become a standalone library in the future.
package dirly

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Directory represents a directory with file operations.
type Directory struct {
	basePath      string
	filterConfig  FilterConfig // set only during construction via builder
	useFastPath   bool         // computed at construction time - true if no filters (fast path)
	matchNested   bool         // default: true, matches against full relative path
	caseSensitive bool         // default: false (OS default), set to true for explicit case sensitivity
	defaultUID    int          // default UID for new files (-1 = no change)
	defaultGID    int          // default GID for new files (-1 = no change)
	defaultPerm   os.FileMode  // default permissions for new files (0 = no change)
}

// NewDirectory creates a new unfiltered Directory instance.
func NewDirectory(basePath string) *Directory {
	return &Directory{
		basePath:      basePath,
		filterConfig:  FilterConfig{}, // empty config
		useFastPath:   true,           // optimized path!
		matchNested:   true,           // default behavior
		caseSensitive: false,          // OS default (filepath.Match handles this)
		defaultUID:    -1,             // no change by default
		defaultGID:    -1,             // no change by default
		defaultPerm:   0,              // no change by default
	}
}

// isValidName checks if the name contains only allowed characters.
// For directory names: alphanumeric, _, -, +, ~, !, @, [, ], (, ), $, %, #, |, €
// Dot is allowed but only once, not at start or end (to prevent hidden files).
// For glob patterns, dots are allowed freely (e.g., *.yaml).
func isValidName(name string, isGlobPattern bool) error {
	dotCount := 0
	for _, r := range name {
		valid := false
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '_' || r == '-' || r == '+' || r == '~' || r == '!' || r == '@' ||
			r == '[' || r == ']' || r == '(' || r == ')' || r == '$' || r == '%' ||
			r == '#' || r == '|' || r == '€' {
			valid = true
		}
		if !valid && isGlobPattern && (r == '*' || r == '?' || r == '/') {
			valid = true
		}
		if !valid && (r != '.' || (!isGlobPattern && dotCount >= 1)) {
			return fmt.Errorf("invalid character %q in name: only alphanumeric, _, -, +, ~, !, @, [, ], (, ), $, %% , #, |, and € are allowed", r)
		}
		if r == '.' && !isGlobPattern {
			dotCount++
		}
	}

	if !isGlobPattern && dotCount > 0 {
		if dotCount > 1 {
			return fmt.Errorf("name can contain at most one dot")
		}
		if name[0] == '.' || name[len(name)-1] == '.' {
			return fmt.Errorf("name cannot start or end with a dot")
		}
	}

	return nil
}

// SubDirectory returns a new Directory instance for the given child directory name.
// Returns an error if the path does not exist, is not a directory, or uses invalid characters.
func (d *Directory) SubDirectory(name string) (*Directory, error) {
	if name == "" {
		return nil, fmt.Errorf("directory name cannot be empty")
	}

	if err := isValidName(name, false); err != nil {
		return nil, err
	}

	// Use resolvePath to validate path stays within base directory
	fullPath, err := d.resolvePath(name)
	if err != nil {
		return nil, err
	}

	// Check if it exists and is a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory %q: %w", name, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", name)
	}

	// Reuse NewDirectory with the validated path
	dir := NewDirectory(fullPath)

	// Preserve filter config from parent if it was set
	if !d.useFastPath {
		dir.filterConfig = d.filterConfig
		dir.useFastPath = false
	}

	return dir, nil
}

// BasePath returns the base directory path.
func (d *Directory) BasePath() string {
	return d.basePath
}

// FilterConfig returns a copy of the filter configuration.
func (d *Directory) FilterConfig() FilterConfig {
	return FilterConfig{
		IncludePatterns:   append([]string(nil), d.filterConfig.IncludePatterns...),
		ExcludePatterns:   append([]string(nil), d.filterConfig.ExcludePatterns...),
		AllowedExtensions: append([]string(nil), d.filterConfig.AllowedExtensions...),
	}
}

// Exists checks if a file exists in this directory.
// Returns true if the file exists and passes all filters, false otherwise.
func (d *Directory) Exists(filename string) bool {
	if filename == "" {
		return false
	}

	filePath, err := d.resolvePath(filename)
	if err != nil {
		return false
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// Check if it's a file (not a directory)
	if info.IsDir() {
		return false
	}

	// Apply filters if configured
	if !d.useFastPath {
		files := []string{filePath}
		files = d.applyFilters(files)
		return len(files) > 0
	}

	return true
}

// ReadToBytes reads a file in this directory and returns raw bytes.
func (d *Directory) ReadToBytes(filename string) ([]byte, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}
	filePath, err := d.resolvePath(filename)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filePath)
}

func (d *Directory) ReadToString(filename string) (string, error) {
	data, err := d.ReadToBytes(filename)
	var str string
	if err == nil {
		str = string(data)
	}
	return str, err
}

// ReadToLines reads a file and returns its content as a slice of strings (one per line).
func (d *Directory) ReadToLines(filename string) ([]string, error) {
	data, err := d.ReadToBytes(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	// Handle trailing newline - if file ends with \n, last element will be empty string
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines, nil
}

// WriteFromBytes writes content to a file in this directory with specified permissions.
func (d *Directory) WriteFromBytes(filename string, data []byte, perm os.FileMode) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return err
	}

	finalPerm := d.applyDefaultPermissions(perm)

	if err := d.writeWithOwnership(fullPath, data, finalPerm); err != nil {
		return err
	}

	return nil
}

// applyDefaultPermissions returns the effective permission for a file.
// If perm is 0 and defaultPerm is set, uses defaultPerm. Otherwise falls back to 0644.
func (d *Directory) applyDefaultPermissions(perm os.FileMode) os.FileMode {
	if perm == 0 && d.defaultPerm != 0 {
		return d.defaultPerm
	} else if perm == 0 {
		return 0644 // Default file permission when neither perm nor defaultPerm is set
	}
	return perm
}

// writeWithOwnership writes data to a file with specified permissions and applies ownership.
func (d *Directory) writeWithOwnership(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Apply default ownership if configured (only if both UID and GID are valid)
	// Note: Chown may fail on non-root systems; we log but don't fail the operation
	if d.defaultUID >= 0 && d.defaultGID >= 0 {
		_ = os.Chown(path, d.defaultUID, d.defaultGID) // Best effort
	}

	return nil
}

// Chown changes the owner and group of a file in this directory.
func (d *Directory) Chown(filename string, uid, gid int) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return err
	}

	if err := os.Chown(fullPath, uid, gid); err != nil {
		return fmt.Errorf("failed to change owner/group: %w", err)
	}
	return nil
}

// Chmod changes the permissions of a file in this directory.
func (d *Directory) Chmod(filename string, perm os.FileMode) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return err
	}

	if err := os.Chmod(fullPath, perm); err != nil {
		return fmt.Errorf("failed to change permissions: %w", err)
	}
	return nil
}

// WriteFromBuff writes content from a buffered reader to a file with specified permissions.
func (d *Directory) WriteFromBuff(filename string, r *bufio.Reader, perm os.FileMode) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return err
	}

	finalPerm := d.applyDefaultPermissions(perm)

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read from buffer: %w", err)
	}

	return d.writeWithOwnership(fullPath, data, finalPerm)
}

// WriteFromString writes content from a string to a file with specified permissions.
func (d *Directory) WriteFromString(filename string, s string, perm os.FileMode) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return err
	}

	finalPerm := d.applyDefaultPermissions(perm)

	return d.writeWithOwnership(fullPath, []byte(s), finalPerm)
}

// WriteFromScanner writes content from a scanner to a file with specified permissions.
func (d *Directory) WriteFromScanner(filename string, s *bufio.Scanner, perm os.FileMode) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return err
	}

	finalPerm := d.applyDefaultPermissions(perm)

	var data []byte
	for s.Scan() {
		data = append(data, s.Bytes()...)
		data = append(data, '\n')
	}

	if err := s.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return d.writeWithOwnership(fullPath, data, finalPerm)
}

// WriteFromLines writes content from a slice of strings (lines) to a file with specified permissions.
func (d *Directory) WriteFromLines(filename string, lines []string, perm os.FileMode) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return err
	}

	finalPerm := d.applyDefaultPermissions(perm)

	var data []byte
	for _, line := range lines {
		data = append(data, []byte(line)...)
		data = append(data, '\n')
	}

	return d.writeWithOwnership(fullPath, data, finalPerm)
}

// ReadToBuff returns a buffered reader for efficient streaming reads.
func (d *Directory) ReadToBuff(filename string) (*os.File, *bufio.Reader, error) {
	if filename == "" {
		return nil, nil, fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	reader := bufio.NewReader(file)
	return file, reader, nil
}

// ReadToScanner returns a scanner for reading a file line by line.
func (d *Directory) ReadToScanner(filename string) (*os.File, *bufio.Scanner, error) {
	if filename == "" {
		return nil, nil, fmt.Errorf("filename cannot be empty")
	}

	fullPath, err := d.resolvePath(filename)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	scanner := bufio.NewScanner(file)
	return file, scanner, nil
}

// Remove removes a file from this directory.
func (d *Directory) Remove(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	filePath, err := d.resolvePath(filename)
	if err != nil {
		return err
	}
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// BatchRemoveAllFromDir removes a file or directory and all its contents.
func (d *Directory) BatchRemoveAllFromDir(path string) error {
	fullPath, err := d.resolvePath(path)
	if err != nil {
		return err
	}
	return os.RemoveAll(fullPath)
}

// BatchRemove removes multiple files or directories.
// Returns an error if any path cannot be removed.
func (d *Directory) BatchRemove(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	for _, path := range paths {
		fullPath, err := d.resolvePath(path)
		if err != nil {
			return fmt.Errorf("invalid path %q: %w", path, err)
		}
		if err := os.RemoveAll(fullPath); err != nil {
			return fmt.Errorf("failed to remove %q: %w", path, err)
		}
	}

	return nil
}

// resolvePath safely joins and validates a relative path against the base directory.
// It prevents directory traversal attacks by ensuring the resolved path stays within basePath.
func (d *Directory) resolvePath(relative string) (string, error) {
	if relative == "" {
		return "", fmt.Errorf("relative path cannot be empty")
	}

	// Reject absolute paths
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("absolute paths not allowed: %q", relative)
	}

	// Clean and join paths
	joined := filepath.Join(d.basePath, relative)
	cleaned := filepath.Clean(joined)

	// Normalize basePath for comparison (ensure trailing separator)
	baseForCheck := d.basePath
	if !strings.HasSuffix(baseForCheck, string(filepath.Separator)) {
		baseForCheck += string(filepath.Separator)
	}

	// Verify the resolved path is within the base directory
	// Allow exact match for basePath itself
	if cleaned != d.basePath && !strings.HasPrefix(cleaned, baseForCheck) {
		return "", fmt.Errorf("path traversal detected: %q resolves to %q (outside base %q)",
			relative, cleaned, d.basePath)
	}

	return cleaned, nil
}

// getByGlob searches for files matching a pattern and returns absolute paths.
// It first gets all files (applying filters if configured), then applies the glob pattern
// to those filtered results, ensuring matchNested setting is respected.
func (d *Directory) getByGlob(pattern string) ([]string, error) {
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	if err := isValidName(pattern, true); err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	var allFiles []string
	var err error

	// If filters are configured, get filtered files first
	if !d.useFastPath {
		allFiles, err = d.GetAllAbs()
		if err != nil {
			return nil, fmt.Errorf("failed to get filtered files: %w", err)
		}
	} else {
		// No filters - walk directory to get all files (so matchNested can be applied)
		err = filepath.WalkDir(d.basePath, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			allFiles = append(allFiles, path)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	}

	// Apply glob pattern to all files (respecting matchNested setting)
	var matches []string
	for _, file := range allFiles {
		relPath := strings.TrimPrefix(file, d.basePath+string(filepath.Separator))

		// Handle **/ prefix for recursive matching
		if strings.HasPrefix(pattern, "**/") {
			suffix := strings.TrimPrefix(pattern, "**/")

			// In flat mode (matchNested=false), **/ should NOT recurse - treat as simple pattern
			if !d.matchNested {
				// Only match files in root directory with the suffix pattern
				if !strings.Contains(relPath, string(filepath.Separator)) && matchSimpleGlob(suffix, filepath.Base(relPath)) {
					matches = append(matches, file)
				}
			} else {
				// Nested mode: match suffix against any path component or filename
				if matchPatternInPath(relPath, suffix) || matchSimpleGlob(suffix, relPath) {
					matches = append(matches, file)
				}
			}
		} else if strings.Contains(pattern, "/") {
			// Path pattern like "config/*.yaml" - match against relative path
			if matchPatternInPath(relPath, pattern) {
				matches = append(matches, file)
			}
		} else {
			// Simple pattern like "*.yaml" - respect matchNested setting
			if d.matchNested {
				// Match against filename AND any path component
				if matchSimpleGlob(pattern, relPath) || matchSimpleGlob(pattern, filepath.Base(relPath)) {
					matches = append(matches, file)
				}
			} else {
				// Flat mode - only match files directly in root directory (no subdirectories)
				// Check if file is in a subdirectory by looking for path separator
				if !strings.Contains(relPath, string(filepath.Separator)) {
					// File is in root, check if filename matches pattern
					if matchSimpleGlob(pattern, filepath.Base(relPath)) {
						matches = append(matches, file)
					}
				}
			}
		}
	}

	return matches, nil
}

// matchSimpleGlob checks if a path matches a simple glob pattern (no / in pattern).
func matchSimpleGlob(pattern string, path string) bool {
	matched, _ := filepath.Match(pattern, path)
	return matched
}

// GetAllByGlobAbs searches for files matching a pattern and returns absolute paths.
func (d *Directory) GetAllByGlobAbs(pattern string) ([]string, error) {
	return d.getByGlob(pattern)
}

// GetAllByGlobRel searches for files matching a pattern and returns relative paths.
func (d *Directory) GetAllByGlobRel(pattern string) ([]string, error) {
	absPaths, err := d.GetAllByGlobAbs(pattern)
	if err != nil {
		return nil, err
	}

	var relPaths []string
	for _, absPath := range absPaths {
		relPath, err := filepath.Rel(d.basePath, absPath)
		if err == nil {
			relPaths = append(relPaths, relPath)
		}
	}

	return relPaths, nil
}

// GetAllAbs returns all files with absolute paths (applies builder filters).
func (d *Directory) GetAllAbs() ([]string, error) {
	var files []string

	err := filepath.WalkDir(d.basePath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if entry.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Apply filters if configured
	if !d.useFastPath {
		files = d.applyFilters(files)
	}

	return files, nil
}

// GetAllDirsAbs returns all directories with absolute paths.
func (d *Directory) GetAllDirsAbs() ([]string, error) {
	var dirs []string

	err := filepath.WalkDir(d.basePath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Include directories only (skip the base directory itself)
		if entry.IsDir() && path != d.basePath {
			dirs = append(dirs, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return dirs, nil
}

// GetAllDirsRel returns all directories with relative paths.
func (d *Directory) GetAllDirsRel() ([]string, error) {
	absPaths, err := d.GetAllDirsAbs()
	if err != nil {
		return nil, err
	}

	var relPaths []string
	for _, absPath := range absPaths {
		relPath, err := filepath.Rel(d.basePath, absPath)
		if err == nil {
			relPaths = append(relPaths, relPath)
		}
	}

	return relPaths, nil
}

// GetAllRel returns all files with relative paths (applies builder filters).
func (d *Directory) GetAllRel() ([]string, error) {
	absPaths, err := d.GetAllAbs()
	if err != nil {
		return nil, err
	}

	var relPaths []string
	for _, absPath := range absPaths {
		relPath, err := filepath.Rel(d.basePath, absPath)
		if err == nil {
			relPaths = append(relPaths, relPath)
		}
	}

	return relPaths, nil
}

// applyFilters applies include/exclude/extension filters.
func (d *Directory) applyFilters(files []string) []string {
	config := d.filterConfig

	if len(config.IncludePatterns) > 0 {
		if d.caseSensitive {
			files = matchFilesCaseSensitive(files, config.IncludePatterns, d.basePath, d.matchNested)
		} else {
			files = matchFiles(files, config.IncludePatterns, d.basePath, d.matchNested)
		}
	}

	if len(config.ExcludePatterns) > 0 {
		if d.caseSensitive {
			files = filterOutFilesCaseSensitive(files, config.ExcludePatterns, d.basePath, d.matchNested)
		} else {
			files = filterOutFiles(files, config.ExcludePatterns, d.basePath, d.matchNested)
		}
	}

	if len(config.AllowedExtensions) > 0 {
		files = filterByExtension(files, config.AllowedExtensions, d.basePath, d.matchNested)
	}

	return files
}

// isPathWithinBase checks if a path is within the base directory.
func (d *Directory) isPathWithinBase(path string) bool {
	cleaned := filepath.Clean(path)
	baseForCheck := d.basePath
	if !strings.HasSuffix(baseForCheck, string(filepath.Separator)) {
		baseForCheck += string(filepath.Separator)
	}

	return cleaned == d.basePath || strings.HasPrefix(cleaned, baseForCheck)
}

// AbsPath returns the resolved absolute path for a relative path.
func (d *Directory) AbsPath(name string) (string, error) {
	return d.resolvePath(name)
}

// BatchReadToBytes reads multiple files and returns their contents as a map.
// Returns an error if any file cannot be read.
func (d *Directory) BatchReadToBytes(filenames []string) (map[string][]byte, error) {
	if len(filenames) == 0 {
		return make(map[string][]byte), nil
	}

	result := make(map[string][]byte, len(filenames))
	for _, filename := range filenames {
		content, err := d.ReadToBytes(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read %q: %w", filename, err)
		}
		result[filename] = content
	}

	return result, nil
}

// BatchWriteFromBytes writes multiple files with their contents using default permissions from Directory.
// Returns an error if any file cannot be written.
func (d *Directory) BatchWriteFromBytes(files map[string][]byte) error {
	if len(files) == 0 {
		return nil
	}

	for filename, content := range files {
		if err := d.WriteFromBytes(filename, content, 0); err != nil {
			return fmt.Errorf("failed to write %q: %w", filename, err)
		}
	}

	return nil
}

// BatchReadToString reads multiple files and returns their contents as a map of strings.
// Returns an error if any file cannot be read.
func (d *Directory) BatchReadToString(filenames []string) (map[string]string, error) {
	if len(filenames) == 0 {
		return make(map[string]string), nil
	}

	result := make(map[string]string, len(filenames))
	for _, filename := range filenames {
		content, err := d.ReadToString(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read %q: %w", filename, err)
		}
		result[filename] = content
	}

	return result, nil
}

// BatchWriteFromString writes multiple files from strings using default permissions.
// Returns an error if any file cannot be written.
func (d *Directory) BatchWriteFromString(files map[string]string) error {
	if len(files) == 0 {
		return nil
	}

	for filename, content := range files {
		if err := d.WriteFromString(filename, content, 0); err != nil {
			return fmt.Errorf("failed to write %q: %w", filename, err)
		}
	}

	return nil
}

// BatchReadToBuff reads multiple files and returns their contents as a map of buffered readers.
// Returns an error if any file cannot be read.
func (d *Directory) BatchReadToBuff(filenames []string) (map[string]*bufio.Reader, error) {
	if len(filenames) == 0 {
		return make(map[string]*bufio.Reader), nil
	}

	result := make(map[string]*bufio.Reader, len(filenames))
	for _, filename := range filenames {
		_, reader, err := d.ReadToBuff(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read %q: %w", filename, err)
		}
		result[filename] = reader
	}

	return result, nil
}

// BatchWriteFromBuff writes multiple files from buffered readers using default permissions.
// Returns an error if any file cannot be written.
func (d *Directory) BatchWriteFromBuff(files map[string]*bufio.Reader) error {
	if len(files) == 0 {
		return nil
	}

	for filename, reader := range files {
		if err := d.WriteFromBuff(filename, reader, 0); err != nil {
			return fmt.Errorf("failed to write %q: %w", filename, err)
		}
	}

	return nil
}

// BatchReadToScanner reads multiple files and returns their contents as a map of scanners.
// Returns an error if any file cannot be read.
func (d *Directory) BatchReadToScanner(filenames []string) (map[string]*bufio.Scanner, error) {
	if len(filenames) == 0 {
		return make(map[string]*bufio.Scanner), nil
	}

	result := make(map[string]*bufio.Scanner, len(filenames))
	for _, filename := range filenames {
		_, scanner, err := d.ReadToScanner(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read %q: %w", filename, err)
		}
		result[filename] = scanner
	}

	return result, nil
}

// BatchWriteFromScanner writes multiple files from scanners using default permissions.
// Returns an error if any file cannot be written.
func (d *Directory) BatchWriteFromScanner(files map[string]*bufio.Scanner) error {
	if len(files) == 0 {
		return nil
	}

	for filename, scanner := range files {
		if err := d.WriteFromScanner(filename, scanner, 0); err != nil {
			return fmt.Errorf("failed to write %q: %w", filename, err)
		}
	}

	return nil
}

// BatchReadToLines reads multiple files and returns their contents as a map of line slices.
// Returns an error if any file cannot be read.
func (d *Directory) BatchReadToLines(filenames []string) (map[string][]string, error) {
	if len(filenames) == 0 {
		return make(map[string][]string), nil
	}

	result := make(map[string][]string, len(filenames))
	for _, filename := range filenames {
		lines, err := d.ReadToLines(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read %q: %w", filename, err)
		}
		result[filename] = lines
	}

	return result, nil
}

// BatchWriteFromLines writes multiple files from line slices using default permissions.
// Returns an error if any file cannot be written.
func (d *Directory) BatchWriteFromLines(files map[string][]string) error {
	if len(files) == 0 {
		return nil
	}

	for filename, lines := range files {
		if err := d.WriteFromLines(filename, lines, 0); err != nil {
			return fmt.Errorf("failed to write %q: %w", filename, err)
		}
	}

	return nil
}
