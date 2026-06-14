# pouch

`pouch` is a small CLI that creates a file or directory from a path.

It uses one rule in auto mode:

- If the final path segment contains a dot (`.`), treat the path as a file.
- Otherwise, treat the path as a directory.

That rule lets you create common paths without stopping to choose between `mkdir -p` and `touch`.

## Examples

```sh
pouch notes
pouch notes/today.md
pouch src/main.go test
```

What those commands do:

- `pouch notes` creates the `notes` directory.
- `pouch notes/today.md` creates parent directories and then creates `today.md`.
- `pouch src/main.go test` handles each path in order.

## Why it exists

Creating paths often means switching between commands:

```sh
mkdir -p notes
mkdir -p src && touch src/main.go
```

`pouch` reduces that to one command with one detection rule.

## How auto mode works

Auto mode looks only at the final path segment.

Examples:

- `sample` -> directory
- `sample.ts` -> file
- `sample/temp.ts` -> file
- `.env` -> file

`pouch` keeps this rule intentionally small. It does not try to infer intent from well-known filenames, MIME types, or trailing slashes.

## When to use `--mode`

Some names are ambiguous under the auto rule:

- `Dockerfile` becomes a directory in auto mode.
- `Makefile` becomes a directory in auto mode.
- `dir.with.dot` becomes a file in auto mode.

Use `--mode` when you want a different result:

```sh
pouch --mode file Dockerfile
pouch --mode dir dir.with.dot
```

## Behavior

### File paths

If `pouch` treats a path as a file, it:

1. Creates missing parent directories.
2. Creates the file if it does not exist.
3. Leaves the file unchanged if it already exists.

### Directory paths

If `pouch` treats a path as a directory, it:

1. Creates the directory with `mkdir -p` semantics.
2. Succeeds if the directory already exists.

## CLI

Basic usage:

```sh
pouch [flags] PATH...
```

Flags:

- `-m, --mode <auto|file|dir>`: force file or directory mode.
- `-n, --dry-run`: print planned actions without changing the filesystem.
- `-v, --verbose`: print each action in input order.
- `-h, --help`: show help.
- `--version`: show version.

## Exit behavior

- Exit `0` on full success.
- Exit non-zero on the first error.
- Write errors to stderr.

## Scope

`pouch` is intentionally narrow.

It includes:

- path creation from CLI arguments
- simple auto detection
- explicit mode overrides
- a reusable Go package for the CLI and tests

It does not include:

- template generation
- file content generation
- project scaffolding
- config files in the first release
- interactive prompts in the first release

## Project docs

- Repository overview: `README.md`
- Japanese overview: `README-ja.md`
- Agent guidance: `AGENTS.md`
- Japanese agent guidance: `AGENTS-ja.md`
