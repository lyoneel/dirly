# Changelog

All notable changes in the dirly project are documented in this file. The format is based on Keep a Changelog, and the project follows Semantic Versioning.

## v1.2.0 - Unreleased

Minor release with a security-relevant behavior change: by default, symlinks that resolve outside the directory are now refused.

### Security

- Symlinks can no longer be used to read or write outside the directory. The path check used to be purely textual, so a symlink inside the directory (for example `link -> /etc`) let `ReadToString("link/hostname")` read `/etc/hostname`, and a symlinked file let `WriteFromString` overwrite a file outside the directory. The default policy now resolves every path and refuses targets outside the resolved root with an error wrapping `ErrPathEscapesRoot`.
- Default-policy operations go through `os.Root`, so a symlink swapped in after the check still cannot escape the root.

### Added

- `SymlinkPolicy` with `FollowWithinRoot` (default), `FollowAny`, and `NoFollow`, set with `DirectoryBuilder.WithSymlinkPolicy`
- `Directory.SymlinkPolicy()` and `Directory.ResolvedBasePath()`
- Error values `ErrPathEscapesRoot`, `ErrSymlinkNotAllowed`, `ErrOwnership`, and the `*SymlinkError` type
- The base directory is resolved through symlinks once, at construction, so a base path that is itself a symlink (for example `~/.config/app -> ~/Sync/app`) works, and its real location becomes the boundary. Listing (`GetAll*`, globs) now walks a symlinked base path instead of reporting it as a single file.

### Changed

- Behavior change: under the default `FollowWithinRoot`, reads, writes, `Exists`, `AbsPath`, `Chmod`, `Chown`, and `SubDirectory` refuse paths that resolve outside the root through a symlink. Use `WithSymlinkPolicy(dirly.FollowAny)` to keep the v1.1.0 behavior.
- Writes are atomic. Data goes to a temporary file in the same directory, which is fsynced and then renamed over the target. The directory is then fsynced on POSIX (best effort). Writing through an allowed symlink updates the link target and keeps the link.
- Existing files keep their mode and, on Unix, their owner and group when they are rewritten.
- Ownership failures are returned, not ignored. When `WithDefaultOwnership` (or keeping the original owner) cannot be applied, the write fails with an error wrapping `ErrOwnership`, and the target is left unchanged. Ownership is applied when either the UID or the GID is set (>= 0).
- Removing a symlink (`Remove`, `BatchRemove`, `BatchRemoveAllFromDir`) removes the link itself, never its target. Parent directories are checked against the policy.
- `NewFilteredDirectory` now defaults to no ownership change (UID/GID -1). It used to default to 0/0, a silently failing `chown` to root.
- `Build()` copies the filter configuration, so reusing a builder no longer changes a `Directory` that was already built.
- The minimum Go version is now 1.26.11, which clears the standard-library advisories reported against 1.26.0. CI now uses the pinned `golang:1.26.11` image.

### Upgrade Notes

- If your app intentionally links files or folders inside its directory to places outside it, opt back in with `WithSymlinkPolicy(dirly.FollowAny)`. A base path that is itself a symlink needs no change.
- Check errors with `errors.Is(err, dirly.ErrPathEscapesRoot)`, `errors.Is(err, dirly.ErrSymlinkNotAllowed)`, or `errors.Is(err, dirly.ErrOwnership)`.
- Atomic replacement needs write permission on the containing directory, not only on the file.

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
