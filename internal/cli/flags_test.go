package cli

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/zawa-kyo/pouch/internal/pouch"
)

func TestParseBuildsConfigFromInterspersedFlagsAndPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want Config
	}{
		{
			name: "path only uses auto mode",
			args: []string{"sample"},
			want: Config{
				Paths:   []string{"sample"},
				Options: pouch.Options{Mode: pouch.ModeAuto},
			},
		},
		{
			name: "flags before paths",
			args: []string{"--mode", "file", "--dry-run", "--verbose", "Dockerfile"},
			want: Config{
				Paths:   []string{"Dockerfile"},
				Options: pouch.Options{Mode: pouch.ModeFile, DryRun: true},
				Verbose: true,
			},
		},
		{
			name: "boolean flag after paths",
			args: []string{"src/main.go", "test", "--dry-run"},
			want: Config{
				Paths:   []string{"src/main.go", "test"},
				Options: pouch.Options{Mode: pouch.ModeAuto, DryRun: true},
			},
		},
		{
			name: "value flag after path",
			args: []string{"Dockerfile", "--mode", "file"},
			want: Config{
				Paths:   []string{"Dockerfile"},
				Options: pouch.Options{Mode: pouch.ModeFile},
			},
		},
		{
			name: "long inline mode flag after path",
			args: []string{"Dockerfile", "--mode=file"},
			want: Config{
				Paths:   []string{"Dockerfile"},
				Options: pouch.Options{Mode: pouch.ModeFile},
			},
		},
		{
			name: "short inline mode flag after path",
			args: []string{"Dockerfile", "-m=file"},
			want: Config{
				Paths:   []string{"Dockerfile"},
				Options: pouch.Options{Mode: pouch.ModeFile},
			},
		},
		{
			name: "file shortcut applies to all paths",
			args: []string{"Dockerfile", "test", "--file"},
			want: Config{
				Paths:   []string{"Dockerfile", "test"},
				Options: pouch.Options{Mode: pouch.ModeFile},
			},
		},
		{
			name: "directory shortcut applies to all paths",
			args: []string{"dir.with.dot", "--dir"},
			want: Config{
				Paths:   []string{"dir.with.dot"},
				Options: pouch.Options{Mode: pouch.ModeDir},
			},
		},
		{
			name: "short flags set dry-run strict and verbose",
			args: []string{"sample", "-n", "-s", "-V"},
			want: Config{
				Paths:   []string{"sample"},
				Options: pouch.Options{Mode: pouch.ModeAuto, DryRun: true, Strict: true},
				Verbose: true,
			},
		},
		{
			name: "double dash keeps later flag-shaped values as paths",
			args: []string{"--dry-run", "--", "--dry-run", "-n"},
			want: Config{
				Paths:   []string{"--dry-run", "-n"},
				Options: pouch.Options{Mode: pouch.ModeAuto, DryRun: true},
			},
		},
		{
			name: "single dash is a path value",
			args: []string{"-"},
			want: Config{
				Paths:   []string{"-"},
				Options: pouch.Options{Mode: pouch.ModeAuto},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, stdout, stderr, err := parseForTest(tt.args)

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			assertConfig(t, config, tt.want)
			assertOutput(t, stdout, "")
			assertOutput(t, stderr, "")
		})
	}
}

func TestParseRejectsInvalidInvocations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantErr    string
		wantStderr string
	}{
		{
			name:    "path is required unless a special flag handles the invocation",
			args:    []string{"--mode", "auto"},
			wantErr: "PATH... is required",
		},
		{
			name:    "mode value must be one of the documented values",
			args:    []string{"sample", "--mode", "weird"},
			wantErr: `invalid mode "weird": must be auto, file, or dir`,
		},
		{
			name:    "mode flag requires a value",
			args:    []string{"sample", "--mode"},
			wantErr: "flag needs an argument: --mode",
		},
		{
			name:    "file and directory shortcuts conflict",
			args:    []string{"sample", "--file", "--dir"},
			wantErr: "--file and --dir cannot be used together",
		},
		{
			name:    "explicit mode conflicts with file shortcut",
			args:    []string{"Dockerfile", "--mode", "auto", "--file"},
			wantErr: "--mode cannot be used with --file or --dir",
		},
		{
			name:    "explicit mode conflicts with directory shortcut",
			args:    []string{"dir.with.dot", "-m", "file", "--dir"},
			wantErr: "--mode cannot be used with --file or --dir",
		},
		{
			name:       "unknown flag is reported by the flag parser",
			args:       []string{"--unknown"},
			wantErr:    "flag provided but not defined",
			wantStderr: "flag provided but not defined",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, stdout, stderr, err := parseForTest(tt.args)

			assertErrorContains(t, err, tt.wantErr)
			assertOutput(t, stdout, "")
			if tt.wantStderr != "" {
				assertStringContains(t, stderr, tt.wantStderr)
			} else {
				assertOutput(t, stderr, "")
			}
		})
	}
}

func TestParseHandlesSpecialFlagsBeforePathValidation(t *testing.T) {
	t.Parallel()

	t.Run("help writes usage to stdout", func(t *testing.T) {
		t.Parallel()

		config, stdout, stderr, err := parseForTest([]string{"--help"})

		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("Parse() error = %v, want %v", err, flag.ErrHelp)
		}
		assertConfig(t, config, Config{})
		assertStringContains(t, stdout, "Usage: pouch [flags] PATH...")
		assertOutput(t, stderr, "")
	})

	t.Run("version does not require paths", func(t *testing.T) {
		t.Parallel()

		config, stdout, stderr, err := parseForTest([]string{"-v"})

		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		assertConfig(t, config, Config{ShowVersion: true})
		assertOutput(t, stdout, "")
		assertOutput(t, stderr, "")
	})
}

func parseForTest(args []string) (Config, string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	config, err := Parse(args, &stdout, &stderr)

	return config, stdout.String(), stderr.String(), err
}

func assertConfig(t *testing.T, got, want Config) {
	t.Helper()
	assertPaths(t, got.Paths, want.Paths...)
	if got.Options != want.Options {
		t.Fatalf("Options = %+v, want %+v", got.Options, want.Options)
	}
	if got.Verbose != want.Verbose {
		t.Fatalf("Verbose = %v, want %v", got.Verbose, want.Verbose)
	}
	if got.ShowVersion != want.ShowVersion {
		t.Fatalf("ShowVersion = %v, want %v", got.ShowVersion, want.ShowVersion)
	}
}

func assertPaths(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(paths) = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("paths[%d] = %q, want %q: %+v", i, got[i], want[i], got)
		}
	}
}

func assertOutput(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want message containing %q", want)
	}
	assertStringContains(t, err.Error(), want)
}

func assertStringContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("%q does not contain %q", got, want)
	}
}
