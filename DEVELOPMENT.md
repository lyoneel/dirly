# dirly Developer Reference

Developer guide for the dirly package: architecture, full API reference, matching semantics, security internals, and the extended example set. The README holds the overview and quick start; this document holds the depth.

> Main repository: https://gitlab.com/lyoneel/dirly
> Any other host that serves this repository is a mirror. Open issues
> and merge requests on GitLab.

## Architecture

The package is a single root package with these source files:

| File | Responsibility |
|------|----------------|
| `directory.go` | The `Directory` type: all read, write, glob, and management operations |
| `builder.go` | The `DirectoryBuilder` fluent API and filter immutability |
| `filter.go` | The `FilterConfig` type and the include, exclude, and extension evaluation logic |
| `*_test.go` | Unit, edge-case, nested-matching, and batch test suites |

Design patterns:

- Builder pattern for construction. `NewFilteredDirectory` returns a builder; `Build()` produces an immutable `Directory`.
- Embedding of standard I/O types (`*os.File`, `*bufio.Reader`, `*bufio.Scanner`) so callers keep full control of streaming and closing.
- Defensive copies for configuration access (`FilterConfig()` returns a copy).

Extension points:

- Add a new I/O mode by pairing a `ReadTo*` method with a `WriteFrom*` method and their batch variants.
- Add a new filter rule inside `filter.go` and wire it into the evaluation order below.
- Keep the fast path in `NewDirectory` free of filter checks; it exists so unfiltered use pays no filter overhead.

## API Reference

### Directory Creation

#### `NewDirectory(basePath string) *Directory`

Creates an unfiltered directory instance. All files are returned without filtering. Uses a fast path optimization when no filters are configured.

```go
dir := dirly.NewDirectory("/path/to/dir")
allFiles, _ := dir.GetAllRel() // Returns all files
```

#### `NewFilteredDirectory(basePath string) *DirectoryBuilder`

Returns a builder for creating filtered directories with immutable configuration. The builder allows chaining multiple configuration methods before calling `Build()` to create the final Directory instance.

### Builder Methods

All builder methods return the same builder instance for method chaining. Configuration is applied when `Build()` is called, and filters cannot be modified afterward.

| Method | Description |
|--------|-------------|
| `WithExtensions(extensions ...string)` | Add file extensions to filter by (e.g., `"yaml"`, `"json"`). Extensions are compared case-insensitively after removing the leading dot. |
| `Include(patterns ...string)` | Add include patterns that files must match to be included. Panics if any pattern starts with `"!"`. Patterns support wildcards (`*`, `?`) and path components. |
| `Exclude(patterns ...string)` | Add exclude patterns that remove matching files from results. Panics if any pattern starts with `"!"`. Exclude patterns are applied after include patterns. |
| `Match(patterns []string, include bool)` | Low-level API to add patterns without validation. If `include` is true, patterns go to IncludePatterns; otherwise to ExcludePatterns. Use for advanced use cases. |
| `MatchNested(match bool)` | Enable/disable nested path matching (default: `true`). When false, only filenames are matched unless the pattern contains `/`. When true, both filename and full relative path are checked. |
| `CaseSensitive(sensitive bool)` | Set explicit case sensitivity (default: OS behavior via `filepath.Match`). Set to true for always case-sensitive matching regardless of platform. |
| `WithDefaultOwnership(uid, gid int)` | Set default UID/GID for all new files written to this directory. Use `-1` to skip setting a specific value. Ownership is applied as best-effort after file creation. |
| `WithDefaultPermissions(perm os.FileMode)` | Set default permissions for all new files written to this directory. Use `0` to skip setting permissions and use the provided value instead. Defaults to `0644` when neither perm nor defaultPerm is set. |
| `Build() *Directory` | Create the final Directory instance with configured filters. Filters are immutable after this point. |

### File Operations

#### Basic Operations

| Method | Description |
|--------|-------------|
| `BasePath() string` | Returns the base directory path that all operations are scoped to. |
| `FilterConfig() FilterConfig` | Returns a copy of the filter configuration, including include patterns, exclude patterns, and allowed extensions. |
| `Exists(filename string) bool` | Checks if a file exists in this directory and passes all configured filters. Returns false for directories or non-existent files. |
| `AbsPath(name string) (string, error)` | Converts a relative path name to its absolute resolved path. Validates against path traversal attacks. |

#### Read Operations

