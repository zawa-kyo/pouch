package pouch

import (
	"path/filepath"
	"strings"
)

// Detect classifies a path using the final path segment only.
func Detect(path string) Kind {
	if hasTrailingPathSeparator(path) {
		return KindDir
	}
	base := filepath.Base(path)
	if strings.Contains(base, ".") {
		return KindFile
	}
	return KindDir
}

func hasTrailingPathSeparator(path string) bool {
	return strings.HasSuffix(path, string(filepath.Separator))
}
