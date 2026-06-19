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