**Single File Reads:**

| Method | Description |
|--------|-------------|
| `ReadToBytes(filename string) ([]byte, error)` | Reads the entire file into memory as raw bytes. Returns an error if the filename is empty or the file cannot be read. |
| `ReadToString(filename string) (string, error)` | Reads the file and converts its contents to a string. Internally uses ReadToBytes then converts to string. |
| `ReadToLines(filename string) ([]string, error)` | Reads the file and splits it into lines by newline characters (`\n`). Trailing newlines are handled automatically (empty last element removed if present). |
| `ReadToBuff(filename string) (*os.File, *bufio.Reader, error)` | Opens a file and returns both the file handle and a buffered reader for efficient streaming reads. Caller is responsible for closing the file. Returns an error if filename is empty or file cannot be opened. |
| `ReadToScanner(filename string) (*os.File, *bufio.Scanner, error)` | Opens a file and returns both the file handle and a scanner for line-by-line reading. Scanner uses default buffer size (64KB). Caller is responsible for closing the file. Returns an error if filename is empty or file cannot be opened. |

**Batch Reads:**

| Method | Description |
|--------|-------------|
| `BatchReadToBytes(filenames []string) (map[string][]byte, error)` | Reads multiple files and returns a map from filename to byte contents. Returns an error if any single file fails to read; all successfully read files are still included in the result map. Empty input returns empty map. |
| `BatchReadToString(filenames []string) (map[string]string, error)` | Reads multiple files as strings. Same behavior as BatchReadToBytes but with string values. |
| `BatchReadToLines(filenames []string) (map[string][]string, error)` | Reads multiple files and splits each into lines. Returns a map from filename to line slices. Handles trailing newlines per file independently. |
| `BatchReadToBuff(filenames []string) (map[string]*bufio.Reader, error)` | Opens multiple files and returns buffered readers for streaming. Caller is responsible for closing all returned files. |
| `BatchReadToScanner(filenames []string) (map[string]*bufio.Scanner, error)` | Opens multiple files and returns scanners for line-by-line reading. Caller is responsible for closing all returned files. |

#### Write Operations

**Single File Writes:**

| Method | Description |
|--------|-------------|
| `WriteFromBytes(filename string, data []byte, perm os.FileMode) error` | Writes byte content to a file. If `perm` is 0 and defaultPerm is configured, uses defaultPerm; otherwise falls back to `0644`. Creates parent directories as needed. Applies ownership if configured. |
| `WriteFromString(filename string, s string, perm os.FileMode) error` | Writes string content to a file by converting to bytes first. Same permission and ownership handling as WriteFromBytes. |
| `WriteFromLines(filename string, lines []string, perm os.FileMode) error` | Writes multiple lines to a file, joining them with newline characters. Each line gets a trailing newline. Same permission and ownership handling. |
| `WriteFromBuff(filename string, r *bufio.Reader, perm os.FileMode) error` | Reads all data from a buffered reader and writes it to a file. Useful for streaming large files without loading everything into memory first. |
| `WriteFromScanner(filename string, s *bufio.Scanner, perm os.FileMode) error` | Reads lines from a scanner and writes them to a file with trailing newlines. Ideal for processing line-by-line input and writing output. |

**Batch Writes:**

| Method | Description |
|--------|-------------|
| `BatchWriteFromBytes(files map[string][]byte) error` | Writes multiple files from byte maps using default permissions (from Directory config or 0644). Returns an error if any single file fails; other writes may have succeeded. Empty input returns nil. |
| `BatchWriteFromString(files map[string]string) error` | Writes multiple files from string maps. Same behavior as BatchWriteFromBytes but with string values. |
| `BatchWriteFromLines(files map[string][]string) error` | Writes multiple files from line slice maps. Each file gets lines joined with newlines. |
| `BatchWriteFromBuff(files map[string]*bufio.Reader) error` | Writes multiple files from buffered readers. Useful for streaming writes without loading all data into memory. |
| `BatchWriteFromScanner(files map[string]*bufio.Scanner) error` | Writes multiple files from scanners, processing each line-by-line. |

#### Pattern Matching

| Method | Description |
|--------|-------------|
| `GetAllByGlobAbs(pattern string) ([]string, error)` | Searches for files matching a glob pattern and returns absolute paths. When filters are configured, the glob operates on filtered results and respects `matchNested`: with true, matches filename AND path; with false, only root-level files. Rejects patterns containing `..`. |
| `GetAllByGlobRel(pattern string) ([]string, error)` | Same as GetAllByGlobAbs but returns relative paths instead of absolute ones. Converts each match using `filepath.Rel`. |

