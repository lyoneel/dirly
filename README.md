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
- **Path traversal protection** against security vulnerabilities, including symlink escapes
- **Atomic writes** (temp file, fsync, rename), so a crash never leaves a half-written file
- **Batch operations** for efficient file handling
- **Multiple I/O modes**: bytes, strings, lines, bufio.Reader, *bufio.Scanner

## Installation

```bash
go get gitlab.com/lyoneel/dirly
```

```go
import "gitlab.com/lyoneel/dirly"
```

Requires Go 1.26.11 or later. The library has no external dependencies.

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
| `NewDirectory(basePath string) *Directory` | Unfiltered directory. All files pass. Uses a fast path when no filters are configured. Default symlink policy. |
| `NewFilteredDirectory(basePath string) *DirectoryBuilder` | Builder for a filtered, immutable `Directory`. Chain configuration methods, then call `Build()`. |

Key operations on the resulting `*Directory`:

| Group | Methods |
|-------|---------|
| Inspection | `BasePath`, `ResolvedBasePath`, `SymlinkPolicy`, `FilterConfig`, `Exists`, `AbsPath` |
| Single-file read | `Open`, `ReadToBytes`, `ReadToString`, `ReadToLines`, `ReadToBuff`, `ReadToScanner` |
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
4. Symlinks are resolved and checked against the directory root, as described under Symlinks.

### Symlinks

The base path is resolved through symlinks once, when the `Directory` is created. A base path that is itself a symlink, such as `~/.config/app -> ~/Sync/app` for a config folder synced across devices, works under every policy, and its real location becomes the boundary. If the base path doesn't exist yet, it is resolved on each operation until it does.

Symlinks below the root follow the policy set with `WithSymlinkPolicy`:

| Policy | Behavior |
|--------|----------|
| `FollowWithinRoot` (default) | Symlinks are followed, but an operation whose resolved target is outside the root fails with an error wrapping `ErrPathEscapesRoot`. This also covers writes to new files and dangling links. |
| `FollowAny` | Symlinks are followed anywhere. This is the v1.1.0 behavior. `../` and absolute names are still rejected. |
| `NoFollow` | Any path through a symlink below the root fails with an error wrapping `ErrSymlinkNotAllowed`. |

```go
dir := dirly.NewFilteredDirectory("~/.config/app"). // may be a symlink to a synced folder
    WithSymlinkPolicy(dirly.FollowWithinRoot).        // the default
    Build()

_, err := dir.ReadToString("link-to-etc/passwd")
if errors.Is(err, dirly.ErrPathEscapesRoot) {
    // refused: the link resolves outside the directory
}
```

Removing a symlink removes the link, never its target. Writing through an allowed symlink updates the target file and leaves the link in place.

Limitations, stated plainly:

- `FollowWithinRoot` and `NoFollow` run file operations through Go's `os.Root`, opened at the resolved root. On Linux, macOS, BSD, and Windows, a symlink swapped in after dirly's check still cannot make an operation escape the root. A swap can still redirect an operation to a different file *inside* the root, and under `NoFollow` a symlink created after the check may be followed, though only within the root. On `GOOS=js`, `os.Root` itself is vulnerable to these races. `Chmod` and `Chown` on Unix share `os.Root`'s documented race, where a file replaced by a symlink mid-call may have the link changed instead.
- If the base path symlink is later pointed somewhere else, an existing `Directory` keeps the location it resolved at construction.
- Hard links and bind mounts are not detected. Anyone who can already write into the directory can still replace its files. The policy stops links from quietly pulling reads and writes outside the directory.

### Streaming reads with Open

`Open(name)` returns a plain `*os.File` for reading, with the same checks as the `Read*` helpers: relative names only, no traversal, and the symlink policy enforced. Under `FollowWithinRoot` and `NoFollow` the file is opened through `os.Root` at the resolved root. Use it when your code does its own buffering or scanning, for example a `bufio.Scanner` with a larger buffer; `ReadToBuff` and `ReadToScanner` are `Open` plus a `bufio` wrapper.

```go
f, err := dir.Open("notes/today.md")
switch {
case errors.Is(err, os.ErrNotExist): // os.IsNotExist(err) works too
case errors.Is(err, dirly.ErrPathEscapesRoot), errors.Is(err, dirly.ErrSymlinkNotAllowed):
    // refused by the symlink policy
case err != nil:
    return err
}
defer f.Close()
sc := bufio.NewScanner(f)
sc.Buffer(make([]byte, 0, 64*1024), 1<<20) // your own limits
```

Like the other read helpers, `Open` does not apply the builder's extension or pattern filters; filter with `Exists` or the glob methods first if needed. The returned file stays valid after dirly closes its `os.Root`.

### Writes and Ownership

Every write is atomic. Data goes to a temporary `.<name>.tmp-*` file in the target's directory, is fsynced, and is then renamed over the target. On POSIX the directory is then fsynced, best effort. A crash leaves either the old or the new content, and at worst a stale temp file.

- New files get the requested permissions (or the default), subject to umask. Existing files keep their mode and, on Unix, their owner and group.
- If `WithDefaultOwnership`, or keeping the original owner, can't be applied, the write fails with an error wrapping `ErrOwnership` and the target is left unchanged. Ownership is never ignored silently.
- Replacing a file needs write permission on its directory.

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
go test -race -cover ./...
```

Tests cover:
- Builder pattern methods (WithExtensions, Include, Exclude, MatchNested, CaseSensitive, WithDefaultOwnership, WithDefaultPermissions)
- File matching with various patterns (`*.yaml`, `config/*.yaml`, `**/*.yaml`)
- Both nested and flat matching modes for filters AND glob operations
- Path traversal protection (absolute paths, `..` sequences, encoded traversals)
- Symlink policies (escape refusal, in-root links, symlinked base paths, `FollowAny`, `NoFollow`)
- Atomic writes (no temp files left, mode preserved, writes through symlinks, concurrent writers, ownership errors)
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
