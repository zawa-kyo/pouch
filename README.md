<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="./assets/logo.jpeg" alt="logo">
  A simple, path-aware touch command.
</p>

<div align="center">
  <a href="https://github.com/koki-develop/gat/releases/latest"><img src="https://img.shields.io/github/v/release/zawa-kyo/pouch" alt="release"></a>
  <a href="https://github.com/zawa-kyo/pouch?tab=MIT-1-ov-file"><img src="https://img.shields.io/github/license/zawa-kyo/pouch" alt="license"></a>
  <a href="https://github.com/zawa-kyo/pouch/actions/workflows/ci.yml"><img src="https://github.com/zawa-kyo/pouch/actions/workflows/ci.yml/badge.svg?branch=main" alt="ci"></a>
  <a href="https://github.com/zawa-kyo/pouch/tree/main"><img src="https://goreportcard.com/badge/github.com/zawa-kyo/pouch" alt="report"></a>
  <a href="https://github.com/zawa-kyo/pouch/tree/main"><img src="https://img.shields.io/github/repo-size/zawa-kyo/pouch" alt="size"></a>
</div>
<!-- markdownlint-enable MD033 -->

# 👜 pouch

`pouch` is a small CLI that creates a file or directory from a path.
It creates missing paths, but it leaves existing files unchanged.

It uses one small rule set in auto mode:

| Path shape                 | Result               |
| -------------------------- | -------------------- |
| Ends with `/`              | Treat as a directory |
| Final segment contains `.` | Treat as a file      |
| Otherwise                  | Treat as a directory |

That rule lets you create common paths without stopping to choose between `mkdir -p` and `touch`.

## Examples

```sh
pouch notes
pouch notes/today.md
pouch src/main.go test
```

| Command                  | Result                                      |
| ------------------------ | ------------------------------------------- |
| `pouch notes`            | Creates the `notes` directory               |
| `pouch notes/today.md`   | Creates parent directories, then `today.md` |
| `pouch src/main.go test` | Processes each path in input order          |

## Installation

### Go

```sh
go install github.com/zawa-kyo/pouch/cmd/pouch@latest
```

### mise

Use the GitHub backend directly:

```sh
mise use -g github:zawa-kyo/pouch@latest
```

If you prefer a short tool name, add an alias to your mise config:

```toml
[tool_alias]
pouch = "github:zawa-kyo/pouch"
```

Then install and activate it with:

```sh
mise use -g pouch@latest
```

## Why it exists

Creating paths often means switching between commands:

```sh
mkdir -p notes
mkdir -p src && touch src/main.go
```

The name `pouch` comes from that muscle memory: `mkdir -p` for directories, `touch` for files.
This tool folds those two habits into one small command that accepts a path and does the obvious thing with one detection rule set.

It is meant for the moment when you know the path you want, but you do not want to stop and spell out whether this one needs `mkdir -p`, `touch`, or both.

## Related tools

`pouch` sits on top of a familiar idea rather than replacing an existing standard tool.

- `mkdir -p` creates parent directories as needed.
- `touch` creates an empty file when the target does not exist.
- In Go, the closest building blocks are `os.MkdirAll` and `os.OpenFile`.

If you already like the explicit Unix primitives, keep using them. `pouch` is for the narrower case where you want the path itself to drive the operation.

## How auto mode works

Auto mode first checks whether the path ends with `/`. If it does not, it looks at the final path segment.

| Path             | Auto mode result |
| ---------------- | ---------------- |
| `sample`         | Directory        |
| `sample.ts`      | File             |
| `sample/temp.ts` | File             |
| `.env`           | File             |
| `dir.with.dot/`  | Directory        |

> [!NOTE]
> `pouch` keeps this rule intentionally small. It does not infer intent from well-known filenames or MIME types. A trailing slash is the only explicit directory hint in auto mode.

## When to use `--mode`

Some names are ambiguous under the auto rule:

| Path           | Auto mode result | Common override |
| -------------- | ---------------- | --------------- |
| `Dockerfile`   | Directory        | `--mode file`   |
| `Makefile`     | Directory        | `--mode file`   |
| `dir.with.dot` | File             | `--mode dir`    |

> [!IMPORTANT]
> `Dockerfile` and `Makefile` are treated as directories in auto mode. Use `--mode file` when you want file creation semantics.

Use `--mode` when you want a different result:

```sh
pouch --mode file Dockerfile
pouch --mode dir dir.with.dot
```

If a path ends with `/`, `--mode file` returns an error instead of creating a file.

## Behavior

| Target kind | Behavior                                                   |
| ----------- | ---------------------------------------------------------- |
| File        | Creates missing parent directories                         |
| File        | Creates the file if it does not exist                      |
| File        | Leaves an existing file unchanged                          |
| File        | Returns an error if the path already exists as a directory |
| Directory   | Creates the directory with `mkdir -p` semantics            |
| Directory   | Succeeds if the directory already exists                   |
| Directory   | Returns an error if the path already exists as a file      |

## CLI

Basic usage:

```sh
pouch [flags] PATH...
```

| Flag                             | Meaning                                               |
| -------------------------------- | ----------------------------------------------------- |
| `-h`, `--help`                   | Show help                                             |
| `-m`, `--mode <auto\|file\|dir>` | Force file or directory mode                          |
| `-n`, `--dry-run`                | Print planned actions without changing the filesystem |
| `-v`, `--version`                | Show version                                          |
| `-V`, `--verbose`                | Print each action in input order                      |

## Exit behavior

- Exit `0` on full success.
- Exit non-zero on the first error.
- Write errors to stderr.
- Process input paths in order and stop at the first error.

## Scope

`pouch` is intentionally narrow.

| Included in v0.1                          | Not included in v0.1    |
| ----------------------------------------- | ----------------------- |
| macOS support                             | Windows support         |
| Linux support                             | Template generation     |
| Path creation from CLI arguments          | File content generation |
| Simple auto detection                     | Project scaffolding     |
| Explicit mode overrides                   | Config files            |
| Reusable Go package for the CLI and tests | Interactive prompts     |
