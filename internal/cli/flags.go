package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/zawa-kyo/pouch"
)

// Config is the parsed CLI input for one invocation.
type Config struct {
	Paths   []string
	Options pouch.Options
	Verbose bool
}

// Parse converts CLI arguments into a validated Config.
func Parse(args []string, stdout, stderr io.Writer) (Config, error) {
	fs := flag.NewFlagSet("pouch", flag.ContinueOnError)
	fs.SetOutput(stderr)

	mode := fs.String("mode", "auto", "")
	fs.StringVar(mode, "m", "auto", "")

	dryRun := fs.Bool("dry-run", false, "")
	fs.BoolVar(dryRun, "n", false, "")

	verbose := fs.Bool("verbose", false, "")
	fs.BoolVar(verbose, "v", false, "")

	help := fs.Bool("help", false, "")
	fs.BoolVar(help, "h", false, "")

	fs.Usage = func() { writeUsage(stderr) }

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if *help {
		writeUsage(stdout)
		return Config{}, flag.ErrHelp
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

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: pouch [flags] PATH...")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -m, --mode auto|file|dir")
	fmt.Fprintln(w, "      Force file or directory mode.")
	fmt.Fprintln(w, "  -n, --dry-run")
	fmt.Fprintln(w, "      Print planned actions without changing the filesystem.")
	fmt.Fprintln(w, "  -v, --verbose")
	fmt.Fprintln(w, "      Print each action in input order.")
	fmt.Fprintln(w, "  -h, --help")
	fmt.Fprintln(w, "      Show help.")
	fmt.Fprintln(w, "      --version")
	fmt.Fprintln(w, "      Show version.")
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
