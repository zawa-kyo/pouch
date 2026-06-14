package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/zawa-kyo/pouch/pkg/pouch"
)

type Config struct {
	Paths   []string
	Options pouch.Options
	Verbose bool
}

func Parse(args []string, stdout io.Writer) (Config, error) {
	fs := flag.NewFlagSet("pouch", flag.ContinueOnError)
	fs.SetOutput(stdout)

	mode := fs.String("mode", "auto", "")
	fs.StringVar(mode, "m", "auto", "")

	dryRun := fs.Bool("dry-run", false, "")
	fs.BoolVar(dryRun, "n", false, "")

	verbose := fs.Bool("verbose", false, "")
	fs.BoolVar(verbose, "v", false, "")

	fs.Usage = func() {
		fmt.Fprintln(stdout, "Usage: pouch [flags] PATH...")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Flags:")
		fmt.Fprintln(stdout, "  -m, --mode auto|file|dir")
		fmt.Fprintln(stdout, "      Force file or directory mode.")
		fmt.Fprintln(stdout, "  -n, --dry-run")
		fmt.Fprintln(stdout, "      Print planned actions without changing the filesystem.")
		fmt.Fprintln(stdout, "  -v, --verbose")
		fmt.Fprintln(stdout, "      Print each action in input order.")
		fmt.Fprintln(stdout, "  -h, --help")
		fmt.Fprintln(stdout, "      Show help.")
		fmt.Fprintln(stdout, "      --version")
		fmt.Fprintln(stdout, "      Show version.")
	}

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	paths := fs.Args()
	if len(paths) == 0 {
		return Config{}, errors.New("PATH... is required")
	}

	parsedMode, err := parseMode(*mode)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Paths: paths,
		Options: pouch.Options{
			Mode:   parsedMode,
			DryRun: *dryRun,
		},
		Verbose: *verbose,
	}, nil
}

func parseMode(value string) (pouch.Mode, error) {
	switch strings.ToLower(value) {
	case "auto":
		return pouch.ModeAuto, nil
	case "file":
		return pouch.ModeFile, nil
	case "dir":
		return pouch.ModeDir, nil
	default:
		return pouch.ModeAuto, fmt.Errorf("invalid mode %q: must be auto, file, or dir", value)
	}
}
