package fonts

import (
	"sort"
	"strings"
)

var baseFamilies = []string{
	"system-ui",
	"ui-sans-serif",
	"ui-monospace",
	"sans-serif",
	"serif",
	"monospace",
}

func ListSystemFonts() []string {
	detected := detectPlatformFonts()
	seen := map[string]bool{}
	var out []string

	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}

	for _, f := range baseFamilies {
		add(f)
	}
	for _, f := range detected {
		add(f)
	}

	sort.Strings(out)
	return out
}
