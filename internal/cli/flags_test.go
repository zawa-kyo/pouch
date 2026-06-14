package cli

import (
	"bytes"
	"testing"

	"github.com/zawa-kyo/pouch/pkg/pouch"
)

func TestParse(t *testing.T) {
	t.Parallel()

	t.Run("parses flags and paths", func(t *testing.T) {
		t.Parallel()
		var out bytes.Buffer

		config, err := Parse([]string{"--mode", "file", "--dry-run", "--verbose", "Dockerfile"}, &out)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if len(config.Paths) != 1 || config.Paths[0] != "Dockerfile" {
			t.Fatalf("unexpected paths: %+v", config.Paths)
		}
		if config.Options.Mode != pouch.ModeFile || !config.Options.DryRun || !config.Verbose {
			t.Fatalf("unexpected config: %+v", config)
		}
	})

	t.Run("requires path", func(t *testing.T) {
		t.Parallel()
		var out bytes.Buffer

		_, err := Parse([]string{"--mode", "auto"}, &out)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects invalid mode", func(t *testing.T) {
		t.Parallel()
		var out bytes.Buffer

		_, err := Parse([]string{"--mode", "weird", "sample"}, &out)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
