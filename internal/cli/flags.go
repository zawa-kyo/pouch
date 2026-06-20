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

type flagValues struct {
	mode        *string
	fileMode    *bool
	dirMode     *bool
	dryRun      *bool
	strict      *bool
	verbose     *bool
	help        *bool
	showVersion *bool
}

type splitResult struct {
	flagArgs     []string
	paths        []string
	explicitMode bool
}

type flagKind int

const (
	flagKindUnknown flagKind = iota
	flagKindBool
	flagKindMode
	flagKindInlineMode
)

// Parse converts CLI arguments into a validated Config.
func Parse(args []string, stdout, stderr io.Writer) (Config, error) {
	fs, values := newFlagSet(stdout, stderr)
	split, err := splitArgs(args)
	if err != nil {
		return Config{}, err
	}
	if err := fs.Parse(split.flagArgs); err != nil {
		return Config{}, err
	}
	if handled, config := handleSpecialFlags(values, stdout); handled {
		return config, flag.ErrHelp
	}
	if *values.showVersion {
		return Config{ShowVersion: true}, nil
	}
	if err := validateFlags(values, split.explicitMode); err != nil {
		return Config{}, err
	}
	if err := validatePaths(split.paths); err != nil {
		return Config{}, err
	}
	parsedMode, err := resolveMode(values, split.explicitMode)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Paths: split.paths,
		Options: pouch.Options{
			Mode:   parsedMode,
			DryRun: *values.dryRun,
			Strict: *values.strict,
		},
		Verbose: *values.verbose,
	}, nil
}

func newFlagSet(stdout, stderr io.Writer) (*flag.FlagSet, flagValues) {
	fs := flag.NewFlagSet("pouch", flag.ContinueOnError)
	fs.SetOutput(stderr)

	mode := fs.String("mode", "auto", "")
	fs.StringVar(mode, "m", "auto", "")

	fileMode := fs.Bool("file", false, "")
	dirMode := fs.Bool("dir", false, "")

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

	return fs, flagValues{
		mode:        mode,
		fileMode:    fileMode,
		dirMode:     dirMode,
		dryRun:      dryRun,
		strict:      strict,
		verbose:     verbose,
		help:        help,
		showVersion: showVersion,
	}
}

func handleSpecialFlags(values flagValues, stdout io.Writer) (bool, Config) {
	if !*values.help {
		return false, Config{}
	}

	writeUsage(stdout)
	return true, Config{}
}

func validateFlags(values flagValues, explicitMode bool) error {
	if *values.fileMode && *values.dirMode {
		return errors.New("--file and --dir cannot be used together")
	}
	if explicitMode && (*values.fileMode || *values.dirMode) {
		return errors.New("--mode cannot be used with --file or --dir")
	}
	return nil
}

func validatePaths(paths []string) error {
	if len(paths) == 0 {
		return errors.New("PATH... is required")
	}
	return nil
}

func resolveMode(values flagValues, explicitMode bool) (pouch.Mode, error) {
	parsedMode, err := parseMode(*values.mode)
	if err != nil {
		return parsedMode, err
	}
	if !explicitMode {
		if *values.fileMode {
			return pouch.ModeFile, nil
		}
		if *values.dirMode {
			return pouch.ModeDir, nil
		}
	}
	return parsedMode, nil
}

func splitArgs(args []string) (splitResult, error) {
	result := splitResult{
		flagArgs: make([]string, 0, len(args)),
		paths:    make([]string, 0, len(args)),
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			result.paths = append(result.paths, args[i+1:]...)
			return result, nil
		}

		if !strings.HasPrefix(arg, "-") || arg == "-" {
			result.paths = append(result.paths, arg)
			continue
		}

		switch classifyFlag(arg) {
		case flagKindBool:
			result.flagArgs = append(result.flagArgs, arg)
		case flagKindInlineMode:
			result.flagArgs = append(result.flagArgs, arg)
			result.explicitMode = true
		case flagKindMode:
			if i+1 >= len(args) {
				return splitResult{}, fmt.Errorf("flag needs an argument: %s", arg)
			}
			result.flagArgs = append(result.flagArgs, arg, args[i+1])
			result.explicitMode = true
			i++
		default:
			result.flagArgs = append(result.flagArgs, arg)
		}
	}

	return result, nil
}

func classifyFlag(arg string) flagKind {
	switch arg {
	case "--file", "--dir", "--dry-run", "-n", "--strict", "-s", "--verbose", "-V", "--help", "-h", "--version", "-v":
		return flagKindBool
	case "--mode", "-m":
		return flagKindMode
	default:
		if strings.HasPrefix(arg, "--mode=") || strings.HasPrefix(arg, "-m=") {
			return flagKindInlineMode
		}
		return flagKindUnknown
	}
}

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: pouch [flags] PATH...")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -h, --help")
	fmt.Fprintln(w, "      Show help.")
	fmt.Fprintln(w, "  --file")
	fmt.Fprintln(w, "      Use file mode.")
	fmt.Fprintln(w, "  --dir")
	fmt.Fprintln(w, "      Use directory mode.")
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
