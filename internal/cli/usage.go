package cli

import (
	"fmt"
	"io"
)

var usageLines = []string{
	"Usage: pouch [flags] PATH...",
	"",
	"Flags:",
	"  -h, --help",
	"      Show help.",
	"  --file",
	"      Use file mode.",
	"  --dir",
	"      Use directory mode.",
	"  -m, --mode auto|file|dir",
	"      Force file or directory mode.",
	"  -n, --dry-run",
	"      Print planned actions without changing the filesystem.",
	"  -s, --strict",
	"      Fail if a target already exists.",
	"  -v, --version",
	"      Show version.",
	"  -V, --verbose",
	"      Print each action in input order.",
}

func writeUsage(w io.Writer) {
	for _, line := range usageLines {
		fmt.Fprintln(w, line)
	}
}
