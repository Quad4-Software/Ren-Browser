// SPDX-License-Identifier: MIT
package fonts

import (
	"os"
	"path/filepath"
	"strings"
)

func userHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func scanFontDirs(dirs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			name := fontNameFromPath(d.Name())
			if name != "" && !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
			return nil
		})
	}
	return out
}

func fontNameFromPath(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".ttf"),
		strings.HasSuffix(lower, ".otf"),
		strings.HasSuffix(lower, ".ttc"):
		base := strings.TrimSuffix(filename, filepath.Ext(filename))
		base = strings.TrimPrefix(base, ".")
		base = strings.ReplaceAll(base, "_", " ")
		base = strings.ReplaceAll(base, "-", " ")
		return strings.TrimSpace(base)
	default:
		return ""
	}
}
