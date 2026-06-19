package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/zawa-kyo/pouch/internal/pouch"
)

// Config is the parsed CLI input for one invocation.
type Config struct {
	Paths       []string
	Options     pouch.Options
	Verbose     bool
	ShowVersion bool
}

// Parse converts CLI arguments into a validated Config.
func Parse(args []string, stdout, stderr io.Writer) (Config, error) {
	fs := flag.NewFlagSet("pouch", flag.ContinueOnError)
	fs.SetOutput(stderr)

	mode := fs.String("mode", "auto", "")
	fs.StringVar(mode, "m", "auto", "")

	dryRun := fs.Bool("dry-run", false, "")
	fs.BoolVar(dryRun, "n", false, "")

	strict := fs.Bool("strict", false, "")
	fs.BoolVar(strict, "s", false, "")

	verbose := fs.Bool("verbose", false, "")
	fs.BoolVar(verbose, "V", false, "")

	help := fs.Bool("help", false, "")
	fs.BoolVar(help, "h", false, "")

	showVersion := fs.Bool("version", false, "")
	fs.BoolVar(showVersion, "v", false, "")

	fs.Usage = func() { writeUsage(stderr) }

	flagArgs, paths, err := splitArgs(args)
	if err != nil {
		return Config{}, err
	}

	if err := fs.Parse(flagArgs); err != nil {
		return Config{}, err
	}

	if *help {
		writeUsage(stdout)
		return Config{}, flag.ErrHelp
	}

	if *showVersion {
		return Config{ShowVersion: true}, nil
	}

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
			Strict: *strict,
		},
		Verbose: *verbose,
	}, nil
}

func splitArgs(args []string) ([]string, []string, error) {
	flagArgs := make([]string, 0, len(args))
	paths := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			paths = append(paths, args[i+1:]...)
			return flagArgs, paths, nil
		}

		if !strings.HasPrefix(arg, "-") || arg == "-" {
			paths = append(paths, arg)
			continue
		}

		switch {
		case isBoolFlag(arg), isInlineModeFlag(arg):
			flagArgs = append(flagArgs, arg)
		case isModeFlag(arg):
			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("flag needs an argument: %s", arg)
			}
			flagArgs = append(flagArgs, arg, args[i+1])
			i++
		default:
			flagArgs = append(flagArgs, arg)
		}
	}

	return flagArgs, paths, nil
}

func isBoolFlag(arg string) bool {
	switch arg {
	case "--dry-run", "-n", "--strict", "-s", "--verbose", "-V", "--help", "-h", "--version", "-v":
		return true
	default:
		return false
	}
}

func isModeFlag(arg string) bool {
	return arg == "--mode" || arg == "-m"
}

func isInlineModeFlag(arg string) bool {
	return strings.HasPrefix(arg, "--mode=") || strings.HasPrefix(arg, "-m=")
}

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: pouch [flags] PATH...")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -h, --help")
	fmt.Fprintln(w, "      Show help.")
	fmt.Fprintln(w, "  -m, --mode auto|file|dir")
	fmt.Fprintln(w, "      Force file or directory mode.")
	fmt.Fprintln(w, "  -n, --dry-run")
	fmt.Fprintln(w, "      Print planned actions without changing the filesystem.")
	fmt.Fprintln(w, "  -s, --strict")
	fmt.Fprintln(w, "      Fail if a target already exists.")
	fmt.Fprintln(w, "  -v, --version")
	fmt.Fprintln(w, "      Show version.")
	fmt.Fprintln(w, "  -V, --verbose")
	fmt.Fprintln(w, "      Print each action in input order.")
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
