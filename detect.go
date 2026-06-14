package pouch

import (
	"path/filepath"
	"strings"
)

// Detect classifies a path using the final path segment only.
func Detect(path string) Kind {
	base := filepath.Base(path)
	if strings.Contains(base, ".") {
		return KindFile
	}
	return KindDir
}
