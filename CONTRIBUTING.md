# Contributing to dirly

Thank you for considering a contribution. This document explains how to set up the development environment, follow the project conventions, and submit changes.

> Main repository: https://gitlab.com/lyoneel/dirly
> Any other host that serves this repository is a mirror. Open issues
> and merge requests on GitLab.

## Code of Conduct

This project keeps collaboration practical and respectful. Treat maintainers and contributors as professional peers.

## Development Environment Setup

1. Install Go 1.26 or later. Check with `go version`.
2. Clone the repository:

```bash
git clone https://gitlab.com/lyoneel/dirly.git
cd dirly
```

3. Confirm the toolchain works:

```bash
go build ./...
go test ./...
```

The library has no external dependencies. No vendor directory, no module proxy tricks, no code generation step.

## Project Layout

| Path | Content |
|------|---------|
| `directory.go` | The `Directory` type and all file operations |
| `builder.go` | The `DirectoryBuilder` fluent construction API |
| `filter.go` | Filter configuration and evaluation logic |
| `*_test.go` | Unit tests, edge-case tests, nested-matching tests, batch tests |

## Code Style Guidelines

1. Follow the conventions of Effective Go and the Go Code Review Comments guide.
2. Run `gofmt` on every changed file. The project has no custom formatter configuration.
3. Keep the public API stable within a major version. Breaking changes require a major version bump and a CHANGELOG entry.
4. Document every exported symbol with a doc comment that starts with the symbol name.
5. Return errors rather than panicking. The two builder methods that panic (`Include`, `Exclude` on invalid patterns) are documented exceptions.
6. Keep path traversal protection intact. Every new file-access path must resolve through `resolvePath` or glob validation.

## Testing Requirements

1. Run the full suite before you submit anything:

```bash
go test ./... -cover
```

2. New features need table-driven tests in the style of the existing `*_test.go` files.
3. Bug fixes need a regression test that fails without the fix.
4. Keep coverage at or above the current 85 percent. The CI pipeline runs `go test -race -cover ./...` on every merge request.
5. Tests must pass with `-race`. The batch and streaming paths must stay safe for concurrent use.

## Git Workflow

1. The default branch is the integration target. Create a feature branch from it:

```bash
git checkout -b feat/my-change
```

2. Use conventional commit messages in the form `type: message` or `type!: message` for breaking changes. Known types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`, `perf`.

```bash
git commit -m "feat: add symlink-aware Exists check"
```

3. Keep each commit focused. One logical change per commit.
4. Rebase your branch on the default branch before you open a merge request.

## Pull Request Process

1. Push your branch to the same repository and open a merge request on GitLab.
2. Fill in the merge request template: describe the change, the type, and how you tested it.
3. The CI pipeline must pass: vet, tests with race detection and coverage, build.
4. Update documentation in the same change set:
   - README.md for user-facing behavior
   - DEVELOPMENT.md for internals and matching semantics
   - CHANGELOG.md under an Unreleased heading

## Code Review Expectations

1. A maintainer reviews every merge request. Expect a review within a few days.
2. Reviewers check correctness, test coverage, API stability, and documentation accuracy.
3. Address review comments with new commits; keep the merge request history readable.
4. Squash only at the discretion of the maintainer who merges.

## Onboarding for New Contributors

1. Read the README quick start, then DEVELOPMENT.md for the architecture and matching semantics.
2. Pick an issue labeled with the bug label as a first change.
3. Build and run the suite first; the tests document the expected behavior of every operation.
4. Ask questions in the merge request or in an issue. Questions in the open are welcome.

## License

The project uses the MIT license. All contributions are submitted under the MIT license. See the LICENSE file for the full text.
