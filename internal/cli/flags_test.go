package cli

import (
	"bytes"
	"flag"
	"testing"

	"github.com/zawa-kyo/pouch/internal/pouch"
)

func TestParseSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		verify func(*testing.T, Config, string, string)
	}{
		{
			name: "parses flags and paths",
			args: []string{"--mode", "file", "--dry-run", "--verbose", "Dockerfile"},
			verify: func(t *testing.T, config Config, stdout, stderr string) {
				assertPaths(t, config.Paths, "Dockerfile")
				assertMode(t, config, pouch.ModeFile)
				assertDryRun(t, config, true)
				assertVerbose(t, config, true)
				assertOutputEmpty(t, stdout, stderr)
			},
		},
		{
			name: "parses trailing boolean flag after paths",
			args: []string{"src/main.go", "test", "--dry-run"},
			verify: func(t *testing.T, config Config, _, _ string) {
				assertPaths(t, config.Paths, "src/main.go", "test")
				assertDryRun(t, config, true)
			},
		},
		{
			name: "parses trailing value flag after path",
			args: []string{"Dockerfile", "--mode", "file"},
			verify: func(t *testing.T, config Config, _, _ string) {
				assertPaths(t, config.Paths, "Dockerfile")
				assertMode(t, config, pouch.ModeFile)
			},
		},
		{
			name: "parses trailing inline mode flag after path",
			args: []string{"Dockerfile", "--mode=file"},
			verify: func(t *testing.T, config Config, _, _ string) {
				assertPaths(t, config.Paths, "Dockerfile")
				assertMode(t, config, pouch.ModeFile)
			},
		},
		{
			name: "parses file mode flag",
			args: []string{"Dockerfile", "--file"},
			verify: func(t *testing.T, config Config, _, _ string) {
				assertPaths(t, config.Paths, "Dockerfile")
				assertMode(t, config, pouch.ModeFile)
			},
		},
		{
			name: "parses dir mode flag",
			args: []string{"dir.with.dot", "--dir"},
			verify: func(t *testing.T, config Config, _, _ string) {
				assertPaths(t, config.Paths, "dir.with.dot")
				assertMode(t, config, pouch.ModeDir)
			},
		},
		{
			name: "parses strict flag",
			args: []string{"sample", "--strict"},
			verify: func(t *testing.T, config Config, _, _ string) {
				assertStrict(t, config, true)
			},
		},
		{
			name: "treats arguments after double dash as paths",
			args: []string{"--dry-run", "--", "--dry-run"},
			verify: func(t *testing.T, config Config, _, _ string) {
				assertPaths(t, config.Paths, "--dry-run")
				assertDryRun(t, config, true)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config, stdout, stderr := parseForTest(t, tt.args)
			tt.verify(t, config, stdout, stderr)
		})
	}
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

func parseForTest(t *testing.T, args []string) (Config, string, string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	config, err := Parse(args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	return config, stdout.String(), stderr.String()
}

func assertPaths(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(paths) = %d, want %d (%+v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("paths[%d] = %q, want %q (%+v)", i, got[i], want[i], got)
		}
	}
}

func assertMode(t *testing.T, config Config, want pouch.Mode) {
	t.Helper()
	if config.Options.Mode != want {
		t.Fatalf("config.Options.Mode = %v, want %v", config.Options.Mode, want)
	}
}

func assertDryRun(t *testing.T, config Config, want bool) {
	t.Helper()
	if config.Options.DryRun != want {
		t.Fatalf("config.Options.DryRun = %v, want %v", config.Options.DryRun, want)
	}
}

func assertStrict(t *testing.T, config Config, want bool) {
	t.Helper()
	if config.Options.Strict != want {
		t.Fatalf("config.Options.Strict = %v, want %v", config.Options.Strict, want)
	}
}

func assertVerbose(t *testing.T, config Config, want bool) {
	t.Helper()
	if config.Verbose != want {
		t.Fatalf("config.Verbose = %v, want %v", config.Verbose, want)
	}
}

func assertOutputEmpty(t *testing.T, stdout, stderr string) {
	t.Helper()
	if stdout != "" || stderr != "" {
		t.Fatalf("unexpected output: stdout=%q stderr=%q", stdout, stderr)
	}
}
