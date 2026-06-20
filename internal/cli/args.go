package cli

import (
	"fmt"
	"strings"
)

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
