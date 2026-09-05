# dirly - Cross-Platform Directory Operations with File Filtering

A standalone Go library for directory operations with powerful file filtering capabilities.

> Main repository: https://gitlab.com/lyoneel/dirly
> Any other host that serves this repository is a mirror. Open issues
> and merge requests on GitLab.

## Features

- **Cross-platform** directory operations using standard `filepath` functions
- **Builder pattern** for immutable configuration
- **Flexible file matching** with support for wildcards, recursive patterns (`**/`), and path-based filters
- **Case-sensitive or case-insensitive** matching (OS default)
- **Path traversal protection** against security vulnerabilities
- **Batch operations** for efficient file handling
- **Multiple I/O modes**: bytes, strings, lines, bufio.Reader, *bufio.Scanner

## Installation

```bash
go get gitlab.com/lyoneel/dirly
```

```go
import "gitlab.com/lyoneel/dirly"
```

Requires Go 1.26 or later. The library has no external dependencies.

## Quick Start

```go
package main

import (
    "fmt"
    "gitlab.com/lyoneel/dirly"
)

func main() {
    // Create a filtered directory with YAML files only
    dir := dirly.NewFilteredDirectory("/path/to/config").
        WithExtensions("yaml", "json").
        Build()

    // Get all matching files (relative paths)
    files, err := dir.GetAllRel()
    if err != nil {
        panic(err)
    }

    fmt.Println(files)
}
```

## Core API

| Constructor | Purpose |
|-------------|---------|
| `NewDirectory(basePath string) *Directory` | Unfiltered directory. All files pass. Uses a fast path when no filters are configured. |
| `NewFilteredDirectory(basePath string) *DirectoryBuilder` | Builder for a filtered, immutable `Directory`. Chain configuration methods, then call `Build()`. |

Key operations on the resulting `*Directory`:

| Group | Methods |
|-------|---------|
| Inspection | `BasePath`, `FilterConfig`, `Exists`, `AbsPath` |
| Single-file read | `ReadToBytes`, `ReadToString`, `ReadToLines`, `ReadToBuff`, `ReadToScanner` |
| Single-file write | `WriteFromBytes`, `WriteFromString`, `WriteFromLines`, `WriteFromBuff`, `WriteFromScanner` |
| Batch I/O | `BatchReadToBytes`, `BatchReadToLines`, `BatchWriteFromBytes`, `BatchWriteFromLines`, and the matching Buff and Scanner variants |
| Glob search | `GetAllByGlobAbs`, `GetAllByGlobRel` |
| File management | `Chown`, `Chmod`, `Remove`, `BatchRemove`, `BatchRemoveAllFromDir` |

The full method reference with signatures, semantics, and all builder methods lives in [DEVELOPMENT.md](DEVELOPMENT.md). The generated reference lives on pkg.go.dev.

## File Matching

Filters and glob operations share the same matching engine:

1. Simple wildcards (`*.yaml`) use `filepath.Match` on filenames.
2. Path patterns (`config/*.yaml`) match files in a specific subdirectory.
3. Recursive patterns (`**/*.yaml`) match at any depth.
4. `MatchNested(true)` (default) matches filenames and full relative paths. `MatchNested(false)` restricts simple patterns to the root directory.

See [DEVELOPMENT.md](DEVELOPMENT.md) for the complete matching semantics, the filter evaluation order, and the mode tables.

## Security

Path traversal protection is built in:

1. Absolute paths and `..` sequences in file names are rejected.
2. Glob patterns containing `..` are rejected before any file system access.
3. Glob results are validated to stay within the base directory.

See [DEVELOPMENT.md](DEVELOPMENT.md) for the path resolution and validation internals.

## Examples

### Example 1: Load All YAML Config Files

```go
dir := dirly.NewFilteredDirectory("./config").
    WithExtensions("yaml", "yml").
    Build()

files, err := dir.GetAllRel()
if err != nil {
    panic(err)
}

// files = ["database.yaml", "server.yml", "cache.yaml"]
```

### Example 2: Get Specific Subdirectory Files Only

```go
dir := dirly.NewFilteredDirectory("./src").
    MatchNested(false).  // Flat mode for precise control
    Build()

// Get only Go files directly in src/ (not in subdirectories)
goFiles, _ := dir.GetAllByGlobRel("*.go")

// Get only config files in src/config/ subdirectory
configFiles, _ := dir.GetAllByGlobRel("src/config/*.yaml")
```

### Example 3: Recursive Search with Extension Filter

```go
dir := dirly.NewFilteredDirectory("./").
    WithExtensions("md").
    MatchNested(true).
    Build()

// Find all markdown files anywhere in the project
docs, _ := dir.GetAllByGlobRel("**/*.md")
// docs = ["README.md", "docs/guide.md", "api/CHANGELOG.md"]
```

More examples cover batch operations, streaming with scanners and buffered readers, default permissions and ownership, case sensitivity, and complex filter chains. See [DEVELOPMENT.md](DEVELOPMENT.md).

## Testing

Run the comprehensive test suite:

```bash
go test ./... -v
```

Tests cover:
- Builder pattern methods (WithExtensions, Include, Exclude, MatchNested, CaseSensitive, WithDefaultOwnership, WithDefaultPermissions)
- File matching with various patterns (`*.yaml`, `config/*.yaml`, `**/*.yaml`)
- Both nested and flat matching modes for filters AND glob operations
- Path traversal protection (absolute paths, `..` sequences, encoded traversals)
- All directory operations (ReadToBytes, ReadToString, ReadToLines, WriteFromBytes, WriteFromString, WriteFromLines, etc.)
- Batch operations for all I/O modes (bytes, strings, lines, bufio.Reader, *bufio.Scanner)
- Glob pattern behavior with `matchNested` setting (new tests in `glob_nested_test.go`)
- File management operations (Chown, Chmod, Remove, BatchRemove, BatchRemoveAllFromDir)
- Edge cases (empty filenames, empty patterns, non-existent files, permission errors)

## Comparison Table

| Feature | `NewDirectory()` | `NewFilteredDirectory().Build()` |
|---------|------------------|----------------------------------|
| Extension filtering | ❌ No | ✅ Yes |
| Include/Exclude patterns | ❌ No | ✅ Yes |
| Path traversal protection | ✅ Yes | ✅ Yes |
| Builder pattern | ❌ No | ✅ Yes |
| Default ownership/permissions | ❌ No | ✅ Yes |
| Fast path optimization | ✅ Yes (no filters) | ⚠️ Only if no filters configured |

## License

MIT License - See [LICENSE](LICENSE) file for details.
