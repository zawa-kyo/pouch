# pouch Agent Guide

## Documentation sync

- Keep `README.md` and `README-ja.md` aligned in meaning.
- Keep `AGENTS.md` and `AGENTS-ja.md` aligned in meaning.
- Update both files in a language pair within the same change unless the user explicitly asks for a temporary mismatch.
- Prefer natural Japanese in `*-ja.md` files rather than line-by-line translation.

## Purpose

Use this file while building the first Go implementation of `pouch` from the contract in `README.md`.

## Scope

- `pouch` is primarily a CLI.
- The Go package exists to support the CLI and tests.
- Keep the project focused on path creation.
- Do not expand it into a general scaffolding tool.
- Target macOS and Linux only for v0.1.0.
- Do not add Windows-specific behavior unless the product scope changes.

## Product rules

### Detection

- In auto mode, first check whether the path ends with `/`.
- If the path ends with `/`, treat it as a directory.
- Otherwise, inspect only the final path segment.
- If the final segment contains a dot, treat the path as a file.
- Otherwise, treat the path as a directory.
- Do not add heuristics for hidden files or well-known filenames unless the external spec changes.

Examples:

- `sample` -> directory
- `sample.ts` -> file
- `sample/temp.ts` -> file
- `.env` -> file
- `dir.with.dot/` -> directory in auto mode
- `Dockerfile` -> directory in auto mode
- `dir.with.dot` -> file in auto mode

### Filesystem behavior

- If a path is treated as a file, create missing parent directories first.
- If the file does not exist, create it.
- If the file already exists, leave it unchanged.
- If the target path already exists as a directory, return an error.
- If the path ends with `/` and mode is `file`, return an error.
- If a path is treated as a directory, use `mkdir -p` semantics.
- If the directory already exists, treat that as success.
- If the target path already exists as a file, return an error.

### Non-goals

- No template generation
- No file content generation
- No MIME or language detection
- No project initialization workflows
- No config file in the first release
- No interactive mode in the first release

## CLI contract

Expected flags:

- `--mode auto|file|dir`
- `--dry-run`
- `--verbose`
- `--help`
- `--version`

Expected behavior:

- Exit `0` on full success.
- Exit non-zero on the first error.
- Write errors to stderr.
- In dry-run mode, do not mutate the filesystem.
- In verbose mode, print each planned or executed action in input order.
- Process input paths in order and stop at the first error.

## Package API

Keep the public surface small.

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

type Action int

const (
    ActionNone Action = iota
    ActionCreateFile
    ActionCreateDir
    ActionSkipExisting
)

type Options struct {
    Mode     Mode
    DirPerm  os.FileMode
    FilePerm os.FileMode
    DryRun   bool
}

type Result struct {
    Path   string
    Kind   Kind
    Action Action
}

func Detect(path string) Kind
func Create(path string, opts Options) (Result, error)
func CreateMany(paths []string, opts Options) ([]Result, error)
```

Notes:

- `Detect` should remain deterministic and side-effect free.
- `Create` should be the main unit of behavior.
- `CreateMany` should preserve input order.
- `CreateMany` should stop at the first error.
- If `CreateMany` fails, it should return the successful results collected before the failure together with the error.
- `Result.Action` should describe the intended or executed action.
- In dry-run mode, `Result.Action` should still report what would happen.

## Implementation notes

### Path handling

- Use `path/filepath` from the standard library.
- Avoid shelling out to `mkdir`, `touch`, or other OS utilities.
- Prefer standard library behavior over custom path normalization.

### File creation

- Ensure the parent directory exists first.
- Use `os.OpenFile` with create-capable flags.
- Close files promptly.
- Return success without changing timestamps when the file already exists.

### Permissions

Use explicit defaults in code:

- Directories: `0o755`
- Files: `0o644`

The effective permissions will still be influenced by the user's `umask`.

### Errors

Wrap errors with path context.

Examples:

- `detect "sample.ts": ...`
- `create parent directory for "sample/temp.ts": ...`
- `create file "sample.ts": ...`
- `create directory "sample": ...`

Prefer plain, direct error messages over custom error hierarchies in the first release.

### Package layout

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

Responsibilities:

- `types.go`: public enums and option/result types
- `detect.go`: detection logic only
- `create.go`: core filesystem behavior
- `pouch.go`: public entry points and small orchestration helpers
- `cmd/pouch/main.go`: CLI entry point
- `internal/cli/flags.go`: CLI flag parsing and validation

Do not split packages further unless the code clearly demands it.

## Testing

- Use table-driven tests when they make the cases easier to scan.
- Use `t.TempDir()` for filesystem isolation.
- Cover detection logic, file creation, directory creation, parent directory creation, explicit mode overrides, ambiguous names, and dry-run behavior.
- Cover type-conflict cases such as `--mode file` against an existing directory and `--mode dir` against an existing file.
- Cover trailing-slash cases in auto mode and `--mode file`.

Representative cases:

- `sample` creates a directory
- `sample.ts` creates a file
- `sample/temp.ts` creates parent directories and file
- `.env` creates a file
- `dir.with.dot/` with auto mode creates a directory
- `Dockerfile` with auto mode creates a directory
- `Dockerfile` with file mode creates a file
- `dir.with.dot` with auto mode creates a file
- `dir.with.dot` with dir mode creates a directory

## Release scope

### v0.1.0

Include:

- CLI with `PATH...`
- Auto detection
- `--mode auto|file|dir`
- `--dry-run`
- `--verbose`
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

- Does the code reflect the documented detection rule?
- Is the public API smaller than the internal structure?
- Are ambiguous cases documented rather than hidden?
- Does the CLI output stay predictable and low-noise?
- Are tests written from observable behavior rather than implementation trivia?

## Change discipline

- Do not add heuristics silently.
- Do not broaden the scope without updating the README.
- Prefer an explicit mode override over smarter auto detection.
- Preserve the core identity: a small, predictable path creation tool.
