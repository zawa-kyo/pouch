package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreatesPathsFromCLIArguments(t *testing.T) {
	t.Parallel()

	t.Run("auto mode is silent on success", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		target := filepath.Join(root, "sample")

		result := runForTest([]string{target})

		assertExitCode(t, result.code, 0)
		assertOutput(t, result.stdout, "")
		assertOutput(t, result.stderr, "")
		assertDir(t, target)
	})

	t.Run("file shortcut applies to every path", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		dockerfile := filepath.Join(root, "Dockerfile")
		test := filepath.Join(root, "test")

		result := runForTest([]string{dockerfile, test, "--file"})

		assertExitCode(t, result.code, 0)
		assertOutput(t, result.stdout, "")
		assertOutput(t, result.stderr, "")
		assertFile(t, dockerfile)
		assertFile(t, test)
	})

	t.Run("directory shortcut applies to every path", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		first := filepath.Join(root, "first.go")
		second := filepath.Join(root, "second")

		result := runForTest([]string{first, second, "--dir"})

		assertExitCode(t, result.code, 0)
		assertOutput(t, result.stdout, "")
		assertOutput(t, result.stderr, "")
		assertDir(t, first)
		assertDir(t, second)
	})

	t.Run("dry-run leaves every planned path missing", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		fileTarget := filepath.Join(root, "src", "main.go")
		dirTarget := filepath.Join(root, "test")

		result := runForTest([]string{fileTarget, dirTarget, "--dry-run"})

		assertExitCode(t, result.code, 0)
		assertOutput(t, result.stdout, "")
		assertOutput(t, result.stderr, "")
		assertMissing(t, filepath.Dir(fileTarget))
		assertMissing(t, fileTarget)
		assertMissing(t, dirTarget)
	})
}

func TestRunDoubleDashAllowsFlagShapedPathArguments(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	result := runForTest([]string{"--", "--dry-run"})

	assertExitCode(t, result.code, 0)
	assertOutput(t, result.stdout, "")
	assertOutput(t, result.stderr, "")
	assertDir(t, filepath.Join(root, "--dry-run"))
}

func TestRunWritesRequestedCommandOutput(t *testing.T) {
	t.Parallel()

	t.Run("verbose output lists actions in input order", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		first := filepath.Join(root, "first")
		second := filepath.Join(root, "second.go")

		result := runForTest([]string{first, second, "--verbose"})

		assertExitCode(t, result.code, 0)
		assertLines(t, result.stdout,
			formatResultLine("create dir", first),
			formatResultLine("create file", second),
		)
		assertOutput(t, result.stderr, "")
	})

	t.Run("version prints only the version", func(t *testing.T) {
		t.Parallel()

		result := runForTest([]string{"-v"})

		assertExitCode(t, result.code, 0)
		assertOutput(t, strings.TrimSpace(result.stdout), version)
		assertOutput(t, result.stderr, "")
	})

	t.Run("help prints usage to stdout", func(t *testing.T) {
		t.Parallel()

		result := runForTest([]string{"--help"})

		assertExitCode(t, result.code, 0)
		assertStringContains(t, result.stdout, "Usage: pouch [flags] PATH...")
		assertOutput(t, result.stderr, "")
	})
}

func TestRunReportsFailuresAndStopsAtFirstError(t *testing.T) {
	t.Parallel()

	t.Run("parse errors exit non-zero and write a warning", func(t *testing.T) {
		t.Parallel()

		result := runForTest([]string{"target", "--mode", "auto", "--file"})

		assertExitCode(t, result.code, 1)
		assertOutput(t, result.stdout, "")
		assertWarningContains(t, result.stderr, "--mode cannot be used with --file or --dir")
	})

	t.Run("creation errors exit non-zero and leave later paths untouched", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		first := filepath.Join(root, "first.go")
		conflict := filepath.Join(root, "conflict")
		later := filepath.Join(root, "later.go")
		mkdir(t, conflict)

		result := runForTest([]string{first, conflict, later, "--file"})

		assertExitCode(t, result.code, 1)
		assertOutput(t, result.stdout, "")
		assertWarningContains(t, result.stderr, "path exists as directory")
		assertFile(t, first)
		assertMissing(t, later)
	})

	t.Run("verbose failure prints completed actions before the warning", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		first := filepath.Join(root, "first.go")
		conflict := filepath.Join(root, "conflict")
		mkdir(t, conflict)

		result := runForTest([]string{first, conflict, "--file", "--verbose"})

		assertExitCode(t, result.code, 1)
		assertLines(t, result.stdout, formatResultLine("create file", first))
		assertWarningContains(t, result.stderr, "path exists as directory")
	})
}

type runResult struct {
	code   int
	stdout string
	stderr string
}

func runForTest(args []string) runResult {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(args, &stdout, &stderr)

	return runResult{
		code:   code,
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", path, err)
	}
}

func assertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("exit code = %d, want %d", got, want)
	}
}

func assertFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%q is a directory, want file", path)
	}
}

func assertDir(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is a file, want directory", path)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("os.Stat(%q) error = %v, want missing path", path, err)
	}
}

func assertOutput(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func assertLines(t *testing.T, got string, want ...string) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	if len(lines) != len(want) {
		t.Fatalf("lines = %+v, want %+v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("lines[%d] = %q, want %q", i, lines[i], want[i])
		}
	}
}

func assertWarningContains(t *testing.T, got, want string) {
	t.Helper()
	assertStringContains(t, got, warnPrefix)
	assertStringContains(t, got, want)
}

func assertStringContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("%q does not contain %q", got, want)
	}
}

func formatResultLine(action, path string) string {
	return infoPrefix + " " + action + " " + path
}
