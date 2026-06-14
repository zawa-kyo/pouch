# pouch

`pouch` is a simple, path-aware `touch` command.

It creates either a directory or a file from a single path argument:

- If the final path segment contains a dot (`.`), `pouch` treats it as a file path.
- Otherwise, `pouch` treats it as a directory path.

When creating a file, `pouch` also creates any missing parent directories automatically.

## Why

Creating paths usually requires switching between commands:

```sh
mkdir -p sample
mkdir -p "$(dirname sample/temp.ts)" && touch sample/temp.ts
touch sample.ts
```

`pouch` collapses that into one mental model:

```sh
pouch sample
pouch sample.ts
pouch sample/temp.ts
```

The command is intentionally small and opinionated.

## Behavior

### Auto detection

`pouch` inspects the final path segment.

Examples:

- `sample` -> directory
- `sample.ts` -> file
- `sample/temp.ts` -> file
- `.env` -> file

### File path behavior

If a path is treated as a file:

1. Create parent directories with `mkdir -p` semantics.
2. Create the file if it does not exist.
3. If the file already exists, update its modification time like `touch`.

### Directory path behavior

If a path is treated as a directory:

1. Create the directory with `mkdir -p` semantics.
2. Do nothing if it already exists.

## Non-goals

`pouch` does not try to solve every path-creation case.

Out of scope for the default auto mode:

- Detecting extensionless files automatically
- Inferring file type from content or MIME
- Generating templates or starter code
- Language-aware scaffolding
- Project bootstrapping

## Known limitations

The auto-detection rule is intentionally simple, so it has tradeoffs.

Examples:

- `Dockerfile` is treated as a directory in auto mode
- `Makefile` is treated as a directory in auto mode
- `dir.with.dot` is treated as a file in auto mode

To handle those cases, `pouch` should support explicit mode overrides.

## CLI

### Basic usage

```sh
pouch PATH...
```

Examples:

```sh
pouch sample
pouch sample.ts
pouch sample/temp.ts
pouch src/index.ts test/index.test.ts docs
```

### Proposed flags

```sh
pouch [flags] PATH...
```

Flags:

- `-m, --mode <auto|file|dir>`
  Force path interpretation instead of using auto detection.
- `-n, --dry-run`
  Print planned actions without changing the filesystem.
- `-v, --verbose`
  Print each action as it happens.
- `--no-touch`
  When the file already exists, do not update its modification time.
- `-h, --help`
  Show help.
- `--version`
  Show version.

### Recommended behavior for explicit mode

Examples:

```sh
pouch --mode file Dockerfile
pouch --mode dir dir.with.dot
```

## Exit behavior

Recommended initial behavior:

- Exit `0` if all paths are processed successfully.
- Exit non-zero on the first error.
- Print a clear error message to stderr.

Future expansion could add:

- `--continue-on-error`

That should not be part of the first release unless there is a clear need.

## Examples

### Create a directory

```sh
pouch sample
```

Equivalent intent:

```sh
mkdir -p sample
```

### Create a file in the current directory

```sh
pouch sample.ts
```

Equivalent intent:

```sh
touch sample.ts
```

### Create a file and all missing parents

```sh
pouch sample/temp.ts
```

Equivalent intent:

```sh
mkdir -p sample
touch sample/temp.ts
```

### Create an extensionless file explicitly

```sh
pouch --mode file Dockerfile
```

### Create a dot-containing directory explicitly

```sh
pouch --mode dir dir.with.dot
```

## Design principles

`pouch` should stay small.

Core principles:

- One job: create a path
- Prefer predictable rules over clever inference
- Keep auto mode simple and documented
- Offer explicit overrides for ambiguous cases
- Match standard filesystem behavior as closely as possible
- Be useful as both a CLI and a Go package

## Positioning

Short description:

> `pouch` is a path-aware `touch` command.

Slightly longer description:

> `pouch` creates directories or files from path-like arguments using a simple, explicit detection rule.