**How Glob Patterns Work:**

The glob methods first retrieve all files (applying filters if configured), then match the pattern against those results. The `matchNested` setting controls matching behavior:

| Pattern Type | Behavior with `matchNested=true` | Behavior with `matchNested=false` |
|--------------|----------------------------------|-----------------------------------|
| **Simple** (`*.yaml`) | Matches filename anywhere in tree (root + all subdirectories) | Only matches files directly in root directory |
| **Path** (`config/*.yaml`) | Matches files in config/ AND nested paths under it | Only matches files directly in config/ (not deeper) |
| **Recursive** (`**/*.yaml`) | Matches at any depth recursively | Treated as simple pattern - only root level |

The pattern is matched against the filtered file list using Go's `filepath.Match` for filename matching and custom logic for path-based patterns.

**Important Notes:**
- Patterns containing `..` are rejected before any file system access occurs
- Results are validated to ensure they stay within the base directory
- When filters are configured, glob operates on filtered results (not raw filesystem)
- The pattern is treated as a relative path from the base directory

#### File Management

| Method | Description |
|--------|-------------|
| `Chown(filename string, uid, gid int) error` | Changes the owner and group of an existing file. Returns an error if filename is empty or chown fails (may fail on non-root systems). |
| `Chmod(filename string, perm os.FileMode) error` | Changes the permissions of an existing file. Returns an error if filename is empty or chmod fails. |
| `Remove(filename string) error` | Deletes a single file from the directory. Returns an error if filename is empty or removal fails. Does not follow symlinks. |
| `BatchRemoveAllFromDir(path string) error` | Recursively removes a file or directory and all its contents using `os.RemoveAll`. Useful for cleaning up entire subdirectories. |
| `BatchRemove(paths []string) error` | Removes multiple files or directories by their paths. Returns an error if any single removal fails, but continues attempting others first. Empty input returns nil. |

### Filter Configuration Access

| Method | Description |
|--------|-------------|
| `FilterConfig() FilterConfig` | Returns a defensive copy of the current filter configuration, including all include patterns, exclude patterns, and allowed extensions. Useful for debugging or inspecting active filters. |

## File Matching Behavior

### Pattern Types (for Filters)

The package supports several pattern types for file matching when using filters (Include/Exclude/WithExtensions):

#### 1. Simple Wildcards (`*.yaml`)

Matches filenames using standard glob patterns with `filepath.Match`. Supports:
- `*` - Matches any sequence of characters (except path separators)
- `?` - Matches any single character (except path separators)
- `[abc]` - Matches any character in the set
- `[a-z]` - Matches any character in the range

```go
dir := dirly.NewFilteredDirectory("/path").
    Include("*.yaml", "*.json").  // Filter-based matching
    Build()

// With matchNested=true: Matches config.yaml, data.yaml, test.yaml anywhere
// With matchNested=false: Only matches *.yaml files in root directory
files, _ := dir.GetAllRel()
```

#### 2. Path Patterns (`config/*.yaml`)

Matches files in specific subdirectories using path components in the pattern.

```go
dir := dirly.NewFilteredDirectory("/path").
    Include("config/*.yaml").
    Build()

// With matchNested=true: Matches config/settings.yaml, config/backup/data.yaml
// With matchNested=false: Only matches config/settings.yaml (not deeper)
files, _ := dir.GetAllRel()
```

#### 3. Recursive Patterns (`**/*.yaml`)

The `**/` prefix enables recursive matching at any directory depth when filters are applied.

```go
dir := dirly.NewFilteredDirectory("/path").
    Include("**/*.yaml").
    Build()

// With matchNested=true: Matches root.yaml, config/data.yaml, deep/nested/test.yaml
// With matchNested=false: Only matches *.yaml files in root directory (no recursion)
files, _ := dir.GetAllRel()
```

**Note:** These pattern types apply to filter-based matching. For glob operations (`GetAllByGlob*`), see the "Pattern Matching" section above for how `matchNested` affects behavior.

### Matching Modes

The `MatchNested` setting controls how patterns are evaluated when using filters (Include/Exclude/WithExtensions) or glob operations:

