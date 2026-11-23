# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/2.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2026-11-23

Minor release with a security-relevant behavior change: by default, symlinks that resolve outside the directory are now refused.

### Added

- `SymlinkPolicy` with `FollowWithinRoot` (default), `FollowAny`, and `NoFollow`, set with `DirectoryBuilder.WithSymlinkPolicy`.
- `Directory.SymlinkPolicy()` and `Directory.ResolvedBasePath()`.
- Error values `ErrPathEscapesRoot`, `ErrSymlinkNotAllowed`, `ErrOwnership`, and the `*SymlinkError` type. Check them with `errors.Is` / `errors.As`.
- A base path that is itself a symlink (for example `~/.config/app -> ~/Sync/app`) is resolved once, at construction, and its real location becomes the boundary. Listing (`GetAll*`, globs) now walks a symlinked base path instead of reporting it as a single file.

### Changed

- **Breaking:** Writes are atomic. Data goes to a temporary file in the same directory, which is fsynced and then renamed over the target; the directory is then fsynced on POSIX (best effort). Writing through an allowed symlink updates the link target and keeps the link. To upgrade, make sure the process can write to the containing directory, not only to the file.
- **Breaking:** Ownership failures are returned, not ignored. When `WithDefaultOwnership` (or keeping the original owner) cannot be applied, the write fails with an error wrapping `ErrOwnership`, and the target is left unchanged. Ownership is applied when either the UID or the GID is set (>= 0). To upgrade, handle `errors.Is(err, dirly.ErrOwnership)` or stop requesting ownership the process cannot apply.
- Existing files keep their mode and, on Unix, their owner and group when they are rewritten.
- Removing a symlink (`Remove`, `BatchRemove`, `BatchRemoveAllFromDir`) removes the link itself, never its target. Parent directories are checked against the policy.

### Fixed

- `NewFilteredDirectory` now defaults to no ownership change (UID/GID -1). It used to default to 0/0, a silently failing `chown` to root.
- `Build()` copies the filter configuration, so reusing a builder no longer changes a `Directory` that was already built.

### Security

- GO-2026-4602: **Breaking:** the minimum Go version is now 1.26.11, which clears this and the other standard-library advisories reported against 1.26.0. CI now uses the pinned `golang:1.26.11` image. To upgrade, build with Go 1.26.11 or later.
- **Breaking:** Symlinks can no longer be used to read or write outside the directory. The path check used to be purely textual, so a symlink inside the directory (for example `link -> /etc`) let `ReadToString("link/hostname")` read `/etc/hostname`, and a symlinked file let `WriteFromString` overwrite a file outside the directory. Under the default `FollowWithinRoot`, reads, writes, `Exists`, `AbsPath`, `Chmod`, `Chown`, and `SubDirectory` resolve every path and refuse targets outside the resolved root with an error wrapping `ErrPathEscapesRoot`. If your app intentionally links files or folders inside its directory to places outside it, opt back in with `WithSymlinkPolicy(dirly.FollowAny)`; a base path that is itself a symlink needs no change.
- Default-policy operations go through `os.Root`, so a symlink swapped in after the check still cannot escape the root.

## [1.1.0] - 2026-08-17

### Changed

- **Breaking:** The package is flattened to a single root package `dirly`; the former `lydir` subpackage and its leftover directory are gone. To upgrade, change imports from `gitlab.com/lyoneel/dirly/lydir` to `gitlab.com/lyoneel/dirly`.

## [1.0.0] - 2026-08-13

### Added

- First stable release of the standalone library, with no dependencies beyond the Go standard library.
- Cross-platform directory operations using standard `filepath` functions.
- Builder pattern with immutable filter configuration: `WithExtensions`, `Include`, `Exclude`, `Match`, `MatchNested`, `CaseSensitive`, `WithDefaultOwnership`, `WithDefaultPermissions`.
- File matching with simple wildcards, path patterns, and recursive `**/` patterns.
- Case-sensitive or case-insensitive matching with OS default behavior.
- Path traversal protection for file operations and glob patterns.
- Single-file and batch operations across five I/O modes: bytes, strings, lines, `*bufio.Reader`, and `*bufio.Scanner`.
- File management operations: `Chown`, `Chmod`, `Remove`, `BatchRemove`, `BatchRemoveAllFromDir`.
- Glob search with absolute and relative result paths.

## [0.0.0] - 2026-07-02

### Added

- Initial public snapshot of the library with README and MIT license.

[Unreleased]: https://gitlab.com/lyoneel/dirly/-/compare/v1.2.0...HEAD
[1.2.0]: https://gitlab.com/lyoneel/dirly/-/compare/v1.1.0...v1.2.0
[1.1.0]: https://gitlab.com/lyoneel/dirly/-/tags/v1.1.0
