package dirly

import (
	"os"
	"strings"
)

// DirectoryBuilder constructs a filtered Directory with immutable configuration.
type DirectoryBuilder struct {
	basePath      string
	config        FilterConfig
	matchNested   bool
	caseSensitive bool
	defaultUID    int // -1 = no change
	defaultGID    int // -1 = no change
	defaultPerm   os.FileMode
}

// NewFilteredDirectory returns a builder for filtered directories.
func NewFilteredDirectory(basePath string) *DirectoryBuilder {
	return &DirectoryBuilder{
		basePath:      basePath,
		config:        FilterConfig{},
		matchNested:   true,  // default: match nested directories
		caseSensitive: false, // default: OS behavior via filepath.Match
	}
}

// WithExtensions adds the specified extensions to the allowedExtensions filter.
func (b *DirectoryBuilder) WithExtensions(extensions ...string) *DirectoryBuilder {
	b.config.AllowedExtensions = append(b.config.AllowedExtensions, extensions...)
	return b
}

// Include adds patterns to IncludePatterns. Panics if any pattern starts with "!".
func (b *DirectoryBuilder) Include(patterns ...string) *DirectoryBuilder {
	for _, p := range patterns {
		if strings.HasPrefix(p, "!") {
			panic("Include() does not support ! prefix. Use clear patterns instead.")
		}
	}
	b.config.IncludePatterns = append(b.config.IncludePatterns, patterns...)
	return b
}

// Exclude adds patterns to ExcludePatterns. Panics if any pattern starts with "!".
func (b *DirectoryBuilder) Exclude(patterns ...string) *DirectoryBuilder {
	for _, p := range patterns {
		if strings.HasPrefix(p, "!") {
			panic("Exclude() does not support ! prefix. Use clear patterns instead.")
		}
	}
	b.config.ExcludePatterns = append(b.config.ExcludePatterns, patterns...)
	return b
}

// Match adds patterns without validation (power user API).
// If include is true, patterns are added to IncludePatterns.
// If include is false, patterns are added to ExcludePatterns.
func (b *DirectoryBuilder) Match(patterns []string, include bool) *DirectoryBuilder {
	if include {
		b.config.IncludePatterns = append(b.config.IncludePatterns, patterns...)
	} else {
		b.config.ExcludePatterns = append(b.config.ExcludePatterns, patterns...)
	}
	return b
}

// MatchNested sets whether to match against full relative paths (default: true).
// If false, only filenames are matched (flat matching).
func (b *DirectoryBuilder) MatchNested(match bool) *DirectoryBuilder {
	b.matchNested = match
	return b
}

// CaseSensitive sets explicit case sensitivity.
// Default is OS behavior via filepath.Match. Set to true for always case-sensitive.
func (b *DirectoryBuilder) CaseSensitive(sensitive bool) *DirectoryBuilder {
	b.caseSensitive = sensitive
	return b
}

// WithDefaultOwnership sets default UID/GID for all new files written to this directory.
// Use -1 to skip setting a specific value.
func (b *DirectoryBuilder) WithDefaultOwnership(uid, gid int) *DirectoryBuilder {
	b.defaultUID = uid
	b.defaultGID = gid
	return b
}

// WithDefaultPermissions sets default file permissions for all new files written to this directory.
// Use 0 to skip setting permissions (use provided value).
func (b *DirectoryBuilder) WithDefaultPermissions(perm os.FileMode) *DirectoryBuilder {
	b.defaultPerm = perm
	return b
}

// Build creates the Directory with the configured filters.
// Filters are immutable after this point.
func (b *DirectoryBuilder) Build() *Directory {
	return &Directory{
		basePath:      b.basePath,
		filterConfig:  b.config,
		useFastPath:   b.config.isEmpty(),
		matchNested:   b.matchNested,
		caseSensitive: b.caseSensitive,
		defaultUID:    b.defaultUID,
		defaultGID:    b.defaultGID,
		defaultPerm:   b.defaultPerm,
	}
}