#### `matchNested = true` (Default)

Matches against both filename and full relative path. This provides flexible matching where:
- Simple patterns like `*.yaml` match the filename anywhere in the tree
- Path patterns like `config/*.yaml` match files in config/ or any nested subdirectory
- Recursive patterns like `**/*.yaml` match at any depth

| Pattern | Matches | Does NOT Match |
|---------|---------|----------------|
| `*.yaml` | Any `.yaml` file anywhere (root, config/, deep/nested/) | Non-YAML files |
| `config/*.yaml` | Files in `config/` + all nested paths under it | Files outside config/ tree |
| `**/*.yaml` | All YAML files recursively at any depth | Non-YAML files |

```go
dir := dirly.NewFilteredDirectory("/path").
    MatchNested(true).  // default behavior
    Build()

// Gets ALL yaml files from all subdirectories
files, _ := dir.GetAllByGlobRel("*.yaml")
```

#### `matchNested = false` (Flat Mode)

Matches only filenames by default. **Path patterns with `/` are still supported** for specific directory filtering, but without recursive matching.

| Pattern | Matches | Does NOT Match |
|---------|---------|----------------|
| `*.yaml` | Only `.yaml` files directly in root directory | Files in subdirectories |
| `config/*.yaml` | Files directly in `config/` subdirectory only | Files in deeper paths like `config/subdir/file.yaml` |
| `**/*.yaml` | Same as `*.yaml` (no recursive matching) - only root level | All nested files |

```go
dir := dirly.NewFilteredDirectory("/path").
    MatchNested(false).  // flat mode!
    Build()

// Gets ONLY yaml files directly in root directory
files, _ := dir.GetAllByGlobRel("*.yaml")

// Gets ONLY yaml files directly in config/ subdirectory (not nested deeper)
configFiles, _ := dir.GetAllByGlobRel("config/*.yaml")
```

**Important:** When `matchNested=false`, glob patterns do NOT recurse into subdirectories. Only files at the root level match simple patterns like `*.yaml`.

### Filter Evaluation Order

When using `GetAllAbs()` or `GetAllRel()`, filters are applied in this specific order:

1. **Include patterns** - Only files matching include patterns proceed to next step
2. **Exclude patterns** - Files matching exclude patterns are removed from results
3. **Extension filter** - Only files with allowed extensions remain

```go
dir := dirly.NewFilteredDirectory("/path").
    Include("*.yaml", "*.json").      // Step 1: Keep only yaml/json
    Exclude("*.tmp.yaml").            // Step 2: Remove temp files
    WithExtensions("yaml").           // Step 3: Keep only yaml
    Build()

// Result: Only .yaml files that are NOT *.tmp.yaml
files, _ := dir.GetAllRel()
```

**Important:** If no include patterns are specified, all files pass step 1. The extension filter is always applied last regardless of configuration order.

## Security Features

### Path Traversal Protection

The package prevents directory traversal attacks by validating all paths at multiple levels:

#### Input Validation

All file operations validate input to reject dangerous paths:

```go
dir := dirly.NewFilteredDirectory("/safe/path").Build()

// ✅ Allowed - relative path within base
dir.ReadToBytes("config.yaml")

// ❌ Blocked - absolute path attempt
dir.ReadToBytes("/etc/passwd")  // Error: "absolute paths not allowed"

// ❌ Blocked - parent directory traversal
dir.ReadToBytes("../etc/passwd")  // Error: "path traversal detected"

// ❌ Blocked - encoded traversal
dir.ReadToBytes("config/../../etc/passwd")  // Error: "path traversal detected"
```

#### Path Resolution

The `resolvePath` method:
1. Rejects absolute paths immediately
2. Joins and cleans the path using `filepath.Join` and `filepath.Clean`
3. Verifies the resolved path stays within the base directory
4. Allows exact match for basePath itself (for edge cases)

#### Glob Pattern Validation

Glob patterns containing `..` are rejected before any file system access:

```go
// ❌ Blocked - traversal in glob pattern
dir.GetAllByGlobRel("../../etc/passwd")  // Error: "path traversal not allowed in glob pattern"

// ✅ Allowed - safe recursive search
dir.GetAllByGlobRel("**/*.yaml")  // Works correctly
```

#### Result Validation

After glob operations, all results are validated to ensure they stay within the base directory using `isPathWithinBase`, providing defense-in-depth against any edge cases.

## Additional Examples

