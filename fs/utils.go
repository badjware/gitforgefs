package fs

import (
	"path/filepath"
	"strings"
)

func stripSlash(fn string) string {
	return strings.TrimRight(fn, string(filepath.Separator))
}
