//go:build linux && !android

// SPDX-License-Identifier: MIT

package fonts

import (
	"os/exec"
	"strings"
)

func detectPlatformFonts() []string {
	if out, err := exec.Command("fc-list", "--format=%{family}\n").Output(); err == nil {
		return parseLines(string(out))
	}
	return scanFontDirs([]string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
	})
}

func parseLines(raw string) []string {
	seen := map[string]bool{}
	var out []string
	for line := range strings.SplitSeq(raw, "\n") {
		family := strings.TrimSpace(line)
		if family == "" {
			continue
		}
		if idx := strings.Index(family, ","); idx >= 0 {
			family = strings.TrimSpace(family[:idx])
		}
		if !seen[family] {
			seen[family] = true
			out = append(out, family)
		}
	}
	return out
}
