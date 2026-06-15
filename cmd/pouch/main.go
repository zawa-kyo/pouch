package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zawa-kyo/pouch/internal/cli"
	"github.com/zawa-kyo/pouch/internal/pouch"
)

var version = "dev"

const infoPrefix = "󰄳"
const warnPrefix = "󰅙"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	config, err := cli.Parse(args, stdout, stderr)
	if err != nil {
		if errors.Is(err, os.ErrInvalid) {
			warn(stderr, err)
		} else if err.Error() != "flag: help requested" {
			warn(stderr, err)
		}
		if err.Error() == "flag: help requested" {
			return 0
		}
		return 1
	}
	if config.ShowVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	results, err := pouch.CreateMany(config.Paths, config.Options)
	for _, result := range results {
		if config.Verbose {
			fmt.Fprintln(stdout, formatResult(result))
		}
	}
	if err != nil {
		warn(stderr, err)
		return 1
	}
	return 0
}

func warn(stderr io.Writer, err error) {
	fmt.Fprintf(stderr, "%s %v\n", warnPrefix, err)
}

func formatResult(result pouch.Result) string {
	switch result.Action {
	case pouch.ActionCreateFile:
		return fmt.Sprintf("%s create file %s", infoPrefix, result.Path)
	case pouch.ActionCreateDir:
		return fmt.Sprintf("%s create dir %s", infoPrefix, result.Path)
	case pouch.ActionSkipExisting:
		if result.Kind == pouch.KindDir {
			return fmt.Sprintf("%s skip existing dir %s", infoPrefix, result.Path)
		}
		return fmt.Sprintf("%s skip existing file %s", infoPrefix, result.Path)
	default:
		return fmt.Sprintf("%s noop %s", infoPrefix, result.Path)
	}
}
