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

	mode := registerStringFlag(fs, "mode", "m", "auto")
	fileMode := fs.Bool("file", false, "")
	dirMode := fs.Bool("dir", false, "")
	dryRun := registerBoolFlag(fs, "dry-run", "n")
	strict := registerBoolFlag(fs, "strict", "s")
	verbose := registerBoolFlag(fs, "verbose", "V")
	help := registerBoolFlag(fs, "help", "h")
	showVersion := registerBoolFlag(fs, "version", "v")

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

func registerBoolFlag(fs *flag.FlagSet, name, short string) *bool {
	value := fs.Bool(name, false, "")
	fs.BoolVar(value, short, false, "")
	return value
}

func registerStringFlag(fs *flag.FlagSet, name, short, defaultValue string) *string {
	value := fs.String(name, defaultValue, "")
	fs.StringVar(value, short, defaultValue, "")
	return value
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
