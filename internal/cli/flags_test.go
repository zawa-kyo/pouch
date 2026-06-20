package cli

import (
	"bytes"
	"flag"
	"testing"

	"github.com/zawa-kyo/pouch/internal/pouch"
)

func TestParseSuccess(t *testing.T) {
	t.Parallel()

	t.Run("parses flags and paths", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"--mode", "file", "--dry-run", "--verbose", "Dockerfile"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(config.Paths) != 1 || config.Paths[0] != "Dockerfile" {
			t.Fatalf("unexpected paths: %+v", config.Paths)
		}
		if config.Options.Mode != pouch.ModeFile || !config.Options.DryRun || !config.Verbose {
			t.Fatalf("unexpected config: %+v", config)
		}
		if stdout.Len() != 0 || stderr.Len() != 0 {
			t.Fatalf("unexpected output: stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	})

	t.Run("parses trailing boolean flag after paths", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"src/main.go", "test", "--dry-run"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(config.Paths) != 2 || config.Paths[0] != "src/main.go" || config.Paths[1] != "test" {
			t.Fatalf("unexpected paths: %+v", config.Paths)
		}
		if !config.Options.DryRun {
			t.Fatalf("config.Options.DryRun = %v, want true", config.Options.DryRun)
		}
	})

	t.Run("parses trailing value flag after path", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"Dockerfile", "--mode", "file"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(config.Paths) != 1 || config.Paths[0] != "Dockerfile" {
			t.Fatalf("unexpected paths: %+v", config.Paths)
		}
		if config.Options.Mode != pouch.ModeFile {
			t.Fatalf("config.Options.Mode = %v, want %v", config.Options.Mode, pouch.ModeFile)
		}
	})

	t.Run("parses file mode flag", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"Dockerfile", "--file"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(config.Paths) != 1 || config.Paths[0] != "Dockerfile" {
			t.Fatalf("unexpected paths: %+v", config.Paths)
		}
		if config.Options.Mode != pouch.ModeFile {
			t.Fatalf("config.Options.Mode = %v, want %v", config.Options.Mode, pouch.ModeFile)
		}
	})

	t.Run("parses dir mode flag", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"dir.with.dot", "--dir"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(config.Paths) != 1 || config.Paths[0] != "dir.with.dot" {
			t.Fatalf("unexpected paths: %+v", config.Paths)
		}
		if config.Options.Mode != pouch.ModeDir {
			t.Fatalf("config.Options.Mode = %v, want %v", config.Options.Mode, pouch.ModeDir)
		}
	})

	t.Run("parses strict flag", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"sample", "--strict"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if !config.Options.Strict {
			t.Fatalf("config.Options.Strict = %v, want true", config.Options.Strict)
		}
	})

	t.Run("treats arguments after double dash as paths", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"--dry-run", "--", "--dry-run"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(config.Paths) != 1 || config.Paths[0] != "--dry-run" {
			t.Fatalf("unexpected paths: %+v", config.Paths)
		}
		if !config.Options.DryRun {
			t.Fatalf("config.Options.DryRun = %v, want true", config.Options.DryRun)
		}
	})

}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	t.Run("requires path", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		_, err := Parse([]string{"--mode", "auto"}, &stdout, &stderr)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects invalid mode", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		_, err := Parse([]string{"--mode", "weird", "sample"}, &stdout, &stderr)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("writes flag parse errors to stderr", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		_, err := Parse([]string{"--unknown"}, &stdout, &stderr)
		if err == nil {
			t.Fatal("expected error")
		}
		if stdout.Len() != 0 {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
		if stderr.Len() == 0 {
			t.Fatal("stderr is empty, want parse error output")
		}
	})

	t.Run("rejects file and dir together", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		_, err := Parse([]string{"sample", "--file", "--dir"}, &stdout, &stderr)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects mode with file shortcut", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		_, err := Parse([]string{"Dockerfile", "--mode", "auto", "--file"}, &stdout, &stderr)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects short mode with dir shortcut", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		_, err := Parse([]string{"dir.with.dot", "-m", "file", "--dir"}, &stdout, &stderr)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestParseSpecialFlags(t *testing.T) {
	t.Parallel()

	t.Run("writes help to stdout", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		_, err := Parse([]string{"--help"}, &stdout, &stderr)
		if err != flag.ErrHelp {
			t.Fatalf("Parse() error = %v, want %v", err, flag.ErrHelp)
		}
		if stdout.Len() == 0 {
			t.Fatal("stdout is empty, want help output")
		}
		if stderr.Len() != 0 {
			t.Fatalf("stderr = %q, want empty", stderr.String())
		}
	})

	t.Run("parses uppercase short verbose flag", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"-V", "sample"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if !config.Verbose {
			t.Fatalf("config.Verbose = %v, want true", config.Verbose)
		}
	})

	t.Run("parses version flag without requiring path", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config, err := Parse([]string{"-v"}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if !config.ShowVersion {
			t.Fatalf("config.ShowVersion = %v, want true", config.ShowVersion)
		}
		if len(config.Paths) != 0 {
			t.Fatalf("config.Paths = %+v, want empty", config.Paths)
		}
	})
}
