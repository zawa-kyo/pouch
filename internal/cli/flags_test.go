package cli

import (
	"bytes"
	"flag"
	"testing"

	"github.com/zawa-kyo/pouch"
)

func TestParse(t *testing.T) {
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
}
