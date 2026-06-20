package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSilentOnSuccess(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "sample")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{target}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunWarnsOnFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "Dockerfile")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--mode", "file", target}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), warnPrefix) {
		t.Fatalf("stderr = %q, want %q prefix", stderr.String(), warnPrefix)
	}
}

func TestRunVerboseUsesInfoPrefix(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "sample")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--verbose", target}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), infoPrefix) {
		t.Fatalf("stdout = %q, want %q prefix", stdout.String(), infoPrefix)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunVerbosePrintsActionsInInputOrder(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second.go")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{first, second, "--verbose"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	output := stdout.String()
	firstIndex := strings.Index(output, first)
	secondIndex := strings.Index(output, second)
	if firstIndex < 0 || secondIndex < 0 {
		t.Fatalf("stdout = %q, want both paths", output)
	}
	if firstIndex > secondIndex {
		t.Fatalf("stdout = %q, want paths in input order", output)
	}
}

func TestRunVersionShortFlag(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-v"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if strings.TrimSpace(stdout.String()) != version {
		t.Fatalf("stdout = %q, want %q", stdout.String(), version)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDryRunTrailingFlagDoesNotCreatePaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	fileTarget := filepath.Join(root, "src", "main.go")
	dirTarget := filepath.Join(root, "test")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{fileTarget, dirTarget, "--dry-run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if _, err := os.Stat(fileTarget); !os.IsNotExist(err) {
		t.Fatalf("fileTarget exists after dry run, err = %v", err)
	}
	if _, err := os.Stat(dirTarget); !os.IsNotExist(err) {
		t.Fatalf("dirTarget exists after dry run, err = %v", err)
	}
}

func TestRunStrictFailsForExistingPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "sample")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{target, "--strict"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "already exists") {
		t.Fatalf("stderr = %q, want existing-path error", stderr.String())
	}
}

func TestRunFileFlagCreatesAmbiguousFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "Dockerfile")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{target, "--file"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", target)
	}
}

func TestRunFileFlagAppliesToAllPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dockerfile := filepath.Join(root, "Dockerfile")
	test := filepath.Join(root, "test")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{dockerfile, test, "--file"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertFile(t, dockerfile)
	assertFile(t, test)
}

func TestRunDirFlagCreatesAmbiguousDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "dir.with.dot")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{target, "--dir"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is a file, want directory", target)
	}
}

func TestRunDirFlagAppliesToAllPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := filepath.Join(root, "first.go")
	second := filepath.Join(root, "second")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{first, second, "--dir"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertDir(t, first)
	assertDir(t, second)
}

func TestRunModeFlagAppliesToAllPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second.go")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{first, second, "--mode", "dir"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertDir(t, first)
	assertDir(t, second)
}

func TestRunRejectsModeShortcutConflict(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "Dockerfile")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{target, "--mode", "auto", "--file"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "--mode cannot be used with --file or --dir") {
		t.Fatalf("stderr = %q, want mode shortcut conflict", stderr.String())
	}
}

func TestRunStopsAtFirstError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := filepath.Join(root, "first.go")
	conflict := filepath.Join(root, "conflict")
	later := filepath.Join(root, "later.go")
	if err := os.Mkdir(conflict, 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{first, conflict, later, "--file"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr is empty, want error output")
	}
	assertFile(t, first)
	if _, err := os.Stat(later); !os.IsNotExist(err) {
		t.Fatalf("later path exists after first error, err = %v", err)
	}
}

func assertFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}

func assertDir(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is a file, want directory", path)
	}
}
