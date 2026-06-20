<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="./assets/logo.jpeg" alt="logo" width="400">
</p>

<div align="center">
  <a href="https://github.com/zawa-kyo/pouch/releases/latest"><img src="https://img.shields.io/github/v/release/zawa-kyo/pouch" alt="release"></a>
  <a href="https://github.com/zawa-kyo/pouch?tab=MIT-1-ov-file"><img src="https://img.shields.io/github/license/zawa-kyo/pouch" alt="license"></a>
  <a href="https://github.com/zawa-kyo/pouch/actions/workflows/ci.yml"><img src="https://github.com/zawa-kyo/pouch/actions/workflows/ci.yml/badge.svg?branch=main" alt="ci"></a>
  <a href="https://goreportcard.com/report/github.com/zawa-kyo/pouch"><img src="https://goreportcard.com/badge/github.com/zawa-kyo/pouch" alt="report"></a>
  <a href="https://github.com/zawa-kyo/pouch/tree/main"><img src="https://img.shields.io/github/repo-size/zawa-kyo/pouch" alt="size"></a>
</div>
<!-- markdownlint-enable MD033 -->

# 👜 pouch

`pouch` creates files and directories from path-like CLI arguments.
It creates missing paths and leaves existing files unchanged.

It is for the moment when you already know the path you want, but you do not want to stop and spell out whether this one needs `mkdir -p`, `touch`, or both.

## Why pouch

Creating paths often means switching between commands:

```sh
mkdir -p notes
mkdir -p src && touch src/main.go
```

The name `pouch` comes from that muscle memory: `mkdir -p` for directories, `touch` for files.
`pouch` folds those two habits into one small command. You pass a path, and it follows one simple rule set to do the obvious thing:

| Path shape                 | Result               |
| -------------------------- | -------------------- |
| Ends with `/`              | Treat as a directory |
| Final segment contains `.` | Treat as a file      |
| Otherwise                  | Treat as a directory |

That rule is intentionally small. `pouch` is not a scaffolding tool; it just creates the path you asked for.

## Examples

```sh
pouch foo
pouch bar/baz.go
pouch src/main.go test
```

| Command                  | Result                                      |
| ------------------------ | ------------------------------------------- |
| `pouch foo`              | Creates the `foo` directory                 |
| `pouch bar/baz.go`       | Creates parent directories, then `baz.go`   |
| `pouch src/main.go test` | Processes each path in input order          |

In practice it looks like this:

<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="./assets/demo.gif" alt="demo" width="640">
</p>
<!-- markdownlint-enable MD033 -->

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

## Compared with `mkdir -p` and `touch`

`pouch` sits on top of a familiar idea rather than replacing an existing standard tool.

- `mkdir -p` creates parent directories as needed.
- `touch` creates an empty file when the target does not exist.

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

When that rule matches your intent, auto mode is enough. When it does not, use `--file`, `--dir`, or `--mode` to be explicit.

## When to choose file or directory mode

Some names are ambiguous under the auto rule:

| Path           | Auto mode result | Common override |
| -------------- | ---------------- | --------------- |
| `Dockerfile`   | Directory        | `--file`        |
| `Makefile`     | Directory        | `--file`        |
| `dir.with.dot` | File             | `--dir`         |

> [!IMPORTANT]
> `Dockerfile` and `Makefile` are treated as directories in auto mode. Use `--file` when you want file creation semantics.

Use `--file` or `--dir` when you want a different result:

```sh
pouch Dockerfile --file
pouch dir.with.dot --dir
```

Mode flags apply to the whole command. For example, `pouch Dockerfile test --file` treats both `Dockerfile` and `test` as files.

`--mode file` and `--mode dir` remain available. Do not combine `--mode` with `--file` or `--dir` in the same command.

If a path ends with `/`, file mode returns an error instead of creating a file.

## When to use `--strict`

By default, `pouch` is safe to run again. If the target already exists with the expected kind, the command succeeds without changing it.

Use `--strict` when an existing target should fail instead:

```sh
pouch src/main.go test --strict
```

## Behavior

| Target kind | Behavior                                                   |
| ----------- | ---------------------------------------------------------- |
| File        | Creates missing parent directories                         |
| File        | Creates the file if it does not exist                      |
| File        | Leaves an existing file unchanged by default               |
| File        | Returns an error if the path already exists as a directory |
| Directory   | Creates the directory with `mkdir -p` semantics            |
| Directory   | Succeeds if the directory already exists                   |
| Directory   | Returns an error if the path already exists as a file      |

## CLI

Basic usage:

```sh
pouch [flags] PATH...
```

Flags can appear before or after `PATH...`. Use `--` if a path itself starts with `-`.

| Flag                             | Meaning                                               |
| -------------------------------- | ----------------------------------------------------- |
| `-h`, `--help`                   | Show help                                             |
| `--file`                         | Treat each path as a file                             |
| `--dir`                          | Treat each path as a directory                        |
| `-m`, `--mode <auto\|file\|dir>` | Force file or directory mode                          |
| `-n`, `--dry-run`                | Print planned actions without changing the filesystem |
| `-s`, `--strict`                 | Fail if a target already exists                       |
| `-v`, `--version`                | Show version                                          |
| `-V`, `--verbose`                | Print each action in input order                      |

## Exit behavior

- Exit `0` on full success.
- Exit non-zero on the first error.
- Write errors to stderr.
- Process input paths in order and stop at the first error.

## Scope

`pouch` is intentionally narrow.

- Platform: focus on macOS and Linux.
- Responsibility: turn CLI paths into files or directories. It does not define project structure or file contents.
- Detection: use one small auto detection rule set, with `--file`, `--dir`, or `--mode` when explicit control matters.
- UX: keep each invocation predictable and non-interactive.
- Configuration: keep behavior local to each command instead of relying on config files.
