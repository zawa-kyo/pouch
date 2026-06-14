package pouch

import (
	"path/filepath"
	"strings"
)

func Detect(path string) Kind {
	base := filepath.Base(path)
	if strings.Contains(base, ".") {
		return KindFile
	}
	return KindDir
}
