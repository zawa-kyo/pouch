# pouch Agent Guide

## Documentation sync

- Keep `README.md` and `README-ja.md` aligned in meaning.
- Keep `AGENTS.md` and `AGENTS-ja.md` aligned in meaning.
- When updating one file in a paired set, update the other in the same change unless the user explicitly asks for a temporary mismatch.
- Prefer natural Japanese in `*-ja.md` files rather than line-by-line translation.

## Purpose

This file captures the internal design and implementation expectations for `pouch`.
Use it while building the initial Go implementation from the external contract in `README.md`.

## Product boundary

- `pouch` is primarily a CLI.
- The reusable Go package exists to support the CLI and tests.
- Do not grow this project into a general scaffolding tool.
- Keep the implementation centered on path creation only.

## Primary behavior

- In auto mode, inspect only the final path segment.
- If the final segment contains a dot, treat the path as a file.
- Otherwise, treat the path as a directory.
- If a path is treated as a file, create missing parent directories first.
- If a path is treated as a directory, use `mkdir -p` semantics.

## Non-goals

- No template generation
- No file content generation
- No MIME or language detection
- No project initialization workflows
- No config file in the first release
- No interactive mode in the first release

## API direction

Prefer a small public surface.

Suggested package:

```go
package pouch
```

Suggested API:

```go
package pouch

import "os"

type Mode int

const (
    ModeAuto Mode = iota
    ModeFile
    ModeDir
)

type Kind int

const (
    KindFile Kind = iota
    KindDir
)

type Options struct {
    Mode     Mode
    DirPerm  os.FileMode
    FilePerm os.FileMode
    Touch    bool
    DryRun   bool
}

type Result struct {
    Path    string
    Kind    Kind
    Created bool
    Touched bool
}

func Detect(path string) Kind
func Create(path string, opts Options) (Result, error)
func CreateMany(paths []string, opts Options) ([]Result, error)
```

Notes:

- `Detect` should remain deterministic and side-effect free.
- `Create` should be the main unit of behavior.
- `CreateMany` should preserve input order.

## Detection rules

Reference logic:

1. Compute `filepath.Base(path)`.
2. If the base contains a dot, return file.
3. Otherwise, return directory.

Important examples:

- `sample` -> directory
- `sample.ts` -> file
- `sample/temp.ts` -> file
- `.env` -> file
- `Dockerfile` -> directory in auto mode
- `dir.with.dot` -> file in auto mode

Keep this logic simple. Do not add heuristics for hidden files, well-known filenames, or trailing slashes unless the external spec changes.

## CLI requirements

Expected flags:

- `--mode auto|file|dir`
- `--dry-run`
- `--verbose`
- `--no-touch`
- `--help`
- `--version`

Expected behavior:

- Exit `0` on full success.
- Exit non-zero on the first error.
- Write errors to stderr.
- In dry-run mode, do not mutate the filesystem.
- In verbose mode, print each planned or executed action in input order.

## Filesystem behavior

### File mode

- Create parent directories with `os.MkdirAll`.
- Create the file if it does not exist.
- If the file exists and touch is enabled, update timestamps.
- If touch is disabled and the file exists, leave timestamps unchanged.

### Directory mode

- Create the directory with `os.MkdirAll`.
- If it already exists, treat that as success.

## Permissions

Use explicit defaults in code:

- Directories: `0o755`
- Files: `0o644`

The effective permissions will still be influenced by the user's `umask`.

## Error handling

Wrap errors with path context.

Examples:

- `detect "sample.ts": ...`
- `create parent directory for "sample/temp.ts": ...`
- `create file "sample.ts": ...`
- `touch file "sample.ts": ...`

Prefer plain, direct error messages over custom error hierarchies in the first release.

## Package layout

Keep the initial layout shallow.

```text
.
├── cmd/
│   └── pouch/
│       └── main.go
├── internal/
│   └── cli/
│       └── flags.go
├── pouch.go
├── create.go
├── detect.go
├── types.go
└── README.md
```

Suggested responsibilities:

- `types.go`
  Public enums and option/result types.
- `detect.go`
  Detection logic only.
- `create.go`
  Core filesystem behavior.
- `pouch.go`
  Public entry points and small orchestration helpers.
- `cmd/pouch/main.go`
  CLI entry point.
- `internal/cli/flags.go`
  CLI flag parsing and validation.

Avoid splitting packages further unless the code clearly demands it.

## Implementation notes

### Path handling

- Use `path/filepath` from the standard library.
- Avoid shelling out to `mkdir`, `touch`, or other OS utilities.
- Prefer standard library behavior over custom path normalization.

### File creation

Recommended approach:

- Ensure the parent directory exists first.
- Use `os.OpenFile` with create-capable flags.
- Close files promptly.
- If touch behavior is enabled, update timestamps explicitly.

### Timestamps

Make touch behavior deliberate.

- A newly created file should count as `Created`.
- An existing file with timestamp update should count as `Touched`.
- A dry-run result should reflect the intended action without mutating the filesystem.

### Batch processing

- `CreateMany` should process inputs in order.
- Stop on the first error for v0.1.0.
- Return all successful results collected before the failure only if that shape is useful to the CLI.
- If the API becomes awkward, prefer a simpler contract and keep partial reporting inside the CLI layer.

## Testing strategy

Use table-driven tests where they keep the cases easy to scan.
Use `t.TempDir()` for filesystem isolation.

Required test coverage:

- Detection logic
- Directory creation
- File creation
- Parent directory creation
- Existing file touch behavior
- `--mode file` override
- `--mode dir` override
- Dotfile handling
- Ambiguous names such as `Dockerfile`
- Dot-containing directory names such as `dir.with.dot`
- Dry-run behavior

Representative cases:

- `sample` creates a directory
- `sample.ts` creates a file
- `sample/temp.ts` creates parent directories and file
- `.env` creates a file
- `Dockerfile` with auto mode creates a directory
- `Dockerfile` with file mode creates a file
- `dir.with.dot` with auto mode creates a file
- `dir.with.dot` with dir mode creates a directory

## Release plan

### v0.1.0

Include:

- CLI with `PATH...`
- Auto detection
- `--mode auto|file|dir`
- `--dry-run`
- `--verbose`
- `--no-touch`
- Public Go package
- Unit tests
- README

Do not include:

- Shell completion
- Config files
- Template generation
- Interactive prompts
- Plugin systems

## Review criteria

Before merging changes, check:

- Does the code still reflect the simple external rule?
- Is the public API smaller than the internal code structure?
- Are ambiguous cases documented rather than hidden?
- Does the CLI output stay predictable and low-noise?
- Are tests written from observable behavior, not implementation trivia?

## Change discipline

- Do not add heuristics silently.
- Do not broaden the scope without updating the README.
- If a new edge case requires a special rule, prefer an explicit mode override over smarter auto detection.
- Preserve the core identity: a small, predictable, path-aware `touch`.
