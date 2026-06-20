# Agent Guide

## Documentation sync

- Keep `README.md` and `README-ja.md` aligned in meaning.
- Keep `AGENTS.md` and `AGENTS-ja.md` aligned in meaning.
- Update both files in a language pair within the same change unless the user explicitly asks for a temporary mismatch.
- Prefer natural Japanese in `*-ja.md` files rather than line-by-line translation.

## Source of truth

- Treat `README.md` and `README-ja.md` as the source of truth for public behavior, CLI usage, examples, and user-facing scope.
- Use this file for agent workflows, internal design boundaries, repository maintenance notes, and implementation guidance that does not belong in the README.
- If public behavior changes, update the README pair in the same change.
- If internal guidance changes, update the AGENTS pair in the same change.

## README media

- The README demo GIF is generated from `assets/demo.tape`.
- When the demo changes, regenerate `assets/demo.gif` instead of editing the GIF by hand.
- Demo generation requires `vhs`, `ttyd`, `ffmpeg`, `tree`, and a runnable `pouch` binary on `PATH`.
- Standard regeneration flow:
  - Create an empty working directory outside the repository.
  - Ensure `vhs`, `ttyd`, `ffmpeg`, `tree`, and `pouch` are available on `PATH`.
  - Run `vhs /absolute/path/to/repo/assets/demo.tape` from that working directory.
  - Copy the generated `demo.gif` to `assets/demo.gif` in the repository.
  - Verify that `assets/demo.gif` still matches the README examples and visible CLI behavior.
  - Ask the user before deleting the temporary working directory.

## Repository boundaries

- `pouch` is a CLI tool.
- Do not treat the repository root package as a supported external library API.
- Keep the project focused on intuitive file and directory creation from paths.
- Do not expand the project into a general scaffolding tool.

## Internal design boundaries

- Keep path detection deterministic and side-effect free.
- Keep single-path creation as the main behavior unit.
- Keep multi-path orchestration small: preserve input order, stop at the first error, and return the successful results collected before the failure together with the error.
- Prefer explicit mode overrides over smarter auto detection.
- Keep the current separation of concerns:
  - `internal/cli` owns argument parsing and CLI validation.
  - `internal/pouch/detect.go` owns detection logic only.
  - `internal/pouch/create.go` owns filesystem behavior.
  - `internal/pouch/pouch.go` owns orchestration across multiple paths.
- Do not split packages further unless the code clearly demands it.

## Implementation notes

### Path handling

- Use `path/filepath` from the standard library.
- Avoid shelling out to `mkdir`, `touch`, or other OS utilities.
- Prefer standard library behavior over custom path normalization.

### File creation

- Ensure the parent directory exists before creating a file.
- Use `os.OpenFile` with create-capable flags.
- Close files promptly.
- Leave existing files unchanged unless the documented public behavior changes.

### Permissions

- Use explicit defaults in code:
  - Directories: `0o755`
  - Files: `0o644`
- The effective permissions still depend on the user's `umask`.

### Errors

- Wrap errors with path context.
- Prefer direct messages over custom error hierarchies.
- Keep error wording consistent with the CLI behavior documented in the README.

## Testing guidance

- Use `go test ./...` from the repository root as the default test command.
- Do not assume the repository root itself contains a buildable Go package.
- Use table-driven tests when they make cases easier to scan.
- Use `t.TempDir()` for filesystem isolation.
- Cover observable behavior rather than implementation trivia.
- Keep coverage aligned with the current public behavior in the README, including:
  - auto detection
  - explicit mode overrides with `--file`, `--dir`, and `--mode`
  - conflicts between `--mode` and the mode shortcut flags
  - parent directory creation
  - dry-run behavior
  - strict behavior for existing paths
  - trailing flag parsing
  - `--` handling for path-like arguments

## Review criteria

- Does the code still match the public behavior documented in the README?
- Is the CLI-facing structure still smaller and simpler than the implementation behind it?
- Are ambiguous cases documented rather than hidden behind new heuristics?
- Are tests written from observable behavior?
- If README examples change, does the demo tape still match them?

## Change discipline

- Do not add heuristics silently.
- Do not broaden public behavior without updating the README.
- If README examples or demo-facing behavior change, update `assets/demo.tape` and regenerate `assets/demo.gif`.
- Preserve the core identity: a small, predictable path creation tool.
