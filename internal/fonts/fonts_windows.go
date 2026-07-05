//go:build windows

package fonts

import (
	"os"
	"path/filepath"
	"strings"
)

func detectPlatformFonts() []string {
	windir := os.Getenv("WINDIR")
	if windir == "" {
		windir = `C:\Windows`
	}
	fontsDir := filepath.Join(windir, "Fonts")
	entries, err := os.ReadDir(fontsDir)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := fontFamilyFromFilename(e.Name())
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

func fontFamilyFromFilename(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".ttf"),
		strings.HasSuffix(lower, ".otf"),
		strings.HasSuffix(lower, ".ttc"):
		base := strings.TrimSuffix(name, filepath.Ext(name))
		base = strings.ReplaceAll(base, "_", " ")
		base = strings.ReplaceAll(base, "-", " ")
		return strings.TrimSpace(base)
	default:
		return ""
	}
}
