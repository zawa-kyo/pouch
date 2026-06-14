package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zawa-kyo/pouch"
	"github.com/zawa-kyo/pouch/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	for _, arg := range args {
		if arg == "--version" {
			fmt.Fprintln(stdout, version)
			return 0
		}
	}

	config, err := cli.Parse(args, stdout, stderr)
	if err != nil {
		if errors.Is(err, os.ErrInvalid) {
			fmt.Fprintln(stderr, err)
		} else if err.Error() != "flag: help requested" {
			fmt.Fprintln(stderr, err)
		}
		if err.Error() == "flag: help requested" {
			return 0
		}
		return 1
	}

	results, err := pouch.CreateMany(config.Paths, config.Options)
	for _, result := range results {
		if config.Verbose {
			fmt.Fprintln(stdout, formatResult(result))
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func formatResult(result pouch.Result) string {
	switch result.Action {
	case pouch.ActionCreateFile:
		return fmt.Sprintf("create file %s", result.Path)
	case pouch.ActionCreateDir:
		return fmt.Sprintf("create dir %s", result.Path)
	case pouch.ActionSkipExisting:
		if result.Kind == pouch.KindDir {
			return fmt.Sprintf("skip existing dir %s", result.Path)
		}
		return fmt.Sprintf("skip existing file %s", result.Path)
	default:
		return fmt.Sprintf("noop %s", result.Path)
	}
}
