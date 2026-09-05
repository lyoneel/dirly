# Changelog

All notable changes in the dirly project are documented in this file. The format is based on Keep a Changelog, and the project follows Semantic Versioning.

## v1.1.0 - 2026-08-17

### Changed

- Breaking: flatten the package to a single root package `dirly` (the former `lydir` subpackage is gone)
- Remove the leftover `lydir` directory after flattening

### Upgrade Notes

Code that imported the subpackage changes its import path:

```go
// before
import "gitlab.com/lyoneel/dirly/lydir"
// after
import "gitlab.com/lyoneel/dirly"
```

## v1.0.0 - 2026-08-13

### Added

- First stable release of the standalone library
- Cross-platform directory operations using standard `filepath` functions
- Builder pattern with immutable filter configuration: `WithExtensions`, `Include`, `Exclude`, `Match`, `MatchNested`, `CaseSensitive`, `WithDefaultOwnership`, `WithDefaultPermissions`
- File matching with simple wildcards, path patterns, and recursive `**/` patterns
- Case-sensitive or case-insensitive matching with OS default behavior
- Path traversal protection for file operations and glob patterns
- Single-file and batch operations across five I/O modes: bytes, strings, lines, `*bufio.Reader`, and `*bufio.Scanner`
- File management operations: `Chown`, `Chmod`, `Remove`, `BatchRemove`, `BatchRemoveAllFromDir`
- Glob search with absolute and relative result paths

### Dependencies

- None. The library uses the Go standard library only.

## v0.0.0 - 2026-07-02

### Added

- Initial public snapshot of the library with README and MIT license

## Statistics

- 6 commits across all tags
- 38 tracked files, 34 Go source files, about 11000 lines of Go
- Test coverage 85.0 percent of statements