### Example 4: Exclude Temporary Files

```go
dir := dirly.NewFilteredDirectory("./data").
    Include("*.yaml").
    Exclude("*.tmp.yaml", "*.bak").
    Build()

// Get only production YAML files, excluding temp and backup files
files, _ := dir.GetAllRel()
```

### Example 5: Batch File Operations with Bytes

```go
dir := dirly.NewFilteredDirectory("./templates").Build()

// Read multiple templates at once as bytes
contents, err := dir.BatchReadToBytes([]string{"header.html", "footer.html"})
if err != nil {
    panic(err)
}

// Write multiple files at once using default permissions from Directory
err = dir.BatchWriteFromBytes(map[string][]byte{
    "output1.txt": []byte("content1"),
    "output2.txt": []byte("content2"),
})
if err != nil {
    panic(err)
}
```

### Example 6: Stream Large Files with Scanner

```go
dir := dirly.NewFilteredDirectory("./logs").Build()

// Get scanner for line-by-line reading of large log files
file, scanner, err := dir.ReadToScanner("app.log")
if err != nil {
    panic(err)
}
defer file.Close()

for scanner.Scan() {
    line := scanner.Text()
    // Process each line
    fmt.Println(line)
}

if err := scanner.Err(); err != nil {
    panic(err)
}
```

### Example 7: Set Default File Permissions and Ownership

```go
dir := dirly.NewFilteredDirectory("./config").
    WithDefaultOwnership(1000, 1000).      // UID/GID for new files
    WithDefaultPermissions(0644).          // Permissions for new files
    Build()

// All WriteFrom* operations will use these defaults unless specified otherwise
err := dir.WriteFromString("settings.yaml", "key: value\n", 0)
if err != nil {
    panic(err)
}
```

### Example 8: Read and Process Multiple Files as Lines

```go
dir := dirly.NewFilteredDirectory("./scripts").Build()

// Read multiple shell scripts as line slices
scripts, err := dir.BatchReadToLines([]string{"build.sh", "test.sh"})
if err != nil {
    panic(err)
}

// Process each script
for name, lines := range scripts {
    fmt.Printf("Script: %s (%d lines)\n", name, len(lines))
    for i, line := range lines {
        fmt.Printf("  Line %d: %s\n", i+1, line)
    }
}
```

### Example 9: Case-Sensitive Matching on Linux

```go
dir := dirly.NewFilteredDirectory("./src").
    CaseSensitive(true).
    Include("*.Go", "*.go").  // Both cases explicitly listed
    Build()

// On case-sensitive filesystems, this matches both main.Go and main.go
files, _ := dir.GetAllRel()
```

### Example 10: Complex Filter Chain

```go
dir := dirly.NewFilteredDirectory("./data").
    Include("*.yaml", "*.json").           // Only config files
    Exclude("*.dev.yaml", "*-test.json").  // Remove dev/test variants
    WithExtensions("yaml").                // Keep only YAML (overrides include)
    Build()

// Result: All .yaml files except *-test.yaml files
files, _ := dir.GetAllRel()
```

### Example 11: Using Buffered Readers for Memory Efficiency

```go
dir := dirly.NewFilteredDirectory("./large-files").Build()

// Open file with buffered reader for streaming
file, reader, err := dir.ReadToBuff("huge.log")
if err != nil {
    panic(err)
}
defer file.Close()

// Read in chunks without loading entire file into memory
buf := make([]byte, 4096)
for {
    n, err := reader.Read(buf)
    if n > 0 {
        processChunk(buf[:n])
    }
    if err == io.EOF {
        break
    }
    if err != nil {
        panic(err)
    }
}
```

### Example 12: Batch Operations with Error Handling

```go
dir := dirly.NewFilteredDirectory("./backup").Build()

// Read multiple files, handling partial failures
contents, err := dir.BatchReadToBytes([]string{"file1.txt", "missing.txt", "file2.txt"})
if err != nil {
    // contents still contains successfully read files
    fmt.Printf("Partial success: %d files read\n", len(contents))
}

// Write multiple files, all-or-nothing semantics per file
err = dir.BatchWriteFromBytes(map[string][]byte{
    "out1.txt": []byte("data1"),
    "out2.txt": []byte("data2"),
})
if err != nil {
    // Some writes may have succeeded before the error occurred
    fmt.Printf("Write failed: %v\n", err)
}
```
